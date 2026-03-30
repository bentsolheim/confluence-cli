package cmd

import (
"fmt"
"os"

"github.com/spf13/cobra"

"github.com/bentsolheim/confluence-cli/internal/auth"
"github.com/bentsolheim/confluence-cli/internal/confluence"
"github.com/bentsolheim/confluence-cli/internal/formatter"
"github.com/bentsolheim/confluence-cli/internal/markdown"
)

var diffCmd = &cobra.Command{
Use:   "diff [markdown-file]",
Short: "Show differences between local Markdown and a Confluence page",
Long: `Diff compares a local Markdown file with the existing content on Confluence.

It fetches the remote page, converts its Confluence storage HTML back to Markdown
using the built-in converter, and displays a unified diff.

The page is identified via frontmatter (pageId) or the --url flag.`,
Args: cobra.ExactArgs(1),
RunE: runDiff,
}

var diffPageURL string

func init() {
rootCmd.AddCommand(diffCmd)

diffCmd.Flags().StringVar(&diffPageURL, "url", "", "Confluence page URL (overrides frontmatter)")
}

func runDiff(cmd *cobra.Command, args []string) error {
markdownFile := args[0]

parser := markdown.NewParser()
doc, err := parser.ParseFile(markdownFile)
if err != nil {
return fmt.Errorf("parsing markdown file: %w", err)
}

// Determine page configuration
cfg := doc.Frontmatter.Confluence
if diffPageURL != "" {
pageID, baseURL, err := parsePublishURL(diffPageURL)
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

// Resolve auth and create client
token, err := auth.ResolveToken(cfg.URL)
if err != nil {
return err
}
client := confluence.NewClient(cfg.URL, token, verbose)

fmt.Fprintf(os.Stderr, "🔍 Comparing local file with Confluence page...\n")
fmt.Fprintf(os.Stderr, "   File:    %s\n", markdownFile)
fmt.Fprintf(os.Stderr, "   Page ID: %s\n", cfg.PageID)

// Fetch existing page
existingPage, err := client.GetPage(cfg.PageID)
if err != nil {
return fmt.Errorf("fetching existing page: %w", err)
}

// Convert existing Confluence storage HTML → Markdown using the v2 converter
existingHTML := ""
if existingPage.Body != nil && existingPage.Body.Storage != nil {
existingHTML = existingPage.Body.Storage.Value
}
existingMarkdown := formatter.ConvertBody(existingHTML)

// Show diff
fmt.Fprintf(os.Stderr, "\n")
return confluence.ShowDiff(existingMarkdown, doc.Content, markdownFile)
}
