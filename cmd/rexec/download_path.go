package main

import (
	"fmt"
	"path/filepath"
	"regexp"
)

// downloadNamePattern matches published binary names such as
// rexec-cli-linux-amd64 and rexec-agent-darwin-arm64.exe.
// The character set rejects separators, quotes, and parent-directory segments.
var downloadNamePattern = regexp.MustCompile(`^rexec-[A-Za-z0-9][A-Za-z0-9._-]*$`)

// resolveDownloadPath joins filename onto downloadsDir and refuses any name
// that could escape that directory.
func resolveDownloadPath(downloadsDir, filename string) (string, error) {
	if !downloadNamePattern.MatchString(filename) || !filepath.IsLocal(filename) {
		return "", fmt.Errorf("invalid download name")
	}
	base, err := filepath.Abs(downloadsDir)
	if err != nil {
		return "", err
	}
	full := filepath.Join(base, filename)
	rel, err := filepath.Rel(base, full)
	if err != nil || !filepath.IsLocal(rel) {
		return "", fmt.Errorf("download path escapes directory")
	}
	return full, nil
}
