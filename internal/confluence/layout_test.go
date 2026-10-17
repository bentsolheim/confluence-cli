package confluence

import (
	"strings"
	"testing"
)

const existingWithSidebar = `<ac:layout>
  <ac:layout-section ac:type="two_right_sidebar">
    <ac:layout-cell>
      <p>Old main content</p>
    </ac:layout-cell>
    <ac:layout-cell>
      <p><ac:structured-macro ac:name="toc" ac:schema-version="1" /></p>
    </ac:layout-cell>
  </ac:layout-section>
</ac:layout>`

func TestPreserveSidebarWrapsContentWithoutLayout(t *testing.T) {
	got, wrapped := PreserveSidebar("<p>New main content</p>", existingWithSidebar)
	if !wrapped {
		t.Fatal("expected content without a layout to be wrapped")
	}
	if strings.Count(got, "<ac:layout>") != 1 {
		t.Errorf("expected one layout, got:\n%s", got)
	}
	if !strings.Contains(got, "<p>New main content</p>") || !strings.Contains(got, `ac:name="toc"`) {
		t.Errorf("expected new main content and the existing sidebar, got:\n%s", got)
	}
	if strings.Contains(got, "Old main content") {
		t.Errorf("expected old main content to be replaced, got:\n%s", got)
	}
}

func TestPreserveSidebarLeavesContentWithLayoutUnchanged(t *testing.T) {
	// A page fetched in storage format, edited and published back already has its layout.
	storage := strings.Replace(existingWithSidebar, "Old main content", "Edited main content", 1)
	got, wrapped := PreserveSidebar(storage, existingWithSidebar)
	if wrapped {
		t.Fatal("expected content with its own layout not to be wrapped")
	}
	if got != storage {
		t.Errorf("expected content unchanged, got:\n%s", got)
	}
}

func TestPreserveSidebarLeavesContentUnchangedWithoutExistingLayout(t *testing.T) {
	got, wrapped := PreserveSidebar("<p>New</p>", "<p>Old</p>")
	if wrapped || got != "<p>New</p>" {
		t.Errorf("expected content unchanged, got wrapped=%v:\n%s", wrapped, got)
	}
}

func TestHasLayout(t *testing.T) {
	cases := map[string]bool{
		"<ac:layout><ac:layout-section>":        true,
		"<p>x</p>\n<ac:layout>":                 true,
		`<ac:layout ac:foo="bar">`:              true,
		"<ac:layout-section>":                   false,
		"<p>No layout, just ac:layout text</p>": false,
	}
	for html, want := range cases {
		if got := HasLayout(html); got != want {
			t.Errorf("HasLayout(%q) = %v, want %v", html, got, want)
		}
	}
}
