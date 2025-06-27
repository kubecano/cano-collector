package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

// MSTeamsSender sends alerts to a Microsoft Teams webhook
type MSTeamsSender struct {
	WebhookURL string
	httpClient util.HTTPClient
	logger     logger.LoggerInterface
}

// NewMSTeamsSender creates a new MSTeamsSender with functional options
func NewMSTeamsSender(webhookURL string, logger logger.LoggerInterface, opts ...Option) *MSTeamsSender {
	sender := &MSTeamsSender{
		WebhookURL: webhookURL,
		httpClient: util.DefaultHTTPClient(), // Default client
		logger:     logger,                   // Default logger
	}

	// Apply functional options
	for _, opt := range opts {
		opt(sender)
	}

	return sender
}

// SetClient allows setting a custom HTTP client
func (s *MSTeamsSender) SetClient(client util.HTTPClient) {
	s.httpClient = client
}

// Send sends a message to Microsoft Teams
func (s *MSTeamsSender) Send(ctx context.Context, message string) error {
	payload := map[string]string{
		"text": message,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal MS Teams message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message to MS Teams: %w", err)
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			s.logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send to Microsoft Teams: non-OK status %d", resp.StatusCode)
	}

	s.logger.Infof("Successfully sent message to MS Teams")
	return nil
}
