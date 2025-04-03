package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kubecano/cano-collector/pkg/utils"

	"github.com/kubecano/cano-collector/pkg/logger"
)

// SlackSender sends alerts to a Slack webhook
type SlackSender struct {
	WebhookURL string
	httpClient utils.HTTPClient
	logger     logger.LoggerInterface
}

// NewSlackSender creates a new SlackSender
func NewSlackSender(webhookURL string, logger logger.LoggerInterface, opts ...Option) *SlackSender {
	sender := &SlackSender{
		WebhookURL: webhookURL,
		httpClient: utils.DefaultHTTPClient(),
		logger:     logger,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(sender)
	}

	return sender
}

// SetClient allows setting a custom HTTP client
func (s *SlackSender) SetClient(client utils.HTTPClient) {
	s.httpClient = client
}

// Send sends an alert to Slack
func (s *SlackSender) Send(alert Alert) error {
	payload := map[string]string{
		"text": fmt.Sprintf("*%s*\n%s", alert.Title, alert.Message),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send alert to Slack: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			s.logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned non-OK status: %d", resp.StatusCode)
	}

	s.logger.Infof("Successfully sent alert to Slack: %s", alert.Title)
	return nil
}
