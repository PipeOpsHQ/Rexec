package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SandboxSnapshot is a point-in-time filesystem image of a sandbox (docker commit).
type SandboxSnapshot struct {
	ID                string    `db:"id" json:"id"`
	UserID            string    `db:"user_id" json:"user_id"`
	Name              string    `db:"name" json:"name"`
	Description       string    `db:"description" json:"description"`
	SourceContainerID string    `db:"source_container_id" json:"source_container_id,omitempty"`
	SourceDockerID    string    `db:"source_docker_id" json:"source_docker_id,omitempty"`
	BaseImage         string    `db:"base_image" json:"base_image,omitempty"`
	DockerImage       string    `db:"docker_image" json:"docker_image"`
	Status            string    `db:"status" json:"status"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

// CreateSandboxSnapshot inserts a snapshot row.
func (s *PostgresStore) CreateSandboxSnapshot(ctx context.Context, snap *SandboxSnapshot) error {
	if snap.CreatedAt.IsZero() {
		snap.CreatedAt = time.Now().UTC()
	}
	if snap.Status == "" {
		snap.Status = "ready"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sandbox_snapshots (
			id, user_id, name, description, source_container_id, source_docker_id,
			base_image, docker_image, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		snap.ID, snap.UserID, snap.Name, snap.Description, snap.SourceContainerID, snap.SourceDockerID,
		snap.BaseImage, snap.DockerImage, snap.Status, snap.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create sandbox snapshot: %w", err)
	}
	return nil
}

// GetSandboxSnapshotByID returns a snapshot by id.
func (s *PostgresStore) GetSandboxSnapshotByID(ctx context.Context, id string) (*SandboxSnapshot, error) {
	var snap SandboxSnapshot
	err := s.db.GetContext(ctx, &snap, `
		SELECT id, user_id, name, description, source_container_id, source_docker_id,
		       base_image, docker_image, status, created_at
		FROM sandbox_snapshots WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sandbox snapshot: %w", err)
	}
	return &snap, nil
}

// ListSandboxSnapshotsByUserID lists snapshots for a user (newest first).
func (s *PostgresStore) ListSandboxSnapshotsByUserID(ctx context.Context, userID string) ([]*SandboxSnapshot, error) {
	var rows []*SandboxSnapshot
	err := s.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, name, description, source_container_id, source_docker_id,
		       base_image, docker_image, status, created_at
		FROM sandbox_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list sandbox snapshots: %w", err)
	}
	if rows == nil {
		rows = []*SandboxSnapshot{}
	}
	return rows, nil
}

// DeleteSandboxSnapshot deletes a snapshot owned by userID.
func (s *PostgresStore) DeleteSandboxSnapshot(ctx context.Context, id, userID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM sandbox_snapshots WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete sandbox snapshot: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
