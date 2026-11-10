package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

var sampleLinks = []confluence.Link{
	{Type: confluence.LinkPage, Text: "Existing", Href: "https://wiki.sits.no/display/MUP/Existing", Title: "Existing", SpaceKey: "MUP"},
	{Type: confluence.LinkMissing, Text: "Ghost", Href: "https://wiki.sits.no/pages/createpage.action?spaceKey=MUP&title=Ghost", Title: "Ghost", SpaceKey: "MUP", Missing: true},
	{Type: confluence.LinkExternal, Text: "Ext", Href: "https://example.com"},
}

func TestBuildLinkReport(t *testing.T) {
	page := &confluence.Page{ID: "1", Title: "T"}
	r := buildLinkReport(page, sampleLinks, false)
	if r.Total != 3 || r.Missing != 1 || len(r.Links) != 3 {
		t.Errorf("unexpected report: %+v", r)
	}
	r = buildLinkReport(page, sampleLinks, true)
	if r.Total != 3 || r.Missing != 1 || len(r.Links) != 1 || r.Links[0].Title != "Ghost" {
		t.Errorf("unexpected missing-only report: %+v", r)
	}
}

func TestWriteLinkReport(t *testing.T) {
	r := buildLinkReport(&confluence.Page{ID: "1", Title: "T"}, sampleLinks, false)

	var buf bytes.Buffer
	if err := writeLinkReport(&buf, "markdown", r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "| MISSING | missing | Ghost | MUP:Ghost |") {
		t.Errorf("unexpected markdown output: %q", buf.String())
	}

	buf.Reset()
	if err := writeLinkReport(&buf, "text", r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "3 link(s), 1 missing") || !strings.Contains(buf.String(), "MUP:Ghost") {
		t.Errorf("unexpected text output: %q", buf.String())
	}

	buf.Reset()
	if err := writeLinkReport(&buf, "json", r); err != nil {
		t.Fatal(err)
	}
	var out linkReport
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out.Missing != 1 || len(out.Links) != 3 || !out.Links[1].Missing {
		t.Errorf("unexpected json output: %+v", out)
	}
}
