package container

import (
	"strings"
	"testing"
)

func TestInSandboxCLIInstallScript(t *testing.T) {
	t.Parallel()

	script := inSandboxCLIInstallScript()
	for _, want := range []string{
		"#!/bin/sh",
		"cat > /root/.local/bin/rexec",
		"/usr/local/bin/rexec",
		"/usr/bin/rexec",
		"tools|ls",
		"show_tools",
		"rexec CLI ready",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("install script missing %q", want)
		}
	}

	if strings.Contains(script, "rexec login") {
		t.Error("in-sandbox helper must not be the host rexec-cli")
	}
}
