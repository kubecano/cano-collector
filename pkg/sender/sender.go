package sender

import (
	"github.com/kubecano/cano-collector/pkg/interfaces"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

// Option defines a function signature for configuring the sender
type Option func(interfaces.DestinationSenderInterface)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client util.HTTPClient) Option {
	return func(s interfaces.DestinationSenderInterface) {
		if senderWithClient, ok := s.(interface{ SetClient(util.HTTPClient) }); ok {
			senderWithClient.SetClient(client)
		}
	}
}

// WithLogger sets a custom logger
func WithLogger(log logger.LoggerInterface) Option {
	return func(s interfaces.DestinationSenderInterface) {
		if senderWithLogger, ok := s.(interface{ SetLogger(logger.LoggerInterface) }); ok {
			senderWithLogger.SetLogger(log)
		}
	}
}

// ApplyOptions applies functional options to a sender
func ApplyOptions(sender interfaces.DestinationSenderInterface, opts ...Option) interfaces.DestinationSenderInterface {
	for _, opt := range opts {
		opt(sender)
	}
	return sender
}
