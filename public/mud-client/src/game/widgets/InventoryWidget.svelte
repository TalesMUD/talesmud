<script>
  // Placeholder widget - will be connected to server data later
  const mockInventory = [
    { id: 1, name: 'Health Potion', icon: 'local_drink', quantity: 5, rarity: 'common' },
    { id: 2, name: 'Mana Potion', icon: 'water_drop', quantity: 3, rarity: 'common' },
    { id: 3, name: 'Iron Sword', icon: 'bolt', quantity: 1, rarity: 'uncommon' },
    { id: 4, name: 'Leather Armor', icon: 'security', quantity: 1, rarity: 'common' },
    { id: 5, name: 'Gold Ring', icon: 'circle', quantity: 1, rarity: 'rare' },
    { id: 6, name: 'Torch', icon: 'local_fire_department', quantity: 10, rarity: 'common' },
    { id: 7, name: 'Rope', icon: 'cable', quantity: 1, rarity: 'common' },
    { id: 8, name: 'Ancient Scroll', icon: 'description', quantity: 1, rarity: 'epic' },
  ];

  function getRarityColor(rarity) {
    switch (rarity) {
      case 'uncommon': return '#22c55e';
      case 'rare': return '#3b82f6';
      case 'epic': return '#a855f7';
      case 'legendary': return '#f59e0b';
      default: return '#9ca3af';
    }
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

  .item-count {
    margin-left: auto;
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

  .placeholder-notice {
    font-size: 0.7em;
    color: #6b7280;
    text-align: center;
    margin-top: 1em;
    padding-top: 0.75em;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }
</style>

<div class="inventory-widget">
  <div class="widget-header">
    <i class="material-icons">inventory_2</i>
    <span class="widget-title">Inventory</span>
    <span class="item-count">{mockInventory.length} items</span>
  </div>

  <div class="inventory-grid">
    {#each mockInventory as item (item.id)}
      <div
        class="item-slot"
        title="{item.name} ({item.rarity})"
        style="border-color: {getRarityColor(item.rarity)}"
      >
        <i class="material-icons item-icon" style="color: {getRarityColor(item.rarity)}">{item.icon}</i>
        <span class="item-name">{item.name}</span>
        {#if item.quantity > 1}
          <span class="item-quantity">x{item.quantity}</span>
        {/if}
      </div>
    {/each}
  </div>

  <div class="placeholder-notice">
    Placeholder data - server integration pending
  </div>
</div>
