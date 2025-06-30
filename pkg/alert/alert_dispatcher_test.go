package alert

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/interfaces"
)

type alertDispatcherTestDeps struct {
	ctrl       *gomock.Controller
	registry   *mocks.MockDestinationRegistryInterface
	formatter  *mocks.MockAlertFormatterInterface
	logger     *mocks.MockLoggerInterface
	dispatcher *AlertDispatcher
}

// setupAlertDispatcherTest initializes mocks and dispatcher for tests
func setupAlertDispatcherTest(t *testing.T) alertDispatcherTestDeps {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRegistry := mocks.NewMockDestinationRegistryInterface(ctrl)
	mockFormatter := mocks.NewMockAlertFormatterInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	dispatcher := NewAlertDispatcher(mockRegistry, mockFormatter, mockLogger)

	return alertDispatcherTestDeps{
		ctrl:       ctrl,
		registry:   mockRegistry,
		formatter:  mockFormatter,
		logger:     mockLogger,
		dispatcher: dispatcher,
	}
}

func TestAlertDispatcher_DispatchAlert_Success(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	formattedMessage := "🚨 **Alert: firing**\n**Alert:** HighCPUUsage\n**Status:** firing\n**Severity:** critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	deps.logger.EXPECT().Info("Alert sent successfully to destination", gomock.Any()).Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	assert.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_MultipleDestinations(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination1 := mocks.NewMockDestinationInterface(deps.ctrl)
	mockDestination2 := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test", "email-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	formattedMessage := "🚨 **Alert: firing**\n**Alert:** HighCPUUsage\n**Status:** firing\n**Severity:** critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test", "email-test"}).Return([]interfaces.DestinationInterface{mockDestination1, mockDestination2}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination1.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	mockDestination2.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	deps.logger.EXPECT().Info("Alert sent successfully to destination", gomock.Any()).Times(2)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	assert.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_NilTeam(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	// Setup expectations
	deps.logger.EXPECT().Info("No team resolved for alert, skipping dispatch").Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, nil)

	assert.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_TeamWithoutDestinations(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{}, // empty destinations list
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	// Setup expectations
	deps.logger.EXPECT().Info("Team has no destinations configured", "team", "test-team").Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	assert.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_RegistryError(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	// Setup expectations
	expectedError := errors.New("destination not found")
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return(nil, expectedError)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get destinations for team 'test-team'")
	assert.Contains(t, err.Error(), "destination not found")
}

func TestAlertDispatcher_DispatchAlert_DestinationSendError(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	formattedMessage := "🚨 **Alert: firing**\n**Alert:** HighCPUUsage\n**Status:** firing\n**Severity:** critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	sendError := errors.New("slack API error")
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(sendError)
	deps.logger.EXPECT().Error("Failed to send alert to destination", gomock.Any(), gomock.Any()).Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "some destinations failed")
	assert.Contains(t, err.Error(), "slack API error")
}

func TestAlertDispatcher_DispatchAlert_PartialFailure(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination1 := mocks.NewMockDestinationInterface(deps.ctrl)
	mockDestination2 := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test", "email-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	formattedMessage := "🚨 **Alert: firing**\n**Alert:** HighCPUUsage\n**Status:** firing\n**Severity:** critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test", "email-test"}).Return([]interfaces.DestinationInterface{mockDestination1, mockDestination2}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination1.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	sendError := errors.New("email service unavailable")
	mockDestination2.EXPECT().Send(gomock.Any(), formattedMessage).Return(sendError)
	deps.logger.EXPECT().Info("Alert sent successfully to destination", gomock.Any()).Times(1)
	deps.logger.EXPECT().Error("Failed to send alert to destination", gomock.Any(), gomock.Any()).Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "some destinations failed")
	assert.Contains(t, err.Error(), "email service unavailable")
}

func TestAlertDispatcher_DispatchAlert_ComplexAlert(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		GroupLabels: map[string]string{
			"namespace": "production",
			"service":   "api",
		},
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
					"instance":  "api-1",
				},
				Annotations: map[string]string{
					"summary":     "High CPU usage detected",
					"description": "CPU usage exceeded 90%",
				},
			},
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighMemoryUsage",
					"severity":  "warning",
					"instance":  "api-2",
				},
				Annotations: map[string]string{
					"summary": "High memory usage detected",
				},
			},
		},
	}

	formattedMessage := "Complex formatted message with multiple alerts"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	deps.logger.EXPECT().Info("Alert sent successfully to destination", gomock.Any()).Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	assert.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_ResolvedAlert(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "resolved",
		Alerts: []template.Alert{
			{
				Status: "resolved",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	formattedMessage := "🚨 **Alert: resolved**\n**Alert:** HighCPUUsage\n**Status:** resolved\n**Severity:** critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	deps.logger.EXPECT().Info("Alert sent successfully to destination", gomock.Any()).Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	assert.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_ContextCancellation(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	formattedMessage := "🚨 **Alert: firing**\n**Alert:** HighCPUUsage\n**Status:** firing\n**Severity:** critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	// Simulate context-related error
	contextError := errors.New("context deadline exceeded")
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(contextError)
	deps.logger.EXPECT().Error("Failed to send alert to destination", gomock.Any(), gomock.Any()).Times(1)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
