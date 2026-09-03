/**
 * Player HUD preferences: action-bar pins + inventory open mode + hotbar binds.
 * Pure helpers (testable) + SettingsStore-backed persistence.
 *
 * Option C layout:
 * - Action bar = room only (dirs, room actions, Shop) + fixed INV/MAP/SAY chrome
 * - Hotbar = skills + consumables; Look/Rest/Talk/Flee bindable but not seeded
 * - Search must never alias look
 */

/** Bump when default pin/chrome layout changes; migrates saved settings once. */
export const ACTION_BAR_LAYOUT_REVISION = 2;

/** Option C: no default command pins — Look/Rest/etc. are optional via ⋯ */
export const DEFAULT_ACTION_BAR_PINS = [];

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

/**
 * Fixed action-bar chrome (always shown out of combat). Not pins — cannot be
 * removed via Customize; INV/MAP/SAY live here under Option C.
 */
export const ACTION_BAR_CHROME = [
  { id: 'inv', name: 'inv', icon: 'inventory_2', label: 'Inv', kind: 'inventory' },
  { id: 'map', name: 'map', icon: 'map', label: 'Map', kind: 'map' },
  { id: 'say', name: 'say', icon: 'chat', label: 'Say', kind: 'say' },
];

const CHROME_IDS = new Set(ACTION_BAR_CHROME.map((c) => c.id));

/** Optional pins via ⋯ (Look/Rest/Help/…). Chrome ids are rejected. */
export const PINNABLE_COMMANDS = [
  { id: 'look', name: 'look', icon: 'visibility', label: 'Look', kind: 'command' },
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
  // Empty array is valid Option C default (room + chrome only).
  const src = Array.isArray(pins) ? pins : DEFAULT_ACTION_BAR_PINS;
  const out = [];
  const seen = new Set();
  for (const raw of src) {
    const id = String(raw || '').trim().toLowerCase();
    if (!PIN_IDS.has(id) || CHROME_IDS.has(id) || seen.has(id)) continue;
    seen.add(id);
    out.push(id);
  }
  return out;
}

/**
 * One-shot migration to Option C: reset legacy pin layouts so room bar is
 * clean. INV/MAP/SAY become fixed chrome (not pins). Custom pins from
 * revision 2+ are preserved.
 */
export function migrateActionBarPins(pins, revision) {
  const rev = Number(revision) || 0;
  if (rev >= ACTION_BAR_LAYOUT_REVISION) {
    return normalizeActionBarPins(pins);
  }
  return [...DEFAULT_ACTION_BAR_PINS];
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
  if (!PIN_IDS.has(key) || CHROME_IDS.has(key)) return list;
  if (list.includes(key)) {
    return list.filter((p) => p !== key);
  }
  return [...list, key];
}

export function commandForPin(pin) {
  if (!pin) return null;
  // Special kinds (inventory / map / say) need UI handlers — never bare commands.
  if (pin.kind === 'command') return pin.name;
  return null;
}

