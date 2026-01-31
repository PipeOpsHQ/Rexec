package firecracker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ImageManager handles rootfs image management
type ImageManager struct {
	basePath string
}

// NewImageManager creates a new image manager
func NewImageManager(basePath string) (*ImageManager, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create image directory: %w", err)
	}

	return &ImageManager{
		basePath: basePath,
	}, nil
}

// ListImages returns available rootfs images
func (im *ImageManager) ListImages(ctx context.Context) ([]string, error) {
	imagesDir := filepath.Join(im.basePath, "images")
	
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read images directory: %w", err)
	}

	var images []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".ext4" {
			images = append(images, entry.Name())
		}
	}

	return images, nil
}

// GetImagePath returns the full path to an image
func (im *ImageManager) GetImagePath(imageName string) string {
	return filepath.Join(im.basePath, "images", imageName)
}

// ImageExists checks if an image exists
func (im *ImageManager) ImageExists(imageName string) bool {
	path := im.GetImagePath(imageName)
	_, err := os.Stat(path)
	return err == nil
}

// GetImageInfo returns information about an image
func (im *ImageManager) GetImageInfo(imageName string) (map[string]interface{}, error) {
	path := im.GetImagePath(imageName)
	
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	return map[string]interface{}{
		"name":     imageName,
		"path":     path,
		"size":     info.Size(),
		"mod_time": info.ModTime(),
	}, nil
}

// SupportedImageTypes returns the list of supported image types
func SupportedImageTypes() []string {
	return []string{
		"ubuntu",
		"ubuntu-24.04",
		"ubuntu-22.04",
		"debian",
		"debian-12",
		"debian-11",
	}
}

// MapImageTypeToFileName maps image type to actual image file name
func MapImageTypeToFileName(imageType string) string {
	// Default mapping
	mapping := map[string]string{
		"ubuntu":      "ubuntu-24.04.ext4",
		"ubuntu-24":   "ubuntu-24.04.ext4",
		"ubuntu-24.04": "ubuntu-24.04.ext4",
		"ubuntu-22":   "ubuntu-22.04.ext4",
		"ubuntu-22.04": "ubuntu-22.04.ext4",
		"debian":      "debian-12.ext4",
		"debian-12":   "debian-12.ext4",
		"debian-11":   "debian-11.ext4",
	}

	if filename, ok := mapping[imageType]; ok {
		return filename
	}

	// Default: assume image type is the filename
	return fmt.Sprintf("%s.ext4", imageType)
}
