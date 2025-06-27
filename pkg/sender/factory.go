package sender

import (
	"fmt"

	"github.com/kubecano/cano-collector/pkg/util"

	"github.com/kubecano/cano-collector/config/destination"
	"github.com/kubecano/cano-collector/pkg/logger"
)

// SenderFactory creates appropriate DestinationSender based on destination type
type SenderFactory struct {
	logger logger.LoggerInterface
	client util.HTTPClient
}

// NewSenderFactory creates a new SenderFactory
func NewSenderFactory(logger logger.LoggerInterface, client util.HTTPClient) *SenderFactory {
	if client == nil {
		client = util.DefaultHTTPClient()
	}
	return &SenderFactory{
		logger: logger,
		client: client,
	}
}

// CreateSender creates appropriate DestinationSender based on destination configuration
func (f *SenderFactory) CreateSender(dest interface{}, opts ...Option) (DestinationSender, error) {
	var sender DestinationSender

	switch d := dest.(type) {
	case destination.SlackDestination:
		if d.APIKey == "" {
			return nil, fmt.Errorf("slack destination '%s' must have api_key", d.Name)
		}
		sender = NewSlackSenderWithAPIKey(d.APIKey, d.SlackChannel, f.logger)

	case destination.TeamsDestination:
		if d.WebhookURL == "" {
			return nil, fmt.Errorf("teams destination '%s' must have webhookURL", d.Name)
		}
		sender = NewMSTeamsSender(d.WebhookURL, f.logger)

	default:
		return nil, fmt.Errorf("unsupported destination type: %T", dest)
	}

	// Apply default logger and client if not overridden
	sender = ApplyOptions(sender,
		WithLogger(f.logger),
		WithHTTPClient(f.client),
	)

	// Apply additional options
	sender = ApplyOptions(sender, opts...)

	return sender, nil
}
