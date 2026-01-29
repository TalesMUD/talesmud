<script>
  import { createEventDispatcher } from "svelte";
  import ItemViewModal from "./ItemViewModal.svelte";

  export let open = false;
  export let itemIds = [];  // Current item IDs in the room (array of strings)
  export let items = [];     // Full item objects (loaded from API)
  export let itemTemplates = [];  // Available templates for creating new items

  const dispatch = createEventDispatcher();

  // Local copy for editing
  let localItemIds = [];
  let selectedItem = null;
  let showItemView = false;
  let selectedTemplateId = "";
  let wasOpen = false;

  // Reset selection only when modal first opens
  $: if (open && !wasOpen) {
    wasOpen = true;
    selectedTemplateId = "";
  } else if (!open && wasOpen) {
    wasOpen = false;
  }

  // Keep localItemIds in sync with itemIds prop (for when new items are added)
  $: if (open) {
    localItemIds = itemIds ? [...itemIds] : [];
  }

  const close = () => {
    dispatch("close", localItemIds);
  };

  const handleKey = (event) => {
    if (!open) return;
    if (event.key === "Escape" && !showItemView) {
      close();
    }
  };

  function addItemFromTemplate() {
    if (!selectedTemplateId) return;
    dispatch("createFromTemplate", selectedTemplateId);
    selectedTemplateId = "";
  }

  function removeItem(itemId) {
    localItemIds = localItemIds.filter(id => id !== itemId);
  }

  function viewItem(item) {
    selectedItem = item;
    showItemView = true;
  }

  function closeItemView() {
    showItemView = false;
    selectedItem = null;
  }

  // Get item object by ID from the items array
  function getItemById(id) {
    return items.find(item => item.id === id);
  }

  // Quality color mapping
  const qualityColors = {
    normal: "#9ca3af",
    magic: "#3b82f6",
    rare: "#eab308",
    legendary: "#f97316",
    mythic: "#a855f7",
  };
</script>

