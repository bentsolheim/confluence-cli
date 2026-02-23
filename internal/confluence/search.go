package confluence

import (
	"net/url"
	"strconv"
)

// Search executes a CQL query against the content search endpoint.
// Use this for standard CQL queries (e.g. "space = DEV AND type = page").
func (c *Client) Search(cql string, maxResults int) (*SearchResult, error) {
	var result SearchResult
	params := url.Values{}
	params.Set("cql", cql)
	params.Set("limit", strconv.Itoa(maxResults))
	params.Set("expand", "content.space,content.version")
	path := "/rest/api/content/search?" + params.Encode()
	if err := c.do("GET", path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SiteSearch executes a CQL query against the general search endpoint.
// This endpoint is required for siteSearch queries and returns results
// with content properly nested.
func (c *Client) SiteSearch(cql string, maxResults int) (*SearchResult, error) {
	var result SearchResult
	params := url.Values{}
	params.Set("cql", cql)
	params.Set("limit", strconv.Itoa(maxResults))
	params.Set("expand", "content.space,content.version")
	path := "/rest/api/search?" + params.Encode()
	if err := c.do("GET", path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
