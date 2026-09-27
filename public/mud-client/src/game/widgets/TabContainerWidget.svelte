<script>
  import { layoutStore } from '../layout/LayoutStore.js';
  import { getWidgetConfig } from '../layout/WidgetRegistry.js';
  import { childWidgetComponents, getChildWidgetProps } from '../layout/WidgetComponents.js';
  import WidgetChrome from '../layout/WidgetChrome.svelte';

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
  $: collapsed = !!($layoutStore.widgets.find((w) => w.id === widget.id) || {}).collapsed;
  let moreOpen = false;

  function toggleMore() {
    moreOpen = !moreOpen;
  }

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
    background: transparent;
    overflow: hidden;
    border: none;
    box-shadow: none;
  }

  .more {
    flex: 0 0 auto;
    height: 36px;
    border: none;
    background: transparent;
    color: #f5e6c0;
    font-family: var(--font-display, 'Cinzel', serif);
    cursor: pointer;
  }

  .more-menu {
    position: absolute;
    top: 36px;
    right: 72px;
    z-index: 30;
    display: flex;
    flex-direction: column;
    min-width: 10rem;
    background: rgba(16, 12, 8, 0.98);
    border: 1px solid rgba(212, 164, 74, 0.55);
    border-radius: 8px;
  }

  .more-menu button {
    text-align: left;
    background: transparent;
    border: none;
    color: #f5e6c0;
    padding: 0.45rem 0.7rem;
    cursor: pointer;
  }

  .tab-bar {
    position: relative;
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    background: var(--panel-header-bg, rgba(0, 0, 0, 0.35));
    border-bottom: 1px solid var(--panel-header-border, rgba(180, 130, 60, 0.22));
    min-height: 36px;
    height: 36px;
    flex-shrink: 0;
    padding-right: 0.15rem;
  }

  .tab-bar :global(.widget-chrome.buttons-only) {
    position: static;
    height: 36px;
    min-height: 36px;
    margin-left: auto;
    align-items: center;
  }

  .tab-scroll {
    display: flex;
    align-items: stretch;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: thin;
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
    flex: 0 0 auto;
    height: 36px;
    padding: 0 0.8em;
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

  /* Hide child widget title chrome inside tabs — the tab bar already shows the name.
     !important needed because Svelte's double-hash scoping on child components
     gives them equal specificity (0,3,0) and they load later in the bundle.

     Title-only headers (shared game-panel / widget chrome) are fully suppressed.
     Hybrid headers that also host utility controls keep a slim toolbar: only the
     title text is hidden (Quest Log, Terminal X, future .widget-chrome-title). */
  .tab-pane :global(.widget-header),
  .tab-pane :global(.game-panel-header) {
    display: none !important;
  }

  .tab-pane :global(.widget-chrome-title),
  .tab-pane :global(.questlog-header h2),
  .tab-pane :global(.tx-title) {
    display: none !important;
  }

  .tab-pane :global(.questlog-header .header-title-row),
  .tab-pane :global(.tx-titlebar) {
    justify-content: flex-end;
  }

  .tab-pane :global(.questlog-header) {
    padding: 0.25em 0.5em;
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

<div class="tab-container" class:collapsed>
  <div class="tab-bar">
    <div class="tab-scroll">
      {#each tabs as tab, i}
        <button
          class="tab"
          class:active={i === safeActiveIndex}
          on:click={() => { moreOpen = false; switchTab(i); }}
        >
          {#if getWidgetConfig(tab.widgetType)?.icon}
            <i class="material-icons tab-icon">{getWidgetConfig(tab.widgetType).icon}</i>
          {/if}
          <span>{getWidgetConfig(tab.widgetType)?.name || tab.widgetType}</span>
        </button>
      {/each}
    </div>
    {#if tabs.length > 3}
      <button type="button" class="tab more" aria-expanded={moreOpen} on:click={toggleMore}>More</button>
      {#if moreOpen}
        <div class="more-menu" role="menu">
          {#each tabs as tab, i}
            <button type="button" role="menuitem" on:click={() => { moreOpen = false; switchTab(i); }}>
              {getWidgetConfig(tab.widgetType)?.name || tab.widgetType}
            </button>
          {/each}
        </div>
      {/if}
    {/if}
    <WidgetChrome widgetId={widget.id} {collapsed} showTitle={false} inline={true} />
  </div>

  {#if !collapsed}
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
  {/if}
</div>
