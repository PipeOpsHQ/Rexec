<script lang="ts">
    import StatusIcon from "./icons/StatusIcon.svelte";

    export let onback: (() => void) | undefined = undefined;

    let copiedCommand = "";
    let activeTab = "javascript";

    function copyToClipboard(text: string, id: string) {
        navigator.clipboard.writeText(text);
        copiedCommand = id;
        setTimeout(() => {
            copiedCommand = "";
        }, 2000);
    }

    function handleBack() {
        if (onback) onback();
        else if (typeof window !== "undefined") window.history.back();
    }

    const sdks: {
        id: string;
        name: string;
        install: string;
        registry: string;
        github: string;
    }[] = [
        {
            id: "javascript",
            name: "JavaScript / TypeScript",
            install: "npm install pipeops-rexec",
            registry: "https://www.npmjs.com/package/pipeops-rexec",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/js",
        },
        {
            id: "python",
            name: "Python",
            install: "pip install pipeops-rexec",
            registry: "https://pypi.org/project/pipeops-rexec/",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/python",
        },
        {
            id: "go",
            name: "Go",
            install: "go get github.com/PipeOpsHQ/rexec-go@v1.0.1",
            registry: "https://github.com/PipeOpsHQ/rexec-go",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/go",
        },
        {
            id: "rust",
            name: "Rust",
            install: "cargo add pipeops-rexec",
            registry: "https://crates.io/crates/pipeops-rexec",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/rust",
        },
        {
            id: "ruby",
            name: "Ruby",
            install: "gem install pipeops-rexec",
            registry: "https://rubygems.org/gems/pipeops-rexec",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/ruby",
        },
        {
            id: "dotnet",
            name: "C# / .NET",
            install: "dotnet add package PipeOps.Rexec",
            registry: "https://www.nuget.org/packages/PipeOps.Rexec",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/dotnet",
        },
        {
            id: "java",
            name: "Java / Kotlin",
            install: "io.pipeops:rexec:1.0.1 (Maven/Gradle)",
            registry: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/java",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/java",
        },
        {
            id: "php",
            name: "PHP",
            install: "composer require pipeopshq/rexec",
            registry: "https://packagist.org/packages/pipeopshq/rexec",
            github: "https://github.com/PipeOpsHQ/rexec/tree/main/sdk/php",
        },
    ];

    const codeExamples: Record<string, string> = {
        javascript: `import { RexecClient } from 'pipeops-rexec';

const client = new RexecClient({
  baseURL: process.env.REXEC_URL,
  token: process.env.REXEC_TOKEN,
});

const list = await client.containers.list();
const container = await client.containers.create({
  image: 'ubuntu',
  name: 'demo',
});
console.log(container.id, container.status);
await client.containers.get(container.id);
await client.containers.delete(container.id);`,
        python: `import asyncio, os
from rexec import RexecClient

async def main():
    async with RexecClient(os.environ["REXEC_URL"], os.environ["REXEC_TOKEN"]) as client:
        print(await client.containers.list())
        c = await client.containers.create(image="ubuntu", name="demo")
        print(c.id, c.status)
        await client.containers.get(c.id)
        await client.containers.delete(c.id)

asyncio.run(main())`,
        go: `package main

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
    fmt.Println(len(list))
    c, err := client.Containers.Create(ctx, &rexec.CreateContainerRequest{
        Image: "ubuntu",
        Name:  "demo",
    })
    if err != nil { panic(err) }
    fmt.Println(c.ID, c.Status)
    client.Containers.Get(ctx, c.ID)
    client.Containers.Delete(ctx, c.ID)
}`,
        rust: `use rexec::{CreateContainerRequest, RexecClient};

#[tokio::main]
async fn main() -> Result<(), rexec::Error> {
    let client = RexecClient::new(
        std::env::var("REXEC_URL").unwrap(),
        std::env::var("REXEC_TOKEN").unwrap(),
    );
    let list = client.containers().list().await?;
    println!("{}", list.len());
    let c = client.containers()
        .create(CreateContainerRequest::new("ubuntu").name("demo"))
        .await?;
    println!("{} {}", c.id, c.status);
    client.containers().delete(&c.id).await?;
    Ok(())
}`,
        ruby: `require 'rexec'

client = Rexec::Client.new(ENV['REXEC_URL'], ENV['REXEC_TOKEN'])
puts client.containers.list.length
c = client.containers.create(image: 'ubuntu', name: 'demo')
puts "#{c.id} #{c.status}"
client.containers.get(c.id)
client.containers.delete(c.id)`,
        dotnet: `using Rexec;

using var client = new RexecClient(
    Environment.GetEnvironmentVariable("REXEC_URL")!,
    Environment.GetEnvironmentVariable("REXEC_TOKEN")!);

var list = await client.Containers.ListAsync();
var c = await client.Containers.CreateAsync(new CreateContainerRequest("ubuntu") {
    Name = "demo"
});
Console.WriteLine($"{c?.Id} {c?.Status}");
await client.Containers.GetAsync(c!.Id);
await client.Containers.DeleteAsync(c.Id);`,
        java: `import io.pipeops.rexec.*;

RexecClient client = new RexecClient(
    System.getenv("REXEC_URL"),
    System.getenv("REXEC_TOKEN"));

System.out.println(client.containers().list().size());
Container c = client.containers().create(
    new CreateContainerRequest("ubuntu").setName("demo"));
System.out.println(c.getId() + " " + c.getStatus());
client.containers().get(c.getId());
client.containers().delete(c.getId());`,
        php: `<?php
require 'vendor/autoload.php';
use Rexec\\RexecClient;

$client = new RexecClient(getenv('REXEC_URL'), getenv('REXEC_TOKEN'));
echo count($client->containers()->list()), PHP_EOL;
$c = $client->containers()->create('ubuntu', ['name' => 'demo']);
echo $c->id, ' ', $c->status, PHP_EOL;
$client->containers()->get($c->id);
$client->containers()->delete($c->id);`,
    };

    $: activeSdk = sdks.find((s) => s.id === activeTab) ?? sdks[0];
    $: activeCode = codeExamples[activeTab] ?? "";
