<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { collab } from '../stores/collab';
  import { auth, isAuthenticated } from '../stores/auth';
  import { toast } from '../stores/toast';

  export let code: string = '';

  const dispatch = createEventDispatcher();

  let isLoading = true;
  let error = '';
  let needsAuth = false;
  let guestEmail = '';
  let isSubmittingGuest = false;
  let sessionInfo: {
    sessionId: string;
    containerId: string;
    containerName: string;
    mode: string;
    role: string;
    expiresAt: string;
  } | null = null;

  async function attemptJoin() {
    if (!code) {
      error = 'Invalid session code';
      isLoading = false;
      return;
    }

    // If not authenticated, show guest login form
    if (!$isAuthenticated) {
      needsAuth = true;
      isLoading = false;
      return;
    }

    // Try to join the session
    try {
      const result = await collab.joinSession(code);
      if (result) {
        sessionInfo = {
          sessionId: result.id,
          containerId: result.containerId,
          containerName: result.containerName,
          mode: result.mode,
          role: result.role,
          expiresAt: result.expiresAt
        };
        // Connect to the collab websocket
        collab.connectWebSocket(code);
        toast.success(`Joined terminal as ${result.role}`);
      } else {
        error = 'Terminal not found or sharing has ended';
      }
    } catch (e) {
      error = 'Failed to join terminal';
    }
    
    isLoading = false;
  }

  onMount(() => {
    attemptJoin();
  });

  // React to auth changes - retry join when user authenticates
  $: if ($isAuthenticated && needsAuth) {
    needsAuth = false;
    isLoading = true;
    attemptJoin();
  }

  async function handleGuestSubmit() {
    if (!guestEmail.trim()) {
      toast.error('Please enter your email');
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(guestEmail.trim())) {
      toast.error('Please enter a valid email');
      return;
    }

    isSubmittingGuest = true;
    const result = await auth.guestLogin(guestEmail.trim());
    isSubmittingGuest = false;

    if (result.success) {
      // attemptJoin will be triggered by the reactive statement above
      toast.success('Guest session started!');
    } else {
      toast.error('Failed to start guest session');
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !isSubmittingGuest) {
      handleGuestSubmit();
    }
  }

  function joinTerminal() {
    if (sessionInfo) {
      dispatch('joined', {
        containerId: sessionInfo.containerId,
        containerName: `Collab Session (${code})`
      });
    }
  }

  function cancel() {
    dispatch('cancel');
  }
</script>

