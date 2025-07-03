package alert

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/interfaces"
	"github.com/kubecano/cano-collector/pkg/logger"
)

// AlertDispatcher dispatches alerts to team destinations
type AlertDispatcher struct {
	destinationRegistry interfaces.DestinationRegistryInterface
	alertFormatter      interfaces.AlertFormatterInterface
	logger              logger.LoggerInterface
}

// NewAlertDispatcher creates a new alert dispatcher
func NewAlertDispatcher(registry interfaces.DestinationRegistryInterface, formatter interfaces.AlertFormatterInterface, logger logger.LoggerInterface) *AlertDispatcher {
	return &AlertDispatcher{
		destinationRegistry: registry,
		alertFormatter:      formatter,
		logger:              logger,
	}
}

// DispatchAlert sends the alert to all destinations of the specified team
func (d *AlertDispatcher) DispatchAlert(ctx context.Context, alert interface{}, team *config_team.Team) error {
	// Type assertion - in practice, this should always be *AlertManagerEvent
	alertEvent, ok := alert.(*AlertManagerEvent)
	if !ok {
		d.logger.Error("Invalid alert type passed to DispatchAlert")
		return fmt.Errorf("invalid alert type")
	}

	if team == nil {
		d.logger.Info("No team resolved for alert, skipping dispatch")
		return nil
	}

	if len(team.Destinations) == 0 {
		d.logger.Info("Team has no destinations configured", "team", team.Name)
		return nil
	}

	// Get destinations for the team
	destinations, err := d.destinationRegistry.GetDestinations(team.Destinations)
	if err != nil {
		return fmt.Errorf("failed to get destinations for team '%s': %w", team.Name, err)
	}

	// Convert alert to message format using formatter
	message := d.alertFormatter.FormatAlert(alertEvent)

	// Send to all destinations
	var errors []string
	for _, dest := range destinations {
		if err := dest.Send(ctx, message); err != nil {
			errorMsg := fmt.Sprintf("failed to send to destination: %v", err)
			errors = append(errors, errorMsg)
			d.logger.Error("Failed to send alert to destination",
				zap.Error(err),
				zap.String("team", team.Name))
		} else {
			d.logger.Info("Alert sent successfully to destination",
				zap.String("team", team.Name))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some destinations failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
