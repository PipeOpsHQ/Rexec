# Rexec Agent Documentation

The Rexec Agent allows you to connect any server or local machine to Rexec as a terminal. This enables remote terminal access through the Rexec dashboard without exposing SSH ports or managing complex network configurations.

## Getting Started

The easiest way to set up an agent is through the Rexec dashboard:

1. Go to **[Settings → Agents](https://rexec.pipeops.io/settings)** in your Rexec dashboard
2. Click **"Add Agent"** to create a new agent
3. Copy the generated command (includes your token and agent configuration)
4. Paste and run the command on your server

The dashboard provides a ready-to-use command that handles registration automatically.

## Overview

The agent system consists of two components:

1. **rexec-agent** - A standalone agent binary that runs on your server and maintains a persistent connection to Rexec
2. **rexec agent** - Agent commands integrated into the rexec-cli for unified management

## Quick Start

> **Tip:** For the fastest setup, use the [Settings page](https://rexec.pipeops.io/settings) to generate a ready-to-use install command with your token pre-configured.

### 1. Install the Agent

```bash
# Using the install script
curl -sSL https://rexec.pipeops.io/install-agent.sh | bash

# Or download directly
wget https://github.com/rexec/rexec/releases/latest/download/rexec-agent-linux-amd64 -O /usr/local/bin/rexec-agent
chmod +x /usr/local/bin/rexec-agent
```

### 2. Register Your Server

```bash
# Interactive registration
rexec-agent register

# Or with options
rexec-agent register --name "prod-server-1" --tags "production,aws,us-east"
```

### 3. Start the Agent

```bash
# Foreground (for testing)
rexec-agent start

# Background daemon
rexec-agent start --daemon

# Or install as a system service
sudo rexec-agent install
```

## Agent Commands

### rexec-agent register

Register this machine as a Rexec terminal.

```bash
rexec-agent register [options]
```

**Options:**
| Option | Short | Description |
|--------|-------|-------------|
| `--name` | `-n` | Name for this terminal (default: hostname) |
| `--description` | `-d` | Description of this server |
| `--shell` | `-s` | Shell to use (default: $SHELL or /bin/bash) |
| `--tags` | `-t` | Comma-separated tags for organization |
| `--token` | | Auth token (or set REXEC_TOKEN env var) |
| `--host` | | API host (default: https://rexec.pipeops.io) |

**Examples:**

```bash
# Interactive registration (prompts for token)
rexec-agent register --name "my-server"

# With all options
rexec-agent register \
  --name "prod-api-1" \
  --description "Production API server" \
  --shell /bin/zsh \
  --tags "production,api,aws" \
  --token "your-api-token"

# Using environment variable for token
export REXEC_TOKEN="your-api-token"
rexec-agent register --name "staging-server"
```

### rexec-agent start

Start the agent and connect to Rexec.

```bash
rexec-agent start [options]
```

**Options:**
| Option | Short | Description |
|--------|-------|-------------|
| `--daemon` | `-d` | Run in background |

**Examples:**

```bash
# Run in foreground (useful for testing/debugging)
rexec-agent start

# Run as background daemon
rexec-agent start --daemon
```

### rexec-agent stop

Stop a running agent.

```bash
rexec-agent stop
```

### rexec-agent status

Show the current status of the agent.

```bash
rexec-agent status
```

**Output:**

```
Agent Status
─────────────────────────────────────────
  Registered: ✓
  ID:         abc123-def456-...
  Name:       prod-server-1
  Host:       https://rexec.pipeops.io
  Shell:      /bin/bash
  Running:    ✓ Yes
```

### rexec-agent unregister

Remove this machine from Rexec.

```bash
rexec-agent unregister
```

### rexec-agent install

Install the agent as a systemd service (Linux only).

```bash
sudo rexec-agent install
```

This creates a systemd service that:

- Starts automatically on boot
- Restarts on failure
- Manages credentials securely

**Managing the service:**

```bash
# Check status
sudo systemctl status rexec-agent

# Stop the agent
sudo systemctl stop rexec-agent

# Restart the agent
sudo systemctl restart rexec-agent

# View logs
sudo journalctl -u rexec-agent -f
```

## Environment Variables

| Variable       | Description                                 |
| -------------- | ------------------------------------------- |
| `REXEC_TOKEN`  | Authentication token (overrides config)     |
| `REXEC_HOST`   | API host URL (overrides config)             |
| `REXEC_API`    | API host URL (alias of `REXEC_HOST`)        |
| `REXEC_CONFIG` | Config file path (overrides default search) |

## Configuration Files

The agent loads configuration from:

- `/etc/rexec/agent.yaml` (system-wide installs), or
- `~/.rexec/agent.json` (user installs)

Override the path with `--config` or `REXEC_CONFIG`.

Example (`/etc/rexec/agent.yaml`):

```yaml
api_url: https://rexec.pipeops.io
token: rexec_...
agent_id: abc123-def456-...
name: prod-server-1
labels:
  - production
  - aws
shell: /bin/bash
working_dir: /root
```

## How It Works

1. **Registration**: The agent registers with the Rexec API, providing system information (OS, architecture, shell)

2. **Connection**: The agent establishes a WebSocket connection to Rexec's agent endpoint

3. **Shell Sessions**: When you connect to the terminal from the dashboard:
   - Rexec sends a `shell_start` message
   - The agent spawns a PTY with your configured shell
   - Input/output is proxied through the WebSocket

4. **Reconnection**: If the connection drops, the agent automatically reconnects with exponential backoff

## Security Considerations

- **Token Security**: Store tokens securely; use environment variables in production
- **Shell Access**: The agent provides full shell access to your configured shell
- **Network**: All communication is encrypted via WSS (WebSocket Secure)
- **Firewall**: No inbound ports required; agent initiates outbound connection

## Troubleshooting

### Agent won't connect

```bash
# Check agent status
rexec-agent status

# Run in foreground to see errors
rexec-agent start

# Verify token is valid
curl -H "Authorization: Bearer $REXEC_TOKEN" https://rexec.pipeops.io/api/profile
```

### Rotate an API token

If you see repeated `401` reconnect errors, update the agent to use a long-lived API token:

```bash
rexec-agent refresh-token
sudo systemctl restart rexec-agent
```

### Connection keeps dropping

- Check network stability
- Ensure no firewall blocking WebSocket connections
- Check server logs: `journalctl -u rexec-agent`

### Shell not working correctly

- Verify the shell path exists: `which bash`
- Check shell permissions
- Try specifying a different shell: `rexec-agent register --shell /bin/sh`

## Use Cases

### 1. Remote Server Management

Connect production servers for quick debugging:

```bash
rexec-agent register --name "prod-db-1" --tags "production,database,postgres"
```

### 2. Development Environment

Connect your local machine for remote access:

```bash
rexec-agent register --name "macbook-dev" --description "Personal dev machine"
```

### 3. CI/CD Runners

Connect CI runners for debugging failed builds:

```bash
rexec-agent register --name "ci-runner-1" --tags "ci,github-actions"
```

### 4. IoT/Edge Devices

Connect Raspberry Pi or edge devices:

```bash
rexec-agent register --name "rpi-sensor-hub" --shell /bin/sh
```

## Comparison with Containers

| Feature     | Agent                      | Container                 |
| ----------- | -------------------------- | ------------------------- |
| Setup       | Install on existing server | Provisioned by Rexec      |
| Resources   | Uses host resources        | Dedicated container       |
| Persistence | Full disk access           | Ephemeral or volume-based |
| Network     | Host network               | Isolated network          |
| Use Case    | Existing servers           | Development environments  |

## API Reference

### Register Agent

```
POST /api/agents/register

{
  "name": "server-name",
  "description": "Server description",
  "type": "agent",
  "os": "linux",
  "arch": "amd64",
  "shell": "/bin/bash",
  "tags": ["tag1", "tag2"]
}
```

### Unregister Agent

```
DELETE /api/agents/{agent_id}
```

### WebSocket Connection

```
WSS /ws/agent/{agent_id}?token={token}

Headers:
  X-Agent-Name: server-name
  X-Agent-OS: linux
  X-Agent-Arch: amd64
  X-Agent-Shell: /bin/bash
```

### WebSocket Messages

**From Server:**

- `shell_start` - Start a new shell session
- `shell_input` - Send input to shell
- `shell_resize` - Resize terminal
- `shell_stop` - Stop shell session
- `ping` - Keep-alive ping
- `exec` - Execute a command

**From Agent:**

- `shell_started` - Shell session started
- `shell_output` - Shell output data
- `shell_stopped` - Shell session ended
- `shell_error` - Error message
- `pong` - Keep-alive response
- `exec_result` - Command execution result
