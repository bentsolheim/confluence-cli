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
	Metadata  *PageMetadata  `json:"metadata"`
}

// PageMetadata contains page metadata including labels.
type PageMetadata struct {
	Labels *LabelListResponse `json:"labels"`
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
	Content               *Page            `json:"content"`
	Title                 string           `json:"title"`
	Excerpt               string           `json:"excerpt"`
	URL                   string           `json:"url"`
	ResultGlobalContainer *ResultContainer `json:"resultGlobalContainer"`
	LastModified          string           `json:"lastModified"`
	FriendlyLastModified  string           `json:"friendlyLastModified"`
	// Top-level fields present in siteSearch results (where content is nil)
	ID    string    `json:"id"`
	Space *Space    `json:"space"`
	Links PageLinks `json:"_links"`
}

// ResultContainer holds space info for a search result.
type ResultContainer struct {
	Title      string `json:"title"`
	DisplayURL string `json:"displayUrl"`
}

// Attachment represents a Confluence attachment.
type Attachment struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Links struct {
		Download string `json:"download"`
	} `json:"_links"`
}

// AttachmentListResponse models the response from listing/uploading attachments.
type AttachmentListResponse struct {
	Results []Attachment `json:"results"`
}

// ImageRef represents a discovered image tag in HTML (used in the publish flow).
type ImageRef struct {
	OriginalTag string
	Src         string
	Alt         string
	TitleAttr   string
	Width       string
	Height      string
	IsExternal  bool
	IsDataURI   bool
}

// PublishConfig represents the Confluence configuration from Markdown frontmatter.
type PublishConfig struct {
	URL                  string   `yaml:"url"`
	SpaceKey             string   `yaml:"spaceKey,omitempty"`
	PageID               string   `yaml:"pageId,omitempty"`
	Title                string   `yaml:"title"`
	Format               string   `yaml:"format,omitempty"`
	Labels               []string `yaml:"labels,omitempty"`
	AssetsBase           string   `yaml:"assetsBase,omitempty"`
	OverwriteAttachments bool     `yaml:"overwriteAttachments,omitempty"`
	FailOnMissingImages  bool     `yaml:"failOnMissingImages,omitempty"`
}

// Label represents a Confluence content label.
type Label struct {
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
	ID     string `json:"id,omitempty"`
}

// LabelListResponse is the response from the label list endpoint.
type LabelListResponse struct {
	Results []Label `json:"results"`
	Start   int     `json:"start"`
	Limit   int     `json:"limit"`
	Size    int     `json:"size"`
}
