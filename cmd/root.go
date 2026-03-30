package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFormat   string
	confluenceURL  string
	verbose        bool
	defaultSpaces  string
)

var rootCmd = &cobra.Command{
	Use:   "confluence",
	Short: "CLI for reading and publishing Confluence content",
	Long: `A command-line tool for your internal Confluence installation.

Read:    Fetch pages by ID/URL and search content via CQL or text.
Write:   Publish Markdown files to Confluence with image uploads.
Diff:    Compare local Markdown against existing pages.

Output formats (for reads): JSON, Markdown, Storage (native Confluence XML),
or plain text. The storage format enables lossless round-trip editing — fetch
with -o storage, edit the XML, then publish back.

Authentication uses a Personal Access Token stored in the macOS Keychain,
with CONFLUENCE_PAT environment variable as a fallback.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "markdown", "Output format: markdown, json, storage, text")
	rootCmd.PersistentFlags().StringVar(&confluenceURL, "url", "https://wiki.sits.no", "Confluence base URL")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show raw HTTP responses from Confluence")
	rootCmd.PersistentFlags().StringVar(&defaultSpaces, "spaces", os.Getenv("CONFLUENCE_SPACES"), "Comma-separated list of space keys to search (env: CONFLUENCE_SPACES)")
}
