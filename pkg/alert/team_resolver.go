package alert

import (
	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/prometheus/alertmanager/template"
)

//go:generate mockgen -destination=../../mocks/team_resolver_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert TeamResolverInterface
type TeamResolverInterface interface {
	ResolveTeam(alert template.Data) (*config_team.Team, error)
}

// TeamResolver resolves which team should handle an alert
type TeamResolver struct {
	teams  config_team.TeamsConfig
	logger logger.LoggerInterface
}

// NewTeamResolver creates a new team resolver
func NewTeamResolver(teams config_team.TeamsConfig, logger logger.LoggerInterface) *TeamResolver {
	return &TeamResolver{
		teams:  teams,
		logger: logger,
	}
}

// ResolveTeam determines which team should handle the alert
// For now, returns the first team (default team) as specified in requirements
func (r *TeamResolver) ResolveTeam(alert template.Data) (*config_team.Team, error) {
	if len(r.teams.Teams) == 0 {
		return nil, nil // No teams configured
	}

	// TODO: Implement proper routing logic based on namespace, pod names, etc.
	// For now, return the first team as the default team
	defaultTeam := r.teams.Teams[0]
	r.logger.Info("Resolved team for alert", "team", defaultTeam.Name, "destinations", defaultTeam.Destinations)

	return &defaultTeam, nil
}
