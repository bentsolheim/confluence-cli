package cmd

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/bentsolheim/confluence-cli/internal/formatter"
	"github.com/bentsolheim/confluence-cli/internal/keychain"
	"github.com/spf13/cobra"
)

var pageCmd = &cobra.Command{
	Use:   "page [ID or URL]",
	Short: "Get details of a Confluence page",
	Long: `Fetch full details of a Confluence page by its ID or URL.

You can pass a numeric page ID or copy-paste a full URL from the browser.

Examples:
  confluence page 12345
  confluence page https://wiki.sits.no/spaces/~k77319/pages/1481704261/Page+Title
  confluence page https://wiki.sits.no/display/SPACE/Page+Title
  confluence page "https://wiki.sits.no/pages/viewpage.action?pageId=12345"
  confluence page 12345 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		token, err := keychain.GetPAT(confluenceURL)
		if err != nil {
			return err
		}
		client := confluence.NewClient(confluenceURL, token, verbose)

		page, err := resolveAndFetchPage(client, input)
		if err != nil {
			return err
		}

		f, err := formatter.New(outputFormat, confluenceURL)
		if err != nil {
			return err
		}
		return f.FormatPage(os.Stdout, page)
	},
}

// pagesPathRe matches /spaces/SPACEKEY/pages/PAGEID/...
var pagesPathRe = regexp.MustCompile(`/spaces/([^/]+)/pages/(\d+)`)

// pageRef represents a parsed page reference — either by ID or by space+title.
type pageRef struct {
	ID       string // non-empty if resolved by ID
	SpaceKey string // non-empty (with Title) if resolved by title
	Title    string
}

// parsePageInput parses a page ID or Confluence URL into a pageRef.
func parsePageInput(input string) (pageRef, error) {
	// Plain ID (no slashes or query params)
	if !strings.Contains(input, "/") && !strings.Contains(input, "?") {
		return pageRef{ID: input}, nil
	}

	parsed, err := url.Parse(input)
	if err != nil {
		return pageRef{}, fmt.Errorf("invalid URL %q: %w", input, err)
	}

	path := parsed.Path

	// Format: /pages/viewpage.action?pageId=12345
	if pageID := parsed.Query().Get("pageId"); pageID != "" {
		return pageRef{ID: pageID}, nil
	}

	// Format: /spaces/SPACEKEY/pages/PAGEID/Title
	if m := pagesPathRe.FindStringSubmatch(path); m != nil {
		return pageRef{ID: m[2]}, nil
	}

	// Format: /display/SPACEKEY/Page+Title
	if strings.HasPrefix(path, "/display/") {
		parts := strings.SplitN(strings.TrimPrefix(path, "/display/"), "/", 2)
		if len(parts) == 2 {
			spaceKey := parts[0]
			title, err := url.PathUnescape(parts[1])
			if err != nil {
				title = parts[1]
			}
			title = strings.ReplaceAll(title, "+", " ")
			return pageRef{SpaceKey: spaceKey, Title: title}, nil
		}
	}

	return pageRef{}, fmt.Errorf("could not extract page ID from URL: %s", input)
}

// resolveAndFetchPage parses the input as a page ID or Confluence URL and fetches the page.
func resolveAndFetchPage(client *confluence.Client, input string) (*confluence.Page, error) {
	ref, err := parsePageInput(input)
	if err != nil {
		return nil, err
	}
	if ref.ID != "" {
		return client.GetPage(ref.ID)
	}
	return client.GetPageByTitle(ref.SpaceKey, ref.Title)
}

func init() {
	rootCmd.AddCommand(pageCmd)
}
