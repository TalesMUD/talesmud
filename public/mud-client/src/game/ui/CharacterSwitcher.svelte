<script>
  import { onMount, onDestroy } from "svelte";
  import { getMyCharacters } from "../../api/characters.js";
  import { getUser } from "../../api/user.js";
  import { getAuth } from "../../auth.js";
  import { user } from "../../stores.js";
  import { layoutStore } from "../layout/LayoutStore.js";
  import { settingsStore } from "../SettingsStore.js";

  export let store;
  export let authToken;
  export let sendMessage;

  const { login, logout } = getAuth();

  let characters = [];
  let loading = false;
  let open = false;
  let error = "";
  let narrow = false;
  let root;

  $: activeCharacter = $store.character;
  $: connectionStatus = $store.connectionStatus;
  $: canSwitch = connectionStatus === "connected";

  onMount(() => {
    loadCharacters();
    measure();
    window.addEventListener("resize", measure);
    window.addEventListener("keydown", onKey);
    window.addEventListener("pointerdown", onPointerDown, true);
  });

  onDestroy(() => {
    window.removeEventListener("resize", measure);
    window.removeEventListener("keydown", onKey);
    window.removeEventListener("pointerdown", onPointerDown, true);
  });

  function measure() {
    narrow = window.innerWidth < 1100;
  }

  function onKey(event) {
    if (event.key === "Escape" && open) {
      open = false;
    }
  }

  function onPointerDown(event) {
    if (!open || !root) return;
    if (root.contains(event.target)) return;
    open = false;
  }

  $: if (authToken && characters.length === 0 && !loading) {
    loadCharacters();
  }

  function isGuestClient() {
    try {
      return typeof sessionStorage !== "undefined" && !!sessionStorage.getItem("talesmud_guest_token");
    } catch (err) {
      return false;
    }
  }

  $: showFriends = !isGuestClient();
  $: showParty = !isGuestClient();
  $: resting = !!(!$store.inCombat && $store.characterStats?.resting);
  $: if (showFriends && authToken) {
    loadAccount();
  }

  let accountLoaded = false;
  function loadAccount() {
    if (accountLoaded || !authToken || isGuestClient()) return;
    accountLoaded = true;
    getUser(authToken, (u) => user.set(u), () => {});
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
    if (open && characters.length === 0) loadCharacters();
  }

  function selectCharacter(character) {
    if (!character || !canSwitch || character.id === activeCharacter?.id) return;
    sendMessage(`sc ${character.name}`);
    open = false;
  }

  function className(character) {
    return character?.class?.name || character?.class?.Name || "Adventurer";
  }

  function openFriends() {
    if (!showFriends) return;
    open = false;
    if (store && store.openFriendsOverlay) store.openFriendsOverlay();
  }

  function openParty() {
    if (!showParty) return;
    open = false;
    if (store && store.openPartyOverlay) store.openPartyOverlay();
  }

  function editLayout() {
    open = false;
    layoutStore.enterEditMode();
  }

  function openSettings() {
    open = false;
    settingsStore.openModal();
  }

  function endSession() {
    open = false;
    if (isGuestClient()) {
      try { sessionStorage.removeItem("talesmud_guest_token"); } catch (err) { /* ignore */ }
      window.location.reload();
      return;
    }
    logout();
  }

  function createAccount() {
    open = false;
    if (login) login(null, { screen_hint: "signup" });
  }
</script>

