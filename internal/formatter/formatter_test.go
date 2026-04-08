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
				Title:                "Result One",
				Excerpt:              "First result excerpt",
				URL:                  "/spaces/DEV/pages/100/Result+One",
				FriendlyLastModified: "Feb 06, 2026",
				Content: &confluence.Page{
					ID:    "100",
					Space: &confluence.Space{Key: "DEV"},
				},
			},
			{
				Title:                "Result Two",
				Excerpt:              "Second result excerpt",
				URL:                  "/spaces/OPS/pages/200/Result+Two",
				FriendlyLastModified: "Jan 15, 2026",
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
	for _, format := range []string{"json", "markdown", "md", "storage", "text"} {
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

func TestToAgentSearchHit_SiteSearchFallback(t *testing.T) {
	// siteSearch results have no Content field; ID and space are top-level
	item := &confluence.SearchResultItem{
		ID:    "555",
		Title: "SiteSearch Hit",
		Space: &confluence.Space{Key: "MUP"},
		Links: confluence.PageLinks{WebUI: "/spaces/MUP/pages/555/SiteSearch+Hit"},
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	if hit.ID != "555" {
		t.Errorf("expected ID=555 from top-level, got %q", hit.ID)
	}
	if hit.SpaceKey != "MUP" {
		t.Errorf("expected SpaceKey=MUP from top-level, got %q", hit.SpaceKey)
	}
	if hit.WebURL != "https://wiki.example.com/spaces/MUP/pages/555/SiteSearch+Hit" {
		t.Errorf("expected WebURL from _links.WebUI, got %q", hit.WebURL)
	}
}

func TestToAgentSearchHit_ContentTakesPrecedence(t *testing.T) {
	// When both content and top-level fields exist, content wins
	item := &confluence.SearchResultItem{
		ID:    "top-level-id",
		Title: "Page",
		URL:   "/spaces/DEV/pages/777/Page",
		Content: &confluence.Page{
			ID:    "777",
			Space: &confluence.Space{Key: "DEV"},
		},
		Links: confluence.PageLinks{WebUI: "/fallback"},
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	if hit.ID != "777" {
		t.Errorf("expected content ID=777 to take precedence, got %q", hit.ID)
	}
	if hit.WebURL != "https://wiki.example.com/spaces/DEV/pages/777/Page" {
		t.Errorf("expected URL from item.URL, got %q", hit.WebURL)
	}
}

func TestToAgentSearchHit_StripsHighlightMarkers(t *testing.T) {
	item := &confluence.SearchResultItem{
		Title:   "@@@hl@@@Oppsett@@@endhl@@@ av @@@hl@@@runners@@@endhl@@@",
		Excerpt: "bruk @@@hl@@@av@@@endhl@@@ Harden-@@@hl@@@Runner@@@endhl@@@",
		URL:     "/spaces/MUP/pages/123/Page",
		Content: &confluence.Page{ID: "123"},
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	if hit.Title != "Oppsett av runners" {
		t.Errorf("expected cleaned title, got %q", hit.Title)
	}
	if hit.Excerpt != "bruk av Harden-Runner" {
		t.Errorf("expected cleaned excerpt, got %q", hit.Excerpt)
	}
}

func TestToAgentSearchHit_LastModified_FriendlyPreferred(t *testing.T) {
	item := &confluence.SearchResultItem{
		Title:                "Page",
		URL:                  "/spaces/DEV/pages/1/Page",
		FriendlyLastModified: "Feb 06, 2026",
		LastModified:         "2026-02-06T13:09:56.000+01:00",
		Content: &confluence.Page{
			ID: "1",
			Version: &confluence.Version{
				When: "2026-02-06T13:09:56.000+01:00",
			},
		},
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	if hit.LastModified != "Feb 06, 2026" {
		t.Errorf("expected friendly date, got %q", hit.LastModified)
	}
}

func TestToAgentSearchHit_LastModified_FallbackToVersionWhen(t *testing.T) {
	item := &confluence.SearchResultItem{
		Title: "Page",
		URL:   "/spaces/DEV/pages/1/Page",
		Content: &confluence.Page{
			ID: "1",
			Version: &confluence.Version{
				When: "2026-02-06T13:09:56.000+01:00",
			},
		},
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	// No friendly date, so falls back to content.version.when
	if hit.LastModified != "2026-02-06T13:09:56.000+01:00" {
		t.Errorf("expected version.when fallback, got %q", hit.LastModified)
	}
}

func TestToAgentSearchHit_LastModified_FallbackToRaw(t *testing.T) {
	item := &confluence.SearchResultItem{
		Title:        "Page",
		URL:          "/spaces/DEV/pages/1/Page",
		LastModified: "2026-02-06T13:09:56.000+01:00",
		Content:      &confluence.Page{ID: "1"},
	}
	hit := toAgentSearchHit(item, "https://wiki.example.com")
	if hit.LastModified != "2026-02-06T13:09:56.000+01:00" {
		t.Errorf("expected raw lastModified fallback, got %q", hit.LastModified)
	}
}

func TestCleanHighlightMarkers(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"no markers here", "no markers here"},
		{"@@@hl@@@bold@@@endhl@@@", "bold"},
		{"before @@@hl@@@mid@@@endhl@@@ after", "before mid after"},
		{"@@@hl@@@a@@@endhl@@@ @@@hl@@@b@@@endhl@@@", "a b"},
		{"", ""},
	}
	for _, tt := range tests {
		got := cleanHighlightMarkers(tt.input)
		if got != tt.want {
			t.Errorf("cleanHighlightMarkers(%q) = %q, want %q", tt.input, got, tt.want)
		}
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

	// Should start with YAML frontmatter
	if !strings.HasPrefix(out, "---\n") {
		t.Errorf("expected YAML frontmatter start, got %q", out[:40])
	}
	if !strings.Contains(out, "confluence:") {
		t.Errorf("expected 'confluence:' key in frontmatter, got %q", out)
	}
	if !strings.Contains(out, `pageId: "12345"`) {
		t.Errorf("expected pageId in frontmatter, got %q", out)
	}
	if !strings.Contains(out, `spaceKey: "DEV"`) {
		t.Errorf("expected spaceKey in frontmatter, got %q", out)
	}
	if !strings.Contains(out, `title: "Test Page"`) {
		t.Errorf("expected title in frontmatter, got %q", out)
	}
	if !strings.Contains(out, "url: https://wiki.example.com") {
		t.Errorf("expected url in frontmatter, got %q", out)
	}

	// Should contain converted body (no metadata bullets, no ## Content wrapper)
	if !strings.Contains(out, "**world**") {
		t.Errorf("expected converted body, got %q", out)
	}
	if strings.Contains(out, "## Content") {
		t.Errorf("should NOT contain ## Content wrapper, got %q", out)
	}
	if strings.Contains(out, "- **ID:**") {
		t.Errorf("should NOT contain metadata bullets, got %q", out)
	}
	if strings.Contains(out, "## Child Pages") {
		t.Errorf("should NOT contain child pages section, got %q", out)
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
	if !strings.Contains(out, "Modified") {
		t.Errorf("expected Modified column header, got %q", out)
	}
	if !strings.Contains(out, "Result One") {
		t.Errorf("expected first result, got %q", out)
	}
	if !strings.Contains(out, "100") {
		t.Errorf("expected result ID in table, got %q", out)
	}
	if !strings.Contains(out, "Feb 06, 2026") {
		t.Errorf("expected modified date in table, got %q", out)
	}
	// Verify columns are consistently padded
	lines := strings.Split(out, "\n")
	var tablePipeCount int
	for _, line := range lines {
		if strings.HasPrefix(line, "|") {
			count := strings.Count(line, "|")
			if tablePipeCount == 0 {
				tablePipeCount = count
			} else if count != tablePipeCount {
				t.Errorf("inconsistent pipe count: expected %d, got %d in line %q", tablePipeCount, count, line)
			}
		}
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

// --- StorageFormatter ---

func TestStorageFormatter_FormatPage(t *testing.T) {
	var buf bytes.Buffer
	f := &StorageFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatPage(&buf, testPage())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()

	// Should have frontmatter with format: storage
	if !strings.HasPrefix(out, "---\n") {
		t.Errorf("expected YAML frontmatter start, got %q", out[:40])
	}
	if !strings.Contains(out, "format: storage") {
		t.Errorf("expected format: storage in frontmatter, got %q", out)
	}
	if !strings.Contains(out, `pageId: "12345"`) {
		t.Errorf("expected pageId in frontmatter, got %q", out)
	}
	if !strings.Contains(out, `title: "Test Page"`) {
		t.Errorf("expected title in frontmatter, got %q", out)
	}

	// Body should be raw storage format, NOT converted to markdown
	// Content inside <p> is now indented
	if !strings.Contains(out, "Hello <strong>world</strong>") {
		t.Errorf("expected raw storage XML body, got %q", out)
	}
	// Should NOT contain markdown conversion artifacts
	if strings.Contains(out, "**world**") {
		t.Errorf("should NOT contain markdown-converted body, got %q", out)
	}
	// Should be formatted (the body ends with newline from FormatStorageXML)
	if !strings.Contains(out, "format: storage") {
		t.Errorf("expected format: storage in frontmatter, got %q", out)
	}
}

func TestStorageFormatter_FormatSearchResult_ReturnsError(t *testing.T) {
	f := &StorageFormatter{BaseURL: "https://wiki.example.com"}
	err := f.FormatSearchResult(&bytes.Buffer{}, testSearchResult())
	if err == nil {
		t.Fatal("expected error for storage format search results")
	}
}

// --- Label support ---

func testPageWithLabels() *confluence.Page {
	page := testPage()
	page.Metadata = &confluence.PageMetadata{
		Labels: &confluence.LabelListResponse{
			Results: []confluence.Label{
				{Prefix: "global", Name: "backend"},
				{Prefix: "global", Name: "architecture"},
			},
		},
	}
	return page
}

func TestToAgentPage_ExtractsLabels(t *testing.T) {
	ap := toAgentPage(testPageWithLabels(), "https://wiki.example.com")
	if len(ap.Labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(ap.Labels))
	}
	if ap.Labels[0] != "backend" || ap.Labels[1] != "architecture" {
		t.Errorf("expected [backend, architecture], got %v", ap.Labels)
	}
}

func TestToAgentPage_NoLabels(t *testing.T) {
	ap := toAgentPage(testPage(), "https://wiki.example.com")
	if len(ap.Labels) != 0 {
		t.Errorf("expected no labels, got %v", ap.Labels)
	}
}

func TestMarkdownFormatter_IncludesLabels(t *testing.T) {
	var buf bytes.Buffer
	f := &MarkdownFormatter{BaseURL: "https://wiki.example.com"}
	if err := f.FormatPage(&buf, testPageWithLabels()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "labels:") {
		t.Errorf("expected labels in frontmatter, got %q", out)
	}
	if !strings.Contains(out, "- backend") {
		t.Errorf("expected backend label, got %q", out)
	}
	if !strings.Contains(out, "- architecture") {
		t.Errorf("expected architecture label, got %q", out)
	}
}

func TestMarkdownFormatter_NoLabelsOmitted(t *testing.T) {
	var buf bytes.Buffer
	f := &MarkdownFormatter{BaseURL: "https://wiki.example.com"}
	if err := f.FormatPage(&buf, testPage()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(buf.String(), "labels:") {
		t.Errorf("should not contain labels when none exist")
	}
}

func TestStorageFormatter_IncludesLabels(t *testing.T) {
	var buf bytes.Buffer
	f := &StorageFormatter{BaseURL: "https://wiki.example.com"}
	if err := f.FormatPage(&buf, testPageWithLabels()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "labels:") {
		t.Errorf("expected labels in frontmatter, got %q", out)
	}
	if !strings.Contains(out, "- backend") {
		t.Errorf("expected backend label, got %q", out)
	}
	// Labels should come before format: storage
	labelsIdx := strings.Index(out, "labels:")
	formatIdx := strings.Index(out, "format: storage")
	if labelsIdx > formatIdx {
		t.Errorf("labels should appear before format: storage in frontmatter")
	}
}

func TestJSONFormatter_IncludesLabels(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{BaseURL: "https://wiki.example.com"}
	if err := f.FormatPage(&buf, testPageWithLabels()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var ap AgentPage
	if err := json.Unmarshal(buf.Bytes(), &ap); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(ap.Labels) != 2 {
		t.Fatalf("expected 2 labels in JSON, got %d", len(ap.Labels))
	}
	if ap.Labels[0] != "backend" || ap.Labels[1] != "architecture" {
		t.Errorf("expected [backend, architecture], got %v", ap.Labels)
	}
}

func TestTextFormatter_IncludesLabels(t *testing.T) {
	var buf bytes.Buffer
	f := &TextFormatter{BaseURL: "https://wiki.example.com"}
	if err := f.FormatPage(&buf, testPageWithLabels()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Labels: backend, architecture") {
		t.Errorf("expected labels line in text output, got %q", out)
	}
}

func TestTextFormatter_NoLabelsOmitted(t *testing.T) {
	var buf bytes.Buffer
	f := &TextFormatter{BaseURL: "https://wiki.example.com"}
	if err := f.FormatPage(&buf, testPage()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(buf.String(), "Labels:") {
		t.Errorf("should not contain Labels line when none exist")
	}
}
