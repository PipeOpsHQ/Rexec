package container

import (
	"testing"
)

func TestParseWarmPoolConfig_Empty(t *testing.T) {
	t.Setenv("WARM_POOL", "")
	t.Setenv("WARM_POOL_ENABLED", "")
	cfg := ParseWarmPoolConfig()
	if cfg.Enabled {
		t.Fatal("expected disabled")
	}
}

func TestParseWarmPoolConfig_Targets(t *testing.T) {
	t.Setenv("WARM_POOL", "ubuntu:2, debian:1")
	t.Setenv("WARM_POOL_ENABLED", "")
	cfg := ParseWarmPoolConfig()
	if !cfg.Enabled {
		t.Fatal("expected enabled")
	}
	if cfg.Targets["ubuntu"] != 2 || cfg.Targets["debian"] != 1 {
		t.Fatalf("targets = %#v", cfg.Targets)
	}
}

func TestParseWarmPoolConfig_Disabled(t *testing.T) {
	t.Setenv("WARM_POOL", "ubuntu:2")
	t.Setenv("WARM_POOL_ENABLED", "false")
	cfg := ParseWarmPoolConfig()
	if cfg.Enabled {
		t.Fatal("expected disabled via WARM_POOL_ENABLED=false")
	}
}

func TestIsWarmUnclaimed(t *testing.T) {
	info := &ContainerInfo{
		ID:     "abc",
		UserID: SystemPoolUserID,
		Status: "running",
		Labels: map[string]string{
			"rexec.warm_pool":    "true",
			"rexec.warm_claimed": "false",
			"rexec.warm_image":   "ubuntu",
		},
		ImageType: "ubuntu",
	}
	if !isWarmUnclaimed(info, "ubuntu") {
		t.Fatal("expected unclaimed")
	}
	info.Labels["rexec.warm_claimed"] = "true"
	if isWarmUnclaimed(info, "ubuntu") {
		t.Fatal("expected claimed")
	}
}
