<script>
  import { createEventDispatcher } from "svelte";

  export let dialog = null;
  export let selectedNode = null;
  export let saving = false;

  const dispatch = createEventDispatcher();

  // Generate a short random ID with a prefix
  function generateShortId(prefix) {
    const randomPart = Math.floor(Math.random() * 10000000).toString().padStart(7, '0');
    return `${prefix}_${randomPart}`;
  }

  // Track if viewing the full dialog or a specific node
  $: isViewingNode = selectedNode !== null && selectedNode !== dialog;
  $: displayNode = isViewingNode ? selectedNode : dialog;

  function handleSave() {
    dispatch("save", dialog);
  }

  function handleClose() {
    dispatch("close");
  }

  function handleNodeClick(node) {
    dispatch("selectNode", node);
  }

  function handleBackToRoot() {
    dispatch("selectNode", dialog);
  }

  function addOption() {
    if (!displayNode) return;
    if (!displayNode.options) displayNode.options = [];
    displayNode.options = [
      ...displayNode.options,
      {
        nodeId: generateShortId("Q"),
        text: "New option",
        options: [],
      }
    ];
    dialog = dialog; // Trigger reactivity
  }

  function removeOption(index) {
    if (!displayNode?.options) return;
    displayNode.options = displayNode.options.filter((_, i) => i !== index);
    dialog = dialog;
  }

  function addAlternateText() {
    if (!displayNode) return;
    if (!displayNode.alternateTexts) displayNode.alternateTexts = [];
    displayNode.alternateTexts = [...displayNode.alternateTexts, ""];
    dialog = dialog;
  }

  function removeAlternateText(index) {
    if (!displayNode?.alternateTexts) return;
    displayNode.alternateTexts = displayNode.alternateTexts.filter((_, i) => i !== index);
    dialog = dialog;
  }

  function toggleShowOnlyOnce() {
    if (!displayNode) return;
    displayNode.show_only_once = !displayNode.show_only_once;
    dialog = dialog;
  }

  function toggleIsDialogExit() {
    if (!displayNode) return;
    displayNode.is_dialog_exit = !displayNode.is_dialog_exit;
    dialog = dialog;
  }

  function toggleOrderedTexts() {
    if (!displayNode) return;
    displayNode.orderedTexts = !displayNode.orderedTexts;
    dialog = dialog;
  }

  function addAnswer() {
    if (!displayNode) return;
    displayNode.answer = {
      nodeId: generateShortId("A"),
      text: "NPC response...",
      options: [],
    };
    dialog = dialog;
  }

  function removeAnswer() {
    if (!displayNode) return;
    displayNode.answer = null;
    dialog = dialog;
  }

  function addRequiredDialog() {
    if (!displayNode) return;
    if (!displayNode.requires_visited_dialogs) displayNode.requires_visited_dialogs = [];
    displayNode.requires_visited_dialogs = [...displayNode.requires_visited_dialogs, ""];
    dialog = dialog;
  }

  function removeRequiredDialog(index) {
    if (!displayNode?.requires_visited_dialogs) return;
    displayNode.requires_visited_dialogs = displayNode.requires_visited_dialogs.filter((_, i) => i !== index);
    dialog = dialog;
  }
</script>

