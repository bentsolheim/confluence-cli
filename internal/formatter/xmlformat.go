package formatter

import (
	"regexp"
	"strings"
)

// blockElements are elements that get their own line with indentation.
var blockElements = map[string]bool{
	// HTML block elements
	"p": true, "div": true, "table": true, "tbody": true, "thead": true,
	"tfoot": true, "tr": true, "td": true, "th": true,
	"ul": true, "ol": true, "li": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"hr": true, "pre": true, "blockquote": true,
	"colgroup": true,
	// Confluence block elements
	"ac:structured-macro": true, "ac:rich-text-body": true,
	"ac:plain-text-body": true, "ac:parameter": true,
	"ac:layout": true, "ac:layout-section": true, "ac:layout-cell": true,
	"ac:task-list": true, "ac:task": true,
	"ac:task-body": true, "ac:task-status": true,
}

// indentContentElems are block elements whose content is always indented
// on a new line, even when the content is inline.
var indentContentElems = map[string]bool{
	"p": true,
}

// lineBreakElems are elements that force a newline after themselves,
// with the next content continuing at the same indent level.
var lineBreakElems = map[string]bool{
	"br": true,
}

type xmlToken struct {
	tokType string // "open", "close", "self-close", "cdata", "comment", "text"
	name    string // tag name (empty for text/cdata/comment)
	raw     string // original text
}

// tokenRe matches XML/HTML tokens in order: CDATA, comments, closing tags,
// self-closing tags, opening tags, and text content.
var tokenRe = regexp.MustCompile(
	`<!\[CDATA\[[\s\S]*?\]\]>` + // CDATA sections
		`|<!--[\s\S]*?-->` + // comments
		`|</[^>]+>` + // closing tags
		`|<[^>]+/>` + // self-closing tags
		`|<[^>]+>` + // opening tags
		`|[^<]+`) // text

// extractTagName returns the element name from a raw tag string.
func extractTagName(tag string) string {
	s := strings.TrimPrefix(tag, "</")
	s = strings.TrimPrefix(s, "<")
	s = strings.TrimSuffix(s, "/>")
	s = strings.TrimSuffix(s, ">")
	s = strings.TrimSpace(s)
	if idx := strings.IndexAny(s, " \t\n\r"); idx >= 0 {
		return s[:idx]
	}
	return s
}

