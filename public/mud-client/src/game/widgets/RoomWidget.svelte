<script>
  import EntityPanel from '../ui/EntityPanel.svelte';
  import DialogOverlay from '../ui/DialogOverlay.svelte';
  import ShopOverlay from '../ui/ShopOverlay.svelte';
  import RecipesOverlay from '../ui/RecipesOverlay.svelte';
  import RoomTextOverlay from '../ui/RoomTextOverlay.svelte';
  import PlayersOverlay from '../ui/PlayersOverlay.svelte';
  import { findNpcByName } from '../MUDXPlusStore';
  import { settingsStore } from '../SettingsStore.js';
  import { backend } from '../../api/base.js';

  export let store;
  export let sendMessage;

  let toggleImage = true;

  // Derive NPC for dialog overlay
  $: dialogNpc = $store.dialogActive
    ? findNpcByName($store.npcs, $store.dialogNpcName)
    : null;

  $: dialogNpcType = (() => {
    if (!$store.dialogActive) return 'npc';
    if (dialogNpc?.isEnemy) return 'enemy';
    if (dialogNpc?.isQuestGiver) return 'quest';
    if (dialogNpc?.isMerchant) return 'merchant';
    return 'npc';
  })();

  // Handle background changes — only when the asset id actually changes
  // (roomUpdate also refreshes NPCs and used to retrigger the crossfade).
  let appliedBackground = null;
  $: if ($store.background && $store.background !== appliedBackground) {
    appliedBackground = $store.background;
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
  /* Aspect-aware room chrome:
     - Art region grows to fill leftover vertical space
     - Description is a bottom-anchored content-sized block (no black void)
     - Fade band sits just above the description, not mid-painting
     Portrait / square / landscape via container size queries on .room-widget */
  .room-widget {
    display: flex;
    flex-direction: column;
    background: #000;
    border: 1px solid var(--panel-border, rgba(255, 255, 255, 0.1));
    border-radius: var(--panel-radius, 12px);
    overflow: hidden;
    height: 100%;
    min-height: 0;
    box-shadow: var(--panel-shadow, none);
    container-type: size;
    container-name: room-widget;
  }

  .roomImageSection {
    position: relative;
    flex: 1 1 auto;
    min-height: 140px;
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

  /* Narrow fade band hugging the description top edge */
  .roomImageGradient {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: clamp(56px, 18%, 110px);
    background-image: linear-gradient(
      to bottom,
      rgba(0, 0, 0, 0) 0%,
      rgba(0, 0, 0, 0.35) 45%,
      rgba(0, 0, 0, 0.85) 78%,
      rgba(0, 0, 0, 1) 100%
    );
    pointer-events: none;
    z-index: 4;
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

  /* Content-sized, flush to widget bottom — no flex-grow black void */
  .roomContentSection {
    flex: 0 0 auto;
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    padding: 0.55em 1em 0.75em;
    overflow-y: auto;
    background: #000;
    position: relative;
    z-index: 5;
    margin-top: 0;
    max-height: 32%;
    min-height: 0;
  }

  .roomDescription {
    color: #e5e7eb;
    font-size: 1.1em;
    line-height: 1.6;
    margin: 0;
    flex-shrink: 0;
    font-style: italic;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
    padding: 0.65em 0.9em;
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
    container-type: inline-size;
    container-name: room-entities;
  }

  /* Portrait: tall widget — maximize art, keep desc compact at bottom */
  @container room-widget (aspect-ratio < 0.85) {
    .roomImageSection {
      min-height: 55%;
    }

    .roomContentSection {
      max-height: 26%;
      padding: 0.45em 0.9em 0.65em;
    }

    .roomDescription {
      font-size: 1.02em;
      line-height: 1.5;
      padding: 0.55em 0.8em;
    }

    .roomImageGradient {
      height: clamp(48px, 14%, 96px);
    }

    .roomName {
      font-size: 1.25em;
      padding: 0.9em 1em;
    }
  }

  /* Square-ish: balanced art / desc */
  @container room-widget (aspect-ratio >= 0.85) and (aspect-ratio <= 1.25) {
    .roomContentSection {
      max-height: 34%;
    }

    .roomImageGradient {
      height: clamp(56px, 16%, 100px);
    }
  }

  /* Landscape: horizontal art priority, tighter description chrome */
  @container room-widget (aspect-ratio > 1.25) {
    .roomImageSection {
      min-height: 48%;
    }

    .roomContentSection {
      max-height: 38%;
      padding: 0.4em 1em 0.55em;
    }

    .roomDescription {
      font-size: 0.98em;
      line-height: 1.45;
      padding: 0.45em 0.75em;
    }

    .roomImageGradient {
      height: clamp(44px, 22%, 88px);
    }

    .roomName {
      font-size: 1.15em;
      padding: 0.75em 1em;
    }
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

    <PlayersOverlay {store} {sendMessage} />

    {#if $store.dialogActive}
      <DialogOverlay
        npcName={$store.dialogNpcName}
        npcText={$store.dialogNpcText}
        options={$store.dialogOptions}
        npcType={dialogNpcType}
        npc={dialogNpc}
        sendMessage={sendMessage}
      />
    {/if}

    {#if $store.shop}
      <ShopOverlay {store} {sendMessage} />
    {/if}

    {#if $store.recipes}
      <RecipesOverlay {store} {sendMessage} />
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
