<style>
  .mobile-action-bar {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 8px 10px;
    flex-shrink: 0;
  }

  .action-row {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-items: center;
    justify-content: center;
  }

  .separator {
    width: 1px;
    height: 28px;
    background: rgba(255, 255, 255, 0.12);
    flex-shrink: 0;
  }

  .btn {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    padding: 8px 12px;
    color: #e5e7eb;
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    min-height: 44px;
    min-width: 44px;
    transition: all 0.15s ease;
    -webkit-tap-highlight-color: transparent;
    margin: 0;
  }

  .btn:active {
    transform: scale(0.95);
    background: rgba(255, 255, 255, 0.15);
  }

  .btn i {
    font-size: 16px;
  }

  .exit-btn {
    background: rgba(249, 115, 22, 0.2);
    border-color: rgba(249, 115, 22, 0.4);
    color: #fdba74;
  }

  .exit-btn:active {
    background: rgba(249, 115, 22, 0.35);
  }

  .exit-btn.unavailable,
  .exit-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .cmd-btn {
    background: rgba(59, 130, 246, 0.2);
    border-color: rgba(59, 130, 246, 0.4);
    color: #93c5fd;
  }

  .cmd-btn:active {
    background: rgba(59, 130, 246, 0.35);
  }

  .combat-btn {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.4);
    color: #fca5a5;
  }

  .combat-btn:active {
    background: rgba(239, 68, 68, 0.35);
  }

  .context-btn {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .context-btn:active {
    background: rgba(34, 197, 94, 0.35);
  }

  .room-action-btn {
    background: rgba(168, 85, 247, 0.2);
    border-color: rgba(168, 85, 247, 0.4);
    color: #d8b4fe;
  }

  .room-action-btn:active {
    background: rgba(168, 85, 247, 0.35);
  }

  .more-btn {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.2);
  }

  .pickup-icon-btn {
    position: relative;
    flex: 0 0 44px;
    width: 44px;
    min-width: 44px;
    max-width: 44px;
    padding: 8px;
  }

  .pickup-count {
    position: absolute;
    top: 2px;
    right: 2px;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 8px;
    background: #22c55e;
    color: #052e16;
    font-size: 10px;
    font-weight: 800;
    line-height: 16px;
    text-align: center;
  }

  .dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: 9999;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    animation: overlayFadeIn 0.15s ease-out;
  }

  @keyframes overlayFadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .dialog {
    background: rgba(15, 15, 25, 0.95);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 14px;
    padding: 16px;
    width: 90vw;
    max-width: 400px;
    max-height: 60vh;
    overflow-y: auto;
    animation: dialogSlideIn 0.2s ease-out;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  }

  @keyframes dialogSlideIn {
    from { opacity: 0; transform: scale(0.95) translateY(10px); }
    to { opacity: 1; transform: scale(1) translateY(0); }
  }

  .dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .dialog-title {
    color: #e5e7eb;
    font-size: 13px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 1px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .dialog-title i {
    font-size: 18px;
  }

  .dialog-close {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 6px;
    color: #9ca3af;
    cursor: pointer;
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 36px;
    min-height: 36px;
    margin: 0;
  }

  .dialog-close i {
    font-size: 18px;
  }

  .dialog-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .popup-btn {
    background: rgba(59, 130, 246, 0.2);
    border: 1px solid rgba(59, 130, 246, 0.4);
    border-radius: 8px;
    padding: 10px 8px;
    color: #93c5fd;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    text-transform: uppercase;
    min-height: 44px;
  }

  .popup-btn i {
    font-size: 16px;
  }

  .pickup-popup-btn {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .dialog.dialog-say {
    border-color: rgba(96, 165, 250, 0.45);
    max-width: 360px;
  }

  .dialog-say .dialog-title {
    color: #93c5fd;
  }

  .say-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .say-input {
    width: 100%;
    box-sizing: border-box;
    padding: 12px;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.35);
    background: rgba(0, 0, 0, 0.35);
    color: #e5e7eb;
    font: inherit;
    font-size: 16px;
    min-height: 44px;
  }

  .say-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .say-actions .popup-btn {
    text-transform: none;
    flex: 1;
  }

  .pin-toggle {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 8px;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.25);
    background: rgba(255, 255, 255, 0.04);
    color: #cbd5e1;
    cursor: pointer;
    font-size: 12px;
    min-height: 44px;
  }
  .pin-toggle.pinned {
    border-color: rgba(59, 130, 246, 0.55);
    background: rgba(59, 130, 246, 0.15);
    color: #bfdbfe;
  }
  .dialog-section-label {
    grid-column: 1 / -1;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #64748b;
  }
  .pref-row {
    grid-column: 1 / -1;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    margin-top: 4px;
    padding-top: 10px;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
    color: #94a3b8;
    font-size: 11px;
  }
  .pref-row label {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: #cbd5e1;
  }
