<script>
  import EntityPanel from '../ui/EntityPanel.svelte';
  import DialogOverlay from '../ui/DialogOverlay.svelte';
  import ShopOverlay from '../ui/ShopOverlay.svelte';
  import RoomTextOverlay from '../ui/RoomTextOverlay.svelte';
  import PlayersOverlay from '../ui/PlayersOverlay.svelte';
  import HotbarWidget from '../widgets/HotbarWidget.svelte';
  import MobileActionBar from './MobileActionBar.svelte';
  import { findNpcByName } from '../MUDXPlusStore';
  import { settingsStore } from '../SettingsStore.js';
  import { backend } from '../../api/base.js';

  export let store;
  export let sendMessage;

  let toggleImage = true;
  let img1El;
  let img2El;
  let descriptionExpanded = false;
  let showPickupMenu = false;

  $: groundItems = ($store.groundItems || []).filter(item => !item.noPickup);

  function pickupItem(item) {
    sendMessage('pickup ' + item.name);
    store.removeGroundItem(item.id);
    showPickupMenu = false;
  }

  // Derive NPC type for dialog overlay
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

  // Handle background changes
  $: if ($store.background) {
    updateRoomImage($store.background);
  }

  function updateRoomImage(background) {
    const oldImg = toggleImage ? img1El : img2El;
    const newImg = !toggleImage ? img1El : img2El;

    toggleImage = !toggleImage;

    if (newImg && oldImg) {
      const bgUrl = `${backend}/backgrounds/${background}.png`;
      const placeholderUrl = 'img/placeholder.png';

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

  function toggleDescription() {
    descriptionExpanded = !descriptionExpanded;
  }
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&display=swap" rel="stylesheet">
</svelte:head>

<style>
  .mobile-room {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  /* Hero image section */
  .room-image-section {
    position: relative;
    width: 100%;
    height: 40vh;
    min-height: 180px;
    max-height: 360px;
    overflow: hidden;
    flex-shrink: 0;
  }

  .room-image-inner {
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

  .room-image-inner.hidden {
    opacity: 0;
  }

  .room-image-gradient {
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
    z-index: 5;
  }

  .entity-section {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 10;
    padding: 0.3em 0.5em;
    container-type: inline-size;
    container-name: room-entities;
  }

  /* Portrait entity cards in room overlay */
  .entity-section :global(.entity-panel) {
    gap: 0.35em;
  }

  .entity-section :global(.entity-card) {
    width: clamp(72px, 20cqw, 104px);
    border-radius: 6px;
  }

  .entity-section :global(.entity-name) {
    font-size: 10px;
  }

  .entity-section :global(.entity-meta) {
    font-size: 7px;
  }

  .entity-section :global(.entity-health) {
    height: 2px;
  }

  .entity-section :global(.action-btn) {
    font-size: 8px;
    padding: 0.2em 0.35em;
    min-height: 22px;
  }

  .entity-section :global(.action-btn.menu-btn) {
    flex: 0 0 22px;
    min-width: 22px;
  }

  /* Description section ? size to content; do not grow into a black void.
     Android Chrome collapses -webkit-box to 0 height inside flex:1 + min-height:0. */
  .room-description-section {
    padding: 10px 14px;
    background: #000;
    flex: 0 0 auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
  }

  .room-description {
    color: #e5e7eb;
    font-size: 0.95em;
    line-height: 1.6;
    font-style: italic;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
    padding: 8px 10px;
    border-left: 2px solid rgba(168, 130, 90, 0.6);
    border-radius: 0 4px 4px 0;
    background: linear-gradient(90deg, rgba(168, 130, 90, 0.08) 0%, transparent 100%);
    cursor: pointer;
  }

  .room-description.clamped {
    /* Prefer max-height clamp; keep -webkit-line-clamp as progressive enhancement */
    display: block;
    max-height: calc(1.6em * 3 + 16px);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    /* Floor so a flex parent cannot squash the box to 0 on mobile WebKit */
    min-height: calc(1.6em * 3 + 16px);
  }

  .room-description.parchment {
    background:
      url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E"),
      linear-gradient(90deg, rgba(139, 105, 65, 0.15) 0%, rgba(139, 105, 65, 0.05) 50%, transparent 100%);
    background-blend-mode: overlay, normal;
    box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.3);
  }

  .expand-hint {
    text-align: center;
    padding: 4px 0 0;
    color: #6b7280;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  /* Action bar area */
  .action-section {
    background: rgba(0, 0, 0, 0.6);
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    flex-shrink: 0;
    margin-top: auto;
  }

  .room-pickup-fab {
    position: absolute;
    right: 8px;
    bottom: 10px;
    z-index: 12;
    width: 44px;
    height: 44px;
    min-width: 44px;
    padding: 0;
    border-radius: 10px;
    border: 1px solid rgba(34, 197, 94, 0.5);
    background: rgba(20, 83, 45, 0.88);
    color: #86efac;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .room-pickup-fab i {
    font-size: 20px;
  }

  .room-pickup-count {
    position: absolute;
    top: -4px;
    right: -4px;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 8px;
    background: #22c55e;
    color: #052e16;
    font-size: 10px;
    font-weight: 800;
    line-height: 16px;
    text-align: center;
  }

  .pickup-dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: 9999;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .pickup-dialog {
    background: rgba(15, 15, 25, 0.95);
    border: 1px solid rgba(34, 197, 94, 0.4);
    border-radius: 14px;
    padding: 16px;
    width: 90vw;
    max-width: 400px;
    max-height: 60vh;
    overflow-y: auto;
  }

  .pickup-dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .pickup-dialog-title {
    color: #86efac;
    font-size: 13px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 1px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .pickup-dialog-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .pickup-dialog-grid button {
    background: rgba(34, 197, 94, 0.2);
    border: 1px solid rgba(34, 197, 94, 0.4);
    border-radius: 8px;
    padding: 10px 8px;
    color: #86efac;
    min-height: 44px;
    text-transform: uppercase;
    font-size: 12px;
  }
</style>

<div class="mobile-room">
  <!-- Hero room image -->
  <div class="room-image-section">
    <div class="room-image-inner" bind:this={img1El}></div>
    <div class="room-image-inner hidden" bind:this={img2El}></div>
    <div class="room-image-gradient"></div>

    <!-- Room name is already shown in MobileHeader -->

    <RoomTextOverlay />
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

    <div class="entity-section">
      <EntityPanel {store} {sendMessage} />
    </div>

    {#if groundItems.length > 0}
      <button
        class="room-pickup-fab"
        type="button"
        aria-label="Pick up {groundItems.length} items"
        on:click={() => showPickupMenu = true}
      >
        <i class="material-icons">back_hand</i>
        <span class="room-pickup-count">{groundItems.length}</span>
      </button>
    {/if}
  </div>

  {#if showPickupMenu && groundItems.length > 0}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="pickup-dialog-overlay" on:click={() => showPickupMenu = false}>
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="pickup-dialog" on:click|stopPropagation>
        <div class="pickup-dialog-header">
          <span class="pickup-dialog-title">
            <i class="material-icons">back_hand</i>
            Pick Up
          </span>
          <button type="button" on:click={() => showPickupMenu = false} aria-label="Close">
            <i class="material-icons">close</i>
          </button>
        </div>
        <div class="pickup-dialog-grid">
          {#each groundItems as item}
            <button type="button" on:click={() => pickupItem(item)}>{item.name}</button>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <!-- Room description (tap to expand) -->
  {#if $store.roomDescription}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="room-description-section" on:click={toggleDescription}>
      <div
        class="room-description"
        class:clamped={!descriptionExpanded}
        class:parchment={$settingsStore.interface?.parchmentBackground}
      >
        {$store.roomDescription}
      </div>
      {#if !descriptionExpanded}
        <div class="expand-hint">tap to expand</div>
      {/if}
    </div>
  {/if}

  <!-- Spell / consumable hotbar above dirs + LOOK/INV/MAP -->
  <HotbarWidget {store} {sendMessage} compact={true} />

  <!-- Action bar -->
  <div class="action-section">
    <MobileActionBar {store} {sendMessage} />
  </div>
</div>
