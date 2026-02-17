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
  $: race = typeof character?.race === 'object' ? (character.race.name || '') : (character?.race || '');
  $: charClass = typeof character?.class === 'object' ? (character.class.name || '') : (character?.class || '');
  $: level = stats.level || character?.level || 0;
  $: currentHp = stats.currentHitPoints || 0;
  $: maxHp = stats.maxHitPoints || 1;
  $: xp = stats.xp || 0;
  $: gold = stats.gold || 0;
  $: inCombat = stats.inCombat || false;
  $: attributes = stats.attributes || character?.attributes || [];
  $: currentMana = stats.currentMana || 0;
  $: maxMana = stats.maxMana || 0;
  $: hasMana = maxMana > 0;
  $: attackPower = stats.attackPower || 0;
  $: attackAttr = stats.attackAttr || 'STR';
  $: weaponDamage = stats.weaponDamage || 0;
  $: attackMod = stats.attackMod || 0;
  $: defense = stats.defense || 0;
  $: manaRegen = stats.manaRegen || 0;

  // HP percentage and color
  $: hpPercent = maxHp > 0 ? Math.min(100, (currentHp / maxHp) * 100) : 0;
  $: hpColor = hpPercent > 60 ? '#22c55e' : hpPercent > 30 ? '#f59e0b' : '#ef4444';
  $: hpGlow = hpPercent > 60 ? 'rgba(34,197,94,0.4)' : hpPercent > 30 ? 'rgba(245,158,11,0.4)' : 'rgba(239,68,68,0.4)';

  // Mana percentage
  $: manaPercent = maxMana > 0 ? Math.min(100, (currentMana / maxMana) * 100) : 0;

  // XP - use server-provided threshold, fallback to level * 1000
  $: xpToNext = stats.xpForNextLevel > 0 ? stats.xpForNextLevel : (level > 0 ? level * 1000 : 1000);
  $: xpPercent = xpToNext > 0 ? Math.min(100, (xp / xpToNext) * 100) : 0;

  // Attribute points
  $: unspentPoints = stats.unspentAttributePoints || 0;
  $: hasUnspentPoints = unspentPoints > 0;

  $: hasData = character !== null;

  function formatGold(value) {
    if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M';
    if (value >= 1000) return (value / 1000).toFixed(1) + 'K';
    return value.toLocaleString();
  }

  function getAttrShort(attr) {
    return attr.short || attr.name?.slice(0, 3).toUpperCase() || '???';
  }

  // Compute D&D-style modifier: (value - 10) / 2, rounded down
  function getModifier(value) {
    return Math.floor((value - 10) / 2);
  }

  function formatMod(mod) {
    return mod >= 0 ? '+' + mod : String(mod);
  }

  function isPrimaryAttr(attr) {
    return getAttrShort(attr) === attackAttr;
  }

  function getAttrTooltip(attr) {
    const short = getAttrShort(attr);
    const mod = getModifier(attr.value);
    let tip = `${attr.name || short}: ${attr.value} (modifier: ${formatMod(mod)})`;
    if (short === attackAttr) {
      tip += `\nPrimary attack attribute for ${charClass}`;
    }
    return tip;
  }

  // Build ATK formula tooltip
  $: atkFormula = weaponDamage === 1
    ? `ATK = Unarmed (1) ${formatMod(attackMod)} ${attackAttr} mod = ${attackPower}`
    : `ATK = Weapon (${weaponDamage}) ${formatMod(attackMod)} ${attackAttr} mod = ${attackPower}`;

  function spendPoint(attrShort) {
    if (sendMessage && hasUnspentPoints) {
      sendMessage('spend ' + attrShort);
    }
  }
</script>

