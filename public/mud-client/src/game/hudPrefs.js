/**
 * Player HUD preferences: action-bar pins + inventory open mode + hotbar binds.
 * Pure helpers (testable) + SettingsStore-backed persistence.
 */

export const DEFAULT_ACTION_BAR_PINS = ['look', 'inv', 'map'];

export const INVENTORY_OPEN_OVERLAY = 'overlay';
export const INVENTORY_OPEN_WIDGET = 'widget';
export const DEFAULT_INVENTORY_OPEN_MODE = INVENTORY_OPEN_OVERLAY;

export const HOTBAR_SLOT_COUNT = 8;
export const DEFAULT_HOTBAR_BINDS = Object.freeze(
  Array.from({ length: HOTBAR_SLOT_COUNT }, () => null)
);

/** Seed skill id → display name (cast matches Name, not id). */
export const SKILL_LABELS = {
  warrior_power_strike: 'Power Strike',
  warrior_shield_bash: 'Shield Bash',
  warrior_battle_cry: 'Battle Cry',
  warrior_cleave: 'Cleave',
  warrior_berserker_rage: 'Berserker Rage',
  rogue_backstab: 'Backstab',
  rogue_poison_strike: 'Poison Strike',
  rogue_evasion: 'Evasion',
  rogue_shadow_strike: 'Shadow Strike',
  rogue_flurry: 'Flurry',
  mage_fireball: 'Fireball',
  mage_frost_shield: 'Frost Shield',
  mage_lightning_bolt: 'Lightning Bolt',
  mage_arcane_burst: 'Arcane Burst',
  mage_mana_shield: 'Mana Shield',
  cleric_heal: 'Heal',
  cleric_holy_strike: 'Holy Strike',
  cleric_shield_of_faith: 'Shield of Faith',
  cleric_smite: 'Smite',
  cleric_divine_light: 'Divine Light',
  ranger_aimed_shot: 'Aimed Shot',
  ranger_volley: 'Volley',
  ranger_natures_gift: "Nature's Gift",
  ranger_pin_down: 'Pin Down',
  druid_wrath: 'Wrath',
  druid_rejuvenation: 'Rejuvenation',
  druid_entangle: 'Entangle',
  druid_starfire: 'Starfire',
  druid_barkskin: 'Barkskin',
};

/** Catalog of pinnable action-bar commands (id used in pins array). */
export const PINNABLE_COMMANDS = [
  { id: 'look', name: 'look', icon: 'visibility', label: 'Look', kind: 'command' },
  { id: 'say', name: 'say', icon: 'chat', label: 'Say', kind: 'say' },
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
  // Special kinds (inventory / map / say) need UI handlers — never bare commands.
  if (pin.kind === 'command') return pin.name;
  return null;
}

export function skillDisplayName(idOrName) {
  const raw = String(idOrName || '').trim();
  if (!raw) return 'Skill';
  if (SKILL_LABELS[raw]) return SKILL_LABELS[raw];
  const lower = raw.toLowerCase();
  for (const [id, name] of Object.entries(SKILL_LABELS)) {
    if (id === lower || name.toLowerCase() === lower) return name;
  }
  if (!raw.includes('_')) return raw;
  const parts = raw.split('_');
  const body = parts.length > 1 ? parts.slice(1) : parts;
  return body.map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
}

export function skillMaterialIcon(idOrName) {
  const key = String(idOrName || '').toLowerCase();
  if (/fire|flame|smite|wrath|starfire/.test(key)) return 'local_fire_department';
  if (/frost|ice|cold/.test(key)) return 'ac_unit';
  if (/heal|divine|rejuven|faith|holy|light/.test(key)) return 'healing';
  if (/shield|bark|mana_shield|frost_shield/.test(key)) return 'security';
  if (/lightning|arcane|bolt/.test(key)) return 'bolt';
  if (/poison|shadow|backstab|flurry/.test(key)) return 'visibility_off';
  if (/shot|volley|pin|aimed/.test(key)) return 'my_location';
  if (/strike|bash|cleave|rage|cry/.test(key)) return 'swords';
  return 'auto_awesome';
}

export function normalizeHotbarBind(raw) {
  if (raw == null || raw === false) return null;
  if (typeof raw !== 'object') return null;
  const kind = String(raw.kind || '').trim().toLowerCase();
  if (kind === 'item') {
    const name = String(raw.name || '').trim();
    const id = String(raw.id || raw.templateId || '').trim();
    if (!name && !id) return null;
    const bind = { kind: 'item' };
    if (name) bind.name = name;
    if (id) bind.id = id;
    return bind;
  }
  if (kind === 'skill') {
    const id = String(raw.id || '').trim();
    const name = String(raw.name || '').trim() || (id ? skillDisplayName(id) : '');
    if (!id && !name) return null;
    const bind = { kind: 'skill' };
    if (id) bind.id = id;
    if (name) bind.name = name;
    return bind;
  }
  return null;
}

export function normalizeHotbarBinds(binds) {
  const src = Array.isArray(binds) ? binds : [];
  const out = [];
  for (let i = 0; i < HOTBAR_SLOT_COUNT; i++) {
    out.push(normalizeHotbarBind(src[i]));
  }
  return out;
}

export function isConsumableItem(item) {
  if (!item) return false;
  return item.type === 'consumable' || item.consumable === true;
}

export function findInventoryItem(inventory, bind) {
  if (!bind || bind.kind !== 'item') return null;
  const items = Array.isArray(inventory) ? inventory : [];
  const wantName = String(bind.name || '').trim().toLowerCase();
  const wantId = String(bind.id || '').trim();
  if (wantId) {
    const byId = items.find((it) => {
      const tid = String(it.templateId || it.id || '').trim();
      return tid === wantId || tid.startsWith(`${wantId}~`) || String(it.id || '') === wantId;
    });
    if (byId) return byId;
  }
  if (wantName) {
    return items.find((it) => String(it.name || '').trim().toLowerCase() === wantName) || null;
  }
  return null;
}

/**
 * Resolve whether a hotbar bind can fire and which command to send.
 * Skills are combat-gated; consumables are not.
 */
export function resolveHotbarActivation(bind, { inCombat = false, inventory = [] } = {}) {
  const normalized = normalizeHotbarBind(bind);
  if (!normalized) {
    return { ok: false, reason: 'empty', command: null };
  }
  if (normalized.kind === 'skill') {
    if (!inCombat) {
      return {
        ok: false,
        reason: 'You can only use skills in combat.',
        command: null,
        gated: true,
      };
    }
    const name = normalized.name || skillDisplayName(normalized.id);
    return { ok: true, reason: null, command: `cast ${name}`, gated: false };
  }
  if (normalized.kind === 'item') {
    const item = findInventoryItem(inventory, normalized);
    if (!item) {
      const label = normalized.name || 'that item';
      return {
        ok: false,
        reason: `You don't have ${label}.`,
        command: null,
        missing: true,
      };
    }
    return { ok: true, reason: null, command: `use ${item.name}`, gated: false };
  }
  return { ok: false, reason: 'empty', command: null };
}

export function makeSkillBind(skillId) {
  const id = String(skillId || '').trim();
  if (!id) return null;
  return normalizeHotbarBind({ kind: 'skill', id, name: skillDisplayName(id) });
}

export function makeItemBind(item) {
  if (!item) return null;
  return normalizeHotbarBind({
    kind: 'item',
    name: item.name,
    id: item.templateId || item.id,
  });
}
