package alert

import (
	"fmt"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/alert/model"
	"github.com/kubecano/cano-collector/pkg/interfaces"
	"github.com/kubecano/cano-collector/pkg/logger"
)

// TeamResolver resolves which team should handle an alert
type TeamResolver struct {
	teams   config_team.TeamsConfig
	logger  logger.LoggerInterface
	metrics interfaces.MetricsInterface
}

// NewTeamResolver creates a new team resolver
func NewTeamResolver(teams config_team.TeamsConfig, logger logger.LoggerInterface, metrics interfaces.MetricsInterface) *TeamResolver {
	return &TeamResolver{
		teams:   teams,
		logger:  logger,
		metrics: metrics,
	}
}

// ValidateTeamDestinations validates that all team destinations exist in the registry
func (r *TeamResolver) ValidateTeamDestinations(registry interfaces.DestinationRegistryInterface) error {
	for _, team := range r.teams.Teams {
		for _, destName := range team.Destinations {
			if _, err := registry.GetDestination(destName); err != nil {
				return fmt.Errorf("team '%s' references non-existent destination '%s': %w", team.Name, destName, err)
			}
		}
	}
	return nil
}

// ResolveTeam determines which team should handle the alert
// For now, returns the first team (default team) as specified in requirements
func (r *TeamResolver) ResolveTeam(alertEvent *model.AlertManagerEvent) (*config_team.Team, error) {
	if len(r.teams.Teams) == 0 {
		r.metrics.IncRoutingDecisions("no_team", "none", "no_teams_configured")
		return nil, nil // No teams configured
	}

	// TODO: Implement proper routing logic based on namespace, pod names, etc.
	// For now, return the first team as the default team
	defaultTeam := r.teams.Teams[0]
	r.logger.Info("Resolved team for alert",
		"team", defaultTeam.Name,
		"destinations", defaultTeam.Destinations,
		"alert_name", alertEvent.GetAlertName())

	// Record team matching metrics
	r.metrics.IncTeamsMatched(defaultTeam.Name, alertEvent.GetAlertName())

	// Record routing decision metrics based on team's destinations
	for range defaultTeam.Destinations {
		r.metrics.IncRoutingDecisions(defaultTeam.Name, "unknown", "routed") // TODO: Get actual destination type
	}

	return &defaultTeam, nil
}
