package provider

import (
	"context"
	"io"
	"time"
)

// ProviderType represents the type of terminal provider
type ProviderType string

const (
	ProviderTypeDocker      ProviderType = "docker"
	ProviderTypeFirecracker ProviderType = "firecracker"
	ProviderTypeAgent       ProviderType = "agent"
)

// TerminalInfo holds information about a terminal instance
type TerminalInfo struct {
	ID            string            `json:"id"`
	Provider      ProviderType      `json:"provider"`
	UserID        string            `json:"user_id"`
	Name          string            `json:"name"`
	Status        string            `json:"status"` // creating, running, stopped, error
	IPAddress     string            `json:"ip_address,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	LastUsedAt    time.Time         `json:"last_used_at"`
	Labels        map[string]string `json:"labels,omitempty"`
	ProviderID    string            `json:"provider_id"` // Docker ID, VM ID, or Agent ID
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"` // Provider-specific data
}

// CreateConfig holds configuration for creating a new terminal
type CreateConfig struct {
	UserID      string            `json:"user_id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`        // Image type or name
	CustomImage string            `json:"custom_image,omitempty"` // For custom images
	Role        string            `json:"role,omitempty"`
	MemoryMB    int64             `json:"memory_mb"`
	CPUShares   int64             `json:"cpu_shares"` // In millicores (500 = 0.5 CPU)
	DiskMB      int64             `json:"disk_mb"`
	Labels      map[string]string `json:"labels,omitempty"`
	UserData    string            `json:"user_data,omitempty"` // Cloud-init userdata for VMs
}

// ResourceStats represents resource usage statistics
type ResourceStats struct {
	CPUPercent  float64 `json:"cpu_percent"`
	Memory      float64 `json:"memory"`       // in bytes
	MemoryLimit float64 `json:"memory_limit"` // in bytes
	DiskUsage   float64 `json:"disk_usage"`   // in bytes
	DiskLimit   float64 `json:"disk_limit"`   // in bytes
	NetRx       float64 `json:"net_rx"`       // bytes received
	NetTx       float64 `json:"net_tx"`       // bytes transmitted
}

// TerminalConnection represents an active terminal connection
type TerminalConnection interface {
	io.ReadWriteCloser
	Resize(cols, rows uint16) error
}

// ExecResult represents the result of executing a command
type ExecResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Duration time.Duration `json:"duration"`
}

// TerminalProvider defines the interface that all terminal providers must implement
type TerminalProvider interface {
	// Type returns the provider type
	Type() ProviderType

	// Create creates a new terminal instance
	Create(ctx context.Context, cfg CreateConfig) (*TerminalInfo, error)

	// Start starts a stopped terminal instance
	Start(ctx context.Context, id string) error

	// Stop stops a running terminal instance
	Stop(ctx context.Context, id string) error

	// Delete removes a terminal instance
	Delete(ctx context.Context, id string) error

	// Get retrieves information about a terminal instance
	Get(ctx context.Context, id string) (*TerminalInfo, error)

	// List returns all terminal instances for a user
	List(ctx context.Context, userID string) ([]*TerminalInfo, error)

	// ConnectTerminal establishes a terminal connection
	ConnectTerminal(ctx context.Context, id string) (TerminalConnection, error)

	// Exec executes a command in the terminal
	Exec(ctx context.Context, id string, cmd []string) (*ExecResult, error)

	// GetStats returns resource usage statistics
	GetStats(ctx context.Context, id string) (*ResourceStats, error)

	// StreamStats streams resource statistics to a channel
	StreamStats(ctx context.Context, id string, statsCh chan<- ResourceStats) error

	// IsAvailable checks if the provider is available/configured
	IsAvailable(ctx context.Context) bool
}

// ProviderRegistry manages multiple terminal providers
type ProviderRegistry struct {
	providers map[ProviderType]TerminalProvider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[ProviderType]TerminalProvider),
	}
}

// Register registers a provider
func (r *ProviderRegistry) Register(provider TerminalProvider) {
	r.providers[provider.Type()] = provider
}

// Get retrieves a provider by type
func (r *ProviderRegistry) Get(providerType ProviderType) (TerminalProvider, bool) {
	provider, ok := r.providers[providerType]
	return provider, ok
}

// GetDefault returns the default provider (docker) or first available
func (r *ProviderRegistry) GetDefault() TerminalProvider {
	// Prefer docker as default
	if docker, ok := r.providers[ProviderTypeDocker]; ok && docker.IsAvailable(context.Background()) {
		return docker
	}
	// Fallback to first available provider
	for _, provider := range r.providers {
		if provider.IsAvailable(context.Background()) {
			return provider
		}
	}
	return nil
}

// ListAvailable returns all available providers
func (r *ProviderRegistry) ListAvailable(ctx context.Context) []ProviderType {
	var available []ProviderType
	for providerType, provider := range r.providers {
		if provider.IsAvailable(ctx) {
			available = append(available, providerType)
		}
	}
	return available
}
