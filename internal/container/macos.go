package container

import (
	"errors"
	"os"
	"strings"

	"github.com/moby/moby/api/types/container"
)

// errMacOSNotAllowed is returned when a macOS image would otherwise run privileged.
var errMacOSNotAllowed = errors.New("macos images require ALLOW_MACOS_VM=true and an allowlisted administrator")

// IsMacOSImage reports whether imageType is a macOS VM image.
// Matching is exact on the normalised alias. A substring of the caller-supplied
// name must not select the privileged host configuration.
func IsMacOSImage(imageType string) bool {
	switch strings.ToLower(strings.TrimSpace(imageType)) {
	case "macos", "macos-legacy":
		return true
	default:
		return false
	}
}

// MacOSPrivilegedAllowed reports whether this caller may receive a privileged
// macOS VM. The host must opt in with ALLOW_MACOS_VM=true, the caller must be
// an administrator, and MACOS_VM_ALLOWLIST (when set) must include their user id.
func MacOSPrivilegedAllowed(userID string, isAdmin bool) bool {
	if !isAdmin || strings.TrimSpace(userID) == "" {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("ALLOW_MACOS_VM")), "true") {
		return false
	}
	allow := strings.TrimSpace(os.Getenv("MACOS_VM_ALLOWLIST"))
	if allow == "" {
		return true
	}
	for _, id := range strings.Split(allow, ",") {
		if strings.TrimSpace(id) == userID {
			return true
		}
	}
	return false
}

// macOSHostEnabled is the manager-side gate. The API sets callerAllowed only
// after MacOSPrivilegedAllowed; the host flag is checked again here so a
// direct CreateContainer call cannot drop privileges on its own.
func macOSHostEnabled(callerAllowed bool) bool {
	if !callerAllowed {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ALLOW_MACOS_VM")), "true")
}

// configureMacOSIsolation applies the KVM/privileged host config.
// It refuses to weaken isolation unless the host and caller were authorized.
func configureMacOSIsolation(hostConfig *container.HostConfig, callerAllowed bool) error {
	if hostConfig == nil {
		return errMacOSNotAllowed
	}
	if !macOSHostEnabled(callerAllowed) {
		return errMacOSNotAllowed
	}
	hostConfig.Privileged = true
	hostConfig.SecurityOpt = nil
	hostConfig.CapDrop = nil
	hostConfig.CapAdd = nil
	hostConfig.ReadonlyRootfs = false
	hostConfig.Tmpfs = nil
	if _, err := os.Stat("/dev/kvm"); err == nil {
		hostConfig.Devices = []container.DeviceMapping{
			{
				PathOnHost:        "/dev/kvm",
				PathInContainer:   "/dev/kvm",
				CgroupPermissions: "rwm",
			},
		}
	}
	return nil
}
