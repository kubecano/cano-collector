package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnrichmentType_String(t *testing.T) {
	tests := []struct {
		enrichmentType EnrichmentType
		expected       string
	}{
		{EnrichmentTypeAlertLabels, "alert_labels"},
		{EnrichmentTypeAlertAnnotations, "alert_annotations"},
		{EnrichmentTypeGraph, "graph"},
		{EnrichmentTypeAIAnalysis, "ai_analysis"},
		{EnrichmentTypeNodeInfo, "node_info"},
		{EnrichmentTypeContainerInfo, "container_info"},
		{EnrichmentTypeK8sEvents, "k8s_events"},
		{EnrichmentTypeDiff, "diff"},
		{EnrichmentTypeTextFile, "text_file"},
		{EnrichmentTypeCrashInfo, "crash_info"},
		{EnrichmentTypeImagePullBackoffInfo, "image_pull_backoff_info"},
		{EnrichmentTypePendingPodInfo, "pending_pod_info"},
		{EnrichmentType(999), "unknown"}, // default case
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.enrichmentType.String())
	}
}

func TestEnrichmentType_Constants(t *testing.T) {
	// Test that all constants are defined correctly
	assert.Equal(t, EnrichmentTypeAlertLabels, EnrichmentType(0))
	assert.Equal(t, EnrichmentTypeAlertAnnotations, EnrichmentType(1))
	assert.Equal(t, EnrichmentTypeGraph, EnrichmentType(2))
	assert.Equal(t, EnrichmentTypeAIAnalysis, EnrichmentType(3))
	assert.Equal(t, EnrichmentTypeNodeInfo, EnrichmentType(4))
	assert.Equal(t, EnrichmentTypeContainerInfo, EnrichmentType(5))
	assert.Equal(t, EnrichmentTypeK8sEvents, EnrichmentType(6))
	assert.Equal(t, EnrichmentTypeDiff, EnrichmentType(7))
	assert.Equal(t, EnrichmentTypeTextFile, EnrichmentType(8))
	assert.Equal(t, EnrichmentTypeCrashInfo, EnrichmentType(9))
	assert.Equal(t, EnrichmentTypeImagePullBackoffInfo, EnrichmentType(10))
	assert.Equal(t, EnrichmentTypePendingPodInfo, EnrichmentType(11))
}
