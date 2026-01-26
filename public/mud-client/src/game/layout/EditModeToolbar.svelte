<script>
  import { createEventDispatcher } from 'svelte';
  import { layoutStore } from './LayoutStore.js';

  const dispatch = createEventDispatcher();

  function save() {
    layoutStore.exitEditMode(true);
  }

  function cancel() {
    layoutStore.exitEditMode(false);
  }

  function reset() {
    if (confirm('Reset layout to default? This will discard your custom layout.')) {
      layoutStore.resetToDefault();
    }
  }

  function openAddPanel() {
    dispatch('openAddPanel');
  }
</script>

<style>
  .edit-toolbar {
    position: fixed;
    bottom: 24px;
    right: 24px;
    display: flex;
    gap: 0.5em;
    padding: 0.75em;
    background: rgba(0, 0, 0, 0.9);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    z-index: 1000;
    animation: slideUp 0.3s ease-out;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .toolbar-btn {
    display: flex;
    align-items: center;
    gap: 0.4em;
    padding: 0.6em 1em;
    border-radius: 8px;
    border: 1px solid transparent;
    font-size: 0.9em;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .toolbar-btn i {
    font-size: 1.1em;
  }

  .add-btn {
    background: rgba(59, 130, 246, 0.2);
    border-color: rgba(59, 130, 246, 0.5);
    color: #93c5fd;
  }

  .add-btn:hover {
    background: rgba(59, 130, 246, 0.3);
    border-color: rgba(59, 130, 246, 0.7);
  }

  .reset-btn {
    background: rgba(107, 114, 128, 0.2);
    border-color: rgba(107, 114, 128, 0.5);
    color: #d1d5db;
  }

  .reset-btn:hover {
    background: rgba(107, 114, 128, 0.3);
    border-color: rgba(107, 114, 128, 0.7);
  }

  .cancel-btn {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.5);
    color: #fca5a5;
  }

  .cancel-btn:hover {
    background: rgba(239, 68, 68, 0.3);
    border-color: rgba(239, 68, 68, 0.7);
  }

  .save-btn {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.5);
    color: #86efac;
  }

  .save-btn:hover {
    background: rgba(34, 197, 94, 0.3);
    border-color: rgba(34, 197, 94, 0.7);
  }

  .divider {
    width: 1px;
    background: rgba(255, 255, 255, 0.1);
    margin: 0 0.25em;
  }

  .edit-mode-label {
    display: flex;
    align-items: center;
    gap: 0.4em;
    padding: 0.6em;
    color: #f59e0b;
    font-size: 0.85em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .edit-mode-label i {
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
</style>

<div class="edit-toolbar">
  <div class="edit-mode-label">
    <i class="material-icons">edit</i>
    Edit Mode
  </div>

  <div class="divider"></div>

  <button class="toolbar-btn add-btn" on:click={openAddPanel}>
    <i class="material-icons">add</i>
    Add Widget
  </button>

  <button class="toolbar-btn reset-btn" on:click={reset}>
    <i class="material-icons">refresh</i>
    Reset
  </button>

  <div class="divider"></div>

  <button class="toolbar-btn cancel-btn" on:click={cancel}>
    <i class="material-icons">close</i>
    Cancel
  </button>

  <button class="toolbar-btn save-btn" on:click={save}>
    <i class="material-icons">save</i>
    Save
  </button>
</div>
