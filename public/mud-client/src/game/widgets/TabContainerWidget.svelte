<script>
  import { layoutStore } from '../layout/LayoutStore.js';
  import { getWidgetConfig } from '../layout/WidgetRegistry.js';
  import { childWidgetComponents, getChildWidgetProps } from '../layout/WidgetComponents.js';

  export let store;
  export let sendMessage;
  export let onTerminalReady = () => {};
  export let onTerminalInput = () => {};
  export let widget; // full widget data item (needs .id)

  // Reactively read tab data from layout store
  let tabs = [];
  let activeTabIndex = 0;

  $: {
    const w = $layoutStore.widgets.find(w => w.id === widget.id);
    if (w) {
      tabs = w.tabs || [];
      activeTabIndex = w.activeTabIndex || 0;
    }
  }

  // Clamp activeTabIndex to valid range
  $: safeActiveIndex = tabs.length > 0 ? Math.min(activeTabIndex, tabs.length - 1) : 0;

  // Props deps for child widgets
  $: propDeps = { store, sendMessage, onTerminalReady, onTerminalInput };

  function switchTab(index) {
    layoutStore.setActiveTab(widget.id, index);
  }

  function getComponent(widgetType) {
    return childWidgetComponents[widgetType] || null;
  }
</script>

<style>
  .tab-container {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
    background: rgba(15, 15, 25, 0.95);
    border-radius: 8px;
    overflow: hidden;
  }

  .tab-bar {
    display: flex;
    align-items: stretch;
    background: rgba(0, 0, 0, 0.4);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    min-height: 36px;
    flex-shrink: 0;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.2) transparent;
  }

  .tab-bar::-webkit-scrollbar {
    height: 3px;
  }
  .tab-bar::-webkit-scrollbar-track {
    background: transparent;
  }
  .tab-bar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 2px;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 0.35em;
    padding: 0.4em 0.75em;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: #9ca3af;
    font-size: 0.8em;
    font-weight: 500;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.15s ease;
  }

  .tab:hover {
    color: #e5e7eb;
    background: rgba(255, 255, 255, 0.05);
  }

  .tab.active {
    color: #f59e0b;
    border-bottom-color: #f59e0b;
    background: rgba(245, 158, 11, 0.08);
  }

  .tab i.tab-icon {
    font-size: 1em;
  }

  .tab-content {
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  /* Hide child widget headers inside tabs — the tab bar already shows the name.
     !important needed because Svelte's double-hash scoping on child components
     gives them equal specificity (0,3,0) and they load later in the bundle. */
  .tab-pane :global(.widget-header) {
    display: none !important;
  }

  .tab-pane {
    width: 100%;
    height: 100%;
    position: absolute;
    top: 0;
    left: 0;
  }

  .tab-pane.hidden {
    visibility: hidden;
    pointer-events: none;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #6b7280;
    gap: 0.75em;
    padding: 1em;
    text-align: center;
  }

  .empty-state i {
    font-size: 2.5em;
    color: #4b5563;
  }

  .empty-state .hint {
    font-size: 0.85em;
    max-width: 200px;
  }
</style>

<div class="tab-container">
  <div class="tab-bar">
    {#each tabs as tab, i}
      <button
        class="tab"
        class:active={i === safeActiveIndex}
        on:click={() => switchTab(i)}
      >
        {#if getWidgetConfig(tab.widgetType)?.icon}
          <i class="material-icons tab-icon">{getWidgetConfig(tab.widgetType).icon}</i>
        {/if}
        <span>{getWidgetConfig(tab.widgetType)?.name || tab.widgetType}</span>
      </button>
    {/each}
  </div>

  <div class="tab-content">
    {#if tabs.length === 0}
      <div class="empty-state">
        <i class="material-icons">tab</i>
        <div class="hint">
          Use the <i class="material-icons" style="font-size: 1em; vertical-align: middle;">settings</i> button to add widgets
        </div>
      </div>
    {:else}
      {#each tabs as tab, i (tab.id)}
        <div class="tab-pane" class:hidden={i !== safeActiveIndex}>
          {#if getComponent(tab.widgetType)}
            <svelte:component
              this={getComponent(tab.widgetType)}
              {...getChildWidgetProps(tab.widgetType, propDeps)}
            />
          {:else}
            <div>Unknown widget: {tab.widgetType}</div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</div>
