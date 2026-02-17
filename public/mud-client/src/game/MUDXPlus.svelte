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

  export let store;
  // svelte-ignore unused-export-let
  export let term;
  export let sendMessage;

  // Menu state
  let showMoreMenu = false;
  let showActionsMenu = false;
  let showPickupMenu = false;

  // Derive exits from store
  $: cardinalExits = getCardinalExits($store.exits);
  $: specialExits = getSpecialExits($store.exits);
  $: verticalExits = getVerticalExits($store.exits);

  // Only show available cardinal directions
  $: availableCardinals = cardinalExits.filter(e => e.available);

  // Direction display config
  const directionMeta = {
    north: { label: "N", icon: "north" },
    south: { label: "S", icon: "south" },
    east:  { label: "E", icon: "east" },
    west:  { label: "W", icon: "west" },
  };

  // Combined special + vertical exits for display
  $: allSpecialExits = [...verticalExits, ...specialExits];

  // All directional exits in one section (cardinal + vertical + special)
  $: allExits = [...availableCardinals, ...allSpecialExits];

  // Standard commands (shown when not in combat)
  const standardCommands = [
    { name: "look", icon: "visibility", label: "Look" },
    { name: "i", icon: "inventory_2", label: "Inv" },
    { name: "equipment", icon: "shield", label: "Equip" },
    { name: "character", icon: "person", label: "Stats" },
  ];

  // Combat commands (replace standard when in combat)
  const combatCommands = [
    { name: "attack", icon: "swords", label: "Attack" },
    { name: "defend", icon: "security", label: "Defend" },
    { name: "flee", icon: "directions_run", label: "Flee" },
    { name: "status", icon: "monitor_heart", label: "Status" },
  ];

  // Additional commands for "more" menu
  const moreCommands = [
    { name: "who", icon: "people", label: "Who" },
    { name: "bind", icon: "location_on", label: "Bind" },
    { name: "drop", icon: "delete", label: "Drop" },
    { name: "use", icon: "touch_app", label: "Use" },
    { name: "examine", icon: "search", label: "Examine" },
    { name: "shrug", icon: "sentiment_neutral", label: "Shrug" },
    { name: "scream", icon: "campaign", label: "Scream" },
    { name: "help", icon: "help", label: "Help" },
  ];

  // Select commands based on combat state
  $: activeCommands = $store.inCombat ? combatCommands : standardCommands;

  // Context-specific commands (only in normal mode)
  $: contextCommands = $store.inCombat ? [] : [
    ...($store.hasMerchant ? [{ name: "list", icon: "store", label: "Shop" }] : []),
  ];

  // Room actions
  $: roomActions = $store.actions || [];

  // Ground items (pickable)
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

<div class="mudx">
  <!-- Popup area: fills space above action bar, popup aligns to bottom -->
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
              title={action.description || action.name}
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
      <div class="dialog dialog-more" on:click|stopPropagation>
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
              title={cmd.label}
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

  <!-- Action bar pinned to bottom -->
  <div class="action-bar">
    <!-- Direction exits: cardinal (N/E/S/W) + vertical + special -->
    {#if allExits.length > 0}
      <div class="exits-section">
        {#each availableCardinals as exit}
          <button
            class="btn exit-btn"
            on:click={() => executeCommand(exit.name)}
            title={exit.name}
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
      </div>
      <div class="separator"></div>
    {/if}

    <!-- Main Commands (standard or combat) -->
    <div class="commands-section">
      {#each activeCommands as cmd}
        <button
          class="btn {$store.inCombat ? 'combat-btn' : 'cmd-btn'}"
          on:click={() => executeCommand(cmd.name)}
          title={cmd.label}
        >
          <i class="material-icons">{cmd.icon}</i>
          {cmd.label}
        </button>
      {/each}

      <!-- Context commands (merchant, items) -->
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

    <!-- Pickup button (only when items on ground) -->
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

    <!-- Room-specific actions -->
    {#if roomActions.length === 1}
      <!-- Single action: show inline -->
      <div class="separator"></div>
      <button
        class="btn room-action-btn"
        on:click={() => executeCommand(roomActions[0].name)}
        title={roomActions[0].description || roomActions[0].name}
      >
        {roomActions[0].name}
      </button>
    {:else if roomActions.length > 1}
      <!-- Multiple actions: overflow menu trigger -->
      <div class="separator"></div>
      <button
        class="btn room-actions-btn"
        class:active={showActionsMenu}
        on:click={toggleActionsMenu}
        title="Room actions"
      >
        <i class="material-icons">auto_awesome</i>
        Actions
        <span class="action-count">{roomActions.length}</span>
      </button>
    {/if}

    <!-- More menu button -->
    <button
      class="btn more-btn"
      class:active={showMoreMenu}
      on:click={toggleMoreMenu}
      title="More commands"
    >
      <i class="material-icons">more_horiz</i>
    </button>
  </div>
</div>
