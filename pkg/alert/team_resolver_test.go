package alert

import (
	"testing"

	"github.com/golang/mock/gomock"
	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

type teamResolverTestDeps struct {
	ctrl     *gomock.Controller
	logger   *mocks.MockLoggerInterface
	resolver *TeamResolver
}

func setupTeamResolverTest(t *testing.T, teams config_team.TeamsConfig) teamResolverTestDeps {
	ctrl := gomock.NewController(t)
	logger := mocks.NewMockLoggerInterface(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{{
			Status: "firing",
			Labels: map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{{
			Status: "firing",
			Labels: map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "team-1", team.Name)
	assert.Equal(t, []string{"slack-1"}, team.Destinations)
}

func TestTeamResolver_ResolveTeam_NoTeams(t *testing.T) {
	teams := config_team.TeamsConfig{Teams: []config_team.Team{}}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{{
			Status: "firing",
			Labels: map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
	assert.Nil(t, team)
}

func TestTeamResolver_ResolveTeam_TeamWithoutDestinations(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "team-no-dest", Destinations: []string{}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{{
			Status: "firing",
			Labels: map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
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

	alert := template.Data{
		Receiver:    "test-receiver",
		Status:      "firing",
		GroupLabels: map[string]string{"namespace": "production", "service": "api"},
		Alerts: []template.Alert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "HighCPUUsage", "severity": "critical", "instance": "api-1"},
				Annotations: map[string]string{"summary": "High CPU usage detected", "description": "CPU usage exceeded 90%"},
			},
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "HighMemoryUsage", "severity": "warning", "instance": "api-2"},
				Annotations: map[string]string{"summary": "High memory usage detected"},
			},
		},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "resolved",
		Alerts: []template.Alert{{
			Status: "resolved",
			Labels: map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
}

func TestTeamResolver_ResolveTeam_EmptyAlert(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   []template.Alert{},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
}

func TestTeamResolver_ResolveTeam_AlertWithSpecialCharacters(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "default-team", Destinations: []string{"slack-default"}}},
	}
	deps := setupTeamResolverTest(t, teams)
	defer deps.ctrl.Finish()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{{
			Status:      "firing",
			Labels:      map[string]string{"alertname": "High CPU Usage (API)", "severity": "critical", "instance": "api-1.prod.example.com"},
			Annotations: map[string]string{"summary": "CPU usage > 90% for 5 minutes"},
		}},
	}

	team, err := deps.resolver.ResolveTeam(alert)
	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "default-team", team.Name)
}

func TestTeamResolver_ResolveTeam_LoggingVerification(t *testing.T) {
	teams := config_team.TeamsConfig{
		Teams: []config_team.Team{{Name: "test-team", Destinations: []string{"slack-test"}}},
	}
	ctrl := gomock.NewController(t)
	logger := mocks.NewMockLoggerInterface(ctrl)
	logger.EXPECT().Info(
		"Resolved team for alert",
		"team", "test-team",
		"destinations", []string{"slack-test"},
	).Times(1)
	resolver := NewTeamResolver(teams, logger)
	defer ctrl.Finish()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{{
			Status: "firing",
			Labels: map[string]string{"alertname": "HighCPUUsage", "severity": "critical"},
		}},
	}

	team, err := resolver.ResolveTeam(alert)
	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "test-team", team.Name)
}
