<script>
  import { settingsStore } from '../SettingsStore.js';
  import { overlayStore } from '../ui/overlayStore.js';
  import { itemArtSrc, onItemArtError } from '../itemArtSrc.js';
  import {
    HOTBAR_ACTIONS,
    findInventoryItem,
    isConsumableItem,
    makeActionBind,
    makeItemBind,
    makeSkillBind,
    normalizeHotbarBinds,
    resolveHotbarActivation,
    skillDisplayName,
    skillGenericArtUrl,
    actionGenericArtUrl,
  } from '../hudPrefs.js';

  export let store;
  export let sendMessage;
  /** Compact strip for mobile (no widget chrome). */
  export let compact = false;

  let pickerIndex = -1;
  let longPressTimer = null;

  $: binds = normalizeHotbarBinds($settingsStore.interface?.hotbarBinds);
  $: inventory = $store.inventory || [];
  $: equippedSkills = $store.characterStats?.equippedSkills || [];
  $: inCombat = !!(
    $store.inCombat || $store.characterStats?.inCombat
  );
  $: consumables = inventory.filter(isConsumableItem);

  function saveBinds(next) {
    settingsStore.setSetting('interface', 'hotbarBinds', normalizeHotbarBinds(next));
  }

  function toast(text) {
    if (overlayStore?.pushMessage) overlayStore.pushMessage(text);
  }

  function openPicker(index) {
    pickerIndex = index;
  }

  function closePicker() {
    pickerIndex = -1;
  }

  function clearSlot(index) {
    const next = [...binds];
    next[index] = null;
    saveBinds(next);
    closePicker();
  }

  function bindSkill(skillId) {
    if (pickerIndex < 0) return;
    const next = [...binds];
    next[pickerIndex] = makeSkillBind(skillId);
    saveBinds(next);
    closePicker();
  }

  function bindItem(item) {
    if (pickerIndex < 0) return;
    const next = [...binds];
    next[pickerIndex] = makeItemBind(item);
    saveBinds(next);
    closePicker();
  }

  function bindAction(actionId) {
    if (pickerIndex < 0) return;
    const next = [...binds];
    next[pickerIndex] = makeActionBind(actionId);
    saveBinds(next);
    closePicker();
  }

  function slotDisabled(bind) {
    if (!bind) return false;
    if (bind.kind === 'skill' && !inCombat) return true;
    if (bind.kind === 'item' && !findInventoryItem(inventory, bind)) return true;
    return false;
  }

  function activateSlot(index) {
    const bind = binds[index];
    if (!bind) {
      openPicker(index);
      return;
    }
    const result = resolveHotbarActivation(bind, { inCombat, inventory });
    if (!result.ok) {
      if (result.reason && result.reason !== 'empty') toast(result.reason);
      return;
    }
    if (result.command && sendMessage) sendMessage(result.command);
  }

  function onSlotContext(index, event) {
    event.preventDefault();
    openPicker(index);
  }

  function onPointerDown(index) {
    clearTimeout(longPressTimer);
    longPressTimer = setTimeout(() => {
      openPicker(index);
      longPressTimer = null;
    }, 550);
  }

  function onPointerUp() {
    clearTimeout(longPressTimer);
    longPressTimer = null;
  }

  function slotTitle(bind) {
    if (!bind) return 'Empty — click to bind';
    if (bind.kind === 'skill') {
      const name = bind.name || skillDisplayName(bind.id);
      return inCombat ? `Cast ${name}` : `${name} (combat only)`;
    }
    if (bind.kind === 'item') {
      const item = findInventoryItem(inventory, bind);
      return item ? `Use ${item.name}` : `${bind.name || 'Item'} (missing)`;
    }
    if (bind.kind === 'action') {
      return bind.name || bind.id || 'Action';
    }
    return 'Empty';
  }
</script>

