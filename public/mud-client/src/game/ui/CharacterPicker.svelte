<script>
  import { onMount, onDestroy } from "svelte";
  import { getMyCharacters } from "../../api/characters.js";
  import { showCharacterWizard } from "../../onboarding/onboardingStore.js";

  export let authToken = "";
  export let activeCharacter = null;
  export let canSwitch = false;
  export let sendMessage = () => {};
  export let onClose = () => {};

  let characters = [];
  let loading = false;
  let error = "";
  let panel;
  let loadedFor = "";

  function asList(data) {
    if (Array.isArray(data)) return data;
    if (data && Array.isArray(data.characters)) return data.characters;
    return [];
  }

  function className(character) {
    return character?.class?.name || character?.class?.Name || "Adventurer";
  }

  function loadCharacters() {
    if (!authToken || loading) return;
    loading = true;
    error = "";
    getMyCharacters(
      authToken,
      (data) => {
        characters = asList(data);
        loading = false;
      },
      () => {
        error = "Could not load characters";
        loading = false;
      }
    );
  }

  function choose(character) {
    if (!character || !canSwitch || character.id === activeCharacter?.id) return;
    const name = (character.name || "").trim();
    if (!name) return;
    sendMessage(`sc ${name}`);
    onClose();
  }

  function newCharacter() {
    showCharacterWizard.set(true);
    onClose();
  }

  function onKey(event) {
    if (event.key === "Escape") onClose();
  }

  onMount(() => {
    window.addEventListener("keydown", onKey);
    if (panel) panel.focus();
  });

  onDestroy(() => {
    window.removeEventListener("keydown", onKey);
  });

  $: if (authToken && authToken !== loadedFor && !loading) {
    loadedFor = authToken;
    loadCharacters();
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="char-picker" role="presentation" on:click={onClose} on:keydown={onKey}>
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div
    class="picker-card"
    bind:this={panel}
    role="dialog"
    aria-modal="true"
    aria-label="Characters"
    tabindex="-1"
    on:click|stopPropagation
  >
    <div class="picker-head">
      <h2>Characters</h2>
      <div class="head-actions">
        <button class="icon" type="button" title="Refresh characters" aria-label="Refresh characters" on:click={loadCharacters}>
          <i class="material-icons">refresh</i>
        </button>
        <button class="icon" type="button" title="Close" aria-label="Close" on:click={onClose}>
          <i class="material-icons">close</i>
        </button>
      </div>
    </div>

    <div class="list">
      {#if error}
        <p class="notice">{error}</p>
      {:else if loading && characters.length === 0}
        <p class="notice">Loading characters...</p>
      {:else if characters.length === 0}
        <p class="notice">No characters yet.</p>
      {:else}
        {#each characters as character (character.id || character.name)}
          <button
            class="row"
            class:active={character.id === activeCharacter?.id}
            type="button"
            disabled={!canSwitch || character.id === activeCharacter?.id}
            on:click={() => choose(character)}
          >
            <span class="identity">
              <span class="name">{character.name}</span>
              <span class="meta">{className(character)} | {character.race?.name || "Unknown race"}</span>
            </span>
            <span class="level">
              {#if character.id === activeCharacter?.id}
                Playing
              {:else}
                Lv {character.level || 1}
              {/if}
            </span>
          </button>
        {/each}
      {/if}
    </div>

    <div class="picker-foot">
      {#if !canSwitch}
        <p class="notice">Connecting… you can switch when the game is online.</p>
      {/if}
      <button class="new-btn" type="button" on:click={newCharacter}>
        <i class="material-icons">person_add</i>
        New character
      </button>
    </div>
  </div>
</div>

<style>
  .char-picker {
    position: fixed;
    inset: 0;
    z-index: 2000;
    width: auto;
    height: auto;
    margin: 0;
    visibility: visible;
    opacity: 1;
    background: rgba(0, 0, 0, 0.72);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    box-sizing: border-box;
  }

  .picker-card {
    width: min(92vw, 420px);
    max-height: min(80vh, 640px);
    display: flex;
    flex-direction: column;
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 10px;
    background: rgba(7, 9, 12, 0.98);
    color: #f0e6d3;
    font-family: "Cinzel", serif;
    box-shadow: 0 18px 46px rgba(0, 0, 0, 0.55);
    outline: none;
  }

  .picker-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 0.75rem 0.85rem;
    border-bottom: 1px solid rgba(251, 191, 36, 0.28);
  }

  h2 {
    margin: 0;
    font-size: 1rem;
    letter-spacing: 0.04em;
    color: #fbbf24;
  }

  .head-actions {
    display: flex;
    gap: 6px;
  }

  .icon {
    width: 32px;
    height: 32px;
    padding: 0;
    margin: 0;
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.55);
    color: #fbbf24;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .icon i { font-size: 18px; }

  .list {
    overflow: auto;
    min-height: 0;
  }

  .row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.75rem;
    width: 100%;
    margin: 0;
    padding: 0.7rem 0.85rem;
    border: 0;
    border-radius: 0;
    background: transparent;
    color: #f0e6d3;
    cursor: pointer;
    text-align: left;
    font-family: "Cinzel", serif;
  }

  .row:hover,
  .row.active {
    background: rgba(251, 191, 36, 0.16);
  }

  .row:disabled {
    cursor: default;
    opacity: 0.85;
  }

  .identity {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .name {
    font-size: 0.92rem;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .meta {
    font-size: 0.68rem;
    color: rgba(240, 230, 211, 0.68);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .level {
    color: #fbbf24;
    font-size: 0.74rem;
    align-self: center;
  }

  .notice {
    margin: 0;
    padding: 0.75rem 0.85rem;
    color: rgba(240, 230, 211, 0.72);
    font-size: 0.78rem;
  }

  .picker-foot {
    border-top: 1px solid rgba(251, 191, 36, 0.28);
  }

  .new-btn {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    width: 100%;
    margin: 0;
    padding: 0.75rem 0.85rem;
    border: 0;
    border-radius: 0 0 10px 10px;
    background: transparent;
    color: #fbbf24;
    cursor: pointer;
    text-align: left;
    font-family: "Cinzel", serif;
    font-size: 0.85rem;
  }

  .new-btn:hover { background: rgba(251, 191, 36, 0.16); }
  .new-btn i { font-size: 18px; }
</style>
