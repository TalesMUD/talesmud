/**
 * Player HUD preferences: action-bar pins + inventory open mode.
 * Pure helpers (testable) + SettingsStore-backed persistence.
 */

export const DEFAULT_ACTION_BAR_PINS = ['look', 'inv', 'map'];

export const INVENTORY_OPEN_OVERLAY = 'overlay';
export const INVENTORY_OPEN_WIDGET = 'widget';
export const DEFAULT_INVENTORY_OPEN_MODE = INVENTORY_OPEN_OVERLAY;

/** Catalog of pinnable action-bar commands (id used in pins array). */
export const PINNABLE_COMMANDS = [
  { id: 'look', name: 'look', icon: 'visibility', label: 'Look', kind: 'command' },
  { id: 'inv', name: 'inv', icon: 'inventory_2', label: 'Inv', kind: 'inventory' },
  { id: 'map', name: 'map', icon: 'map', label: 'Map', kind: 'map' },
  { id: 'rest', name: 'rest', icon: 'hotel', label: 'Rest', kind: 'command' },
  { id: 'who', name: 'who', icon: 'people', label: 'Who', kind: 'command' },
  { id: 'help', name: 'help', icon: 'help', label: 'Help', kind: 'command' },
  { id: 'equipment', name: 'equipment', icon: 'shield', label: 'Equip', kind: 'command' },
  { id: 'character', name: 'character', icon: 'person', label: 'Stats', kind: 'command' },
  { id: 'bind', name: 'bind', icon: 'location_on', label: 'Bind', kind: 'command' },
  { id: 'drop', name: 'drop', icon: 'delete', label: 'Drop', kind: 'command' },
  { id: 'use', name: 'use', icon: 'touch_app', label: 'Use', kind: 'command' },
  { id: 'examine', name: 'examine', icon: 'search', label: 'Examine', kind: 'command' },
];

const PIN_IDS = new Set(PINNABLE_COMMANDS.map((c) => c.id));

export function normalizeActionBarPins(pins) {
  const src = Array.isArray(pins) ? pins : DEFAULT_ACTION_BAR_PINS;
  const out = [];
  const seen = new Set();
  for (const raw of src) {
    const id = String(raw || '').trim().toLowerCase();
    if (!PIN_IDS.has(id) || seen.has(id)) continue;
    seen.add(id);
    out.push(id);
  }
  return out.length > 0 ? out : [...DEFAULT_ACTION_BAR_PINS];
}

export function normalizeInventoryOpenMode(mode) {
  const m = String(mode || '').trim().toLowerCase();
  if (m === INVENTORY_OPEN_WIDGET) return INVENTORY_OPEN_WIDGET;
  return INVENTORY_OPEN_OVERLAY;
}

export function resolvePinnedCommands(pins) {
  const ids = normalizeActionBarPins(pins);
  return ids
    .map((id) => PINNABLE_COMMANDS.find((c) => c.id === id))
    .filter(Boolean);
}

export function togglePin(pins, id) {
  const list = normalizeActionBarPins(pins);
  const key = String(id || '').trim().toLowerCase();
  if (!PIN_IDS.has(key)) return list;
  if (list.includes(key)) {
    const next = list.filter((p) => p !== key);
    return next.length > 0 ? next : [...DEFAULT_ACTION_BAR_PINS];
  }
  return [...list, key];
}

export function commandForPin(pin) {
  if (!pin) return null;
  if (pin.kind === 'command') return pin.name;
  return null;
}
