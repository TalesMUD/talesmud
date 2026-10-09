// Player-facing effect lines for items (onHit / onEquip / onUse).
// Authored effects come from content. A script without authored text still
// gets a neutral "Special effect" line so it is never silently hidden.

export const SPECIAL_EFFECT_TEXT = 'Special effect';

const TRIGGER_LABELS = {
  onHit: 'On hit',
  onEquip: 'While equipped',
  onUse: 'On use',
};

export function effectTriggerLabel(trigger) {
  if (!trigger) return 'Effect';
  return TRIGGER_LABELS[trigger] || String(trigger);
}

function clean(value) {
  return String(value == null ? '' : value).trim();
}

/** @returns {Array<{trigger:string,label:string,name:string,text:string}>} */
export function itemEffectLines(item) {
  if (!item) return [];
  const out = [];
  const covered = new Set();
  for (const eff of Array.isArray(item.effects) ? item.effects : []) {
    if (!eff) continue;
    const name = clean(eff.name);
    const text = clean(eff.text);
    if (!name && !text) continue;
    const trigger = clean(eff.trigger);
    covered.add(trigger);
    out.push({ trigger, label: effectTriggerLabel(trigger), name, text });
  }
  const fallback = (trigger, scriptId) => {
    if (!clean(scriptId) || covered.has(trigger)) return;
    out.push({ trigger, label: effectTriggerLabel(trigger), name: '', text: SPECIAL_EFFECT_TEXT });
  };
  fallback('onHit', item.onHitScriptId);
  fallback('onUse', item.onUseScriptId);
  return out;
}

/** Victory loot: rarity-unique drops first, the rest keep server order. */
export function sortLootUniqueFirst(list) {
  if (!Array.isArray(list)) return [];
  return list
    .map((item, index) => ({ item, index }))
    .sort((a, b) => (b.item && b.item.unique ? 1 : 0) - (a.item && a.item.unique ? 1 : 0) || a.index - b.index)
    .map((entry) => entry.item);
}

/** "UNIQUE: Unmarked Vigil Blade" → "Unmarked Vigil Blade" for the unique drop line. */
export function uniqueDropName(text) {
  return clean(text).replace(/^unique\s*:\s*/i, '');
}
