# Rexec CLI Documentation

The Rexec CLI (`rexec`) is a command-line interface for managing cloud terminal environments. It provides full access to all Rexec features including terminal management, snippets, and agent mode.

## Installation

### From GitHub Releases

```bash
# Linux (amd64)
curl -sSL https://github.com/rexec/rexec/releases/latest/download/rexec-linux-amd64 -o /usr/local/bin/rexec
chmod +x /usr/local/bin/rexec

# Linux (arm64)
curl -sSL https://github.com/rexec/rexec/releases/latest/download/rexec-linux-arm64 -o /usr/local/bin/rexec
chmod +x /usr/local/bin/rexec

# macOS (Intel)
curl -sSL https://github.com/rexec/rexec/releases/latest/download/rexec-darwin-amd64 -o /usr/local/bin/rexec
chmod +x /usr/local/bin/rexec

# macOS (Apple Silicon)
curl -sSL https://github.com/rexec/rexec/releases/latest/download/rexec-darwin-arm64 -o /usr/local/bin/rexec
chmod +x /usr/local/bin/rexec
```

### Using Go

```bash
go install github.com/rexec/rexec/cmd/rexec-cli@latest
```

### Using Homebrew (macOS)

```bash
brew install rexec/tap/rexec
```

## Quick Start

```bash
# Login to Rexec
rexec login

# List your terminals
rexec ls

# Create a new terminal
rexec create --name mydev --image ubuntu:22.04

# Connect to a terminal
rexec connect abc123

# Launch interactive TUI
rexec -i
```

## Commands Reference

### Authentication

#### login

Login to Rexec.

```bash
rexec login [--token TOKEN]
```

**Options:**

- `--token` - Provide token directly (skips interactive prompt)

**Examples:**

```bash
# Interactive login
rexec login

# With token
rexec login --token "your-api-token"

# Using environment variable
REXEC_TOKEN="your-token" rexec whoami
```

#### logout

Clear saved credentials.

```bash
rexec logout
```

#### whoami

Show current user information.

```bash
rexec whoami
```

**Output:**

```
User Profile
─────────────────────────────
  Username: johndoe
  Email:    john@example.com
  Tier:     pro
  Host:     https://rexec.pipeops.io
```

### Terminal Management

#### ls / list

List all terminals.

```bash
rexec ls
rexec list
```

**Output:**

```
Terminals
─────────────────────────────────────────────────────────────────────────────
ID           NAME                 IMAGE           STATUS       ROLE
─────────────────────────────────────────────────────────────────────────────
abc123def... my-dev               ubuntu:22.04    running      devops
def456ghi... staging-server       alpine:latest   stopped      default
```

#### create

Create a new terminal.

```bash
rexec create [options]
```

**Options:**
| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--name` | `-n` | auto-generated | Terminal name |
| `--image` | `-i` | ubuntu:22.04 | Container image |
| `--role` | `-r` | default | Environment role |
| `--memory` | `-m` | 512m | Memory limit |
| `--cpu` | `-c` | 0.5 | CPU limit |

**Available Roles:**

- `default` - Basic terminal
- `devops` - DevOps tools (docker, kubectl, terraform)
- `fullstack` - Web development (node, python, databases)
- `backend` - Backend development
- `frontend` - Frontend development
- `data` - Data science (python, jupyter, pandas)
- `security` - Security tools (nmap, metasploit)

**Examples:**

```bash
# Basic terminal
rexec create --name mydev

# DevOps environment
rexec create --name k8s-admin --role devops --memory 1g

# Data science environment
rexec create --name jupyter-lab --role data --image python:3.11
```

#### connect / ssh

Connect to a terminal (interactive shell).

```bash
rexec connect <terminal-id>
rexec ssh <terminal-id>
```

**Features:**

- Full PTY support
- Automatic terminal resize handling
- Press `Ctrl+]` to disconnect

**Examples:**

```bash
# Connect using ID
rexec connect abc123

# Connect using full ID
rexec connect abc123def456-7890-...
```

#### start

Start a stopped terminal.

```bash
rexec start <terminal-id>
```

#### stop

Stop a running terminal.

```bash
rexec stop <terminal-id>
```

#### rm / delete

Delete a terminal.

```bash
rexec rm <terminal-id>
rexec delete <terminal-id>
```

### Snippets & Macros

#### snippets

List and manage snippets.

```bash
# List your snippets
rexec snippets

