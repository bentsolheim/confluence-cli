package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/auth"
	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/bentsolheim/confluence-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var (
	maxResults int
	rawCQL     bool
	ancestorID string
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for Confluence content",
	Long: `Search for content in Confluence using plain text or raw CQL.

By default, the search terms are used as a text search (siteSearch).
Results are scoped to spaces configured via --spaces or CONFLUENCE_SPACES env.

Use --ancestor to restrict results to descendants of a specific page (by ID).
Use --cql to pass a raw CQL query instead.

Examples:
  confluence search GitHub-pilotering
  confluence search "deployment pipeline"
  confluence search "deployment pipeline" --spaces MUP,DEV
  confluence search "beslutning" --ancestor 997497294
  confluence search --cql "space = DEV AND type = page"
  confluence search --cql "label = backend AND lastModified > now('-30d')"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := auth.ResolveToken(confluenceURL)
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
			cql := buildTextSearchCQL(query, defaultSpaces, ancestorID)
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
func buildTextSearchCQL(query, spaces, ancestor string) string {
	cql := fmt.Sprintf("siteSearch ~ %q AND type = page", query)

	spaceKeys := parseSpaces(spaces)
	if len(spaceKeys) > 0 {
		quoted := make([]string, len(spaceKeys))
		for i, s := range spaceKeys {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		cql += fmt.Sprintf(" AND space in (%s)", strings.Join(quoted, ","))
	}

	if ancestor != "" {
		cql += fmt.Sprintf(" AND ancestor = %s", ancestor)
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
	searchCmd.Flags().StringVar(&ancestorID, "ancestor", "", "Restrict results to descendants of this page ID")
	rootCmd.AddCommand(searchCmd)
}
