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

    function goToAITools() {
        dispatch("navigate", { view: "ai-tools" });
    }

    const useCases = [
        {
            title: "Autonomous Code Generation",
            iconType: "ai",
            description: "Let AI agents write, test, and refactor code autonomously. Describe what you want, and watch it build.",
            tools: ["aider", "opencode", "tgpt"],
            example: {
                scenario: "Build a REST API",
                commands: [
                    { cmd: 'aider --message "create a Python FastAPI app with user authentication"', desc: "AI writes the code" },
                    { cmd: 'aider --message "add JWT token support"', desc: "Iterate with natural language" },
                    { cmd: 'aider --message "write tests for all endpoints"', desc: "Generate tests automatically" },
                ],
            },
        },
        {
            title: "Automated Bug Fixing",
            iconType: "wrench",
            description: "Paste error logs, stack traces, or describe the bug. AI agents analyze and fix issues automatically.",
            tools: ["aider", "mods", "tgpt"],
            example: {
                scenario: "Fix a runtime error",
                commands: [
                    { cmd: 'cat error.log | mods "analyze this error and suggest fixes"', desc: "Analyze error" },
                    { cmd: 'aider --message "fix the null pointer exception in user.py line 42"', desc: "Auto-fix" },
                    { cmd: 'git diff | mods "explain what was changed"', desc: "Understand the fix" },
                ],
            },
        },
        {
            title: "Code Review Agent",
            iconType: "code",
            description: "Automated code reviews that catch bugs, suggest improvements, and ensure best practices.",
            tools: ["mods", "aichat", "llm"],
            example: {
                scenario: "Review a pull request",
                commands: [
                    { cmd: 'git diff main..feature | mods "review this code for bugs and security issues"', desc: "Security review" },
                    { cmd: 'cat src/*.py | mods "suggest performance improvements"', desc: "Performance review" },
                    { cmd: 'git log --oneline -10 | mods "summarize recent changes"', desc: "Change summary" },
                ],
            },
        },
        {
            title: "Documentation Generator",
            iconType: "info",
            description: "Automatically generate documentation, README files, API docs, and inline comments.",
            tools: ["aider", "mods", "tgpt"],
            example: {
                scenario: "Document a codebase",
                commands: [
                    { cmd: 'aider --message "add docstrings to all functions in src/"', desc: "Add docstrings" },
                    { cmd: 'cat src/*.py | mods "generate a README.md for this project"', desc: "Generate README" },
                    { cmd: 'tgpt -c "write API documentation for these endpoints"', desc: "API docs" },
                ],
            },
        },
        {
            title: "DevOps Automation",
            iconType: "settings",
            description: "Generate Dockerfiles, Kubernetes manifests, CI/CD pipelines, and infrastructure as code.",
            tools: ["tgpt", "mods", "aichat"],
            example: {
                scenario: "Containerize an application",
                commands: [
                    { cmd: 'tgpt "write a Dockerfile for a Python Flask app"', desc: "Generate Dockerfile" },
                    { cmd: 'tgpt "create kubernetes deployment and service yaml"', desc: "K8s manifests" },
                    { cmd: 'cat .github/workflows/*.yml | mods "optimize this CI pipeline"', desc: "Optimize CI" },
                ],
            },
        },
        {
            title: "Data Pipeline Assistant",
            iconType: "chart",
            description: "Build ETL pipelines, write SQL queries, analyze data, and generate visualizations.",
            tools: ["aichat", "mods", "tgpt"],
            example: {
                scenario: "Build a data pipeline",
                commands: [
                    { cmd: 'tgpt "write a Python script to extract data from this API and load to PostgreSQL"', desc: "ETL script" },
                    { cmd: 'cat schema.sql | mods "write optimized queries for reporting"', desc: "Query optimization" },
                    { cmd: 'aichat "analyze this CSV and suggest data cleaning steps"', desc: "Data analysis" },
                ],
            },
        },
    ];

    const workflows = [
        {
            title: "The Vibe Coder Loop",
            description: "Natural language → Code → Test → Deploy",
            steps: [
                { step: "1", text: "Describe what you want in plain English" },
                { step: "2", text: "AI generates the code and makes changes" },
                { step: "3", text: "Review and iterate with natural language" },
                { step: "4", text: "AI writes tests and documentation" },
                { step: "5", text: "Deploy with AI-generated scripts" },
            ],
        },
        {
            title: "Zero-Config AI Pipeline",
            description: "Pipe any data through AI for instant processing",
            steps: [
                { step: "1", text: "Pipe command output: kubectl get pods | mods 'find issues'" },
                { step: "2", text: "Pipe files: cat code.py | mods 'review'" },
                { step: "3", text: "Chain commands: git diff | mods 'commit message'" },
                { step: "4", text: "Save results: tgpt 'query' > result.txt" },
            ],
        },
    ];