# Browse marketplace snippets
rexec snippets marketplace
```

#### run

Run a snippet on a terminal.

```bash
rexec run <snippet-name> [--terminal <id>]
```

**Examples:**

```bash
# Run with terminal selection prompt
rexec run docker-install

# Run on specific terminal
rexec run setup-nodejs --terminal abc123
```

### Agent Mode

See [AGENTS.md](AGENTS.md) for detailed agent documentation.

> **Tip:** The easiest way to set up an agent is through the [Settings page](https://rexec.pipeops.io/settings) in your dashboard. It generates a ready-to-use command with your token pre-configured that you can copy and paste directly onto your server.

```bash
# Register this machine
rexec agent register --name "my-server"

# Start the agent
rexec agent start

# Check status
rexec agent status

# Stop the agent
rexec agent stop
```

### Interactive Mode (TUI)

#### -i / tui / dashboard

Launch the interactive TUI dashboard.

```bash
rexec -i
rexec --interactive
rexec tui
rexec dashboard
rexec ui
```

**TUI Features:**

- Visual terminal list with status indicators
- Quick connect (press 1-9)
- Create new terminals
- Browse snippets
- Keyboard navigation

**TUI Controls:**
| Key | Action |
|-----|--------|
| `1-9` | Connect to terminal |
| `c` | Create new terminal |
| `s` | View snippets |
| `r` | Refresh |
| `q` | Quit |

### Configuration

#### config

View or modify configuration.

```bash
# View configuration
rexec config

# Set host
rexec config set host https://custom.rexec.io
```

**Configuration file location:** `~/.rexec/config.json`

```json
{
  "host": "https://rexec.pipeops.io",
  "token": "your-token",
  "username": "johndoe",
  "email": "john@example.com",
  "tier": "pro"
}
```

## Environment Variables

| Variable         | Description                             |
| ---------------- | --------------------------------------- |
| `REXEC_HOST`     | API host URL (overrides config)         |
| `REXEC_TOKEN`    | Authentication token (overrides config) |
| `REXEC_TUI_PATH` | Custom path to TUI binary               |

## Exit Codes

| Code | Description   |
| ---- | ------------- |
| 0    | Success       |
| 1    | General error |

## Tips & Tricks

### Aliases

Add to your shell config (`~/.bashrc`, `~/.zshrc`):

```bash
# Quick connect to favorite terminal
alias dev="rexec connect abc123"

# Create common environments
alias newdev="rexec create --name dev-\$(date +%s) --role devops"
alias newdata="rexec create --name data-\$(date +%s) --role data"

# Quick dashboard
alias r="rexec -i"
```

### Shell Completion

```bash
# Bash
rexec completion bash > /etc/bash_completion.d/rexec

# Zsh
rexec completion zsh > ~/.zsh/completions/_rexec

# Fish
rexec completion fish > ~/.config/fish/completions/rexec.fish
```

### SSH Integration

Use rexec as an SSH replacement:

```bash
# In ~/.ssh/config
Host rexec-*
    ProxyCommand rexec connect %h
```

### CI/CD Integration

```yaml
# GitHub Actions
- name: Deploy to terminal
  run: |
    rexec login --token ${{ secrets.REXEC_TOKEN }}
    rexec run deploy-script --terminal ${{ vars.DEPLOY_TERMINAL }}
```

## Troubleshooting

### "Not logged in" error

```bash
# Check if logged in
rexec whoami

# Re-login
rexec login
```

### Connection timeout

```bash
# Check network
curl -I https://rexec.pipeops.io

# Try with verbose
REXEC_DEBUG=1 rexec connect abc123
```

### Terminal resize not working

Ensure your terminal emulator supports SIGWINCH. Try:

```bash
# Manual resize
printf '\e[8;50;120t'
```

### TUI not launching

```bash
# Check if TUI binary exists
which rexec-tui

# Set custom path
export REXEC_TUI_PATH=/path/to/rexec-tui
rexec -i
```
