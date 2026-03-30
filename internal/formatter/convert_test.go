package formatter

import (
	"strings"
	"testing"
)

func TestConvertBody_Empty(t *testing.T) {
	if got := ConvertBody(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	if got := ConvertBody("   "); got != "" {
		t.Errorf("expected empty string for whitespace, got %q", got)
	}
}

func TestConvertBody_PlainHTML(t *testing.T) {
	got := ConvertBody("<p>Hello <strong>world</strong></p>")
	if !strings.Contains(got, "Hello") || !strings.Contains(got, "**world**") {
		t.Errorf("expected markdown with bold, got %q", got)
	}
}

func TestConvertBody_HeadingsAndParagraphs(t *testing.T) {
	input := "<h1>Title</h1><p>Some text</p><h2>Subtitle</h2><p>More text</p>"
	got := ConvertBody(input)
	if !strings.Contains(got, "# Title") {
		t.Errorf("expected h1 conversion, got %q", got)
	}
	if !strings.Contains(got, "## Subtitle") {
		t.Errorf("expected h2 conversion, got %q", got)
	}
}

func TestConvertBody_CodeMacro_WithLanguage(t *testing.T) {
	input := `<ac:structured-macro ac:name="code" ac:macro-id="abc123">
		<ac:parameter ac:name="language">go</ac:parameter>
		<ac:plain-text-body><![CDATA[fmt.Println("hello")]]></ac:plain-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "```go") {
		t.Errorf("expected fenced code block with language, got %q", got)
	}
	if !strings.Contains(got, `fmt.Println("hello")`) {
		t.Errorf("expected code content preserved, got %q", got)
	}
	if !strings.Contains(got, "```") {
		t.Errorf("expected closing fence, got %q", got)
	}
}

func TestConvertBody_CodeMacro_WithoutLanguage(t *testing.T) {
	input := `<ac:structured-macro ac:name="code" ac:macro-id="abc123">
		<ac:plain-text-body><![CDATA[some code]]></ac:plain-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "```\nsome code\n```") {
		t.Errorf("expected fenced code block without language, got %q", got)
	}
}

func TestConvertBody_CodeMacro_PreservesNewlines(t *testing.T) {
	input := `<ac:structured-macro ac:name="code" ac:macro-id="abc123">
		<ac:parameter ac:name="language">python</ac:parameter>
		<ac:plain-text-body><![CDATA[def hello():
    print("hello")
    return True]]></ac:plain-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "def hello():") {
		t.Errorf("expected first line preserved, got %q", got)
	}
	if !strings.Contains(got, `    print("hello")`) {
		t.Errorf("expected indented line preserved, got %q", got)
	}
}

func TestConvertBody_MarkdownMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="markdown" ac:macro-id="md1">
		<ac:plain-text-body><![CDATA[## Heading

Some **bold** text.

- item 1
- item 2]]></ac:plain-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "## Heading") {
		t.Errorf("expected heading preserved, got %q", got)
	}
	if !strings.Contains(got, "**bold**") {
		t.Errorf("expected bold preserved, got %q", got)
	}
	if !strings.Contains(got, "- item 1") {
		t.Errorf("expected list items preserved, got %q", got)
	}
}

func TestConvertBody_MarkdownMacro_PreservesNewlines(t *testing.T) {
	input := `<ac:structured-macro ac:name="markdown" ac:macro-id="md1">
		<ac:plain-text-body><![CDATA[## First

Paragraph one.

## Second

Paragraph two.]]></ac:plain-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	// Newlines between sections must be preserved
	if !strings.Contains(got, "## First\n\nParagraph one.") {
		t.Errorf("expected newlines between heading and paragraph, got %q", got)
	}
	if !strings.Contains(got, "## Second\n\nParagraph two.") {
		t.Errorf("expected newlines between second heading and paragraph, got %q", got)
	}
}

func TestConvertBody_InfoMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="info">
		<ac:rich-text-body><p>Important note here</p></ac:rich-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "> ℹ️ **Info:**") {
		t.Errorf("expected info blockquote prefix, got %q", got)
	}
	if !strings.Contains(got, "Important note here") {
		t.Errorf("expected info content, got %q", got)
	}
}

func TestConvertBody_WarningMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="warning">
		<ac:rich-text-body><p>Danger!</p></ac:rich-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "⚠️ **Warning:**") {
		t.Errorf("expected warning prefix, got %q", got)
	}
	if !strings.Contains(got, "Danger!") {
		t.Errorf("expected warning content, got %q", got)
	}
}

func TestConvertBody_NoteMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="note">
		<ac:rich-text-body><p>Remember this</p></ac:rich-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "📝 **Note:**") {
		t.Errorf("expected note prefix, got %q", got)
	}
}

func TestConvertBody_TipMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="tip">
		<ac:rich-text-body><p>Pro tip</p></ac:rich-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "💡 **Tip:**") {
		t.Errorf("expected tip prefix, got %q", got)
	}
}

func TestConvertBody_StatusMacro_Green(t *testing.T) {
	input := `<ac:structured-macro ac:name="status">
		<ac:parameter ac:name="title">Done</ac:parameter>
		<ac:parameter ac:name="colour">Green</ac:parameter>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "🟢") {
		t.Errorf("expected green emoji, got %q", got)
	}
	if !strings.Contains(got, "**Done**") {
		t.Errorf("expected bold title, got %q", got)
	}
}

