package container

import "testing"

func TestHostAllowed(t *testing.T) {
	allow := []string{"pypi.org", "*.ubuntu.com", "github.com"}
	cases := []struct {
		host string
		ok   bool
	}{
		{"pypi.org", true},
		{"files.pythonhosted.org", false},
		{"archive.ubuntu.com", true},
		{"security.ubuntu.com", true},
		{"github.com", true},
		{"evil.com", false},
		{"notubuntu.com", false},
	}
	for _, tc := range cases {
		if got := hostAllowed(tc.host, allow); got != tc.ok {
			t.Errorf("hostAllowed(%q)=%v want %v", tc.host, got, tc.ok)
		}
	}
}

func TestMergeEgressAllow(t *testing.T) {
	t.Setenv("RESTRICTED_EGRESS_ALLOW", "")
	got := MergeEgressAllow([]string{"api.openai.com", "pypi.org"})
	found := map[string]bool{}
	for _, h := range got {
		found[h] = true
	}
	if !found["pypi.org"] || !found["api.openai.com"] || !found["github.com"] {
		t.Fatalf("merge missing defaults or extras: %v", got)
	}
}
