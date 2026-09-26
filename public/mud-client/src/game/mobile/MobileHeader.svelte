<style>
  .mobile-header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: calc(48px + env(safe-area-inset-top, 0px));
    padding-top: env(safe-area-inset-top, 0px);
    background: rgba(0, 0, 0, 0.9);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-left: 12px;
    padding-right: 12px;
    z-index: 1001;
    box-sizing: border-box;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
  }

  .room-icon {
    font-size: 18px;
    color: #f59e0b;
    flex-shrink: 0;
  }

  .room-name {
    font-family: 'Cinzel', serif;
    font-size: 14px;
    font-weight: 600;
    color: #f0e6d3;
    letter-spacing: 0.08em;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .acct {
    position: relative;
  }

  .acct-btn {
    width: 36px;
    height: 36px;
    box-sizing: border-box;
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.55);
    color: #fbbf24;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    cursor: pointer;
  }

  .acct-btn i { font-size: 20px; }

  .acct-btn:hover { background: rgba(251, 191, 36, 0.2); }

  .acct-menu {
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    z-index: 30;
    min-width: 180px;
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 8px;
    background: rgba(7, 9, 12, 0.96);
    box-shadow: 0 18px 46px rgba(0, 0, 0, 0.55);
    display: flex;
    flex-direction: column;
    max-height: min(70vh, 420px);
    overflow: auto;
  }

  .acct-menu button {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    width: 100%;
    padding: 0.7rem 0.8rem;
    border: 0;
    background: transparent;
    color: #f0e6d3;
    font-family: 'Cinzel', serif;
    font-size: 0.82rem;
    text-align: left;
    cursor: pointer;
  }

  .acct-menu button i { font-size: 18px; color: #fbbf24; }

  .acct-menu button:hover {
    background: rgba(251, 191, 36, 0.2);
    color: #fbbf24;
  }

  .hp-pill {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    border-radius: 10px;
    font-size: 11px;
    font-weight: 600;
    border: 1px solid;
  }

  .hp-pill.hp-high {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .hp-pill.hp-mid {
    background: rgba(245, 158, 11, 0.2);
    border-color: rgba(245, 158, 11, 0.4);
    color: #fcd34d;
  }

  .hp-pill.hp-low {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.4);
    color: #fca5a5;
  }

  .hp-pill i {
    font-size: 13px;
  }

  .rest-pill {
    display: flex;
    align-items: center;
    gap: 3px;
    padding: 3px 8px;
    border-radius: 10px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    background: rgba(34, 197, 94, 0.2);
    border: 1px solid rgba(34, 197, 94, 0.45);
    color: #86efac;
  }
  .rest-pill i { font-size: 13px; }

  .combat-indicator {
    animation: combatPulse 2s ease-in-out infinite;
    border-bottom: 2px solid rgba(239, 68, 68, 0.8);
  }

  @keyframes combatPulse {
    0%, 100% { border-bottom-color: rgba(239, 68, 68, 0.8); }
    50% { border-bottom-color: rgba(239, 68, 68, 0.3); }
  }
</style>

<script>
  import { onMount, onDestroy } from "svelte";
  import { getAuth } from "../../auth.js";
  import { isGuestSession, clearGuestToken } from "../../authSession.js";
  import { layoutStore } from "../layout/LayoutStore.js";
  import { settingsStore } from "../SettingsStore.js";
  import { openCharacterPicker } from "../ui/characterPickerStore.js";

  export let store;
  export let authToken = "";

  const { login, logout } = getAuth();
  $: guest = isGuestSession(authToken);
  let open = false;
  let root;

  function toggle(event) {
    event.stopPropagation();
    open = !open;
  }

  function onKey(event) {
    if (event.key === "Escape") open = false;
  }

  function onPointerDown(event) {
    if (!open || !root) return;
    if (root.contains(event.target)) return;
    open = false;
  }

  onMount(() => {
    window.addEventListener("keydown", onKey);
    window.addEventListener("pointerdown", onPointerDown, true);
  });

  onDestroy(() => {
    window.removeEventListener("keydown", onKey);
    window.removeEventListener("pointerdown", onPointerDown, true);
  });

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

  function switchCharacter() {
    open = false;
    openCharacterPicker();
  }

  $: roomName = $store.roomName || 'Unknown';
  $: stats = $store.characterStats || {};
  $: currentHP = stats.currentHitPoints || 0;
  $: maxHP = stats.maxHitPoints || 1;
  $: hpPercent = Math.round((currentHP / maxHP) * 100);
  $: hpClass = hpPercent > 60 ? 'hp-high' : hpPercent > 30 ? 'hp-mid' : 'hp-low';
  $: inCombat = $store.inCombat;
  $: resting = !!(stats.resting && !inCombat);
</script>

<div class="mobile-header" class:combat-indicator={inCombat}>
  <div class="header-left">
    <i class="material-icons room-icon">explore</i>
    <span class="room-name">{roomName}</span>
  </div>

  <div class="header-right">
    {#if resting}
      <span class="rest-pill"><i class="material-icons">hotel</i> Resting</span>
    {/if}
    <div class="hp-pill {hpClass}">
      <i class="material-icons">favorite</i>
      {currentHP}/{maxHP}
    </div>
    <div class="acct" bind:this={root}>
      <button class="acct-btn" type="button" title="Account" aria-label="Account" aria-expanded={open} on:click={toggle}>
        <i class="material-icons">person</i>
      </button>
      {#if open}
        <div class="acct-menu" role="menu">
          <button type="button" role="menuitem" on:click={editLayout}><i class="material-icons">dashboard_customize</i> Edit Layout</button>
          <button type="button" role="menuitem" on:click={switchCharacter}><i class="material-icons">switch_account</i> Switch character</button>
          <button type="button" role="menuitem" on:click={openSettings}><i class="material-icons">settings</i> Settings</button>
          {#if guest}
            <button type="button" role="menuitem" on:click={loginToSave}><i class="material-icons">login</i> Log in / Save progress</button>
            <button type="button" role="menuitem" on:click={endSession}><i class="material-icons">logout</i> End Session</button>
          {:else}
            <button type="button" role="menuitem" on:click={endSession}><i class="material-icons">logout</i> Log out</button>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>
