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
    background: var(--panel-bg);
    border-radius: var(--panel-radius);
    overflow: hidden;
    border: 1px solid var(--panel-border);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: var(--panel-shadow);
  }

  .tab-bar {
    display: flex;
    align-items: stretch;
    background: var(--panel-header-bg);
    border-bottom: 1px solid var(--panel-header-border);
    min-height: 38px;
    flex-shrink: 0;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: thin;
    scrollbar-color: var(--scrollbar-thumb) transparent;
  }

  .tab-bar::-webkit-scrollbar {
    height: 3px;
  }
  .tab-bar::-webkit-scrollbar-track {
    background: transparent;
  }
  .tab-bar::-webkit-scrollbar-thumb {
    background: var(--scrollbar-thumb);
    border-radius: 2px;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 0.35em;
    padding: 0.4em 0.8em;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-secondary);
    font-family: var(--font-display);
    font-size: var(--text-sm);
    font-weight: 500;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.15s ease;
  }

  .tab:hover {
    color: var(--text-primary);
    background: var(--tab-hover-bg);
  }

  .tab.active {
    color: var(--tab-active-color);
    border-bottom-color: var(--tab-active-border);
    background: var(--tab-active-bg);
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
     Also hide game-panel-header since we now use that class.
     !important needed because Svelte's double-hash scoping on child components
     gives them equal specificity (0,3,0) and they load later in the bundle. */
  .tab-pane :global(.widget-header),
  .tab-pane :global(.game-panel-header) {
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
    color: var(--text-dim);
    gap: 0.75em;
    padding: 1em;
    text-align: center;
  }

  .empty-state i {
    font-size: 2.5em;
    color: var(--text-dim);
  }

  .empty-state .hint {
    font-size: var(--text-sm);
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
