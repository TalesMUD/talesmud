<script>
  import { createEventDispatcher } from 'svelte';
  import { layoutStore } from './LayoutStore.js';
  import { WIDGET_TYPES, getWidgetTypeList, WIDGET_CATEGORIES } from './WidgetRegistry.js';

  const dispatch = createEventDispatcher();

  $: availableWidgets = getWidgetTypeList().map(widget => ({
    ...widget,
    canAdd: layoutStore.canAddWidget(widget.type, WIDGET_TYPES)
  }));

  // Group by category
  $: groupedWidgets = Object.entries(WIDGET_CATEGORIES).map(([key, category]) => ({
    key,
    ...category,
    widgets: availableWidgets.filter(w => w.category === key)
  }));

  function addWidget(widgetType) {
    const config = WIDGET_TYPES[widgetType];
    layoutStore.addWidget(widgetType, config);
    close();
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
    max-width: 500px;
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

  .category {
    margin-bottom: 1.5em;
  }

  .category:last-child {
    margin-bottom: 0;
  }

  .category-header {
    font-size: 0.75em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #6b7280;
    margin-bottom: 0.75em;
    padding-left: 0.25em;
  }

  .widget-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 0.75em;
  }

  .widget-card {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 1em;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: center;
  }

  .widget-card:hover:not(.disabled) {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(245, 158, 11, 0.5);
    transform: translateY(-2px);
  }

  .widget-card.disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .widget-icon {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    background: rgba(245, 158, 11, 0.15);
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 0.75em;
  }

  .widget-icon i {
    font-size: 24px;
    color: #f59e0b;
  }

  .widget-card.disabled .widget-icon {
    background: rgba(107, 114, 128, 0.15);
  }

  .widget-card.disabled .widget-icon i {
    color: #6b7280;
  }

  .widget-name {
    font-weight: 600;
    color: #e5e7eb;
    margin-bottom: 0.25em;
    font-size: 0.9em;
  }

  .widget-description {
    font-size: 0.75em;
    color: #9ca3af;
    line-height: 1.4;
  }

  .widget-card.disabled .widget-name,
  .widget-card.disabled .widget-description {
    color: #6b7280;
  }

  .already-added {
    font-size: 0.7em;
    color: #f59e0b;
    margin-top: 0.5em;
  }
</style>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="panel-backdrop" on:click={handleBackdropClick}>
  <div class="panel">
    <div class="panel-header">
      <div class="panel-title">
        <i class="material-icons">widgets</i>
        Add Widget
      </div>
      <button class="close-btn" on:click={close}>
        <i class="material-icons">close</i>
      </button>
    </div>

    {#each groupedWidgets as category}
      <div class="category">
        <div class="category-header">{category.name}</div>
        <div class="widget-grid">
          {#each category.widgets as widget}
            <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
            <div
              class="widget-card"
              class:disabled={!widget.canAdd}
              on:click={() => widget.canAdd && addWidget(widget.type)}
            >
              <div class="widget-icon">
                <i class="material-icons">{widget.icon}</i>
              </div>
              <div class="widget-name">{widget.name}</div>
              <div class="widget-description">{widget.description}</div>
              {#if !widget.canAdd}
                <div class="already-added">Already added</div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {/each}
  </div>
</div>
