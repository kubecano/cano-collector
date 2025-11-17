package issue

// LinkType represents the type of link
type LinkType int

const (
	LinkTypeGeneral LinkType = iota
	LinkTypePrometheusGenerator
	LinkTypeInvestigate
	LinkTypeSilence
	LinkTypeRunbook
)

// String returns the string representation of the link type
func (lt LinkType) String() string {
	switch lt {
	case LinkTypeGeneral:
		return "GENERAL"
	case LinkTypePrometheusGenerator:
		return "PROMETHEUS_GENERATOR"
	case LinkTypeInvestigate:
		return "INVESTIGATE"
	case LinkTypeSilence:
		return "SILENCE"
	case LinkTypeRunbook:
		return "RUNBOOK"
	default:
		return "UNKNOWN"
	}
}

// Link represents a URL relevant to the issue
type Link struct {
	Text string   `json:"text"`
	URL  string   `json:"url"`
	Type LinkType `json:"type"`
}

// NewLink creates a new Link
func NewLink(text, url string, linkType LinkType) *Link {
	return &Link{
		Text: text,
		URL:  url,
		Type: linkType,
	}
}

// BaseBlock represents a base block for enrichments (placeholder for future implementation)
type BaseBlock interface {
	BlockType() string
}

// FileInfo contains metadata about uploaded files
type FileInfo struct {
	ID        string `json:"id,omitempty"`        // Slack file ID (for attaching to messages)
	Permalink string `json:"permalink"`           // Slack file permalink
	Filename  string `json:"filename"`            // Original filename
	Size      int64  `json:"size,omitempty"`      // File size in bytes
	MimeType  string `json:"mime_type,omitempty"` // MIME type
}

// Enrichment provides additional context to an Issue
type Enrichment struct {
	Type        EnrichmentType    `json:"type"`                  // Enrichment type (logs, table, etc.)
	Title       string            `json:"title,omitempty"`       // Enrichment title
	Content     string            `json:"content,omitempty"`     // Text content for inline rendering
	Blocks      []BaseBlock       `json:"blocks,omitempty"`      // Structured blocks (tables, etc.)
	Annotations map[string]string `json:"annotations,omitempty"` // Additional metadata
	FileInfo    *FileInfo         `json:"file_info,omitempty"`   // File upload metadata (for Slack/Teams)
}

// NewEnrichment creates a new Enrichment
func NewEnrichment() *Enrichment {
	return &Enrichment{
		Blocks:      make([]BaseBlock, 0),
		Annotations: make(map[string]string),
	}
}

// NewEnrichmentWithType creates a new Enrichment with type and title
func NewEnrichmentWithType(enrichmentType EnrichmentType, title string) *Enrichment {
	return &Enrichment{
		Type:        enrichmentType,
		Title:       title,
		Blocks:      make([]BaseBlock, 0),
		Annotations: make(map[string]string),
	}
}

// AddBlock adds a block to the enrichment
func (e *Enrichment) AddBlock(block BaseBlock) {
	e.Blocks = append(e.Blocks, block)
}

// AddAnnotation adds an annotation to the enrichment
func (e *Enrichment) AddAnnotation(key, value string) {
	if e.Annotations == nil {
		e.Annotations = make(map[string]string)
	}
	e.Annotations[key] = value
}
