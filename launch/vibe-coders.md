# Rexec for Vibe Coders: An AI‑Ready Coding Terminal

“Vibe coding” is fast feedback loops: prompt → generate → run → test → iterate. The bottleneck isn’t ideas — it’s getting a clean environment with the right tools, quickly, from anywhere.

Rexec is a Terminal‑as‑a‑Service that doubles as a lightweight programming environment: spin up an isolated terminal in seconds, reconnect without losing context, and get a curated set of dev + AI CLI tools installed automatically.

## Why it works for vibe coding
- **Disposable by default**: run AI‑generated scripts in an isolated container instead of your laptop or prod box.
- **Persistent where it matters**: your home/workdir can persist across reconnects, while the environment stays easy to reset.
- **Terminal UX that keeps up**: optimized streaming, large paste support, and tmux-backed persistence so you can drop/reconnect without losing your session.

## What’s preinstalled (or auto-installed during role setup)
Rexec supports “roles” (Minimalist, Node, Python, Go, Neovim, DevOps, and **Vibe Coder**) that install a sensible toolchain for the job.

In the **Vibe Coder** role, you’ll typically have:
- **Shell comfort**: `zsh` with autosuggestions + syntax highlighting, plus `tmux`
- **Core dev tooling**: `git`, `curl`, `wget`, `jq`, `htop`
- **Editor/search**: `neovim`, `ripgrep`, `fzf`
- **Languages**: `python3` (+ `pip`, `venv`), `node` (+ `npm`)
- **AI CLI tools (free/no key)**: `tgpt`, `aichat`, `mods`
- **AI coding assistants**:
  - `opencode` (AI coding assistant)
  - `aider` / `llm` / `sgpt` (install + configure with your own provider keys)
  - `gh` Copilot extension (requires `gh auth login`)
  - `claude` / `gemini` CLIs (require your API keys)

To see what your environment has, run:
```bash
rexec tools
ai-help
```

## A practical vibe‑coding workflow
1) **Create a terminal** with the **Vibe Coder** role.
2) **Clone a repo** and let the LLM generate changes (via your preferred CLI).
3) **Run tests in isolation** (and break things safely):
   - `go test ./...`
   - `npm test`
   - `pytest -q`
4) **Iterate quickly** with `rg` + `fzf` for code navigation and `tmux` to keep long-running processes alive across reconnects.

## Why teams like it
If you’re doing AI‑assisted coding, “safe execution” matters. Rexec gives you a controlled place to run generated code, validate it with tests, and share a session (view/control) when you need a second pair of eyes — without asking everyone to recreate your local setup.

