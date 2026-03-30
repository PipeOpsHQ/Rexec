<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { admin } from "$stores/admin";
    import { formatRelativeTime, formatMemory, formatCPU } from "$utils/api";
    import PlatformIcon from "./icons/PlatformIcon.svelte";

    const statsRanges = [
        { value: "24h", label: "24H" },
        { value: "7d", label: "7D" },
        { value: "30d", label: "30D" },
        { value: "90d", label: "90D" },
        { value: "12m", label: "12M" }
    ] as const;

    // Tabs
    type Tab = "stats" | "users" | "subscribers" | "containers" | "terminals" | "agents";
    let activeTab: Tab = "stats";
    let selectedStatsRange = "30d";

    function setTab(tab: Tab) {
        activeTab = tab;
    }

    async function loadData() {
        await Promise.all([
            admin.fetchStats(selectedStatsRange),
            admin.fetchUsers(),
            admin.fetchContainers(),
            admin.fetchTerminals(),
            admin.fetchAgents()
        ]);
    }

    // Load initial data on mount and start WebSocket
    onMount(async () => {
        await loadData(); // Initial data fetch
        admin.startAdminEvents(); // Start WebSocket for live updates
    });

    // Clean up WebSocket on destroy
    onDestroy(() => {
        admin.stopAdminEvents();
    });

    $: users = $admin.users;
    $: subscribers = users.filter((u) => u.subscriptionActive === true);
    $: containers = $admin.containers;
    $: terminals = $admin.terminals;
    $: agents = $admin.agents;
    $: stats = $admin.stats;
    $: isLoading = $admin.isLoading;
    $: wsConnected = $admin.wsConnected;
    $: wsError = $admin.error;
    $: chartMax = Math.max(
        1,
        ...((stats?.timeline ?? []).map((point) => (
            point.newUsers +
            point.newContainers +
            point.newSessions +
            point.newLogins +
            point.newAgents +
            point.newRecordings
        )))
    );

    // Actions
    async function handleDeleteUser(userId: string) {
        if (confirm("Are you sure you want to delete this user? This action cannot be undone.")) {
            await admin.deleteUser(userId);
        }
    }

    async function handleDeleteContainer(containerId: string) {
        if (confirm("Are you sure you want to delete this container? This will stop and remove the container permanently.")) {
            await admin.deleteContainer(containerId);
        }
    }

    function getDistro(image: string): string {
         if (!image) return "linux";
         const lower = image.toLowerCase();
         if (lower.includes("ubuntu")) return "ubuntu";
         if (lower.includes("debian")) return "debian";
         if (lower.includes("alpine")) return "alpine";
         return "linux";
     }

    function getBarHeight(total: number): string {
        return `${Math.max(8, (total / chartMax) * 100)}%`;
    }

    function getSegmentHeight(value: number, total: number): string {
        if (value === 0 || total === 0) return "0%";
        return `${(value / total) * 100}%`;
    }

    async function handleStatsRangeChange(range: string) {
        selectedStatsRange = range;
        await admin.fetchStats(range);
    }

    function exportStatsCsv() {
        if (!stats) return;

        const rows = [
            ["period", "users", "containers", "sessions", "logins", "agents", "recordings"],
            ...stats.timeline.map((point) => [
                point.bucketLabel,
                point.newUsers,
                point.newContainers,
                point.newSessions,
                point.newLogins,
                point.newAgents,
                point.newRecordings
            ])
        ];

        const csv = rows
            .map((row) => row.map((value) => `"${String(value).replaceAll('"', '""')}"`).join(","))
            .join("\n");

        const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = `admin-stats-${selectedStatsRange}.csv`;
        link.click();
        URL.revokeObjectURL(url);
    }

</script>

