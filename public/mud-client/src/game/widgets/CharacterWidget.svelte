<script>
  import { settingsStore } from '../SettingsStore.js';
  import { overlayStore } from '../ui/overlayStore.js';
  import { onItemArtError } from '../itemArtSrc.js';
  import {
    bindSkillToFirstEmptyHotbar,
    characterClassId,
    classifySkills,
    formatSkillCost,
    formatSkillEffects,
    maxSkillSlots,
    normalizeHotbarBinds,
    skillGenericArtUrl,
  } from '../hudPrefs.js';

  export let store = null;
  export let sendMessage = null;

  let sheetTab = 'stats';

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
  $: resting = !!(stats.resting && !inCombat);
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

  $: classId = characterClassId(character);
  $: equippedSkills = stats.equippedSkills || character?.equippedSkills || [];
  $: computedMaxSlots = Number(stats.maxSkillSlots) > 0
    ? Number(stats.maxSkillSlots)
    : maxSkillSlots(classId, level);
  $: classified = classifySkills(classId, level, equippedSkills);
  $: slotPlaceholders = Array.from({ length: Math.max(computedMaxSlots, 1) }, (_, i) => classified.equipped[i] || null);
  $: slotsFull = classified.equipped.length >= computedMaxSlots;
  $: hotbarBinds = normalizeHotbarBinds($settingsStore.interface?.hotbarBinds);

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

  function toast(text) {
    if (overlayStore?.pushMessage) overlayStore.pushMessage(text);
  }

  function skillName(skill) {
    return skill?.name || skill?.id || 'Skill';
  }

  function equipSkill(skill) {
    if (!sendMessage || !skill) return;
    if (inCombat) {
      toast('You cannot change skills during combat.');
      return;
    }
    if (slotsFull) {
      toast(`All ${computedMaxSlots} skill slots are full. Unequip a skill first.`);
      return;
    }
    sendMessage('skills equip ' + skillName(skill));
  }

  function unequipSkill(skill) {
    if (!sendMessage || !skill) return;
    if (inCombat) {
      toast('You cannot change skills during combat.');
      return;
    }
    sendMessage('skills unequip ' + skillName(skill));
  }

  function addSkillToHotbar(skill) {
    if (!skill) return;
    const result = bindSkillToFirstEmptyHotbar(hotbarBinds, skill.id || skill.name);
    if (result.status === 'already') {
      toast(`${skillName(skill)} is already on the hotbar.`);
      return;
    }
    if (result.status === 'full') {
      toast('Hotbar is full — right-click a slot to replace it.');
      return;
    }
    if (result.status !== 'bound') {
      toast('Could not bind that skill.');
      return;
    }
    settingsStore.setSetting('interface', 'hotbarBinds', result.binds);
    toast(`${skillName(skill)} bound to hotbar slot ${result.index + 1}.`);
  }
</script>

