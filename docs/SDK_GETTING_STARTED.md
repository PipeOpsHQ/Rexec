# SDK Getting Started

## Prerequisites

1. A Rexec instance URL (self-hosted or https://rexec.sh)
2. An API token (**Settings → API Tokens**) or a guest session token

```bash
export REXEC_URL=https://rexec.sh
export REXEC_TOKEN=$(curl -sS -X POST "$REXEC_URL/api/auth/guest" \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo","email":"you@example.com"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
```

## Install

| Language | Command |
|----------|---------|
| JavaScript | `npm install pipeops-rexec` |
| Python | `pip install pipeops-rexec` |
| Go | `go get github.com/PipeOpsHQ/rexec-go@v1.0.1` |
| Rust | `cargo add pipeops-rexec` |
| Ruby | `gem install pipeops-rexec` |
| .NET | `dotnet add package PipeOps.Rexec` |
| Java / Kotlin | `io.pipeops:rexec:1.0.1` (or `cd sdk/java && mvn install`) |
| PHP | `composer require pipeopshq/rexec` |

## Minimal flow (all SDKs)

1. **List** sandboxes  
2. **Create** with an image **alias** (e.g. `ubuntu`, not `ubuntu:24.04`)  
3. **Get** by id  
4. **Delete** when finished  

Full language examples: [SDK.md](SDK.md).

## Image aliases

Hosted Rexec validates images against a fixed catalog. Common aliases:

`ubuntu`, `ubuntu-24`, `ubuntu-22`, `debian`, `alpine`, `fedora`, `archlinux`, …

Fetch the full list: `GET /api/images`.

## Smoke tests

```bash
cd scripts/sdk-e2e
# See README.md for per-language runners (used in CI-style verification)
```

## Publishing packages

Only via GitHub Actions: [SDK_PUBLISHING.md](SDK_PUBLISHING.md).
