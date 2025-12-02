<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { containers, type ProgressEvent } from '$stores/containers';
  import { toast } from '$stores/toast';
  import { api } from '$utils/api';

  const dispatch = createEventDispatcher<{
    cancel: void;
    created: { id: string; name: string };
  }>();

  // State
  let containerName = '';
  let selectedImage = '';
  let customImage = '';
  let isCreating = false;
  let progress = 0;
  let progressMessage = '';
  let progressStage = '';

  // Available images
  let images: Array<{
    name: string;
    display_name: string;
    description: string;
    category: string;
    popular?: boolean;
  }> = [];

  // Image icons
  const imageIcons: Record<string, string> = {
    ubuntu: '🟠',
    debian: '🔴',
    alpine: '🔵',
    fedora: '🔵',
    centos: '🟣',
    rocky: '🟣',
    alma: '🟣',
    arch: '🔷',
    kali: '🐉',
    parrot: '🦜',
    custom: '📦',
  };

  function getIcon(imageName: string): string {
    const lower = imageName.toLowerCase();
    for (const [key, icon] of Object.entries(imageIcons)) {
      if (lower.includes(key)) return icon;
    }
    return '🐧';
  }

  // Load available images
  onMount(async () => {
    const { data, error } = await api.get<{
      images?: typeof images;
      popular?: typeof images;
    }>('/api/images');

    if (data) {
      images = data.images || data.popular || [];
    } else if (error) {
      toast.error('Failed to load images');
    }
  });

  // Generate random name
  function generateName(): string {
    const adjectives = [
      'swift', 'bold', 'calm', 'dark', 'eager', 'fast', 'grand', 'happy',
      'keen', 'light', 'merry', 'noble', 'proud', 'quick', 'rare', 'sharp',
    ];
    const nouns = [
      'ant', 'bear', 'cat', 'fox', 'hawk', 'lion', 'owl', 'wolf',
      'tiger', 'eagle', 'shark', 'cobra', 'raven', 'viper', 'lynx', 'orca',
    ];
    const adj = adjectives[Math.floor(Math.random() * adjectives.length)];
    const noun = nouns[Math.floor(Math.random() * nouns.length)];
    const num = Math.floor(Math.random() * 1000);
    return `${adj}-${noun}-${num}`;
  }

  // Validate form
  $: isValid =
    selectedImage &&
    (selectedImage !== 'custom' || customImage.trim()) &&
    !isCreating;

  // Handle creation
  async function handleCreate() {
    if (!isValid) return;

    isCreating = true;
    progress = 0;
    progressMessage = 'Starting...';
    progressStage = 'initializing';

    const name = containerName.trim() || generateName();
    const image = selectedImage;
    const custom = selectedImage === 'custom' ? customImage.trim() : undefined;

    containers.createContainerWithProgress(
      name,
      image,
      custom,
      // onProgress - updates UI during creation
      (event: ProgressEvent) => {
        progress = event.progress || 0;
        progressMessage = event.message || '';
        progressStage = event.stage || '';

        // Show error toast if there's an error in progress
        if (event.error) {
          isCreating = false;
          toast.error(event.error);
        }
      },
      // onComplete - called when container is fully created
      (container) => {
        isCreating = false;
        toast.success(`Terminal "${container.name}" created!`);
        dispatch('created', { id: container.id, name: container.name });
      },
      // onError
      (error) => {
        isCreating = false;
        toast.error(error);
      }
    );
  }

  function handleCancel() {
    if (!isCreating) {
      dispatch('cancel');
    }
  }

  function selectImage(imageName: string) {
    selectedImage = imageName;
  }
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
      <!-- Terminal Name -->
      <div class="form-group">
        <label for="container-name">
          Terminal Name
          <span class="optional">(optional)</span>
        </label>
        <input
          type="text"
          id="container-name"
          bind:value={containerName}
          placeholder="Auto-generated if empty"
          maxlength="64"
        />
        <span class="form-hint">
          Leave empty for auto-generated name like "swift-tiger-123"
        </span>
      </div>

      <!-- Image Selection -->
      <div class="form-group">
        <label>Select Operating System</label>
        <div class="image-grid">
          {#each images as image (image.name)}
            <button
              type="button"
              class="image-card"
              class:selected={selectedImage === image.name}
              on:click={() => selectImage(image.name)}
            >
              <span class="image-icon">{getIcon(image.name)}</span>
              <span class="image-name">{image.display_name || image.name}</span>
              {#if image.popular}
                <span class="popular-badge">Popular</span>
              {/if}
            </button>
          {/each}

          <!-- Custom Image Option -->
          <button
            type="button"
            class="image-card"
            class:selected={selectedImage === 'custom'}
            on:click={() => selectImage('custom')}
          >
            <span class="image-icon">📦</span>
            <span class="image-name">Custom Image</span>
          </button>
        </div>
      </div>

      <!-- Custom Image Input -->
      {#if selectedImage === 'custom'}
        <div class="form-group">
          <label for="custom-image">Docker Image</label>
          <input
            type="text"
            id="custom-image"
            bind:value={customImage}
            placeholder="e.g., nginx:latest or ghcr.io/user/image:tag"
          />
          <span class="form-hint">
            Enter a valid Docker image name from Docker Hub or any registry
          </span>
        </div>
      {/if}

      <!-- Actions -->
      <div class="form-actions">
        <button
          type="button"
          class="btn btn-secondary"
          on:click={handleCancel}
        >
          Cancel
        </button>
        <button
          type="button"
          class="btn btn-primary btn-lg"
          disabled={!isValid}
          on:click={handleCreate}
        >
          Create Terminal
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .create-container {
    max-width: 800px;
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
    gap: 24px;
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
  }

  .form-hint {
    font-size: 11px;
    color: var(--text-muted);
  }

  /* Image Grid */
  .image-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
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
    border-color: var(--text-muted);
    background: var(--bg-tertiary);
  }

  .image-card.selected {
    border-color: var(--accent);
    background: var(--accent-dim);
  }

  .image-icon {
    font-size: 28px;
  }

  .image-name {
    font-size: 12px;
    color: var(--text);
    text-align: center;
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

  /* Form Actions */
  .form-actions {
    display: flex;
    justify-content: flex-end;
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
