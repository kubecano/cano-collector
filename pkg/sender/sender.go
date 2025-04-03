package sender

import (
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/utils"
)

// Alert represents a structured alert to be sent
type Alert struct {
	Title   string
	Message string
}

// DestinationSender defines the interface for sending alerts to various destinations
type DestinationSender interface {
	Send(alert Alert) error
}

// Option defines a function signature for configuring the sender
type Option func(DestinationSender)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client utils.HTTPClient) Option {
	return func(s DestinationSender) {
		if senderWithClient, ok := s.(interface{ SetClient(utils.HTTPClient) }); ok {
			senderWithClient.SetClient(client)
		}
	}
}

// WithLogger sets a custom logger
func WithLogger(log logger.LoggerInterface) Option {
	return func(s DestinationSender) {
		if senderWithLogger, ok := s.(interface{ SetLogger(logger.LoggerInterface) }); ok {
			senderWithLogger.SetLogger(log)
		}
	}
}

// ApplyOptions applies functional options to a sender
func ApplyOptions(sender DestinationSender, opts ...Option) DestinationSender {
	for _, opt := range opts {
		opt(sender)
	}
	return sender
}
