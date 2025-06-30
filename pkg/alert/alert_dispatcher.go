package alert

import (
	"context"
	"fmt"
	"strings"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/destination"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../mocks/alert_dispatcher_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert AlertDispatcherInterface
type AlertDispatcherInterface interface {
	DispatchAlert(ctx context.Context, alert template.Data, team *config_team.Team) error
}

// AlertDispatcher dispatches alerts to team destinations
type AlertDispatcher struct {
	destinationRegistry destination.DestinationRegistryInterface
	logger              logger.LoggerInterface
}

// NewAlertDispatcher creates a new alert dispatcher
func NewAlertDispatcher(registry destination.DestinationRegistryInterface, logger logger.LoggerInterface) *AlertDispatcher {
	return &AlertDispatcher{
		destinationRegistry: registry,
		logger:              logger,
	}
}

// DispatchAlert sends the alert to all destinations of the specified team
func (d *AlertDispatcher) DispatchAlert(ctx context.Context, alert template.Data, team *config_team.Team) error {
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

	// Convert alert to message format
	message := d.formatAlertMessage(alert)

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

// formatAlertMessage converts alertmanager alert to a readable message
func (d *AlertDispatcher) formatAlertMessage(alert template.Data) string {
	var messages []string

	messages = append(messages, fmt.Sprintf("🚨 **Alert: %s**", alert.Status))

	if alert.GroupLabels != nil {
		for key, value := range alert.GroupLabels {
			messages = append(messages, fmt.Sprintf("**%s:** %s", key, value))
		}
	}

	messages = append(messages, "")

	for _, alertItem := range alert.Alerts {
		messages = append(messages, fmt.Sprintf("**Alert:** %s", alertItem.Labels["alertname"]))
		messages = append(messages, fmt.Sprintf("**Status:** %s", alertItem.Status))
		messages = append(messages, fmt.Sprintf("**Severity:** %s", alertItem.Labels["severity"]))

		if alertItem.Annotations["summary"] != "" {
			messages = append(messages, fmt.Sprintf("**Summary:** %s", alertItem.Annotations["summary"]))
		}

		if alertItem.Annotations["description"] != "" {
			messages = append(messages, fmt.Sprintf("**Description:** %s", alertItem.Annotations["description"]))
		}

		messages = append(messages, "")
	}

	return strings.Join(messages, "\n")
}
