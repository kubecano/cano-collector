package sender

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kubecano/cano-collector/pkg/core/reporting"

	"github.com/slack-go/slack"

	"github.com/kubecano/cano-collector/pkg/logger"
)

const (
	SlackRequestTimeout = 30 * time.Second
	ACTION_LINK         = "link"
)

// SlackClientInterface definiuje interfejs dla klienta Slack
//
//go:generate mockgen -destination=../../mocks/slack_client_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/sender SlackClientInterface
type SlackClientInterface interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	AuthTest() (*slack.AuthTestResponse, error)
	UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
}

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
	slackClient       SlackClientInterface
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

		cert, err := os.ReadFile(additionalCertificate)
		if err != nil {
			logger.Warnf("Cannot read additional certificate: %v", err)
		} else {
			if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
				logger.Warnf("Cannot add additional certificate to the pool")
			}

			tlsConfig = &tls.Config{
				RootCAs: rootCAs,
			}
		}
	}

	// Preparing Slack client options
	slackOptions := []slack.Option{
		slack.OptionHTTPClient(&http.Client{
			Timeout: SlackRequestTimeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}),
	}

	// Only add logger if provided
	if logger != nil {
		slackLogger := logger.GetSlackLogger()
		if slackLogger != nil {
			slackOptions = append(slackOptions, slack.OptionLog(slackLogger))
		}
	}

	// Creating a Slack client with optional TLS configuration
	slackClient := slack.New(slackToken, slackOptions...)

	// Creating SlackSender instance
	sender, err := createSlackSender(slackToken, accountID, clusterName, signingKey, slackChannel, slackClient, logger)
	if err != nil {
		return nil, err
	}

	return sender, nil
}

// createSlackSender creates a new SlackSender instance
func createSlackSender(
	slackToken string,
	accountID string,
	clusterName string,
	signingKey string,
	slackChannel string,
	slackClient SlackClientInterface,
	logger logger.LoggerInterface,
) (*SlackSender, error) {
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
			return nil, fmt.Errorf("cannot connect with Slack API: %w", err)
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

// Send implements the SenderInterface interface for sending messages to Slack
func (s *SlackSender) Send(message interface{}) error {
	slackMsg, ok := message.(SlackMessage)
	if !ok {
		return fmt.Errorf("wiadomość nie jest poprawną wiadomością Slack")
	}

	channelID := slackMsg.Channel
	if channelID == "" {
		channelID = s.channel
	}

	// Pobierz ID kanału, jeśli podano nazwę
	if id, exists := s.channelNameToID[channelID]; exists {
		channelID = id
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

	// Przygotuj tekst nagłówka
	headerText := details.Title
	if details.Severity != "" {
		headerText = fmt.Sprintf("[%s] %s", details.Severity, details.Title)
	}

	// Dodaj nagłówek
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
