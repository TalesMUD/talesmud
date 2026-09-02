<script>
  import { itemArtSrc, onItemArtError } from '../itemArtSrc.js';

  export let store;
  export let sendMessage;

  let tab = 'buy';
  let errorText = '';

  $: shop = $store.shop || null;
  $: stock = shop?.stock || [];
  $: gold = shop?.gold ?? $store.gold ?? 0;
  $: inventory = $store.inventory || [];
  $: equipped = $store.equippedItems || {};
  $: sellables = inventory.filter((item) => canSell(item));

  function closeShop() {
    errorText = '';
    if (store?.clearShop) store.clearShop();
  }

  function stockLabel(qty) {
    if (qty == null || qty < 0) return '∞';
    return String(qty);
  }

  function canSell(item) {
    if (!item || !shop) return false;
    if (item.boundToCharacterId) return false;
    if (String(item.type || '').toLowerCase() === 'quest') return false;
    const equippedIds = new Set(Object.values(equipped || {}).map((e) => e && e.id).filter(Boolean));
    if (equippedIds.has(item.id)) return false;
    const accepted = shop.acceptedTypes || [];
    if (accepted.length > 0 && !accepted.includes(item.type)) return false;
    const rejected = new Set(shop.rejectedTags || []);
    for (const tag of item.tags || []) {
      if (rejected.has(tag)) return false;
    }
    return true;
  }

  function sellPrice(item) {
    const base = Number(item.basePrice || 0);
    const mult = Number(shop?.sellMultiplier || 0.5);
    const price = Math.floor(base * mult);
    return price > 0 ? price : 1;
  }

  function buyItem(row) {
    errorText = '';
    if (!row?.name) return;
    if (row.quantity === 0) {
      errorText = 'That item is out of stock.';
      return;
    }
    if (Number(row.price || 0) > Number(gold || 0)) {
      errorText = `You don't have enough gold. Need ${row.price} gold.`;
      return;
    }
    sendMessage(`buy ${row.name}`);
  }

  function sellItem(item) {
    errorText = '';
    if (!item?.name) return;
    const name = item.instanceSuffix ? `${item.name}-${item.instanceSuffix}` : item.name;
    sendMessage(`sell ${name}`);
  }

  // Surface engine error replies while shop is open.
  $: if ($store.shopError) {
    errorText = $store.shopError;
    if (store?.clearShopError) store.clearShopError();
  }
</script>

<style>
  .shop-overlay {
    position: absolute;
    inset: 0;
    z-index: 120;
    background: rgba(0, 0, 0, 0.82);
    backdrop-filter: blur(6px);
    display: flex;
    flex-direction: column;
    padding: 1em;
    overflow: hidden;
  }
  .shop-panel {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: rgba(12, 16, 24, 0.96);
    border: 1px solid rgba(148, 163, 184, 0.28);
    border-radius: 12px;
    overflow: hidden;
  }
  .shop-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
    padding: 0.85em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
  }
  .shop-title {
    flex: 1;
    font-weight: 700;
    color: #f8fafc;
  }
  .shop-gold {
    color: #fbbf24;
    font-weight: 600;
    font-size: 0.95em;
  }
  .shop-close {
    border: none;
    background: transparent;
    color: #94a3b8;
    font-size: 1.4em;
    cursor: pointer;
    line-height: 1;
  }
  .shop-tabs {
    display: flex;
    gap: 0.35em;
    padding: 0.6em 0.85em 0;
  }
  .shop-tab {
    border: 1px solid rgba(148, 163, 184, 0.25);
    background: rgba(255, 255, 255, 0.04);
    color: #cbd5e1;
    border-radius: 8px;
    padding: 0.45em 0.9em;
    cursor: pointer;
  }
  .shop-tab.active {
    border-color: rgba(59, 130, 246, 0.55);
    background: rgba(59, 130, 246, 0.18);
    color: #bfdbfe;
  }
  .shop-error {
    margin: 0.6em 0.85em 0;
    padding: 0.55em 0.75em;
    border-radius: 8px;
    background: rgba(127, 29, 29, 0.45);
    border: 1px solid rgba(248, 113, 113, 0.4);
    color: #fecaca;
    font-size: 0.9em;
  }
  .shop-grid {
    flex: 1;
    min-height: 0;
    overflow: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(75px, 1fr));
    gap: 0.35em;
    padding: 0.45em;
  }
  .shop-card {
    display: flex;
    flex-direction: column;
    gap: 0.2em;
    text-align: left;
    padding: 0.35em;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.22);
    background: rgba(255, 255, 255, 0.04);
    color: #e2e8f0;
    cursor: pointer;
  }
  .shop-card:hover {
    border-color: rgba(96, 165, 250, 0.5);
    background: rgba(59, 130, 246, 0.12);
  }
  .shop-card.oos {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .shop-art {
    width: 100%;
    aspect-ratio: 1;
    object-fit: contain;
    image-rendering: pixelated;
    background: #0b1119;
    border-radius: 4px;
  }
  .shop-name {
    font-weight: 700;
    font-size: 0.72em;
    line-height: 1.15;
  }
  .shop-meta {
    display: flex;
    justify-content: space-between;
    gap: 0.25em;
    font-size: 0.68em;
    color: #94a3b8;
  }
  .shop-price {
    color: #fbbf24;
    font-weight: 600;
  }
  .shop-empty {
    grid-column: 1 / -1;
    text-align: center;
    color: #94a3b8;
    padding: 2em 1em;
  }
</style>

{#if shop}
  <div class="shop-overlay" role="dialog" aria-label="Shop">
    <div class="shop-panel">
      <div class="shop-header">
        <div class="shop-title">{shop.merchantName || 'Merchant'}'s Shop</div>
        <div class="shop-gold">{gold} gold</div>
        <button class="shop-close" type="button" on:click={closeShop} aria-label="Close shop">×</button>
      </div>
      <div class="shop-tabs">
        <button class="shop-tab" class:active={tab === 'buy'} type="button" on:click={() => (tab = 'buy')}>Buy</button>
        <button class="shop-tab" class:active={tab === 'sell'} type="button" on:click={() => (tab = 'sell')}>Sell</button>
      </div>
      {#if errorText}
        <div class="shop-error">{errorText}</div>
      {/if}
      <div class="shop-grid">
        {#if tab === 'buy'}
          {#each stock as row}
            <button
              class="shop-card"
              class:oos={row.quantity === 0}
              type="button"
              on:click={() => buyItem(row)}
              disabled={row.quantity === 0}
            >
              <img class="shop-art" src={itemArtSrc(row)} alt="" on:error={(e) => onItemArtError(e, row)} />
              <div class="shop-name">{row.name}</div>
              <div class="shop-meta">
                <span class="shop-price">{row.price}g</span>
                <span>×{stockLabel(row.quantity)}</span>
              </div>
            </button>
          {:else}
            <div class="shop-empty">Nothing for sale.</div>
          {/each}
        {:else}
          {#each sellables as item}
            <button class="shop-card" type="button" on:click={() => sellItem(item)}>
              <img class="shop-art" src={itemArtSrc(item)} alt="" on:error={(e) => onItemArtError(e, item)} />
              <div class="shop-name">{item.name}</div>
              <div class="shop-meta">
                <span class="shop-price">{sellPrice(item)}g</span>
                <span>×{item.quantity || 1}</span>
              </div>
            </button>
          {:else}
            <div class="shop-empty">Nothing you can sell here.</div>
          {/each}
        {/if}
      </div>
    </div>
  </div>
{/if}
