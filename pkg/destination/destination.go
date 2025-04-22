package destination

import (
	"github.com/kubecano/cano-collector/pkg/core/reporting"
)

// Destination defines the interface for sending alerts to different platforms
type Destination interface {
	// Send sends the alert details to the destination
	Send(details reporting.AlertDetails) error
	// Name returns the name of the destination
	Name() string
}
