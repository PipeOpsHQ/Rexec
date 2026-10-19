package middleware

import "testing"

func TestInternalSSHAuthorized(t *testing.T) {
	const key = "gateway-secret"
	cases := []struct {
		name   string
		remote string
		header string
		config string
		ok     bool
	}{
		{"loopback with key", "127.0.0.1:54321", key, key, true},
		{"ipv6 loopback", "[::1]:8080", key, key, true},
		{"rfc1918 with key", "10.1.2.3:1234", key, key, true},
		{"docker bridge", "172.18.0.4:50000", key, key, true},
		{"public with key", "8.8.8.8:443", key, key, false},
		{"loopback wrong key", "127.0.0.1:1", "nope", key, false},
		{"loopback missing key header", "127.0.0.1:1", "", key, false},
		{"loopback unconfigured", "127.0.0.1:1", "", "", false},
		{"loopback empty config ignores header", "127.0.0.1:1", "anything", "  ", false},
		{"spoofed forwarded string is not an ip", "127.0.0.1, 8.8.8.8", key, key, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := InternalSSHAuthorized(tc.remote, tc.header, tc.config); got != tc.ok {
				t.Fatalf("got %v", got)
			}
		})
	}
}
