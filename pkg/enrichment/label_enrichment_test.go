package enrichment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/kubecano/cano-collector/pkg/logger"
)

func TestNewLabelEnrichment(t *testing.T) {
	log := logger.NewLogger("info", "test")

	t.Run("with custom config", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			EnableLabels:      true,
			EnableAnnotations: false,
			ExcludeLabels:     []string{"test"},
			DisplayFormat:     "json",
		}

		enricher := NewLabelEnrichment(log, config)

		assert.NotNil(t, enricher)
		assert.Equal(t, config, enricher.config)
		assert.Equal(t, log, enricher.logger)
	})

	t.Run("with nil config uses default", func(t *testing.T) {
		enricher := NewLabelEnrichment(log, nil)

		assert.NotNil(t, enricher)
		assert.NotNil(t, enricher.config)
		assert.True(t, enricher.config.EnableLabels)
		assert.True(t, enricher.config.EnableAnnotations)
		assert.Equal(t, "table", enricher.config.DisplayFormat)
	})
}

func TestDefaultLabelEnrichmentConfig(t *testing.T) {
	config := DefaultLabelEnrichmentConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableLabels)
	assert.True(t, config.EnableAnnotations)
	assert.Equal(t, "table", config.DisplayFormat)
	assert.Contains(t, config.ExcludeLabels, "__name__")
	assert.Contains(t, config.ExcludeLabels, "job")
	assert.Contains(t, config.ExcludeLabels, "instance")
}

func TestLabelEnrichment_EnrichIssue(t *testing.T) {
	log := logger.NewLogger("info", "test")

	t.Run("nil issue returns error", func(t *testing.T) {
		enricher := NewLabelEnrichment(log, nil)

		err := enricher.EnrichIssue(nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue is nil")
	})

	t.Run("issue without subject skips enrichment", func(t *testing.T) {
		enricher := NewLabelEnrichment(log, nil)
		iss := issue.NewIssue("Test Issue", "test")
		iss.Subject = nil

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Empty(t, iss.Enrichments)
	})

	t.Run("adds labels enrichment", func(t *testing.T) {
		enricher := NewLabelEnrichment(log, nil)
		iss := createTestIssueWithLabels()

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Len(t, iss.Enrichments, 2) // labels + annotations

		// Check labels enrichment
		labelsEnrichment := iss.Enrichments[0]
		assert.Equal(t, "Alert Labels", *labelsEnrichment.Title)
		assert.Equal(t, issue.EnrichmentTypeAlertLabels, *labelsEnrichment.EnrichmentType)
		assert.Len(t, labelsEnrichment.Blocks, 1)

		// Check annotations enrichment
		annotationsEnrichment := iss.Enrichments[1]
		assert.Equal(t, "Alert Annotations", *annotationsEnrichment.Title)
		assert.Equal(t, issue.EnrichmentTypeAlertAnnotations, *annotationsEnrichment.EnrichmentType)
		assert.Len(t, annotationsEnrichment.Blocks, 1)
	})

	t.Run("skips when labels disabled", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			EnableLabels:      false,
			EnableAnnotations: true,
		}
		enricher := NewLabelEnrichment(log, config)
		iss := createTestIssueWithLabels()

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Len(t, iss.Enrichments, 1) // only annotations
	})

	t.Run("skips when annotations disabled", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			EnableLabels:      true,
			EnableAnnotations: false,
		}
		enricher := NewLabelEnrichment(log, config)
		iss := createTestIssueWithLabels()

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Len(t, iss.Enrichments, 1) // only labels
	})
}

func TestLabelEnrichment_filterLabels(t *testing.T) {
	log := logger.NewLogger("info", "test")

	t.Run("excludes labels", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			ExcludeLabels: []string{"job", "instance"},
		}
		enricher := NewLabelEnrichment(log, config)

		labels := map[string]string{
			"alertname": "TestAlert",
			"severity":  "warning",
			"job":       "prometheus",
			"instance":  "localhost:9090",
		}

		filtered := enricher.filterLabels(labels)

		assert.Len(t, filtered, 2)
		assert.Contains(t, filtered, "alertname")
		assert.Contains(t, filtered, "severity")
		assert.NotContains(t, filtered, "job")
		assert.NotContains(t, filtered, "instance")
	})

	t.Run("includes only specified labels", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			IncludeLabels: []string{"alertname", "severity"},
		}
		enricher := NewLabelEnrichment(log, config)

		labels := map[string]string{
			"alertname": "TestAlert",
			"severity":  "warning",
			"job":       "prometheus",
			"namespace": "default",
		}

		filtered := enricher.filterLabels(labels)

		assert.Len(t, filtered, 2)
		assert.Contains(t, filtered, "alertname")
		assert.Contains(t, filtered, "severity")
		assert.NotContains(t, filtered, "job")
		assert.NotContains(t, filtered, "namespace")
	})

	t.Run("empty include list includes all except excluded", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			IncludeLabels: []string{},
			ExcludeLabels: []string{"job"},
		}
		enricher := NewLabelEnrichment(log, config)

		labels := map[string]string{
			"alertname": "TestAlert",
			"severity":  "warning",
			"job":       "prometheus",
		}

		filtered := enricher.filterLabels(labels)

		assert.Len(t, filtered, 2)
		assert.Contains(t, filtered, "alertname")
		assert.Contains(t, filtered, "severity")
		assert.NotContains(t, filtered, "job")
	})
}

