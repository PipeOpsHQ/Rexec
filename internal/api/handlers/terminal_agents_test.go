package handlers

import (
	"strings"
	"testing"
)

func TestLookupAgentCLI(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID string
		wantOK bool
	}{
		{"known claude", "claude", "claude", true},
		{"case insensitive", "Claude", "claude", true},
		{"trims whitespace", "  gemini  ", "gemini", true},
		{"codex", "codex", "codex", true},
		{"empty", "", "", false},
		{"unknown", "rm-rf", "", false},
		// Ensure the query param can't smuggle a shell command through.
		{"injection attempt", "claude; rm -rf /", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, ok := lookupAgentCLI(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("lookupAgentCLI(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && def.ID != tt.wantID {
				t.Fatalf("lookupAgentCLI(%q) id = %q, want %q", tt.input, def.ID, tt.wantID)
			}
		})
	}
}

func TestBuildAgentLaunchScript(t *testing.T) {
	def, ok := lookupAgentCLI("claude")
	if !ok {
		t.Fatal("expected claude to be in the allowlist")
	}

	script := buildAgentLaunchScript("/bin/bash", def)

	// Must run the CLI command...
	if !strings.Contains(script, def.Command) {
		t.Errorf("script missing agent command %q: %s", def.Command, script)
	}
	// ...guard on its presence so a missing CLI shows a hint, not just an error...
	if !strings.Contains(script, "command -v") {
		t.Errorf("script should guard on command availability: %s", script)
	}
	// ...and always drop the user into an interactive login shell afterwards.
	if !strings.Contains(script, "exec /bin/bash -l") {
		t.Errorf("script should exec a login shell after the agent: %s", script)
	}

	// Empty shell falls back to /bin/sh so we never exec an empty command.
	fallback := buildAgentLaunchScript("", def)
	if !strings.Contains(fallback, "exec /bin/sh -l") {
		t.Errorf("empty shell should fall back to /bin/sh: %s", fallback)
	}
}
