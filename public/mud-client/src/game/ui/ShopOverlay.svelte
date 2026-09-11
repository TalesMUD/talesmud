<script>
  import { onDestroy } from 'svelte';
  import { itemArtSrc, onItemArtError } from '../itemArtSrc.js';

  export let store;
  export let sendMessage;

  const PAGE_SIZE = 10; // 5 rows × 2 columns

  let tab = 'buy';
  let errorText = '';
  let page = 0;
  let selected = null;
  let buyQty = 1;
  let escHandler = null;

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

  // Keep selection in sync after shop refresh (post-buy) or inventory changes.
  $: if (selected && shop) {
    const key = itemKey(selected);
    const next =
      tab === 'buy'
        ? stock.find((r) => itemKey(r) === key)
        : sellables.find((r) => itemKey(r) === key);
    if (!next) {
      selected = null;
      buyQty = 1;
    } else if (next !== selected) {
      selected = next;
      buyQty = clampBuyQty(buyQty, next);
    }
  }

  $: if (shop && !escHandler) {
    escHandler = (e) => {
      if (e.key !== 'Escape') return;
      if (selected) {
        e.preventDefault();
        e.stopPropagation();
        clearSelection();
        return;
      }
      // No selection — close shop (same as ×). Don't swallow if already closed.
      e.preventDefault();
      closeShop();
    };
    if (typeof window !== 'undefined') window.addEventListener('keydown', escHandler, true);
  }

  $: if (!shop && escHandler) {
    if (typeof window !== 'undefined') window.removeEventListener('keydown', escHandler, true);
    escHandler = null;
    selected = null;
    buyQty = 1;
    errorText = '';
    page = 0;
  }

  onDestroy(() => {
    if (escHandler && typeof window !== 'undefined') {
      window.removeEventListener('keydown', escHandler, true);
    }
    escHandler = null;
  });

  function itemKey(item) {
    if (!item) return '';
    return String(item.templateId || item.id || item.name || '');
  }

  function closeShop() {
    errorText = '';
    page = 0;
    selected = null;
    buyQty = 1;
    if (store?.clearShop) store.clearShop();
  }

  function clearSelection() {
    selected = null;
    buyQty = 1;
    errorText = '';
  }

  function setTab(next) {
    tab = next;
    page = 0;
    errorText = '';
    clearSelection();
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

  function maxBuyQty(item) {
    if (!item) return 1;
    if (item.quantity == null || item.quantity < 0) {
      // Unlimited stock — allow stacks up to maxStack or a sane cap.
      if (item.stackable && item.maxStack > 0) return Math.min(99, item.maxStack);
      return item.stackable ? 99 : 1;
    }
    if (item.stackable) {
      const cap = item.maxStack > 0 ? Math.min(item.quantity, item.maxStack) : item.quantity;
      return Math.max(1, Math.min(99, cap));
    }
    return 1;
  }

  function clampBuyQty(value, item) {
    const limit = maxBuyQty(item);
    const parsed = Number.parseInt(value, 10);
    if (!Number.isFinite(parsed) || parsed < 1) return 1;
    if (parsed > limit) return limit;
    return parsed;
  }

  function selectItem(item) {
    if (!item) return;
    if (tab === 'buy' && item.quantity === 0) return;
    if (selected && itemKey(selected) === itemKey(item)) {
      // Toggle off on second click of same row.
      clearSelection();
      return;
    }
    selected = item;
    buyQty = 1;
    errorText = '';
  }

  function buyItem(row, quantity = 1) {
    errorText = '';
    if (!row?.name) return;
    if (row.quantity === 0) {
      errorText = 'That item is out of stock.';
      return;
    }
    const qty = clampBuyQty(quantity, row);
    const unit = Number(row.price || 0);
    if (unit * qty > Number(gold || 0)) {
      errorText = `You don't have enough gold. Need ${unit * qty} gold.`;
      return;
    }
    if (qty > 1) sendMessage(`buy ${row.name} ${qty}`);
    else sendMessage(`buy ${row.name}`);
    // Keep panel open; shop refresh will update selection/stock.
  }

  function sellItem(item, quantity = 1) {
    errorText = '';
    if (!item?.name) return;
    const name = item.instanceSuffix ? `${item.name}-${item.instanceSuffix}` : item.name;
    const qty = Math.max(1, Number.parseInt(quantity, 10) || 1);
    if (qty > 1) sendMessage(`sell ${name} ${qty}`);
    else sendMessage(`sell ${name}`);
  }

  function confirmTrade() {
    if (!selected) return;
    if (tab === 'buy') buyItem(selected, buyQty);
    else {
      const qty = selected.stackable && (selected.quantity || 1) > 1 ? buyQty : 1;
      sellItem(selected, qty);
    }
  }

  function getQualityColor(quality) {
    switch (quality) {
      case 'magic': return '#22c55e';
      case 'rare': return '#3b82f6';
      case 'legendary': return '#a855f7';
      case 'mythic': return '#f59e0b';
      default: return '#9ca3af';
    }
  }

  function formatTypeName(str) {
    if (!str) return '';
    return String(str).replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
  }

  function formatAttributeLabel(key) {
    const labels = {
      damage: 'Damage',
      defense: 'Defense',
      armor: 'Armor',
      strength: 'Strength',
      agility: 'Agility',
      intelligence: 'Intelligence',
      health: 'Health',
      mana: 'Mana',
      speed: 'Speed',
      critical: 'Critical',
    };
    return labels[key] || key.charAt(0).toUpperCase() + key.slice(1).replace(/_/g, ' ');
  }

  function isOffensiveStat(key) {
    return ['damage', 'critical', 'speed'].includes(key);
  }

  function isDefensiveStat(key) {
    return ['defense', 'armor', 'health'].includes(key);
  }

  function formatStatValue(key, value) {
    if (typeof value === 'number') {
      const n = Number.isInteger(value) ? value : Math.round(value * 100) / 100;
      if (!isOffensiveStat(key) && !isDefensiveStat(key)) return `+${n}`;
      return String(n);
    }
    return String(value);
  }

  function canShowQty(item) {
    if (!item) return false;
    if (tab === 'buy') return maxBuyQty(item) > 1;
    return !!(item.stackable && (item.quantity || 1) > 1);
  }

  function maxQtyForSelected() {
    if (!selected) return 1;
    if (tab === 'buy') return maxBuyQty(selected);
    return Math.max(1, selected.quantity || 1);
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
    position: relative;
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
    min-height: 72px;
  }
  .shop-row:hover {
    border-color: rgba(96, 165, 250, 0.5);
    background: rgba(59, 130, 246, 0.12);
  }
  .shop-row.selected {
    border-color: rgba(251, 191, 36, 0.65);
    background: rgba(251, 191, 36, 0.12);
    box-shadow: inset 0 0 0 1px rgba(251, 191, 36, 0.25);
  }
  .shop-row.oos {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .shop-icon {
    position: relative;
    flex: 0 0 64px;
    width: 64px;
    height: 64px;
    border-radius: 4px;
    background: #0b1119;
    border: 1px solid rgba(148, 163, 184, 0.25);
    overflow: hidden;
  }
  .shop-icon img {
    width: 64px;
    height: 64px;
    object-fit: contain;
    image-rendering: pixelated;
    image-rendering: crisp-edges;
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

  /* Detail inspect panel (mirrors InventoryWidget richness) */
  .detail-backdrop {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    z-index: 5;
  }
  .detail-overlay {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: calc(100% - 2em);
    max-width: 320px;
    max-height: calc(100% - 2em);
    overflow-y: auto;
    background: rgba(12, 16, 24, 0.98);
    border: 1px solid rgba(212, 175, 55, 0.35);
    border-radius: 10px;
    padding: 1em;
    z-index: 6;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.7);
  }
  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 0.75em;
    padding-bottom: 0.6em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
  }
  .detail-title-row {
    display: flex;
    align-items: center;
    gap: 0.6em;
    min-width: 0;
  }
  .detail-art-wrap {
    width: 56px;
    height: 56px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.35);
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
  .detail-art {
    width: 48px;
    height: 48px;
    object-fit: contain;
    image-rendering: pixelated;
  }
  .detail-title-info {
    display: flex;
    flex-direction: column;
    gap: 0.15em;
    min-width: 0;
  }
  .detail-name {
    font-size: 1em;
    font-weight: 700;
    line-height: 1.2;
  }
  .detail-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
    font-size: 0.75em;
    text-transform: capitalize;
  }
  .detail-quality { font-weight: 600; }
  .detail-type { color: #94a3b8; }
  .detail-close {
    cursor: pointer;
    color: #64748b;
    font-size: 1.2em;
    transition: color 0.15s;
    flex-shrink: 0;
    border: none;
    background: transparent;
    padding: 0;
    line-height: 1;
  }
  .detail-close:hover { color: #f8fafc; }
  .detail-slot {
    display: flex;
    align-items: center;
    gap: 0.35em;
    font-size: 0.85em;
    color: #94a3b8;
    text-transform: capitalize;
    margin-bottom: 0.6em;
  }
  .detail-stats {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(148, 163, 184, 0.18);
    border-radius: 6px;
    padding: 0.5em 0.6em;
    margin-bottom: 0.6em;
  }
  .stat-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.25em 0;
    font-size: 0.85em;
  }
  .stat-row + .stat-row {
    border-top: 1px solid rgba(148, 163, 184, 0.12);
  }
  .stat-label { color: #94a3b8; }
  .stat-value { font-weight: 600; color: #f8fafc; }
  .stat-offensive { color: #ef4444; }
  .stat-defensive { color: #3b82f6; }
  .detail-description {
    font-size: 0.85em;
    color: #94a3b8;
    line-height: 1.5;
    margin-bottom: 0.6em;
    font-style: italic;
  }
  .detail-info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(70px, 1fr));
    gap: 0.4em;
    margin-bottom: 0.6em;
  }
  .info-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(148, 163, 184, 0.18);
    border-radius: 6px;
    padding: 0.35em 0.4em;
  }
  .info-label {
    font-size: 0.65em;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .info-value {
    font-size: 0.85em;
    font-weight: 600;
    color: #f8fafc;
  }
  .detail-gold { color: #fbbf24; }
  .detail-qty {
    display: flex;
    flex-direction: column;
    gap: 0.35em;
    margin-bottom: 0.6em;
  }
  .detail-qty label {
    font-size: 0.8em;
    color: #94a3b8;
    font-weight: 600;
  }
  .qty-controls {
    display: flex;
    align-items: center;
    gap: 0.35em;
  }
  .qty-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2em;
    height: 2em;
    border: 1px solid rgba(148, 163, 184, 0.3);
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    color: #e2e8f0;
    cursor: pointer;
    font-size: 1em;
    font-family: inherit;
    font-weight: 700;
  }
  .qty-btn:disabled { opacity: 0.35; cursor: default; }
  .qty-controls input {
    width: 3.5em;
    text-align: center;
    border: 1px solid rgba(148, 163, 184, 0.3);
    border-radius: 6px;
    background: rgba(0, 0, 0, 0.35);
    color: #f8fafc;
    padding: 0.35em;
    font-family: inherit;
  }
  .qty-all-btn {
    border: 1px solid rgba(148, 163, 184, 0.3);
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    color: #cbd5e1;
    padding: 0.35em 0.6em;
    cursor: pointer;
    font-size: 0.8em;
    font-family: inherit;
  }
  .detail-actions {
    display: flex;
    gap: 0.4em;
    flex-wrap: wrap;
    margin-top: 0.5em;
    padding-top: 0.6em;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
  }
  .detail-action-btn {
    display: flex;
    align-items: center;
    gap: 0.3em;
    padding: 0.5em 0.85em;
    border: 1px solid rgba(148, 163, 184, 0.25);
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.04);
    color: #e2e8f0;
    cursor: pointer;
    font-size: 0.9em;
    font-family: inherit;
    font-weight: 600;
  }
  .detail-action-btn.buy {
    color: #86efac;
    border-color: rgba(34, 197, 94, 0.4);
    background: rgba(34, 197, 94, 0.12);
  }
  .detail-action-btn.buy:hover { background: rgba(34, 197, 94, 0.22); }
  .detail-action-btn.sell {
    color: #fbbf24;
    border-color: rgba(245, 158, 11, 0.4);
    background: rgba(245, 158, 11, 0.12);
  }
  .detail-action-btn.sell:hover { background: rgba(245, 158, 11, 0.22); }
  .detail-action-btn.cancel {
    color: #94a3b8;
  }
  .detail-action-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .detail-total {
    font-size: 0.85em;
    color: #cbd5e1;
    margin-bottom: 0.35em;
  }

  @media screen and (max-width: 520px) {
    .shop-list {
      grid-template-columns: 1fr;
    }
    .shop-panel {
      max-width: none;
    }
    .detail-overlay {
      top: auto;
      bottom: 0;
      left: 0;
      right: 0;
      transform: none;
      width: 100%;
      max-width: none;
      max-height: 70%;
      border-radius: 12px 12px 0 0;
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
          {@const isSelected = selected && itemKey(selected) === itemKey(item)}
          <button
            class="shop-row"
            class:oos={oos}
            class:selected={isSelected}
            type="button"
            on:click={() => selectItem(item)}
            disabled={oos}
            aria-pressed={isSelected}
          >
            <div class="shop-icon">
              <img
                src={itemArtSrc(item)}
                alt=""
                on:error={(e) => onItemArtError(e, item)}
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

      {#if selected}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div class="detail-backdrop" on:click={clearSelection}></div>
        <div class="detail-overlay" role="dialog" aria-label="{selected.name} details">
          <div class="detail-header">
            <div class="detail-title-row">
              <div class="detail-art-wrap">
                <img
                  class="detail-art"
                  src={itemArtSrc(selected)}
                  alt=""
                  on:error={(e) => onItemArtError(e, selected)}
                />
              </div>
              <div class="detail-title-info">
                <span class="detail-name" style="color: {getQualityColor(selected.quality)}">{selected.name}</span>
                <span class="detail-meta">
                  {#if selected.quality}
                    <span class="detail-quality" style="color: {getQualityColor(selected.quality)}">{formatTypeName(selected.quality)}</span>
                  {/if}
                  {#if selected.type}
                    <span class="detail-type">{formatTypeName(selected.type)}{#if selected.subType} ({formatTypeName(selected.subType)}){/if}</span>
                  {/if}
                </span>
              </div>
            </div>
            <button class="detail-close" type="button" on:click={clearSelection} aria-label="Close item details">×</button>
          </div>

          {#if selected.slot && selected.slot !== 'inventory' && selected.slot !== 'container' && selected.slot !== 'purse'}
            <div class="detail-slot">
              <i class="material-icons" style="font-size: 0.85em">straighten</i>
              Slot: {String(selected.slot).replace('_', ' ')}
            </div>
          {/if}

          {#if selected.attributes && Object.keys(selected.attributes).length > 0}
            <div class="detail-stats">
              {#each Object.entries(selected.attributes) as [key, value]}
                <div class="stat-row">
                  <span class="stat-label">{formatAttributeLabel(key)}</span>
                  <span class="stat-value" class:stat-offensive={isOffensiveStat(key)} class:stat-defensive={isDefensiveStat(key)}>
                    {formatStatValue(key, value)}
                  </span>
                </div>
              {/each}
            </div>
          {/if}

          {#if selected.description}
            <div class="detail-description">{selected.description}</div>
          {/if}

          <div class="detail-info-grid">
            <div class="info-item">
              <span class="info-label">{tab === 'buy' ? 'Price' : 'Sell for'}</span>
              <span class="info-value detail-gold">{rowPrice(selected)}g</span>
            </div>
            {#if tab === 'buy'}
              {#if selected.quantity == null || selected.quantity < 0}
                <div class="info-item">
                  <span class="info-label">Stock</span>
                  <span class="info-value">∞</span>
                </div>
              {:else}
                <div class="info-item">
                  <span class="info-label">Stock</span>
                  <span class="info-value">{selected.quantity}</span>
                </div>
              {/if}
            {:else if selected.quantity}
              <div class="info-item">
                <span class="info-label">Owned</span>
                <span class="info-value">{selected.quantity}</span>
              </div>
            {/if}
            {#if selected.level && selected.level > 0}
              <div class="info-item">
                <span class="info-label">Level</span>
                <span class="info-value">{selected.level}</span>
              </div>
            {/if}
            {#if selected.requiredLevel && selected.requiredLevel > 0}
              <div class="info-item">
                <span class="info-label">Req Lv</span>
                <span class="info-value">{selected.requiredLevel}</span>
              </div>
            {/if}
            {#if selected.basePrice && tab === 'buy'}
              <div class="info-item">
                <span class="info-label">Value</span>
                <span class="info-value detail-gold">{selected.basePrice}g</span>
              </div>
            {/if}
          </div>

          {#if canShowQty(selected)}
            <div class="detail-qty">
              <label for="shop-qty">Quantity</label>
              <div class="qty-controls">
                <button
                  class="qty-btn"
                  type="button"
                  disabled={buyQty <= 1}
                  on:click={() => (buyQty = Math.max(1, buyQty - 1))}
                >−</button>
                <input
                  id="shop-qty"
                  type="number"
                  min="1"
                  max={maxQtyForSelected()}
                  bind:value={buyQty}
                  on:input={() => {
                    buyQty = tab === 'buy'
                      ? clampBuyQty(buyQty, selected)
                      : Math.min(Math.max(1, Number.parseInt(buyQty, 10) || 1), selected.quantity || 1);
                  }}
                />
                <button
                  class="qty-btn"
                  type="button"
                  disabled={buyQty >= maxQtyForSelected()}
                  on:click={() => (buyQty = Math.min(maxQtyForSelected(), buyQty + 1))}
                >+</button>
                <button class="qty-all-btn" type="button" on:click={() => (buyQty = maxQtyForSelected())}>All</button>
              </div>
            </div>
            <div class="detail-total">
              Total: <span class="detail-gold">{rowPrice(selected) * buyQty}g</span>
            </div>
          {/if}

          <div class="detail-actions">
            {#if tab === 'buy'}
              <button
                class="detail-action-btn buy"
                type="button"
                disabled={selected.quantity === 0 || rowPrice(selected) * buyQty > gold}
                on:click={confirmTrade}
              >
                <i class="material-icons" style="font-size: 1em">shopping_cart</i>
                Buy{buyQty > 1 ? ` x${buyQty}` : ''}
              </button>
            {:else}
              <button class="detail-action-btn sell" type="button" on:click={confirmTrade}>
                <i class="material-icons" style="font-size: 1em">sell</i>
                Sell{canShowQty(selected) && buyQty > 1 ? ` x${buyQty}` : ''}
              </button>
            {/if}
            <button class="detail-action-btn cancel" type="button" on:click={clearSelection}>Cancel</button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}
