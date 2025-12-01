<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import {
    terminal,
    activeSession,
    sessionCount,
    isFloating,
    isDocked,
  } from '$stores/terminal';
  import TerminalPanel from './TerminalPanel.svelte';

  // Get active container ID for new tab functionality
  $: activeContainerId = $activeSession?.containerId || null;
  $: activeContainerName = $activeSession?.name?.replace(/\s*\(\d+\)$/, '') || 'Terminal';

  // Floating window state
  let floatingWindow: HTMLDivElement;
  let isDragging = false;
  let isResizing = false;
  let dragOffset = { x: 0, y: 0 };

  // Get state from store
  $: viewMode = $terminal.viewMode;
  $: isMinimized = $terminal.isMinimized;
  $: floatingPosition = $terminal.floatingPosition;
  $: floatingSize = $terminal.floatingSize;
  $: sessions = Array.from($terminal.sessions.entries());
  $: activeId = $terminal.activeSessionId;

  // Floating drag handlers
  function handleMouseDown(event: MouseEvent) {
    if (
      event.target instanceof HTMLElement &&
      (event.target.tagName === 'BUTTON' || event.target.closest('button'))
    ) {
      return;
    }

    isDragging = true;
    dragOffset = {
      x: event.clientX - floatingPosition.x,
      y: event.clientY - floatingPosition.y,
    };
  }

  function handleMouseMove(event: MouseEvent) {
    if (isDragging) {
      const x = Math.max(0, Math.min(window.innerWidth - 100, event.clientX - dragOffset.x));
      const y = Math.max(0, Math.min(window.innerHeight - 100, event.clientY - dragOffset.y));
      terminal.setFloatingPosition(x, y);
    }

    if (isResizing) {
      const width = Math.max(400, event.clientX - floatingPosition.x);
      const height = Math.max(300, event.clientY - floatingPosition.y);
      terminal.setFloatingSize(width, height);
    }
  }

  function handleMouseUp() {
    if (isDragging || isResizing) {
      isDragging = false;
      isResizing = false;
      // Fit terminals after resize
      setTimeout(() => terminal.fitAll(), 50);
    }
  }

  function handleResizeStart(event: MouseEvent) {
    event.preventDefault();
    event.stopPropagation();
    isResizing = true;
  }

  // Actions
  function toggleViewMode() {
    terminal.toggleViewMode();
  }

  function minimize() {
    terminal.minimize();
  }

  function restore() {
    terminal.restore();
  }

  function closeAll() {
    if (confirm('Close all terminal sessions?')) {
      terminal.closeAllSessions();
    }
  }

  function closeSession(sessionId: string) {
    terminal.closeSession(sessionId);
  }

  function setActive(sessionId: string) {
    terminal.setActiveSession(sessionId);
  }

  function getStatusClass(status: string): string {
    switch (status) {
      case 'connected':
        return 'status-connected';
      case 'connecting':
        return 'status-connecting';
      default:
        return 'status-disconnected';
    }
  }

  // Create a new tab for the active container
  function createNewTab() {
    if (activeContainerId) {
      terminal.createNewTab(activeContainerId, activeContainerName);
    }
  }

  // Window event listeners
  onMount(() => {
    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);
  });

  onDestroy(() => {
    window.removeEventListener('mousemove', handleMouseMove);
    window.removeEventListener('mouseup', handleMouseUp);
  });
</script>

