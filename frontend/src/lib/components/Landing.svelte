<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { auth } from "$stores/auth";
    import { toast } from "$stores/toast";
    import StatusIcon from "./icons/StatusIcon.svelte";

    const dispatch = createEventDispatcher<{
        guest: void;
        navigate: { view: string };
    }>();

    let isOAuthLoading = false;

    function handleGuestClick() {
        dispatch("guest");
    }

    function navigateTo(view: string) {
        dispatch("navigate", { view });
    }

    async function handleOAuthLogin() {
        if (isOAuthLoading) return;

        isOAuthLoading = true;
        try {
            const url = await auth.getOAuthUrl();
            if (url) {
                window.location.href = url;
            } else {
                toast.error(
                    "Unable to connect to PipeOps. Please try again later.",
                );
                isOAuthLoading = false;
            }
        } catch (e) {
            toast.error("Failed to connect to PipeOps. Please try again.");
            isOAuthLoading = false;
        }
    }
</script>

<div class="landing">
    <div class="landing-content">
        <div class="landing-badge">
            <span class="dot"></span>
            <span>Terminal as a Service</span>
        </div>

        <h1>
            Instant <span class="accent">Linux</span> Terminals
            <br />
            In Your Browser
        </h1>

        <p class="description">
            Create your first terminal to access a cloud environment, GPU workspace, 
            or connect to remote resources. No setup required.
        </p>

        <div class="landing-actions">
            <button class="btn btn-primary btn-lg" onclick={handleGuestClick}>
                Try Now — No Sign Up
            </button>
            <button
                class="btn btn-secondary btn-lg"
                onclick={handleOAuthLogin}
                disabled={isOAuthLoading}
            >
                {#if isOAuthLoading}
                    <span class="btn-spinner"></span>
                    Connecting...
                {:else}
                    Sign in with PipeOps
                {/if}
            </button>
        </div>

        <div class="landing-links">
            <button class="link-btn" onclick={() => navigateTo('use-cases')}>
                <StatusIcon status="bolt" size={14} /> Use Cases
            </button>
            <span class="divider"></span>
            <button class="link-btn" onclick={() => navigateTo('guides')}>
                <StatusIcon status="book" size={14} /> Product Guide
            </button>
        </div>

        <div class="terminal-preview">
            <div class="terminal-preview-header">
                <span class="terminal-dot dot-red"></span>
                <span class="terminal-dot dot-yellow"></span>
                <span class="terminal-dot dot-green"></span>
                <span class="terminal-title">ubuntu-24 — rexec</span>
            </div>
            <div class="terminal-preview-body">
                <div class="terminal-line">
                    <span class="prompt">root@rexec:~#</span>
                    <span class="command">whoami</span>
                </div>
                <div class="terminal-output">root</div>
                <div class="terminal-line">
                    <span class="prompt">root@rexec:~#</span>
                    <span class="command">uname -a</span>
                </div>
                <div class="terminal-output">
                    Linux rexec 6.5.0-44-generic #44-Ubuntu SMP x86_64 GNU/Linux
                </div>
                <div class="terminal-line">
                    <span class="prompt">root@rexec:~#</span>
                    <span class="cursor">_</span>
                </div>
            </div>
        </div>

        <div class="features">
            <div class="feature">
                <span class="feature-icon"><StatusIcon status="bolt" size={24} /></span>
                <h3>Instant</h3>
                <p>Rexec terminals launch in seconds with pre-configured shells</p>
            </div>
            <div class="feature">
                <span class="feature-icon"><StatusIcon status="connected" size={24} /></span>
                <h3>Isolated</h3>
                <p>Each terminal is fully isolated with its own filesystem</p>
            </div>
            <div class="feature">
                <span class="feature-icon"><StatusIcon status="terminal" size={24} /></span>
                <h3>Accessible</h3>
                <p>Access from any browser, anywhere. SSH support included</p>
            </div>
        </div>

        <div class="faq-section">
            <h2 class="faq-title">Frequently Asked Questions</h2>

            <div class="faq-grid">
                <details class="faq-item">
                    <summary class="faq-question">
                        <span class="faq-icon">?</span>
                        Is this related to the old rexec protocol?
                    </summary>
                    <div class="faq-answer">
                        <p><strong>No, not at all!</strong> We understand the concern — the legacy <code>rexec</code> (Remote Execution) protocol from the 1980s is indeed deprecated and insecure. Our Rexec is a completely different, modern product.</p>
                        <p>We chose the name "Rexec" as shorthand for <strong>"Remote Execution"</strong> — which perfectly describes what we do: execute commands on remote cloud terminals. Think of it as reclaiming a cool name for a secure, modern use case.</p>
                        <p>Our platform uses <strong>TLS encryption</strong>, <strong>JWT authentication</strong>, <strong>container isolation</strong>, and follows modern security best practices. It's as secure as any enterprise cloud platform.</p>
                    </div>
                </details>

                <details class="faq-item">
                    <summary class="faq-question">
                        <span class="faq-icon">🔒</span>
                        How is Rexec secure?
                    </summary>
                    <div class="faq-answer">
                        <p>Security is built into every layer:</p>
                        <ul>
                            <li><strong>TLS/HTTPS</strong> — All connections are encrypted end-to-end</li>
                            <li><strong>Container Isolation</strong> — Each terminal runs in its own isolated container</li>
                            <li><strong>OAuth + MFA</strong> — Secure authentication with multi-factor support</li>
                            <li><strong>No Shared State</strong> — Your terminal is yours alone</li>
                            <li><strong>Automatic Cleanup</strong> — Sessions are destroyed when you're done</li>
                        </ul>
                    </div>
                </details>

                <details class="faq-item">
                    <summary class="faq-question">
                        <span class="faq-icon">⚡</span>
                        What can I use Rexec for?
                    </summary>
                    <div class="faq-answer">
                        <p>Rexec is perfect for:</p>
                        <ul>
                            <li><strong>Learning Linux</strong> — Safe environment to experiment</li>
                            <li><strong>Development</strong> — Instant dev environments with pre-installed tools</li>
                            <li><strong>DevOps</strong> — Test scripts, debug issues, run CI/CD tasks</li>
                            <li><strong>Teaching</strong> — Provide students with instant terminals</li>
                            <li><strong>AI/ML Workloads</strong> — Access GPU-enabled environments</li>
                            <li><strong>Remote Access</strong> — Connect to your servers from anywhere</li>
                        </ul>
                    </div>
                </details>

                <details class="faq-item">
                    <summary class="faq-question">
                        <span class="faq-icon">💾</span>
                        Is my data persistent?
                    </summary>
                    <div class="faq-answer">
                        <p>By default, terminals are ephemeral — they're destroyed when your session ends. This is great for security and quick experimentation.</p>
                        <p>For persistent storage, you can attach volumes to your containers or connect to your own infrastructure using Rexec Agents.</p>
                    </div>
                </details>
            </div>
        </div>


    </div>
</div>

<style>
    .landing {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: calc(100vh - 100px);
        min-height: calc(100dvh - 100px);
        text-align: center;
        border: 1px solid var(--border);
        background: var(--bg-elevated);
        position: relative;
        padding: 40px;
    }

    .landing::before {
        content: "";
        position: absolute;
        top: -1px;
        left: -1px;
        width: 10px;
        height: 10px;
        border-top: 2px solid var(--accent);
        border-left: 2px solid var(--accent);
    }

    .landing::after {
        content: "";
        position: absolute;
        bottom: -1px;
        right: -1px;
        width: 10px;
        height: 10px;
        border-bottom: 2px solid var(--accent);
        border-right: 2px solid var(--accent);
    }

    .landing-content {
        max-width: 800px;
        width: 100%;
    }

    .landing-badge {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 4px 12px;
        background: var(--bg-card);
        border: 1px solid var(--border);
        font-size: 11px;
        color: var(--text-secondary);
        margin-bottom: 24px;
        text-transform: uppercase;
        letter-spacing: 1px;
    }

    .landing-badge .dot {
        width: 6px;
        height: 6px;
        background: var(--accent);
        animation: blink 1s step-end infinite;
    }

    h1 {
        font-size: 36px;
        font-weight: 700;
        margin-bottom: 20px;
        text-transform: uppercase;
        letter-spacing: 2px;
        line-height: 1.3;
    }

    h1 .accent {
        color: var(--accent);
        text-shadow: var(--accent-glow);
    }

    .description {
        font-size: 14px;
        color: var(--text-muted);
        max-width: 500px;
        margin: 0 auto 40px;
        line-height: 1.6;
    }

    .landing-actions {
        display: flex;
        gap: 16px;
        justify-content: center;
        margin-bottom: 40px;
    }

    .landing-links {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 16px;
        margin-bottom: 40px;
    }

    .link-btn {
        background: none;
        border: none;
        color: var(--text-secondary);
        font-size: 13px;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 8px 12px;
        border-radius: 6px;
        transition: all 0.2s;
        border: 1px solid transparent;
    }

    .link-btn:hover {
        color: var(--text);
        background: var(--bg-card);
        border-color: var(--border);
    }

    .divider {
        width: 4px;
        height: 4px;
        background: var(--border);
        border-radius: 50%;
    }

    .btn-spinner {
        display: inline-block;
        width: 14px;
        height: 14px;
        border: 2px solid transparent;
        border-top-color: currentColor;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
        margin-right: 8px;
        vertical-align: middle;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    .btn:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }

    .terminal-preview {
        width: 100%;
        max-width: 600px;
        margin: 0 auto 40px;
        background: var(--terminal-bg, #0a0a0a);
        border: 1px solid var(--border);
        text-align: left;
    }

    .terminal-preview-header {
        display: flex;
        align-items: center;
        padding: 8px 12px;
        background: var(--terminal-header-bg, #111);
        border-bottom: 1px solid var(--border);
        gap: 6px;
    }

    .terminal-dot {
        width: 10px;
        height: 10px;
        border-radius: 50%;
        background: var(--border);
    }

    .terminal-dot.dot-red {
        background: #ff5f56;
    }

    .terminal-dot.dot-yellow {
        background: #ffbd2e;
    }

    .terminal-dot.dot-green {
        background: #27c93f;
    }

    .terminal-title {
        flex: 1;
        text-align: center;
        font-size: 11px;
        color: var(--text-muted);
    }

    .terminal-preview-body {
        padding: 16px;
        font-family: var(--font-mono);
        font-size: 13px;
    }

    .terminal-line {
        margin-bottom: 4px;
    }

    .prompt {
        color: var(--accent);
        margin-right: 8px;
    }

    .command {
        color: var(--text);
    }

    .terminal-output {
        color: var(--text-muted);
        margin-bottom: 8px;
        padding-left: 16px;
    }

    .cursor {
        background: var(--accent);
        color: var(--bg);
        animation: blink 1s step-end infinite;
    }

    .features {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 16px;
    }

    .feature {
        padding: 20px;
        background: var(--bg-card);
        border: 1px solid var(--border);
        text-align: left;
        transition: border-color 0.2s;
    }

    .feature:hover {
        border-color: var(--accent);
    }

    .feature-icon {
        font-size: 24px;
        display: block;
        margin-bottom: 12px;
    }

    .feature h3 {
        font-size: 14px;
        text-transform: uppercase;
        margin-bottom: 8px;
        color: var(--text);
        letter-spacing: 0.5px;
    }

    .feature p {
        font-size: 12px;
        color: var(--text-muted);
        line-height: 1.5;
    }

    /* FAQ Section */
    .faq-section {
        margin-top: 60px;
        width: 100%;
        text-align: left;
    }

    .faq-title {
        font-size: 20px;
        text-transform: uppercase;
        letter-spacing: 2px;
        text-align: center;
        margin-bottom: 32px;
        color: var(--text);
    }

    .faq-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 16px;
    }

    .faq-item {
        background: var(--bg-card);
        border: 1px solid var(--border);
        transition: border-color 0.2s;
    }

    .faq-item[open] {
        border-color: var(--accent);
    }

    .faq-question {
        padding: 16px;
        cursor: pointer;
        font-size: 13px;
        font-weight: 600;
        color: var(--text);
        display: flex;
        align-items: center;
        gap: 12px;
        list-style: none;
        user-select: none;
    }

    .faq-question::-webkit-details-marker {
        display: none;
    }

    .faq-question::after {
        content: "+";
        margin-left: auto;
        font-size: 16px;
        color: var(--text-muted);
        transition: transform 0.2s;
    }

    .faq-item[open] .faq-question::after {
        content: "−";
        color: var(--accent);
    }

    .faq-icon {
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-elevated);
        border: 1px solid var(--border);
        font-size: 12px;
        flex-shrink: 0;
    }

    .faq-answer {
        padding: 0 16px 16px 52px;
        font-size: 12px;
        color: var(--text-secondary);
        line-height: 1.7;
    }

    .faq-answer p {
        margin-bottom: 12px;
    }

    .faq-answer p:last-child {
        margin-bottom: 0;
    }

    .faq-answer code {
        background: var(--bg-elevated);
        padding: 2px 6px;
        border: 1px solid var(--border);
        font-family: var(--font-mono);
        font-size: 11px;
        color: var(--accent);
    }

    .faq-answer ul {
        margin: 8px 0;
        padding-left: 20px;
    }

    .faq-answer li {
        margin-bottom: 6px;
    }

    .faq-answer strong {
        color: var(--text);
    }

    @keyframes blink {
        0%,
        100% {
            opacity: 1;
        }
        50% {
            opacity: 0;
        }
    }

    @media (max-width: 768px) {
        h1 {
            font-size: 24px;
        }

        .landing-actions {
            flex-direction: column;
        }

        .features {
            grid-template-columns: 1fr;
        }

        .faq-grid {
            grid-template-columns: 1fr;
        }

        .faq-answer {
            padding-left: 16px;
        }
    }
</style>