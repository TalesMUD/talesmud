<script>
  import { createEventDispatcher } from "svelte";

  /** When true, the dialog is visible. */
  export let open = false;
  /** Short type label, for example "npc" or "room". */
  export let entityType = "entity";
  /** Display name of the thing being deleted. */
  export let entityName = "";
  /** Stored id. Shown in full. */
  export let entityId = "";
  /** Optional extra line, for a node path or a batch. */
  export let detail = "";
  export let title = "";
  export let confirmLabel = "Delete";
  /** Replaces the default delete warning. Ops dialogs pass their own line. An empty hint hides the line. */
  export let hint = "This removes it from the world. There is no undo.";
  /** "danger" keeps the red button. "ops" is for live actions. */
  export let tone = "danger";

  const dispatch = createEventDispatcher();

  $: heading = title || `Delete ${entityType}?`;

  function confirm() {
    dispatch("confirm");
  }

  function cancel() {
    dispatch("cancel");
  }

  function onKey(event) {
    if (!open) return;
    if (event.key === "Escape") {
      event.preventDefault();
      cancel();
    }
  }
</script>

<svelte:window on:keydown={onKey} />

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="cd-backdrop" on:click|self={cancel}>
    <div class="cd-card" role="dialog" aria-modal="true" aria-labelledby="cd-title">
      <h2 id="cd-title">{heading}</h2>
      <p class="cd-body">
        {#if entityName}
          <span class="cd-name">{entityName}</span>
        {/if}
        {#if entityId}
          <span class="cd-id">{entityId}</span>
        {/if}
        {#if !entityName && !entityId}
          This cannot be undone.
        {/if}
      </p>
      {#if detail}
        <p class="cd-detail">{detail}</p>
      {/if}
      {#if hint}
        <p class="cd-hint">{hint}</p>
      {/if}
      <div class="cd-actions">
        <button class="cd-cancel" type="button" on:click={cancel}>Cancel</button>
        <button class={tone === "ops" ? "cd-run" : "cd-delete"} type="button" on:click={confirm}>{confirmLabel}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .cd-backdrop {
    position: fixed;
    inset: 0;
    z-index: 260;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(4px);
    padding: 24px;
  }

  .cd-card {
    width: 100%;
    max-width: 420px;
    background: #1e1e1e;
    border: 1px solid #3a3a3a;
    border-radius: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    padding: 20px;
  }

  h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #fff;
  }

  .cd-body {
    margin: 12px 0 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .cd-name {
    color: #e2e8f0;
    font-size: 14px;
    font-weight: 600;
  }

  .cd-id {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 12px;
    color: #94a3b8;
    word-break: break-all;
  }

  .cd-detail,
  .cd-hint {
    margin: 10px 0 0;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.4;
  }

  .cd-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 18px;
  }

  button {
    border-radius: 6px;
    padding: 8px 14px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }

  .cd-cancel {
    background: transparent;
    color: #cbd5e1;
    border: 1px solid #475569;
  }

  .cd-cancel:hover {
    background: #333;
  }

  .cd-delete {
    background: #dc2626;
    color: #fff;
    border: 1px solid #b91c1c;
  }

  .cd-delete:hover {
    background: #ef4444;
  }

  .cd-run {
    background: #b45309;
    color: #fff;
    border: 1px solid #92400e;
  }

  .cd-run:hover {
    background: #d97706;
  }
</style>
