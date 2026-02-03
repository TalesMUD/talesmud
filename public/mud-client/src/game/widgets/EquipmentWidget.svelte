<script>
  export let store = null;
  export let sendMessage = null;

  let equippedItems = {};

  // Subscribe to store
  $: if (store) {
    equippedItems = $store.equippedItems || {};
  }

  // Backend slot keys mapped to display labels, grouped into rows
  const slotRows = [
    [
      { key: 'head', label: 'Head' },
      { key: 'neck', label: 'Neck' },
    ],
    [
      { key: 'chest', label: 'Chest' },
    ],
    [
      { key: 'main_hand', label: 'Main Hand' },
      { key: 'hands', label: 'Hands' },
      { key: 'off_hand', label: 'Off Hand' },
    ],
    [
      { key: 'legs', label: 'Legs' },
    ],
    [
      { key: 'boots', label: 'Boots' },
      { key: 'ring1', label: 'Ring 1' },
      { key: 'ring2', label: 'Ring 2' },
    ],
  ];

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
    if (item.meta && item.meta.img) return null;
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

  function getSlotIcon(slotKey) {
    switch (slotKey) {
      case 'head': return 'face';
      case 'neck': return 'more_horiz';
      case 'chest': return 'checkroom';
      case 'hands': return 'front_hand';
      case 'legs': return 'straighten';
      case 'boots': return 'do_not_step';
      case 'main_hand': return 'bolt';
      case 'off_hand': return 'shield';
      case 'ring1': return 'circle';
      case 'ring2': return 'circle';
      default: return 'help_outline';
    }
  }

  function getItemTooltip(item, slotLabel) {
    if (!item) return slotLabel + ' - Empty';
    let tip = item.name;
    if (item.quality && item.quality !== 'normal') {
      tip += ' [' + item.quality.toUpperCase() + ']';
    }
    if (item.type) {
      tip += ' (' + item.type + ')';
    }
    // Show key attributes
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
    return tip;
  }

  function handleUnequip(item) {
    if (!sendMessage || !item) return;
    const name = item.instanceSuffix ? item.name + '-' + item.instanceSuffix : item.name;
    sendMessage('unequip ' + name);
  }
</script>

<style>
  .equipment-widget {
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

  .equipment-grid {
    display: flex;
    flex-direction: column;
    gap: 0.5em;
  }

  .equipment-row {
    display: flex;
    gap: 0.5em;
  }

  .equipment-slot {
    flex: 1;
    min-height: 60px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0.5em;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .equipment-slot:hover {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.2);
  }

  .equipment-slot.empty {
    opacity: 0.5;
    cursor: default;
  }

  .equipment-slot.filled:hover {
    background: rgba(239, 68, 68, 0.1);
    border-color: rgba(239, 68, 68, 0.3);
  }

  .slot-label {
    font-size: 0.65em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #6b7280;
    margin-bottom: 0.3em;
  }

  .slot-icon {
    font-size: 1.3em;
    margin-bottom: 0.2em;
  }

  .slot-name {
    font-size: 0.7em;
    text-align: center;
    color: #d1d5db;
    line-height: 1.2;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .slot-empty {
    font-size: 0.7em;
    color: #4b5563;
    font-style: italic;
  }

  .unequip-hint {
    font-size: 0.55em;
    color: #ef4444;
    opacity: 0;
    transition: opacity 0.15s ease;
    margin-top: 0.2em;
  }

  .equipment-slot.filled:hover .unequip-hint {
    opacity: 1;
  }
</style>

<div class="equipment-widget">
  <div class="widget-header">
    <i class="material-icons">shield</i>
    <span class="widget-title">Equipment</span>
  </div>

  <div class="equipment-grid">
    {#each slotRows as row}
      <div class="equipment-row">
        {#each row as slot}
          {@const item = equippedItems[slot.key]}
          <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
          <div
            class="equipment-slot"
            class:empty={!item}
            class:filled={!!item}
            style="border-color: {item ? getQualityColor(item.quality) : 'rgba(255, 255, 255, 0.1)'}"
            title={getItemTooltip(item, slot.label)}
            on:click={() => item && handleUnequip(item)}
          >
            <span class="slot-label">{slot.label}</span>
            {#if item}
              <i class="material-icons slot-icon" style="color: {getQualityColor(item.quality)}">
                {getItemIcon(item) || getSlotIcon(slot.key)}
              </i>
              <span class="slot-name">{item.name}</span>
              <span class="unequip-hint">click to unequip</span>
            {:else}
              <i class="material-icons slot-icon" style="color: #4b5563">{getSlotIcon(slot.key)}</i>
              <span class="slot-empty">Empty</span>
            {/if}
          </div>
        {/each}
      </div>
    {/each}
  </div>
</div>
