<script>
  import { createEventDispatcher, onDestroy, tick } from 'svelte';
  import { layoutStore } from './LayoutStore.js';
  import { widgetsEqual } from './layoutTemplates.js';

  const dispatch = createEventDispatcher();

  $: templates = $layoutStore.templates || [];
  $: activeId = $layoutStore.activeTemplateId;
  $: activeTpl = templates.find((t) => t.id === activeId) || null;
  $: dirty =
    !!activeTpl &&
    !widgetsEqual(
      ($layoutStore.widgets || []).map((item) => ({
        id: item.id,
        widgetType: item.widgetType,
        x: item[24]?.x ?? item.x,
        y: item[24]?.y ?? item.y,
        w: item[24]?.w ?? item.w,
        h: item[24]?.h ?? item.h,
        visible: item.visible ?? true,
        ...(item.widgetType === 'tabcontainer'
          ? { tabs: item.tabs || [], activeTabIndex: item.activeTabIndex || 0 }
          : {}),
      })),
      activeTpl.widgets
    );

  let showTemplatesMenu = false;
  let renameId = null;
  let renameValue = '';

  /** @type {null | { kind: 'save' } | { kind: 'confirm', title: string, message: string, confirmLabel: string, danger?: boolean, onConfirm: () => void }} */
  let dialog = null;
  let dialogName = '';
  let dialogError = '';
  let dialogInputEl = null;
  let escHandler = null;

  function bindEsc(active) {
    if (active && !escHandler) {
      escHandler = (e) => {
        if (e.key !== 'Escape') return;
        if (!dialog) return;
        e.preventDefault();
        e.stopPropagation();
        closeDialog();
      };
      if (typeof window !== 'undefined') window.addEventListener('keydown', escHandler, true);
    } else if (!active && escHandler) {
      if (typeof window !== 'undefined') window.removeEventListener('keydown', escHandler, true);
      escHandler = null;
    }
  }

  $: bindEsc(!!dialog);

  onDestroy(() => {
    if (escHandler && typeof window !== 'undefined') {
      window.removeEventListener('keydown', escHandler, true);
    }
    escHandler = null;
  });

  async function focusDialogInput() {
    await tick();
    if (dialogInputEl && typeof dialogInputEl.focus === 'function') {
      dialogInputEl.focus();
      if (typeof dialogInputEl.select === 'function') dialogInputEl.select();
    }
  }

  function closeDialog() {
    dialog = null;
    dialogName = '';
    dialogError = '';
    dialogInputEl = null;
  }

  function save() {
    layoutStore.exitEditMode(true);
  }

  function cancel() {
    layoutStore.exitEditMode(false);
  }

  function reset() {
    dialog = {
      kind: 'confirm',
      title: 'Reset layout',
      message: 'Reset layout to default? This will discard your custom layout.',
      confirmLabel: 'Reset',
      danger: true,
      onConfirm: () => {
        layoutStore.resetToDefault();
        closeDialog();
      },
    };
  }

  function openAddPanel() {
    dispatch('openAddPanel');
  }

  function saveAsTemplate() {
    dialogName = activeTpl?.name || '';
    dialogError = '';
    dialog = { kind: 'save' };
    focusDialogInput();
  }

  function submitSaveTemplate() {
    const result = layoutStore.saveAsTemplate(dialogName);
    if (!result) {
      dialogError = 'Please enter a template name.';
      focusDialogInput();
      return;
    }
    showTemplatesMenu = false;
    closeDialog();
  }

  function onSelectTemplate(e) {
    const id = e.target.value;
    if (!id) return;
    if (id === activeId && !dirty) return;
    layoutStore.applyTemplate(id);
    e.target.value = '';
    showTemplatesMenu = false;
  }

  function applyTemplate(id) {
    layoutStore.applyTemplate(id);
    showTemplatesMenu = false;
  }

  function startRename(tpl) {
    renameId = tpl.id;
    renameValue = tpl.name;
  }

  function commitRename() {
    if (renameId && renameValue.trim()) {
      layoutStore.renameTemplate(renameId, renameValue);
    }
    renameId = null;
    renameValue = '';
  }

  function deleteTpl(tpl) {
    dialog = {
      kind: 'confirm',
      title: 'Delete template',
      message: `Delete template "${tpl.name}"?`,
      confirmLabel: 'Delete',
      danger: true,
      onConfirm: () => {
        layoutStore.deleteTemplate(tpl.id);
        closeDialog();
      },
    };
  }

  function toggleMenu() {
    showTemplatesMenu = !showTemplatesMenu;
    renameId = null;
  }

  function onDialogBackdrop(e) {
    if (e.target === e.currentTarget) closeDialog();
  }
</script>

