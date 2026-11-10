package confluence

import (
	"strings"
	"testing"
)

const sampleView = `<p>See <a href="/display/MUP/Existing+Page" data-linked-resource-id="123" data-linked-resource-type="page" data-linked-resource-default-alias="Existing Page">Existing Page</a>
and <a href="/pages/createpage.action?spaceKey=MUP&amp;title=Does+Not+Exist&amp;linkCreation=true&amp;fromPageId=1" class="createlink">Does Not Exist</a>.</p>
<p><a href="/pages/viewpage.action?pageId=456">By id</a>
<a href="/spaces/DEV/pages/789/Some+Title">Spaces URL</a>
<a href="https://example.com/x">External</a>
<a href="#section">Anchor</a>
<a href="/download/attachments/1/file.pdf" data-linked-resource-type="attachment" data-linked-resource-container-id="1">file.pdf</a>
<a class="confluence-userlink user-mention" data-username="k77319" href="/display/~k77319">Solheim, Bent Andre</a>
<a href="https://wiki.sits.no/dashboard.action">Dashboard</a>
<span class="unresolved">Broken Space Link</span></p>`

func TestExtractLinks(t *testing.T) {
	links := ExtractLinks(sampleView, "https://wiki.sits.no")

	want := []struct {
		typ     string
		text    string
		missing bool
	}{
		{LinkPage, "Existing Page", false},
		{LinkMissing, "Does Not Exist", true},
		{LinkPage, "By id", false},
		{LinkPage, "Spaces URL", false},
		{LinkExternal, "External", false},
		{LinkAnchor, "Anchor", false},
		{LinkAttachment, "file.pdf", false},
		{LinkUser, "Solheim, Bent Andre", false},
		{LinkInternal, "Dashboard", false},
		{LinkMissing, "Broken Space Link", true},
	}
	if len(links) != len(want) {
		t.Fatalf("expected %d links, got %d: %+v", len(want), len(links), links)
	}
	for i, w := range want {
		l := links[i]
		if l.Type != w.typ || l.Text != w.text || l.Missing != w.missing {
			t.Errorf("link %d: expected {%s %q missing=%v}, got {%s %q missing=%v}", i, w.typ, w.text, w.missing, l.Type, l.Text, l.Missing)
		}
	}

	if links[0].PageID != "123" || links[0].Title != "Existing Page" || links[0].SpaceKey != "MUP" {
		t.Errorf("unexpected page link details: %+v", links[0])
	}
	if links[0].Href != "https://wiki.sits.no/display/MUP/Existing+Page" {
		t.Errorf("expected absolute href, got %q", links[0].Href)
	}
	if links[1].Title != "Does Not Exist" || links[1].SpaceKey != "MUP" {
		t.Errorf("unexpected missing link details: %+v", links[1])
	}
	if links[2].PageID != "456" {
		t.Errorf("expected pageId 456, got %q", links[2].PageID)
	}
	if links[3].PageID != "789" || links[3].SpaceKey != "DEV" {
		t.Errorf("unexpected spaces link details: %+v", links[3])
	}
	if links[7].Username != "k77319" {
		t.Errorf("expected username k77319, got %q", links[7].Username)
	}
}

func TestExtractLinks_Empty(t *testing.T) {
	if links := ExtractLinks("<p>No links here</p>", "https://wiki.sits.no"); len(links) != 0 {
		t.Errorf("expected no links, got %+v", links)
	}
}

// realView is rendered HTML returned by wiki.sits.no (contentbody/convert/view)
// for a link to a missing page, an existing page, a link into a non-existent
// space and a user mention.
const realView = `<p><a class="createlink" href="/pages/createpage.action?spaceKey=~k77319&amp;title=Finnes+ikke+denne+siden&amp;atl_token=82c64f0d08a46a849cd48991ec66f4354c009f1b">Finnes ikke denne siden</a> og <a href="/spaces/~k77319/pages/1651447216/confluence-cli+testside">confluence-cli testside</a> og <a class="unresolved" href="#">Hei</a> og <a class="confluence-userlink user-mention current-user-mention" data-username="k77319" href="/display/~k77319" data-linked-resource-id="66130591" data-linked-resource-version="3" data-linked-resource-type="userinfo" data-base-url="https://wiki.sits.no">Solheim, Bent Andre</a></p>`

func TestExtractLinks_RealConfluenceView(t *testing.T) {
	links := ExtractLinks(realView, "https://wiki.sits.no")
	if len(links) != 4 {
		t.Fatalf("expected 4 links, got %d: %+v", len(links), links)
	}

	missing := links[0]
	if !missing.Missing || missing.Type != LinkMissing || missing.Title != "Finnes ikke denne siden" || missing.SpaceKey != "~k77319" {
		t.Errorf("unexpected missing link: %+v", missing)
	}
	if strings.Contains(missing.Href, "atl_token") {
		t.Errorf("expected atl_token to be stripped from href, got %q", missing.Href)
	}

	page := links[1]
	if page.Missing || page.Type != LinkPage || page.PageID != "1651447216" || page.Title != "confluence-cli testside" {
		t.Errorf("unexpected page link: %+v", page)
	}

	if !links[2].Missing || links[2].Title != "Hei" {
		t.Errorf("expected unresolved link to be missing: %+v", links[2])
	}

	if links[3].Type != LinkUser || links[3].Username != "k77319" {
		t.Errorf("unexpected user link: %+v", links[3])
	}
}