<div class="detail-panel" class:open={dialog !== null}>
  {#if dialog}
    <div class="panel-header">
      <div class="header-title">
        {#if isViewingNode}
          <button class="back-btn" on:click={handleBackToRoot} title="Back to root">
            <span class="material-symbols-outlined">arrow_back</span>
          </button>
          <span class="node-path">
            <span class="path-root">{dialog.name || dialog.nodeId}</span>
            <span class="path-separator">/</span>
            <span class="path-current">{displayNode.nodeId}</span>
          </span>
        {:else}
          <span>{dialog.name || "Dialog"}</span>
        {/if}
      </div>
      <button class="close-btn" on:click={handleClose}>
        <span class="material-symbols-outlined">close</span>
      </button>
    </div>

    <div class="panel-content">
      <!-- Basic Info -->
      <section class="section">
        <h4 class="section-title">Basic Info</h4>

        {#if !isViewingNode}
          <div class="field">
            <label for="dialog-name">Dialog Name</label>
            <input id="dialog-name" type="text" bind:value={dialog.name} placeholder="e.g., Guard Greeting" />
          </div>
        {/if}

        <div class="field">
          <label for="dialog-node-id">Node ID</label>
          <input id="dialog-node-id" type="text" bind:value={displayNode.nodeId} placeholder="e.g., main, greeting" />
        </div>

        <div class="field">
          <label for="dialog-text">Text</label>
          <textarea id="dialog-text" bind:value={displayNode.text} rows="3" placeholder="What the NPC says..."></textarea>
        </div>
      </section>

      <!-- Flags -->
      <section class="section">
        <h4 class="section-title">Behavior Flags</h4>

        <div class="flags-row">
          <label class="flag-toggle">
            <input type="checkbox" checked={displayNode.show_only_once} on:change={toggleShowOnlyOnce} />
            <span>Show Only Once</span>
          </label>

          <label class="flag-toggle">
            <input type="checkbox" checked={displayNode.is_dialog_exit} on:change={toggleIsDialogExit} />
            <span>Dialog Exit</span>
          </label>

          <label class="flag-toggle">
            <input type="checkbox" checked={displayNode.orderedTexts} on:change={toggleOrderedTexts} />
            <span>Ordered Texts</span>
          </label>
        </div>
      </section>

      <!-- Alternate Texts -->
      <section class="section">
        <div class="section-header">
          <h4 class="section-title">Alternate Texts</h4>
          <button class="add-btn" on:click={addAlternateText}>+ Add</button>
        </div>
        <p class="section-hint">Random variations of the main text</p>

        {#if displayNode.alternateTexts?.length > 0}
          <div class="list">
            {#each displayNode.alternateTexts as text, index}
              <div class="list-item">
                <input type="text" bind:value={displayNode.alternateTexts[index]} placeholder="Alternate text..." />
                <button class="remove-btn" on:click={() => removeAlternateText(index)}>
                  <span class="material-symbols-outlined">delete</span>
                </button>
              </div>
            {/each}
          </div>
        {:else}
          <p class="empty-hint">No alternate texts</p>
        {/if}
      </section>

      <!-- Required Dialogs -->
      <section class="section">
        <div class="section-header">
          <h4 class="section-title">Required Visited Dialogs</h4>
          <button class="add-btn" on:click={addRequiredDialog}>+ Add</button>
        </div>
        <p class="section-hint">This option only shows if these nodes were visited</p>

        {#if displayNode.requires_visited_dialogs?.length > 0}
          <div class="list">
            {#each displayNode.requires_visited_dialogs as reqId, index}
              <div class="list-item">
                <input type="text" bind:value={displayNode.requires_visited_dialogs[index]} placeholder="Node ID..." />
                <button class="remove-btn" on:click={() => removeRequiredDialog(index)}>
                  <span class="material-symbols-outlined">delete</span>
                </button>
              </div>
            {/each}
          </div>
        {:else}
          <p class="empty-hint">No requirements</p>
        {/if}
      </section>

      <!-- Answer Node -->
      <section class="section">
        <div class="section-header">
          <h4 class="section-title">Auto-Response (Answer)</h4>
          {#if !displayNode.answer}
            <button class="add-btn" on:click={addAnswer}>+ Add Answer</button>
          {:else}
            <button class="remove-btn-text" on:click={removeAnswer}>Remove</button>
          {/if}
        </div>
        <p class="section-hint">Automatic NPC response after this option is selected</p>

        {#if displayNode.answer}
          <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
          <div class="answer-card" on:click={() => handleNodeClick(displayNode.answer)}>
            <div class="answer-header">
              <span class="answer-badge">ANSWER</span>
              <span class="answer-id">{displayNode.answer.nodeId}</span>
            </div>
            <p class="answer-preview">{displayNode.answer.text?.slice(0, 60) || "(no text)"}...</p>
            <span class="click-hint">Click to edit</span>
          </div>
        {/if}
      </section>

      <!-- Options (Child Nodes) -->
      <section class="section">
        <div class="section-header">
          <h4 class="section-title">Dialog Options</h4>
          <button class="add-btn" on:click={addOption}>+ Add Option</button>
        </div>
        <p class="section-hint">Player choices that branch the conversation</p>

        {#if displayNode.options?.length > 0}
          <div class="options-list">
            {#each displayNode.options as option, index}
              <div class="option-card">
                <div class="option-header">
                  <span class="option-number">{index + 1}</span>
                  <input type="text" bind:value={option.nodeId} placeholder="Node ID" class="option-id-input" />
                  <button class="remove-btn" on:click={() => removeOption(index)}>
                    <span class="material-symbols-outlined">delete</span>
                  </button>
                </div>
                <input type="text" bind:value={option.text} placeholder="Option text..." class="option-text-input" />
                <button class="edit-option-btn" on:click={() => handleNodeClick(option)}>
                  Edit Sub-Dialog
                  <span class="material-symbols-outlined">arrow_forward</span>
                </button>
              </div>
            {/each}
          </div>
        {:else}
          <p class="empty-hint">No options - this is a leaf node</p>
        {/if}
      </section>
    </div>

    <div class="panel-footer">
      <button class="cancel-btn" on:click={handleClose} disabled={saving}>
        Cancel
      </button>
      <button class="save-btn" on:click={handleSave} disabled={saving}>
        {saving ? "Saving..." : "Save Dialog"}
      </button>
    </div>
  {/if}
</div>

<style>
  .detail-panel {
    width: 0;
    height: 100%;
    background: #1e1e1e;
    border-left: 1px solid #333;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    transition: width 0.3s ease;
  }

  .detail-panel.open {
    width: 380px;
    min-width: 380px;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
    background: #252525;
    border-bottom: 1px solid #333;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    color: #fff;
  }

  .back-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: #333;
    border: none;
    border-radius: 4px;
    color: #aaa;
    cursor: pointer;
  }

  .back-btn:hover {
    background: #444;
    color: #fff;
  }

  .node-path {
    font-size: 13px;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .path-root {
    color: #888;
  }

  .path-separator {
    color: #555;
  }

  .path-current {
    color: #00bcd4;
    font-weight: 600;
  }

  .close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: #888;
    cursor: pointer;
  }

  .close-btn:hover {
    background: #333;
    color: #fff;
  }

  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }

  .section {
    margin-bottom: 24px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .section-title {
    font-size: 12px;
    font-weight: 600;
    color: #00bcd4;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 0 0 8px;
  }

  .section-hint {
    font-size: 11px;
    color: #666;
    margin: 0 0 12px;
  }

  .field {
    margin-bottom: 12px;
  }

  .field label {
    display: block;
    font-size: 11px;
    font-weight: 500;
    color: #888;
    margin-bottom: 4px;
  }

  .field input,
  .field textarea {
    width: 100%;
    padding: 8px 10px;
    background: #2a2a2a;
    border: 1px solid #444;
    border-radius: 4px;
    color: #fff;
    font-size: 13px;
    font-family: inherit;
  }

  .field input:focus,
  .field textarea:focus {
    outline: none;
    border-color: #00bcd4;
  }

  .field textarea {
    resize: vertical;
    min-height: 60px;
  }

  .flags-row {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .flag-toggle {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #aaa;
    cursor: pointer;
  }

  .flag-toggle input {
    accent-color: #00bcd4;
  }

  .add-btn {
    font-size: 11px;
    padding: 4px 10px;
    background: #333;
    border: 1px solid #555;
    border-radius: 4px;
    color: #00bcd4;
    cursor: pointer;
  }

  .add-btn:hover {
    background: #444;
  }

  .remove-btn-text {
    font-size: 11px;
    padding: 4px 10px;
    background: transparent;
    border: none;
    color: #f44336;
    cursor: pointer;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .list-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .list-item input {
    flex: 1;
    padding: 6px 8px;
    background: #2a2a2a;
    border: 1px solid #444;
    border-radius: 4px;
    color: #fff;
    font-size: 12px;
  }

  .remove-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: transparent;
    border: none;
    color: #f44336;
    cursor: pointer;
    border-radius: 4px;
  }

  .remove-btn:hover {
    background: rgba(244, 67, 54, 0.1);
  }

  .remove-btn .material-symbols-outlined {
    font-size: 18px;
  }

  .empty-hint {
    font-size: 12px;
    color: #555;
    font-style: italic;
  }

  .answer-card {
    background: #1a3a4a;
    border: 1px solid #2196f3;
    border-radius: 6px;
    padding: 12px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .answer-card:hover {
    background: #1e4a5e;
  }

  .answer-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
  }

  .answer-badge {
    font-size: 9px;
    font-weight: 700;
    padding: 2px 6px;
    background: #2196f3;
    border-radius: 3px;
    color: #fff;
  }

  .answer-id {
    font-size: 12px;
    color: #888;
  }

  .answer-preview {
    font-size: 12px;
    color: #aaa;
    margin: 0;
    line-height: 1.4;
  }

  .click-hint {
    font-size: 10px;
    color: #2196f3;
    display: block;
    margin-top: 8px;
  }

  .options-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .option-card {
    background: #2a2a2a;
    border: 1px solid #444;
    border-radius: 6px;
    padding: 12px;
  }

  .option-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }

  .option-number {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    background: #ff9800;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 700;
    color: #000;
  }

  .option-id-input {
    flex: 1;
    padding: 4px 8px;
    background: #333;
    border: 1px solid #555;
    border-radius: 4px;
    color: #fff;
    font-size: 11px;
  }

  .option-text-input {
    width: 100%;
    padding: 6px 8px;
    background: #333;
    border: 1px solid #555;
    border-radius: 4px;
    color: #fff;
    font-size: 12px;
    margin-bottom: 8px;
  }

  .edit-option-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    width: 100%;
    padding: 6px;
    background: transparent;
    border: 1px dashed #555;
    border-radius: 4px;
    color: #888;
    font-size: 11px;
    cursor: pointer;
  }

  .edit-option-btn:hover {
    border-color: #00bcd4;
    color: #00bcd4;
  }

  .edit-option-btn .material-symbols-outlined {
    font-size: 14px;
  }

  .panel-footer {
    display: flex;
    gap: 12px;
    padding: 16px;
    background: #252525;
    border-top: 1px solid #333;
  }

  .cancel-btn,
  .save-btn {
    flex: 1;
    padding: 10px 16px;
    font-size: 13px;
    font-weight: 500;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .cancel-btn {
    background: #333;
    border: 1px solid #555;
    color: #aaa;
  }

  .cancel-btn:hover {
    background: #444;
    color: #fff;
  }

  .save-btn {
    background: #00bcd4;
    border: none;
    color: #000;
  }

  .save-btn:hover {
    background: #00a5bb;
  }

  .save-btn:disabled,
  .cancel-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
