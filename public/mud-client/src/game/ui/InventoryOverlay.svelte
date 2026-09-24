<script>
  import InventoryWidget from '../widgets/InventoryWidget.svelte';

  export let store;
  export let sendMessage;

  function close() {
    if (store?.closeInventoryOverlay) store.closeInventoryOverlay();
  }

  function openRecipes() {
    close();
    if (sendMessage) sendMessage('recipes');
  }
</script>

<style>
  .inv-overlay {
    position: fixed;
    inset: 0;
    z-index: 120;
    background: rgba(0, 0, 0, 0.82);
    backdrop-filter: blur(6px);
    display: flex;
    flex-direction: column;
    padding: 1em;
    overflow: hidden;
  }
  .inv-panel {
    flex: 1;
    min-height: 0;
    max-width: 720px;
    width: 100%;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    background: rgba(12, 16, 24, 0.97);
    border: 1px solid rgba(212, 175, 55, 0.28);
    border-radius: 10px;
    overflow: hidden;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.55);
  }
  .inv-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
    flex-shrink: 0;
    margin: 0 !important;
  }
  .inv-header :global(.widget-title) {
    flex: 1;
  }
  .inv-gold {
    color: #fbbf24;
    font-weight: 600;
    font-size: 0.95em;
  }
  .inv-craft {
    border: 1px solid rgba(34, 197, 94, 0.4);
    background: rgba(34, 197, 94, 0.12);
    color: #86efac;
    border-radius: 6px;
    padding: 0.3em 0.7em;
    cursor: pointer;
    font-size: 0.82em;
    font-weight: 700;
    font-family: inherit;
    display: flex;
    align-items: center;
    gap: 0.25em;
  }
  .inv-craft:hover { background: rgba(34, 197, 94, 0.22); }
  .inv-craft i { font-size: 1em; }
  .inv-close {
    border: none;
    background: transparent;
    color: #94a3b8;
    font-size: 1.4em;
    cursor: pointer;
    line-height: 1;
  }
  .inv-body {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }
</style>

{#if $store.inventoryOverlayOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="inv-overlay" role="dialog" aria-label="Inventory" on:click={(e) => { if (e.target === e.currentTarget) close(); }}>
    <div class="inv-panel" on:click|stopPropagation>
      <div class="inv-header game-panel-header">
        <i class="material-icons">inventory_2</i>
        <span class="widget-chrome-title widget-title">Inventory</span>
        <div class="inv-gold">{$store.gold || 0} gold</div>
        <button class="inv-craft" type="button" on:click={openRecipes} aria-label="Open crafting recipes">
          <i class="material-icons">construction</i>
          Craft
        </button>
        <button class="inv-close" type="button" on:click={close} aria-label="Close inventory">×</button>
      </div>
      <div class="inv-body widget-scroll">
        <InventoryWidget {store} {sendMessage} embedded={true} />
      </div>
    </div>
  </div>
{/if}
