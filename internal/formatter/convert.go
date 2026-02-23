package formatter

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"golang.org/x/net/html"
)

// confluencePlugin handles Confluence-specific XML elements during HTML-to-Markdown conversion.
type confluencePlugin struct {
	// rawHTML stores the original storage format string so we can extract
	// CDATA content via string parsing (the HTML parser collapses whitespace in CDATA).
	rawHTML string
}

func (p *confluencePlugin) Name() string {
	return "confluence"
}

func (p *confluencePlugin) Init(conv *converter.Converter) error {
	conv.Register.RendererFor("ac:structured-macro", converter.TagTypeBlock, p.handleMacro, converter.PriorityStandard)
	conv.Register.RendererFor("ac:image", converter.TagTypeInline, p.handleImage, converter.PriorityStandard)
	conv.Register.RendererFor("ac:emoticon", converter.TagTypeInline, p.handleEmoticon, converter.PriorityStandard)
	conv.Register.RendererFor("ac:link", converter.TagTypeInline, p.handleLink, converter.PriorityStandard)
	conv.Register.RendererFor("ac:inline-comment-marker", converter.TagTypeInline, p.handleInlineComment, converter.PriorityStandard)
	conv.Register.RendererFor("ac:placeholder", converter.TagTypeInline, p.handlePlaceholder, converter.PriorityStandard)
	conv.Register.RendererFor("time", converter.TagTypeInline, p.handleTime, converter.PriorityStandard)
	return nil
}

// --- Macro handling ---

func (p *confluencePlugin) handleMacro(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	macroName := getAttr(n, "ac:name")
	switch macroName {
	case "code":
		_, _ = w.WriteString(p.handleCodeMacro(n))
	case "markdown":
		// Content is already Markdown inside ac:plain-text-body
		_, _ = w.WriteString(p.handleMarkdownMacro(n))
	case "info":
		_, _ = w.WriteString(p.handleBlockquoteMacro(ctx, n, "ℹ️", "Info"))
	case "warning":
		_, _ = w.WriteString(p.handleBlockquoteMacro(ctx, n, "⚠️", "Warning"))
	case "note":
		_, _ = w.WriteString(p.handleBlockquoteMacro(ctx, n, "📝", "Note"))
	case "tip":
		_, _ = w.WriteString(p.handleBlockquoteMacro(ctx, n, "💡", "Tip"))
	case "expand":
		content := p.convertNestedHTML(ctx, n)
		if content != "" {
			_, _ = w.WriteString(content + "\n\n")
		}
	case "details":
		content := p.convertNestedHTML(ctx, n)
		if content != "" {
			_, _ = w.WriteString(content + "\n\n")
		}
	case "status":
		_, _ = w.WriteString(p.handleStatusMacro(n))
	case "toc":
		_, _ = w.WriteString("<!-- Table of Contents -->")
	case "children":
		_, _ = w.WriteString("<!-- Child Pages -->")
	default:
		_, _ = w.WriteString(fmt.Sprintf("<!-- Unsupported macro: %s -->", macroName))
	}
	return converter.RenderSuccess
}

func (p *confluencePlugin) handleCodeMacro(n *html.Node) string {
	language := ""
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "ac:parameter" && getAttr(child, "ac:name") == "language" {
			if child.FirstChild != nil {
				language = child.FirstChild.Data
			}
		}
	}
	// Extract code from the raw HTML using the macro ID to find the right CDATA block
	code := p.extractPlainTextBody(n)
	code = strings.TrimSpace(code)
	if language != "" {
		return fmt.Sprintf("```%s\n%s\n```\n", language, code)
	}
	return fmt.Sprintf("```\n%s\n```\n", code)
}

func (p *confluencePlugin) handleMarkdownMacro(n *html.Node) string {
	content := p.extractPlainTextBody(n)
	return strings.TrimSpace(content)
}

