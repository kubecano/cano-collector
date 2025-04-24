package sender

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/kubecano/cano-collector/pkg/core/reporting"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/slack-go/slack"

	"github.com/kubecano/cano-collector/pkg/logger"
)

const (
	SlackRequestTimeout = 30 * time.Second
	ACTION_LINK         = "link"
)

// SlackBlock represents a block in Slack message
type SlackBlock map[string]interface{}

// SlackMessage represents a message to be sent to Slack
type SlackMessage struct {
	Channel     string
	Text        string
	Blocks      []slack.Block
	Attachments []slack.Attachment
}

// SlackSender sends alerts to a Slack webhook
type SlackSender struct {
	slackClient       *slack.Client
	signingKey        string
	accountID         string
	clusterName       string
	channel           string
	channelNameToID   map[string]string
	verifiedAPITokens sync.Map
	logger            logger.LoggerInterface
}

// NewSlackSender creates a new SlackSender using the Slack SDK and verifies Slack token
func NewSlackSender(slackToken, accountID, clusterName, signingKey, slackChannel string, additionalCertificate string, logger logger.LoggerInterface) (*SlackSender, error) {
	var tlsConfig *tls.Config

	// Configure additional certificate, if provided
	if additionalCertificate != "" {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		cert, err := ioutil.ReadFile(additionalCertificate)
		if err != nil {
			logger.Warnf("Nie udało się wczytać dodatkowego certyfikatu: %v", err)
		} else {
			if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
				logger.Warnf("Nie udało się dodać certyfikatu do puli")
			}

			tlsConfig = &tls.Config{
				RootCAs: rootCAs,
			}
		}
	}

	// Creating a Slack client with optional TLS configuration
	slackClient := slack.New(
		slackToken,
		slack.OptionHTTPClient(&http.Client{
			Timeout: SlackRequestTimeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}),
		slack.OptionLog(logger.GetSlackLogger()),
	)

	// Creating SlackSender instance
	sender := &SlackSender{
		slackClient:     slackClient,
		signingKey:      signingKey,
		accountID:       accountID,
		clusterName:     clusterName,
		channel:         slackChannel,
		channelNameToID: make(map[string]string),
		logger:          logger,
	}

	// Verifying Slack token
	_, alreadyVerified := sender.verifiedAPITokens.Load(slackToken)
	if !alreadyVerified {
		// Auth test
		_, err := slackClient.AuthTest()
		if err != nil {
			return nil, fmt.Errorf("nie można połączyć się ze Slack API: %w", err)
		}

		// Saved verified token
		sender.verifiedAPITokens.Store(slackToken, true)
	}

	return sender, nil
}

// ToSlackLinks converts a slice of LinkProp to SlackBlock
func (s *SlackSender) ToSlackLinks(links []reporting.LinkProp) []SlackBlock {
	if len(links) == 0 {
		return []SlackBlock{}
	}

	buttons := []map[string]interface{}{}

	for i, link := range links {
		button := map[string]interface{}{
			"type": "button",
			"text": map[string]interface{}{
				"type": "plain_text",
				"text": link.Text,
			},
			"action_id": fmt.Sprintf("%s_%d", ACTION_LINK, i),
			"url":       link.URL,
		}

		buttons = append(buttons, button)
	}

	return []SlackBlock{
		{
			"type":     "actions",
			"elements": buttons,
		},
	}
}

// UpdateMessage updates a message in a Slack channel
func (s *SlackSender) UpdateMessage(channel, timestamp, text string, blocks []slack.Block) (string, error) {
	// Check if the channel exists
	channelID, exists := s.channelNameToID[channel]
	if !exists {
		return "", fmt.Errorf("channel ID for %s could not be determined, update aborted", channel)
	}

	_, timestamp, _, err := s.slackClient.UpdateMessage(
		channelID,
		timestamp,
		slack.MsgOptionText(text, false),
		slack.MsgOptionBlocks(blocks...),
	)

	s.logger.Debugf("message updated successfully: %s", timestamp)
	return timestamp, err
}

// Send implements the Sender interface for sending messages to Slack
func (s *SlackSender) Send(message interface{}) error {
	slackMsg, ok := message.(SlackMessage)
	if !ok {
		return fmt.Errorf("wiadomość nie jest poprawną wiadomością Slack")
	}

	channelID := slackMsg.Channel
	if channelID == "" {
		channelID = s.channel
	}

	_, _, err := s.slackClient.PostMessage(
		channelID,
		slack.MsgOptionText(slackMsg.Text, false),
		slack.MsgOptionBlocks(slackMsg.Blocks...),
		slack.MsgOptionAttachments(slackMsg.Attachments...),
	)

	return err
}

// FormatMessage formats the alert details for Slack
func (s *SlackSender) FormatMessage(details reporting.AlertDetails) interface{} {
	var blocks []slack.Block

	// Dodaj nagłówek z tytułem i powagą alertu
	headerText := fmt.Sprintf("*%s*", details.Title)
	if details.Severity != "" {
		headerText = fmt.Sprintf("*[%s]* %s", details.Severity, details.Title)
	}

	headerBlock := slack.NewHeaderBlock(
		slack.NewTextBlockObject("plain_text", headerText, false, false),
	)
	blocks = append(blocks, headerBlock)

	// Dodaj opis
	if details.Description != "" {
		descBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", details.Description, false, false),
			nil, nil,
		)
		blocks = append(blocks, descBlock)
	}

	// Dodaj linki jako przyciski akcji
	if len(details.Links) > 0 {
		actionElements := []slack.BlockElement{}
		for i, link := range details.Links {
			actionElements = append(actionElements,
				slack.NewButtonBlockElement(
					fmt.Sprintf("%s_%d", ACTION_LINK, i),
					link.URL,
					slack.NewTextBlockObject("plain_text", link.Text, false, false),
				),
			)
		}

		blocks = append(blocks, slack.NewActionBlock(
			"links",
			actionElements...,
		))
	}

	// Dodaj metadane jako pola kontekstu
	if len(details.Metadata) > 0 {
		contextElements := []slack.MixedElement{}
		for key, value := range details.Metadata {
			contextElements = append(contextElements,
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*: %s", key, value), false, false),
			)
		}

		blocks = append(blocks, slack.NewContextBlock(
			"metadata",
			contextElements...,
		))
	}

	return SlackMessage{
		Channel: s.channel,
		Text:    details.Title,
		Blocks:  blocks,
	}
}