</style>

<script>
  import { getCardinalExits, getSpecialExits, getVerticalExits } from "../MUDXPlusStore";
  import { mobileStore } from './mobileStore.js';
  import { settingsStore } from '../SettingsStore.js';
  import {
    PINNABLE_COMMANDS,
    resolvePinnedCommands,
    togglePin,
    INVENTORY_OPEN_OVERLAY,
    INVENTORY_OPEN_WIDGET,
  } from '../hudPrefs.js';

  export let store;
  export let sendMessage;

  let showMoreMenu = false;
  let showPickupMenu = false;
  let showSayPrompt = false;
  let sayText = '';

  $: cardinalExits = getCardinalExits($store.exits);
  $: specialExits = getSpecialExits($store.exits);
  $: verticalExits = getVerticalExits($store.exits);
  $: allSpecialExits = [...verticalExits, ...specialExits];

  const directionMeta = {
    north: { label: "N", icon: "north" },
    south: { label: "S", icon: "south" },
    east:  { label: "E", icon: "east" },
    west:  { label: "W", icon: "west" },
  };

  const combatCommands = [
    { id: "attack", name: "attack", icon: "swords", label: "Attack", kind: "command" },
    { id: "defend", name: "defend", icon: "security", label: "Defend", kind: "command" },
    { id: "flee", name: "flee", icon: "directions_run", label: "Flee", kind: "command" },
    { id: "status", name: "status", icon: "monitor_heart", label: "Status", kind: "command" },
  ];

  $: pins = $settingsStore.interface?.actionBarPins;
  $: inventoryOpenMode = $settingsStore.interface?.inventoryOpenMode || INVENTORY_OPEN_OVERLAY;
  $: pinnedCommands = $store.inCombat ? combatCommands : resolvePinnedCommands(pins);
  $: contextCommands = $store.inCombat ? [] : [
    ...($store.hasMerchant ? [{ name: "list", icon: "store", label: "Shop" }] : []),
  ];
  $: roomActions = $store.actions || [];
  $: groundItems = ($store.groundItems || []).filter(item => !item.noPickup);

  function executeCommand(cmd) {
    sendMessage(cmd);
    closeMenus();
  }

  function activatePin(pin) {
    if (!pin) return;
    if (pin.kind === "map") {
      if (store?.openMapOverview) store.openMapOverview();
      closeMenus();
      return;
    }
    if (pin.kind === "inventory") {
      openInventory();
      return;
    }
    if (pin.kind === "say") {
      openSayPrompt();
      return;
    }
    executeCommand(pin.name);
  }

  function openSayPrompt() {
    showMoreMenu = false;
    showPickupMenu = false;
    sayText = '';
    showSayPrompt = true;
  }

  function closeSayPrompt() {
    showSayPrompt = false;
    sayText = '';
  }

  function submitSay() {
    const text = String(sayText || '').trim();
    if (!text) return;
    sendMessage(`say ${text}`);
    closeSayPrompt();
    closeMenus();
  }

  function openInventory() {
    closeMenus();
    if (inventoryOpenMode === INVENTORY_OPEN_WIDGET) {
      mobileStore.openBottomSheet('inventory');
      return;
    }
    if (store?.openInventoryOverlay) store.openInventoryOverlay();
    else mobileStore.openBottomSheet('inventory');
  }

  function pickupItem(item) {
    sendMessage('pickup ' + item.name);
    store.removeGroundItem(item.id);
  }

  function toggleMoreMenu() {
    showMoreMenu = !showMoreMenu;
    showPickupMenu = false;
  }

  function togglePickupMenu() {
    showPickupMenu = !showPickupMenu;
    showMoreMenu = false;
  }

  function closeMenus() {
    showMoreMenu = false;
    showPickupMenu = false;
  }

  function onTogglePin(id) {
    settingsStore.setSetting('interface', 'actionBarPins', togglePin(pins, id));
  }

  function setInvMode(mode) {
    settingsStore.setSetting('interface', 'inventoryOpenMode', mode);
  }

  function isPinned(id) {
    return (pins || []).includes(id);
  }
