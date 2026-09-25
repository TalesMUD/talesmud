<script>
  import { layoutStore } from './LayoutStore.js';

  export let title = '';
  export let icon = '';
  export let widgetId = '';
  /** Title bar. False renders only the collapse / focus buttons. */
  export let showTitle = true;
  /** Inline in a tab bar. False floats the buttons over a scene widget. */
  export let inline = false;
  export let collapsed = false;
  /** 'tr' (default) or 'tl' when the top-right of the widget is already used. */
  export let corner = 'tr';

  $: focused = !!(widgetId && $layoutStore.focusId === widgetId);

  function onCollapse(event) {
    event.stopPropagation();
    event.preventDefault();
    if (widgetId) layoutStore.toggleCollapse(widgetId);
  }

  function onFocus(event) {
    event.stopPropagation();
    event.preventDefault();
    if (widgetId) layoutStore.toggleFocus(widgetId);
  }
</script>

<div
  class="game-panel-header widget-chrome"
  class:buttons-only={!showTitle}
  class:inline
  class:corner-tl={corner === 'tl'}
>
  {#if showTitle}
    {#if icon}<i class="material-icons" aria-hidden="true">{icon}</i>{/if}
    <span class="widget-chrome-title widget-title">{title}</span>
  {/if}
  <slot />
  {#if widgetId}
    <button
      type="button"
      class="chrome-btn"
      title={collapsed ? 'Expand panel' : 'Collapse panel'}
      aria-label={collapsed ? 'Expand panel' : 'Collapse panel'}
      on:click={onCollapse}
    >
      <i class="material-icons">{collapsed ? 'expand_more' : 'expand_less'}</i>
    </button>
    <button
      type="button"
      class="chrome-btn"
      title={focused ? 'Restore layout' : 'Focus this panel'}
      aria-label={focused ? 'Restore layout' : 'Focus this panel'}
      on:click={onFocus}
    >
      <i class="material-icons">{focused ? 'close_fullscreen' : 'open_in_full'}</i>
    </button>
  {/if}
</div>

<style>
  .widget-chrome {
    margin: 0 !important;
    flex-shrink: 0;
    min-height: 36px;
    padding: 0.35em 0.55em 0.35em 0.75em;
    gap: 0.35em;
    border-radius: 0;
    box-sizing: border-box;
  }

  .widget-chrome-title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .buttons-only {
    position: absolute;
    top: 8px;
    right: 8px;
    z-index: 20;
    width: auto;
    min-height: 0;
    padding: 0;
    background: transparent;
    border: none;
    box-shadow: none;
    gap: 4px;
  }

  .buttons-only.corner-tl {
    right: auto;
    left: 8px;
  }

  .buttons-only.inline {
    position: static;
    margin-left: auto;
    background: transparent;
    border: none;
    padding: 0 4px;
  }

  .chrome-btn,
  .widget-chrome :global(.chrome-font) {
    width: 28px;
    height: 28px;
    flex: 0 0 28px;
    border-radius: 6px;
    border: 1px solid rgba(251, 191, 36, 0.45);
    background: rgba(0, 0, 0, 0.55);
    color: #fbbf24;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    font-family: var(--font-display, 'Cinzel', serif);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
  }

  .chrome-btn i {
    font-size: 18px;
    opacity: 1;
  }

  .chrome-btn:hover,
  .widget-chrome :global(.chrome-font:hover) {
    background: rgba(251, 191, 36, 0.2);
  }
</style>
