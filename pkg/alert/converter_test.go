package alert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/alert/model"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/kubecano/cano-collector/pkg/enrichment"
	"github.com/kubecano/cano-collector/pkg/logger"
)

func TestConverter_ConvertAlertManagerEventToIssues(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	tests := []struct {
		name          string
		alertEvent    *model.AlertManagerEvent
		expectedCount int
		expectError   bool
	}{
		{
			name: "successful conversion with single alert",
			alertEvent: &model.AlertManagerEvent{
				Alerts: []model.PrometheusAlert{
					{
						Status:       "firing",
						Fingerprint:  "test-fingerprint",
						StartsAt:     time.Now(),
						EndsAt:       time.Time{},
						GeneratorURL: "http://prometheus.example.com/graph",
						Labels: map[string]string{
							"alertname": "HighCPUUsage",
							"severity":  "critical",
							"pod":       "test-pod",
							"namespace": "test-namespace",
						},
						Annotations: map[string]string{
							"summary":     "High CPU usage detected",
							"description": "CPU usage is above 90%",
						},
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "successful conversion with multiple alerts",
			alertEvent: &model.AlertManagerEvent{
				Alerts: []model.PrometheusAlert{
					{
						Status:      "firing",
						Fingerprint: "test-fingerprint-1",
						StartsAt:    time.Now(),
						Labels: map[string]string{
							"alertname": "HighCPUUsage",
							"severity":  "critical",
						},
						Annotations: map[string]string{
							"summary": "High CPU usage detected",
						},
					},
					{
						Status:      "resolved",
						Fingerprint: "test-fingerprint-2",
						StartsAt:    time.Now().Add(-time.Hour),
						EndsAt:      time.Now(),
						Labels: map[string]string{
							"alertname": "HighMemoryUsage",
							"severity":  "warning",
						},
						Annotations: map[string]string{
							"summary": "High memory usage resolved",
						},
					},
				},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "nil alert event",
			alertEvent:    nil,
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "empty alerts",
			alertEvent: &model.AlertManagerEvent{
				Alerts: []model.PrometheusAlert{},
			},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "alert without alertname",
			alertEvent: &model.AlertManagerEvent{
				Alerts: []model.PrometheusAlert{
					{
						Status:      "firing",
						Fingerprint: "test-fingerprint",
						StartsAt:    time.Now(),
						Labels: map[string]string{
							"severity": "critical",
						},
						Annotations: map[string]string{
							"summary": "Alert without name",
						},
					},
				},
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues, err := converter.ConvertAlertManagerEventToIssues(tt.alertEvent)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, issues)
			} else {
				require.NoError(t, err)
				assert.Len(t, issues, tt.expectedCount)

				if len(issues) > 0 {
					// Verify first issue
					firstIssue := issues[0]
					assert.NotEmpty(t, firstIssue.ID)
					assert.NotEmpty(t, firstIssue.Title)
					assert.NotEmpty(t, firstIssue.AggregationKey)
					assert.Equal(t, issue.SourcePrometheus, firstIssue.Source)
					assert.NotEmpty(t, firstIssue.Fingerprint)
				}
			}
		})
	}
}

func TestConverter_convertPrometheusAlertToIssue(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	tests := []struct {
		name             string
		alert            model.PrometheusAlert
		expectedTitle    string
		expectedSeverity issue.Severity
		expectedStatus   issue.Status
		expectError      bool
	}{
		{
			name: "alert with summary annotation",
			alert: model.PrometheusAlert{
				Status:      "firing",
				Fingerprint: "test-fingerprint",
				StartsAt:    time.Now(),
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary":     "High CPU usage detected",
					"description": "CPU usage is above 90%",
				},
			},
			expectedTitle:    "High CPU usage detected",
			expectedSeverity: issue.SeverityHigh,
			expectedStatus:   issue.StatusFiring,
			expectError:      false,
		},
		{
			name: "resolved alert",
			alert: model.PrometheusAlert{
				Status:      "resolved",
				Fingerprint: "test-fingerprint",
				StartsAt:    time.Now().Add(-time.Hour),
				EndsAt:      time.Now(),
				Labels: map[string]string{
					"alertname": "HighMemoryUsage",
					"severity":  "warning",
				},
				Annotations: map[string]string{
					"summary": "Memory usage normal",
				},
			},
			expectedTitle:    "Memory usage normal",
			expectedSeverity: issue.SeverityLow,
			expectedStatus:   issue.StatusResolved,
			expectError:      false,
		},
		{
			name: "alert without summary - uses alertname",
			alert: model.PrometheusAlert{
				Status:      "firing",
				Fingerprint: "test-fingerprint",
				StartsAt:    time.Now(),
				Labels: map[string]string{
					"alertname": "DiskSpaceLow",
					"severity":  "info",
				},
				Annotations: map[string]string{
					"description": "Disk space is running low",
				},
			},
			expectedTitle:    "DiskSpaceLow",
			expectedSeverity: issue.SeverityInfo,
			expectedStatus:   issue.StatusFiring,
			expectError:      false,
		},
		{
			name: "alert without alertname",
			alert: model.PrometheusAlert{
				Status:      "firing",
				Fingerprint: "test-fingerprint",
				StartsAt:    time.Now(),
				Labels: map[string]string{
					"severity": "critical",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iss, err := converter.convertPrometheusAlertToIssue(tt.alert)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, iss)
			} else {
				require.NoError(t, err)
				require.NotNil(t, iss)

				assert.Equal(t, tt.expectedTitle, iss.Title)
				assert.Equal(t, tt.expectedSeverity, iss.Severity)
				assert.Equal(t, tt.expectedStatus, iss.Status)
				assert.Equal(t, issue.SourcePrometheus, iss.Source)
				assert.Equal(t, tt.alert.Fingerprint, iss.Fingerprint)
				assert.Equal(t, tt.alert.Labels["alertname"], iss.AggregationKey)
				assert.False(t, iss.StartsAt.IsZero())

				if tt.expectedStatus == issue.StatusResolved {
					assert.NotNil(t, iss.EndsAt)
				}
			}
		})
	}
}

