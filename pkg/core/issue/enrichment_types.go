package issue

// EnrichmentType represents the type of enrichment to categorize different kinds of contextual data
type EnrichmentType int

const (
	EnrichmentTypeAlertLabels EnrichmentType = iota
	EnrichmentTypeAlertAnnotations
	EnrichmentTypeGraph
	EnrichmentTypeAIAnalysis
	EnrichmentTypeNodeInfo
	EnrichmentTypeContainerInfo
	EnrichmentTypeK8sEvents
	EnrichmentTypeDiff
	EnrichmentTypeTextFile
	EnrichmentTypeCrashInfo
	EnrichmentTypeImagePullBackoffInfo
	EnrichmentTypePendingPodInfo
)

// String returns the string representation of the enrichment type
func (et EnrichmentType) String() string {
	switch et {
	case EnrichmentTypeAlertLabels:
		return "alert_labels"
	case EnrichmentTypeAlertAnnotations:
		return "alert_annotations"
	case EnrichmentTypeGraph:
		return "graph"
	case EnrichmentTypeAIAnalysis:
		return "ai_analysis"
	case EnrichmentTypeNodeInfo:
		return "node_info"
	case EnrichmentTypeContainerInfo:
		return "container_info"
	case EnrichmentTypeK8sEvents:
		return "k8s_events"
	case EnrichmentTypeDiff:
		return "diff"
	case EnrichmentTypeTextFile:
		return "text_file"
	case EnrichmentTypeCrashInfo:
		return "crash_info"
	case EnrichmentTypeImagePullBackoffInfo:
		return "image_pull_backoff_info"
	case EnrichmentTypePendingPodInfo:
		return "pending_pod_info"
	default:
		return "unknown"
	}
}
