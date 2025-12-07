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

    const steps = [
        {
            title: "1. Essential Tooling Pre-installed",
            icon: "box",
            description: "We focus on getting you straight to code. Instead of spending hours configuring a new VPS, Rexec provides specialized 'Roles' that come pre-loaded with the perfect toolchain.",
            details: [
                "**Standard**: Docker, Zsh, Git, Curl, Wget",
                "**Node.js**: Node 20/22, Npm, Pnpm, Bun",
                "**Python**: Python 3.11, Pip, Venv, Poetry",
                "**Vibe Coder**: Aider, LLM CLI, specialized AI tools"
            ]
        },
        {
            title: "2. Instant Browser Access",
            icon: "terminal",
            description: "Getting access to the UI is instantaneous. No SSH keys to generate, no IP addresses to copy-paste. We use secure WebSockets to tunnel a full TTY directly to your browser.",
            details: [
                "Zero-latency typing feel",
                "Copy/Paste support via browser clipboard",
                "Mobile-friendly interface with virtual keyboard",
                "Works on iPad, Chromebook, and restricted corporate networks"
            ]
        },
        {
            title: "3. Background Orchestration",
            icon: "settings",
            description: "What happens in the background? We spin up isolated containers that act like full virtual machines. Your data is persisted, but the environment is ephemeral and clean.",
            details: [
                "**Isolation**: Each terminal is a separate Docker/Podman container",
                "**Persistence**: Your `/home/user` directory is saved to a persistent volume",
                "**Networking**: Each container gets its own IP and port space",
                "**Security**: Root access within the container, but isolated from the host"
            ]
        }
    ];
</script>

<div class="guides-page">
    <div class="page-header">
        <div class="header-badge">
            <span class="dot"></span>
            <span>How Rexec Works</span>
        </div>
        <h1>The <span class="accent">Rexec</span> Architecture</h1>
        <p class="subtitle">
            From zero to a full Linux environment in seconds. Here is how we handle the complexity so you can focus on building.
        </p>
    </div>

    <div class="steps-container">
        {#each steps as step, i}
            <div class="step-card">
                <div class="step-icon">
                    <StatusIcon status={step.icon} size={32} />
                    <div class="step-line" class:last={i === steps.length - 1}></div>
                </div>
                <div class="step-content">
                    <h2>{step.title}</h2>
                    <p class="description">{step.description}</p>
                    <div class="details-grid">
                        {#each step.details as detail}
                            <div class="detail-item">
                                <span class="check">✓</span>
                                <!-- Render rudimentary markdown bolding -->
                                {@html detail.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')}
                            </div>
                        {/each}
                    </div>
                </div>
            </div>
        {/each}
    </div>

    <section class="cta-section">
        <h2>Experience the Magic</h2>
        <p>See the orchestration in action. Launch a terminal now.</p>
        <button class="btn btn-primary btn-lg" on:click={handleTryNow}>
            <StatusIcon status="rocket" size={16} />
            <span>Start Terminal</span>
        </button>
    </section>
</div>

<style>
    .guides-page {
        max-width: 1000px;
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
        font-size: 42px;
        font-weight: 700;
        margin-bottom: 16px;
        letter-spacing: -1px;
    }

    h1 .accent {
        color: var(--accent);
        text-shadow: var(--accent-glow);
    }

    .subtitle {
        font-size: 18px;
        color: var(--text-muted);
        max-width: 600px;
        margin: 0 auto;
        line-height: 1.6;
    }

    .steps-container {
        display: flex;
        flex-direction: column;
        gap: 0;
        margin-bottom: 60px;
    }

    .step-card {
        display: flex;
        gap: 30px;
        padding-bottom: 40px;
    }

    .step-icon {
        display: flex;
        flex-direction: column;
        align-items: center;
        flex-shrink: 0;
        width: 60px;
    }

    .step-line {
        flex: 1;
        width: 2px;
        background: var(--border);
        margin-top: 20px;
    }

    .step-line.last {
        background: linear-gradient(to bottom, var(--border), transparent);
    }

    .step-content {
        flex: 1;
        background: var(--bg-card);
        border: 1px solid var(--border);
        padding: 30px;
        border-radius: 8px;
        margin-top: -10px; /* Align with icon */
    }

    .step-content h2 {
        font-size: 24px;
        margin: 0 0 12px 0;
        color: var(--text);
    }

    .description {
        font-size: 15px;
        color: var(--text-secondary);
        line-height: 1.6;
        margin-bottom: 24px;
    }

    .details-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 12px;
    }

    .detail-item {
        display: flex;
        align-items: flex-start;
        gap: 10px;
        font-size: 13px;
        color: var(--text-muted);
        line-height: 1.4;
    }

    .check {
        color: var(--accent);
        font-weight: bold;
    }

    /* Use global styles for strong tags inside svelte html */
    .detail-item :global(strong) {
        color: var(--text);
        font-weight: 600;
    }

    .cta-section {
        text-align: center;
        padding: 60px;
        background: var(--bg-card);
        border: 1px solid var(--border);
        border-radius: 12px;
    }

    .cta-section h2 {
        font-size: 28px;
        margin-bottom: 12px;
    }

    .cta-section p {
        color: var(--text-muted);
        margin-bottom: 24px;
        font-size: 16px;
    }

    @keyframes blink {
        0%, 100% { opacity: 1; }
        50% { opacity: 0; }
    }

    @media (max-width: 768px) {
        .step-card {
            flex-direction: column;
            gap: 16px;
        }
        
        .step-icon {
            flex-direction: row;
            align-items: center;
            gap: 16px;
            width: 100%;
        }

        .step-line {
            display: none;
        }

        .step-content {
            margin-top: 0;
        }
    }
</style>