<style>
  .edit-toolbar {
    position: fixed;
    bottom: 24px;
    right: 24px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5em;
    padding: 0.75em;
    background: rgba(0, 0, 0, 0.9);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    z-index: 1000;
    animation: slideUp 0.3s ease-out;
    max-width: min(960px, calc(100vw - 32px));
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
    white-space: nowrap;
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

  .template-btn {
    background: rgba(168, 85, 247, 0.2);
    border-color: rgba(168, 85, 247, 0.5);
    color: #d8b4fe;
  }

  .template-btn:hover {
    background: rgba(168, 85, 247, 0.3);
    border-color: rgba(168, 85, 247, 0.7);
  }

  .divider {
    width: 1px;
    align-self: stretch;
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

  .active-chip {
    display: flex;
    align-items: center;
    gap: 0.35em;
    padding: 0.35em 0.7em;
    border-radius: 999px;
    background: rgba(245, 158, 11, 0.15);
    border: 1px solid rgba(245, 158, 11, 0.4);
    color: #fbbf24;
    font-size: 0.8em;
    font-weight: 600;
    max-width: 160px;
  }

  .active-chip.dirty {
    border-color: rgba(251, 191, 36, 0.8);
    color: #fde68a;
  }

  .active-chip span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .templates-wrap {
    position: relative;
  }

  .templates-panel {
    position: absolute;
    bottom: calc(100% + 10px);
    right: 0;
    width: min(320px, calc(100vw - 48px));
    background: rgba(12, 12, 16, 0.98);
    border: 1px solid rgba(255, 255, 255, 0.18);
    border-radius: 12px;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.55);
    padding: 0.75em;
    z-index: 1001;
  }

  .templates-panel h4 {
    margin: 0 0 0.5em;
    font-size: 0.85em;
    color: #e5e7eb;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .templates-panel select {
    width: 100%;
    margin-bottom: 0.6em;
    padding: 0.45em 0.5em;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(0, 0, 0, 0.5);
    color: #f3f4f6;
    font-size: 0.85em;
  }

  .tpl-list {
    list-style: none;
    margin: 0;
    padding: 0;
    max-height: 220px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.35em;
  }

  .tpl-row {
    display: flex;
    align-items: center;
    gap: 0.35em;
    padding: 0.4em 0.45em;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid transparent;
  }

  .tpl-row.active {
    border-color: rgba(245, 158, 11, 0.5);
    background: rgba(245, 158, 11, 0.1);
  }

  .tpl-name {
    flex: 1;
    min-width: 0;
    font-size: 0.85em;
    color: #e5e7eb;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    background: none;
    border: none;
    text-align: left;
    cursor: pointer;
    padding: 0;
  }

  .tpl-name:hover {
    color: #fbbf24;
  }

  .tpl-actions {
    display: flex;
    gap: 0.15em;
  }

  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(255, 255, 255, 0.05);
    color: #d1d5db;
    cursor: pointer;
    padding: 0;
  }

  .icon-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #fff;
  }

  .icon-btn.danger:hover {
    background: rgba(239, 68, 68, 0.25);
    color: #fca5a5;
    border-color: rgba(239, 68, 68, 0.4);
  }

  .icon-btn i {
    font-size: 1rem;
  }

  .rename-input {
    flex: 1;
    min-width: 0;
    padding: 0.25em 0.4em;
    border-radius: 6px;
    border: 1px solid rgba(168, 85, 247, 0.5);
    background: rgba(0, 0, 0, 0.4);
    color: #f3f4f6;
    font-size: 0.85em;
  }

  .empty-hint {
    font-size: 0.8em;
    color: #9ca3af;
    margin: 0.25em 0 0.5em;
  }

  .panel-footer {
    margin-top: 0.55em;
    display: flex;
    gap: 0.4em;
  }

  .panel-footer .toolbar-btn {
    flex: 1;
    justify-content: center;
    font-size: 0.8em;
    padding: 0.5em 0.6em;
  }

  .tpl-dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: 11000;
    background: rgba(0, 0, 0, 0.72);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1em;
    animation: tplOverlayIn 0.15s ease-out;
  }

  @keyframes tplOverlayIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .tpl-dialog {
    width: min(400px, calc(100vw - 32px));
    background: rgba(12, 16, 24, 0.97);
    border: 1px solid rgba(212, 175, 55, 0.35);
    border-radius: 12px;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.65), 0 0 0 1px rgba(212, 175, 55, 0.12);
    overflow: hidden;
    animation: tplDialogIn 0.2s ease-out;
  }

  @keyframes tplDialogIn {
    from { opacity: 0; transform: scale(0.96) translateY(8px); }
    to { opacity: 1; transform: scale(1) translateY(0); }
  }

  .tpl-dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75em;
    padding: 0.85em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: rgba(20, 26, 36, 0.92);
  }

  .tpl-dialog-title {
    display: flex;
    align-items: center;
    gap: 0.45em;
    color: #f8fafc;
    font-size: 0.95em;
    font-weight: 700;
    letter-spacing: 0.02em;
  }

  .tpl-dialog-title i {
    color: #fbbf24;
    font-size: 1.15em;
  }

  .tpl-dialog-close {
    border: none;
    background: transparent;
    color: #94a3b8;
    cursor: pointer;
    padding: 0.25em;
    border-radius: 6px;
    display: inline-flex;
    line-height: 1;
  }

  .tpl-dialog-close:hover {
    color: #e2e8f0;
    background: rgba(255, 255, 255, 0.08);
  }

  .tpl-dialog-body {
    padding: 1em;
    display: flex;
    flex-direction: column;
    gap: 0.75em;
  }

  .tpl-dialog-message {
    margin: 0;
    color: #cbd5e1;
    font-size: 0.9em;
    line-height: 1.45;
  }

  .tpl-dialog-label {
    margin: 0;
    color: #94a3b8;
    font-size: 0.8em;
  }

  .tpl-dialog-input {
    width: 100%;
    box-sizing: border-box;
    padding: 0.65em 0.75em;
    border-radius: 8px;
    border: 1px solid rgba(212, 175, 55, 0.35);
    background: rgba(0, 0, 0, 0.4);
    color: #f3f4f6;
    font: inherit;
    font-size: 0.95em;
  }

  .tpl-dialog-input:focus {
    outline: none;
    border-color: rgba(251, 191, 36, 0.75);
    box-shadow: 0 0 0 2px rgba(251, 191, 36, 0.15);
  }

  .tpl-dialog-error {
    margin: 0;
    padding: 0.45em 0.6em;
    border-radius: 6px;
    background: rgba(127, 29, 29, 0.45);
    border: 1px solid rgba(248, 113, 113, 0.4);
    color: #fecaca;
    font-size: 0.82em;
  }

  .tpl-dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5em;
    padding: 0 1em 1em;
  }

  .tpl-dialog-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.35em;
    padding: 0.55em 1em;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: rgba(255, 255, 255, 0.05);
    color: #e2e8f0;
    font-size: 0.88em;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s ease, border-color 0.15s ease;
  }

  .tpl-dialog-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .tpl-dialog-btn.primary {
    border-color: rgba(212, 175, 55, 0.55);
    background: rgba(212, 175, 55, 0.18);
    color: #fde68a;
  }

  .tpl-dialog-btn.primary:hover {
    background: rgba(212, 175, 55, 0.28);
    border-color: rgba(251, 191, 36, 0.7);
  }

  .tpl-dialog-btn.danger {
    border-color: rgba(239, 68, 68, 0.45);
    background: rgba(239, 68, 68, 0.18);
    color: #fca5a5;
  }

  .tpl-dialog-btn.danger:hover {
    background: rgba(239, 68, 68, 0.28);
    border-color: rgba(248, 113, 113, 0.65);
  }

  .tpl-dialog-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

