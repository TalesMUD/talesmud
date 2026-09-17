<script>
  export let store;
  export let sendMessage;

  let showPlayersOverlay = false;
  let chatEl;

  $: playerCount = ($store.players || []).length;
  $: roomChat = $store.roomChat || [];
  $: roomName = $store.roomName;

  // Close when the room itself changes, not on every store tick.
  let lastRoomName = $store.roomName;
  $: if (roomName !== lastRoomName) {
    lastRoomName = roomName;
    showPlayersOverlay = false;
  }

  $: if (showPlayersOverlay && chatEl && roomChat.length) {
    chatEl.scrollTop = chatEl.scrollHeight;
  }

  function togglePlayersOverlay() {
    showPlayersOverlay = !showPlayersOverlay;
  }

  function whisperPlayer(player) {
    sendMessage(`tell ${player.name} `);
    showPlayersOverlay = false;
  }

  function inspectPlayer(player) {
    sendMessage(`inspect ${player.name}`);
    showPlayersOverlay = false;
  }

  function isGuestClient() {
    try {
      return typeof sessionStorage !== 'undefined' && !!sessionStorage.getItem('talesmud_guest_token');
    } catch (err) {
      return false;
    }
  }

  function addFriend(player) {
    if (!player || !player.name || isGuestClient()) return;
    sendMessage(`friend add ${player.name}`);
  }

  function openFriends() {
    if (isGuestClient()) return;
    if (store && store.openFriendsOverlay) store.openFriendsOverlay();
    showPlayersOverlay = false;
  }
</script>

