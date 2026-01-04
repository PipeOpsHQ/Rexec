<script lang="ts">
    import StatusIcon from "./icons/StatusIcon.svelte";

    export let onback: (() => void) | undefined = undefined;

    let copiedCommand = "";

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

    const currentHost =
        typeof window !== "undefined" ? window.location.host : "rexec.dev";
    const protocol =
        typeof window !== "undefined" ? window.location.protocol : "https:";
    const baseUrl = `${protocol}//${currentHost}`;

    // Helper to build code snippets with script tags (avoids Svelte parsing issues)
    function getQuickStartCode() {
        return `<script src="${baseUrl}/embed/rexec.min.js"><\/script>

<div id="terminal" style="width: 100%; height: 400px;"></div>

<script>
  const term = Rexec.embed('#terminal', {
    shareCode: 'your-share-code'
  });
<\/script>`;
    }
</script>

<div class="docs-page">
    <button class="back-btn" onclick={handleBack}>
        <span class="back-icon">←</span>
        <span>Back</span>
    </button>

    <div class="docs-content">
        <header class="docs-header">
            <div class="header-icon">
                <StatusIcon status="code" size={48} />
            </div>
            <h1>Embeddable Terminal Widget</h1>
            <p class="subtitle">
                Add a cloud terminal to any website with a single script tag
            </p>
        </header>

        <section class="docs-section">
            <h2>What is the Embed Widget?</h2>
            <p>
                The Rexec embed widget lets you add a fully-featured cloud
                terminal to any website. Similar to Google Cloud Shell, you can
                embed interactive terminal sessions in documentation, tutorials,
                learning platforms, or anywhere you need live command execution.
            </p>
            <div class="feature-grid">
                <div class="feature-card">
                    <StatusIcon status="bolt" size={20} />
                    <h4>One-Line Integration</h4>
                    <p>
                        Add a terminal with just a script tag and one line of
                        JavaScript
                    </p>
                </div>
                <div class="feature-card">
                    <StatusIcon status="user" size={20} />
                    <h4>Guest & Auth Modes</h4>
                    <p>
                        Support share codes for guests or API tokens for
                        authenticated access
                    </p>
                </div>
                <div class="feature-card">
                    <StatusIcon status="palette" size={20} />
                    <h4>Fully Customizable</h4>
                    <p>
                        Themes, fonts, sizes, and event callbacks for full
                        control
                    </p>
                </div>
                <div class="feature-card">
                    <StatusIcon status="cloud" size={20} />
                    <h4>Cloud Powered</h4>
                    <p>
                        Runs on Rexec infrastructure - no server setup required
                    </p>
                </div>
            </div>
        </section>

        <section class="docs-section">
            <h2>Quick Start</h2>
            <p>Add this to your HTML to embed a terminal:</p>
            <div class="code-block large">
                <code
                    >&lt;!-- Include the embed script --&gt;<br />&lt;script
                    src="{baseUrl}/embed/rexec.min.js"&gt;&lt;/script&gt;<br
                    /><br />&lt;!-- Create a container --&gt;<br />&lt;div
                    id="terminal" style="width: 100%; height:
                    400px;"&gt;&lt;/div&gt;<br /><br />&lt;!-- Initialize --&gt;<br
                    />&lt;script&gt;<br /> const term = Rexec.embed('#terminal',
                    &#123;<br /> shareCode: 'your-share-code'<br /> &#125;);<br
                    />&lt;/script&gt;</code
                >
                <button
                    class="copy-btn"
                    onclick={() =>
                        copyToClipboard(getQuickStartCode(), "quick")}
                >
                    {copiedCommand === "quick" ? "Copied!" : "Copy"}
                </button>
            </div>
            <p class="hint">
                Replace 'your-share-code' with an actual session share code from
                your Rexec dashboard.
            </p>
        </section>

        <section class="docs-section">
            <h2>Connection Methods</h2>

            <div class="method-card">
                <h3>
                    <StatusIcon status="user" size={20} />
                    Join via Share Code (Guest Access)
                </h3>
                <p>
                    Join an existing shared session. No authentication required
                    - perfect for tutorials and demos.
                </p>
                <div class="code-block">
                    <code
                        >const term = Rexec.embed('#terminal', &#123;<br />
                        shareCode: 'ABC123'<br />&#125;);</code
                    >
                    <button
                        class="copy-btn"
                        onclick={() =>
                            copyToClipboard(
                                `const term = Rexec.embed('#terminal', {\n  shareCode: 'ABC123'\n});`,
                                "share",
                            )}
                    >
                        {copiedCommand === "share" ? "Copied!" : "Copy"}
                    </button>
                </div>
            </div>

            <div class="method-card">
                <h3>
                    <StatusIcon status="key" size={20} />
                    Connect to Existing Container
                </h3>
                <p>Connect to a container you own. Requires an API token.</p>
                <div class="code-block">
                    <code
                        >const term = Rexec.embed('#terminal', &#123;<br />
                        token: 'your-api-token',<br /> container: 'container-id'<br
                        />&#125;);</code
                    >
                    <button
                        class="copy-btn"
                        onclick={() =>
                            copyToClipboard(
                                `const term = Rexec.embed('#terminal', {\n  token: 'your-api-token',\n  container: 'container-id'\n});`,
                                "container",
                            )}
                    >
                        {copiedCommand === "container" ? "Copied!" : "Copy"}
                    </button>
                </div>
            </div>

            <div class="method-card">
                <h3>
                    <StatusIcon status="plus" size={20} />
                    Create New Container On-Demand
                </h3>
                <p>
                    Spin up a fresh container for each user. Great for
                    interactive learning platforms.
                </p>
                <div class="code-block">
                    <code
                        >const term = Rexec.embed('#terminal', &#123;<br />
                        token: 'your-api-token',<br /> role: 'ubuntu' // or
                        'node', 'python', 'go', 'rust'<br />&#125;);</code
                    >
                    <button
                        class="copy-btn"
                        onclick={() =>
                            copyToClipboard(
                                `const term = Rexec.embed('#terminal', {\n  token: 'your-api-token',\n  role: 'ubuntu'\n});`,
                                "create",
                            )}
                    >
                        {copiedCommand === "create" ? "Copied!" : "Copy"}
                    </button>
                </div>
            </div>
        </section>

        <section class="docs-section">
            <h2>Configuration Options</h2>
            <div class="options-table">
                <div class="option-row header">
                    <span class="option-name">Option</span>
                    <span class="option-type">Type</span>
                    <span class="option-default">Default</span>
                    <span class="option-desc">Description</span>
                </div>
                <div class="option-row">
                    <span class="option-name"><code>token</code></span>
                    <span class="option-type">string</span>
                    <span class="option-default">-</span>
                    <span class="option-desc">API token for authentication</span
                    >
                </div>
                <div class="option-row">
                    <span class="option-name"><code>container</code></span>
                    <span class="option-type">string</span>
                    <span class="option-default">-</span>
                    <span class="option-desc">Container ID to connect to</span>
                </div>
                <div class="option-row">
                    <span class="option-name"><code>shareCode</code></span>
                    <span class="option-type">string</span>
                    <span class="option-default">-</span>
                    <span class="option-desc"
                        >Share code for joining sessions</span
                    >
                </div>
                <div class="option-row">
                    <span class="option-name"><code>role</code></span>
                    <span class="option-type">string</span>
                    <span class="option-default">-</span>
                    <span class="option-desc"
                        >Environment type for new containers</span
                    >
                </div>
                <div class="option-row">
                    <span class="option-name"><code>baseUrl</code></span>
                    <span class="option-type">string</span>
                    <span class="option-default">'https://rexec.dev'</span>
                    <span class="option-desc">API base URL</span>
                </div>
                <div class="option-row">
                    <span class="option-name"><code>theme</code></span>
                    <span class="option-type">'dark' | 'light' | object</span>
                    <span class="option-default">'dark'</span>
                    <span class="option-desc">Terminal color theme</span>
                </div>
                <div class="option-row">
                    <span class="option-name"><code>fontSize</code></span>
                    <span class="option-type">number</span>
                    <span class="option-default">14</span>
                    <span class="option-desc">Font size in pixels</span>
                </div>
                <div class="option-row">
                    <span class="option-name"><code>cursorStyle</code></span>
                    <span class="option-type"
                        >'block' | 'underline' | 'bar'</span
                    >
                    <span class="option-default">'block'</span>
                    <span class="option-desc">Cursor appearance</span>
                </div>
                <div class="option-row">
                    <span class="option-name"><code>scrollback</code></span>
                    <span class="option-type">number</span>
                    <span class="option-default">5000</span>
                    <span class="option-desc">Lines in scrollback buffer</span>
                </div>
                <div class="option-row">
                    <span class="option-name"><code>showStatus</code></span>
                    <span class="option-type">boolean</span>
                    <span class="option-default">true</span>
                    <span class="option-desc"
                        >Show connection status overlay</span
                    >
                </div>
                <div class="option-row">
                    <span class="option-name"><code>autoReconnect</code></span>
                    <span class="option-type">boolean</span>
                    <span class="option-default">true</span>
                    <span class="option-desc">Auto-reconnect on disconnect</span
                    >
                </div>
                <div class="option-row">
                    <span class="option-name"><code>initialCommand</code></span>
                    <span class="option-type">string</span>
                    <span class="option-default">-</span>
                    <span class="option-desc">Command to run after connect</span
                    >
                </div>
            </div>
        </section>

        <section class="docs-section">
            <h2>Event Callbacks</h2>
            <p>Listen to terminal events for custom behavior:</p>
            <div class="code-block">
                <code
                    >const term = Rexec.embed('#terminal', &#123;<br />
                    shareCode: 'ABC123',<br /><br /> onReady: (terminal) =&gt;
                    &#123;<br /> console.log('Terminal connected!');<br />
                    console.log('Session:', terminal.session);<br /> &#125;,<br
                    /><br /> onStateChange: (state) =&gt; &#123;<br /> //
                    'idle', 'connecting', 'connected', 'reconnecting', 'error'<br
                    />
                    console.log('State:', state);<br /> &#125;,<br /><br />
                    onData: (data) =&gt; &#123;<br /> console.log('Output:',
                    data);<br /> &#125;,<br /><br /> onError: (error) =&gt;
                    &#123;<br /> console.error('Error:', error.code,
                    error.message);<br /> &#125;<br />&#125;);</code
                >
                <button
                    class="copy-btn"
                    onclick={() =>
                        copyToClipboard(
                            `const term = Rexec.embed('#terminal', {\n  shareCode: 'ABC123',\n\n  onReady: (terminal) => {\n    console.log('Terminal connected!');\n  },\n\n  onStateChange: (state) => {\n    console.log('State:', state);\n  },\n\n  onData: (data) => {\n    console.log('Output:', data);\n  },\n\n  onError: (error) => {\n    console.error('Error:', error.code, error.message);\n  }\n});`,
                            "events",
                        )}
                >
                    {copiedCommand === "events" ? "Copied!" : "Copy"}
                </button>
            </div>
        </section>

        <section class="docs-section">
            <h2>Terminal API</h2>
            <p>Control the terminal programmatically:</p>

            <div class="api-group">
                <h3>Methods</h3>
                <div class="cli-commands">
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command"
                                >term.write('echo "Hello"')</code
                            >
                            <span class="command-desc"
                                >Write to terminal input</span
                            >
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.writeln('ls -la')</code>
                            <span class="command-desc"
                                >Write with newline (executes command)</span
                            >
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.clear()</code>
                            <span class="command-desc"
                                >Clear the terminal screen</span
                            >
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.fit()</code>
                            <span class="command-desc"
                                >Fit terminal to container size</span
                            >
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.focus()</code>
                            <span class="command-desc">Focus the terminal</span>
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.setTheme('light')</code>
                            <span class="command-desc"
                                >Change theme dynamically</span
                            >
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.setFontSize(16)</code>
                            <span class="command-desc">Change font size</span>
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.destroy()</code>
                            <span class="command-desc"
                                >Clean up and disconnect</span
                            >
                        </div>
                    </div>
                </div>
            </div>

            <div class="api-group">
                <h3>Properties</h3>
                <div class="cli-commands">
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.state</code>
                            <span class="command-desc"
                                >Current connection state</span
                            >
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.session</code>
                            <span class="command-desc"
                                >Session info (id, containerId, etc.)</span
                            >
                        </div>
                    </div>
                    <div class="command-item">
                        <div class="command-header">
                            <code class="command">term.stats</code>
                            <span class="command-desc"
                                >Container stats (cpu, memory, disk)</span
                            >
                        </div>
                    </div>
                </div>
            </div>
        </section>

        <section class="docs-section">
            <h2>Custom Themes</h2>
            <p>Create your own color scheme:</p>
            <div class="code-block">
                <code
                    >const term = Rexec.embed('#terminal', &#123;<br />
                    shareCode: 'ABC123',<br /> theme: &#123;<br /> background:
                    '#1a1b26',<br /> foreground: '#a9b1d6',<br /> cursor:
                    '#c0caf5',<br /> black: '#15161e',<br /> red: '#f7768e',<br
                    />
                    green: '#9ece6a',<br /> yellow: '#e0af68',<br /> blue:
                    '#7aa2f7',<br /> magenta: '#bb9af7',<br /> cyan: '#7dcfff',<br
                    />
                    white: '#a9b1d6'<br /> &#125;<br />&#125;);</code
                >
                <button
                    class="copy-btn"
                    onclick={() =>
                        copyToClipboard(
                            `const term = Rexec.embed('#terminal', {\n  shareCode: 'ABC123',\n  theme: {\n    background: '#1a1b26',\n    foreground: '#a9b1d6',\n    cursor: '#c0caf5',\n    black: '#15161e',\n    red: '#f7768e',\n    green: '#9ece6a',\n    yellow: '#e0af68',\n    blue: '#7aa2f7',\n    magenta: '#bb9af7',\n    cyan: '#7dcfff',\n    white: '#a9b1d6'\n  }\n});`,
                            "theme",
                        )}
                >
                    {copiedCommand === "theme" ? "Copied!" : "Copy"}
                </button>
            </div>
            <p class="hint">
                Use Rexec.DARK_THEME or Rexec.LIGHT_THEME as presets.
            </p>
        </section>

        <section class="docs-section">
            <h2>Use Cases</h2>
            <div class="feature-grid">
                <div class="feature-card">
                    <StatusIcon status="book" size={20} />
                    <h4>Interactive Documentation</h4>
                    <p>Let users try commands directly in your docs</p>
                </div>
                <div class="feature-card">
                    <StatusIcon status="graduation" size={20} />
                    <h4>Learning Platforms</h4>
                    <p>Hands-on coding exercises with real environments</p>
                </div>
                <div class="feature-card">
                    <StatusIcon status="play" size={20} />
                    <h4>Product Demos</h4>
                    <p>Showcase CLI tools without users installing anything</p>
                </div>
                <div class="feature-card">
                    <StatusIcon status="robot" size={20} />
                    <h4>AI Agent Sandboxes</h4>
                    <p>Give AI assistants a safe execution environment</p>
                </div>
            </div>
        </section>

        <section class="docs-section">
            <h2>Browser Support</h2>
            <div class="browser-grid">
                <div class="browser-item">
                    <StatusIcon status="check" size={20} />
                    <span>Chrome 80+</span>
                </div>
                <div class="browser-item">
                    <StatusIcon status="check" size={20} />
                    <span>Firefox 75+</span>
                </div>
                <div class="browser-item">
                    <StatusIcon status="check" size={20} />
                    <span>Safari 13+</span>
                </div>
                <div class="browser-item">
                    <StatusIcon status="check" size={20} />
                    <span>Edge 80+</span>
                </div>
            </div>
        </section>

        <section class="docs-section">
            <h2>FAQ</h2>
            <div class="faq-list">
                <div class="faq-item">
                    <h4>How do I get a share code?</h4>
                    <p>
                        Create a terminal in your Rexec dashboard, click the
                        share button, and copy the code. Share codes allow guest
                        access to your terminal session.
                    </p>
                </div>
                <div class="faq-item">
                    <h4>How do I get an API token?</h4>
                    <p>
                        Go to <a href="/account/api">Account → API Tokens</a> to generate
                        tokens for programmatic access.
                    </p>
                </div>
                <div class="faq-item">
                    <h4>Is there a rate limit?</h4>
                    <p>
                        Free tier has limits on concurrent sessions. Upgrade to
                        Pro for higher limits and priority access.
                    </p>
                </div>
                <div class="faq-item">
                    <h4>Can I self-host?</h4>
                    <p>
                        The embed widget connects to Rexec cloud infrastructure.
                        For on-premise needs, contact us about enterprise
                        options.
                    </p>
                </div>
            </div>
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

    .docs-section h3 {
        display: flex;
        align-items: center;
        gap: 10px;
        font-size: 16px;
        margin: 0 0 12px 0;
        color: var(--text);
    }

    .docs-section h3 :global(svg) {
        color: var(--accent);
    }

    .docs-section p {
        font-size: 14px;
        color: var(--text-muted);
        line-height: 1.7;
        margin: 0 0 16px 0;
    }

    .feature-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 16px;
        margin-top: 20px;
    }

    .feature-card {
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 20px;
    }

    .feature-card :global(svg) {
        color: var(--accent);
        margin-bottom: 12px;
    }

    .feature-card h4 {
        font-size: 14px;
        margin: 0 0 8px 0;
        color: var(--text);
    }

    .feature-card p {
        font-size: 12px;
        color: var(--text-muted);
        margin: 0;
        line-height: 1.5;
    }

    .method-card {
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 20px;
        margin-bottom: 16px;
    }

    .method-card h3 {
        margin-top: 0;
    }

    .method-card p {
        margin-bottom: 12px;
    }

    .code-block {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        background: var(--bg-tertiary);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 16px;
        margin: 12px 0;
    }

    .code-block.large {
        padding: 20px;
    }

    .code-block code {
        flex: 1;
        font-family: var(--font-mono);
        font-size: 13px;
        color: var(--accent);
        line-height: 1.6;
        white-space: pre-wrap;
        word-break: break-all;
    }

    .copy-btn {
        flex-shrink: 0;
        padding: 6px 12px;
        background: transparent;
        border: 1px solid var(--border);
        border-radius: 4px;
        color: var(--text-muted);
        font-size: 11px;
        font-family: var(--font-mono);
        cursor: pointer;
        transition: all 0.15s ease;
    }

    .copy-btn:hover {
        border-color: var(--accent);
        color: var(--accent);
        background: var(--accent-dim);
    }

    .hint {
        font-size: 12px !important;
        color: var(--text-muted) !important;
        font-style: italic;
    }

    .options-table {
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
        overflow: hidden;
    }

    .option-row {
        display: grid;
        grid-template-columns: 140px 150px 130px 1fr;
        padding: 12px 16px;
        border-bottom: 1px solid var(--border);
        font-size: 13px;
    }

    .option-row:last-child {
        border-bottom: none;
    }

    .option-row.header {
        background: var(--bg-tertiary);
        font-weight: 600;
        color: var(--text);
        text-transform: uppercase;
        font-size: 11px;
        letter-spacing: 0.5px;
    }

    .option-name code {
        font-family: var(--font-mono);
        color: var(--accent);
        font-size: 12px;
    }

    .option-type {
        color: var(--text-muted);
        font-family: var(--font-mono);
        font-size: 11px;
    }

    .option-default {
        color: var(--text-muted);
        font-family: var(--font-mono);
        font-size: 11px;
    }

    .option-desc {
        color: var(--text);
    }

    .api-group {
        margin-bottom: 24px;
    }

    .api-group h3 {
        font-size: 14px;
        margin: 0 0 12px 0;
        color: var(--text);
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .cli-commands {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .command-item {
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 6px;
        padding: 12px 16px;
    }

    .command-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
    }

    .command {
        font-family: var(--font-mono);
        font-size: 13px;
        color: var(--accent);
        background: var(--bg-tertiary);
        padding: 4px 8px;
        border-radius: 4px;
    }

    .command-desc {
        color: var(--text-muted);
        font-size: 13px;
    }

    .browser-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
        gap: 12px;
    }

    .browser-item {
        display: flex;
        align-items: center;
        gap: 10px;
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 6px;
        padding: 12px 16px;
        font-size: 14px;
        color: var(--text);
    }

    .browser-item :global(svg) {
        color: var(--success);
    }

    .faq-list {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .faq-item {
        background: var(--bg-secondary);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 20px;
    }

    .faq-item h4 {
        font-size: 14px;
        margin: 0 0 8px 0;
        color: var(--text);
    }

    .faq-item p {
        font-size: 13px;
        color: var(--text-muted);
        margin: 0;
        line-height: 1.6;
    }

    .faq-item a {
        color: var(--accent);
        text-decoration: none;
    }

    .faq-item a:hover {
        text-decoration: underline;
    }

    @media (max-width: 768px) {
        .docs-page {
            padding: 16px;
        }

        .docs-header h1 {
            font-size: 28px;
        }

        .option-row {
            grid-template-columns: 1fr;
            gap: 4px;
        }

        .option-row.header {
            display: none;
        }

        .option-name::before {
            content: "Option: ";
            color: var(--text-muted);
            font-size: 10px;
        }

        .option-type::before {
            content: "Type: ";
            color: var(--text-muted);
        }

        .option-default::before {
            content: "Default: ";
            color: var(--text-muted);
        }

        .command-header {
            flex-direction: column;
            align-items: flex-start;
            gap: 8px;
        }

        .code-block {
            flex-direction: column;
        }

        .copy-btn {
            align-self: flex-end;
        }
    }

    @media (max-width: 480px) {
        .feature-grid {
            grid-template-columns: 1fr;
        }

        .browser-grid {
            grid-template-columns: 1fr 1fr;
        }
    }
</style>
