<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { terminal, type TerminalSession } from '$stores/terminal';

  export let session: TerminalSession;

  let containerElement: HTMLDivElement;
  let isAttached = false;

  onMount(() => {
    if (containerElement && session && !isAttached) {
      // Attach terminal to DOM
      terminal.attachTerminal(session.id, containerElement);
      isAttached = true;

      // Connect WebSocket if not already connected
      if (!session.ws || session.ws.readyState !== WebSocket.OPEN) {
        terminal.connectWebSocket(session.id);
      }
    }
  });

  onDestroy(() => {
    // Cleanup is handled by the terminal store when session is closed
  });

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
        session.ws.send(JSON.stringify({ type: 'input', data: text }));
      }
    });
  }

  // Reactive status
  $: status = session?.status || 'disconnected';
  $: isConnected = status === 'connected';
  $: isConnecting = status === 'connecting';
  $: isDisconnected = status === 'disconnected' || status === 'error';
</script>

<div class="terminal-panel-wrapper">
  <!-- Toolbar -->
  <div class="terminal-toolbar">
    <div class="toolbar-left">
      <span class="terminal-name">{session.name}</span>
      <span class="terminal-status" class:connected={isConnected} class:connecting={isConnecting} class:disconnected={isDisconnected}>
        <span class="status-indicator"></span>
        {status}
      </span>
    </div>

    <div class="toolbar-actions">
      {#if isDisconnected}
        <button class="toolbar-btn" on:click={handleReconnect} title="Reconnect">
          ↻ Reconnect
        </button>
      {/if}
      <button class="toolbar-btn" on:click={handleCopy} title="Copy Selection">
        📋 Copy
      </button>
      <button class="toolbar-btn" on:click={handlePaste} title="Paste">
        📥 Paste
      </button>
      <button class="toolbar-btn" on:click={handleClear} title="Clear Terminal">
        🗑 Clear
      </button>
    </div>
  </div>

  <!-- Terminal Container -->
  <div class="terminal-container" bind:this={containerElement}></div>

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
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }
</style>
