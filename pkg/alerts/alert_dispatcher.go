package alerts

type AlertDispatcher interface {
	Send(alert EnrichedAlert) error
}