<style>
  /* Base panel styling comes from global .game-panel class in themes.css */
  .character-widget {
    transition: border-color 0.3s ease, box-shadow 0.3s ease;
  }

  .character-widget.in-combat {
    border-color: rgba(239, 68, 68, 0.5) !important;
    box-shadow: 0 0 20px rgba(239, 68, 68, 0.15), inset 0 0 20px rgba(239, 68, 68, 0.05);
    animation: combatPulse 2s ease-in-out infinite;
  }

  @keyframes combatPulse {
    0%, 100% { border-color: rgba(239, 68, 68, 0.3); box-shadow: 0 0 15px rgba(239, 68, 68, 0.1); }
    50% { border-color: rgba(239, 68, 68, 0.6); box-shadow: 0 0 25px rgba(239, 68, 68, 0.2); }
  }

  /* Header uses global .game-panel-header */
  .widget-title {
    flex: 1;
  }

  .combat-badge {
    font-size: var(--text-xs);
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
    font-family: var(--font-display);
    font-size: var(--text-lg);
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 0.15em;
  }

  .char-info {
    font-size: var(--text-sm);
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: 0.4em;
  }

  .level-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--accent-subtle);
    color: var(--accent-primary);
    font-weight: 700;
    font-size: 0.85em;
    padding: 0.1em 0.45em;
    border-radius: 4px;
    border: 1px solid var(--panel-inner-border);
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
    font-size: var(--text-sm);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .bar-label.hp { color: #4ade80; }
  .bar-label.hp.danger { color: #f59e0b; }
  .bar-label.hp.critical { color: #ef4444; }
  .bar-label.mana { color: #60a5fa; }
  .bar-label.xp { color: #c084fc; }

  .bar-value {
    font-size: var(--text-sm);
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
  }

  .bar-track {
    height: var(--bar-height);
    background: var(--bar-track-bg);
    border: var(--bar-track-border);
    border-radius: var(--bar-radius);
    overflow: hidden;
    position: relative;
  }

  .bar-fill {
    height: 100%;
    border-radius: var(--bar-radius);
    transition: width 0.5s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
  }

  .bar-fill.hp {
    background: linear-gradient(90deg, var(--hp-color), var(--hp-color));
    box-shadow: 0 0 8px var(--hp-glow);
  }

  .bar-fill.mana {
    background: linear-gradient(90deg, #3b82f6, #60a5fa);
    box-shadow: 0 0 8px rgba(59, 130, 246, 0.3);
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
    border-radius: var(--bar-radius) var(--bar-radius) 0 0;
  }

  /* Gold row */
  .gold-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5em 0.6em;
    background: rgba(251, 191, 36, 0.08);
    border-radius: 6px;
    border: 1px solid rgba(251, 191, 36, 0.15);
    margin-bottom: 0.75em;
  }

  .gold-label {
    display: flex;
    align-items: center;
    gap: 0.35em;
    font-size: var(--text-sm);
    color: var(--text-primary);
  }

  .gold-label i {
    font-size: 1.1em;
    color: var(--color-gold);
  }

  .gold-value {
    font-size: var(--text-base);
    font-weight: 700;
    color: var(--color-gold);
    font-variant-numeric: tabular-nums;
  }

  /* Attributes grid */
  .attributes-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 0.4em;
  }

  .attr-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0.45em 0.35em;
    background: var(--panel-inner-bg);
    border-radius: 6px;
    border: 1px solid var(--panel-inner-border);
    transition: background 0.15s ease, border-color 0.15s ease;
    position: relative;
  }

  .attr-item:hover {
    background: var(--panel-inner-hover);
  }

  .attr-item.has-points {
    border-color: rgba(74, 222, 128, 0.2);
  }

  .attr-item.primary-attr {
    border-color: rgba(248, 113, 113, 0.3);
    background: rgba(248, 113, 113, 0.08);
  }

  .attr-value {
    font-size: var(--text-base);
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1.2;
  }

  .attr-mod {
    font-size: var(--text-xs);
    font-weight: 600;
    line-height: 1;
  }

  .attr-mod.positive { color: #4ade80; }
  .attr-mod.negative { color: #f87171; }

  .attr-name {
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-dim);
    margin-top: 0.1em;
  }

  /* Unspent points badge */
  .unspent-badge {
    font-size: var(--text-xs);
    font-weight: 700;
    color: #4ade80;
    background: rgba(74, 222, 128, 0.15);
    padding: 0.15em 0.5em;
    border-radius: 4px;
    border: 1px solid rgba(74, 222, 128, 0.3);
    animation: unspentPulse 2s ease-in-out infinite;
    white-space: nowrap;
  }

  @keyframes unspentPulse {
    0%, 100% { opacity: 0.8; }
    50% { opacity: 1; }
  }

  /* Attribute spend button */
  .attr-spend-btn {
    position: absolute;
    top: 2px;
    right: 2px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 1px solid rgba(74, 222, 128, 0.4);
    background: rgba(74, 222, 128, 0.15);
    color: #4ade80;
    font-size: var(--text-xs);
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
    padding: 0;
    line-height: 1;
  }

  .attr-spend-btn:hover {
    background: rgba(74, 222, 128, 0.3);
    border-color: rgba(74, 222, 128, 0.6);
    transform: scale(1.1);
  }

  /* Combat stats grid */
  .combat-stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.4em;
  }

  .combat-stat {
    display: flex;
    align-items: center;
    gap: 0.5em;
    padding: 0.45em 0.55em;
    background: var(--panel-inner-bg);
    border-radius: 6px;
    border: 1px solid var(--panel-inner-border);
  }

  .stat-icon {
    font-size: 1.1em;
    opacity: 0.8;
  }

  .stat-icon.atk { color: #f87171; }
  .stat-icon.def { color: #60a5fa; }
  .stat-icon.mpr { color: #a78bfa; }

  .combat-stat-info {
    display: flex;
    flex-direction: column;
  }

  .combat-stat-value {
    font-size: var(--text-base);
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1.2;
  }

  .combat-stat-label {
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-dim);
  }

  .combat-stat-formula {
    font-size: var(--text-xs);
    color: var(--text-secondary);
    font-style: italic;
    white-space: nowrap;
  }

  /* Empty state */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: calc(100% - 3em);
    color: var(--text-dim);
    text-align: center;
    gap: 0.5em;
  }

  .empty-state i {
    font-size: 2em;
    opacity: 0.4;
  }

  .empty-state span {
    font-size: var(--text-sm);
  }
</style>

<div class="character-widget game-panel" class:in-combat={inCombat}>
  <div class="game-panel-header">
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

      <!-- Mana Bar (only for caster classes) -->
      {#if hasMana}
        <div class="bar-container">
          <div class="bar-header">
            <span class="bar-label mana">MP</span>
            <span class="bar-value">{currentMana} / {maxMana}</span>
          </div>
          <div class="bar-track">
            <div
              class="bar-fill mana"
              style="width: {manaPercent}%"
            ></div>
          </div>
        </div>
      {/if}

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
      <div class="game-panel-divider">
        <span>Attributes</span>
        {#if hasUnspentPoints}
          <span class="unspent-badge">{unspentPoints} pts</span>
        {/if}
      </div>
      <div class="attributes-grid">
        {#each attributes as attr}
          <div class="attr-item" class:has-points={hasUnspentPoints} class:primary-attr={isPrimaryAttr(attr)} title={getAttrTooltip(attr)}>
            <span class="attr-value">{attr.value}</span>
            <span class="attr-mod" class:positive={getModifier(attr.value) >= 0} class:negative={getModifier(attr.value) < 0}>{formatMod(getModifier(attr.value))}</span>
            <span class="attr-name">{getAttrShort(attr)}</span>
            {#if hasUnspentPoints}
              <button class="attr-spend-btn" on:click={() => spendPoint(getAttrShort(attr))} title="Spend 1 point on {getAttrShort(attr)}">+</button>
            {/if}
          </div>
        {/each}
      </div>
    {/if}

    <!-- Derived Combat Stats -->
    <div class="game-panel-divider">
      <span>Combat Stats</span>
    </div>
    <div class="combat-stats-grid">
      <div class="combat-stat" title={atkFormula}>
        <i class="material-icons stat-icon atk">gavel</i>
        <div class="combat-stat-info">
          <span class="combat-stat-value">{attackPower}</span>
          <span class="combat-stat-label">ATK</span>
          <span class="combat-stat-formula">{weaponDamage === 1 ? 'Unarmed' : 'Wpn ' + weaponDamage} {formatMod(attackMod)} {attackAttr}</span>
        </div>
      </div>
      <div class="combat-stat" title="Total armor defense from equipped items">
        <i class="material-icons stat-icon def">shield</i>
        <div class="combat-stat-info">
          <span class="combat-stat-value">{defense}</span>
          <span class="combat-stat-label">DEF</span>
          <span class="combat-stat-formula">Armor</span>
        </div>
      </div>
      {#if hasMana}
        <div class="combat-stat" title="Mana regenerated per combat round">
          <i class="material-icons stat-icon mpr">auto_fix_high</i>
          <div class="combat-stat-info">
            <span class="combat-stat-value">{manaRegen}</span>
            <span class="combat-stat-label">MP/RND</span>
          </div>
        </div>
      {/if}
    </div>
  {:else}
    <div class="empty-state">
      <i class="material-icons">person_off</i>
      <span>No character selected</span>
    </div>
  {/if}
</div>
