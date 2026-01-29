<script>
  export let store = null;
  export let sendMessage = null;

  let inventory = [];
  let equippedItems = {};
  let gold = 0;
  let selectedItemId = null;

  // Subscribe to store
  $: if (store) {
    inventory = $store.inventory || [];
    equippedItems = $store.equippedItems || {};
    gold = $store.gold || 0;
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
    return Object.values(equippedItems).some(eq => eq && eq.ID === item.ID);
  }

  function toggleActions(itemId) {
    selectedItemId = selectedItemId === itemId ? null : itemId;
  }

  function sendCmd(cmd) {
    if (sendMessage) {
      sendMessage(cmd);
    }
    selectedItemId = null;
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
</script>

<style>
  .inventory-widget {
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 1em;
    height: 100%;
    overflow-y: auto;
    color: #e5e7eb;
  }

  .widget-header {
    display: flex;
    align-items: center;
    gap: 0.5em;
    margin-bottom: 1em;
    padding-bottom: 0.75em;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .widget-header i {
    color: #f59e0b;
  }

  .widget-title {
    font-size: 1em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .header-right {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 0.75em;
  }

  .gold-display {
    display: flex;
    align-items: center;
    gap: 0.25em;
    font-size: 0.8em;
    color: #f59e0b;
  }

  .gold-display i {
    font-size: 1em;
  }

  .item-count {
    font-size: 0.8em;
    color: #6b7280;
  }

  .inventory-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(70px, 1fr));
    gap: 0.5em;
  }

  .item-slot {
    aspect-ratio: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
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
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.2);
    transform: translateY(-2px);
  }

  .item-slot.selected {
    background: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .item-slot.equipped-item {
    background: rgba(34, 197, 94, 0.1);
  }

  .equipped-badge {
    position: absolute;
    top: 2px;
    left: 2px;
    font-size: 0.55em;
    background: rgba(34, 197, 94, 0.8);
    color: #fff;
    padding: 0.1em 0.3em;
    border-radius: 3px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .item-icon {
    font-size: 1.5em;
    margin-bottom: 0.2em;
  }

  .item-name {
    font-size: 0.65em;
    text-align: center;
    line-height: 1.2;
    color: #d1d5db;
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
    font-size: 0.65em;
    font-weight: 600;
  }

  .action-popup {
    position: absolute;
    bottom: 100%;
    left: 50%;
    transform: translateX(-50%);
    background: rgba(20, 20, 30, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 6px;
    padding: 0.3em;
    display: flex;
    gap: 0.2em;
    z-index: 100;
    white-space: nowrap;
    margin-bottom: 4px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5);
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: 0.2em;
    padding: 0.25em 0.4em;
    border: none;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
    cursor: pointer;
    font-size: 0.65em;
    font-family: inherit;
    transition: background 0.1s;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .action-btn i {
    font-size: 1.1em;
  }

  .action-btn.equip { color: #22c55e; }
  .action-btn.unequip { color: #f59e0b; }
  .action-btn.use { color: #3b82f6; }
  .action-btn.drop { color: #ef4444; }

  .empty-state {
    text-align: center;
    color: #6b7280;
    padding: 2em 1em;
    font-size: 0.85em;
  }

  .empty-state i {
    font-size: 2.5em;
    display: block;
    margin-bottom: 0.5em;
    opacity: 0.4;
  }
</style>

<div class="inventory-widget">
  <div class="widget-header">
    <i class="material-icons">inventory_2</i>
    <span class="widget-title">Inventory</span>
    <div class="header-right">
      {#if gold > 0}
        <span class="gold-display">
          <i class="material-icons">paid</i>
          {gold}
        </span>
      {/if}
      <span class="item-count">{inventory.length} items</span>
    </div>
  </div>

  {#if inventory.length === 0}
    <div class="empty-state">
      <i class="material-icons">inventory_2</i>
      Your inventory is empty.
    </div>
  {:else}
    <div class="inventory-grid">
      {#each inventory as item (item.ID || item.name)}
        {@const equipped = isEquipped(item)}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div
          class="item-slot"
          class:selected={selectedItemId === item.ID}
          class:equipped-item={equipped}
          title="{item.name}{item.quality && item.quality !== 'normal' ? ' [' + item.quality.toUpperCase() + ']' : ''}{item.type ? ' (' + item.type + ')' : ''}"
          style="border-color: {getQualityColor(item.quality)}"
          on:click={() => toggleActions(item.ID)}
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

          {#if selectedItemId === item.ID}
            <div class="action-popup">
              {#if equipped}
                <button class="action-btn unequip" on:click|stopPropagation={() => handleUnequip(item)}>
                  <i class="material-icons">remove_circle_outline</i> Unequip
                </button>
              {:else if isEquippable(item)}
                <button class="action-btn equip" on:click|stopPropagation={() => handleEquip(item)}>
                  <i class="material-icons">shield</i> Equip
                </button>
              {/if}
              {#if isConsumable(item) && !equipped}
                <button class="action-btn use" on:click|stopPropagation={() => handleUse(item)}>
                  <i class="material-icons">local_drink</i> Use
                </button>
              {/if}
              {#if !equipped}
                <button class="action-btn drop" on:click|stopPropagation={() => handleDrop(item)}>
                  <i class="material-icons">delete_outline</i> Drop
                </button>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
