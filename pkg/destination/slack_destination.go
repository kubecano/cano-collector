package destination

import (
	"github.com/kubecano/cano-collector/pkg/core/reporting"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/sender"
)

// SlackDestination implements the Destination interface for sending alerts to Slack
type SlackDestination struct {
	name    string
	sender  *sender.SlackSender
	channel string
	logger  logger.LoggerInterface
}

// NewSlackDestination creates a new SlackDestination instance
func NewSlackDestination(
	name string,
	token string,
	channel string,
	signingKey string,
	accountID string,
	clusterName string,
	logger logger.LoggerInterface,
) (*SlackDestination, error) {
	s, err := sender.NewSlackSender(token, accountID, clusterName, signingKey, channel, "", logger)
	if err != nil {
		return nil, err
	}

	return &SlackDestination{
		name:    name,
		sender:  s,
		channel: channel,
		logger:  logger,
	}, nil
}

// Send sends the alert details to the Slack channel
func (d *SlackDestination) Send(details reporting.AlertDetails) error {
	d.logger.Debugf("Wysyłanie alertu '%s' do kanału Slack %s", details.Title, d.channel)

	// Formats the message using the SlackSender
	message := d.sender.FormatMessage(details)

	// Sends the message using the SlackSender
	return d.sender.Send(message)
}

// Name returns the name of the Slack destination
func (d *SlackDestination) Name() string {
	return d.name
}
