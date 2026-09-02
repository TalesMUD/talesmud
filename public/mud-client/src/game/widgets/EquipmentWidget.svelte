<script>
  import { itemArtSrc, onItemArtError } from '../itemArtSrc.js';
  import { portraitSrc, onPortraitError } from '../portraitSrc.js';

  export let store = null;
  export let sendMessage = null;

  let equippedItems = {};
  let character = null;
  let stats = {};

  $: if (store) {
    equippedItems = $store.equippedItems || {};
    character = $store.character || null;
    stats = $store.characterStats || {};
  }

  $: name = character?.name || 'Adventurer';
  $: race = typeof character?.race === 'object' ? (character.race.name || '') : (character?.race || '');
  $: charClass = typeof character?.class === 'object' ? (character.class.name || '') : (character?.class || '');
  $: level = character?.level || stats.level || 1;
  $: attackPower = stats.attackPower ?? character?.attackPower ?? 0;
  $: defense = stats.defense ?? character?.defense ?? 0;
  $: currentHp = stats.currentHitPoints ?? character?.currentHitPoints ?? 0;
  $: maxHp = stats.maxHitPoints ?? character?.maxHitPoints ?? 0;
  $: currentMana = stats.currentMana ?? character?.currentMana ?? 0;
  $: maxMana = stats.maxMana ?? character?.maxMana ?? 0;

  // Paper-doll columns — existing engine slots only
  const leftSlots = [
    { key: 'head', label: 'Head' },
    { key: 'neck', label: 'Neck' },
    { key: 'chest', label: 'Chest' },
    { key: 'hands', label: 'Hands' },
  ];
  const rightSlots = [
    { key: 'legs', label: 'Legs' },
    { key: 'boots', label: 'Boots' },
    { key: 'ring1', label: 'Ring 1' },
    { key: 'ring2', label: 'Ring 2' },
  ];
  const weaponSlots = [
    { key: 'main_hand', label: 'Main Hand' },
    { key: 'off_hand', label: 'Off Hand' },
  ];

  function getQualityColor(quality) {
    switch (quality) {
      case 'magic': return '#22c55e';
      case 'rare': return '#3b82f6';
      case 'legendary': return '#a855f7';
      case 'mythic': return '#f59e0b';
      default: return 'rgba(148, 163, 184, 0.45)';
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
    if (!item) return slotLabel + ' — Empty';
    let tip = item.name;
    if (item.quality && item.quality !== 'normal') {
      tip += ' [' + item.quality.toUpperCase() + ']';
    }
    if (item.type) tip += ' (' + item.type + ')';
    tip += '\nClick to unequip';
    if (item.attributes) {
      const parts = [];
      if (item.attributes.damage != null) parts.push('Dmg: ' + item.attributes.damage);
      if (item.attributes.defense != null) parts.push('Def: ' + item.attributes.defense);
      if (item.attributes.armor != null) parts.push('Armor: ' + item.attributes.armor);
      if (item.attributes.strength != null) parts.push('Str: +' + item.attributes.strength);
      if (item.attributes.agility != null) parts.push('Agi: +' + item.attributes.agility);
      if (item.attributes.intelligence != null) parts.push('Int: +' + item.attributes.intelligence);
      if (parts.length) tip += '\n' + parts.join(', ');
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
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .widget-title {
    flex: 1;
  }

  .doll {
    display: grid;
    grid-template-columns: auto 1fr auto;
    grid-template-rows: auto auto;
    gap: 0.45em 0.55em;
    padding: 0.55em 0.65em 0.35em;
    align-items: start;
    justify-items: center;
  }

  .slot-col {
    display: flex;
    flex-direction: column;
    gap: 0.4em;
  }

  .slot-col.left { grid-column: 1; grid-row: 1; }
  .slot-col.right { grid-column: 3; grid-row: 1; }

  .portrait-col {
    grid-column: 2;
    grid-row: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.35em;
    min-width: 0;
    width: 100%;
    max-width: 140px;
    align-self: stretch;
  }

  .portrait-frame {
    width: 100%;
    aspect-ratio: 2 / 3;
    max-height: 210px;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: rgba(0, 0, 0, 0.35);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .portrait-frame img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    image-rendering: pixelated;
  }

  .portrait-plate {
    text-align: center;
    width: 100%;
  }

  .portrait-name {
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text-primary, #e5e7eb);
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .portrait-meta {
    font-size: 0.68rem;
    color: var(--text-dim, #94a3b8);
    line-height: 1.25;
  }

  .weapon-row {
    grid-column: 1 / -1;
    grid-row: 2;
    display: flex;
    justify-content: center;
    gap: 0.55em;
  }

  .equip-slot {
    width: 72px;
    height: 72px;
    border-radius: 6px;
    border: 1.5px solid rgba(148, 163, 184, 0.35);
    background: rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    flex-shrink: 0;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .equip-slot.empty {
    opacity: 0.7;
    cursor: default;
    border-style: dashed;
  }

  .equip-slot.filled {
    cursor: pointer;
  }

  .equip-slot.filled:hover {
    background: rgba(239, 68, 68, 0.12);
  }

  .equip-slot img {
    width: 64px;
    height: 64px;
    object-fit: contain;
    image-rendering: pixelated;
  }

  .equip-slot .slot-glyph {
    font-size: 26px;
    color: #4b5563;
  }

  .equip-slot .slot-tag {
    position: absolute;
    left: 3px;
    top: 2px;
    font-size: 0.55rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgba(148, 163, 184, 0.85);
    line-height: 1;
    pointer-events: none;
  }

  .compact-stats {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.35em 0.75em;
    padding: 0.35em 0.65em 0.65em;
    border-top: 1px solid rgba(148, 163, 184, 0.15);
    font-size: 0.72rem;
    color: var(--text-dim, #94a3b8);
  }

  .compact-stats span strong {
    color: var(--text-primary, #e5e7eb);
    font-weight: 600;
  }

  @media (max-width: 420px) {
    .equip-slot {
      width: 64px;
      height: 64px;
    }
    .equip-slot img {
      width: 56px;
      height: 56px;
    }
    .portrait-col {
      max-width: 110px;
    }
  }
</style>

<div class="equipment-widget game-panel">
  <div class="game-panel-header">
    <i class="material-icons">shield</i>
    <span class="widget-title">Equipment</span>
  </div>

  <div class="doll">
    <div class="slot-col left">
      {#each leftSlots as slot}
        {@const item = equippedItems[slot.key]}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div
          class="equip-slot"
          class:empty={!item}
          class:filled={!!item}
          style="border-color: {item ? getQualityColor(item.quality) : 'rgba(148, 163, 184, 0.35)'}"
          title={getItemTooltip(item, slot.label)}
          on:click={() => item && handleUnequip(item)}
        >
          <span class="slot-tag">{slot.label}</span>
          {#if item}
            <img src={itemArtSrc(item)} alt="" on:error={(e) => onItemArtError(e, item)} />
          {:else}
            <i class="material-icons slot-glyph">{getSlotIcon(slot.key)}</i>
          {/if}
        </div>
      {/each}
    </div>

    <div class="portrait-col">
      <div class="portrait-frame">
        {#if character}
          <img
            src={portraitSrc(character)}
            alt=""
            on:error={(e) => onPortraitError(e, character)}
          />
        {:else}
          <i class="material-icons" style="font-size:48px;color:#4b5563">person</i>
        {/if}
      </div>
      <div class="portrait-plate">
        <div class="portrait-name">{name}</div>
        <div class="portrait-meta">Lv {level} {race} {charClass}</div>
      </div>
    </div>

    <div class="slot-col right">
      {#each rightSlots as slot}
        {@const item = equippedItems[slot.key]}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div
          class="equip-slot"
          class:empty={!item}
          class:filled={!!item}
          style="border-color: {item ? getQualityColor(item.quality) : 'rgba(148, 163, 184, 0.35)'}"
          title={getItemTooltip(item, slot.label)}
          on:click={() => item && handleUnequip(item)}
        >
          <span class="slot-tag">{slot.label}</span>
          {#if item}
            <img src={itemArtSrc(item)} alt="" on:error={(e) => onItemArtError(e, item)} />
          {:else}
            <i class="material-icons slot-glyph">{getSlotIcon(slot.key)}</i>
          {/if}
        </div>
      {/each}
    </div>

    <div class="weapon-row">
      {#each weaponSlots as slot}
        {@const item = equippedItems[slot.key]}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div
          class="equip-slot"
          class:empty={!item}
          class:filled={!!item}
          style="border-color: {item ? getQualityColor(item.quality) : 'rgba(148, 163, 184, 0.35)'}"
          title={getItemTooltip(item, slot.label)}
          on:click={() => item && handleUnequip(item)}
        >
          <span class="slot-tag">{slot.label}</span>
          {#if item}
            <img src={itemArtSrc(item)} alt="" on:error={(e) => onItemArtError(e, item)} />
          {:else}
            <i class="material-icons slot-glyph">{getSlotIcon(slot.key)}</i>
          {/if}
        </div>
      {/each}
    </div>
  </div>

  <div class="compact-stats">
    <span>HP <strong>{currentHp}/{maxHp}</strong></span>
    {#if maxMana > 0}
      <span>MP <strong>{currentMana}/{maxMana}</strong></span>
    {/if}
    <span>ATK <strong>{attackPower}</strong></span>
    <span>DEF <strong>{defense}</strong></span>
  </div>
</div>
