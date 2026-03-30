package markdown

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

// ExtractImages scans HTML for <img ...></img> tags and returns ImageRef slices.
// It expects the parser to have normalized self-closing <img/> to <img></img>.
func ExtractImages(htmlStr string) ([]confluence.ImageRef, error) {
	re := regexp.MustCompile(`(?is)<img([^>]*)></img>`)
	matches := re.FindAllStringSubmatch(htmlStr, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	attrRe := regexp.MustCompile(`([a-zA-Z0-9:_-]+)="([^"]*)"`)
	var refs []confluence.ImageRef

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		fullTag := m[0]
		attrBlock := strings.TrimSpace(m[1])
		ref := confluence.ImageRef{OriginalTag: fullTag}

		for _, a := range attrRe.FindAllStringSubmatch(attrBlock, -1) {
			if len(a) != 3 {
				continue
			}
			switch strings.ToLower(a[1]) {
			case "src":
				ref.Src = strings.TrimSpace(a[2])
			case "alt":
				ref.Alt = strings.TrimSpace(a[2])
			case "title":
				ref.TitleAttr = strings.TrimSpace(a[2])
			case "width":
				ref.Width = strings.TrimSpace(a[2])
			case "height":
				ref.Height = strings.TrimSpace(a[2])
			}
		}

		if ref.Src == "" {
			continue
		}

		lower := strings.ToLower(ref.Src)
		if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
			ref.IsExternal = true
		}
		if strings.HasPrefix(lower, "data:") {
			ref.IsExternal = true
			ref.IsDataURI = true
		}

		refs = append(refs, ref)
	}
	return refs, nil
}

// DebugFormat returns a human-readable description of an ImageRef.
func DebugFormat(ref confluence.ImageRef) string {
	kind := "local"
	if ref.IsDataURI {
		kind = "data-uri"
	} else if ref.IsExternal {
		kind = "external"
	}
	dim := ref.Width
	if ref.Height != "" {
		if dim != "" {
			dim += "x"
		} else {
			dim = ref.Height
		}
	}
	if dim == "" {
		dim = "(auto)"
	}
	return fmt.Sprintf("%s src=%s alt=%q title=%q size=%s", kind, ref.Src, ref.Alt, ref.TitleAttr, dim)
}
