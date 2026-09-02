<script>
  import { itemArtSrc, itemArtGenericKey } from '../itemArtSrc.js';

  export let store;
  export let sendMessage;

  const PAGE_SIZE = 10; // 5 rows × 2 columns

  let tab = 'buy';
  let errorText = '';
  let page = 0;

  $: shop = $store.shop || null;
  $: stock = shop?.stock || [];
  $: gold = shop?.gold ?? $store.gold ?? 0;
  $: inventory = $store.inventory || [];
  $: equipped = $store.equippedItems || {};
  $: sellables = inventory.filter((item) => canSell(item));
  $: rows = tab === 'buy' ? stock : sellables;
  $: pageCount = Math.max(1, Math.ceil(rows.length / PAGE_SIZE));
  $: if (page >= pageCount) page = Math.max(0, pageCount - 1);
  $: pageRows = rows.slice(page * PAGE_SIZE, page * PAGE_SIZE + PAGE_SIZE);

  function closeShop() {
    errorText = '';
    page = 0;
    if (store?.clearShop) store.clearShop();
  }

  function setTab(next) {
    tab = next;
    page = 0;
    errorText = '';
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

  function qtyFor(item) {
    if (tab === 'buy') {
      if (item.quantity == null || item.quantity < 0) return null; // unlimited — no badge
      return item.quantity;
    }
    return item.quantity || 1;
  }

  function rowPrice(item) {
    return tab === 'buy' ? Number(item.price || 0) : sellPrice(item);
  }

  /** Prefer real art, but fall through to local SVG generics — never leave a broken img. */
  function shopArtSrc(item) {
    return itemArtSrc(item);
  }

  function onShopArtError(ev, item) {
    const img = ev && ev.currentTarget;
    if (!img) return;
    // Skip often-missing generic PNG; go straight to local SVG generics.
    const stage = img.dataset.fallback || '0';
    const key = itemArtGenericKey(item);
    if (stage === '0') {
      img.dataset.fallback = '1';
      img.src = `sprites/items/generic-${key}.svg`;
      return;
    }
    if (stage === '1') {
      img.dataset.fallback = '2';
      img.src = 'sprites/items/generic-default.svg';
    }
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

  function activate(item) {
    if (tab === 'buy') buyItem(item);
    else sellItem(item);
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
    max-width: 560px;
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
  .shop-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
    padding: 0.75em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: rgba(20, 26, 36, 0.9);
  }
  .shop-title {
    flex: 1;
    font-weight: 700;
    color: #f8fafc;
    letter-spacing: 0.02em;
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
    padding: 0.55em 0.85em 0;
  }
  .shop-tab {
    border: 1px solid rgba(148, 163, 184, 0.25);
    background: rgba(255, 255, 255, 0.04);
    color: #cbd5e1;
    border-radius: 6px;
    padding: 0.4em 0.85em;
    cursor: pointer;
    font-size: 0.9em;
  }
  .shop-tab.active {
    border-color: rgba(59, 130, 246, 0.55);
    background: rgba(59, 130, 246, 0.18);
    color: #bfdbfe;
  }
  .shop-error {
    margin: 0.5em 0.85em 0;
    padding: 0.5em 0.7em;
    border-radius: 6px;
    background: rgba(127, 29, 29, 0.45);
    border: 1px solid rgba(248, 113, 113, 0.4);
    color: #fecaca;
    font-size: 0.85em;
  }
  .shop-list {
    flex: 1;
    min-height: 0;
    overflow: auto;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.35em 0.65em;
    padding: 0.7em 0.85em;
    align-content: start;
  }
  .shop-row {
    display: flex;
    align-items: center;
    gap: 0.55em;
    text-align: left;
    padding: 0.35em 0.45em;
    border-radius: 6px;
    border: 1px solid rgba(148, 163, 184, 0.18);
    background: rgba(255, 255, 255, 0.03);
    color: #e2e8f0;
    cursor: pointer;
    min-height: 48px;
  }
  .shop-row:hover {
    border-color: rgba(96, 165, 250, 0.5);
    background: rgba(59, 130, 246, 0.12);
  }
  .shop-row.oos {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .shop-icon {
    position: relative;
    flex: 0 0 40px;
    width: 40px;
    height: 40px;
    border-radius: 4px;
    background: #0b1119;
    border: 1px solid rgba(148, 163, 184, 0.25);
    overflow: hidden;
  }
  .shop-icon img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    image-rendering: pixelated;
    display: block;
  }
  .shop-stack {
    position: absolute;
    right: 2px;
    bottom: 1px;
    font-size: 0.65em;
    font-weight: 700;
    color: #fff;
    text-shadow: 0 0 2px #000, 0 1px 2px #000;
    line-height: 1;
    pointer-events: none;
  }
  .shop-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.1em;
  }
  .shop-name {
    font-weight: 600;
    font-size: 0.82em;
    line-height: 1.2;
    color: #f8fafc;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .shop-price {
    color: #fbbf24;
    font-weight: 600;
    font-size: 0.75em;
  }
  .shop-empty {
    grid-column: 1 / -1;
    text-align: center;
    color: #94a3b8;
    padding: 2em 1em;
  }
  .shop-pager {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75em;
    padding: 0.55em 0.85em 0.75em;
    border-top: 1px solid rgba(148, 163, 184, 0.15);
  }
  .shop-pager button {
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: rgba(255, 255, 255, 0.05);
    color: #cbd5e1;
    border-radius: 6px;
    padding: 0.3em 0.75em;
    cursor: pointer;
    font-size: 0.8em;
  }
  .shop-pager button:disabled {
    opacity: 0.35;
    cursor: default;
  }
  .shop-page-label {
    color: #94a3b8;
    font-size: 0.8em;
    min-width: 5.5em;
    text-align: center;
  }

  @media screen and (max-width: 520px) {
    .shop-list {
      grid-template-columns: 1fr;
    }
    .shop-panel {
      max-width: none;
    }
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
        <button class="shop-tab" class:active={tab === 'buy'} type="button" on:click={() => setTab('buy')}>Buy</button>
        <button class="shop-tab" class:active={tab === 'sell'} type="button" on:click={() => setTab('sell')}>Sell</button>
      </div>
      {#if errorText}
        <div class="shop-error">{errorText}</div>
      {/if}
      <div class="shop-list">
        {#each pageRows as item (item.templateId || item.id || item.name)}
          {@const qty = qtyFor(item)}
          {@const oos = tab === 'buy' && item.quantity === 0}
          <button
            class="shop-row"
            class:oos={oos}
            type="button"
            on:click={() => activate(item)}
            disabled={oos}
          >
            <div class="shop-icon">
              <img
                src={shopArtSrc(item)}
                alt=""
                on:error={(e) => onShopArtError(e, item)}
              />
              {#if qty != null && qty > 1}
                <span class="shop-stack">{qty}</span>
              {/if}
            </div>
            <div class="shop-text">
              <div class="shop-name">{item.name}</div>
              <div class="shop-price">{rowPrice(item)}g</div>
            </div>
          </button>
        {:else}
          <div class="shop-empty">{tab === 'buy' ? 'Nothing for sale.' : 'Nothing you can sell here.'}</div>
        {/each}
      </div>
      {#if rows.length > PAGE_SIZE}
        <div class="shop-pager">
          <button type="button" disabled={page <= 0} on:click={() => (page = Math.max(0, page - 1))}>Prev</button>
          <span class="shop-page-label">Page {page + 1} of {pageCount}</span>
          <button type="button" disabled={page >= pageCount - 1} on:click={() => (page = Math.min(pageCount - 1, page + 1))}>Next</button>
        </div>
      {/if}
    </div>
  </div>
{/if}
