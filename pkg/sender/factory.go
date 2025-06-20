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

// Create creates a DestinationSender based on destination type
func (f *SenderFactory) Create(destination destination.Destination, opts ...Option) (DestinationSender, error) {
	var sender DestinationSender

	switch destination.Name {
	case "slack":
		sender = NewSlackSender(destination.WebhookURL, f.logger)
	case "teams":
		sender = NewMSTeamsSender(destination.WebhookURL, f.logger)
	default:
		return nil, fmt.Errorf("unsupported destination type: %s", destination.Name)
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
