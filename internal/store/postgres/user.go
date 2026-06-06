package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/truongpnx/mind-forge/internal/models"
)

// UserStore implements store.UserStore using a pgxpool connection pool.
type UserStore struct {
	pool *pgxpool.Pool
}

// NewUserStore returns a UserStore backed by the given pool.
func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

// CreateUser inserts a new user row and returns the populated record.
func (s *UserStore) CreateUser(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	const q = `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, password_hash, created_at`

	u := &models.User{}
	err := s.pool.QueryRow(ctx, q, username, email, passwordHash).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("postgres: CreateUser: %w", err)
	}
	return u, nil
}

// GetUserByID returns the user with the given ID.
func (s *UserStore) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	const q = `
		SELECT id, username, email, password_hash, created_at
		FROM users WHERE id = $1`

	u := &models.User{}
	err := s.pool.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("postgres: GetUserByID: %w", err)
	}
	return u, nil
}

// GetUserByEmail returns the user with the given email, or an error when not found.
func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
		SELECT id, username, email, password_hash, created_at
		FROM users WHERE email = $1`

	u := &models.User{}
	err := s.pool.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("postgres: GetUserByEmail: %w", err)
	}
	return u, nil
}

// GetUserByUsername returns the user with the given username, or an error when not found.
func (s *UserStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const q = `
		SELECT id, username, email, password_hash, created_at
		FROM users WHERE username = $1`

	u := &models.User{}
	err := s.pool.QueryRow(ctx, q, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("postgres: GetUserByUsername: %w", err)
	}
	return u, nil
}

// UpdateUsername changes the username for the given user ID.
func (s *UserStore) UpdateUsername(ctx context.Context, userID, newUsername string) error {
	const q = `UPDATE users SET username = $1 WHERE id = $2`
	if _, err := s.pool.Exec(ctx, q, newUsername, userID); err != nil {
		return fmt.Errorf("postgres: UpdateUsername: %w", err)
	}
	return nil
}

// UpdatePassword replaces the stored password hash for the given user ID.
func (s *UserStore) UpdatePassword(ctx context.Context, userID, newPasswordHash string) error {
	const q = `UPDATE users SET password_hash = $1 WHERE id = $2`
	if _, err := s.pool.Exec(ctx, q, newPasswordHash, userID); err != nil {
		return fmt.Errorf("postgres: UpdatePassword: %w", err)
	}
	return nil
}
