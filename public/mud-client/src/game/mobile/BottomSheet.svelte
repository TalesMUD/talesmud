<style>
  .bottom-sheet-backdrop {
    position: fixed;
    top: var(--sheet-top, 48px);
    right: 0;
    bottom: var(--sheet-bottom, 56px);
    left: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: var(--layer-sheet-backdrop);
    animation: backdropFadeIn 0.2s ease-out;
  }

  @keyframes backdropFadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .bottom-sheet {
    position: fixed;
    top: calc(var(--sheet-top, 48px) + 4px);
    right: 8px;
    bottom: calc(var(--sheet-bottom, 56px) + 4px);
    left: 8px;
    height: auto;
    background: rgba(10, 10, 18, 0.97);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 16px;
    z-index: var(--layer-sheet);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    animation: sheetSlideUp 0.3s ease-out;
    box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.5);
    padding-bottom: env(safe-area-inset-bottom, 0px);
  }

  @keyframes sheetSlideUp {
    from { transform: translateY(100%); }
    to { transform: translateY(0); }
  }

  .drag-handle-area {
    display: none;
  }

  .drag-handle {
    width: 36px;
    height: 4px;
    background: rgba(255, 255, 255, 0.3);
    border-radius: 2px;
  }

  .sheet-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px 4px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    flex-shrink: 0;
  }

  .sheet-title {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #e5e7eb;
    font-size: 14px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .sheet-title i {
    font-size: 20px;
    color: #f59e0b;
  }

  .sheet-close {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    color: #9ca3af;
    cursor: pointer;
    padding: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
    min-width: 44px;
    min-height: 44px;
  }

  .sheet-close:hover, .sheet-close:active {
    background: rgba(255, 255, 255, 0.15);
    color: #e5e7eb;
  }

  .sheet-close i {
    font-size: 20px;
  }

  .sheet-content {
    position: relative;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    padding: 12px 0;
    display: flex;
    flex-direction: column;
  }

  .sheet-content.roomy :global(.questlog-widget.embedded) {
    position: absolute;
    inset: 0;
    overflow: hidden;
  }

  .sheet-content.roomy {
    font-size: 14px;
    line-height: 1.45;
  }

  .sheet-content.roomy :global(button),
  .sheet-content.roomy :global([role="button"]),
  .sheet-content.roomy :global([role="tab"]),
  .sheet-content.roomy :global(input),
  .sheet-content.roomy :global(select) {
    min-width: 44px;
    min-height: 44px;
  }

  .sheet-content.roomy :global(.skill-card-desc),
  .sheet-content.roomy :global(.skills-note),
  .sheet-content.roomy :global(.char-info),
  .sheet-content.roomy :global(.quest-description),
  .sheet-content.roomy :global(.obj-text),
  .sheet-content.roomy :global(.empty-state),
  .sheet-content.roomy :global(.skills-empty) {
    font-size: 14px;
    line-height: 1.45;
  }

  /* Hide the widget's own header since the sheet already has one */
  .sheet-content :global(.widget-header) {
    display: none !important;
  }
  @media (prefers-reduced-motion: reduce) {
    .bottom-sheet, .bottom-sheet-backdrop { animation: none; transition: none; }
  }
</style>

<script>
  import { createEventDispatcher } from 'svelte';

  export let title = '';
  export let icon = '';
  export let roomy = false;

  const dispatch = createEventDispatcher();

  let sheetEl;
  let startY = 0;
  let currentY = 0;
  let dragging = false;

  function close() {
    dispatch('close');
  }

  function handleTouchStart(e) {
    startY = e.touches[0].clientY;
    currentY = 0;
    dragging = true;
  }

  function handleTouchMove(e) {
    if (!dragging) return;
    const delta = e.touches[0].clientY - startY;
    currentY = Math.max(0, delta);
    if (sheetEl) {
      sheetEl.style.transform = `translateY(${currentY}px)`;
      sheetEl.style.transition = 'none';
    }
  }

  function handleTouchEnd() {
    if (!dragging) return;
    dragging = false;
    if (currentY > 100) {
      close();
    } else if (sheetEl) {
      sheetEl.style.transition = 'transform 0.2s ease-out';
      sheetEl.style.transform = 'translateY(0)';
    }
    currentY = 0;
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="bottom-sheet-backdrop" on:click={close}></div>

<div class="bottom-sheet" bind:this={sheetEl}>
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="drag-handle-area"
    on:touchstart={handleTouchStart}
    on:touchmove={handleTouchMove}
    on:touchend={handleTouchEnd}
  >
    <div class="drag-handle"></div>
  </div>

  <div class="sheet-header">
    <span class="sheet-title">
      {#if icon}<i class="material-icons">{icon}</i>{/if}
      {title}
    </span>
    <button class="sheet-close" type="button" aria-label={`Close ${title}`} on:click={close}>
      <i class="material-icons">close</i>
    </button>
  </div>

  <div class="sheet-content" class:roomy>
    <slot></slot>
  </div>
</div>
