<style>
  .entity-panel {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5em;
    justify-content: flex-start;
    align-items: flex-end;
  }

  .entity-card {
    position: relative;
    width: clamp(88px, 22cqw, 128px);
    aspect-ratio: 2 / 3;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.18);
    animation: slideUp 0.3s ease-out;
    flex-shrink: 0;
    background: #111;
  }

  .entity-card.clickable {
    cursor: pointer;
  }

  .entity-card.clickable:focus-visible {
    outline: 2px solid rgba(59, 130, 246, 0.65);
    outline-offset: 2px;
  }

  .entity-card.enemy {
    border-color: rgba(239, 68, 68, 0.45);
    box-shadow: 0 0 0 1px rgba(239, 68, 68, 0.2);
  }

  .entity-card.merchant {
    border-color: rgba(34, 197, 94, 0.45);
    box-shadow: 0 0 0 1px rgba(34, 197, 94, 0.2);
  }

  .entity-card.quest {
    border-color: rgba(245, 158, 11, 0.45);
    box-shadow: 0 0 0 1px rgba(245, 158, 11, 0.2);
  }

  .entity-card.friendly {
    border-color: rgba(59, 130, 246, 0.4);
  }

  .entity-bg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: top center;
    image-rendering: pixelated;
    display: block;
    z-index: 0;
    /* ~10% padding each side — scale after fit so view-box still uses full card */
    transform: scale(0.8);
    transform-origin: center center;
  }

  /* Wide sprite content (animals): show whole figure, letterbox */
  .entity-bg.fit-wide {
    object-fit: contain;
    object-position: center center;
  }

  /* Tall/narrow sprite content (standing NPCs): fill 2:3, keep head */
  .entity-bg.fit-tall {
    object-fit: cover;
    object-position: top center;
  }

  .entity-health {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: rgba(0, 0, 0, 0.45);
    z-index: 3;
  }

  .health-fill {
    height: 100%;
    background: linear-gradient(90deg, #ef4444, #f87171);
    transition: width 0.3s ease;
  }

  .health-fill.healthy {
    background: linear-gradient(90deg, #22c55e, #4ade80);
  }

  .health-fill.wounded {
    background: linear-gradient(90deg, #f59e0b, #fbbf24);
  }

  .entity-badges {
    position: absolute;
    top: 5px;
    left: 5px;
    z-index: 2;
    display: flex;
    flex-wrap: wrap;
    gap: 3px;
    max-width: calc(100% - 10px);
  }

  .state-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    padding: 0;
    border-radius: 4px;
    background: rgba(0, 0, 0, 0.72);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #e5e7eb;
    line-height: 1;
    flex-shrink: 0;
  }

  .state-badge i {
    font-size: 15px;
  }

  .state-badge.enemy {
    color: #fca5a5;
    border-color: rgba(239, 68, 68, 0.45);
  }

  .state-badge.merchant {
    color: #86efac;
    border-color: rgba(34, 197, 94, 0.45);
  }

  .state-badge.quest {
    color: #fcd34d;
    border-color: rgba(245, 158, 11, 0.5);
  }

  .state-badge.dialog {
    color: #93c5fd;
    border-color: rgba(59, 130, 246, 0.45);
  }

  .entity-footer {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 2;
    padding: 0.45em 0.35em 0.4em;
    display: flex;
    flex-direction: column;
    gap: 0.3em;
    pointer-events: none;
  }

  .entity-footer::before {
    content: '';
    position: absolute;
    inset: -12px 0 0;
    background: linear-gradient(
      to top,
      rgba(0, 0, 0, 0.94) 0%,
      rgba(0, 0, 0, 0.72) 45%,
      rgba(0, 0, 0, 0.2) 75%,
      transparent 100%
    );
    z-index: -1;
    pointer-events: none;
  }

  .entity-name {
    font-weight: 700;
    font-size: 11px;
    color: #f9fafb;
    line-height: 1.15;
    text-align: center;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.9);
    word-break: break-word;
  }

  .entity-meta {
    font-size: 8px;
    color: rgba(255, 255, 255, 0.75);
    text-align: center;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
  }

  .entity-actions {
    display: flex;
    gap: 0.2em;
    justify-content: center;
    pointer-events: auto;
  }

  .action-btn {
    font-size: 9px;
    padding: 0.28em 0.4em;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.22);
    background: rgba(0, 0, 0, 0.55);
    color: #f3f4f6;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.2em;
    min-height: 24px;
    flex: 1;
    backdrop-filter: blur(4px);
  }

  .action-btn i {
    font-size: 12px;
  }

  .action-btn.primary.attack {
    border-color: rgba(239, 68, 68, 0.55);
    color: #fecaca;
    background: rgba(127, 29, 29, 0.65);
  }

  .action-btn.primary.trade {
    border-color: rgba(34, 197, 94, 0.55);
    color: #bbf7d0;
    background: rgba(20, 83, 45, 0.65);
  }

  .action-btn.primary.talk {
    border-color: rgba(59, 130, 246, 0.55);
    color: #bfdbfe;
    background: rgba(30, 58, 138, 0.65);
  }

  .action-btn.menu-btn {
    flex: 0 0 24px;
    padding: 0.2em;
    min-width: 24px;
  }

  .overflow-menu {
    position: absolute;
    left: 0;
    right: 0;
    bottom: calc(100% + 4px);
    background: rgba(9, 10, 12, 0.96);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 6px;
    padding: 0.25em;
    z-index: 20;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
    pointer-events: auto;
  }

  .overflow-item {
    display: block;
    width: 100%;
    text-align: left;
    font-size: 10px;
    padding: 0.4em 0.5em;
    border: 0;
    border-radius: 4px;
    background: transparent;
    color: #e5e7eb;
    cursor: pointer;
  }

  .overflow-item:hover,
  .overflow-item:focus {
    background: rgba(255, 255, 255, 0.1);
  }

  .overflow-item.muted {
    color: #9ca3af;
    cursor: default;
    font-style: italic;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>

<script>
  import { portraitSrc, onPortraitError } from '../portraitSrc.js';

  export let store;
  export let sendMessage;

  let openMenuId = null;

  function getHealthClass(currentHp, maxHp) {
    if (maxHp === 0) return 'healthy';
    const ratio = currentHp / maxHp;
    if (ratio > 0.6) return 'healthy';
    if (ratio > 0.3) return 'wounded';
    return '';
  }

  function attack(npc) {
    openMenuId = null;
    sendMessage(`attack ${npc.displayName}`);
  }

  function talk(npc) {
    openMenuId = null;
    sendMessage(`speak to ${npc.displayName}`);
  }

  function trade(npc) {
    openMenuId = null;
    sendMessage(`list`);
  }

  function getEntityType(npc) {
    if (npc.isEnemy) return 'enemy';
    if (npc.isQuestGiver) return 'quest';
    if (npc.isMerchant) return 'merchant';
    return 'friendly';
  }

  function toggleMenu(npc) {
    openMenuId = openMenuId === npc.id ? null : npc.id;
  }

  function canTalk(npc) {
    return npc.hasDialog || npc.isQuestGiver;
  }

  function canAttack(npc) {
    return npc.isEnemy;
  }

  function canTrade(npc) {
    return npc.isMerchant;
  }

  function getPrimaryAction(npc) {
    if (canTalk(npc)) {
      return { label: 'Talk', kind: 'talk', fn: () => talk(npc) };
    }
    if (canAttack(npc)) {
      return { label: 'Attack', kind: 'attack', fn: () => attack(npc) };
    }
    if (canTrade(npc)) {
      return { label: 'Trade', kind: 'trade', fn: () => trade(npc) };
    }
    return null;
  }

  function getExtraVerbs(npc) {
    const primary = getPrimaryAction(npc);
    const verbs = [];

    if (canAttack(npc) && primary?.label !== 'Attack') {
      verbs.push({ label: 'Attack', kind: 'attack', fn: () => attack(npc) });
    }
    if (canTrade(npc) && primary?.label !== 'Trade') {
      verbs.push({ label: 'Trade', kind: 'trade', fn: () => trade(npc) });
    }
    if (canTalk(npc) && primary?.label !== 'Talk') {
      verbs.push({ label: 'Talk', kind: 'talk', fn: () => talk(npc) });
    }

    return verbs;
  }

  function hasExtraVerbs(npc) {
    return getExtraVerbs(npc).length > 0;
  }

  function handleCardClick(npc, primary) {
    if (primary) {
      primary.fn();
    }
  }

  function handleCardKeydown(event, npc, primary) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleCardClick(npc, primary);
    }
  }

  // Card chrome is 2:3. Sprites are often square canvases with tall or wide
  // content — crop to opaque bbox, then contain (wide) or cover-top (tall).
  const CARD_ASPECT = 2 / 3;
  const fitCache = new Map();

  function opaqueBBox(img) {
    const w = img.naturalWidth;
    const h = img.naturalHeight;
    if (!w || !h) return null;
    try {
      const canvas = document.createElement('canvas');
      canvas.width = w;
      canvas.height = h;
      const ctx = canvas.getContext('2d', { willReadFrequently: true });
      if (!ctx) return null;
      ctx.drawImage(img, 0, 0);
      const data = ctx.getImageData(0, 0, w, h).data;
      let minX = w, minY = h, maxX = 0, maxY = 0;
      let found = false;
      for (let y = 0; y < h; y++) {
        for (let x = 0; x < w; x++) {
          if (data[(y * w + x) * 4 + 3] > 12) {
            found = true;
            if (x < minX) minX = x;
            if (y < minY) minY = y;
            if (x > maxX) maxX = x;
            if (y > maxY) maxY = y;
          }
        }
      }
      if (!found) return null;
      // Small pad so pixel edges aren't flush against the crop
      const pad = 2;
      minX = Math.max(0, minX - pad);
      minY = Math.max(0, minY - pad);
      maxX = Math.min(w - 1, maxX + pad);
      maxY = Math.min(h - 1, maxY + pad);
      return { x: minX, y: minY, w: maxX - minX + 1, h: maxY - minY + 1 };
    } catch (_) {
      return null;
    }
  }

  function fitModeForBox(box, img) {
    if (box && box.h > 0) {
      return box.w / box.h > CARD_ASPECT ? 'fit-wide' : 'fit-tall';
    }
    if (img.naturalWidth && img.naturalHeight) {
      return img.naturalWidth / img.naturalHeight > CARD_ASPECT ? 'fit-wide' : 'fit-tall';
    }
    return 'fit-tall';
  }

  function applyCardFit(ev) {
    const img = ev && ev.currentTarget;
    if (!img) return;
    const src = img.currentSrc || img.src;
    let cached = fitCache.get(src);
    if (!cached) {
      const box = opaqueBBox(img);
      cached = { mode: fitModeForBox(box, img), box };
      fitCache.set(src, cached);
    }
    img.classList.remove('fit-wide', 'fit-tall');
    img.classList.add(cached.mode);
    if (cached.box) {
      const { x, y, w, h } = cached.box;
      img.style.objectViewBox = `xywh(${x}px ${y}px ${w}px ${h}px)`;
    } else {
      img.style.objectViewBox = '';
    }
  }

  function handlePortraitError(ev, npc) {
    onPortraitError(ev, npc);
    const img = ev && ev.currentTarget;
    if (!img) return;
    img.classList.remove('fit-wide', 'fit-tall');
    img.style.objectViewBox = '';
    // Avatar fallbacks are square faces — contain looks fine
    img.classList.add('fit-wide');
  }
