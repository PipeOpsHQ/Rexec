package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mgr "github.com/rexec/rexec/internal/container"
	"github.com/rexec/rexec/internal/storage"
)

// SnapshotHandler manages sandbox filesystem snapshots (docker commit).
type SnapshotHandler struct {
	containerManager *mgr.Manager
	store            *storage.PostgresStore
}

// NewSnapshotHandler creates a SnapshotHandler.
func NewSnapshotHandler(cm *mgr.Manager, store *storage.PostgresStore) *SnapshotHandler {
	return &SnapshotHandler{containerManager: cm, store: store}
}

// CreateSnapshotRequest is POST /api/containers/:id/snapshot body.
type CreateSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateSnapshot commits the sandbox filesystem and stores metadata.
// POST /api/containers/:id/snapshot
func (h *SnapshotHandler) CreateSnapshot(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateSnapshotRequest
	_ = c.ShouldBindJSON(&req)

	ctx := c.Request.Context()
	dockerID, baseImage, status, owner, sourceID, err := h.resolveSandbox(ctx, userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if owner != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if !strings.EqualFold(status, "running") && !strings.EqualFold(status, "stopped") {
		// docker commit works on stopped containers too; block creating/error
		if strings.EqualFold(status, "creating") || strings.EqualFold(status, "error") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sandbox not ready for snapshot", "status": status})
			return
		}
	}
	if dockerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sandbox has no docker id yet"})
		return
	}

	snapID := uuid.New().String()
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "snap-" + time.Now().UTC().Format("20060102-150405")
	}
	ref := fmt.Sprintf("rexec-snapshot/%s/%s:latest", shortID(userID), shortID(snapID))

	imageRef, err := h.containerManager.CommitSandboxImage(ctx, dockerID, ref,
		fmt.Sprintf("rexec snapshot %s from %s", name, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to snapshot: " + err.Error()})
		return
	}

	snap := &storage.SandboxSnapshot{
		ID:                snapID,
		UserID:            userID,
		Name:              name,
		Description:       strings.TrimSpace(req.Description),
		SourceContainerID: sourceID,
		SourceDockerID:    dockerID,
		BaseImage:         baseImage,
		DockerImage:       imageRef,
		Status:            "ready",
		CreatedAt:         time.Now().UTC(),
	}
	if err := h.store.CreateSandboxSnapshot(ctx, snap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save snapshot: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, snap)
}

// ListSnapshots lists the caller's snapshots.
// GET /api/snapshots
func (h *SnapshotHandler) ListSnapshots(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	list, err := h.store.ListSandboxSnapshotsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"snapshots": list, "count": len(list)})
}

// GetSnapshot returns one snapshot.
// GET /api/snapshots/:id
func (h *SnapshotHandler) GetSnapshot(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	snap, err := h.store.GetSandboxSnapshotByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if snap == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}
	if snap.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	c.JSON(http.StatusOK, snap)
}

