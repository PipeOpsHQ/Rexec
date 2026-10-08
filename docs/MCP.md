# Rexec MCP server

`rexec-mcp` speaks the [Model Context Protocol](https://modelcontextprotocol.io) so an LLM client can use your Rexec account. It creates sandboxes, runs commands, and reads and writes files. It uses the same login as the CLI.

## Build

```bash
make mcp
# bin/rexec-mcp
```

The CLI subcommand runs the same server:

```bash
rexec mcp
```

## Credentials

Resolution order:

1. `~/.rexec/config.json` from `rexec login` (or `REXEC_CONFIG`)
2. `REXEC_TOKEN` and `REXEC_URL` (aliases: `REXEC_HOST`, `REXEC_API`)
3. `--token` and `--host`

The default API host, when nothing else is set, is `https://rexec.pipeops.io`.

## On a running Rexec server

The API serves the same tools at `/mcp`. A client sends the bearer token on each request. No extra process is required.

```json
{
  "mcpServers": {
    "rexec": {
      "url": "https://rexec.pipeops.io/mcp",
      "headers": {
        "Authorization": "Bearer your-api-token"
      }
    }
  }
}
```

For a server on your machine, use `http://127.0.0.1:8080/mcp`. The tools call the API on loopback with your token. Set `REXEC_INTERNAL_URL` when that loopback address is not `http://127.0.0.1:$PORT`.

## Claude Desktop / Cursor (stdio)

```json
{
  "mcpServers": {
    "rexec": {
      "command": "rexec-mcp",
      "env": {
        "REXEC_URL": "https://rexec.pipeops.io",
        "REXEC_TOKEN": "your-api-token"
      }
    }
  }
}
```

After `rexec login`, `env` can be omitted if the client can read your home directory config. Use an absolute path for `command` when the binary is not on the client's `PATH`.

## Streamable HTTP

```bash
rexec-mcp --http 127.0.0.1:8090
```

Point a remote MCP client at `http://127.0.0.1:8090`. The process still uses one token, from the environment or config file above.

## Tools

| Tool | What the model uses it for |
|------|----------------------------|
| `whoami` | Confirm the token and account |
| `list_sandboxes` | Reuse an existing sandbox |
| `create_sandbox` | Create one. Defaults to `ubuntu` and waits until it is running |
| `get_sandbox` / `wait_running` | Check status |
| `start_sandbox` / `stop_sandbox` | Start or stop without deleting |
| `delete_sandbox` | Permanently remove a sandbox |
| `exec` | Run one shell command. Returns stdout, stderr, and `exit_code` |
| `list_files` / `read_file` / `write_file` | Inspect and edit files under paths such as `/home/user` |
| `mkdir` / `delete_file` | Create or remove a path inside the sandbox |
| `list_images` | Image aliases for `create_sandbox` |
| `list_templates` / `create_template` / `delete_template` | Reusable committed images |
| `list_snapshots` / `snapshot_sandbox` / `fork_sandbox` | Copy a filesystem |

The server sends instructions on connect: reuse a running sandbox, treat a non-zero `exit_code` as the command's result, and delete a sandbox only when it is disposable or the user asked.

`read_file` returns at most 256KB of UTF-8 text. `write_file` accepts at most 512KB. `exec` responses are clipped to about 24KB so a noisy command does not fill the context window.

## Node package

[`sdk/mcp`](../sdk/mcp/) (`@pipeops/rexec-mcp`) is the npm stdio server. The Go binary is the one that shares the CLI config and includes file read and write.

## Tests

```bash
go test ./internal/mcp/
```
