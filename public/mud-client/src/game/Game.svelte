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
</style>

<script>
  import { writable } from "svelte/store";
  import { createStore } from "./MUDXPlusStore";
  import { layoutStore } from "./layout/LayoutStore.js";
  import WidgetGrid from "./layout/WidgetGrid.svelte";
  import EditModeToolbar from "./layout/EditModeToolbar.svelte";
  import AddWidgetPanel from "./layout/AddWidgetPanel.svelte";

  import CharacterCreator from "../characters/CharacterCreator.svelte";
  import { onMount, onDestroy } from "svelte";
  import { getAuth } from "../auth.js";
  import { createClient } from "./Client";
  import { wsbackend } from "../api/base.js";

  let client;
  let term;
  let renderer;
  let ws;

  const muxStore = createStore();
  const muxClient = writable({});

  const { isLoading, isAuthenticated, authToken } = getAuth();

  let showAddPanel = false;

  $: editMode = $layoutStore.editMode;

  $: {
    // Only connect when: client exists, no existing ws, auth is loaded, user is authenticated, and token exists
    if (client && !ws && !$isLoading && $isAuthenticated && $authToken) {
      console.log("Connecting to websocket with token:", $authToken.slice(0, 20) + "...");
      const url = wsbackend + "?access_token=";
      ws = new WebSocket(url + $authToken);
      client.setWSClient(ws);
    }

    // set document background (blurred)
    if ($muxStore.background) {
      const bgUrl = "/api/backgrounds/" + $muxStore.background + ".png";
      const placeholderUrl = "/play/img/placeholder.png";
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

  const characterCreator = () => {
    console.log("CREATE CHARACTER");
    var Modalelem = document.querySelector(".modal");
    var instance = M.Modal.init(Modalelem);
    instance.open();
  };

  function handleTerminalReady(terminal, termRenderer) {
    term = terminal;
    renderer = termRenderer;

    // Now create the client with the renderer
    client = createClient(
      renderer,
      characterCreator,
      muxStore
    );

    muxClient.set(client);

    // Show message if not authenticated
    if (!$isLoading && !$isAuthenticated) {
      term.writeln("Please log in to connect to the game.");
    }
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
    document.body.style.backgroundImage = "url('/api/backgrounds/oldtown-griphon.png')";
    document.body.style.backdropFilter =
      "blur(10px) saturate(30%) brightness(50%)";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000000";
    }

    // Initialize layout from storage
    layoutStore.loadFromStorage();
  });

  onDestroy(async () => {
    document.body.style.backgroundImage = "";
    document.body.style.backdropFilter = "";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000055";
    }
  });
</script>

<CharacterCreator />

<div class="gameContainer">
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
</div>
