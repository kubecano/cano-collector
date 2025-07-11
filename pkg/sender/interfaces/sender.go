package interfaces

import (
	"context"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

//go:generate mockgen -destination=../../../mocks/sender_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/sender/interfaces DestinationSenderInterface
type DestinationSenderInterface interface {
	Send(ctx context.Context, issue *issuepkg.Issue) error
}
