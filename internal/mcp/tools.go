package mcp

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerTools(s *sdk.Server, api *Client) {
	sdk.AddTool(s, tool("whoami", "Who am I",
		"Show the Rexec account for the configured token: id, username, email, and tier.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		p, err := api.Profile(ctx)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(p)
	})

	sdk.AddTool(s, tool("list_sandboxes", "List sandboxes",
		"List sandboxes for the signed-in user, including id, name, image, and status. Call this before creating a new sandbox.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		list, err := api.ListSandboxes(ctx)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"count": len(list), "sandboxes": list})
	})

	sdk.AddTool(s, tool("get_sandbox", "Get sandbox",
		"Get one sandbox by id, including its status (creating, running, stopped, or error).",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, in idArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.SandboxID)
		if err != nil {
			return fail(err)
		}
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		sb, err := api.GetSandbox(ctx, id)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(sb)
	})

	sdk.AddTool(s, tool("list_images", "List images",
		"List sandbox image aliases you can pass to create_sandbox, such as ubuntu, debian, and alpine. Set all to true to include less common aliases.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, in listImagesArgs) (*sdk.CallToolResult, any, error) {
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		images, err := api.ListImages(ctx, in.All)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"count": len(images), "images": images})
	})

	sdk.AddTool(s, tool("create_sandbox", "Create sandbox",
		"Create a sandbox. Pass an image alias, or template_id, or snapshot_id. Defaults to ubuntu. Waits until status is running unless wait_running is false. network_mode none disables outbound network.",
		additive(false)), func(ctx context.Context, _ *sdk.CallToolRequest, in createArgs) (*sdk.CallToolResult, any, error) {
		body, err := in.body()
		if err != nil {
			return fail(err)
		}
		wait := in.wait()
		timeout := clamp(in.TimeoutSeconds, 120, 5, 600)
		ctx, cancel := withTimeout(ctx, time.Duration(timeout+20)*time.Second)
		defer cancel()
		sb, err := api.CreateSandbox(ctx, body)
		if err != nil {
			return toolErr(err)
		}
		if !wait {
			return jsonResult(sb)
		}
		ready, err := waitRunning(ctx, api, sb.ID, time.Duration(timeout)*time.Second)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(ready)
	})

	sdk.AddTool(s, tool("wait_running", "Wait until running",
		"Poll a sandbox until status is running, or until it errors or the timeout elapses.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, in waitArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.SandboxID)
		if err != nil {
			return fail(err)
		}
		timeout := clamp(in.TimeoutSeconds, 120, 5, 600)
		ctx, cancel := withTimeout(ctx, time.Duration(timeout+20)*time.Second)
		defer cancel()
		sb, err := waitRunning(ctx, api, id, time.Duration(timeout)*time.Second)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(sb)
	})

	sdk.AddTool(s, tool("start_sandbox", "Start sandbox",
		"Start a stopped sandbox.",
		additive(true)), func(ctx context.Context, _ *sdk.CallToolRequest, in idArgs) (*sdk.CallToolResult, any, error) {
		return sandboxAction(ctx, in.SandboxID, api.StartSandbox)
	})

	sdk.AddTool(s, tool("stop_sandbox", "Stop sandbox",
		"Stop a running sandbox. The filesystem is kept. Use delete_sandbox to remove it.",
		additive(true)), func(ctx context.Context, _ *sdk.CallToolRequest, in idArgs) (*sdk.CallToolResult, any, error) {
		return sandboxAction(ctx, in.SandboxID, api.StopSandbox)
	})

	sdk.AddTool(s, tool("delete_sandbox", "Delete sandbox",
		"Permanently delete a sandbox and its filesystem. Use this when the user asked you to remove it, or when you created it for a finished one-off task.",
		destructive(true)), func(ctx context.Context, _ *sdk.CallToolRequest, in idArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.SandboxID)
		if err != nil {
			return fail(err)
		}
		ctx, cancel := withTimeout(ctx, 60*time.Second)
		defer cancel()
		if err := api.DeleteSandbox(ctx, id); err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"deleted": true, "sandbox_id": id})
	})

	sdk.AddTool(s, tool("exec", "Run command",
		"Run one non-interactive shell command in a running sandbox. Returns stdout, stderr, and exit_code. Prefer this over an interactive terminal. A non-zero exit_code means the command failed.",
		destructive(false)), func(ctx context.Context, _ *sdk.CallToolRequest, in execArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.SandboxID)
		if err != nil {
			return fail(err)
		}
		req, err := in.request()
		if err != nil {
			return fail(err)
		}
		timeout := req.TimeoutSeconds
		if timeout <= 0 {
			timeout = 60
		}
		ctx, cancel := withTimeout(ctx, time.Duration(timeout+15)*time.Second)
		defer cancel()
		result, err := api.Exec(ctx, id, req)
		if err != nil {
			return toolErr(err)
		}
		return textResult(formatExec(result))
	})

	sdk.AddTool(s, tool("list_files", "List files",
		"List files in a directory inside a running sandbox. path defaults to /home/user.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, in pathArgs) (*sdk.CallToolResult, any, error) {
		id, dir, err := sandboxPath(in.SandboxID, in.Path, "/home/user")
		if err != nil {
			return fail(err)
		}
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		files, err := api.ListFiles(ctx, id, dir)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"path": dir, "count": len(files), "files": files})
	})

	sdk.AddTool(s, tool("read_file", "Read file",
		"Read a text file from a running sandbox. path is absolute, for example /home/user/main.py. Returns at most 256KB. Binary files are reported and not inlined.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, in pathArgs) (*sdk.CallToolResult, any, error) {
		id, filePath, err := sandboxPath(in.SandboxID, in.Path, "")
		if err != nil {
			return fail(err)
		}
		if filePath == "/" {
			return fail(errors.New("path is a directory. Use list_files"))
		}
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		data, truncated, err := api.ReadFile(ctx, id, filePath)
		if err != nil {
			return toolErr(err)
		}
		if !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
			return textResult(fmt.Sprintf("path: %s\nbytes: %d\nbinary: true\nThis file is not UTF-8 text. Use exec if you need a bounded inspection.", filePath, len(data)))
		}
		var b strings.Builder
		fmt.Fprintf(&b, "path: %s\nbytes: %d\n", filePath, len(data))
		if truncated {
			b.WriteString("truncated: true\nThe file is longer than 256KB. Use exec to read a smaller slice.\n")
		}
		b.WriteString("---\n")
		b.Write(data)
		return textResult(b.String())
	})

	sdk.AddTool(s, tool("write_file", "Write file",
		"Write a UTF-8 file in a running sandbox. path is the absolute file path, for example /home/user/src/main.py. Overwrites an existing file. Content is limited to 512KB. Parent directories must already exist; call mkdir first.",
		destructive(false)), func(ctx context.Context, _ *sdk.CallToolRequest, in writeArgs) (*sdk.CallToolResult, any, error) {
		id, filePath, err := sandboxPath(in.SandboxID, in.Path, "")
		if err != nil {
			return fail(err)
		}
		if filePath == "/" || strings.HasSuffix(in.Path, "/") {
			return fail(errors.New("path must be a file, not a directory"))
		}
		if err := refuseSystemPath(filePath); err != nil {
			return fail(err)
		}
		content := []byte(in.Content)
		if len(content) > maxWriteBytes {
			return fail(fmt.Errorf("content is %d bytes. The limit is %d", len(content), maxWriteBytes))
		}
		dir, name := path.Split(filePath)
		dir = strings.TrimRight(dir, "/")
		if dir == "" {
			dir = "/"
		}
		ctx, cancel := withTimeout(ctx, 60*time.Second)
		defer cancel()
		if err := api.WriteFile(ctx, id, dir, name, content); err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"ok": true, "path": filePath, "bytes": len(content)})
	})

	sdk.AddTool(s, tool("mkdir", "Create directory",
		"Create a directory in a running sandbox, including parents. path is absolute.",
		additive(true)), func(ctx context.Context, _ *sdk.CallToolRequest, in pathArgs) (*sdk.CallToolResult, any, error) {
		id, dir, err := sandboxPath(in.SandboxID, in.Path, "")
		if err != nil {
			return fail(err)
		}
		if err := refuseSystemPath(dir); err != nil {
			return fail(err)
		}
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := api.Mkdir(ctx, id, dir); err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"ok": true, "path": dir})
	})

	sdk.AddTool(s, tool("delete_file", "Delete file",
		"Delete a file or directory inside a running sandbox. path is absolute. Refuses filesystem roots such as / and /etc.",
		destructive(true)), func(ctx context.Context, _ *sdk.CallToolRequest, in pathArgs) (*sdk.CallToolResult, any, error) {
		id, filePath, err := sandboxPath(in.SandboxID, in.Path, "")
		if err != nil {
			return fail(err)
		}
		if err := refuseSystemPath(filePath); err != nil {
			return fail(err)
		}
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := api.DeleteFile(ctx, id, filePath); err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"deleted": true, "path": filePath})
	})

	sdk.AddTool(s, tool("list_templates", "List templates",
		"List saved sandbox templates (committed images) you can pass as template_id to create_sandbox.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		list, err := api.ListTemplates(ctx)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"count": len(list), "templates": list})
	})

	sdk.AddTool(s, tool("create_template", "Create template",
		"Commit a running sandbox into a reusable template. Later create_sandbox calls can pass template_id.",
		additive(false)), func(ctx context.Context, _ *sdk.CallToolRequest, in templateArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.FromSandboxID)
		if err != nil {
			return fail(err)
		}
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return fail(errors.New("name is required"))
		}
		ctx, cancel := withTimeout(ctx, 120*time.Second)
		defer cancel()
		tpl, err := api.CreateTemplate(ctx, name, id, strings.TrimSpace(in.Description))
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(tpl)
	})

	sdk.AddTool(s, tool("delete_template", "Delete template",
		"Delete a sandbox template by id. Sandboxes already created from it are kept.",
		destructive(true)), func(ctx context.Context, _ *sdk.CallToolRequest, in templateIDArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.TemplateID)
		if err != nil {
			return fail(err)
		}
		ctx, cancel := withTimeout(ctx, 60*time.Second)
		defer cancel()
		if err := api.DeleteTemplate(ctx, id); err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"deleted": true, "template_id": id})
	})

	sdk.AddTool(s, tool("list_snapshots", "List snapshots",
		"List filesystem snapshots for the signed-in user. Pass a snapshot id to create_sandbox as snapshot_id.",
		readOnly()), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		ctx, cancel := withTimeout(ctx, 30*time.Second)
		defer cancel()
		list, err := api.ListSnapshots(ctx)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(map[string]any{"count": len(list), "snapshots": list})
	})

	sdk.AddTool(s, tool("snapshot_sandbox", "Snapshot sandbox",
		"Create a point-in-time filesystem snapshot of a sandbox. Use the returned id as snapshot_id when creating a sandbox.",
		additive(false)), func(ctx context.Context, _ *sdk.CallToolRequest, in snapshotArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.SandboxID)
		if err != nil {
			return fail(err)
		}
		ctx, cancel := withTimeout(ctx, 120*time.Second)
		defer cancel()
		snap, err := api.CreateSnapshot(ctx, id, strings.TrimSpace(in.Name), strings.TrimSpace(in.Description))
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(snap)
	})

	sdk.AddTool(s, tool("fork_sandbox", "Fork sandbox",
		"Copy a sandbox filesystem into a new sandbox. The source sandbox is left in place.",
		additive(false)), func(ctx context.Context, _ *sdk.CallToolRequest, in forkArgs) (*sdk.CallToolResult, any, error) {
		id, err := cleanID(in.SandboxID)
		if err != nil {
			return fail(err)
		}
		body := map[string]any{}
		if name := strings.TrimSpace(in.Name); name != "" {
			body["name"] = name
		}
		if mode := strings.TrimSpace(in.NetworkMode); mode != "" {
			if err := checkNetwork(mode); err != nil {
				return fail(err)
			}
			body["network_mode"] = mode
		}
		if in.SaveSnapshot {
			body["save_snapshot"] = true
		}
		ctx, cancel := withTimeout(ctx, 180*time.Second)
		defer cancel()
		fork, err := api.ForkSandbox(ctx, id, body)
		if err != nil {
			return toolErr(err)
		}
		return jsonResult(fork)
	})
}

