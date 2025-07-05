package alert

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	metrics             interfaces.MetricsInterface
}

// NewAlertDispatcher creates a new alert dispatcher
func NewAlertDispatcher(registry interfaces.DestinationRegistryInterface, formatter interfaces.AlertFormatterInterface, logger logger.LoggerInterface, metrics interfaces.MetricsInterface) *AlertDispatcher {
	return &AlertDispatcher{
		destinationRegistry: registry,
		alertFormatter:      formatter,
		logger:              logger,
		metrics:             metrics,
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

	// Convert alert to message format using formatter
	message := d.alertFormatter.FormatAlert(alertEvent)

	// Send to each destination individually to avoid index mismatch issues
	var errors []string
	for _, destName := range team.Destinations {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get individual destination by name
		dest, err := d.destinationRegistry.GetDestination(destName)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to get destination '%s': %v", destName, err)
			errors = append(errors, errorMsg)
			d.logger.Error("Failed to get destination",
				"destination", destName,
				"team", team.Name,
				"error", err)
			d.metrics.IncDestinationErrors(destName, "unknown", "destination_not_found")
			continue
		}

		// Send message to destination with timing
		start := time.Now()
		if err := dest.Send(ctx, message); err != nil {
			duration := time.Since(start)
			errorMsg := fmt.Sprintf("failed to send to destination '%s': %v", destName, err)
			errors = append(errors, errorMsg)
			d.logger.Error("Failed to send alert to destination",
				"destination", destName,
				"team", team.Name,
				"error", err)
			d.metrics.IncDestinationErrors(destName, "unknown", "send_failed") // TODO: Get actual destination type
			d.metrics.ObserveDestinationSendDuration(destName, "unknown", duration)
		} else {
			duration := time.Since(start)
			d.logger.Info("Alert sent successfully",
				"destination", destName,
				"team", team.Name,
				"alert_name", alertEvent.GetAlertName())
			d.metrics.IncDestinationMessagesSent(destName, "unknown", "success") // TODO: Get actual destination type
			d.metrics.ObserveDestinationSendDuration(destName, "unknown", duration)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some destinations: %s", strings.Join(errors, "; "))
	}

	return nil
}
