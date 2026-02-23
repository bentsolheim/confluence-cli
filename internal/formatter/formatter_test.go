package formatter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

func testPage() *confluence.Page {
	return &confluence.Page{
		ID:     "12345",
		Title:  "Test Page",
		Status: "current",
		Space:  &confluence.Space{Key: "DEV", Name: "Development"},
		Version: &confluence.Version{
			Number: 3,
			When:   "2026-01-15T10:00:00.000+01:00",
			By:     &confluence.User{DisplayName: "Jane Doe"},
		},
		History: &confluence.PageHistory{
			CreatedDate: "2026-01-10T08:00:00.000+01:00",
			CreatedBy:   &confluence.User{DisplayName: "John Smith"},
		},
		Body: &confluence.Body{
			Storage: &confluence.BodyContent{
				Value: "<p>Hello <strong>world</strong></p>",
			},
		},
		Links: confluence.PageLinks{
			WebUI: "/spaces/DEV/pages/12345/Test+Page",
		},
		Ancestors: []confluence.Page{
			{Title: "Parent Page"},
			{Title: "Grandparent Page"},
		},
		Children: &confluence.PageChildren{
			Page: &confluence.PageResults{
				Results: []confluence.Page{
					{ID: "111", Title: "Child A"},
					{ID: "222", Title: "Child B"},
				},
			},
		},
	}
}

func testSearchResult() *confluence.SearchResult {
	return &confluence.SearchResult{
		TotalSize: 42,
		Size:      2,
		Results: []confluence.SearchResultItem{
			{
				Title:   "Result One",
				Excerpt: "First result excerpt",
				URL:     "/spaces/DEV/pages/100/Result+One",
				Content: &confluence.Page{
					ID:    "100",
					Space: &confluence.Space{Key: "DEV"},
				},
			},
			{
				Title:   "Result Two",
				Excerpt: "Second result excerpt",
				URL:     "/spaces/OPS/pages/200/Result+Two",
				Content: &confluence.Page{
					ID:    "200",
					Space: &confluence.Space{Key: "OPS"},
				},
			},
		},
	}
}

// --- New() ---

func TestNew_ValidFormats(t *testing.T) {
	for _, format := range []string{"json", "markdown", "md", "text"} {
		f, err := New(format, "https://example.com")
		if err != nil {
			t.Errorf("New(%q) returned error: %v", format, err)
		}
		if f == nil {
			t.Errorf("New(%q) returned nil formatter", format)
		}
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	_, err := New("xml", "https://example.com")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error should mention the bad format, got %q", err.Error())
	}
}

// --- toAgentPage ---

func TestToAgentPage_FullPage(t *testing.T) {
	page := testPage()
	ap := toAgentPage(page, "https://wiki.example.com")

	if ap.ID != "12345" {
		t.Errorf("expected ID=12345, got %q", ap.ID)
	}
	if ap.Title != "Test Page" {
		t.Errorf("expected Title='Test Page', got %q", ap.Title)
	}
	if ap.SpaceKey != "DEV" {
		t.Errorf("expected SpaceKey=DEV, got %q", ap.SpaceKey)
	}
	if ap.SpaceName != "Development" {
		t.Errorf("expected SpaceName=Development, got %q", ap.SpaceName)
	}
	if ap.Version != 3 {
		t.Errorf("expected Version=3, got %d", ap.Version)
	}
	if ap.CreatedBy != "John Smith" {
		t.Errorf("expected CreatedBy='John Smith', got %q", ap.CreatedBy)
	}
	if ap.UpdatedBy != "Jane Doe" {
		t.Errorf("expected UpdatedBy='Jane Doe', got %q", ap.UpdatedBy)
	}
	if ap.WebURL != "https://wiki.example.com/spaces/DEV/pages/12345/Test+Page" {
		t.Errorf("expected full WebURL, got %q", ap.WebURL)
	}
	if len(ap.Ancestors) != 2 || ap.Ancestors[0] != "Parent Page" {
		t.Errorf("expected 2 ancestors, got %v", ap.Ancestors)
	}
	if len(ap.Children) != 2 || ap.Children[0].ID != "111" {
		t.Errorf("expected 2 children, got %v", ap.Children)
	}
}

func TestToAgentPage_MinimalPage(t *testing.T) {
	page := &confluence.Page{ID: "1", Title: "Minimal"}
	ap := toAgentPage(page, "https://wiki.example.com")
	if ap.ID != "1" {
		t.Errorf("expected ID=1, got %q", ap.ID)
	}
	if ap.SpaceKey != "" {
		t.Errorf("expected empty SpaceKey, got %q", ap.SpaceKey)
	}
	if ap.Version != 0 {
		t.Errorf("expected Version=0, got %d", ap.Version)
	}
	if len(ap.Ancestors) != 0 {
		t.Errorf("expected no ancestors, got %v", ap.Ancestors)
	}
	if len(ap.Children) != 0 {
		t.Errorf("expected no children, got %v", ap.Children)
	}
}

// --- toAgentSearchHit ---