/** Resolve chrome + optional pins for out-of-combat action bar. */
export function resolveActionBarChrome() {
  return ACTION_BAR_CHROME.slice();
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

/** Equipped skill id/name → generic item-art stem (no .png). */
export const SKILL_GENERIC_ART = {
  warrior_power_strike: 'generic-action-melee',
  warrior_shield_bash: 'generic-spell-stun',
  warrior_battle_cry: 'generic-spell-strength',
  warrior_cleave: 'generic-action-melee',
  warrior_berserker_rage: 'generic-spell-strength',
  rogue_backstab: 'generic-action-melee',
  rogue_poison_strike: 'generic-spell-poison',
  rogue_evasion: 'generic-spell-shield',
  rogue_shadow_strike: 'generic-spell-curse',
  rogue_flurry: 'generic-action-melee',
  mage_fireball: 'generic-spell-fire',
  mage_frost_shield: 'generic-spell-ice',
  mage_lightning_bolt: 'generic-spell-lightning',
  mage_arcane_burst: 'generic-spell-arcane',
  mage_mana_shield: 'generic-spell-shield',
  cleric_heal: 'generic-spell-heal',
  cleric_holy_strike: 'generic-spell-holy',
  cleric_shield_of_faith: 'generic-spell-shield',
  cleric_smite: 'generic-spell-holy',
  cleric_divine_light: 'generic-spell-heal',
  ranger_aimed_shot: 'generic-action-ranged',
  ranger_volley: 'generic-action-ranged',
  ranger_natures_gift: 'generic-spell-heal',
  ranger_pin_down: 'generic-spell-stun',
  druid_wrath: 'generic-spell-nature',
  druid_rejuvenation: 'generic-spell-heal',
  druid_entangle: 'generic-spell-nature',
  druid_starfire: 'generic-spell-holy',
  druid_barkskin: 'generic-spell-shield',
};

/** Bindable hotbar actions. Look/Rest/Talk/Flee are optional — never seeded. */
export const HOTBAR_ACTIONS = [
  { id: 'melee', label: 'Attack', command: 'attack', art: 'generic-action-melee' },
  { id: 'look', label: 'Look', command: 'look', art: 'generic-action-look' },
  { id: 'rest', label: 'Rest', command: 'rest', art: 'generic-action-rest' },
  { id: 'flee', label: 'Flee', command: 'flee', art: 'generic-action-flee' },
  { id: 'talk', label: 'Talk', command: 'talk', art: 'generic-action-talk' },
];

/** Drop legacy Search=look binds if present in saved hotbar slots. */
export function scrubLegacySearchBinds(binds) {
  const normalized = normalizeHotbarBinds(binds);
  return normalized.map((b) => {
    if (!b || b.kind !== 'action') return b;
    if (b.id === 'search' || String(b.command || '').toLowerCase() === 'search') {
      return null;
    }
    return b;
  });
}

export function skillGenericArtStem(idOrName) {
  const raw = String(idOrName || '').trim();
  const lower = raw.toLowerCase();
  if (SKILL_GENERIC_ART[raw] || SKILL_GENERIC_ART[lower]) {
    return SKILL_GENERIC_ART[raw] || SKILL_GENERIC_ART[lower];
  }
  for (const [id, stem] of Object.entries(SKILL_GENERIC_ART)) {
    if (SKILL_LABELS[id] && SKILL_LABELS[id].toLowerCase() === lower) return stem;
  }
  if (/fire|flame/.test(lower)) return 'generic-spell-fire';
  if (/frost|ice|cold/.test(lower)) return 'generic-spell-ice';
  if (/lightning|bolt/.test(lower)) return 'generic-spell-lightning';
  if (/heal|rejuven|divine/.test(lower)) return 'generic-spell-heal';
  if (/shield|bark|evasion/.test(lower)) return 'generic-spell-shield';
  if (/poison/.test(lower)) return 'generic-spell-poison';
  if (/holy|smite|starfire/.test(lower)) return 'generic-spell-holy';
  if (/curse|shadow|dark/.test(lower)) return 'generic-spell-curse';
  if (/rage|cry|strength/.test(lower)) return 'generic-spell-strength';
  if (/stun|sleep|bash|pin/.test(lower)) return 'generic-spell-stun';
  if (/arcane/.test(lower)) return 'generic-spell-arcane';
  if (/wrath|entangle|nature/.test(lower)) return 'generic-spell-nature';
  if (/shot|volley|bow/.test(lower)) return 'generic-action-ranged';
  if (/strike|cleave|slash|flurry|backstab/.test(lower)) return 'generic-action-melee';
  return 'generic-spell-arcane';
}

export function skillGenericArtFile(idOrName) {
  return `${skillGenericArtStem(idOrName)}.png`;
}

export function skillGenericArtUrl(idOrName) {
  return `/api/item-art/${skillGenericArtFile(idOrName)}`;
}

export function actionGenericArtUrl(actionId) {
  const found = HOTBAR_ACTIONS.find((a) => a.id === actionId);
  const stem = found ? found.art : 'generic-default';
  return `/api/item-art/${stem}.png`;
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
  if (kind === 'action') {
    const id = String(raw.id || '').trim().toLowerCase();
    const found = HOTBAR_ACTIONS.find((a) => a.id === id);
    if (!found) return null;
    return { kind: 'action', id: found.id, name: found.label, command: found.command };
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
  if (normalized.kind === 'action') {
    const found = HOTBAR_ACTIONS.find((a) => a.id === normalized.id);
    if (!found) return { ok: false, reason: 'empty', command: null };
    return { ok: true, reason: null, command: found.command, gated: false };
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

export function makeActionBind(actionId) {
  return normalizeHotbarBind({ kind: 'action', id: actionId });
}
