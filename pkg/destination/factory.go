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

// config: może być np. config_destination.SlackDestination
func (f *DestinationFactory) CreateDestination(config interface{}) (interface{}, error) {
	switch d := config.(type) {
	case config_destination.DestinationSlack:
		return f.createSlackDestination(&d)
	default:
		return nil, fmt.Errorf("unsupported destination type: %T", config)
	}
}

func (f *DestinationFactory) createSlackDestination(d *config_destination.DestinationSlack) (*destslack.DestinationSlack, error) {
	if d.APIKey == "" {
		return nil, fmt.Errorf("slack destination '%s' must have api_key", d.Name)
	}
	cfg := &destslack.DestinationSlackConfig{
		Name:             d.Name,
		APIKey:           d.APIKey,
		SlackChannel:     d.SlackChannel,
		GroupingInterval: d.GroupingInterval,
		UnfurlLinks:      d.UnfurlLinks != nil && *d.UnfurlLinks,
	}
	return destslack.NewDestinationSlack(cfg, f.logger, f.httpClient), nil
}
