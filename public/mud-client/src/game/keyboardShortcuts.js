/** 1–9 → hotbar index. Anything else is -1. */
export function hotbarSlotFromKey(key) {
  if (typeof key !== 'string' || key.length !== 1) return -1;
  const index = key.charCodeAt(0) - 49;
  if (index < 0 || index > 8) return -1;
  return index;
}

/** Highest open panel wins. Empty string when nothing is open. */
export const PANEL_ORDER = [
  'cheatSheet',
  'layoutDialog',
  'characterPicker',
  'settings',
  'addWidget',
  'map',
  'friends',
  'party',
  'inventory',
  'battleOutcome',
  'battleDock',
  'accountMenu',
  'widgetFocus',
  'editMode',
];

export function topOpenPanel(flags) {
  if (!flags) return '';
  for (const id of PANEL_ORDER) {
    if (flags[id]) return id;
  }
  return '';
}

/** True when the event target is a command line or any text field. */
export function isTextEntry(target) {
  if (!target || typeof target !== 'object') return false;
  const tag = String(target.tagName || '').toLowerCase();
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return true;
  if (target.isContentEditable) return true;
  if (typeof target.closest === 'function') {
    if (target.closest('input, textarea, select, [contenteditable="true"], .xterm')) return true;
  }
  return false;
}

export function rarityClass(quality) {
  const value = String(quality || 'normal').toLowerCase();
  if (value === 'magic' || value === 'rare' || value === 'legendary' || value === 'mythic') return value;
  return 'normal';
}

export function prefersReducedMotion() {
  if (typeof window === 'undefined' || !window.matchMedia) return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}
