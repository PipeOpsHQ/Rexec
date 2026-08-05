package container

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SystemPoolUserID owns unclaimed warm-pool sandboxes (same as marketplace system user).
const SystemPoolUserID = "00000000-0000-0000-0000-000000000000"

// WarmPoolConfig describes how many ready sandboxes to keep per image alias.
type WarmPoolConfig struct {
	Enabled bool
	// Targets maps image alias → desired ready count (e.g. "ubuntu": 2).
	Targets       map[string]int
	CheckInterval time.Duration
}

// ParseWarmPoolConfig reads WARM_POOL env: "ubuntu:2,debian:1" or empty/disabled.
// WARM_POOL_ENABLED=false disables even if WARM_POOL is set.
func ParseWarmPoolConfig() WarmPoolConfig {
	cfg := WarmPoolConfig{
		Targets:       map[string]int{},
		CheckInterval: 30 * time.Second,
	}
	if strings.EqualFold(os.Getenv("WARM_POOL_ENABLED"), "false") {
		return cfg
	}
	raw := strings.TrimSpace(os.Getenv("WARM_POOL"))
	if raw == "" {
		return cfg
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		image, nStr, ok := strings.Cut(part, ":")
		if !ok {
			image, nStr = part, "1"
		}
		image = strings.TrimSpace(image)
		n, err := strconv.Atoi(strings.TrimSpace(nStr))
		if err != nil || n < 1 {
			n = 1
		}
		if n > 20 {
			n = 20 // safety cap per image
		}
		if image != "" {
			cfg.Targets[image] = n
			cfg.Enabled = true
		}
	}
	if v := strings.TrimSpace(os.Getenv("WARM_POOL_INTERVAL_SEC")); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec >= 10 {
			cfg.CheckInterval = time.Duration(sec) * time.Second
		}
	}
	return cfg
}

// WarmPoolService keeps a stock of pre-created sandboxes and hands them out on claim.
type WarmPoolService struct {
	manager *Manager
	cfg     WarmPoolConfig
	mu      sync.Mutex
	stopCh  chan struct{}
}

// NewWarmPoolService creates a warm pool service (may be disabled).
func NewWarmPoolService(manager *Manager, cfg WarmPoolConfig) *WarmPoolService {
	return &WarmPoolService{
		manager: manager,
		cfg:     cfg,
		stopCh:  make(chan struct{}),
	}
}

// Start begins background refill if enabled.
func (s *WarmPoolService) Start() {
	if s == nil || !s.cfg.Enabled || len(s.cfg.Targets) == 0 {
		log.Printf("[WarmPool] disabled (set WARM_POOL=ubuntu:2 to enable)")
		return
	}
	go s.run()
	log.Printf("[WarmPool] started targets=%v interval=%v", s.cfg.Targets, s.cfg.CheckInterval)
}

// Stop stops the refill loop.
func (s *WarmPoolService) Stop() {
	if s == nil {
		return
	}
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}

func (s *WarmPoolService) run() {
	// Initial fill
	s.refill(context.Background())
	ticker := time.NewTicker(s.cfg.CheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.refill(context.Background())
		case <-s.stopCh:
			return
		}
	}
}

// CountReady returns unclaimed warm sandboxes for an image alias.
func (s *WarmPoolService) CountReady(image string) int {
	if s == nil {
		return 0
	}
	s.manager.mu.RLock()
	defer s.manager.mu.RUnlock()
	n := 0
	for _, info := range s.manager.containers {
		if isWarmUnclaimed(info, image) {
			n++
		}
	}
	return n
}

func isWarmUnclaimed(info *ContainerInfo, image string) bool {
	if info == nil || info.Status != "running" || info.Labels == nil {
		return false
	}
	if info.Labels["rexec.warm_pool"] != "true" {
		return false
	}
	if info.Labels["rexec.warm_claimed"] == "true" {
		return false
	}
	if info.UserID != SystemPoolUserID {
		return false
	}
	want := image
	if info.Labels["rexec.warm_image"] != want && info.ImageType != want {
		return false
	}
	return true
}

