<style>
  .mudx {
    display: flex;
    flex-direction: column;
    height: 100%;
    max-width: 100%;
    padding: 0 1em;
    position: relative;
  }


  .action-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5em;
    align-items: center;
    justify-content: center;
    padding: 0.45em 0.85em;
    background: var(--panel-bg, rgba(0, 0, 0, 0.6));
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border-radius: var(--panel-radius, 12px);
    border: 1px solid var(--panel-border, rgba(255, 255, 255, 0.1));
    box-shadow: var(--panel-shadow, none);
    flex-shrink: 0;
    margin-bottom: 1em;
  }

  .exits-section {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
    align-items: center;
  }

  .commands-section {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
    align-items: center;
  }

  /* Glass button base style */
  .btn {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    padding: 0.5em 0.85em;
    color: var(--text-primary, #e5e7eb);
    font-size: var(--text-sm, 12px);
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.3em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .btn:hover {
    background: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.25);
    transform: translateY(-1px);
  }

  .btn:active {
    transform: translateY(0);
  }

  .btn i {
    font-size: 16px;
  }

  /* Exit buttons - orange */
  .exit-btn {
    background: rgba(249, 115, 22, 0.2);
    border-color: rgba(249, 115, 22, 0.4);
    color: #fdba74;
  }

  .exit-btn:hover {
    background: rgba(249, 115, 22, 0.35);
    border-color: rgba(249, 115, 22, 0.6);
  }

  /* Command buttons - blue */
  .cmd-btn {
    background: rgba(59, 130, 246, 0.2);
    border-color: rgba(59, 130, 246, 0.4);
    color: #93c5fd;
  }

  .cmd-btn:hover {
    background: rgba(59, 130, 246, 0.35);
    border-color: rgba(59, 130, 246, 0.6);
  }

  /* Combat buttons - red */
  .combat-btn {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.4);
    color: #fca5a5;
  }

  .combat-btn:hover {
    background: rgba(239, 68, 68, 0.35);
    border-color: rgba(239, 68, 68, 0.6);
  }

  /* Context buttons - green */
  .context-btn {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .context-btn:hover, .context-btn.active {
    background: rgba(34, 197, 94, 0.35);
    border-color: rgba(34, 197, 94, 0.6);
  }

  /* Room action buttons - purple */
  .room-action-btn {
    background: rgba(168, 85, 247, 0.2);
    border-color: rgba(168, 85, 247, 0.4);
    color: #d8b4fe;
  }

  .room-action-btn:hover {
    background: rgba(168, 85, 247, 0.35);
    border-color: rgba(168, 85, 247, 0.6);
  }

  /* Room actions overflow button - purple */
  .room-actions-btn {
    background: rgba(168, 85, 247, 0.2);
    border-color: rgba(168, 85, 247, 0.4);
    color: #d8b4fe;
  }

  .room-actions-btn:hover, .room-actions-btn.active {
    background: rgba(168, 85, 247, 0.35);
    border-color: rgba(168, 85, 247, 0.6);
  }

  .separator {
    width: 1px;
    height: 24px;
    background: var(--panel-border, rgba(255, 255, 255, 0.15));
    margin: 0 0.3em;
    flex-shrink: 0;
  }

  .more-btn {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.2);
    font-weight: 700;
    letter-spacing: 2px;
  }

  .more-btn:hover, .more-btn.active {
    background: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .exit-btn.unavailable,
  .exit-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
    transform: none;
  }
  .exit-btn.unavailable:hover,
  .exit-btn:disabled:hover {
    background: rgba(249, 115, 22, 0.2);
    border-color: rgba(249, 115, 22, 0.4);
    transform: none;
  }

  .pin-toggle {
    display: flex;
    align-items: center;
    gap: 0.45em;
    padding: 0.45em 0.55em;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.25);
    background: rgba(255, 255, 255, 0.04);
    color: #cbd5e1;
    cursor: pointer;
    font-size: 12px;
  }
  .pin-toggle.pinned {
    border-color: rgba(59, 130, 246, 0.55);
    background: rgba(59, 130, 246, 0.15);
    color: #bfdbfe;
  }
  .pin-toggle i {
    font-size: 16px;
  }
  .pref-row {
    grid-column: 1 / -1;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5em;
    align-items: center;
    margin-top: 0.35em;
    padding-top: 0.65em;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
    color: #94a3b8;
    font-size: 11px;
  }
  .pref-row label {
    display: inline-flex;
    align-items: center;
    gap: 0.35em;
    cursor: pointer;
    color: #cbd5e1;
  }
  .dialog-section-label {
    grid-column: 1 / -1;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #64748b;
    margin-top: 0.25em;
  }

  /* Popup buttons */
  .popup-btn {
    background: rgba(59, 130, 246, 0.2);
    border: 1px solid rgba(59, 130, 246, 0.4);
    border-radius: 8px;
    padding: 0.5em 0.85em;
    color: #93c5fd;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.3em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .popup-btn:hover {
    background: rgba(59, 130, 246, 0.35);
    border-color: rgba(59, 130, 246, 0.6);
    transform: translateY(-1px);
  }

  .popup-btn:active {
    transform: translateY(0);
  }

  .popup-btn i {
    font-size: 16px;
  }

  /* Purple popup buttons for room actions */
  .popup-btn.room-action-popup-btn {
    background: rgba(168, 85, 247, 0.2);
    border-color: rgba(168, 85, 247, 0.4);
    color: #d8b4fe;
  }

  .popup-btn.room-action-popup-btn:hover {
    background: rgba(168, 85, 247, 0.35);
    border-color: rgba(168, 85, 247, 0.6);
  }

  /* Green popup buttons for pickup items */
  .popup-btn.pickup-popup-btn {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .popup-btn.pickup-popup-btn:hover {
    background: rgba(34, 197, 94, 0.35);
    border-color: rgba(34, 197, 94, 0.6);
  }

  /* Dialog overlay for actions */
  .dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
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
    border: 1px solid rgba(168, 85, 247, 0.3);
    border-radius: 14px;
    padding: 1.2em;
    min-width: 320px;
    max-width: 520px;
    width: 90vw;
    max-height: 70vh;
    overflow-y: auto;
    animation: dialogSlideIn 0.2s ease-out;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(168, 85, 247, 0.1);
  }

  .dialog.dialog-more {
    border-color: rgba(255, 255, 255, 0.2);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.05);
  }

  .dialog-more .dialog-title {
    color: #e5e7eb;
  }

  .dialog.dialog-pickup {
    border-color: rgba(34, 197, 94, 0.3);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(34, 197, 94, 0.1);
  }

  .dialog-pickup .dialog-title {
    color: #86efac;
  }

  @keyframes dialogSlideIn {
    from {
      opacity: 0;
      transform: scale(0.95) translateY(10px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  .dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1em;
    padding-bottom: 0.6em;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .dialog-title {
    color: #d8b4fe;
    font-size: 13px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 1px;
    display: flex;
    align-items: center;
    gap: 0.4em;
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
    padding: 0.2em;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .dialog-close:hover {
    background: rgba(255, 255, 255, 0.15);
    color: #e5e7eb;
  }

  .dialog-close i {
    font-size: 18px;
  }

  .dialog-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 0.5em;
  }

  /* Badge for action count */
  .action-count {
    background: rgba(168, 85, 247, 0.5);
    border-radius: 8px;
    padding: 0.1em 0.4em;
    font-size: 10px;
    font-weight: 700;
    margin-left: 0.2em;
  }

  /* Responsive adjustments */
  @media screen and (max-width: 600px) {
    .action-bar {
      padding: 0.5em;
      gap: 0.4em;
    }

    .btn {
      padding: 0.4em 0.6em;
      font-size: 11px;
    }

    .btn i {
      font-size: 14px;
    }

    .dialog {
      min-width: 0;
      width: 95vw;
      max-height: 60vh;
    }

    .dialog-grid {
      grid-template-columns: repeat(2, 1fr);
    }

    .popup-btn {
      font-size: 11px;
      padding: 0.4em 0.6em;
    }

    .popup-btn i {
      font-size: 14px;
    }
  }
</style>

<script>
  import { getCardinalExits, getSpecialExits, getVerticalExits } from "./MUDXPlusStore";
  import { settingsStore } from "./SettingsStore.js";
  import {
    PINNABLE_COMMANDS,
    resolvePinnedCommands,
    togglePin,
    INVENTORY_OPEN_OVERLAY,
    INVENTORY_OPEN_WIDGET,
  } from "./hudPrefs.js";

  export let store;
  // svelte-ignore unused-export-let
  export let term;
  export let sendMessage;

  let showMoreMenu = false;
  let showPickupMenu = false;

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
    executeCommand(pin.name);
  }

  function openInventory() {
    closeMenus();
    if (inventoryOpenMode === INVENTORY_OPEN_WIDGET) {
      // Prefer on-screen inventory widget when chosen; still open overlay if user has none.
      executeCommand("i");
      return;
    }
    if (store?.openInventoryOverlay) store.openInventoryOverlay();
    else executeCommand("i");
  }

  function pickupItem(item) {
    sendMessage("pickup " + item.name);
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
    const next = togglePin(pins, id);
    settingsStore.setSetting("interface", "actionBarPins", next);
  }

  function setInvMode(mode) {
    settingsStore.setSetting("interface", "inventoryOpenMode", mode);
  }

  function isPinned(id) {
    return (pins || []).includes(id);
  }
</script>

<div class="mudx">
  {#if showMoreMenu}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="dialog-overlay" on:click={closeMenus}>
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="dialog dialog-more" on:click|stopPropagation>
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
          <div class="dialog-section-label">Tap to pin / unpin on the action bar</div>
          {#each PINNABLE_COMMANDS as cmd}
            <button
              class="pin-toggle"
              class:pinned={isPinned(cmd.id)}
              type="button"
              on:click={() => onTogglePin(cmd.id)}
              title={isPinned(cmd.id) ? `Unpin ${cmd.label}` : `Pin ${cmd.label}`}
            >
              <i class="material-icons">{isPinned(cmd.id) ? "push_pin" : cmd.icon}</i>
              {cmd.label}
            </button>
          {/each}
          <div class="dialog-section-label">Run command</div>
          {#each PINNABLE_COMMANDS as cmd}
            <button
              class="popup-btn"
              type="button"
              on:click={() => activatePin(cmd)}
              title={cmd.label}
            >
              <i class="material-icons">{cmd.icon}</i>
              {cmd.label}
            </button>
          {/each}
          <div class="pref-row">
            Inventory opens as:
            <label>
              <input
                type="radio"
                name="inv-mode-desktop"
                checked={inventoryOpenMode === INVENTORY_OPEN_OVERLAY}
                on:change={() => setInvMode(INVENTORY_OPEN_OVERLAY)}
              />
              Popup overlay
            </label>
            <label>
              <input
                type="radio"
                name="inv-mode-desktop"
                checked={inventoryOpenMode === INVENTORY_OPEN_WIDGET}
                on:change={() => setInvMode(INVENTORY_OPEN_WIDGET)}
              />
              On-screen / log
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
      <div class="dialog dialog-pickup" on:click|stopPropagation>
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
            <button
              class="popup-btn pickup-popup-btn"
              on:click={() => pickupItem(item)}
              title={item.name}
            >
              {item.name}
            </button>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <div class="action-bar">
    <!-- Cardinals always; special/vertical when present; room actions adjacent -->
    <div class="exits-section">
      {#each cardinalExits as exit}
        <button
          class="btn exit-btn"
          class:unavailable={!exit.available}
          disabled={!exit.available}
          on:click={() => exit.available && executeCommand(exit.name)}
          title={exit.available ? exit.name : `${exit.name} (no exit)`}
        >
          <i class="material-icons">{directionMeta[exit.name].icon}</i>
          {directionMeta[exit.name].label}
        </button>
      {/each}
      {#each allSpecialExits as exit}
        <button
          class="btn exit-btn"
          on:click={() => executeCommand(exit.name)}
          title={exit.name}
        >
          {exit.name}
        </button>
      {/each}
      {#each roomActions as action}
        <button
          class="btn room-action-btn"
          on:click={() => executeCommand(action.name)}
          title={action.description || action.name}
        >
          {action.name}
        </button>
      {/each}
    </div>

    <div class="separator"></div>

    <div class="commands-section">
      {#each pinnedCommands as cmd}
        <button
          class="btn {$store.inCombat ? 'combat-btn' : 'cmd-btn'}"
          on:click={() => activatePin(cmd)}
          title={cmd.label}
        >
          <i class="material-icons">{cmd.icon}</i>
          {cmd.label}
        </button>
      {/each}

      {#each contextCommands as cmd}
        <button
          class="btn context-btn"
          on:click={() => executeCommand(cmd.name)}
          title={cmd.label}
        >
          <i class="material-icons">{cmd.icon}</i>
          {cmd.label}
        </button>
      {/each}
    </div>

    {#if groundItems.length > 0}
      <button
        class="btn context-btn"
        class:active={showPickupMenu}
        on:click={togglePickupMenu}
        title="Pick up items"
      >
        <i class="material-icons">back_hand</i>
        Pickup
      </button>
    {/if}

    <button
      class="btn more-btn"
      class:active={showMoreMenu}
      on:click={toggleMoreMenu}
      title="Customize pinned commands"
    >
      <i class="material-icons">more_horiz</i>
    </button>
  </div>
</div>