<style>
  /* Base panel styling comes from global .game-panel class in themes.css */
  .character-widget {
    container-type: inline-size;
    container-name: character;
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

  .rest-badge {
    font-size: var(--text-xs);
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #86efac;
    background: rgba(34, 197, 94, 0.15);
    padding: 0.2em 0.6em;
    border-radius: 4px;
    border: 1px solid rgba(34, 197, 94, 0.35);
  }

  .rest-chip {
    font-size: 0.65rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: #86efac;
    background: rgba(34, 197, 94, 0.16);
    border: 1px solid rgba(34, 197, 94, 0.4);
    border-radius: 999px;
    padding: 0.08em 0.5em;
    margin-right: 0.35em;
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

  /* ── Responsive: narrow containers (< 240px) ── */
  @container character (max-width: 240px) {
    .char-identity {
      margin-bottom: 0.6em;
    }

    .char-name {
      font-size: var(--text-base);
    }

    .char-info {
      font-size: var(--text-xs);
    }

    .stat-bars {
      gap: 0.4em;
      margin-bottom: 0.5em;
    }

    .bar-label, .bar-value {
      font-size: var(--text-xs);
    }

    .gold-row {
      padding: 0.35em 0.45em;
      margin-bottom: 0.5em;
    }

    .gold-label {
      font-size: var(--text-xs);
    }

    .gold-value {
      font-size: var(--text-sm);
    }

    .attributes-grid {
      grid-template-columns: 1fr 1fr;
      gap: 0.3em;
    }

    .attr-item {
      padding: 0.35em 0.25em;
    }

    .attr-value {
      font-size: var(--text-sm);
    }

    .combat-stats-grid {
      gap: 0.3em;
    }

    .combat-stat {
      padding: 0.35em 0.4em;
      gap: 0.35em;
    }

    .combat-stat-value {
      font-size: var(--text-sm);
    }

    .combat-stat-formula {
      font-size: 0.7em;
    }
  }

  /* ── Responsive: very narrow containers (< 180px) ── */
  @container character (max-width: 180px) {
    .char-name {
      font-size: var(--text-sm);
    }

    .stat-bars {
      gap: 0.3em;
      margin-bottom: 0.4em;
    }

    .bar-header {
      margin-bottom: 0.1em;
    }

    .gold-row {
      padding: 0.25em 0.35em;
      margin-bottom: 0.4em;
    }

    .attributes-grid {
      grid-template-columns: 1fr 1fr;
      gap: 0.25em;
    }

    .attr-item {
      padding: 0.25em 0.15em;
    }

    .attr-value {
      font-size: var(--text-xs);
    }

    .attr-mod {
      font-size: 0.65em;
    }

    .combat-stats-grid {
      grid-template-columns: 1fr;
      gap: 0.25em;
    }

    .combat-stat {
      padding: 0.3em 0.35em;
      gap: 0.3em;
    }

    .stat-icon {
      font-size: 0.95em;
    }

    .combat-stat-value {
      font-size: var(--text-xs);
    }
  }

  /* ── Responsive: wider containers (> 320px) ── */
  @container character (min-width: 320px) {
    .attributes-grid {
      grid-template-columns: repeat(5, 1fr);
    }

    .combat-stats-grid {
      grid-template-columns: repeat(3, 1fr);
    }
  }

  /* Internal Stats | Skills tabs */
  .sheet-tabs {
    display: flex;
    gap: 0.35em;
    margin: 0 0 0.85em;
    padding: 0.2em;
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 8px;
  }

  .sheet-tab {
    flex: 1;
    min-height: 40px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.35em;
    padding: 0.45em 0.6em;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--text-secondary);
    font-family: var(--font-display);
    font-size: var(--text-sm);
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease;
  }

  .sheet-tab:hover {
    color: var(--text-primary);
    background: var(--tab-hover-bg);
  }

  .sheet-tab.active {
    color: var(--tab-active-color);
    background: var(--tab-active-bg);
    box-shadow: inset 0 0 0 1px var(--tab-active-border);
  }

  .sheet-tab i {
    font-size: 1.05em;
    opacity: 0.85;
  }

  .tab-count {
    font-family: inherit;
    font-size: 0.75em;
    font-weight: 700;
    letter-spacing: 0;
    color: var(--accent-primary);
    background: var(--accent-subtle);
    border: 1px solid var(--panel-inner-border);
    border-radius: 999px;
    padding: 0.05em 0.45em;
  }

  .skills-note {
    font-size: var(--text-xs);
    color: var(--text-dim);
    margin: -0.35em 0 0.65em;
  }

  .skills-note.combat {
    color: #fca5a5;
  }

  .skill-slots {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(118px, 1fr));
    gap: 0.4em;
    margin-bottom: 0.35em;
  }

  .skill-slot {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 0.35em;
    min-height: 88px;
    padding: 0.5em 0.45em;
    background: var(--panel-inner-bg);
    border: 1px solid rgba(212, 164, 74, 0.28);
    border-radius: 8px;
  }

  .skill-slot.empty {
    border-style: dashed;
    border-color: rgba(148, 163, 184, 0.28);
    background: rgba(0, 0, 0, 0.22);
    align-items: center;
    justify-content: center;
    color: var(--text-dim);
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .skill-slot-top {
    display: flex;
    align-items: center;
    gap: 0.4em;
  }

  .skill-icon {
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    object-fit: contain;
    image-rendering: pixelated;
    background: rgba(0, 0, 0, 0.35);
    border-radius: 6px;
    border: 1px solid var(--panel-inner-border);
  }

  .skill-slot-name {
    font-size: var(--text-sm);
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.2;
  }

  .skill-slot-meta {
    font-size: var(--text-xs);
    color: var(--text-dim);
  }

  .skill-list {
    display: flex;
    flex-direction: column;
    gap: 0.45em;
  }

  .skill-card {
    display: flex;
    flex-direction: column;
    gap: 0.35em;
    padding: 0.55em 0.6em;
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 8px;
  }

  .skill-card.locked {
    opacity: 0.55;
    filter: grayscale(0.55);
  }

  .skill-card-head {
    display: flex;
    align-items: flex-start;
    gap: 0.5em;
  }

  .skill-card-copy {
    flex: 1;
    min-width: 0;
  }

  .skill-card-name {
    font-size: var(--text-sm);
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1.2;
  }

  .skill-card-desc {
    font-size: var(--text-xs);
    color: var(--text-secondary);
    line-height: 1.35;
    margin-top: 0.15em;
  }

  .skill-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25em;
  }

  .skill-chip {
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    color: #e8dcc8;
    background: rgba(212, 164, 74, 0.12);
    border: 1px solid rgba(212, 164, 74, 0.28);
    border-radius: 999px;
    padding: 0.12em 0.5em;
  }

  .skill-chip.cost {
    color: #93c5fd;
    background: rgba(59, 130, 246, 0.12);
    border-color: rgba(59, 130, 246, 0.3);
  }

  .skill-chip.level {
    color: var(--text-secondary);
    background: rgba(255, 255, 255, 0.04);
    border-color: var(--panel-inner-border);
  }

  .skill-chip.lock {
    color: #fbbf24;
    background: rgba(251, 191, 36, 0.1);
    border-color: rgba(251, 191, 36, 0.28);
  }

  .skill-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35em;
    margin-top: 0.15em;
  }

  .skill-btn {
    flex: 1 1 auto;
    min-height: 36px;
    min-width: 88px;
    padding: 0.35em 0.6em;
    border-radius: 6px;
    border: 1px solid rgba(212, 164, 74, 0.35);
    background: rgba(212, 164, 74, 0.12);
    color: var(--accent-primary);
    font: inherit;
    font-size: var(--text-xs);
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    cursor: pointer;
  }

  .skill-btn:hover:not(:disabled) {
    background: rgba(212, 164, 74, 0.22);
    border-color: rgba(212, 164, 74, 0.55);
  }

  .skill-btn.ghost {
    background: transparent;
    color: var(--text-secondary);
    border-color: var(--panel-inner-border);
  }

  .skill-btn.danger {
    color: #fca5a5;
    background: rgba(239, 68, 68, 0.1);
    border-color: rgba(239, 68, 68, 0.35);
  }

  .skill-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .skills-empty {
    font-size: var(--text-sm);
    color: var(--text-dim);
    font-style: italic;
    padding: 0.25em 0.1em;
  }

  @container character (max-width: 240px) {
    .sheet-tab {
      min-height: 44px;
      font-size: var(--text-xs);
      padding: 0.5em 0.35em;
    }

    .skill-btn {
      min-height: 40px;
    }
  }
