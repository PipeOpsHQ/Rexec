package firecracker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// NetworkManager handles network setup for Firecracker VMs
type NetworkManager struct {
	bridgeName string
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(bridgeName string) (*NetworkManager, error) {
	return &NetworkManager{
		bridgeName: bridgeName,
	}, nil
}

// EnsureBridge ensures the bridge exists
func (nm *NetworkManager) EnsureBridge(ctx context.Context) error {
	// Check if bridge exists
	cmd := exec.CommandContext(ctx, "ip", "link", "show", nm.bridgeName)
	if err := cmd.Run(); err == nil {
		return nil // Bridge already exists
	}

	// Create bridge
	cmd = exec.CommandContext(ctx, "ip", "link", "add", "name", nm.bridgeName, "type", "bridge")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge: %w", err)
	}

	// Bring bridge up
	cmd = exec.CommandContext(ctx, "ip", "link", "set", nm.bridgeName, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring bridge up: %w", err)
	}

	return nil
}

// CreateTapDevice creates a tap device for a VM
func (nm *NetworkManager) CreateTapDevice(ctx context.Context, vmID string) (string, error) {
	tapName := fmt.Sprintf("tap-%s", vmID[:12]) // Use first 12 chars of VM ID

	// Create tap device
	cmd := exec.CommandContext(ctx, "ip", "tuntap", "add", "name", tapName, "mode", "tap")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create tap device: %w", err)
	}

	// Add tap to bridge
	cmd = exec.CommandContext(ctx, "ip", "link", "set", tapName, "master", nm.bridgeName)
	if err := cmd.Run(); err != nil {
		// Clean up tap device on error
		_ = nm.DeleteTapDevice(ctx, tapName)
		return "", fmt.Errorf("failed to add tap to bridge: %w", err)
	}

	// Bring tap up
	cmd = exec.CommandContext(ctx, "ip", "link", "set", tapName, "up")
	if err := cmd.Run(); err != nil {
		_ = nm.DeleteTapDevice(ctx, tapName)
		return "", fmt.Errorf("failed to bring tap up: %w", err)
	}

	return tapName, nil
}

// DeleteTapDevice deletes a tap device
func (nm *NetworkManager) DeleteTapDevice(ctx context.Context, tapName string) error {
	// Remove from bridge first
	cmd := exec.CommandContext(ctx, "ip", "link", "set", tapName, "nomaster")
	_ = cmd.Run() // Ignore error if not in bridge

	// Delete tap device
	cmd = exec.CommandContext(ctx, "ip", "link", "delete", tapName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete tap device: %w", err)
	}

	return nil
}

// GetVMIP gets the IP address assigned to a VM's tap device
func (nm *NetworkManager) GetVMIP(ctx context.Context, tapName string) (string, error) {
	// Check IP address on tap device
	cmd := exec.CommandContext(ctx, "ip", "addr", "show", tapName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get IP address: %w", err)
	}

	// Parse output to find IP address
	// Format: "inet 10.0.0.1/24 ..."
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ip := strings.Split(fields[1], "/")[0]
				return ip, nil
			}
		}
	}

	return "", fmt.Errorf("no IP address found for tap device %s", tapName)
}
