package interfaces

import (
	"context"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

//go:generate mockgen -destination=../../../mocks/destination_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/destination/interfaces DestinationInterface
type DestinationInterface interface {
	Send(ctx context.Context, issue *issuepkg.Issue) error
}
