package confluence

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"time"
)

// Client is an authenticated Confluence REST API client.
type Client struct {
	baseURL    string
	token      string
	verbose    bool
	httpClient *http.Client
}

// NewClient creates a new Confluence API client.
func NewClient(baseURL, token string, verbose bool) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		verbose: verbose,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do executes an authenticated HTTP request and decodes the JSON response.
func (c *Client) do(method, path string, result interface{}) error {
	reqURL := c.baseURL + path
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	if c.verbose {
		fmt.Fprintf(os.Stderr, ">>> %s %s\n", method, reqURL)
		if parsed, err := neturl.Parse(reqURL); err == nil {
			if addrs, err := net.LookupHost(parsed.Hostname()); err == nil {
				fmt.Fprintf(os.Stderr, "    DNS: %s -> %v\n", parsed.Hostname(), addrs)
			} else {
				fmt.Fprintf(os.Stderr, "    DNS lookup failed: %v\n", err)
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "<<< HTTP %d\n%s\n", resp.StatusCode, string(body))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// CurrentUser returns the currently authenticated user. Useful for testing auth.
func (c *Client) CurrentUser() (*User, error) {
	var user User
	if err := c.do("GET", "/rest/api/user/current", &user); err != nil {
		return nil, err
	}
	return &user, nil
}
