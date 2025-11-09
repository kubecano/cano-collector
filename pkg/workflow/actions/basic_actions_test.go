package actions

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/core/event"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// createTestAlertManagerEvent creates a proper AlertManagerEvent for testing
func createTestAlertManagerEvent(status, alertname, severity, namespace string, extraLabels map[string]string) event.WorkflowEvent {
	labels := map[string]string{
		"alertname": alertname,
	}
	if severity != "" {
		labels["severity"] = severity
	}
	if namespace != "" {
		labels["namespace"] = namespace
	}
	// Add any extra labels
	for k, v := range extraLabels {
		labels[k] = v
	}

	alertManagerEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      event.EventTypeAlertManager,
		},
		Alerts: []event.PrometheusAlert{
			{
				Status: status,
				Labels: labels,
				Annotations: map[string]string{
					"summary": "Test alert summary",
				},
			},
		},
		Receiver: "test-receiver",
		Status:   status,
	}

	return event.NewAlertManagerWorkflowEvent(alertManagerEvent)
}

// createTestAlertManagerEventNoAlerts creates an AlertManagerEvent with no alerts for testing
func createTestAlertManagerEventNoAlerts() event.WorkflowEvent {
	alertManagerEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      event.EventTypeAlertManager,
		},
		Alerts:   []event.PrometheusAlert{},
		Receiver: "test-receiver",
		Status:   "firing",
	}

	return event.NewAlertManagerWorkflowEvent(alertManagerEvent)
}

func TestLabelFilterAction_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-label-filter",
		Type:    "label_filter",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_labels": map[string]interface{}{
				"severity": "critical",
				"team":     "backend",
			},
			"required_labels": []interface{}{"alertname"},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewLabelFilterAction(config, mockLogger, mockMetrics)

	// Create test event with matching labels
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "default", map[string]string{
		"team": "backend",
	})
	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, true, result.Data.(map[string]interface{})["filter_passed"])
	assert.Contains(t, result.Data.(map[string]interface{})["reason"], "all label filters passed")
}

func TestLabelFilterAction_Execute_FilteredOut_MissingIncludeLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-label-filter",
		Type:    "label_filter",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_labels": map[string]interface{}{
				"severity": "critical",
				"team":     "backend",
			},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewLabelFilterAction(config, mockLogger, mockMetrics)

	// Create test event missing required include label (no team label)
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "default", nil)
	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, false, result.Data.(map[string]interface{})["filter_passed"])
	assert.Contains(t, result.Data.(map[string]interface{})["reason"], "missing required include label: team")
}

func TestLabelFilterAction_Execute_FilteredOut_ExcludeLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-label-filter",
		Type:    "label_filter",
		Enabled: true,
		Parameters: map[string]interface{}{
			"exclude_labels": map[string]interface{}{
				"environment": "test",
				"team":        "qa",
			},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewLabelFilterAction(config, mockLogger, mockMetrics)

	// Create test event with excluded label
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "default", map[string]string{
		"environment": "test",
	})
	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, false, result.Data.(map[string]interface{})["filter_passed"])
	assert.Contains(t, result.Data.(map[string]interface{})["reason"], "alert has excluded label: environment=test")
}

func TestLabelFilterAction_Execute_FilteredOut_MissingRequiredLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-label-filter",
		Type:    "label_filter",
		Enabled: true,
		Parameters: map[string]interface{}{
			"required_labels": []interface{}{"alertname", "severity", "namespace"},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewLabelFilterAction(config, mockLogger, mockMetrics)

	// Create test event missing required label (missing namespace)
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "", nil)
	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, false, result.Data.(map[string]interface{})["filter_passed"])
	assert.Contains(t, result.Data.(map[string]interface{})["reason"], "missing required label:")
}

func TestLabelFilterAction_Execute_NoAlertsInEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-label-filter",
		Type:    "label_filter",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_labels": map[string]interface{}{
				"severity": "critical",
			},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewLabelFilterAction(config, mockLogger, mockMetrics)

	// Create test event with no alerts
	alertEvent := createTestAlertManagerEventNoAlerts()
	result, err := action.Execute(context.Background(), alertEvent)

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "no alerts found in event")
}

