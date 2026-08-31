<style>
  .entity-panel {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5em;
    justify-content: flex-start;
  }

  .entity-card {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 0.55em;
    background: rgba(0, 0, 0, 0.72);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    padding: 0.35em 0.55em 0.35em 0.35em;
    min-width: 0;
    max-width: 220px;
    transition: all 0.2s ease;
    animation: slideUp 0.3s ease-out;
  }

  .entity-card:hover {
    transform: translateY(-2px);
    border-color: rgba(255, 255, 255, 0.3);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
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

  .entity-portrait {
    width: 56px;
    height: 56px;
    border-radius: 6px;
    object-fit: contain;
    object-position: center bottom;
    image-rendering: pixelated;
    display: block;
    flex-shrink: 0;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255,255,255, 0.1);
  }

  .entity-body {
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .entity-name {
    font-weight: 600;
    font-size: 13px;
    color: #e5e7eb;
    display: block;
    margin-bottom: 0.15em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .badge-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25em;
    margin: 0.35em 0 0.45em;
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

  .state-badge.dialog {
    color: #93c5fd;
    border-color: rgba(59, 130, 246, 0.35);
    background: rgba(59, 130, 246, 0.12);
  }

  .state-badge.chatter {
    color: #c4b5fd;
    border-color: rgba(139, 92, 246, 0.35);
    background: rgba(139, 92, 246, 0.12);
  }

  .entity-type {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 0.4em;
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
  }

  .action-btn {
    font-size: 10px;
    padding: 0.3em 0.5em;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(255, 255, 255, 0.05);
    color: #e5e7eb;
    cursor: pointer;
    transition: all 0.15s ease;
    display: inline-flex;
    align-items: center;
    gap: 0.25em;
  }

  .action-btn i {
    font-size: 13px;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .action-btn.attack {
    border-color: rgba(239, 68, 68, 0.5);
    color: #fca5a5;
  }

  .action-btn.attack:hover {
    background: rgba(239, 68, 68, 0.2);
  }

  .action-btn.trade {
    border-color: rgba(34, 197, 94, 0.5);
    color: #86efac;
  }

  .action-btn.trade:hover {
    background: rgba(34, 197, 94, 0.2);
  }

  .action-btn.talk {
    border-color: rgba(59, 130, 246, 0.5);
    color: #93c5fd;
  }

  .action-btn.talk:hover {
    background: rgba(59, 130, 246, 0.2);
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

  /* Empty state */
  .empty-panel {
    display: none;
  }
</style>

<script>
  import { portraitSrc, onPortraitError } from '../portraitSrc.js';

  export let store;
  export let sendMessage;

  function getHealthClass(currentHp, maxHp) {
    if (maxHp === 0) return 'healthy';
    const ratio = currentHp / maxHp;
    if (ratio > 0.6) return 'healthy';
    if (ratio > 0.3) return 'wounded';
    return '';
  }

  function attack(npc) {
    sendMessage(`attack ${npc.displayName}`);
  }

  function talk(npc) {
    sendMessage(`speak to ${npc.displayName}`);
  }

  function trade(npc) {
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
</script>

{#if $store.npcs && $store.npcs.length > 0}
  <div class="entity-panel">
    {#each $store.npcs as npc (npc.id)}
      <div class="entity-card {getEntityType(npc)}">
        <img class="entity-portrait" src={portraitSrc(npc)} alt={npc.displayName} on:error={(e) => onPortraitError(e, npc)} />
        <div class="entity-body">
        <span class="entity-name">{npc.displayName}</span>
        <span class="entity-type {getEntityType(npc)}">{getEntityTypeLabel(npc)}</span>

        <div class="badge-row">
          {#if npc.isEnemy}
            <span class="state-badge enemy" title="Enemy">
              <i class="material-icons">swords</i> Enemy
            </span>
          {/if}
          {#if npc.isMerchant}
            <span class="state-badge merchant" title="Merchant">
              <i class="material-icons">store</i> Shop
            </span>
          {/if}
          {#if npc.isQuestGiver}
            <span class="state-badge quest" title="Quest giver">
              <i class="material-icons">assignment</i> Quest
            </span>
          {/if}
          {#if npc.hasDialog}
            <span class="state-badge dialog" title="Interactive dialog">
              <i class="material-icons">chat_bubble</i> Talk
            </span>
          {/if}
          {#if npc.state && npc.state !== 'idle'}
            <span class="state-badge" title="Current state">
              <i class="material-icons">{npc.state === 'patrol' ? 'route' : 'radio_button_checked'}</i>
              {getStateLabel(npc.state)}
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
          {#if npc.isEnemy}
            <button class="action-btn attack" on:click={() => attack(npc)} title="Attack {npc.displayName}">
              <i class="material-icons">swords</i> Attack
            </button>
          {/if}
          {#if npc.isMerchant}
            <button class="action-btn trade" on:click={() => trade(npc)} title="Trade with {npc.displayName}">
              <i class="material-icons">store</i> Trade
            </button>
          {/if}
          {#if !npc.isEnemy && (npc.hasDialog || npc.isQuestGiver)}
            <button class="action-btn talk" on:click={() => talk(npc)} title="Speak to {npc.displayName}">
              <i class="material-icons">chat</i> Speak
            </button>
          {/if}
        </div>
        </div>
      </div>
    {/each}
  </div>
{/if}
