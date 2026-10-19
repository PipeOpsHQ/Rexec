package container

import "testing"

func TestClaimRejectsMacOS(t *testing.T) {
	svc := &WarmPoolService{cfg: WarmPoolConfig{Enabled: true}}
	for _, image := range []string{"macos", "macos-legacy", " macos "} {
		if _, err := svc.Claim(image, "user", "box"); err == nil {
			t.Fatalf("Claim(%q) succeeded", image)
		}
	}
}
