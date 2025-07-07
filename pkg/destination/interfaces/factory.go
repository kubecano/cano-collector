package interfaces

//go:generate mockgen -destination=../../../mocks/destination_factory_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/destination/interfaces DestinationFactoryInterface

// DestinationFactoryInterface defines the interface for creating destinations
type DestinationFactoryInterface interface {
	CreateDestination(config interface{}) (interface{}, error)
}