func TestConverter_createSubject(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	tests := []struct {
		name                string
		alert               model.PrometheusAlert
		expectedSubjectType issue.SubjectType
		expectedSubjectName string
		expectedNamespace   string
	}{
		{
			name: "pod subject",
			alert: model.PrometheusAlert{
				Labels: map[string]string{
					"pod":       "test-pod",
					"namespace": "test-namespace",
					"node":      "test-node",
				},
			},
			expectedSubjectType: issue.SubjectTypePod,
			expectedSubjectName: "test-pod",
			expectedNamespace:   "test-namespace",
		},
		{
			name: "deployment subject",
			alert: model.PrometheusAlert{
				Labels: map[string]string{
					"deployment": "test-deployment",
					"namespace":  "test-namespace",
				},
			},
			expectedSubjectType: issue.SubjectTypeDeployment,
			expectedSubjectName: "test-deployment",
			expectedNamespace:   "test-namespace",
		},
		{
			name: "node subject",
			alert: model.PrometheusAlert{
				Labels: map[string]string{
					"node": "test-node",
				},
			},
			expectedSubjectType: issue.SubjectTypeNode,
			expectedSubjectName: "test-node",
			expectedNamespace:   "",
		},
		{
			name: "instance as node subject",
			alert: model.PrometheusAlert{
				Labels: map[string]string{
					"instance": "test-instance",
				},
			},
			expectedSubjectType: issue.SubjectTypeNode,
			expectedSubjectName: "test-instance",
			expectedNamespace:   "",
		},
		{
			name: "unknown subject",
			alert: model.PrometheusAlert{
				Labels: map[string]string{
					"custom_label": "custom_value",
				},
			},
			expectedSubjectType: issue.SubjectTypeNone,
			expectedSubjectName: "Unknown",
			expectedNamespace:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := converter.createSubject(tt.alert)

			assert.Equal(t, tt.expectedSubjectType, subject.SubjectType)
			assert.Equal(t, tt.expectedSubjectName, subject.Name)
			assert.Equal(t, tt.expectedNamespace, subject.Namespace)

			// Verify labels are copied
			for k, v := range tt.alert.Labels {
				assert.Equal(t, v, subject.Labels[k])
			}
		})
	}
}

