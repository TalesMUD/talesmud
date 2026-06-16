<script>
  export let store = null;
  export let sendMessage = null;

  let inventory = [];
  let equippedItems = {};
  let gold = 0;
  let hasMerchant = false;
  let detailItem = null;
  let collapsedCategories = {};
  let viewMode = localStorage.getItem('inventoryViewMode') || 'grid';

  // Sell quantity popup state
  let showSellPopup = false;
  let sellItem = null;
  let sellQuantity = 1;

  function toggleViewMode() {
    viewMode = viewMode === 'grid' ? 'list' : 'grid';
    localStorage.setItem('inventoryViewMode', viewMode);
  }

  // Category definitions
  const CATEGORIES = [
    { key: 'equipment', label: 'Equipment', icon: 'shield', color: '#22c55e',
      filter: (item) => item.type === 'weapon' || item.type === 'armor' },
    { key: 'consumables', label: 'Consumables', icon: 'local_drink', color: '#3b82f6',
      filter: (item) => item.type === 'consumable' },
    { key: 'quest', label: 'Quest Items', icon: 'auto_stories', color: '#f59e0b',
      filter: (item) => item.type === 'quest' },
    { key: 'other', label: 'Other', icon: 'category', color: '#9ca3af',
      filter: (item) => !['weapon', 'armor', 'consumable', 'quest'].includes(item.type) },
  ];

  // Subscribe to store
  $: if (store) {
    inventory = $store.inventory || [];
    equippedItems = $store.equippedItems || {};
    gold = $store.gold || 0;
    hasMerchant = $store.hasMerchant || false;
  }

  // Group items by category, only include non-empty categories
  $: categorizedItems = CATEGORIES
    .map(cat => ({ ...cat, items: inventory.filter(cat.filter) }))
    .filter(cat => cat.items.length > 0);

  function toggleCategory(key) {
    collapsedCategories[key] = !collapsedCategories[key];
    collapsedCategories = collapsedCategories;
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

  function getItemIcon(item) {
    if (item.meta && item.meta.img) return null; // use image instead
    switch (item.type) {
      case 'weapon': return 'bolt';
      case 'armor': return 'security';
      case 'consumable': return 'local_drink';
      case 'quest': return 'auto_stories';
      case 'currency': return 'paid';
      case 'collectible': return 'star';
      case 'crafting_material': return 'build';
      default: return 'inventory_2';
    }
  }

  function isEquippable(item) {
    return item.slot && item.slot !== 'inventory' && item.slot !== 'container' && item.slot !== 'purse';
  }

  function isConsumable(item) {
    return item.type === 'consumable' || item.consumable;
  }

  function isEquipped(item) {
    if (!equippedItems) return false;
    return Object.values(equippedItems).some(eq => eq && eq.id === item.id);
  }

  function isBound(item) {
    return !!item.boundToCharacterId;
  }

  function canSell(item) {
    return hasMerchant && item && item.type !== 'quest' && !isBound(item) && !isEquipped(item);
  }

  function toggleActions(item) {
    if (detailItem && detailItem.id === item.id) {
      detailItem = null;
    } else {
      detailItem = item;
    }
  }

  function closeDetail() {
    detailItem = null;
  }

  function sendCmd(cmd) {
    if (sendMessage) {
      sendMessage(cmd);
    }
    detailItem = null;
  }

  function formatTypeName(str) {
    if (!str) return '';
    return str.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
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

  function handleEquip(item) {
    const name = item.instanceSuffix ? item.name + '-' + item.instanceSuffix : item.name;
    sendCmd('equip ' + name);
  }

  function handleDrop(item) {
    const name = item.instanceSuffix ? item.name + '-' + item.instanceSuffix : item.name;
    sendCmd('drop ' + name);
  }

  function handleUse(item) {
    const name = item.instanceSuffix ? item.name + '-' + item.instanceSuffix : item.name;
    sendCmd('use ' + name);
  }

  function handleUnequip(item) {
    const name = item.instanceSuffix ? item.name + '-' + item.instanceSuffix : item.name;
    sendCmd('unequip ' + name);
  }

  function handleExamine(item) {
    const name = item.instanceSuffix ? item.name + '-' + item.instanceSuffix : item.name;
    sendCmd('examine ' + name);
  }

  function handleSell(item) {
    if (!canSell(item)) return;
    if (item.stackable && item.quantity > 1) {
      // Show quantity popup for stacked items
      sellItem = item;
      sellQuantity = 1;
      detailItem = null;
      showSellPopup = true;
    } else {
      // Sell single item directly
      const name = item.instanceSuffix ? item.name + '-' + item.instanceSuffix : item.name;
      sendCmd('sell ' + name);
    }
  }

  function confirmSell() {
    if (!sellItem) return;
    const name = sellItem.instanceSuffix ? sellItem.name + '-' + sellItem.instanceSuffix : sellItem.name;
    const quantity = clampSellQuantity(sellQuantity, sellItem.quantity || 1);
    if (quantity > 1) {
      sendCmd('sell ' + name + ' ' + quantity);
    } else {
      sendCmd('sell ' + name);
    }
    closeSellPopup();
  }

  function closeSellPopup() {
    showSellPopup = false;
    sellItem = null;
    sellQuantity = 1;
  }

  function sellAll() {
    if (!sellItem) return;
    sellQuantity = clampSellQuantity(sellItem.quantity || 1, sellItem.quantity || 1);
  }

  function clampSellQuantity(value, max) {
    const limit = Math.max(1, Number(max) || 1);
    const parsed = Number.parseInt(value, 10);
    if (!Number.isFinite(parsed) || parsed < 1) return 1;
    if (parsed > limit) return limit;
    return parsed;
  }

  function onSellQuantityInput() {
    if (!sellItem) return;
    sellQuantity = clampSellQuantity(sellQuantity, sellItem.quantity || 1);
  }

  function getItemTooltip(item) {
    let tip = item.name;
    if (item.quality && item.quality !== 'normal') {
      tip += ' [' + item.quality.toUpperCase() + ']';
    }
    if (item.type) {
      tip += ' (' + item.type + ')';
    }
    if (item.slot && item.slot !== 'inventory') {
      tip += ' - Slot: ' + item.slot;
    }
    if (item.attributes) {
      const stats = [];
      if (item.attributes.damage != null) stats.push('Dmg: ' + item.attributes.damage);
      if (item.attributes.defense != null) stats.push('Def: ' + item.attributes.defense);
      if (item.attributes.armor != null) stats.push('Armor: ' + item.attributes.armor);
      if (item.attributes.strength != null) stats.push('Str: +' + item.attributes.strength);
      if (item.attributes.agility != null) stats.push('Agi: +' + item.attributes.agility);
      if (item.attributes.intelligence != null) stats.push('Int: +' + item.attributes.intelligence);
      if (stats.length > 0) {
        tip += '\n' + stats.join(', ');
      }
    }
    if (item.description) {
      tip += '\n' + item.description;
    }
    return tip;
  }
</script>

<style>
  /* Base panel styling comes from global .game-panel class in themes.css */
  .inventory-widget {
    position: relative;
  }

  .widget-title {
    flex: 1;
  }

  .widget-toolbar {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.75em;
    margin-bottom: 0.75em;
    padding-bottom: 0.5em;
    border-bottom: 1px solid var(--divider-color);
  }

  .gold-display {
    display: flex;
    align-items: center;
    gap: 0.25em;
    font-size: var(--text-sm);
    color: var(--color-gold);
  }

  .gold-display i {
    font-size: 1em;
  }

  .item-count {
    font-size: var(--text-sm);
    color: var(--text-dim);
  }

  .view-toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 0.2em;
    border-radius: 4px;
    transition: background 0.15s ease, color 0.15s ease;
    color: var(--text-dim);
  }

  .view-toggle:hover {
    background: var(--panel-inner-hover);
    color: var(--text-primary);
  }

  .view-toggle i {
    font-size: 1.1em;
    color: inherit;
  }

  .category-section {
    margin-bottom: 0.75em;
  }

  .category-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.45em 0.55em;
    margin-bottom: 0.4em;
    cursor: pointer;
    border-radius: 6px;
    background: var(--panel-inner-bg);
    border-bottom: 1px solid var(--panel-inner-border);
    transition: background 0.15s ease;
    user-select: none;
  }

  .category-header:hover {
    background: var(--panel-inner-hover);
  }

  .category-header-left {
    display: flex;
    align-items: center;
    gap: 0.4em;
  }

  .category-icon {
    font-size: 1em;
  }

  .category-label {
    font-size: var(--text-sm);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-primary);
  }

  .category-count {
    font-size: var(--text-xs);
    color: var(--text-dim);
    background: var(--panel-inner-bg);
    padding: 0.1em 0.4em;
    border-radius: 8px;
  }

  .category-chevron {
    font-size: 1em;
    color: var(--text-dim);
    transition: transform 0.2s ease;
  }

  .category-chevron.collapsed {
    transform: rotate(-90deg);
  }

  .inventory-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(85px, 1fr));
    gap: 0.5em;
  }

  .item-slot {
    aspect-ratio: 1;
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0.4em;
    cursor: pointer;
    transition: all 0.15s ease;
    position: relative;
  }

  .item-slot:hover {
    background: var(--panel-inner-hover);
    border-color: var(--panel-border-hover);
    transform: translateY(-2px);
  }

  .item-slot.selected {
    background: var(--accent-glow);
    border-color: var(--accent-primary);
  }

  .item-slot.equipped-item {
    background: rgba(34, 197, 94, 0.1);
  }

  .equipped-badge {
    position: absolute;
    top: 3px;
    left: 3px;
    font-size: var(--text-xs);
    background: rgba(34, 197, 94, 0.8);
    color: #fff;
    padding: 0.1em 0.3em;
    border-radius: 3px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .item-icon {
    font-size: 1.6em;
    margin-bottom: 0.2em;
  }

  .item-name {
    font-size: var(--text-xs);
    text-align: center;
    line-height: 1.2;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .item-quantity {
    position: absolute;
    bottom: 4px;
    right: 4px;
    background: rgba(0, 0, 0, 0.7);
    padding: 0.15em 0.35em;
    border-radius: 4px;
    font-size: var(--text-xs);
    font-weight: 600;
  }

  /* List view */
  .inventory-list {
    display: flex;
    flex-direction: column;
    gap: 0.25em;
  }

  .list-item {
    display: flex;
    align-items: center;
    gap: 0.5em;
    padding: 0.4em 0.55em;
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
    position: relative;
  }

  .list-item:hover {
    background: var(--panel-inner-hover);
    border-color: var(--panel-border-hover);
  }

  .list-item.selected {
    background: var(--accent-glow);
    border-color: var(--accent-primary);
  }

  .list-item.equipped-item {
    background: rgba(34, 197, 94, 0.06);
  }

  .list-item-icon {
    font-size: 1.2em;
    flex-shrink: 0;
  }

  .list-item-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.05em;
  }

  .list-item-name {
    font-size: var(--text-base);
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .list-item-meta {
    font-size: var(--text-xs);
    color: var(--text-dim);
    display: flex;
    align-items: center;
    gap: 0.4em;
  }

  .list-item-quality {
    font-weight: 600;
    text-transform: capitalize;
  }

  .list-item-type {
    text-transform: capitalize;
  }

  .list-item-right {
    display: flex;
    align-items: center;
    gap: 0.4em;
    flex-shrink: 0;
  }

  .list-equipped-tag {
    font-size: var(--text-xs);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.3px;
    color: #22c55e;
    background: rgba(34, 197, 94, 0.15);
    padding: 0.15em 0.35em;
    border-radius: 3px;
  }

  .list-item-qty {
    font-size: var(--text-xs);
    color: var(--text-secondary);
    font-weight: 600;
  }

  /* Detail overlay */
  .detail-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 99;
  }

  .detail-overlay {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: calc(100vw - 4em);
    max-width: 300px;
    max-height: calc(100vh - 4em);
    overflow-y: auto;
    background: var(--panel-bg);
    border: 1px solid var(--panel-border);
    border-radius: var(--panel-radius);
    padding: 1em;
    z-index: 100;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
  }

  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 0.75em;
    padding-bottom: 0.6em;
    border-bottom: 1px solid var(--divider-color);
  }

  .detail-title-row {
    display: flex;
    align-items: center;
    gap: 0.6em;
  }

  .detail-icon {
    font-size: 2em;
  }

  .detail-title-info {
    display: flex;
    flex-direction: column;
    gap: 0.15em;
  }

  .detail-name {
    font-family: var(--font-display);
    font-size: var(--text-base);
    font-weight: 700;
  }

  .detail-meta {
    display: flex;
    gap: 0.4em;
    font-size: var(--text-xs);
    text-transform: capitalize;
  }

  .detail-quality {
    font-weight: 600;
  }

  .detail-type {
    color: var(--text-secondary);
  }

  .detail-close {
    cursor: pointer;
    color: var(--text-dim);
    font-size: 1.2em;
    transition: color 0.15s;
    flex-shrink: 0;
  }

  .detail-close:hover {
    color: var(--text-primary);
  }

  .detail-slot {
    display: flex;
    align-items: center;
    gap: 0.35em;
    font-size: var(--text-sm);
    color: var(--text-secondary);
    text-transform: capitalize;
    margin-bottom: 0.6em;
  }

  .detail-stats {
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    padding: 0.5em 0.6em;
    margin-bottom: 0.6em;
  }

  .stat-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.25em 0;
    font-size: var(--text-sm);
  }

  .stat-row + .stat-row {
    border-top: 1px solid var(--divider-color);
  }

  .stat-label {
    color: var(--text-secondary);
  }

  .stat-value {
    font-weight: 600;
    color: var(--text-primary);
  }

  .stat-offensive {
    color: #ef4444;
  }

  .stat-defensive {
    color: #3b82f6;
  }

  .detail-description {
    font-size: var(--text-sm);
    color: var(--text-secondary);
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
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    padding: 0.35em 0.4em;
  }

  .info-label {
    font-size: var(--text-xs);
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .info-value {
    font-size: var(--text-sm);
    font-weight: 600;
    color: var(--text-primary);
  }

  .detail-gold {
    color: var(--color-gold);
  }

  .detail-equipped-tag {
    display: inline-flex;
    align-items: center;
    gap: 0.3em;
    font-size: var(--text-xs);
    color: #22c55e;
    background: rgba(34, 197, 94, 0.1);
    padding: 0.2em 0.5em;
    border-radius: 4px;
    margin-bottom: 0.6em;
  }

  .detail-actions {
    display: flex;
    gap: 0.4em;
    flex-wrap: wrap;
    margin-top: 0.5em;
    padding-top: 0.6em;
    border-top: 1px solid var(--divider-color);
  }

  .detail-action-btn {
    display: flex;
    align-items: center;
    gap: 0.3em;
    padding: 0.45em 0.75em;
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    background: var(--panel-inner-bg);
    color: var(--text-primary);
    cursor: pointer;
    font-size: var(--text-sm);
    font-family: inherit;
    transition: all 0.15s;
  }

  .detail-action-btn:hover {
    background: var(--panel-inner-hover);
    border-color: var(--panel-border-hover);
  }

  .detail-action-btn i {
    font-size: 1.1em;
  }

  .detail-action-btn.examine { color: #a78bfa; border-color: rgba(167, 139, 250, 0.3); }
  .detail-action-btn.examine:hover { background: rgba(167, 139, 250, 0.15); }
  .detail-action-btn.equip { color: #22c55e; border-color: rgba(34, 197, 94, 0.3); }
  .detail-action-btn.equip:hover { background: rgba(34, 197, 94, 0.15); }
  .detail-action-btn.unequip { color: #f59e0b; border-color: rgba(245, 158, 11, 0.3); }
  .detail-action-btn.unequip:hover { background: rgba(245, 158, 11, 0.15); }
  .detail-action-btn.use { color: #3b82f6; border-color: rgba(59, 130, 246, 0.3); }
  .detail-action-btn.use:hover { background: rgba(59, 130, 246, 0.15); }
  .detail-action-btn.sell { color: #f59e0b; border-color: rgba(245, 158, 11, 0.3); }
  .detail-action-btn.sell:hover { background: rgba(245, 158, 11, 0.15); }
  .detail-action-btn.drop { color: #ef4444; border-color: rgba(239, 68, 68, 0.3); }
  .detail-action-btn.drop:hover { background: rgba(239, 68, 68, 0.15); }

  /* Sell quantity popup */
  .sell-popup-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 110;
  }

  .sell-popup {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: calc(100vw - 4em);
    max-width: 280px;
    background: var(--panel-bg);
    border: 1px solid var(--panel-border);
    border-radius: var(--panel-radius);
    padding: 1em;
    z-index: 111;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
  }

  .sell-popup-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75em;
    padding-bottom: 0.5em;
    border-bottom: 1px solid var(--divider-color);
  }

  .sell-popup-title {
    font-family: var(--font-display);
    font-size: var(--text-base);
    font-weight: 700;
    color: var(--text-primary);
  }

  .sell-popup-close {
    cursor: pointer;
    color: var(--text-dim);
    font-size: 1.2em;
    transition: color 0.15s;
  }

  .sell-popup-close:hover {
    color: var(--text-primary);
  }

  .sell-popup-body {
    display: flex;
    flex-direction: column;
    gap: 0.75em;
    margin-bottom: 0.75em;
  }

  .sell-popup-info {
    font-size: var(--text-sm);
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .sell-popup-quantity {
    display: flex;
    flex-direction: column;
    gap: 0.35em;
  }

  .sell-popup-quantity label {
    font-size: var(--text-sm);
    color: var(--text-secondary);
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
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    background: var(--panel-inner-bg);
    color: var(--text-primary);
    cursor: pointer;
    font-size: var(--text-base);
    font-family: inherit;
    font-weight: 700;
    transition: all 0.15s;
  }

  .qty-btn:hover {
    background: var(--panel-inner-hover);
    border-color: var(--panel-border-hover);
  }

  .qty-all-btn {
    padding: 0.35em 0.6em;
    border: 1px solid rgba(245, 158, 11, 0.3);
    border-radius: 6px;
    background: var(--panel-inner-bg);
    color: #f59e0b;
    cursor: pointer;
    font-size: var(--text-sm);
    font-family: inherit;
    font-weight: 600;
    transition: all 0.15s;
  }

  .qty-all-btn:hover {
    background: rgba(245, 158, 11, 0.15);
  }

  .qty-controls input[type="number"] {
    width: 3.5em;
    text-align: center;
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    background: var(--panel-inner-bg);
    color: var(--text-primary);
    font-size: var(--text-base);
    font-family: inherit;
    padding: 0.3em;
    -moz-appearance: textfield;
  }

  .qty-controls input[type="number"]::-webkit-inner-spin-button,
  .qty-controls input[type="number"]::-webkit-outer-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  .sell-popup-total {
    font-size: var(--text-sm);
    font-weight: 600;
    color: var(--text-primary);
    padding: 0.4em 0.6em;
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    text-align: center;
  }

  .sell-popup-actions {
    display: flex;
    gap: 0.4em;
    justify-content: flex-end;
    padding-top: 0.5em;
    border-top: 1px solid var(--divider-color);
  }

  .empty-state {
    text-align: center;
    color: var(--text-dim);
    padding: 2em 1em;
    font-size: var(--text-sm);
  }

  .empty-state i {
    font-size: 2.5em;
    display: block;
    margin-bottom: 0.5em;
    opacity: 0.4;
  }
</style>

<div class="inventory-widget game-panel">
  <div class="game-panel-header">
    <i class="material-icons">inventory_2</i>
    <span class="widget-title">Inventory</span>
  </div>

  <div class="widget-toolbar">
    <span class="gold-display">
      <i class="material-icons">paid</i>
      {gold}
    </span>
    <span class="item-count">{inventory.length} items</span>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <span
      class="view-toggle"
      on:click={toggleViewMode}
      title={viewMode === 'grid' ? 'Switch to list view' : 'Switch to grid view'}
    >
      <i class="material-icons">{viewMode === 'grid' ? 'view_list' : 'grid_view'}</i>
    </span>
  </div>

  {#if inventory.length === 0}
    <div class="empty-state">
      <i class="material-icons">inventory_2</i>
      Your inventory is empty.
    </div>
  {:else}
    {#each categorizedItems as category (category.key)}
      <div class="category-section">
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div class="category-header" on:click={() => toggleCategory(category.key)}>
          <div class="category-header-left">
            <i class="material-icons category-icon" style="color: {category.color}">{category.icon}</i>
            <span class="category-label">{category.label}</span>
            <span class="category-count">{category.items.length}</span>
          </div>
          <i class="material-icons category-chevron" class:collapsed={collapsedCategories[category.key]}>
            expand_more
          </i>
        </div>

        {#if !collapsedCategories[category.key]}
          {#if viewMode === 'grid'}
            <div class="inventory-grid">
              {#each category.items as item (item.id || item.name)}
                {@const equipped = isEquipped(item)}
                <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
                <div
                  class="item-slot"
                  class:selected={detailItem && detailItem.id === item.id}
                  class:equipped-item={equipped}
                  title={getItemTooltip(item)}
                  style="border-color: {getQualityColor(item.quality)}"
                  on:click={() => toggleActions(item)}
                >
                  {#if equipped}
                    <span class="equipped-badge">E</span>
                  {/if}

                  <i class="material-icons item-icon" style="color: {getQualityColor(item.quality)}">
                    {getItemIcon(item)}
                  </i>
                  <span class="item-name">{item.name}</span>

                  {#if item.stackable && item.quantity > 1}
                    <span class="item-quantity">x{item.quantity}</span>
                  {/if}
                </div>
              {/each}
            </div>
          {:else}
            <div class="inventory-list">
              {#each category.items as item (item.id || item.name)}
                {@const equipped = isEquipped(item)}
                <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
                <div
                  class="list-item"
                  class:selected={detailItem && detailItem.id === item.id}
                  class:equipped-item={equipped}
                  title={getItemTooltip(item)}
                  on:click={() => toggleActions(item)}
                >
                  <i class="material-icons list-item-icon" style="color: {getQualityColor(item.quality)}">
                    {getItemIcon(item)}
                  </i>
                  <div class="list-item-info">
                    <span class="list-item-name" style="color: {getQualityColor(item.quality)}">{item.name}</span>
                    <span class="list-item-meta">
                      {#if item.quality && item.quality !== 'normal'}
                        <span class="list-item-quality" style="color: {getQualityColor(item.quality)}">{item.quality}</span>
                      {/if}
                      {#if item.type}
                        <span class="list-item-type">{formatTypeName(item.type)}</span>
                      {/if}
                    </span>
                  </div>
                  <div class="list-item-right">
                    {#if equipped}
                      <span class="list-equipped-tag">Equipped</span>
                    {/if}
                    {#if item.stackable && item.quantity > 1}
                      <span class="list-item-qty">x{item.quantity}</span>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </div>
    {/each}
  {/if}

  {#if detailItem}
    {@const equipped = isEquipped(detailItem)}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="detail-backdrop" on:click={closeDetail}></div>
    <div class="detail-overlay">
      <div class="detail-header">
        <div class="detail-title-row">
          <i class="material-icons detail-icon" style="color: {getQualityColor(detailItem.quality)}">
            {getItemIcon(detailItem)}
          </i>
          <div class="detail-title-info">
            <span class="detail-name" style="color: {getQualityColor(detailItem.quality)}">{detailItem.name}</span>
            <span class="detail-meta">
              {#if detailItem.quality}
                <span class="detail-quality" style="color: {getQualityColor(detailItem.quality)}">{formatTypeName(detailItem.quality)}</span>
              {/if}
              {#if detailItem.type}
                <span class="detail-type">{formatTypeName(detailItem.type)}{#if detailItem.subType} ({formatTypeName(detailItem.subType)}){/if}</span>
              {/if}
            </span>
          </div>
        </div>
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <i class="material-icons detail-close" on:click={closeDetail}>close</i>
      </div>

      {#if detailItem.slot && detailItem.slot !== 'inventory' && detailItem.slot !== 'container' && detailItem.slot !== 'purse'}
        <div class="detail-slot">
          <i class="material-icons" style="font-size: 0.85em">straighten</i>
          Slot: {detailItem.slot.replace('_', ' ')}
        </div>
      {/if}

      {#if detailItem.attributes && Object.keys(detailItem.attributes).length > 0}
        <div class="detail-stats">
          {#each Object.entries(detailItem.attributes) as [key, value]}
            <div class="stat-row">
              <span class="stat-label">{formatAttributeLabel(key)}</span>
              <span class="stat-value" class:stat-offensive={isOffensiveStat(key)} class:stat-defensive={isDefensiveStat(key)}>
                {#if !isOffensiveStat(key) && !isDefensiveStat(key)}+{/if}{value}
              </span>
            </div>
          {/each}
        </div>
      {/if}

      {#if detailItem.description}
        <div class="detail-description">{detailItem.description}</div>
      {/if}

      {#if detailItem.level || detailItem.stackable || detailItem.basePrice || equipped}
        <div class="detail-info-grid">
          {#if detailItem.level && detailItem.level > 0}
            <div class="info-item">
              <span class="info-label">Level</span>
              <span class="info-value">{detailItem.level}</span>
            </div>
          {/if}
          {#if detailItem.stackable}
            <div class="info-item">
              <span class="info-label">Stack</span>
              <span class="info-value">{detailItem.quantity || 1}/{detailItem.maxStack || '?'}</span>
            </div>
          {/if}
          {#if detailItem.basePrice}
            <div class="info-item">
              <span class="info-label">Value</span>
              <span class="info-value detail-gold">{detailItem.basePrice} gold</span>
            </div>
          {/if}
        </div>
      {/if}

      {#if equipped}
        <div class="detail-equipped-tag">
          <i class="material-icons" style="font-size: 0.85em">check_circle</i> Equipped
        </div>
      {/if}

      <div class="detail-actions">
        <button class="detail-action-btn examine" on:click={() => handleExamine(detailItem)}>
          <i class="material-icons">search</i> Examine
        </button>
        {#if equipped}
          <button class="detail-action-btn unequip" on:click={() => handleUnequip(detailItem)}>
            <i class="material-icons">remove_circle_outline</i> Unequip
          </button>
        {:else}
          {#if isEquippable(detailItem)}
            <button class="detail-action-btn equip" on:click={() => handleEquip(detailItem)}>
              <i class="material-icons">shield</i> Equip
            </button>
          {/if}
          {#if isConsumable(detailItem)}
            <button class="detail-action-btn use" on:click={() => handleUse(detailItem)}>
              <i class="material-icons">local_drink</i> Use
            </button>
          {/if}
        {/if}
        {#if canSell(detailItem)}
          <button class="detail-action-btn sell" on:click={() => handleSell(detailItem)}>
            <i class="material-icons">sell</i> Sell
          </button>
        {/if}
        <button class="detail-action-btn drop" on:click={() => handleDrop(detailItem)}>
          <i class="material-icons">delete_outline</i> Drop
        </button>
      </div>
    </div>
  {/if}

  {#if showSellPopup && sellItem}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="sell-popup-backdrop" on:click={closeSellPopup}></div>
    <div class="sell-popup">
      <div class="sell-popup-header">
        <span class="sell-popup-title">Sell {sellItem.name}</span>
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <i class="material-icons sell-popup-close" on:click={closeSellPopup}>close</i>
      </div>
      <div class="sell-popup-body">
        <div class="sell-popup-info">
          You have <strong>{sellItem.quantity}</strong> in your inventory.
          {#if sellItem.basePrice}
            <br/>Base value: <span class="detail-gold">{sellItem.basePrice} gold each</span>
          {/if}
        </div>
        <div class="sell-popup-quantity">
          <label for="sell-qty">Quantity:</label>
          <div class="qty-controls">
            <button class="qty-btn" on:click={() => { if (sellQuantity > 1) sellQuantity--; }}>-</button>
            <input
              id="sell-qty"
              type="number"
	              min="1"
	              max={sellItem.quantity}
	              bind:value={sellQuantity}
	              on:input={onSellQuantityInput}
	            />
            <button class="qty-btn" on:click={() => { if (sellQuantity < sellItem.quantity) sellQuantity++; }}>+</button>
            <button class="qty-all-btn" on:click={sellAll}>All</button>
          </div>
        </div>
        {#if sellItem.basePrice}
          <div class="sell-popup-total">
            Base value: <span class="detail-gold">{sellItem.basePrice * sellQuantity} gold</span>
          </div>
        {/if}
      </div>
      <div class="sell-popup-actions">
        <button class="detail-action-btn sell" on:click={confirmSell}>
          <i class="material-icons">sell</i> Sell {sellQuantity > 1 ? `x${sellQuantity}` : ''}
        </button>
        <button class="detail-action-btn drop" on:click={closeSellPopup}>
          Cancel
        </button>
      </div>
    </div>
  {/if}
</div>
