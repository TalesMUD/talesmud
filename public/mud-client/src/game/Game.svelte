<style>
  /* CSS Custom Properties for theming */
  :root {
    --terminal-bg: rgba(0, 0, 0, 0.85);
    --terminal-blur: 12px;
    --glass-border: rgba(255, 255, 255, 0.1);
    --panel-bg: rgba(0, 0, 0, 0.7);
    --accent-color: #f59e0b;
    --text-primary: #e5e7eb;
    --border-radius: 12px;
    --panel-gap: 1em;
    --transition-fast: 150ms ease;
    --transition-normal: 300ms ease;
  }

  /* Main game container */
  .gameContainer {
    display: flex;
    flex-direction: column;
    box-sizing: border-box;
    /* Top band holds the account chip so it does not cover a panel corner. */
    padding: 52px 12px 8px;
    margin: 0 auto;
    max-width: 100vw;
    height: 100dvh;
    max-height: 100dvh;
    min-height: 0;
    overflow: hidden;
    gap: 0;
  }

  .grid-container {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  /* Animation for panel appearance */
  @keyframes fadeSlideIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .gameContainer {
    animation: fadeSlideIn 0.5s ease-out;
  }

  .gameContainer.mobile {
    padding: 0;
    max-width: 100vw;
    height: 100dvh;
    max-height: 100dvh;
    overflow: auto;
  }

  .gameContainer.mobile :global(.switcher) {
    display: none;
  }

  /* Edit mode can scroll inside the shell so a new widget under the fold stays reachable. */
  .gameContainer.edit-mode {
    overflow: auto;
  }

  .gameContainer.combat-dimmed {
    filter: brightness(0.35) saturate(0.7);
    pointer-events: none;
  }

  .manual-battle-button {
    position: fixed;
    right: max(1rem, env(safe-area-inset-right));
    bottom: max(1rem, env(safe-area-inset-bottom));
    z-index: 1002;
    padding: 0.6rem 0.9rem;
    border: 1px solid #d4a44a;
    border-radius: 6px;
    background: #21180e;
    color: #f5d78c;
    font-weight: 700;
    cursor: pointer;
  }

  /* Old pink combat action-bar / hotbars must not bleed through the stage */
  .gameContainer.combat-dimmed :global(.action-bar),
  .gameContainer.combat-dimmed :global(.hotbar-widget),
  .gameContainer.combat-dimmed :global(.mobile-action-bar),
  .gameContainer.combat-dimmed :global(.actionbar),
  .gameContainer.combat-dimmed :global([data-widget-type="actionbar"]),
  .gameContainer.combat-dimmed :global([data-widget-type="hotbar"]) {
    visibility: hidden !important;
    opacity: 0 !important;
    pointer-events: none !important;
    display: none !important;
  }

  .gameContainer.combat-dimmed :global(.quest-notifications),
  .gameContainer.combat-dimmed :global(.character-switcher) {
    /* keep dimmed with parent; BattleStage is outside pointer-events:none sibling */
  }


  /* Fixed overlay for darkening/blurring the background image */
  .bg-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(10px) saturate(30%) brightness(50%);
    z-index: -1;
    pointer-events: none;
  }
</style>

