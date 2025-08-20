package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"storage-service/internal/models"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbURL string) (*Storage, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	s := &Storage{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS processed_transactions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		amount DOUBLE PRECISION NOT NULL,
		currency TEXT NOT NULL,
		type TEXT NOT NULL,
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
		metadata JSONB,
		is_valid BOOLEAN NOT NULL,
		risk_score DOUBLE PRECISION NOT NULL,
		enriched_at TIMESTAMP WITH TIME ZONE NOT NULL,
		failure_reason TEXT
	);`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

func (s *Storage) SaveProcessedTransaction(ctx context.Context, tx *models.ProcessedTransaction) error {
	var metadataJSON []byte
	if tx.Metadata != nil {
		b, err := json.Marshal(tx.Metadata)
		if err != nil {
			return fmt.Errorf("failed to encode metadata: %w", err)
		}
		metadataJSON = b
	}

	query := `
		INSERT INTO processed_transactions (
			id, user_id, amount, currency, type, timestamp, metadata, is_valid, risk_score, enriched_at, failure_reason
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			type = EXCLUDED.type,
			timestamp = EXCLUDED.timestamp,
			metadata = EXCLUDED.metadata,
			is_valid = EXCLUDED.is_valid,
			risk_score = EXCLUDED.risk_score,
			enriched_at = EXCLUDED.enriched_at,
			failure_reason = EXCLUDED.failure_reason;`

	_, err := s.db.ExecContext(ctx, query,
		tx.ID,
		tx.UserID,
		tx.Amount,
		tx.Currency,
		tx.Type,
		tx.Timestamp,
		metadataJSON,
		tx.IsValid,
		tx.RiskScore,
		tx.EnrichedAt,
		tx.FailureReason,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert processed transaction: %w", err)
	}
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
