package destination

import (
	"fmt"
	"sync"

	config_destination "github.com/kubecano/cano-collector/config/destination"
	"github.com/kubecano/cano-collector/pkg/logger"
)

//go:generate mockgen -destination=../../mocks/destination_registry_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/destination DestinationRegistryInterface
type DestinationRegistryInterface interface {
	GetDestination(name string) (DestinationInterface, error)
	GetDestinations(names []string) ([]DestinationInterface, error)
	RegisterDestination(name string, destination DestinationInterface)
}

// DestinationRegistry manages a registry of destinations
type DestinationRegistry struct {
	destinations map[string]DestinationInterface
	factory      *DestinationFactory
	logger       logger.LoggerInterface
	mu           sync.RWMutex
}

// NewDestinationRegistry creates a new destination registry
func NewDestinationRegistry(factory *DestinationFactory, logger logger.LoggerInterface) *DestinationRegistry {
	return &DestinationRegistry{
		destinations: make(map[string]DestinationInterface),
		factory:      factory,
		logger:       logger,
	}
}

// LoadFromConfig loads destinations from configuration
func (r *DestinationRegistry) LoadFromConfig(config config_destination.DestinationsConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load Slack destinations
	for _, slackConfig := range config.Destinations.Slack {
		destination, err := r.factory.CreateDestination(slackConfig)
		if err != nil {
			return fmt.Errorf("failed to create slack destination '%s': %w", slackConfig.Name, err)
		}

		if dest, ok := destination.(DestinationInterface); ok {
			r.destinations[slackConfig.Name] = dest
			r.logger.Info("Registered destination", "name", slackConfig.Name, "type", "slack")
		} else {
			return fmt.Errorf("destination '%s' does not implement DestinationInterface", slackConfig.Name)
		}
	}

	return nil
}

// GetDestination returns a destination by name
func (r *DestinationRegistry) GetDestination(name string) (DestinationInterface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	destination, exists := r.destinations[name]
	if !exists {
		return nil, fmt.Errorf("destination '%s' not found", name)
	}

	return destination, nil
}

// GetDestinations returns multiple destinations by names
func (r *DestinationRegistry) GetDestinations(names []string) ([]DestinationInterface, error) {
	var destinations []DestinationInterface

	for _, name := range names {
		destination, err := r.GetDestination(name)
		if err != nil {
			return nil, err
		}
		destinations = append(destinations, destination)
	}

	return destinations, nil
}

// RegisterDestination manually registers a destination
func (r *DestinationRegistry) RegisterDestination(name string, destination DestinationInterface) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.destinations[name] = destination
	r.logger.Info("Manually registered destination", "name", name)
}
