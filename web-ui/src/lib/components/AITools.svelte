<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import StatusIcon from "./icons/StatusIcon.svelte";
    
    const dispatch = createEventDispatcher<{
        navigate: { view: string };
        tryNow: void;
    }>();

    function handleTryNow() {
        dispatch("tryNow");
    }

    function goToAgentic() {
        dispatch("navigate", { view: "agentic" });
    }

    const freeTools = [
        {
            name: "tgpt",
            iconType: "free",
            tagline: "Free GPT in your terminal",
            description: "Ask questions, generate code, get explanations - all without an API key. Uses free AI providers.",
            examples: [
                { cmd: 'tgpt "explain kubernetes pods"', desc: "Get explanations" },
                { cmd: 'tgpt -c "write a bash script to backup files"', desc: "Generate code" },
                { cmd: "tgpt -i", desc: "Interactive chat mode" },
                { cmd: 'tgpt -s "summarize this error log"', desc: "Quick summaries" },
            ],
            features: ["No API key required", "Multiple AI providers", "Code generation mode", "Interactive chat"],
        },
        {
            name: "aichat",
            iconType: "chat",
            tagline: "Feature-rich AI chat with Ollama support",
            description: "A powerful AI chat CLI that supports local models via Ollama, multiple providers, and advanced features like roles and sessions.",
            examples: [
                { cmd: "aichat", desc: "Start interactive chat" },
                { cmd: 'aichat "review this code"', desc: "Quick questions" },
                { cmd: "aichat -m ollama:llama3.2", desc: "Use local Ollama model" },
                { cmd: "aichat --role coder", desc: "Use predefined roles" },
            ],
            features: ["Ollama integration", "Session management", "Custom roles", "Streaming responses"],
        },
        {
            name: "mods",
            iconType: "zap",
            tagline: "Pipe anything to AI",
            description: "By Charm. Pipe any command output to AI for analysis, transformation, or explanation. Perfect for CLI workflows.",
            examples: [
                { cmd: 'cat error.log | mods "explain these errors"', desc: "Analyze logs" },
                { cmd: 'git diff | mods "write commit message"', desc: "Generate commit messages" },
                { cmd: 'cat code.py | mods "add type hints"', desc: "Transform code" },
                { cmd: 'kubectl get pods | mods "which pods have issues?"', desc: "Analyze output" },
            ],
            features: ["Pipe-friendly", "Works with any command", "Ollama support", "Beautiful TUI"],
        },
    ];

    const proTools = [
        {
            name: "aider",
            iconType: "code",
            tagline: "AI pair programming in your terminal",
            description: "The most popular AI coding assistant. Edit multiple files, understands your codebase, works with git. Supports Claude, GPT-4, and more.",
            examples: [
                { cmd: "aider", desc: "Start with current directory" },
                { cmd: "aider --model claude-3.5-sonnet", desc: "Use Claude" },
                { cmd: "aider src/*.py", desc: "Add specific files" },
                { cmd: 'aider --message "add dark mode"', desc: "Non-interactive mode" },
            ],
            features: ["Multi-file editing", "Git integration", "Codebase understanding", "Voice mode"],
            requiresKey: "ANTHROPIC_API_KEY or OPENAI_API_KEY",
        },
        {
            name: "opencode",
            iconType: "wrench",
            tagline: "AI coding assistant by SST",
            description: "A terminal-based AI coding assistant with a beautiful TUI. Designed for developers who want a modern, fast coding experience.",
            examples: [
                { cmd: "opencode", desc: "Launch TUI" },
                { cmd: 'opencode "add authentication"', desc: "Quick task" },
            ],
            features: ["Beautiful TUI", "Fast responses", "Context-aware", "Multi-file support"],
            requiresKey: "ANTHROPIC_API_KEY",
        },
        {
            name: "llm",
            iconType: "ai",
            tagline: "CLI for large language models",
            description: "Access various LLMs from the command line. Supports plugins for different providers and local models.",
            examples: [
                { cmd: 'llm "explain quantum computing"', desc: "Quick query" },
                { cmd: "llm chat", desc: "Interactive chat" },
                { cmd: "cat file.txt | llm -s 'summarize'", desc: "Pipe content" },
                { cmd: "llm models", desc: "List available models" },
            ],
            features: ["Multiple providers", "Plugin ecosystem", "Local model support", "Conversation history"],
            requiresKey: "OPENAI_API_KEY (or configure other providers)",
        },
        {
            name: "sgpt",
            iconType: "terminal",
            tagline: "Shell GPT - AI in your shell",
            description: "Command-line productivity tool powered by AI. Generate shell commands, code, and get answers directly in your terminal.",
            examples: [
                { cmd: 'sgpt "find large files over 1GB"', desc: "Generate commands" },
                { cmd: "sgpt --shell 'compress all images'", desc: "Shell mode" },
                { cmd: "sgpt --code 'python http server'", desc: "Code generation" },
                { cmd: "sgpt --repl temp", desc: "REPL mode" },
            ],
            features: ["Shell command generation", "Code generation", "REPL mode", "Chat conversations"],
            requiresKey: "OPENAI_API_KEY",
        },
    ];
