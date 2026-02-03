<script>
  export let store = null;
  export let sendMessage = null;

  let character = null;
  let stats = {
    currentHitPoints: 0,
    maxHitPoints: 0,
    xp: 0,
    level: 0,
    gold: 0,
    inCombat: false,
    attributes: [],
  };

  // Subscribe to store for reactive updates
  $: if (store) {
    character = $store.character;
    stats = $store.characterStats || stats;
  }

  // Derived values
  $: name = character?.name || '—';
  $: race = character?.race || '';
  $: charClass = character?.class || '';
  $: level = stats.level || character?.level || 0;
  $: currentHp = stats.currentHitPoints || 0;
  $: maxHp = stats.maxHitPoints || 1;
  $: xp = stats.xp || 0;
  $: gold = stats.gold || 0;
  $: inCombat = stats.inCombat || false;
  $: attributes = stats.attributes || character?.attributes || [];

  // HP percentage and color
  $: hpPercent = maxHp > 0 ? Math.min(100, (currentHp / maxHp) * 100) : 0;
  $: hpColor = hpPercent > 60 ? '#22c55e' : hpPercent > 30 ? '#f59e0b' : '#ef4444';
  $: hpGlow = hpPercent > 60 ? 'rgba(34,197,94,0.4)' : hpPercent > 30 ? 'rgba(245,158,11,0.4)' : 'rgba(239,68,68,0.4)';

  // XP - estimate xpToNext as level * 1000 if not provided by server
  $: xpToNext = level > 0 ? level * 1000 : 1000;
  $: xpPercent = xpToNext > 0 ? Math.min(100, (xp / xpToNext) * 100) : 0;

  $: hasData = character !== null;

  function formatGold(value) {
    if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M';
    if (value >= 1000) return (value / 1000).toFixed(1) + 'K';
    return value.toLocaleString();
  }

  function getAttrShort(attr) {
    return attr.short || attr.name?.slice(0, 3).toUpperCase() || '???';
  }
</script>

<style>
  .character-widget {
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 1em;
    height: 100%;
    overflow-y: auto;
    color: #e5e7eb;
    transition: border-color 0.3s ease, box-shadow 0.3s ease;
  }

  .character-widget.in-combat {
    border-color: rgba(239, 68, 68, 0.5);
    box-shadow: 0 0 20px rgba(239, 68, 68, 0.15), inset 0 0 20px rgba(239, 68, 68, 0.05);
    animation: combatPulse 2s ease-in-out infinite;
  }

  @keyframes combatPulse {
    0%, 100% { border-color: rgba(239, 68, 68, 0.3); box-shadow: 0 0 15px rgba(239, 68, 68, 0.1); }
    50% { border-color: rgba(239, 68, 68, 0.6); box-shadow: 0 0 25px rgba(239, 68, 68, 0.2); }
  }

  .widget-header {
    display: flex;
    align-items: center;
    gap: 0.5em;
    margin-bottom: 1em;
    padding-bottom: 0.75em;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .widget-header i {
    color: #f59e0b;
  }

  .widget-title {
    font-size: 1em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    flex: 1;
  }

  .combat-badge {
    font-size: 0.65em;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #ef4444;
    background: rgba(239, 68, 68, 0.15);
    padding: 0.2em 0.6em;
    border-radius: 4px;
    border: 1px solid rgba(239, 68, 68, 0.3);
    animation: combatBadgePulse 1.5s ease-in-out infinite;
  }

  @keyframes combatBadgePulse {
    0%, 100% { opacity: 0.8; }
    50% { opacity: 1; }
  }

  /* Character identity */
  .char-identity {
    margin-bottom: 1em;
  }

  .char-name {
    font-size: 1.1em;
    font-weight: 700;
    color: #f3f4f6;
    margin-bottom: 0.15em;
  }

  .char-info {
    font-size: 0.8em;
    color: #9ca3af;
    display: flex;
    align-items: center;
    gap: 0.4em;
  }

  .level-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: rgba(245, 158, 11, 0.15);
    color: #f59e0b;
    font-weight: 700;
    font-size: 0.85em;
    padding: 0.1em 0.45em;
    border-radius: 4px;
    border: 1px solid rgba(245, 158, 11, 0.25);
  }

  /* Stat bars */
  .stat-bars {
    display: flex;
    flex-direction: column;
    gap: 0.6em;
    margin-bottom: 0.75em;
  }

  .bar-container {
    position: relative;
  }

  .bar-header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: 0.2em;
  }

  .bar-label {
    font-size: 0.75em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .bar-label.hp { color: #4ade80; }
  .bar-label.hp.danger { color: #f59e0b; }
  .bar-label.hp.critical { color: #ef4444; }
  .bar-label.xp { color: #c084fc; }

  .bar-value {
    font-size: 0.75em;
    color: #9ca3af;
    font-variant-numeric: tabular-nums;
  }

  .bar-track {
    height: 8px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 4px;
    overflow: hidden;
    position: relative;
  }

  .bar-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.5s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
  }

  .bar-fill.hp {
    background: linear-gradient(90deg, var(--hp-color), var(--hp-color));
    box-shadow: 0 0 8px var(--hp-glow);
  }

  .bar-fill.xp {
    background: linear-gradient(90deg, #a855f7, #c084fc);
    box-shadow: 0 0 8px rgba(168, 85, 247, 0.3);
  }

  .bar-fill::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 50%;
    background: linear-gradient(to bottom, rgba(255,255,255,0.15), transparent);
    border-radius: 4px 4px 0 0;
  }

  /* Gold row */
  .gold-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5em 0.6em;
    background: rgba(251, 191, 36, 0.08);
    border-radius: 6px;
    border: 1px solid rgba(251, 191, 36, 0.12);
    margin-bottom: 0.75em;
  }

  .gold-label {
    display: flex;
    align-items: center;
    gap: 0.35em;
    font-size: 0.8em;
    color: #d1d5db;
  }

  .gold-label i {
    font-size: 1.1em;
    color: #fbbf24;
  }

  .gold-value {
    font-size: 1em;
    font-weight: 700;
    color: #fbbf24;
    font-variant-numeric: tabular-nums;
  }

  /* Attributes section */
  .section-divider {
    display: flex;
    align-items: center;
    gap: 0.5em;
    margin: 0.75em 0 0.5em 0;
  }

  .section-divider span {
    font-size: 0.7em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #6b7280;
    white-space: nowrap;
  }

  .section-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: rgba(255, 255, 255, 0.08);
  }

  .attributes-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 0.35em;
  }

  .attr-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0.4em 0.3em;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.06);
    transition: background 0.15s ease;
  }

  .attr-item:hover {
    background: rgba(255, 255, 255, 0.08);
  }

  .attr-value {
    font-size: 1em;
    font-weight: 700;
    color: #f3f4f6;
    line-height: 1.2;
  }

  .attr-name {
    font-size: 0.6em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #6b7280;
    margin-top: 0.1em;
  }

  /* Empty state */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: calc(100% - 3em);
    color: #6b7280;
    text-align: center;
    gap: 0.5em;
  }

  .empty-state i {
    font-size: 2em;
    opacity: 0.4;
  }

  .empty-state span {
    font-size: 0.85em;
  }
