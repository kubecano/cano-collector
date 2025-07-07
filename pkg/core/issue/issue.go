package issue

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	issue := &Issue{
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

	// Generate fingerprint based on issue attributes
	issue.Fingerprint = issue.generateFingerprint()

	return issue
}

// generateFingerprint generates a unique fingerprint for the issue
// Logic similar to Robusta's implementation for deduplication
func (i *Issue) generateFingerprint() string {
	subjectName := ""
	subjectNamespace := ""
	subjectNode := ""
	subjectType := ""

	if i.Subject != nil {
		subjectName = i.Subject.Name
		subjectNamespace = i.Subject.Namespace
		subjectNode = i.Subject.Node
		subjectType = i.Subject.SubjectType.String()
	}

	// Create fingerprint string combining key attributes
	fingerprintStr := fmt.Sprintf("%s,%s,%s,%s,%s,%s",
		subjectType,
		subjectName,
		subjectNamespace,
		subjectNode,
		i.Source.String(),
		i.AggregationKey,
	)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(fingerprintStr))
	return hex.EncodeToString(hash[:])
}

// SetFingerprint sets a custom fingerprint (e.g., from Prometheus/Alertmanager)
func (i *Issue) SetFingerprint(fingerprint string) {
	if fingerprint != "" {
		i.Fingerprint = fingerprint
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
	// Always regenerate fingerprint when subject changes
	i.Fingerprint = i.generateFingerprint()
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
