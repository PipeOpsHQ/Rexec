# AI-Native Vision & FAQ

Rexec is designed for the "vibe coding" era. Its API-driven architecture is perfect for integration with AI agents and coding tools.

## AI Integration Features

*   **🤖 AI-Native**: Designed for the "vibe coding" era. API-driven architecture perfect for integration with AI agents and coding tools.
*   **Headless Capability**: The Rexec Agent allows AI coding tools (like Claude Code, Cursor, Windsurf) to execute commands in a secure, remote environment without requiring a local shell.

## FAQ

### Is Rexec a Virtual Machine?

**No.** Rexec is not a VM. It's a **Terminal as a Service** platform that gives you instant access to containerized Linux environments or your own machines via the agent. Think of it as a cloud-based terminal multiplexer, not a hypervisor.

### What's the difference between Rexec containers and the Agent?

| Feature        | Container                               | Agent                                                   |
| -------------- | --------------------------------------- | ------------------------------------------------------- |
| **What it is** | A Docker container provisioned by Rexec | A lightweight binary running on your existing machine   |
| **Setup**      | One click in the dashboard              | Install agent, register, done                           |
| **Resources**  | Dedicated, isolated container           | Uses your machine's resources                           |
| **Use case**   | Disposable dev environments, sandboxes  | Access your laptop, servers, Raspberry Pi from anywhere |
| **Network**    | Isolated container network              | Your machine's full network                             |

### Can I access my own laptop/server remotely?

**Yes!** This is exactly what the **Rexec Agent** is for. Install the agent on any machine, and you get instant terminal access from the Rexec dashboard—no SSH port exposure, no VPN, no complex setup.

```bash
# Install and register your machine in 30 seconds
curl -sSL https://rexec.pipeops.io/install-agent.sh | bash
rexec-agent register --name "my-laptop"
rexec-agent start --daemon
```

Now you can access your laptop's terminal from any browser, phone, or AI CLI tool.

### Why would I use Rexec instead of SSH?

- **No inbound ports** – The agent connects outbound, so no firewall holes needed
- **Works from anywhere** – Access from browser, mobile, or AI coding tools
- **No VPN required** – Direct access without complex network setup
- **Unified dashboard** – Manage all your machines (containers + agents) in one place
- **AI-native** – Built for integration with AI coding assistants

### Can I use Rexec with AI coding tools like Claude, Cursor, or Windsurf?

**Absolutely.** Rexec was designed with AI-native workflows in mind. Register your dev machine as an agent, and you can continue building in your environment from any AI CLI without opening your laptop.

### What happens if my connection drops?

The agent automatically reconnects with exponential backoff. Your terminal sessions are maintained, and you can resume right where you left off.

### Is it secure?

- All traffic is encrypted via WSS (WebSocket Secure)
- No inbound ports required on your machines
- JWT-based authentication
- Tokens can be rotated anytime
- See [SECURITY.md](docs/SECURITY.md) for details

### What operating systems are supported?

**Containers:** Ubuntu, Debian, Alpine, Fedora (pre-built images)

**Agent:** Linux (amd64, arm64), macOS (amd64, arm64), Windows (experimental)