<div class="join-container">
  <div class="join-card">
    <div class="join-header">
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
        <circle cx="9" cy="7" r="4"/>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
        <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
      </svg>
      <h1>Join Terminal</h1>
      <p class="code-display">{code}</p>
    </div>

    {#if isLoading}
      <div class="loading">
        <div class="spinner"></div>
        <p>Connecting to terminal...</p>
      </div>
    {:else if needsAuth}
      <div class="auth-prompt">
        <p class="auth-description">Enter your email to join this shared terminal</p>
        <div class="form-group">
          <label for="guest-email">Email Address</label>
          <input
            type="email"
            id="guest-email"
            bind:value={guestEmail}
            on:keydown={handleKeydown}
            placeholder="you@example.com"
            disabled={isSubmittingGuest}
          />
        </div>
        <div class="actions">
          <button class="btn btn-secondary" on:click={cancel} disabled={isSubmittingGuest}>Cancel</button>
          <button class="btn btn-primary" on:click={handleGuestSubmit} disabled={isSubmittingGuest || !guestEmail.trim()}>
            {isSubmittingGuest ? 'Connecting...' : 'Join Terminal'}
          </button>
        </div>
      </div>
    {:else if error}
      <div class="error-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="12" r="10"/>
          <line x1="15" y1="9" x2="9" y2="15"/>
          <line x1="9" y1="9" x2="15" y2="15"/>
        </svg>
        <p>{error}</p>
        <button class="btn btn-secondary" on:click={cancel}>Go Back</button>
      </div>
    {:else if sessionInfo}
      <div class="session-info">
        <div class="terminal-card">
          <div class="terminal-icon">⚡</div>
          <div class="terminal-details">
            <span class="terminal-name">{sessionInfo.containerName}</span>
            <span class="terminal-shared">Shared Terminal</span>
          </div>
        </div>
        
        <div class="info-row">
          <span class="label">Access</span>
          <span class="value mode-badge" class:control={sessionInfo.mode === 'control'}>
            {sessionInfo.mode === 'control' ? 'Full Control' : 'View Only'}
          </span>
        </div>
        <div class="info-row">
          <span class="label">Your Role</span>
          <span class="value role-badge">{sessionInfo.role}</span>
        </div>
        <div class="info-row">
          <span class="label">Expires</span>
          <span class="value">{new Date(sessionInfo.expiresAt).toLocaleTimeString()}</span>
        </div>
      </div>

      <div class="actions">
        <button class="btn btn-secondary" on:click={cancel}>Cancel</button>
        <button class="btn btn-primary" on:click={joinTerminal}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4 17 10 11 4 5"/>
            <line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
          Open Terminal
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  .join-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    padding: 20px;
  }

  .join-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    width: 100%;
    max-width: 400px;
    padding: 32px;
  }

  .join-header {
    text-align: center;
    margin-bottom: 24px;
  }

  .join-header svg {
    color: var(--accent);
    margin-bottom: 16px;
  }

  .join-header h1 {
    font-size: 20px;
    text-transform: uppercase;
    letter-spacing: 2px;
    margin: 0 0 12px;
  }

  .code-display {
    font-family: var(--font-mono);
    font-size: 28px;
    font-weight: 700;
    letter-spacing: 4px;
    color: var(--accent);
    margin: 0;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding: 32px 0;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 2px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading p {
    color: var(--text-muted);
    font-size: 14px;
  }

  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding: 24px 0;
    text-align: center;
  }

  .error-state svg {
    color: var(--error);
    opacity: 0.7;
  }

  .error-state p {
    color: var(--text-secondary);
    margin: 0;
  }

  .session-info {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 20px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    margin-bottom: 24px;
  }

  .terminal-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: rgba(0, 255, 65, 0.05);
    border: 1px solid rgba(0, 255, 65, 0.2);
    margin-bottom: 8px;
  }

  .terminal-icon {
    font-size: 24px;
  }

  .terminal-details {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .terminal-name {
    font-size: 16px;
    font-weight: 600;
    color: var(--accent);
    font-family: var(--font-mono);
  }

  .terminal-shared {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .label {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
  }

  .value {
    font-size: 14px;
    color: var(--text);
  }

  .mode-badge {
    padding: 4px 10px;
    background: var(--bg-card);
    border: 1px solid var(--border);
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .mode-badge.control {
    border-color: var(--accent);
    color: var(--accent);
  }

  .role-badge {
    padding: 4px 10px;
    background: var(--accent);
    color: #000;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .actions {
    display: flex;
    gap: 12px;
  }

  .actions .btn {
    flex: 1;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 20px;
    font-size: 14px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    border: 1px solid var(--border);
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-secondary {
    background: transparent;
    color: var(--text-secondary);
  }

  .btn-secondary:hover {
    background: var(--bg-secondary);
    color: var(--text);
  }

  .btn-primary {
    background: var(--accent);
    border-color: var(--accent);
    color: #000;
  }

  .btn-primary:hover {
    filter: brightness(1.1);
  }

  .auth-prompt {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .auth-description {
    color: var(--text-secondary);
    font-size: 14px;
    text-align: center;
    margin: 0;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .form-group label {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
  }

  .form-group input {
    width: 100%;
    padding: 12px 14px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    color: var(--text);
    font-family: var(--font-mono);
    font-size: 14px;
  }

  .form-group input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .form-group input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
