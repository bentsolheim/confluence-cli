package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/bentsolheim/confluence-cli/internal/formatter"
	"github.com/bentsolheim/confluence-cli/internal/keychain"
	"github.com/spf13/cobra"
)

var maxResults int

var searchCmd = &cobra.Command{
	Use:   "search [CQL]",
	Short: "Search for Confluence content using CQL",
	Long: `Search for content using Confluence Query Language (CQL).

Examples:
  confluence search "space = DEV AND type = page"
  confluence search "title ~ 'architecture'"
  confluence search "label = backend AND lastModified > now('-30d')" --max-results 20 -o markdown`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cql := strings.Join(args, " ")
		token, err := keychain.GetPAT(confluenceURL)
		if err != nil {
			return err
		}
		client := confluence.NewClient(confluenceURL, token, verbose)
		result, err := client.Search(cql, maxResults)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}
		f, err := formatter.New(outputFormat, confluenceURL)
		if err != nil {
			return err
		}
		return f.FormatSearchResult(os.Stdout, result)
	},
}

func init() {
	searchCmd.Flags().IntVar(&maxResults, "max-results", 25, "Maximum number of results to return")
	rootCmd.AddCommand(searchCmd)
}
