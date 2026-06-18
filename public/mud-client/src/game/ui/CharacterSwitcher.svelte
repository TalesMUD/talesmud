<script>
  import { onMount } from "svelte";
  import { getMyCharacters } from "../../api/characters.js";

  export let authToken = "";
  export let store;
  export let sendMessage;

  let characters = [];
  let selectedCharacterId = "";
  let loading = false;
  let error = "";
  let menuOpen = false;

  $: currentCharacter = $store.character;
  $: connectionStatus = $store.connectionStatus || "idle";
  $: connectionMessage = $store.connectionMessage || "";
  $: reconnectAttempt = $store.reconnectAttempt || 0;
  $: currentCharacterName = currentCharacter?.name || "No character";
  $: if (currentCharacter?.id && selectedCharacterId !== currentCharacter.id) {
    selectedCharacterId = currentCharacter.id;
  }

  onMount(() => {
    loadCharacters();
  });

  $: if (authToken && characters.length === 0 && !loading && !error) {
    loadCharacters();
  }

  function loadCharacters() {
    if (!authToken || loading) return;
    loading = true;
    error = "";
    getMyCharacters(
      authToken,
      (data) => {
        characters = data || [];
        loading = false;
      },
      () => {
        error = "Could not load characters";
        loading = false;
      }
    );
  }

  function selectCharacter(character) {
    if (!character || !sendMessage || character.id === currentCharacter?.id) {
      menuOpen = false;
      return;
    }
    selectedCharacterId = character.id;
    sendMessage(`sc ${character.name}`);
    menuOpen = false;
  }

  function statusLabel(status) {
    if (status === "connected") return "Online";
    if (status === "connecting") return "Connecting";
    if (status === "reconnecting") return reconnectAttempt > 0 ? `Reconnecting ${reconnectAttempt}` : "Reconnecting";
    if (status === "disconnected") return "Offline";
    return "Waiting";
  }
</script>

<style>
  .switcher {
    position: fixed;
    top: 14px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 1002;
    display: flex;
    align-items: stretch;
    min-width: min(420px, calc(100vw - 32px));
    max-width: calc(100vw - 32px);
    border: 1px solid rgba(245, 158, 11, 0.28);
    border-radius: 8px;
    background: rgba(5, 8, 13, 0.88);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    color: #e5e7eb;
    font-size: 12px;
  }

  .status {
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 8px 10px;
    border-right: 1px solid rgba(255, 255, 255, 0.08);
    white-space: nowrap;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #6b7280;
    box-shadow: 0 0 0 3px rgba(107, 114, 128, 0.12);
  }

  .dot.connected {
    background: #22c55e;
    box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.18);
  }

  .dot.connecting,
  .dot.reconnecting {
    background: #f59e0b;
    box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.18);
  }

  .dot.disconnected {
    background: #ef4444;
    box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.18);
  }

  .character-area {
    position: relative;
    flex: 1;
    min-width: 0;
  }

  .trigger {
    width: 100%;
    height: 100%;
    min-height: 36px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 7px 10px;
    border: 0;
    background: transparent;
    color: inherit;
    cursor: pointer;
    text-align: left;
  }

  .identity {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .label {
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 9px;
  }

  .name {
    color: #f0e6d3;
    font-family: "Cinzel", serif;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .chevron {
    color: #f59e0b;
    font-size: 18px;
    line-height: 1;
  }

  .menu {
    position: absolute;
    top: calc(100% + 6px);
    left: 0;
    right: 0;
    border: 1px solid rgba(245, 158, 11, 0.25);
    border-radius: 8px;
    background: rgba(6, 8, 12, 0.96);
    box-shadow: 0 18px 40px rgba(0, 0, 0, 0.45);
    overflow: hidden;
  }

  .menu-item {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 9px 11px;
    border: 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    background: transparent;
    color: #e5e7eb;
    cursor: pointer;
    text-align: left;
  }

  .menu-item:hover,
  .menu-item.active {
    background: rgba(245, 158, 11, 0.12);
  }

  .meta {
    color: #9ca3af;
    font-size: 11px;
    white-space: nowrap;
  }

  .message {
    padding: 10px 11px;
    color: #9ca3af;
  }

  @media screen and (max-width: 768px) {
    .switcher {
      top: calc(54px + env(safe-area-inset-top, 0px));
      min-width: calc(100vw - 24px);
    }
  }
</style>

<div class="switcher" title={connectionMessage}>
  <div class="status">
    <span class="dot {connectionStatus}"></span>
    <span>{statusLabel(connectionStatus)}</span>
  </div>

  <div class="character-area">
    <button class="trigger" type="button" on:click={() => menuOpen = !menuOpen}>
      <span class="identity">
        <span class="label">Playing as</span>
        <span class="name">{currentCharacterName}</span>
      </span>
      <span class="chevron">v</span>
    </button>

    {#if menuOpen}
      <div class="menu">
        {#if loading}
          <div class="message">Loading characters...</div>
        {:else if error}
          <button class="menu-item" type="button" on:click={loadCharacters}>
            <span>{error}</span>
            <span class="meta">Retry</span>
          </button>
        {:else if characters.length === 0}
          <div class="message">No characters available</div>
        {:else}
          {#each characters as character (character.id)}
            <button
              class="menu-item"
              class:active={character.id === selectedCharacterId}
              type="button"
              on:click={() => selectCharacter(character)}
            >
              <span>{character.name}</span>
              <span class="meta">Lv {character.level || 1}</span>
            </button>
          {/each}
        {/if}
      </div>
    {/if}
  </div>
</div>
