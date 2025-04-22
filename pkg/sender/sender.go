package sender

import (
	"github.com/kubecano/cano-collector/pkg/core/reporting"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

// Sender defines the interface for sending alerts to different platforms
type Sender interface {
	// Send sends formatted alert details to the destination
	Send(message interface{}) error
	// FormatMessage formats the alert details for the destination
	FormatMessage(details reporting.AlertDetails) interface{}
}

// Option defines a functional option type for configuring the sender
type Option func(interface{})

// WithHTTPClient sets a custom HTTP client for the sender
func WithHTTPClient(client util.HTTPClient) Option {
	return func(s interface{}) {
		if senderWithClient, ok := s.(interface{ SetClient(util.HTTPClient) }); ok {
			senderWithClient.SetClient(client)
		}
	}
}

// WithLogger sets a custom logger for the sender
func WithLogger(log logger.LoggerInterface) Option {
	return func(s interface{}) {
		if senderWithLogger, ok := s.(interface{ SetLogger(logger.LoggerInterface) }); ok {
			senderWithLogger.SetLogger(log)
		}
	}
}

// ApplyOptions sets the provided options on the target sender
func ApplyOptions(target interface{}, opts ...Option) interface{} {
	for _, opt := range opts {
		opt(target)
	}
	return target
}
