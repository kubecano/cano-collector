package alert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/pkg/alert/model"
	"github.com/kubecano/cano-collector/pkg/core/issue"
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
