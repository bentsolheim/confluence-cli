package confluence

// User represents a Confluence user.
type User struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// Page represents a Confluence page.
type Page struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"`
	Status  string    `json:"status"`
	Title   string    `json:"title"`
	Space   *Space    `json:"space"`
	Version *Version  `json:"version"`
	Body    *Body     `json:"body"`
	Links   PageLinks `json:"_links"`

	Ancestors []Page         `json:"ancestors"`
	Children  *PageChildren  `json:"children"`
	History   *PageHistory   `json:"history"`
}

// Space represents a Confluence space.
type Space struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Version represents a page version.
type Version struct {
	Number  int    `json:"number"`
	When    string `json:"when"`
	By      *User  `json:"by"`
	Message string `json:"message"`
}

// Body contains the page body in different representations.
type Body struct {
	Storage *BodyContent `json:"storage"`
	View    *BodyContent `json:"view"`
}

// BodyContent holds the actual content value and representation.
type BodyContent struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// PageLinks contains links related to a page.
type PageLinks struct {
	WebUI  string `json:"webui"`
	Self   string `json:"self"`
	TinyUI string `json:"tinyui"`
	Base   string `json:"base"`
}

// PageChildren contains child pages.
type PageChildren struct {
	Page *PageResults `json:"page"`
}

// PageResults is a paginated list of pages.
type PageResults struct {
	Results []Page `json:"results"`
	Start   int    `json:"start"`
	Limit   int    `json:"limit"`
	Size    int    `json:"size"`
}

// PageHistory contains creation info for a page.
type PageHistory struct {
	CreatedBy   *User  `json:"createdBy"`
	CreatedDate string `json:"createdDate"`
}

// SearchResult represents the result of a CQL search.
type SearchResult struct {
	Results    []SearchResultItem `json:"results"`
	Start      int                `json:"start"`
	Limit      int                `json:"limit"`
	Size       int                `json:"size"`
	TotalSize  int                `json:"totalSize"`
}

// SearchResultItem represents a single search hit.
type SearchResultItem struct {
	Content   *Page  `json:"content"`
	Title     string `json:"title"`
	Excerpt   string `json:"excerpt"`
	URL       string `json:"url"`
	ResultGlobalContainer *ResultContainer `json:"resultGlobalContainer"`
}

// ResultContainer holds space info for a search result.
type ResultContainer struct {
	Title      string `json:"title"`
	DisplayURL string `json:"displayUrl"`
}
