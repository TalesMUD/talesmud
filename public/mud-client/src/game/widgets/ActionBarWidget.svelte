<script>
  import MUDXPlus from '../MUDXPlus.svelte';
  import HotbarWidget from './HotbarWidget.svelte';
  import { layoutStore } from '../layout/LayoutStore.js';

  export let store;
  export let sendMessage;
  export let term = null;

  $: separateHotbar = ($layoutStore.widgets || []).some(
    (w) => w.widgetType === 'hotbar' && w.visible !== false
  );
</script>

<style>
  .actionbar-widget {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .actionbar-widget.with-hotbar {
    background: var(--panel-bg, rgba(0, 0, 0, 0.6));
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1px solid var(--panel-border, rgba(255, 255, 255, 0.1));
    border-radius: var(--panel-radius, 12px);
    box-shadow: var(--panel-shadow, none);
    overflow: hidden;
  }

  .docked-hotbar {
    flex: 0 0 auto;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .actionbar-widget.with-hotbar :global(.mudx) {
    flex: 1;
    min-height: 0;
    height: auto;
  }

  .actionbar-widget.with-hotbar :global(.action-bar) {
    flex: 1;
    margin-bottom: 0;
    border: none;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
    height: 100%;
    box-sizing: border-box;
  }
</style>

<div class="actionbar-widget" class:with-hotbar={!separateHotbar}>
  {#if !separateHotbar}
    <div class="docked-hotbar">
      <HotbarWidget {store} {sendMessage} docked={true} />
    </div>
  {/if}
  <MUDXPlus {store} {sendMessage} {term} />
</div>
