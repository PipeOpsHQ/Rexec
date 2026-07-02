package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AgentCLI describes an AI coding CLI that a terminal can auto-launch into.
//
// The set of agents is a fixed allowlist: the terminal WebSocket accepts an
// "agent" query param whose value MUST match one of these IDs. The launch
// Command strings are hardcoded here (never derived from user input), so the
// query param can never be used to inject arbitrary shell commands into a
// container.
type AgentCLI struct {
	ID          string `json:"id"`          // stable identifier used in the ?agent= query param
	Label       string `json:"label"`       // human-friendly name shown in the picker
	Command     string `json:"command"`     // command invoked inside the container
	Description string `json:"description"` // short blurb for the picker
	InstallHint string `json:"installHint"` // shown in-terminal when the command is missing
}

// agentCLIRegistry is the allowlist of AI CLIs a terminal can boot into.
// IDs mirror the presets already installed by the "ai"/coding container roles
// (see internal/container/roles.go) so an embedded terminal finds them on PATH.
var agentCLIRegistry = []AgentCLI{
	{
		ID:          "claude",
		Label:       "Claude Code",
		Command:     "claude",
		Description: "Anthropic's Claude Code agent",
		InstallHint: "npm install -g @anthropic-ai/claude-code",
	},
	{
		ID:          "codex",
		Label:       "Codex",
		Command:     "codex",
		Description: "OpenAI Codex CLI",
		InstallHint: "npm install -g @openai/codex",
	},
	{
		ID:          "gemini",
		Label:       "Gemini CLI",
		Command:     "gemini",
		Description: "Google Gemini CLI",
		InstallHint: "npm install -g @google/gemini-cli",
	},
	{
		ID:          "aider",
		Label:       "Aider",
		Command:     "aider",
		Description: "AI pair programming in your terminal",
		InstallHint: "pip install aider-chat",
	},
	{
		ID:          "opencode",
		Label:       "opencode",
		Command:     "opencode",
		Description: "Open-source terminal coding agent",
		InstallHint: "curl -fsSL https://opencode.ai/install | bash",
	},
	{
		ID:          "amp",
		Label:       "Sourcegraph Amp",
		Command:     "amp",
		Description: "Sourcegraph's agentic coding tool",
		InstallHint: "npm install -g @sourcegraph/amp",
	},
}

// lookupAgentCLI resolves an agent ID from the allowlist. The lookup is
// case-insensitive and tolerant of surrounding whitespace. Returns false for
// unknown IDs so callers can safely ignore untrusted query params.
func lookupAgentCLI(id string) (AgentCLI, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return AgentCLI{}, false
	}
	for _, a := range agentCLIRegistry {
		if a.ID == id {
			return a, true
		}
	}
	return AgentCLI{}, false
}

// buildAgentLaunchScript returns the argument for `<shell> -c` that auto-runs
// the agent CLI and then drops the user into an interactive login shell.
//
// Behavior:
//   - If the CLI is on PATH, it runs; on exit (or Ctrl-C) the user lands in a
//     normal login shell with full scrollback preserved.
//   - If the CLI is missing, a friendly install hint is printed instead of a
//     bare "command not found", then the user still gets a shell.
func buildAgentLaunchScript(shell string, def AgentCLI) string {
	// Fall back to sh for the trailing interactive shell if none was detected.
	if shell == "" {
		shell = "/bin/sh"
	}
	return fmt.Sprintf(
		"if command -v %s >/dev/null 2>&1; then %s; else "+
			"printf '\\033[33m%s is not installed in this sandbox.\\033[0m\\n'; "+
			"printf 'Install it with: \\033[36m%s\\033[0m\\n'; fi; "+
			"exec %s -l",
		def.Command, def.Command, def.Label, def.InstallHint, shell,
	)
}

// ListAgentCLIs returns the allowlist of AI CLIs the terminal can boot into.
// The frontend uses this to populate the create-terminal picker.
func (h *TerminalHandler) ListAgentCLIs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"agents": agentCLIRegistry})
}
