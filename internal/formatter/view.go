package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// ViewFormatter outputs the rendered HTML of a page (body.view) as is.
// The page must have been fetched with the "body.view" expand.
type ViewFormatter struct {
	BaseURL string
}

func (f *ViewFormatter) FormatPage(w io.Writer, page *confluence.Page) error {
	if page.Body == nil || page.Body.View == nil {
		return fmt.Errorf("page %s has no rendered view body (fetch with expand=body.view)", page.ID)
	}
	body := strings.TrimSpace(page.Body.View.Value)
	if body == "" {
		return nil
	}
	_, err := io.WriteString(w, body+"\n")
	return err
}

func (f *ViewFormatter) FormatSearchResult(w io.Writer, result *confluence.SearchResult) error {
	return fmt.Errorf("view format is not supported for search results (use markdown or json)")
}
