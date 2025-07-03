package alert

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"
)

type teamResolverTestDeps struct {
	ctrl     *gomock.Controller
	logger   *mocks.MockLoggerInterface
	resolver *TeamResolver
}

func setupTeamResolverTest(t *testing.T, teams config_team.TeamsConfig) teamResolverTestDeps {
	t.Helper()
	ctrl := gomock.NewController(t)
	logger := mocks.NewMockLoggerInterface(ctrl)

	// Allow any number of arguments
	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	resolver := NewTeamResolver(teams, logger)
	return teamResolverTestDeps{ctrl, logger, resolver}
}

func TestTeamResolver_ResolveTeam_DefaultTeam(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{
			Name: "default-team", Destinations: []string{"slack-default"},
		}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{{
			Status:   "firing",
			Labels:   map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
			StartsAt: now,
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
	assert.Equal(t, []string{"slack-default"}, team.Destinations)
}

func TestTeamResolver_ResolveTeam_MultipleTeams(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{
			{Name: "team-1", Destinations: []string{"slack-1"}},
			{Name: "team-2", Destinations: []string{"slack-2", "email-2"}},
			{Name: "team-3", Destinations: []string{}},
		},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{{
			Status:   "firing",
			Labels:   map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
			StartsAt: now,
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "team-1", team.Name)
	assert.Equal(t, []string{"slack-1"}, team.Destinations)
}

func TestTeamResolver_ResolveTeam_NoTeams(t *testing.T) {
	teams := config_team.TeamsConfig{Teams: []config_team.Team{}}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{{
			Status:   "firing",
			Labels:   map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
			StartsAt: now,
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.Nil(t, team)
}

func TestTeamResolver_ResolveTeam_TeamWithoutDestinations(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "team-no-dest", Destinations: []string{}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{{
			Status:   "firing",
			Labels:   map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
			StartsAt: now,
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "team-no-dest", team.Name)
	assert.Empty(t, team.Destinations)
}

func TestTeamResolver_ResolveTeam_ComplexAlert(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default", "email-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver:    "test-receiver",
		Status:      "firing",
		GroupLabels: map[string]string{"namespace": "production", "service": "api"},
		Alerts: []PrometheusAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "HighCPUUsage", "severity": "critical", "instance": "api-1"},
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

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
	assert.Equal(t, []string{"slack-default", "email-default"}, team.Destinations)
}

func TestTeamResolver_ResolveTeam_ResolvedAlert(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "resolved",
		Alerts: []PrometheusAlert{{
			Status:   "resolved",
			Labels:   map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
			StartsAt: now,
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
}

func TestTeamResolver_ResolveTeam_EmptyAlert(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   []PrometheusAlert{},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
}

func TestTeamResolver_ResolveTeam_AlertWithSpecialCharacters(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{{
			Status: "firing",
			Labels: map[string]string{
				"alertname": "HighCPUUsage",
				"severity":  "critical",
				"special":   "test@example.com",
				"unicode":   "🚨🔥💻",
			},
			StartsAt: now,
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
}

func TestTeamResolver_ResolveTeam_LoggingVerification(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	now := time.Now()
	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{{
			Status:   "firing",
			Labels:   map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
			StartsAt: now,
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
}

func TestTeamResolver_ResolveTeam_InvalidAlertType(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	team, err := deps.resolver.ResolveTeam("not an alert")
	require.Error(t, err)
	assert.Nil(t, team)
	assert.Contains(t, err.Error(), "invalid alert type")
}
