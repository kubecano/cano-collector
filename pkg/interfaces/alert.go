package interfaces

import (
	"context"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/alert/model"
	"github.com/kubecano/cano-collector/pkg/core/issue"
)

// AlertDispatcherInterface defines the interface for dispatching issues to team destinations.
//
//go:generate mockgen -destination=../../mocks/alert_dispatcher_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces AlertDispatcherInterface
type AlertDispatcherInterface interface {
	DispatchIssues(ctx context.Context, issues []*issue.Issue, team *config_team.Team) error
}

// TeamResolverInterface defines the interface for resolving which team should handle an alert.
//
//go:generate mockgen -destination=../../mocks/team_resolver_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces TeamResolverInterface
type TeamResolverInterface interface {
	ResolveTeam(alert *model.AlertManagerEvent) (*config_team.Team, error)
	ValidateTeamDestinations(registry DestinationRegistryInterface) error
}