</script>

<div class="ai-tools-page">
    <div class="page-header">
        <div class="header-badge">
            <span class="dot"></span>
            <span>AI-Powered Development</span>
        </div>
        <h1>AI Tools for <span class="accent">Vibe Coding</span></h1>
        <p class="subtitle">
            Every Rexec terminal comes with powerful AI tools pre-installed. 
            Start coding with AI assistance immediately - no setup required.
        </p>
        <div class="header-actions">
            <button class="btn btn-primary" on:click={handleTryNow}>
                <StatusIcon status="rocket" size={14} />
                <span>Try Now — Free</span>
            </button>
            <button class="btn btn-secondary" on:click={goToAgentic}>
                <StatusIcon status="ai" size={14} />
                <span>Agentic Use Cases →</span>
            </button>
        </div>
    </div>

    <section class="tools-section">
        <div class="section-header">
            <h2><StatusIcon status="free" size={20} /> Free Tools — No API Key Required</h2>
            <p>These tools work immediately without any configuration. Just type and go.</p>
        </div>
        
        <div class="tools-grid">
            {#each freeTools as tool}
                <div class="tool-card free">
                    <div class="tool-header">
                        <span class="tool-icon"><StatusIcon status={tool.iconType} size={28} /></span>
                        <div class="tool-title">
                            <h3>{tool.name}</h3>
                            <span class="tool-tagline">{tool.tagline}</span>
                        </div>
                        <span class="tool-badge free-badge">FREE</span>
                    </div>
                    <p class="tool-description">{tool.description}</p>
                    
                    <div class="tool-examples">
                        <h4>Examples</h4>
                        {#each tool.examples as example}
                            <div class="example">
                                <code>{example.cmd}</code>
                                <span class="example-desc">{example.desc}</span>
                            </div>
                        {/each}
                    </div>
                    
                    <div class="tool-features">
                        {#each tool.features as feature}
                            <span class="feature-tag"><StatusIcon status="check" size={10} /> {feature}</span>
                        {/each}
                    </div>
                </div>
            {/each}
        </div>
    </section>

    <section class="tools-section">
        <div class="section-header">
            <h2><StatusIcon status="bolt" size={20} /> Pro Tools — Vibe Coder Environment</h2>
            <p>Advanced AI coding tools for serious development. Available in the Vibe Coder environment.</p>
        </div>
        
        <div class="tools-grid">
            {#each proTools as tool}
                <div class="tool-card pro">
                    <div class="tool-header">
                        <span class="tool-icon"><StatusIcon status={tool.iconType} size={28} /></span>
                        <div class="tool-title">
                            <h3>{tool.name}</h3>
                            <span class="tool-tagline">{tool.tagline}</span>
                        </div>
                        <span class="tool-badge pro-badge">VIBE CODER</span>
                    </div>
                    <p class="tool-description">{tool.description}</p>
                    
                    <div class="tool-examples">
                        <h4>Examples</h4>
                        {#each tool.examples as example}
                            <div class="example">
                                <code>{example.cmd}</code>
                                <span class="example-desc">{example.desc}</span>
                            </div>
                        {/each}
                    </div>
                    
                    <div class="tool-features">
                        {#each tool.features as feature}
                            <span class="feature-tag"><StatusIcon status="check" size={10} /> {feature}</span>
                        {/each}
                    </div>
                    
                    {#if tool.requiresKey}
                        <div class="api-key-note">
                            <StatusIcon status="key" size={14} />
                            <span>Requires: <code>{tool.requiresKey}</code></span>
                        </div>
                    {/if}
                </div>
            {/each}
        </div>
    </section>

    <section class="quick-start-section">
        <h2><StatusIcon status="zap" size={20} /> Quick Start</h2>
        <div class="quick-start-grid">
            <div class="quick-start-card">
                <h3>1. Launch Terminal</h3>
                <p>Create a new terminal with any environment. All environments include free AI tools.</p>
            </div>
            <div class="quick-start-card">
                <h3>2. Ask AI Anything</h3>
                <p>Type <code>tgpt "your question"</code> to get instant AI assistance - no setup needed.</p>
            </div>
            <div class="quick-start-card">
                <h3>3. Vibe Code</h3>
                <p>Use <code>ai-help</code> to see all available tools and start your AI-powered workflow.</p>
            </div>
        </div>
    </section>

    <section class="cta-section">
        <h2>Ready to Code with AI?</h2>
        <p>Get a cloud terminal with all these tools in seconds.</p>
        <button class="btn btn-primary btn-lg" on:click={handleTryNow}>
            <StatusIcon status="ai" size={16} />
            <span>Launch AI Terminal — Free</span>
        </button>
    </section>
</div>

<style>
    .ai-tools-page {
        max-width: 1200px;
        margin: 0 auto;
        padding: 40px 20px;
    }

    .page-header {
        text-align: center;
        margin-bottom: 60px;
    }

    .header-badge {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 4px 12px;
        background: var(--bg-card);
        border: 1px solid var(--border);
        font-size: 11px;
        color: var(--text-secondary);
        margin-bottom: 20px;
        text-transform: uppercase;
        letter-spacing: 1px;
    }

    .header-badge .dot {
        width: 6px;
        height: 6px;
        background: var(--accent);
        animation: blink 1s step-end infinite;
    }

    h1 {
        font-size: 36px;
        font-weight: 700;
        margin-bottom: 16px;
        text-transform: uppercase;
        letter-spacing: 2px;
    }

    h1 .accent {
        color: var(--accent);
        text-shadow: var(--accent-glow);
    }

    .subtitle {
        font-size: 16px;
        color: var(--text-muted);
        max-width: 600px;
        margin: 0 auto 30px;
        line-height: 1.6;
    }

    .header-actions {
        display: flex;
        gap: 16px;
        justify-content: center;
    }

    .tools-section {
        margin-bottom: 60px;
    }

    .section-header {
        margin-bottom: 30px;
    }

    .section-header h2 {
        font-size: 24px;
        margin-bottom: 8px;
        color: var(--text);
    }

    .section-header p {
        color: var(--text-muted);
        font-size: 14px;
    }

    .tools-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
        gap: 24px;
    }

    .tool-card {
        background: var(--bg-card);
        border: 1px solid var(--border);
        padding: 24px;
        transition: border-color 0.2s, transform 0.2s;
    }

    .tool-card:hover {
        border-color: var(--accent);
        transform: translateY(-2px);
    }

    .tool-card.free {
        border-left: 3px solid #00ff41;
    }

    .tool-card.pro {
        border-left: 3px solid #ff6b00;
    }

    .tool-header {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        margin-bottom: 16px;
    }

    .tool-icon {
        font-size: 32px;
    }

    .tool-title {
        flex: 1;
    }

    .tool-title h3 {
        font-size: 18px;
        margin: 0 0 4px 0;
        color: var(--text);
        font-family: var(--font-mono);
    }

    .tool-tagline {
        font-size: 12px;
        color: var(--text-muted);
    }

    .tool-badge {
        padding: 4px 8px;
        font-size: 10px;
        font-weight: 600;
        border-radius: 3px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .free-badge {
        background: rgba(0, 255, 65, 0.15);
        color: #00ff41;
        border: 1px solid rgba(0, 255, 65, 0.3);
    }

    .pro-badge {
        background: rgba(255, 107, 0, 0.15);
        color: #ff6b00;
        border: 1px solid rgba(255, 107, 0, 0.3);
    }

    .tool-description {
        font-size: 13px;
        color: var(--text-secondary);
        line-height: 1.6;
        margin-bottom: 20px;
    }

    .tool-examples {
        margin-bottom: 16px;
    }

    .tool-examples h4 {
        font-size: 11px;
        text-transform: uppercase;
        letter-spacing: 1px;
        color: var(--text-muted);
        margin-bottom: 10px;
    }

    .example {
        display: flex;
        flex-direction: column;
        gap: 4px;
        margin-bottom: 10px;
        padding: 8px;
        background: rgba(0, 0, 0, 0.3);
        border-left: 2px solid var(--border);
    }

    .example code {
        font-family: var(--font-mono);
        font-size: 12px;
        color: var(--accent);
    }

    .example-desc {
        font-size: 11px;
        color: var(--text-muted);
    }

    .tool-features {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .feature-tag {
        font-size: 11px;
        padding: 4px 8px;
        background: rgba(255, 255, 255, 0.05);
        border: 1px solid var(--border);
        color: var(--text-secondary);
    }

    .api-key-note {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-top: 16px;
        padding: 10px;
        background: rgba(255, 193, 7, 0.1);
        border: 1px solid rgba(255, 193, 7, 0.3);
        font-size: 11px;
        color: #ffc107;
    }

    .api-key-note code {
        font-family: var(--font-mono);
        font-size: 10px;
    }

    .quick-start-section {
        margin-bottom: 60px;
    }

    .quick-start-section h2 {
        font-size: 24px;
        margin-bottom: 24px;
        text-align: center;
    }

    .quick-start-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 24px;
    }

    .quick-start-card {
        background: var(--bg-card);
        border: 1px solid var(--border);
        padding: 24px;
        text-align: center;
    }

    .quick-start-card h3 {
        font-size: 16px;
        margin-bottom: 12px;
        color: var(--accent);
    }

    .quick-start-card p {
        font-size: 13px;
        color: var(--text-muted);
        line-height: 1.5;
    }

    .quick-start-card code {
        font-family: var(--font-mono);
        background: rgba(0, 255, 65, 0.1);
        padding: 2px 6px;
        font-size: 12px;
        color: var(--accent);
    }

    .cta-section {
        text-align: center;
        padding: 60px;
        background: var(--bg-card);
        border: 1px solid var(--border);
    }

    .cta-section h2 {
        font-size: 28px;
        margin-bottom: 12px;
    }

    .cta-section p {
        color: var(--text-muted);
        margin-bottom: 24px;
    }

    @keyframes blink {
        0%, 100% { opacity: 1; }
        50% { opacity: 0; }
    }

    @media (max-width: 768px) {
        h1 {
            font-size: 24px;
        }

        .tools-grid {
            grid-template-columns: 1fr;
        }

        .quick-start-grid {
            grid-template-columns: 1fr;
        }

        .header-actions {
            flex-direction: column;
        }
    }
</style>
