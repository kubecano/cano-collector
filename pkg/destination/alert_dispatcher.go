package destination

import (
	"github.com/kubecano/cano-collector/pkg/alert"
)

type AlertDispatcher interface {
	Send(alert alert.EnrichedAlert) error
}
