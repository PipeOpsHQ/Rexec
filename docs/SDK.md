# Rexec SDK Documentation

Official client libraries for the Rexec API (sandboxes, files, and terminals).

> **Verified E2E** against a live Rexec instance (`list` → `create` → `get` → `delete`).  
> Source smoke tests: [`scripts/sdk-e2e/`](../scripts/sdk-e2e/).

## Available SDKs (v1.0.1)

| Language | Package | Install |
|----------|---------|---------|
| **JavaScript / TypeScript** | [pipeops-rexec](https://www.npmjs.com/package/pipeops-rexec) | `npm install pipeops-rexec` |
| **Python** | [pipeops-rexec](https://pypi.org/project/pipeops-rexec/) | `pip install pipeops-rexec` |
| **Go** | [github.com/PipeOpsHQ/rexec-go](https://github.com/PipeOpsHQ/rexec-go) | `go get github.com/PipeOpsHQ/rexec-go@v1.0.1` |
| **Rust** | [pipeops-rexec](https://crates.io/crates/pipeops-rexec) | `cargo add pipeops-rexec` |
| **Ruby** | [pipeops-rexec](https://rubygems.org/gems/pipeops-rexec) | `gem install pipeops-rexec` |
| **C# / .NET** | [PipeOps.Rexec](https://www.nuget.org/packages/PipeOps.Rexec) | `dotnet add package PipeOps.Rexec` |
| **Java / Kotlin** | `io.pipeops:rexec` | Maven/Gradle (source in repo; Maven Central when secrets set) |
| **PHP** | [pipeopshq/rexec](https://packagist.org/packages/pipeopshq/rexec) | `composer require pipeopshq/rexec` |

Publishing is **GitHub Actions only** — see [SDK_PUBLISHING.md](SDK_PUBLISHING.md).

## Auth

1. Open your Rexec UI → **Settings** → **API Tokens** → generate a token  
2. Or use guest login for short-lived tests:

```bash
export REXEC_URL=https://rexec.sh   # or your instance
export REXEC_TOKEN=$(curl -sS -X POST "$REXEC_URL/api/auth/guest" \
  -H 'Content-Type: application/json' \
  -d '{"username":"sdk_demo","email":"you@example.com"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
```

Use the token as `Authorization: Bearer <token>` (all SDKs set this for you).

## Images

Hosted Rexec accepts **image aliases**, not arbitrary Docker Hub tags:

- Prefer: `ubuntu`, `ubuntu-24`, `debian`, `alpine`, `fedora`, …
- Avoid: `ubuntu:24.04` (may return `unsupported image type`)

List aliases with `GET /api/images` on your instance.

## Common API surface

Every SDK wraps the same REST resources:

| Operation | HTTP |
|-----------|------|
| List sandboxes | `GET /api/containers` → `{ containers: [...], count, limit }` |
| Create | `POST /api/containers` `{ image, name? }` (often async `status: creating`) |
| Get | `GET /api/containers/:id` |
| Start / stop | `POST .../start` · `POST .../stop` |
| Delete | `DELETE /api/containers/:id` |
| Files list | `GET /api/containers/:id/files/list?path=` |
| Terminal | WebSocket `/ws/terminal/:id` |

There is **no** high-level `exec()` helper in the published SDKs; run commands via the terminal WebSocket (or your own HTTP tooling).

---

## JavaScript / TypeScript

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

const list = await client.containers.list();
const container = await client.containers.create({
  image: 'ubuntu',
  name: 'demo',
});
console.log(container.id, container.status);

const got = await client.containers.get(container.id);
await client.containers.delete(container.id);

// Terminal (browser WebSocket or `ws` in Node)
const term = await client.terminal.connect(container.id);
term.onData((data) => process.stdout.write(String(data)));
term.write('echo hello\n');
term.close();
```

## Python

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

## Go

```bash
go get github.com/PipeOpsHQ/rexec-go@v1.0.1
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

    list, _ := client.Containers.List(ctx)
    fmt.Println("count", len(list))

    c, err := client.Containers.Create(ctx, &rexec.CreateContainerRequest{
        Image: "ubuntu",
        Name:  "demo",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(c.ID, c.Status)
    _, _ = client.Containers.Get(ctx, c.ID)
    _ = client.Containers.Delete(ctx, c.ID)
}
```

## Rust

```bash
cargo add pipeops-rexec tokio --features tokio/full
```

```rust
use rexec::{CreateContainerRequest, RexecClient};

#[tokio::main]
async fn main() -> Result<(), rexec::Error> {
    let client = RexecClient::new(
        std::env::var("REXEC_URL").unwrap(),
        std::env::var("REXEC_TOKEN").unwrap(),
    );

    let list = client.containers().list().await?;
    println!("count {}", list.len());

    let c = client
        .containers()
        .create(CreateContainerRequest::new("ubuntu").name("demo"))
        .await?;
    println!("{} {}", c.id, c.status);
    client.containers().get(&c.id).await?;
    client.containers().delete(&c.id).await?;
    Ok(())
}
```

Import crate name: `use rexec::...` (package name is `pipeops-rexec`).

## Ruby

```bash
gem install pipeops-rexec   # Ruby >= 3.0
```

```ruby
require "rexec"

client = Rexec::Client.new(ENV["REXEC_URL"], ENV["REXEC_TOKEN"])
puts client.containers.list.length

c = client.containers.create(image: "ubuntu", name: "demo")
puts "#{c.id} #{c.status}"
client.containers.get(c.id)
client.containers.delete(c.id)
```

## C# / .NET

```bash
dotnet add package PipeOps.Rexec
```

```csharp
using Rexec;

using var client = new RexecClient(
    Environment.GetEnvironmentVariable("REXEC_URL")!,
    Environment.GetEnvironmentVariable("REXEC_TOKEN")!);

var list = await client.Containers.ListAsync();
var c = await client.Containers.CreateAsync(new CreateContainerRequest("ubuntu") {
    Name = "demo"
});
Console.WriteLine($"{c?.Id} {c?.Status}");
await client.Containers.GetAsync(c!.Id);
await client.Containers.DeleteAsync(c.Id);
```

## Java / Kotlin

```xml
<dependency>
  <groupId>io.pipeops</groupId>
  <artifactId>rexec</artifactId>
  <version>1.0.1</version>
</dependency>
```

```java
import io.pipeops.rexec.*;

RexecClient client = new RexecClient(
    System.getenv("REXEC_URL"),
    System.getenv("REXEC_TOKEN"));

System.out.println(client.containers().list().size());
Container c = client.containers().create(
    new CreateContainerRequest("ubuntu").setName("demo"));
System.out.println(c.getId() + " " + c.getStatus());
client.containers().get(c.getId());
client.containers().delete(c.getId());
```

Kotlin can use the same JAR:

```kotlin
val client = RexecClient(System.getenv("REXEC_URL"), System.getenv("REXEC_TOKEN"))
val c = client.containers().create(CreateContainerRequest("ubuntu").setName("demo"))
client.containers().delete(c.id)
```

Install from source until Maven Central is configured: `cd sdk/java && mvn install -DskipTests`.

## PHP

```bash
composer require pipeopshq/rexec
```

```php
<?php
require 'vendor/autoload.php';

use Rexec\RexecClient;

$client = new RexecClient(getenv('REXEC_URL'), getenv('REXEC_TOKEN'));
echo count($client->containers()->list()), PHP_EOL;

$c = $client->containers()->create('ubuntu', ['name' => 'demo']);
echo $c->id, ' ', $c->status, PHP_EOL;
$client->containers()->get($c->id);
$client->containers()->delete($c->id);
```

Packagist syncs from GitHub once the package is submitted (see [SDK_PUBLISHING.md](SDK_PUBLISHING.md)).

## Error handling

| SDK | Type |
|-----|------|
| JS | `RexecError` (`statusCode`, `message`) |
| Python | `RexecAPIError`, `RexecConnectionError` |
| Go | `*rexec.APIError` |
| Rust | `rexec::Error` |
| Ruby | `Rexec::APIError` |
| .NET | `RexecException` |
| Java | `RexecException` |
| PHP | `RexecException` |

## End-to-end smoke tests

```bash
cd scripts/sdk-e2e
# see README.md — list → create → get → delete for each language
```

## Further reading

- [Getting started](SDK_GETTING_STARTED.md)
- [Publishing / CI](SDK_PUBLISHING.md)
- Per-language READMEs under [`sdk/`](../sdk/)
