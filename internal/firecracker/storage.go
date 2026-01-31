package firecracker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// StorageManager handles storage operations for Firecracker VMs
type StorageManager struct {
	rootfsBasePath string
}

// NewStorageManager creates a new storage manager
func NewStorageManager(rootfsBasePath string) (*StorageManager, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(rootfsBasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create rootfs directory: %w", err)
	}

	return &StorageManager{
		rootfsBasePath: rootfsBasePath,
	}, nil
}

// GetRootfsPath returns the path to a rootfs image for the given image type
func (sm *StorageManager) GetRootfsPath(ctx context.Context, imageType string) (string, error) {
	// Map image type to filename
	rootfsName := MapImageTypeToFileName(imageType)
	rootfsPath := filepath.Join(sm.rootfsBasePath, "images", rootfsName)

	// Check if rootfs exists
	if _, err := os.Stat(rootfsPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("rootfs image not found for %s: %s (run scripts/build-rootfs.sh to create)", imageType, rootfsPath)
		}
		return "", fmt.Errorf("failed to check rootfs: %w", err)
	}

	return rootfsPath, nil
}

// CreateSnapshot creates a snapshot of a VM's rootfs
func (sm *StorageManager) CreateSnapshot(ctx context.Context, vmID, snapshotName string) (string, error) {
	// TODO: Implement ZFS snapshot or copy-based snapshot
	return "", fmt.Errorf("snapshots not yet implemented")
}

// CloneFromSnapshot creates a new rootfs from a snapshot
func (sm *StorageManager) CloneFromSnapshot(ctx context.Context, snapshotName string, newVMID string) (string, error) {
	// TODO: Implement ZFS clone or copy-based clone
	return "", fmt.Errorf("cloning not yet implemented")
}
