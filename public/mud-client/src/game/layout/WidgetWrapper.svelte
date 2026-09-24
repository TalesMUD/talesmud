<script>
  import { createEventDispatcher } from 'svelte';
  import { getWidgetConfig } from './WidgetRegistry.js';
  import WidgetChrome from './WidgetChrome.svelte';

  export let widget;
  export let editMode = false;
  /** False while the layout lock is on: edit chrome stays, drag and resize do not. */
  export let canResize = true;
  export let resizePointerDown = null;

  const dispatch = createEventDispatcher();

  const TITLE_CHROME = new Set([
    'terminal', 'terminalx', 'inventory', 'equipment', 'character', 'questlog', 'tabcontainer',
  ]);
  const BUTTON_CHROME = new Set(['room']);
  const TERM_FONTS = ['small', 'medium', 'large'];
  const TERM_LABEL = { small: 'S', medium: 'M', large: 'L' };

  $: config = getWidgetConfig(widget.widgetType);
  $: isTabContainer = widget.widgetType === 'tabcontainer';
  $: collapsed = !!widget.collapsed;
  $: titleChrome = TITLE_CHROME.has(widget.widgetType);
  $: buttonChrome = BUTTON_CHROME.has(widget.widgetType);
  $: showTitleBar = !editMode && (titleChrome || (buttonChrome && collapsed));
  $: showFloatButtons = !editMode && buttonChrome && !collapsed;

  let termFontLabel = 'M';

  function readTermFontLabel() {
    try {
      const key = localStorage.getItem('talesmud_term_fontsize') || 'medium';
      termFontLabel = TERM_LABEL[key] || 'M';
    } catch (e) {
      termFontLabel = 'M';
    }
  }

  function cycleTermFont(event) {
    event.stopPropagation();
    event.preventDefault();
    let idx = 1;
    try {
      const cur = localStorage.getItem('talesmud_term_fontsize');
      const found = TERM_FONTS.indexOf(cur);
      if (found >= 0) idx = found;
      const next = TERM_FONTS[(idx + 1) % TERM_FONTS.length];
      localStorage.setItem('talesmud_term_fontsize', next);
      termFontLabel = TERM_LABEL[next];
    } catch (e) { /* ignore */ }
    window.dispatchEvent(new Event('talesmud-term-font'));
  }

  readTermFontLabel();

  function handleRemove() {
    dispatch('remove', { id: widget.id });
  }

  function handleConfigure() {
    dispatch('configure', { id: widget.id });
  }
</script>

