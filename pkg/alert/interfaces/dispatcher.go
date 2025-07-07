package interfaces

import (
	"context"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/core/issue"
)

// AlertDispatcherInterface defines the interface for dispatching issues to team destinations.
//
//go:generate mockgen -destination=../../../mocks/alert_dispatcher_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert/interfaces AlertDispatcherInterface
type AlertDispatcherInterface interface {
	DispatchIssues(ctx context.Context, issues []*issue.Issue, team *config_team.Team) error
}
