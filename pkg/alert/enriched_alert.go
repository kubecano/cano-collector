package alert

import (
	"github.com/prometheus/alertmanager/template"
)

// EnrichedAlert represents the base alert structure that can be extended later
type EnrichedAlert struct {
	Original template.Data `json:"original"`
	// Additional fields like logs, events can be added later
}
