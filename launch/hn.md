# Show HN Draft — Rexec

## Title Options
1) Show HN: Rexec — Terminal as a Service (cloud terminals + remote agents)
2) Show HN: Rexec — disposable cloud terminals and a secure BYO server agent

## Post Body (paste into HN)
Hi HN — I’m building Rexec, a Terminal‑as‑a‑Service platform.

The problem I wanted to solve: I often need a safe “jump box” or a disposable dev environment (with decent UX), and I also want to reach my own servers without exposing SSH ports or maintaining brittle networking rules.

Rexec gives you:
- On‑demand, network‑isolated cloud terminals (Docker-backed)
- A safe place to run AI-generated scripts/patches and tests before they touch your machine
- A lightweight agent to connect your own machines to the dashboard (outbound connection)
- Real‑time terminal streaming over WebSockets
- Collaboration modes (view/control)

I’d love feedback on what would make this useful for your workflow, and what you’d expect from a “terminal control room” product.

Link: https://rexec.pipeops.io

## Maker Follow‑Up Comment (optional)
Happy to answer questions about the security model, isolation boundaries, and what “agent mode” looks like operationally. If you want to self-host, it’s Go + Docker + Postgres.
