package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDownloadPath(t *testing.T) {
	base := t.TempDir()
	ok := []string{
		"rexec-cli-linux-amd64",
		"rexec-cli-windows-amd64.exe",
		"rexec-agent-darwin-arm64",
		"rexec-notes",
	}
	for _, name := range ok {
		got, err := resolveDownloadPath(base, name)
		if err != nil {
			t.Fatalf("resolveDownloadPath(%q): %v", name, err)
		}
		rel, err := filepath.Rel(base, got)
		if err != nil || !filepath.IsLocal(rel) || rel != name {
			t.Fatalf("path %s escaped %s (rel %s)", got, base, rel)
		}
	}

	bad := []string{
		"",
		"rexec-",
		"rexec-../../../etc/passwd",
		"rexec-cli-linux-amd64/../../etc/passwd",
		"/etc/passwd",
		"rexec-cli/linux",
		`rexec-cli\windows`,
		"../rexec-cli-linux-amd64",
		"rexec-cli-linux-amd64\n",
		"rexec-cli-linux-amd64\r\nSet-Cookie: x=y",
		`rexec-cli-"quoted"`,
		"other-cli-linux-amd64",
		"rexec-cli-linux-amd64\x00",
	}
	for _, name := range bad {
		if _, err := resolveDownloadPath(base, name); err == nil {
			t.Fatalf("resolveDownloadPath(%q) succeeded", name)
		}
	}
}

func TestResolveDownloadPathRejectsSibling(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "downloads")
	got, err := resolveDownloadPath(base, "rexec-cli-linux-amd64")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, filepath.Clean(base)+string(filepath.Separator)) {
		t.Fatalf("path %s is not inside %s", got, base)
	}
}
