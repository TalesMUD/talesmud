<script>
  import EntityPanel from '../ui/EntityPanel.svelte';
  import DialogOverlay from '../ui/DialogOverlay.svelte';
  import { findNpcByName } from '../MUDXPlusStore';

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
      newImg.style.backgroundImage = `url('/play/img/bg/${background}.png')`;
      newImg.classList.remove('hidden');
      oldImg.classList.add('hidden');
    }
  }
</script>

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
    padding: 1em 1.2em;
    text-align: center;
    z-index: 10;
    font-size: 1.3em;
    font-weight: 600;
    color: #e5e7eb;
    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.8);
    background: radial-gradient(
      ellipse 60% 100% at 50% 0%,
      rgba(0, 0, 0, 0.7) 0%,
      rgba(0, 0, 0, 0.4) 50%,
      rgba(0, 0, 0, 0) 100%
    );
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
    color: #d1d5db;
    font-size: 1em;
    line-height: 1.6;
    margin-bottom: 1em;
    flex-shrink: 0;
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
      <div class="roomName">{$store.roomName}</div>
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
      <div class="roomDescription">{$store.roomDescription}</div>
    {/if}

    <div class="entitySection">
      <EntityPanel {store} {sendMessage} />
    </div>
  </div>
</div>
