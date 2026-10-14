package container

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Metadata destinations sandboxes must not reach. Host-generated packets do
// not traverse DOCKER-USER, so the host agent can still read IMDS.
//
// Keep this list identical to scripts/block-sandbox-metadata.sh and the
// copy installed by docker/docker-host/setup.sh.
var metadataDropDestinations = []struct {
	bin  string
	dest string
}{
	{"iptables", "169.254.169.254/32"}, // Azure, AWS, and GCP IMDS
	{"iptables", "169.254.169.253/32"},
	{"iptables", "169.254.170.2/32"},   // ECS task credentials
	{"iptables", "168.63.129.16/32"},   // Azure wireserver
	{"ip6tables", "fd00:ec2::254/128"}, // AWS IPv6 IMDS
}

// CommandRunner executes a firewall command. Tests supply a fake.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
}

// ExecRunner runs commands on the host.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, msg)
	}
	return nil
}

type metadataRule struct {
	Bin    string
	Dest   string
	Check  []string
	Insert []string
}

// MetadataDropRules is the DOCKER-USER DROP list applied to forwarded
// container traffic.
func MetadataDropRules() []metadataRule {
	rules := make([]metadataRule, 0, len(metadataDropDestinations))
	for _, d := range metadataDropDestinations {
		rules = append(rules, metadataRule{
			Bin:    d.bin,
			Dest:   d.dest,
			Check:  []string{"-C", "DOCKER-USER", "-d", d.dest, "-j", "DROP"},
			Insert: []string{"-I", "DOCKER-USER", "1", "-d", d.dest, "-j", "DROP"},
		})
	}
	return rules
}

// DockerHostIsLocal reports whether the daemon is the local socket. A remote
// DOCKER_HOST cannot have its filter table changed from this process.
func DockerHostIsLocal(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return true
	}
	return strings.HasPrefix(host, "unix://")
}

// ApplyMetadataGuard inserts each drop rule when it is not already present.
// A missing rule (iptables -C failure) is expected and is followed by an insert.
func ApplyMetadataGuard(ctx context.Context, run CommandRunner) error {
	if run == nil {
		run = ExecRunner{}
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
	}
	var failed []string
	for _, rule := range MetadataDropRules() {
		if err := run.Run(ctx, rule.Bin, rule.Check...); err == nil {
			continue
		}
		if err := run.Run(ctx, rule.Bin, rule.Insert...); err != nil {
			failed = append(failed, fmt.Sprintf("%s %s: %v", rule.Bin, rule.Dest, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("sandbox metadata guard: %s", strings.Join(failed, "; "))
	}
	return nil
}
