package cmd

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/auth"
	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/bentsolheim/confluence-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var pageResolveUsers bool

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
  confluence page 12345 -o json
  confluence page 12345 --resolve-users
  confluence page 12345 -o view

Output formats:
  markdown  Frontmatter + body converted to Markdown (default)
  json      Flattened page metadata with Markdown body
  text      Human-readable summary with Markdown body
  storage   Native Confluence storage XML for lossless round-trip editing
  view      Rendered HTML (body.view) as shown in the browser. Links to pages
            that do not exist have class="createlink" (see 'confluence links')

User mentions are shown as @user(<userkey>). With --resolve-users, each unique
user is looked up and shown as @user(Display Name / username) in markdown,
json and text output. Use 'confluence user' to look up single users.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		token, err := auth.ResolveToken(confluenceURL)
		if err != nil {
			return err
		}
		client := confluence.NewClient(confluenceURL, token, verbose)

		var extraExpand []string
		if outputFormat == "view" {
			extraExpand = append(extraExpand, "body.view")
		}

		page, err := resolveAndFetchPage(client, input, extraExpand...)
		if err != nil {
			return err
		}

		var opts formatter.Options
		if pageResolveUsers {
			switch outputFormat {
			case "storage", "view":
				fmt.Fprintf(os.Stderr, "warning: --resolve-users has no effect with -o %s\n", outputFormat)
			default:
				if page.Body != nil && page.Body.Storage != nil {
					opts.UserNames = resolveUserNames(client, page.Body.Storage.Value)
				}
			}
		}

		f, err := formatter.New(outputFormat, confluenceURL, opts)
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
// Extra expand values (e.g. "body.view") are passed on to the API.
func resolveAndFetchPage(client *confluence.Client, input string, extraExpand ...string) (*confluence.Page, error) {
	ref, err := parsePageInput(input)
	if err != nil {
		return nil, err
	}
	if ref.ID != "" {
		return client.GetPage(ref.ID, extraExpand...)
	}
	return client.GetPageByTitle(ref.SpaceKey, ref.Title, extraExpand...)
}

// resolveUserNames looks up every unique user mentioned in the storage body and
// returns a map from user key/username to "Display Name / username".
// Users that cannot be looked up are reported on stderr and left unresolved.
func resolveUserNames(client *confluence.Client, storage string) map[string]string {
	refs := formatter.ExtractUserRefs(storage)
	names := make(map[string]string, len(refs.Keys)+len(refs.Usernames))
	for _, key := range refs.Keys {
		u, err := client.GetUserByKey(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not look up user key %s: %v\n", key, err)
			continue
		}
		names[key] = formatter.UserLabel(u.DisplayName, u.Username)
	}
	for _, username := range refs.Usernames {
		u, err := client.GetUserByUsername(username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not look up username %s: %v\n", username, err)
			continue
		}
		names[username] = formatter.UserLabel(u.DisplayName, u.Username)
	}
	return names
}

func init() {
	pageCmd.Flags().BoolVar(&pageResolveUsers, "resolve-users", false, "Show @user mentions as names instead of user keys (markdown, json, text)")
	rootCmd.AddCommand(pageCmd)
}
