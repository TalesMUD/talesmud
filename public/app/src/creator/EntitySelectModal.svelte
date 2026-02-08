<script>
  import { createEventDispatcher } from "svelte";
  import DataTable from "./DataTable.svelte";

  export let open = false;
  export let title = "Select";
  export let columns = [];
  export let elements = [];
  export let value = "";

  const dispatch = createEventDispatcher();

  let filterValues = {};
  let sortKey = "";
  let sortDir = "asc";
  let highlightedElement = null;
  let lastClickTime = 0;
  let lastClickId = null;

  // Reset state when modal opens
  let wasOpen = false;
  $: if (open && !wasOpen) {
    wasOpen = true;
    highlightedElement = value ? elements.find((el) => el.id === value) || null : null;
    filterValues = {};
    sortKey = "";
    sortDir = "asc";
    lastClickTime = 0;
    lastClickId = null;
  } else if (!open && wasOpen) {
    wasOpen = false;
  }

  // Auto-populate select column options from data when not already set
  $: populatedColumns = columns.map((col) => {
    if (col.type === "select" && (!col.options || col.options.length === 0)) {
      const values = [
        ...new Set(
          elements
            .map((el) => {
              if (col.accessor) return String(col.accessor(el));
              const path = col.key.split(".");
              const v = path.reduce((acc, part) => acc?.[part], el);
              return v != null ? String(v) : "";
            })
            .filter((v) => v && v !== "-" && v !== "")
        ),
      ].sort();
      return { ...col, options: values };
    }
    return col;
  });

  function handleRowSelect(event) {
    const element = event.detail;
    const now = Date.now();

    // Double-click detection: select + close immediately
    if (lastClickId === element.id && now - lastClickTime < 400) {
      dispatch("select", element.id);
      dispatch("close");
    } else {
      highlightedElement = element;
    }

    lastClickTime = now;
    lastClickId = element.id;
  }

  function confirm() {
    if (highlightedElement) {
      dispatch("select", highlightedElement.id);
    }
    dispatch("close");
  }

  function cancel() {
    dispatch("close");
  }

  function handleFilterChange(event) {
    filterValues = { ...filterValues, [event.detail.key]: event.detail.value };
  }

  function handleResetFilters() {
    filterValues = {};
  }

  function handleSort(event) {
    sortKey = event.detail.key;
    sortDir = event.detail.dir;
  }

  function handleKey(event) {
    if (!open) return;
    if (event.key === "Escape") cancel();
    if (event.key === "Enter" && highlightedElement) confirm();
  }
</script>

<svelte:window on:keydown={handleKey} />

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="esm-backdrop" on:click|self={cancel}>
    <div class="esm-container">
      <div class="esm-header">
        <div class="esm-header-text">
          <h2>{title}</h2>
          <p class="esm-subtitle">
            Click a row to highlight, double-click to confirm.
          </p>
        </div>
        <button class="esm-close-btn" type="button" on:click={cancel}>
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="esm-content">
        <DataTable
          columns={populatedColumns}
          {elements}
          selectedElement={highlightedElement}
          {filterValues}
          {sortKey}
          {sortDir}
          compact={true}
          on:select={handleRowSelect}
          on:sort={handleSort}
          on:filterChange={handleFilterChange}
          on:resetFilters={handleResetFilters}
        />
      </div>

      <div class="esm-footer">
        <span class="esm-count">{elements.length} entries</span>
        <div class="esm-actions">
          <button class="esm-btn-cancel" type="button" on:click={cancel}>
            Cancel
          </button>
          <button
            class="esm-btn-select"
            type="button"
            on:click={confirm}
            disabled={!highlightedElement}
          >
            <span class="material-symbols-outlined" style="font-size: 16px">check</span>
            Select
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .esm-backdrop {
    position: fixed;
    inset: 0;
    z-index: 200;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(4px);
    padding: 40px 20px;
    overflow-y: auto;
  }

  .esm-container {
    width: 100%;
    max-width: 860px;
    background: #1e1e1e;
    border: 1px solid #3a3a3a;
    border-radius: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 80px);
  }

  .esm-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid #3a3a3a;
    background: #252525;
    flex-shrink: 0;
  }

  .esm-header-text h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #fff;
  }

  .esm-subtitle {
    margin: 2px 0 0;
    font-size: 11px;
    color: #888;
  }

  .esm-close-btn {
    background: transparent;
    border: none;
    color: #888;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .esm-close-btn:hover {
    background: #333;
    color: #fff;
  }

  .esm-content {
    padding: 16px 20px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .esm-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px;
    border-top: 1px solid #3a3a3a;
    background: #252525;
    flex-shrink: 0;
  }

  .esm-count {
    font-size: 11px;
    color: #666;
  }

  .esm-actions {
    display: flex;
    gap: 10px;
  }

  .esm-btn-cancel {
    padding: 8px 18px;
    font-size: 12px;
    font-weight: 500;
    background: transparent;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    color: #888;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .esm-btn-cancel:hover {
    border-color: #555;
    color: #ccc;
  }

  .esm-btn-select {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 8px 18px;
    font-size: 12px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #fff;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .esm-btn-select:hover:not(:disabled) {
    background: #00a5bb;
  }

  .esm-btn-select:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