func (p *confluencePlugin) handleBlockquoteMacro(ctx converter.Context, n *html.Node, emoji, label string) string {
	content := p.convertNestedHTML(ctx, n)
	prefix := fmt.Sprintf("%s **%s:**", emoji, label)
	if content == "" {
		return "> " + prefix
	}
	lines := strings.Split(content, "\n")
	if len(lines) > 1 {
		result := "> " + prefix + "\n"
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result += "> " + line + "\n"
			} else {
				result += ">\n"
			}
		}
		return strings.TrimRight(result, "\n")
	}
	return fmt.Sprintf("> %s %s", prefix, content)
}

func (p *confluencePlugin) handleStatusMacro(n *html.Node) string {
	title, colour := "", ""
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "ac:parameter" {
			name := getAttr(child, "ac:name")
			if name == "title" && child.FirstChild != nil {
				title = child.FirstChild.Data
			} else if name == "colour" && child.FirstChild != nil {
				colour = child.FirstChild.Data
			}
		}
	}
	emoji := ""
	switch strings.ToLower(colour) {
	case "red":
		emoji = "🔴"
	case "yellow":
		emoji = "🟡"
	case "green":
		emoji = "🟢"
	case "blue":
		emoji = "🔵"
	case "grey", "gray":
		emoji = "⚪"
	}
	if title != "" {
		if emoji != "" {
			return fmt.Sprintf("%s **%s**", emoji, title)
		}
		return fmt.Sprintf("**[%s]**", title)
	}
	return ""
}

// --- Image handling ---

func (p *confluencePlugin) handleImage(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	filename := getAttr(n, "ac:filename")
	if filename == "" {
		// Try to find ri:attachment child
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "ri:attachment" {
				filename = getAttr(child, "ri:filename")
				break
			}
		}
	}
	if filename == "" {
		_, _ = w.WriteString("<!-- Image attachment not found -->")
		return converter.RenderSuccess
	}
	_, _ = fmt.Fprintf(w, "![%s](%s)", filename, url.PathEscape(filename))
	return converter.RenderSuccess
}

// --- Emoticon handling ---

func (p *confluencePlugin) handleEmoticon(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	if fallback := getAttr(n, "ac:emoji-fallback"); fallback != "" {
		_, _ = w.WriteString(fallback + " ")
		return converter.RenderTryNext
	}
	if shortname := getAttr(n, "ac:emoji-shortname"); shortname != "" {
		_, _ = w.WriteString(shortname + " ")
		return converter.RenderTryNext
	}
	if name := getAttr(n, "ac:name"); name != "" {
		_, _ = fmt.Fprintf(w, ":%s:", name)
		return converter.RenderTryNext
	}
	_, _ = w.WriteString(":emoji: ")
	return converter.RenderTryNext
}

// --- Link handling ---

func (p *confluencePlugin) handleLink(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "ri:user" {
			accountID := getAttr(child, "ri:account-id")
			userKey := getAttr(child, "ri:userkey")
			if accountID != "" {
				_, _ = fmt.Fprintf(w, " @user(%s) ", accountID)
				return converter.RenderTryNext
			}
			if userKey != "" {
				_, _ = fmt.Fprintf(w, " @user(%s) ", userKey)
				return converter.RenderTryNext
			}
		}
	}
	return converter.RenderTryNext
}

// --- Inline comment handling ---

func (p *confluencePlugin) handleInlineComment(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	// Render child content inline, stripping the comment marker wrapper
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		ctx.RenderNodes(ctx, w, child)
	}
	return converter.RenderSuccess
}

// --- Placeholder handling ---

func (p *confluencePlugin) handlePlaceholder(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	text := extractTextContent(n)
	if text != "" {
		_, _ = fmt.Fprintf(w, "<!-- %s -->", text)
	}
	return converter.RenderSuccess
}

// --- Time handling ---

