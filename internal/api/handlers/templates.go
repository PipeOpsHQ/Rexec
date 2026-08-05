package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mgr "github.com/rexec/rexec/internal/container"
	"github.com/rexec/rexec/internal/storage"
)

var templateNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,62}$`)

// TemplateHandler manages sandbox templates (committed images).
type TemplateHandler struct {
	containerManager *mgr.Manager
	store            *storage.PostgresStore
}

// NewTemplateHandler creates a TemplateHandler.
func NewTemplateHandler(cm *mgr.Manager, store *storage.PostgresStore) *TemplateHandler {
	return &TemplateHandler{containerManager: cm, store: store}
}

// CreateTemplateRequest is POST /api/templates body.
type CreateTemplateRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	FromSandboxID string `json:"from_sandbox_id" binding:"required"` // DB or Docker id
}

// CreateTemplate commits a running sandbox to a local image and saves metadata.
// POST /api/templates
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}
	name := strings.TrimSpace(req.Name)
	if !templateNameRe.MatchString(name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name: use 1-63 alphanumeric/._- starting with alnum"})
		return
	}

	ctx := c.Request.Context()
	dockerID, baseImage, status, owner, err := h.resolveSandbox(ctx, userID, strings.TrimSpace(req.FromSandboxID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if owner != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if !strings.EqualFold(status, "running") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sandbox must be running to create a template", "status": status})
		return
	}

	templateID := uuid.New().String()
	// Local-only image tag; not pushed to a registry in v1.
	ref := fmt.Sprintf("rexec-template/%s/%s:latest", shortID(userID), shortID(templateID))

	imageRef, err := h.containerManager.CommitSandboxImage(ctx, dockerID, ref,
		fmt.Sprintf("rexec template %s from %s", name, req.FromSandboxID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit sandbox: " + err.Error()})
		return
	}

	t := &storage.SandboxTemplate{
		ID:                templateID,
		UserID:            userID,
		Name:              name,
		Description:       strings.TrimSpace(req.Description),
		BaseImage:         baseImage,
		DockerImage:       imageRef,
		SourceContainerID: req.FromSandboxID,
		Status:            "ready",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := h.store.CreateSandboxTemplate(ctx, t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save template: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

// ListTemplates lists the caller's templates.
// GET /api/templates
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	list, err := h.store.ListSandboxTemplatesByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": list, "count": len(list)})
}

// GetTemplate returns one template.
// GET /api/templates/:id
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	t, err := h.store.GetSandboxTemplateByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	if t.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// DeleteTemplate removes template metadata (local image may remain until GC).
// DELETE /api/templates/:id
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err := h.store.DeleteSandboxTemplate(c.Request.Context(), id, userID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TemplateHandler) resolveSandbox(ctx context.Context, userID, id string) (dockerID, baseImage, status, owner string, err error) {
	if info, ok := h.containerManager.GetContainer(id); ok {
		return info.ID, info.ImageType, info.Status, info.UserID, nil
	}
	if rec, e := h.store.GetContainerByID(ctx, id); e == nil && rec != nil {
		if rec.DockerID == "" {
			return "", rec.Image, rec.Status, rec.UserID, fmt.Errorf("sandbox not ready")
		}
		if info, ok := h.containerManager.GetContainer(rec.DockerID); ok {
			return info.ID, rec.Image, info.Status, info.UserID, nil
		}
		return rec.DockerID, rec.Image, rec.Status, rec.UserID, nil
	}
	if rec, e := h.store.GetContainerByDockerID(ctx, id); e == nil && rec != nil {
		if info, ok := h.containerManager.GetContainer(rec.DockerID); ok {
			return info.ID, rec.Image, info.Status, info.UserID, nil
		}
		return rec.DockerID, rec.Image, rec.Status, rec.UserID, nil
	}
	_ = userID
	return "", "", "", "", fmt.Errorf("sandbox not found")
}

func shortID(id string) string {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) > 12 {
		return strings.ToLower(id[:12])
	}
	return strings.ToLower(id)
}
