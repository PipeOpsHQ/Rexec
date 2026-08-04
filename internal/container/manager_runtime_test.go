package container

import (
	"context"
	"net/netip"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func TestSandboxDNSServers_Default(t *testing.T) {
	t.Setenv("CONTAINER_DNS", "")
	got := sandboxDNSServers()
	want := []string{"8.8.8.8", "1.1.1.1"}
	if len(got) != len(want) {
		t.Fatalf("len(dns)=%d, want %d (%v)", len(got), len(want), got)
	}
	for i, addr := range got {
		if addr.String() != want[i] {
			t.Errorf("dns[%d]=%s, want %s", i, addr.String(), want[i])
		}
	}
}

func TestSandboxDNSServers_Override(t *testing.T) {
	t.Setenv("CONTAINER_DNS", "9.9.9.9, 208.67.222.222")
	got := sandboxDNSServers()
	if len(got) != 2 || got[0].String() != "9.9.9.9" || got[1].String() != "208.67.222.222" {
		t.Fatalf("got %v", got)
	}
}

func TestSandboxDNSServers_InvalidFallsBack(t *testing.T) {
	t.Setenv("CONTAINER_DNS", "not-an-ip,also-bad")
	got := sandboxDNSServers()
	if len(got) != 2 || got[0].String() != "8.8.8.8" {
		t.Fatalf("expected hard fallback, got %v", got)
	}
}

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

			mockClient.ContainerInspectFunc = func(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
				return client.ContainerInspectResult{
					Container: container.InspectResponse{
						HostConfig:      &container.HostConfig{},
						Config:          &container.Config{},
						NetworkSettings: &container.NetworkSettings{},
					},
				}, nil
			}

			var gotDNS []netip.Addr
			mockClient.ContainerCreateFunc = func(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
				gotRuntime = options.HostConfig.Runtime
				gotStorageSize = options.HostConfig.StorageOpt["size"]
				gotDNS = options.HostConfig.DNS
				return client.ContainerCreateResult{ID: "test-container-id"}, nil
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

			if len(gotDNS) == 0 {
				t.Fatal("expected HostConfig.DNS to be set")
			}
			if gotDNS[0].String() != "8.8.8.8" {
				t.Fatalf("dns[0]=%s, want 8.8.8.8", gotDNS[0].String())
			}
		})
	}
}
