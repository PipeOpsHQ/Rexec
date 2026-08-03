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
            github: "https://github.com/PipeOpsHQ/Rexec/tree/main/sdk/js",
        },
        {
            id: "python",
            name: "Python",
            install: "pip install pipeops-rexec",
            registry: "https://pypi.org/project/pipeops-rexec/",
            github: "https://github.com/PipeOpsHQ/Rexec/tree/main/sdk/python",
        },
        {
            id: "go",
            name: "Go",
            install: "go get github.com/PipeOpsHQ/rexec-go@v1.0.1",
            registry: "https://github.com/PipeOpsHQ/rexec-go",
            github: "https://github.com/PipeOpsHQ/Rexec/tree/main/sdk/go",
        },
        {
            id: "rust",
            name: "Rust",
            install: "cargo add pipeops-rexec",
            registry: "https://crates.io/crates/pipeops-rexec",
            github: "https://github.com/PipeOpsHQ/Rexec/tree/main/sdk/rust",
        },
        {
            id: "ruby",
            name: "Ruby",
            install: "gem install pipeops-rexec",
            registry: "https://rubygems.org/gems/pipeops-rexec",
            github: "https://github.com/PipeOpsHQ/Rexec/tree/main/sdk/ruby",
        },
        {
            id: "dotnet",
            name: "C# / .NET",
            install: "dotnet add package PipeOps.Rexec",
            registry: "https://www.nuget.org/packages/PipeOps.Rexec",
            github: "https://github.com/PipeOpsHQ/Rexec/tree/main/sdk/dotnet",
        },
        {
            id: "java",
            name: "Java / Kotlin",
            // Maven: <dependency>…</dependency> · Gradle: implementation '…'
            install: "io.pipeops:rexec:1.0.1 (Maven/Gradle)",
            registry: "https://repo1.maven.org/maven2/io/pipeops/rexec/",
            github: "https://github.com/PipeOpsHQ/Rexec/tree/main/sdk/java",
        },
        {
            id: "php",
            name: "PHP",
            install: "composer require pipeopshq/rexec",
            // Packagist is the registry; monorepo source stays under GitHub Source link
            registry: "https://packagist.org/packages/pipeopshq/rexec",
            github: "https://github.com/PipeOpsHQ/rexec-php",
        },
    ];

    const codeExamples: Record<string, string> = {
        javascript: `import { RexecClient } from 'pipeops-rexec';

const client = new RexecClient({
  baseURL: process.env.REXEC_URL,
  token: process.env.REXEC_TOKEN,
});

// List sandboxes (SDK normalizes API list wrapper)
const list = await client.containers.list();

// Create a sandbox (image alias — not ubuntu:24.04 on hosted)
const sandbox = await client.containers.create({
  image: 'ubuntu',
  name: 'demo',
});
console.log(sandbox.id, sandbox.status); // may be "creating"

await client.containers.get(sandbox.id);
await client.containers.delete(sandbox.id);`,
        python: `import asyncio, os
from rexec import RexecClient

async def main():
    async with RexecClient(os.environ["REXEC_URL"], os.environ["REXEC_TOKEN"]) as client:
        print(await client.containers.list())
        sandbox = await client.containers.create(image="ubuntu", name="demo")
        print(sandbox.id, sandbox.status)
        await client.containers.get(sandbox.id)
        await client.containers.delete(sandbox.id)

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
    if err != nil {
        panic(err)
    }
    fmt.Println(c.ID, c.Status)
    _, _ = client.Containers.Get(ctx, c.ID)
    _ = client.Containers.Delete(ctx, c.ID)
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
    let c = client
        .containers()
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

    // Single-line-friendly guest auth (no awkward soft-wrap mid-flag)
    const guestAuthSnippet = [
        "export REXEC_URL=https://rexec.sh",
        "export REXEC_TOKEN=$(curl -sS -X POST \\",
        '  "$REXEC_URL/api/auth/guest" \\',
        "  -H 'Content-Type: application/json' \\",
        `  -d '{"username":"docs_demo","email":"you@example.com"}' \\`,
        "  | python3 -c \"import sys,json; print(json.load(sys.stdin)['token'])\")",
    ].join("\n");
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
                and delete <strong>sandboxes</strong> from your code. Same REST +
                WebSocket surface in JS, Python, Go, Rust, Ruby, .NET, Java/Kotlin,
                and PHP.
            </p>
        </header>

        <section class="docs-section">
            <h2>Auth</h2>
            <p>
                Prefer an <strong>API token</strong> from
                <strong>Settings → API Tokens</strong>. For short smoke tests, use a
                guest JWT:
            </p>
            <div class="code-block">
                <div class="code-toolbar">
                    <span class="code-label">shell</span>
                    <button
                        type="button"
                        class="copy-btn"
                        onclick={() => copyToClipboard(guestAuthSnippet, "auth")}
                    >
                        {copiedCommand === "auth" ? "Copied" : "Copy"}
                    </button>
                </div>
                <pre><code>{guestAuthSnippet}</code></pre>
            </div>
            <p class="muted">
                Full docs:
                <a
                    href="https://github.com/PipeOpsHQ/Rexec/blob/main/docs/SDK_GETTING_STARTED.md"
                    target="_blank"
                    rel="noreferrer">Quick start</a
                >
                ·
                <a
                    href="https://github.com/PipeOpsHQ/Rexec/blob/main/docs/SDK.md"
                    target="_blank"
                    rel="noreferrer">API reference</a
                >
                ·
                <a
                    href="https://docs.pipeops.io/docs/rexec/overview"
                    target="_blank"
                    rel="noreferrer">PipeOps · Sandboxes</a
                >
            </p>
        </section>

        <section class="docs-section">
            <h2>Install</h2>

            <ul class="callouts">
                <li>
                    <span class="callout-label">Images</span>
                    Use aliases like <code>ubuntu</code>, <code>debian</code>, or
                    <code>alpine</code> — not
                    <code class="warn">ubuntu:24.04</code> on hosted Rexec.
                </li>
                <li>
                    <span class="callout-label">Create</span>
                    May return <code>status: "creating"</code>. Poll
                    <code>get</code> until <code>"running"</code> if you need a
                    ready shell.
                </li>
                <li>
                    <span class="callout-label">Commands</span>
                    No primary HTTP <code>exec()</code> — use the terminal
                    WebSocket for interactive or scripted shell input.
                </li>
            </ul>

            <p class="muted tip">
                Select a language to preview a sample. <code>containers.*</code> methods
                manage <strong>sandboxes</strong>.
            </p>

            <div class="sdk-list" role="tablist" aria-label="SDK languages">
                {#each sdks as sdk (sdk.id)}
                    <div
                        class="sdk-row"
                        class:active={activeTab === sdk.id}
                    >
                        <button
                            type="button"
                            class="sdk-select"
                            role="tab"
                            id={`sdk-tab-${sdk.id}`}
                            aria-selected={activeTab === sdk.id}
                            aria-controls="sdk-example-panel"
                            tabindex={activeTab === sdk.id ? 0 : -1}
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
                                    copyToClipboard(
                                        // Copy bare coordinates for Java (no parenthetical)
                                        sdk.id === "java"
                                            ? "io.pipeops:rexec:1.0.1"
                                            : sdk.install,
                                        sdk.id,
                                    )}
                            >
                                {copiedCommand === sdk.id ? "Copied" : "Copy"}
                            </button>
                            <a
                                class="link-btn"
                                href={sdk.registry}
                                target="_blank"
                                rel="noreferrer">Registry</a
                            >
                            <a
                                class="link-btn"
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
                <h2 id="sdk-example-heading">{activeSdk.name} example</h2>
                <button
                    type="button"
                    class="copy-btn"
                    onclick={() => copyToClipboard(activeCode, "code")}
                >
                    {copiedCommand === "code" ? "Copied" : "Copy code"}
                </button>
            </div>
            {#if activeSdk.id === "java"}
                <p class="muted tip">
                    Maven:
                    <code
                        >&lt;dependency&gt;…&lt;artifactId&gt;rexec&lt;/artifactId&gt;…&lt;/dependency&gt;</code
                    >
                    · Gradle:
                    <code>implementation 'io.pipeops:rexec:1.0.1'</code>
                </p>
            {/if}
            {#if activeSdk.id === "php"}
                <p class="muted tip">
                    If Packagist is not linked yet:
                    <code
                        >composer config repositories.rexec-php vcs https://github.com/PipeOpsHQ/rexec-php</code
                    >
                </p>
            {/if}
            <div
                class="code-block"
                role="tabpanel"
                id="sdk-example-panel"
                aria-labelledby={`sdk-tab-${activeSdk.id}`}
            >
                <div class="code-toolbar">
                    <span class="code-label">{activeSdk.id}</span>
                </div>
                <pre><code>{activeCode}</code></pre>
            </div>
        </section>

        <section class="docs-section">
            <h2>References</h2>
            <ul class="ref-list">
                <li>
                    <a
                        href="https://github.com/PipeOpsHQ/Rexec/blob/main/docs/SDK.md"
                        target="_blank"
                        rel="noreferrer">Full SDK API reference</a
                    >
                </li>
                <li>
                    <a
                        href="https://github.com/PipeOpsHQ/Rexec/blob/main/docs/SDK_GETTING_STARTED.md"
                        target="_blank"
                        rel="noreferrer">Getting started guide</a
                    >
                </li>
                <li>
                    <a
                        href="https://docs.pipeops.io/docs/rexec/overview"
                        target="_blank"
                        rel="noreferrer">PipeOps docs · Rexec sandboxes</a
                    >
                </li>
                <li>
                    <a
                        href="https://github.com/PipeOpsHQ/Rexec/tree/main/scripts/sdk-e2e"
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
        padding-bottom: 48px;
    }

    .docs-header {
        text-align: center;
        margin-bottom: 40px;
        padding-bottom: 28px;
        border-bottom: 1px solid var(--border);
    }

    .header-icon {
        margin-bottom: 16px;
    }

    .header-icon :global(svg) {
        color: var(--accent);
    }

    .docs-header h1 {
        font-size: clamp(28px, 5vw, 36px);
        margin: 0 0 12px 0;
        background: linear-gradient(135deg, var(--accent), #00d4ff);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .subtitle {
        font-size: 15px;
        color: var(--text-muted);
        margin: 0 auto;
        line-height: 1.6;
        max-width: 40rem;
    }

    .subtitle strong {
        color: var(--text);
        font-weight: 600;
        -webkit-text-fill-color: var(--text);
    }

    .docs-section {
        margin-bottom: 40px;
    }

    .docs-section h2 {
        font-size: 13px;
        margin: 0 0 14px 0;
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.08em;
        font-weight: 600;
    }

    .docs-section p {
        font-size: 14px;
        color: var(--text-muted);
        line-height: 1.7;
        margin: 0 0 14px 0;
    }

    .docs-section p strong {
        color: var(--text);
        font-weight: 600;
    }

    .docs-section :not(pre) > code,
    .docs-section p code,
    .callouts code {
        font-family: var(--font-mono);
        font-size: 12px;
        color: var(--accent);
        background: var(--bg-secondary);
        padding: 2px 6px;
        border-radius: 4px;
        white-space: nowrap;
    }

    .callouts code.warn {
        color: #f87171;
    }

    .muted {
        font-size: 13px;
        color: var(--text-muted);
    }

    .muted a {
        color: var(--accent);
        text-decoration: none;
    }

    .muted a:hover {
        text-decoration: underline;
    }

    .tip {
        margin-top: 4px;
        margin-bottom: 16px !important;
    }

    .callouts {
        list-style: none;
        margin: 0 0 16px 0;
        padding: 0;
        display: grid;
        gap: 10px;
    }

    .callouts li {
        display: grid;
        grid-template-columns: auto 1fr;
        gap: 10px 12px;
        align-items: start;
        padding: 12px 14px;
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
        font-size: 13px;
        color: var(--text-muted);
        line-height: 1.55;
    }

    .callout-label {
        font-family: var(--font-mono);
        font-size: 10px;
        font-weight: 700;
        letter-spacing: 0.06em;
        text-transform: uppercase;
        color: var(--accent);
        padding-top: 2px;
        min-width: 4.5rem;
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
        transition:
            border-color 0.15s ease,
            box-shadow 0.15s ease;
    }

    .sdk-row.active {
        border-color: var(--accent);
        /* Fallback for browsers without color-mix (e.g. Safari < 16.2) */
        box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.35);
        box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 35%, transparent);
    }

    .sdk-select {
        flex: 1;
        min-width: min(100%, 240px);
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
        word-break: break-word;
        overflow-wrap: anywhere;
        background: transparent;
        padding: 0;
        white-space: normal;
    }

    .sdk-actions {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 8px;
        font-size: 12px;
    }

    .link-btn {
        color: var(--accent);
        text-decoration: none;
        padding: 6px 8px;
        border-radius: 6px;
        border: 1px solid transparent;
    }

    .link-btn:hover {
        text-decoration: none;
        border-color: rgba(0, 212, 255, 0.4);
        background: rgba(0, 212, 255, 0.08);
        border-color: color-mix(in srgb, var(--accent) 40%, transparent);
        background: color-mix(in srgb, var(--accent) 8%, transparent);
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
        white-space: nowrap;
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
        margin-bottom: 12px;
    }

    .section-head h2 {
        margin: 0;
        font-size: 13px;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: var(--text-muted);
    }

    .code-block {
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
        overflow: hidden;
        margin-bottom: 12px;
    }

    .code-toolbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        padding: 8px 12px;
        border-bottom: 1px solid var(--border);
        background: var(--bg-secondary);
        background: color-mix(in srgb, var(--bg) 55%, var(--bg-secondary));
    }

    .code-label {
        font-family: var(--font-mono);
        font-size: 11px;
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
    }

    .code-block pre {
        margin: 0;
        padding: 16px;
        font-size: 12px;
        line-height: 1.55;
        color: var(--text);
        font-family: var(--font-mono);
        white-space: pre;
        overflow-x: auto;
        tab-size: 2;
    }

    /* Higher specificity than section-level code styles — avoid !important */
    .docs-section .code-block pre code {
        background: transparent;
        padding: 0;
        color: inherit;
        font-size: inherit;
        white-space: inherit;
        border-radius: 0;
        word-break: normal;
        overflow-wrap: normal;
    }

    .ref-list {
        margin: 0;
        padding-left: 1.25rem;
        color: var(--text-muted);
        font-size: 14px;
        line-height: 1.9;
    }

    .ref-list a {
        color: var(--accent);
        text-decoration: none;
    }

    .ref-list a:hover {
        text-decoration: underline;
    }

    @media (max-width: 640px) {
        .docs-page {
            padding: 16px;
        }

        .sdk-row {
            flex-direction: column;
            align-items: stretch;
        }

        .sdk-actions {
            justify-content: flex-start;
        }

        .callouts li {
            grid-template-columns: 1fr;
            gap: 6px;
        }
    }
</style>
