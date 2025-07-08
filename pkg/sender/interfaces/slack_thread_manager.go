package interfaces

import (
	"context"
)

//go:generate mockgen -destination=../../../mocks/slack_thread_manager_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/sender/interfaces SlackThreadManagerInterface

// SlackThreadManagerInterface defines the contract for managing Slack thread relationships
type SlackThreadManagerInterface interface {
	// GetThreadTS returns the thread timestamp for an existing alert with the given fingerprint
	// Returns empty string if no thread exists or if the thread has expired
	GetThreadTS(ctx context.Context, fingerprint string) (string, error)

	// SetThreadTS caches a new thread relationship for the given fingerprint
	SetThreadTS(fingerprint, threadTS string)

	// InvalidateThread removes a thread relationship for the given fingerprint
	InvalidateThread(fingerprint string)

	// Cleanup removes expired thread relationships
	Cleanup()
}