<div class="dashboard">
    <div class="dashboard-header">
        <div class="dashboard-title">
            <h1>Admin Dashboard</h1>
            {#if !wsConnected}
                <span class="status-indicator error">Disconnected</span>
            {:else}
                <span class="status-indicator connected">Live</span>
            {/if}
        </div>
        <div class="dashboard-actions">
            {#if wsError}
                <div class="alert alert-error">
                    {wsError}
                </div>
            {/if}
            <button
                class="btn btn-secondary btn-sm"
                onclick={loadData}
                disabled={isLoading}
            >
                <svg
                    class="icon"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                >
                    <path
                        d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"
                    />
                </svg>
                Refresh
            </button>
        </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
        <button
            class="tab-btn"
            class:active={activeTab === "stats"}
            onclick={() => setTab("stats")}
        >
            Stats
        </button>
        <button
            class="tab-btn"
            class:active={activeTab === "users"}
            onclick={() => setTab("users")}
        >
            Users ({users.length})
        </button>
        <button
            class="tab-btn"
            class:active={activeTab === "subscribers"}
            onclick={() => setTab("subscribers")}
        >
            Subscribers ({subscribers.length})
        </button>
        <button
            class="tab-btn"
            class:active={activeTab === "containers"}
            onclick={() => setTab("containers")}
        >
            Containers ({containers.length})
        </button>
        <button
            class="tab-btn"
            class:active={activeTab === "terminals"}
            onclick={() => setTab("terminals")}
        >
            Active Terminals ({terminals.length})
        </button>
        <button
            class="tab-btn"
            class:active={activeTab === "agents"}
            onclick={() => setTab("agents")}
        >
            Agents ({agents.length})
        </button>
    </div>

    {#if isLoading}
        <div class="loading-state">
            <div class="spinner"></div>
            <p>Loading data...</p>
        </div>
    {:else}
        <div class="tab-content">
            {#if activeTab === "stats"}
                <div class="stats-panel">
                    <div class="stats-toolbar">
                        <div>
                            <h2>Usage overview</h2>
                            <p>Track signups, sessions, containers, and agents over time.</p>
                        </div>
                        <div class="stats-actions">
                            <div class="range-filter" role="group" aria-label="Stats time range">
                                {#each statsRanges as range}
                                    <button
                                        class="range-btn"
                                        class:active={selectedStatsRange === range.value}
                                        onclick={() => handleStatsRangeChange(range.value)}
                                    >
                                        {range.label}
                                    </button>
                                {/each}
                            </div>
                            <button class="btn btn-secondary btn-sm export-btn" onclick={exportStatsCsv}>
                                Export CSV
                            </button>
                        </div>
                    </div>

                    {#if stats}
                        <div class="metric-grid metric-grid-primary">
                            <article class="metric-card">
                                <span class="metric-label">Total users</span>
                                <strong>{stats.totals.users}</strong>
                                <span class="metric-note">{stats.activity.newUsers} new in range</span>
                            </article>
                            <article class="metric-card">
                                <span class="metric-label">Live terminals</span>
                                <strong>{stats.totals.activeSessions}</strong>
                                <span class="metric-note">{stats.activity.newSessions} sessions started</span>
                            </article>
                            <article class="metric-card">
                                <span class="metric-label">Logins</span>
                                <strong>{stats.totals.logins}</strong>
                                <span class="metric-note">{stats.activity.newLogins} in range</span>
                            </article>
                            <article class="metric-card">
                                <span class="metric-label">Active containers</span>
                                <strong>{stats.totals.containers}</strong>
                                <span class="metric-note">{stats.activity.newContainers} created in range</span>
                            </article>
                            <article class="metric-card">
                                <span class="metric-label">Agents online</span>
                                <strong>{stats.totals.onlineAgents} / {stats.totals.agents}</strong>
                                <span class="metric-note">{stats.activity.newAgents} new registrations</span>
                            </article>
                            <article class="metric-card accent-recordings">
                                <span class="metric-label">Recordings</span>
                                <strong>{stats.totals.recordings}</strong>
                                <span class="metric-note">{stats.activity.newRecordings} saved in range</span>
                            </article>
                            <article class="metric-card accent-recordings">
                                <span class="metric-label">Recorded hours</span>
                                <strong>{stats.totals.recordingHours}</strong>
                                <span class="metric-note">{stats.activity.recordingHours} hours in range</span>
                            </article>
                        </div>

                        <div class="chart-card">
                            <div class="chart-header">
                                <div>
                                    <h3>Usage over time</h3>
                                    <p>{new Date(stats.from).toLocaleDateString()} - {new Date(stats.to).toLocaleDateString()}</p>
                                </div>
                                <div class="chart-legend">
                                    <span class="legend-item users">Users</span>
                                    <span class="legend-item containers">Containers</span>
                                    <span class="legend-item sessions">Sessions</span>
                                    <span class="legend-item logins">Logins</span>
                                    <span class="legend-item agents">Agents</span>
                                    <span class="legend-item recordings">Recordings</span>
                                </div>
                            </div>

                            <div class="chart-wrap">
                                <div class="stacked-chart" aria-label="Usage timeline chart">
                                    {#each stats.timeline as point (point.bucketStart)}
                                        {@const total = point.newUsers + point.newContainers + point.newSessions + point.newLogins + point.newAgents + point.newRecordings}
                                        <div class="chart-bar-group" title={`${point.bucketLabel}: ${total} total events`}>
                                            <div class="chart-bar" style={`height: ${getBarHeight(total)}`}>
                                                <span class="bar-segment users" style={`height: ${getSegmentHeight(point.newUsers, total)}`}></span>
                                                <span class="bar-segment containers" style={`height: ${getSegmentHeight(point.newContainers, total)}`}></span>
                                                <span class="bar-segment sessions" style={`height: ${getSegmentHeight(point.newSessions, total)}`}></span>
                                                <span class="bar-segment logins" style={`height: ${getSegmentHeight(point.newLogins, total)}`}></span>
                                                <span class="bar-segment agents" style={`height: ${getSegmentHeight(point.newAgents, total)}`}></span>
                                                <span class="bar-segment recordings" style={`height: ${getSegmentHeight(point.newRecordings, total)}`}></span>
                                            </div>
                                        </div>
                                    {/each}
                                </div>
                            </div>

                            <div class="chart-axis">
                                {#each stats.timeline as point (point.bucketStart)}
                                    <span>{point.bucketLabel}</span>
                                {/each}
                            </div>
                        </div>

                        <div class="data-table-container stats-table-container">
                            <table class="data-table stats-table">
                                <thead>
                                    <tr>
                                        <th>Period</th>
                                        <th>Users</th>
                                        <th>Containers</th>
                                        <th>Sessions</th>
                                        <th>Logins</th>
                                        <th>Agents</th>
                                        <th>Recordings</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each stats.timeline as point (point.bucketStart)}
                                        <tr>
                                            <td>{point.bucketLabel}</td>
                                            <td>{point.newUsers}</td>
                                            <td>{point.newContainers}</td>
                                            <td>{point.newSessions}</td>
                                            <td>{point.newLogins}</td>
                                            <td>{point.newAgents}</td>
                                            <td>{point.newRecordings}</td>
                                        </tr>
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                    {/if}
                </div>
            {:else if activeTab === "users"}
                <div class="data-table-container">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>User</th>
                                <th>Email</th>
                                <th>Role</th>
                                <th>Tier</th>
                                <th>Containers</th>
                                <th>Created</th>
                                <th>Last Login</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each users as user (user.id)}
                                <tr>
                                    <td>
                                        <div class="user-info">
                                            <div class="avatar">{user.username.charAt(0).toUpperCase()}</div>
                                            <span>{user.username}</span>
                                        </div>
                                    </td>
                                    <td>{user.email}</td>
                                    <td>
                                        {#if user.isAdmin}
                                            <span class="badge admin">Admin</span>
                                        {:else}
                                            <span class="badge user">User</span>
                                        {/if}
                                    </td>
                                    <td><span class="badge tier-{user.tier}">{user.tier}</span></td>
                                    <td>{user.containerCount}</td>
                                    <td>{new Date(user.created_at).toLocaleDateString()}</td>
                                    <td>{user.updated_at ? formatRelativeTime(user.updated_at) : '-'}</td>
                                    <td>
                                        <button class="btn-icon danger" onclick={() => handleDeleteUser(user.id)} title="Delete User">
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {:else if activeTab === "subscribers"}
                <div class="data-table-container">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>User</th>
                                <th>Email</th>
                                <th>Tier</th>
                                <th>Containers</th>
                                <th>Created</th>
                                <th>Last Login</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each subscribers as user (user.id)}
                                <tr>
                                    <td>
                                        <div class="user-info">
                                            <div class="avatar">{user.username.charAt(0).toUpperCase()}</div>
                                            <span>{user.username}</span>
                                        </div>
                                    </td>
                                    <td>{user.email}</td>
                                    <td><span class="badge tier-{user.tier}">{user.tier}</span></td>
                                    <td>{user.containerCount}</td>
                                    <td>{new Date(user.created_at).toLocaleDateString()}</td>
                                    <td>{user.updated_at ? formatRelativeTime(user.updated_at) : "-"}</td>
                                    <td>
                                        <button
                                            class="btn-icon danger"
                                            onclick={() => handleDeleteUser(user.id)}
                                            title="Delete User"
                                        >
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {:else if activeTab === "containers"}
                <div class="data-table-container">
                    <table class="data-table">
                         <thead>
                            <tr>
                                <th>Name</th>
                                <th>User</th>
                                <th>Image</th>
                                <th>Status</th>
                                <th>Resources</th>
                                <th>Created</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each containers as container (container.id)}
                                <tr>
                                    <td>
                                        <div class="container-name-cell">
                                            <PlatformIcon platform={getDistro(container.image)} size={16} />
                                            <span>{container.name}</span>
                                        </div>
                                    </td>
                                    <td>
                                        <div class="user-cell">
                                            <span class="user-name">{container.username}</span>
                                            <span class="user-email">{container.userEmail}</span>
                                        </div>
                                    </td>
                                    <td class="mono">{container.image}</td>
                                    <td>
                                        <span class="status-badge {container.status}">
                                            {container.status}
                                        </span>
                                    </td>
                                    <td>
                                        {#if container.resources}
                                            <div class="resources-cell">
                                                <span>{formatMemory(container.resources.memory_mb)}</span>
                                                <span>/</span>
                                                <span>{formatCPU(container.resources.cpu_shares)}</span>
                                            </div>
                                        {:else}
                                            -
                                        {/if}
                                    </td>
                                    <td>{formatRelativeTime(container.created_at)}</td>
                                    <td>
                                        <button class="btn-icon danger" onclick={() => handleDeleteContainer(container.id)} title="Delete Container">
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {:else if activeTab === "terminals"}
                <div class="data-table-container">
                     <table class="data-table">
                        <thead>
                            <tr>
                                <th>Terminal ID</th>
                                <th>User</th>
                                <th>Container</th>
                                <th>Status</th>
                                <th>Connected At</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each terminals as term (term.id)}
                                <tr>
                                    <td class="mono">{term.id}</td>
                                    <td>{term.username}</td>
                                    <td>{term.name}</td>
                                    <td>
                                        <span class="status-badge {term.status}">
                                            {term.status}
                                        </span>
                                    </td>
                                    <td>{formatRelativeTime(term.connectedAt)}</td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {:else if activeTab === "agents"}
                <div class="data-table-container">
                     <table class="data-table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>User</th>
                                <th>Status</th>
                                <th>Platform</th>
                                <th>Specs</th>
                                <th>Last Seen</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each agents as agent (agent.id)}
                                <tr>
                                    <td>
                                        <div class="container-name-cell">
                                            <PlatformIcon platform={agent.distro || agent.os} size={16} />
                                            <span>{agent.name}</span>
                                        </div>
                                    </td>
                                    <td>{agent.username || agent.user_id.slice(0, 8)}</td>
                                    <td>
                                        <span class="status-badge {agent.status}">
                                            {agent.status}
                                        </span>
                                    </td>
                                    <td class="mono">{agent.distro || agent.os}/{agent.arch}</td>
                                    <td>
                                        {#if agent.system_info && agent.system_info.memory && agent.system_info.num_cpu}
                                            <div class="resources-cell">
                                                <span>{formatMemory(Math.round((agent.system_info.memory.total || 0) / 1024 / 1024))}</span>
                                                <span>/</span>
                                                <span>{agent.system_info.num_cpu} CPU</span>
                                            </div>
                                        {:else}
                                            -
                                        {/if}
                                    </td>
                                    <td>{agent.last_ping ? formatRelativeTime(agent.last_ping) : '-'}</td>
                                    <td>
                                        <button class="btn-icon danger" onclick={() => handleDeleteUser(agent.id)} title="Delete Agent (TODO)">
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {/if}
        </div>
    {/if}
</div>

<style>
    .dashboard {
        animation: fadeIn 0.2s ease;
        padding: 20px;
    }

    .dashboard-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 24px;
        padding-bottom: 16px;
        border-bottom: 1px solid var(--border);
    }

    .dashboard-title {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .dashboard-title h1 {
        font-size: 20px;
        text-transform: uppercase;
        letter-spacing: 1px;
        margin: 0;
    }

    .status-indicator {
        font-size: 11px;
        font-weight: 600;
        text-transform: uppercase;
        padding: 4px 8px;
        border-radius: 12px;
    }

    .status-indicator.connected {
        background: rgba(0, 255, 65, 0.1);
        color: var(--green);
        border: 1px solid rgba(0, 255, 65, 0.2);
    }

    .status-indicator.error, .status-indicator.disconnected {
        background: rgba(255, 0, 0, 0.1);
        color: var(--red);
        border: 1px solid rgba(255, 0, 0, 0.2);
    }

    .alert.alert-error {
        background: rgba(255, 0, 0, 0.1);
        color: var(--red);
        padding: 8px 16px;
        border-radius: 4px;
        border: 1px solid rgba(255, 0, 0, 0.2);
        margin-right: 16px;
        font-size: 13px;
    }


    /* Tabs */
    .tabs {
        display: flex;
        gap: 16px;
        margin-bottom: 24px;
        border-bottom: 1px solid var(--border);
        overflow-x: auto;
        scrollbar-width: none;
        -ms-overflow-style: none;
        -webkit-overflow-scrolling: touch;
    }

    .tabs::-webkit-scrollbar {
        display: none;
    }

    .tab-btn {
        background: none;
        border: none;
        padding: 12px 16px;
        color: var(--text-muted);
        font-size: 14px;
        cursor: pointer;
        position: relative;
        white-space: nowrap;
    }

    .tab-btn:hover {
        color: var(--text);
    }

    .tab-btn.active {
        color: var(--accent);
    }

    .tab-btn.active::after {
        content: "";
        position: absolute;
        bottom: -1px;
        left: 0;
        width: 100%;
        height: 2px;
        background: var(--accent);
    }

    .stats-panel {
        display: grid;
        gap: 20px;
    }

    .stats-toolbar {
        display: flex;
        justify-content: space-between;
        gap: 16px;
        align-items: flex-start;
    }

    .stats-actions {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        justify-content: flex-end;
        align-items: center;
    }

    .stats-toolbar h2,
    .chart-header h3 {
        margin: 0;
        font-size: 18px;
    }

    .stats-toolbar p,
    .chart-header p {
        margin: 6px 0 0;
        color: var(--text-muted);
        font-size: 13px;
    }

    .range-filter {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .range-btn {
        border: 1px solid var(--border);
        background: var(--bg-secondary);
        color: var(--text-muted);
        padding: 8px 12px;
        border-radius: 999px;
        font-size: 12px;
        font-weight: 600;
        cursor: pointer;
    }

    .range-btn:hover,
    .range-btn.active {
        color: var(--text);
        border-color: var(--accent);
        box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 45%, transparent);
    }

    .metric-grid {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 16px;
    }

    .metric-card,
    .chart-card {
        border: 1px solid var(--border);
        border-radius: 12px;
        background: linear-gradient(180deg, color-mix(in srgb, var(--bg-secondary) 92%, white 8%), var(--bg-secondary));
    }

    .metric-card {
        padding: 18px;
        display: grid;
        gap: 8px;
    }

    .metric-label,
    .metric-note {
        font-size: 12px;
    }

    .metric-label {
        text-transform: uppercase;
        color: var(--text-muted);
        letter-spacing: 0.08em;
    }

    .metric-card strong {
        font-size: 28px;
        line-height: 1;
    }

    .metric-note {
        color: var(--text-secondary);
    }

    .accent-recordings strong {
        color: #ff7a59;
    }

    .chart-card {
        padding: 20px;
    }

    .chart-header {
        display: flex;
        justify-content: space-between;
        gap: 16px;
        align-items: flex-start;
        margin-bottom: 20px;
    }

    .chart-legend {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        justify-content: flex-end;
    }

    .legend-item {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        color: var(--text-muted);
    }

    .legend-item::before {
        content: "";
        width: 10px;
        height: 10px;
        border-radius: 50%;
        background: currentColor;
    }

    .legend-item.users,
    .bar-segment.users {
        color: #4aa3ff;
        background: #4aa3ff;
    }

    .legend-item.containers,
    .bar-segment.containers {
        color: #f8b84e;
        background: #f8b84e;
    }

    .legend-item.sessions,
    .bar-segment.sessions {
        color: #3ddc97;
        background: #3ddc97;
    }

    .legend-item.logins,
    .bar-segment.logins {
        color: #b38cff;
        background: #b38cff;
    }

    .legend-item.agents,
    .bar-segment.agents {
        color: #ff7a59;
        background: #ff7a59;
    }

    .legend-item.recordings,
    .bar-segment.recordings {
        color: #56d4c1;
        background: #56d4c1;
    }

    .chart-wrap {
        height: 240px;
        padding: 16px 0 8px;
        border-top: 1px solid var(--border);
        border-bottom: 1px solid var(--border);
        background:
            linear-gradient(to top, transparent 24%, color-mix(in srgb, var(--border) 70%, transparent) 25%, transparent 26%),
            linear-gradient(to top, transparent 49%, color-mix(in srgb, var(--border) 70%, transparent) 50%, transparent 51%),
            linear-gradient(to top, transparent 74%, color-mix(in srgb, var(--border) 70%, transparent) 75%, transparent 76%);
    }

    .stacked-chart {
        width: 100%;
        height: 100%;
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(12px, 1fr));
        gap: 8px;
        align-items: end;
    }

    .chart-bar-group {
        min-width: 0;
        height: 100%;
        display: flex;
        align-items: end;
    }

    .chart-bar {
        width: 100%;
        min-height: 8px;
        display: flex;
        flex-direction: column-reverse;
        border-radius: 8px 8px 4px 4px;
        overflow: hidden;
        background: color-mix(in srgb, var(--bg-tertiary) 88%, transparent);
        box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--border) 75%, transparent);
    }

    .bar-segment {
        width: 100%;
    }

    .chart-axis {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(12px, 1fr));
        gap: 8px;
        margin-top: 12px;
        color: var(--text-muted);
        font-size: 11px;
    }

    .chart-axis span {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .stats-table {
        min-width: 760px;
    }

    /* Table */
    .data-table-container {
        overflow-x: auto;
        border: 1px solid var(--border);
        border-radius: 4px;
        width: 100%;
        -webkit-overflow-scrolling: touch;
    }

    .data-table {
        width: 100%;
        min-width: 900px;
        border-collapse: collapse;
        font-size: 14px;
    }

    .data-table th, .data-table td {
        padding: 12px 16px;
        text-align: left;
        border-bottom: 1px solid var(--border);
        white-space: nowrap;
    }

    .data-table th {
        background: var(--bg-secondary);
        color: var(--text-muted);
        text-transform: uppercase;
        font-size: 11px;
        font-weight: 600;
        position: sticky;
        top: 0;
        z-index: 1;
    }

    .data-table tr:last-child td {
        border-bottom: none;
    }

    .data-table tr:hover {
        background: var(--bg-secondary);
    }
    
    .mono {
        font-family: var(--font-mono);
        font-size: 12px;
    }

    /* User Info */
    .user-info {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .avatar {
        width: 24px;
        height: 24px;
        background: var(--accent);
        color: var(--bg);
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: bold;
        font-size: 12px;
    }

    .user-cell {
        display: flex;
        flex-direction: column;
    }
    
    .user-email {
        font-size: 11px;
        color: var(--text-muted);
    }

    /* Badges */
    .badge {
        display: inline-block;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 10px;
        text-transform: uppercase;
        font-weight: bold;
    }

    .badge.admin {
        background: var(--accent);
        color: var(--bg);
    }

    .badge.user {
        background: var(--bg-tertiary);
        color: var(--text-muted);
        border: 1px solid var(--border);
    }
    
    .badge.tier-free {
        background: var(--bg-tertiary);
        color: var(--text);
    }
    
    .badge.tier-pro {
        background: #ffd700;
        color: #000;
    }
    
    .badge.tier-guest {
        background: var(--bg-tertiary);
        color: var(--text-muted);
    }

    .status-badge {
        display: inline-flex;
        align-items: center;
        padding: 2px 8px;
        border-radius: 12px;
        font-size: 11px;
        text-transform: uppercase;
        font-weight: 600;
    }
    
    .status-badge.running, .status-badge.connected {
        background: rgba(0, 255, 65, 0.1);
        color: var(--green);
        border: 1px solid rgba(0, 255, 65, 0.2);
    }
    
    .status-badge.stopped, .status-badge.disconnected {
        background: var(--bg-tertiary);
        color: var(--text-muted);
        border: 1px solid var(--border);
    }

    /* Container Info */
    .container-name-cell {
        display: flex;
        align-items: center;
        gap: 8px;
    }
    
    .resources-cell {
        display: flex;
        gap: 4px;
        font-family: var(--font-mono);
        font-size: 11px;
        color: var(--text-secondary);
    }

    /* Buttons */
    .btn-icon {
        background: none;
        border: none;
        padding: 4px;
        cursor: pointer;
        color: var(--text-muted);
        border-radius: 4px;
    }
    
    .btn-icon:hover {
        background: var(--bg-tertiary);
        color: var(--text);
    }
    
    .btn-icon.danger:hover {
        color: var(--red);
        background: rgba(255, 0, 0, 0.1);
    }
    
    .btn-icon svg {
        width: 16px;
        height: 16px;
    }

    /* Loading */
    .loading-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 60px 20px;
        gap: 16px;
        color: var(--text-muted);
    }
    
    .spinner {
        width: 24px;
        height: 24px;
        border: 2px solid var(--border);
        border-top-color: var(--accent);
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }
    
    @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
    }

    @media (max-width: 768px) {
        .dashboard {
            padding: 12px;
        }

        .stats-toolbar,
        .chart-header,
        .stats-actions {
            flex-direction: column;
            align-items: stretch;
        }

        .metric-grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }

        .chart-wrap {
            height: 200px;
        }

        .dashboard-header {
            flex-direction: column;
            align-items: flex-start;
            gap: 16px;
        }

        .dashboard-actions {
            width: 100%;
            display: flex;
            justify-content: flex-start;
            flex-wrap: wrap;
        }

        .alert.alert-error {
            margin-right: 0;
            margin-bottom: 8px;
            width: 100%;
        }

        .data-table {
            font-size: 13px;
        }

        .data-table th, .data-table td {
            padding: 10px 12px;
        }
    }

    @media (max-width: 520px) {
        .metric-grid {
            grid-template-columns: 1fr;
        }
    }
</style>
