package handler

import (
	"context"
	"encoding/json"

	"storage-service/internal/models"
	"storage-service/internal/storage"
)

type TransactionHandler struct {
	store *storage.Storage
}

func NewTransactionHandler(store *storage.Storage) *TransactionHandler {
	return &TransactionHandler{store: store}
}

// Handle satisfies consumer.Handler by decoding a processed transaction and persisting it
func (h *TransactionHandler) Handle(ctx context.Context, message []byte) error {
	var tx models.ProcessedTransaction
	if err := json.Unmarshal(message, &tx); err != nil {
		return err
	}
	return h.store.SaveProcessedTransaction(ctx, &tx)
}
