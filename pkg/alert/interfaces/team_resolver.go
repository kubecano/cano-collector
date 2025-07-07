package interfaces

import (
	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/alert/model"
	destination_interfaces "github.com/kubecano/cano-collector/pkg/destination/interfaces"
)

// TeamResolverInterface defines the interface for resolving which team should handle an alert.
//
//go:generate mockgen -destination=../../../mocks/team_resolver_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert/interfaces TeamResolverInterface
type TeamResolverInterface interface {
	ResolveTeam(alert *model.AlertManagerEvent) (*config_team.Team, error)
	ValidateTeamDestinations(registry destination_interfaces.DestinationRegistryInterface) error
}
