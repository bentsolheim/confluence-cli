package formatter

import (
	"encoding/json"
	"io"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// JSONFormatter outputs pages as JSON.
type JSONFormatter struct {
	BaseURL string
}

func (f *JSONFormatter) FormatPage(w io.Writer, page *confluence.Page) error {
	ap := toAgentPage(page, f.BaseURL)
	ap.Body = convertBody(ap.Body)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ap)
}

func (f *JSONFormatter) FormatSearchResult(w io.Writer, result *confluence.SearchResult) error {
	type output struct {
		Total   int              `json:"total"`
		Count   int              `json:"count"`
		Results []AgentSearchHit `json:"results"`
	}
	o := output{
		Total: result.TotalSize,
		Count: result.Size,
	}
	for _, item := range result.Results {
		o.Results = append(o.Results, toAgentSearchHit(&item, f.BaseURL))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(o)
}
