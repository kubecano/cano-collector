package issue

import (
	"time"

	"github.com/google/uuid"
)

// Severity defines the severity level of an issue.
type Severity string

const (
	SeverityDebug Severity = "DEBUG"
	SeverityInfo  Severity = "INFO"
	SeverityLow   Severity = "LOW"
	SeverityHigh  Severity = "HIGH"
)

// Status defines the status of an issue.
type Status string

const (
	StatusFiring   Status = "FIRING"
	StatusResolved Status = "RESOLVED"
)

// Source defines the origin of an issue.
type Source string

const (
	SourceNone                Source = "NONE"
	SourcePrometheus          Source = "PROMETHEUS"
	SourceKubernetesAPIServer Source = "KUBERNETES_API_SERVER"
	SourceManual              Source = "MANUAL"
	SourceCallback            Source = "CALLBACK"
)

// SubjectType defines the type of the Kubernetes resource the issue is about.
type SubjectType string

const (
	SubjectTypeNone       SubjectType = "none"
	SubjectTypeDeployment SubjectType = "deployment"
	SubjectTypePod        SubjectType = "pod"
	SubjectTypeJob        SubjectType = "job"
	SubjectTypeNode       SubjectType = "node"
	SubjectTypeDaemonSet  SubjectType = "daemonset"
)

// Subject holds information about the resource related to the issue.
type Subject struct {
	Name        string
	SubjectType SubjectType
	Namespace   string
	Node        string
	Container   string
	Labels      map[string]string
	Annotations map[string]string
}

// Link represents a URL link to be included in an issue report.
type Link struct {
	URL  string
	Name string
}

// Issue represents a single event or problem identified in the cluster.
// It is the central data structure for reporting.
type Issue struct {
	ID             uuid.UUID
	Title          string
	Description    string
	AggregationKey string
	Severity       Severity
	Status         Status
	Source         Source
	Subject        Subject
	Enrichments    []Enrichment
	Links          []Link
	Fingerprint    string
	StartsAt       time.Time
	EndsAt         *time.Time
}

// Enrichment adds contextual data to an Issue.
type Enrichment struct {
	Blocks []BaseBlock
	// Annotations can be used by senders to control rendering.
	Annotations map[string]string
}