func TestSeverityRouterAction_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-severity-router",
		Type:    "severity_router",
		Enabled: true,
		Parameters: map[string]interface{}{
			"severity_mapping": map[string]interface{}{
				"critical": "team-backend-critical",
				"info":     "team-backend-info",
			},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewSeverityRouterAction(config, mockLogger, mockMetrics)

	// Create test event with critical severity
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "default", nil)
	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "HIGH", result.Data.(map[string]interface{})["severity"])
	assert.Equal(t, "team-backend-critical", result.Data.(map[string]interface{})["destination"])
	assert.Equal(t, true, result.Data.(map[string]interface{})["routed"])
}

func TestSeverityRouterAction_Execute_NoMappingFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-severity-router",
		Type:    "severity_router",
		Enabled: true,
		Parameters: map[string]interface{}{
			"severity_mapping": map[string]interface{}{
				"critical": "team-backend-critical",
			},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewSeverityRouterAction(config, mockLogger, mockMetrics)

	// Create test event with info severity (not in mapping)
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "info", "default", nil)
	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "INFO", result.Data.(map[string]interface{})["severity"])
	assert.Empty(t, result.Data.(map[string]interface{})["destination"])
	assert.Equal(t, false, result.Data.(map[string]interface{})["routed"])
}

func TestSeverityRouterAction_Execute_DefaultMapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-severity-router",
		Type:    "severity_router",
		Enabled: true,
		Parameters: map[string]interface{}{
			"severity_mapping": map[string]interface{}{
				"critical": "team-backend-critical",
				"default":  "team-backend-default",
			},
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewSeverityRouterAction(config, mockLogger, mockMetrics)

	// Create test event with info severity (should use default)
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "info", "default", nil)
	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "INFO", result.Data.(map[string]interface{})["severity"])
	assert.Equal(t, "team-backend-default", result.Data.(map[string]interface{})["destination"])
	assert.Equal(t, true, result.Data.(map[string]interface{})["routed"])
}

func TestLabelFilterAction_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-label-filter",
		Type:    "label_filter",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_labels": map[string]interface{}{
				"severity": "critical",
			},
		},
	}

	action := NewLabelFilterAction(config, mockLogger, mockMetrics)
	err := action.Validate()
	assert.NoError(t, err)
}

func TestLabelFilterAction_Validate_NoFiltersConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:       "test-label-filter",
		Type:       "label_filter",
		Enabled:    true,
		Parameters: map[string]interface{}{},
	}

	action := NewLabelFilterAction(config, mockLogger, mockMetrics)
	err := action.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have at least one filter configured")
}

func TestSeverityRouterAction_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-severity-router",
		Type:    "severity_router",
		Enabled: true,
		Parameters: map[string]interface{}{
			"severity_mapping": map[string]interface{}{
				"critical": "team-backend",
				"info":     "team-frontend",
			},
		},
	}

	action := NewSeverityRouterAction(config, mockLogger, mockMetrics)
	err := action.Validate()
	assert.NoError(t, err)
}

func TestSeverityRouterAction_Validate_NoMappingConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:       "test-severity-router",
		Type:       "severity_router",
		Enabled:    true,
		Parameters: map[string]interface{}{},
	}

	action := NewSeverityRouterAction(config, mockLogger, mockMetrics)
	err := action.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have severity_mapping parameter configured")
}

func TestSeverityRouterAction_Validate_InvalidSeverity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-severity-router",
		Type:    "severity_router",
		Enabled: true,
		Parameters: map[string]interface{}{
			"severity_mapping": map[string]interface{}{
				"invalid": "team-backend",
			},
		},
	}

	action := NewSeverityRouterAction(config, mockLogger, mockMetrics)
	err := action.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid severity level in mapping: invalid")
}

// Factory tests

func TestLabelFilterActionFactory_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	factory := NewLabelFilterActionFactory(mockLogger, mockMetrics)

	config := actions_interfaces.ActionConfig{
		Name:    "test-label-filter",
		Type:    "label_filter",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_labels": map[string]interface{}{
				"severity": "critical",
			},
		},
	}

	action, err := factory.Create(config)
	require.NoError(t, err)
	assert.NotNil(t, action)
	assert.Equal(t, "test-label-filter", action.GetName())
}

func TestLabelFilterActionFactory_GetActionType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	factory := NewLabelFilterActionFactory(mockLogger, mockMetrics)
	assert.Equal(t, "label_filter", factory.GetActionType())
}

