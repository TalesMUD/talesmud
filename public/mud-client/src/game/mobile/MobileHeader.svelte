<style>
  .mobile-header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: calc(48px + env(safe-area-inset-top, 0px));
    padding-top: env(safe-area-inset-top, 0px);
    background: rgba(0, 0, 0, 0.9);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-left: 12px;
    padding-right: 12px;
    z-index: 1001;
    box-sizing: border-box;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
  }

  .room-icon {
    font-size: 18px;
    color: #f59e0b;
    flex-shrink: 0;
  }

  .room-name {
    font-family: 'Cinzel', serif;
    font-size: 14px;
    font-weight: 600;
    color: #f0e6d3;
    letter-spacing: 0.08em;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .hp-pill {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    border-radius: 10px;
    font-size: 11px;
    font-weight: 600;
    border: 1px solid;
  }

  .hp-pill.hp-high {
    background: rgba(34, 197, 94, 0.2);
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }

  .hp-pill.hp-mid {
    background: rgba(245, 158, 11, 0.2);
    border-color: rgba(245, 158, 11, 0.4);
    color: #fcd34d;
  }

  .hp-pill.hp-low {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.4);
    color: #fca5a5;
  }

  .hp-pill i {
    font-size: 13px;
  }

  .combat-indicator {
    animation: combatPulse 2s ease-in-out infinite;
    border-bottom: 2px solid rgba(239, 68, 68, 0.8);
  }

  @keyframes combatPulse {
    0%, 100% { border-bottom-color: rgba(239, 68, 68, 0.8); }
    50% { border-bottom-color: rgba(239, 68, 68, 0.3); }
  }
</style>

<script>
  export let store;

  $: roomName = $store.roomName || 'Unknown';
  $: stats = $store.characterStats || {};
  $: currentHP = stats.currentHitPoints || 0;
  $: maxHP = stats.maxHitPoints || 1;
  $: hpPercent = Math.round((currentHP / maxHP) * 100);
  $: hpClass = hpPercent > 60 ? 'hp-high' : hpPercent > 30 ? 'hp-mid' : 'hp-low';
  $: inCombat = $store.inCombat;
</script>

<div class="mobile-header" class:combat-indicator={inCombat}>
  <div class="header-left">
    <i class="material-icons room-icon">explore</i>
    <span class="room-name">{roomName}</span>
  </div>

  <div class="header-right">
    <div class="hp-pill {hpClass}">
      <i class="material-icons">favorite</i>
      {currentHP}/{maxHP}
    </div>
  </div>
</div>
