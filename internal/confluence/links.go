package confluence

import (
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Link types reported by ExtractLinks.
const (
	LinkPage       = "page"
	LinkMissing    = "missing"
	LinkAttachment = "attachment"
	LinkUser       = "user"
	LinkAnchor     = "anchor"
	LinkInternal   = "internal"
	LinkExternal   = "external"
)

// Link is a hyperlink found in the rendered (body.view) HTML of a page.
type Link struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Href     string `json:"href,omitempty"`
	Title    string `json:"title,omitempty"`
	SpaceKey string `json:"spaceKey,omitempty"`
	PageID   string `json:"pageId,omitempty"`
	Username string `json:"username,omitempty"`
	Missing  bool   `json:"missing"`
}

var (
	displayPathRe = regexp.MustCompile(`^/display/([^/]+)/(.+)$`)
	spacesPathRe  = regexp.MustCompile(`^/spaces/([^/]+)/pages/(\d+)(?:/([^/]+))?`)
)

// ExtractLinks parses rendered page HTML (body.view) and returns every link,
// classified by type. Links to pages that do not exist are rendered by
// Confluence with class "createlink" (or as "unresolved" elements) and are
// reported with Type "missing" and Missing=true.
func ExtractLinks(viewHTML, baseURL string) []Link {
	doc, err := html.Parse(strings.NewReader(viewHTML))
	if err != nil {
		return nil
	}
	base, _ := url.Parse(baseURL)

	var links []Link
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "a" && hasAttr(n, "href") {
				links = append(links, classifyLink(n, base))
				return
			}
			if hasClass(n, "unresolved") {
				links = append(links, Link{
					Type:    LinkMissing,
					Text:    nodeText(n),
					Title:   nodeText(n),
					Missing: true,
				})
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return links
}

func classifyLink(n *html.Node, base *url.URL) Link {
	rawHref := attr(n, "href")
	link := Link{
		Text: nodeText(n),
		Href: absoluteHref(rawHref, base),
	}

	if hasClass(n, "createlink") || hasClass(n, "unresolved") {
		link.Type = LinkMissing
		link.Missing = true
		if u, err := url.Parse(rawHref); err == nil {
			link.Title = u.Query().Get("title")
			link.SpaceKey = u.Query().Get("spaceKey")
		}
		if link.Title == "" {
			link.Title = link.Text
		}
		return link
	}

	if hasClass(n, "confluence-userlink") || hasClass(n, "user-mention") {
		link.Type = LinkUser
		link.Username = attr(n, "data-username")
		return link
	}

	if strings.HasPrefix(rawHref, "#") {
		link.Type = LinkAnchor
		return link
	}

	switch attr(n, "data-linked-resource-type") {
	case "page", "blogpost":
		link.Type = LinkPage
		link.PageID = attr(n, "data-linked-resource-id")
		link.Title = attr(n, "data-linked-resource-default-alias")
	case "attachment":
		link.Type = LinkAttachment
		link.PageID = attr(n, "data-linked-resource-container-id")
		return link
	case "userinfo":
		link.Type = LinkUser
		link.Username = attr(n, "data-username")
		return link
	}

	u, err := url.Parse(rawHref)
	if err != nil {
		if link.Type == "" {
			link.Type = LinkExternal
		}
		return link
	}

	if u.IsAbs() && (base == nil || !strings.EqualFold(u.Host, base.Host)) {
		if link.Type == "" {
			link.Type = LinkExternal
		}
		return link
	}

	path := u.Path
	if base != nil && base.Path != "" && base.Path != "/" {
		path = strings.TrimPrefix(path, strings.TrimSuffix(base.Path, "/"))
	}

	switch {
	case strings.HasPrefix(path, "/download/"):
		if link.Type == "" {
			link.Type = LinkAttachment
		}
	case path == "/pages/viewpage.action":
		link.Type = LinkPage
		if link.PageID == "" {
			link.PageID = u.Query().Get("pageId")
		}
	case spacesPathRe.MatchString(path):
		m := spacesPathRe.FindStringSubmatch(path)
		link.Type = LinkPage
		link.SpaceKey = m[1]
		if link.PageID == "" {
			link.PageID = m[2]
		}
		if link.Title == "" && m[3] != "" {
			link.Title = unescapeTitle(m[3])
		}
	case displayPathRe.MatchString(path):
		m := displayPathRe.FindStringSubmatch(path)
		link.Type = LinkPage
		link.SpaceKey = m[1]
		if link.Title == "" {
			link.Title = unescapeTitle(m[2])
		}
	case link.Type == "":
		link.Type = LinkInternal
	}
	return link
}

// unescapeTitle turns a URL path segment like "Page+Title" into "Page Title".
func unescapeTitle(segment string) string {
	title, err := url.PathUnescape(segment)
	if err != nil {
		title = segment
	}
	return strings.ReplaceAll(title, "+", " ")
}

// absoluteHref resolves href against base and strips the session-bound
// atl_token (XSRF token) that Confluence adds to "create page" links.
func absoluteHref(href string, base *url.URL) string {
	u, err := url.Parse(href)
	if err != nil || strings.HasPrefix(href, "#") {
		return href
	}
	if q := u.Query(); q.Has("atl_token") {
		q.Del("atl_token")
		u.RawQuery = q.Encode()
	}
	if base != nil && !u.IsAbs() {
		u = base.ResolveReference(u)
	}
	return u.String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}
	return false
}

func hasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
