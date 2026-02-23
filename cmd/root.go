package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFormat  string
	confluenceURL string
	verbose       bool
)

var rootCmd = &cobra.Command{
	Use:   "confluence",
	Short: "CLI for querying Confluence pages, optimized for AI agent consumption",
	Long: `A command-line tool that queries your internal Confluence installation
and presents pages in structured formats (JSON, Markdown, text)
suitable for AI/KI agent consumption.

Authentication uses a Personal Access Token stored in the macOS Keychain.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "markdown", "Output format: markdown, json, text")
	rootCmd.PersistentFlags().StringVar(&confluenceURL, "url", "https://wiki.sits.no", "Confluence base URL")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show raw HTTP responses from Confluence")
}
