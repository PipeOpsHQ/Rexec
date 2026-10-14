package container

import (
	"context"
	"os"
	"strings"
	"testing"
)

type fakeRunner struct {
	calls []string
	fail  map[string]error
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	line := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, line)
	if err, ok := f.fail[line]; ok {
		return err
	}
	return nil
}

func TestMetadataDropRulesCoverCloudMetadata(t *testing.T) {
	want := []string{
		"169.254.169.254/32",
		"169.254.169.253/32",
		"169.254.170.2/32",
		"168.63.129.16/32",
		"fd00:ec2::254/128",
	}
	got := map[string]string{}
	for _, rule := range MetadataDropRules() {
		got[rule.Dest] = rule.Bin
		if !strings.Contains(strings.Join(rule.Check, " "), "DOCKER-USER") || rule.Check[0] != "-C" || rule.Check[len(rule.Check)-1] != "DROP" {
			t.Fatalf("check args = %v", rule.Check)
		}
		if rule.Insert[0] != "-I" || rule.Insert[len(rule.Insert)-1] != "DROP" {
			t.Fatalf("insert args = %v", rule.Insert)
		}
	}
	for _, dest := range want {
		if got[dest] == "" {
			t.Errorf("missing drop for %s", dest)
		}
	}
	if got["fd00:ec2::254/128"] != "ip6tables" {
		t.Fatalf("ipv6 rule bin = %s", got["fd00:ec2::254/128"])
	}
}

func TestApplyMetadataGuardSkipsExistingRules(t *testing.T) {
	run := &fakeRunner{}
	if err := ApplyMetadataGuard(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	for _, call := range run.calls {
		if strings.Contains(call, " -I ") {
			t.Fatalf("inserted a rule that was already present: %s", call)
		}
	}
	if len(run.calls) != len(MetadataDropRules()) {
		t.Fatalf("calls = %d, want %d", len(run.calls), len(MetadataDropRules()))
	}
}

func TestApplyMetadataGuardInsertsMissingRules(t *testing.T) {
	run := &fakeRunner{fail: map[string]error{}}
	for _, rule := range MetadataDropRules() {
		run.fail[rule.Bin+" "+strings.Join(rule.Check, " ")] = errMissingRule
	}
	if err := ApplyMetadataGuard(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	inserts := 0
	for _, call := range run.calls {
		if strings.Contains(call, " -I ") {
			inserts++
		}
	}
	if inserts != len(MetadataDropRules()) {
		t.Fatalf("inserts = %d, want %d", inserts, len(MetadataDropRules()))
	}
}

func TestApplyMetadataGuardReportsInsertFailure(t *testing.T) {
	run := &fakeRunner{fail: map[string]error{}}
	for _, rule := range MetadataDropRules() {
		run.fail[rule.Bin+" "+strings.Join(rule.Check, " ")] = errMissingRule
		run.fail[rule.Bin+" "+strings.Join(rule.Insert, " ")] = errMissingRule
	}
	err := ApplyMetadataGuard(context.Background(), run)
	if err == nil || !strings.Contains(err.Error(), "169.254.169.254/32") {
		t.Fatalf("err = %v", err)
	}
}

func TestDockerHostIsLocal(t *testing.T) {
	if !DockerHostIsLocal("") || !DockerHostIsLocal("unix:///var/run/docker.sock") {
		t.Fatal("local socket should install the guard here")
	}
	if DockerHostIsLocal("tcp://10.0.0.5:2376") {
		t.Fatal("remote daemon cannot be filtered from this process")
	}
}

func TestHostScriptsDropTheSameDestinations(t *testing.T) {
	files := []string{
		"../../scripts/block-sandbox-metadata.sh",
		"../../docker/docker-host/setup.sh",
	}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(body)
		for _, rule := range MetadataDropRules() {
			if !strings.Contains(text, rule.Dest) {
				t.Errorf("%s does not drop %s", file, rule.Dest)
			}
		}
	}
}

type missingRuleError struct{}

func (missingRuleError) Error() string { return "rule missing" }

var errMissingRule missingRuleError
