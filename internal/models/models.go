package models

import (
	"time"
)

// User represents a registered user
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`                    // Never serialize password
	Tier      string    `json:"tier"`                 // free, pro, enterprise
	PipeOpsID string    `json:"pipeops_id,omitempty"` // PipeOps OAuth user ID
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Container represents a user's terminal container
type Container struct {
	ID         string            `json:"id"`
	UserID     string            `json:"user_id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"` // ubuntu, debian, arch, etc.
	Status     ContainerStatus   `json:"status"`
	IPAddress  string            `json:"ip_address,omitempty"`
	VolumePath string            `json:"volume_path,omitempty"`
	Resources  ResourceLimits    `json:"resources"`
	Labels     map[string]string `json:"labels,omitempty"`
	DockerID   string            `json:"docker_id,omitempty"` // Actual Docker container ID
	CreatedAt  time.Time         `json:"created_at"`
	LastUsedAt time.Time         `json:"last_used_at"`
}

// ContainerStatus represents the state of a container
type ContainerStatus string

const (
	StatusCreating ContainerStatus = "creating"
	StatusRunning  ContainerStatus = "running"
	StatusStopped  ContainerStatus = "stopped"
	StatusPaused   ContainerStatus = "paused"
	StatusError    ContainerStatus = "error"
)

// ResourceLimits defines resource constraints for a container
type ResourceLimits struct {
	CPUShares int64 `json:"cpu_shares"` // CPU shares (relative weight)
	MemoryMB  int64 `json:"memory_mb"`  // Memory limit in MB
	DiskMB    int64 `json:"disk_mb"`    // Disk quota in MB
	NetworkMB int64 `json:"network_mb"` // Network bandwidth limit in MB/s
}

// Session represents an active terminal session
type Session struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ContainerID string    `json:"container_id"`
	PTY         string    `json:"-"` // PTY file descriptor path
	Cols        uint16    `json:"cols"`
	Rows        uint16    `json:"rows"`
	CreatedAt   time.Time `json:"created_at"`
	LastPingAt  time.Time `json:"last_ping_at"`
}

// TierLimits returns resource limits based on user tier
func TierLimits(tier string) ResourceLimits {
	switch tier {
	case "trial", "guest", "free": // Unified 60-day trial
		return ResourceLimits{
			CPUShares: 512,
			MemoryMB:  512,
			DiskMB:    2048, // Increased from 512-1024 for better trial experience
			NetworkMB: 10,
		}
	case "pro":
		return ResourceLimits{
			CPUShares: 2048,
			MemoryMB:  2048,
			DiskMB:    10240,
			NetworkMB: 100,
		}
	case "enterprise":
		return ResourceLimits{
			CPUShares: 4096,
			MemoryMB:  4096,
			DiskMB:    51200,
			NetworkMB: 500,
		}
	default: // Default to trial tier
		return ResourceLimits{
			CPUShares: 512,
			MemoryMB:  512,
			DiskMB:    2048,
			NetworkMB: 10,
		}
	}
}

// CreateContainerRequest represents a request to create a new container
type CreateContainerRequest struct {
	Name        string `json:"name"`                     // Optional - auto-generated if empty
	Image       string `json:"image" binding:"required"` // Image type (ubuntu, debian, etc.) or "custom"
	CustomImage string `json:"custom_image,omitempty"`   // Required when Image is "custom"
	Role        string `json:"role,omitempty"`           // Optional role (node, python, etc.)
}

// ResizeRequest represents a terminal resize request
type ResizeRequest struct {
	Cols uint16 `json:"cols" binding:"required"`
	Rows uint16 `json:"rows" binding:"required"`
}

// AuthRequest represents login/register request
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Username string `json:"username,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// APIError represents an API error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