</script>

<div class="agentic-page">
    <div class="page-header">
        <div class="header-badge">
            <span class="dot"></span>
            <span>Autonomous AI Agents</span>
        </div>
        <h1>Agentic <span class="accent">Development</span></h1>
        <p class="subtitle">
            Let AI agents handle the heavy lifting. Describe what you want, 
            and autonomous agents write code, fix bugs, and build features.
        </p>
        <div class="header-actions">
            <button class="btn btn-primary" on:click={handleTryNow}>
                <StatusIcon status="ai" size={14} />
                <span>Start Building with Agents</span>
            </button>
            <button class="btn btn-secondary" on:click={goToAITools}>
                <StatusIcon status="wrench" size={14} />
                <span>View All AI Tools →</span>
            </button>
        </div>
    </div>

    <section class="workflows-section">
        <h2><StatusIcon status="workflow" size={20} /> Agentic Workflows</h2>
        <div class="workflows-grid">
            {#each workflows as workflow}
                <div class="workflow-card">
                    <h3>{workflow.title}</h3>
                    <p class="workflow-desc">{workflow.description}</p>
                    <div class="workflow-steps">
                        {#each workflow.steps as step}
                            <div class="workflow-step">
                                <span class="step-number">{step.step}</span>
                                <span class="step-text">{step.text}</span>
                            </div>
                        {/each}
                    </div>
                </div>
            {/each}
        </div>
    </section>

    <section class="use-cases-section">
        <h2><StatusIcon status="sparkles" size={20} /> Agentic Use Cases</h2>
        <p class="section-desc">Real-world examples of AI agents automating development tasks</p>
        
        <div class="use-cases-grid">
            {#each useCases as useCase}
                <div class="use-case-card">
                    <div class="use-case-header">
                        <span class="use-case-icon"><StatusIcon status={useCase.iconType} size={32} /></span>
                        <div>
                            <h3>{useCase.title}</h3>
                            <div class="use-case-tools">
                                {#each useCase.tools as tool}
                                    <span class="tool-chip">{tool}</span>
                                {/each}
                            </div>
                        </div>
                    </div>
                    <p class="use-case-desc">{useCase.description}</p>
                    
                    <div class="use-case-example">
                        <div class="example-header">
                            <StatusIcon status="terminal" size={14} />
                            <span>Example: {useCase.example.scenario}</span>
                        </div>
                        <div class="example-commands">
                            {#each useCase.example.commands as cmd}
                                <div class="example-cmd">
                                    <code>{cmd.cmd}</code>
                                    <span class="cmd-desc"># {cmd.desc}</span>
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    </section>

    <section class="getting-started-section">
        <h2><StatusIcon status="rocket" size={20} /> Get Started in 60 Seconds</h2>
        <div class="terminal-demo">
            <div class="terminal-header">
                <span class="dot red"></span>
                <span class="dot yellow"></span>
                <span class="dot green"></span>
                <span class="terminal-title">rexec — agentic-demo</span>
            </div>
            <div class="terminal-body">
                <div class="term-line">
                    <span class="prompt">$</span>
                    <span class="cmd">tgpt "create a todo app in Python with SQLite"</span>
                </div>
                <div class="term-output">
                    <span class="ai-badge"><StatusIcon status="ai" size={12} /> AI</span> Here's a complete todo app with SQLite...
                </div>
                <div class="term-line">
                    <span class="prompt">$</span>
                    <span class="cmd">aider todo.py</span>
                </div>
                <div class="term-output">
                    <span class="ai-badge"><StatusIcon status="ai" size={12} /> Aider</span> Added todo.py to the chat. What would you like to change?
                </div>
                <div class="term-line">
                    <span class="prompt">$</span>
                    <span class="cmd typing">"add due dates and priority levels"</span>
                    <span class="cursor">_</span>
                </div>
            </div>
        </div>
    </section>

    <section class="cta-section">
        <h2>Build Faster with AI Agents</h2>
        <p>No setup. No configuration. Just describe and build.</p>
        <div class="cta-buttons">
            <button class="btn btn-primary btn-lg" on:click={handleTryNow}>
                <StatusIcon status="ai" size={16} />
                <span>Launch Agentic Terminal</span>
            </button>
            <button class="btn btn-secondary btn-lg" on:click={goToAITools}>
                <StatusIcon status="info" size={16} />
                <span>Explore All Tools</span>
            </button>
        </div>
    </section>
</div>

<style>
    .agentic-page {
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
        background: #ff6b00;
        animation: pulse 2s ease-in-out infinite;
    }

    @keyframes pulse {
        0%, 100% { opacity: 1; transform: scale(1); }
        50% { opacity: 0.5; transform: scale(1.2); }
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

    section {
        margin-bottom: 60px;
    }

    section h2 {
        font-size: 24px;
        margin-bottom: 16px;
    }

    .section-desc {
        color: var(--text-muted);
        margin-bottom: 30px;
    }

    /* Workflows */
    .workflows-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 24px;
    }

    .workflow-card {
        background: var(--bg-card);
        border: 1px solid var(--border);
        padding: 24px;
    }

    .workflow-card h3 {
        font-size: 18px;
        margin-bottom: 8px;
        color: var(--accent);
    }

    .workflow-desc {
        font-size: 13px;
        color: var(--text-muted);
        margin-bottom: 20px;
    }

    .workflow-steps {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .workflow-step {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .step-number {
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--accent);
        color: var(--bg);
        font-size: 12px;
        font-weight: bold;
        border-radius: 50%;
    }

    .step-text {
        font-size: 13px;
        color: var(--text-secondary);
        font-family: var(--font-mono);
    }

    /* Use Cases */
    .use-cases-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
        gap: 24px;
    }

    .use-case-card {
        background: var(--bg-card);
        border: 1px solid var(--border);
        padding: 24px;
        transition: border-color 0.2s;
    }

    .use-case-card:hover {
        border-color: var(--accent);
    }

    .use-case-header {
        display: flex;
        gap: 16px;
        margin-bottom: 16px;
    }

    .use-case-icon {
        font-size: 36px;
    }

    .use-case-header h3 {
        font-size: 16px;
        margin-bottom: 8px;
    }

    .use-case-tools {
        display: flex;
        gap: 6px;
    }

    .tool-chip {
        font-size: 10px;
        padding: 2px 6px;
        background: rgba(0, 255, 65, 0.1);
        border: 1px solid rgba(0, 255, 65, 0.3);
        color: var(--accent);
        font-family: var(--font-mono);
    }

    .use-case-desc {
        font-size: 13px;
        color: var(--text-secondary);
        line-height: 1.6;
        margin-bottom: 20px;
    }

    .use-case-example {
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid var(--border);
        padding: 16px;
    }

    .example-header {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 12px;
        color: var(--text-muted);
        margin-bottom: 12px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .example-commands {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .example-cmd {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .example-cmd code {
        font-family: var(--font-mono);
        font-size: 11px;
        color: var(--accent);
        word-break: break-all;
    }

    .cmd-desc {
        font-size: 10px;
        color: var(--text-muted);
    }

    /* Terminal Demo */
    .terminal-demo {
        max-width: 700px;
        margin: 30px auto;
        background: #000;
        border: 1px solid var(--border);
    }

    .terminal-header {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 10px 12px;
        background: #111;
        border-bottom: 1px solid var(--border);
    }

    .terminal-header .dot {
        width: 10px;
        height: 10px;
        border-radius: 50%;
    }

    .terminal-header .dot.red { background: #ff5f56; }
    .terminal-header .dot.yellow { background: #ffbd2e; }
    .terminal-header .dot.green { background: #27c93f; }

    .terminal-title {
        flex: 1;
        text-align: center;
        font-size: 11px;
        color: var(--text-muted);
    }

    .terminal-body {
        padding: 20px;
        font-family: var(--font-mono);
        font-size: 13px;
    }

    .term-line {
        margin-bottom: 8px;
    }

    .prompt {
        color: var(--accent);
        margin-right: 8px;
    }

    .cmd {
        color: var(--text);
    }

    .term-output {
        color: var(--text-muted);
        margin-bottom: 12px;
        padding-left: 16px;
    }

    .ai-badge {
        background: rgba(0, 255, 65, 0.1);
        padding: 2px 6px;
        font-size: 10px;
        margin-right: 8px;
    }

    .cursor {
        background: var(--accent);
        color: var(--bg);
        animation: blink 1s step-end infinite;
    }

    @keyframes blink {
        0%, 100% { opacity: 1; }
        50% { opacity: 0; }
    }

    .typing {
        color: var(--accent);
    }

    /* CTA */
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

    .cta-buttons {
        display: flex;
        gap: 16px;
        justify-content: center;
    }

    @media (max-width: 768px) {
        h1 {
            font-size: 24px;
        }

        .workflows-grid,
        .use-cases-grid {
            grid-template-columns: 1fr;
        }

        .header-actions,
        .cta-buttons {
            flex-direction: column;
        }

        .use-cases-grid {
            grid-template-columns: 1fr;
        }
    }
</style>
