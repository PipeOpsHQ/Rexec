package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SandboxTemplate is a reusable sandbox image (usually from docker commit).
type SandboxTemplate struct {
	ID                string    `db:"id" json:"id"`
	UserID            string    `db:"user_id" json:"user_id"`
	Name              string    `db:"name" json:"name"`
	Description       string    `db:"description" json:"description"`
	BaseImage         string    `db:"base_image" json:"base_image"`
	DockerImage       string    `db:"docker_image" json:"docker_image"`
	SourceContainerID string    `db:"source_container_id" json:"source_container_id,omitempty"`
	Status            string    `db:"status" json:"status"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// CreateSandboxTemplate inserts a new template row.
func (s *PostgresStore) CreateSandboxTemplate(ctx context.Context, t *SandboxTemplate) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}
	if t.Status == "" {
		t.Status = "ready"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sandbox_templates (
			id, user_id, name, description, base_image, docker_image,
			source_container_id, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.UserID, t.Name, t.Description, t.BaseImage, t.DockerImage,
		t.SourceContainerID, t.Status, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create sandbox template: %w", err)
	}
	return nil
}

// GetSandboxTemplateByID returns a template by id (any user — caller enforces ownership).
func (s *PostgresStore) GetSandboxTemplateByID(ctx context.Context, id string) (*SandboxTemplate, error) {
	var t SandboxTemplate
	err := s.db.GetContext(ctx, &t, `
		SELECT id, user_id, name, description, base_image, docker_image,
		       source_container_id, status, created_at, updated_at
		FROM sandbox_templates WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sandbox template: %w", err)
	}
	return &t, nil
}

// ListSandboxTemplatesByUserID lists templates for a user (newest first).
func (s *PostgresStore) ListSandboxTemplatesByUserID(ctx context.Context, userID string) ([]*SandboxTemplate, error) {
	var rows []*SandboxTemplate
	err := s.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, name, description, base_image, docker_image,
		       source_container_id, status, created_at, updated_at
		FROM sandbox_templates
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list sandbox templates: %w", err)
	}
	if rows == nil {
		rows = []*SandboxTemplate{}
	}
	return rows, nil
}

// DeleteSandboxTemplate deletes a template owned by userID.
func (s *PostgresStore) DeleteSandboxTemplate(ctx context.Context, id, userID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM sandbox_templates WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete sandbox template: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
