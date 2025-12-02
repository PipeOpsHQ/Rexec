<script lang="ts">
    import { onMount, onDestroy, tick } from "svelte";
    import { terminal, type TerminalSession } from "$stores/terminal";
    import { toast } from "$stores/toast";

    export let session: TerminalSession;

    let containerElement: HTMLDivElement;
    let attachedToContainer: HTMLDivElement | null = null;

    // Synchronously check and attach terminal if needed
    async function ensureAttached() {
        if (!containerElement || !session?.terminal) return;

        // Only reattach if the container has changed
        if (attachedToContainer === containerElement) return;

        // Wait for DOM to be ready
        await tick();

        // Clear container before reattaching
        containerElement.innerHTML = "";

        try {
            // Check if terminal is already attached somewhere else
            if (session.terminal.element?.parentElement) {
                // Detach from previous parent first
                session.terminal.element.parentElement.innerHTML = "";
            }

            session.terminal.open(containerElement);
            attachedToContainer = containerElement;

            // Fit and focus after attachment with a small delay
            setTimeout(() => {
                if (session?.terminal && containerElement) {
                    terminal.fitSession(session.id);
                    session.terminal.focus();
                }
            }, 50);
        } catch (e) {
            console.error("Failed to attach terminal:", e);
            // Try to recover by recreating attachment
            attachedToContainer = null;
        }
    }

    onMount(async () => {
        if (containerElement && session) {
            if (session.terminal?.element) {
                // Terminal already exists, reattach to new container
                await ensureAttached();
            } else {
                // First time - let the store create and attach the terminal
                terminal.attachTerminal(session.id, containerElement);
                attachedToContainer = containerElement;
            }

            // Connect WebSocket if not already connected
            if (
                !session.ws ||
                (session.ws.readyState !== WebSocket.OPEN &&
                    session.ws.readyState !== WebSocket.CONNECTING)
            ) {
                terminal.connectWebSocket(session.id);
            }

            // Fit terminal after mount
            setTimeout(() => {
                if (session?.terminal) {
                    terminal.fitSession(session.id);
                }
            }, 100);
        }
    });

    onDestroy(() => {
        // Don't dispose terminal - it's managed by the store
        attachedToContainer = null;
    });

    // Use reactive statement to handle container changes (dock/float switch)
    $: if (
        containerElement &&
        session?.terminal &&
        attachedToContainer !== containerElement
    ) {
        // Use async IIFE to handle the async ensureAttached
        (async () => {
            await ensureAttached();
        })();
    }

    // Actions
    function handleReconnect() {
        terminal.reconnectSession(session.id);
    }

    function handleClear() {
        if (session.terminal) {
            session.terminal.clear();
        }
    }

    function handleCopy() {
        if (session.terminal) {
            const selection = session.terminal.getSelection();
            if (selection) {
                navigator.clipboard.writeText(selection);
            }
        }
    }

    function handlePaste() {
        navigator.clipboard.readText().then((text) => {
            if (session.ws && session.ws.readyState === WebSocket.OPEN) {
                session.ws.send(JSON.stringify({ type: "input", data: text }));
            }
        });
    }

    function handleCopyLink() {
        const url = `${window.location.origin}/terminal/${session.containerId}`;
        navigator.clipboard
            .writeText(url)
            .then(() => {
                toast.success("Terminal link copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy link");
            });
    }

    // Focus terminal when clicking on container
    function handleContainerClick() {
        if (session.terminal) {
            session.terminal.focus();
        }
    }

    // Reactive status
    $: status = session?.status || "disconnected";
    $: isConnected = status === "connected";
    $: isConnecting = status === "connecting";
    $: isDisconnected = status === "disconnected" || status === "error";
    $: isSettingUp = session?.isSettingUp || false;
    $: setupMessage = session?.setupMessage || "";
</script>

