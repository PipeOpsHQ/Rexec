<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import StatusIcon from "./icons/StatusIcon.svelte";
    import { auth } from "$stores/auth";
    import { toast } from "$stores/toast";

    const dispatch = createEventDispatcher<{
        guest: void;
        navigate: { view: string; slug?: string };
    }>();

    let isOAuthLoading = false;
    let text = "Initialize...";
    let typedText = "";
    let cursorVisible = true;
    let scrollY = 0;

    const featuredUseCases = [
        {
            slug: "ephemeral-dev-environments",
            title: "EPHEMERAL_DEV",
            icon: "bolt",
            desc: "Spin up fresh environments in milliseconds."
        },
        {
            slug: "collaborative-intelligence",
            title: "COLLAB_INTELLIGENCE",
            icon: "ai",
            desc: "Shared workspace for humans and AI agents."
        },
        {
            slug: "universal-jump-host",
            title: "UNIVERSAL_JUMP",
            icon: "shield",
            desc: "Secure browser-based access to private infrastructure."
        }
    ];

    onMount(() => {
        let i = 0;
        const typeInterval = setInterval(() => {
            typedText = text.slice(0, i + 1);
            i++;
            if (i > text.length) clearInterval(typeInterval);
        }, 100);

        const cursorInterval = setInterval(() => {
            cursorVisible = !cursorVisible;
        }, 500);

        return () => {
            clearInterval(typeInterval);
            clearInterval(cursorInterval);
        };
    });

    function handleGuestClick() {
        dispatch("guest");
    }

    function navigateToUseCase(slug: string) {
        window.location.href = `/use-cases/${slug}`;
    }

    async function handleOAuthLogin() {
        if (isOAuthLoading) return;

        isOAuthLoading = true;
        try {
            const url = await auth.getOAuthUrl();
            if (url) {
                window.location.href = url;
            } else {
                toast.error("Connection failed.");
                isOAuthLoading = false;
            }
        } catch (e) {
            toast.error("Connection failed.");
            isOAuthLoading = false;
        }
    }
</script>

<svelte:window bind:scrollY={scrollY} />

