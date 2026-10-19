package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rexec/rexec/internal/container"
	"github.com/rexec/rexec/internal/models"
)

func TestRecordingContainerAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mgr := container.NewMemoryManager(map[string]*container.ContainerInfo{
		"abc123def456": {ID: "abc123def456", UserID: "owner", ContainerName: "box"},
	})
	h := &RecordingHandler{containerManager: mgr}
	h.SetCollabHandler(&CollabHandler{
		sessions: map[string]*CollabSession{
			"share": {
				ContainerID: "abc123def456",
				Participants: map[string]*CollabParticipant{
					"viewer": {UserID: "viewer", Role: "viewer"},
				},
			},
		},
	})

	tests := []struct {
		name   string
		userID string
		id     string
		want   int
	}{
		{name: "owner", userID: "owner", id: "abc123def456", want: http.StatusOK},
		{name: "name lookup", userID: "owner", id: "box", want: http.StatusOK},
		{name: "collaborator", userID: "viewer", id: "abc123def456", want: http.StatusOK},
		{name: "stranger", userID: "other", id: "abc123def456", want: http.StatusForbidden},
		{name: "unknown", userID: "owner", id: "missing", want: http.StatusNotFound},
		{name: "empty user", userID: "", id: "abc123def456", want: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/recordings/start", nil)
			c.Set("userID", tt.userID)
			info, status := h.recordingContainerAccess(c, tt.id)
			if status != tt.want {
				t.Fatalf("status = %d, want %d", status, tt.want)
			}
			if tt.want == http.StatusNotFound && info != nil {
				t.Fatal("unknown container leaked info")
			}
			if tt.want == http.StatusOK && (info == nil || info.ID != "abc123def456") {
				t.Fatalf("info = %#v", info)
			}
			if w.Body.Len() != 0 {
				t.Fatalf("access check wrote a response: %s", w.Body.String())
			}
		})
	}
}

func TestStatsHidesByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mgr := container.NewMemoryManager(map[string]*container.ContainerInfo{
		"c1": {ID: "c1", UserID: "tenant-a", ImageType: "ubuntu", Status: "running"},
		"c2": {ID: "c2", UserID: "tenant-b", ImageType: "debian", Status: "stopped"},
	})
	h := &ContainerHandler{manager: mgr}

	t.Run("member", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		c.Set("user", &models.User{ID: "tenant-a"})
		h.Stats(c)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d", w.Code)
		}
		body := w.Body.String()
		if strings.Contains(body, "by_user") || strings.Contains(body, "tenant-b") {
			t.Fatalf("non-admin response leaked tenants: %s", body)
		}
		if !strings.Contains(body, "ubuntu") {
			t.Fatalf("missing by_image: %s", body)
		}
	})

	t.Run("admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		c.Set("user", &models.User{ID: "admin", IsAdmin: true})
		h.Stats(c)
		body := w.Body.String()
		if !strings.Contains(body, "by_user") || !strings.Contains(body, "tenant-a") {
			t.Fatalf("admin response missing by_user: %s", body)
		}
	})
}
