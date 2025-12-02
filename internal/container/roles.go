package container

import (
	"fmt"
)

// RoleInfo represents a user role and its configuration
type RoleInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Packages    []string `json:"packages"` // Generic package names
}

// AvailableRoles returns the list of supported roles
func AvailableRoles() []RoleInfo {
	return []RoleInfo{
		{
			ID:          "standard",
			Name:        "Standard",
			Description: "Basic tools: git, curl, vim, nano",
			Icon:        "🛠️",
			Packages:    []string{"git", "curl", "wget", "vim", "nano", "htop", "jq"},
		},
		{
			ID:          "node",
			Name:        "Node.js Dev",
			Description: "Node.js, npm, yarn, pnpm",
			Icon:        "🟢",
			Packages:    []string{"nodejs", "npm", "yarn"},
		},
		{
			ID:          "python",
			Name:        "Python Dev",
			Description: "Python 3, pip, venv",
			Icon:        "🐍",
			Packages:    []string{"python3", "python3-pip", "python3-venv"},
		},
		{
			ID:          "go",
			Name:        "Go Dev",
			Description: "Go language environment",
			Icon:        "🐹",
			Packages:    []string{"go", "git", "make"},
		},
		{
			ID:          "rust",
			Name:        "Rust Dev",
			Description: "Rust, cargo (via rustup)",
			Icon:        "🦀",
			Packages:    []string{"rust", "cargo", "build-essential"}, // Note: rustup might be better but pkg is simpler
		},
		{
			ID:          "devops",
			Name:        "DevOps",
			Description: "Docker, kubectl, ansible",
			Icon:        "☸️",
			Packages:    []string{"docker-cli", "kubectl", "ansible", "terraform"},
		},
		{
			ID:          "cpp",
			Name:        "C/C++ Dev",
			Description: "GCC, G++, Make, CMake",
			Icon:        "⚡",
			Packages:    []string{"gcc", "g++", "make", "cmake", "gdb"},
		},
	}
}

// GenerateRoleScript generates the installation script for a specific role
func GenerateRoleScript(roleID string) (string, error) {
	var role *RoleInfo
	for _, r := range AvailableRoles() {
		if r.ID == roleID {
			role = &r
			break
		}
	}

	if role == nil {
		return "", fmt.Errorf("role not found: %s", roleID)
	}

	// Build the script reusing the package manager detection from shell_setup.go
	// We'll inject the specific packages for this role
	packages := ""
	for _, p := range role.Packages {
		packages += p + " "
	}

	script := fmt.Sprintf(`#!/bin/sh
set -e

echo "Installing tools for role: %s..."

# Function to install packages based on detected manager
install_role_packages() {
    PACKAGES="%s"

    if command -v apt-get >/dev/null 2>&1; then
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        # Map generic names to apt names if needed
        apt-get install -y -qq $PACKAGES >/dev/null 2>&1
    elif command -v apk >/dev/null 2>&1; then
        # Alpine mapping
        apk add --no-cache $PACKAGES >/dev/null 2>&1
    elif command -v dnf >/dev/null 2>&1; then
        dnf install -y -q $PACKAGES >/dev/null 2>&1
    elif command -v yum >/dev/null 2>&1; then
        yum install -y -q $PACKAGES >/dev/null 2>&1
    elif command -v pacman >/dev/null 2>&1; then
        pacman -Sy --noconfirm $PACKAGES >/dev/null 2>&1
    else
        echo "Unsupported package manager"
        exit 1
    fi
}

install_role_packages
echo "Role setup complete!"
`, role.Name, packages)

	return script, nil
}
