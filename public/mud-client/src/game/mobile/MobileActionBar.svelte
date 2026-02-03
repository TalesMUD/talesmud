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

  /* Base button: 44px min touch target */
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

  /* Exit buttons - orange */
  .exit-btn {
    background: rgba(249, 115, 22, 0.2);
    border-color: rgba(249, 115, 22, 0.4);
    color: #fdba74;
  }

  .exit-btn:active {
    background: rgba(249, 115, 22, 0.35);
  }

  /* Command buttons - blue */
  .cmd-btn {
    background: rgba(59, 130, 246, 0.2);
    border-color: rgba(59, 130, 246, 0.4);
    color: #93c5fd;
  }

  .cmd-btn:active {
    background: rgba(59, 130, 246, 0.35);
  }

  /* Combat buttons - red */
  .combat-btn {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.4);
    color: #fca5a5;
  }

  .combat-btn:active {
    background: rgba(239, 68, 68, 0.35);
  }

  /* Context buttons - green */
  .context-btn {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .context-btn:active {
    background: rgba(34, 197, 94, 0.35);
  }

  /* Room action buttons - purple */
  .room-action-btn {
    background: rgba(168, 85, 247, 0.2);
    border-color: rgba(168, 85, 247, 0.4);
    color: #d8b4fe;
  }

  .room-action-btn:active {
    background: rgba(168, 85, 247, 0.35);
  }

  /* More button */
  .more-btn {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.2);
  }

  /* Quick access row */
  .quick-access-row {
    display: flex;
    gap: 6px;
    justify-content: center;
  }

  .quick-btn {
    flex: 1;
    max-width: 120px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 8px 6px;
    color: #9ca3af;
    font-size: 10px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    min-height: 44px;
    transition: all 0.15s ease;
    -webkit-tap-highlight-color: transparent;
    margin: 0;
  }

  .quick-btn:active {
    transform: scale(0.95);
    background: rgba(255, 255, 255, 0.12);
    color: #e5e7eb;
  }

  .quick-btn i {
    font-size: 18px;
  }

  /* Dialog overlay for menus (reused pattern from MUDXPlus) */
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
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    min-height: 44px;
    margin: 0;
  }

  .popup-btn:active {
    transform: scale(0.95);
  }

  .popup-btn i {
    font-size: 16px;
  }

  .popup-btn.room-action-popup-btn {
    background: rgba(168, 85, 247, 0.2);
    border-color: rgba(168, 85, 247, 0.4);
    color: #d8b4fe;
  }

  .popup-btn.pickup-popup-btn {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .action-count {
    background: rgba(168, 85, 247, 0.5);
    border-radius: 8px;
    padding: 1px 5px;
    font-size: 10px;
    font-weight: 700;
  }
</style>

<script>
  import { getCardinalExits, getSpecialExits, getVerticalExits } from "../MUDXPlusStore";
  import { mobileStore } from './mobileStore.js';

  export let store;
  export let sendMessage;

  // Menu state
  let showMoreMenu = false;
  let showActionsMenu = false;
  let showPickupMenu = false;

  // Derive exits from store
  $: cardinalExits = getCardinalExits($store.exits);
  $: specialExits = getSpecialExits($store.exits);
  $: verticalExits = getVerticalExits($store.exits);
  $: availableCardinals = cardinalExits.filter(e => e.available);
  $: allSpecialExits = [...verticalExits, ...specialExits];
  $: allExits = [...availableCardinals, ...allSpecialExits];

  const directionMeta = {
    north: { label: "N", icon: "north" },
    south: { label: "S", icon: "south" },
    east:  { label: "E", icon: "east" },
    west:  { label: "W", icon: "west" },
  };

  // Standard commands
  const standardCommands = [
    { name: "look", icon: "visibility", label: "Look" },
  ];

  // Combat commands
  const combatCommands = [
    { name: "attack", icon: "swords", label: "Attack" },
    { name: "defend", icon: "security", label: "Defend" },
    { name: "flee", icon: "directions_run", label: "Flee" },
    { name: "status", icon: "monitor_heart", label: "Status" },
  ];

  // More menu commands
  const moreCommands = [
    { name: "look", icon: "visibility", label: "Look" },
    { name: "who", icon: "people", label: "Who" },
    { name: "bind", icon: "location_on", label: "Bind" },
    { name: "drop", icon: "delete", label: "Drop" },
    { name: "use", icon: "touch_app", label: "Use" },
    { name: "examine", icon: "search", label: "Examine" },
    { name: "shrug", icon: "sentiment_neutral", label: "Shrug" },
    { name: "scream", icon: "campaign", label: "Scream" },
    { name: "help", icon: "help", label: "Help" },
  ];

  $: activeCommands = $store.inCombat ? combatCommands : standardCommands;
  $: contextCommands = $store.inCombat ? [] : [
    ...($store.hasMerchant ? [{ name: "list", icon: "store", label: "Shop" }] : []),
  ];
  $: roomActions = $store.actions || [];
  $: groundItems = ($store.groundItems || []).filter(item => !item.noPickup);

  function executeCommand(cmd) {
    sendMessage(cmd);
    closeMenus();
  }

  function pickupItem(item) {
    sendMessage('pickup ' + item.name);
    store.removeGroundItem(item.id);
  }

  function toggleMoreMenu() {
    showMoreMenu = !showMoreMenu;
    showActionsMenu = false;
    showPickupMenu = false;
  }

  function toggleActionsMenu() {
    showActionsMenu = !showActionsMenu;
    showMoreMenu = false;
    showPickupMenu = false;
  }

  function togglePickupMenu() {
    showPickupMenu = !showPickupMenu;
    showMoreMenu = false;
    showActionsMenu = false;
  }

  function closeMenus() {
    showMoreMenu = false;
    showActionsMenu = false;
    showPickupMenu = false;
  }
</script>

<!-- Dialog overlays -->
{#if showActionsMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="dialog-overlay" on:click={closeMenus}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="dialog-title">
          <i class="material-icons">auto_awesome</i>
          Actions
        </span>
        <button class="dialog-close" on:click={closeMenus}>
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="dialog-grid">
        {#each roomActions as action}
          <button
            class="popup-btn room-action-popup-btn"
            on:click={() => executeCommand(action.name)}
          >
            {action.name}
          </button>
        {/each}
      </div>
    </div>
  </div>
{/if}

{#if showMoreMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="dialog-overlay" on:click={closeMenus}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="dialog-title">
          <i class="material-icons">more_horiz</i>
          Commands
        </span>
        <button class="dialog-close" on:click={closeMenus}>
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="dialog-grid">
        {#each moreCommands as cmd}
          <button
            class="popup-btn"
            on:click={() => executeCommand(cmd.name)}
          >
            <i class="material-icons">{cmd.icon}</i>
            {cmd.label}
          </button>
        {/each}
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
          <button
            class="popup-btn pickup-popup-btn"
            on:click={() => pickupItem(item)}
          >
            {item.name}
          </button>
        {/each}
      </div>
    </div>
  </div>
{/if}

<div class="mobile-action-bar">
  <!-- Row 1: Exits + primary commands -->
  <div class="action-row">
    {#each availableCardinals as exit}
      <button
        class="btn exit-btn"
        on:click={() => executeCommand(exit.name)}
      >
        <i class="material-icons">{directionMeta[exit.name].icon}</i>
        {directionMeta[exit.name].label}
      </button>
    {/each}
    {#each allSpecialExits as exit}
      <button
        class="btn exit-btn"
        on:click={() => executeCommand(exit.name)}
      >
        {exit.name}
      </button>
    {/each}

    {#if allExits.length > 0}
      <div class="separator"></div>
    {/if}

    {#each activeCommands as cmd}
      <button
        class="btn {$store.inCombat ? 'combat-btn' : 'cmd-btn'}"
        on:click={() => executeCommand(cmd.name)}
      >
        <i class="material-icons">{cmd.icon}</i>
        {cmd.label}
      </button>
    {/each}

    {#each contextCommands as cmd}
      <button
        class="btn context-btn"
        on:click={() => executeCommand(cmd.name)}
      >
        <i class="material-icons">{cmd.icon}</i>
        {cmd.label}
      </button>
    {/each}

    {#if groundItems.length > 0}
      <button class="btn context-btn" on:click={togglePickupMenu}>
        <i class="material-icons">back_hand</i>
        Pickup
      </button>
    {/if}

    {#if roomActions.length === 1}
      <button
        class="btn room-action-btn"
        on:click={() => executeCommand(roomActions[0].name)}
      >
        {roomActions[0].name}
      </button>
    {:else if roomActions.length > 1}
      <button class="btn room-action-btn" on:click={toggleActionsMenu}>
        <i class="material-icons">auto_awesome</i>
        Actions
        <span class="action-count">{roomActions.length}</span>
      </button>
    {/if}

    <button class="btn more-btn" on:click={toggleMoreMenu}>
      <i class="material-icons">more_horiz</i>
    </button>
  </div>

  <!-- Row 2: Quick access to Character / Inventory / Equipment sheets -->
  <div class="quick-access-row">
    <button class="quick-btn" on:click={() => mobileStore.openBottomSheet('character')}>
      <i class="material-icons">person</i>
      <span>Char</span>
    </button>
    <button class="quick-btn" on:click={() => mobileStore.openBottomSheet('inventory')}>
      <i class="material-icons">inventory_2</i>
      <span>Inv</span>
    </button>
    <button class="quick-btn" on:click={() => mobileStore.openBottomSheet('equipment')}>
      <i class="material-icons">shield</i>
      <span>Equip</span>
    </button>
  </div>
</div>
