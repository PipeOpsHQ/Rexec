# Product Hunt Listing — Rexec

## Basics
- **Name:** Rexec
- **Tagline (pick one):**
  1) Terminal as a Service — cloud sandboxes + your own machines via agent
  2) Instant cloud terminals with a secure “bring your own server” agent
  3) Ephemeral dev environments + remote agent terminals in one dashboard

## Short Description (60–140 chars)
Instant, network‑isolated terminals in the cloud — and a secure agent to connect your own servers without opening SSH.

## Long Description
Rexec is a Terminal‑as‑a‑Service platform for developers and ops teams who need fast, disposable environments without sacrificing security or UX.

Create per‑user cloud terminals on demand, or connect your own servers/VMs/local machines using the Rexec Agent. Everything shows up in one dashboard with real‑time terminal streaming.

### What you can do today
- Spin up an isolated Linux terminal in seconds
- Safely run AI-generated code/tests in a disposable environment
- Reconnect without losing context
- Connect your own machines with the agent (no inbound ports required)
- Share terminals (view/control collaboration modes)
- Manage lifecycle: create/start/stop/delete

## Maker Comment (first comment)
Hey Product Hunt — I built Rexec because I kept needing safe “jump boxes” and disposable dev environments, but wanted the UX of a modern terminal + a secure way to reach my own servers without exposing ports.

Rexec gives you:
- Cloud terminals on demand
- A lightweight agent to connect your own machines
- Collaboration modes for pair debugging and demos

I’d love feedback on the onboarding flow, terminal UX, and what features would make this indispensable for your workflow.

## FAQ
**Is it self-hostable?**  
You can run the server yourself (Go + Docker + Postgres). Hosted is available at `https://rexec.pipeops.io`.

**How does the agent authenticate?**  
Agents use long‑lived API tokens (prefixed `rexec_`) intended for persistent connections.

**What’s the security model?**  
Short‑lived sandboxes, token-based auth, isolated terminal streaming over WebSockets; bring-your-own machines connect outbound to the server.

## Assets
- **Gallery SVG:** `launch/assets/producthunt-gallery.svg`
- **Thumbnail SVG:** `launch/assets/producthunt-thumbnail.svg`
