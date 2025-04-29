package destination

import (
	"github.com/kubecano/cano-collector/pkg/core/reporting"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/sender"
)

// TeamsDestination implements the Destination interface for sending alerts to Slack
type TeamsDestination struct {
	name       string
	sender     *sender.TeamsSender
	webhookURL string
	logger     logger.LoggerInterface
}

// NewTeamsDestination creates a new SlackDestination instance
func NewTeamsDestination(
	name string,
	webhookURL string,
	logger logger.LoggerInterface,
) (*TeamsDestination, error) {
	s, err := sender.NewTeamsSender(webhookURL, logger)
	if err != nil {
		return nil, err
	}

	return &TeamsDestination{
		name:       name,
		sender:     s,
		webhookURL: webhookURL,
		logger:     logger,
	}, nil
}

// Send sends the alert details to the Microsoft Teams channel
func (d *TeamsDestination) Send(details reporting.AlertDetails) error {
	d.logger.Debugf("Wysyłanie alertu '%s' do kanału MS Tems", details.Title)

	// Formats the message using the SlackSender
	message := d.sender.FormatMessage(details)

	// Sends the message using the SlackSender
	return d.sender.Send(message)
}

// Name returns the name of the Microsoft Teams destination
func (d *TeamsDestination) Name() string {
	return d.name
}
