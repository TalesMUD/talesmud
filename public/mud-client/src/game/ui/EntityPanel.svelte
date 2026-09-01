<style>
  .entity-panel {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5em;
    justify-content: flex-start;
  }

  .entity-card {
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    padding: 0.5em 0.75em;
    min-width: 100px;
    max-width: 150px;
    animation: slideUp 0.3s ease-out;
    position: relative;
  }

  .entity-card.enemy {
    border-left: 3px solid #ef4444;
  }

  .entity-card.merchant {
    border-left: 3px solid #22c55e;
  }

  .entity-card.quest {
    border-left-color: #f59e0b;
  }

  .entity-card.friendly {
    border-left: 3px solid #3b82f6;
  }

  .entity-portrait-wrap {
    width: 48px;
    height: 48px;
    overflow: hidden;
    border-radius: 6px;
    margin: 0 auto 0.45em;
    background: #000;
    border: 1px solid rgba(255, 255, 255, 0.12);
    flex-shrink: 0;
    position: relative;
  }

  .entity-portrait {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: top center;
    transform: scale(2.6);
    transform-origin: top center;
    image-rendering: pixelated;
    display: block;
  }

  .entity-name {
    font-weight: 600;
    font-size: 13px;
    color: #e5e7eb;
    display: block;
    margin-bottom: 0.3em;
    text-align: center;
  }

  .badge-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25em;
    justify-content: center;
    margin: 0.2em 0 0.45em;
  }

  .state-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.2em;
    min-height: 18px;
    padding: 0.15em 0.4em;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.07);
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: #cbd5e1;
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
  }

  .state-badge i {
    font-size: 12px;
  }

  .state-badge.enemy {
    color: #fca5a5;
    border-color: rgba(239, 68, 68, 0.35);
    background: rgba(239, 68, 68, 0.12);
  }

  .state-badge.merchant {
    color: #86efac;
    border-color: rgba(34, 197, 94, 0.35);
    background: rgba(34, 197, 94, 0.12);
  }

  .state-badge.quest {
    color: #fcd34d;
    border-color: rgba(245, 158, 11, 0.4);
    background: rgba(245, 158, 11, 0.14);
  }

  .entity-type {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 0.4em;
    text-align: center;
    display: block;
  }

  .entity-type.enemy {
    color: #ef4444;
  }

  .entity-type.merchant {
    color: #22c55e;
  }

  .entity-type.quest {
    color: #f59e0b;
  }

  .entity-type.friendly {
    color: #3b82f6;
  }

  .entity-level {
    font-size: 10px;
    color: #9ca3af;
    margin-bottom: 0.4em;
    text-align: center;
    display: block;
  }

  .health-bar {
    width: 100%;
    height: 4px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
    overflow: hidden;
    margin-bottom: 0.5em;
  }

  .health-fill {
    height: 100%;
    background: linear-gradient(90deg, #ef4444, #f87171);
    border-radius: 2px;
    transition: width 0.3s ease;
  }

  .health-fill.healthy {
    background: linear-gradient(90deg, #22c55e, #4ade80);
  }

  .health-fill.wounded {
    background: linear-gradient(90deg, #f59e0b, #fbbf24);
  }

  .entity-actions {
    display: flex;
    gap: 0.3em;
    flex-wrap: wrap;
    justify-content: center;
  }

  .action-btn {
    font-size: 10px;
    padding: 0.35em 0.55em;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(255, 255, 255, 0.08);
    color: #e5e7eb;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 0.25em;
    min-height: 28px;
  }

  .action-btn i {
    font-size: 13px;
  }

  .action-btn.primary.attack {
    border-color: rgba(239, 68, 68, 0.5);
    color: #fca5a5;
    background: rgba(239, 68, 68, 0.15);
  }

  .action-btn.primary.trade {
    border-color: rgba(34, 197, 94, 0.5);
    color: #86efac;
    background: rgba(34, 197, 94, 0.15);
  }

  .action-btn.primary.talk {
    border-color: rgba(59, 130, 246, 0.5);
    color: #93c5fd;
    background: rgba(59, 130, 246, 0.15);
  }

  .action-btn.menu-btn {
    padding: 0.35em;
    min-width: 28px;
    justify-content: center;
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

  function getEntityTypeLabel(npc) {
    if (npc.isEnemy) return 'Enemy';
    if (npc.isMerchant) return 'Merchant';
    if (npc.isQuestGiver) return 'Quest';
    return 'NPC';
  }

  function getStateLabel(state) {
    if (!state) return 'Idle';
    return state.charAt(0).toUpperCase() + state.slice(1);
  }

  function toggleMenu(npc) {
    openMenuId = openMenuId === npc.id ? null : npc.id;
  }

  function getPrimaryAction(npc) {
    if (npc.isEnemy) {
      return { label: 'Attack', kind: 'attack', fn: () => attack(npc) };
    }
    if (npc.hasDialog || npc.isQuestGiver) {
      return { label: 'Talk', kind: 'talk', fn: () => talk(npc) };
    }
    if (npc.isMerchant) {
      return { label: 'Trade', kind: 'trade', fn: () => trade(npc) };
    }
    return null;
  }

  function getOverflowActions(npc) {
    const primary = getPrimaryAction(npc);
    const actions = [];

    if (npc.isEnemy && primary?.label !== 'Attack') {
      actions.push({ label: 'Attack', fn: () => attack(npc) });
    }
    if (npc.isMerchant && primary?.label !== 'Trade') {
      actions.push({ label: 'Trade', fn: () => trade(npc) });
    }
    if ((npc.hasDialog || npc.isQuestGiver) && primary?.label !== 'Talk') {
      actions.push({ label: 'Talk', fn: () => talk(npc) });
    }

    actions.push({
      label: `State: ${getStateLabel(npc.state)}`,
      fn: null,
      muted: true,
    });

    if (npc.hasIdleDialog) {
      actions.push({
        label: 'Idle chatter',
        fn: null,
        muted: true,
      });
    }

    return actions;
  }

  function hasOverflow(npc) {
    return getOverflowActions(npc).some((a) => !a.muted || a.fn);
  }
</script>

{#if $store.npcs && $store.npcs.length > 0}
  <div class="entity-panel">
    {#each $store.npcs as npc (npc.id)}
      {@const primary = getPrimaryAction(npc)}
      <div class="entity-card {getEntityType(npc)}">
        <div class="entity-portrait-wrap">
          <img
            class="entity-portrait"
            src={portraitSrc(npc)}
            alt=""
            width="48"
            height="48"
            on:error={(e) => onPortraitError(e, npc)}
          />
        </div>
        <span class="entity-name">{npc.displayName}</span>

        <div class="badge-row">
          {#if npc.isEnemy}
            <span class="state-badge enemy">
              <i class="material-icons">swords</i> Enemy
            </span>
          {/if}
          {#if npc.isMerchant}
            <span class="state-badge merchant">
              <i class="material-icons">store</i> Shop
            </span>
          {/if}
          {#if npc.isQuestGiver}
            <span class="state-badge quest">
              <i class="material-icons">assignment</i> Quest
            </span>
          {/if}
        </div>

        {#if npc.level > 0}
          <span class="entity-level">Level {npc.level}</span>
        {/if}

        {#if npc.isEnemy && npc.maxHp > 0}
          <div class="health-bar">
            <div
              class="health-fill {getHealthClass(npc.currentHp, npc.maxHp)}"
              style="width: {(npc.currentHp / npc.maxHp) * 100}%"
            ></div>
          </div>
        {/if}

        <div class="entity-actions">
          {#if primary}
            <button
              class="action-btn primary {primary.kind}"
              on:click={primary.fn}
              title={primary.label}
            >
              <i class="material-icons">
                {primary.kind === 'attack' ? 'swords' : primary.kind === 'trade' ? 'store' : 'chat'}
              </i>
              {primary.label}
            </button>
          {/if}
          {#if hasOverflow(npc)}
            <button
              class="action-btn menu-btn"
              on:click={() => toggleMenu(npc)}
              aria-label="More actions"
              title="More actions"
            >
              <i class="material-icons">more_horiz</i>
            </button>
          {/if}
        </div>

        {#if openMenuId === npc.id}
          <div class="overflow-menu">
            {#each getOverflowActions(npc) as action}
              {#if action.muted}
                <span class="overflow-item muted">{action.label}</span>
              {:else}
                <button class="overflow-item" on:click={action.fn}>{action.label}</button>
              {/if}
            {/each}
          </div>
        {/if}
      </div>
    {/each}
  </div>
{/if}
