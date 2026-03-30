package markdown

import (
	"strings"
	"testing"
)

func TestParse_StorageFormat_SkipsConversion(t *testing.T) {
	input := "---\nconfluence:\n  url: https://wiki.example.com\n  pageId: \"123\"\n  title: \"Test\"\n  format: storage\n---\n\n<p>Hello <ac:structured-macro ac:name=\"jira\"><ac:parameter ac:name=\"key\">PROJ-1</ac:parameter></ac:structured-macro></p>"

	p := NewParser()
	doc, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter.Confluence.Format != "storage" {
		t.Errorf("expected format=storage, got %q", doc.Frontmatter.Confluence.Format)
	}

	// HTML should be the raw content, not goldmark-converted
	if !strings.Contains(doc.HTML, "ac:structured-macro") {
		t.Errorf("expected raw Confluence XML preserved, got %q", doc.HTML)
	}
	if strings.Contains(doc.HTML, "&lt;") {
		t.Errorf("should NOT have HTML-escaped the storage content, got %q", doc.HTML)
	}
}

func TestParse_MarkdownFormat_ConvertsNormally(t *testing.T) {
	input := "---\nconfluence:\n  url: https://wiki.example.com\n  pageId: \"123\"\n  title: \"Test\"\n---\n\n# Hello World"

	p := NewParser()
	doc, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter.Confluence.Format != "" {
		t.Errorf("expected empty format, got %q", doc.Frontmatter.Confluence.Format)
	}

	if !strings.Contains(doc.HTML, "<h1") {
		t.Errorf("expected goldmark-converted HTML, got %q", doc.HTML)
	}
}
