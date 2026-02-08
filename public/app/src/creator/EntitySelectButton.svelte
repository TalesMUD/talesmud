<script>
  import { createEventDispatcher } from "svelte";
  import EntitySelectModal from "./EntitySelectModal.svelte";

  /** Currently selected entity ID */
  export let value = "";
  /** Array of entity objects to pick from (each must have .id and .name) */
  export let elements = [];
  /** Column definitions for the modal table (from tableColumns.js) */
  export let columns = [];
  /** Modal header title, e.g. "Select Room" */
  export let title = "Select";
  /** Placeholder text when nothing is selected */
  export let placeholder = "Select...";
  /** Allow clearing the selection back to "" */
  export let allowNone = true;

  const dispatch = createEventDispatcher();

  let modalOpen = false;

  // Find the currently selected element for display
  $: selectedElement = value ? elements.find((el) => el.id === value) : null;
  $: displayName = selectedElement ? selectedElement.name : "";
  $: displayId = value && value !== displayName ? value : "";

  function openModal() {
    modalOpen = true;
  }

  function handleSelect(event) {
    dispatch("change", event.detail);
    modalOpen = false;
  }

  function handleClose() {
    modalOpen = false;
  }

  function clear(e) {
    e.stopPropagation();
    dispatch("change", "");
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="esb-button" on:click={openModal}>
  <span class="material-symbols-outlined esb-icon">search</span>
  <div class="esb-text">
    {#if displayName}
      <span class="esb-name">{displayName}</span>
      {#if displayId}
        <span class="esb-id">{displayId.length > 12 ? displayId.slice(0, 10) + "\u2026" : displayId}</span>
      {/if}
    {:else if value}
      <span class="esb-name esb-raw">{value.length > 20 ? value.slice(0, 18) + "\u2026" : value}</span>
    {:else}
      <span class="esb-placeholder">{placeholder}</span>
    {/if}
  </div>
  {#if value && allowNone}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <span class="esb-clear" on:click={clear} title="Clear selection">
      <span class="material-symbols-outlined" style="font-size: 14px">close</span>
    </span>
  {/if}
</div>

<EntitySelectModal
  open={modalOpen}
  {title}
  {columns}
  {elements}
  {value}
  on:select={handleSelect}
  on:close={handleClose}
/>

<style>
  .esb-button {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 6px 10px;
    font-size: 13px;
    background: rgba(15, 23, 42, 0.5);
    border: 1px solid #334155;
    border-radius: 6px;
    color: #e2e8f0;
    cursor: pointer;
    transition: border-color 0.2s ease;
    min-height: 38px;
  }

  .esb-button:hover {
    border-color: #00bcd4;
  }

  .esb-icon {
    font-size: 16px;
    color: #64748b;
    flex-shrink: 0;
  }

  .esb-text {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 6px;
    overflow: hidden;
  }

  .esb-name {
    font-size: 13px;
    font-weight: 500;
    color: #e2e8f0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .esb-name.esb-raw {
    font-family: monospace;
    font-size: 11px;
    color: #94a3b8;
  }

  .esb-id {
    font-size: 10px;
    font-family: monospace;
    color: #64748b;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .esb-placeholder {
    font-size: 13px;
    color: #64748b;
    font-style: italic;
  }

  .esb-clear {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border-radius: 3px;
    color: #64748b;
    flex-shrink: 0;
    transition: all 0.15s ease;
  }

  .esb-clear:hover {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
  }
</style>
