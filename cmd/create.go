package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bentsolheim/confluence-cli/internal/auth"
	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/bentsolheim/confluence-cli/internal/markdown"
)

var createCmd = &cobra.Command{
	Use:   "create [markdown-file]",
	Short: "Create a new Confluence page from a Markdown file",
	Long: `Create converts a Markdown file to HTML and creates a new Confluence page.

The space key and title are required — provide them via frontmatter or flags.
After creation the new page ID is printed so you can add it to your frontmatter
for future publish/diff operations.

Example:
  confluence create page.md --space "~username" --title "My Page"`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

var (
	createSpaceKey string
	createTitle    string
	createParentID string
	createURL      string
	createDryRun   bool
	createForce    bool
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&createSpaceKey, "space", "", "Confluence space key (e.g. '~username' for personal space)")
	createCmd.Flags().StringVar(&createTitle, "title", "", "Page title (overrides frontmatter)")
	createCmd.Flags().StringVar(&createParentID, "parent-id", "", "Parent page ID (creates as child page)")
	createCmd.Flags().StringVar(&createURL, "url", "", "Confluence base URL")
	createCmd.Flags().BoolVar(&createDryRun, "dry-run", false, "Preview without creating")
	createCmd.Flags().BoolVar(&createForce, "force", false, "Skip confirmation prompts")
}

