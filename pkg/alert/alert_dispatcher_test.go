package alert

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/alert/model"
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

	// Allow any number of arguments for logger methods
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	dispatcher := NewAlertDispatcher(mockRegistry, mockFormatter, mockLogger)

	return alertDispatcherTestDeps{
		ctrl:       ctrl,
		registry:   mockRegistry,
		formatter:  mockFormatter,
		logger:     mockLogger,
		dispatcher: dispatcher,
	}
}

func createTestAlertManagerEvent() *model.AlertManagerEvent {
	return &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []model.PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "TestAlert",
					"severity":  "critical",
				},
			},
		},
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

	alert := createTestAlertManagerEvent()

	formattedMessage := "🚨 Alert: TestAlert\nStatus: firing\nSeverity: critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.NoError(t, err)
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

	alert := createTestAlertManagerEvent()

	formattedMessage := "🚨 Alert: TestAlert\nStatus: firing\nSeverity: critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test", "email-test"}).Return([]interfaces.DestinationInterface{mockDestination1, mockDestination2}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination1.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	mockDestination2.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_NilTeam(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	alert := createTestAlertManagerEvent()

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, nil)

	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_TeamWithoutDestinations(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{}, // empty destinations list
	}

	alert := createTestAlertManagerEvent()

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_RegistryError(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := createTestAlertManagerEvent()

	registryError := errors.New("failed to get destinations")

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return(nil, registryError)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get destinations")
}

func TestAlertDispatcher_DispatchAlert_DestinationSendError(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := createTestAlertManagerEvent()

	formattedMessage := "🚨 Alert: TestAlert\nStatus: firing\nSeverity: critical"
	sendError := errors.New("failed to send alert")

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(sendError)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send alert")
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

	alert := createTestAlertManagerEvent()

	formattedMessage := "🚨 Alert: TestAlert\nStatus: firing\nSeverity: critical"
	sendError := errors.New("failed to send alert to destination 2")

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test", "email-test"}).Return([]interfaces.DestinationInterface{mockDestination1, mockDestination2}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination1.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)
	mockDestination2.EXPECT().Send(gomock.Any(), formattedMessage).Return(sendError)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send alert to destination 2")
}

func TestAlertDispatcher_DispatchAlert_ComplexAlert(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	now := time.Now()
	alert := &model.AlertManagerEvent{
		Receiver:    "test-receiver",
		Status:      "firing",
		GroupLabels: map[string]string{"namespace": "production", "service": "api"},
		Alerts: []model.PrometheusAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert", "severity": "critical", "instance": "api-1"},
				Annotations: map[string]string{"summary": "High CPU usage detected", "description": "CPU usage exceeded 90%"},
				StartsAt:    now,
			},
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "HighMemoryUsage", "severity": "warning", "instance": "api-2"},
				Annotations: map[string]string{"summary": "High memory usage detected"},
				StartsAt:    now,
			},
		},
	}

	formattedMessage := "🚨 Alert: TestAlert\nStatus: firing\nSeverity: critical\nSummary: High CPU usage detected\nDescription: CPU usage exceeded 90%\nSummary: High memory usage detected"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_ResolvedAlert(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	now := time.Now()
	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "resolved",
		Alerts: []model.PrometheusAlert{
			{
				Status: "resolved",
				Labels: map[string]string{
					"alertname": "TestAlert",
					"severity":  "critical",
				},
				StartsAt: now,
			},
		},
	}

	formattedMessage := "🚨 Alert: TestAlert\nStatus: resolved\nSeverity: critical"

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(nil)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchAlert_ContextCancellation(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)

	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"slack-test"},
	}

	alert := createTestAlertManagerEvent()

	formattedMessage := "🚨 Alert: TestAlert\nStatus: firing\nSeverity: critical"
	ctxError := context.DeadlineExceeded

	// Setup expectations
	deps.registry.EXPECT().GetDestinations([]string{"slack-test"}).Return([]interfaces.DestinationInterface{mockDestination}, nil)
	deps.formatter.EXPECT().FormatAlert(alert).Return(formattedMessage)
	mockDestination.EXPECT().Send(gomock.Any(), formattedMessage).Return(ctxError)

	err := deps.dispatcher.DispatchAlert(context.Background(), alert, team)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
