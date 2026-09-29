// Phase state comes from the engine, never inferred from client HP percentages.
export function phaseCaption(enemy) {
  if (!enemy?.bossPhase || !enemy?.bossPhaseLabel) return '';
  return `Phase ${enemy.bossPhase}${enemy.bossPhaseCount ? ` / ${enemy.bossPhaseCount}` : ''} · ${enemy.bossPhaseLabel}`;
}

export function phaseBanner(event, now, visible) {
  return visible && event?.text && now - event.at >= 0 && now - event.at < 4000 ? event.text : '';
}
