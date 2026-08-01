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
    }

    const sdks = [
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
    ];

    // Examples match live API: image aliases (ubuntu), list/create/get/delete.
    const codeExamples: Record<string, string> = {
        javascript: `import { RexecClient } from 'pipeops-rexec';

const client = new RexecClient({
  baseURL: process.env.REXEC_URL,
  token: process.env.REXEC_TOKEN,
});

// List sandboxes
const list = await client.containers.list();

// Create (use image aliases: ubuntu, debian, alpine, …)
const container = await client.containers.create({
  image: 'ubuntu',
  name: 'demo',
});
console.log(container.id, container.status);

await client.containers.get(container.id);
await client.containers.delete(container.id);

// Terminal over WebSocket
// const term = await client.terminal.connect(container.id);
// term.onData((d) => console.log(d));
// term.write('echo hello\\n');`,
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
    };
</script>

<div class="sdk-docs">
    <header class="docs-header">
        <button class="back-btn" on:click={handleBack} type="button">
            <StatusIcon status="arrow-left" size={16} />
            Back
        </button>
        <div>
            <h1>SDKs</h1>
            <p class="lead">
                Official clients for the Rexec API — verified end-to-end
                (<code>list → create → get → delete</code>). Latest:
                <strong>v1.0.1</strong>.
            </p>
        </div>
    </header>

    <section class="callout">
        <h2>Quick notes</h2>
        <ul>
            <li>
                Use <strong>image aliases</strong> such as <code>ubuntu</code>,
                <code>debian</code>, or <code>alpine</code> (not
                <code>ubuntu:24.04</code> on hosted Rexec).
            </li>
            <li>
                Set <code>REXEC_URL</code> and <code>REXEC_TOKEN</code> (API token
                from Settings, or guest login).
            </li>
            <li>
                Full reference:
                <a href="https://github.com/PipeOpsHQ/rexec/blob/main/docs/SDK.md" target="_blank" rel="noreferrer">docs/SDK.md</a>
            </li>
        </ul>
    </section>

    <section class="sdk-grid">
        {#each sdks as sdk}
            <article class="sdk-card" class:active={activeTab === sdk.id}>
                <button type="button" class="sdk-tab" on:click={() => (activeTab = sdk.id)}>
                    <h3>{sdk.name}</h3>
                    <code class="install">{sdk.install}</code>
                </button>
                <div class="card-actions">
                    <button
                        type="button"
                        class="copy"
                        on:click={() => copyToClipboard(sdk.install, sdk.id)}
                    >
                        {copiedCommand === sdk.id ? "Copied" : "Copy install"}
                    </button>
                    <a href={sdk.registry} target="_blank" rel="noreferrer">Registry</a>
                    <a href={sdk.github} target="_blank" rel="noreferrer">Source</a>
                </div>
            </article>
        {/each}
    </section>

    <section class="example">
        <div class="example-header">
            <h2>{sdks.find((s) => s.id === activeTab)?.name} example</h2>
            <button
                type="button"
                class="copy"
                on:click={() => copyToClipboard(codeExamples[activeTab], "code")}
            >
                {copiedCommand === "code" ? "Copied" : "Copy code"}
            </button>
        </div>
        <pre><code>{codeExamples[activeTab]}</code></pre>
    </section>
</div>

<style>
    .sdk-docs {
        max-width: 960px;
        margin: 0 auto;
        padding: 1.5rem;
        color: var(--text, #e2e8f0);
    }
    .docs-header {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        margin-bottom: 1.5rem;
    }
    .back-btn {
        align-self: flex-start;
        display: inline-flex;
        align-items: center;
        gap: 0.35rem;
        background: transparent;
        border: 1px solid rgba(148, 163, 184, 0.3);
        color: inherit;
        border-radius: 0.5rem;
        padding: 0.35rem 0.75rem;
        cursor: pointer;
    }
    h1 {
        font-size: 1.75rem;
        margin: 0;
    }
    .lead {
        color: rgba(226, 232, 240, 0.75);
        margin: 0.35rem 0 0;
    }
    .callout {
        background: rgba(15, 23, 42, 0.6);
        border: 1px solid rgba(148, 163, 184, 0.2);
        border-radius: 0.75rem;
        padding: 1rem 1.25rem;
        margin-bottom: 1.5rem;
    }
    .callout ul {
        margin: 0.5rem 0 0;
        padding-left: 1.2rem;
    }
    .sdk-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
        gap: 0.75rem;
        margin-bottom: 1.5rem;
    }
    .sdk-card {
        border: 1px solid rgba(148, 163, 184, 0.2);
        border-radius: 0.75rem;
        padding: 0.75rem;
        background: rgba(15, 23, 42, 0.45);
    }
    .sdk-card.active {
        border-color: #22d3ee;
        box-shadow: 0 0 0 1px rgba(34, 211, 238, 0.35);
    }
    .sdk-tab {
        width: 100%;
        text-align: left;
        background: transparent;
        border: 0;
        color: inherit;
        cursor: pointer;
        padding: 0;
    }
    .sdk-tab h3 {
        margin: 0 0 0.35rem;
        font-size: 1rem;
    }
    .install {
        display: block;
        font-size: 0.75rem;
        color: #67e8f9;
        word-break: break-all;
    }
    .card-actions {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        margin-top: 0.65rem;
        font-size: 0.8rem;
    }
    .card-actions a {
        color: #93c5fd;
    }
    .copy {
        background: rgba(34, 211, 238, 0.12);
        border: 1px solid rgba(34, 211, 238, 0.35);
        color: #a5f3fc;
        border-radius: 0.4rem;
        padding: 0.25rem 0.55rem;
        cursor: pointer;
        font-size: 0.8rem;
    }
    .example {
        border: 1px solid rgba(148, 163, 184, 0.2);
        border-radius: 0.75rem;
        overflow: hidden;
    }
    .example-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem 1rem;
        background: rgba(15, 23, 42, 0.8);
    }
    .example-header h2 {
        margin: 0;
        font-size: 1rem;
    }
    pre {
        margin: 0;
        padding: 1rem;
        overflow: auto;
        background: #0b1220;
        font-size: 0.8rem;
        line-height: 1.45;
    }
    code {
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    }
</style>
