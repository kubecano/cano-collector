package alert

import (
	"fmt"

	"github.com/prometheus/alertmanager/template"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/interfaces"
	"github.com/kubecano/cano-collector/pkg/logger"
)

//go:generate mockgen -destination=../../mocks/team_resolver_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert TeamResolverInterface
type TeamResolverInterface interface {
	ResolveTeam(alert template.Data) (*config_team.Team, error)
	ValidateTeamDestinations(registry interfaces.DestinationRegistryInterface) error
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
