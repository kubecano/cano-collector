package sender

import "github.com/kubecano/cano-collector/pkg/alerts"

type DestinationSender interface {
	Send(alert alerts.EnrichedAlert) error
}