</style>

<div class="character-widget" class:in-combat={inCombat}>
  <div class="widget-header">
    <i class="material-icons">{inCombat ? 'swords' : 'person'}</i>
    <span class="widget-title">Character</span>
    {#if inCombat}
      <span class="combat-badge">In Combat</span>
    {/if}
  </div>

  {#if hasData}
    <div class="char-identity">
      <div class="char-name">{name}</div>
      <div class="char-info">
        <span class="level-badge">Lv {level}</span>
        <span>{race} {charClass}</span>
      </div>
    </div>

    <div class="stat-bars">
      <!-- HP Bar -->
      <div class="bar-container">
        <div class="bar-header">
          <span class="bar-label hp" class:danger={hpPercent <= 60 && hpPercent > 30} class:critical={hpPercent <= 30}>HP</span>
          <span class="bar-value">{currentHp} / {maxHp}</span>
        </div>
        <div class="bar-track">
          <div
            class="bar-fill hp"
            style="width: {hpPercent}%; --hp-color: {hpColor}; --hp-glow: {hpGlow}"
          ></div>
        </div>
      </div>

      <!-- XP Bar -->
      <div class="bar-container">
        <div class="bar-header">
          <span class="bar-label xp">XP</span>
          <span class="bar-value">{xp} / {xpToNext}</span>
        </div>
        <div class="bar-track">
          <div
            class="bar-fill xp"
            style="width: {xpPercent}%"
          ></div>
        </div>
      </div>
    </div>

    <!-- Gold -->
    <div class="gold-row">
      <span class="gold-label">
        <i class="material-icons">paid</i>
        Gold
      </span>
      <span class="gold-value">{formatGold(gold)}</span>
    </div>

    <!-- Attributes -->
    {#if attributes && attributes.length > 0}
      <div class="section-divider">
        <span>Attributes</span>
      </div>
      <div class="attributes-grid">
        {#each attributes as attr}
          <div class="attr-item" title={attr.name || ''}>
            <span class="attr-value">{attr.value}</span>
            <span class="attr-name">{getAttrShort(attr)}</span>
          </div>
        {/each}
      </div>
    {/if}
  {:else}
    <div class="empty-state">
      <i class="material-icons">person_off</i>
      <span>No character selected</span>
    </div>
  {/if}
</div>