<style>
  .hotbar-widget {
    height: 100%;
    min-height: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25em 0.5em;
    box-sizing: border-box;
  }

  .hotbar-widget.compact {
    padding: 6px 8px 4px;
    background: rgba(0, 0, 0, 0.45);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .slots {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.4em;
    width: 100%;
    max-width: 560px;
  }

  .slot {
    flex: 0 0 auto;
    width: clamp(40px, 7.2vw, 56px);
    aspect-ratio: 1;
    border-radius: 8px;
    border: 1.5px dashed rgba(148, 163, 184, 0.35);
    background: rgba(0, 0, 0, 0.35);
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    cursor: pointer;
    padding: 0;
    color: #94a3b8;
    transition: border-color 0.15s ease, background 0.15s ease, opacity 0.15s ease;
  }

  .slot.filled {
    border-style: solid;
    border-color: rgba(148, 163, 184, 0.45);
    background: rgba(15, 23, 42, 0.75);
  }

  .slot.skill.filled {
    border-color: rgba(167, 139, 250, 0.55);
  }

  .slot.item.filled {
    border-color: rgba(96, 165, 250, 0.5);
  }

  .slot:hover:not(.disabled) {
    border-color: rgba(251, 191, 36, 0.65);
    background: rgba(30, 41, 59, 0.9);
  }

  .slot.disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .slot img {
    width: 70%;
    height: 70%;
    object-fit: contain;
    image-rendering: pixelated;
    pointer-events: none;
  }

  .slot i {
    font-size: clamp(18px, 3.2vw, 26px);
    color: #c4b5fd;
    pointer-events: none;
  }

  .slot-index {
    position: absolute;
    left: 3px;
    top: 2px;
    font-size: 0.55rem;
    color: rgba(148, 163, 184, 0.7);
    line-height: 1;
    pointer-events: none;
  }

  .picker-overlay {
    position: fixed;
    inset: 0;
    z-index: 10050;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }

  .picker {
    width: min(92vw, 360px);
    max-height: min(70vh, 480px);
    overflow: auto;
    background: rgba(12, 14, 22, 0.96);
    border: 1px solid rgba(148, 163, 184, 0.3);
    border-radius: 12px;
    padding: 0.9rem 1rem 1rem;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.55);
  }

  .picker-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.75rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
  }

  .picker-title {
    font-size: 0.85rem;
    font-weight: 600;
    color: #e5e7eb;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .picker-close {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    font-size: 1.25rem;
    line-height: 1;
    padding: 0.15rem;
  }

  .section-label {
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #64748b;
    margin: 0.65rem 0 0.35rem;
  }

  .pick-list {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .pick-btn {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    width: 100%;
    text-align: left;
    padding: 0.5rem 0.6rem;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.25);
    background: rgba(255, 255, 255, 0.04);
    color: #e5e7eb;
    cursor: pointer;
    font: inherit;
    font-size: 0.85rem;
  }

  .pick-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(251, 191, 36, 0.45);
  }

  .pick-btn img {
    width: 32px;
    height: 32px;
    object-fit: contain;
    image-rendering: pixelated;
  }

  .pick-btn i {
    font-size: 22px;
    color: #c4b5fd;
    width: 32px;
    text-align: center;
  }

  .empty-hint {
    font-size: 0.78rem;
    color: #64748b;
    font-style: italic;
    padding: 0.25rem 0.15rem;
  }

  .clear-btn {
    margin-top: 0.75rem;
    width: 100%;
    padding: 0.5rem;
    border-radius: 8px;
    border: 1px solid rgba(239, 68, 68, 0.35);
    background: rgba(239, 68, 68, 0.1);
    color: #fca5a5;
    cursor: pointer;
    font: inherit;
    font-size: 0.8rem;
  }
</style>

<div class="hotbar-widget" class:compact aria-label="Spell bar">
  <div class="slots">
    {#each binds as bind, index}
      {@const item = bind?.kind === 'item' ? findInventoryItem(inventory, bind) : null}
      <button
        type="button"
        class="slot"
        class:filled={!!bind}
        class:skill={bind?.kind === 'skill'}
        class:item={bind?.kind === 'item'}
        class:disabled={slotDisabled(bind)}
        title={slotTitle(bind)}
        aria-label={slotTitle(bind)}
        on:click={() => activateSlot(index)}
        on:contextmenu={(e) => onSlotContext(index, e)}
        on:pointerdown={() => onPointerDown(index)}
        on:pointerup={onPointerUp}
        on:pointerleave={onPointerUp}
        on:pointercancel={onPointerUp}
      >
        <span class="slot-index">{index + 1}</span>
        {#if bind?.kind === 'item'}
          <img
            src={itemArtSrc(item || { name: bind.name, templateId: bind.id, type: 'consumable' })}
            alt=""
            on:error={(e) => onItemArtError(e, item || { type: 'consumable' })}
          />
        {:else if bind?.kind === 'skill'}
          <img
            src={skillGenericArtUrl(bind.id || bind.name)}
            alt=""
            on:error={(e) => onItemArtError(e, { type: 'default' })}
          />
        {:else if bind?.kind === 'action'}
          <img
            src={actionGenericArtUrl(bind.id)}
            alt=""
            on:error={(e) => onItemArtError(e, { type: 'default' })}
          />
        {/if}
      </button>
    {/each}
  </div>
</div>

{#if pickerIndex >= 0}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="picker-overlay" on:click={closePicker}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="picker" on:click|stopPropagation role="dialog" aria-label="Bind hotbar slot">
      <div class="picker-header">
        <span class="picker-title">Bind slot {pickerIndex + 1}</span>
        <button type="button" class="picker-close" on:click={closePicker} aria-label="Close">×</button>
      </div>

      <div class="section-label">Skills (equipped)</div>
      {#if equippedSkills.length === 0}
        <div class="empty-hint">spellbook empty — equip skills with `skills equip &lt;name&gt;`</div>
      {:else}
        <div class="pick-list">
          {#each equippedSkills as skillId}
            <button type="button" class="pick-btn" on:click={() => bindSkill(skillId)}>
              <img src={skillGenericArtUrl(skillId)} alt="" on:error={(e) => onItemArtError(e, { type: 'default' })} />
              {skillDisplayName(skillId)}
            </button>
          {/each}
        </div>
      {/if}

      <div class="section-label">Actions</div>
      <div class="pick-list">
        {#each HOTBAR_ACTIONS as action}
          <button type="button" class="pick-btn" on:click={() => bindAction(action.id)}>
            <img src={actionGenericArtUrl(action.id)} alt="" on:error={(e) => onItemArtError(e, { type: 'default' })} />
            {action.label}
          </button>
        {/each}
      </div>

      <div class="section-label">Consumables</div>
      {#if consumables.length === 0}
        <div class="empty-hint">No consumables in inventory</div>
      {:else}
        <div class="pick-list">
          {#each consumables as item}
            <button type="button" class="pick-btn" on:click={() => bindItem(item)}>
              <img src={itemArtSrc(item)} alt="" on:error={(e) => onItemArtError(e, item)} />
              {item.name}
            </button>
          {/each}
        </div>
      {/if}

      {#if binds[pickerIndex]}
        <button type="button" class="clear-btn" on:click={() => clearSlot(pickerIndex)}>
          Clear slot
        </button>
      {/if}
    </div>
  </div>
{/if}