func TestToAgentSearchHit(t *testing.T) {
	item := &confluence.SearchResultItem{
		Title:   "My Page",
		Excerpt: "  some excerpt  ",
		URL:     "/spaces/DEV/pages/999/My+Page",
		Content: &confluence.Page{
			ID:    "999",
			Space: &confluence.Space{Key: "DEV"},
		},
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	if hit.ID != "999" {
		t.Errorf("expected ID=999, got %q", hit.ID)
	}
	if hit.SpaceKey != "DEV" {
		t.Errorf("expected SpaceKey=DEV, got %q", hit.SpaceKey)
	}
	if hit.Excerpt != "some excerpt" {
		t.Errorf("expected trimmed excerpt, got %q", hit.Excerpt)
	}
	if hit.WebURL != "https://wiki.example.com/spaces/DEV/pages/999/My+Page" {
		t.Errorf("expected full WebURL, got %q", hit.WebURL)
	}
}

func TestToAgentSearchHit_NilContent(t *testing.T) {
	item := &confluence.SearchResultItem{
		Title: "Orphan",
		URL:   "/some/path",
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	if hit.ID != "" {
		t.Errorf("expected empty ID, got %q", hit.ID)
	}
}

// --- JSONFormatter ---

func TestJSONFormatter_FormatPage(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatPage(&buf, testPage())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var ap AgentPage
	if err := json.Unmarshal(buf.Bytes(), &ap); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if ap.ID != "12345" {
		t.Errorf("expected ID=12345 in JSON, got %q", ap.ID)
	}
	if ap.Title != "Test Page" {
		t.Errorf("expected Title in JSON, got %q", ap.Title)
	}
	// Body should be converted to markdown
	if !strings.Contains(ap.Body, "**world**") {
		t.Errorf("expected markdown body in JSON, got %q", ap.Body)
	}
}

func TestJSONFormatter_FormatSearchResult(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatSearchResult(&buf, testSearchResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result struct {
		Total   int              `json:"total"`
		Count   int              `json:"count"`
		Results []AgentSearchHit `json:"results"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result.Total != 42 {
		t.Errorf("expected total=42, got %d", result.Total)
	}
	if result.Count != 2 {
		t.Errorf("expected count=2, got %d", result.Count)
	}
	if len(result.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Results))
	}
}

// --- MarkdownFormatter ---

func TestMarkdownFormatter_FormatPage(t *testing.T) {
	var buf bytes.Buffer
	f := &MarkdownFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatPage(&buf, testPage())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# [Test Page]") {
		t.Errorf("expected markdown title, got %q", out)
	}
	if !strings.Contains(out, "- **ID:** 12345") {
		t.Errorf("expected ID metadata, got %q", out)
	}
	if !strings.Contains(out, "- **Space:** Development (DEV)") {
		t.Errorf("expected space metadata, got %q", out)
	}
	if !strings.Contains(out, "- **Created by:** John Smith") {
		t.Errorf("expected created by, got %q", out)
	}
	if !strings.Contains(out, "- **Breadcrumb:** Parent Page → Grandparent Page") {
		t.Errorf("expected breadcrumb, got %q", out)
	}
	if !strings.Contains(out, "## Child Pages") {
		t.Errorf("expected child pages section, got %q", out)
	}
	if !strings.Contains(out, "- Child A (id: 111)") {
		t.Errorf("expected child A, got %q", out)
	}
	if !strings.Contains(out, "## Content") {
		t.Errorf("expected content section, got %q", out)
	}
	if !strings.Contains(out, "**world**") {
		t.Errorf("expected converted body, got %q", out)
	}
}

func TestMarkdownFormatter_FormatSearchResult(t *testing.T) {
	var buf bytes.Buffer
	f := &MarkdownFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatSearchResult(&buf, testSearchResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# Search Results (2 of 42)") {
		t.Errorf("expected search results header, got %q", out)
	}
	if !strings.Contains(out, "Result One") {
		t.Errorf("expected first result, got %q", out)
	}
	if !strings.Contains(out, "| 100 |") {
		t.Errorf("expected result ID in table, got %q", out)
	}
}

// --- TextFormatter ---

func TestTextFormatter_FormatPage(t *testing.T) {
	var buf bytes.Buffer
	f := &TextFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatPage(&buf, testPage())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Test Page\n") {
		t.Errorf("expected plain text title, got %q", out)
	}
	if !strings.Contains(out, "ID:") {
		t.Errorf("expected ID field, got %q", out)
	}
	if !strings.Contains(out, "Breadcrumb: Parent Page > Grandparent Page") {
		t.Errorf("expected breadcrumb with > separator, got %q", out)
	}
	if !strings.Contains(out, "Child Pages:") {
		t.Errorf("expected child pages section, got %q", out)
	}
	if !strings.Contains(out, "Content:") {
		t.Errorf("expected content section, got %q", out)
	}
}

func TestTextFormatter_FormatSearchResult(t *testing.T) {
	var buf bytes.Buffer
	f := &TextFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatSearchResult(&buf, testSearchResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Results: 2 of 42") {
		t.Errorf("expected result count, got %q", out)
	}
	if !strings.Contains(out, "Result One") {
		t.Errorf("expected first result, got %q", out)
	}
}
