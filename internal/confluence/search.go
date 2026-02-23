package confluence

import (
	"net/url"
	"strconv"
)

// Search executes a CQL query and returns matching content.
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