func TestConvertBody_StatusMacro_Red(t *testing.T) {
	input := `<ac:structured-macro ac:name="status">
		<ac:parameter ac:name="title">Failed</ac:parameter>
		<ac:parameter ac:name="colour">Red</ac:parameter>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "🔴 **Failed**") {
		t.Errorf("expected red status badge, got %q", got)
	}
}

func TestConvertBody_StatusMacro_NoColour(t *testing.T) {
	input := `<ac:structured-macro ac:name="status">
		<ac:parameter ac:name="title">Pending</ac:parameter>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "**[Pending]**") {
		t.Errorf("expected fallback status format, got %q", got)
	}
}

func TestConvertBody_TocMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="toc"></ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "<!-- Table of Contents -->") {
		t.Errorf("expected TOC comment, got %q", got)
	}
}

func TestConvertBody_ChildrenMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="children"></ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "<!-- Child Pages -->") {
		t.Errorf("expected children comment, got %q", got)
	}
}

func TestConvertBody_UnsupportedMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="somefuturemacro"></ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "<!-- Unsupported macro: somefuturemacro -->") {
		t.Errorf("expected unsupported macro comment, got %q", got)
	}
}

func TestConvertBody_JiraMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="jira" ac:schema-version="1"><ac:parameter ac:name="server">Aurora JIRA</ac:parameter><ac:parameter ac:name="serverId">1cca82b1-f588-33c4-acd7-802ed9996b69</ac:parameter><ac:parameter ac:name="key">MUP-1192</ac:parameter></ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "MUP-1192") {
		t.Errorf("expected Jira key MUP-1192, got %q", got)
	}
}

func TestConvertBody_ExpandMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="expand">
		<ac:rich-text-body><p>Hidden content</p></ac:rich-text-body>
	</ac:structured-macro>`
	got := ConvertBody(input)
	if !strings.Contains(got, "Hidden content") {
		t.Errorf("expected expanded content, got %q", got)
	}
}

func TestConvertBody_Table(t *testing.T) {
	input := `<table><tbody>
		<tr><th>Name</th><th>Value</th></tr>
		<tr><td>foo</td><td>bar</td></tr>
	</tbody></table>`
	got := ConvertBody(input)
	if !strings.Contains(got, "| Name") || !strings.Contains(got, "| foo") {
		t.Errorf("expected markdown table with pipes, got %q", got)
	}
}

func TestConvertBody_TableMultiParagraphCell(t *testing.T) {
	input := `<table><tbody><tr><th>Date</th><th>Notes</th></tr><tr><td>2026-03-25</td><td><p>Line one.</p><p>Line two.</p></td></tr></tbody></table>`
	got := ConvertBody(input)
	if !strings.Contains(got, "| Date") {
		t.Errorf("expected markdown table, got %q", got)
	}
	if !strings.Contains(got, "Line one. / Line two.") {
		t.Errorf("expected flattened multi-paragraph cell, got %q", got)
	}
}

func TestConvertBody_Links(t *testing.T) {
	input := `<p>Visit <a href="https://example.com">example</a></p>`
	got := ConvertBody(input)
	if !strings.Contains(got, "[example](https://example.com)") {
		t.Errorf("expected markdown link, got %q", got)
	}
}

func TestConvertBody_Lists(t *testing.T) {
	input := `<ul><li>one</li><li>two</li><li>three</li></ul>`
	got := ConvertBody(input)
	if !strings.Contains(got, "- one") || !strings.Contains(got, "- two") {
		t.Errorf("expected unordered list, got %q", got)
	}
}

func TestCleanCDATA(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"<![CDATA[hello]]>", "hello"},
		{"hello]]>", "hello"},
		{"<![CDATA[hello", "hello"},
		{"plain text", "plain text"},
		{"", ""},
	}
	for _, tt := range tests {
		got := cleanCDATA(tt.input)
		if got != tt.want {
			t.Errorf("cleanCDATA(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestConvertBody_ConfluenceWrappedTable(t *testing.T) {
	input := `<table class="wrapped"><thead><tr><th scope="row">Tabell</th><th> Test</th></tr></thead><tbody><tr><th scope="row">A</th><td>B</td></tr><tr><th scope="row">C</th><td>D</td></tr></tbody></table>`
	got := ConvertBody(input)
	t.Logf("Output:\n%s", got)
	if !strings.Contains(got, "|") {
		t.Errorf("expected markdown table with pipes, got:\n%s", got)
	}
}