</style>

<div class="character-widget game-panel" class:in-combat={inCombat}>
  <div class="game-panel-header">
    <i class="material-icons">{inCombat ? 'swords' : 'person'}</i>
    <span class="widget-title">Character</span>
    {#if inCombat}
      <span class="combat-badge">In Combat</span>
    {:else if resting}
      <span class="rest-badge">Resting</span>
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

    <div class="sheet-tabs" role="tablist" aria-label="Character sheet">
      <button
        type="button"
        class="sheet-tab"
        class:active={sheetTab === 'stats'}
        role="tab"
        aria-selected={sheetTab === 'stats'}
        on:click={() => sheetTab = 'stats'}
      >
        <i class="material-icons">bar_chart</i>
        Stats
      </button>
      <button
        type="button"
        class="sheet-tab"
        class:active={sheetTab === 'skills'}
        role="tab"
        aria-selected={sheetTab === 'skills'}
        on:click={() => sheetTab = 'skills'}
      >
        <i class="material-icons">auto_awesome</i>
        Skills
        <span class="tab-count">{classified.equipped.length}/{computedMaxSlots}</span>
      </button>
    </div>

    {#if sheetTab === 'stats'}
    <div class="stat-bars">
      <!-- HP Bar -->
      <div class="bar-container">
        <div class="bar-header">
          <span class="bar-label hp" class:danger={hpPercent <= 60 && hpPercent > 30} class:critical={hpPercent <= 30}>HP</span>
          <span class="bar-value">{#if resting}<span class="rest-chip">Resting</span>{/if}{currentHp} / {maxHp}</span>
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
    <!-- Skills tab -->
    <div class="game-panel-divider">
      <span>Equipped {classified.equipped.length}/{computedMaxSlots}</span>
    </div>
    {#if inCombat}
      <div class="skills-note combat">Skill slots are locked during combat.</div>
    {:else}
      <div class="skills-note">Bind unlocked skills into slots, then add them to the hotbar.</div>
    {/if}

    <div class="skill-slots">
      {#each slotPlaceholders as slot, i}
        {#if slot}
          <div class="skill-slot">
            <div class="skill-slot-top">
              <img
                class="skill-icon"
                src={skillGenericArtUrl(slot.id || slot.name)}
                alt=""
                on:error={(e) => onItemArtError(e, { type: 'default' })}
              />
              <div>
                <div class="skill-slot-name">{skillName(slot)}</div>
                <div class="skill-slot-meta">Slot {i + 1} · {formatSkillCost(slot)}</div>
              </div>
            </div>
            <div class="skill-actions">
              <button
                type="button"
                class="skill-btn ghost"
                on:click={() => addSkillToHotbar(slot)}
              >Add to hotbar</button>
              <button
                type="button"
                class="skill-btn danger"
                disabled={inCombat}
                on:click={() => unequipSkill(slot)}
              >Unequip</button>
            </div>
          </div>
        {:else}
          <div class="skill-slot empty">Empty slot {i + 1}</div>
        {/if}
      {/each}
    </div>

    <div class="game-panel-divider">
      <span>Available</span>
    </div>
    {#if classified.available.length === 0}
      <div class="skills-empty">
        {classified.equipped.length > 0 ? 'All unlocked skills are equipped.' : 'No skills unlocked yet.'}
      </div>
    {:else}
      <div class="skill-list">
        {#each classified.available as skill (skill.id)}
          <div class="skill-card">
            <div class="skill-card-head">
              <img
                class="skill-icon"
                src={skillGenericArtUrl(skill.id)}
                alt=""
                on:error={(e) => onItemArtError(e, { type: 'default' })}
              />
              <div class="skill-card-copy">
                <div class="skill-card-name">{skill.name}</div>
                <div class="skill-card-desc">{skill.description}</div>
              </div>
            </div>
            <div class="skill-chips">
              <span class="skill-chip level">Lv {skill.levelRequired}</span>
              <span class="skill-chip cost">{formatSkillCost(skill)}</span>
              {#each formatSkillEffects(skill) as chip}
                <span class="skill-chip">{chip}</span>
              {/each}
            </div>
            <div class="skill-actions">
              <button
                type="button"
                class="skill-btn"
                disabled={inCombat || slotsFull}
                title={slotsFull ? 'All skill slots are full' : 'Equip into an empty slot'}
                on:click={() => equipSkill(skill)}
              >Equip</button>
              <button
                type="button"
                class="skill-btn ghost"
                title="Equip first to cast; you can still pin it on the hotbar"
                on:click={() => addSkillToHotbar(skill)}
              >Add to hotbar</button>
            </div>
          </div>
        {/each}
      </div>
    {/if}

    {#if classified.locked.length > 0}
      <div class="game-panel-divider">
        <span>Locked</span>
      </div>
      <div class="skill-list">
        {#each classified.locked as skill (skill.id)}
          <div class="skill-card locked">
            <div class="skill-card-head">
              <img
                class="skill-icon"
                src={skillGenericArtUrl(skill.id)}
                alt=""
                on:error={(e) => onItemArtError(e, { type: 'default' })}
              />
              <div class="skill-card-copy">
                <div class="skill-card-name">{skill.name}</div>
                <div class="skill-card-desc">{skill.description}</div>
              </div>
            </div>
            <div class="skill-chips">
              <span class="skill-chip lock">Unlocks at L{skill.levelRequired}</span>
              <span class="skill-chip cost">{formatSkillCost(skill)}</span>
              {#each formatSkillEffects(skill) as chip}
                <span class="skill-chip">{chip}</span>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
    {/if}
  {:else}
    <div class="empty-state">
      <i class="material-icons">person_off</i>
      <span>No character selected</span>
    </div>
  {/if}
</div>