// DeleteSnapshot removes snapshot metadata.
// DELETE /api/snapshots/:id
func (h *SnapshotHandler) DeleteSnapshot(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err := h.store.DeleteSandboxSnapshot(c.Request.Context(), id, userID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ForkRequest is POST /api/containers/:id/fork body.
type ForkRequest struct {
	Name               string            `json:"name"`
	NetworkMode        string            `json:"network_mode"`
	EgressAllow        []string          `json:"egress_allow"`
	IdleTimeoutSeconds int               `json:"idle_timeout_seconds"`
	MaxLifetimeSeconds int               `json:"max_lifetime_seconds"`
	Labels             map[string]string `json:"labels"`
	// SaveSnapshot also stores the intermediate commit as a named snapshot.
	SaveSnapshot bool   `json:"save_snapshot"`
	SnapshotName string `json:"snapshot_name"`
}

// ForkSandbox commits the source sandbox and creates a new sandbox from that image.
// POST /api/containers/:id/fork
func (h *SnapshotHandler) ForkSandbox(c *gin.Context) {
	userID := c.GetString("userID")
	tier := c.GetString("tier")
	id := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ForkRequest
	_ = c.ShouldBindJSON(&req)

	ctx := c.Request.Context()
	dockerID, baseImage, status, owner, sourceID, err := h.resolveSandbox(ctx, userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if owner != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if dockerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source sandbox not ready"})
		return
	}
	if strings.EqualFold(status, "creating") || strings.EqualFold(status, "error") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source sandbox not ready", "status": status})
		return
	}

	// Limit check
	existing, err := h.store.GetContainersByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check limit"})
		return
	}
	limit := mgr.UserContainerLimit(tier)
	if len(existing) >= limit {
		c.JSON(http.StatusForbidden, gin.H{"error": "container limit reached", "limit": limit})
		return
	}

	forkID := uuid.New().String()
	ref := fmt.Sprintf("rexec-fork/%s/%s:latest", shortID(userID), shortID(forkID))
	imageRef, err := h.containerManager.CommitSandboxImage(ctx, dockerID, ref,
		fmt.Sprintf("rexec fork from %s", id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit for fork: " + err.Error()})
		return
	}

	var savedSnap *storage.SandboxSnapshot
	if req.SaveSnapshot {
		snapName := strings.TrimSpace(req.SnapshotName)
		if snapName == "" {
			snapName = "pre-fork-" + time.Now().UTC().Format("20060102-150405")
		}
		savedSnap = &storage.SandboxSnapshot{
			ID:                uuid.New().String(),
			UserID:            userID,
			Name:              snapName,
			Description:       "auto-saved before fork",
			SourceContainerID: sourceID,
			SourceDockerID:    dockerID,
			BaseImage:         baseImage,
			DockerImage:       imageRef,
			Status:            "ready",
			CreatedAt:         time.Now().UTC(),
		}
		if err := h.store.CreateSandboxSnapshot(ctx, savedSnap); err != nil {
			// non-fatal for fork
			savedSnap = nil
		}
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "fork-" + shortID(forkID)
	}
	// uniqueness
	for _, rec := range existing {
		if rec.Name == name {
			name = name + "-" + shortID(forkID)[:6]
			break
		}
	}

	labels := map[string]string{
		"rexec.tier":           tier,
		"rexec.user_id":        userID,
		"rexec.forked_from":    sourceID,
		"rexec.fork_source_id": id,
	}
	for k, v := range req.Labels {
		if k != "" && v != "" {
			labels[k] = v
		}
	}
	if req.IdleTimeoutSeconds > 0 {
		labels["rexec.idle_timeout_sec"] = fmt.Sprintf("%d", req.IdleTimeoutSeconds)
	}
	if req.MaxLifetimeSeconds > 0 {
		exp := time.Now().Add(time.Duration(req.MaxLifetimeSeconds) * time.Second)
		labels["rexec.expires_at"] = exp.UTC().Format(time.RFC3339)
	}

	cfg := mgr.ContainerConfig{
		UserID:        userID,
		ContainerName: name,
		ImageType:     "custom",
		CustomImage:   imageRef,
		NetworkMode:   req.NetworkMode,
		EgressAllow:   req.EgressAllow,
		Labels:        labels,
		MemoryLimit:   512 * 1024 * 1024,
		CPULimit:      500,
		DiskQuota:     2 * 1024 * 1024 * 1024,
	}

	// Persist DB row then create docker container
	record := &storage.ContainerRecord{
		ID:         uuid.New().String(),
		UserID:     userID,
		Name:       name,
		Image:      "custom:" + imageRef,
		Status:     "creating",
		DockerID:   "",
		VolumeName: "rexec-" + userID + "-" + name,
		MemoryMB:   512,
		CPUShares:  500,
		DiskMB:     2048,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}
	if err := h.store.CreateContainer(ctx, record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create record: " + err.Error()})
		return
	}

	info, err := h.containerManager.CreateContainer(ctx, cfg)
	if err != nil {
		_ = h.store.UpdateContainerStatus(ctx, record.ID, "error")
		_ = h.store.UpdateContainerError(ctx, record.ID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create forked sandbox: " + err.Error()})
		return
	}
	_ = h.store.UpdateContainerDockerID(ctx, record.ID, info.ID)
	_ = h.store.UpdateContainerStatus(ctx, record.ID, "running")

	resp := gin.H{
		"id":         info.ID,
		"db_id":      record.ID,
		"name":       name,
		"status":     "running",
		"image":      record.Image,
		"forked":     true,
		"source_id":  sourceID,
		"created_at": record.CreatedAt,
	}
	if savedSnap != nil {
		resp["snapshot"] = savedSnap
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *SnapshotHandler) resolveSandbox(ctx context.Context, userID, id string) (dockerID, baseImage, status, owner, sourceID string, err error) {
	if info, ok := h.containerManager.GetContainer(id); ok {
		return info.ID, info.ImageType, info.Status, info.UserID, id, nil
	}
	if rec, e := h.store.GetContainerByID(ctx, id); e == nil && rec != nil {
		src := rec.ID
		if rec.DockerID == "" {
			return "", rec.Image, rec.Status, rec.UserID, src, fmt.Errorf("sandbox not ready")
		}
		if info, ok := h.containerManager.GetContainer(rec.DockerID); ok {
			return info.ID, rec.Image, info.Status, info.UserID, src, nil
		}
		return rec.DockerID, rec.Image, rec.Status, rec.UserID, src, nil
	}
	if rec, e := h.store.GetContainerByDockerID(ctx, id); e == nil && rec != nil {
		src := rec.ID
		if info, ok := h.containerManager.GetContainer(rec.DockerID); ok {
			return info.ID, rec.Image, info.Status, info.UserID, src, nil
		}
		return rec.DockerID, rec.Image, rec.Status, rec.UserID, src, nil
	}
	_ = userID
	return "", "", "", "", "", fmt.Errorf("sandbox not found")
}
