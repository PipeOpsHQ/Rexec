package storage

import (
	"context"
	"time"
)

// ============================================================================
// Session operations
// ============================================================================

// SessionRecord represents a terminal session in the database
type SessionRecord struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	ContainerID string    `db:"container_id"`
	CreatedAt   time.Time `db:"created_at"`
	LastPingAt  time.Time `db:"last_ping_at"`
}

// CreateSession creates a new session record
func (s *PostgresStore) CreateSession(ctx context.Context, session *SessionRecord) error {
	query := `
		INSERT INTO sessions (id, user_id, container_id, created_at, last_ping_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.ContainerID,
		session.CreatedAt,
		session.LastPingAt,
	)
	return err
}

// DeleteSession deletes a session record
func (s *PostgresStore) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// UpdateSessionLastPing updates the last_ping_at timestamp for a session
func (s *PostgresStore) UpdateSessionLastPing(ctx context.Context, id string) error {
	query := `UPDATE sessions SET last_ping_at = $2 WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id, time.Now())
	return err
}

// GetActiveSessionsCount returns the count of sessions active in the last 5 minutes
func (s *PostgresStore) GetActiveSessionsCount(ctx context.Context) (int, error) {
	var count int
	// Define "active" as having pinged in the last 5 minutes
	query := `SELECT COUNT(*) FROM sessions WHERE last_ping_at > NOW() - INTERVAL '5 minutes'`
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}