<script>
  import { writable } from "svelte/store";
  import { createStore } from "./MUDXPlusStore";
  import { layoutStore } from "./layout/LayoutStore.js";
  import WidgetGrid from "./layout/WidgetGrid.svelte";
  import EditModeToolbar from "./layout/EditModeToolbar.svelte";
  import AddWidgetPanel from "./layout/AddWidgetPanel.svelte";
  import { mobileStore } from "./mobile/mobileStore.js";
  import MobileLayout from "./mobile/MobileLayout.svelte";
  import QuestNotifications from "./ui/QuestNotifications.svelte";
  import CharacterSwitcher from "./ui/CharacterSwitcher.svelte";
  import CharacterPicker from "./ui/CharacterPicker.svelte";
  import { characterPickerOpen, openCharacterPicker, closeCharacterPicker } from "./ui/characterPickerStore.js";
  import { getMyCharacters } from "../api/characters.js";
  import { isGuestSession, PICKER_SEEN_KEY, shouldAutoOpenCharacterPicker } from "../authSession.js";
  import InventoryOverlay from "./ui/InventoryOverlay.svelte";
  import BattleStage from "./ui/BattleStage.svelte";
  import MapOverviewOverlay from "./ui/MapOverviewOverlay.svelte";
  import FriendsOverlay from "./ui/FriendsOverlay.svelte";
  import PartyOverlay from "./ui/PartyOverlay.svelte";

  import { onMount, onDestroy } from "svelte";
  import { get } from "svelte/store";
  import { settingsStore } from "./SettingsStore.js";
  import { overlayStore } from "./ui/overlayStore.js";
  import { normalizeHotbarBinds, resolveHotbarActivation } from "./hudPrefs.js";
  import { hotbarSlotFromKey, isTextEntry, topOpenPanel } from "./keyboardShortcuts.js";
  import {
    accountMenuOpen,
    battleDockOpen,
    cheatSheetOpen,
    layoutDialogOpen,
    requestCloseLayoutDialog,
  } from "./uiChrome.js";
  import ShortcutSheet from "./ui/ShortcutSheet.svelte";
  import { ROOM_PLACEHOLDER } from "./portraitSrc.js";
  import { getAuth } from "../auth.js";
  import { showCharacterWizard } from "../onboarding/onboardingStore.js";
  import { createClient } from "./Client";
  import { backend, wsbackend } from "../api/base.js";
  import {
    WS_CLOSE_SESSION_REPLACED,
    beginWsConnect,
    registerWs,
    clearWs,
    closeLive,
    wsBusy,
    wsTakenOver,
    markWsTakenOver,
    liveSocket,
    wsGeneration,
  } from "./websocketGate.js";

  let client;
  let term;
  let ws;
  let renderers = [];
  let reconnectTimer;
  let reconnectPending = false;
  let reconnectAttempt = 0;
  let destroyed = false;

  // Multi-renderer dispatches output to all registered terminal widgets
  function multiRenderer(data) {
    for (const r of renderers) {
      r(data);
    }
  }

  const muxStore = createStore();
  const muxClient = writable({});

  const { isLoading, isAuthenticated, authToken } = getAuth();

  let showAddPanel = false;

  const { isMobile } = mobileStore;

  $: editMode = $layoutStore.editMode;

  // Gate on the asset id — any muxStore notify used to re-run this and
  // rewrite document.body.style (visible full-screen flicker ~every ambient tick).
  let appliedBodyBackground = null;
  $: if ($muxStore.background && $muxStore.background !== appliedBodyBackground) {
    const bgId = $muxStore.background;
    appliedBodyBackground = bgId;
    const bgUrl = backend + "/backgrounds/" + bgId + ".png";
    const placeholderUrl = ROOM_PLACEHOLDER;
    const testImg = new Image();
    testImg.onload = () => {
      if (appliedBodyBackground === bgId) {
        document.body.style.backgroundImage = "url('" + bgUrl + "')";
      }
    };
    testImg.onerror = () => {
      if (appliedBodyBackground === bgId) {
        document.body.style.backgroundImage = "url('" + placeholderUrl + "')";
      }
    };
    testImg.src = bgUrl;
  }

  $: if (client && $authToken) {
    client.setAuthToken($authToken);
  }

  // Token/auth becoming ready: connect once. Must NOT depend on `ws` —
  // assigning ws=null on close used to re-enter this and dual-open.
  $: if (client && !$isLoading && $isAuthenticated && $authToken && !destroyed && !wsTakenOver()) {
    connectWebSocket(false);
  }

  async function connectWebSocket(isReconnect = false) {
    if (!client || !$authToken || destroyed || wsTakenOver()) return;
    if (wsBusy()) return;
    if (!beginWsConnect()) return;

    clearTimeout(reconnectTimer);
    reconnectTimer = null;
    reconnectPending = false;

    muxStore.setConnectionState(
      isReconnect ? "reconnecting" : "connecting",
      isReconnect ? "Reconnecting to the game server..." : "Connecting to the game server...",
      reconnectAttempt
    );

    let nextWs;
    try {
      const ticketRes = await fetch(`${backend}/ws-ticket`, {
        method: "POST",
        headers: { Authorization: `Bearer ${$authToken}` },
      });
      if (!ticketRes.ok) {
        throw new Error("ws-ticket " + ticketRes.status);
      }
      const ticketBody = await ticketRes.json();
      if (!ticketBody || !ticketBody.ticket) {
        throw new Error("ws-ticket empty");
      }
      if (destroyed || wsTakenOver()) {
        clearWs(null);
        return;
      }
      nextWs = new WebSocket(wsbackend + "?ticket=" + encodeURIComponent(ticketBody.ticket));
    } catch (e) {
      clearWs(null);
      console.info("[ws] construct failed", e);
      scheduleReconnect();
      return;
    }
    registerWs(nextWs);
    ws = nextWs;
    const gen = wsGeneration();
    client.setWSClient(nextWs);
    console.info("[ws] construct", { gen, isReconnect });

    nextWs.addEventListener("open", () => {
      if (destroyed || liveSocket() !== nextWs) return;
      reconnectAttempt = 0;
      muxStore.setConnectionState("connected", "Connected");
      console.info("[ws] open", { gen });
    });

    nextWs.addEventListener("close", (ev) => {
      const code = ev && typeof ev.code === "number" ? ev.code : 0;
      const reason = (ev && ev.reason) || "";
      console.info("[ws] close", { gen, code, reason, wasClean: !!(ev && ev.wasClean) });
      clearWs(nextWs);
      if (ws === nextWs) ws = null;
      if (destroyed) return;

      if (code === WS_CLOSE_SESSION_REPLACED) {
        markWsTakenOver();
        reconnectPending = false;
        muxStore.setConnectionState(
          "disconnected",
          "Session taken over by another connection. Refresh to reclaim."
        );
        return;
      }
      scheduleReconnect();
    });

    nextWs.addEventListener("error", () => {
      if (destroyed || liveSocket() !== nextWs) return;
      console.info("[ws] error", { gen, readyState: nextWs.readyState });
    });
  }

  function scheduleReconnect() {
    if (destroyed || wsTakenOver() || !$isAuthenticated || !$authToken) {
      reconnectPending = false;
      if (!wsTakenOver()) muxStore.setConnectionState("disconnected", "Disconnected");
      return;
    }
    if (wsBusy() || reconnectPending) return;
    reconnectAttempt += 1;
    const delay = Math.min(1000 * reconnectAttempt, 5000);
    muxStore.setConnectionState(
      "reconnecting",
      `Connection lost. Reconnecting in ${Math.ceil(delay / 1000)}s...`,
      reconnectAttempt
    );
    clearTimeout(reconnectTimer);
    reconnectPending = true;
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;
      connectWebSocket(true);
    }, delay);
  }

  const characterCreator = () => {
    console.log("CREATE CHARACTER");
    showCharacterWizard.set(true);
  };

  function handleTerminalReady(terminal, termRenderer) {
    // Track xterm instance if provided (classic Terminal passes it, Terminal X passes null)
    if (terminal) {
      term = terminal;
    }

    // Register this terminal's renderer for multi-output
    renderers.push(termRenderer);

    // Create client once using the multi-renderer so all terminals receive output
    if (!client) {
      client = createClient(
        multiRenderer,
        characterCreator,
        muxStore
      );
      muxClient.set(client);
    }

    // Unauthenticated users are handled by the onboarding flow in App.svelte
  }

  function handleTerminalInput(input) {
    if (client) {
      client.onInput(input);
    }
  }

  function sendMessage(msg) {
    if (client) {
      client.sendMessage(msg);
    }
  }

  function shortcutFlags() {
    const play = get(muxStore);
    const layout = get(layoutStore);
    return {
      cheatSheet: get(cheatSheetOpen),
      layoutDialog: get(layoutDialogOpen),
      characterPicker: get(characterPickerOpen),
      settings: !!get(settingsStore).modalOpen,
      addWidget: showAddPanel,
      map: !!play.mapOverviewOpen,
      friends: !!play.friendsOverlayOpen,
      party: !!play.partyOverlayOpen,
      inventory: !!play.inventoryOverlayOpen,
      battleOutcome: play.combatPhase === "ending",
      battleDock: get(battleDockOpen),
      accountMenu: get(accountMenuOpen),
      widgetFocus: !!layout.focusId,
      editMode: !!layout.editMode,
    };
  }

  function closeTopPanel(id) {
    if (id === "cheatSheet") cheatSheetOpen.set(false);
    else if (id === "layoutDialog") requestCloseLayoutDialog();
    else if (id === "characterPicker") closeCharacterPicker();
    else if (id === "settings") settingsStore.closeModal();
    else if (id === "addWidget") showAddPanel = false;
    else if (id === "map") muxStore.closeMapOverview();
    else if (id === "friends") muxStore.closeFriendsOverlay();
    else if (id === "party") muxStore.closePartyOverlay();
    else if (id === "inventory") muxStore.closeInventoryOverlay();
    else if (id === "battleOutcome") muxStore.dismissCombat();
    else if (id === "battleDock") battleDockOpen.set(false);
    else if (id === "accountMenu") accountMenuOpen.set(false);
    else if (id === "widgetFocus") layoutStore.toggleFocus(get(layoutStore).focusId);
    else if (id === "editMode") layoutStore.exitEditMode(false);
  }

  function fireHotbarSlot(index) {
    const binds = normalizeHotbarBinds(get(settingsStore).interface?.hotbarBinds);
    const bind = binds[index];
    if (!bind) return;
    const play = get(muxStore);
    const inCombat = play.combatPhase === "active" || !!play.inCombat;
    const result = resolveHotbarActivation(bind, {
      inCombat,
      inventory: play.inventory || [],
    });
    if (!result.ok) {
      if (result.reason && result.reason !== "empty" && overlayStore?.pushMessage) {
        overlayStore.pushMessage(result.reason);
      }
      return;
    }
    if (result.command) sendMessage(result.command);
  }

  function onShortcutKey(event) {
    if (event.metaKey || event.ctrlKey || event.altKey) return;
    if (isTextEntry(event.target)) return;
    if (event.key === "?" || (event.key === "/" && event.shiftKey)) {
      event.preventDefault();
      cheatSheetOpen.update((open) => !open);
      return;
    }
    if (event.key === "Escape") {
      const top = topOpenPanel(shortcutFlags());
      if (!top) return;
      event.preventDefault();
      closeTopPanel(top);
      return;
    }
    if (event.key === "Tab") {
      const play = get(muxStore);
      if (play.combatPhase !== "active") return;
      event.preventDefault();
      if (muxStore.cycleCombatTarget) muxStore.cycleCombatTarget(event.shiftKey ? -1 : 1);
      return;
    }
    const slot = hotbarSlotFromKey(event.key);
    if (slot >= 0) {
      event.preventDefault();
      fireHotbarSlot(slot);
    }
  }

  // The preference also applies mid-fight; an explicit manual cover uses the same restore contract.
  let manualBattleOpen = false;
  $: if ($muxStore.combatPhase === 'idle') manualBattleOpen = false;
  $: showBattleStage = $settingsStore.interface?.combatAutoFocus !== false || manualBattleOpen;
  $: layoutStore.syncCombatFocus($muxStore.combatPhase, showBattleStage);

  let pickerChecked = false;
  $: if ($authToken && $muxStore.connectionStatus === "connected" && !pickerChecked) {
    pickerChecked = true;
    if (!isGuestSession($authToken)) {
      let seen = false;
      try { seen = sessionStorage.getItem(PICKER_SEEN_KEY) === "1"; } catch (err) { seen = false; }
      if (!seen) {
        getMyCharacters($authToken, (chars) => {
          const list = Array.isArray(chars) ? chars : [];
          if (shouldAutoOpenCharacterPicker({ guest: false, seen: false, characterCount: list.length })) {
            try { sessionStorage.setItem(PICKER_SEEN_KEY, "1"); } catch (err) { /* ignore */ }
            openCharacterPicker();
          }
        }, () => {});
      }
    }
  }

  onMount(async () => {
    document.body.style.backgroundImage = "url('" + backend + "/backgrounds/oldtown-griphon.png')";
    document.body.style.backgroundAttachment = "fixed";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000000";
    }

    // Initialize layout from storage
    layoutStore.loadFromStorage();
    window.addEventListener("keydown", onShortcutKey);
  });

  onDestroy(async () => {
    destroyed = true;
    layoutStore.syncCombatFocus("idle");
    window.removeEventListener("keydown", onShortcutKey);
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
    reconnectPending = false;
    closeLive();
    ws = null;

    document.body.style.backgroundImage = "";
    document.body.style.backgroundAttachment = "";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000055";
    }
  });