<div class="switcher play-header" bind:this={root}>
  <div class="switcher-row">
    {#if resting}
      <span class="hdr-btn rest-launch" title="Resting"><i class="material-icons">hotel</i> Resting</span>
    {/if}
    {#if !narrow}
      <button class="hdr-btn text-btn" type="button" on:click={editLayout}>Edit Layout</button>
    {/if}
    {#if showParty}
      <button class="hdr-btn icon-btn" type="button" title="Party" aria-label="Party" on:click={openParty}>
        <i class="material-icons">groups</i>
      </button>
    {/if}
    {#if showFriends}
      <button class="hdr-btn icon-btn" type="button" title="Friends" aria-label="Friends" on:click={openFriends}>
        <i class="material-icons">group</i>
      </button>
    {/if}
    <button
      class="hdr-btn switcher-button"
      type="button"
      aria-haspopup="menu"
      aria-expanded={open}
      on:click={toggleOpen}
      disabled={loading && characters.length === 0}
    >
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
  </div>

  {#if open}
    <div class="menu" role="menu">
      {#if narrow}
        <button class="menu-item" type="button" role="menuitem" on:click={editLayout}>
          <i class="material-icons">dashboard_customize</i>
          Edit Layout
        </button>
      {/if}
      <div class="menu-header">
        <span>Characters</span>
        <button class="refresh" type="button" on:click={loadCharacters} title="Refresh characters" aria-label="Refresh characters">
          <i class="material-icons">refresh</i>
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
            type="button"
            role="menuitem"
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
      <div class="menu-rule"></div>
      <button class="menu-item" type="button" role="menuitem" on:click={openSettings}>
        <i class="material-icons">settings</i>
        Settings
      </button>
      {#if isGuestClient()}
        <button class="menu-item" type="button" role="menuitem" on:click={createAccount}>
          <i class="material-icons">person_add</i>
          Create Account
        </button>
      {:else if $user && ($user.role === "creator" || $user.role === "admin")}
        <a class="menu-item" role="menuitem" href="/creator" target="_blank" rel="noreferrer">
          <i class="material-icons">public</i>
          World Builder
        </a>
      {/if}
      <button class="menu-item" type="button" role="menuitem" on:click={endSession}>
        <i class="material-icons">logout</i>
        {isGuestClient() ? "End Session" : "Logout"}
      </button>
    </div>
  {/if}
</div>

<style>
  .switcher {
    position: fixed;
    top: 8px;
    right: 12px;
    z-index: 80;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    font-family: 'Cinzel', serif;
    color: #f0e6d3;
    pointer-events: none;
  }

  .switcher-row {
    display: flex;
    align-items: center;
    gap: 8px;
    pointer-events: auto;
  }

  .hdr-btn {
    pointer-events: auto;
    box-sizing: border-box;
    height: 36px;
    min-height: 36px;
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.55);
    color: #fbbf24;
    cursor: pointer;
    font-family: 'Cinzel', serif;
  }

  .hdr-btn:hover {
    background: rgba(251, 191, 36, 0.2);
  }

  .icon-btn {
    width: 36px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .icon-btn i { font-size: 20px; color: #fbbf24; }

  .text-btn {
    padding: 0 0.75rem;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    white-space: nowrap;
  }

  .rest-launch {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 0 0.7rem;
    color: #86efac;
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    cursor: default;
  }
  .rest-launch i { font-size: 16px; color: #86efac; }

  .switcher-button {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0 0.7rem 0 0.55rem;
    color: #f0e6d3;
    max-width: min(42vw, 280px);
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
    line-height: 1.1;
  }

  .name {
    font-size: 0.82rem;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .meta {
    font-size: 0.62rem;
    color: rgba(240, 230, 211, 0.68);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chevron {
    font-size: 1.1rem;
    color: #fbbf24;
    line-height: 1;
  }

  .menu {
    pointer-events: auto;
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    z-index: 81;
    width: min(86vw, 280px);
    max-height: min(62vh, 440px);
    overflow: auto;
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 8px;
    background: rgba(7, 9, 12, 0.96);
    box-shadow: 0 18px 46px rgba(0, 0, 0, 0.55);
  }

  .menu-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.55rem 0.75rem;
    color: rgba(251, 191, 36, 0.8);
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .refresh {
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 6px;
    background: rgba(0, 0, 0, 0.55);
    color: #fbbf24;
    cursor: pointer;
    width: 28px;
    height: 28px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0;
  }

  .refresh:hover { background: rgba(251, 191, 36, 0.2); }
  .refresh i { font-size: 16px; }

  .menu-item {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    width: 100%;
    box-sizing: border-box;
    padding: 0.65rem 0.85rem;
    border: 0;
    background: transparent;
    color: #f0e6d3;
    cursor: pointer;
    text-align: left;
    text-decoration: none;
    font-family: 'Cinzel', serif;
    font-size: 0.82rem;
  }

  .menu-item i {
    font-size: 18px;
    color: #fbbf24;
  }

  .menu-item:hover {
    background: rgba(251, 191, 36, 0.2);
    color: #fbbf24;
  }

  .menu-rule {
    height: 1px;
    margin: 0.15rem 0.75rem;
    background: rgba(251, 191, 36, 0.28);
  }

  .character-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.75rem;
    width: 100%;
    padding: 0.6rem 0.8rem;
    border: 0;
    background: transparent;
    color: #f0e6d3;
    cursor: pointer;
    text-align: left;
    font-family: 'Cinzel', serif;
  }

  .character-row:hover,
  .character-row.active {
    background: rgba(251, 191, 36, 0.16);
  }

  .character-row:disabled {
    cursor: default;
    opacity: 0.7;
  }

  .level {
    color: #fbbf24;
    font-size: 0.74rem;
    align-self: center;
  }

  .notice {
    padding: 0.7rem 0.8rem;
    color: rgba(240, 230, 211, 0.68);
    font-size: 0.78rem;
  }
</style>
