package confluence

import (
	"fmt"
	"net/url"
)

// GetLabels returns all labels for the given content ID.
func (c *Client) GetLabels(contentID string) ([]Label, error) {
	path := fmt.Sprintf("/rest/api/content/%s/label", url.PathEscape(contentID))
	var result LabelListResponse
	if err := c.do("GET", path, &result); err != nil {
		return nil, fmt.Errorf("getting labels: %w", err)
	}
	return result.Results, nil
}

// AddLabels adds the given labels to the content. Labels use prefix "global".
func (c *Client) AddLabels(contentID string, names []string) ([]Label, error) {
	labels := make([]Label, len(names))
	for i, name := range names {
		labels[i] = Label{Prefix: "global", Name: name}
	}
	path := fmt.Sprintf("/rest/api/content/%s/label", url.PathEscape(contentID))
	var result LabelListResponse
	if err := c.doJSON("POST", path, labels, &result); err != nil {
		return nil, fmt.Errorf("adding labels: %w", err)
	}
	return result.Results, nil
}

// RemoveLabel removes a single label from the content.
func (c *Client) RemoveLabel(contentID, labelName string) error {
	path := fmt.Sprintf("/rest/api/content/%s/label/%s",
		url.PathEscape(contentID), url.PathEscape(labelName))
	if err := c.do("DELETE", path, nil); err != nil {
		return fmt.Errorf("removing label %q: %w", labelName, err)
	}
	return nil
}
