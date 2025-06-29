package sender

import (
	"context"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

//go:generate mockgen -destination=../../mocks/sender_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/sender SenderInterface
type DestinationSenderInterface interface {
	Send(ctx context.Context, message string) error
}

// Option defines a function signature for configuring the sender
type Option func(DestinationSenderInterface)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client util.HTTPClient) Option {
	return func(s DestinationSenderInterface) {
		if senderWithClient, ok := s.(interface{ SetClient(util.HTTPClient) }); ok {
			senderWithClient.SetClient(client)
		}
	}
}

// WithLogger sets a custom logger
func WithLogger(log logger.LoggerInterface) Option {
	return func(s DestinationSenderInterface) {
		if senderWithLogger, ok := s.(interface{ SetLogger(logger.LoggerInterface) }); ok {
			senderWithLogger.SetLogger(log)
		}
	}
}

// ApplyOptions applies functional options to a sender
func ApplyOptions(sender DestinationSenderInterface, opts ...Option) DestinationSenderInterface {
	for _, opt := range opts {
		opt(sender)
	}
	return sender
}
