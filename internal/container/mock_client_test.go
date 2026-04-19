package container

import (
	"context"
	"net"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// MockDockerClient implements client.APIClient for testing
type MockDockerClient struct {
	client.APIClient
	ContainerCreateFunc  func(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error)
	ContainerStartFunc   func(ctx context.Context, containerID string, options client.ContainerStartOptions) (client.ContainerStartResult, error)
	ContainerStopFunc    func(ctx context.Context, containerID string, options client.ContainerStopOptions) (client.ContainerStopResult, error)
	ContainerRemoveFunc  func(ctx context.Context, containerID string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error)
	ContainerListFunc    func(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error)
	ContainerInspectFunc func(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error)
	ImagePullFunc        func(ctx context.Context, ref string, options client.ImagePullOptions) (client.ImagePullResponse, error)
	ImageInspectFunc     func(ctx context.Context, imageID string, opts ...client.ImageInspectOption) (client.ImageInspectResult, error)
	NetworkListFunc      func(ctx context.Context, options client.NetworkListOptions) (client.NetworkListResult, error)
	NetworkCreateFunc    func(ctx context.Context, name string, options client.NetworkCreateOptions) (client.NetworkCreateResult, error)
	NetworkInspectFunc   func(ctx context.Context, networkID string, options client.NetworkInspectOptions) (client.NetworkInspectResult, error)
	ExecCreateFunc       func(ctx context.Context, container string, config client.ExecCreateOptions) (client.ExecCreateResult, error)
	ExecAttachFunc       func(ctx context.Context, execID string, config client.ExecAttachOptions) (client.ExecAttachResult, error)
	ContainerStatsFunc   func(ctx context.Context, containerID string, options client.ContainerStatsOptions) (client.ContainerStatsResult, error)
}

func (m *MockDockerClient) ExecCreate(ctx context.Context, container string, config client.ExecCreateOptions) (client.ExecCreateResult, error) {
	if m.ExecCreateFunc != nil {
		return m.ExecCreateFunc(ctx, container, config)
	}
	return client.ExecCreateResult{ID: "test-exec-id"}, nil
}

func (m *MockDockerClient) ExecAttach(ctx context.Context, execID string, config client.ExecAttachOptions) (client.ExecAttachResult, error) {
	if m.ExecAttachFunc != nil {
		return m.ExecAttachFunc(ctx, execID, config)
	}

	// Return a valid ExecAttachResult with a dummy connection to avoid panic on Close()
	conn, _ := net.Pipe()
	result := client.ExecAttachResult{}
	result.Conn = conn
	return result, nil
}

func (m *MockDockerClient) ContainerStats(ctx context.Context, containerID string, options client.ContainerStatsOptions) (client.ContainerStatsResult, error) {
	if m.ContainerStatsFunc != nil {
		return m.ContainerStatsFunc(ctx, containerID, options)
	}
	return client.ContainerStatsResult{}, nil
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
	if m.ContainerListFunc != nil {
		return m.ContainerListFunc(ctx, options)
	}
	return client.ContainerListResult{}, nil
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	if m.ContainerInspectFunc != nil {
		return m.ContainerInspectFunc(ctx, containerID, options)
	}
	return client.ContainerInspectResult{
		Container: container.InspectResponse{
			HostConfig: &container.HostConfig{},
			Config:     &container.Config{},
		},
	}, nil
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
	if m.ContainerCreateFunc != nil {
		return m.ContainerCreateFunc(ctx, options)
	}
	return client.ContainerCreateResult{ID: "test-container-id"}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options client.ContainerStartOptions) (client.ContainerStartResult, error) {
	if m.ContainerStartFunc != nil {
		return m.ContainerStartFunc(ctx, containerID, options)
	}
	return client.ContainerStartResult{}, nil
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options client.ContainerStopOptions) (client.ContainerStopResult, error) {
	if m.ContainerStopFunc != nil {
		return m.ContainerStopFunc(ctx, containerID, options)
	}
	return client.ContainerStopResult{}, nil
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
	if m.ContainerRemoveFunc != nil {
		return m.ContainerRemoveFunc(ctx, containerID, options)
	}
	return client.ContainerRemoveResult{}, nil
}

func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options client.ImagePullOptions) (client.ImagePullResponse, error) {
	if m.ImagePullFunc != nil {
		return m.ImagePullFunc(ctx, ref, options)
	}
	return nil, nil
}

func (m *MockDockerClient) ImageInspect(ctx context.Context, imageID string, opts ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	if m.ImageInspectFunc != nil {
		return m.ImageInspectFunc(ctx, imageID, opts...)
	}
	return client.ImageInspectResult{}, nil
}

func (m *MockDockerClient) NetworkInspect(ctx context.Context, networkID string, options client.NetworkInspectOptions) (client.NetworkInspectResult, error) {
	if m.NetworkInspectFunc != nil {
		return m.NetworkInspectFunc(ctx, networkID, options)
	}
	return client.NetworkInspectResult{}, nil
}

func (m *MockDockerClient) NetworkList(ctx context.Context, options client.NetworkListOptions) (client.NetworkListResult, error) {
	if m.NetworkListFunc != nil {
		return m.NetworkListFunc(ctx, options)
	}
	return client.NetworkListResult{}, nil
}

func (m *MockDockerClient) NetworkCreate(ctx context.Context, name string, options client.NetworkCreateOptions) (client.NetworkCreateResult, error) {
	if m.NetworkCreateFunc != nil {
		return m.NetworkCreateFunc(ctx, name, options)
	}
	return client.NetworkCreateResult{ID: "test-network-id"}, nil
}

func (m *MockDockerClient) Close() error {
	return nil
}
