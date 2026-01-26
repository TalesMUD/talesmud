<script>
  // Placeholder widget - will be connected to server data later
  const mockEquipment = {
    head: { name: 'Steel Helm', icon: 'person', rarity: 'uncommon' },
    body: { name: 'Chainmail', icon: 'checkroom', rarity: 'uncommon' },
    hands: null,
    feet: { name: 'Leather Boots', icon: 'do_not_step', rarity: 'common' },
    mainHand: { name: 'Iron Sword', icon: 'bolt', rarity: 'uncommon' },
    offHand: { name: 'Wooden Shield', icon: 'shield', rarity: 'common' },
    ring1: { name: 'Gold Ring', icon: 'circle', rarity: 'rare' },
    ring2: null,
    neck: null,
  };

  const slotLabels = {
    head: 'Head',
    body: 'Body',
    hands: 'Hands',
    feet: 'Feet',
    mainHand: 'Main Hand',
    offHand: 'Off Hand',
    ring1: 'Ring 1',
    ring2: 'Ring 2',
    neck: 'Neck',
  };

  function getRarityColor(rarity) {
    if (!rarity) return '#4b5563';
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
  }

  .slot-empty {
    font-size: 0.7em;
    color: #4b5563;
    font-style: italic;
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

<div class="equipment-widget">
  <div class="widget-header">
    <i class="material-icons">shield</i>
    <span class="widget-title">Equipment</span>
  </div>

  <div class="equipment-grid">
    <!-- Head & Neck row -->
    <div class="equipment-row">
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.head}
        style="border-color: {getRarityColor(mockEquipment.head?.rarity)}"
      >
        <span class="slot-label">{slotLabels.head}</span>
        {#if mockEquipment.head}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.head.rarity)}">{mockEquipment.head.icon}</i>
          <span class="slot-name">{mockEquipment.head.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.neck}
        style="border-color: {getRarityColor(mockEquipment.neck?.rarity)}"
      >
        <span class="slot-label">{slotLabels.neck}</span>
        {#if mockEquipment.neck}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.neck.rarity)}">{mockEquipment.neck.icon}</i>
          <span class="slot-name">{mockEquipment.neck.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
    </div>

    <!-- Body row -->
    <div class="equipment-row">
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.body}
        style="border-color: {getRarityColor(mockEquipment.body?.rarity)}"
      >
        <span class="slot-label">{slotLabels.body}</span>
        {#if mockEquipment.body}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.body.rarity)}">{mockEquipment.body.icon}</i>
          <span class="slot-name">{mockEquipment.body.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
    </div>

    <!-- Hands row -->
    <div class="equipment-row">
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.mainHand}
        style="border-color: {getRarityColor(mockEquipment.mainHand?.rarity)}"
      >
        <span class="slot-label">{slotLabels.mainHand}</span>
        {#if mockEquipment.mainHand}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.mainHand.rarity)}">{mockEquipment.mainHand.icon}</i>
          <span class="slot-name">{mockEquipment.mainHand.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.hands}
        style="border-color: {getRarityColor(mockEquipment.hands?.rarity)}"
      >
        <span class="slot-label">{slotLabels.hands}</span>
        {#if mockEquipment.hands}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.hands.rarity)}">{mockEquipment.hands.icon}</i>
          <span class="slot-name">{mockEquipment.hands.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.offHand}
        style="border-color: {getRarityColor(mockEquipment.offHand?.rarity)}"
      >
        <span class="slot-label">{slotLabels.offHand}</span>
        {#if mockEquipment.offHand}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.offHand.rarity)}">{mockEquipment.offHand.icon}</i>
          <span class="slot-name">{mockEquipment.offHand.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
    </div>

    <!-- Feet & Rings row -->
    <div class="equipment-row">
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.feet}
        style="border-color: {getRarityColor(mockEquipment.feet?.rarity)}"
      >
        <span class="slot-label">{slotLabels.feet}</span>
        {#if mockEquipment.feet}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.feet.rarity)}">{mockEquipment.feet.icon}</i>
          <span class="slot-name">{mockEquipment.feet.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.ring1}
        style="border-color: {getRarityColor(mockEquipment.ring1?.rarity)}"
      >
        <span class="slot-label">{slotLabels.ring1}</span>
        {#if mockEquipment.ring1}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.ring1.rarity)}">{mockEquipment.ring1.icon}</i>
          <span class="slot-name">{mockEquipment.ring1.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
      <div
        class="equipment-slot"
        class:empty={!mockEquipment.ring2}
        style="border-color: {getRarityColor(mockEquipment.ring2?.rarity)}"
      >
        <span class="slot-label">{slotLabels.ring2}</span>
        {#if mockEquipment.ring2}
          <i class="material-icons slot-icon" style="color: {getRarityColor(mockEquipment.ring2.rarity)}">{mockEquipment.ring2.icon}</i>
          <span class="slot-name">{mockEquipment.ring2.name}</span>
        {:else}
          <span class="slot-empty">Empty</span>
        {/if}
      </div>
    </div>
  </div>

  <div class="placeholder-notice">
    Placeholder data - server integration pending
  </div>
</div>