{#if $sessionCount > 0}
  {#if $isFloating}
    <!-- Floating Terminal -->
    <div class="floating-container">
      <div
        bind:this={floatingWindow}
        class="floating-terminal"
        class:minimized={isMinimized}
        class:focused={true}
        style="left: {floatingPosition.x}px; top: {floatingPosition.y}px; width: {floatingSize.width}px; height: {floatingSize.height}px;"
      >
        <!-- Header -->
        <div
          class="floating-header"
          on:mousedown={handleMouseDown}
          role="toolbar"
        >
          <div class="floating-tabs">
            {#each sessions as [id, session] (id)}
              <button
                class="floating-tab"
                class:active={id === activeId}
                on:click={() => setActive(id)}
              >
                <span class="status-dot {getStatusClass(session.status)}"></span>
                <span class="tab-name">{session.name}</span>
                <button
                  class="tab-close"
                  on:click|stopPropagation={() => closeSession(id)}
                  title="Close"
                >
                  ×
                </button>
              </button>
            {/each}
          </div>

          <div class="floating-actions">
            <button on:click={createNewTab} title="New Tab" class="new-tab-btn">
              +
            </button>
            <button on:click={toggleViewMode} title="Dock Terminal">
              ⬒ Dock
            </button>
            <button on:click={minimize} title="Minimize">−</button>
            <button on:click={closeAll} title="Close All">×</button>
          </div>
        </div>

        <!-- Body -->
        <div class="floating-body">
          {#each sessions as [id, session] (id)}
            <div
              class="terminal-panel"
              class:active={id === activeId}
            >
              <TerminalPanel {session} />
            </div>
          {/each}
        </div>

        <!-- Resize Handle -->
        <div
          class="resize-handle"
          on:mousedown={handleResizeStart}
          role="separator"
        ></div>
      </div>
    </div>

    <!-- Minimized bar -->
    {#if isMinimized}
      <div class="minimized-bar">
        <button class="restore-btn" on:click={restore}>
          <span class="restore-icon">↑</span>
          <span>{$sessionCount} Terminal{$sessionCount > 1 ? 's' : ''}</span>
        </button>
      </div>
    {/if}
  {:else}
    <!-- Docked Terminal -->
    <div class="docked-terminal">
      <!-- Header -->
      <div class="docked-header">
        <div class="docked-tabs">
          {#each sessions as [id, session] (id)}
            <button
              class="docked-tab"
              class:active={id === activeId}
              on:click={() => setActive(id)}
            >
              <span class="status-dot {getStatusClass(session.status)}"></span>
              <span class="tab-name">{session.name}</span>
              <button
                class="tab-close"
                on:click|stopPropagation={() => closeSession(id)}
                title="Close"
              >
                ×
              </button>
            </button>
          {/each}
        </div>

        <div class="docked-actions">
          <button class="btn btn-primary btn-sm" on:click={createNewTab} title="New Tab">
            + New Tab
          </button>
          <button class="btn btn-secondary btn-sm" on:click={toggleViewMode}>
            ⬔ Float
          </button>
          <button class="btn btn-secondary btn-sm" on:click={() => window.history.back()}>
            ← Back
          </button>
          <button class="btn btn-danger btn-sm" on:click={closeAll}>
            ✕ Close All
          </button>
        </div>
      </div>

      <!-- Body -->
      <div class="docked-body">
        {#each sessions as [id, session] (id)}
          <div
            class="terminal-panel"
            class:active={id === activeId}
          >
            <TerminalPanel {session} />
          </div>
        {/each}
      </div>
    </div>
  {/if}
{/if}

<style>
  /* Floating Container */
  .floating-container {
    position: fixed;
    inset: 0;
    pointer-events: none;
    z-index: 1000;
  }

  .floating-terminal {
    position: absolute;
    display: flex;
    flex-direction: column;
    background: var(--bg-card);
    border: 1px solid var(--border);
    box-shadow: 0 0 40px rgba(0, 0, 0, 0.9), 0 0 1px var(--accent);
    pointer-events: auto;
    overflow: hidden;
    min-width: 400px;
    min-height: 300px;
  }

  .floating-terminal.focused {
    border-color: var(--accent);
    box-shadow: 0 0 40px rgba(0, 0, 0, 0.9), 0 0 10px rgba(0, 255, 65, 0.2);
  }

  .floating-terminal.minimized {
    display: none;
  }

  .floating-header {
    display: flex;
    align-items: center;
    padding: 6px 12px;
    background: #111;
    border-bottom: 1px solid var(--border);
    cursor: move;
    user-select: none;
    gap: 8px;
  }

  .floating-tabs {
    display: flex;
    flex: 1;
    gap: 2px;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .floating-tabs::-webkit-scrollbar {
    display: none;
  }

  .floating-tab {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-muted);
    font-size: 11px;
    font-family: var(--font-mono);
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.15s ease;
  }

  .floating-tab:hover {
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-secondary);
  }

  .floating-tab.active {
    background: rgba(0, 255, 65, 0.1);
    border-color: var(--accent);
    color: var(--accent);
  }

  .floating-actions {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .floating-actions button {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px 8px;
    font-size: 12px;
    font-family: var(--font-mono);
    transition: color 0.15s ease;
  }

  .floating-actions button:hover {
    color: var(--text);
  }

  .floating-actions .new-tab-btn {
    color: var(--accent);
    font-weight: bold;
    font-size: 14px;
    padding: 2px 10px;
    border: 1px solid var(--accent);
  }

  .floating-actions .new-tab-btn:hover {
    background: var(--accent);
    color: var(--bg);
  }

  .floating-body {
    flex: 1;
    overflow: hidden;
    background: #0a0a0a;
    position: relative;
  }

  .resize-handle {
    position: absolute;
    bottom: 0;
    right: 0;
    width: 16px;
    height: 16px;
    cursor: nwse-resize;
    background: linear-gradient(135deg, transparent 50%, var(--border) 50%);
  }

  .resize-handle:hover {
    background: linear-gradient(135deg, transparent 50%, var(--accent) 50%);
  }

  /* Docked Terminal */
  .docked-terminal {
    position: fixed;
    top: 60px;
    left: 0;
    right: 0;
    bottom: 0;
    background: var(--bg-card);
    z-index: 998;
    display: flex;
    flex-direction: column;
  }

  .docked-header {
    display: flex;
    align-items: center;
    padding: 8px 16px;
    background: #111;
    border-bottom: 1px solid var(--border);
    gap: 16px;
  }

  .docked-tabs {
    display: flex;
    flex: 1;
    gap: 4px;
    overflow-x: auto;
  }

  .docked-tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 14px;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-bottom: none;
    color: var(--text-muted);
    font-size: 12px;
    font-family: var(--font-mono);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .docked-tab:hover {
    color: var(--text);
    background: var(--bg-tertiary);
  }

  .docked-tab.active {
    color: var(--accent);
    border-color: var(--accent);
    border-bottom: 1px solid var(--bg-card);
    background: var(--bg-card);
  }

  .docked-actions {
    display: flex;
    gap: 8px;
  }

  .docked-body {
    flex: 1;
    position: relative;
    overflow: hidden;
    background: #050505;
  }

  /* Common Styles */
  .terminal-panel {
    position: absolute;
    inset: 0;
    display: none;
    overflow: hidden;
  }

  .terminal-panel.active {
    display: flex;
    flex-direction: column;
  }

  .status-dot {
    width: 6px;
    height: 6px;
  }

  .status-connected {
    background: var(--green);
  }

  .status-connecting {
    background: var(--yellow);
    animation: pulse 1s infinite;
  }

  .status-disconnected {
    background: var(--red);
  }

  .tab-name {
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tab-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 14px;
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
    opacity: 0.5;
  }

  .tab-close:hover {
    color: var(--red);
    opacity: 1;
  }

  /* Minimized Bar */
  .minimized-bar {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    background: var(--bg-card);
    border-top: 1px solid var(--border);
    padding: 8px 16px;
    z-index: 1001;
  }

  .restore-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--accent-dim);
    border: 1px solid var(--accent);
    color: var(--accent);
    padding: 6px 16px;
    font-family: var(--font-mono);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .restore-btn:hover {
    background: var(--accent);
    color: var(--bg);
  }

  .restore-icon {
    font-size: 14px;
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