type idArgs struct {
	SandboxID string `json:"sandbox_id" jsonschema:"sandbox id from list_sandboxes or create_sandbox"`
}

type listImagesArgs struct {
	All bool `json:"all,omitempty" jsonschema:"include less common image aliases"`
}

type createArgs struct {
	Image              string   `json:"image,omitempty" jsonschema:"image alias such as ubuntu, debian, or alpine. Omit when template_id or snapshot_id is set"`
	Name               string   `json:"name,omitempty" jsonschema:"optional display name"`
	TemplateID         string   `json:"template_id,omitempty" jsonschema:"create from a saved template"`
	SnapshotID         string   `json:"snapshot_id,omitempty" jsonschema:"create from a filesystem snapshot"`
	NetworkMode        string   `json:"network_mode,omitempty" jsonschema:"default, none, or restricted. none disables outbound network"`
	EgressAllow        []string `json:"egress_allow,omitempty" jsonschema:"extra egress hosts when network_mode is restricted"`
	Role               string   `json:"role,omitempty" jsonschema:"optional environment role such as python, node, or devops"`
	IdleTimeoutSeconds *int     `json:"idle_timeout_seconds,omitempty" jsonschema:"stop the sandbox after this many idle seconds"`
	MaxLifetimeSeconds *int     `json:"max_lifetime_seconds,omitempty" jsonschema:"hard lifetime in seconds from create time"`
	WaitRunning        *bool    `json:"wait_running,omitempty" jsonschema:"wait until status is running. Defaults to true"`
	TimeoutSeconds     int      `json:"timeout_seconds,omitempty" jsonschema:"seconds to wait for running, from 5 to 600. Default 120"`
}

