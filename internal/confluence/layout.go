package confluence

import (
	"fmt"
	"regexp"
	"strings"
)

// WrapInTwoColumnLayout wraps main content in a Confluence two-column layout,
// preserving existing sidebar content if available.
func WrapInTwoColumnLayout(mainContent, existingSidebarContent string) string {
	sidebar := existingSidebarContent
	if sidebar == "" {
		sidebar = getDefaultSidebarTemplate()
	}

	return fmt.Sprintf(`<ac:layout><ac:layout-section ac:type="two_right_sidebar"><ac:layout-cell>
%s</ac:layout-cell>
<ac:layout-cell>
%s</ac:layout-cell>
</ac:layout-section></ac:layout>`, mainContent, sidebar)
}

// ExtractLayoutContent extracts main content and sidebar from an existing
// Confluence page that uses a two-column layout.
func ExtractLayoutContent(html string) (mainContent, sidebarContent string, hasLayout bool) {
	if !strings.Contains(html, `ac:type="two_right_sidebar"`) {
		return html, "", false
	}

	// Extract main content (left column) — first ac:layout-cell
	mainRe := regexp.MustCompile(`(?s)<ac:layout-section[^>]*ac:type="two_right_sidebar"[^>]*>[\s\n]*<ac:layout-cell>(.*?)</ac:layout-cell>[\s\n]*<ac:layout-cell>`)
	if m := mainRe.FindStringSubmatch(html); len(m) > 1 {
		mainContent = strings.TrimSpace(m[1])
	} else {
		// Fallback: try without the two_right_sidebar qualifier
		altRe := regexp.MustCompile(`(?s)<ac:layout-section[^>]*>[\s\n]*<ac:layout-cell>(.*?)</ac:layout-cell>`)
		if m2 := altRe.FindStringSubmatch(html); len(m2) > 1 {
			mainContent = strings.TrimSpace(m2[1])
		}
	}

	// Extract sidebar content (right column) — second ac:layout-cell
	sidebarRe := regexp.MustCompile(`(?s)<ac:layout-section[^>]*ac:type="two_right_sidebar"[^>]*>[\s\n]*<ac:layout-cell>.*?</ac:layout-cell>[\s\n]*<ac:layout-cell>(.*?)</ac:layout-cell>[\s\n]*</ac:layout-section>`)
	if m := sidebarRe.FindStringSubmatch(html); len(m) > 1 {
		sidebarContent = strings.TrimSpace(m[1])
	}

	return mainContent, sidebarContent, true
}

func getDefaultSidebarTemplate() string {
	return `<ac:structured-macro ac:name="toc" ac:schema-version="1"/>`
}
