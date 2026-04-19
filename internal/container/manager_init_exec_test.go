package container

import (
	"context"
	"net"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

func TestManager_EnsureIsolatedNetwork(t *testing.T) {
	// Setup mock client
	mockClient := &MockDockerClient{}

	// Setup Manager
	manager := &Manager{
		client: mockClient,
	}

	// Test Case 1: Network already exists
	t.Run("NetworkExists", func(t *testing.T) {
		mockClient.NetworkInspectFunc = func(ctx context.Context, networkID string, options client.NetworkInspectOptions) (client.NetworkInspectResult, error) {
			if networkID == IsolatedNetworkName {
				return client.NetworkInspectResult{
					Network: network.Inspect{Network: network.Network{Name: IsolatedNetworkName, ID: "existing-network-id"}},
				}, nil
			}
			return client.NetworkInspectResult{}, errdefs.ErrNotFound
		}

		// We can call ensureIsolatedNetwork directly on our manager instance since we are in package container.

		err := manager.ensureIsolatedNetwork()
		if err != nil {
			t.Errorf("ensureIsolatedNetwork failed: %v", err)
		}
	})

	// Test Case 2: Network does not exist, create it
	t.Run("CreateNetwork", func(t *testing.T) {
		mockClient.NetworkInspectFunc = func(ctx context.Context, networkID string, options client.NetworkInspectOptions) (client.NetworkInspectResult, error) {
			return client.NetworkInspectResult{}, errdefs.ErrNotFound
		}

		created := false
		mockClient.NetworkCreateFunc = func(ctx context.Context, name string, options client.NetworkCreateOptions) (client.NetworkCreateResult, error) {
			if name != IsolatedNetworkName {
				t.Errorf("Expected network name %s, got %s", IsolatedNetworkName, name)
			}
			created = true
			return client.NetworkCreateResult{ID: "new-network-id"}, nil
		}

		err := manager.ensureIsolatedNetwork()
		if err != nil {
			t.Errorf("ensureIsolatedNetwork failed: %v", err)
		}

		if !created {
			t.Error("Expected network to be created")
		}
	})
}

func TestManager_ExecInContainer(t *testing.T) {
	// Setup mock client
	mockClient := &MockDockerClient{}

	// Setup Manager
	manager := &Manager{
		client:     mockClient,
		containers: make(map[string]*ContainerInfo),
	}

	containerID := "test-container"
	manager.containers[containerID] = &ContainerInfo{
		ID:     containerID,
		Status: "running",
	}

	// Mock ExecCreate
	mockClient.ExecCreateFunc = func(ctx context.Context, ctr string, config client.ExecCreateOptions) (client.ExecCreateResult, error) {
		if ctr != containerID {
			t.Errorf("Expected container ID %s, got %s", containerID, ctr)
		}
		return client.ExecCreateResult{ID: "exec-id"}, nil
	}

	// Mock ExecAttach
	mockClient.ExecAttachFunc = func(ctx context.Context, execID string, config client.ExecAttachOptions) (client.ExecAttachResult, error) {
		if execID != "exec-id" {
			t.Errorf("Expected exec ID exec-id, got %s", execID)
		}
		conn, _ := net.Pipe()
		result := client.ExecAttachResult{}
		result.Conn = conn
		return result, nil
	}

	// Test ExecInContainer
	err := manager.ExecInContainer(context.Background(), containerID, []string{"ls"})
	if err != nil {
		t.Errorf("ExecInContainer failed: %v", err)
	}
}
