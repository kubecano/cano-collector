package interfaces

import (
	"context"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

//go:generate mockgen -source=destination.go -destination=../../../mocks/destination_mock.go -package=mocks
type DestinationInterface interface {
	Send(ctx context.Context, issue *issuepkg.Issue) error
}