</style>

<div class="edit-toolbar">
  <div class="edit-mode-label">
    <i class="material-icons">edit</i>
    Edit Mode
  </div>

  {#if activeTpl}
    <div class="active-chip" class:dirty title={dirty ? 'Unsaved edits vs template' : 'Active template'}>
      <i class="material-icons" style="font-size: 1em;">dashboard</i>
      <span>{activeTpl.name}{dirty ? ' *' : ''}</span>
    </div>
  {/if}

  <div class="divider"></div>

  <button class="toolbar-btn add-btn" on:click={openAddPanel}>
    <i class="material-icons">add</i>
    Add Widget
  </button>

  <button class="toolbar-btn template-btn" type="button" title="Stack room and terminal" on:click={() => layoutStore.applyPreset('compact', { lock: true })}>
    Compact
  </button>
  <button class="toolbar-btn template-btn" type="button" title="Room and terminal side by side" on:click={() => layoutStore.applyPreset('desktop', { lock: true })}>
    Desktop
  </button>
  <button class="toolbar-btn template-btn" type="button" title="Side by side, using the full height" on:click={() => layoutStore.applyPreset('wide', { lock: true })}>
    Wide
  </button>

  <button class="toolbar-btn template-btn" type="button" title="Undo the last layout change" on:click={() => layoutStore.undo()}>
    Undo
  </button>
  <button
    class="toolbar-btn template-btn"
    type="button"
    title={$layoutStore.layoutLocked ? 'Unlock dragging and resizing' : 'Lock the layout in place'}
    on:click={() => layoutStore.toggleLock()}
  >
    {$layoutStore.layoutLocked ? 'Unlock' : 'Lock'}
  </button>

  <button class="toolbar-btn reset-btn" on:click={reset}>
    <i class="material-icons">refresh</i>
    Reset
  </button>

  <div class="divider"></div>

  <div class="templates-wrap">
    <button class="toolbar-btn template-btn" on:click={toggleMenu} title="Layout templates">
      <i class="material-icons">view_quilt</i>
      Templates
      {#if templates.length}
        ({templates.length})
      {/if}
    </button>

    {#if showTemplatesMenu}
      <div class="templates-panel" role="dialog" aria-label="Layout templates">
        <h4>Personal templates</h4>
        {#if templates.length === 0}
          <p class="empty-hint">No templates yet. Arrange widgets, then Save as template.</p>
        {:else}
          <select on:change={onSelectTemplate} aria-label="Switch template">
            <option value="">Switch to…</option>
            {#each templates as tpl}
              <option value={tpl.id}>
                {tpl.name}{tpl.id === activeId ? (dirty ? ' (active *)' : ' (active)') : ''}
              </option>
            {/each}
          </select>
          <ul class="tpl-list">
            {#each templates as tpl}
              <li class="tpl-row" class:active={tpl.id === activeId}>
                {#if renameId === tpl.id}
                  <input
                    class="rename-input"
                    bind:value={renameValue}
                    on:keydown={(e) => e.key === 'Enter' && commitRename()}
                    on:blur={commitRename}
                  />
                {:else}
                  <button class="tpl-name" type="button" on:click={() => applyTemplate(tpl.id)}>
                    {tpl.name}{tpl.id === activeId && dirty ? ' *' : ''}
                  </button>
                {/if}
                <div class="tpl-actions">
                  <button class="icon-btn" type="button" title="Rename" on:click={() => startRename(tpl)}>
                    <i class="material-icons">edit</i>
                  </button>
                  <button class="icon-btn danger" type="button" title="Delete" on:click={() => deleteTpl(tpl)}>
                    <i class="material-icons">delete</i>
                  </button>
                </div>
              </li>
            {/each}
          </ul>
        {/if}
        <div class="panel-footer">
          <button class="toolbar-btn template-btn" type="button" on:click={saveAsTemplate}>
            <i class="material-icons">bookmark_add</i>
            Save as template…
          </button>
        </div>
      </div>
    {/if}
  </div>

  <button class="toolbar-btn template-btn" on:click={saveAsTemplate} title="Save current layout as a named template">
    <i class="material-icons">bookmark_add</i>
    Save as template…
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

{#if dialog}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="tpl-dialog-overlay" on:click={onDialogBackdrop} role="presentation">
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div
      class="tpl-dialog"
      role="dialog"
      aria-modal="true"
      aria-label={dialog.kind === 'save' ? 'Save as template' : dialog.title}
      on:click|stopPropagation
    >
      <div class="tpl-dialog-header">
        <span class="tpl-dialog-title">
          <i class="material-icons">{dialog.kind === 'save' ? 'bookmark_add' : 'warning'}</i>
          {dialog.kind === 'save' ? 'Save as template' : dialog.title}
        </span>
        <button class="tpl-dialog-close" type="button" on:click={closeDialog} aria-label="Close">
          <i class="material-icons">close</i>
        </button>
      </div>

      {#if dialog.kind === 'save'}
        <form class="tpl-dialog-body" on:submit|preventDefault={submitSaveTemplate}>
          <p class="tpl-dialog-label">Name for this layout template</p>
          <input
            class="tpl-dialog-input"
            type="text"
            bind:this={dialogInputEl}
            bind:value={dialogName}
            placeholder="e.g. Combat focus"
            maxlength="48"
            aria-label="Template name"
            autocomplete="off"
          />
          {#if dialogError}
            <p class="tpl-dialog-error">{dialogError}</p>
          {/if}
          <div class="tpl-dialog-actions" style="padding: 0;">
            <button class="tpl-dialog-btn" type="button" on:click={closeDialog}>Cancel</button>
            <button class="tpl-dialog-btn primary" type="submit">Save</button>
          </div>
        </form>
      {:else}
        <div class="tpl-dialog-body">
          <p class="tpl-dialog-message">{dialog.message}</p>
        </div>
        <div class="tpl-dialog-actions">
          <button class="tpl-dialog-btn" type="button" on:click={closeDialog}>Cancel</button>
          <button
            class="tpl-dialog-btn"
            class:danger={dialog.danger}
            class:primary={!dialog.danger}
            type="button"
            on:click={() => dialog.onConfirm && dialog.onConfirm()}
          >
            {dialog.confirmLabel}
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}

