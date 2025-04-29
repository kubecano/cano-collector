package destination

import (
	"fmt"

	"github.com/kubecano/cano-collector/config/destination"
	"github.com/kubecano/cano-collector/pkg/sender"
	"github.com/kubecano/cano-collector/pkg/util"

	"github.com/kubecano/cano-collector/pkg/logger"
)

// DestinationFactory creates instances of different destinations
type DestinationFactory struct {
	logger      logger.LoggerInterface
	client      util.HTTPClient
	slackSender sender.SenderInterface
}

// NewDestinationFactory creates a new instance of DestinationFactory
func NewDestinationFactory(logger logger.LoggerInterface, client util.HTTPClient) *DestinationFactory {
	if client == nil {
		client = util.DefaultHTTPClient()
	}
	return &DestinationFactory{
		logger: logger,
		client: client,
	}
}

// WithSlackSender ustawia niestandardowy SlackSender dla fabryki (głównie dla testów)
func (f *DestinationFactory) WithSlackSender(slackSender sender.SenderInterface) *DestinationFactory {
	f.slackSender = slackSender
	return f
}

// CreateSlackDestination creates a Slack destination
func (f *DestinationFactory) CreateSlackDestination(config destination.SlackDestinationConfig) (Destination, error) {
	return NewSlackDestination(
		config.Name,
		config.Token,
		config.Channel,
		config.SigningKey,
		config.AccountID,
		config.ClusterName,
		f.logger,
		f.slackSender,
	)
}

// CreateTeamsDestination creates a Teams destination
func (f *DestinationFactory) CreateTeamsDestination(config destination.TeamsDestinationConfig) (Destination, error) {
	return NewTeamsDestination(
		config.Name,
		config.WebhookURL,
		f.logger,
	)
}

// CreateAllDestinations creates all destinations based on the provided configuration
func (f *DestinationFactory) CreateAllDestinations(config destination.DestinationsConfig) (map[string]Destination, error) {
	result := make(map[string]Destination)
	var errs []error

	// Creates Slack destinations
	for _, slackConfig := range config.Destinations.Slack {
		dest, err := f.CreateSlackDestination(slackConfig)
		if err != nil {
			errs = append(errs, fmt.Errorf("error creating Slack destination %s: %w", slackConfig.Name, err))
			continue
		}
		result[slackConfig.Name] = dest
	}

	// Creates Teams destinations
	for _, teamsConfig := range config.Destinations.Teams {
		dest, err := f.CreateTeamsDestination(teamsConfig)
		if err != nil {
			errs = append(errs, fmt.Errorf("error creating Teams destination %s: %w", teamsConfig.Name, err))
			continue
		}
		result[teamsConfig.Name] = dest
	}

	if len(errs) > 0 {
		return result, fmt.Errorf("errors occurred during destinations creation: %v", errs)
	}

	return result, nil
}
