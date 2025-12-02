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
			Name:        "The Minimalist",
			Description: "I use Arch btw. Just give me a shell.",
			Icon:        "🧘",
			Packages:    []string{"git", "curl", "wget", "vim", "nano", "htop", "jq", "neofetch"},
		},
		{
			ID:          "node",
			Name:        "10x JS Ninja",
			Description: "Ship fast, break things, npm install everything.",
			Icon:        "🚀",
			Packages:    []string{"nodejs", "npm", "yarn"},
		},
		{
			ID:          "python",
			Name:        "Data Wizard",
			Description: "Import antigravity. I speak in list comprehensions.",
			Icon:        "🧙‍♂️",
			Packages:    []string{"python3", "python3-pip", "python3-venv"},
		},
		{
			ID:          "go",
			Name:        "The Gopher",
			Description: "If err != nil { panic(err) }. Simplicity is key.",
			Icon:        "🐹",
			Packages:    []string{"go", "git", "make"},
		},
		{
			ID:          "neovim",
			Name:        "Neovim God",
			Description: "My config is longer than your code. Mouse? What mouse?",
			Icon:        "⌨️",
			Packages:    []string{"neovim", "ripgrep", "git", "gcc", "make", "curl"},
		},
		{
			ID:          "devops",
			Name:        "YAML Herder",
			Description: "I don't write code, I write config. Prod is my playground.",
			Icon:        "☸️",
			Packages:    []string{"docker-cli", "kubectl", "ansible", "terraform"},
		},
		{
			ID:          "overemployed",
			Name:        "The Overemployed",
			Description: "Working 4 remote jobs. Need max efficiency & automation.",
			Icon:        "💼",
			Packages:    []string{"tmux", "screen", "python3", "cron", "htop", "vim"},
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
