# Rexec SDK Quick Start

**~5 minutes** to list, create, inspect, and delete a sandbox with an official client.

**Rexec** is an AI-native sandbox platform: instant, isolated Linux environments (containers) via API, UI, and CLI. Official language SDKs wrap the same REST + WebSocket API so agents and apps can manage sandboxes, files, and interactive terminals.

| | |
|--|--|
| **Hosted API** | `https://rexec.sh` |
| **Auth** | `Authorization: Bearer <token>` |
| **SDK line** | **v1.0.1** |
| **Full reference** | [SDK.md](SDK.md) |
| **In-app docs** | `https://rexec.sh` → `/docs/sdk` |
| **Monorepo** | [github.com/PipeOpsHQ/Rexec](https://github.com/PipeOpsHQ/Rexec) |

> **E2E verified:** `list` → `create` → `get` → `delete` against a live instance.  
> Smoke runners: [`scripts/sdk-e2e/`](../scripts/sdk-e2e/).

---

## 1. Prerequisites

1. A Rexec base URL (hosted `https://rexec.sh` or your self-hosted origin).
2. A Bearer token:
   - **API token** (production / apps): Rexec UI → **Settings** → **API Tokens**
   - **Guest JWT** (short-lived smoke tests):

```bash
export REXEC_URL=https://rexec.sh
export REXEC_TOKEN=$(curl -sS -X POST "$REXEC_URL/api/auth/guest" \
  -H 'Content-Type: application/json' \
  -d '{"username":"docs_demo","email":"you@example.com"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
```

All official SDKs inject `Authorization: Bearer <token>` for you.

Guest sessions are limited (fewer concurrent sandboxes / shorter lifetime). Prefer API tokens for real apps.

---

## 2. Install (v1.0.1)

| Language | Package | Install |
|----------|---------|---------|
| **JavaScript / TypeScript** | [`pipeops-rexec`](https://www.npmjs.com/package/pipeops-rexec) | `npm install pipeops-rexec` |
| **Python** | [`pipeops-rexec`](https://pypi.org/project/pipeops-rexec/) (`import rexec`) | `pip install pipeops-rexec` |
| **Go** | [`github.com/PipeOpsHQ/rexec-go`](https://github.com/PipeOpsHQ/rexec-go) | `go get github.com/PipeOpsHQ/rexec-go@v1.0.1` |
| **Rust** | crate [`pipeops-rexec`](https://crates.io/crates/pipeops-rexec) (`use rexec::…`) | `cargo add pipeops-rexec` |
| **Ruby** | gem [`pipeops-rexec`](https://rubygems.org/gems/pipeops-rexec) | `gem install pipeops-rexec` |
| **C# / .NET** | [`PipeOps.Rexec`](https://www.nuget.org/packages/PipeOps.Rexec) (namespace `Rexec`) | `dotnet add package PipeOps.Rexec` |
| **Java / Kotlin** | `io.pipeops:rexec:1.0.1` | Maven/Gradle (see below) |
| **PHP** | [`pipeopshq/rexec`](https://packagist.org/packages/pipeopshq/rexec) | `composer require pipeopshq/rexec` |

### Fallback installs

```bash
# PHP until Packagist is live
composer config repositories.rexec-php vcs https://github.com/PipeOpsHQ/rexec-php
composer require pipeopshq/rexec:^1.0

# Java from monorepo
cd sdk/java && mvn install -DskipTests
```

Maven:

```xml
<dependency>
  <groupId>io.pipeops</groupId>
  <artifactId>rexec</artifactId>
  <version>1.0.1</version>
</dependency>
```

> Package names are prefixed (`pipeops-rexec`, `PipeOps.Rexec`, `pipeopshq/rexec`) because bare `rexec` is taken on several registries. Rust and Python still **import** as `rexec`.

---

## 3. Image aliases (critical)

Hosted Rexec validates images against a fixed catalog of **aliases**, not arbitrary Docker Hub tags.

| Do use | Don’t use on hosted |
|--------|---------------------|
| `ubuntu`, `ubuntu-24`, `ubuntu-22` | `ubuntu:24.04` |
| `debian`, `alpine`, `fedora` | `debian:latest` |
| `archlinux`, … | random Hub tags |

Full catalog: `GET /api/images` on your instance.

Create often returns immediately with `status: "creating"` (async provision). If you need a ready shell, poll `get(id)` until `status === "running"`.

---

## 4. Minimal happy path

Every language does the same four steps:

1. **List** sandboxes  
2. **Create** with `image: "ubuntu"` (optional `name`)  
3. **Get** by id  
4. **Delete** when finished  

There is **no** first-class `exec()` HTTP API. Run commands via the **terminal WebSocket** (see [SDK.md](SDK.md#terminal-websocket)).

### JavaScript / TypeScript (primary)

```bash
npm install pipeops-rexec
# Node WebSockets: npm install ws
```

```typescript
import { RexecClient } from 'pipeops-rexec';

const client = new RexecClient({
  baseURL: process.env.REXEC_URL!,
  token: process.env.REXEC_TOKEN!,
});

const list = await client.containers.list(); // always an array (SDK normalizes API wrapper)
const c = await client.containers.create({ image: 'ubuntu', name: 'demo' });
console.log(c.id, c.status); // often "creating" at first

const got = await client.containers.get(c.id);
await client.containers.delete(c.id);
```

### Python (primary)

```bash
pip install pipeops-rexec
```

```python
import asyncio
import os
from rexec import RexecClient

async def main():
    async with RexecClient(os.environ["REXEC_URL"], os.environ["REXEC_TOKEN"]) as client:
        print(await client.containers.list())
        c = await client.containers.create(image="ubuntu", name="demo")
        print(c.id, c.status)
        await client.containers.get(c.id)
        await client.containers.delete(c.id)

asyncio.run(main())
```

### Other languages (same flow)

<details>
<summary>Go</summary>

```bash
go get github.com/PipeOpsHQ/rexec-go@v1.0.1
```

```go
client := rexec.NewClient(os.Getenv("REXEC_URL"), os.Getenv("REXEC_TOKEN"))
c, _ := client.Containers.Create(ctx, &rexec.CreateContainerRequest{Image: "ubuntu", Name: "demo"})
_ = client.Containers.Delete(ctx, c.ID)
```

</details>

<details>
<summary>Rust</summary>

```bash
cargo add pipeops-rexec
```

```rust
use rexec::{CreateContainerRequest, RexecClient};
let client = RexecClient::new(url, token);
let c = client.containers().create(CreateContainerRequest::new("ubuntu").name("demo")).await?;
client.containers().delete(&c.id).await?;
```

</details>

<details>
<summary>Ruby</summary>

```bash
gem install pipeops-rexec
```

```ruby
require "rexec"
client = Rexec::Client.new(ENV["REXEC_URL"], ENV["REXEC_TOKEN"])
c = client.containers.create(image: "ubuntu", name: "demo")
client.containers.delete(c.id)
```

</details>

<details>
<summary>C# / .NET</summary>

```bash
dotnet add package PipeOps.Rexec
```

```csharp
using Rexec;
using var client = new RexecClient(url, token);
var c = await client.Containers.CreateAsync(new CreateContainerRequest("ubuntu") { Name = "demo" });
await client.Containers.DeleteAsync(c!.Id);
```

</details>

<details>
<summary>Java / Kotlin</summary>

```java
RexecClient client = new RexecClient(url, token);
Container c = client.containers().create(new CreateContainerRequest("ubuntu").setName("demo"));
client.containers().delete(c.getId());
```

</details>

<details>
<summary>PHP</summary>

```bash
composer require pipeopshq/rexec
```

```php
$client = new Rexec\RexecClient(getenv('REXEC_URL'), getenv('REXEC_TOKEN'));
$c = $client->containers()->create('ubuntu', ['name' => 'demo']);
$client->containers()->delete($c->id);
```

</details>

Full copy-paste samples for every language: [SDK.md](SDK.md#language-examples).

---

## 5. What the SDKs cover

| Surface | Capabilities |
|---------|----------------|
| **Containers** | list, create, get, start, stop, delete |
| **Files** | list dir, read, write (often base64), mkdir (JS+), delete |
| **Terminal** | WebSocket connect, `write` / `onData`, resize, close |

Raw HTTP equivalents and error types live in the [API Reference](SDK.md).

---

## 6. Next steps

| Topic | Where |
|-------|--------|
| Auth, lifecycle, files, terminal protocol, errors | [SDK.md](SDK.md) |
| Publish / CI | [SDK_PUBLISHING.md](SDK_PUBLISHING.md) |
| Per-language README | [`sdk/*/README.md`](../sdk/) |
| E2E runners | [`scripts/sdk-e2e/`](../scripts/sdk-e2e/) |
| Product | [rexec.sh](https://rexec.sh) |

### Registry links

- [npm `pipeops-rexec`](https://www.npmjs.com/package/pipeops-rexec)
- [PyPI `pipeops-rexec`](https://pypi.org/project/pipeops-rexec/)
- [crates.io `pipeops-rexec`](https://crates.io/crates/pipeops-rexec)
- [RubyGems `pipeops-rexec`](https://rubygems.org/gems/pipeops-rexec)
- [NuGet `PipeOps.Rexec`](https://www.nuget.org/packages/PipeOps.Rexec)
- [Maven `io/pipeops/rexec`](https://repo1.maven.org/maven2/io/pipeops/rexec/)
- [Go module](https://github.com/PipeOpsHQ/rexec-go)
- [PHP source](https://github.com/PipeOpsHQ/rexec-php)
