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

There is **no** primary HTTP `exec()` on hosted Rexec — run commands via the [terminal WebSocket](#terminal-websocket).

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

### Explicit non-features

- **No first-class `exec()` HTTP API** on hosted Rexec. Do not treat `exec` as the primary way to run commands. Use the **terminal WebSocket** (or raw HTTP if you build your own client).
- **Create is often async** — poll `get()` until `running` when you need a ready PTY.
- **Hosted images use aliases**, not arbitrary Docker Hub tags.

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
