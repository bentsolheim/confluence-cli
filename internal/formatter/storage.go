package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// StorageFormatter outputs pages in Confluence's native storage format (XML),
// enabling lossless round-trip fetch→edit→publish.
type StorageFormatter struct {
	BaseURL string
}

func (f *StorageFormatter) FormatPage(w io.Writer, page *confluence.Page) error {
	ap := toAgentPage(page, f.BaseURL)
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("confluence:\n")
	b.WriteString(fmt.Sprintf("  url: %s\n", f.BaseURL))
	b.WriteString(fmt.Sprintf("  pageId: %q\n", ap.ID))
	if ap.SpaceKey != "" {
		b.WriteString(fmt.Sprintf("  spaceKey: %q\n", ap.SpaceKey))
	}
	b.WriteString(fmt.Sprintf("  title: %q\n", ap.Title))
	writeFrontmatterLabels(&b, ap.Labels)
	b.WriteString("  format: storage\n")
	b.WriteString("---\n")

	if ap.Body != "" {
		b.WriteString("\n")
		b.WriteString(FormatStorageXML(ap.Body))
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func (f *StorageFormatter) FormatSearchResult(w io.Writer, result *confluence.SearchResult) error {
	return fmt.Errorf("storage format is not supported for search results (use markdown or json)")
}
