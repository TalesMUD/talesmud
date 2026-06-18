<script>
  import EntityPanel from '../ui/EntityPanel.svelte';
  import DialogOverlay from '../ui/DialogOverlay.svelte';
  import RoomTextOverlay from '../ui/RoomTextOverlay.svelte';
  import { findNpcByName } from '../MUDXPlusStore';
  import { settingsStore } from '../SettingsStore.js';
  import { backend } from '../../api/base.js';

  export let store;
  export let sendMessage;

  let toggleImage = true;
  let showPlayersOverlay = false;

  // Room population from live backend presence
  $: playerCount = ($store.players || []).length;

  // Auto-close overlay when changing rooms
  $: if ($store.roomName) {
    showPlayersOverlay = false;
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

  // Derive NPC type for dialog overlay
  $: dialogNpcType = (() => {
    if (!$store.dialogActive) return 'npc';
    const npc = findNpcByName($store.npcs, $store.dialogNpcName);
    if (npc?.isEnemy) return 'enemy';
    if (npc?.isMerchant) return 'merchant';
    return 'npc';
  })();

  // Handle background changes
  $: if ($store.background) {
    updateRoomImage($store.background);
  }

  function updateRoomImage(background) {
    const oldImg = document.querySelector(toggleImage ? '#roomImg1' : '#roomImg2');
    const newImg = document.querySelector(!toggleImage ? '#roomImg1' : '#roomImg2');

    toggleImage = !toggleImage;

    if (newImg && oldImg) {
      const bgUrl = `${backend}/backgrounds/${background}.png`;
      const placeholderUrl = 'img/placeholder.png';

      // Test if image exists before displaying
      const testImg = new Image();
      testImg.onload = () => {
        newImg.style.backgroundImage = `url('${bgUrl}')`;
        newImg.classList.remove('hidden');
        oldImg.classList.add('hidden');
      };
      testImg.onerror = () => {
        newImg.style.backgroundImage = `url('${placeholderUrl}')`;
        newImg.classList.remove('hidden');
        oldImg.classList.add('hidden');
      };
      testImg.src = bgUrl;
    }
  }
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&display=swap" rel="stylesheet">
</svelte:head>

<style>
  .room-widget {
    display: flex;
    flex-direction: column;
    background: #000;
    border: 1px solid var(--panel-border, rgba(255, 255, 255, 0.1));
    border-radius: var(--panel-radius, 12px);
    overflow: hidden;
    height: 100%;
    box-shadow: var(--panel-shadow, none);
  }

  .roomImageSection {
    position: relative;
    flex: 0 0 60%;
    min-height: 120px;
    overflow: hidden;
  }

  .roomImageInner {
    width: 100%;
    height: 100%;
    background-repeat: no-repeat;
    background-size: cover;
    background-position: center top;
    image-rendering: pixelated;
    opacity: 1;
    transition: opacity 0.8s ease-in-out;
    position: absolute;
    top: 0;
    left: 0;
  }

  .roomImageInner.hidden {
    opacity: 0;
  }

  .roomImageGradient {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 50%;
    background-image: linear-gradient(
      to bottom,
      rgba(0, 0, 0, 0) 0%,
      rgba(0, 0, 0, 0.4) 40%,
      rgba(0, 0, 0, 0.8) 70%,
      rgba(0, 0, 0, 1) 100%
    );
    pointer-events: none;
  }

  .roomName {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    padding: 1.1em 1.2em;
    text-align: center;
    z-index: 10;
    font-family: 'Cinzel', serif;
    font-size: 1.4em;
    font-weight: 600;
    letter-spacing: 0.12em;
    color: #f0e6d3;
    text-shadow:
      0 0 10px rgba(255, 215, 140, 0.3),
      0 2px 4px rgba(0, 0, 0, 0.8),
      0 4px 12px rgba(0, 0, 0, 0.6);
    background: radial-gradient(
      ellipse 70% 100% at 50% 0%,
      rgba(0, 0, 0, 0.75) 0%,
      rgba(0, 0, 0, 0.4) 50%,
      rgba(0, 0, 0, 0) 100%
    );
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.6em;
  }

  .roomName::before,
  .roomName::after {
    content: '◈';
    font-size: 0.6em;
    color: rgba(168, 130, 90, 0.7);
    text-shadow: 0 0 8px rgba(168, 130, 90, 0.4);
  }

  .room-name-text {
    display: flex;
    align-items: center;
    gap: 0.5em;
  }

  .flourish {
    display: inline-block;
    color: rgba(168, 130, 90, 0.6);
    font-size: 0.85em;
  }

  .flourish-left {
    transform: scaleX(-1);
  }

  .roomContentSection {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 1em 1.2em;
    overflow-y: auto;
    background: #000;
    position: relative;
    z-index: 5;
    margin-top: -0.5em;
  }

  .roomDescription {
    color: #e5e7eb;
    font-size: 1.15em;
    line-height: 1.7;
    margin-bottom: 1em;
    flex-shrink: 0;
    font-style: italic;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
    padding: 0.8em 1em;
    border-left: 2px solid rgba(168, 130, 90, 0.6);
    border-radius: 0 4px 4px 0;
    background: linear-gradient(90deg, rgba(168, 130, 90, 0.08) 0%, transparent 100%);
  }

  /* Parchment style variant */
  .roomDescription.parchment {
    background:
      url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E"),
      linear-gradient(90deg, rgba(139, 105, 65, 0.15) 0%, rgba(139, 105, 65, 0.05) 50%, transparent 100%);
    background-blend-mode: overlay, normal;
    box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.3);
  }

  .entitySection {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 10;
    padding: 0.5em 0.8em;
  }

  /* Player count badge */
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

  /* Players overlay */
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
  }

  .players-overlay-title {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.8px;
    color: #9ca3af;
    padding: 0.3em 0.5em;
    margin-bottom: 0.3em;
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

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-4px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>

<div class="room-widget">
  <div class="roomImageSection">
    <div id="roomImg1" class="roomImageInner"></div>
    <div id="roomImg2" class="roomImageInner hidden"></div>
    <div class="roomImageGradient"></div>

    <RoomTextOverlay />

    {#if $store.roomName}
      <div class="roomName">
        <span class="flourish flourish-left">━━</span>
        <span class="room-name-text">{$store.roomName}</span>
        <span class="flourish">━━</span>
      </div>
    {/if}

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
        <div class="players-overlay">
          <div class="players-overlay-title">In this room</div>
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
      {/if}
    {/if}

    {#if $store.dialogActive}
      <DialogOverlay
        npcName={$store.dialogNpcName}
        npcText={$store.dialogNpcText}
        options={$store.dialogOptions}
        npcType={dialogNpcType}
        sendMessage={sendMessage}
      />
    {/if}

    <div class="entitySection">
      <EntityPanel {store} {sendMessage} />
    </div>
  </div>

  <div class="roomContentSection">
    {#if $store.roomDescription}
      <div class="roomDescription" class:parchment={$settingsStore.interface?.parchmentBackground}>{$store.roomDescription}</div>
    {/if}
  </div>
</div>
