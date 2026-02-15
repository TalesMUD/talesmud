// playerColors.js — Assigns persistent, distinct colors to player names.
// 16 hand-picked colors readable on dark backgrounds (#0c0c0c).

const PLAYER_COLORS = [
  '#e06c75', // soft red
  '#98c379', // green
  '#e5c07b', // gold
  '#61afef', // blue
  '#c678dd', // purple
  '#56b6c2', // cyan
  '#d19a66', // orange
  '#ff6ac1', // pink
  '#a3be8c', // sage
  '#ebcb8b', // light gold
  '#88c0d0', // light blue
  '#b48ead', // lavender
  '#bf616a', // coral
  '#d08770', // salmon
  '#5e81ac', // steel blue
  '#8fbcbb', // teal
];

const STORAGE_KEY = 'talesmud_player_colors';

/** Load the name→colorIndex map from localStorage. */
function loadMap() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch (_) { /* ignore */ }
  return {};
}

/** Persist the map. */
function saveMap(map) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(map));
  } catch (_) { /* ignore */ }
}

/** Simple string hash → index in [0, 16). */
function hashName(name) {
  let h = 0;
  for (let i = 0; i < name.length; i++) {
    h = ((h << 5) - h + name.charCodeAt(i)) | 0;
  }
  return ((h % PLAYER_COLORS.length) + PLAYER_COLORS.length) % PLAYER_COLORS.length;
}

/**
 * Return the hex color assigned to `playerName`.
 * First call for a given name picks a color via hash and persists it.
 */
export function getPlayerColor(playerName) {
  if (!playerName) return PLAYER_COLORS[0];

  const map = loadMap();
  if (map[playerName] !== undefined) {
    return PLAYER_COLORS[map[playerName] % PLAYER_COLORS.length];
  }

  const idx = hashName(playerName);
  map[playerName] = idx;
  saveMap(map);
  return PLAYER_COLORS[idx];
}

/**
 * Parse a hex color "#rrggbb" into {r, g, b} integers.
 */
export function hexToRgb(hex) {
  const n = parseInt(hex.slice(1), 16);
  return { r: (n >> 16) & 0xff, g: (n >> 8) & 0xff, b: n & 0xff };
}