// Claim takes one warm sandbox for image and reassigns it to userID.
// Returns the ContainerInfo if successful.
func (s *WarmPoolService) Claim(image, userID, name string) (*ContainerInfo, error) {
	if s == nil || !s.cfg.Enabled {
		return nil, fmt.Errorf("warm pool disabled")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.manager.mu.Lock()
	defer s.manager.mu.Unlock()

	var target *ContainerInfo
	for _, info := range s.manager.containers {
		if isWarmUnclaimed(info, image) {
			target = info
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("no warm sandbox available for %s", image)
	}

	// Re-index user
	oldUser := target.UserID
	if ids, ok := s.manager.userIndex[oldUser]; ok {
		newIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != target.ID {
				newIDs = append(newIDs, id)
			}
		}
		if len(newIDs) > 0 {
			s.manager.userIndex[oldUser] = newIDs
		} else {
			delete(s.manager.userIndex, oldUser)
		}
	}

	target.UserID = userID
	if name != "" {
		target.ContainerName = name
	}
	if target.Labels == nil {
		target.Labels = map[string]string{}
	}
	target.Labels["rexec.warm_claimed"] = "true"
	target.Labels["rexec.warm_pool"] = "false"
	target.Labels["rexec.user_id"] = userID
	target.Labels["rexec.claimed_at"] = time.Now().UTC().Format(time.RFC3339)
	target.LastUsedAt = time.Now()

	s.manager.userIndex[userID] = append(s.manager.userIndex[userID], target.ID)

	// Copy for caller
	out := *target
	log.Printf("[WarmPool] claimed %s for user %s (image=%s)", target.ID[:12], userID, image)

	// Refill async
	go s.refill(context.Background())

	return &out, nil
}

func (s *WarmPoolService) refill(ctx context.Context) {
	if s == nil || !s.cfg.Enabled {
		return
	}
	// Serialize refill with claim
	s.mu.Lock()
	defer s.mu.Unlock()

	for image, want := range s.cfg.Targets {
		have := 0
		s.manager.mu.RLock()
		for _, info := range s.manager.containers {
			if isWarmUnclaimed(info, image) {
				have++
			}
		}
		s.manager.mu.RUnlock()

		need := want - have
		for i := 0; i < need; i++ {
			if err := s.createWarm(ctx, image); err != nil {
				log.Printf("[WarmPool] failed to create warm %s: %v", image, err)
				break
			}
		}
	}
}

func (s *WarmPoolService) createWarm(ctx context.Context, image string) error {
	if _, ok := SupportedImages[image]; !ok {
		return fmt.Errorf("unsupported warm image %q", image)
	}
	name := "warm-" + image + "-" + uuid.New().String()[:8]
	cfg := ContainerConfig{
		UserID:        SystemPoolUserID,
		ContainerName: name,
		ImageType:     image,
		MemoryLimit:   512 * 1024 * 1024,
		CPULimit:      500,
		DiskQuota:     2 * 1024 * 1024 * 1024,
		Labels: map[string]string{
			"rexec.warm_pool":    "true",
			"rexec.warm_claimed": "false",
			"rexec.warm_image":   image,
			"rexec.user_id":      SystemPoolUserID,
			"rexec.tier":         "system",
			"rexec.managed":      "true",
		},
	}
	// Pull if needed
	if err := s.manager.PullImage(ctx, image); err != nil {
		return err
	}
	info, err := s.manager.CreateContainer(ctx, cfg)
	if err != nil {
		return err
	}
	log.Printf("[WarmPool] created warm sandbox %s image=%s", info.ID[:12], image)
	return nil
}

// Stats returns ready counts per image.
func (s *WarmPoolService) Stats() map[string]int {
	out := map[string]int{}
	if s == nil {
		return out
	}
	for image := range s.cfg.Targets {
		out[image] = s.CountReady(image)
	}
	return out
}
