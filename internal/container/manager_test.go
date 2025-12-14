package container

import (
	"errors"
	"testing"
)

func TestUserContainerLimit(t *testing.T) {
	tests := []struct {
		tier string
		want int
	}{
		{"trial", 5},
		{"guest", 1},
		{"free", 5},
		{"pro", 10},
		{"enterprise", 20},
		{"unknown", 5},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			if got := UserContainerLimit(tt.tier); got != tt.want {
				t.Errorf("UserContainerLimit(%s) = %d, want %d", tt.tier, got, tt.want)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "No error",
			err:  nil,
			want: "",
		},
		{
			name: "Generic error",
			err:  errors.New("something went wrong"),
			want: "something went wrong",
		},
		{
			name: "Docker daemon connection error",
			err:  errors.New("Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?"),
			want: "Container service temporarily unavailable. Please try again.",
		},
		{
			name: "TCP connection error",
			err:  errors.New("dial tcp: lookup tcp://1.2.3.4:2376: no such host"),
			want: "Container service temporarily unavailable. Please try again.",
		},
		{
			name: "Error with IP address",
			err:  errors.New("failed to connect to tcp://192.168.1.1:2376"),
			want: "Container service temporarily unavailable. Please try again.",
		},
		{
			name: "Error with Unix socket path",
			err:  errors.New("dial unix:///var/run/docker.sock: connect: permission denied"),
			want: "dial [docker-socket] connect: permission denied", // Parser consumes the colon
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeError(tt.err); got != tt.want {
				t.Errorf("SanitizeError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeErrorString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Normal string",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "String with tcp:// IP",
			input: "Connection to tcp://127.0.0.1:8080 failed",
			want:  "Container service temporarily unavailable. Please try again.", // Generic message triggers on tcp://
		},
		{
			name:  "String with unix:// path",
			input: "Socket at unix:///tmp/sock not found",
			want:  "Socket at [docker-socket] not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeErrorString(tt.input); got != tt.want {
				t.Errorf("SanitizeErrorString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"Zero", 0, "2G"},
		{"Small", 100, "0M"},
		{"Megabytes", 500 * 1024 * 1024, "500M"},
		{"Gigabytes", 2 * 1024 * 1024 * 1024, "2G"},
		{"Exact GB", 1024 * 1024 * 1024, "1G"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBytes(tt.bytes); got != tt.want {
				t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestGetImageMetadata(t *testing.T) {
	metadata := GetImageMetadata()
	if len(metadata) == 0 {
		t.Error("GetImageMetadata() returned empty list")
	}

	// Verify some known images exist
	foundUbuntu := false
	foundDebian := false
	for _, img := range metadata {
		if img.Name == "ubuntu" {
			foundUbuntu = true
		}
		if img.Name == "debian" {
			foundDebian = true
		}
	}

	if !foundUbuntu {
		t.Error("GetImageMetadata() missing ubuntu")
	}
	if !foundDebian {
		t.Error("GetImageMetadata() missing debian")
	}
}

func TestGetPopularImages(t *testing.T) {
	popular := GetPopularImages()
	if len(popular) == 0 {
		t.Error("GetPopularImages() returned empty list")
	}

	for _, img := range popular {
		if !img.Popular {
			t.Errorf("Image %s in popular list but Popular flag is false", img.Name)
		}
	}
}

func TestGetImagesByCategory(t *testing.T) {
	categories := GetImagesByCategory()
	if len(categories) == 0 {
		t.Error("GetImagesByCategory() returned empty map")
	}

	if _, ok := categories["debian"]; !ok {
		t.Error("GetImagesByCategory() missing 'debian' category")
	}
}

func TestIsCustomImageSupported(t *testing.T) {
	// Temporarily mock CustomImages if needed, but for now we test with default values
	// Assuming "ubuntu" is in CustomImages map in manager.go
	if !IsCustomImageSupported("ubuntu") {
		t.Error("IsCustomImageSupported('ubuntu') should be true")
	}
	if IsCustomImageSupported("nonexistent-image") {
		t.Error("IsCustomImageSupported('nonexistent-image') should be false")
	}
}

func TestGetImageName(t *testing.T) {
	// Test standard image
	name := GetImageName("ubuntu")
	if name == "" {
		t.Error("GetImageName('ubuntu') returned empty string")
	}

	// Test non-existent image
	name = GetImageName("nonexistent-image-type")
	if name != "" {
		t.Errorf("GetImageName('nonexistent-image-type') = %s, want empty string", name)
	}
}

func TestMergeLabels(t *testing.T) {
	base := map[string]string{"a": "1", "b": "2"}
	custom := map[string]string{"b": "3", "c": "4"}
	
	merged := mergeLabels(base, custom)
	
	if len(merged) != 3 {
		t.Errorf("mergeLabels returned map of size %d, want 3", len(merged))
	}
	
	if merged["a"] != "1" {
		t.Errorf("merged['a'] = %s, want 1", merged["a"])
	}
	if merged["b"] != "3" { // custom should overwrite base
		t.Errorf("merged['b'] = %s, want 3", merged["b"])
	}
	if merged["c"] != "4" {
		t.Errorf("merged['c'] = %s, want 4", merged["c"])
	}
}
