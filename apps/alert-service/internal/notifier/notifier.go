package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"alert-service/internal/models"
)

// Notifier handles sending alerts to external services
type Notifier struct {
	webhookURL string
}

// NewNotifier creates a new notifier instance
func NewNotifier(webhookURL string) *Notifier {
	return &Notifier{webhookURL: webhookURL}
}

// SlackPayload defines the JSON structure for Slack messages
type SlackPayload struct {
	Text string `json:"text"`
}

// SendAlert sends an alert to the configured notification channel
func (n *Notifier) SendAlert(ctx context.Context, alert *models.Alert) error {
	message := fmt.Sprintf("🚨 *%s Alert* (%s)\n%s",
		alert.Severity, alert.Type, alert.Message)

	if alert.TransactionID != "" {
		message += fmt.Sprintf("\nTransaction: %s", alert.TransactionID)
	}
	if alert.UserID != "" {
		message += fmt.Sprintf("\nUser: %s", alert.UserID)
	}

	return n.sendSlackNotification(ctx, message)
}

// sendSlackNotification posts a message to Slack using the webhook URL
func (n *Notifier) sendSlackNotification(ctx context.Context, message string) error {
	if n.webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	payload := SlackPayload{Text: message}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response from Slack: %s", resp.Status)
	}

	return nil
}
