package firecracker

import (
	"fmt"
	"path/filepath"
	"strings"
)

// containedJoin joins base and name and refuses any result that escapes base.
// name must be a single local path segment.
func containedJoin(base, name string) (string, error) {
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid path segment")
	}
	if !filepath.IsLocal(name) {
		return "", fmt.Errorf("invalid path segment")
	}
	base = filepath.Clean(base)
	full := filepath.Join(base, name)
	rel, err := filepath.Rel(base, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes base directory")
	}
	return full, nil
}

// ValidVMName matches the sandbox name rule: 1-64 letters, digits, hyphens
// and underscores, not starting with a hyphen or underscore.
func ValidVMName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for i, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
		if i == 0 && (c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// RejectUnsafeSocket refuses socket paths that do not end in .socket or that
// contain a parent-directory segment. The raw path is checked before Clean,
// because Clean turns "/tmp/../etc/passwd.socket" into "/etc/passwd.socket".
func RejectUnsafeSocket(socketPath string) error {
	if strings.TrimSpace(socketPath) == "" {
		return fmt.Errorf("empty socket path")
	}
	if strings.Contains(socketPath, "..") {
		return fmt.Errorf("socket path escapes its directory")
	}
	cleaned := filepath.Clean(socketPath)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("socket path escapes its directory")
	}
	if !strings.HasSuffix(cleaned, ".socket") {
		return fmt.Errorf("socket path must end in .socket")
	}
	return nil
}
