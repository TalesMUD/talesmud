<script>
  import { onDestroy } from 'svelte';

  export let store = null;
  export let sendMessage = null;

  let addName = '';
  let whisperFor = '';
  let whisperText = '';
  let escHandler = null;

  $: open = !!(store && $store && $store.friendsOverlayOpen);
  $: friends = (store && $store && $store.friends) || [];
  $: guest = isGuestClient();
  let wasOpen = false;

  function isGuestClient() {
    try {
      return typeof sessionStorage !== 'undefined' && !!sessionStorage.getItem('talesmud_guest_token');
    } catch (err) {
      return false;
    }
  }

  $: if (open && !wasOpen) {
    wasOpen = true;
    if (!guest && sendMessage) sendMessage('friends');
  } else if (!open && wasOpen) {
    wasOpen = false;
  }

  $: if (open) {
    if (!escHandler) {
      escHandler = (e) => {
        if (e.key === 'Escape') {
          e.preventDefault();
          close();
        }
      };
      if (typeof window !== 'undefined') window.addEventListener('keydown', escHandler);
    }
  } else if (escHandler) {
    if (typeof window !== 'undefined') window.removeEventListener('keydown', escHandler);
    escHandler = null;
    whisperFor = '';
    whisperText = '';
    addName = '';
  }

  function close() {
    if (store && store.closeFriendsOverlay) store.closeFriendsOverlay();
  }

  function addFriend() {
    const name = String(addName || '').trim();
    if (!name || !sendMessage) return;
    sendMessage(`friend add ${name}`);
    addName = '';
  }

  function removeFriend(friend) {
    if (!friend || !friend.name || !sendMessage) return;
    sendMessage(`friend remove ${friend.name}`);
  }

  function startWhisper(friend) {
    whisperFor = friend && friend.name ? friend.name : '';
    whisperText = '';
  }

  function sendWhisper() {
    const text = String(whisperText || '').trim();
    if (!whisperFor || !text || !sendMessage) return;
    sendMessage(`tell ${whisperFor} ${text}`);
    whisperText = '';
    whisperFor = '';
  }

  onDestroy(() => {
    if (escHandler && typeof window !== 'undefined') window.removeEventListener('keydown', escHandler);
  });
</script>

<style>
  .friends-overlay {
    position: fixed;
    inset: 0;
    z-index: 130;
    background: rgba(0, 0, 0, 0.82);
    backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1em;
    padding-top: max(1em, env(safe-area-inset-top, 0px));
    padding-bottom: max(1em, env(safe-area-inset-bottom, 0px));
    box-sizing: border-box;
  }
  .friends-panel {
    width: min(480px, 100%);
    max-height: min(80vh, 640px);
    display: flex;
    flex-direction: column;
    background: rgba(12, 16, 24, 0.97);
    border: 1px solid rgba(212, 175, 55, 0.28);
    border-radius: 10px;
    overflow: hidden;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.55);
  }
  .friends-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
    padding: 0.75em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: rgba(20, 26, 36, 0.9);
    flex-shrink: 0;
  }
  .friends-title {
    flex: 1;
    font-weight: 700;
    color: #f8fafc;
    display: flex;
    align-items: center;
    gap: 0.4em;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    font-size: 13px;
  }
  .friends-title i { color: #c4b5fd; font-size: 1.2em; }
  .friends-close {
    border: none;
    background: transparent;
    color: #94a3b8;
    font-size: 1.4em;
    cursor: pointer;
    line-height: 1;
    min-width: 44px;
    min-height: 44px;
  }
  .friends-body {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 0.5em 0.75em 0.75em;
  }
  .empty {
    color: #8a8070;
    font-size: 13px;
    padding: 1.2em 0.4em;
    text-align: center;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 6px;
    border-bottom: 1px solid rgba(255,255,255,0.05);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #6b7280;
    flex-shrink: 0;
  }
  .dot.online {
    background: #4ade80;
    box-shadow: 0 0 8px rgba(74, 222, 128, 0.7);
  }
  .who {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .who-name { color: #f3ead4; font-weight: 700; font-size: 14px; }
  .who-status { color: #8a8070; font-size: 10px; text-transform: uppercase; letter-spacing: 0.06em; }
  .who-status.online { color: #86efac; }
  .actions { display: flex; gap: 6px; flex-shrink: 0; }
  .act {
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: rgba(255,255,255,0.05);
    color: #e5e7eb;
    border-radius: 6px;
    min-height: 40px;
    min-width: 40px;
    padding: 0 10px;
    cursor: pointer;
    font-size: 12px;
  }
  .act.whisper { border-color: rgba(59, 130, 246, 0.45); color: #93c5fd; }
  .act.remove { border-color: rgba(248, 113, 113, 0.4); color: #fca5a5; }
  .compose, .adder {
    display: flex;
    gap: 6px;
    padding: 8px 6px 4px;
  }
  .compose input, .adder input {
    flex: 1;
    min-width: 0;
    background: rgba(0,0,0,0.35);
    border: 1px solid rgba(148,163,184,0.35);
    color: #e5e7eb;
    border-radius: 6px;
    padding: 8px 10px;
    font: inherit;
    font-size: 14px;
  }
  .adder {
    border-top: 1px solid rgba(148,163,184,0.15);
    padding: 10px 12px 12px;
  }
  .act.add {
    background: #c9a227;
    border-color: #d4af37;
    color: #1a1408;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }
  .guest-note {
    color: #c4b5fd;
    font-size: 13px;
    padding: 1.5em 1em;
    text-align: center;
    line-height: 1.45;
  }
  @media (max-width: 768px) {
    .friends-panel {
      max-height: 90dvh;
      border-radius: 12px;
    }
    .act { min-height: 44px; min-width: 44px; }
  }
</style>

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="friends-overlay" role="dialog" aria-modal="true" aria-label="Friends" on:click={(e) => { if (e.target === e.currentTarget) close(); }}>
    <div class="friends-panel" on:click|stopPropagation>
      <div class="friends-header">
        <div class="friends-title"><i class="material-icons">group</i> Friends</div>
        <button class="friends-close" type="button" on:click={close} aria-label="Close friends">×</button>
      </div>
      {#if guest}
        <div class="guest-note">Friends are for lasting adventurers. Sign in to keep a list.</div>
      {:else}
        <div class="friends-body">
          {#if friends.length === 0}
            <div class="empty">No friends yet. Add someone by name.</div>
          {:else}
            {#each friends as friend (friend.id || friend.name)}
              <div class="row">
                <span class="dot" class:online={friend.online}></span>
                <div class="who">
                  <span class="who-name">{friend.name}</span>
                  <span class="who-status" class:online={friend.online}>{friend.online ? 'Online' : 'Offline'}</span>
                </div>
                <div class="actions">
                  <button class="act whisper" type="button" disabled={!friend.online} on:click={() => startWhisper(friend)} title="Whisper">Tell</button>
                  <button class="act remove" type="button" on:click={() => removeFriend(friend)} title="Remove">Remove</button>
                </div>
              </div>
              {#if whisperFor === friend.name}
                <form class="compose" on:submit|preventDefault={sendWhisper}>
                  <input bind:value={whisperText} placeholder="Whisper to {friend.name}…" maxlength="240" />
                  <button class="act add" type="submit">Send</button>
                </form>
              {/if}
            {/each}
          {/if}
        </div>
        <form class="adder" on:submit|preventDefault={addFriend}>
          <input bind:value={addName} placeholder="Add friend by name" maxlength="40" />
          <button class="act add" type="submit">Add</button>
        </form>
      {/if}
    </div>
  </div>
{/if}
