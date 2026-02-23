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

var (
	maxResults int
	rawCQL     bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for Confluence content",
	Long: `Search for content in Confluence using plain text or raw CQL.

By default, the search terms are used as a text search (siteSearch).
Results are scoped to spaces configured via --spaces or CONFLUENCE_SPACES env.

Use --cql to pass a raw CQL query instead.

Examples:
  confluence search GitHub-pilotering
  confluence search "deployment pipeline"
  confluence search "deployment pipeline" --spaces MUP,DEV
  confluence search --cql "space = DEV AND type = page"
  confluence search --cql "label = backend AND lastModified > now('-30d')"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := keychain.GetPAT(confluenceURL)
		if err != nil {
			return err
		}
		client := confluence.NewClient(confluenceURL, token, verbose)

		var result *confluence.SearchResult
		if rawCQL {
			cql := strings.Join(args, " ")
			result, err = client.Search(cql, maxResults)
		} else {
			query := strings.Join(args, " ")
			cql := buildTextSearchCQL(query, defaultSpaces)
			result, err = client.SiteSearch(cql, maxResults)
		}
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

// buildTextSearchCQL builds a CQL query from plain text and optional space keys.
func buildTextSearchCQL(query, spaces string) string {
	cql := fmt.Sprintf("siteSearch ~ %q AND type = page", query)

	spaceKeys := parseSpaces(spaces)
	if len(spaceKeys) > 0 {
		quoted := make([]string, len(spaceKeys))
		for i, s := range spaceKeys {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		cql += fmt.Sprintf(" AND space in (%s)", strings.Join(quoted, ","))
	}

	return cql
}

// parseSpaces splits a comma-separated spaces string into trimmed, non-empty keys.
func parseSpaces(spaces string) []string {
	if strings.TrimSpace(spaces) == "" {
		return nil
	}
	parts := strings.Split(spaces, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func init() {
	searchCmd.Flags().IntVar(&maxResults, "max-results", 25, "Maximum number of results to return")
	searchCmd.Flags().BoolVar(&rawCQL, "cql", false, "Treat the query as raw CQL")
	rootCmd.AddCommand(searchCmd)
}