func (in createArgs) wait() bool {
	if in.WaitRunning == nil {
		return true
	}
	return *in.WaitRunning
}

func (in createArgs) body() (map[string]any, error) {
	if err := checkNetwork(in.NetworkMode); err != nil {
		return nil, err
	}
	image := strings.TrimSpace(in.Image)
	templateID := strings.TrimSpace(in.TemplateID)
	snapshotID := strings.TrimSpace(in.SnapshotID)
	if image == "" && templateID == "" && snapshotID == "" {
		image = "ubuntu"
	}
	body := map[string]any{}
	if image != "" {
		body["image"] = image
	}
	if name := strings.TrimSpace(in.Name); name != "" {
		body["name"] = name
	}
	if templateID != "" {
		body["template_id"] = templateID
	}
	if snapshotID != "" {
		body["snapshot_id"] = snapshotID
	}
	if mode := strings.TrimSpace(in.NetworkMode); mode != "" {
		body["network_mode"] = mode
	}
	if len(in.EgressAllow) > 0 {
		body["egress_allow"] = in.EgressAllow
	}
	if role := strings.TrimSpace(in.Role); role != "" {
		body["role"] = role
	}
	if in.IdleTimeoutSeconds != nil && *in.IdleTimeoutSeconds > 0 {
		body["idle_timeout_seconds"] = *in.IdleTimeoutSeconds
	}
	if in.MaxLifetimeSeconds != nil && *in.MaxLifetimeSeconds > 0 {
		body["max_lifetime_seconds"] = *in.MaxLifetimeSeconds
	}
	return body, nil
}

