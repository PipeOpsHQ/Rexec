<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { containers, type ProgressEvent } from "$stores/containers";
    import { toast } from "$stores/toast";
    import { api } from "$utils/api";

    const dispatch = createEventDispatcher<{
        cancel: void;
        created: { id: string; name: string };
    }>();

    // State
    let containerName = "";
    let selectedImage = "";
    let selectedRole = "standard";
    let customImage = "";
    let isCreating = false;
    let progress = 0;
    let progressMessage = "";
    let progressStage = "";

    // Available images
    let images: Array<{
        name: string;
        display_name: string;
        description: string;
        category: string;
        popular?: boolean;
    }> = [];

    // Role to preferred OS mapping
    const roleToOS: Record<string, string> = {
        standard: "alpine", // Minimalist loves lightweight
        node: "ubuntu", // Best Node.js support
        python: "ubuntu", // Best Python/data science support
        go: "alpine", // Go's preferred container OS
        neovim: "arch", // Power users love Arch
        devops: "alpine", // Container standard
        overemployed: "alpine", // Fast startup
    };

    // Auto-select OS when role changes
    $: if (selectedRole && roleToOS[selectedRole]) {
        const preferredOS = roleToOS[selectedRole];
        // Only auto-select if images are loaded and the preferred OS exists
        if (
            images.length > 0 &&
            images.some((img) => img.name === preferredOS)
        ) {
            selectedImage = preferredOS;
        }
    }

    // Available roles with fun descriptions
    const roles = [
        {
            id: "standard",
            name: "The Minimalist",
            icon: "🧘",
            desc: "I use Arch btw. Just give me a shell.",
        },
        {
            id: "node",
            name: "10x JS Ninja",
            icon: "🚀",
            desc: "Ship fast, break things, npm install everything.",
        },
        {
            id: "python",
            name: "Data Wizard",
            icon: "🧙‍♂️",
            desc: "Import antigravity. I speak in list comprehensions.",
        },
        {
            id: "go",
            name: "The Gopher",
            icon: "🐹",
            desc: "If err != nil { panic(err) }. Simplicity is key.",
        },
        {
            id: "neovim",
            name: "Neovim God",
            icon: "⌨️",
            desc: "My config is longer than your code. Mouse? What mouse?",
        },
        {
            id: "devops",
            name: "YAML Herder",
            icon: "☸️",
            desc: "I don't write code, I write config. Prod is my playground.",
        },
        {
            id: "overemployed",
            name: "The Overemployed",
            icon: "💼",
            desc: "Working 4 remote jobs. Need max efficiency.",
        },
    ];

    // Image icons
    const imageIcons: Record<string, string> = {
        ubuntu: "🟠",
        debian: "🔴",
        alpine: "🔵",
        fedora: "🔵",
        centos: "🟣",
        rocky: "🟢",
        alma: "🟣",
        arch: "🔷",
        kali: "🐉",
        parrot: "🦜",
        mint: "🌿",
        elementary: "🪶",
        devuan: "🔘",
        blackarch: "🖤",
        manjaro: "🟩",
        opensuse: "🦎",
        tumbleweed: "🌀",
        gentoo: "🗿",
        void: "⬛",
        nixos: "❄️",
        slackware: "📦",
        busybox: "📦",
        amazonlinux: "🟧",
        oracle: "🔶",
        rhel: "🎩",
        openeuler: "🔵",
        clearlinux: "💎",
        photon: "☀️",
        raspberrypi: "🍓",
        scientific: "🔬",
        rancheros: "🐄",
        custom: "📦",
    };

    function getIcon(imageName: string): string {
        const lower = imageName.toLowerCase();
        for (const [key, icon] of Object.entries(imageIcons)) {
            if (lower.includes(key)) return icon;
        }
        return "🐧";
    }

    // Load available images
    onMount(async () => {
        const { data, error } = await api.get<{
            images?: typeof images;
            popular?: typeof images;
        }>("/api/images?all=true");

        if (data) {
            images = data.images || data.popular || [];
        } else if (error) {
            toast.error("Failed to load images");
        }
    });

    // Generate random name
    function generateName(): string {
        const adjectives = [
            "swift",
            "bold",
            "calm",
            "dark",
            "eager",
            "fast",
            "grand",
            "happy",
            "keen",
            "light",
            "merry",
            "noble",
            "proud",
            "quick",
            "rare",
            "sharp",
        ];
        const nouns = [
            "ant",
            "bear",
            "cat",
            "fox",
            "hawk",
            "lion",
            "owl",
            "wolf",
            "tiger",
            "eagle",
            "shark",
            "cobra",
            "raven",
            "viper",
            "lynx",
            "orca",
        ];
        const adj = adjectives[Math.floor(Math.random() * adjectives.length)];
        const noun = nouns[Math.floor(Math.random() * nouns.length)];
        const num = Math.floor(Math.random() * 1000);
        return `${adj}-${noun}-${num}`;
    }

    // Handle image selection and creation
    function selectAndCreate(imageName: string) {
        if (isCreating) return;

        selectedImage = imageName;

        // For custom image, prompt for input
        if (imageName === "custom") {
            const input = prompt(
                "Enter Docker image name (e.g., nginx:latest):",
            );
            if (!input || !input.trim()) {
                selectedImage = "";
                return;
            }
            customImage = input.trim();
        }

        // Start creation
        isCreating = true;
        progress = 0;
        progressMessage = "Starting...";
        progressStage = "initializing";

        const name = containerName.trim() || generateName();
        const image = selectedImage;
        const custom = selectedImage === "custom" ? customImage : undefined;

        containers.createContainerWithProgress(
            name,
            image,
            custom,
            selectedRole,
            // onProgress
            (event: ProgressEvent) => {
                progress = event.progress || 0;
                progressMessage = event.message || "";
                progressStage = event.stage || "";
            },
            // onComplete
            (container) => {
                isCreating = false;
                const containerName = container?.name || name;
                const containerId = container?.id || container?.db_id;

                if (!containerId) {
                    toast.error(
                        "Container created but ID not found. Please refresh.",
                    );
                    return;
                }

                toast.success(`Terminal "${containerName}" created!`);

                // Delay before dispatching to ensure container is ready
                setTimeout(() => {
                    dispatch("created", {
                        id: containerId,
                        name: containerName,
                    });
                }, 1000);
            },
            // onError
            (error) => {
                isCreating = false;
                let errorMsg = "Failed to create terminal";
                if (typeof error === "string" && error.trim()) {
                    errorMsg = error;
                } else if (error && typeof error === "object") {
                    errorMsg =
                        error.message || error.error || JSON.stringify(error);
                }
                if (!errorMsg || errorMsg === "undefined") {
                    errorMsg = "Failed to create terminal. Please try again.";
                }
                toast.error(errorMsg);
            },
        );
    }

    function handleCancel() {
        if (!isCreating) {
            dispatch("cancel");
        }
    }

    // Get current role description
    $: currentRoleDesc = roles.find((r) => r.id === selectedRole)?.desc || "";