</script>

{#if showMoreMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="dialog-overlay" on:click={closeMenus}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="dialog-title">
          <i class="material-icons">tune</i>
          Customize pins
        </span>
        <button class="dialog-close" on:click={closeMenus}>
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="dialog-grid">
        <div class="dialog-section-label">Tap to pin / unpin</div>
        {#each PINNABLE_COMMANDS as cmd}
          <button
            class="pin-toggle"
            class:pinned={isPinned(cmd.id)}
            type="button"
            on:click={() => onTogglePin(cmd.id)}
          >
            <i class="material-icons">{isPinned(cmd.id) ? 'push_pin' : cmd.icon}</i>
            {cmd.label}
          </button>
        {/each}
        <div class="dialog-section-label">Run command</div>
        {#each PINNABLE_COMMANDS as cmd}
          <button class="popup-btn" type="button" on:click={() => activatePin(cmd)}>
            <i class="material-icons">{cmd.icon}</i>
            {cmd.label}
          </button>
        {/each}
        <div class="pref-row">
          Inventory:
          <label>
            <input
              type="radio"
              name="inv-mode-mobile"
              checked={inventoryOpenMode === INVENTORY_OPEN_OVERLAY}
              on:change={() => setInvMode(INVENTORY_OPEN_OVERLAY)}
            />
            Popup
          </label>
          <label>
            <input
              type="radio"
              name="inv-mode-mobile"
              checked={inventoryOpenMode === INVENTORY_OPEN_WIDGET}
              on:change={() => setInvMode(INVENTORY_OPEN_WIDGET)}
            />
            Sheet
          </label>
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showPickupMenu && groundItems.length > 0}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="dialog-overlay" on:click={closeMenus}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="dialog-title">
          <i class="material-icons">back_hand</i>
          Pick Up
        </span>
        <button class="dialog-close" on:click={closeMenus}>
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="dialog-grid">
        {#each groundItems as item}
          <button class="popup-btn pickup-popup-btn" on:click={() => pickupItem(item)}>
            {item.name}
          </button>
        {/each}
      </div>
    </div>
  </div>
{/if}

{#if showSayPrompt}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="dialog-overlay" on:click={closeSayPrompt}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="dialog dialog-say" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="dialog-title">
          <i class="material-icons">chat</i>
          Say
        </span>
        <button class="dialog-close" type="button" on:click={closeSayPrompt}>
          <i class="material-icons">close</i>
        </button>
      </div>
      <form class="say-form" on:submit|preventDefault={submitSay}>
        <input
          class="say-input"
          type="text"
          bind:value={sayText}
          placeholder="What do you say?"
          maxlength="240"
          autofocus
          aria-label="Say message"
        />
        <div class="say-actions">
          <button class="popup-btn" type="button" on:click={closeSayPrompt}>Cancel</button>
          <button class="popup-btn" type="submit" disabled={!String(sayText || '').trim()}>
            Send
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<div class="mobile-action-bar">
  <div class="action-row">
    {#each cardinalExits as exit}
      <button
        class="btn exit-btn"
        class:unavailable={!exit.available}
        disabled={!exit.available}
        on:click={() => exit.available && executeCommand(exit.name)}
      >
        <i class="material-icons">{directionMeta[exit.name].icon}</i>
        {directionMeta[exit.name].label}
      </button>
    {/each}
    {#each allSpecialExits as exit}
      <button class="btn exit-btn" on:click={() => executeCommand(exit.name)}>
        {exit.name}
      </button>
    {/each}
    {#each roomActions as action}
      <button class="btn room-action-btn" on:click={() => executeCommand(action.name)}>
        {action.name}
      </button>
    {/each}

    <div class="separator"></div>

    {#each pinnedCommands as cmd}
      <button
        class="btn {$store.inCombat ? 'combat-btn' : 'cmd-btn'}"
        on:click={() => activatePin(cmd)}
      >
        <i class="material-icons">{cmd.icon}</i>
        {cmd.label}
      </button>
    {/each}

    {#each contextCommands as cmd}
      <button class="btn context-btn" on:click={() => executeCommand(cmd.name)}>
        <i class="material-icons">{cmd.icon}</i>
        {cmd.label}
      </button>
    {/each}


    <button class="btn more-btn" on:click={toggleMoreMenu}>
      <i class="material-icons">more_horiz</i>
    </button>
  </div>
</div>
