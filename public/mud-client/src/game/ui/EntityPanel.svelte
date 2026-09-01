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
    width: 96px;
    aspect-ratio: 2 / 3;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.18);
    animation: slideUp 0.3s ease-out;
    flex-shrink: 0;
    background: #111;
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
    top: 4px;
    left: 4px;
    z-index: 2;
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    max-width: calc(100% - 8px);
  }

  .state-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.15em;
    padding: 0.1em 0.35em;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.65);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #e5e7eb;
    font-size: 7px;
    font-weight: 700;
    text-transform: uppercase;
    line-height: 1.2;
  }

  .state-badge i {
    font-size: 10px;
  }

  .state-badge.enemy {
    color: #fca5a5;
    border-color: rgba(239, 68, 68, 0.4);
  }

  .state-badge.merchant {
    color: #86efac;
    border-color: rgba(34, 197, 94, 0.4);
  }

  .state-badge.quest {
    color: #fcd34d;
    border-color: rgba(245, 158, 11, 0.45);
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

  function getStateLabel(state) {
    if (!state) return 'Idle';
    return state.charAt(0).toUpperCase() + state.slice(1);
  }

  function toggleMenu(npc) {
    openMenuId = openMenuId === npc.id ? null : npc.id;
  }

  function getPrimaryAction(npc) {
    if (npc.hasDialog || npc.isQuestGiver) {
      return { label: 'Talk', kind: 'talk', fn: () => talk(npc) };
    }
    if (npc.isEnemy) {
      return { label: 'Attack', kind: 'attack', fn: () => attack(npc) };
    }
    if (npc.isMerchant) {
      return { label: 'Trade', kind: 'trade', fn: () => trade(npc) };
    }
    return null;
  }

  function getOverflowActions(npc) {
    const primary = getPrimaryAction(npc);
    const actions = [];

    if (npc.isEnemy) {
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
    return getOverflowActions(npc).length > 0;
  }
</script>

{#if $store.npcs && $store.npcs.length > 0}
  <div class="entity-panel">
    {#each $store.npcs as npc (npc.id)}
      {@const primary = getPrimaryAction(npc)}
      <div class="entity-card {getEntityType(npc)}">
        <img
          class="entity-bg"
          src={portraitSrc(npc)}
          alt=""
          on:error={(e) => onPortraitError(e, npc)}
        />

        {#if npc.isEnemy && npc.maxHp > 0}
          <div class="entity-health">
            <div
              class="health-fill {getHealthClass(npc.currentHp, npc.maxHp)}"
              style="width: {(npc.currentHp / npc.maxHp) * 100}%"
            ></div>
          </div>
        {/if}

        {#if npc.isEnemy || npc.isMerchant || npc.isQuestGiver}
          <div class="entity-badges">
            {#if npc.isEnemy}
              <span class="state-badge enemy"><i class="material-icons">swords</i></span>
            {/if}
            {#if npc.isMerchant}
              <span class="state-badge merchant"><i class="material-icons">store</i></span>
            {/if}
            {#if npc.isQuestGiver}
              <span class="state-badge quest"><i class="material-icons">assignment</i></span>
            {/if}
          </div>
        {/if}

        <div class="entity-footer">
          <span class="entity-name">{npc.displayName}</span>
          {#if npc.level > 0}
            <span class="entity-meta">Lv {npc.level}</span>
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