</script>

{#if $store.npcs && $store.npcs.length > 0}
  <div class="entity-panel">
    {#each $store.npcs as npc (npc.id)}
      {@const primary = getPrimaryAction(npc)}
      {@const extraVerbs = hasExtraVerbs(npc)}
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div
        class="entity-card {getEntityType(npc)}"
        class:clickable={primary}
        on:click={() => handleCardClick(npc, primary)}
        on:keydown={(e) => handleCardKeydown(e, npc, primary)}
        role={primary ? 'button' : undefined}
        tabindex={primary ? 0 : undefined}
      >
        <img
          class="entity-bg"
          src={portraitSrc(npc)}
          alt=""
          on:load={applyCardFit}
          on:error={(e) => handlePortraitError(e, npc)}
        />

        {#if npc.isEnemy && npc.maxHp > 0}
          <div class="entity-health">
            <div
              class="health-fill {getHealthClass(npc.currentHp, npc.maxHp)}"
              style="width: {(npc.currentHp / npc.maxHp) * 100}%"
            ></div>
          </div>
        {/if}

        {#if npc.isEnemy || npc.isMerchant || npc.isQuestGiver || npc.hasDialog}
          <div class="entity-badges">
            {#if npc.isEnemy}
              <span class="state-badge enemy" title="Enemy"><i class="material-icons">swords</i></span>
            {/if}
            {#if npc.hasDialog}
              <span class="state-badge dialog" title="Dialog"><i class="material-icons">chat_bubble</i></span>
            {/if}
            {#if npc.isMerchant}
              <span class="state-badge merchant" title="Merchant"><i class="material-icons">store</i></span>
            {/if}
            {#if npc.isQuestGiver}
              <span class="state-badge quest" title="Quest"><i class="material-icons">assignment</i></span>
            {/if}
          </div>
        {/if}

        <div class="entity-footer">
          <span class="entity-name">{npc.displayName}</span>
          {#if npc.level > 0}
            <span class="entity-meta">Lv {npc.level}</span>
          {/if}
          {#if extraVerbs}
            <div class="entity-actions">
              <button
                class="action-btn menu-btn"
                on:click|stopPropagation={() => toggleMenu(npc)}
                aria-label="More actions"
                title="More actions"
              >
                <i class="material-icons">more_horiz</i>
              </button>
            </div>
          {/if}
        </div>

        {#if openMenuId === npc.id}
          <div class="overflow-menu">
            {#each getExtraVerbs(npc) as action}
              <button class="overflow-item" on:click|stopPropagation={action.fn}>{action.label}</button>
            {/each}
          </div>
        {/if}
      </div>
    {/each}
  </div>
{/if}
