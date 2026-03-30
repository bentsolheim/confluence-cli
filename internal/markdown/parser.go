package markdown

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// Frontmatter represents the YAML frontmatter in markdown files.
type Frontmatter struct {
	Confluence confluence.PublishConfig `yaml:"confluence"`
}

// Document represents a parsed markdown document.
type Document struct {
	Frontmatter Frontmatter
	Content     string // raw markdown (without frontmatter)
	HTML        string // converted and Confluence-fixed HTML
}

// Parser handles markdown parsing and conversion.
type Parser struct {
	goldmarkParser goldmark.Markdown
}

// NewParser creates a new markdown parser with GFM extensions.
func NewParser() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	return &Parser{goldmarkParser: md}
}

// ParseFile parses a markdown file with frontmatter.
func (p *Parser) ParseFile(filename string) (*Document, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filename, err)
	}
	return p.Parse(string(content))
}

// Parse parses markdown content with frontmatter.
func (p *Parser) Parse(content string) (*Document, error) {
	doc := &Document{}

	// Extract frontmatter
	if strings.HasPrefix(content, "---\n") {
		parts := strings.SplitN(content, "\n---\n", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid frontmatter format")
		}

		yamlPart := parts[0]
		if len(yamlPart) >= 4 {
			yamlContent := yamlPart[4:] // Skip "---\n"
			trimmedYaml := strings.TrimSpace(yamlContent)
			if trimmedYaml != "" {
				if err := yaml.Unmarshal([]byte(yamlContent), &doc.Frontmatter); err != nil {
					return nil, fmt.Errorf("parsing frontmatter: %w", err)
				}
			}
		}

		doc.Content = strings.TrimSpace(parts[1])
	} else {
		doc.Content = content
	}

	// Storage format: use content as-is (Confluence native XML, no conversion)
	if doc.Frontmatter.Confluence.Format == "storage" {
		doc.HTML = doc.Content
		return doc, nil
	}

	// Convert markdown to HTML
	var buf bytes.Buffer
	if err := p.goldmarkParser.Convert([]byte(doc.Content), &buf); err != nil {
		return nil, fmt.Errorf("converting markdown to HTML: %w", err)
	}

	doc.HTML = fixConfluenceHTML(buf.String())
	return doc, nil
}

// ConvertToHTML converts a markdown string to Confluence-compatible HTML.
func (p *Parser) ConvertToHTML(markdownContent string) (string, error) {
	var buf bytes.Buffer
	if err := p.goldmarkParser.Convert([]byte(markdownContent), &buf); err != nil {
		return "", fmt.Errorf("converting markdown to HTML: %w", err)
	}
	return fixConfluenceHTML(buf.String()), nil
}

// WriteHTML writes HTML content to a writer (useful for debugging).
func (p *Parser) WriteHTML(w io.Writer, markdownContent string) error {
	return p.goldmarkParser.Convert([]byte(markdownContent), w)
}

// transformCodeBlocks converts HTML <pre><code> blocks to Confluence code macros.
func transformCodeBlocks(htmlStr string) string {
	re := regexp.MustCompile(`(?s)<pre><code(?:\s+class="language-([^"]+)")?>([\s\S]*?)</code></pre>`)

	return re.ReplaceAllStringFunc(htmlStr, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}

		language := submatches[1]
		content := decodeHTMLEntities(submatches[2])
		content = strings.TrimSuffix(content, "\n")
		// Escape ]]> sequences to prevent CDATA issues
		content = strings.ReplaceAll(content, "]]>", "]]]]><![CDATA[>")

		var macro strings.Builder
		macro.WriteString(`<ac:structured-macro ac:name="code">`)
		macro.WriteString(`<ac:parameter ac:name="linenumbers">true</ac:parameter>`)
		if language != "" {
			macro.WriteString(`<ac:parameter ac:name="language">`)
			macro.WriteString(language)
			macro.WriteString(`</ac:parameter>`)
		}
		macro.WriteString(`<ac:plain-text-body><![CDATA[`)
		macro.WriteString(content)
		macro.WriteString(`]]></ac:plain-text-body>`)
		macro.WriteString(`</ac:structured-macro>`)
		return macro.String()
	})
}

func decodeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&") // must be last
	return s
}

// fixConfluenceHTML applies Confluence-specific HTML fixes.
func fixConfluenceHTML(htmlStr string) string {
	// Encode bare & in URLs while preserving existing entities
	re1 := regexp.MustCompile(`&([a-z0-9_]+)=([^&"\s]+)`)
	htmlStr = re1.ReplaceAllString(htmlStr, "&amp;$1=$2")

	// Mark existing entities
	htmlStr = strings.ReplaceAll(htmlStr, "&amp;", "___AMP___")
	htmlStr = strings.ReplaceAll(htmlStr, "&lt;", "___LT___")
	htmlStr = strings.ReplaceAll(htmlStr, "&gt;", "___GT___")
	htmlStr = strings.ReplaceAll(htmlStr, "&quot;", "___QUOT___")
	htmlStr = strings.ReplaceAll(htmlStr, "&apos;", "___APOS___")
	htmlStr = strings.ReplaceAll(htmlStr, "&nbsp;", "___NBSP___")
	htmlStr = strings.ReplaceAll(htmlStr, "&ndash;", "___NDASH___")
	htmlStr = strings.ReplaceAll(htmlStr, "&mdash;", "___MDASH___")

	// Encode remaining &
	htmlStr = strings.ReplaceAll(htmlStr, "&", "&amp;")

	// Restore marked entities
	htmlStr = strings.ReplaceAll(htmlStr, "___AMP___", "&amp;")
	htmlStr = strings.ReplaceAll(htmlStr, "___LT___", "&lt;")
	htmlStr = strings.ReplaceAll(htmlStr, "___GT___", "&gt;")
	htmlStr = strings.ReplaceAll(htmlStr, "___QUOT___", "&quot;")
	htmlStr = strings.ReplaceAll(htmlStr, "___APOS___", "&apos;")
	htmlStr = strings.ReplaceAll(htmlStr, "___NBSP___", "&nbsp;")
	htmlStr = strings.ReplaceAll(htmlStr, "___NDASH___", "&ndash;")
	htmlStr = strings.ReplaceAll(htmlStr, "___MDASH___", "&mdash;")

	// Transform code blocks AFTER entity encoding
	htmlStr = transformCodeBlocks(htmlStr)

	// Fix self-closing tags for Confluence
	htmlStr = regexp.MustCompile(`(?i)<hr\s*/?>`).ReplaceAllString(htmlStr, "<hr></hr>")
	htmlStr = regexp.MustCompile(`(?i)<br\s*/?>`).ReplaceAllString(htmlStr, "<br></br>")
	htmlStr = regexp.MustCompile(`(?i)<img([^>]*)/?>`).ReplaceAllString(htmlStr, "<img$1></img>")

	return htmlStr
}
