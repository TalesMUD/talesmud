<script>
  import { createEventDispatcher } from 'svelte';
  import { layoutStore } from './LayoutStore.js';
  import { WIDGET_TYPES, getWidgetConfig, canBeTabChild } from './WidgetRegistry.js';

  export let widgetId;

  const dispatch = createEventDispatcher();

  // Read tab data reactively from the store
  $: widget = $layoutStore.widgets.find(w => w.id === widgetId);
  $: tabs = widget?.tabs || [];

  // Available widget types to add (exclude tabcontainer, respect maxInstances)
  $: availableTypes = Object.entries(WIDGET_TYPES)
    .filter(([type]) => canBeTabChild(type))
    .filter(([type]) => layoutStore.canAddWidget(type, WIDGET_TYPES))
    .map(([type, config]) => ({ type, ...config }));

  function addTab(widgetType) {
    layoutStore.addTabToContainer(widgetId, widgetType);
  }

  function removeTab(index) {
    layoutStore.removeTabFromContainer(widgetId, index);
  }

  function close() {
    dispatch('close');
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) {
      close();
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      close();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<style>
  .panel-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1001;
    animation: fadeIn 0.2s ease-out;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .panel {
    background: rgba(20, 20, 30, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 16px;
    padding: 1.5em;
    max-width: 420px;
    width: 90%;
    max-height: 80vh;
    overflow-y: auto;
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

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.5em;
    padding-bottom: 1em;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .panel-title {
    display: flex;
    align-items: center;
    gap: 0.5em;
    font-size: 1.2em;
    font-weight: 600;
    color: #e5e7eb;
  }

  .panel-title i {
    color: #f59e0b;
  }

  .close-btn {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.1);
    border: none;
    color: #9ca3af;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.2);
    color: #e5e7eb;
  }

  .section-header {
    font-size: 0.75em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #6b7280;
    margin-bottom: 0.75em;
    padding-left: 0.25em;
  }

  .current-tabs {
    margin-bottom: 1.5em;
  }

  .tab-item {
    display: flex;
    align-items: center;
    gap: 0.5em;
    padding: 0.6em 0.75em;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    margin-bottom: 0.5em;
  }

  .tab-item i.tab-icon {
    font-size: 18px;
    color: #f59e0b;
    width: 24px;
    text-align: center;
  }

  .tab-item .tab-name {
    flex: 1;
    font-size: 0.9em;
    color: #e5e7eb;
  }

  .tab-remove-btn {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: rgba(239, 68, 68, 0.2);
    border: none;
    color: #ef4444;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .tab-remove-btn:hover {
    background: rgba(239, 68, 68, 0.4);
  }

  .tab-remove-btn i {
    font-size: 16px;
  }

  .empty-tabs {
    padding: 1em;
    text-align: center;
    color: #6b7280;
    font-size: 0.85em;
    font-style: italic;
  }

  .add-section {
    margin-top: 0.5em;
  }

  .widget-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 0.5em;
  }

  .widget-card {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 0.75em 0.5em;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: center;
  }

  .widget-card:hover:not(.disabled) {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(245, 158, 11, 0.5);
    transform: translateY(-1px);
  }

  .widget-card-icon {
    width: 36px;
    height: 36px;
    border-radius: 6px;
    background: rgba(245, 158, 11, 0.15);
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 0.5em;
  }

  .widget-card-icon i {
    font-size: 18px;
    color: #f59e0b;
  }

  .widget-card-name {
    font-size: 0.8em;
    font-weight: 500;
    color: #e5e7eb;
  }

  .no-widgets {
    padding: 1em;
    text-align: center;
    color: #6b7280;
    font-size: 0.85em;
  }
</style>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="panel-backdrop" on:click={handleBackdropClick}>
  <div class="panel">
    <div class="panel-header">
      <div class="panel-title">
        <i class="material-icons">tab</i>
        Configure Tabs
      </div>
      <button class="close-btn" on:click={close}>
        <i class="material-icons">close</i>
      </button>
    </div>

    <div class="current-tabs">
      <div class="section-header">Current Tabs ({tabs.length})</div>
      {#if tabs.length > 0}
        {#each tabs as tab, i}
          {@const tabConfig = getWidgetConfig(tab.widgetType)}
          <div class="tab-item">
            <i class="material-icons tab-icon">{tabConfig?.icon || 'widgets'}</i>
            <span class="tab-name">{tabConfig?.name || tab.widgetType}</span>
            <button class="tab-remove-btn" on:click={() => removeTab(i)} title="Remove tab">
              <i class="material-icons">close</i>
            </button>
          </div>
        {/each}
      {:else}
        <div class="empty-tabs">No tabs added yet</div>
      {/if}
    </div>

    <div class="add-section">
      <div class="section-header">Add Widget</div>
      {#if availableTypes.length > 0}
        <div class="widget-grid">
          {#each availableTypes as wt}
            <div class="widget-card" on:click={() => addTab(wt.type)}>
              <div class="widget-card-icon">
                <i class="material-icons">{wt.icon}</i>
              </div>
              <div class="widget-card-name">{wt.name}</div>
            </div>
          {/each}
        </div>
      {:else}
        <div class="no-widgets">All widgets have been added</div>
      {/if}
    </div>
  </div>
</div>
