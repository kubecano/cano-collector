package interfaces

//go:generate mockgen -source=factory.go -destination=../../../mocks/destination_factory_mock.go -package=mocks

// DestinationFactoryInterface defines the interface for creating destinations
type DestinationFactoryInterface interface {
	CreateDestination(config interface{}) (interface{}, error)
}
