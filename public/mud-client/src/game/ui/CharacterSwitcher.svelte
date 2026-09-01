<script>
  import { onMount } from "svelte";
  import { getMyCharacters } from "../../api/characters.js";

  export let store;
  export let authToken;
  export let sendMessage;

  let characters = [];
  let loading = false;
  let open = false;
  let error = "";

  $: activeCharacter = $store.character;
  $: connectionStatus = $store.connectionStatus;
  $: canSwitch = connectionStatus === "connected";

  onMount(() => {
    loadCharacters();
  });

  $: if (authToken && characters.length === 0 && !loading) {
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

  function toggleOpen() {
    open = !open;
    if (open && characters.length === 0) {
      loadCharacters();
    }
  }

  function selectCharacter(character) {
    if (!character || !canSwitch || character.id === activeCharacter?.id) return;
    sendMessage(`sc ${character.name}`);
    open = false;
  }

  function className(character) {
    return character?.class?.name || character?.class?.Name || "Adventurer";
  }
</script>

<style>
  .switcher {
    position: fixed;
    top: 15px;
    right: 60px;
    z-index: 900;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    font-family: 'Cinzel', serif;
    color: #f0e6d3;
    pointer-events: none;
  }

  .switcher-button {
    pointer-events: auto;
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    min-height: 38px;
    padding: 0.35rem 0.85rem;
    border: 1px solid rgba(194, 162, 99, 0.38);
    border-radius: 8px;
    background:
      linear-gradient(180deg, rgba(28, 23, 18, 0.92), rgba(9, 10, 12, 0.86)),
      radial-gradient(circle at 20% 0%, rgba(194, 162, 99, 0.18), transparent 40%);
    color: #f0e6d3;
    box-shadow: 0 10px 26px rgba(0, 0, 0, 0.35);
    cursor: pointer;
    max-width: min(78vw, 380px);
  }

  .switcher-button:disabled {
    cursor: wait;
    opacity: 0.72;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: #9ca3af;
    box-shadow: 0 0 10px rgba(156, 163, 175, 0.5);
    flex: 0 0 auto;
  }

  .status-dot.connected {
    background: #4ade80;
    box-shadow: 0 0 12px rgba(74, 222, 128, 0.75);
  }

  .status-dot.connecting,
  .status-dot.reconnecting {
    background: #fbbf24;
    box-shadow: 0 0 12px rgba(251, 191, 36, 0.75);
  }

  .identity {
    display: flex;
    flex-direction: column;
    min-width: 0;
    text-align: left;
  }

  .name {
    font-size: 0.92rem;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    letter-spacing: 0;
  }

  .meta {
    font-size: 0.66rem;
    color: rgba(240, 230, 211, 0.68);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    letter-spacing: 0;
  }

  .chevron {
    font-size: 1.2rem;
    color: rgba(240, 230, 211, 0.72);
    line-height: 1;
  }

  .menu {
    pointer-events: auto;
    margin-top: 0.45rem;
    width: min(86vw, 420px);
    max-height: min(62vh, 440px);
    overflow: auto;
    border: 1px solid rgba(194, 162, 99, 0.28);
    border-radius: 8px;
    background: rgba(7, 9, 12, 0.94);
    box-shadow: 0 18px 46px rgba(0, 0, 0, 0.52);
    backdrop-filter: blur(14px);
    align-self: flex-end;
  }

  .menu-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.65rem 0.75rem;
    color: rgba(240, 230, 211, 0.72);
    font-size: 0.68rem;
    text-transform: uppercase;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    letter-spacing: 0;
  }

  .refresh {
    border: 0;
    background: transparent;
    color: #c2a263;
    cursor: pointer;
    padding: 0.1rem 0.2rem;
  }

  .character-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.75rem;
    width: 100%;
    padding: 0.7rem 0.8rem;
    border: 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    background: transparent;
    color: #f0e6d3;
    cursor: pointer;
    text-align: left;
  }

  .character-row:hover,
  .character-row.active {
    background: rgba(194, 162, 99, 0.12);
  }

  .character-row:disabled {
    cursor: default;
    opacity: 0.58;
  }

  .level {
    color: #fbbf24;
    font-size: 0.74rem;
    align-self: center;
  }

  .notice {
    padding: 0.8rem;
    color: rgba(240, 230, 211, 0.68);
    font-size: 0.78rem;
  }

  @media screen and (max-width: 768px) {
    .switcher {
      top: 10px;
      right: 10px;
      left: auto;
      transform: none;
      align-items: flex-end;
    }
  }
</style>

<div class="switcher">
  <button class="switcher-button" on:click={toggleOpen} disabled={loading && characters.length === 0}>
    <span class="status-dot" class:connected={connectionStatus === 'connected'} class:connecting={connectionStatus === 'connecting'} class:reconnecting={connectionStatus === 'reconnecting'}></span>
    <span class="identity">
      <span class="name">{activeCharacter?.name || "Selecting character"}</span>
      <span class="meta">
        {#if connectionStatus !== "connected"}
          {$store.connectionMessage || "Connecting"}
        {:else}
          {className(activeCharacter)}{activeCharacter?.level ? `, Level ${activeCharacter.level}` : ""}
        {/if}
      </span>
    </span>
    <i class="material-icons chevron">{open ? "expand_less" : "expand_more"}</i>
  </button>

  {#if open}
    <div class="menu">
      <div class="menu-header">
        <span>Characters</span>
        <button class="refresh" on:click={loadCharacters} title="Refresh characters">
          <i class="material-icons" style="font-size: 16px;">refresh</i>
        </button>
      </div>
      {#if error}
        <div class="notice">{error}</div>
      {:else if loading && characters.length === 0}
        <div class="notice">Loading characters...</div>
      {:else if characters.length === 0}
        <div class="notice">No characters found.</div>
      {:else}
        {#each characters as character (character.id)}
          <button
            class="character-row"
            class:active={character.id === activeCharacter?.id}
            disabled={!canSwitch || character.id === activeCharacter?.id}
            on:click={() => selectCharacter(character)}
          >
            <span class="identity">
              <span class="name">{character.name}</span>
              <span class="meta">{className(character)} | {character.race?.name || "Unknown race"}</span>
            </span>
            <span class="level">Lv {character.level || 1}</span>
          </button>
        {/each}
      {/if}
    </div>
  {/if}
</div>
