const STAT_LABELS = {
  damage: 'ATK', defense: 'DEF', armor: 'ARM', strength: 'STR',
  agility: 'AGI', intelligence: 'INT', health: 'HP', mana: 'MP',
  speed: 'SPD', critical: 'CRIT', wisdom: 'WIS', stamina: 'STA',
  spellPower: 'SPELL', holyDamage: 'HOLY', attackSpeed: 'SPD',
  damageMin: 'MIN ATK', damageMax: 'MAX ATK', criticalChance: 'CRIT',
};

const CLASS_WEIGHTS = {
  warrior: { damage: 3, damageMin: 2, damageMax: 2, strength: 2, armor: 2, defense: 2, health: 1 },
  rogue: { damage: 3, damageMin: 2, damageMax: 2, agility: 2, critical: 2, criticalChance: 2, speed: 1, attackSpeed: 1, armor: 1 },
  ranger: { damage: 3, damageMin: 2, damageMax: 2, agility: 2, critical: 1, criticalChance: 1, speed: 1, attackSpeed: 1, armor: 1 },
  hunter: { damage: 3, damageMin: 2, damageMax: 2, agility: 2, critical: 1, criticalChance: 1, speed: 1, attackSpeed: 1, armor: 1 },
  mage: { damage: 2, spellPower: 4, intelligence: 3, mana: 2, wisdom: 1, armor: 1 },
  wizard: { damage: 2, spellPower: 4, intelligence: 3, mana: 2, wisdom: 1, armor: 1 },
  cleric: { damage: 1, holyDamage: 3, spellPower: 3, wisdom: 3, mana: 2, armor: 2, health: 1 },
  druid: { damage: 1, spellPower: 3, wisdom: 3, intelligence: 2, mana: 2, health: 1 },
};

const ARMOR_RANK = { cloth: 1, leather: 2, plate: 3 };

export function classId(character) {
  const raw = character?.class || character?.classId || '';
  return String(typeof raw === 'object' ? (raw.id || raw.name || '') : raw).toLowerCase();
}

export function itemWeight(item) {
  const raw = item?.weight ?? item?.properties?.weight;
  if (raw === undefined || raw === null || raw === '') return null;
  const value = Number(raw);
  return Number.isFinite(value) && value >= 0 ? value : null;
}

export function numericStats(item) {
  const stats = {};
  for (const [key, raw] of Object.entries(item?.attributes || {})) {
    const value = Number(raw);
    if (Number.isFinite(value)) stats[key] = value;
  }
  return stats;
}

export function itemUsabilityReason(item, character) {
  if (!item || !character) return '';
  const allowed = (item.tags || []).filter((tag) => String(tag).startsWith('class:'))
    .map((tag) => String(tag).slice(6).toLowerCase());
  const cls = classId(character);
  if (allowed.length && !allowed.includes(cls)) return `Requires ${allowed.join(' or ')} class`;
  if (Number(item.level) > Number(character.level || 1)) return `Requires level ${item.level}`;
  const armorWeight = String(item.properties?.armorWeight ||
    (item.tags || []).find((tag) => String(tag).startsWith('armor:'))?.slice(6) || '').toLowerCase();
  const classArmor = String(character.class?.armorType || character.armorType || '').toLowerCase();
  if (item.type === 'armor' && ARMOR_RANK[armorWeight] && ARMOR_RANK[classArmor] &&
      ARMOR_RANK[armorWeight] > ARMOR_RANK[classArmor]) {
    return `Requires ${armorWeight} armor training`;
  }
  return '';
}

export function isTwoHanded(item) {
  return String(item?.subType || '').toLowerCase() === 'twohandsword' ||
    item?.properties?.twoHanded === true;
}

function uniqueItems(items) {
  const seen = new Set();
  return items.filter((item) => {
    if (!item) return false;
    const key = item.id || item.name;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

export function scoreItem(item, character) {
  const weights = CLASS_WEIGHTS[classId(character)] || CLASS_WEIGHTS.warrior;
  return Object.entries(numericStats(item)).reduce((sum, [key, value]) =>
    sum + value * (weights[key] || 0), 0);
}

/** The pieces replaced by this candidate. Rings use the weaker worn ring. */
export function comparisonItems(item, equippedItems = {}, character = null) {
  if (!item?.slot) return [];
  if (item.slot === 'ring1' || item.slot === 'ring2') {
    const rings = [equippedItems.ring1, equippedItems.ring2];
    if (!rings[0] || !rings[1]) return [];
    return [scoreItem(rings[0], character) <= scoreItem(rings[1], character) ? rings[0] : rings[1]];
  }
  if (isTwoHanded(item)) return uniqueItems([equippedItems.main_hand, equippedItems.off_hand]);
  if (item.slot === 'main_hand' && isTwoHanded(equippedItems.main_hand)) {
    return uniqueItems([equippedItems.main_hand, equippedItems.off_hand]);
  }
  return uniqueItems([equippedItems[item.slot]]);
}

export function comparisonRows(item, equippedItems = {}, character = null) {
  const worn = comparisonItems(item, equippedItems, character);
  const candidate = numericStats(item);
  const current = {};
  for (const old of worn) {
    for (const [key, value] of Object.entries(numericStats(old))) {
      current[key] = (current[key] || 0) + value;
    }
  }
  const keys = new Set([...Object.keys(candidate), ...Object.keys(current)]);
  const rows = [...keys].map((key) => ({
    key, label: STAT_LABELS[key] || key.toUpperCase(),
    value: candidate[key] || 0, worn: current[key] || 0,
    delta: (candidate[key] || 0) - (current[key] || 0),
  }));
  const weight = itemWeight(item);
  const wornWeights = worn.map(itemWeight);
  if (weight !== null && wornWeights.every((value) => value !== null)) {
    const oldWeight = wornWeights.reduce((sum, value) => sum + value, 0);
    rows.push({ key: 'weight', label: 'WT', value: weight, worn: oldWeight,
      delta: weight - oldWeight, lowerIsBetter: true });
  }
  return rows;
}

export function isClearUpgrade(item, equippedItems = {}, character = null) {
  if (!item?.slot || itemUsabilityReason(item, character)) return false;
  const weights = CLASS_WEIGHTS[classId(character)] || CLASS_WEIGHTS.warrior;
  const rows = comparisonRows(item, equippedItems, character).filter((row) => weights[row.key] > 0);
  if (!rows.length) return false;
  const gain = rows.reduce((sum, row) => sum + row.delta * weights[row.key], 0);
  return gain > 0 && rows.some((row) => row.delta > 0);
}
