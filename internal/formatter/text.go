package formatter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// TextFormatter outputs pages as human-readable plain text.
type TextFormatter struct {
	BaseURL string
}

func (f *TextFormatter) FormatPage(w io.Writer, page *confluence.Page) error {
	ap := toAgentPage(page, f.BaseURL)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s\n", ap.Title))
	b.WriteString(strings.Repeat("=", len(ap.Title)) + "\n\n")

	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "ID:\t%s\n", ap.ID)
	fmt.Fprintf(tw, "Space:\t%s (%s)\n", ap.SpaceName, ap.SpaceKey)
	fmt.Fprintf(tw, "Status:\t%s\n", ap.Status)
	fmt.Fprintf(tw, "Version:\t%d\n", ap.Version)
	if ap.CreatedBy != "" {
		fmt.Fprintf(tw, "Created by:\t%s\n", ap.CreatedBy)
	}
	if ap.CreatedDate != "" {
		fmt.Fprintf(tw, "Created:\t%s\n", ap.CreatedDate)
	}
	if ap.UpdatedBy != "" {
		fmt.Fprintf(tw, "Updated by:\t%s\n", ap.UpdatedBy)
	}
	if ap.UpdatedDate != "" {
		fmt.Fprintf(tw, "Updated:\t%s\n", ap.UpdatedDate)
	}
	fmt.Fprintf(tw, "URL:\t%s\n", ap.WebURL)
	tw.Flush()

	if len(ap.Ancestors) > 0 {
		b.WriteString(fmt.Sprintf("\nBreadcrumb: %s\n", strings.Join(ap.Ancestors, " > ")))
	}

	if len(ap.Children) > 0 {
		b.WriteString("\nChild Pages:\n")
		for _, c := range ap.Children {
			b.WriteString(fmt.Sprintf("  - %s (id: %s)\n", c.Title, c.ID))
		}
	}

	body := convertBody(ap.Body)
	if body != "" {
		b.WriteString("\nContent:\n")
		b.WriteString(strings.Repeat("-", 40) + "\n")
		b.WriteString(body + "\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func (f *TextFormatter) FormatSearchResult(w io.Writer, result *confluence.SearchResult) error {
	fmt.Fprintf(w, "Results: %d of %d\n\n", result.Size, result.TotalSize)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSPACE\tMODIFIED\tTITLE\tEXCERPT")
	fmt.Fprintln(tw, "--\t-----\t--------\t-----\t-------")
	for _, item := range result.Results {
		hit := toAgentSearchHit(&item, f.BaseURL)
		excerpt := strings.ReplaceAll(hit.Excerpt, "\n", " ")
		if len(excerpt) > 80 {
			excerpt = excerpt[:77] + "..."
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", hit.ID, hit.SpaceKey, hit.LastModified, hit.Title, excerpt)
	}
	return tw.Flush()
}