func TestConverter_SeverityMapping(t *testing.T) {
	tests := []struct {
		prometheusLabel  string
		expectedSeverity issue.Severity
	}{
		{"critical", issue.SeverityHigh},
		{"high", issue.SeverityHigh},
		{"error", issue.SeverityHigh},
		{"warning", issue.SeverityLow},
		{"low", issue.SeverityLow},
		{"info", issue.SeverityInfo},
		{"debug", issue.SeverityDebug},
		{"unknown", issue.SeverityInfo}, // default
		{"", issue.SeverityInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.prometheusLabel, func(t *testing.T) {
			severity := issue.SeverityFromPrometheusLabel(tt.prometheusLabel)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

func TestConverter_StatusMapping(t *testing.T) {
	tests := []struct {
		prometheusStatus string
		expectedStatus   issue.Status
	}{
		{"firing", issue.StatusFiring},
		{"resolved", issue.StatusResolved},
		{"unknown", issue.StatusFiring}, // default
		{"", issue.StatusFiring},        // default
	}

	for _, tt := range tests {
		t.Run(tt.prometheusStatus, func(t *testing.T) {
			status := issue.StatusFromPrometheusStatus(tt.prometheusStatus)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestConverter_WithGeneratorURL(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	alert := model.PrometheusAlert{
		Status:       "firing",
		Fingerprint:  "test-fingerprint",
		StartsAt:     time.Now(),
		GeneratorURL: "http://prometheus.example.com/graph?g0.expr=up",
		Labels: map[string]string{
			"alertname": "TargetDown",
		},
		Annotations: map[string]string{
			"summary": "Target is down",
		},
	}

	iss, err := converter.convertPrometheusAlertToIssue(alert)
	require.NoError(t, err)
	require.NotNil(t, iss)

	assert.Len(t, iss.Links, 1)
	assert.Equal(t, "Generator URL", iss.Links[0].Text)
	assert.Equal(t, alert.GeneratorURL, iss.Links[0].URL)
	assert.Equal(t, issue.LinkTypePrometheusGenerator, iss.Links[0].Type)
}

func TestConverter_WithRunbookURL(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	alert := model.PrometheusAlert{
		Status:      "firing",
		Fingerprint: "test-fingerprint",
		StartsAt:    time.Now(),
		Labels: map[string]string{
			"alertname": "KubePodCrashLooping",
		},
		Annotations: map[string]string{
			"summary":     "Pod is crash looping",
			"runbook_url": "https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping",
		},
	}

	iss, err := converter.convertPrometheusAlertToIssue(alert)
	require.NoError(t, err)
	require.NotNil(t, iss)

	assert.Len(t, iss.Links, 1)
	assert.Equal(t, "Runbook", iss.Links[0].Text)
	assert.Equal(t, alert.Annotations["runbook_url"], iss.Links[0].URL)
	assert.Equal(t, issue.LinkTypeRunbook, iss.Links[0].Type)
}

func TestConverter_WithBothGeneratorAndRunbookURLs(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	alert := model.PrometheusAlert{
		Status:       "firing",
		Fingerprint:  "test-fingerprint",
		StartsAt:     time.Now(),
		GeneratorURL: "http://prometheus.example.com/graph?g0.expr=up",
		Labels: map[string]string{
			"alertname": "KubePodCrashLooping",
		},
		Annotations: map[string]string{
			"summary":     "Pod is crash looping",
			"runbook_url": "https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping",
		},
	}

	iss, err := converter.convertPrometheusAlertToIssue(alert)
	require.NoError(t, err)
	require.NotNil(t, iss)

	// Should have both links
	assert.Len(t, iss.Links, 2)

	// Generator URL should be first
	assert.Equal(t, "Generator URL", iss.Links[0].Text)
	assert.Equal(t, alert.GeneratorURL, iss.Links[0].URL)
	assert.Equal(t, issue.LinkTypePrometheusGenerator, iss.Links[0].Type)

	// Runbook URL should be second
	assert.Equal(t, "Runbook", iss.Links[1].Text)
	assert.Equal(t, alert.Annotations["runbook_url"], iss.Links[1].URL)
	assert.Equal(t, issue.LinkTypeRunbook, iss.Links[1].Type)
}

func TestConverter_FingerprintHandling(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	t.Run("uses alert fingerprint when provided", func(t *testing.T) {
		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "custom-alert-fingerprint-123",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"pod":       "test-pod",
				"namespace": "default",
			},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)
		require.NoError(t, err)
		require.NotNil(t, iss)

		// Should use the fingerprint from the alert
		assert.Equal(t, "custom-alert-fingerprint-123", iss.Fingerprint)
	})

	t.Run("generates fingerprint when alert fingerprint is empty", func(t *testing.T) {
		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "", // Empty fingerprint
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"pod":       "test-pod",
				"namespace": "default",
			},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)
		require.NoError(t, err)
		require.NotNil(t, iss)

		// Should generate a fingerprint automatically
		assert.NotEmpty(t, iss.Fingerprint)
		assert.Len(t, iss.Fingerprint, 64) // SHA256 hex string
		assert.Regexp(t, "^[a-f0-9]+$", iss.Fingerprint)
	})

	t.Run("different alerts generate different fingerprints", func(t *testing.T) {
		alert1 := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "", // Let it generate
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert1",
				"pod":       "test-pod-1",
				"namespace": "default",
			},
		}

		alert2 := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "", // Let it generate
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert2",
				"pod":       "test-pod-2",
				"namespace": "default",
			},
		}

		iss1, err := converter.convertPrometheusAlertToIssue(alert1)
		require.NoError(t, err)

		iss2, err := converter.convertPrometheusAlertToIssue(alert2)
		require.NoError(t, err)

		// Should generate different fingerprints
		assert.NotEqual(t, iss1.Fingerprint, iss2.Fingerprint)
	})

	t.Run("same alert parameters generate same fingerprint", func(t *testing.T) {
		alert1 := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "", // Let it generate
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"pod":       "test-pod",
				"namespace": "default",
			},
		}

		alert2 := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "", // Let it generate
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"pod":       "test-pod",
				"namespace": "default",
			},
		}

		iss1, err := converter.convertPrometheusAlertToIssue(alert1)
		require.NoError(t, err)

		iss2, err := converter.convertPrometheusAlertToIssue(alert2)
		require.NoError(t, err)

		// Should generate same fingerprints for identical alerts
		assert.Equal(t, iss1.Fingerprint, iss2.Fingerprint)
	})
}

