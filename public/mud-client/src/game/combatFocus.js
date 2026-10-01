export const WARNING_TIERS = new Set(['orange', 'red', 'skull']);

export function livingFocus(enemies, targetId) {
  const living = (enemies || []).filter((enemy) => enemy && (enemy.hp ?? 0) > 0);
  return living.find((enemy) => enemy.id === targetId) || living[0] || null;
}

export function changeFocus(state, targetId, warn = true) {
  const enemy = (state.combatEnemies || []).find((entry) => entry.id === targetId && (entry.hp ?? 0) > 0);
  if (!enemy || state.combatTargetId === enemy.id) return false;
  state.combatTargetId = enemy.id;
  state.combatThreatWarning = warn && WARNING_TIERS.has(enemy.threat)
    ? { text: `${enemy.name} is much stronger than you.`, tier: enemy.threat, at: Date.now() }
    : null;
  return true;
}
