package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kubecano/cano-collector/pkg/core/reporting"
	"github.com/kubecano/cano-collector/pkg/util"

	"github.com/kubecano/cano-collector/pkg/logger"
)

// TeamsSender sends alerts to a Microsoft Teams webhook
type TeamsSender struct {
	WebhookURL string
	httpClient util.HTTPClient
	logger     logger.LoggerInterface
}

// NewTeamsSender creates a new TeamsSender with functional options
func NewTeamsSender(webhookURL string, logger logger.LoggerInterface, opts ...Option) (*TeamsSender, error) {
	sender := &TeamsSender{
		WebhookURL: webhookURL,
		httpClient: util.DefaultHTTPClient(), // Default client
		logger:     logger,                   // Default logger
	}

	// Apply functional options
	for _, opt := range opts {
		opt(sender)
	}

	return sender, nil
}

// SetClient allows setting a custom HTTP client
func (s *TeamsSender) SetClient(client util.HTTPClient) {
	s.httpClient = client
}

func (s *TeamsSender) FormatMessage(details reporting.AlertDetails) interface{} {
	// Format the message as needed for Microsoft Teams
	return details
}

// Send sends an alert to Microsoft Teams
func (s *TeamsSender) Send(message interface{}) error {
	teamsMsg := message.(reporting.AlertDetails)

	payload := map[string]string{
		"text": fmt.Sprintf("**%s**\n%s", teamsMsg.Title, teamsMsg.Description),
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
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			s.logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send to Microsoft Teams: non-OK status %d", resp.StatusCode)
	}

	s.logger.Infof("Successfully sent alert to MS Teams: %s", teamsMsg.Title)
	return nil
}
