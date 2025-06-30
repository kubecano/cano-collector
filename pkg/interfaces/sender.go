package interfaces

import "context"

//go:generate mockgen -destination=../../mocks/sender_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces DestinationSenderInterface
type DestinationSenderInterface interface {
	Send(ctx context.Context, message string) error
}
