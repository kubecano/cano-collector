package sender

import (
	"context"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

// Alert represents a structured alert to be sent
type Alert struct {
	Title   string
	Message string
}

// DestinationSender defines the interface for sending notifications to various destinations
type DestinationSender interface {
	Send(ctx context.Context, message string) error
}

// Option defines a function signature for configuring the sender
type Option func(DestinationSender)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client util.HTTPClient) Option {
	return func(s DestinationSender) {
		if senderWithClient, ok := s.(interface{ SetClient(util.HTTPClient) }); ok {
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
