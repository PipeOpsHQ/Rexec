<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { collab } from '../stores/collab';
  import { containers } from '../stores/containers';
  import { slide } from 'svelte/transition';
  import PlatformIcon from './icons/PlatformIcon.svelte';
  import StatusIcon from './icons/StatusIcon.svelte';

  export let containerId: string;
  export let isOpen = false;
  export let compact = false;

  const dispatch = createEventDispatcher();

  let mode: 'view' | 'control' = 'view';
  let maxUsers = 5;
  let isStarting = false;
  let shareCode = '';
  let shareUrl = '';
  let copied = false;

  $: session = $collab.activeSession;
  $: participants = $collab.participants;
  $: isConnected = $collab.isConnected;
  
  // Get terminal info for display
  $: terminal = $containers.containers.find(c => c.id === containerId);
  $: terminalName = terminal?.name || containerId.slice(0, 12);
  $: terminalOS = terminal?.image || 'unknown';

  async function startSession() {
    isStarting = true;
    const result = await collab.startSession(containerId, mode, maxUsers);
    isStarting = false;
    
    if (result) {
      shareCode = result.shareCode;
      shareUrl = `${window.location.origin}/join/${shareCode}`;
      collab.connectWebSocket(shareCode);
    }
  }

  async function endSession() {
    if (session) {
      await collab.endSession(session.id);
    }
    close();
  }

  function copyLink() {
    navigator.clipboard.writeText(shareUrl);
    copied = true;
    setTimeout(() => copied = false, 2000);
  }

  function copyCode() {
    navigator.clipboard.writeText(shareCode);
    copied = true;
    setTimeout(() => copied = false, 2000);
  }

  function close() {
    isOpen = false;
    dispatch('close');
  }
</script>