</script>

<div class="docs-page">
    <button class="back-btn" type="button" onclick={handleBack}>
        <span class="back-icon">←</span>
        <span>Back</span>
    </button>

    <div class="docs-content">
        <header class="docs-header">
            <div class="header-icon">
                <StatusIcon status="code" size={48} />
            </div>
            <h1>Rexec SDKs</h1>
            <p class="subtitle">
                Official clients for the Rexec API (v1.0.1) — list, create, get,
                and delete sandboxes from your code
            </p>
        </header>

        <section class="docs-section">
            <h2>Install</h2>
            <p>
                Use image aliases such as <code>ubuntu</code>,
                <code>debian</code>, or <code>alpine</code> (not
                <code>ubuntu:24.04</code> on hosted Rexec). Set
                <code>REXEC_URL</code> and <code>REXEC_TOKEN</code>.
            </p>

            <div class="sdk-list">
                {#each sdks as sdk (sdk.id)}
                    <div
                        class="sdk-row"
                        class:active={activeTab === sdk.id}
                    >
                        <button
                            type="button"
                            class="sdk-select"
                            onclick={() => (activeTab = sdk.id)}
                        >
                            <strong>{sdk.name}</strong>
                            <code>{sdk.install}</code>
                        </button>
                        <div class="sdk-actions">
                            <button
                                type="button"
                                class="copy-btn"
                                onclick={() =>
                                    copyToClipboard(sdk.install, sdk.id)}
                            >
                                {copiedCommand === sdk.id ? "Copied" : "Copy"}
                            </button>
                            <a
                                href={sdk.registry}
                                target="_blank"
                                rel="noreferrer">Registry</a
                            >
                            <a
                                href={sdk.github}
                                target="_blank"
                                rel="noreferrer">Source</a
                            >
                        </div>
                    </div>
                {/each}
            </div>
        </section>

        <section class="docs-section">
            <div class="section-head">
                <h2>{activeSdk.name} example</h2>
                <button
                    type="button"
                    class="copy-btn"
                    onclick={() => copyToClipboard(activeCode, "code")}
                >
                    {copiedCommand === "code" ? "Copied" : "Copy code"}
                </button>
            </div>
            <div class="code-block">
                <pre><code>{activeCode}</code></pre>
            </div>
        </section>

        <section class="docs-section">
            <h2>References</h2>
            <ul class="ref-list">
                <li>
                    <a
                        href="https://github.com/PipeOpsHQ/rexec/blob/main/docs/SDK.md"
                        target="_blank"
                        rel="noreferrer">Full SDK reference (docs/SDK.md)</a
                    >
                </li>
                <li>
                    <a
                        href="https://github.com/PipeOpsHQ/rexec/blob/main/docs/SDK_GETTING_STARTED.md"
                        target="_blank"
                        rel="noreferrer">Getting started</a
                    >
                </li>
                <li>
                    <a
                        href="https://github.com/PipeOpsHQ/rexec/tree/main/scripts/sdk-e2e"
                        target="_blank"
                        rel="noreferrer">E2E smoke tests</a
                    >
                </li>
            </ul>
        </section>
    </div>
</div>

<style>
    .docs-page {
        min-height: 100vh;
        background: var(--bg);
        padding: 24px;
        overflow-y: auto;
    }

    .back-btn {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 8px 14px;
        background: transparent;
        border: 1px solid var(--border);
        border-radius: 6px;
        color: var(--text-muted);
        font-size: 13px;
        font-family: var(--font-mono);
        cursor: pointer;
        transition: all 0.15s ease;
        margin-bottom: 24px;
    }

    .back-btn:hover {
        border-color: var(--accent);
        color: var(--accent);
    }

    .back-icon {
        font-size: 16px;
    }

    .docs-content {
        max-width: 900px;
        margin: 0 auto;
    }

    .docs-header {
        text-align: center;
        margin-bottom: 48px;
        padding-bottom: 32px;
        border-bottom: 1px solid var(--border);
    }

    .header-icon {
        margin-bottom: 16px;
    }

    .header-icon :global(svg) {
        color: var(--accent);
    }

    .docs-header h1 {
        font-size: 36px;
        margin: 0 0 12px 0;
        background: linear-gradient(135deg, var(--accent), #00d4ff);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .subtitle {
        font-size: 16px;
        color: var(--text-muted);
        margin: 0;
        line-height: 1.5;
    }

    .docs-section {
        margin-bottom: 48px;
    }

    .docs-section h2 {
        font-size: 20px;
        margin: 0 0 16px 0;
        color: var(--text);
        text-transform: uppercase;
        letter-spacing: 1px;
    }

    .docs-section p {
        font-size: 14px;
        color: var(--text-muted);
        line-height: 1.7;
        margin: 0 0 16px 0;
    }

    .docs-section code {
        font-family: var(--font-mono);
        font-size: 12px;
        color: var(--accent);
        background: var(--bg-secondary);
        padding: 2px 6px;
        border-radius: 4px;
    }

    .sdk-list {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .sdk-row {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        padding: 14px 16px;
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
    }

    .sdk-row.active {
        border-color: var(--accent);
        box-shadow: 0 0 0 1px rgba(var(--accent-rgb, 0, 212, 255), 0.25);
    }

    .sdk-select {
        flex: 1;
        min-width: 220px;
        text-align: left;
        background: transparent;
        border: 0;
        color: inherit;
        cursor: pointer;
        padding: 0;
    }

    .sdk-select strong {
        display: block;
        font-size: 14px;
        color: var(--text);
        margin-bottom: 4px;
    }

    .sdk-select code {
        display: block;
        font-size: 12px;
        color: var(--accent);
        word-break: break-all;
        background: transparent;
        padding: 0;
    }

    .sdk-actions {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 10px;
        font-size: 12px;
    }

    .sdk-actions a {
        color: var(--accent);
        text-decoration: none;
    }

    .sdk-actions a:hover {
        text-decoration: underline;
    }

    .copy-btn {
        padding: 6px 10px;
        background: transparent;
        border: 1px solid var(--border);
        border-radius: 6px;
        color: var(--text-muted);
        font-size: 12px;
        font-family: var(--font-mono);
        cursor: pointer;
    }

    .copy-btn:hover {
        border-color: var(--accent);
        color: var(--accent);
    }

    .section-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        margin-bottom: 16px;
    }

    .section-head h2 {
        margin: 0;
    }

    .code-block {
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
        overflow: auto;
    }

    .code-block pre {
        margin: 0;
        padding: 16px;
        font-size: 12px;
        line-height: 1.55;
        color: var(--text);
        font-family: var(--font-mono);
        white-space: pre;
    }

    .code-block code {
        background: transparent;
        padding: 0;
        color: inherit;
        font-size: inherit;
    }

    .ref-list {
        margin: 0;
        padding-left: 1.25rem;
        color: var(--text-muted);
        font-size: 14px;
        line-height: 1.8;
    }

    .ref-list a {
        color: var(--accent);
    }
</style>
