package sender

import (
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	sender_interfaces "github.com/kubecano/cano-collector/pkg/sender/interfaces"
	"github.com/kubecano/cano-collector/pkg/util"
)

// Option defines a function signature for configuring the sender
type Option func(sender_interfaces.DestinationSenderInterface)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client util.HTTPClient) Option {
	return func(s sender_interfaces.DestinationSenderInterface) {
		if senderWithClient, ok := s.(interface{ SetClient(util.HTTPClient) }); ok {
			senderWithClient.SetClient(client)
		}
	}
}

// WithLogger sets a custom logger
func WithLogger(log logger_interfaces.LoggerInterface) Option {
	return func(s sender_interfaces.DestinationSenderInterface) {
		if senderWithLogger, ok := s.(interface {
			SetLogger(logger_interfaces.LoggerInterface)
		}); ok {
			senderWithLogger.SetLogger(log)
		}
	}
}

// ApplyOptions applies functional options to a sender
func ApplyOptions(sender sender_interfaces.DestinationSenderInterface, opts ...Option) sender_interfaces.DestinationSenderInterface {
	for _, opt := range opts {
		opt(sender)
	}
	return sender
}
