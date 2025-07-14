package interfaces

import (
	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/core/event"
	destination_interfaces "github.com/kubecano/cano-collector/pkg/destination/interfaces"
)

// TeamResolverInterface defines the interface for resolving which team should handle an alert.
//
//go:generate mockgen -source=team_resolver.go -destination=../../../mocks/team_resolver_mock.go -package=mocks
type TeamResolverInterface interface {
	ResolveTeam(alert *event.AlertManagerEvent) (*config_team.Team, error)
	ValidateTeamDestinations(registry destination_interfaces.DestinationRegistryInterface) error
}
