export const TOUCH_TARGET_PX = 44;

export const PHONE_SHEET_LAUNCHERS = [
  { id: 'character', label: 'Stats', icon: 'person' },
  { id: 'equipment', label: 'Equip', icon: 'shield' },
  { id: 'skills', label: 'Skills', icon: 'auto_awesome' },
  { id: 'quests', label: 'Quests', icon: 'assignment' },
];

const BARE_SHEET = {
  character: 'character',
  stats: 'character',
  equipment: 'equipment',
  equip: 'equipment',
  skills: 'skills',
  spells: 'skills',
  quests: 'quests',
  ql: 'quests',
  quest: 'quests',
};

export function sheetIdForCommand(command) {
  const raw = String(command || '').trim().toLowerCase();
  if (!raw || /\s/.test(raw)) return null;
  return BARE_SHEET[raw] || null;
}

export function floorPhoneFontPx(size, phone) {
  const text = String(size || '').trim();
  if (!phone) return text;
  const rem = /^([\d.]+)rem$/i.exec(text);
  const px = /^([\d.]+)px$/i.exec(text);
  const value = rem ? Number(rem[1]) * 16 : px ? Number(px[1]) : NaN;
  if (!Number.isFinite(value)) return text;
  if (value < 14) return '14px';
  return text;
}

function channel(hex, index) {
  const raw = String(hex || '').replace('#', '');
  const n = parseInt(raw.slice(index * 2, index * 2 + 2), 16) / 255;
  return n <= 0.03928 ? n / 12.92 : ((n + 0.055) / 1.055) ** 2.4;
}

export function relativeLuminance(hex) {
  return 0.2126 * channel(hex, 0) + 0.7152 * channel(hex, 1) + 0.0722 * channel(hex, 2);
}

export function contrastRatio(fg, bg) {
  const lighter = Math.max(relativeLuminance(fg), relativeLuminance(bg));
  const darker = Math.min(relativeLuminance(fg), relativeLuminance(bg));
  return (lighter + 0.05) / (darker + 0.05);
}

export const GILDED_PAIRS = [
  { fg: '#ffd24a', bg: '#15110e', role: 'gold on panel' },
  { fg: '#ffd24a', bg: '#2d2017', role: 'gold on row' },
  { fg: '#ffd24a', bg: '#1c140e', role: 'gold on row-b' },
  { fg: '#3ee04a', bg: '#15110e', role: 'green stat on panel' },
  { fg: '#3ee04a', bg: '#1c140e', role: 'green stat on row-b' },
  { fg: '#efe6d6', bg: '#15110e', role: 'body on panel' },
  { fg: '#a89c8a', bg: '#15110e', role: 'dim on panel' },
];

export function hotbarSlotLayout(viewportWidth, slots = 9, opts = {}) {
  const padding = opts.padding ?? 8;
  const gap = opts.gap ?? 4;
  const min = opts.min ?? TOUCH_TARGET_PX;
  const max = opts.max ?? 56;
  const inner = viewportWidth - padding * 2;
  let perRow = 1;
  for (let n = slots; n >= 1; n -= 1) {
    if (n * min + (n - 1) * gap <= inner) {
      perRow = n;
      break;
    }
  }
  const rows = Math.ceil(slots / perRow);
  const width = Math.min(max, Math.max(min, (inner - (perRow - 1) * gap) / perRow));
  const rowWidth = perRow * width + (perRow - 1) * gap;
  return { perRow, rows, width, fits: width >= min && rowWidth <= inner + 0.01 };
}
