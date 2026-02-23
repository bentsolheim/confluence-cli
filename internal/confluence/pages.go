package confluence

import (
	"fmt"
	"net/url"
)

// GetPage fetches a single page by ID with expanded content, ancestors, children, and history.
func (c *Client) GetPage(id string) (*Page, error) {
	var page Page
	params := url.Values{}
	params.Set("expand", "body.storage,version,space,ancestors,children.page,history")
	path := fmt.Sprintf("/rest/api/content/%s?%s", url.PathEscape(id), params.Encode())
	if err := c.do("GET", path, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
