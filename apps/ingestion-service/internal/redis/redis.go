package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client
func NewClient(addr, password string, db int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// SetIdempotencyKey sets an idempotency key with TTL
func (c *Client) SetIdempotencyKey(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.rdb.Set(ctx, fmt.Sprintf("idempotency:%s", key), data, ttl).Err()
}

// GetIdempotencyKey retrieves an idempotency key
func (c *Client) GetIdempotencyKey(ctx context.Context, key string) ([]byte, error) {
	data, err := c.rdb.Get(ctx, fmt.Sprintf("idempotency:%s", key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key not found
		}
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	return data, nil
}

// SetAccountBalance sets account balance cache
func (c *Client) SetAccountBalance(ctx context.Context, accountID string, balance float64, ttl time.Duration) error {
	return c.rdb.Set(ctx, fmt.Sprintf("balance:%s", accountID), balance, ttl).Err()
}

// GetAccountBalance retrieves account balance from cache
func (c *Client) GetAccountBalance(ctx context.Context, accountID string) (float64, error) {
	balance, err := c.rdb.Get(ctx, fmt.Sprintf("balance:%s", accountID)).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // No cached balance
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.rdb.Close()
}
