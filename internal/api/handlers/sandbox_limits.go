package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rexec/rexec/internal/container"
	"github.com/rexec/rexec/internal/models"
	"github.com/rexec/rexec/internal/providers"
	"github.com/rexec/rexec/internal/storage"
)

// callerIsAdmin reads the user loaded by AuthMiddleware.
func callerIsAdmin(c *gin.Context) bool {
	if c == nil {
		return false
	}
	v, ok := c.Get("user")
	if !ok || v == nil {
		return false
	}
	switch u := v.(type) {
	case *models.User:
		return u != nil && u.IsAdmin
	case models.User:
		return u.IsAdmin
	default:
		return false
	}
}

// macOSImageDenied returns a client error when image is a macOS VM the caller
// may not run. An empty string means the image is allowed or is not macOS.
func macOSImageDenied(userID string, isAdmin bool, image string) string {
	if !container.IsMacOSImage(image) {
		return ""
	}
	if container.MacOSPrivilegedAllowed(userID, isAdmin) {
		return ""
	}
	return container.MacOSDenyReason(userID, isAdmin)
}

// clampSandboxResources applies the caller's tier cap.
// A zero request uses the tier default. A request above the cap is clamped down.
// Guest accounts always use GuestResourceLimits, even if a subscription flag is set.
func clampSandboxResources(memoryMB, cpuShares, diskMB int64, tier string, subscriptionActive bool) (int64, int64, int64) {
	limits := models.GetUserResourceLimits(tier, subscriptionActive)
	if tier == "guest" {
		limits = models.GuestResourceLimits
	}
	return clampToLimit(memoryMB, limits.MemoryMB),
		clampToLimit(cpuShares, limits.CPUShares),
		clampToLimit(diskMB, limits.DiskMB)
}

func clampToLimit(requested, cap int64) int64 {
	if cap <= 0 {
		if requested > 0 {
			return requested
		}
		return 0
	}
	if requested <= 0 || requested > cap {
		return cap
	}
	return requested
}

// countUserSandboxes counts distinct sandboxes for a user across provider
// listings and database records. A provider id that is also stored as a
// container's Docker id is counted once.
func countUserSandboxes(ctx context.Context, store *storage.PostgresStore, registry *providers.Registry, userID string) (int, error) {
	if store == nil {
		return 0, fmt.Errorf("store unavailable")
	}
	ids := make(map[string]struct{})
	if registry != nil {
		for _, p := range registry.GetAvailable(ctx) {
			terms, err := p.List(ctx, userID)
			if err != nil {
				return 0, err
			}
			for _, t := range terms {
				if t != nil && t.ID != "" {
					ids[t.ID] = struct{}{}
				}
			}
		}
	}
	records, err := store.GetContainersByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	for _, rec := range records {
		if rec == nil {
			continue
		}
		if rec.DockerID != "" {
			ids[rec.DockerID] = struct{}{}
		} else if rec.ID != "" {
			ids["db:"+rec.ID] = struct{}{}
		}
	}
	return len(ids), nil
}
