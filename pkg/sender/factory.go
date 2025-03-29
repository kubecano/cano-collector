package sender

import "github.com/kubecano/cano-collector/config/destinations"

type SenderFactory interface {
	Create(destination destinations.Destination) (DestinationSender, error)
}