<style>
  .widget-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  .widget-wrapper.has-title-chrome {
    display: flex;
    flex-direction: column;
    background: var(--panel-bg, rgba(0, 0, 0, 0.78));
    border: 1px solid var(--panel-border, rgba(255, 255, 255, 0.1));
    border-radius: var(--panel-radius, 12px);
    box-shadow: var(--panel-shadow, none);
  }

  .widget-wrapper.has-title-chrome .widget-content {
    flex: 1;
    min-height: 0;
  }

  /* One header: the wrapper bar. Inner titles and the old terminal font row stay hidden. */
  .widget-wrapper.has-title-chrome .widget-content :global(.game-panel-header),
  .widget-wrapper.has-title-chrome .widget-content :global(.terminal-toolbar),
  .widget-wrapper.has-title-chrome .widget-content :global(.tx-title),
  .widget-wrapper.has-title-chrome .widget-content :global(.questlog-header h2) {
    display: none !important;
  }

  .widget-wrapper.has-title-chrome .widget-content :global(.game-panel),
  .widget-wrapper.has-title-chrome .widget-content :global(.terminal-widget),
  .widget-wrapper.has-title-chrome .widget-content :global(.tx-window),
  .widget-wrapper.has-title-chrome .widget-content :global(.tab-container),
  .widget-wrapper.has-title-chrome .widget-content :global(.questlog-widget),
  .widget-wrapper.has-title-chrome .widget-content :global(.room-widget) {
    border: none !important;
    border-radius: 0 !important;
    box-shadow: none !important;
    height: 100%;
  }

  .widget-content {
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  .widget-content.is-collapsed {
    display: none;
  }

  .widget-content.disabled {
    pointer-events: none;
    opacity: 0.7;
  }

  .edit-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(245, 158, 11, 0.1);
    border: 2px dashed rgba(245, 158, 11, 0.5);
    border-radius: 12px;
    z-index: 100;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    pointer-events: none;
  }

  .widget-label {
    background: rgba(245, 158, 11, 0.9);
    color: #000;
    padding: 0.4em 0.8em;
    border-radius: 4px;
    font-size: 0.85em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    display: flex;
    align-items: center;
    gap: 0.5em;
  }

  .widget-label i {
    font-size: 1.1em;
  }

  .remove-btn {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: rgba(239, 68, 68, 0.9);
    border: none;
    color: white;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: auto;
    transition: all 0.15s ease;
    z-index: 120;
  }

  .remove-btn:hover {
    background: #ef4444;
    transform: scale(1.1);
  }

  .remove-btn i {
    font-size: 18px;
  }

  .configure-btn {
    position: absolute;
    top: 8px;
    right: 44px;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: rgba(59, 130, 246, 0.9);
    border: none;
    color: white;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: auto;
    transition: all 0.15s ease;
    z-index: 120;
  }

  .configure-btn:hover {
    background: #3b82f6;
    transform: scale(1.1);
  }

  .configure-btn i {
    font-size: 18px;
  }

  .resize-hint {
    position: absolute;
    bottom: 8px;
    right: 36px;
    font-size: 0.7em;
    color: rgba(245, 158, 11, 0.8);
    background: rgba(0, 0, 0, 0.5);
    padding: 0.2em 0.4em;
    border-radius: 4px;
  }

  /* Resize handles */
  .resize-handle {
    position: absolute;
    pointer-events: auto;
    z-index: 110;
    opacity: 1;
  }

  /* Corner handles — always visible in edit mode */
  .resize-handle.corner {
    width: 32px;
    height: 32px;
  }

  .resize-handle.corner::after {
    content: '';
    position: absolute;
    width: 16px;
    height: 16px;
    border-color: #fbbf24;
    border-style: solid;
    border-width: 0;
    filter: drop-shadow(0 0 1px #000);
  }

  .resize-handle.se {
    right: 0;
    bottom: 0;
    cursor: se-resize;
  }
  .resize-handle.se::after {
    right: 4px;
    bottom: 4px;
    border-right-width: 4px;
    border-bottom-width: 4px;
  }

  .resize-handle.sw {
    left: 0;
    bottom: 0;
    cursor: sw-resize;
  }
  .resize-handle.sw::after {
    left: 4px;
    bottom: 4px;
    border-left-width: 4px;
    border-bottom-width: 4px;
  }

  .resize-handle.ne {
    right: 0;
    top: 0;
    cursor: ne-resize;
  }
  .resize-handle.ne::after {
    right: 4px;
    top: 4px;
    border-right-width: 4px;
    border-top-width: 4px;
  }

  .resize-handle.nw {
    left: 0;
    top: 0;
    cursor: nw-resize;
  }
  .resize-handle.nw::after {
    left: 4px;
    top: 4px;
    border-left-width: 4px;
    border-top-width: 4px;
  }

  /* Side handles */
  .resize-handle.side {
    background: transparent;
  }

  .resize-handle.side::after {
    content: '';
    position: absolute;
    background: #f59e0b;
    border-radius: 2px;
    transition: transform 0.15s ease;
  }

  .resize-handle.side:hover::after,
  .resize-handle.side:active::after {
    transform: scale(1.3);
  }

  .resize-handle.n,
  .resize-handle.s {
    left: 50%;
    transform: translateX(-50%);
    width: 60px;
    height: 12px;
    cursor: ns-resize;
  }
  .resize-handle.n::after,
  .resize-handle.s::after {
    left: 50%;
    transform: translateX(-50%);
    width: 40px;
    height: 4px;
  }

  .resize-handle.n {
    top: 0;
  }
  .resize-handle.n::after {
    top: 4px;
  }

  .resize-handle.s {
    bottom: 0;
  }
  .resize-handle.s::after {
    bottom: 4px;
  }

  .resize-handle.e,
  .resize-handle.w {
    top: 50%;
    transform: translateY(-50%);
    width: 12px;
    height: 60px;
    cursor: ew-resize;
  }
  .resize-handle.e::after,
  .resize-handle.w::after {
    top: 50%;
    transform: translateY(-50%);
    width: 4px;
    height: 40px;
  }

  .resize-handle.e {
    right: 0;
  }
  .resize-handle.e::after {
    right: 4px;
  }

  .resize-handle.w {
    left: 0;
  }
  .resize-handle.w::after {
    left: 4px;
  }
</style>

<div class="widget-wrapper" class:has-title-chrome={showTitleBar}>
  {#if showTitleBar}
    <WidgetChrome
      title={config?.name || widget.widgetType}
      icon={config?.icon || ''}
      widgetId={widget.id}
      {collapsed}
      showTitle={true}
    >
      {#if widget.widgetType === 'terminal'}
        <button type="button" class="chrome-font" title="Terminal font size" on:click={cycleTermFont}>{termFontLabel}</button>
      {/if}
    </WidgetChrome>
  {/if}
  {#if showFloatButtons}
    <WidgetChrome
      widgetId={widget.id}
      {collapsed}
      showTitle={false}
      corner="tl"
    />
  {/if}
  {#if editMode}
    <div class="edit-overlay">
      <span class="widget-label">
        {#if config?.icon}
          <i class="material-icons">{config.icon}</i>
        {/if}
        {config?.name || widget.widgetType}
      </span>
      <span class="resize-hint">{canResize ? 'Drag to move, corners to resize' : 'Layout locked'}</span>
    </div>
    {#if isTabContainer}
      <button class="configure-btn" on:click={handleConfigure} title="Configure tabs">
        <i class="material-icons">settings</i>
      </button>
    {/if}
    <button class="remove-btn" on:click={handleRemove} title="Remove widget">
      <i class="material-icons">close</i>
    </button>

    {#if canResize}
      <!-- Corner resize handles -->
      <div class="resize-handle corner se" on:pointerdown={resizePointerDown}></div>
      <div class="resize-handle corner sw" on:pointerdown={resizePointerDown}></div>
      <div class="resize-handle corner ne" on:pointerdown={resizePointerDown}></div>
      <div class="resize-handle corner nw" on:pointerdown={resizePointerDown}></div>

      <!-- Side resize handles -->
      <div class="resize-handle side n" on:pointerdown={resizePointerDown}></div>
      <div class="resize-handle side s" on:pointerdown={resizePointerDown}></div>
      <div class="resize-handle side e" on:pointerdown={resizePointerDown}></div>
      <div class="resize-handle side w" on:pointerdown={resizePointerDown}></div>
    {/if}
  {/if}

  <div class="widget-content" class:disabled={editMode} class:is-collapsed={collapsed}>
    <slot />
  </div>
</div>
