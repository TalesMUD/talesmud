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

  .entity-card.friendly {
    border-left: 3px solid #3b82f6;
  }

  .entity-name {
    font-weight: 600;
    font-size: 13px;
    color: #e5e7eb;
    display: block;
    margin-bottom: 0.3em;
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
    padding: 0.3em 0.6em;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(255, 255, 255, 0.05);
    color: #e5e7eb;
    cursor: pointer;
    transition: all 0.15s ease;
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
    sendMessage(`talk ${npc.displayName}`);
  }

  function trade(npc) {
    sendMessage(`trade ${npc.displayName}`);
  }

  function getEntityType(npc) {
    if (npc.isEnemy) return 'enemy';
    if (npc.isMerchant) return 'merchant';
    return 'friendly';
  }

  function getEntityTypeLabel(npc) {
    if (npc.isEnemy) return 'Enemy';
    if (npc.isMerchant) return 'Merchant';
    return 'NPC';
  }
</script>

{#if $store.npcs && $store.npcs.length > 0}
  <div class="entity-panel">
    {#each $store.npcs as npc (npc.id)}
      <div class="entity-card {getEntityType(npc)}">
        <span class="entity-name">{npc.displayName}</span>
        <span class="entity-type {getEntityType(npc)}">{getEntityTypeLabel(npc)}</span>

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
            <button class="action-btn attack" on:click={() => attack(npc)}>Attack</button>
          {/if}
          {#if npc.isMerchant}
            <button class="action-btn trade" on:click={() => trade(npc)}>Trade</button>
          {/if}
          {#if !npc.isEnemy}
            <button class="action-btn talk" on:click={() => talk(npc)}>Talk</button>
          {/if}
        </div>
      </div>
    {/each}
  </div>
{/if}
