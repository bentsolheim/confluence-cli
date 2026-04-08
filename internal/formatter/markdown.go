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

	// YAML frontmatter for round-trip publish workflow
	b.WriteString("---\n")
	b.WriteString("confluence:\n")
	b.WriteString(fmt.Sprintf("  url: %s\n", f.BaseURL))
	b.WriteString(fmt.Sprintf("  pageId: %q\n", ap.ID))
	if ap.SpaceKey != "" {
		b.WriteString(fmt.Sprintf("  spaceKey: %q\n", ap.SpaceKey))
	}
	b.WriteString(fmt.Sprintf("  title: %q\n", ap.Title))
	writeFrontmatterLabels(&b, ap.Labels)
	b.WriteString("---\n")

	body := ConvertBody(ap.Body)
	if body != "" {
		b.WriteString("\n")
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