func TestConverter_LabelEnrichmentIntegration(t *testing.T) {
	converter := NewConverter(logger.NewLogger("info", "test"))

	t.Run("adds label and annotation enrichments", func(t *testing.T) {
		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "test-fingerprint",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "HighCPUUsage",
				"severity":  "critical",
				"pod":       "test-pod",
				"namespace": "test-namespace",
				"job":       "prometheus", // This should be excluded by default
			},
			Annotations: map[string]string{
				"summary":     "High CPU usage detected",
				"description": "CPU usage is above 90%",
			},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)

		require.NoError(t, err)
		require.NotNil(t, iss)

		// Verify that enrichments were added
		assert.Len(t, iss.Enrichments, 2) // labels + annotations

		// Verify labels enrichment
		labelsEnrichment := iss.Enrichments[0]
		assert.Equal(t, "Alert Labels", *labelsEnrichment.Title)
		assert.Equal(t, issue.EnrichmentTypeAlertLabels, *labelsEnrichment.EnrichmentType)
		assert.Len(t, labelsEnrichment.Blocks, 1)

		// Verify that the labels block is a table
		tableBlock, ok := labelsEnrichment.Blocks[0].(*issue.TableBlock)
		require.True(t, ok, "Expected table block for labels")
		assert.Equal(t, []string{"Label", "Value"}, tableBlock.Headers)

		// Verify that "job" label was excluded by default
		jobFound := false
		for _, row := range tableBlock.Rows {
			if len(row) >= 2 && row[0] == "job" {
				jobFound = true
				break
			}
		}
		assert.False(t, jobFound, "Expected 'job' label to be excluded by default")

		// Verify annotations enrichment
		annotationsEnrichment := iss.Enrichments[1]
		assert.Equal(t, "Alert Annotations", *annotationsEnrichment.Title)
		assert.Equal(t, issue.EnrichmentTypeAlertAnnotations, *annotationsEnrichment.EnrichmentType)
		assert.Len(t, annotationsEnrichment.Blocks, 1)

		// Verify that the annotations block is a table
		annotationsTableBlock, ok := annotationsEnrichment.Blocks[0].(*issue.TableBlock)
		require.True(t, ok, "Expected table block for annotations")
		assert.Equal(t, []string{"Annotation", "Value"}, annotationsTableBlock.Headers)
		assert.Len(t, annotationsTableBlock.Rows, 2) // summary + description
	})

	t.Run("works with custom enrichment config", func(t *testing.T) {
		// Create converter with custom config that includes "job" label
		enrichmentConfig := &enrichment.LabelEnrichmentConfig{
			EnableLabels:      true,
			EnableAnnotations: false,                // Disable annotations
			ExcludeLabels:     []string{"__name__"}, // Only exclude __name__
			DisplayFormat:     "json",
		}
		customConverter := NewConverterWithEnrichmentConfig(logger.NewLogger("info", "test"), enrichmentConfig)

		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "test-fingerprint",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "HighCPUUsage",
				"severity":  "critical",
				"job":       "prometheus", // This should now be included
			},
			Annotations: map[string]string{
				"summary": "High CPU usage detected",
			},
		}

		iss, err := customConverter.convertPrometheusAlertToIssue(alert)

		require.NoError(t, err)
		require.NotNil(t, iss)

		// Should only have labels enrichment (annotations disabled)
		assert.Len(t, iss.Enrichments, 1)

		// Verify that the block is a JSON block
		jsonBlock, ok := iss.Enrichments[0].Blocks[0].(*issue.JsonBlock)
		require.True(t, ok, "Expected JSON block for labels")

		// Verify that "job" label is included
		labelsData, ok := jsonBlock.Data.(map[string]string)
		require.True(t, ok, "Expected map[string]string data")
		assert.Contains(t, labelsData, "job")
		assert.Equal(t, "prometheus", labelsData["job"])
	})

	t.Run("handles enrichment errors gracefully", func(t *testing.T) {
		// Create alert without labels/annotations to test edge case
		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "test-fingerprint",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
			},
			Annotations: map[string]string{},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)

		// Should not fail even if enrichment has issues
		require.NoError(t, err)
		require.NotNil(t, iss)

		// Should have one enrichment (labels only, since annotations is empty)
		assert.Len(t, iss.Enrichments, 1)
	})
}

