package alert

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	destination_interfaces "github.com/kubecano/cano-collector/pkg/destination/interfaces"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
)

// AlertDispatcher dispatches issues to team destinations
type AlertDispatcher struct {
	destinationRegistry destination_interfaces.DestinationRegistryInterface
	logger              logger_interfaces.LoggerInterface
	metrics             metric_interfaces.MetricsInterface
}

// NewAlertDispatcher creates a new alert dispatcher
func NewAlertDispatcher(registry destination_interfaces.DestinationRegistryInterface, logger logger_interfaces.LoggerInterface, metrics metric_interfaces.MetricsInterface) *AlertDispatcher {
	return &AlertDispatcher{
		destinationRegistry: registry,
		logger:              logger,
		metrics:             metrics,
	}
}

// DispatchIssues sends the issues to all destinations of the specified team
func (d *AlertDispatcher) DispatchIssues(ctx context.Context, issues []*issue.Issue, team *config_team.Team) error {
	if team == nil {
		d.logger.Info("No team resolved for issues, skipping dispatch")
		return nil
	}

	if len(team.Destinations) == 0 {
		d.logger.Info("Team has no destinations configured",
			zap.String("team", team.Name),
		)
		return nil
	}

	if len(issues) == 0 {
		d.logger.Info("No issues to dispatch for team",
			zap.String("team", team.Name),
		)
		return nil
	}

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
				zap.String("destination", destName),
				zap.String("team", team.Name),
				zap.Error(err),
			)
			d.metrics.IncDestinationErrors(destName, "unknown", "destination_not_found")
			continue
		}

		// Send each issue to destination with timing
		for _, iss := range issues {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			start := time.Now()
			if err := dest.Send(ctx, iss); err != nil {
				duration := time.Since(start)
				errorMsg := fmt.Sprintf("failed to send issue to destination '%s': %v", destName, err)
				errors = append(errors, errorMsg)
				d.logger.Error("Failed to send issue",
					zap.String("issue", iss.Title),
					zap.String("destination", destName),
					zap.String("team", team.Name),
					zap.Error(err),
				)
				d.metrics.IncDestinationErrors(destName, "unknown", "send_failed") // TODO: Get actual destination type
				d.metrics.ObserveDestinationSendDuration(destName, "unknown", duration)
			} else {
				duration := time.Since(start)
				d.logger.Info("Issue sent successfully",
					zap.String("issue", iss.Title),
					zap.String("destination", destName),
					zap.String("team", team.Name),
					zap.String("severity", iss.Severity.String()),
				)
				d.metrics.IncDestinationMessagesSent(destName, "unknown", "success") // TODO: Get actual destination type
				d.metrics.ObserveDestinationSendDuration(destName, "unknown", duration)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some destinations: %s", strings.Join(errors, "; "))
	}

	return nil
}