func tokenize(input string) []xmlToken {
	matches := tokenRe.FindAllString(input, -1)
	tokens := make([]xmlToken, 0, len(matches))
	for _, m := range matches {
		var tok xmlToken
		tok.raw = m
		switch {
		case strings.HasPrefix(m, "<![CDATA["):
			tok.tokType = "cdata"
		case strings.HasPrefix(m, "<!--"):
			tok.tokType = "comment"
		case strings.HasPrefix(m, "</"):
			tok.tokType = "close"
			tok.name = extractTagName(m)
		case strings.HasSuffix(strings.TrimSpace(m), "/>"):
			tok.tokType = "self-close"
			tok.name = extractTagName(m)
		case strings.HasPrefix(m, "<"):
			tok.tokType = "open"
			tok.name = extractTagName(m)
		default:
			tok.tokType = "text"
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

// nextIsBlock returns true if the next non-whitespace token is a block element.
func nextIsBlock(tokens []xmlToken, from int) bool {
	for i := from; i < len(tokens); i++ {
		t := tokens[i]
		if t.tokType == "text" && strings.TrimFunc(t.raw, isFormatSpace) == "" {
			continue
		}
		if t.tokType == "cdata" || t.tokType == "text" {
			return false
		}
		return blockElements[t.name]
	}
	return true
}

// isFormatSpace reports whether r is whitespace the formatter may rewrite.
// Non-breaking spaces are content, not formatting, so they are left alone.
func isFormatSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// isBoundary reports whether whitespace next to the token is insignificant:
// block elements, line breaks and comments, where the formatter writes its own
// newlines and indentation.
func isBoundary(t xmlToken) bool {
	switch t.tokType {
	case "comment":
		return true
	case "open", "close", "self-close":
		return blockElements[t.name] || lineBreakElems[t.name]
	}
	return false
}

// normalizeEdge rewrites the whitespace at one end of a text token. Next to a
// boundary it is dropped: the formatter writes its own newline there, which
// also renders as a space when the boundary is a macro used inline, like a
// Jira or status macro. Next to an inline element, whitespace with newlines is
// collapsed to a single space, and whitespace without newlines is kept as it
// is, since it separates words in the rendered page.
func normalizeEdge(ws string, boundary bool) string {
	if ws == "" || boundary {
		return ""
	}
	if strings.ContainsAny(ws, "\n\r") {
		return " "
	}
	return ws
}

// normalizeText rewrites the formatting whitespace in a text token, given
// whether the tokens before and after it are boundaries.
func normalizeText(raw string, afterBoundary, beforeBoundary bool) string {
	left := strings.TrimLeftFunc(raw, isFormatSpace)
	if left == "" {
		// Whitespace only. Between two inline elements it separates them.
		if afterBoundary || beforeBoundary {
			return ""
		}
		return normalizeEdge(raw, false)
	}
	core := strings.TrimRightFunc(left, isFormatSpace)
	lead := raw[:len(raw)-len(left)]
	trail := left[len(core):]
	return normalizeEdge(lead, afterBoundary) + core + normalizeEdge(trail, beforeBoundary)
}

func writeIndent(buf *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		buf.WriteString("  ")
	}
}

// FormatStorageXML pretty-prints Confluence storage format XML.
// It handles namespace prefixes (ac:, ri:) natively, preserves CDATA sections
// and <pre> block content verbatim, and applies block/inline formatting rules.
func FormatStorageXML(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	tokens := tokenize(input)
	var buf strings.Builder
	indent := 0
	inPre := false
	// prevType tracks what was last written: "start", "block", or "inline"
	prevType := "start"

	for i, tok := range tokens {
		// Inside <pre>: preserve everything verbatim
		if inPre {
			buf.WriteString(tok.raw)
			if tok.tokType == "close" && tok.name == "pre" {
				inPre = false
				prevType = "block"
			}
			continue
		}

		isBlock := blockElements[tok.name]

		switch tok.tokType {
		case "cdata":
			if prevType == "block" {
				buf.WriteString("\n")
				writeIndent(&buf, indent)
			}
			buf.WriteString(tok.raw)
			prevType = "inline"

		case "comment":
			if prevType != "start" {
				buf.WriteString("\n")
			}
			writeIndent(&buf, indent)
			buf.WriteString(tok.raw)
			prevType = "block"

		case "open":
			if tok.name == "pre" {
				if prevType != "start" {
					buf.WriteString("\n")
				}
				writeIndent(&buf, indent)
				buf.WriteString(tok.raw)
				inPre = true
				continue
			}

			if isBlock {
				if prevType != "start" {
					buf.WriteString("\n")
				}
				writeIndent(&buf, indent)
				buf.WriteString(tok.raw)
				indent++
				// Elements in indentContentElems always indent their content
				// on a new line. Otherwise, look ahead to decide.
				if indentContentElems[tok.name] || nextIsBlock(tokens, i+1) {
					prevType = "block"
				} else {
					prevType = "inline"
				}
			} else {
				if prevType == "block" {
					buf.WriteString("\n")
					writeIndent(&buf, indent)
				}
				buf.WriteString(tok.raw)
				prevType = "inline"
			}

		case "self-close":
			if isBlock {
				if prevType != "start" {
					buf.WriteString("\n")
				}
				writeIndent(&buf, indent)
				buf.WriteString(tok.raw)
				prevType = "block"
			} else {
				if prevType == "block" {
					buf.WriteString("\n")
					writeIndent(&buf, indent)
				}
				buf.WriteString(tok.raw)
				// Line-break elements force a newline after themselves
				if lineBreakElems[tok.name] {
					buf.WriteString("\n")
					writeIndent(&buf, indent)
				}
				prevType = "inline"
			}

		case "close":
			if isBlock {
				indent--
				if indent < 0 {
					indent = 0
				}
				// Always newline before closing tag of indentContentElems
				if prevType == "block" || indentContentElems[tok.name] {
					buf.WriteString("\n")
					writeIndent(&buf, indent)
				}
				buf.WriteString(tok.raw)
				prevType = "block"
			} else {
				buf.WriteString(tok.raw)
				prevType = "inline"
			}

		case "text":
			// Newlines and indentation from an already formatted page are
			// replaced by the formatter's own, but the whitespace that
			// separates text from inline elements, like the space before a
			// link, must survive.
			afterBoundary := i == 0 || isBoundary(tokens[i-1])
			beforeBoundary := i == len(tokens)-1 || isBoundary(tokens[i+1])
			content := normalizeText(tok.raw, afterBoundary, beforeBoundary)
			if content == "" {
				continue // skip whitespace between block elements
			}
			if prevType == "block" {
				buf.WriteString("\n")
				writeIndent(&buf, indent)
			}
			buf.WriteString(content)
			prevType = "inline"
		}
	}

	result := buf.String()
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}
