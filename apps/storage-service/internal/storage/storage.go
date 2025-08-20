package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"storage-service/internal/models"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// Storage handles database operations and caching
type Storage struct {
	db    *sql.DB
	redis *redis.Client
}

// NewStorage creates a new storage instance
func NewStorage(dbURL string) (*Storage, error) {
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize Redis client (optional, for caching)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis not available, caching disabled: %v", err)
		redisClient = nil
	}

	storage := &Storage{
		db:    db,
		redis: redisClient,
	}

	// Initialize database schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary tables and indexes
func (s *Storage) initSchema() error {
	log.Println("Initializing database schema...")

	// Create tables
	for _, sql := range models.CreateTablesSQL() {
		if _, err := s.db.Exec(sql); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create indexes
	for _, sql := range models.CreateIndexesSQL() {
		if _, err := s.db.Exec(sql); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// StoreTransaction stores a processed transaction in the database
func (s *Storage) StoreTransaction(ctx context.Context, txn *models.StoredTransaction) error {
	start := time.Now()

	// Check if transaction already exists (idempotency)
	exists, err := s.transactionExists(ctx, txn.ID)
	if err != nil {
		return fmt.Errorf("failed to check transaction existence: %w", err)
	}

	if exists {
		log.Printf("Transaction %s already exists, skipping", txn.ID)
		return nil
	}

	// Prepare the SQL statement
	query := `
		INSERT INTO transactions (
			id, idempotency_key, account_id, user_id, amount, currency, type, category,
			merchant, reference, status, timestamp, metadata, risk_score, risk_level,
			is_approved, rejection_reason, is_valid, validation_errors, country,
			ip_address, device_info, processed_at, processing_time, processor_id,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27
		)
	`

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(txn.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Convert validation errors to array
	var validationErrors []string
	if txn.ValidationErrors != nil {
		validationErrors = txn.ValidationErrors
	}

	// Execute the insert
	_, err = s.db.ExecContext(ctx, query,
		txn.ID, txn.IdempotencyKey, txn.AccountID, txn.UserID, txn.Amount,
		txn.Currency, txn.Type, txn.Category, txn.Merchant, txn.Reference,
		txn.Status, txn.Timestamp, metadataJSON, txn.RiskScore, txn.RiskLevel,
		txn.IsApproved, txn.RejectionReason, txn.IsValid, validationErrors,
		txn.Country, txn.IPAddress, txn.DeviceInfo, txn.ProcessedAt,
		txn.ProcessingTime, txn.ProcessorID, time.Now(), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	// Update risk metrics
	if err := s.updateRiskMetrics(ctx, txn); err != nil {
		log.Printf("Warning: failed to update risk metrics: %v", err)
	}

	// Cache the transaction
	if s.redis != nil {
		s.cacheTransaction(ctx, txn)
	}

	log.Printf("Transaction %s stored successfully in %v", txn.ID, time.Since(start))
	return nil
}

// transactionExists checks if a transaction already exists
func (s *Storage) transactionExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE id = $1)`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

// updateRiskMetrics updates the risk metrics for an account
func (s *Storage) updateRiskMetrics(ctx context.Context, txn *models.StoredTransaction) error {
	query := `
		INSERT INTO risk_metrics (account_id, risk_score, risk_level, total_flagged, total_rejected, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (account_id) DO UPDATE SET
			risk_score = CASE 
				WHEN risk_metrics.risk_score < $2 THEN $2 
				ELSE risk_metrics.risk_score 
			END,
			risk_level = CASE 
				WHEN $2 > 0.7 THEN 'high'
				WHEN $2 > 0.4 THEN 'medium'
				ELSE 'low'
			END,
			total_flagged = risk_metrics.total_flagged + CASE WHEN $6 = 'flagged' THEN 1 ELSE 0 END,
			total_rejected = risk_metrics.total_rejected + CASE WHEN $6 = 'rejected' THEN 1 ELSE 0 END,
			last_updated = $6
	`

	var flaggedCount, rejectedCount int64
	if txn.Status == models.StatusFlagged {
		flaggedCount = 1
	}
	if txn.Status == models.StatusRejected {
		rejectedCount = 1
	}

	_, err := s.db.ExecContext(ctx, query,
		txn.AccountID, txn.RiskScore, txn.RiskLevel, flaggedCount, rejectedCount, time.Now(),
	)

	return err
}

// cacheTransaction caches a transaction in Redis
func (s *Storage) cacheTransaction(ctx context.Context, txn *models.StoredTransaction) {
	if s.redis == nil {
		return
	}

	key := fmt.Sprintf("txn:%s", txn.ID)
	data, err := json.Marshal(txn)
	if err != nil {
		log.Printf("Failed to marshal transaction for caching: %v", err)
		return
	}

	// Cache for 1 hour
	err = s.redis.Set(ctx, key, data, time.Hour).Err()
	if err != nil {
		log.Printf("Failed to cache transaction: %v", err)
	}
}

// GetTransaction retrieves a transaction by ID
func (s *Storage) GetTransaction(ctx context.Context, id string) (*models.StoredTransaction, error) {
	// Try cache first
	if s.redis != nil {
		if cached, err := s.getCachedTransaction(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}

	// Query database
	query := `SELECT * FROM transactions WHERE id = $1`
	row := s.db.QueryRowContext(ctx, query, id)

	var txn models.StoredTransaction
	var metadataJSON []byte
	var validationErrors []string

	err := row.Scan(
		&txn.ID, &txn.IdempotencyKey, &txn.AccountID, &txn.UserID, &txn.Amount,
		&txn.Currency, &txn.Type, &txn.Category, &txn.Merchant, &txn.Reference,
		&txn.Status, &txn.Timestamp, &metadataJSON, &txn.RiskScore, &txn.RiskLevel,
		&txn.IsApproved, &txn.RejectionReason, &txn.IsValid, &validationErrors,
		&txn.Country, &txn.IPAddress, &txn.DeviceInfo, &txn.ProcessedAt,
		&txn.ProcessingTime, &txn.ProcessorID, &txn.CreatedAt, &txn.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan transaction: %w", err)
	}

	// Parse metadata JSON
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &txn.Metadata); err != nil {
			log.Printf("Warning: failed to unmarshal metadata: %v", err)
		}
	}

	txn.ValidationErrors = validationErrors

	// Cache the result
	if s.redis != nil {
		s.cacheTransaction(ctx, &txn)
	}

	return &txn, nil
}

// getCachedTransaction retrieves a transaction from Redis cache
func (s *Storage) getCachedTransaction(ctx context.Context, id string) (*models.StoredTransaction, error) {
	key := fmt.Sprintf("txn:%s", id)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var txn models.StoredTransaction
	if err := json.Unmarshal(data, &txn); err != nil {
		return nil, err
	}

	return &txn, nil
}

// GetTransactionsByAccount retrieves transactions for a specific account
func (s *Storage) GetTransactionsByAccount(ctx context.Context, accountID string, limit, offset int) ([]*models.StoredTransaction, error) {
	query := `
		SELECT * FROM transactions 
		WHERE account_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.StoredTransaction
	for rows.Next() {
		var txn models.StoredTransaction
		var metadataJSON []byte
		var validationErrors []string

		err := rows.Scan(
			&txn.ID, &txn.IdempotencyKey, &txn.AccountID, &txn.UserID, &txn.Amount,
			&txn.Currency, &txn.Type, &txn.Category, &txn.Merchant, &txn.Reference,
			&txn.Status, &txn.Timestamp, &metadataJSON, &txn.RiskScore, &txn.RiskLevel,
			&txn.IsApproved, &txn.RejectionReason, &txn.IsValid, &validationErrors,
			&txn.Country, &txn.IPAddress, &txn.DeviceInfo, &txn.ProcessedAt,
			&txn.ProcessingTime, &txn.ProcessorID, &txn.CreatedAt, &txn.UpdatedAt,
		)

		if err != nil {
			log.Printf("Failed to scan transaction row: %v", err)
			continue
		}

		// Parse metadata JSON
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &txn.Metadata); err != nil {
				log.Printf("Warning: failed to unmarshal metadata: %v", err)
			}
		}

		txn.ValidationErrors = validationErrors
		transactions = append(transactions, &txn)
	}

	return transactions, nil
}

// GetTransactionSummary returns a summary of transactions for an account
func (s *Storage) GetTransactionSummary(ctx context.Context, accountID string) (*models.TransactionSummary, error) {
	query := `
		SELECT 
			account_id,
			COUNT(*) as total_transactions,
			SUM(amount) as total_amount,
			AVG(amount) as average_amount,
			MAX(timestamp) as last_transaction,
			MAX(risk_level) as risk_level
		FROM transactions 
		WHERE account_id = $1
		GROUP BY account_id
	`

	var summary models.TransactionSummary
	err := s.db.QueryRowContext(ctx, query, accountID).Scan(
		&summary.AccountID, &summary.TotalTransactions, &summary.TotalAmount,
		&summary.AverageAmount, &summary.LastTransaction, &summary.RiskLevel,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	return &summary, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.redis != nil {
		s.redis.Close()
	}
	return s.db.Close()
}
