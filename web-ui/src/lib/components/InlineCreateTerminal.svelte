<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { containers, type ProgressEvent } from "$stores/containers";
    import { api } from "$utils/api";
    import PlatformIcon from "./icons/PlatformIcon.svelte";

    export let compact = false;

    const dispatch = createEventDispatcher<{
        created: { id: string; name: string };
        cancel: void;
    }>();

    let selectedImage = "";
    let isCreating = false;
    let selectedRole = "standard";
    let progress = 0;
    let progressMessage = "";
    let progressStage = "";
    
    // Resource customization
    let showResources = false;
    let memoryMB = 512;
    let cpuShares = 512;
    let diskMB = 2048;
    
    // Trial limits - allow more memory than CPU
    const resourceLimits = {
        minMemory: 256,
        maxMemory: 2048,  // Allow up to 2GB for trial
        minCPU: 256,
        maxCPU: 1024,
        minDisk: 1024,
        maxDisk: 8192
    };

    // Slider event handlers
    function handleMemoryChange(e: Event) {
        memoryMB = parseInt((e.target as HTMLInputElement).value);
    }

    function handleCpuChange(e: Event) {
        cpuShares = parseInt((e.target as HTMLInputElement).value);
    }

    function handleDiskChange(e: Event) {
        diskMB = parseInt((e.target as HTMLInputElement).value);
    }

    const progressSteps = [
        { id: "validating", label: "Validating" },
        { id: "pulling", label: "Pulling Image" },
        { id: "creating", label: "Creating Container" },
        { id: "starting", label: "Starting" },
        { id: "configuring", label: "Configuring" },
        { id: "ready", label: "Ready" },
    ];

    // Hacker-style log messages
    let logMessages: Array<{ text: string; type: 'info' | 'success' | 'cmd' | 'data' }> = [];
    let logContainer: HTMLDivElement;
    
    const stageLogMessages: Record<string, Array<{ text: string; type: 'info' | 'success' | 'cmd' | 'data' }>> = {
        validating: [
            { text: '$ rexec init --validate', type: 'cmd' },
            { text: '[SYS] Authenticating session...', type: 'info' },
            { text: '[AUTH] Token verified ✓', type: 'success' },
            { text: '[QUOTA] Checking resource allocation...', type: 'info' },
        ],
        pulling: [
            { text: '$ docker pull registry.rexec.io/base', type: 'cmd' },
            { text: '[NET] Connecting to registry...', type: 'info' },
            { text: '[PULL] Downloading layers...', type: 'data' },
            { text: '[CACHE] Layer sha256:a3ed... cached', type: 'info' },
            { text: '[PULL] Extracting filesystem...', type: 'data' },
        ],
        creating: [
            { text: '$ rexec container create --secure', type: 'cmd' },
            { text: '[DOCKER] Allocating container ID...', type: 'info' },
            { text: '[NET] Configuring network namespace...', type: 'info' },
            { text: '[FS] Mounting overlay filesystem...', type: 'data' },
            { text: '[SEC] Applying seccomp profile...', type: 'info' },
        ],
        starting: [
            { text: '$ rexec container start', type: 'cmd' },
            { text: '[INIT] Starting container process...', type: 'info' },
            { text: '[PID] Process spawned: 1', type: 'data' },
            { text: '[TTY] Allocating pseudo-terminal...', type: 'info' },
            { text: '[WS] WebSocket channel ready', type: 'success' },
        ],
        configuring: [
            { text: '$ rexec setup --role ${role}', type: 'cmd' },
            { text: '[PKG] Updating package index...', type: 'info' },
            { text: '[INSTALL] Installing development tools...', type: 'data' },
            { text: '[CONFIG] Writing shell configuration...', type: 'info' },
            { text: '[ENV] Setting environment variables...', type: 'data' },
        ],
        ready: [
            { text: '[SYS] Container ready ✓', type: 'success' },
            { text: '[WS] Terminal connection established', type: 'success' },
            { text: '$ echo "Welcome to Rexec"', type: 'cmd' },
        ],
    };
    
    let prevStage = '';
    $: if (progressStage && progressStage !== prevStage) {
        prevStage = progressStage;
        const newLogs = stageLogMessages[progressStage] || [];
        // Add logs with slight delay for effect
        newLogs.forEach((log, i) => {
            setTimeout(() => {
                logMessages = [...logMessages, log];
                // Auto-scroll to bottom
                if (logContainer) {
                    logContainer.scrollTop = logContainer.scrollHeight;
                }
            }, i * 150);
        });
    }

    function getStepStatus(stepId: string): "pending" | "active" | "completed" {
        const stepOrder = progressSteps.map((s) => s.id);
        const currentIndex = stepOrder.indexOf(progressStage);
        const stepIndex = stepOrder.indexOf(stepId);

        if (stepIndex < currentIndex) return "completed";
        if (stepIndex === currentIndex) return "active";
        return "pending";
    }

    $: displayProgress = Math.round(progress);

    let images: Array<{
        name: string;
        display_name: string;
        description: string;
        category: string;
        popular?: boolean;
    }> = [];

    const roleToOS: Record<string, string> = {
        standard: "alpine",
        node: "ubuntu",
        python: "ubuntu",
        go: "alpine",
        neovim: "arch",
        devops: "alpine",
        overemployed: "alpine",
    };

    $: if (selectedRole && roleToOS[selectedRole]) {
        const preferredOS = roleToOS[selectedRole];
        if (images.length > 0 && images.some((img) => img.name === preferredOS)) {
            selectedImage = preferredOS;
        }
    }

    const roles = [
        {
            id: "standard",
            name: "The Minimalist",
            desc: "I use Arch btw. Just give me a shell.",
            tools: ["bash", "git", "curl", "vim"],
            recommendedOS: "Alpine",
        },
        {
            id: "node",
            name: "10x JS Ninja",
            desc: "Ship fast, break things, npm install everything.",
            tools: ["node", "npm", "yarn", "pnpm", "git"],
            recommendedOS: "Ubuntu",
        },
        {
            id: "python",
            name: "Data Wizard",
            desc: "Import antigravity. I speak in list comprehensions.",
            tools: ["python3", "pip", "jupyter", "pandas", "numpy"],
            recommendedOS: "Ubuntu",
        },
        {
            id: "go",
            name: "The Gopher",
            desc: "If err != nil { panic(err) }. Simplicity is key.",
            tools: ["go", "git", "make", "delve"],
            recommendedOS: "Alpine",
        },
        {
            id: "neovim",
            name: "Neovim God",
            desc: "My config is longer than your code. Mouse? What mouse?",
            tools: ["neovim", "tmux", "fzf", "ripgrep", "lazygit"],
            recommendedOS: "Arch",
        },
        {
            id: "devops",
            name: "YAML Herder",
            desc: "I don't write code, I write config. Prod is my playground.",
            tools: ["kubectl", "docker", "terraform", "helm", "aws-cli"],
            recommendedOS: "Alpine",
        },
        {
            id: "overemployed",
            name: "The Overemployed",
            desc: "Working 4 remote jobs. Need max efficiency.",
            tools: ["tmux", "git", "ssh", "docker", "zsh"],
            recommendedOS: "Alpine",
        },
    ];

    $: currentRole = roles.find((r) => r.id === selectedRole);

    async function loadImages() {
        if (images.length > 0) return;

        const { data, error } = await api.get<{
            images?: typeof images;
            popular?: typeof images;
        }>("/api/images?all=true");

        if (data) {
            images = data.images || data.popular || [];
        }
    }

    onMount(() => {
        loadImages();
    });

    function selectAndCreate(imageName: string) {
        selectedImage = imageName;
        createContainer();
    }

    async function createContainer() {
        if (!selectedImage || isCreating) return;

        isCreating = true;
        progress = 0;
        progressMessage = "Starting...";
        progressStage = "validating";

        // Generate a unique name for the container
        const containerName = `${selectedImage}-${Date.now().toString(36)}`;

        function handleProgress(event: ProgressEvent) {
            progress = event.progress;
            progressMessage = event.message;
            progressStage = event.stage;
        }

        function handleComplete(container: { id: string; name: string }) {
            dispatch("created", { id: container.id, name: container.name });
            isCreating = false;
            progress = 0;
            progressMessage = "";
            progressStage = "";
            logMessages = [];
            prevStage = '';
        }

        function handleError(error: string) {
            logMessages = [...logMessages, { text: `[ERROR] ${error}`, type: 'info' as const }];
            progressMessage = error || "Failed to create terminal";
            setTimeout(() => {
                isCreating = false;
                progress = 0;
                progressMessage = "";
                progressStage = "";
                logMessages = [];
                prevStage = '';
            }, 3000);
        }

        // Call with correct parameters: name, image, customImage, role, onProgress, onComplete, onError, resources
        containers.createContainerWithProgress(
            containerName,
            selectedImage,
            undefined,  // customImage
            selectedRole,
            handleProgress,
            handleComplete,
            handleError,
            { memory_mb: memoryMB, cpu_shares: cpuShares, disk_mb: diskMB }
        );
    }