func (p *confluencePlugin) handleTime(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	if datetime := getAttr(n, "datetime"); datetime != "" {
		_, _ = w.WriteString(datetime + " ")
	}
	return converter.RenderTryNext
}

// --- Helper: convert nested rich-text-body content ---

func (p *confluencePlugin) convertNestedHTML(ctx converter.Context, n *html.Node) string {
	body := findChild(n, "ac:rich-text-body")
	if body == nil {
		return ""
	}
	var buf strings.Builder
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			text := strings.TrimSpace(child.Data)
			if text != "" {
				buf.WriteString(text)
			}
			continue
		}
		if child.Type == html.ElementNode {
			if child.Data == "p" && child.FirstChild == nil {
				continue
			}
			ctx.RenderNodes(ctx, &buf, child)
		}
	}
	return strings.TrimSpace(buf.String())
}

// --- Utility functions ---

// cleanCDATA strips CDATA wrapper artifacts that may leak through the HTML parser.
func cleanCDATA(s string) string {
	s = strings.TrimPrefix(s, "<![CDATA[")
	s = strings.TrimSuffix(s, "]]>")
	return s
}

// extractPlainTextBody extracts the content of <ac:plain-text-body><![CDATA[...]]></ac:plain-text-body>
// from the raw storage HTML using string parsing, preserving newlines that the HTML parser would collapse.
func (p *confluencePlugin) extractPlainTextBody(macroNode *html.Node) string {
	// Try to find the macro in the raw HTML using its macro-id attribute
	macroID := getAttr(macroNode, "ac:macro-id")
	searchHTML := p.rawHTML

	if macroID != "" {
		// Narrow search to this specific macro
		idx := strings.Index(searchHTML, macroID)
		if idx >= 0 {
			searchHTML = searchHTML[idx:]
		}
	}

	// Find <ac:plain-text-body><![CDATA[ ... ]]></ac:plain-text-body>
	const cdataStart = "<ac:plain-text-body><![CDATA["
	const cdataEnd = "]]></ac:plain-text-body>"
	startIdx := strings.Index(searchHTML, cdataStart)
	if startIdx < 0 {
		// Fallback: try without CDATA wrapper
		const altStart = "<ac:plain-text-body>"
		const altEnd = "</ac:plain-text-body>"
		startIdx = strings.Index(searchHTML, altStart)
		if startIdx < 0 {
			// Last resort: use DOM extraction
			return cleanCDATA(extractTextContent(macroNode))
		}
		startIdx += len(altStart)
		endIdx := strings.Index(searchHTML[startIdx:], altEnd)
		if endIdx < 0 {
			return cleanCDATA(extractTextContent(macroNode))
		}
		return searchHTML[startIdx : startIdx+endIdx]
	}
	startIdx += len(cdataStart)
	endIdx := strings.Index(searchHTML[startIdx:], cdataEnd)
	if endIdx < 0 {
		return cleanCDATA(extractTextContent(macroNode))
	}
	return searchHTML[startIdx : startIdx+endIdx]
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func findChild(n *html.Node, name string) *html.Node {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == name {
			return child
		}
		if found := findChild(child, name); found != nil {
			return found
		}
	}
	return nil
}

func extractTextContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return sb.String()
}

// --- Public conversion function ---

// newConfluenceConverter creates an html-to-markdown converter with Confluence plugin.
func newConfluenceConverter(rawHTML string) *converter.Converter {
	return converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			&confluencePlugin{rawHTML: rawHTML},
		),
	)
}

// convertBody converts Confluence storage format (HTML/XML) to Markdown.
// If conversion fails, the raw HTML is returned as-is.
func convertBody(storageHTML string) string {
	storageHTML = strings.TrimSpace(storageHTML)
	if storageHTML == "" {
		return ""
	}
	conv := newConfluenceConverter(storageHTML)
	md, err := conv.ConvertString(storageHTML)
	if err != nil {
		return storageHTML
	}
	return strings.TrimSpace(md)
}
