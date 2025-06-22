package destination

import (
	"github.com/kubecano/cano-collector/pkg/core/issue"
)

type AlertDispatcher interface {
	Send(issue issue.Issue) error
}
