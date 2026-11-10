package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bentsolheim/confluence-cli/internal/auth"
	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/spf13/cobra"
)

var linksMissingOnly bool

var linksCmd = &cobra.Command{
	Use:   "links [ID or URL]",
	Short: "List links on a page and flag links to pages that do not exist",
	Long: `List the links in the rendered version of a page and flag links to pages
that do not exist.

Confluence renders links to non-existent pages as "create page" links. Use
this command after 'confluence publish' to verify that
<ac:link><ri:page ri:content-title="..."/></ac:link> links resolve.

Each link gets one of these types: page, missing, attachment, user, anchor,
internal, external. The command exits with a non-zero status if any link
points to a page that does not exist.

Examples:
  confluence links 12345
  confluence links 12345 --missing-only
  confluence links https://wiki.sits.no/spaces/DEV/pages/12345/Title -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch outputFormat {
		case "markdown", "md", "json", "text":
		default:
			return fmt.Errorf("unsupported output format for links: %q (use markdown, json, or text)", outputFormat)
		}

		token, err := auth.ResolveToken(confluenceURL)
		if err != nil {
			return err
		}
		client := confluence.NewClient(confluenceURL, token, verbose)

		page, err := resolveAndFetchPage(client, args[0], "body.view")
		if err != nil {
			return err
		}
		if page.Body == nil || page.Body.View == nil {
			return fmt.Errorf("page %s has no rendered view body", page.ID)
		}

		all := confluence.ExtractLinks(page.Body.View.Value, confluenceURL)
		report := buildLinkReport(page, all, linksMissingOnly)

		if err := writeLinkReport(os.Stdout, outputFormat, report); err != nil {
			return err
		}
		if report.Missing > 0 {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return fmt.Errorf("%d link(s) point to pages that do not exist", report.Missing)
		}
		return nil
	},
}

// linkReport is the output of the links command.
type linkReport struct {
	PageID  string            `json:"pageId"`
	Title   string            `json:"title"`
	Total   int               `json:"total"`
	Missing int               `json:"missing"`
	Links   []confluence.Link `json:"links"`
}

func buildLinkReport(page *confluence.Page, links []confluence.Link, missingOnly bool) linkReport {
	r := linkReport{PageID: page.ID, Title: page.Title, Total: len(links), Links: []confluence.Link{}}
	for _, l := range links {
		if l.Missing {
			r.Missing++
		}
		if missingOnly && !l.Missing {
			continue
		}
		r.Links = append(r.Links, l)
	}
	return r
}

// linkTarget returns a human-readable target for a link.
func linkTarget(l confluence.Link) string {
	if l.Missing {
		if l.SpaceKey != "" {
			return fmt.Sprintf("%s:%s", l.SpaceKey, l.Title)
		}
		return l.Title
	}
	return l.Href
}

func linkStatus(l confluence.Link) string {
	if l.Missing {
		return "MISSING"
	}
	return "ok"
}

// truncate shortens s to at most n runes, adding "..." when cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-3]) + "..."
}

func writeLinkReport(w io.Writer, format string, r linkReport) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	case "text":
		fmt.Fprintf(w, "%s (id: %s): %d link(s), %d missing\n", r.Title, r.PageID, r.Total, r.Missing)
		if len(r.Links) == 0 {
			return nil
		}
		fmt.Fprintln(w)
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "STATUS\tTYPE\tTEXT\tTARGET")
		for _, l := range r.Links {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", linkStatus(l), l.Type, truncate(l.Text, 60), linkTarget(l))
		}
		return tw.Flush()
	default:
		var b strings.Builder
		fmt.Fprintf(&b, "# Links in %q (%d links, %d missing)\n", r.Title, r.Total, r.Missing)
		if len(r.Links) > 0 {
			b.WriteString("\n| Status | Type | Text | Target |\n")
			b.WriteString("|--------|------|------|--------|\n")
			for _, l := range r.Links {
				fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", linkStatus(l), l.Type, escapeTableCell(l.Text), escapeTableCell(linkTarget(l)))
			}
		}
		_, err := io.WriteString(w, b.String())
		return err
	}
}

func init() {
	linksCmd.Flags().BoolVar(&linksMissingOnly, "missing-only", false, "Only list links to pages that do not exist")
	rootCmd.AddCommand(linksCmd)
}
