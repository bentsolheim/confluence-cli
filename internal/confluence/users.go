package confluence

import (
	"net/url"
)

// GetUserByKey looks up a user by their Confluence user key
// (the value of ri:userkey in storage format).
func (c *Client) GetUserByKey(key string) (*User, error) {
	return c.getUser("key", key)
}

// GetUserByUsername looks up a user by username (login name).
func (c *Client) GetUserByUsername(username string) (*User, error) {
	return c.getUser("username", username)
}

func (c *Client) getUser(param, value string) (*User, error) {
	params := url.Values{}
	params.Set(param, value)
	var user User
	if err := c.do("GET", "/rest/api/user?"+params.Encode(), &user); err != nil {
		return nil, err
	}
	return &user, nil
}
