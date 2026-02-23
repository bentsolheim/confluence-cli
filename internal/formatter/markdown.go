package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// MarkdownFormatter outputs pages as Markdown, suitable for LLM/agent context.
type MarkdownFormatter struct {
	BaseURL string
}

func (f *MarkdownFormatter) FormatPage(w io.Writer, page *confluence.Page) error {
	ap := toAgentPage(page, f.BaseURL)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# [%s](%s)\n\n", ap.Title, ap.WebURL))
	b.WriteString(fmt.Sprintf("- **ID:** %s\n", ap.ID))
	b.WriteString(fmt.Sprintf("- **Space:** %s (%s)\n", ap.SpaceName, ap.SpaceKey))
	b.WriteString(fmt.Sprintf("- **Status:** %s\n", ap.Status))
	b.WriteString(fmt.Sprintf("- **Version:** %d\n", ap.Version))
	if ap.CreatedBy != "" {
		b.WriteString(fmt.Sprintf("- **Created by:** %s\n", ap.CreatedBy))
	}
	if ap.CreatedDate != "" {
		b.WriteString(fmt.Sprintf("- **Created:** %s\n", ap.CreatedDate))
	}
	if ap.UpdatedBy != "" {
		b.WriteString(fmt.Sprintf("- **Updated by:** %s\n", ap.UpdatedBy))
	}
	if ap.UpdatedDate != "" {
		b.WriteString(fmt.Sprintf("- **Updated:** %s\n", ap.UpdatedDate))
	}

	if len(ap.Ancestors) > 0 {
		b.WriteString(fmt.Sprintf("- **Breadcrumb:** %s\n", strings.Join(ap.Ancestors, " → ")))
	}

	if len(ap.Children) > 0 {
		b.WriteString("\n## Child Pages\n\n")
		for _, c := range ap.Children {
			b.WriteString(fmt.Sprintf("- %s (id: %s)\n", c.Title, c.ID))
		}
	}

	body := convertBody(ap.Body)
	if body != "" {
		b.WriteString("\n## Content\n\n")
		b.WriteString(body)
		b.WriteString("\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func (f *MarkdownFormatter) FormatSearchResult(w io.Writer, result *confluence.SearchResult) error {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Search Results (%d of %d)\n\n", result.Size, result.TotalSize))

	headers := []string{"ID", "Space", "Modified", "Title", "Excerpt"}

	// Build rows
	rows := make([][]string, 0, len(result.Results))
	for _, item := range result.Results {
		hit := toAgentSearchHit(&item, f.BaseURL)
		excerpt := strings.ReplaceAll(hit.Excerpt, "\n", " ")
		title := fmt.Sprintf("[%s](%s)", hit.Title, hit.WebURL)
		rows = append(rows, []string{hit.ID, hit.SpaceKey, hit.LastModified, title, excerpt})
	}

	// Compute max width per column
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Write header
	b.WriteString("|")
	for i, h := range headers {
		b.WriteString(fmt.Sprintf(" %-*s |", widths[i], h))
	}
	b.WriteString("\n|")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w+2) + "|")
	}
	b.WriteString("\n")

	// Write rows
	for _, row := range rows {
		b.WriteString("|")
		for i, cell := range row {
			b.WriteString(fmt.Sprintf(" %-*s |", widths[i], cell))
		}
		b.WriteString("\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}
