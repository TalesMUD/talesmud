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
    padding: 1em;
    margin: 0 auto;
    max-width: min(95vw, 2400px);
    height: calc(100vh - 2em);
    gap: var(--panel-gap);
  }

  .grid-container {
    flex: 1;
    min-height: 0;
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
    height: 100vh;
    height: 100dvh;
  }

  .gameContainer.combat-dimmed {
    filter: brightness(0.35) saturate(0.7);
    pointer-events: none;
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
  import InventoryOverlay from "./ui/InventoryOverlay.svelte";
  import BattleStage from "./ui/BattleStage.svelte";
  import MapOverviewOverlay from "./ui/MapOverviewOverlay.svelte";
  import FriendsOverlay from "./ui/FriendsOverlay.svelte";

  import { onMount, onDestroy } from "svelte";
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
    const placeholderUrl = "img/placeholder.png";
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

  function connectWebSocket(isReconnect = false) {
    if (!client || !$authToken || destroyed || wsTakenOver()) return;
    if (wsBusy()) return;
    if (!beginWsConnect()) return;

    clearTimeout(reconnectTimer);
    reconnectTimer = null;
    reconnectPending = false;

    const url = wsbackend + "?access_token=";
    let nextWs;
    try {
      nextWs = new WebSocket(url + $authToken);
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

    muxStore.setConnectionState(
      isReconnect ? "reconnecting" : "connecting",
      isReconnect ? "Reconnecting to the game server..." : "Connecting to the game server...",
      reconnectAttempt
    );

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

  onMount(async () => {
    document.body.style.backgroundImage = "url('" + backend + "/backgrounds/oldtown-griphon.png')";
    document.body.style.backgroundAttachment = "fixed";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000000";
    }

    // Initialize layout from storage
    layoutStore.loadFromStorage();
  });

  onDestroy(async () => {
    destroyed = true;
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

<div class="gameContainer" class:mobile={$isMobile} class:combat-dimmed={($muxStore.combatPhase === "active" || $muxStore.combatPhase === "ending" || $muxStore.inCombat)}>
  <CharacterSwitcher
    store={muxStore}
    authToken={$authToken}
    {sendMessage}
  />

  {#if $isMobile}
    <MobileLayout
      store={muxStore}
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

</div>

<!-- Map overview: sibling of BattleStage, outside .gameContainer (body portal) -->
<MapOverviewOverlay store={muxStore} {sendMessage} />

<!-- C2: full-screen battle stage over dimmed room chrome -->
<BattleStage store={muxStore} {sendMessage} />
