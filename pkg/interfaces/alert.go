package interfaces

import (
	"context"

	config_team "github.com/kubecano/cano-collector/config/team"
)

// AlertDispatcherInterface defines the interface for dispatching alerts to team destinations.
// Note: We use interface{} instead of *AlertManagerEvent to avoid import cycles.
// In practice, this interface should always be called with *AlertManagerEvent.
//
//go:generate mockgen -destination=../../mocks/alert_dispatcher_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces AlertDispatcherInterface
type AlertDispatcherInterface interface {
	DispatchAlert(ctx context.Context, alert interface{}, team *config_team.Team) error
}

// AlertFormatterInterface defines the interface for formatting alerts into readable messages.
// Note: We use interface{} instead of *AlertManagerEvent to avoid import cycles.
// In practice, this interface should always be called with *AlertManagerEvent.
//
//go:generate mockgen -destination=../../mocks/alert_formatter_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces AlertFormatterInterface
type AlertFormatterInterface interface {
	FormatAlert(alert interface{}) string
}

// TeamResolverInterface defines the interface for resolving which team should handle an alert.
// Note: We use interface{} instead of *AlertManagerEvent to avoid import cycles.
// In practice, this interface should always be called with *AlertManagerEvent.
//
//go:generate mockgen -destination=../../mocks/team_resolver_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces TeamResolverInterface
type TeamResolverInterface interface {
	ResolveTeam(alert interface{}) (*config_team.Team, error)
	ValidateTeamDestinations(registry DestinationRegistryInterface) error
}
