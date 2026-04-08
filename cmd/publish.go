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

var publishCmd = &cobra.Command{
	Use:   "publish [markdown-file]",
	Short: "Publish a Markdown file to an existing Confluence page",
	Long: `Publish converts a Markdown file to HTML and updates an existing Confluence page.

The Confluence configuration (URL, page ID, title) is read from YAML frontmatter
in the Markdown file. Flags can override frontmatter values.

Local images referenced in the Markdown are uploaded as attachments.

Example frontmatter:
  ---
  confluence:
    url: https://wiki.sits.no
    pageId: "12345"
    title: "My Page"
  ---`,
	Args: cobra.ExactArgs(1),
	RunE: runPublish,
}

var (
	publishDryRun bool
	publishURL    string
	publishForce  bool
)

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().BoolVar(&publishDryRun, "dry-run", false, "Preview changes without publishing")
	publishCmd.Flags().StringVar(&publishURL, "url", "", "Confluence page URL (overrides frontmatter)")
	publishCmd.Flags().BoolVar(&publishForce, "force", false, "Skip confirmation prompts")
}

func runPublish(cmd *cobra.Command, args []string) error {
	markdownFile := args[0]

	parser := markdown.NewParser()
	doc, err := parser.ParseFile(markdownFile)
	if err != nil {
		return fmt.Errorf("parsing markdown file: %w", err)
	}

	// Determine configuration — flags override frontmatter
	cfg := doc.Frontmatter.Confluence
	if publishURL != "" {
		pageID, baseURL, err := parsePublishURL(publishURL)
		if err != nil {
			return fmt.Errorf("parsing page URL: %w", err)
		}
		cfg.PageID = pageID
		cfg.URL = baseURL
	}
	if cfg.URL == "" {
		cfg.URL = confluenceURL
	}
	if cfg.URL == "" {
		return fmt.Errorf("Confluence URL not specified (use --url flag or add to frontmatter)")
	}
	if cfg.PageID == "" {
		return fmt.Errorf("page ID not specified (use --url with full page URL or add pageId to frontmatter)")
	}
	if cfg.Title == "" {
		return fmt.Errorf("page title not specified in frontmatter")
	}

	// Resolve auth and create client
	token, err := auth.ResolveToken(cfg.URL)
	if err != nil {
		return err
	}
	client := confluence.NewClient(cfg.URL, token, verbose)

	fmt.Fprintf(os.Stderr, "Publishing to Confluence:\n")
	fmt.Fprintf(os.Stderr, "   URL:     %s\n", cfg.URL)
	fmt.Fprintf(os.Stderr, "   Page ID: %s\n", cfg.PageID)
	fmt.Fprintf(os.Stderr, "   Title:   %s\n", cfg.Title)
	if cfg.Format == "storage" {
		fmt.Fprintf(os.Stderr, "   Format:  storage (native Confluence XML)\n")
	} else {
		fmt.Fprintf(os.Stderr, "   Format:  markdown\n")
	}

	// Default image handling options
	if !cfg.OverwriteAttachments {
		cfg.OverwriteAttachments = true
	}
	if !cfg.FailOnMissingImages {
		cfg.FailOnMissingImages = true
	}

	// Process images (skip for storage format — images are already Confluence macros)
	if cfg.Format != "storage" {
		if err := processAndUploadImages(client, doc, cfg, markdownFile); err != nil {
			return err
		}
	}

	// Get existing page version
	existingPage, err := client.GetPage(cfg.PageID)
	if err != nil {
		return fmt.Errorf("fetching existing page: %w", err)
	}

	currentVersion := 0
	if existingPage.Version != nil {
		currentVersion = existingPage.Version.Number
	}

	fmt.Fprintf(os.Stderr, "   Current version: %d\n", currentVersion)

	if publishDryRun {
		fmt.Fprintf(os.Stderr, "\n=== DRY RUN ===\nWould update to version: %d\n", currentVersion+1)
		fmt.Fprintf(os.Stderr, "\nHTML preview (first 500 chars):\n%s\n", truncateStr(doc.HTML, 500))
		return nil
	}

	if !publishForce {
		fmt.Fprint(os.Stderr, "\nContinue with publish? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Fprintln(os.Stderr, "Publish cancelled.")
			return nil
		}
	}

	_, err = client.UpdatePage(cfg.PageID, cfg.Title, doc.HTML, currentVersion)
	if err != nil {
		return fmt.Errorf("updating page: %w", err)
	}

	// Sync labels if specified in frontmatter
	if len(cfg.Labels) > 0 {
		if _, err := client.AddLabels(cfg.PageID, cfg.Labels); err != nil {
			return fmt.Errorf("syncing labels: %w", err)
		}
		fmt.Fprintf(os.Stderr, "   Labels: %s\n", strings.Join(cfg.Labels, ", "))
	}

	fmt.Fprintf(os.Stderr, "✅ Published successfully (version %d)\n", currentVersion+1)
	return nil
}

// processAndUploadImages extracts images from HTML, uploads local ones as
// attachments, and replaces <img> tags with Confluence macros in doc.HTML.
func processAndUploadImages(client *confluence.Client, doc *markdown.Document, cfg confluence.PublishConfig, markdownFile string) error {
	imageRefs, err := markdown.ExtractImages(doc.HTML)
	if err != nil {
		return fmt.Errorf("extracting images: %w", err)
	}
	if len(imageRefs) == 0 {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d image(s) — processing attachments...\n", len(imageRefs))
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

		// Resolve local file path
		chosenPath := filepath.Join(baseDir, ref.Src)
		if _, statErr := os.Stat(chosenPath); statErr != nil {
			// Try URL-decoded path
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

		uploadedName, err := client.UploadAttachment(cfg.PageID, chosenPath, cfg.OverwriteAttachments)
		if err != nil {
			return fmt.Errorf("uploading attachment %s: %w", chosenPath, err)
		}
		replacements[ref.OriginalTag] = confluence.BuildImageMacro(ref, uploadedName)
	}

	if len(replacements) > 0 {
		doc.HTML = confluence.ReplaceImageTags(doc.HTML, replacements)
	}
	return nil
}

// parsePublishURL extracts page ID and base URL from a Confluence page URL.
func parsePublishURL(pageURL string) (pageID, baseURL string, err error) {
	if pageURL == "" {
		return "", "", fmt.Errorf("page URL is required")
	}
	u, err := url.Parse(pageURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for _, part := range pathParts {
		if isNumericString(part) {
			pageID = part
			break
		}
	}
	// Also check query param
	if pageID == "" {
		if qid := u.Query().Get("pageId"); qid != "" {
			pageID = qid
		}
	}
	if pageID == "" {
		return "", "", fmt.Errorf("cannot extract page ID from URL: %s", pageURL)
	}

	baseURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	return pageID, baseURL, nil
}

func isNumericString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