<div class="terminal-panel-wrapper">
    <!-- Toolbar -->
    <div class="terminal-toolbar">
        <div class="toolbar-left">
            <span class="terminal-name">{session.name}</span>
            <span
                class="terminal-status"
                class:connected={isConnected}
                class:connecting={isConnecting}
                class:disconnected={isDisconnected}
            >
                <span class="status-indicator"></span>
                {status}
            </span>
            {#if isSettingUp}
                <span class="setup-indicator">
                    <span class="setup-spinner"></span>
                    Installing...
                </span>
            {/if}
        </div>

        <div class="toolbar-actions">
            {#if isDisconnected}
                <button
                    class="toolbar-btn"
                    on:click={handleReconnect}
                    title="Reconnect"
                >
                    ↻ Reconnect
                </button>
            {/if}
            <button
                class="toolbar-btn"
                on:click={handleCopyLink}
                title="Copy Terminal Link"
            >
                🔗 Copy Link
            </button>
            <button
                class="toolbar-btn"
                on:click={handleCopy}
                title="Copy Selection"
            >
                📋 Copy
            </button>
            <button class="toolbar-btn" on:click={handlePaste} title="Paste">
                📥 Paste
            </button>
            <button
                class="toolbar-btn"
                on:click={handleClear}
                title="Clear Terminal"
            >
                🗑 Clear
            </button>
        </div>
    </div>

    <!-- Terminal Container -->
    <div
        class="terminal-container"
        bind:this={containerElement}
        on:click={handleContainerClick}
        on:keydown={() => {}}
        role="textbox"
        tabindex="0"
    ></div>

    <!-- Connection overlay -->
    {#if isConnecting}
        <div class="connection-overlay">
            <div class="connection-spinner"></div>
            <span>Connecting...</span>
        </div>
    {/if}

    {#if isDisconnected}
        <div class="disconnected-overlay">
            <span class="disconnected-icon">⚠</span>
            <span>Disconnected</span>
            <button class="reconnect-btn" on:click={handleReconnect}>
                ↻ Reconnect
            </button>
        </div>
    {/if}

    {#if isSettingUp}
        <div class="setup-overlay">
            <div class="setup-content">
                <div class="setup-spinner-large"></div>
                <span class="setup-title">Installing packages...</span>
                <span class="setup-detail">{setupMessage}</span>
            </div>
        </div>
    {/if}
</div>

<style>
    .terminal-panel-wrapper {
        display: flex;
        flex-direction: column;
        height: 100%;
        width: 100%;
        position: relative;
        background: #0a0a0a;
    }

    /* Toolbar */
    .terminal-toolbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 6px 12px;
        background: #111;
        border-bottom: 1px solid var(--border);
        flex-shrink: 0;
    }

    .toolbar-left {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .terminal-name {
        font-size: 12px;
        color: var(--text);
        font-weight: 500;
    }

    .terminal-status {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 10px;
        text-transform: uppercase;
        padding: 2px 8px;
        background: var(--bg-tertiary);
        border: 1px solid var(--border);
    }

    .status-indicator {
        width: 6px;
        height: 6px;
    }

    .terminal-status.connected {
        border-color: var(--green);
        color: var(--green);
    }

    .terminal-status.connected .status-indicator {
        background: var(--green);
    }

    .terminal-status.connecting {
        border-color: var(--yellow);
        color: var(--yellow);
    }

    .terminal-status.connecting .status-indicator {
        background: var(--yellow);
        animation: pulse 1s infinite;
    }

    .terminal-status.disconnected {
        border-color: var(--red);
        color: var(--red);
    }

    .terminal-status.disconnected .status-indicator {
        background: var(--red);
    }

    .toolbar-actions {
        display: flex;
        gap: 4px;
    }

    .toolbar-btn {
        background: none;
        border: 1px solid transparent;
        color: var(--text-muted);
        font-size: 11px;
        font-family: var(--font-mono);
        padding: 4px 8px;
        cursor: pointer;
        transition: all 0.15s;
    }

    .toolbar-btn:hover {
        color: var(--text);
        background: var(--bg-tertiary);
        border-color: var(--border);
    }

    /* Terminal Container */
    .terminal-container {
        flex: 1;
        width: 100%;
        overflow: hidden;
        padding: 8px;
    }

    .terminal-container:focus {
        outline: none;
    }

    .terminal-container :global(.xterm) {
        height: 100% !important;
        width: 100% !important;
    }

    .terminal-container :global(.xterm-viewport) {
        overflow-y: auto !important;
    }

    .terminal-container :global(.xterm-screen) {
        height: 100% !important;
    }

    /* Connection Overlay */
    .connection-overlay {
        position: absolute;
        inset: 0;
        top: 40px; /* Below toolbar */
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 16px;
        background: rgba(10, 10, 10, 0.9);
        z-index: 10;
    }

    .connection-spinner {
        width: 32px;
        height: 32px;
        border: 3px solid var(--border);
        border-top-color: var(--accent);
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }

    .connection-overlay span {
        color: var(--text-muted);
        font-size: 13px;
    }

    /* Disconnected Overlay */
    .disconnected-overlay {
        position: absolute;
        bottom: 16px;
        left: 50%;
        transform: translateX(-50%);
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 10px 16px;
        background: rgba(255, 0, 60, 0.1);
        border: 1px solid var(--red);
        z-index: 10;
    }

    .disconnected-icon {
        font-size: 16px;
    }

    .disconnected-overlay span {
        color: var(--red);
        font-size: 12px;
    }

    .reconnect-btn {
        background: var(--red);
        border: none;
        color: var(--bg);
        font-size: 11px;
        font-family: var(--font-mono);
        padding: 4px 10px;
        cursor: pointer;
        transition: opacity 0.15s;
    }

    .reconnect-btn:hover {
        opacity: 0.9;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    @keyframes pulse {
        0%,
        100% {
            opacity: 1;
        }
        50% {
            opacity: 0.5;
        }
    }

    /* Setup Indicator */
    .setup-indicator {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 10px;
        text-transform: uppercase;
        padding: 2px 8px;
        background: rgba(0, 200, 255, 0.1);
        border: 1px solid var(--cyan, #00c8ff);
        color: var(--cyan, #00c8ff);
        animation: fadeIn 0.2s ease;
    }

    .setup-spinner {
        width: 8px;
        height: 8px;
        border: 1.5px solid rgba(0, 200, 255, 0.3);
        border-top-color: var(--cyan, #00c8ff);
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
    }

    /* Setup Overlay */
    .setup-overlay {
        position: absolute;
        bottom: 16px;
        right: 16px;
        z-index: 10;
        animation: fadeIn 0.2s ease;
    }

    .setup-content {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 10px 16px;
        background: rgba(0, 200, 255, 0.1);
        border: 1px solid var(--cyan, #00c8ff);
        backdrop-filter: blur(4px);
    }

    .setup-spinner-large {
        width: 16px;
        height: 16px;
        border: 2px solid rgba(0, 200, 255, 0.3);
        border-top-color: var(--cyan, #00c8ff);
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
    }

    .setup-title {
        font-size: 12px;
        color: var(--cyan, #00c8ff);
        font-weight: 500;
    }

    .setup-detail {
        font-size: 11px;
        color: var(--text-muted);
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }
</style>
