package container

import (
	"context"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/container"
)

func TestIsMacOSImageExactMatch(t *testing.T) {
	if !IsMacOSImage("macos") || !IsMacOSImage(" macos-legacy ") {
		t.Fatal("expected exact macos aliases")
	}
	for _, name := range []string{"ubuntu", "my-macos-box", "notosx", "custom", "docker-osx", ""} {
		if IsMacOSImage(name) {
			t.Fatalf("IsMacOSImage(%q) = true", name)
		}
	}
}

func TestMacOSPrivilegedAllowed(t *testing.T) {
	const admin = "admin-user"
	t.Setenv("ALLOW_MACOS_VM", "")
	t.Setenv("MACOS_VM_ALLOWLIST", "")
	if MacOSPrivilegedAllowed(admin, true) {
		t.Fatal("host flag off must deny")
	}
	t.Setenv("ALLOW_MACOS_VM", "true")
	if MacOSPrivilegedAllowed(admin, false) {
		t.Fatal("non-admin must be denied")
	}
	if !MacOSPrivilegedAllowed(admin, true) {
		t.Fatal("admin should be allowed when the host flag is on and the list is empty")
	}
	t.Setenv("MACOS_VM_ALLOWLIST", "other-user")
	if MacOSPrivilegedAllowed(admin, true) {
		t.Fatal("admin not on the allowlist must be denied")
	}
	t.Setenv("MACOS_VM_ALLOWLIST", "other-user, "+admin)
	if !MacOSPrivilegedAllowed(admin, true) {
		t.Fatal("listed admin should be allowed")
	}
	if MacOSPrivilegedAllowed("other-user", false) {
		t.Fatal("listed non-admin must be denied")
	}
}

func TestConfigureMacOSIsolationRefusesByDefault(t *testing.T) {
	t.Setenv("ALLOW_MACOS_VM", "true")
	host := &container.HostConfig{
		Privileged:  false,
		CapDrop:     []string{"ALL"},
		SecurityOpt: []string{"no-new-privileges:true"},
	}
	if err := configureMacOSIsolation(host, false); err == nil {
		t.Fatal("expected denial without caller approval")
	}
	if host.Privileged || host.CapDrop == nil || host.SecurityOpt == nil {
		t.Fatalf("host config was weakened: privileged=%v capDrop=%v securityOpt=%v", host.Privileged, host.CapDrop, host.SecurityOpt)
	}

	t.Setenv("ALLOW_MACOS_VM", "")
	if err := configureMacOSIsolation(host, true); err == nil {
		t.Fatal("expected denial without ALLOW_MACOS_VM")
	}
	if host.Privileged {
		t.Fatal("privileged bit was set")
	}
}

func TestCreateContainerRejectsUnapprovedMacOS(t *testing.T) {
	t.Setenv("ALLOW_MACOS_VM", "true")
	m := &Manager{
		containers: make(map[string]*ContainerInfo),
		userIndex:  make(map[string][]string),
	}
	_, err := m.CreateContainer(context.Background(), ContainerConfig{
		UserID:               "user-1",
		ContainerName:        "term",
		ImageType:            "macos",
		AllowPrivilegedMacOS: false,
	})
	if err == nil || !strings.Contains(err.Error(), "ALLOW_MACOS_VM") {
		t.Fatalf("err = %v", err)
	}

	_, err = m.CreateContainer(context.Background(), ContainerConfig{
		UserID:        "user-1",
		ContainerName: "term",
		ImageType:     "totally-macos-evil",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported image") {
		t.Fatalf("substring image type should be unsupported, err = %v", err)
	}
}
