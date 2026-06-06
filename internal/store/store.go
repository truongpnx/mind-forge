package store

import (
	"context"
	"time"

	"github.com/truongpnx/mind-forge/internal/models"
)

// UserStore persists and retrieves user records.
type UserStore interface {
	// CreateUser inserts a new user and returns the created record.
	CreateUser(ctx context.Context, username, email, passwordHash string) (*models.User, error)
	// GetUserByID returns the user with the given ID.
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	// GetUserByEmail returns the user matching the given email address.
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	// GetUserByUsername returns the user matching the given username.
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	// UpdateUsername changes the username for the given user ID.
	UpdateUsername(ctx context.Context, userID, newUsername string) error
	// UpdatePassword replaces the stored password hash for the given user ID.
	UpdatePassword(ctx context.Context, userID, newPasswordHash string) error
}

// SessionStore manages short-lived session tokens backed by Redis.
type SessionStore interface {
	// SetSession stores a mapping from token → userID with the given TTL.
	SetSession(ctx context.Context, token, userID string, ttl time.Duration) error
	// GetSession returns the userID associated with a token.
	// Returns an empty string and no error when the token does not exist.
	GetSession(ctx context.Context, token string) (userID string, err error)
	// DeleteSession removes a session token.
	DeleteSession(ctx context.Context, token string) error
}
