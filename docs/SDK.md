# Rexec SDK — API Reference

Official client libraries for the Rexec **sandbox** API (files and terminals).

**Product:** Rexec is PipeOps’ **AI-native sandbox** platform — instant, isolated Linux environments via API / UI / CLI. SDKs share one REST + WebSocket surface.

| | |
|--|--|
| **Hosted base URL** | `https://rexec.sh` |
| **Auth header** | `Authorization: Bearer <token>` |
| **SDK version** | **v1.1.0** |
| **Quick start** | [SDK_GETTING_STARTED.md](SDK_GETTING_STARTED.md) |
| **Publishing** | [SDK_PUBLISHING.md](SDK_PUBLISHING.md) |
| **Source monorepo** | [github.com/PipeOpsHQ/Rexec](https://github.com/PipeOpsHQ/Rexec) (`sdk/{js,python,go,rust,ruby,dotnet,java,php}`) |
| **In-app docs** | `/docs/sdk` on the product UI |
| **PipeOps docs** | [docs.pipeops.io — Rexec Sandboxes](https://docs.pipeops.io/docs/rexec/overview) (when published) |
| **E2E smoke** | [`scripts/sdk-e2e/`](../scripts/sdk-e2e/) (`test-js.mjs`, `test_py.py`, Go/Rust/Ruby/.NET/Java/PHP runners) |
| **MCP (agents)** | [`sdk/mcp`](../sdk/mcp/) — `@pipeops/rexec-mcp` (stdio tools for create/exec/templates) |

> **Verified E2E** against a live Rexec instance: `list` → `create` → `get` → `delete`.

---

## Sandboxes (product concept)

A **sandbox** is the isolated Linux workspace you create, use, and delete.

| Term | Meaning |
|------|---------|
| **Sandbox** | Product language and preferred SDK names (`client.sandboxes`, `Sandbox`) |
| **Container** | HTTP wire name (`/api/containers`) and **deprecated** SDK aliases (`client.containers`, `Container`) |

Typical use cases: AI agents running code safely, ephemeral dev shells, demos, and CI-style throwaway environments.

**Lifecycle:** `create` → often `creating` (async) → `running` ⇄ `stopped` → `delete` (or `error`).

**Interactive commands** use the [terminal WebSocket](#terminal-websocket). **Non-interactive / agent commands** use the [exec API](#exec-api).

---

## Available SDKs (v1.1.0)

| Language | Package / module | Install | Import notes |
|----------|------------------|---------|--------------|
| **JS / TS** | [pipeops-rexec](https://www.npmjs.com/package/pipeops-rexec) | `npm install pipeops-rexec` | `import { RexecClient } from 'pipeops-rexec'` |
| **Python** | [pipeops-rexec](https://pypi.org/project/pipeops-rexec/) | `pip install pipeops-rexec` | `from rexec import RexecClient` |
| **Go** | [github.com/PipeOpsHQ/rexec-go](https://github.com/PipeOpsHQ/rexec-go) | `go get github.com/PipeOpsHQ/rexec-go@v1.1.0` | `import rexec "github.com/PipeOpsHQ/rexec-go"` |
| **Rust** | [pipeops-rexec](https://crates.io/crates/pipeops-rexec) | `cargo add pipeops-rexec` | `use rexec::{…}` (crate name ≠ import name) |
| **Ruby** | [pipeops-rexec](https://rubygems.org/gems/pipeops-rexec) | `gem install pipeops-rexec` | `require "rexec"` |
| **C# / .NET** | [PipeOps.Rexec](https://www.nuget.org/packages/PipeOps.Rexec) | `dotnet add package PipeOps.Rexec` | `using Rexec;` |
| **Java / Kotlin** | `io.pipeops:rexec:1.1.0` | Maven/Gradle | `import io.pipeops.rexec.*` |
| **PHP** | [pipeopshq/rexec](https://packagist.org/packages/pipeopshq/rexec) | `composer require pipeopshq/rexec` | `use Rexec\RexecClient` |

Publishing is **GitHub Actions only** — see [SDK_PUBLISHING.md](SDK_PUBLISHING.md).

### Fallback installs

```bash
# PHP (if Packagist 404) — VCS from standalone mirror
composer config repositories.rexec-php vcs https://github.com/PipeOpsHQ/rexec-php
composer require pipeopshq/rexec:^1.0

# Java from monorepo until Maven Central is configured
cd sdk/java && mvn install -DskipTests
```

---

## Auth

### 1. API token (recommended)

Rexec UI → **Settings** → **API Tokens** → create token. Use for production apps and agents.

### 2. Guest JWT (smoke tests)

```bash
export REXEC_URL=https://rexec.sh
export REXEC_TOKEN=$(curl -sS -X POST "$REXEC_URL/api/auth/guest" \
  -H 'Content-Type: application/json' \
  -d '{"username":"sdk_demo","email":"you@example.com"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
```

SDKs set `Authorization: Bearer <token>` on REST and (where applicable) WebSocket upgrades.

Guest accounts have **limited** concurrent sandboxes and session lifetime; do not use for production.

---

## Image aliases

Hosted Rexec rejects many raw Docker Hub tags.

| Prefer (aliases) | Avoid on hosted |
|------------------|-----------------|
| `ubuntu`, `ubuntu-24`, `ubuntu-22` | `ubuntu:24.04`, `ubuntu:latest` |
| `debian`, `alpine`, `fedora`, `archlinux`, … | Arbitrary Hub digests/tags |

Catalog: **`GET /api/images`**.

---

## Shared API surface

Every SDK exposes roughly three services over the same HTTP/WS endpoints.

### Sandboxes (`sandboxes` service; wire: `/api/containers`)

Prefer **`client.sandboxes`** (or `Sandboxes` / `sandboxes()`). Legacy **`client.containers`** is the same service.

| Method | HTTP | Notes |
|--------|------|--------|
| **List** | `GET /api/containers` | Response is often `{ containers: [...] \| null, count, limit }`. SDKs **normalize to an array** of sandboxes. |
| **Create** | `POST /api/containers` body `{ image, name? }` | Creates a sandbox. Often returns immediately with `status: "creating"` (**async**). |
| **Get** | `GET /api/containers/:id` | Poll until `status === "running"` if you need a ready shell. |
| **Start** | `POST /api/containers/:id/start` | |
| **Stop** | `POST /api/containers/:id/stop` | |
| **Delete** | `DELETE /api/containers/:id` | Destroys the sandbox |
| **Exec** | `POST /api/containers/:id/exec` | One-shot command; see [Exec API](#exec-api) |
| **Templates** | `GET/POST /api/templates` | Commit a running sandbox and create from it; see [Templates](#templates) |

Create options: `template_id`, `network_mode` (`default` \| `none` \| `restricted`).

### Files (per sandbox)

| Method | HTTP |
|--------|------|
| **List dir** | `GET /api/containers/:id/files/list?path=` |
| **Read** | `GET /api/containers/:id/files?path=` |
| **Write** | `POST /api/containers/:id/files` (content often base64) |
| **Mkdir** | `POST /api/containers/:id/files/mkdir` (JS+; check language README) |
| **Delete** | `DELETE /api/containers/:id/files?path=` |

### Terminal

| Method | Transport |
|--------|-----------|
| **Connect** | WebSocket `wss://<host>/ws/terminal/:id?cols=&rows=` with Bearer auth |
| **write / onData / resize / close** | SDK helpers over the socket |

### Explicit non-features / caveats

- **Create is often async** — poll `get()` until `running` before terminal or exec.
- **Hosted images use aliases**, not arbitrary Docker Hub tags.
- Prefer **exec** for agents; prefer **terminal WebSocket** for interactive shells.

---

## Exec API {#exec-api}

Run a non-interactive command in a **running** sandbox (agent-friendly).

| | |
|--|--|
| **HTTP** | `POST /api/containers/:id/exec` |
| **Auth** | Bearer token (owner of sandbox) |
| **SDK** | `client.sandboxes.exec(id, …)` (JS / Python) |

### Request body

```json
{
  "command": "echo hello && uname -a",
  "timeout_seconds": 60
}
```

Or argv form (takes precedence when non-empty):

```json
{
  "cmd": ["uname", "-a"],
  "workdir": "/home/user",
  "env": ["FOO=bar"],
  "user": "root",
  "timeout_seconds": 30
}
```

| Field | Notes |
|-------|--------|
| `command` | Shell string → run as `sh -c <command>` when `cmd` is empty |
| `cmd` | Argv vector |
| `workdir` | Optional working directory |
| `env` | Optional `KEY=VALUE` list |
| `user` | Optional user inside sandbox |
| `timeout_seconds` | Default **60**, max **300** |

### Response

```json
{
  "stdout": "hello\n",
  "stderr": "",
  "output": "hello\n",
  "exit_code": 0,
  "duration_ms": 42,
  "command": "echo hello"
}
```

| Field | Notes |
|-------|--------|
| `stdout` / `stderr` | Demuxed Docker attach streams |
| `output` | Combined for simple clients |
| `exit_code` | Process exit code (`-1` on timeout) |
| `duration_ms` | Wall time |
| `truncated` | `true` if output hit the 1 MiB cap |

Timeouts return **504** with partial `stdout`/`stderr` when available.

### Examples

```typescript
const r = await client.sandboxes.exec(sandbox.id, { command: 'echo hello' });
console.log(r.stdout, r.exit_code);

await client.sandboxes.exec(sandbox.id, { cmd: ['python3', '-c', 'print(1+1)'] });
// shorthand:
await client.sandboxes.exec(sandbox.id, 'uname -a');
```

```python
r = await client.sandboxes.exec(sandbox.id, "echo hello")
print(r.stdout, r.exit_code)

r = await client.sandboxes.exec(sandbox.id, cmd=["uname", "-a"])
```

---

## Templates {#templates}

Save a running sandbox as a **local Docker image** and spawn new sandboxes from it (agent warm-start pattern).

| Method | HTTP |
|--------|------|
| **List** | `GET /api/templates` → `{ templates, count }` |
| **Create** | `POST /api/templates` body `{ name, from_sandbox_id, description? }` |
| **Get** | `GET /api/templates/:id` |
| **Delete** | `DELETE /api/templates/:id` |
| **Create sandbox from template** | `POST /api/containers` body `{ template_id, name? }` |

Create-template requires the source sandbox to be **running** (uses `docker commit`). Images are tagged locally as `rexec-template/<user>/<id>:latest` (not pushed to a registry in v1).

```typescript
const tpl = await client.sandboxes.createTemplate({
  name: 'with-deps',
  from_sandbox_id: sandbox.id,
});
const clone = await client.sandboxes.create({ template_id: tpl.id, name: 'job-2' });
```

### Network mode on create

| `network_mode` | Behavior |
|----------------|----------|
| `default` / omit | Isolated bridge (`rexec-isolated`) with full outbound |
| `none` | No network (strong isolation for pure compute) |
| `restricted` | HTTP(S) **only** via host egress proxy + host allowlist |

### Restricted egress allowlist

When `network_mode` is `restricted`, the sandbox gets `HTTP_PROXY`/`HTTPS_PROXY` pointing at a host-side CONNECT proxy. Raw TCP (SSH, custom ports) is not proxied.

- **Defaults:** package mirrors (Ubuntu/Debian/Alpine), PyPI, npm, GitHub, Go modules, crates.io, Maven, RubyGems (see `DefaultRestrictedEgressAllow`).
- **Override defaults:** server env `RESTRICTED_EGRESS_ALLOW=host1,*.example.com`
- **Per create extras:** `egress_allow: ["api.openai.com","*.anthropic.com"]` (union with defaults)
- **Disable proxy:** `RESTRICTED_EGRESS_ENABLED=false` (restricted then falls back to no network)
- **Listen:** `RESTRICTED_EGRESS_PROXY_ADDR=:13128`

```json
{
  "image": "ubuntu",
  "network_mode": "restricted",
  "egress_allow": ["api.openai.com", "api.anthropic.com"]
}
```

```json
{ "image": "ubuntu", "network_mode": "none" }
```

---

## Snapshots & fork {#snapshots}

Point-in-time **filesystem** copies via `docker commit` (same underlying mechanism as templates; separate API for agent “branch my env” flows).

| Method | HTTP |
|--------|------|
| **Snapshot** | `POST /api/containers/:id/snapshot` `{ name?, description? }` |
| **List** | `GET /api/snapshots` |
| **Get** | `GET /api/snapshots/:id` |
| **Delete** | `DELETE /api/snapshots/:id` |
| **Fork** | `POST /api/containers/:id/fork` — commit + create new running sandbox |
| **Create from snapshot** | `POST /api/containers` `{ snapshot_id, name? }` |

```typescript
const snap = await client.sandboxes.snapshot(id, { name: 'before-refactor' });
const clone = await client.sandboxes.create({ snapshot_id: snap.id });

// or one-shot fork (new running sandbox from current FS)
const forked = await client.sandboxes.fork(id, {
  name: 'experiment',
  save_snapshot: true,
  network_mode: 'restricted',
});
```

```python
snap = await client.sandboxes.snapshot(sid, name="before-refactor")
clone = await client.sandboxes.create(snapshot_id=snap.id)
forked = await client.sandboxes.fork(sid, name="experiment", save_snapshot=True)
```

Local image tags: `rexec-snapshot/<user>/<id>:latest` / `rexec-fork/...` (not pushed to a registry in v1).

---

## Warm pool & lifecycle {#warm-pool}

### Warm pool (server config)

Pre-create sandboxes so `create` can return **`status: running`** immediately (`warm: true`).

```bash
# Host env (rexec server)
WARM_POOL=ubuntu:2,debian:1
# WARM_POOL_ENABLED=false
# WARM_POOL_INTERVAL_SEC=30
```

Create still works without a pool (async `creating`). With stock available and `prefer_warm` not false:

```json
{ "image": "ubuntu", "name": "job-1" }
// → 200 { "status": "running", "warm": true, ... }
```

```json
{ "image": "ubuntu", "prefer_warm": false }
// → always cold create (async)
```

### Per-sandbox lifecycle (create body)

| Field | Effect |
|-------|--------|
| `idle_timeout_seconds` | Stop after N seconds without activity (exec/terminal Touch) |
| `max_lifetime_seconds` | Hard TTL from create → `rexec.expires_at` |

```json
{
  "image": "ubuntu",
  "idle_timeout_seconds": 600,
  "max_lifetime_seconds": 3600
}
```

Guests still get platform idle/session limits; these fields add **explicit** timeouts for agent workloads on any tier.

---

## Client construction

| Lang | Construct | Sandboxes (preferred) |
|------|-----------|------------------------|
| **JS** | `new RexecClient({ baseURL, token })` | `client.sandboxes.list()` / `.create({ image, name? })` |
| **Python** | `async with RexecClient(url, token)` | `await client.sandboxes.list()` / `.create(image=…, name=…)` |
| **Go** | `rexec.NewClient(url, token)` | `client.Sandboxes.List(ctx)` / `.Create(ctx, &CreateSandboxRequest{…})` |
| **Rust** | `RexecClient::new(url, token)` | `client.sandboxes().list().await` / `.create(CreateSandboxRequest::new("ubuntu").name(…))` |
| **Ruby** | `Rexec::Client.new(url, token)` | `client.sandboxes.list` / `.create(image: "ubuntu", name: "demo")` |
| **.NET** | `new RexecClient(url, token)` | `await client.Sandboxes.ListAsync()` / `CreateAsync(new CreateSandboxRequest("ubuntu") { Name = … })` |
| **Java** | `new RexecClient(url, token)` | `client.sandboxes().list()` / `.create(new CreateSandboxRequest("ubuntu").setName(…))` |
| **PHP** | `new Rexec\RexecClient($url, $token)` | `$client->sandboxes()->list()` / `->create('ubuntu', ['name' => 'demo'])` |

Legacy `containers` / `Container` / `CreateContainerRequest` aliases remain through 1.x.

Optional client options (language-dependent): custom `fetch` (JS), timeouts, TLS. Defaults point at your `baseURL` + Bearer token.

---

## Sandbox (container) model

Conceptual fields (names may be camelCase or snake_case per language):

| Field | Description |
|-------|-------------|
| `id` | Sandbox id |
| `name` | Display name |
| `image` | Alias used at create |
| `status` | `"creating"` \| `"running"` \| `"stopped"` \| `"error"` (and similar) |
| `created_at` | Create time |
| `started_at` | Optional |
| `labels` | Optional map |
| `environment` | Optional map |

### Status lifecycle

```
create → creating → running ⇄ stopped → (delete)
                 ↘ error
```

### List payload quirks

Raw API:

```json
{ "containers": [ /* or null */ ], "count": 0, "limit": 1 }
```

SDKs return a plain array so callers never special-case `null`.

---

## Sandboxes API (detail)

Happy path for every language (via `sandboxes` service; `containers` alias works):

1. `list()` — sandboxes for the current token  
2. `create({ image: "ubuntu", name?: "…" })` — new sandbox  
3. `get(id)` — optionally loop until `status == "running"`  
4. `delete(id)` — destroy sandbox  

Also available: `start(id)`, `stop(id)`.

### Optional: wait until running (pattern)

```typescript
async function waitRunning(client: RexecClient, id: string, ms = 60_000) {
  const deadline = Date.now() + ms;
  while (Date.now() < deadline) {
    const c = await client.sandboxes.get(id);
    if (c.status === 'running') return c;
    if (c.status === 'error') throw new Error('sandbox failed');
    await new Promise((r) => setTimeout(r, 1000));
  }
  throw new Error('timeout waiting for running');
}
```

---

## Files API (detail)

Typical flow after the **sandbox** is `running`:

```typescript
// JS — containerId is the sandbox id
const entries = await client.files.list(containerId, '/home');
const bytes = await client.files.read(containerId, '/etc/hostname');
await client.files.write(containerId, '/tmp/hello.txt', 'hi\n');
await client.files.mkdir?.(containerId, '/tmp/work'); // if available
await client.files.delete(containerId, '/tmp/hello.txt');
```

Write payloads may be base64-encoded depending on language bindings; SDKs accept strings/bytes and encode as needed.

---

## Terminal WebSocket {#terminal-websocket}

### Connect

```
wss://<host>/ws/terminal/<container_id>?cols=80&rows=24
```

Auth: Bearer token (query or header depending on environment; SDKs handle this).

### SDK helpers (common)

| Helper | Purpose |
|--------|---------|
| `write(data)` | Send keystrokes / shell input (often UTF-8 text or bytes) |
| `onData(cb)` | Receive PTY output |
| `resize(cols, rows)` | Often sends JSON `{ "type": "resize", "cols", "rows" }` on the socket |
| `close()` | Tear down the WebSocket |

### Example (JS)

```typescript
const term = await client.terminal.connect(container.id, { cols: 120, rows: 40 });
term.onData((data) => process.stdout.write(String(data)));
term.write('echo hello from rexec\n');
term.resize(100, 30);
term.close();
```

Do **not** document a primary `exec()` RPC for hosted Rexec; interactive and one-shot commands go through this socket (or your own HTTP client if you implement one).

---

## Errors

| SDK | Exception / error type |
|-----|------------------------|
| JS | `RexecError` (`statusCode`, `message`) |
| Python | `RexecAPIError`, `RexecConnectionError` |
| Go | `*rexec.APIError` |
| Rust | `rexec::Error` |
| Ruby | `Rexec::APIError` |
| .NET | `RexecException` |
| Java | `RexecException` |
| PHP | `RexecException` |

HTTP 4xx/5xx from the API map into these types. Network failures may use a separate connection error (Python).

---

## Language examples

### JavaScript / TypeScript

```bash
npm install pipeops-rexec
# optional (Node): npm install ws
```

```typescript
import { RexecClient } from 'pipeops-rexec';

const client = new RexecClient({
  baseURL: process.env.REXEC_URL!,
  token: process.env.REXEC_TOKEN!,
});

const list = await client.sandboxes.list();
const sandbox = await client.sandboxes.create({
  image: 'ubuntu',
  name: 'demo',
});
console.log(sandbox.id, sandbox.status);

const got = await client.sandboxes.get(sandbox.id);
await client.sandboxes.delete(sandbox.id);

// Terminal (browser WebSocket or `ws` in Node)
// Prefer after status === "running"
const term = await client.terminal.connect(sandbox.id);
term.onData((data) => process.stdout.write(String(data)));
term.write('echo hello\n');
term.close();
```

### Python

```bash
pip install pipeops-rexec
```

```python
import asyncio
import os
from rexec import RexecClient

async def main():
    async with RexecClient(os.environ["REXEC_URL"], os.environ["REXEC_TOKEN"]) as client:
        print(await client.sandboxes.list())
        c = await client.sandboxes.create(image="ubuntu", name="demo")
        print(c.id, c.status)
        await client.sandboxes.get(c.id)
        await client.sandboxes.delete(c.id)

asyncio.run(main())
```

### Go

```bash
go get github.com/PipeOpsHQ/rexec-go@v1.1.0
```

```go
package main

import (
    "context"
    "fmt"
    "os"

    rexec "github.com/PipeOpsHQ/rexec-go"
)

func main() {
    client := rexec.NewClient(os.Getenv("REXEC_URL"), os.Getenv("REXEC_TOKEN"))
    ctx := context.Background()

    list, _ := client.Sandboxes.List(ctx)
    fmt.Println("count", len(list))

    c, err := client.Sandboxes.Create(ctx, &rexec.CreateSandboxRequest{
        Image: "ubuntu",
        Name:  "demo",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(c.ID, c.Status)
    _, _ = client.Sandboxes.Get(ctx, c.ID)
    _ = client.Sandboxes.Delete(ctx, c.ID)
}
```

Standalone module repo: [PipeOpsHQ/rexec-go](https://github.com/PipeOpsHQ/rexec-go) (synced from monorepo `sdk/go`).

### Rust

```bash
cargo add pipeops-rexec tokio --features tokio/full
```

```rust
use rexec::{CreateSandboxRequest, RexecClient};

#[tokio::main]
async fn main() -> Result<(), rexec::Error> {
    let client = RexecClient::new(
        std::env::var("REXEC_URL").unwrap(),
        std::env::var("REXEC_TOKEN").unwrap(),
    );

    let list = client.sandboxes().list().await?;
    println!("count {}", list.len());

    let c = client
        .sandboxes()
        .create(CreateSandboxRequest::new("ubuntu").name("demo"))
        .await?;
    println!("{} {}", c.id, c.status);
    client.sandboxes().get(&c.id).await?;
    client.sandboxes().delete(&c.id).await?;
    Ok(())
}
```

Import crate name: `use rexec::...` (package name is `pipeops-rexec`).

### Ruby

```bash
gem install pipeops-rexec   # Ruby >= 3.0
```

```ruby
require "rexec"

client = Rexec::Client.new(ENV["REXEC_URL"], ENV["REXEC_TOKEN"])
puts client.sandboxes.list.length

c = client.sandboxes.create(image: "ubuntu", name: "demo")
puts "#{c.id} #{c.status}"
client.sandboxes.get(c.id)
client.sandboxes.delete(c.id)
```

### C# / .NET

```bash
dotnet add package PipeOps.Rexec
```

```csharp
using Rexec;

using var client = new RexecClient(
    Environment.GetEnvironmentVariable("REXEC_URL")!,
    Environment.GetEnvironmentVariable("REXEC_TOKEN")!);

var list = await client.Sandboxes.ListAsync();
var c = await client.Sandboxes.CreateAsync(new CreateSandboxRequest("ubuntu") {
    Name = "demo"
});
Console.WriteLine($"{c?.Id} {c?.Status}");
await client.Sandboxes.GetAsync(c!.Id);
await client.Sandboxes.DeleteAsync(c.Id);
// Run commands via Terminal WebSocket helpers — not a primary HTTP exec() API.
```

### Java / Kotlin

```xml
<dependency>
  <groupId>io.pipeops</groupId>
  <artifactId>rexec</artifactId>
  <version>1.1.0</version>
</dependency>
```

```java
import io.pipeops.rexec.*;

RexecClient client = new RexecClient(
    System.getenv("REXEC_URL"),
    System.getenv("REXEC_TOKEN"));

System.out.println(client.sandboxes().list().size());
Sandbox c = client.sandboxes().create(
    new CreateSandboxRequest("ubuntu").setName("demo"));
System.out.println(c.getId() + " " + c.getStatus());
client.sandboxes().get(c.getId());
client.sandboxes().delete(c.getId());
```

Kotlin:

```kotlin
val client = RexecClient(System.getenv("REXEC_URL"), System.getenv("REXEC_TOKEN"))
val c = client.sandboxes().create(CreateSandboxRequest("ubuntu").setName("demo"))
client.sandboxes().delete(c.id)
```

Install from source: `cd sdk/java && mvn install -DskipTests`.  
Maven Central coordinates: `io.pipeops:rexec:1.1.0` · [repo path](https://repo1.maven.org/maven2/io/pipeops/rexec/).

### PHP

```bash
composer require pipeopshq/rexec
```

Canonical Packagist source: [PipeOpsHQ/rexec-php](https://github.com/PipeOpsHQ/rexec-php) (synced from `sdk/php`).  
If Packagist is not linked yet:

```bash
composer config repositories.rexec-php vcs https://github.com/PipeOpsHQ/rexec-php
composer require pipeopshq/rexec:^1.1
```

```php
<?php
require 'vendor/autoload.php';

use Rexec\RexecClient;

$client = new RexecClient(getenv('REXEC_URL'), getenv('REXEC_TOKEN'));
echo count($client->sandboxes()->list()), PHP_EOL;

$c = $client->sandboxes()->create('ubuntu', ['name' => 'demo']);
echo $c->id, ' ', $c->status, PHP_EOL;
$client->sandboxes()->get($c->id);
$client->sandboxes()->delete($c->id);
// Legacy: $client->containers() is the same service
```

---

## End-to-end smoke tests

```bash
cd scripts/sdk-e2e
# See README.md — list → create → get → delete for each language
# Runners: test-js.mjs, test_py.py, go/, rust_e2e, test_rb.rb,
#          dotnet_e2e, java_e2e, test_php.php
# Legacy containers.* paths still pass (aliases).
```

---

## MCP server (AI agents) {#mcp}

Official MCP package: **`@pipeops/rexec-mcp`** ([`sdk/mcp`](../sdk/mcp/)).

Exposes tools: `list_sandboxes`, `create_sandbox`, `exec`, `list_files`, `create_template`, `list_templates`, `wait_running`, and more.

```bash
cd sdk/js && npm run build
cd ../mcp && npm install && npm run build
REXEC_URL=https://rexec.sh REXEC_TOKEN=... node dist/index.js
```

Claude Desktop / Cursor config:

```json
{
  "mcpServers": {
    "rexec": {
      "command": "node",
      "args": ["/absolute/path/to/rexec/sdk/mcp/dist/index.js"],
      "env": {
        "REXEC_URL": "https://rexec.sh",
        "REXEC_TOKEN": "your-token"
      }
    }
  }
}
```

See [sdk/mcp/README.md](../sdk/mcp/README.md).

---

## Migration: `containers` → `sandboxes` (v1.1.0)

Product language is **sandbox**. SDKs expose preferred **`sandboxes`** accessors and **`Sandbox`** types.

| Prefer (1.1+) | Still works (deprecated alias) |
|---------------|--------------------------------|
| `client.sandboxes` / `Sandboxes` / `sandboxes()` | `client.containers` / `Containers` / `containers()` |
| `Sandbox` | `Container` |
| `CreateSandboxRequest` | `CreateContainerRequest` |
| `SandboxService` | `ContainerService` |

- **Wire protocol unchanged:** HTTP remains `/api/containers` and list JSON still uses a `containers` array.
- **Same service instance:** legacy accessors point at the new service (no double clients).
- **Removal:** aliases may be removed in **2.0.0**.

## References

| Resource | Link |
|----------|------|
| Quick start | [SDK_GETTING_STARTED.md](SDK_GETTING_STARTED.md) |
| Publishing / CI | [SDK_PUBLISHING.md](SDK_PUBLISHING.md) |
| Per-language READMEs | [`sdk/*/README.md`](../sdk/) |
| Product site | https://rexec.sh |
| PipeOps docs (sandboxes) | https://docs.pipeops.io/docs/rexec/overview |
| Monorepo | https://github.com/PipeOpsHQ/Rexec |
| npm | https://www.npmjs.com/package/pipeops-rexec |
| PyPI | https://pypi.org/project/pipeops-rexec/ |
| crates.io | https://crates.io/crates/pipeops-rexec |
| RubyGems | https://rubygems.org/gems/pipeops-rexec |
| NuGet | https://www.nuget.org/packages/PipeOps.Rexec |
| Maven | https://repo1.maven.org/maven2/io/pipeops/rexec/ |
| Go module | https://github.com/PipeOpsHQ/rexec-go |
| PHP source | https://github.com/PipeOpsHQ/rexec-php |
