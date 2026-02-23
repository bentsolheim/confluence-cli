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

// GetPageByTitle fetches a page by space key and title.
func (c *Client) GetPageByTitle(spaceKey, title string) (*Page, error) {
	params := url.Values{}
	params.Set("spaceKey", spaceKey)
	params.Set("title", title)
	params.Set("expand", "body.storage,version,space,ancestors,children.page,history")
	path := "/rest/api/content?" + params.Encode()

	var result PageResults
	if err := c.do("GET", path, &result); err != nil {
		return nil, err
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no page found with title %q in space %q", title, spaceKey)
	}
	return &result.Results[0], nil
}