type waitArgs struct {
	SandboxID      string `json:"sandbox_id" jsonschema:"sandbox id"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"seconds to wait, from 5 to 600. Default 120"`
}

type execArgs struct {
	SandboxID      string   `json:"sandbox_id" jsonschema:"sandbox id"`
	Command        string   `json:"command,omitempty" jsonschema:"shell command run with sh -c. Omit when cmd is set"`
	Cmd            []string `json:"cmd,omitempty" jsonschema:"argv vector. Used instead of command when set"`
	WorkDir        string   `json:"workdir,omitempty" jsonschema:"working directory inside the sandbox"`
	Env            []string `json:"env,omitempty" jsonschema:"extra environment entries as KEY=VALUE"`
	User           string   `json:"user,omitempty" jsonschema:"user to run as inside the sandbox"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty" jsonschema:"timeout in seconds, from 1 to 300. Default 60"`
}

func (in execArgs) request() (ExecRequest, error) {
	command := strings.TrimSpace(in.Command)
	if command == "" && len(in.Cmd) == 0 {
		return ExecRequest{}, errors.New("command or cmd is required")
	}
	timeout := in.TimeoutSeconds
	if timeout == 0 {
		timeout = 60
	}
	if timeout < 1 || timeout > 300 {
		return ExecRequest{}, errors.New("timeout_seconds must be from 1 to 300")
	}
	for _, e := range in.Env {
		if !strings.Contains(e, "=") || strings.HasPrefix(e, "=") {
			return ExecRequest{}, fmt.Errorf("env entry %q must be KEY=VALUE", e)
		}
	}
	return ExecRequest{
		Command:        command,
		Cmd:            in.Cmd,
		WorkDir:        strings.TrimSpace(in.WorkDir),
		Env:            in.Env,
		User:           strings.TrimSpace(in.User),
		TimeoutSeconds: timeout,
	}, nil
}

type pathArgs struct {
	SandboxID string `json:"sandbox_id" jsonschema:"sandbox id"`
	Path      string `json:"path,omitempty" jsonschema:"absolute path inside the sandbox"`
}

