package destinations

import "github.com/kubecano/cano-collector/pkg/alerts"

type AlertDispatcher interface {
	Send(alert alerts.EnrichedAlert) error
}
