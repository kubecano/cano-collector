package destination

import (
	"fmt"

	config_destination "github.com/kubecano/cano-collector/config/destination"
	destslack "github.com/kubecano/cano-collector/pkg/destination/slack"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	"github.com/kubecano/cano-collector/pkg/util"
)

type DestinationFactory struct {
	logger     logger_interfaces.LoggerInterface
	httpClient util.HTTPClient
}

func NewDestinationFactory(logger logger_interfaces.LoggerInterface, httpClient util.HTTPClient) *DestinationFactory {
	if httpClient == nil {
		httpClient = util.DefaultHTTPClient()
	}
	return &DestinationFactory{
		logger:     logger,
		httpClient: httpClient,
	}
}

// config: can be e.g. config_destination.SlackDestination
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
		UnfurlLinks:      d.UnfurlLinks == nil || *d.UnfurlLinks,
	}

	// Convert threading configuration if present
	if d.Threading != nil {
		cfg.Threading = &destslack.SlackThreadingConfig{
			Enabled:               d.Threading.Enabled,
			CacheTTL:              d.Threading.CacheTTL,
			SearchLimit:           d.Threading.SearchLimit,
			SearchWindow:          d.Threading.SearchWindow,
			FingerprintInMetadata: d.Threading.FingerprintInMetadata,
		}
	}

	// Convert enrichments configuration if present
	if d.Enrichments != nil {
		cfg.Enrichments = &destslack.SlackEnrichmentsConfig{
			FormatAsBlocks:      d.Enrichments.FormatAsBlocks,
			ColorCoding:         d.Enrichments.ColorCoding,
			TableFormatting:     d.Enrichments.TableFormatting,
			MaxTableRows:        d.Enrichments.MaxTableRows,
			AttachmentThreshold: d.Enrichments.AttachmentThreshold,
		}
	}

	return destslack.NewDestinationSlack(cfg, f.logger, f.httpClient), nil
}
