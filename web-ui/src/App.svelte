<script lang="ts">
  import { onMount } from 'svelte';
  import { auth, isAuthenticated } from '$stores/auth';
  import { containers } from '$stores/containers';
  import { terminal, hasSessions } from '$stores/terminal';

  // Components
  import Header from '$components/Header.svelte';
  import Landing from '$components/Landing.svelte';
  import Dashboard from '$components/Dashboard.svelte';
  import CreateTerminal from '$components/CreateTerminal.svelte';
  import TerminalView from '$components/terminal/TerminalView.svelte';
  import ToastContainer from '$components/ui/ToastContainer.svelte';

  // App state
  let currentView: 'landing' | 'dashboard' | 'create' = 'landing';
  let isLoading = true;

  // Handle OAuth callback
  async function handleOAuthCallback() {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');

    if (code) {
      const result = await auth.exchangeOAuthCode(code);
      if (result.success) {
        // Clear URL params
        window.history.replaceState({}, '', window.location.pathname);
        currentView = 'dashboard';
      }
    }
  }

  // Handle URL-based terminal routing
  async function handleTerminalUrl() {
    const path = window.location.pathname;
    const match = path.match(/^\/(?:terminal\/)?([a-f0-9]{64}|[a-f0-9-]{36})$/i);

    if (match && $isAuthenticated) {
      const containerId = match[1];
      // Fetch container info and connect
      const result = await containers.getContainer(containerId);
      if (result.success && result.container) {
        const sessionId = terminal.createSession(containerId, result.container.name);
        if (sessionId) {
          terminal.connectWebSocket(sessionId);
        }
      }
    }
  }

  // Initialize app
  onMount(async () => {
    // Check for OAuth callback
    await handleOAuthCallback();

    // Validate existing token
    if ($auth.token) {
      const isValid = await auth.validateToken();
      if (isValid) {
        await auth.fetchProfile();
        currentView = 'dashboard';
        await containers.fetchContainers();
        await handleTerminalUrl();
      } else {
        auth.logout();
      }
    }

    isLoading = false;
  });

  // React to auth changes
  $: if ($isAuthenticated && currentView === 'landing') {
    currentView = 'dashboard';
    containers.fetchContainers();
  }

  $: if (!$isAuthenticated && currentView !== 'landing') {
    currentView = 'landing';
    containers.reset();
    terminal.closeAllSessions();
  }

  // Navigation functions
  function goToDashboard() {
    currentView = 'dashboard';
    window.history.pushState({}, '', '/');
  }

  function goToCreate() {
    currentView = 'create';
  }

  function onContainerCreated(event: CustomEvent<{ id: string; name: string }>) {
    const { id, name } = event.detail;
    currentView = 'dashboard';

    // Connect to the new container
    const sessionId = terminal.createSession(id, name);
    if (sessionId) {
      terminal.connectWebSocket(sessionId);
    }
  }

  // Handle browser navigation
  function handlePopState() {
    const path = window.location.pathname;
    if (path === '/' || path === '') {
      currentView = $isAuthenticated ? 'dashboard' : 'landing';
    }
  }
</script>

<svelte:window on:popstate={handlePopState} />

<div class="app">
  {#if isLoading}
    <div class="loading-screen">
      <div class="spinner-large"></div>
      <p>Loading...</p>
    </div>
  {:else}
    <Header
      on:home={goToDashboard}
      on:create={goToCreate}
    />

    <main class="main">
      {#if currentView === 'landing'}
        <Landing />
      {:else if currentView === 'dashboard'}
        <Dashboard
          on:create={goToCreate}
          on:connect={(e) => {
            const sessionId = terminal.createSession(e.detail.id, e.detail.name);
            if (sessionId) {
              terminal.connectWebSocket(sessionId);
            }
          }}
        />
      {:else if currentView === 'create'}
        <CreateTerminal
          on:cancel={goToDashboard}
          on:created={onContainerCreated}
        />
      {/if}
    </main>

    <!-- Terminal overlay (floating or docked) -->
    {#if $hasSessions}
      <TerminalView />
    {/if}

    <!-- Toast notifications -->
    <ToastContainer />
  {/if}
</div>

<style>
  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  .main {
    flex: 1;
    max-width: 1400px;
    margin: 0 auto;
    padding: 20px;
    width: 100%;
  }

  .loading-screen {
    position: fixed;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    background: var(--bg);
  }

  .loading-screen p {
    color: var(--text-muted);
    font-size: 14px;
  }

  .spinner-large {
    width: 40px;
    height: 40px;
    border: 3px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