<div class="promo-container">
    <div class="grid-overlay"></div>
    <div class="scanline"></div>
    
    <!-- Space Background Elements -->
    <div class="space-scene" style="transform: translateY({scrollY * 0.2}px)">
        <div class="solar-system">
            <div class="orbit orbit-1">
                <div class="planet planet-1"></div>
            </div>
            <div class="orbit orbit-2">
                <div class="planet planet-2"></div>
            </div>
            <div class="orbit orbit-3">
                <div class="planet planet-3"></div>
            </div>
            <div class="sun"></div>
        </div>
        
        <div class="rocket-container">
            <svg class="rocket" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path class="rocket-body" d="M12 2c0 0-8 6-8 12 0 4.418 3.582 8 8 8s8-3.582 8-8c0-6-8-12-8-12z" fill="#050505"/>
                <path class="rocket-fin-l" d="M4 14l-2 4 4 1" />
                <path class="rocket-fin-r" d="M20 14l2 4-4 1" />
                <circle class="rocket-window" cx="12" cy="10" r="2" fill="#00ffaa" />
                <path class="rocket-flame" d="M12 22v4M10 22v2M14 22v2" stroke="#ff00ff" />
            </svg>
            <div class="rocket-trail"></div>
        </div>
    </div>

    <div class="content-wrapper">
        <header class="hero">
            <div class="glitch-wrapper">
                <h1 class="glitch" data-text="REXEC_SYSTEMS">REXEC_SYSTEMS</h1>
            </div>
            <div class="sub-hero">
                <span class="prompt">&gt;</span> 
                <span class="typed">{typedText}</span>
                <span class="cursor" style:opacity={cursorVisible ? 1 : 0}>_</span>
            </div>
            
            <p class="mission-statement">
                INSTANT. SECURE. EPHEMERAL.
                <br>
                THE TERMINAL LAYER FOR THE WEB.
            </p>

            <div class="cta-group">
                <button class="cyber-btn primary" on:click={handleGuestClick}>
                    <span class="btn-content">INIT_GUEST_SESSION</span>
                    <span class="btn-glitch"></span>
                </button>
                <button class="cyber-btn secondary" on:click={handleOAuthLogin} disabled={isOAuthLoading}>
                    <span class="btn-content">{isOAuthLoading ? "CONNECTING..." : "AUTH_PIPEOPS_ID"}</span>
                </button>
            </div>
        </header>

        <section class="features-grid">
            <div class="cyber-card">
                <div class="card-header">
                    <StatusIcon status="bolt" size={16} />
                    <h3>ZERO_LATENCY</h3>
                </div>
                <p>Spin up environments in milliseconds. No cold starts. Pure speed.</p>
            </div>

            <div class="cyber-card">
                <div class="card-header">
                    <StatusIcon status="shield" size={16} />
                    <h3>SECURE_ENCLAVE</h3>
                </div>
                <p>Isolated sandboxes. Ephemeral filesystems. Your data vanishes on exit.</p>
            </div>

            <div class="cyber-card">
                <div class="card-header">
                    <StatusIcon status="connected" size={16} />
                    <h3>GLOBAL_MESH</h3>
                </div>
                <p>Access your workspace from any node on the network. Browser-native SSH.</p>
            </div>
        </section>

        <!-- New Use Cases Section -->
        <section class="use-cases-section">
            <div class="section-title">
                <span class="bracket">[</span>
                <h2>DEPLOYMENT_PROTOCOLS</h2>
                <span class="bracket">]</span>
            </div>
            
            <div class="cases-list">
                {#each featuredUseCases as useCase, i}
                    <button 
                        class="case-item" 
                        on:click={() => navigateToUseCase(useCase.slug)}
                        style="animation-delay: {i * 150}ms"
                    >
                        <div class="case-icon-wrapper">
                            <StatusIcon status={useCase.icon} size={20} />
                        </div>
                        <div class="case-info">
                            <h4>{useCase.title}</h4>
                            <p>{useCase.desc}</p>
                        </div>
                        <div class="case-arrow">→</div>
                    </button>
                {/each}
            </div>
        </section>

        <footer class="system-status">
            <div class="status-line">
                <span>SYSTEM: ONLINE</span>
                <span>VERSION: 2.0.4</span>
                <span>LATENCY: &lt;15ms</span>
            </div>
        </footer>
    </div>
</div>

<style>
    :global(body) {
        background-color: #050505;
        overflow-x: hidden;
    }

    .promo-container {
        position: relative;
        min-height: 100vh;
        width: 100%;
        background-color: #030303;
        color: #e0e0e0;
        font-family: "JetBrains Mono", "Fira Code", monospace;
        overflow: hidden;
        display: flex;
        flex-direction: column;
        align-items: center;
        perspective: 1000px;
    }

    /* Grid Background */
    .grid-overlay {
        position: absolute;
        inset: -50%;
        width: 200%;
        height: 200%;
        background-image: 
            linear-gradient(rgba(0, 255, 170, 0.05) 1px, transparent 1px),
            linear-gradient(90deg, rgba(0, 255, 170, 0.05) 1px, transparent 1px);
        background-size: 60px 60px;
        z-index: 1;
        pointer-events: none;
        transform: rotateX(60deg) translateY(-100px) translateZ(-200px);
        animation: grid-move 20s linear infinite;
        opacity: 0.4;
    }

    @keyframes grid-move {
        0% { transform: rotateX(60deg) translateY(0) translateZ(-200px); }
        100% { transform: rotateX(60deg) translateY(60px) translateZ(-200px); }
    }

    /* Space Scene & Animations */
    .space-scene {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        z-index: 2;
        pointer-events: none;
        overflow: hidden;
    }

    .solar-system {
        position: absolute;
        top: 20%;
        right: 10%;
        width: 400px;
        height: 400px;
        transform-style: preserve-3d;
        transform: rotateX(75deg) rotateY(10deg);
        opacity: 0.3;
    }

    .sun {
        position: absolute;
        top: 50%;
        left: 50%;
        width: 40px;
        height: 40px;
        transform: translate(-50%, -50%);
        background: radial-gradient(circle, #ff00ff 0%, transparent 70%);
        border-radius: 50%;
        box-shadow: 0 0 40px #ff00ff;
    }

    .orbit {
        position: absolute;
        top: 50%;
        left: 50%;
        border: 1px solid rgba(0, 255, 170, 0.2);
        border-radius: 50%;
        transform: translate(-50%, -50%);
    }

    .orbit-1 { width: 120px; height: 120px; animation: rotate 10s linear infinite; }
    .orbit-2 { width: 220px; height: 220px; animation: rotate 20s linear infinite reverse; }
    .orbit-3 { width: 340px; height: 340px; animation: rotate 35s linear infinite; }

    .planet {
        position: absolute;
        top: 50%;
        right: -4px;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #00ffaa;
        box-shadow: 0 0 10px #00ffaa;
        transform: translateY(-50%);
    }

    .planet-2 { background: #00ffff; box-shadow: 0 0 10px #00ffff; width: 6px; height: 6px; }
    .planet-3 { background: #ffffff; box-shadow: 0 0 10px #ffffff; width: 4px; height: 4px; }

    @keyframes rotate {
        from { transform: translate(-50%, -50%) rotate(0deg); }
        to { transform: translate(-50%, -50%) rotate(360deg); }
    }

    .rocket-container {
        position: absolute;
        bottom: 10%;
        left: -100px; /* Start off-screen */
        width: 60px;
        height: 60px;
        animation: rocket-flight 30s linear infinite;
        filter: drop-shadow(0 0 10px rgba(0, 255, 170, 0.5));
    }

    .rocket {
        width: 100%;
        height: 100%;
        transform: rotate(45deg);
        overflow: visible;
    }
    
    .rocket path {
        stroke: #00ffaa;
        stroke-width: 1.5;
    }

    .rocket-flame {
        animation: flame-flicker 0.1s linear infinite alternate;
        stroke: #ff00ff !important;
    }

    @keyframes rocket-flight {
        0% { left: -100px; bottom: 10%; transform: rotate(0deg); opacity: 0; }
        10% { opacity: 1; }
        40% { left: 40%; bottom: 60%; transform: rotate(10deg); }
        60% { left: 60%; bottom: 40%; transform: rotate(20deg); }
        90% { opacity: 1; }
        100% { left: 110%; bottom: 80%; transform: rotate(30deg); opacity: 0; }
    }

    @keyframes flame-flicker {
        from { opacity: 0.6; transform: scaleY(0.8); }
        to { opacity: 1; transform: scaleY(1.2); }
    }

    /* CRT Scanline */
    .scanline {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: linear-gradient(
            to bottom,
            transparent 50%,
            rgba(0, 0, 0, 0.2) 51%
        );
        background-size: 100% 4px;
        pointer-events: none;
        z-index: 10;
        opacity: 0.3;
    }

    .content-wrapper {
        position: relative;
        z-index: 20;
        max-width: 1000px;
        width: 100%;
        padding: 4rem 2rem;
        display: flex;
        flex-direction: column;
        gap: 5rem;
        text-align: center;
    }

    /* Hero Typography */
    .hero {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1.5rem;
    }

    .glitch-wrapper {
        position: relative;
    }

    .glitch {
        font-size: 5rem;
        font-weight: 900;
        text-transform: uppercase;
        position: relative;
        text-shadow: 2px 2px 0px #ff00ff, -2px -2px 0px #00ffff;
        animation: glitch-anim 2s infinite linear alternate-reverse;
        margin: 0;
        letter-spacing: -2px;
    }
    
    .glitch::before,
    .glitch::after {
        content: attr(data-text);
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
    }

    .glitch::before {
        left: 2px;
        text-shadow: -1px 0 #ff00c1;
        clip: rect(44px, 450px, 56px, 0);
        animation: glitch-anim-2 5s infinite linear alternate-reverse;
    }

    .glitch::after {
        left: -2px;
        text-shadow: -1px 0 #00fff9;
        clip: rect(44px, 450px, 56px, 0);
        animation: glitch-anim-2 5s infinite linear alternate-reverse;
    }

    .sub-hero {
        font-size: 1.5rem;
        color: #00ffaa;
        background: rgba(0, 255, 170, 0.05);
        padding: 0.5rem 1rem;
        border: 1px solid rgba(0, 255, 170, 0.2);
        box-shadow: 0 0 15px rgba(0, 255, 170, 0.1);
        backdrop-filter: blur(4px);
    }

    .prompt {
        color: #ff00ff;
        margin-right: 0.5rem;
    }

    .mission-statement {
        font-size: 1.1rem;
        line-height: 1.6;
        color: #888;
        letter-spacing: 2px;
        max-width: 600px;
        border-left: 2px solid #333;
        padding-left: 1rem;
    }

    /* Buttons */
    .cta-group {
        display: flex;
        gap: 2rem;
        margin-top: 1rem;
    }

    .cyber-btn {
        position: relative;
        padding: 1rem 2rem;
        background: transparent;
        border: none;
        cursor: pointer;
        font-family: inherit;
        font-size: 1rem;
        text-transform: uppercase;
        letter-spacing: 2px;
        transition: all 0.2s;
        clip-path: polygon(10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%, 0 10px);
    }

    .cyber-btn.primary {
        background: #00ffaa;
        color: #000;
        font-weight: 700;
    }

    .cyber-btn.primary:hover {
        background: #ccffee;
        box-shadow: 0 0 30px rgba(0, 255, 170, 0.6);
        transform: translateY(-2px);
    }

    .cyber-btn.secondary {
        background: transparent;
        border: 1px solid #00ffaa;
        color: #00ffaa;
    }

    .cyber-btn.secondary:hover {
        background: rgba(0, 255, 170, 0.1);
        box-shadow: 0 0 20px rgba(0, 255, 170, 0.3);
    }

    /* Features Grid */
    .features-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
        gap: 2rem;
        width: 100%;
    }

    .cyber-card {
        background: rgba(15, 15, 15, 0.9);
        border: 1px solid #333;
        padding: 2rem;
        text-align: left;
        transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
        position: relative;
        overflow: hidden;
        backdrop-filter: blur(10px);
    }

    .cyber-card::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 2px;
        background: linear-gradient(90deg, transparent, #00ffaa, transparent);
        transform: translateX(-100%);
        transition: transform 0.5s;
    }

    .cyber-card:hover {
        border-color: #00ffaa;
        transform: translateY(-5px);
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.7);
    }

    .cyber-card:hover::before {
        transform: translateX(100%);
    }

    .card-header {
        display: flex;
        align-items: center;
        gap: 1rem;
        margin-bottom: 1rem;
        color: #00ffaa;
    }

    .card-header h3 {
        font-size: 1.1rem;
        margin: 0;
        font-weight: 400;
    }

    .cyber-card p {
        color: #888;
        font-size: 0.9rem;
        line-height: 1.5;
        margin: 0;
    }

    /* Use Cases Section */
    .use-cases-section {
        width: 100%;
        display: flex;
        flex-direction: column;
        gap: 2rem;
        align-items: center;
    }

    .section-title {
        display: flex;
        align-items: center;
        gap: 1rem;
        font-size: 1.5rem;
        color: #e0e0e0;
        margin-bottom: 1rem;
    }

    .section-title .bracket {
        color: #ff00ff;
        font-weight: 300;
    }

    .section-title h2 {
        margin: 0;
        font-weight: 400;
        letter-spacing: 3px;
    }

    .cases-list {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        width: 100%;
        max-width: 800px;
    }

    .case-item {
        display: flex;
        align-items: center;
        gap: 1.5rem;
        padding: 1.5rem;
        background: rgba(10, 10, 10, 0.6);
        border: 1px solid #333;
        border-left: 3px solid #333;
        color: inherit;
        text-align: left;
        cursor: pointer;
        transition: all 0.3s ease;
        position: relative;
        overflow: hidden;
        animation: slide-in 0.6s cubic-bezier(0.23, 1, 0.32, 1) backwards;
    }

    .case-item::after {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: linear-gradient(90deg, transparent, rgba(0, 255, 170, 0.05), transparent);
        transform: translateX(-100%);
        transition: transform 0.4s;
    }

    .case-item:hover {
        background: rgba(20, 20, 20, 0.9);
        border-color: #555;
        border-left-color: #00ffaa;
        transform: translateX(10px);
        box-shadow: -5px 5px 15px rgba(0, 0, 0, 0.5);
    }

    .case-item:hover::after {
        transform: translateX(100%);
    }

    .case-icon-wrapper {
        color: #555;
        transition: color 0.3s;
    }

    .case-item:hover .case-icon-wrapper {
        color: #00ffaa;
    }

    .case-info {
        flex: 1;
    }

    .case-info h4 {
        margin: 0 0 0.5rem 0;
        font-size: 1.1rem;
        color: #fff;
        letter-spacing: 1px;
    }

    .case-info p {
        margin: 0;
        font-size: 0.9rem;
        color: #888;
    }

    .case-arrow {
        color: #333;
        font-size: 1.2rem;
        transition: all 0.3s;
    }

    .case-item:hover .case-arrow {
        color: #00ffaa;
        transform: translateX(5px);
    }

    @keyframes slide-in {
        from { opacity: 0; transform: translateX(-30px); }
        to { opacity: 1; transform: translateX(0); }
    }

    /* Footer Status */
    .system-status {
        margin-top: auto;
        width: 100%;
        border-top: 1px solid #333;
        padding-top: 2rem;
    }

    .status-line {
        display: flex;
        justify-content: space-between;
        color: #555;
        font-size: 0.8rem;
        text-transform: uppercase;
    }

    /* Animations */
    @keyframes glitch-anim {
        0% { transform: skew(0deg); }
        20% { transform: skew(-2deg); }
        40% { transform: skew(2deg); }
        60% { transform: skew(-1deg); }
        80% { transform: skew(3deg); }
        100% { transform: skew(0deg); }
    }

    @keyframes glitch-anim-2 {
        0% { clip: rect(12px, 9999px, 86px, 0); }
        20% { clip: rect(94px, 9999px, 2px, 0); }
        40% { clip: rect(24px, 9999px, 16px, 0); }
        60% { clip: rect(65px, 9999px, 120px, 0); }
        80% { clip: rect(3px, 9999px, 55px, 0); }
        100% { clip: rect(48px, 9999px, 92px, 0); }
    }

    @media (max-width: 768px) {
        .glitch { font-size: 3rem; }
        .cta-group { flex-direction: column; width: 100%; }
        .cyber-btn { width: 100%; }
        .features-grid { grid-template-columns: 1fr; }
        .solar-system { display: none; } /* Hide heavy animation on mobile */
        .rocket-container { display: none; }
    }
</style>
