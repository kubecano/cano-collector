package issue

import (
	"time"

	"github.com/google/uuid"
)

// Issue represents an event that should be sent to destinations
type Issue struct {
	ID             uuid.UUID    `json:"id"`
	Title          string       `json:"title"`
	Description    string       `json:"description,omitempty"`
	AggregationKey string       `json:"aggregation_key"`
	Severity       Severity     `json:"severity"`
	Status         Status       `json:"status"`
	Source         Source       `json:"source"`
	Subject        *Subject     `json:"subject"`
	Enrichments    []Enrichment `json:"enrichments,omitempty"`
	Links          []Link       `json:"links,omitempty"`
	Fingerprint    string       `json:"fingerprint"`
	StartsAt       time.Time    `json:"starts_at"`
	EndsAt         *time.Time   `json:"ends_at,omitempty"`
}

// NewIssue creates a new Issue with default values
func NewIssue(title, aggregationKey string) *Issue {
	return &Issue{
		ID:             uuid.New(),
		Title:          title,
		AggregationKey: aggregationKey,
		Severity:       SeverityInfo,
		Status:         StatusFiring,
		Source:         SourceUnknown,
		Subject:        NewSubject("", SubjectTypeNone),
		Enrichments:    make([]Enrichment, 0),
		Links:          make([]Link, 0),
		StartsAt:       time.Now(),
	}
}

// AddEnrichment adds an enrichment to the issue
func (i *Issue) AddEnrichment(enrichment Enrichment) {
	i.Enrichments = append(i.Enrichments, enrichment)
}

// AddLink adds a link to the issue
func (i *Issue) AddLink(link Link) {
	i.Links = append(i.Links, link)
}

// SetSubject sets the subject of the issue
func (i *Issue) SetSubject(subject *Subject) {
	i.Subject = subject
}

// IsResolved returns true if the issue is resolved
func (i *Issue) IsResolved() bool {
	return i.Status == StatusResolved
}

// GetStatusMessage returns a formatted status message
func (i *Issue) GetStatusMessage() string {
	if i.IsResolved() {
		return "[RESOLVED] " + i.Title
	}
	return i.Title
}
