package handlers

import (
	"net/url"
	"testing"
)

func TestSanitizeProxyPath(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "empty becomes root", in: "", want: "/"},
		{name: "root", in: "/", want: "/"},
		{name: "simple path", in: "/api/health", want: "/api/health"},
		{name: "nested path", in: "/foo/bar/baz", want: "/foo/bar/baz"},
		{name: "cleans dots", in: "/foo/../bar", want: "/bar"},
		{name: "missing leading slash", in: "api/health", want: "/api/health"},
		{name: "allows at-sign in path", in: "/@user/profile", want: "/@user/profile"},
		{name: "allows scheme-like segment under root", in: "/redirect/https://evil.com", want: "/redirect/https:/evil.com"},
		{name: "rejects absolute URI without leading slash", in: "http://evil.com/", wantErr: true},
		{name: "rejects scheme relative", in: "//evil.com/path", wantErr: true},
		{name: "rejects backslash", in: "/foo\\bar", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeProxyPath(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got path %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildContainerProxyURL(t *testing.T) {
	t.Run("builds fixed host from container IP and port", func(t *testing.T) {
		u, addr, err := buildContainerProxyURL("10.0.0.5", 3000, "/app", "q=1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if addr != "10.0.0.5:3000" {
			t.Fatalf("addr = %q, want 10.0.0.5:3000", addr)
		}
		if u.Scheme != "http" {
			t.Fatalf("scheme = %q", u.Scheme)
		}
		if u.Host != "10.0.0.5:3000" {
			t.Fatalf("host = %q, want 10.0.0.5:3000", u.Host)
		}
		if u.Path != "/app" {
			t.Fatalf("path = %q", u.Path)
		}
		if u.RawQuery != "q=1" {
			t.Fatalf("query = %q", u.RawQuery)
		}
		// Host must not be influenced by path-like authority tricks in input.
		if u.Hostname() != "10.0.0.5" {
			t.Fatalf("hostname = %q", u.Hostname())
		}
	})

	t.Run("ipv6 uses join host port form", func(t *testing.T) {
		u, addr, err := buildContainerProxyURL("fd00::1", 8080, "/", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if addr != "[fd00::1]:8080" {
			t.Fatalf("addr = %q", addr)
		}
		if u.Host != "[fd00::1]:8080" {
			t.Fatalf("host = %q", u.Host)
		}
	})

	t.Run("rejects invalid IP", func(t *testing.T) {
		if _, _, err := buildContainerProxyURL("not-an-ip", 80, "/", ""); err == nil {
			t.Fatal("expected error for invalid IP")
		}
	})

	t.Run("rejects invalid port", func(t *testing.T) {
		if _, _, err := buildContainerProxyURL("10.0.0.1", 0, "/", ""); err == nil {
			t.Fatal("expected error for port 0")
		}
		if _, _, err := buildContainerProxyURL("10.0.0.1", 70000, "/", ""); err == nil {
			t.Fatal("expected error for port > 65535")
		}
	})

	t.Run("rejects path that could rewrite authority", func(t *testing.T) {
		if _, _, err := buildContainerProxyURL("10.0.0.1", 80, "//evil.com", ""); err == nil {
			t.Fatal("expected error for scheme-relative path")
		}
		if _, _, err := buildContainerProxyURL("10.0.0.1", 80, "http://evil.com", ""); err == nil {
			t.Fatal("expected error for absolute URL path")
		}
	})

	t.Run("string form keeps authority fixed", func(t *testing.T) {
		u, _, err := buildContainerProxyURL("172.18.0.2", 5000, "/api/v1", "x=y")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		parsed, err := url.Parse(u.String())
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if parsed.Hostname() != "172.18.0.2" || parsed.Port() != "5000" {
			t.Fatalf("parsed authority = %s:%s", parsed.Hostname(), parsed.Port())
		}
	})
}
