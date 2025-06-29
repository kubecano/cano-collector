package destination

import (
	"fmt"

	config_destination "github.com/kubecano/cano-collector/config/destination"
	destslack "github.com/kubecano/cano-collector/pkg/destination/slack"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

type DestinationFactory struct {
	logger     logger.LoggerInterface
	httpClient util.HTTPClient
}

func NewDestinationFactory(logger logger.LoggerInterface, httpClient util.HTTPClient) *DestinationFactory {
	if httpClient == nil {
		httpClient = util.DefaultHTTPClient()
	}
	return &DestinationFactory{
		logger:     logger,
		httpClient: httpClient,
	}
}

// CreateDestination creates appropriate Destination based on destination configuration
func (f *DestinationFactory) CreateDestination(dest interface{}) (DestinationInterface, error) {
	var destination DestinationInterface

	switch d := dest.(type) {
	case config_destination.DestinationSlack:
		if d.APIKey == "" {
			return nil, fmt.Errorf("slack destination '%s' must have api_key", d.Name)
		}
		destination = destslack.NewDestinationSlack(&destslack.DestinationSlackConfig{
			Name:             d.Name,
			APIKey:           d.APIKey,
			SlackChannel:     d.SlackChannel,
			GroupingInterval: d.GroupingInterval,
			UnfurlLinks:      *d.UnfurlLinks,
		}, f.logger, f.httpClient)

	default:
		return nil, fmt.Errorf("unsupported destination type: %T", dest)
	}

	return destination, nil
}
