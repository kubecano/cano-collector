package interfaces

import (
	"context"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

//go:generate mockgen -source=sender.go -destination=../../../mocks/sender_mock.go -package=mocks
type DestinationSenderInterface interface {
	Send(ctx context.Context, issue *issuepkg.Issue) error
}
