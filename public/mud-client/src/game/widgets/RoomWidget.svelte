<script>
  import EntityPanel from '../ui/EntityPanel.svelte';
  import DialogOverlay from '../ui/DialogOverlay.svelte';
  import { findNpcByName } from '../MUDXPlusStore';
  import { settingsStore } from '../SettingsStore.js';
  import { backend } from '../../api/base.js';

  export let store;
  export let sendMessage;

  let toggleImage = true;

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
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    overflow: hidden;
    height: 100%;
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
    flex-shrink: 0;
    padding-top: 0.5em;
  }
</style>

<div class="room-widget">
  <div class="roomImageSection">
    <div id="roomImg1" class="roomImageInner"></div>
    <div id="roomImg2" class="roomImageInner hidden"></div>
    <div class="roomImageGradient"></div>

    {#if $store.roomName}
      <div class="roomName">
        <span class="flourish flourish-left">━━</span>
        <span class="room-name-text">{$store.roomName}</span>
        <span class="flourish">━━</span>
      </div>
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
  </div>

  <div class="roomContentSection">
    {#if $store.roomDescription}
      <div class="roomDescription" class:parchment={$settingsStore.interface?.parchmentBackground}>{$store.roomDescription}</div>
    {/if}

    <div class="entitySection">
      <EntityPanel {store} {sendMessage} />
    </div>
  </div>
</div>
