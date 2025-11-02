package slack

// MessageContext holds all data needed for rendering Slack message templates
type MessageContext struct {
	// Status and state
	Status      string // "firing" or "resolved"
	StatusEmoji string // "🔥" or "✅"
	StatusText  string // "Alert firing" or "Alert resolved"

	// Severity
	Severity      string // "High", "Medium", "Low"
	SeverityEmoji string // "🔴", "🟡", "🟢"

	// Alert basic info
	Title       string
	Description string

	// Alert type
	AlertType      string // "Prometheus Alert" or "K8s Event"
	AlertTypeEmoji string // "📊" or "👀"

	// Resource metadata
	Cluster   string
	Namespace string
	PodName   string
	Source    string

	// Links to external systems
	Links []Link

	// Crash information (for pod alerts)
	CrashInfo *CrashInfo

	// Enrichments (logs, tables, files)
	Enrichments []EnrichmentData
}

// CrashInfo contains information about container crashes
type CrashInfo struct {
	Container string
	Restarts  string
	Status    string
	Reason    string
}

// Link represents a clickable link in the message
type Link struct {
	Text string
	URL  string
}

// EnrichmentData represents an enrichment to be rendered
type EnrichmentData struct {
	Type     string // "file", "table", "markdown", "logs"
	Title    string
	Content  string
	FileLink string // Permalink for file uploads
}
