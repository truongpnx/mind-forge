package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "session:"

// SessionStore implements store.SessionStore using Redis.
type SessionStore struct {
	client *redis.Client
}

// NewSessionStore returns a SessionStore backed by the given Redis client.
func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{client: client}
}

// SetSession stores token → userID with the given TTL.
func (s *SessionStore) SetSession(ctx context.Context, token, userID string, ttl time.Duration) error {
	if err := s.client.Set(ctx, keyPrefix+token, userID, ttl).Err(); err != nil {
		return fmt.Errorf("redis: SetSession: %w", err)
	}
	return nil
}

// GetSession returns the userID for the token.
// Returns ("", nil) when the key does not exist.
func (s *SessionStore) GetSession(ctx context.Context, token string) (string, error) {
	val, err := s.client.Get(ctx, keyPrefix+token).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("redis: GetSession: %w", err)
	}
	return val, nil
}

// DeleteSession removes the session token.
func (s *SessionStore) DeleteSession(ctx context.Context, token string) error {
	if err := s.client.Del(ctx, keyPrefix+token).Err(); err != nil {
		return fmt.Errorf("redis: DeleteSession: %w", err)
	}
	return nil
}
