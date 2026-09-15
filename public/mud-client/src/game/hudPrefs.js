/**
 * Player HUD preferences: action-bar pins + inventory open mode + hotbar binds.
 * Pure helpers (testable) + SettingsStore-backed persistence.
 *
 * Option C layout:
 * - Action bar = room only (dirs, room actions, Shop) + fixed INV/MAP/SAY chrome
 * - Hotbar = skills + consumables; Rest seeded on empty/default bar (slot 7);
 *   Look/Talk/Flee bindable but not seeded
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
/** 0-based index for the default Rest seed (slot 7). */
export const DEFAULT_REST_SLOT = 6;

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

/**
 * Combat skill catalog mirrored from pkg/entities/skills/seed.go.
 * Client uses this for Character → Skills (no extra server round-trip).
 */
export const SKILL_CATALOG = [
  // Warrior
  { id: 'warrior_power_strike', name: 'Power Strike', classIds: ['warrior'], levelRequired: 1, description: 'A powerful strike dealing 150% weapon damage.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 3, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 1.5, scalingAttr: 'STR', basePower: 3, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'warrior_shield_bash', name: 'Shield Bash', classIds: ['warrior'], levelRequired: 5, description: 'Bash the target with your shield, dealing damage and stunning for 1 round.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 4, effect: 'damage', buffStat: 'stun', buffPercent: 0, target: 'enemy', scalingFactor: 1.0, scalingAttr: 'STR', basePower: 2, duration: 1, hitCount: 0, ignoresDefense: false },
  { id: 'warrior_battle_cry', name: 'Battle Cry', classIds: ['warrior'], levelRequired: 10, description: 'Let out a battle cry, increasing attack power by 30% for 3 rounds.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 5, effect: 'buff', buffStat: 'attack', buffPercent: 0.30, target: 'self', scalingFactor: 0, scalingAttr: '', basePower: 0, duration: 3, hitCount: 0, ignoresDefense: false },
  { id: 'warrior_cleave', name: 'Cleave', classIds: ['warrior'], levelRequired: 15, description: 'Swing your weapon in a wide arc, hitting all enemies for 80% damage.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 4, effect: 'damage', buffStat: '', buffPercent: 0, target: 'all_enemies', scalingFactor: 0.8, scalingAttr: 'STR', basePower: 2, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'warrior_berserker_rage', name: 'Berserker Rage', classIds: ['warrior'], levelRequired: 20, description: 'Enter a berserker rage: +50% attack but -25% defense for 3 rounds.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 6, effect: 'buff', buffStat: 'attack', buffPercent: 0.50, target: 'self', scalingFactor: 0, scalingAttr: '', basePower: 0, duration: 3, hitCount: 0, ignoresDefense: false },
  // Rogue
  { id: 'rogue_backstab', name: 'Backstab', classIds: ['rogue'], levelRequired: 1, description: 'Strike from the shadows for 200% DEX-scaled damage.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 3, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 2.0, scalingAttr: 'DEX', basePower: 4, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'rogue_poison_strike', name: 'Poison Strike', classIds: ['rogue'], levelRequired: 5, description: 'Coat your blade with poison. Target takes damage each round for 3 rounds.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 4, effect: 'dot', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 0.5, scalingAttr: 'DEX', basePower: 3, duration: 3, hitCount: 0, ignoresDefense: false },
  { id: 'rogue_evasion', name: 'Evasion', classIds: ['rogue'], levelRequired: 10, description: 'Heighten your reflexes, gaining +75% dodge chance for 2 rounds.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 5, effect: 'buff', buffStat: 'dodge', buffPercent: 0.75, target: 'self', scalingFactor: 0, scalingAttr: '', basePower: 0, duration: 2, hitCount: 0, ignoresDefense: false },
  { id: 'rogue_shadow_strike', name: 'Shadow Strike', classIds: ['rogue'], levelRequired: 15, description: "Strike from the shadows, ignoring the target's armor.", resourceType: 'cooldown', manaCost: 0, cooldownRounds: 5, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 1.5, scalingAttr: 'DEX', basePower: 5, duration: 0, hitCount: 0, ignoresDefense: true },
  { id: 'rogue_flurry', name: 'Flurry', classIds: ['rogue'], levelRequired: 20, description: 'Unleash a flurry of 3 rapid strikes, each at 60% damage.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 6, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 0.6, scalingAttr: 'DEX', basePower: 2, duration: 0, hitCount: 3, ignoresDefense: false },
  // Mage (class id "wizard" aliases to mage)
  { id: 'mage_fireball', name: 'Fireball', classIds: ['mage'], levelRequired: 1, description: 'Hurl a ball of fire at the target.', resourceType: 'mana', manaCost: 8, cooldownRounds: 0, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 1.5, scalingAttr: 'INT', basePower: 6, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'mage_frost_shield', name: 'Frost Shield', classIds: ['mage'], levelRequired: 1, description: 'Surround yourself with a shield of ice, increasing defense by 50% for 2 rounds.', resourceType: 'mana', manaCost: 6, cooldownRounds: 0, effect: 'buff', buffStat: 'defense', buffPercent: 0.50, target: 'self', scalingFactor: 0, scalingAttr: '', basePower: 0, duration: 2, hitCount: 0, ignoresDefense: false },
  { id: 'mage_lightning_bolt', name: 'Lightning Bolt', classIds: ['mage'], levelRequired: 5, description: 'Call down a bolt of lightning that ignores armor.', resourceType: 'mana', manaCost: 15, cooldownRounds: 0, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 2.0, scalingAttr: 'INT', basePower: 10, duration: 0, hitCount: 0, ignoresDefense: true },
  { id: 'mage_arcane_burst', name: 'Arcane Burst', classIds: ['mage'], levelRequired: 10, description: 'Release a burst of arcane energy, damaging all enemies.', resourceType: 'mana', manaCost: 20, cooldownRounds: 0, effect: 'damage', buffStat: '', buffPercent: 0, target: 'all_enemies', scalingFactor: 1.2, scalingAttr: 'INT', basePower: 5, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'mage_mana_shield', name: 'Mana Shield', classIds: ['mage'], levelRequired: 15, description: 'Create a shield that absorbs damage by consuming mana.', resourceType: 'mana', manaCost: 12, cooldownRounds: 0, effect: 'buff', buffStat: 'mana_shield', buffPercent: 0, target: 'self', scalingFactor: 2.0, scalingAttr: 'INT', basePower: 20, duration: 3, hitCount: 0, ignoresDefense: false },
  // Cleric
  { id: 'cleric_heal', name: 'Heal', classIds: ['cleric'], levelRequired: 1, description: 'Channel divine energy to heal yourself.', resourceType: 'mana', manaCost: 8, cooldownRounds: 0, effect: 'heal', buffStat: '', buffPercent: 0, target: 'self', scalingFactor: 1.5, scalingAttr: 'WIS', basePower: 8, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'cleric_holy_strike', name: 'Holy Strike', classIds: ['cleric'], levelRequired: 1, description: 'Strike with holy power, dealing damage and healing yourself slightly.', resourceType: 'mana', manaCost: 6, cooldownRounds: 0, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 1.0, scalingAttr: 'WIS', basePower: 4, duration: 0, hitCount: 0, ignoresDefense: false, secondaryEffect: 'heal' },
  { id: 'cleric_shield_of_faith', name: 'Shield of Faith', classIds: ['cleric'], levelRequired: 5, description: 'Invoke divine protection, increasing defense by 40% for 3 rounds.', resourceType: 'mana', manaCost: 10, cooldownRounds: 0, effect: 'buff', buffStat: 'defense', buffPercent: 0.40, target: 'self', scalingFactor: 0, scalingAttr: '', basePower: 0, duration: 3, hitCount: 0, ignoresDefense: false },
  { id: 'cleric_smite', name: 'Smite', classIds: ['cleric'], levelRequired: 10, description: 'Smite your foe with holy wrath.', resourceType: 'mana', manaCost: 15, cooldownRounds: 0, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 2.0, scalingAttr: 'WIS', basePower: 10, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'cleric_divine_light', name: 'Divine Light', classIds: ['cleric'], levelRequired: 15, description: 'Bathe yourself in divine light, restoring a large amount of health.', resourceType: 'mana', manaCost: 25, cooldownRounds: 0, effect: 'heal', buffStat: '', buffPercent: 0, target: 'self', scalingFactor: 2.5, scalingAttr: 'WIS', basePower: 20, duration: 0, hitCount: 0, ignoresDefense: false },
  // Ranger
  { id: 'ranger_aimed_shot', name: 'Aimed Shot', classIds: ['ranger'], levelRequired: 1, description: 'Take careful aim for 150% DEX-scaled damage.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 3, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 1.5, scalingAttr: 'DEX', basePower: 3, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'ranger_volley', name: 'Volley', classIds: ['ranger'], levelRequired: 5, description: 'Fire a volley of arrows, hitting all enemies for 70% damage.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 4, effect: 'damage', buffStat: '', buffPercent: 0, target: 'all_enemies', scalingFactor: 0.7, scalingAttr: 'DEX', basePower: 2, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'ranger_natures_gift', name: "Nature's Gift", classIds: ['ranger'], levelRequired: 10, description: 'Call upon nature to heal 30% of your maximum HP.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 5, effect: 'heal', buffStat: '', buffPercent: 0, target: 'self', scalingFactor: 0.30, scalingAttr: 'maxhp', basePower: 0, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'ranger_pin_down', name: 'Pin Down', classIds: ['ranger'], levelRequired: 15, description: 'Pin the target down, dealing DEX damage and reducing defense by 50% for 2 rounds.', resourceType: 'cooldown', manaCost: 0, cooldownRounds: 5, effect: 'damage', buffStat: 'defense', buffPercent: -0.50, target: 'enemy', scalingFactor: 1.0, scalingAttr: 'DEX', basePower: 4, duration: 2, hitCount: 0, ignoresDefense: false, secondaryEffect: 'debuff' },
  // Druid
  { id: 'druid_wrath', name: 'Wrath', classIds: ['druid'], levelRequired: 1, description: "Call down nature's wrath on your target.", resourceType: 'mana', manaCost: 6, cooldownRounds: 0, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 1.2, scalingAttr: 'INT', basePower: 5, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'druid_rejuvenation', name: 'Rejuvenation', classIds: ['druid'], levelRequired: 1, description: 'Regenerate health over time for 3 rounds.', resourceType: 'mana', manaCost: 8, cooldownRounds: 0, effect: 'hot', buffStat: '', buffPercent: 0, target: 'self', scalingFactor: 0.8, scalingAttr: 'WIS', basePower: 4, duration: 3, hitCount: 0, ignoresDefense: false },
  { id: 'druid_entangle', name: 'Entangle', classIds: ['druid'], levelRequired: 5, description: 'Entangle the target with roots, reducing attack by 40% for 2 rounds.', resourceType: 'mana', manaCost: 10, cooldownRounds: 0, effect: 'debuff', buffStat: 'attack', buffPercent: -0.40, target: 'enemy', scalingFactor: 0, scalingAttr: '', basePower: 0, duration: 2, hitCount: 0, ignoresDefense: false },
  { id: 'druid_starfire', name: 'Starfire', classIds: ['druid'], levelRequired: 10, description: 'Call down a beam of starlight for heavy damage.', resourceType: 'mana', manaCost: 18, cooldownRounds: 0, effect: 'damage', buffStat: '', buffPercent: 0, target: 'enemy', scalingFactor: 2.0, scalingAttr: 'INT', basePower: 10, duration: 0, hitCount: 0, ignoresDefense: false },
  { id: 'druid_barkskin', name: 'Barkskin', classIds: ['druid'], levelRequired: 15, description: 'Coat yourself in bark, increasing defense by 60% for 3 rounds.', resourceType: 'mana', manaCost: 12, cooldownRounds: 0, effect: 'buff', buffStat: 'defense', buffPercent: 0.60, target: 'self', scalingFactor: 0, scalingAttr: '', basePower: 0, duration: 3, hitCount: 0, ignoresDefense: false },
];

const SKILL_BY_ID = Object.fromEntries(SKILL_CATALOG.map((s) => [s.id, s]));

const CASTER_CLASS_IDS = new Set(['mage', 'cleric', 'druid']);

/** wizard (entity class) → mage (skill classIds). */
export function normalizeClassId(classId) {
  const lower = String(classId || '').trim().toLowerCase();
  if (lower === 'wizard') return 'mage';
  return lower;
}

export function characterClassId(character) {
  const c = character?.class;
  if (!c) return '';
  if (typeof c === 'object') return c.id || c.name || '';
  return String(c);
}

export function skillById(idOrName) {
  const raw = String(idOrName || '').trim();
  if (!raw) return null;
  if (SKILL_BY_ID[raw]) return SKILL_BY_ID[raw];
  const lower = raw.toLowerCase();
  if (SKILL_BY_ID[lower]) return SKILL_BY_ID[lower];
  return SKILL_CATALOG.find((s) => s.name.toLowerCase() === lower) || null;
}

export function skillsForClass(classId) {
  const id = normalizeClassId(classId);
  if (!id) return [];
  return SKILL_CATALOG.filter((s) => (s.classIds || []).includes(id));
}

/** Mirrors pkg/entities/skills.MaxSkillSlots. */
export function maxSkillSlots(classId, level) {
  const id = normalizeClassId(classId);
  const lvl = Number(level) || 0;
  if (CASTER_CLASS_IDS.has(id)) {
    if (lvl >= 30) return 4;
    if (lvl >= 15) return 3;
    return 2;
  }
  if (lvl >= 30) return 4;
  if (lvl >= 20) return 3;
  if (lvl >= 10) return 2;
  return 1;
}

export function formatSkillCost(skill) {
  if (!skill) return '';
  if (skill.resourceType === 'mana') {
    const cost = Number(skill.manaCost) || 0;
    return `${cost} mana`;
  }
  const cd = Number(skill.cooldownRounds) || 0;
  return cd === 1 ? '1 round CD' : `${cd} round CD`;
}

export function formatSkillEffects(skill) {
  if (!skill) return [];
  const chips = [];
  const factor = Number(skill.scalingFactor) || 0;
  const duration = Number(skill.duration) || 0;
  const pct = Number(skill.buffPercent) || 0;
  const effect = String(skill.effect || '');
  const target = String(skill.target || '');
  const attr = skill.scalingAttr ? String(skill.scalingAttr) : '';

  if (effect === 'damage' && factor > 0) {
    chips.push(`${Math.round(factor * 100)}%${attr ? ` ${attr}` : ''} dmg`);
  }
  if (effect === 'heal') {
    if (attr === 'maxhp' && factor > 0) {
      chips.push(`heal ${Math.round(factor * 100)}% max HP`);
    } else {
      chips.push(factor > 0 ? `heal ${Math.round(factor * 100)}% ${attr || ''}`.trim() : 'heal');
    }
  }
  if (effect === 'dot') chips.push(duration ? `DoT ${duration} rnd` : 'DoT');
  if (effect === 'hot') chips.push(duration ? `HoT ${duration} rnd` : 'HoT');
  if (skill.buffStat) {
    if (skill.buffStat === 'stun') {
      chips.push(duration ? `stun ${duration} rnd` : 'stun');
    } else if (pct) {
      const sign = pct > 0 ? '+' : '';
      const label = String(skill.buffStat).replace(/_/g, ' ');
      chips.push(`${sign}${Math.round(pct * 100)}% ${label}${duration ? ` ${duration} rnd` : ''}`);
    } else {
      chips.push(String(skill.buffStat).replace(/_/g, ' '));
    }
  }
  if (skill.secondaryEffect === 'heal') chips.push('self heal');
  if (skill.hitCount && skill.hitCount > 1) chips.push(`${skill.hitCount} hits`);
  if (skill.ignoresDefense) chips.push('ignores armor');
  if (target === 'all_enemies') chips.push('all enemies');
  return chips;
}

export function classifySkills(classId, level, equippedIds) {
  const all = skillsForClass(classId);
  const lvl = Number(level) || 0;
  const ids = Array.isArray(equippedIds) ? equippedIds : [];
  const equippedSet = new Set(ids.map((id) => String(id)));
  const equipped = ids.map((id) => skillById(id) || { id, name: skillDisplayName(id) });
  const available = all.filter((s) => s.levelRequired <= lvl && !equippedSet.has(s.id));
  const locked = all.filter((s) => s.levelRequired > lvl);
  return { equipped, available, locked };
}

/** Bindable hotbar actions. Rest is seeded on empty/default bars; Look/Talk/Flee are optional. */
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
  const skill = skillById(id);
  const name = skill?.name || skillDisplayName(id);
  return normalizeHotbarBind({ kind: 'skill', id: skill?.id || id, name });
}

export function firstEmptyHotbarIndex(binds) {
  const slots = normalizeHotbarBinds(binds);
  return slots.findIndex((b) => b == null);
}

export function isSkillOnHotbar(binds, skillId) {
  const slots = normalizeHotbarBinds(binds);
  const skill = skillById(skillId);
  const id = String(skill?.id || skillId || '').trim();
  const name = (skill?.name || skillDisplayName(id)).toLowerCase();
  return slots.some((b) => {
    if (!b || b.kind !== 'skill') return false;
    if (id && b.id && b.id === id) return true;
    return String(b.name || '').toLowerCase() === name;
  });
}

/**
 * Place a skill on the first empty hotbar slot.
 * status: 'bound' | 'already' | 'full' | 'invalid'
 */
export function bindSkillToFirstEmptyHotbar(binds, skillId) {
  const slots = normalizeHotbarBinds(binds);
  const bind = makeSkillBind(skillId);
  if (!bind) return { binds: slots, index: -1, status: 'invalid' };
  if (isSkillOnHotbar(slots, skillId)) {
    return { binds: slots, index: -1, status: 'already' };
  }
  const idx = slots.findIndex((b) => b == null);
  if (idx < 0) {
    return { binds: slots, index: -1, status: 'full' };
  }
  slots[idx] = bind;
  return { binds: slots, index: idx, status: 'bound' };
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

function emptyHotbarSlots() {
  return Array.from({ length: HOTBAR_SLOT_COUNT }, () => null);
}

/**
 * Default hotbar for fresh prefs: Rest in slot 7, remaining slots empty.
 * Customized bars are never built from this array.
 */
export const DEFAULT_HOTBAR_BINDS = Object.freeze((() => {
  const binds = emptyHotbarSlots();
  binds[DEFAULT_REST_SLOT] = makeActionBind('rest');
  return binds;
})());

/**
 * Seed Rest onto an all-empty hotbar (fresh guest / never customized).
 * If any slot is already bound, the bar is left unchanged.
 */
export function seedRestOnEmptyHotbar(binds) {
  const out = normalizeHotbarBinds(binds);
  if (out.some((b) => b != null)) return out;
  out[DEFAULT_REST_SLOT] = makeActionBind('rest');
  return out;
}
