package container

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestManagerCreateContainerRuntimeSelection(t *testing.T) {
	tests := []struct {
		name             string
		requestedRuntime string
		imageType        string
		diskQuota        int64
		wantRuntime      string
		wantStorageSize  string
	}{
		{
			name:            "auto falls back to runc when disk quotas are enabled",
			imageType:       "ubuntu",
			diskQuota:       2 * 1024 * 1024 * 1024,
			wantRuntime:     "runc",
			wantStorageSize: "2G",
		},
		{
			name:             "explicit runsc skips unsupported storage quota option",
			requestedRuntime: "runsc",
			imageType:        "ubuntu",
			diskQuota:        2 * 1024 * 1024 * 1024,
			wantRuntime:      "runsc",
			wantStorageSize:  "",
		},
		{
			name:             "macos legacy always uses runc",
			requestedRuntime: "runsc",
			imageType:        "macos-legacy",
			wantRuntime:      "runc",
			wantStorageSize:  "20G",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OCI_RUNTIME", tt.requestedRuntime)

			mockClient := &MockDockerClient{}
			var gotRuntime string
			var gotStorageSize string

			mockClient.ContainerInspectFunc = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
				return types.ContainerJSON{
					ContainerJSONBase: &types.ContainerJSONBase{},
					NetworkSettings:   &types.NetworkSettings{},
				}, nil
			}

			mockClient.ContainerCreateFunc = func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				gotRuntime = hostConfig.Runtime
				gotStorageSize = hostConfig.StorageOpt["size"]
				return container.CreateResponse{ID: "test-container-id"}, nil
			}

			manager := &Manager{
				client:            mockClient,
				containers:        make(map[string]*ContainerInfo),
				userIndex:         make(map[string][]string),
				diskQuotaEnabled:  true,
				diskQuotaChecked:  true,
				availableRuntimes: []string{"runc", "runsc"},
			}

			_, err := manager.CreateContainer(context.Background(), ContainerConfig{
				UserID:        "user-1",
				ContainerName: "term-1",
				ImageType:     tt.imageType,
				DiskQuota:     tt.diskQuota,
			})
			if err != nil {
				t.Fatalf("CreateContainer returned error: %v", err)
			}

			if gotRuntime != tt.wantRuntime {
				t.Fatalf("runtime = %q, want %q", gotRuntime, tt.wantRuntime)
			}

			if gotStorageSize != tt.wantStorageSize {
				t.Fatalf("storage size = %q, want %q", gotStorageSize, tt.wantStorageSize)
			}
		})
	}
}
