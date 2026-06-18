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

  import { onMount, onDestroy } from "svelte";
  import { getAuth } from "../auth.js";
  import { showCharacterWizard } from "../onboarding/onboardingStore.js";
  import { createClient } from "./Client";
  import { backend, wsbackend } from "../api/base.js";

  let client;
  let term;
  let ws;
  let renderers = [];
  let reconnectTimer;
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

  $: {
    // set document background (blurred)
    if ($muxStore.background) {
      const bgUrl = backend + "/backgrounds/" + $muxStore.background + ".png";
      const placeholderUrl = "img/placeholder.png";
      const testImg = new Image();
      testImg.onload = () => {
        document.body.style.backgroundImage = "url('" + bgUrl + "')";
      };
      testImg.onerror = () => {
        document.body.style.backgroundImage = "url('" + placeholderUrl + "')";
      };
      testImg.src = bgUrl;
    }
  }

  $: if (client && !ws && !$isLoading && $isAuthenticated && $authToken && !destroyed) {
    connectWebSocket(false);
  }

  function connectWebSocket(isReconnect = false) {
    if (!client || ws || !$authToken) return;

    muxStore.setConnectionState(
      isReconnect ? "reconnecting" : "connecting",
      isReconnect ? "Reconnecting to the game server..." : "Connecting to the game server...",
      reconnectAttempt
    );

    const url = wsbackend + "?access_token=";
    const nextWs = new WebSocket(url + $authToken);
    ws = nextWs;
    client.setWSClient(nextWs);

    nextWs.addEventListener("open", () => {
      reconnectAttempt = 0;
      muxStore.setConnectionState("connected", "Connected");
    });

    nextWs.addEventListener("close", () => {
      if (destroyed || ws !== nextWs) return;
      ws = null;
      scheduleReconnect();
    });

    nextWs.addEventListener("error", () => {
      muxStore.setConnectionState("reconnecting", "Connection interrupted. Reconnecting...", reconnectAttempt);
    });
  }

  function scheduleReconnect() {
    if (destroyed || !$isAuthenticated || !$authToken) {
      muxStore.setConnectionState("disconnected", "Disconnected");
      return;
    }
    reconnectAttempt += 1;
    const delay = Math.min(1000 * reconnectAttempt, 5000);
    muxStore.setConnectionState(
      "reconnecting",
      `Connection lost. Reconnecting in ${Math.ceil(delay / 1000)}s...`,
      reconnectAttempt
    );
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(() => connectWebSocket(true), delay);
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
    if (ws) {
      ws.close();
      ws = null;
    }

    document.body.style.backgroundImage = "";
    document.body.style.backgroundAttachment = "";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000055";
    }
  });
</script>

<div class="bg-overlay"></div>

<div class="gameContainer" class:mobile={$isMobile}>
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
</div>