<style>
  .player-badge {
    position: absolute;
    top: 0.8em;
    right: 0.8em;
    z-index: 15;
    display: flex;
    align-items: center;
    gap: 0.3em;
    padding: 0.28em 0.65em;
    background: var(--panel-bg, rgba(15, 10, 5, 0.88));
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1px solid var(--panel-border, rgba(212, 175, 55, 0.35));
    border-radius: 14px;
    cursor: pointer;
    transition: background 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
    font-size: 12px;
    color: var(--text-primary, #e8dcc8);
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.35), inset 0 1px 0 rgba(212, 175, 55, 0.08);
  }

  .player-badge:hover {
    background: var(--btn-hover-bg, rgba(180, 130, 60, 0.2));
    border-color: var(--panel-border-hover, rgba(212, 175, 55, 0.55));
  }

  .player-badge.active {
    background: var(--accent-glow, rgba(212, 164, 74, 0.18));
    border-color: var(--accent-primary, #d4a44a);
    box-shadow: 0 0 12px rgba(212, 164, 74, 0.15), inset 0 1px 0 rgba(212, 175, 55, 0.12);
  }

  .badge-count {
    font-weight: 700;
    font-size: 13px;
    color: var(--accent-primary, #d4a44a);
    font-variant-numeric: tabular-nums;
  }

  .badge-icon {
    font-size: 12px;
    line-height: 1;
    opacity: 0.9;
  }

  .players-overlay {
    position: absolute;
    top: 3em;
    right: 0.8em;
    z-index: 14;
    min-width: 180px;
    max-width: 240px;
    background:
      linear-gradient(180deg, rgba(180, 130, 60, 0.07) 0%, transparent 28%),
      var(--panel-bg, rgba(12, 8, 4, 0.94));
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border: 1px solid var(--panel-border, rgba(212, 175, 55, 0.32));
    border-radius: var(--panel-radius, 10px);
    padding: 0.45em 0.4em 0.5em;
    animation: fadeIn 0.15s ease-out;
    display: flex;
    flex-direction: column;
    min-height: 0;
    box-shadow: var(--panel-shadow, 0 4px 18px rgba(0, 0, 0, 0.5), inset 0 1px 0 rgba(212, 175, 55, 0.1));
  }

  .players-overlay.expanded {
    width: min(340px, calc(100% - 1.6em));
    max-width: 360px;
    height: 50%;
    min-height: 180px;
  }

  .players-overlay-title {
    font-family: var(--font-display, 'Cinzel', serif);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: var(--panel-header-color, #d4a44a);
    padding: 0.35em 0.55em 0.45em;
    margin-bottom: 0.15em;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    text-shadow: var(--panel-header-shadow, 0 1px 3px rgba(0, 0, 0, 0.45));
    border-bottom: 1px solid var(--panel-header-border, rgba(180, 130, 60, 0.22));
    background: var(--panel-header-bg, linear-gradient(180deg, rgba(180, 130, 60, 0.1) 0%, transparent 100%));
    border-radius: 6px 6px 0 0;
  }

  .roster {
    flex-shrink: 0;
    max-height: 36%;
    overflow-y: auto;
    padding-top: 0.2em;
  }

  .player-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.4em 0.5em;
    border-radius: 6px;
    transition: background 0.15s ease;
  }

  .player-row:hover {
    background: var(--panel-inner-hover, rgba(180, 130, 60, 0.1));
  }

  .player-info {
    display: flex;
    flex-direction: column;
    gap: 0.12em;
    min-width: 0;
  }

  .player-name {
    font-weight: 600;
    font-size: 15px;
    color: var(--text-primary, #e8dcc8);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .player-tag {
    font-family: var(--font-display, 'Cinzel', serif);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .player-tag.you {
    color: var(--color-gold, #fbbf24);
  }

  .player-tag.other {
    color: var(--text-secondary, #a89878);
  }

  .player-actions {
    display: flex;
    gap: 0.3em;
    flex-shrink: 0;
  }

  .player-action-btn {
    font-size: 14px;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    border: 1px solid var(--btn-border, rgba(180, 130, 60, 0.3));
    background: var(--btn-bg, rgba(180, 130, 60, 0.1));
    color: var(--text-primary, #e8dcc8);
    cursor: pointer;
    transition: background 0.15s ease, border-color 0.15s ease;
    padding: 0;
  }

  .player-action-btn:hover {
    background: var(--btn-hover-bg, rgba(180, 130, 60, 0.22));
    border-color: var(--btn-hover-border, rgba(180, 130, 60, 0.5));
  }

  .player-action-btn.whisper {
    border-color: rgba(212, 175, 55, 0.35);
    color: #e8d4a0;
  }

  .player-action-btn.whisper:hover {
    background: rgba(212, 175, 55, 0.16);
  }

  .player-action-btn.inspect {
    border-color: rgba(212, 164, 74, 0.4);
    color: #fde68a;
  }

  .player-action-btn.inspect:hover {
    background: rgba(251, 191, 36, 0.18);
  }

  .player-action-btn.friends-link {
    width: auto;
    padding: 0 9px;
    height: 24px;
    font-family: var(--font-display, 'Cinzel', serif);
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--accent-primary, #d4a44a);
    border-color: var(--btn-border, rgba(180, 130, 60, 0.4));
    background: var(--btn-bg, rgba(180, 130, 60, 0.12));
  }

  .player-action-btn.friends-link:hover {
    color: #f0e6d3;
    border-color: var(--accent-primary, #d4a44a);
    background: var(--accent-glow, rgba(212, 164, 74, 0.2));
  }

  .chat-divider {
    height: 1px;
    margin: 0.45em 0.45em 0.35em;
    flex-shrink: 0;
    border: 0;
    background: linear-gradient(
      90deg,
      transparent 0%,
      var(--divider-color, rgba(180, 130, 60, 0.35)) 18%,
      var(--divider-color, rgba(180, 130, 60, 0.45)) 50%,
      var(--divider-color, rgba(180, 130, 60, 0.35)) 82%,
      transparent 100%
    );
    opacity: 0.9;
  }

  .room-chat {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 0.15em 0.35em 0.35em;
  }

  .chat-empty {
    font-size: 12.5px;
    color: var(--text-dim, #6d5f48);
    padding: 0.45em 0.25em;
    font-style: italic;
  }

  .chat-line {
    display: flex;
    flex-direction: column;
    gap: 0.08em;
    padding: 0.32em 0.2em 0.38em;
    border-bottom: 1px solid transparent;
    background-image: linear-gradient(
      90deg,
      transparent 0%,
      rgba(180, 130, 60, 0.12) 20%,
      rgba(180, 130, 60, 0.12) 80%,
      transparent 100%
    );
    background-size: 100% 1px;
    background-repeat: no-repeat;
    background-position: bottom;
  }

  .chat-line:last-child {
    background-image: none;
  }

  .chat-who {
    font-family: var(--font-display, 'Cinzel', serif);
    font-size: 11.5px;
    font-weight: 600;
    letter-spacing: 0.06em;
    color: var(--accent-primary, #d4a44a);
  }

  .chat-who.you {
    color: var(--color-gold, #fbbf24);
  }

  .chat-text {
    font-size: 13.5px;
    color: var(--text-primary, #e8dcc8);
    line-height: 1.4;
    word-break: break-word;
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-4px); }
    to { opacity: 1; transform: translateY(0); }
  }

  @media (max-width: 640px) {
    .players-overlay.expanded {
      width: min(300px, calc(100% - 1.2em));
      max-width: 320px;
      right: 0.5em;
      height: 45%;
      min-height: 160px;
    }

    .player-badge {
      top: 0.55em;
      right: 0.55em;
    }
  }
</style>

{#if playerCount > 0}
  <button
    class="player-badge"
    class:active={showPlayersOverlay}
    on:click={togglePlayersOverlay}
    title="Players in room"
  >
    <span class="badge-count">{playerCount}</span>
    <span class="badge-icon">&#x1F465;</span>
  </button>

  {#if showPlayersOverlay}
    <div class="players-overlay expanded">
      <div class="players-overlay-title">
        In this room
        {#if !isGuestClient()}
          <button class="player-action-btn friends-link whisper" type="button" on:click={openFriends} title="Open friends">Friends</button>
        {/if}
      </div>
      <div class="roster">
        {#each $store.players as player (player.id)}
          <div class="player-row">
            <div class="player-info">
              <span class="player-name">{player.name}</span>
              <span class="player-tag" class:you={player.isYou} class:other={!player.isYou}>
                {player.isYou ? 'You' : 'Player'}
              </span>
            </div>
            {#if !player.isYou}
              <div class="player-actions">
                <button
                  class="player-action-btn whisper"
                  on:click={() => whisperPlayer(player)}
                  title="Whisper to {player.name}"
                >&#x1F4AC;</button>
                <button
                  class="player-action-btn inspect"
                  on:click={() => inspectPlayer(player)}
                  title="Inspect {player.name}"
                >&#x1F50D;</button>
                {#if !isGuestClient()}
                  <button
                    class="player-action-btn whisper"
                    on:click={() => addFriend(player)}
                    title="Add {player.name} as friend"
                  >+</button>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
      <div class="chat-divider"></div>
      <div class="room-chat" bind:this={chatEl}>
        {#if roomChat.length === 0}
          <div class="chat-empty">No room chat yet. Say something.</div>
        {:else}
          {#each roomChat as line (line.id)}
            <div class="chat-line">
              <span class="chat-who" class:you={line.isYou}>
                {line.isYou ? `${line.name} (YOU)` : (line.name || 'Someone')}
              </span>
              <span class="chat-text">{line.text}</span>
            </div>
          {/each}
        {/if}
      </div>
    </div>
  {/if}
{/if}
