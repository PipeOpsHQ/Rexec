package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	admin_events "github.com/rexec/rexec/internal/api/handlers/admin_events"
	"github.com/rexec/rexec/internal/models"
	"github.com/rexec/rexec/internal/storage"
)

// AdminHandler handles API requests related to admin functionalities.
type AdminHandler struct {
	store          *storage.PostgresStore
	adminEventsHub *admin_events.AdminEventsHub // New field
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(store *storage.PostgresStore, adminEventsHub *admin_events.AdminEventsHub) *AdminHandler {
	return &AdminHandler{store: store, adminEventsHub: adminEventsHub}
}

// ListUsers returns all users in the system.
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.store.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	containerCounts, err := h.store.GetContainerCountsByUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch container counts"})
		return
	}

	// Define a response struct that includes container count
	type AdminUserResponse struct {
		models.User
		ContainerCount int `json:"containerCount"`
	}

	response := make([]AdminUserResponse, len(users))
	for i, user := range users {
		count := containerCounts[user.ID]
		response[i] = AdminUserResponse{
			User:           *user, // Dereference the pointer
			ContainerCount: count,
		}
	}

	c.JSON(http.StatusOK, response)
}

// ListContainers returns all containers in the system with owner information.
func (h *AdminHandler) ListContainers(c *gin.Context) {
	containers, err := h.store.GetAllContainersAdmin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch containers"})
		return
	}
	c.JSON(http.StatusOK, containers)
}

// ListTerminals returns all active terminal sessions.
func (h *AdminHandler) ListTerminals(c *gin.Context) {
	terminals, err := h.store.GetAllSessionsAdmin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch terminals"})
		return
	}
	c.JSON(http.StatusOK, terminals)
}

// DeleteUser deletes a user by ID.
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Fetch user before deleting to include in broadcast payload
	userToDelete, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user for deletion"})
		return
	}
	if userToDelete == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := h.store.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Broadcast user deleted event
	if h.adminEventsHub != nil {
		h.adminEventsHub.Broadcast("user_deleted", userToDelete)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted successfully"})
}

// DeleteContainer deletes a container by ID.
func (h *AdminHandler) DeleteContainer(c *gin.Context) {
	containerID := c.Param("id")

	// Fetch container before deleting to include in broadcast payload
	containerToDelete, err := h.store.GetContainerByID(c.Request.Context(), containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch container for deletion"})
		return
	}
	if containerToDelete == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container not found"})
		return
	}

	// Assumes h.store.DeleteContainer method exists and handles necessary cleanup
	if err := h.store.DeleteContainer(c.Request.Context(), containerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete container"})
		return
	}

	// Broadcast container deleted event
	if h.adminEventsHub != nil {
		h.adminEventsHub.Broadcast("container_deleted", containerToDelete)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Container deleted successfully"})
}

// ListAgents returns all agents in the system.
func (h *AdminHandler) ListAgents(c *gin.Context) {
	agents, err := h.store.GetAllAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	threshold := time.Now().Add(-2 * time.Minute)
	for i := range agents {
		if !agents[i].LastPing.IsZero() && agents[i].LastPing.After(threshold) {
			agents[i].Status = "online"
		} else {
			agents[i].Status = "offline"
		}
	}

	c.JSON(http.StatusOK, agents)
}

// UsageStats returns time-filtered usage stats for the admin dashboard.
func (h *AdminHandler) UsageStats(c *gin.Context) {
	rangeKey := c.DefaultQuery("range", "30d")
	from, to, interval, ok := resolveAdminUsageRange(rangeKey)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid range. Use 24h, 7d, 30d, 90d, or 12m"})
		return
	}

	stats, err := h.store.GetAdminUsageStats(c.Request.Context(), from, to, interval)
	if err != nil {
		// Surface the real failure in logs — a missing table or SQL mismatch
		// previously left the "Usage over time" chart empty with only a generic 500.
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage stats", "detail": err.Error()})
		return
	}

	stats.Range = rangeKey
	c.JSON(http.StatusOK, stats)
}

func resolveAdminUsageRange(rangeKey string) (time.Time, time.Time, string, bool) {
	now := time.Now().UTC()

	switch rangeKey {
	case "24h":
		to := now.Truncate(time.Hour).Add(time.Hour)
		from := to.Add(-24 * time.Hour)
		return from, to, "hour", true
	case "7d":
		to := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
		from := to.AddDate(0, 0, -7)
		return from, to, "day", true
	case "30d":
		to := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
		from := to.AddDate(0, 0, -30)
		return from, to, "day", true
	case "90d":
		to := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
		from := startOfWeekUTC(to.AddDate(0, 0, -90))
		return from, to, "week", true
	case "12m":
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := monthStart.AddDate(0, 1, 0)
		from := monthStart.AddDate(0, -11, 0)
		return from, to, "month", true
	default:
		return time.Time{}, time.Time{}, "", false
	}
}

func startOfWeekUTC(t time.Time) time.Time {
	day := int(t.Weekday())
	if day == 0 {
		day = 7
	}
	start := t.Truncate(24 * time.Hour)
	return start.AddDate(0, 0, -(day - 1))
}