type writeArgs struct {
	SandboxID string `json:"sandbox_id" jsonschema:"sandbox id"`
	Path      string `json:"path" jsonschema:"absolute file path inside the sandbox"`
	Content   string `json:"content" jsonschema:"UTF-8 file contents. Maximum 512KB"`
}

type templateArgs struct {
	Name          string `json:"name" jsonschema:"template name, 1-63 characters, starting with a letter or digit"`
	FromSandboxID string `json:"from_sandbox_id" jsonschema:"running sandbox to commit"`
	Description   string `json:"description,omitempty" jsonschema:"optional description"`
}

type templateIDArgs struct {
	TemplateID string `json:"template_id" jsonschema:"template id"`
}

type snapshotArgs struct {
	SandboxID   string `json:"sandbox_id" jsonschema:"sandbox id"`
	Name        string `json:"name,omitempty" jsonschema:"optional snapshot name"`
	Description string `json:"description,omitempty" jsonschema:"optional description"`
}

type forkArgs struct {
	SandboxID    string `json:"sandbox_id" jsonschema:"sandbox to copy"`
	Name         string `json:"name,omitempty" jsonschema:"name for the new sandbox"`
	NetworkMode  string `json:"network_mode,omitempty" jsonschema:"default, none, or restricted"`
	SaveSnapshot bool   `json:"save_snapshot,omitempty" jsonschema:"also store the copied filesystem as a snapshot"`
}

func sandboxAction(ctx context.Context, rawID string, fn func(context.Context, string) (Sandbox, error)) (*sdk.CallToolResult, any, error) {
	id, err := cleanID(rawID)
	if err != nil {
		return fail(err)
	}
	ctx, cancel := withTimeout(ctx, 60*time.Second)
	defer cancel()
	sb, err := fn(ctx, id)
	if err != nil {
		return toolErr(err)
	}
	return jsonResult(sb)
}

func toolErr(err error) (*sdk.CallToolResult, any, error) {
	if err == nil {
		return fail(errors.New("unknown error"))
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, nil, err
	}
	return fail(err)
}

func waitRunning(ctx context.Context, api *Client, id string, timeout time.Duration) (Sandbox, error) {
	deadline := time.Now().Add(timeout)
	var last Sandbox
	for {
		sb, err := api.GetSandbox(ctx, id)
		if err != nil {
			return Sandbox{}, err
		}
		last = sb
		switch strings.ToLower(sb.Status) {
		case "running":
			return sb, nil
		case "error":
			return sb, fmt.Errorf("sandbox %s entered error state", id)
		}
		if time.Now().After(deadline) {
			return sb, fmt.Errorf("timed out waiting for sandbox %s (status %s)", id, sb.Status)
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return last, ctx.Err()
		case <-timer.C:
		}
	}
}

func clamp(v, def, min, max int) int {
	if v == 0 {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func checkNetwork(mode string) error {
	switch strings.TrimSpace(mode) {
	case "", "default", "none", "restricted":
		return nil
	default:
		return fmt.Errorf("network_mode must be default, none, or restricted")
	}
}

func cleanID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", errors.New("id is required")
	}
	if strings.ContainsAny(id, "/?# \t\r\n") || strings.Contains(id, "..") {
		return "", fmt.Errorf("invalid id %q", id)
	}
	if len(id) > 128 {
		return "", errors.New("id is too long")
	}
	return id, nil
}

func sandboxPath(sandboxID, raw, fallback string) (id, cleaned string, err error) {
	id, err = cleanID(sandboxID)
	if err != nil {
		return "", "", err
	}
	p := strings.TrimSpace(raw)
	if p == "" {
		p = fallback
	}
	if p == "" {
		return "", "", errors.New("path is required")
	}
	if strings.ContainsRune(p, 0) || strings.Contains(p, "\\") || !strings.HasPrefix(p, "/") {
		return "", "", errors.New("path must be absolute, for example /home/user/main.py")
	}
	if strings.Contains(p, "/../") || strings.HasSuffix(p, "/..") || p == "/.." {
		return "", "", errors.New("path must not contain ..")
	}
	cleaned = path.Clean(p)
	if !strings.HasPrefix(cleaned, "/") {
		return "", "", errors.New("path must be absolute, for example /home/user/main.py")
	}
	return id, cleaned, nil
}

func refuseSystemPath(p string) error {
	switch p {
	case "/", "/bin", "/sbin", "/usr", "/etc", "/lib", "/lib64", "/root", "/boot", "/proc", "/sys", "/var":
		return fmt.Errorf("refusing to change %s", p)
	}
	return nil
}
