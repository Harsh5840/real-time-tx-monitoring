package handler

import (
	"context"
	"encoding/json"
	"log"

	"alert-service/internal/models"
	"alert-service/internal/notifier"
)

type AlertHandler struct {
	notifier *notifier.Notifier
}

func NewAlertHandler(webhookURL string) *AlertHandler {
	return &AlertHandler{
		notifier: notifier.NewNotifier(webhookURL),
	}
}

// Handle satisfies consumer.Handler by decoding an alert and sending it via notifier
func (h *AlertHandler) Handle(ctx context.Context, message []byte) error {
	var alert models.Alert
	if err := json.Unmarshal(message, &alert); err != nil {
		return err
	}

	log.Printf("processing alert %s: %s", alert.ID, alert.Message)
	return h.notifier.SendAlert(ctx, &alert)
}
