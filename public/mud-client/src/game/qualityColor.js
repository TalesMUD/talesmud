/** Theme-aware item quality colours. Gilded remaps the variables; other themes keep the fallbacks. */

const ENGINE = {
  normal: '#9ca3af',
  magic: '#22c55e',
  rare: '#3b82f6',
  legendary: '#a855f7',
  mythic: '#f59e0b',
  unique: '#f6d77e',
};

const CARD = {
  normal: '#e5e7eb',
  magic: '#4ade80',
  rare: '#60a5fa',
  legendary: '#c084fc',
  mythic: '#fbbf24',
  unique: '#fde68a',
};

export function qualityColor(quality, tone = 'engine') {
  const table = tone === 'card' ? CARD : ENGINE;
  const key = Object.prototype.hasOwnProperty.call(table, quality) ? quality : 'normal';
  const variable = tone === 'card' ? `--gf-q-${key}` : `--color-quality-${key}`;
  return `var(${variable}, ${table[key]})`;
}

export function itemQualityKey(item) {
  if (!item) return 'empty';
  if (item.unique) return 'unique';
  return item.quality || 'normal';
}
