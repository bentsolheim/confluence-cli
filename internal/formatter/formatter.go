package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// Formatter defines the interface for output formatting.
type Formatter interface {
	FormatPage(w io.Writer, page *confluence.Page) error
	FormatSearchResult(w io.Writer, result *confluence.SearchResult) error
}

// New creates a formatter for the given format name.
func New(format string, baseURL string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{BaseURL: baseURL}, nil
	case "markdown", "md":
		return &MarkdownFormatter{BaseURL: baseURL}, nil
	case "text":
		return &TextFormatter{BaseURL: baseURL}, nil
	default:
		return nil, fmt.Errorf("unknown output format: %q (use json, markdown, or text)", format)
	}
}

// AgentPage is a flattened, agent-friendly representation of a page.
type AgentPage struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	SpaceKey    string       `json:"spaceKey"`
	SpaceName   string       `json:"spaceName"`
	Status      string       `json:"status"`
	Version     int          `json:"version"`
	CreatedBy   string       `json:"createdBy,omitempty"`
	CreatedDate string       `json:"createdDate,omitempty"`
	UpdatedBy   string       `json:"updatedBy,omitempty"`
	UpdatedDate string       `json:"updatedDate,omitempty"`
	WebURL      string       `json:"webUrl"`
	Ancestors   []string     `json:"ancestors,omitempty"`
	Children    []AgentChild `json:"children,omitempty"`
	Body        string       `json:"body,omitempty"`
}

// AgentChild is a simplified child page reference.
type AgentChild struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// AgentSearchHit is a flattened search result item.
type AgentSearchHit struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title"`
	SpaceKey string `json:"spaceKey,omitempty"`
	Excerpt  string `json:"excerpt,omitempty"`
	WebURL   string `json:"webUrl"`
}

func toAgentPage(page *confluence.Page, baseURL string) AgentPage {
	ap := AgentPage{
		ID:     page.ID,
		Title:  page.Title,
		Status: page.Status,
		WebURL: baseURL + page.Links.WebUI,
	}

	if page.Space != nil {
		ap.SpaceKey = page.Space.Key
		ap.SpaceName = page.Space.Name
	}

	if page.Version != nil {
		ap.Version = page.Version.Number
		ap.UpdatedDate = page.Version.When
		if page.Version.By != nil {
			ap.UpdatedBy = page.Version.By.DisplayName
		}
	}

	if page.History != nil {
		ap.CreatedDate = page.History.CreatedDate
		if page.History.CreatedBy != nil {
			ap.CreatedBy = page.History.CreatedBy.DisplayName
		}
	}

	for _, a := range page.Ancestors {
		ap.Ancestors = append(ap.Ancestors, a.Title)
	}

	if page.Children != nil && page.Children.Page != nil {
		for _, c := range page.Children.Page.Results {
			ap.Children = append(ap.Children, AgentChild{
				ID:    c.ID,
				Title: c.Title,
			})
		}
	}

	if page.Body != nil && page.Body.Storage != nil {
		ap.Body = page.Body.Storage.Value
	}

	return ap
}

func toAgentSearchHit(item *confluence.SearchResultItem, baseURL string) AgentSearchHit {
	hit := AgentSearchHit{
		Title:   item.Title,
		Excerpt: strings.TrimSpace(item.Excerpt),
		WebURL:  baseURL + item.URL,
	}
	if item.Content != nil {
		hit.ID = item.Content.ID
		if item.Content.Space != nil {
			hit.SpaceKey = item.Content.Space.Key
		}
	}
	return hit
}