func TestSeverityRouterActionFactory_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	factory := NewSeverityRouterActionFactory(mockLogger, mockMetrics)

	config := actions_interfaces.ActionConfig{
		Name:    "test-severity-router",
		Type:    "severity_router",
		Enabled: true,
		Parameters: map[string]interface{}{
			"severity_mapping": map[string]interface{}{
				"critical": "team-backend",
			},
		},
	}

	action, err := factory.Create(config)
	require.NoError(t, err)
	assert.NotNil(t, action)
	assert.Equal(t, "test-severity-router", action.GetName())
}

func TestSeverityRouterActionFactory_GetActionType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	factory := NewSeverityRouterActionFactory(mockLogger, mockMetrics)
	assert.Equal(t, "severity_router", factory.GetActionType())
}

// ============================================================================
// IssueEnrichmentAction Tests
// ============================================================================

func TestIssueEnrichmentAction_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-issue-enrichment",
		Type:    "issue_enrichment",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_metadata": true,
			"include_labels":   true,
			"custom_title":     "Test Custom Title",
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewIssueEnrichmentAction(config, mockLogger, mockMetrics)

	// Create test event
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "default", map[string]string{
		"pod":     "test-pod",
		"service": "test-service",
	})

	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	// Now returns 2 enrichments: consolidated "Alert Details" + custom title
	assert.Equal(t, 2, result.Data.(map[string]interface{})["enrichments_added"])
	assert.Equal(t, "TestAlert", result.Data.(map[string]interface{})["alert_name"])
	assert.Equal(t, "default", result.Data.(map[string]interface{})["namespace"])
	assert.Equal(t, "issue_enrichment", result.Metadata["action_type"])
	assert.Equal(t, "metadata", result.Metadata["enrichment_type"])
	assert.Len(t, result.Enrichments, 2)
}

func TestIssueEnrichmentAction_Execute_MinimalConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:       "test-issue-enrichment",
		Type:       "issue_enrichment",
		Enabled:    true,
		Parameters: map[string]interface{}{},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewIssueEnrichmentAction(config, mockLogger, mockMetrics)

	// Create test event
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "default", nil)

	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	// Should have 1 enrichment (consolidated "Alert Details" with metadata and labels) since include_metadata and include_labels default to true
	assert.Equal(t, 1, result.Data.(map[string]interface{})["enrichments_added"])
	assert.Len(t, result.Enrichments, 1)
}

func TestIssueEnrichmentAction_Execute_NoAlertsInEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-issue-enrichment",
		Type:    "issue_enrichment",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_metadata": true,
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewIssueEnrichmentAction(config, mockLogger, mockMetrics)

	// Create test event with no alerts
	alertEvent := createTestAlertManagerEventNoAlerts()

	result, err := action.Execute(context.Background(), alertEvent)

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "no alerts found in event")
}

func TestIssueEnrichmentAction_Execute_OnlyCustomTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-issue-enrichment",
		Type:    "issue_enrichment",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_metadata": false,
			"include_labels":   false,
			"custom_title":     "Custom Alert Title",
		},
	}

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	action := NewIssueEnrichmentAction(config, mockLogger, mockMetrics)

	// Create test event
	alertEvent := createTestAlertManagerEvent("firing", "TestAlert", "critical", "default", nil)

	result, err := action.Execute(context.Background(), alertEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 1, result.Data.(map[string]interface{})["enrichments_added"])
	assert.Len(t, result.Enrichments, 1)
}

func TestIssueEnrichmentAction_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-issue-enrichment",
		Type:    "issue_enrichment",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_metadata": true,
		},
	}

	action := NewIssueEnrichmentAction(config, mockLogger, mockMetrics)
	err := action.Validate()
	assert.NoError(t, err)
}

func TestIssueEnrichmentActionFactory_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	factory := NewIssueEnrichmentActionFactory(mockLogger, mockMetrics)

	config := actions_interfaces.ActionConfig{
		Name:    "test-issue-enrichment",
		Type:    "issue_enrichment",
		Enabled: true,
		Parameters: map[string]interface{}{
			"include_metadata": true,
		},
	}

	action, err := factory.Create(config)
	require.NoError(t, err)
	assert.NotNil(t, action)
	assert.Equal(t, "test-issue-enrichment", action.GetName())
}

func TestIssueEnrichmentActionFactory_GetActionType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	factory := NewIssueEnrichmentActionFactory(mockLogger, mockMetrics)
	assert.Equal(t, "issue_enrichment", factory.GetActionType())
}

