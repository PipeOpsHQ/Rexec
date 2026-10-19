package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPITokenAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name   string
		method string
		path   string
		scopes []string
		ok     bool
	}{
		{"agent websocket", http.MethodGet, "/ws/agent/abc", []string{"agent"}, true},
		{"agent terminal is a user route", http.MethodGet, "/ws/agent/abc/terminal", []string{"agent"}, false},
		{"agent list", http.MethodGet, "/api/agents", []string{"agent"}, true},
		{"agent delete", http.MethodDelete, "/api/agents/abc", []string{"agent"}, true},
		{"agent cannot register another", http.MethodPost, "/api/agents/register", []string{"agent"}, false},
		{"agent cannot read profile", http.MethodGet, "/api/profile", []string{"agent"}, false},
		{"agent cannot open admin", http.MethodGet, "/api/admin/users", []string{"agent"}, false},
		{"agent path traversal into admin", http.MethodGet, "/api/agents/../admin/users", []string{"agent"}, false},
		{"read write user api", http.MethodPost, "/api/containers", []string{"read", "write"}, true},
		{"read alone cannot create", http.MethodPost, "/api/containers", []string{"read"}, false},
		{"read alone cannot delete", http.MethodDelete, "/api/containers/abc", []string{"read"}, false},
		{"read alone cannot exec", http.MethodPost, "/api/containers/abc/exec", []string{"read"}, false},
		{"read can list", http.MethodGet, "/api/containers", []string{"read"}, true},
		{"write can delete", http.MethodDelete, "/api/containers/abc", []string{"write"}, true},
		{"admin can post", http.MethodPost, "/api/containers", []string{"admin"}, true},
		{"unknown scope fails closed", http.MethodGet, "/api/profile", []string{"superuser"}, false},
		{"read write blocked from admin", http.MethodGet, "/api/admin/users", []string{"read", "write"}, false},
		{"admin scope on admin", http.MethodDelete, "/api/admin/users/1", []string{"admin"}, true},
		{"admin scope on user api", http.MethodGet, "/api/profile", []string{"admin"}, true},
		{"empty scopes fail closed", http.MethodGet, "/api/profile", nil, false},
		{"user token can validate", http.MethodGet, "/api/tokens/validate", []string{"read"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := APITokenAllowed(tc.method, tc.path, tc.scopes); got != tc.ok {
				t.Fatalf("APITokenAllowed(%s %s %v)=%v", tc.method, tc.path, tc.scopes, got)
			}
		})
	}
}

func TestNormalizeAPITokenScopes(t *testing.T) {
	got, err := NormalizeAPITokenScopes(nil)
	if err != nil || len(got) != 2 || got[0] != "read" || got[1] != "write" {
		t.Fatalf("default = %v err=%v", got, err)
	}
	got, err = NormalizeAPITokenScopes([]string{" Write ", "read", "write"})
	if err != nil || len(got) != 2 {
		t.Fatalf("dedup = %v err=%v", got, err)
	}
	if _, err := NormalizeAPITokenScopes([]string{"root", "read"}); err == nil {
		t.Fatal("expected unknown scope to be rejected")
	}
}

func TestRejectAPITokenWithoutAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	c.Set("api_token", true)
	c.Set("api_token_scopes", []string{"read", "write"})
	if !rejectAPITokenWithoutAdmin(c) {
		t.Fatal("expected admin API token without admin scope to be rejected")
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	c.Set("api_token_scopes", []string{"read"})
	if rejectAPITokenWithoutAdmin(c) {
		t.Fatal("browser JWT must not be rejected here")
	}
}
