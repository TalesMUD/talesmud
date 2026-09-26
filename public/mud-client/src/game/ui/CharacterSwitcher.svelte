<script>
  import { onMount, onDestroy } from "svelte";
  import { getUser } from "../../api/user.js";
  import { getAuth } from "../../auth.js";
  import { isGuestSession, clearGuestToken } from "../../authSession.js";
  import { user } from "../../stores.js";
  import { layoutStore } from "../layout/LayoutStore.js";
  import { settingsStore } from "../SettingsStore.js";
  import { openCharacterPicker } from "./characterPickerStore.js";

  export let store;
  export let authToken;

  const { login, logout } = getAuth();

  let open = false;
  let narrow = false;
  let root;

  $: activeCharacter = $store.character;
  $: connectionStatus = $store.connectionStatus;

  function className(character) {
    return character?.class?.name || character?.class?.Name || "Adventurer";
  }

  onMount(() => {
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

  $: guest = isGuestSession(authToken);
  $: showFriends = !guest;
  $: showParty = !guest;
  $: resting = !!(!$store.inCombat && $store.characterStats?.resting);
  $: if (showFriends && authToken) {
    loadAccount();
  }

  let accountLoaded = false;
  function loadAccount() {
    if (accountLoaded || !authToken || guest) return;
    accountLoaded = true;
    getUser(authToken, (u) => user.set(u), () => {});
  }

  function toggleOpen() {
    open = !open;
  }

  function switchCharacter() {
    open = false;
    openCharacterPicker();
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
    if (guest) {
      clearGuestToken();
      window.location.reload();
      return;
    }
    logout();
  }

  function loginToSave() {
    open = false;
    if (login) login();
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
      <button class="menu-item" type="button" role="menuitem" on:click={switchCharacter}>
        <i class="material-icons">switch_account</i>
        Switch character
      </button>
      <div class="menu-rule"></div>
      <button class="menu-item" type="button" role="menuitem" on:click={openSettings}>
        <i class="material-icons">settings</i>
        Settings
      </button>
      {#if guest}
        <button class="menu-item" type="button" role="menuitem" on:click={loginToSave}>
          <i class="material-icons">login</i>
          Log in / Save progress
        </button>
      {:else if $user && ($user.role === "creator" || $user.role === "admin")}
        <a class="menu-item" role="menuitem" href="/creator" target="_blank" rel="noreferrer">
          <i class="material-icons">public</i>
          World Builder
        </a>
      {/if}
      <button class="menu-item" type="button" role="menuitem" on:click={endSession}>
        <i class="material-icons">logout</i>
        {guest ? "End Session" : "Log out"}
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
</style>
