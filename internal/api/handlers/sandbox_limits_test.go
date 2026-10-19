package handlers

import (
	"testing"

	"github.com/rexec/rexec/internal/models"
)

func TestClampSandboxResources(t *testing.T) {
	guest := models.GuestResourceLimits
	tests := []struct {
		name    string
		mem     int64
		cpu     int64
		disk    int64
		tier    string
		sub     bool
		wantMem int64
		wantCPU int64
		wantDsk int64
	}{
		{name: "guest zeros become defaults", tier: "guest", wantMem: guest.MemoryMB, wantCPU: guest.CPUShares, wantDsk: guest.DiskMB},
		{name: "guest above cap clamps", mem: 8192, cpu: 8000, disk: 99999, tier: "guest", wantMem: guest.MemoryMB, wantCPU: guest.CPUShares, wantDsk: guest.DiskMB},
		{name: "guest below cap kept", mem: 256, cpu: 250, disk: 1024, tier: "guest", wantMem: 256, wantCPU: 250, wantDsk: 1024},
		{name: "guest ignores subscription", mem: 0, cpu: 0, disk: 0, tier: "guest", sub: true, wantMem: guest.MemoryMB, wantCPU: guest.CPUShares, wantDsk: guest.DiskMB},
		{name: "free zero uses tier default", tier: "free", wantMem: 2048, wantCPU: 2000, wantDsk: 10240},
		{name: "free above cap clamps", mem: 99999, cpu: 99999, disk: 99999, tier: "free", wantMem: 2048, wantCPU: 2000, wantDsk: 10240},
		{name: "subscribed zero uses pro defaults", tier: "free", sub: true, wantMem: 4096, wantCPU: 4000, wantDsk: 20480},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem, cpu, disk := clampSandboxResources(tt.mem, tt.cpu, tt.disk, tt.tier, tt.sub)
			if mem != tt.wantMem || cpu != tt.wantCPU || disk != tt.wantDsk {
				t.Fatalf("got mem=%d cpu=%d disk=%d, want %d %d %d", mem, cpu, disk, tt.wantMem, tt.wantCPU, tt.wantDsk)
			}
		})
	}
}

func TestMacOSImageDenied(t *testing.T) {
	t.Setenv("ALLOW_MACOS_VM", "")
	if msg := macOSImageDenied("user", false, "ubuntu"); msg != "" {
		t.Fatalf("ubuntu denied: %s", msg)
	}
	if msg := macOSImageDenied("user", true, "not-macos-at-all"); msg != "" {
		t.Fatalf("substring image denied: %s", msg)
	}
	if msg := macOSImageDenied("user", true, "macos"); msg == "" {
		t.Fatal("macos allowed without ALLOW_MACOS_VM")
	}
	t.Setenv("ALLOW_MACOS_VM", "true")
	if msg := macOSImageDenied("admin", false, "macos"); msg == "" {
		t.Fatal("non-admin allowed")
	}
	if msg := macOSImageDenied("admin", true, "macos-legacy"); msg != "" {
		t.Fatalf("allowlisted admin denied: %s", msg)
	}
}
