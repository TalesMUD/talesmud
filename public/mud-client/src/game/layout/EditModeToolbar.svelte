<script>
  import { createEventDispatcher } from 'svelte';
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

  function saveAsTemplate() {
    const suggested = activeTpl?.name || '';
    const name = prompt('Save current layout as template named:', suggested);
    if (name == null) return;
    const result = layoutStore.saveAsTemplate(name);
    if (!result) {
      alert('Please enter a template name.');
      return;
    }
    showTemplatesMenu = false;
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
    if (!confirm(`Delete template "${tpl.name}"?`)) return;
    layoutStore.deleteTemplate(tpl.id);
  }

  function toggleMenu() {
    showTemplatesMenu = !showTemplatesMenu;
    renameId = null;
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
