package providers

import (
	"context"
	"io"
)

// TerminalInfo represents information about a terminal environment
// (container, VM, or agent)
type TerminalInfo struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	Name          string            `json:"name"`
	Provider      string            `json:"provider"` // "docker", "firecracker", "agent"
	Status        string            `json:"status"`   // "creating", "running", "stopped", "error"
	IPAddress     string            `json:"ip_address,omitempty"`
	CreatedAt     int64             `json:"created_at"`
	LastUsedAt    int64             `json:"last_used_at"`
	Labels        map[string]string `json:"labels,omitempty"`
	ResourceStats *ResourceStats    `json:"resource_stats,omitempty"`
}

// ResourceStats represents resource usage statistics
type ResourceStats struct {
	CPUPercent  float64 `json:"cpu_percent"`
	Memory      int64   `json:"memory"`       // bytes
	MemoryLimit int64   `json:"memory_limit"` // bytes
	DiskUsage   int64   `json:"disk_usage"`   // bytes
	DiskLimit   int64   `json:"disk_limit"`   // bytes
	NetRx       int64   `json:"net_rx"`       // bytes received
	NetTx       int64   `json:"net_tx"`       // bytes transmitted
}

// CreateConfig holds configuration for creating a new terminal environment
type CreateConfig struct {
	UserID      string
	Name        string
	Image       string            // Image type or name
	CustomImage string            // For custom images
	Role        string            // Optional role (node, python, etc.)
	MemoryMB    int64             // Memory limit in MB
	CPUShares   int64             // CPU shares (millicores: 1000 = 1 CPU)
	DiskMB      int64             // Disk quota in MB
	Labels      map[string]string // Custom labels
	UserData    string            // Cloud-init userdata (for VMs)
}

// TerminalConnection represents an active terminal connection
type TerminalConnection struct {
	ID       string
	Provider string
	Reader   io.Reader
	Writer   io.Writer
	Resize   func(cols, rows uint16) error
	Close    func() error
}

// Provider defines the interface for terminal providers (Docker, Firecracker, Agents)
type Provider interface {
	// Name returns the provider name (e.g., "docker", "firecracker", "agent")
	Name() string

	// IsAvailable checks if the provider is available/configured
	IsAvailable(ctx context.Context) bool

	// Create creates a new terminal environment
	Create(ctx context.Context, cfg CreateConfig) (*TerminalInfo, error)

	// Start starts a stopped terminal environment
	Start(ctx context.Context, id string) error

	// Stop stops a running terminal environment
	Stop(ctx context.Context, id string) error

	// Delete removes a terminal environment
	Delete(ctx context.Context, id string) error

	// Get retrieves information about a terminal environment
	Get(ctx context.Context, id string) (*TerminalInfo, error)

	// List returns all terminal environments for a user
	List(ctx context.Context, userID string) ([]*TerminalInfo, error)

	// ConnectTerminal establishes a terminal connection
	ConnectTerminal(ctx context.Context, id string, cols, rows uint16) (*TerminalConnection, error)

	// Exec executes a command in the terminal environment
	Exec(ctx context.Context, id string, cmd []string) ([]byte, error)

	// GetStats retrieves resource usage statistics
	GetStats(ctx context.Context, id string) (*ResourceStats, error)

	// StreamStats streams resource usage statistics
	StreamStats(ctx context.Context, id string) (<-chan *ResourceStats, error)
}

// Registry manages multiple providers
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider
func (r *Registry) Register(provider Provider) {
	r.providers[provider.Name()] = provider
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// List returns all registered providers
func (r *Registry) List() []Provider {
	result := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}

// GetAvailable returns all available providers
func (r *Registry) GetAvailable(ctx context.Context) []Provider {
	var available []Provider
	for _, p := range r.providers {
		if p.IsAvailable(ctx) {
			available = append(available, p)
		}
	}
	return available
}
