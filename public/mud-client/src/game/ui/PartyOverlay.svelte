<script>
  import { onDestroy } from 'svelte';

  export let store = null;
  export let sendMessage = null;

  let inviteName = '';
  let sayText = '';
  let escHandler = null;

  $: open = !!(store && $store && $store.partyOverlayOpen);
  $: party = (store && $store && $store.party) || { inParty: false, partyId: '', partyName: '', members: [] };
  $: members = Array.isArray(party.members) ? party.members : [];
  $: invite = (store && $store && $store.partyInvite) || null;
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
    if (!guest && sendMessage) sendMessage('party list');
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
    inviteName = '';
    sayText = '';
  }

  function close() {
    if (store && store.closePartyOverlay) store.closePartyOverlay();
  }

  function createParty() {
    if (!sendMessage) return;
    sendMessage('party create');
  }

  function invitePlayer() {
    const name = String(inviteName || '').trim();
    if (!name || !sendMessage) return;
    sendMessage(`party invite ${name}`);
    inviteName = '';
  }

  function leaveParty() {
    if (!sendMessage) return;
    sendMessage('party leave');
  }

  function sendSay() {
    const text = String(sayText || '').trim();
    if (!text || !sendMessage) return;
    sendMessage(`party say ${text}`);
    sayText = '';
  }

  function acceptInvite() {
    if (!sendMessage) return;
    sendMessage('party accept');
  }

  function declineInvite() {
    if (!sendMessage) return;
    sendMessage('party decline');
  }

  onDestroy(() => {
    if (escHandler && typeof window !== 'undefined') window.removeEventListener('keydown', escHandler);
  });
</script>

<style>
  .party-overlay {
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
  .party-panel {
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
  .party-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
    padding: 0.75em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: rgba(20, 26, 36, 0.9);
    flex-shrink: 0;
  }
  .party-title {
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
  .party-title i { color: #fbbf24; font-size: 1.2em; }
  .party-close {
    border: none;
    background: transparent;
    color: #94a3b8;
    font-size: 1.4em;
    cursor: pointer;
    line-height: 1;
    min-width: 44px;
    min-height: 44px;
  }
  .party-body {
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
  .party-name {
    color: #fbbf24;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    padding: 4px 6px 10px;
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
  .compose, .adder, .footer-actions {
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
  .adder, .footer-actions, .compose.say {
    border-top: 1px solid rgba(148,163,184,0.15);
    padding: 10px 12px 12px;
  }
  .act {
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: rgba(255,255,255,0.05);
    color: #e5e7eb;
    border-radius: 6px;
    min-height: 40px;
    min-width: 40px;
    padding: 0 12px;
    cursor: pointer;
    font-size: 12px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .act.gold {
    background: #c9a227;
    border-color: #d4af37;
    color: #1a1408;
  }
  .act.leave {
    border-color: rgba(248, 113, 113, 0.4);
    color: #fca5a5;
  }
  .act.accept {
    background: rgba(34, 197, 94, 0.18);
    border-color: rgba(34, 197, 94, 0.45);
    color: #86efac;
  }
  .act.decline {
    border-color: rgba(248, 113, 113, 0.4);
    color: #fca5a5;
  }
  .guest-note {
    color: #fbbf24;
    font-size: 13px;
    padding: 1.5em 1em;
    text-align: center;
    line-height: 1.45;
  }
  .invite-banner {
    position: fixed;
    left: 50%;
    bottom: max(1.2em, env(safe-area-inset-bottom, 0px));
    transform: translateX(-50%);
    z-index: 140;
    width: min(440px, calc(100% - 1.5em));
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: rgba(12, 16, 24, 0.96);
    border: 1px solid rgba(212, 175, 55, 0.4);
    border-radius: 10px;
    box-shadow: 0 12px 36px rgba(0, 0, 0, 0.55);
  }
  .invite-text {
    flex: 1;
    min-width: 0;
    color: #f3ead4;
    font-size: 13px;
    line-height: 1.35;
  }
  .invite-text strong { color: #fbbf24; }
  .invite-actions { display: flex; gap: 6px; flex-shrink: 0; }
  @media (max-width: 768px) {
    .party-panel {
      max-height: 90dvh;
      border-radius: 12px;
    }
    .act { min-height: 44px; }
  }
</style>

{#if invite && invite.pending && !guest}
  <div class="invite-banner" role="status" aria-live="polite">
    <div class="invite-text">
      <strong>{invite.inviterName || 'Someone'}</strong> invited you to a party.
    </div>
    <div class="invite-actions">
      <button class="act accept" type="button" on:click={acceptInvite}>Accept</button>
      <button class="act decline" type="button" on:click={declineInvite}>Decline</button>
    </div>
  </div>
{/if}

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="party-overlay" role="dialog" aria-modal="true" aria-label="Party" on:click={(e) => { if (e.target === e.currentTarget) close(); }}>
    <div class="party-panel" on:click|stopPropagation>
      <div class="party-header">
        <div class="party-title"><i class="material-icons">groups</i> Party</div>
        <button class="party-close" type="button" on:click={close} aria-label="Close party">×</button>
      </div>
      {#if guest}
        <div class="guest-note">Parties are for lasting adventurers. Sign in to form a party.</div>
      {:else if !party.inParty}
        <div class="party-body">
          <div class="empty">You are not in a party yet.</div>
        </div>
        <div class="footer-actions">
          <button class="act gold" type="button" on:click={createParty}>Create Party</button>
        </div>
        <form class="adder" on:submit|preventDefault={invitePlayer}>
          <input bind:value={inviteName} placeholder="Invite by name (creates party)" maxlength="40" />
          <button class="act gold" type="submit">Invite</button>
        </form>
      {:else}
        <div class="party-body">
          {#if party.partyName}
            <div class="party-name">{party.partyName}</div>
          {/if}
          {#if members.length === 0}
            <div class="empty">No members listed yet.</div>
          {:else}
            {#each members as member (member.id || member.name)}
              <div class="row">
                <span class="dot" class:online={member.online}></span>
                <div class="who">
                  <span class="who-name">{member.name}</span>
                  <span class="who-status" class:online={member.online}>{member.online ? 'Online' : 'Offline'}</span>
                </div>
              </div>
            {/each}
          {/if}
        </div>
        <form class="compose say" on:submit|preventDefault={sendSay}>
          <input bind:value={sayText} placeholder="Party say…" maxlength="240" />
          <button class="act gold" type="submit">Say</button>
        </form>
        <form class="adder" on:submit|preventDefault={invitePlayer}>
          <input bind:value={inviteName} placeholder="Invite player by name" maxlength="40" />
          <button class="act gold" type="submit">Invite</button>
        </form>
        <div class="footer-actions">
          <button class="act leave" type="button" on:click={leaveParty}>Leave Party</button>
        </div>
      {/if}
    </div>
  </div>
{/if}