</script>

<div class="bg-overlay"></div>

<div class="gameContainer" class:mobile={$isMobile} class:edit-mode={editMode} class:combat-dimmed={showBattleStage && ($muxStore.combatPhase === "active" || ($muxStore.inCombat && $muxStore.combatPhase !== "ending"))}>
  <CharacterSwitcher
    store={muxStore}
    authToken={$authToken}
  />

  {#if $isMobile}
    <MobileLayout
      store={muxStore}
      authToken={$authToken}
      {sendMessage}
      onTerminalReady={handleTerminalReady}
      onTerminalInput={handleTerminalInput}
    />
  {:else}
    <div class="grid-container">
      <WidgetGrid
        store={muxStore}
        {sendMessage}
        onTerminalReady={handleTerminalReady}
        onTerminalInput={handleTerminalInput}
      />
    </div>

    {#if editMode}
      <EditModeToolbar on:openAddPanel={() => showAddPanel = true} />
    {/if}

    {#if showAddPanel}
      <AddWidgetPanel on:close={() => showAddPanel = false} />
    {/if}
  {/if}

  <!-- Quest notifications - shown on all layouts -->
  <QuestNotifications store={muxStore} />

  <!-- Inventory popup overlay (default inv open mode) -->
  <InventoryOverlay store={muxStore} {sendMessage} />
  <FriendsOverlay store={muxStore} {sendMessage} />
  <PartyOverlay store={muxStore} {sendMessage} />

</div>

<!-- Map overview: sibling of BattleStage, outside .gameContainer (body portal) -->
<MapOverviewOverlay store={muxStore} {sendMessage} />

<!-- C2: full-screen battle stage over dimmed room chrome -->
<BattleStage store={muxStore} {sendMessage} shown={showBattleStage} />
{#if $muxStore.combatPhase === 'active' && $settingsStore.interface?.combatAutoFocus === false}
  <button class="manual-battle-button" on:click={() => manualBattleOpen = !manualBattleOpen}>
    {manualBattleOpen ? 'Return to layout' : 'Open BattleStage'}
  </button>
{/if}
<ShortcutSheet />

{#if $characterPickerOpen}
  <CharacterPicker
    authToken={$authToken}
    activeCharacter={$muxStore.character}
    canSwitch={$muxStore.connectionStatus === "connected"}
    {sendMessage}
    onClose={closeCharacterPicker}
  />
{/if}
