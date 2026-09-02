<script>
  import InventoryWidget from '../widgets/InventoryWidget.svelte';

  export let store;
  export let sendMessage;

  function close() {
    if (store?.closeInventoryOverlay) store.closeInventoryOverlay();
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
    padding: 0.75em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: rgba(20, 26, 36, 0.9);
    flex-shrink: 0;
  }
  .inv-title {
    flex: 1;
    font-weight: 700;
    color: #f8fafc;
    display: flex;
    align-items: center;
    gap: 0.4em;
  }
  .inv-title i {
    color: #fbbf24;
    font-size: 1.2em;
  }
  .inv-gold {
    color: #fbbf24;
    font-weight: 600;
    font-size: 0.95em;
  }
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
      <div class="inv-header">
        <div class="inv-title"><i class="material-icons">inventory_2</i> Inventory</div>
        <div class="inv-gold">{$store.gold || 0} gold</div>
        <button class="inv-close" type="button" on:click={close} aria-label="Close inventory">×</button>
      </div>
      <div class="inv-body">
        <InventoryWidget {store} {sendMessage} />
      </div>
    </div>
  </div>
{/if}