func TestNewConverterWithConfig(t *testing.T) {
	log := logger.NewLogger("info", "test")

	t.Run("creates converter with labels disabled", func(t *testing.T) {
		enrichmentConfig := config.EnrichmentConfig{
			Labels: config.LabelEnrichmentConfig{
				Enabled:       false,
				DisplayFormat: "table",
			},
			Annotations: config.AnnotationEnrichmentConfig{
				Enabled:       true,
				DisplayFormat: "json",
			},
		}

		converter := NewConverterWithConfig(log, enrichmentConfig)

		assert.NotNil(t, converter)
		assert.NotNil(t, converter.labelEnrichment)

		// Test that labels enrichment is disabled
		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "test-fingerprint",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"severity":  "critical",
				"pod":       "test-pod",
				"namespace": "test-namespace",
			},
			Annotations: map[string]string{
				"summary":     "Test summary",
				"description": "Test description",
			},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)
		require.NoError(t, err)

		// Should only have annotations enrichment (labels disabled)
		assert.Len(t, iss.Enrichments, 1)
		assert.Equal(t, issue.EnrichmentTypeAlertAnnotations, *iss.Enrichments[0].EnrichmentType)
	})

	t.Run("creates converter with annotations disabled", func(t *testing.T) {
		enrichmentConfig := config.EnrichmentConfig{
			Labels: config.LabelEnrichmentConfig{
				Enabled:       true,
				DisplayFormat: "json",
			},
			Annotations: config.AnnotationEnrichmentConfig{
				Enabled:       false,
				DisplayFormat: "table",
			},
		}

		converter := NewConverterWithConfig(log, enrichmentConfig)

		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "test-fingerprint",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"severity":  "critical",
			},
			Annotations: map[string]string{
				"summary": "Test summary",
			},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)
		require.NoError(t, err)

		// Should only have labels enrichment (annotations disabled)
		assert.Len(t, iss.Enrichments, 1)
		assert.Equal(t, issue.EnrichmentTypeAlertLabels, *iss.Enrichments[0].EnrichmentType)
	})

	t.Run("creates converter with both disabled", func(t *testing.T) {
		enrichmentConfig := config.EnrichmentConfig{
			Labels: config.LabelEnrichmentConfig{
				Enabled: false,
			},
			Annotations: config.AnnotationEnrichmentConfig{
				Enabled: false,
			},
		}

		converter := NewConverterWithConfig(log, enrichmentConfig)

		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "test-fingerprint",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"severity":  "critical",
			},
			Annotations: map[string]string{
				"summary": "Test summary",
			},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)
		require.NoError(t, err)

		// Should have no enrichments
		assert.Empty(t, iss.Enrichments)
	})

	t.Run("creates converter with custom include/exclude filters", func(t *testing.T) {
		enrichmentConfig := config.EnrichmentConfig{
			Labels: config.LabelEnrichmentConfig{
				Enabled:       true,
				DisplayFormat: "table",
				IncludeLabels: []string{"alertname", "severity"},
				ExcludeLabels: []string{"instance", "job"},
			},
			Annotations: config.AnnotationEnrichmentConfig{
				Enabled:            true,
				DisplayFormat:      "json",
				IncludeAnnotations: []string{"summary"},
				ExcludeAnnotations: []string{"description"},
			},
		}

		converter := NewConverterWithConfig(log, enrichmentConfig)

		alert := model.PrometheusAlert{
			Status:      "firing",
			Fingerprint: "test-fingerprint",
			StartsAt:    time.Now(),
			Labels: map[string]string{
				"alertname": "TestAlert",
				"severity":  "critical",
				"instance":  "localhost:9090", // should be excluded
				"job":       "prometheus",     // should be excluded
				"pod":       "test-pod",       // should be excluded (not in include list)
			},
			Annotations: map[string]string{
				"summary":     "Test summary",     // should be included
				"description": "Test description", // should be excluded
				"runbook":     "Test runbook",     // should be excluded (not in include list)
			},
		}

		iss, err := converter.convertPrometheusAlertToIssue(alert)
		require.NoError(t, err)

		// Should have both enrichments
		assert.Len(t, iss.Enrichments, 2)

		// Find labels enrichment
		var labelsEnrichment *issue.Enrichment
		var annotationsEnrichment *issue.Enrichment
		for i := range iss.Enrichments {
			switch *iss.Enrichments[i].EnrichmentType {
			case issue.EnrichmentTypeAlertLabels:
				labelsEnrichment = &iss.Enrichments[i]
			case issue.EnrichmentTypeAlertAnnotations:
				annotationsEnrichment = &iss.Enrichments[i]
			}
		}

		require.NotNil(t, labelsEnrichment)
		require.NotNil(t, annotationsEnrichment)

		// Check labels enrichment has table block with filtered labels
		tableBlock, ok := labelsEnrichment.Blocks[0].(*issue.TableBlock)
		require.True(t, ok)
		assert.Len(t, tableBlock.Rows, 2) // only alertname and severity

		// Check annotations enrichment has JSON block with filtered annotations
		jsonBlock, ok := annotationsEnrichment.Blocks[0].(*issue.JsonBlock)
		require.True(t, ok)
		assert.Len(t, jsonBlock.Data, 1) // only summary
		assert.Contains(t, jsonBlock.Data, "summary")
	})
}
