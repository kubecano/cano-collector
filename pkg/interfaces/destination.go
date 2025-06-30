package interfaces

import "context"

//go:generate mockgen -destination=../../mocks/destination_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/interfaces DestinationInterface
type DestinationInterface interface {
	Send(ctx context.Context, message string) error
}
