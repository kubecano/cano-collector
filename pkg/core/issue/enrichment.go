package issue

// LinkType represents the type of link
type LinkType int

const (
	LinkTypeGeneral LinkType = iota
	LinkTypePrometheusGenerator
	LinkTypeInvestigate
	LinkTypeSilence
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

// Enrichment provides additional context to an Issue
type Enrichment struct {
	Blocks         []BaseBlock       `json:"blocks"`
	Annotations    map[string]string `json:"annotations,omitempty"`
	EnrichmentType *EnrichmentType   `json:"enrichment_type,omitempty"`
	Title          *string           `json:"title,omitempty"`
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
		Blocks:         make([]BaseBlock, 0),
		Annotations:    make(map[string]string),
		EnrichmentType: &enrichmentType,
		Title:          &title,
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
