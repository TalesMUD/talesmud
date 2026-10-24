// Equipped item score. Twin of pkg/ranking.DefaultWeights.
// piece = qualityWeight * (1 + level) + stat budget. A two-hand item counts once.

export const DEFAULT_QUALITY = {
  normal: 10,
  magic: 16,
  rare: 24,
  legendary: 36,
  mythic: 52,
};

export const UNKNOWN_QUALITY = 10;

export const STAT_KEYS = [
  'damage', 'defense', 'armor',
  'strength', 'agility', 'intelligence', 'wisdom', 'stamina',
  'health', 'mana',
  'spellPower', 'holyDamage', 'naturePower',
  'critChance', 'critical', 'criticalChance',
];

const SLOT_ORDER = [
  'head', 'chest', 'legs', 'boots', 'hands', 'waist', 'neck', 'ring1', 'ring2', 'main_hand', 'off_hand',
];

const NOT_WORN = new Set(['inventory', 'purse', 'container']);

export function scoreEquipped(equipped, weights = null) {
  return scorePieces(equippedPieces(equipped), weights || defaultWeights());
}

export function defaultWeights() {
  return {
    quality: { ...DEFAULT_QUALITY },
    unknown: UNKNOWN_QUALITY,
    statKeys: STAT_KEYS.slice(),
  };
}

export function equippedPieces(equipped) {
  if (!equipped) return [];
  if (Array.isArray(equipped)) return equipped.filter(Boolean);
  const pieces = [];
  const seen = new Set();
  for (const slot of SLOT_ORDER) {
    if (Object.prototype.hasOwnProperty.call(equipped, slot)) {
      if (equipped[slot]) pieces.push(equipped[slot]);
      seen.add(slot);
    }
  }
  for (const slot of Object.keys(equipped).sort()) {
    if (seen.has(slot) || NOT_WORN.has(slot)) continue;
    if (equipped[slot]) pieces.push(equipped[slot]);
  }
  return pieces;
}

export function scorePieces(pieces, weights = null) {
  const w = weights || defaultWeights();
  const seen = new Set();
  let total = 0;
  for (const piece of pieces || []) {
    const id = String(piece?.id || '').trim();
    if (!id || seen.has(id)) continue;
    seen.add(id);
    total += scorePiece(piece, w);
  }
  return total;
}

function scorePiece(piece, weights) {
  let level = Number(piece?.level);
  if (!Number.isFinite(level) || level < 0) level = 0;
  level = Math.trunc(level);
  const quality = String(piece?.quality || '').trim().toLowerCase();
  let weight = weights.quality?.[quality];
  if (!Number.isFinite(weight) || weight < 0) weight = weights.unknown;
  if (!Number.isFinite(weight) || weight < 0) weight = 0;
  return weight * (1 + level) + statBudget(piece?.attributes, weights.statKeys);
}

function statBudget(attrs, keys) {
  if (!attrs || !keys) return 0;
  let total = 0;
  for (const key of keys) {
    total += positiveStat(attrs[key]);
  }
  return total;
}

function positiveStat(raw) {
  if (typeof raw === 'boolean' || raw == null) return 0;
  if (typeof raw === 'number') {
    if (!Number.isFinite(raw) || raw <= 0) return 0;
    return Math.trunc(raw);
  }
  if (typeof raw === 'string') {
    const text = raw.trim();
    if (!text) return 0;
    const n = Number(text);
    if (!Number.isFinite(n) || n <= 0) return 0;
    return Math.trunc(n);
  }
  return 0;
}
