# Reddit Drafts — Rexec

## r/devops (draft)
Title: Rexec — on-demand cloud terminals + a secure agent to connect your own servers (no inbound SSH)

Body:
I’ve been building Rexec: a Terminal-as-a-Service dashboard where you can spin up isolated cloud terminals on demand, and also connect your own servers/VMs via an outbound agent (so you don’t have to expose SSH ports).

It’s also handy for AI-assisted work: generate scripts/patches with an LLM, then run them + tests in an isolated terminal before they touch your laptop or prod.

Looking for feedback from folks who run production infra:
- What would you want for access control/auditing?
- What makes a terminal UX “good enough” for day-to-day ops?
- Any must-have features before you’d consider it?

Disclosure: I’m the builder.
Link: https://rexec.pipeops.io

## r/selfhosted (draft)
Title: Built a self-hostable Terminal-as-a-Service (Go + Docker + Postgres) — feedback?

Body:
Rexec is a dashboard for cloud terminals and agent-connected machines. You can run the server yourself (Go + Docker + Postgres) and connect machines via an outbound agent.

If you self-host similar tools, what would you expect around:
- auth (SSO/MFA), audit logs
- resource limits / multi-tenant isolation
- backups / upgrades

Disclosure: I’m the builder.
Link: https://rexec.pipeops.io