<svelte:window on:keydown={handleKey} />

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="modal-backdrop" on:click|self={close}>
    <div class="modal-container">
      <div class="modal-header">
        <div class="header-text">
          <h2>Room Items</h2>
          <p class="header-subtitle">
            Manage items placed in this room.
          </p>
        </div>
        <button class="close-btn" type="button" on:click={close}>
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="modal-content">
        <!-- Add Item Section -->
        <div class="add-item-section">
          <h3 class="section-title">Add Item from Template</h3>
          <div class="add-item-row">
            <select
              class="template-select"
              bind:value={selectedTemplateId}
            >
              <option value="">Select a template...</option>
              {#each itemTemplates as template}
                <option value={template.id}>
                  {template.name}
                  {#if template.type}({template.type}){/if}
                </option>
              {/each}
            </select>
            <button
              class="add-btn"
              type="button"
              on:click={addItemFromTemplate}
              disabled={!selectedTemplateId}
            >
              <span class="material-symbols-outlined">add</span>
              Create & Add
            </button>
          </div>
          <p class="help-text">Creates a new item instance from the template and adds it to this room.</p>
        </div>

        <div class="divider"></div>

        <!-- Current Items -->
        <div class="items-section">
          <h3 class="section-title">
            Items in Room
            {#if localItemIds.length > 0}
              <span class="count-badge">{localItemIds.length}</span>
            {/if}
          </h3>

          {#if localItemIds.length === 0}
            <div class="empty-state">
              <span class="material-symbols-outlined empty-icon">inventory_2</span>
              <p>No items in this room.</p>
            </div>
          {:else}
            <div class="items-list">
              {#each localItemIds as itemId}
                {@const item = getItemById(itemId)}
                <div class="item-row">
                  {#if item}
                    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
                    <div class="item-info" on:click={() => viewItem(item)}>
                      <span
                        class="item-name"
                        style="color: {qualityColors[item.quality] || '#fff'}"
                      >
                        {item.name}
                        {#if item.instanceSuffix}
                          <span class="instance-suffix">-{item.instanceSuffix}</span>
                        {/if}
                      </span>
                      <span class="item-meta">
                        {#if item.type}
                          <span class="item-type">{item.type}</span>
                        {/if}
                        {#if item.quality && item.quality !== "normal"}
                          <span class="item-quality" style="color: {qualityColors[item.quality]}">{item.quality}</span>
                        {/if}
                      </span>
                    </div>
                    <div class="item-actions">
                      <button
                        class="view-btn"
                        type="button"
                        on:click={() => viewItem(item)}
                        title="View Details"
                      >
                        <span class="material-symbols-outlined">visibility</span>
                      </button>
                      <button
                        class="remove-btn"
                        type="button"
                        on:click={() => removeItem(itemId)}
                        title="Remove from Room"
                      >
                        <span class="material-symbols-outlined">delete</span>
                      </button>
                    </div>
                  {:else}
                    <div class="item-info unknown">
                      <span class="item-name">Unknown Item</span>
                      <span class="item-id">{itemId}</span>
                    </div>
                    <div class="item-actions">
                      <button
                        class="remove-btn"
                        type="button"
                        on:click={() => removeItem(itemId)}
                        title="Remove from Room"
                      >
                        <span class="material-symbols-outlined">delete</span>
                      </button>
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-done" type="button" on:click={close}>
          Done
        </button>
      </div>
    </div>
  </div>
{/if}

<ItemViewModal
  open={showItemView}
  item={selectedItem}
  on:close={closeItemView}
/>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    padding: 40px 20px;
    overflow-y: auto;
  }

  .modal-container {
    width: 100%;
    max-width: 600px;
    background: #1e1e1e;
    border: 1px solid #3a3a3a;
    border-radius: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid #3a3a3a;
    background: #252525;
  }

  .header-text h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #fff;
  }

  .header-subtitle {
    margin: 4px 0 0;
    font-size: 12px;
    color: #888;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #888;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .close-btn:hover {
    background: #333;
    color: #fff;
  }

  .modal-content {
    padding: 20px 24px;
    max-height: 60vh;
    overflow-y: auto;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: #888;
    margin: 0 0 12px;
  }

  .count-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    font-size: 10px;
    font-weight: 700;
    background: rgba(0, 188, 212, 0.2);
    border-radius: 9px;
    color: #00bcd4;
  }

  .add-item-section {
    margin-bottom: 16px;
  }

  .add-item-row {
    display: flex;
    gap: 10px;
  }

  .template-select {
    flex: 1;
    padding: 10px 12px;
    font-size: 13px;
    background: #2a2a2a;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    color: #fff;
  }

  .template-select:focus {
    outline: none;
    border-color: #00bcd4;
  }

  .add-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 16px;
    font-size: 12px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #fff;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.2s ease;
  }

  .add-btn:hover:not(:disabled) {
    background: #00a5bb;
  }

  .add-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .add-btn .material-symbols-outlined {
    font-size: 18px;
  }

  .help-text {
    margin: 8px 0 0;
    font-size: 11px;
    color: #666;
  }

  .divider {
    height: 1px;
    background: #3a3a3a;
    margin: 20px 0;
  }

  .empty-state {
    text-align: center;
    padding: 30px 20px;
    color: #666;
  }

  .empty-icon {
    font-size: 40px;
    margin-bottom: 10px;
    opacity: 0.5;
  }

  .empty-state p {
    margin: 0;
    font-size: 13px;
  }

  .items-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .item-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px;
    background: #252525;
    border: 1px solid #3a3a3a;
    border-radius: 8px;
    transition: border-color 0.2s ease;
  }

  .item-row:hover {
    border-color: #4a4a4a;
  }

  .item-info {
    flex: 1;
    min-width: 0;
    cursor: pointer;
  }

  .item-info.unknown {
    opacity: 0.6;
  }

  .item-name {
    display: block;
    font-size: 14px;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .instance-suffix {
    font-size: 11px;
    font-family: monospace;
    color: #666;
  }

  .item-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 2px;
  }

  .item-type, .item-quality {
    font-size: 10px;
    text-transform: capitalize;
  }

  .item-type {
    color: #888;
  }

  .item-id {
    font-size: 10px;
    font-family: monospace;
    color: #666;
  }

  .item-actions {
    display: flex;
    gap: 6px;
    margin-left: 12px;
  }

  .view-btn, .remove-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background: transparent;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    color: #888;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .view-btn:hover {
    background: rgba(0, 188, 212, 0.1);
    border-color: #00bcd4;
    color: #00bcd4;
  }

  .remove-btn:hover {
    background: rgba(244, 67, 54, 0.1);
    border-color: #f44336;
    color: #f44336;
  }

  .view-btn .material-symbols-outlined,
  .remove-btn .material-symbols-outlined {
    font-size: 18px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid #3a3a3a;
    background: #252525;
  }

  .btn-done {
    padding: 10px 24px;
    font-size: 13px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #fff;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-done:hover {
    background: #00a5bb;
  }
</style>
