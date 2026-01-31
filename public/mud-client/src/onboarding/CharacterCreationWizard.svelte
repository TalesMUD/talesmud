<style>
  .wizard-screen {
    position: fixed;
    inset: 0;
    background: #0a0e14;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .bg-image {
    position: absolute;
    inset: 0;
    background-size: cover;
    background-position: center;
    image-rendering: pixelated;
    opacity: 0.06;
    filter: blur(4px) saturate(0.3) brightness(0.7);
    transition: background-image 0.8s ease;
  }

  .bg-gradient {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse 70% 60% at 50% 45%, transparent 0%, #0a0e14 100%);
  }

  /* Step indicator */
  .step-indicator {
    position: relative;
    z-index: 3;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.9rem;
    padding: 1.8rem 0;
  }

  .step-dot {
    display: flex;
    align-items: center;
    gap: 0.48rem;
    font-size: 0.84rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: #4b5563;
    transition: color 0.3s ease;
  }

  .step-dot.active {
    color: #f59e0b;
  }

  .step-dot.completed {
    color: #16a34a;
  }

  .step-number {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    border: 1px solid #4b5563;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.78rem;
    transition: all 0.3s ease;
  }

  .step-dot.active .step-number {
    border-color: #f59e0b;
    background: rgba(245, 158, 11, 0.1);
  }

  .step-dot.completed .step-number {
    border-color: #16a34a;
    background: rgba(22, 163, 74, 0.1);
  }

  .step-line {
    width: 38px;
    height: 1px;
    background: #374151;
  }

  .step-line.completed {
    background: #16a34a;
  }

  /* Main content area */
  .wizard-content {
    position: relative;
    z-index: 3;
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0 2rem 2rem;
    overflow-y: auto;
    min-height: 0;
  }

  .step-title {
    font-family: 'Cinzel', serif;
    font-size: 1.68rem;
    font-weight: 600;
    color: #e5e7eb;
    text-align: center;
    margin-bottom: 0.48rem;
  }

  .step-description {
    font-size: 1.08rem;
    color: #9ca3af;
    text-align: center;
    margin-bottom: 1.8rem;
    max-width: 552px;
  }

  /* Template grid (Step 1) */
  .template-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 0.9rem;
    width: 100%;
    max-width: 1080px;
    padding: 0.6rem;
  }

  .template-card {
    background: rgba(0, 0, 0, 0.5);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 10px;
    padding: 1.44rem;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.72rem;
  }

  .template-card:hover {
    border-color: rgba(255, 255, 255, 0.15);
    background: rgba(0, 0, 0, 0.6);
    transform: translateY(-2px);
  }

  .template-card.selected {
    border-color: rgba(22, 163, 74, 0.5);
    background: rgba(22, 163, 74, 0.08);
  }

  .template-avatar {
    width: 77px;
    height: 77px;
    image-rendering: pixelated;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.08);
  }

  .template-card.selected .template-avatar {
    border-color: rgba(22, 163, 74, 0.3);
  }

  .template-name {
    font-family: 'Cinzel', serif;
    font-size: 1.14rem;
    font-weight: 600;
    color: #e5e7eb;
  }

  .template-desc {
    font-size: 0.96rem;
    color: #6b7280;
    line-height: 1.4;
    max-height: 2.8em;
    overflow: hidden;
  }

  .template-attrs {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.36rem;
    margin-top: auto;
  }

  .attr-badge {
    font-size: 0.78rem;
    padding: 0.18rem 0.48rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 3px;
    color: #9ca3af;
  }

  .attr-badge .attr-val {
    color: #f59e0b;
    margin-left: 0.2em;
  }

  /* Customize form (Step 2) */
  .customize-layout {
    display: flex;
    gap: 2.4rem;
    width: 100%;
    max-width: 768px;
    align-items: flex-start;
  }

  .preview-card {
    flex: 0 0 216px;
    background: rgba(0, 0, 0, 0.5);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    padding: 1.44rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.72rem;
  }

  .preview-name {
    font-family: 'Cinzel', serif;
    font-size: 1.14rem;
    font-weight: 600;
    color: #e5e7eb;
    word-break: break-word;
  }

  .preview-template {
    font-size: 0.84rem;
    font-weight: 500;
    color: #16a34a;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  .customize-form {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 1.44rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.48rem;
  }

  .form-label {
    font-size: 0.9rem;
    font-weight: 500;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .form-input,
  .form-textarea {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.12);
    outline: none;
    box-shadow: none;
    color: #e5e7eb;
    font-size: 1.2rem;
    padding: 0.84rem 1.08rem;
    border-radius: 7px;
    transition: border-color 0.2s ease;
    box-sizing: border-box;
    width: 100%;
  }

  .form-input:focus,
  .form-textarea:focus {
    border-color: rgba(255, 255, 255, 0.25);
  }

  .form-input::placeholder,
  .form-textarea::placeholder {
    color: #4b5563;
  }

  .form-textarea {
    resize: vertical;
    min-height: 96px;
    line-height: 1.5;
  }

  /* Confirm card (Step 3) */
  .confirm-card {
    background: rgba(0, 0, 0, 0.5);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    padding: 2.4rem;
    max-width: 456px;
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 1rem;
  }

  .confirm-avatar {
    width: 96px;
    height: 96px;
    image-rendering: pixelated;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.1);
  }

  .confirm-name {
    font-family: 'Cinzel', serif;
    font-size: 1.55rem;
    font-weight: 600;
    color: #e5e7eb;
  }

  .confirm-desc {
    font-size: 1.08rem;
    color: #9ca3af;
    line-height: 1.5;
  }

  .confirm-template {
    font-size: 0.84rem;
    font-weight: 500;
    color: #16a34a;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    padding: 0.3rem 0.72rem;
    background: rgba(22, 163, 74, 0.1);
    border-radius: 3px;
  }

  .confirm-attrs {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.36rem;
    width: 100%;
    margin-top: 0.6rem;
  }

  .confirm-attr {
    display: flex;
    justify-content: space-between;
    padding: 0.36rem 0.6rem;
    background: rgba(255, 255, 255, 0.03);
    border-radius: 4px;
    font-size: 0.96rem;
  }

  .confirm-attr-name {
    color: #6b7280;
    font-size: 0.84rem;
    text-transform: uppercase;
  }

  .confirm-attr-value {
    color: #f59e0b;
    font-weight: 600;
  }

  /* Buttons */
  .button-row {
    display: flex;
    gap: 0.9rem;
    margin-top: 1.8rem;
  }

  .btn-wizard {
    font-size: 1.02rem;
    font-weight: 500;
    padding: 0.84rem 2.16rem;
    border-radius: 7px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-wizard.primary {
    border: none;
    color: #fff;
    background: #16a34a;
  }

  .btn-wizard.primary:hover:not(:disabled) {
    background: #15803d;
    box-shadow: 0 4px 12px rgba(22, 163, 74, 0.3);
    transform: translateY(-1px);
  }

  .btn-wizard.secondary {
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: #d1d5db;
    background: transparent;
  }

  .btn-wizard.secondary:hover {
    border-color: rgba(255, 255, 255, 0.25);
    color: #e5e7eb;
    background: rgba(255, 255, 255, 0.04);
  }

  .btn-wizard:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Error */
  .error-banner {
    font-size: 0.96rem;
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
    border: 1px solid rgba(248, 113, 113, 0.2);
    padding: 0.6rem 1.2rem;
    border-radius: 7px;
    text-align: center;
  }

  /* Success animation */
  .success-overlay {
    position: fixed;
    inset: 0;
    z-index: 100;
    background: rgba(10, 14, 20, 0.95);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    animation: fadeIn 0.5s ease;
  }

  .success-text {
    font-family: 'Cinzel', serif;
    font-size: 1.8rem;
    font-weight: 600;
    color: #e5e7eb;
  }

  .success-subtext {
    font-size: 1.08rem;
    color: #6b7280;
  }

  .success-bar {
    width: 192px;
    height: 2px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 1px;
    overflow: hidden;
  }

  .success-bar-fill {
    height: 100%;
    width: 100%;
    background: #16a34a;
    animation: fillBar 1.5s ease-in-out forwards;
  }

  @keyframes fillBar {
    from { width: 0%; }
    to { width: 100%; }
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  /* Loading text */
  .loading-text {
    font-size: 0.96rem;
    color: #6b7280;
  }

  /* Responsive */
  @media (max-width: 600px) {
    .customize-layout {
      flex-direction: column;
      align-items: center;
    }
    .preview-card {
      flex: none;
      width: 100%;
    }
    .template-grid {
      grid-template-columns: 1fr;
    }
  }
</style>

<script>
  import { onMount } from "svelte";
  import { getCharacterTemplates, createNewCharacter } from "../api/characters.js";

  export let authToken;
  export let onComplete;

  // Wizard state
  let step = 1;
  let templates = [];
  let selectedTemplate = null;
  let characterName = "";
  let characterDescription = "";
  let creating = false;
  let error = "";
  let showSuccess = false;

  const backgrounds = [
    'img/bg/castle-1.png',
    'img/bg/forrest1.png',
    'img/bg/oldtown-entrance.png',
  ];

  $: currentBg = backgrounds[step - 1] || backgrounds[0];
  $: canProceedStep1 = selectedTemplate !== null;
  $: canProceedStep2 = characterName.trim().length >= 2;

  onMount(() => {
    getCharacterTemplates(
      (result) => { templates = result || []; },
      (err) => {
        console.error("Failed to load templates:", err);
        error = "Failed to load character templates.";
      }
    );
  });

  function getAvatar(name) {
    let hash = 0;
    for (let i = 0; i < (name || "").length; i++) {
      const char = name.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash |= 0;
    }
    const num = 1 + Math.abs(hash % 12);
    return "img/avatars/" + num + "p.png";
  }

  function selectTemplate(template) {
    selectedTemplate = template;
    characterName = template.name || "";
    characterDescription = template.description || "";
  }

  function nextStep() {
    if (step < 3) step++;
  }

  function prevStep() {
    if (step > 1) step--;
  }

  function handleCreate() {
    if (creating) return;

    creating = true;
    error = "";

    const createDTO = {
      name: characterName.trim(),
      description: characterDescription.trim(),
      templateId: selectedTemplate.id,
    };

    createNewCharacter(
      authToken,
      createDTO,
      (character) => {
        creating = false;
        showSuccess = true;
        setTimeout(() => {
          onComplete(character);
        }, 1800);
      },
      (err) => {
        creating = false;
        error = "Failed to create character. The name may already be taken.";
        console.error("Character creation failed:", err);
      }
    );
  }

  function handleNameKeydown(e) {
    if (e.key === "Enter" && canProceedStep2) {
      nextStep();
    }
  }
</script>

<div class="wizard-screen">
  <div class="bg-image" style="background-image: url('{currentBg}');"></div>
  <div class="bg-gradient"></div>

  <!-- Step indicator -->
  <div class="step-indicator">
    <div class="step-dot" class:active={step === 1} class:completed={step > 1}>
      <span class="step-number">1</span>
      <span>Class</span>
    </div>
    <div class="step-line" class:completed={step > 1}></div>
    <div class="step-dot" class:active={step === 2} class:completed={step > 2}>
      <span class="step-number">2</span>
      <span>Name</span>
    </div>
    <div class="step-line" class:completed={step > 2}></div>
    <div class="step-dot" class:active={step === 3}>
      <span class="step-number">3</span>
      <span>Confirm</span>
    </div>
  </div>

  <!-- Step content -->
  <div class="wizard-content">
    {#if step === 1}
      <!-- Step 1: Choose Template -->
      <h2 class="step-title">Choose a Template</h2>
      <p class="step-description">
        Select a character archetype to get started. You can customize it in the next step.
      </p>

      {#if templates.length === 0 && !error}
        <p class="loading-text">Loading templates...</p>
      {/if}

      <div class="template-grid">
        {#each templates as template}
          <div
            class="template-card"
            class:selected={selectedTemplate && selectedTemplate.id === template.id}
            on:click={() => selectTemplate(template)}
            on:keydown={(e) => e.key === 'Enter' && selectTemplate(template)}
            role="button"
            tabindex="0"
          >
            <img src={getAvatar(template.name)} alt="" class="template-avatar" />
            <span class="template-name">{template.name}</span>
            {#if template.description}
              <span class="template-desc">{template.description}</span>
            {/if}
            {#if template.attributes}
              <div class="template-attrs">
                {#each template.attributes.slice(0, 6) as attr}
                  <span class="attr-badge">
                    {attr.name.slice(0, 3)}<span class="attr-val">{attr.value}</span>
                  </span>
                {/each}
              </div>
            {/if}
          </div>
        {/each}
      </div>

      {#if error}
        <div class="error-banner">{error}</div>
      {/if}

      <div class="button-row">
        <button
          class="btn-wizard primary"
          on:click={nextStep}
          disabled={!canProceedStep1}
        >
          Next
        </button>
      </div>

    {:else if step === 2}
      <!-- Step 2: Name & Description -->
      <h2 class="step-title">Name Your Character</h2>
      <p class="step-description">
        Give your character a name and an optional description.
      </p>

      <div class="customize-layout">
        <div class="preview-card">
          <img src={getAvatar(characterName)} alt="" class="template-avatar" />
          <span class="preview-name">{characterName || "..."}</span>
          <span class="preview-template">{selectedTemplate.name}</span>
          {#if selectedTemplate.attributes}
            <div class="template-attrs">
              {#each selectedTemplate.attributes.slice(0, 6) as attr}
                <span class="attr-badge">
                  {attr.name.slice(0, 3)}<span class="attr-val">{attr.value}</span>
                </span>
              {/each}
            </div>
          {/if}
        </div>

        <div class="customize-form">
          <div class="form-group">
            <label class="form-label" for="char-name">Character Name</label>
            <input
              id="char-name"
              class="form-input"
              type="text"
              bind:value={characterName}
              placeholder="Enter a name..."
              on:keydown={handleNameKeydown}
              maxlength="40"
              autofocus
            />
          </div>

          <div class="form-group">
            <label class="form-label" for="char-desc">Description</label>
            <textarea
              id="char-desc"
              class="form-textarea"
              bind:value={characterDescription}
              placeholder="Describe your character..."
              maxlength="200"
              rows="3"
            ></textarea>
          </div>
        </div>
      </div>

      <div class="button-row">
        <button class="btn-wizard secondary" on:click={prevStep}>
          Back
        </button>
        <button
          class="btn-wizard primary"
          on:click={nextStep}
          disabled={!canProceedStep2}
        >
          Next
        </button>
      </div>

    {:else if step === 3}
      <!-- Step 3: Confirm -->
      <h2 class="step-title">Confirm Character</h2>
      <p class="step-description">
        Review your character before entering the game.
      </p>

      <div class="confirm-card">
        <img src={getAvatar(characterName)} alt="" class="confirm-avatar" />
        <span class="confirm-name">{characterName}</span>
        <span class="confirm-template">{selectedTemplate.name}</span>
        {#if characterDescription}
          <span class="confirm-desc">{characterDescription}</span>
        {/if}

        {#if selectedTemplate.attributes}
          <div class="confirm-attrs">
            {#each selectedTemplate.attributes as attr}
              <div class="confirm-attr">
                <span class="confirm-attr-name">{attr.name.slice(0, 3)}</span>
                <span class="confirm-attr-value">{attr.value}</span>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      {#if error}
        <div class="error-banner">{error}</div>
      {/if}

      <div class="button-row">
        <button class="btn-wizard secondary" on:click={prevStep} disabled={creating}>
          Back
        </button>
        <button
          class="btn-wizard primary"
          on:click={handleCreate}
          disabled={creating}
        >
          {creating ? 'Creating...' : 'Create Character'}
        </button>
      </div>
    {/if}
  </div>

  <!-- Success overlay -->
  {#if showSuccess}
    <div class="success-overlay">
      <span class="success-text">Character Created</span>
      <span class="success-subtext">Entering the game...</span>
      <div class="success-bar">
        <div class="success-bar-fill"></div>
      </div>
    </div>
  {/if}
</div>
