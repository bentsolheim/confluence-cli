package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
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

// doWithBody executes an authenticated HTTP request with a request body and decodes the JSON response.
func (c *Client) doWithBody(method, path string, body io.Reader, contentType string, result interface{}) error {
	reqURL := c.baseURL + path
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "<<< HTTP %d\n%s\n", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// doJSON executes an authenticated HTTP request with a JSON-encoded payload.
func (c *Client) doJSON(method, path string, payload interface{}, result interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}
	return c.doWithBody(method, path, bytes.NewReader(jsonData), "application/json", result)
}

// CurrentUser returns the currently authenticated user. Useful for testing auth.
func (c *Client) CurrentUser() (*User, error) {
	var user User
	if err := c.do("GET", "/rest/api/user/current", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CreatePage creates a new page in Confluence.
func (c *Client) CreatePage(spaceKey, title, htmlContent, parentID string) (*Page, error) {
	payload := map[string]interface{}{
		"type":  "page",
		"title": title,
		"space": map[string]interface{}{
			"key": spaceKey,
		},
		"body": map[string]interface{}{
			"storage": map[string]interface{}{
				"value":          htmlContent,
				"representation": "storage",
			},
		},
	}
	if parentID != "" {
		payload["ancestors"] = []map[string]interface{}{
			{"id": parentID},
		}
	}

	var page Page
	if err := c.doJSON("POST", "/rest/api/content", payload, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// UpdatePage updates an existing page in Confluence.
// It preserves any existing two-column sidebar layout.
func (c *Client) UpdatePage(pageID, title, htmlContent string, currentVersion int) (*Page, error) {
	// Fetch existing page to check for sidebar layout
	existingPage, err := c.GetPage(pageID)
	if err != nil {
		return nil, fmt.Errorf("fetching existing page for layout check: %w", err)
	}

	finalContent := htmlContent
	if existingPage.Body != nil && existingPage.Body.Storage != nil {
		_, existingSidebar, hasLayout := ExtractLayoutContent(existingPage.Body.Storage.Value)
		if hasLayout {
			finalContent = WrapInTwoColumnLayout(htmlContent, existingSidebar)
			if c.verbose {
				fmt.Fprintf(os.Stderr, "[layout] Preserved existing sidebar content\n")
			}
		}
	}

	payload := map[string]interface{}{
		"version": map[string]interface{}{
			"number": currentVersion + 1,
		},
		"title": title,
		"type":  "page",
		"body": map[string]interface{}{
			"storage": map[string]interface{}{
				"value":          finalContent,
				"representation": "storage",
			},
		},
	}

	var page Page
	path := fmt.Sprintf("/rest/api/content/%s", pageID)
	if err := c.doJSON("PUT", path, payload, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// GetAttachmentByFilename retrieves an attachment matching a filename (if it exists).
func (c *Client) GetAttachmentByFilename(pageID, filename string) (*Attachment, error) {
	path := fmt.Sprintf("/rest/api/content/%s/child/attachment?filename=%s",
		pageID, neturl.QueryEscape(filename))

	var ar AttachmentListResponse
	if err := c.do("GET", path, &ar); err != nil {
		return nil, err
	}
	if len(ar.Results) == 0 {
		return nil, nil
	}
	return &ar.Results[0], nil
}

// UploadAttachment uploads a file as an attachment to a page.
// Returns the server-provided filename (title). If overwrite is true and an
// attachment with the same filename already exists, it creates a new version.
func (c *Client) UploadAttachment(pageID, filePath string, overwrite bool) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi != nil && fi.Size() == 0 {
		return "", fmt.Errorf("file is empty: %s", filePath)
	}

	filename := filepath.Base(filePath)

	// Check for existing attachment if overwrite requested
	var existing *Attachment
	if overwrite {
		if ex, err := c.GetAttachmentByFilename(pageID, filename); err == nil && ex != nil {
			existing = ex
		}
	}

	// Build multipart body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("creating multipart part: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", fmt.Errorf("copying file content: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("finalizing multipart body: %w", err)
	}

	var uploadPath string
	updatingExisting := existing != nil && overwrite
	if updatingExisting {
		uploadPath = fmt.Sprintf("/rest/api/content/%s/child/attachment/%s/data", pageID, existing.ID)
	} else {
		uploadPath = fmt.Sprintf("/rest/api/content/%s/child/attachment", pageID)
	}

	// Upload requires a special header to bypass XSRF check
	reqURL := c.baseURL + uploadPath
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return "", fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "no-check")

	if c.verbose {
		fmt.Fprintf(os.Stderr, ">>> POST %s (multipart upload: %s)\n", reqURL, filename)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	if c.verbose {
		fmt.Fprintf(os.Stderr, "<<< HTTP %d\n%s\n", resp.StatusCode, string(respBytes))
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("attachment upload failed (HTTP %d): %s", resp.StatusCode, string(respBytes))
	}

	if updatingExisting {
		var att Attachment
		if err := json.Unmarshal(respBytes, &att); err != nil {
			return "", fmt.Errorf("decoding attachment update response: %w", err)
		}
		if att.Title == "" {
			att.Title = filename
		}
		return att.Title, nil
	}

	var ar AttachmentListResponse
	if err := json.Unmarshal(respBytes, &ar); err != nil {
		return "", fmt.Errorf("decoding attachment upload response: %w", err)
	}
	if len(ar.Results) == 0 {
		return "", fmt.Errorf("no attachment result returned by server")
	}
	return ar.Results[0].Title, nil
}