func TestLabelEnrichment_filterAnnotations(t *testing.T) {
	log := logger.NewLogger("info", "test")

	t.Run("excludes annotations", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			ExcludeAnnotations: []string{"kubectl.kubernetes.io", "prometheus.io"},
		}
		enricher := NewLabelEnrichment(log, config)

		annotations := map[string]string{
			"summary":     "Test alert summary",
			"description": "Test alert description",
			"kubectl.kubernetes.io/last-applied-configuration": "{}",
			"prometheus.io/scrape":                             "true",
		}

		filtered := enricher.filterAnnotations(annotations)

		assert.Len(t, filtered, 2)
		assert.Contains(t, filtered, "summary")
		assert.Contains(t, filtered, "description")
		assert.NotContains(t, filtered, "kubectl.kubernetes.io/last-applied-configuration")
		assert.NotContains(t, filtered, "prometheus.io/scrape")
	})

	t.Run("includes only specified annotations", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			IncludeAnnotations: []string{"summary", "description"},
		}
		enricher := NewLabelEnrichment(log, config)

		annotations := map[string]string{
			"summary":     "Test alert summary",
			"description": "Test alert description",
			"runbook_url": "https://example.com/runbook",
			"dashboard":   "https://example.com/dashboard",
		}

		filtered := enricher.filterAnnotations(annotations)

		assert.Len(t, filtered, 2)
		assert.Contains(t, filtered, "summary")
		assert.Contains(t, filtered, "description")
		assert.NotContains(t, filtered, "runbook_url")
		assert.NotContains(t, filtered, "dashboard")
	})

	t.Run("empty include list includes all except excluded", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			IncludeAnnotations: []string{},
			ExcludeAnnotations: []string{"kubectl.kubernetes.io"},
		}
		enricher := NewLabelEnrichment(log, config)

		annotations := map[string]string{
			"summary":     "Test alert summary",
			"description": "Test alert description",
			"kubectl.kubernetes.io/last-applied-configuration": "{}",
		}

		filtered := enricher.filterAnnotations(annotations)

		assert.Len(t, filtered, 2)
		assert.Contains(t, filtered, "summary")
		assert.Contains(t, filtered, "description")
		assert.NotContains(t, filtered, "kubectl.kubernetes.io/last-applied-configuration")
	})

	t.Run("handles empty annotations map", func(t *testing.T) {
		config := &LabelEnrichmentConfig{}
		enricher := NewLabelEnrichment(log, config)

		filtered := enricher.filterAnnotations(map[string]string{})

		assert.Empty(t, filtered)
	})
}

func TestLabelEnrichment_createTableBlocks(t *testing.T) {
	log := logger.NewLogger("info", "test")
	enricher := NewLabelEnrichment(log, nil)

	t.Run("creates labels table block", func(t *testing.T) {
		labels := map[string]string{
			"alertname": "TestAlert",
			"severity":  "warning",
			"namespace": "default",
		}

		block := enricher.createLabelsTableBlock(labels)

		assert.NotNil(t, block)
		assert.Equal(t, "table", block.BlockType())
		assert.Equal(t, []string{"Label", "Value"}, block.Headers)
		assert.Len(t, block.Rows, 3)

		// Check that rows are sorted by key
		assert.Equal(t, []string{"alertname", "TestAlert"}, block.Rows[0])
		assert.Equal(t, []string{"namespace", "default"}, block.Rows[1])
		assert.Equal(t, []string{"severity", "warning"}, block.Rows[2])
	})

	t.Run("creates annotations table block", func(t *testing.T) {
		annotations := map[string]string{
			"summary":     "Test summary",
			"description": "Test description",
		}

		block := enricher.createAnnotationsTableBlock(annotations)

		assert.NotNil(t, block)
		assert.Equal(t, "table", block.BlockType())
		assert.Equal(t, []string{"Annotation", "Value"}, block.Headers)
		assert.Len(t, block.Rows, 2)

		// Check that rows are sorted by key
		assert.Equal(t, []string{"description", "Test description"}, block.Rows[0])
		assert.Equal(t, []string{"summary", "Test summary"}, block.Rows[1])
	})
}

