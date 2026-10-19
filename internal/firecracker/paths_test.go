package firecracker

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestContainedJoinRejectsTraversal(t *testing.T) {
	base := filepath.Join(t.TempDir(), "sockets")
	cases := []string{
		"../../../../tmp/pwn",
		"..",
		"/etc/passwd",
		"foo/bar",
		"",
		"pwn.socket/../../tmp/x",
	}
	for _, name := range cases {
		if _, err := containedJoin(base, name); err == nil {
			t.Fatalf("containedJoin(%q) succeeded", name)
		}
	}
	got, err := containedJoin(base, "vm-user-box.socket")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, filepath.Clean(base)+string(filepath.Separator)) {
		t.Fatalf("path %s escaped %s", got, base)
	}
}

func TestValidVMName(t *testing.T) {
	if !ValidVMName("sandbox-1") || !ValidVMName("A") {
		t.Fatal("expected valid names")
	}
	for _, name := range []string{"", "../tmp/x", "has/slash", "-bad", "has space", strings.Repeat("a", 65)} {
		if ValidVMName(name) {
			t.Fatalf("ValidVMName(%q) = true", name)
		}
	}
}

func TestRejectUnsafeSocket(t *testing.T) {
	if err := RejectUnsafeSocket("/var/lib/rexec/vm-1.socket"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"", "/tmp/pwn", "/tmp/../etc/passwd.socket", "relative/../../tmp/x.socket"} {
		if err := RejectUnsafeSocket(p); err == nil {
			t.Fatalf("RejectUnsafeSocket(%q) succeeded", p)
		}
	}
}

func TestMapImageTypeRejectsUnknown(t *testing.T) {
	if _, ok := MapImageTypeToFileName("ubuntu"); !ok {
		t.Fatal("ubuntu should be mapped")
	}
	if _, ok := MapImageTypeToFileName("../../../../etc/passwd"); ok {
		t.Fatal("traversal image type must be rejected")
	}
	if _, ok := MapImageTypeToFileName("not-a-real-image"); ok {
		t.Fatal("unknown image must be rejected")
	}
}