</script>

<div class="inline-create" class:compact>
    {#if isCreating}
        <div class="create-progress">
            <!-- Hacker-style terminal header -->
            <div class="hacker-header">
                <div class="terminal-title">
                    <span class="blink">▌</span> REXEC INIT SEQUENCE
                </div>
                <div class="progress-stats">
                    <span class="stat">[{displayProgress}%]</span>
                    <span class="stage">{progressStage.toUpperCase()}</span>
                </div>
            </div>
            
            <!-- Progress bar styled as loading bar -->
            <div class="hacker-progress">
                <div class="progress-track">
                    {#each Array(20) as _, i}
                        <span class="progress-block" class:filled={i < displayProgress / 5}>█</span>
                    {/each}
                </div>
            </div>
            
            <!-- Hacker log display -->
            <div class="hacker-logs" bind:this={logContainer}>
                {#each logMessages as log, i}
                    <div class="log-line" class:cmd={log.type === 'cmd'} class:success={log.type === 'success'} class:data={log.type === 'data'}>
                        <span class="log-prefix">{log.type === 'cmd' ? '' : '>'}</span>
                        <span class="log-text">{log.text}</span>
                        {#if i === logMessages.length - 1}
                            <span class="cursor">_</span>
                        {/if}
                    </div>
                {/each}
            </div>
            
            <!-- Current stage indicator -->
            <div class="stage-indicator">
                {#each progressSteps as step}
                    <div class="step" class:active={progressStage === step.id} class:completed={getStepStatus(step.id) === 'completed'}>
                        <span class="step-dot">{getStepStatus(step.id) === 'completed' ? '●' : getStepStatus(step.id) === 'active' ? '◉' : '○'}</span>
                    </div>
                {/each}
            </div>
        </div>
    {:else}
        <div class="create-content">
            <!-- Role Selection -->
            <div class="create-section">
                <h4>Environment</h4>
                <div class="role-grid">
                    {#each roles as role}
                        <button
                            class="role-card"
                            class:selected={selectedRole === role.id}
                            on:click={() => (selectedRole = role.id)}
                            title={role.desc}
                        >
                            <PlatformIcon platform={role.id} size={28} />
                            <span class="role-name">{role.name}</span>
                        </button>
                    {/each}
                </div>
                {#if currentRole}
                    <div class="role-info">
                        <div class="role-header-row">
                            <PlatformIcon platform={currentRole.id} size={18} />
                            <span class="role-name-sm">{currentRole.name}</span>
                            <span class="role-os-badge">
                                <PlatformIcon platform={currentRole.recommendedOS.toLowerCase()} size={14} />
                                {currentRole.recommendedOS}
                            </span>
                        </div>
                        <div class="role-tools">
                            {#each currentRole.tools as tool}
                                <span class="tool-badge">{tool}</span>
                            {/each}
                        </div>
                    </div>
                {/if}
            </div>

            <!-- Resource Configuration (Trial users can customize) -->
            <div class="create-section">
                <button 
                    class="resource-toggle"
                    on:click={() => showResources = !showResources}
                >
                    <span class="toggle-icon">{showResources ? '▼' : '▶'}</span>
                    <h4>Resources</h4>
                    <span class="resource-preview">
                        {memoryMB}MB / {cpuShares} CPU / {diskMB}MB
                    </span>
                </button>
                
                {#if showResources}
                    <div class="resource-config">
                        <div class="resource-row">
                            <label>
                                <span class="resource-label">Memory</span>
                                <span class="resource-value">{memoryMB} MB</span>
                            </label>
                            <input 
                                type="range" 
                                value={memoryMB}
                                on:input={handleMemoryChange}
                                min={resourceLimits.minMemory}
                                max={resourceLimits.maxMemory}
                                step="128"
                            />
                            <div class="resource-range">
                                <span>{resourceLimits.minMemory}MB</span>
                                <span>{resourceLimits.maxMemory}MB</span>
                            </div>
                        </div>
                        
                        <div class="resource-row">
                            <label>
                                <span class="resource-label">CPU</span>
                                <span class="resource-value">{cpuShares} shares</span>
                            </label>
                            <input 
                                type="range" 
                                value={cpuShares}
                                on:input={handleCpuChange}
                                min={resourceLimits.minCPU}
                                max={resourceLimits.maxCPU}
                                step="128"
                            />
                            <div class="resource-range">
                                <span>{resourceLimits.minCPU}</span>
                                <span>{resourceLimits.maxCPU}</span>
                            </div>
                        </div>
                        
                        <div class="resource-row">
                            <label>
                                <span class="resource-label">Disk</span>
                                <span class="resource-value">{diskMB} MB</span>
                            </label>
                            <input 
                                type="range" 
                                value={diskMB}
                                on:input={handleDiskChange}
                                min={resourceLimits.minDisk}
                                max={resourceLimits.maxDisk}
                                step="256"
                            />
                            <div class="resource-range">
                                <span>{resourceLimits.minDisk}MB</span>
                                <span>{resourceLimits.maxDisk}MB</span>
                            </div>
                        </div>
                        
                        <p class="resource-hint">
                            Trial users can customize resources within these limits
                        </p>
                    </div>
                {/if}
            </div>

            <!-- OS Selection -->
            <div class="create-section">
                <h4>Operating System</h4>
                <div class="os-grid">
                    {#each images as image (image.name)}
                        <button
                            class="os-card"
                            on:click={() => selectAndCreate(image.name)}
                        >
                            <PlatformIcon platform={image.name} size={28} />
                            <span class="os-name">{image.display_name || image.name}</span>
                            {#if image.popular}
                                <span class="popular-badge">Popular</span>
                            {/if}
                        </button>
                    {/each}
                    <button
                        class="os-card"
                        on:click={() => selectAndCreate("custom")}
                    >
                        <PlatformIcon platform="custom" size={28} />
                        <span class="os-name">Custom</span>
                    </button>
                </div>
            </div>
        </div>
    {/if}
</div>

<style>
    .inline-create {
        padding: 16px;
        height: 100%;
        overflow-y: auto;
        background: #0a0a0a;
    }

    .inline-create.compact {
        padding: 12px;
    }

    /* Progress - Hacker Style */
    .create-progress {
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding: 16px;
        background: #000;
        border: 1px solid #00ff41;
        border-radius: 4px;
        font-family: var(--font-mono);
        box-shadow: 0 0 20px rgba(0, 255, 65, 0.1), inset 0 0 40px rgba(0, 255, 65, 0.02);
    }

    .hacker-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(0, 255, 65, 0.3);
    }

    .terminal-title {
        font-size: 12px;
        color: #00ff41;
        font-weight: 600;
        letter-spacing: 2px;
        text-shadow: 0 0 10px rgba(0, 255, 65, 0.5);
    }

    .blink {
        animation: blink-cursor 1s step-end infinite;
    }

    @keyframes blink-cursor {
        0%, 100% { opacity: 1; }
        50% { opacity: 0; }
    }

    .progress-stats {
        display: flex;
        gap: 12px;
        font-size: 11px;
    }

    .progress-stats .stat {
        color: #00ff41;
        font-weight: 600;
    }

    .progress-stats .stage {
        color: #0af;
        text-transform: uppercase;
        letter-spacing: 1px;
    }

    .hacker-progress {
        padding: 4px 0;
    }

    .progress-track {
        display: flex;
        font-size: 10px;
        letter-spacing: 1px;
    }

    .progress-block {
        color: #333;
        transition: color 0.15s ease;
    }

    .progress-block.filled {
        color: #00ff41;
        text-shadow: 0 0 5px rgba(0, 255, 65, 0.8);
    }

    .hacker-logs {
        height: 180px;
        overflow-y: auto;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid #222;
        border-radius: 2px;
        padding: 8px;
        scrollbar-width: thin;
        scrollbar-color: #00ff41 #111;
    }

    .hacker-logs::-webkit-scrollbar {
        width: 4px;
    }

    .hacker-logs::-webkit-scrollbar-track {
        background: #111;
    }

    .hacker-logs::-webkit-scrollbar-thumb {
        background: #00ff41;
        border-radius: 2px;
    }

    .log-line {
        display: flex;
        gap: 6px;
        font-size: 11px;
        line-height: 1.6;
        color: #888;
        animation: fadeIn 0.2s ease;
    }

    @keyframes fadeIn {
        from { opacity: 0; transform: translateX(-5px); }
        to { opacity: 1; transform: translateX(0); }
    }

    .log-line.cmd {
        color: #fff;
        font-weight: 500;
    }

    .log-line.success {
        color: #00ff41;
    }

    .log-line.data {
        color: #0af;
    }

    .log-prefix {
        color: #555;
        user-select: none;
    }

    .log-line.cmd .log-prefix {
        color: #00ff41;
    }

    .cursor {
        animation: blink-cursor 0.7s step-end infinite;
        color: #00ff41;
    }

    .stage-indicator {
        display: flex;
        justify-content: center;
        gap: 8px;
        padding-top: 8px;
        border-top: 1px solid rgba(0, 255, 65, 0.2);
    }

    .step {
        transition: all 0.2s ease;
    }

    .step-dot {
        font-size: 8px;
        color: #333;
    }

    .step.active .step-dot {
        color: #00ff41;
        text-shadow: 0 0 8px rgba(0, 255, 65, 0.8);
        animation: pulse-dot 1s ease infinite;
    }

    .step.completed .step-dot {
        color: #00ff41;
    }

    @keyframes pulse-dot {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
    }

    /* Content */
    .create-content {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .create-section h4 {
        margin: 0 0 12px 0;
        font-size: 13px;
        font-weight: 600;
        color: var(--accent);
        text-transform: uppercase;
        letter-spacing: 1px;
    }

    /* Role Grid */
    .role-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
        gap: 8px;
    }

    .role-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 12px 8px;
        background: #1a1a1a;
        border: 1px solid #333;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.15s ease;
    }

    .role-card:hover {
        border-color: var(--text-muted);
        background: #222;
    }

    .role-card.selected {
        border-color: var(--accent);
        background: rgba(0, 255, 65, 0.05);
        box-shadow: 0 0 8px rgba(0, 255, 65, 0.2);
    }

    .role-card :global(.platform-icon) {
        filter: drop-shadow(0 0 4px rgba(0, 255, 65, 0.3));
    }

    .role-name {
        font-size: 11px;
        color: #e0e0e0;
        text-align: center;
        font-weight: 500;
    }

    /* Role Info */
    .role-info {
        margin-top: 12px;
        padding: 10px;
        background: #111;
        border: 1px solid #333;
        border-radius: 4px;
    }

    .role-header-row {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 8px;
    }

    .role-name-sm {
        font-size: 12px;
        font-weight: 600;
        color: var(--text);
    }

    .role-os-badge {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        margin-left: auto;
        padding: 2px 6px;
        background: var(--bg-tertiary);
        border: 1px solid var(--border);
        border-radius: 3px;
        font-size: 10px;
        color: var(--text-muted);
    }

    .role-tools {
        display: flex;
        flex-wrap: wrap;
        gap: 4px;
    }

    .tool-badge {
        padding: 2px 6px;
        background: var(--bg-tertiary);
        border: 1px solid var(--border);
        border-radius: 3px;
        font-size: 10px;
        color: var(--text-muted);
        font-family: var(--font-mono);
    }

    /* OS Grid */
    .os-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
        gap: 8px;
    }

    .os-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 6px;
        padding: 12px 8px;
        background: #1a1a1a;
        border: 1px solid #333;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.15s ease;
        position: relative;
    }

    .os-card:hover {
        border-color: var(--accent);
        background: rgba(0, 255, 65, 0.05);
        transform: translateY(-2px);
    }

    .os-card :global(.platform-icon) {
        filter: drop-shadow(0 0 4px rgba(0, 255, 65, 0.3));
    }

    .os-name {
        font-size: 11px;
        color: #e0e0e0;
        text-align: center;
    }

    .popular-badge {
        position: absolute;
        top: 4px;
        right: 4px;
        padding: 1px 4px;
        background: var(--accent);
        color: #000;
        font-size: 8px;
        font-weight: 600;
        border-radius: 2px;
        text-transform: uppercase;
    }

    /* Resource Configuration */
    .resource-toggle {
        display: flex;
        align-items: center;
        gap: 8px;
        width: 100%;
        padding: 8px 12px;
        background: #111;
        border: 1px solid #333;
        border-radius: 4px;
        cursor: pointer;
        transition: all 0.15s ease;
    }

    .resource-toggle:hover {
        border-color: var(--text-muted);
        background: #1a1a1a;
    }

    .resource-toggle h4 {
        margin: 0;
        font-size: 12px;
        color: var(--accent);
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .toggle-icon {
        font-size: 10px;
        color: var(--text-muted);
    }

    .resource-preview {
        margin-left: auto;
        font-size: 11px;
        font-family: var(--font-mono);
        color: var(--text-muted);
    }

    .resource-config {
        margin-top: 12px;
        padding: 12px;
        background: #111;
        border: 1px solid #333;
        border-radius: 4px;
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .resource-row {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .resource-row label {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .resource-label {
        font-size: 12px;
        color: var(--text);
        font-weight: 500;
    }

    .resource-value {
        font-size: 12px;
        font-family: var(--font-mono);
        color: var(--accent);
        font-weight: 600;
    }

    .resource-row input[type="range"] {
        width: 100%;
        height: 6px;
        -webkit-appearance: none;
        appearance: none;
        background: #333;
        border-radius: 3px;
        outline: none;
        margin: 8px 0;
        cursor: pointer;
    }

    .resource-row input[type="range"]::-webkit-slider-runnable-track {
        width: 100%;
        height: 6px;
        background: #333;
        border-radius: 3px;
    }

    .resource-row input[type="range"]::-webkit-slider-thumb {
        -webkit-appearance: none;
        width: 16px;
        height: 16px;
        background: var(--accent);
        border-radius: 50%;
        cursor: pointer;
        box-shadow: 0 0 8px rgba(0, 255, 65, 0.5);
        margin-top: -5px;
        transition: transform 0.15s, box-shadow 0.15s;
    }

    .resource-row input[type="range"]::-webkit-slider-thumb:hover {
        transform: scale(1.1);
        box-shadow: 0 0 12px rgba(0, 255, 65, 0.7);
    }

    .resource-row input[type="range"]::-moz-range-track {
        width: 100%;
        height: 6px;
        background: #333;
        border-radius: 3px;
    }

    .resource-row input[type="range"]::-moz-range-thumb {
        width: 16px;
        height: 16px;
        background: var(--accent);
        border: none;
        border-radius: 50%;
        cursor: pointer;
        box-shadow: 0 0 8px rgba(0, 255, 65, 0.5);
    }

    .resource-row input[type="range"]::-moz-range-thumb:hover {
        box-shadow: 0 0 12px rgba(0, 255, 65, 0.7);
    }

    .resource-range {
        display: flex;
        justify-content: space-between;
        font-size: 9px;
        color: var(--text-muted);
        font-family: var(--font-mono);
    }

    .resource-hint {
        font-size: 10px;
        color: var(--text-muted);
        margin: 4px 0 0 0;
        text-align: center;
        font-style: italic;
    }

    /* Compact mode adjustments */
    .compact .role-grid {
        grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    }

    .compact .os-grid {
        grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
    }

    .compact .role-card,
    .compact .os-card {
        padding: 10px 6px;
    }
</style>
