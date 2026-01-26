<script>
  import { createEventDispatcher } from 'svelte';
  import { getWidgetConfig } from './WidgetRegistry.js';

  export let widget;
  export let editMode = false;
  export let resizePointerDown = null;

  const dispatch = createEventDispatcher();

  $: config = getWidgetConfig(widget.widgetType);

  function handleRemove() {
    dispatch('remove', { id: widget.id });
  }
</script>

<style>
  .widget-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  .widget-content {
    width: 100%;
    height: 100%;
    overflow: hidden;
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
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .widget-wrapper:hover .resize-handle,
  .resize-handle:active {
    opacity: 1;
  }

  /* Corner handles */
  .resize-handle.corner {
    width: 20px;
    height: 20px;
  }

  .resize-handle.corner::after {
    content: '';
    position: absolute;
    width: 10px;
    height: 10px;
    border-color: #f59e0b;
    border-style: solid;
    border-width: 0;
  }

  .resize-handle.se {
    right: 0;
    bottom: 0;
    cursor: se-resize;
  }
  .resize-handle.se::after {
    right: 4px;
    bottom: 4px;
    border-right-width: 3px;
    border-bottom-width: 3px;
  }

  .resize-handle.sw {
    left: 0;
    bottom: 0;
    cursor: sw-resize;
  }
  .resize-handle.sw::after {
    left: 4px;
    bottom: 4px;
    border-left-width: 3px;
    border-bottom-width: 3px;
  }

  .resize-handle.ne {
    right: 0;
    top: 0;
    cursor: ne-resize;
  }
  .resize-handle.ne::after {
    right: 4px;
    top: 4px;
    border-right-width: 3px;
    border-top-width: 3px;
  }

  .resize-handle.nw {
    left: 0;
    top: 0;
    cursor: nw-resize;
  }
  .resize-handle.nw::after {
    left: 4px;
    top: 4px;
    border-left-width: 3px;
    border-top-width: 3px;
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

<div class="widget-wrapper">
  {#if editMode}
    <div class="edit-overlay">
      <span class="widget-label">
        {#if config?.icon}
          <i class="material-icons">{config.icon}</i>
        {/if}
        {config?.name || widget.widgetType}
      </span>
      <span class="resize-hint">Drag to move, edges to resize</span>
    </div>
    <button class="remove-btn" on:click={handleRemove} title="Remove widget">
      <i class="material-icons">close</i>
    </button>

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

  <div class="widget-content" class:disabled={editMode}>
    <slot />
  </div>
</div>
