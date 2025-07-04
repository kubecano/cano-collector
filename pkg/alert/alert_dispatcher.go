package alert

import (
	"context"
	"fmt"
	"strings"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/alert/model"
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
func (d *AlertDispatcher) DispatchAlert(ctx context.Context, alertEvent *model.AlertManagerEvent, team *config_team.Team) error {
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
		d.logger.Error("Failed to get destinations for team",
			"team", team.Name,
			"destinations", team.Destinations,
			"error", err)
		return fmt.Errorf("failed to get destinations: %w", err)
	}

	// Convert alert to message format using formatter
	message := d.alertFormatter.FormatAlert(alertEvent)

	// Send to all destinations
	var errors []string
	for i, dest := range destinations {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		destName := team.Destinations[i] // Get the destination name from the team config
		if err := dest.Send(ctx, message); err != nil {
			errorMsg := fmt.Sprintf("failed to send to destination: %v", err)
			errors = append(errors, errorMsg)
			d.logger.Error("Failed to send alert to destination",
				"destination", destName,
				"team", team.Name,
				"error", err)
		} else {
			d.logger.Info("Alert sent successfully",
				"destination", destName,
				"team", team.Name,
				"alert_name", alertEvent.GetAlertName())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some destinations: %s", strings.Join(errors, "; "))
	}

	return nil
}