func TestLabelEnrichment_createJsonBlocks(t *testing.T) {
	log := logger.NewLogger("info", "test")
	enricher := NewLabelEnrichment(log, nil)

	t.Run("creates labels json block", func(t *testing.T) {
		labels := map[string]string{
			"alertname": "TestAlert",
			"severity":  "warning",
		}

		block := enricher.createLabelsJsonBlock(labels)

		assert.NotNil(t, block)
		assert.Equal(t, "json", block.BlockType())
		assert.Equal(t, labels, block.Data)
	})

	t.Run("creates annotations json block", func(t *testing.T) {
		annotations := map[string]string{
			"summary":     "Test summary",
			"description": "Test description",
		}

		block := enricher.createAnnotationsJsonBlock(annotations)

		assert.NotNil(t, block)
		assert.Equal(t, "json", block.BlockType())
		assert.Equal(t, annotations, block.Data)
	})
}

func TestLabelEnrichment_DisplayFormats(t *testing.T) {
	log := logger.NewLogger("info", "test")

	t.Run("uses table format by default", func(t *testing.T) {
		enricher := NewLabelEnrichment(log, nil)
		iss := createTestIssueWithLabels()

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Len(t, iss.Enrichments, 2)

		// Check that blocks are table blocks
		labelsBlock := iss.Enrichments[0].Blocks[0]
		assert.Equal(t, "table", labelsBlock.BlockType())

		annotationsBlock := iss.Enrichments[1].Blocks[0]
		assert.Equal(t, "table", annotationsBlock.BlockType())
	})

	t.Run("uses json format when configured", func(t *testing.T) {
		config := &LabelEnrichmentConfig{
			EnableLabels:      true,
			EnableAnnotations: true,
			DisplayFormat:     "json",
		}
		enricher := NewLabelEnrichment(log, config)
		iss := createTestIssueWithLabels()

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Len(t, iss.Enrichments, 2)

		// Check that blocks are json blocks
		labelsBlock := iss.Enrichments[0].Blocks[0]
		assert.Equal(t, "json", labelsBlock.BlockType())

		annotationsBlock := iss.Enrichments[1].Blocks[0]
		assert.Equal(t, "json", annotationsBlock.BlockType())
	})
}

func TestLabelEnrichment_isExcluded(t *testing.T) {
	log := logger.NewLogger("info", "test")
	config := &LabelEnrichmentConfig{
		ExcludeLabels: []string{"job", "__meta_"},
	}
	enricher := NewLabelEnrichment(log, config)

	tests := []struct {
		key      string
		expected bool
	}{
		{"job", true},
		{"alertname", false},
		{"__meta_kubernetes_pod_uid", true},
		{"__meta_container_id", true},
		{"severity", false},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			result := enricher.isExcluded(test.key)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestLabelEnrichment_isIncluded(t *testing.T) {
	log := logger.NewLogger("info", "test")
	config := &LabelEnrichmentConfig{
		IncludeLabels: []string{"alertname", "severity"},
	}
	enricher := NewLabelEnrichment(log, config)

	tests := []struct {
		key      string
		expected bool
	}{
		{"alertname", true},
		{"severity", true},
		{"job", false},
		{"namespace", false},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			result := enricher.isIncluded(test.key)
			assert.Equal(t, test.expected, result)
		})
	}
}

// Helper function to create a test issue with labels and annotations
func createTestIssueWithLabels() *issue.Issue {
	iss := issue.NewIssue("Test Issue", "test")

	subject := issue.NewSubject("test-pod", issue.SubjectTypePod)
	subject.Labels = map[string]string{
		"alertname": "TestAlert",
		"severity":  "warning",
		"namespace": "default",
		"pod":       "test-pod",
	}
	subject.Annotations = map[string]string{
		"summary":     "Test alert summary",
		"description": "Test alert description",
	}

	iss.SetSubject(subject)
	return iss
}

func TestLabelEnrichment_EmptyMaps(t *testing.T) {
	log := logger.NewLogger("info", "test")
	enricher := NewLabelEnrichment(log, nil)

	t.Run("handles empty labels map", func(t *testing.T) {
		iss := issue.NewIssue("Test Issue", "test")
		subject := issue.NewSubject("test-pod", issue.SubjectTypePod)
		subject.Labels = map[string]string{}
		subject.Annotations = map[string]string{
			"summary": "Test summary",
		}
		iss.SetSubject(subject)

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Len(t, iss.Enrichments, 1) // only annotations
		assert.Equal(t, "Alert Annotations", *iss.Enrichments[0].Title)
	})

	t.Run("handles empty annotations map", func(t *testing.T) {
		iss := issue.NewIssue("Test Issue", "test")
		subject := issue.NewSubject("test-pod", issue.SubjectTypePod)
		subject.Labels = map[string]string{
			"alertname": "TestAlert",
		}
		subject.Annotations = map[string]string{}
		iss.SetSubject(subject)

		err := enricher.EnrichIssue(iss)

		require.NoError(t, err)
		assert.Len(t, iss.Enrichments, 1) // only labels
		assert.Equal(t, "Alert Labels", *iss.Enrichments[0].Title)
	})
}
