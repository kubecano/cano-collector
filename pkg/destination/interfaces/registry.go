package interfaces

import (
	config_destination "github.com/kubecano/cano-collector/config/destination"
)

//go:generate mockgen -source=registry.go -destination=../../../mocks/destination_registry_mock.go -package=mocks

// DestinationRegistryInterface defines the interface for destination registry
type DestinationRegistryInterface interface {
	GetDestination(name string) (DestinationInterface, error)
	GetDestinations(names []string) ([]DestinationInterface, error)
	RegisterDestination(name string, destination DestinationInterface)
	LoadFromConfig(config config_destination.DestinationsConfig) error
}
