package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/rexec/rexec/internal/container"
)

// DockerProvider adapts the existing container.Manager to the Provider interface
type DockerProvider struct {
	manager *container.Manager
}

// NewDockerProvider creates a new Docker provider adapter
func NewDockerProvider(manager *container.Manager) *DockerProvider {
	return &DockerProvider{
		manager: manager,
	}
}

// Name returns the provider name
func (p *DockerProvider) Name() string {
	return "docker"
}

// IsAvailable checks if Docker is available
func (p *DockerProvider) IsAvailable(ctx context.Context) bool {
	if p.manager == nil {
		return false
	}
	// Try to ping Docker
	client := p.manager.GetClient()
	if client == nil {
		return false
	}
	_, err := client.Ping(ctx)
	return err == nil
}

// Create creates a new container
func (p *DockerProvider) Create(ctx context.Context, cfg CreateConfig) (*TerminalInfo, error) {
	containerCfg := container.ContainerConfig{
		UserID:        cfg.UserID,
		ContainerName: cfg.Name,
		ImageType:     cfg.Image,
		CustomImage:   cfg.CustomImage,
		Role:          cfg.Role,
		MemoryLimit:   cfg.MemoryMB * 1024 * 1024, // Convert MB to bytes
		CPULimit:      cfg.CPUShares,              // Already in millicores
		DiskQuota:     cfg.DiskMB * 1024 * 1024,  // Convert MB to bytes
		Labels:        cfg.Labels,
	}

	info, err := p.manager.CreateContainer(ctx, containerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return p.toTerminalInfo(info), nil
}

// Start starts a stopped container
func (p *DockerProvider) Start(ctx context.Context, id string) error {
	return p.manager.StartContainer(ctx, id)
}

// Stop stops a running container
func (p *DockerProvider) Stop(ctx context.Context, id string) error {
	return p.manager.StopContainer(ctx, id)
}

// Delete removes a container
func (p *DockerProvider) Delete(ctx context.Context, id string) error {
	return p.manager.RemoveContainer(ctx, id)
}

// Get retrieves container information
func (p *DockerProvider) Get(ctx context.Context, id string) (*TerminalInfo, error) {
	info, ok := p.manager.GetContainer(id)
	if !ok {
		return nil, fmt.Errorf("container %s not found", id)
	}
	return p.toTerminalInfo(info), nil
}

// List returns all containers for a user
func (p *DockerProvider) List(ctx context.Context, userID string) ([]*TerminalInfo, error) {
	containers := p.manager.GetUserContainers(userID)
	result := make([]*TerminalInfo, len(containers))
	for i, c := range containers {
		result[i] = p.toTerminalInfo(c)
	}
	return result, nil
}

// ConnectTerminal establishes a terminal connection
func (p *DockerProvider) ConnectTerminal(ctx context.Context, id string, cols, rows uint16) (*TerminalConnection, error) {
	// This will be handled by the existing terminal handler
	// Return a placeholder for now
	return nil, fmt.Errorf("use existing terminal handler for Docker containers")
}

// Exec executes a command in the container
func (p *DockerProvider) Exec(ctx context.Context, id string, cmd []string) ([]byte, error) {
	// Use existing exec functionality
	err := p.manager.ExecInContainer(ctx, id, cmd)
	if err != nil {
		return nil, err
	}
	// TODO: Capture output
	return nil, fmt.Errorf("exec output capture not yet implemented")
}

// GetStats retrieves resource usage statistics
func (p *DockerProvider) GetStats(ctx context.Context, id string) (*ResourceStats, error) {
	statsCh := make(chan ContainerResourceStats, 1)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	go func() {
		err := p.manager.StreamContainerStats(ctx, id, statsCh)
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case stats := <-statsCh:
		return &ResourceStats{
			CPUPercent:  stats.CPUPercent,
			Memory:      int64(stats.Memory),
			MemoryLimit: int64(stats.MemoryLimit),
			DiskUsage:   int64(stats.DiskUsage),
			DiskLimit:   int64(stats.DiskLimit),
			NetRx:       int64(stats.NetRx),
			NetTx:       int64(stats.NetTx),
		}, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// StreamStats streams resource usage statistics
func (p *DockerProvider) StreamStats(ctx context.Context, id string) (<-chan *ResourceStats, error) {
	statsCh := make(chan *ResourceStats)
	dockerStatsCh := make(chan container.ContainerResourceStats, 10)

	go func() {
		defer close(statsCh)
		err := p.manager.StreamContainerStats(ctx, id, dockerStatsCh)
		if err != nil {
			return
		}

		for stats := range dockerStatsCh {
			statsCh <- &ResourceStats{
				CPUPercent:  stats.CPUPercent,
				Memory:      int64(stats.Memory),
				MemoryLimit: int64(stats.MemoryLimit),
				DiskUsage:   int64(stats.DiskUsage),
				DiskLimit:   int64(stats.DiskLimit),
				NetRx:       int64(stats.NetRx),
				NetTx:       int64(stats.NetTx),
			}
		}
	}()

	return statsCh, nil
}

// toTerminalInfo converts container.ContainerInfo to TerminalInfo
func (p *DockerProvider) toTerminalInfo(info *container.ContainerInfo) *TerminalInfo {
	return &TerminalInfo{
		ID:         info.ID,
		UserID:     info.UserID,
		Name:       info.ContainerName,
		Provider:   "docker",
		Status:     info.Status,
		IPAddress:  info.IPAddress,
		CreatedAt:  info.CreatedAt.Unix(),
		LastUsedAt: info.LastUsedAt.Unix(),
		Labels:     info.Labels,
	}
}

// ContainerResourceStats is needed for the adapter
type ContainerResourceStats = container.ContainerResourceStats
