package alert

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/core/issue"
)

type alertDispatcherTestDeps struct {
	ctrl       *gomock.Controller
	registry   *mocks.MockDestinationRegistryInterface
	logger     *mocks.MockLoggerInterface
	metrics    *mocks.MockMetricsInterface
	dispatcher *AlertDispatcher
}

// setupAlertDispatcherTest initializes mocks and dispatcher for tests
func setupAlertDispatcherTest(t *testing.T) *alertDispatcherTestDeps {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRegistry := mocks.NewMockDestinationRegistryInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	// Mock logger calls that are always made (variadic arguments)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	// Mock metrics calls
	mockMetrics.EXPECT().IncDestinationMessagesSent(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncDestinationErrors(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().ObserveDestinationSendDuration(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	dispatcher := NewAlertDispatcher(mockRegistry, mockLogger, mockMetrics)

	return &alertDispatcherTestDeps{
		ctrl:       ctrl,
		registry:   mockRegistry,
		logger:     mockLogger,
		metrics:    mockMetrics,
		dispatcher: dispatcher,
	}
}

func TestAlertDispatcher_DispatchIssues_Success(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"test-destination"},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Set up mock expectations
	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)
	deps.registry.EXPECT().GetDestination("test-destination").Return(mockDestination, nil)
	mockDestination.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, team)

	// Verify
	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchIssues_MultipleDestinations(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"dest1", "dest2"},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Set up mock expectations
	mockDestination1 := mocks.NewMockDestinationInterface(deps.ctrl)
	mockDestination2 := mocks.NewMockDestinationInterface(deps.ctrl)
	deps.registry.EXPECT().GetDestination("dest1").Return(mockDestination1, nil)
	deps.registry.EXPECT().GetDestination("dest2").Return(mockDestination2, nil)
	mockDestination1.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	mockDestination2.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, team)

	// Verify
	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchIssues_NilTeam(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, nil)

	// Verify
	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchIssues_EmptyDestinations(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, team)

	// Verify
	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchIssues_EmptyIssues(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"test-destination"},
	}

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), []*issue.Issue{}, team)

	// Verify
	require.NoError(t, err)
}

func TestAlertDispatcher_DispatchIssues_DestinationNotFound(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"non-existent-destination"},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Set up mock expectations
	deps.registry.EXPECT().GetDestination("non-existent-destination").Return(nil, errors.New("destination not found"))

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, team)

	// Verify
	require.Error(t, err)
	assert.Contains(t, err.Error(), "destination not found")
}

func TestAlertDispatcher_DispatchIssues_SendError(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"test-destination"},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Set up mock expectations
	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)
	deps.registry.EXPECT().GetDestination("test-destination").Return(mockDestination, nil)
	mockDestination.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send failed"))

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, team)

	// Verify
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send failed")
}

func TestAlertDispatcher_DispatchIssues_PartialFailure(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"dest1", "dest2"},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Set up mock expectations
	mockDestination1 := mocks.NewMockDestinationInterface(deps.ctrl)
	mockDestination2 := mocks.NewMockDestinationInterface(deps.ctrl)
	deps.registry.EXPECT().GetDestination("dest1").Return(mockDestination1, nil)
	deps.registry.EXPECT().GetDestination("dest2").Return(mockDestination2, nil)
	mockDestination1.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	mockDestination2.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send failed"))

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, team)

	// Verify
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send failed")
}

func TestAlertDispatcher_DispatchIssues_ContextCancellation(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"test-destination"},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue",
			Severity: issue.SeverityHigh,
		},
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Execute
	err := deps.dispatcher.DispatchIssues(ctx, issues, team)

	// Verify
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestAlertDispatcher_DispatchIssues_MultipleIssues(t *testing.T) {
	deps := setupAlertDispatcherTest(t)
	defer deps.ctrl.Finish()

	// Create test data
	team := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"test-destination"},
	}

	issues := []*issue.Issue{
		{
			Title:    "Test Issue 1",
			Severity: issue.SeverityHigh,
		},
		{
			Title:    "Test Issue 2",
			Severity: issue.SeverityLow,
		},
	}

	// Set up mock expectations
	mockDestination := mocks.NewMockDestinationInterface(deps.ctrl)
	deps.registry.EXPECT().GetDestination("test-destination").Return(mockDestination, nil)
	mockDestination.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	// Execute
	err := deps.dispatcher.DispatchIssues(context.Background(), issues, team)

	// Verify
	require.NoError(t, err)
}