func TestIssueEnrichmentActionFactory_ValidateConfig_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	factory := NewIssueEnrichmentActionFactory(mockLogger, mockMetrics)

	config := actions_interfaces.ActionConfig{
		Name:       "test-action",
		Type:       "invalid_type",
		Enabled:    true,
		Parameters: map[string]interface{}{},
	}

	err := factory.ValidateConfig(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action type for IssueEnrichmentActionFactory: invalid_type")
}

// ============================================================================
// Template Parsing Tests
// ============================================================================

func TestIssueEnrichmentAction_ParseTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	action := &IssueEnrichmentAction{
		BaseAction: NewBaseAction(
			actions_interfaces.ActionConfig{
				Name:    "test-enrichment",
				Type:    "issue_enrichment",
				Enabled: true,
			},
			mockLogger,
			mockMetrics,
		),
	}

	t.Run("parses alert_name template variable", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "KubePodCrashLooping",
					},
				},
			},
		}

		result := action.parseTemplate("Alert: {{.alert_name}}", alertEvent)
		assert.Equal(t, "Alert: KubePodCrashLooping", result)
	})

	t.Run("parses namespace template variable", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "PodIssue",
						"namespace": "production",
					},
				},
			},
		}

		result := action.parseTemplate("Namespace: {{.namespace}}", alertEvent)
		assert.Equal(t, "Namespace: production", result)
	})

	t.Run("parses pod label template variable", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "PodIssue",
						"pod":       "nginx-app-123",
					},
				},
			},
		}

		result := action.parseTemplate("Pod: {{.pod}}", alertEvent)
		assert.Equal(t, "Pod: nginx-app-123", result)
	})

	t.Run("parses multiple template variables", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "KubePodCrashLooping",
						"namespace": "production",
						"pod":       "nginx-app-123",
						"container": "nginx",
					},
				},
			},
		}

		result := action.parseTemplate("{{.alert_name}} in {{.namespace}}/{{.pod}}/{{.container}}", alertEvent)
		assert.Equal(t, "KubePodCrashLooping in production/nginx-app-123/nginx", result)
	})

	t.Run("returns raw template on parsing error", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "TestAlert",
					},
				},
			},
		}

		// Invalid template syntax
		result := action.parseTemplate("{{.invalid syntax", alertEvent)
		assert.Equal(t, "{{.invalid syntax", result)
	})

	t.Run("returns string without template markers unchanged", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "TestAlert",
					},
				},
			},
		}

		result := action.parseTemplate("Plain text without templates", alertEvent)
		assert.Equal(t, "Plain text without templates", result)
	})

	t.Run("handles missing label gracefully", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "TestAlert",
						// pod label is missing
					},
				},
			},
		}

		// Go's text/template renders missing fields as "<no value>"
		result := action.parseTemplate("Pod: {{.pod}}", alertEvent)
		assert.Equal(t, "Pod: <no value>", result)
	})
}

func TestIssueEnrichmentAction_CreateTitleEnrichment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	action := &IssueEnrichmentAction{
		BaseAction: NewBaseAction(
			actions_interfaces.ActionConfig{
				Name:    "test-enrichment",
				Type:    "issue_enrichment",
				Enabled: true,
			},
			mockLogger,
			mockMetrics,
		),
	}

	t.Run("creates enrichment with parsed template", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "KubePodCrashLooping",
						"namespace": "production",
					},
				},
			},
		}

		enrichment := action.createTitleEnrichment("Alert {{.alert_name}} in {{.namespace}}", alertEvent)

		require.NotNil(t, enrichment)
		require.Len(t, enrichment.Blocks, 1)

		// Check that template was parsed correctly in the markdown block
		markdownBlock := enrichment.Blocks[0]
		assert.Contains(t, markdownBlock.BlockType(), "markdown")
	})

	t.Run("creates enrichment with plain text", func(t *testing.T) {
		alertEvent := &event.AlertManagerEvent{
			BaseEvent: event.BaseEvent{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Source:    "alertmanager",
				Type:      event.EventTypeAlertManager,
			},
			Alerts: []event.PrometheusAlert{
				{
					Labels: map[string]string{
						"alertname": "TestAlert",
					},
				},
			},
		}

		enrichment := action.createTitleEnrichment("Static Title", alertEvent)

		require.NotNil(t, enrichment)
		require.Len(t, enrichment.Blocks, 1)
	})
}