func runCreate(cmd *cobra.Command, args []string) error {
	markdownFile := args[0]

	parser := markdown.NewParser()
	doc, err := parser.ParseFile(markdownFile)
	if err != nil {
		return fmt.Errorf("parsing markdown file: %w", err)
	}

	// Determine configuration — flags override frontmatter
	cfg := doc.Frontmatter.Confluence
	if createURL != "" {
		cfg.URL = createURL
	}
	if createSpaceKey != "" {
		cfg.SpaceKey = createSpaceKey
	}
	if createTitle != "" {
		cfg.Title = createTitle
	}
	if cfg.URL == "" {
		cfg.URL = confluenceURL
	}
	if cfg.URL == "" {
		return fmt.Errorf("Confluence URL not specified (use --url flag or add to frontmatter)")
	}
	if cfg.SpaceKey == "" {
		return fmt.Errorf("space key not specified (use --space flag or add spaceKey to frontmatter)")
	}
	if cfg.Title == "" {
		return fmt.Errorf("page title not specified (use --title flag or add title to frontmatter)")
	}

	// Resolve auth and create client
	token, err := auth.ResolveToken(cfg.URL)
	if err != nil {
		return err
	}
	client := confluence.NewClient(cfg.URL, token, verbose)

	fmt.Fprintf(os.Stderr, "Creating new Confluence page:\n")
	fmt.Fprintf(os.Stderr, "   URL:   %s\n", cfg.URL)
	fmt.Fprintf(os.Stderr, "   Space: %s\n", cfg.SpaceKey)
	fmt.Fprintf(os.Stderr, "   Title: %s\n", cfg.Title)
	if createParentID != "" {
		fmt.Fprintf(os.Stderr, "   Parent page ID: %s\n", createParentID)
	}

	// Default image handling
	if !cfg.OverwriteAttachments {
		cfg.OverwriteAttachments = true
	}
	if !cfg.FailOnMissingImages {
		cfg.FailOnMissingImages = true
	}

	// Extract images
	imageRefs, err := markdown.ExtractImages(doc.HTML)
	if err != nil {
		return fmt.Errorf("extracting images: %w", err)
	}

	if createDryRun {
		fmt.Fprintf(os.Stderr, "\n=== DRY RUN ===\n")
		fmt.Fprintf(os.Stderr, "Would create page '%s' in space '%s'\n", cfg.Title, cfg.SpaceKey)
		if len(imageRefs) > 0 {
			fmt.Fprintf(os.Stderr, "Found %d image(s) to upload after creation.\n", len(imageRefs))
		}
		fmt.Fprintf(os.Stderr, "\nHTML preview (first 500 chars):\n%s\n", truncateStr(doc.HTML, 500))
		return nil
	}

	if !createForce {
		fmt.Fprint(os.Stderr, "\nContinue with creation? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Fprintln(os.Stderr, "Create cancelled.")
			return nil
		}
	}

	// Create page
	createdPage, err := client.CreatePage(cfg.SpaceKey, cfg.Title, doc.HTML, createParentID)
	if err != nil {
		return fmt.Errorf("creating page: %w", err)
	}

	pageID := createdPage.ID
	fmt.Fprintf(os.Stderr, "✅ Page created with ID: %s\n", pageID)

	// Upload images and update page with macros
	if len(imageRefs) > 0 {
		fmt.Fprintf(os.Stderr, "Uploading %d image(s)...\n", len(imageRefs))
		baseDir := filepath.Dir(markdownFile)
		if cfg.AssetsBase != "" {
			if filepath.IsAbs(cfg.AssetsBase) {
				baseDir = cfg.AssetsBase
			} else {
				baseDir = filepath.Join(baseDir, cfg.AssetsBase)
			}
		}

		replacements := map[string]string{}
		for _, ref := range imageRefs {
			fmt.Fprintf(os.Stderr, "   • %s\n", markdown.DebugFormat(ref))
			if ref.IsExternal || ref.IsDataURI {
				replacements[ref.OriginalTag] = confluence.BuildImageMacro(ref, "")
				continue
			}

			chosenPath := filepath.Join(baseDir, ref.Src)
			if _, statErr := os.Stat(chosenPath); statErr != nil {
				if strings.Contains(ref.Src, "%") {
					if decoded, decErr := url.PathUnescape(ref.Src); decErr == nil && decoded != ref.Src {
						altPath := filepath.Join(baseDir, decoded)
						if _, altErr := os.Stat(altPath); altErr == nil {
							chosenPath = altPath
						}
					}
				}
				if _, statErr2 := os.Stat(chosenPath); statErr2 != nil {
					if cfg.FailOnMissingImages {
						return fmt.Errorf("image not found: %s", chosenPath)
					}
					fmt.Fprintf(os.Stderr, "   [WARN] Missing image (skipping): %s\n", chosenPath)
					continue
				}
			}

			uploadedName, upErr := client.UploadAttachment(pageID, chosenPath, cfg.OverwriteAttachments)
			if upErr != nil {
				return fmt.Errorf("uploading attachment %s: %w", chosenPath, upErr)
			}
			replacements[ref.OriginalTag] = confluence.BuildImageMacro(ref, uploadedName)
		}

		if len(replacements) > 0 {
			updatedHTML := confluence.ReplaceImageTags(doc.HTML, replacements)
			version := 0
			if createdPage.Version != nil {
				version = createdPage.Version.Number
			}
			_, err = client.UpdatePage(pageID, cfg.Title, updatedHTML, version)
			if err != nil {
				return fmt.Errorf("updating page with image macros: %w", err)
			}
			fmt.Fprintln(os.Stderr, "   Images uploaded and page updated.")
		}
	}

	fmt.Fprintf(os.Stderr, "\n✅ Successfully created!\n")
	fmt.Fprintf(os.Stderr, "   Page ID: %s\n", pageID)

	// Add labels if specified in frontmatter
	if len(cfg.Labels) > 0 {
		if _, err := client.AddLabels(pageID, cfg.Labels); err != nil {
			return fmt.Errorf("adding labels: %w", err)
		}
		fmt.Fprintf(os.Stderr, "   Labels: %s\n", strings.Join(cfg.Labels, ", "))
	}

	fmt.Fprintf(os.Stderr, "\nTip: Add this pageId to your frontmatter for future publish/diff:\n")
	fmt.Fprintf(os.Stderr, "  confluence:\n")
	fmt.Fprintf(os.Stderr, "    pageId: \"%s\"\n", pageID)
	return nil
}