</script>

<div class="create-container">
    <div class="create-header">
        <button class="back-btn" on:click={handleCancel} disabled={isCreating}>
            ← Back
        </button>
        <h1>Create Terminal</h1>
    </div>

    {#if isCreating}
        <div class="progress-section">
            <div class="progress-header">
                <h2>Creating Terminal</h2>
                <span class="progress-percent">{progress}%</span>
            </div>

            <div class="progress-bar">
                <div class="progress-fill" style="width: {progress}%"></div>
            </div>

            <div class="progress-info">
                <span class="progress-stage">{progressStage}</span>
                <span class="progress-message">{progressMessage}</span>
            </div>

            <div class="progress-spinner">
                <div class="spinner-large"></div>
            </div>
        </div>
    {:else}
        <div class="create-form">
            <!-- Terminal Name (Optional) -->
            <div class="form-group name-group">
                <label for="container-name">
                    Terminal Name
                    <span class="optional"
                        >(optional - auto-generated if empty)</span
                    >
                </label>
                <input
                    type="text"
                    id="container-name"
                    bind:value={containerName}
                    placeholder="e.g., my-dev-box"
                    maxlength="64"
                />
            </div>

            <!-- Step 1: Environment/Role Selection -->
            <div class="form-section">
                <div class="section-header">
                    <span class="step-number">1</span>
                    <h2>Choose Your Environment</h2>
                </div>
                <p class="section-desc">What kind of developer are you?</p>

                <div class="role-grid">
                    {#each roles as role}
                        <button
                            type="button"
                            class="role-card"
                            class:selected={selectedRole === role.id}
                            on:click={() => (selectedRole = role.id)}
                            title={role.desc}
                        >
                            <span class="role-icon">{role.icon}</span>
                            <span class="role-name">{role.name}</span>
                        </button>
                    {/each}
                </div>

                <p class="role-desc">{currentRoleDesc}</p>
            </div>

            <!-- Step 2: OS Selection -->
            <div class="form-section">
                <div class="section-header">
                    <span class="step-number">2</span>
                    <h2>Select Operating System</h2>
                </div>
                <p class="section-desc">Click to create your terminal</p>

                <div class="image-grid">
                    {#each images as image (image.name)}
                        <button
                            type="button"
                            class="image-card"
                            on:click={() => selectAndCreate(image.name)}
                        >
                            <span class="image-icon">{getIcon(image.name)}</span
                            >
                            <span class="image-name"
                                >{image.display_name || image.name}</span
                            >
                            {#if image.popular}
                                <span class="popular-badge">Popular</span>
                            {/if}
                        </button>
                    {/each}

                    <!-- Custom Image Option -->
                    <button
                        type="button"
                        class="image-card custom-card"
                        on:click={() => selectAndCreate("custom")}
                    >
                        <span class="image-icon">📦</span>
                        <span class="image-name">Custom Image</span>
                    </button>
                </div>
            </div>

            <!-- Cancel Action -->
            <div class="form-actions">
                <button
                    type="button"
                    class="btn btn-secondary"
                    on:click={handleCancel}
                >
                    Cancel
                </button>
            </div>
        </div>
    {/if}
</div>

<style>
    .create-container {
        max-width: 900px;
        margin: 0 auto;
        animation: fadeIn 0.2s ease;
    }

    .create-header {
        display: flex;
        align-items: center;
        gap: 16px;
        margin-bottom: 32px;
        padding-bottom: 16px;
        border-bottom: 1px solid var(--border);
    }

    .back-btn {
        background: none;
        border: 1px solid var(--border);
        color: var(--text-secondary);
        padding: 6px 12px;
        font-family: var(--font-mono);
        font-size: 12px;
        cursor: pointer;
        transition: all 0.2s;
    }

    .back-btn:hover:not(:disabled) {
        border-color: var(--text);
        color: var(--text);
    }

    .back-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .create-header h1 {
        font-size: 20px;
        text-transform: uppercase;
        letter-spacing: 1px;
        margin: 0;
    }

    /* Form Styles */
    .create-form {
        display: flex;
        flex-direction: column;
        gap: 32px;
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
        color: var(--text-secondary);
    }

    .optional {
        color: var(--text-muted);
        text-transform: none;
        font-size: 11px;
    }

    .form-group input {
        width: 100%;
        max-width: 400px;
    }

    .name-group {
        padding-bottom: 16px;
        border-bottom: 1px solid var(--border);
    }

    /* Section Styles */
    .form-section {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .section-header {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .step-number {
        width: 28px;
        height: 28px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--accent);
        color: var(--bg);
        font-size: 14px;
        font-weight: bold;
        font-family: var(--font-mono);
    }

    .section-header h2 {
        font-size: 16px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        margin: 0;
        color: var(--text);
    }

    .section-desc {
        font-size: 13px;
        color: var(--text-muted);
        margin: 0;
        font-family: var(--font-mono);
    }

    /* Role Grid */
    .role-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
        gap: 12px;
    }

    .role-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 16px 12px;
        background: var(--bg-card);
        border: 1px solid var(--border);
        cursor: pointer;
        transition: all 0.2s;
        font-family: var(--font-mono);
    }

    .role-card:hover {
        border-color: var(--accent);
        background: var(--accent-dim);
    }

    .role-card.selected {
        border-color: var(--accent);
        background: rgba(0, 255, 65, 0.1);
        box-shadow: 0 0 10px rgba(0, 255, 65, 0.2);
    }

    .role-icon {
        font-size: 28px;
    }

    .role-name {
        font-size: 11px;
        color: var(--text);
        text-align: center;
    }

    .role-desc {
        font-size: 12px;
        color: var(--accent);
        font-family: var(--font-mono);
        font-style: italic;
        margin: 4px 0 0 0;
        min-height: 20px;
        text-align: center;
    }

    /* Image Grid */
    .image-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
        gap: 12px;
    }

    .image-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 16px 12px;
        background: var(--bg-card);
        border: 1px solid var(--border);
        cursor: pointer;
        transition: all 0.2s;
        font-family: var(--font-mono);
        position: relative;
    }

    .image-card:hover {
        border-color: var(--accent);
        background: var(--accent-dim);
        transform: translateY(-2px);
    }

    .image-icon {
        font-size: 28px;
    }

    .image-name {
        font-size: 11px;
        color: var(--text);
        text-align: center;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
    }

    .popular-badge {
        position: absolute;
        top: 4px;
        right: 4px;
        font-size: 8px;
        padding: 2px 4px;
        background: var(--accent);
        color: var(--bg);
        text-transform: uppercase;
    }

    .custom-card {
        border-style: dashed;
    }

    /* Form Actions */
    .form-actions {
        display: flex;
        justify-content: flex-start;
        gap: 12px;
        padding-top: 16px;
        border-top: 1px solid var(--border);
    }

    /* Progress Section */
    .progress-section {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 60px 20px;
        text-align: center;
    }

    .progress-header {
        display: flex;
        align-items: center;
        gap: 16px;
        margin-bottom: 24px;
    }

    .progress-header h2 {
        font-size: 18px;
        text-transform: uppercase;
        margin: 0;
    }

    .progress-percent {
        font-size: 14px;
        color: var(--accent);
        font-weight: 600;
    }

    .progress-bar {
        width: 100%;
        max-width: 400px;
        height: 4px;
        background: var(--bg-tertiary);
        border: 1px solid var(--border);
        margin-bottom: 16px;
        overflow: hidden;
    }

    .progress-fill {
        height: 100%;
        background: var(--accent);
        transition: width 0.3s ease;
    }

    .progress-info {
        display: flex;
        flex-direction: column;
        gap: 4px;
        margin-bottom: 32px;
    }

    .progress-stage {
        font-size: 12px;
        text-transform: uppercase;
        color: var(--accent);
    }

    .progress-message {
        font-size: 13px;
        color: var(--text-muted);
    }

    .progress-spinner {
        margin-top: 20px;
    }

    .spinner-large {
        width: 40px;
        height: 40px;
        border: 3px solid var(--border);
        border-top-color: var(--accent);
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    @media (max-width: 600px) {
        .role-grid {
            grid-template-columns: repeat(2, 1fr);
        }

        .image-grid {
            grid-template-columns: repeat(2, 1fr);
        }

        .form-actions {
            flex-direction: column;
        }

        .form-actions button {
            width: 100%;
        }
    }
</style>
