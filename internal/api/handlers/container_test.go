package handlers

import (
	"strings"
	"testing"
	"github.com/rexec/rexec/internal/container"
)

func TestGenerateContainerName(t *testing.T) {
	name := generateContainerName()
	if name == "" {
		t.Error("generateContainerName() returned empty string")
	}

	// Format should be adjective-noun-number
	parts := strings.Split(name, "-")
	if len(parts) != 3 {
		t.Errorf("generateContainerName() = %s, expected 3 parts separated by hyphens", name)
	}
}

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected string
	}{
		{"Empty", "", 10, ""},
		{"Short", "hello", 10, "hello"},
		{"Exact", "hello", 5, "hello"},
		{"Long", "hello world", 5, "...world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateOutput(tt.input, tt.n)
			if got != tt.expected {
				t.Errorf("truncateOutput(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.expected)
			}
		})
	}
}

func TestIsValidContainerName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Empty", "", false},
		{"Too long", strings.Repeat("a", 65), false},
		{"Valid alphanumeric", "mycontainer1", true},
		{"Valid hyphen", "my-container", true},
		{"Valid underscore", "my_container", true},
		{"Start with hyphen", "-container", false},
		{"Start with underscore", "_container", false},
		{"Invalid char", "container!", false},
		{"Invalid char space", "container name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidContainerName(tt.input); got != tt.want {
				t.Errorf("isValidContainerName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidDockerImage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Empty", "", false},
		{"Too long", strings.Repeat("a", 257), false},
		{"Valid simple", "ubuntu", true},
		{"Valid with tag", "ubuntu:latest", true},
		{"Valid with registry", "docker.io/library/ubuntu", true},
		{"Valid with port", "registry.example.com:5000/ubuntu", true},
		{"Invalid char", "ubuntu!", false},
		{"Too many colons", "ubuntu:latest:extra", false},
		{"Empty name", ":latest", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidDockerImage(tt.input); got != tt.want {
				t.Errorf("isValidDockerImage(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetImageNames(t *testing.T) {
	names := getImageNames()
	if len(names) == 0 {
		t.Error("getImageNames() returned empty list")
	}
	
	// Check against container.SupportedImages to ensure consistency
	if len(names) != len(container.SupportedImages) {
		t.Errorf("getImageNames() returned %d names, want %d", len(names), len(container.SupportedImages))
	}
}

func TestFormatDiskQuota(t *testing.T) {
	tests := []struct {
		name   string
		input  int64
		want   string
	}{
		{"Megabytes", 512, "512M"},
		{"Gigabytes", 2048, "2G"},
		{"Gigabytes fraction", 1536, "1G"}, // Integer division
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDiskQuota(tt.input); got != tt.want {
				t.Errorf("formatDiskQuota(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
