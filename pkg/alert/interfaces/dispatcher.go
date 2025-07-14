package interfaces

import (
	"context"

	config_team "github.com/kubecano/cano-collector/config/team"
	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

// AlertDispatcherInterface defines the interface for dispatching issues to team destinations.
//
//go:generate mockgen -source=dispatcher.go -destination=../../../mocks/alert_dispatcher_mock.go -package=mocks
type AlertDispatcherInterface interface {
	DispatchIssues(ctx context.Context, issues []*issuepkg.Issue, team *config_team.Team) error
}
