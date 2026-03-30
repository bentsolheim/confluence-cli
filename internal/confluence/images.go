package confluence

import (
	"sort"
	"strings"
)

// BuildImageMacro builds a Confluence storage format macro for an image.
// For external URLs or data URIs it uses ri:url; for attachments it uses ri:attachment.
func BuildImageMacro(ref ImageRef, attachmentFilename string) string {
	var b strings.Builder
	b.WriteString("<ac:image")
	if ref.Width != "" {
		b.WriteString(` ac:width="`)
		b.WriteString(ref.Width)
		b.WriteString(`"`)
	}
	b.WriteString(">")
	if ref.IsExternal || ref.IsDataURI {
		b.WriteString(`<ri:url ri:value="`)
		b.WriteString(ref.Src)
		b.WriteString(`"/>`)
	} else {
		b.WriteString(`<ri:attachment ri:filename="`)
		b.WriteString(attachmentFilename)
		b.WriteString(`"/>`)
	}
	b.WriteString("</ac:image>")
	return b.String()
}

// ReplaceImageTags replaces original <img> tags with Confluence image macros.
// Longer tags are replaced first to avoid nested/partial replacements.
func ReplaceImageTags(html string, replacements map[string]string) string {
	keys := make([]string, 0, len(replacements))
	for k := range replacements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, k := range keys {
		html = strings.ReplaceAll(html, k, replacements[k])
	}
	return html
}
