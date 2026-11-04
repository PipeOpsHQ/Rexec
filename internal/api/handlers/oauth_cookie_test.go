package handlers

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOAuthStateCookieSecureFollowsScheme(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("GIN_MODE", "debug")

	tests := []struct {
		name   string
		scheme string
		tls    bool
		proto  string
		maxAge int
		secure bool
	}{
		{name: "http", scheme: "http", maxAge: 900, secure: false},
		{name: "https tls", scheme: "https", tls: true, maxAge: 900, secure: true},
		{name: "forwarded https", scheme: "http", proto: "https", maxAge: 900, secure: true},
		{name: "forwarded HTTPS", scheme: "http", proto: "HTTPS", maxAge: 900, secure: true},
		{name: "clear on https", scheme: "https", tls: true, maxAge: -1, secure: true},
		{name: "clear on http", scheme: "http", maxAge: -1, secure: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodGet, tt.scheme+"://rexec.example/api/auth/oauth/url", nil)
			if tt.tls {
				req.TLS = &tls.ConnectionState{}
			}
			if tt.proto != "" {
				req.Header.Set("X-Forwarded-Proto", tt.proto)
			}
			c.Request = req

			setOAuthStateCookie(c, "state-token", tt.maxAge)

			res := w.Result()
			defer res.Body.Close()
			var found *http.Cookie
			for _, ck := range res.Cookies() {
				if ck.Name == "oauth_state" {
					found = ck
					break
				}
			}
			if found == nil {
				t.Fatal("oauth_state cookie was not set")
			}
			if found.Secure != tt.secure {
				t.Fatalf("Secure = %v, want %v", found.Secure, tt.secure)
			}
			if !found.HttpOnly {
				t.Fatal("cookie is not HttpOnly")
			}
			if found.Path != "/api/auth" {
				t.Fatalf("path %q", found.Path)
			}
			if found.SameSite != http.SameSiteLaxMode {
				t.Fatalf("SameSite = %v", found.SameSite)
			}
			if found.MaxAge != tt.maxAge {
				t.Fatalf("MaxAge = %d, want %d", found.MaxAge, tt.maxAge)
			}
		})
	}
}
