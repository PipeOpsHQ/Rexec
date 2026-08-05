# @pipeops/rexec-mcp

**Model Context Protocol (MCP)** server for [Rexec](https://rexec.sh) — give AI agents tools to create sandboxes, run commands, list files, and manage templates.

## Install

```bash
# From monorepo (dev)
cd sdk/js && npm run build && cd ../mcp && npm install && npm run build

# Or publish and use
npx -y @pipeops/rexec-mcp
```

## Environment

| Variable | Required | Description |
|----------|----------|-------------|
| `REXEC_TOKEN` | **yes** | API bearer token |
| `REXEC_URL` | no | Base URL (default `https://rexec.sh`) |

## Tools

| Tool | Description |
|------|-------------|
| `list_sandboxes` | List user sandboxes |
| `create_sandbox` | Create (image / template_id / network_mode); waits until running by default |
| `get_sandbox` | Get status |
| `delete_sandbox` | Destroy sandbox |
| `exec` | Non-interactive shell command (`stdout` / `exit_code`) |
| `list_files` | List directory |
| `mkdir` | Create directory |
| `list_templates` | List committed templates |
| `create_template` | Commit running sandbox → template |
| `delete_template` | Remove template |
| `wait_running` | Poll until running |

## Claude Desktop / Cursor

```json
{
  "mcpServers": {
    "rexec": {
      "command": "npx",
      "args": ["-y", "@pipeops/rexec-mcp"],
      "env": {
        "REXEC_URL": "https://rexec.sh",
        "REXEC_TOKEN": "your-api-token"
      }
    }
  }
}
```

Local monorepo:

```json
{
  "mcpServers": {
    "rexec": {
      "command": "node",
      "args": ["/path/to/rexec/sdk/mcp/dist/index.js"],
      "env": {
        "REXEC_URL": "http://localhost:8080",
        "REXEC_TOKEN": "..."
      }
    }
  }
}
```

## Agent workflow example

1. `create_sandbox` with `image: "ubuntu"`
2. `exec` install deps / run tests
3. `create_template` from that sandbox
4. Later: `create_sandbox` with `template_id` for warm starts

## Related

- REST / SDK: [docs/SDK.md](../../docs/SDK.md)
- Package: `pipeops-rexec` (JS client used under the hood)
