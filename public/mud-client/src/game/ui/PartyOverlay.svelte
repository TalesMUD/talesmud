<script>
  import { onDestroy, tick } from 'svelte';
  import { hashedAvatar } from '../portraitSrc.js';
  import { parsePartyChatLine } from '../partyState.js';

  export let store = null;
  export let sendMessage = null;

  let inviteName = '';
  let sayText = '';
  let escHandler = null;
  let closeBtn;
  let localPartyChat = [];
  let wasOpen = false;

  $: open = !!(store && $store && $store.partyOverlayOpen);
  $: party = (store && $store && $store.party) || {
    inParty: false, partyId: '', partyName: '', leaderId: '', maxMembers: 5, members: [],
  };
  $: members = Array.isArray(party.members) ? party.members : [];
  $: invite = (store && $store && $store.partyInvite) || null;
  $: guest = isGuestClient();
  $: me = (store && $store && $store.character) || null;
  $: myId = me && me.id ? String(me.id) : '';
  $: myName = me && me.name ? String(me.name) : '';
  $: storeChat = (store && $store && $store.partyChat) || [];
  $: partyChat = mergeChat(storeChat, localPartyChat);
  $: onlineCount = members.filter((m) => m.online).length;
  $: maxMembers = party.maxMembers || 5;
  $: memberCount = members.length;
  $: iAmLeader = !!(myId && (party.leaderId === myId || members.some((m) => m.id === myId && m.isLeader)));
  $: title = party.inParty && party.partyName ? party.partyName : 'Party';
  $: subtitle = party.inParty
    ? `${memberCount}/${maxMembers} members · ${onlineCount} online`
    : '';

  function isGuestClient() {
    try {
      return typeof sessionStorage !== 'undefined' && !!sessionStorage.getItem('talesmud_guest_token');
    } catch (err) {
      return false;
    }
  }

  function mergeChat(a, b) {
    const seen = new Set();
    const out = [];
    for (const line of [...(a || []), ...(b || [])]) {
      if (!line || !line.id || seen.has(line.id)) continue;
      seen.add(line.id);
      out.push(line);
    }
    return out.slice(-8);
  }

  function isYou(member) {
    if (!member) return false;
    if (myId && member.id && String(member.id) === myId) return true;
    if (myName && member.name && String(member.name).toLowerCase() === myName.toLowerCase()) return true;
    return false;
  }

  function avatarSrc(member) {
    if (member && member.portrait) return member.portrait;
    return hashedAvatar((member && (member.id || member.name)) || 'party');
  }

  function initialOf(member) {
    const n = (member && member.name) || '?';
    return n.charAt(0).toUpperCase();
  }

  function classLevelLine(member) {
    const cls = member && member.class ? member.class : '';
    const lvl = member && member.level ? `Lv ${member.level}` : '';
    if (cls && lvl) return `${cls} · ${lvl}`;
    return cls || lvl || '';
  }

  $: if (open && !wasOpen) {
    wasOpen = true;
    if (!guest && sendMessage) sendMessage('party list');
    tick().then(() => {
      if (closeBtn && typeof closeBtn.focus === 'function') closeBtn.focus();
    });
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
    if (typeof window !== 'undefined' && !window.confirm('Leave this party?')) return;
    sendMessage('party leave');
  }

  function kickMember(member) {
    if (!member || !member.name || !sendMessage || !iAmLeader || isYou(member)) return;
    if (typeof window !== 'undefined' && !window.confirm(`Kick ${member.name} from the party?`)) return;
    sendMessage(`party kick ${member.name}`);
  }

  function sendSay() {
    const text = String(sayText || '').trim();
    if (!text || !sendMessage) return;
    sendMessage(`party say ${text}`);
    const line = parsePartyChatLine(`[Party] ${myName || 'You'}: ${text}`, myName);
    if (line) {
      localPartyChat = [...localPartyChat, line].slice(-8);
    }
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

  function onAvatarError(ev, member) {
    const img = ev && ev.currentTarget;
    if (!img || img.dataset.fallback === '1') {
      if (img) img.style.display = 'none';
      return;
    }
    img.dataset.fallback = '1';
    img.src = hashedAvatar((member && (member.id || member.name)) || 'party');
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
    background: var(--social-panel-bg, rgba(12, 16, 24, 0.97));
    border: 1px solid var(--social-border, rgba(212, 175, 55, 0.28));
    border-radius: 10px;
    overflow: hidden;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.55);
  }
  .party-header {
    display: flex;
    align-items: flex-start;
    gap: 0.75em;
    padding: 0.75em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: var(--social-header-bg, rgba(20, 26, 36, 0.9));
    flex-shrink: 0;
  }
  .party-heading { flex: 1; min-width: 0; }
  .party-title {
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
  .party-subtitle {
    margin-top: 4px;
    color: var(--social-muted, #8a8070);
    font-size: 11px;
    letter-spacing: 0.03em;
  }
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
    padding: 0.35em 0.75em 0.5em;
  }
  .empty {
    color: var(--social-muted, #8a8070);
    font-size: 13px;
    padding: 1.2em 0.4em 0.6em;
    text-align: center;
    line-height: 1.45;
  }
  .empty-hint {
    display: block;
    margin-top: 0.55em;
    font-size: 11px;
    color: #6b7280;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 6px;
    border-bottom: 1px solid rgba(255,255,255,0.05);
  }
  .avatar-wrap {
    position: relative;
    width: 40px;
    height: 40px;
    flex-shrink: 0;
  }
  .avatar, .avatar-fallback {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    border: 1px solid rgba(212, 175, 55, 0.35);
    background: rgba(0,0,0,0.35);
  }
  .avatar-fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fbbf24;
    font-weight: 700;
    font-size: 16px;
  }
  .who {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .who-top {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }
  .who-name { color: var(--social-text, #f3ead4); font-weight: 700; font-size: 14px; }
  .you-tag {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: #1a1408;
    background: #fbbf24;
    border-radius: 999px;
    padding: 1px 6px;
    font-weight: 700;
  }
  .leader-badge {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: #fbbf24;
    border: 1px solid rgba(251, 191, 36, 0.45);
    border-radius: 999px;
    padding: 1px 6px;
    font-weight: 700;
  }
  .who-meta { color: #9ca3af; font-size: 11px; }
  .pill {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    border-radius: 999px;
    padding: 3px 8px;
    font-weight: 700;
    flex-shrink: 0;
    background: rgba(107, 114, 128, 0.25);
    color: #9ca3af;
  }
  .pill.online {
    background: rgba(74, 222, 128, 0.15);
    color: #86efac;
  }
  .kick-btn {
    border: 1px solid var(--social-danger-border, rgba(248, 113, 113, 0.4));
    background: transparent;
    color: var(--social-danger, #fca5a5);
    border-radius: 6px;
    min-height: 32px;
    padding: 0 8px;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    cursor: pointer;
  }
  .action-bar {
    flex-shrink: 0;
    border-top: 1px solid rgba(148,163,184,0.15);
    background: rgba(10, 12, 18, 0.92);
    padding: 10px 12px 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .party-chat {
    max-height: 7.5em;
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 3px;
    padding: 4px 2px 6px;
    border-bottom: 1px solid rgba(148,163,184,0.12);
  }
  .chat-line {
    font-size: 11px;
    color: #c4b5a0;
    line-height: 1.35;
  }
  .chat-line.system { color: #8a8070; font-style: italic; }
  .chat-line .who { color: #fbbf24; font-weight: 700; display: inline; }
  .compose, .adder, .footer-actions {
    display: flex;
    gap: 6px;
    align-items: center;
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
    background: var(--social-gold-fill, #c9a227);
    border-color: var(--social-gold, #d4af37);
    color: var(--social-ink, #1a1408);
  }
  .act.secondary {
    background: transparent;
    border-color: rgba(212, 175, 55, 0.55);
    color: #fbbf24;
  }
  .act.leave {
    border-color: var(--social-danger-border, rgba(248, 113, 113, 0.4));
    color: var(--social-danger, #fca5a5);
    background: transparent;
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
    .party-overlay {
      align-items: flex-end;
      padding: 0;
      padding-bottom: env(safe-area-inset-bottom, 0px);
    }
    .party-panel {
      width: 100%;
      max-height: min(92dvh, 720px);
      border-radius: 14px 14px 0 0;
    }
    .action-bar .compose,
    .action-bar .adder,
    .action-bar .footer-actions {
      flex-wrap: wrap;
    }
    .action-bar .act {
      min-height: 44px;
      flex: 1 1 auto;
    }
    .compose input, .adder input { min-height: 44px; }
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
        <div class="party-heading">
          <div class="party-title"><i class="material-icons" aria-hidden="true">groups</i> {title}</div>
          {#if subtitle}
            <div class="party-subtitle">{subtitle}</div>
          {/if}
        </div>
        <button class="party-close" type="button" bind:this={closeBtn} on:click={close} aria-label="Close party">×</button>
      </div>
      {#if guest}
        <div class="guest-note">Parties are for lasting adventurers. Sign in to form a party.</div>
      {:else if !party.inParty}
        <div class="party-body">
          <div class="empty">
            No party yet — gather up to {maxMembers} adventurers.
            <span class="empty-hint">Invite requires the target online with a lasting (non-guest) account.</span>
          </div>
        </div>
        <div class="action-bar">
          <div class="footer-actions">
            <button class="act gold" type="button" on:click={createParty}>Create</button>
          </div>
          <form class="adder" on:submit|preventDefault={invitePlayer}>
            <input
              bind:value={inviteName}
              placeholder="Invite by name"
              maxlength="40"
              aria-label="Invite player by name"
            />
            <button class="act secondary" type="submit">Invite</button>
          </form>
        </div>
      {:else}
        <div class="party-body">
          {#if members.length === 0}
            <div class="empty">No members listed yet.</div>
          {:else}
            {#each members as member (member.id || member.name)}
              <div class="row">
                <div class="avatar-wrap">
                  {#if member.portrait}
                    <img
                      class="avatar"
                      src={avatarSrc(member)}
                      alt=""
                      on:error={(e) => onAvatarError(e, member)}
                    />
                  {:else}
                    <div class="avatar-fallback" aria-hidden="true">{initialOf(member)}</div>
                  {/if}
                </div>
                <div class="who">
                  <div class="who-top">
                    <span class="who-name">{member.name}</span>
                    {#if isYou(member)}<span class="you-tag">You</span>{/if}
                    {#if member.isLeader || (party.leaderId && member.id === party.leaderId)}
                      <span class="leader-badge">Leader</span>
                    {/if}
                  </div>
                  {#if classLevelLine(member)}
                    <span class="who-meta">{classLevelLine(member)}</span>
                  {/if}
                </div>
                <span class="pill" class:online={member.online}>{member.online ? 'Online' : 'Offline'}</span>
                {#if iAmLeader && !isYou(member)}
                  <button class="kick-btn" type="button" on:click={() => kickMember(member)} aria-label={`Kick ${member.name}`}>Kick</button>
                {/if}
              </div>
            {/each}
          {/if}
        </div>
        <div class="action-bar">
          {#if partyChat.length}
            <div class="party-chat" aria-live="polite" aria-label="Party chat">
              {#each partyChat as line (line.id)}
                <div class="chat-line" class:system={line.system}>
                  {#if line.system}
                    {line.text}
                  {:else}
                    <span class="who">{line.name}{line.isYou ? ' (You)' : ''}:</span> {line.text}
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
          <form class="compose say" on:submit|preventDefault={sendSay}>
            <input
              bind:value={sayText}
              placeholder="Party say…"
              maxlength="240"
              aria-label="Party say message"
            />
            <button class="act gold" type="submit">Say</button>
          </form>
          <form class="adder" on:submit|preventDefault={invitePlayer}>
            <input
              bind:value={inviteName}
              placeholder="Invite player by name"
              maxlength="40"
              aria-label="Invite player by name"
            />
            <button class="act secondary" type="submit">Invite</button>
          </form>
          <div class="footer-actions">
            <button class="act leave" type="button" on:click={leaveParty}>Leave</button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}
