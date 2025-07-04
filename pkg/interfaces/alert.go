package interfaces

import (
	"context"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/alert/model"
)

// AlertDispatcherInterface defines the interface for dispatching alerts to team destinations.
//
//go:generate mockgen -destination=../../mocks/alert_dispatcher_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces AlertDispatcherInterface
type AlertDispatcherInterface interface {
	DispatchAlert(ctx context.Context, alert *model.AlertManagerEvent, team *config_team.Team) error
}

// AlertFormatterInterface defines the interface for formatting alerts into readable messages.
//
//go:generate mockgen -destination=../../mocks/alert_formatter_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces AlertFormatterInterface
type AlertFormatterInterface interface {
	FormatAlert(alert *model.AlertManagerEvent) string
}

// TeamResolverInterface defines the interface for resolving which team should handle an alert.
//
//go:generate mockgen -destination=../../mocks/team_resolver_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces TeamResolverInterface
type TeamResolverInterface interface {
	ResolveTeam(alert *model.AlertManagerEvent) (*config_team.Team, error)
	ValidateTeamDestinations(registry DestinationRegistryInterface) error
}
