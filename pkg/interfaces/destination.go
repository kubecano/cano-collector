package interfaces

import (
	"context"

	"github.com/kubecano/cano-collector/pkg/core/issue"
)

//go:generate mockgen -destination=../../mocks/destination_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces DestinationInterface
type DestinationInterface interface {
	Send(ctx context.Context, issue *issue.Issue) error
}
