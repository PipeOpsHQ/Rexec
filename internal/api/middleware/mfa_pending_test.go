package middleware

import (
	"net/http"
	"testing"
)

func TestMFAPendingRequestAllowed(t *testing.T) {
	cases := []struct {
		method string
		path   string
		ok     bool
	}{
		{http.MethodPost, "/api/mfa/complete-login", true},
		{http.MethodPost, "/api/mfa/complete-login/", true},
		{http.MethodGet, "/api/mfa/complete-login", false},
		{http.MethodGet, "/api/profile", false},
		{http.MethodPut, "/api/profile", false},
		{http.MethodGet, "/api/tokens", false},
		{http.MethodPost, "/api/tokens", false},
		{http.MethodDelete, "/api/tokens/abc", false},
		{http.MethodGet, "/api/admin/users", false},
		{http.MethodPost, "/api/admin/users", false},
		{http.MethodGet, "/api/mfa/setup", false},
		{http.MethodPost, "/api/mfa/verify", false},
		{http.MethodPost, "/api/mfa/validate", false},
		{http.MethodPost, "/api/mfa/disable", false},
		{http.MethodGet, "/api/containers", false},
		{http.MethodGet, "/ws/terminal/abc", false},
		{http.MethodPost, "/api/mfa/../profile", false},
	}
	for _, tc := range cases {
		if got := MFAPendingRequestAllowed(tc.method, tc.path); got != tc.ok {
			t.Errorf("MFAPendingRequestAllowed(%s %s)=%v, want %v", tc.method, tc.path, got, tc.ok)
		}
	}
}
