<style>
  .mudx {
    margin-top: 1em;
    margin-left: auto;
    margin-right: auto;
    margin-bottom: 1em;
    max-width: 100%;
    padding: 0 1em;
    position: relative;
    z-index: 100;
  }

  /* Main row container - D-pad + action bar side by side */
  .action-row {
    display: flex;
    align-items: stretch;
    gap: 0.5em;
  }

  /* Compass container - above action bar, centered */
  .compass-container {
    display: flex;
    justify-content: center;
    margin-bottom: 0.5em;
  }

  .special-exits {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
    align-items: center;
  }

  .action-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5em;
    align-items: center;
    padding: 0.4em 0.75em;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    flex: 1;
  }

  .commands-section {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
    align-items: center;
  }

  .room-actions-section {
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
    color: #e5e7eb;
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

  .context-btn:hover {
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

  .separator {
    width: 1px;
    height: 24px;
    background: rgba(255, 255, 255, 0.15);
    margin: 0 0.3em;
    flex-shrink: 0;
  }

  .spacer {
    flex: 1;
  }

  /* More menu container */
  .more-menu-container {
    position: static;
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

  /* Popup menu - centered above action row */
  .more-popup {
    position: absolute;
    bottom: calc(100% + 8px);
    left: 50%;
    transform: translateX(-50%);
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 10px;
    padding: 0.6em;
    z-index: 1000;
    animation: popupSlideIn 0.2s ease-out;
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.4em;
    min-width: 380px;
  }

  @keyframes popupSlideIn {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* Popup buttons - same style as action bar buttons */
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

  /* Backdrop to close menu */
  .menu-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 999;
  }

  /* Responsive adjustments */
  @media screen and (max-width: 600px) {
    .action-row {
      flex-direction: column;
      gap: 0.4em;
    }

    .compass-container {
      justify-content: center;
    }

    .special-exits {
      flex-direction: row;
    }

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

    .more-popup {
      grid-template-columns: repeat(2, 1fr);
      min-width: 200px;
      left: 50%;
      transform: translateX(-50%);
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
  import CompassExits from "./ui/CompassExits.svelte";
  import { getCardinalExits, getSpecialExits, getVerticalExits } from "./MUDXPlusStore";

  export let store;
  // svelte-ignore unused-export-let
  export let term;
  export let sendMessage;

  // More menu state
  let showMoreMenu = false;

  // Derive exits from store
  $: cardinalExits = getCardinalExits($store.exits);
  $: specialExits = getSpecialExits($store.exits);
  $: verticalExits = getVerticalExits($store.exits);

  // Debug logging for exits
  $: if ($store.exits) {
    console.log("Raw exits from store:", JSON.stringify($store.exits, null, 2));
    console.log("Cardinal exits derived:", JSON.stringify(cardinalExits, null, 2));
    console.log("Special exits:", JSON.stringify(specialExits, null, 2));
  }

  // Combined special + vertical exits for display
  $: allSpecialExits = [...verticalExits, ...specialExits];

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

  // Additional commands for "more" menu (flat list for button grid)
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
    ...($store.hasItems ? [{ name: "pickup", icon: "back_hand", label: "Pickup" }] : []),
  ];

  // Room actions
  $: roomActions = $store.actions || [];

  function executeCommand(cmd) {
    sendMessage(cmd);
    showMoreMenu = false;
  }

  function toggleMoreMenu() {
    showMoreMenu = !showMoreMenu;
  }

  function closeMoreMenu() {
    showMoreMenu = false;
  }
</script>

<div class="mudx">
  <!-- Compass above action bar -->
  <div class="compass-container">
    <CompassExits
      cardinalExits={cardinalExits}
      sendMessage={sendMessage}
    />
  </div>

  <!-- Action bar below compass -->
  <div class="action-row">
    <!-- Action bar with commands -->
    <div class="action-bar">
      <!-- Special and Vertical exits (moved here) -->
      {#if allSpecialExits.length > 0}
        <div class="special-exits">
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

    <!-- Room-specific actions (if any) -->
    {#if roomActions.length > 0}
      <div class="separator"></div>
      <div class="room-actions-section">
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
    {/if}

    <div class="spacer"></div>

    <!-- More menu button -->
    <div class="more-menu-container">
      <button
        class="btn more-btn"
        class:active={showMoreMenu}
        on:click={toggleMoreMenu}
        title="More commands"
      >
        <i class="material-icons">more_horiz</i>
      </button>

      {#if showMoreMenu}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div class="menu-backdrop" on:click={closeMoreMenu}></div>
        <div class="more-popup">
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
      {/if}
    </div>
    </div>
  </div>
</div>
