package mcp

import (
	"context"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Instructions are sent to the client on initialize and are meant for the model.
const Instructions = `You control Rexec sandboxes: isolated Linux machines for the signed-in user. Run installs, builds, and untrusted commands in a sandbox.

Call whoami if you are unsure the token works. Call list_sandboxes before creating another sandbox, and reuse one that is already running when it fits the task. create_sandbox uses image ubuntu when you omit image, template_id, and snapshot_id, and it waits until status is running. list_images returns other aliases such as debian and alpine.

exec runs one non-interactive shell command and returns stdout, stderr, and exit_code. A non-zero exit_code is the command result, not a broken tool. read_file and write_file take absolute paths inside the sandbox, usually under /home/user. list_files lists a directory.

The default network allows outbound traffic and does not accept inbound connections. network_mode none disables outbound network as well.

delete_sandbox is permanent. Delete a sandbox when the user asked you to, or when you created it for a one-off task and the work is finished.

If a tool says unauthorized, ask the user to run rexec login or set REXEC_TOKEN. Do not ask them to paste the token into the chat.`

// pollInterval is how often create_sandbox and wait_running check status.
var pollInterval = 1500 * time.Millisecond

// NewServer builds an MCP server whose tools call api.
func NewServer(api *Client, version string) *sdk.Server {
	if version == "" {
		version = "dev"
	}
	s := sdk.NewServer(&sdk.Implementation{
		Name:    "rexec",
		Version: version,
		Title:   "Rexec sandboxes",
	}, &sdk.ServerOptions{
		Instructions: Instructions,
	})
	registerTools(s, api)
	s.AddPrompt(&sdk.Prompt{
		Name:        "run_in_sandbox",
		Description: "Steps for running a task inside a Rexec sandbox.",
	}, func(_ context.Context, _ *sdk.GetPromptRequest) (*sdk.GetPromptResult, error) {
		return &sdk.GetPromptResult{
			Description: "Run work in a Rexec sandbox",
			Messages: []*sdk.PromptMessage{{
				Role: "user",
				Content: &sdk.TextContent{
					Text: "Use the Rexec tools. list_sandboxes first and reuse a running sandbox when it fits. Otherwise create_sandbox and wait until it is running. Use exec for commands and read_file or write_file for source under /home/user. Check exit_code. delete_sandbox only when the sandbox is disposable or the user asked you to remove it.",
				},
			}},
		}, nil
	})
	return s
}

func boolPtr(v bool) *bool { return &v }

func readOnly() *sdk.ToolAnnotations {
	return &sdk.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPtr(false)}
}

func additive(idempotent bool) *sdk.ToolAnnotations {
	return &sdk.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: boolPtr(false), IdempotentHint: idempotent}
}

func destructive(idempotent bool) *sdk.ToolAnnotations {
	return &sdk.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: boolPtr(true), IdempotentHint: idempotent}
}

func tool(name, title, description string, ann *sdk.ToolAnnotations) *sdk.Tool {
	return &sdk.Tool{
		Name:        name,
		Title:       title,
		Description: description,
		Annotations: ann,
	}
}
