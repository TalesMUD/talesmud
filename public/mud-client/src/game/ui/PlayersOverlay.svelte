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
    padding: 0.3em 0.6em;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    border: 1px solid rgba(167, 139, 250, 0.4);
    border-radius: 16px;
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 12px;
    color: #e5e7eb;
  }

  .player-badge:hover {
    background: rgba(167, 139, 250, 0.2);
    border-color: rgba(167, 139, 250, 0.6);
  }

  .player-badge.active {
    background: rgba(167, 139, 250, 0.25);
    border-color: rgba(167, 139, 250, 0.7);
  }

  .badge-count {
    font-weight: 700;
    font-size: 13px;
    color: #a78bfa;
  }

  .badge-icon {
    font-size: 13px;
    line-height: 1;
  }

  .players-overlay {
    position: absolute;
    top: 3em;
    right: 0.8em;
    z-index: 14;
    min-width: 180px;
    max-width: 240px;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(167, 139, 250, 0.3);
    border-radius: 8px;
    padding: 0.5em;
    animation: fadeIn 0.15s ease-out;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .players-overlay.expanded {
    width: min(340px, calc(100% - 1.6em));
    max-width: 360px;
    height: 50%;
    min-height: 180px;
  }

  .players-overlay-title {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.8px;
    color: #9ca3af;
    padding: 0.3em 0.5em;
    margin-bottom: 0.3em;
    flex-shrink: 0;
  }

  .roster {
    flex-shrink: 0;
    max-height: 36%;
    overflow-y: auto;
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
    background: rgba(255, 255, 255, 0.05);
  }

  .player-info {
    display: flex;
    flex-direction: column;
    gap: 0.1em;
    min-width: 0;
  }

  .player-name {
    font-weight: 600;
    font-size: 13px;
    color: #e5e7eb;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .player-tag {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .player-tag.you {
    color: #fbbf24;
  }

  .player-tag.other {
    color: #a78bfa;
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
    border: 1px solid rgba(255, 255, 255, 0.15);
    background: rgba(255, 255, 255, 0.05);
    color: #e5e7eb;
    cursor: pointer;
    transition: all 0.15s ease;
    padding: 0;
  }

  .player-action-btn:hover {
    background: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .player-action-btn.whisper {
    border-color: rgba(59, 130, 246, 0.4);
    color: #93c5fd;
  }

  .player-action-btn.whisper:hover {
    background: rgba(59, 130, 246, 0.2);
  }

  .player-action-btn.inspect {
    border-color: rgba(251, 191, 36, 0.4);
    color: #fde68a;
  }

  .player-action-btn.inspect:hover {
    background: rgba(251, 191, 36, 0.2);
  }

  .chat-divider {
    height: 1px;
    background: rgba(167, 139, 250, 0.25);
    margin: 0.4em 0.3em;
    flex-shrink: 0;
  }

  .room-chat {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 0.2em 0.4em 0.4em;
  }

  .chat-empty {
    font-size: 11px;
    color: #6b7280;
    padding: 0.4em 0.2em;
    font-style: italic;
  }

  .chat-line {
    display: flex;
    flex-direction: column;
    gap: 0.05em;
    padding: 0.28em 0.15em;
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  }

  .chat-who {
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.3px;
    color: #c4b5fd;
  }

  .chat-who.you {
    color: #fbbf24;
  }

  .chat-text {
    font-size: 12px;
    color: #e5e7eb;
    line-height: 1.35;
    word-break: break-word;
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-4px); }
    to { opacity: 1; transform: translateY(0); }
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
      <div class="players-overlay-title">In this room</div>
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
