package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/utils"
)

// MSTeamsSender sends alerts to a Microsoft Teams webhook
type MSTeamsSender struct {
	WebhookURL string
	httpClient utils.HTTPClient
	logger     logger.LoggerInterface
}

// NewMSTeamsSender creates a new MSTeamsSender with functional options
func NewMSTeamsSender(webhookURL string, logger logger.LoggerInterface, opts ...Option) *MSTeamsSender {
	sender := &MSTeamsSender{
		WebhookURL: webhookURL,
		httpClient: utils.DefaultHTTPClient(), // Default client
		logger:     logger,                    // Default logger
	}

	// Apply functional options
	for _, opt := range opts {
		opt(sender)
	}

	return sender
}

// SetClient allows setting a custom HTTP client
func (s *MSTeamsSender) SetClient(client utils.HTTPClient) {
	s.httpClient = client
}

// Send sends an alert to Microsoft Teams
func (s *MSTeamsSender) Send(alert Alert) error {
	payload := map[string]string{
		"text": fmt.Sprintf("**%s**\n%s", alert.Title, alert.Message),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal MS Teams message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send alert to MS Teams: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			s.logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MS Teams returned non-OK status: %d", resp.StatusCode)
	}

	s.logger.Infof("Successfully sent alert to MS Teams: %s", alert.Title)
	return nil
}