{#if isOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="panel-overlay" on:click|self={close}>
    <div class="collab-panel" class:compact transition:slide={{ duration: 200 }}>
    <div class="panel-header">
      <div class="header-left">
        <span class="collab-icon">👥</span>
        <span class="title">SHARE</span>
      </div>
      <button class="close-btn" on:click={close}>×</button>
    </div>

    {#if !session}
      <div class="panel-content">
        <!-- Terminal being shared -->
        <div class="terminal-info">
          <PlatformIcon platform={terminalOS} size={20} />
          <div class="terminal-details">
            <span class="terminal-name">{terminalName}</span>
            <span class="terminal-os">{terminalOS}</span>
          </div>
        </div>

        <div class="option-row">
          <span class="label">Access Level</span>
          <div class="mode-toggle">
            <button class="mode-btn" class:active={mode === 'view'} on:click={() => mode = 'view'}>
              👁 View Only
            </button>
            <button class="mode-btn" class:active={mode === 'control'} on:click={() => mode = 'control'}>
              ✏️ Full Control
            </button>
          </div>
          <p class="mode-hint">
            {mode === 'view' 
              ? 'Viewers can watch but cannot type or execute commands' 
              : 'Collaborators can type and execute commands'}
          </p>
        </div>

        <div class="option-row">
          <span class="label">Max collaborators</span>
          <div class="slider-row">
            <input type="range" min="2" max="10" bind:value={maxUsers} class="slider" />
            <span class="slider-value">{maxUsers}</span>
          </div>
        </div>

        <button class="start-btn" on:click={startSession} disabled={isStarting}>
          {#if isStarting}
            <span class="spinner-sm"></span>
          {/if}
          {isStarting ? 'Starting...' : 'Share Terminal'}
        </button>
      </div>
    {:else}
      <div class="panel-content">
        <!-- Terminal being shared -->
        <div class="terminal-info shared">
          <PlatformIcon platform={terminalOS} size={20} />
          <div class="terminal-details">
            <span class="terminal-name">{terminalName}</span>
            <span class="sharing-badge">SHARING</span>
          </div>
        </div>

        <div class="share-section">
          <div class="share-link-box">
            <input class="share-url" readonly value={shareUrl} on:click|stopPropagation={(e) => e.currentTarget.select()} />
            <button class="copy-link-btn" on:click={copyLink}>
              {#if copied}
                <StatusIcon status="check" size={12} /> Copied
              {:else}
                <StatusIcon status="copy" size={12} /> Copy Link
              {/if}
            </button>
          </div>
          <p class="share-hint">Share this link with collaborators</p>
        </div>

        <div class="participants-section">
          <div class="section-header">
            <span>Collaborators</span>
            <span class="count">{participants.length}/{maxUsers}</span>
          </div>
          <div class="participants-list">
            {#each participants as p}
              <div class="participant">
                <span class="avatar" style="background: {p.color}">{p.username.charAt(0)}</span>
                <span class="name">{p.username}</span>
                <span class="role-tag" class:owner={p.role === 'owner'} class:viewer={p.role === 'viewer'}>
                  {p.role === 'owner' ? 'Host' : p.role === 'viewer' ? 'Viewing' : 'Editing'}
                </span>
              </div>
            {:else}
              <p class="empty">Waiting for collaborators...</p>
            {/each}
          </div>
        </div>

        <div class="status-row">
          <span class="status-dot" class:live={isConnected}></span>
          <span class="status-text">{isConnected ? 'Live' : 'Connecting'}</span>
          <span class="mode-tag" class:control={mode === 'control'}>
            {mode === 'view' ? 'View Only' : 'Full Control'}
          </span>
        </div>

        <button class="end-btn" on:click={endSession}>End Sharing</button>
      </div>
    {/if}
    </div>
  </div>
{/if}

<style>
  .panel-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10001;
  }

  .collab-panel {
    width: 340px;
    max-width: 95vw;
    max-height: 85vh;
    background: #0c0c10;
    border: 1px solid #1e1e28;
    border-radius: 8px;
    font-size: 11px;
    font-family: var(--font-mono, 'JetBrains Mono', monospace);
    box-shadow: 0 8px 40px rgba(0, 0, 0, 0.9), 0 0 20px rgba(0, 255, 136, 0.1);
    overflow: hidden;
  }

  .collab-panel.compact {
    width: 240px;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: #0f0f14;
    border-bottom: 1px solid #1e1e28;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .collab-icon {
    font-size: 12px;
  }

  .title {
    color: var(--accent, #00ff88);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 1.5px;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #555;
    font-size: 16px;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: 3px;
    transition: all 0.15s;
    line-height: 1;
  }

  .close-btn:hover {
    color: var(--accent, #00ff88);
    background: rgba(0, 255, 136, 0.1);
  }

  .panel-content {
    padding: 12px;
  }

  .terminal-info {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px;
    background: #0a0a0c;
    border: 1px solid #1e1e28;
    border-radius: 4px;
    margin-bottom: 12px;
  }

  .terminal-info.shared {
    background: rgba(0, 255, 136, 0.05);
    border-color: rgba(0, 255, 136, 0.2);
  }

  .terminal-details {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
  }

  .terminal-name {
    font-size: 12px;
    font-weight: 600;
    color: #e0e0e0;
  }

  .terminal-os {
    font-size: 9px;
    color: #666;
    text-transform: uppercase;
  }

  .sharing-badge {
    font-size: 8px;
    padding: 2px 6px;
    background: var(--accent, #00ff88);
    color: #000;
    border-radius: 2px;
    font-weight: 700;
    letter-spacing: 0.5px;
    animation: pulse-badge 2s ease-in-out infinite;
  }

  @keyframes pulse-badge {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }

  .mode-hint {
    margin: 6px 0 0 0;
    font-size: 9px;
    color: #555;
    line-height: 1.3;
  }

  .option-row {
    margin-bottom: 14px;
  }

  .label {
    display: block;
    color: #555;
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 6px;
  }

  .mode-toggle {
    display: flex;
    gap: 4px;
  }

  .mode-btn {
    flex: 1;
    padding: 8px 10px;
    background: #0a0a0c;
    border: 1px solid #1e1e28;
    border-radius: 4px;
    color: #666;
    font-size: 10px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
  }

  .mode-btn:hover {
    background: #0e0e12;
    border-color: #2a2a35;
  }

  .mode-btn.active {
    background: rgba(0, 255, 136, 0.08);
    border-color: rgba(0, 255, 136, 0.3);
    color: var(--accent, #00ff88);
  }

  .slider-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 10px;
    background: #0a0a0c;
    border: 1px solid #1e1e28;
    border-radius: 4px;
  }

  .slider {
    flex: 1;
    height: 3px;
    -webkit-appearance: none;
    background: #1e1e28;
    border-radius: 2px;
  }

  .slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 12px;
    height: 12px;
    background: var(--accent, #00ff88);
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.15s;
  }

  .slider::-webkit-slider-thumb:hover {
    transform: scale(1.1);
  }

  .slider::-moz-range-thumb {
    width: 12px;
    height: 12px;
    background: var(--accent, #00ff88);
    border-radius: 50%;
    cursor: pointer;
    border: none;
  }

  .slider-value {
    color: var(--accent, #00ff88);
    font-weight: 600;
    min-width: 20px;
    text-align: center;
    font-size: 11px;
  }

  .start-btn {
    width: 100%;
    padding: 10px;
    background: var(--accent, #00ff88);
    border: none;
    border-radius: 4px;
    color: #000;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: all 0.2s;
  }

  .start-btn:hover:not(:disabled) {
    background: #00cc6a;
  }

  .start-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .share-section {
    margin-bottom: 14px;
  }

  .share-link-box {
    display: flex;
    gap: 8px;
    margin-bottom: 6px;
  }

  .share-url {
    flex: 1;
    padding: 10px 12px;
    background: #0a0a0c;
    border: 1px solid rgba(0, 255, 136, 0.2);
    border-radius: 4px;
    color: var(--accent, #00ff88);
    font-size: 10px;
    font-family: inherit;
  }

  .share-url:focus {
    outline: none;
    border-color: rgba(0, 255, 136, 0.5);
  }

  .copy-link-btn {
    padding: 8px 14px;
    background: var(--accent, #00ff88);
    border: none;
    border-radius: 4px;
    color: #000;
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.15s;
  }

  .copy-link-btn:hover {
    background: #00cc6a;
  }

  .share-hint {
    margin: 0;
    font-size: 9px;
    color: #555;
    text-align: center;
  }

  .participants-section {
    margin-bottom: 12px;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    color: #555;
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 6px;
  }

  .count {
    color: var(--accent, #00ff88);
  }

  .participants-list {
    background: #0a0a0c;
    border: 1px solid #1e1e28;
    border-radius: 4px;
    max-height: 100px;
    overflow-y: auto;
  }

  .participant {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    border-bottom: 1px solid #151518;
    transition: background 0.15s;
  }

  .participant:last-child {
    border-bottom: none;
  }

  .participant:hover {
    background: rgba(0, 255, 136, 0.03);
  }

  .avatar {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 9px;
    font-weight: 700;
    color: #000;
  }

  .name {
    flex: 1;
    color: #bbb;
    font-size: 11px;
    font-weight: 500;
  }

  .role-tag {
    font-size: 8px;
    padding: 2px 6px;
    background: rgba(0, 255, 136, 0.08);
    border: 1px solid rgba(0, 255, 136, 0.15);
    border-radius: 3px;
    color: var(--accent, #00ff88);
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .role-tag.owner {
    background: rgba(255, 215, 0, 0.1);
    border-color: rgba(255, 215, 0, 0.3);
    color: #ffd700;
  }

  .role-tag.viewer {
    background: rgba(100, 100, 100, 0.1);
    border-color: rgba(100, 100, 100, 0.3);
    color: #888;
  }

  .empty {
    color: #444;
    text-align: center;
    padding: 14px;
    margin: 0;
    font-size: 10px;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    background: #0a0a0c;
    border: 1px solid #1e1e28;
    border-radius: 4px;
    margin-bottom: 10px;
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #444;
    transition: all 0.3s;
  }

  .status-dot.live {
    background: var(--accent, #00ff88);
    animation: pulse-live 1.5s ease-in-out infinite;
  }

  @keyframes pulse-live {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .status-text {
    flex: 1;
    color: #666;
    font-size: 10px;
    font-weight: 500;
  }

  .mode-tag {
    font-size: 8px;
    padding: 2px 6px;
    background: rgba(100, 100, 100, 0.1);
    border: 1px solid rgba(100, 100, 100, 0.2);
    border-radius: 3px;
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .mode-tag.control {
    background: rgba(0, 255, 136, 0.08);
    border-color: rgba(0, 255, 136, 0.2);
    color: var(--accent, #00ff88);
  }

  .end-btn {
    width: 100%;
    padding: 8px;
    background: rgba(255, 71, 87, 0.08);
    border: 1px solid rgba(255, 71, 87, 0.2);
    border-radius: 4px;
    color: #ff4757;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .end-btn:hover {
    background: rgba(255, 71, 87, 0.12);
    border-color: rgba(255, 71, 87, 0.4);
  }

  .spinner-sm {
    width: 12px;
    height: 12px;
    border: 2px solid transparent;
    border-top-color: currentColor;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Scrollbar */
  .participants-list::-webkit-scrollbar {
    width: 3px;
  }

  .participants-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .participants-list::-webkit-scrollbar-thumb {
    background: #222;
    border-radius: 2px;
  }

  .participants-list::-webkit-scrollbar-thumb:hover {
    background: #333;
  }

  /* Firefox scrollbar */
  .participants-list {
    scrollbar-width: thin;
    scrollbar-color: #222 transparent;
  }
</style>
