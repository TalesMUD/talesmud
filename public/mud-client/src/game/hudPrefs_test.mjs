import assert from 'assert';
import {
  ACTION_BAR_CHROME,
  ACTION_BAR_LAYOUT_REVISION,
  DEFAULT_ACTION_BAR_PINS,
  DEFAULT_HOTBAR_BINDS,
  DEFAULT_INVENTORY_OPEN_MODE,
  HOTBAR_SLOT_COUNT,
  INVENTORY_OPEN_OVERLAY,
  INVENTORY_OPEN_WIDGET,
  PINNABLE_COMMANDS,
  commandForPin,
  makeItemBind,
  makeSkillBind,
  migrateActionBarPins,
  normalizeActionBarPins,
  normalizeHotbarBinds,
  normalizeInventoryOpenMode,
  resolveActionBarChrome,
  resolveHotbarActivation,
  resolvePinnedCommands,
  makeActionBind,
  HOTBAR_ACTIONS,
  scrubLegacySearchBinds,
  skillDisplayName,
  skillGenericArtUrl,
  togglePin,
} from './hudPrefs.js';

assert.deepStrictEqual(DEFAULT_ACTION_BAR_PINS, [], 'Option C: no default action-bar pins');
assert.ok(ACTION_BAR_LAYOUT_REVISION >= 2, 'layout revision bumped for Option C');

assert.deepStrictEqual(
  ACTION_BAR_CHROME.map((c) => c.id),
  ['inv', 'map', 'say'],
  'INV/MAP/SAY are fixed chrome'
);
assert.ok(
  !PINNABLE_COMMANDS.some((c) => ['inv', 'map', 'say'].includes(c.id)),
  'chrome ids are not pinnable'
);
assert.ok(
  !DEFAULT_ACTION_BAR_PINS.includes('look'),
  'Look is default OFF the action bar'
);
assert.ok(
  PINNABLE_COMMANDS.some((c) => c.id === 'look'),
  'Look remains pinnable via ⋯'
);

assert.deepStrictEqual(
  normalizeActionBarPins(null),
  [],
  'null pins → empty Option C default'
);
assert.deepStrictEqual(
  normalizeActionBarPins([]),
  [],
  'empty pins stay empty'
);
assert.deepStrictEqual(
  normalizeActionBarPins(['look', 'inv', 'map', 'look', 'nope', 'say']),
  ['look'],
  'chrome ids stripped; look kept if explicitly stored'
);
assert.deepStrictEqual(
  normalizeActionBarPins(['WHO', ' Help ']),
  ['who', 'help'],
  'normalize case/whitespace'
);

assert.deepStrictEqual(
  migrateActionBarPins(['look', 'inv', 'map'], 1),
  [],
  'legacy look+inv+map defaults migrate to empty pins'
);
assert.deepStrictEqual(
  migrateActionBarPins(['look', 'inv', 'map', 'rest', 'help', 'say'], 1),
  [],
  'cluttered legacy pins migrate to empty on revision bump'
);
assert.deepStrictEqual(
  migrateActionBarPins(['look', 'who'], 2),
  ['look', 'who'],
  'revision 2+ preserves optional pins'
);

const toggledOn = togglePin([], 'who');
assert.deepStrictEqual(toggledOn, ['who'], 'pin who onto empty bar');
const toggledOff = togglePin(['who', 'help'], 'help');
assert.deepStrictEqual(toggledOff, ['who'], 'unpin help (empty allowed)');
assert.deepStrictEqual(togglePin(['who'], 'inv'), ['who'], 'cannot pin chrome inv');
assert.deepStrictEqual(togglePin([], 'look'), ['look'], 'Look remains pinnable via ⋯');

const chrome = resolveActionBarChrome();
assert.strictEqual(chrome.length, 3);
assert.strictEqual(chrome[0].kind, 'inventory');
assert.strictEqual(chrome[1].kind, 'map');
assert.strictEqual(chrome[2].kind, 'say');
assert.strictEqual(commandForPin(chrome[2]), null, 'Say chrome must not emit bare say');

const resolved = resolvePinnedCommands(['look', 'help']);
assert.strictEqual(resolved.length, 2);
assert.strictEqual(resolved[0].id, 'look');

assert.strictEqual(normalizeInventoryOpenMode(undefined), DEFAULT_INVENTORY_OPEN_MODE);
assert.strictEqual(normalizeInventoryOpenMode('overlay'), INVENTORY_OPEN_OVERLAY);
assert.strictEqual(normalizeInventoryOpenMode('widget'), INVENTORY_OPEN_WIDGET);
assert.strictEqual(normalizeInventoryOpenMode('bogus'), INVENTORY_OPEN_OVERLAY);

// --- Hotbar binds ---
assert.strictEqual(DEFAULT_HOTBAR_BINDS.length, HOTBAR_SLOT_COUNT);
assert.ok(DEFAULT_HOTBAR_BINDS.every((b) => b === null), 'default hotbar empty');

const normalized = normalizeHotbarBinds([
  { kind: 'skill', id: 'mage_fireball' },
  { kind: 'item', name: 'Health Potion', id: 'ITM0099' },
  { kind: 'nope' },
  'junk',
  null,
  { kind: 'item' },
]);
assert.strictEqual(normalized.length, HOTBAR_SLOT_COUNT);
assert.deepStrictEqual(normalized[0], {
  kind: 'skill',
  id: 'mage_fireball',
  name: 'Fireball',
});
assert.deepStrictEqual(normalized[1], {
  kind: 'item',
  name: 'Health Potion',
  id: 'ITM0099',
});
assert.strictEqual(normalized[2], null, 'junk kind → null');
assert.strictEqual(normalized[3], null, 'string junk → null');
assert.strictEqual(normalized[5], null, 'item without name/id → null');
assert.ok(normalized.slice(6).every((b) => b === null), 'pad to 8 slots');

assert.strictEqual(skillDisplayName('mage_fireball'), 'Fireball');
assert.strictEqual(skillDisplayName('warrior_power_strike'), 'Power Strike');

const skillBind = makeSkillBind('mage_fireball');
const outOfCombat = resolveHotbarActivation(skillBind, { inCombat: false, inventory: [] });
assert.strictEqual(outOfCombat.ok, false);
assert.strictEqual(outOfCombat.gated, true);
assert.match(outOfCombat.reason, /only use skills in combat/i);
assert.strictEqual(outOfCombat.command, null);

const inCombat = resolveHotbarActivation(skillBind, { inCombat: true, inventory: [] });
assert.strictEqual(inCombat.ok, true);
assert.strictEqual(inCombat.command, 'cast Fireball');

const potion = { name: 'Health Potion', type: 'consumable', templateId: 'ITM0099' };
const itemBind = makeItemBind(potion);
const itemOoC = resolveHotbarActivation(itemBind, {
  inCombat: false,
  inventory: [potion],
});
assert.strictEqual(itemOoC.ok, true, 'consumables usable out of combat');
assert.strictEqual(itemOoC.command, 'use Health Potion');

const missing = resolveHotbarActivation(itemBind, { inCombat: true, inventory: [] });
assert.strictEqual(missing.ok, false);
assert.strictEqual(missing.missing, true);
assert.strictEqual(missing.command, null);

console.log('hudPrefs: pins + chrome + hotbar binds/combat gate OK');

assert.strictEqual(skillGenericArtUrl('mage_fireball'), '/api/item-art/generic-spell-fire.png');
assert.strictEqual(skillGenericArtUrl('Fireball'), '/api/item-art/generic-spell-fire.png');
assert.strictEqual(skillGenericArtUrl('cleric_heal'), '/api/item-art/generic-spell-heal.png');
assert.strictEqual(skillGenericArtUrl('ranger_aimed_shot'), '/api/item-art/generic-action-ranged.png');

const lookBind = makeActionBind('look');
assert.deepStrictEqual(lookBind, { kind: 'action', id: 'look', name: 'Look', command: 'look' });
const lookAct = resolveHotbarActivation(lookBind, { inCombat: false, inventory: [] });
assert.strictEqual(lookAct.ok, true);
assert.strictEqual(lookAct.command, 'look');

assert.ok(
  !HOTBAR_ACTIONS.some((a) => a.id === 'search'),
  'Search hotbar action that only sent look is removed'
);
assert.strictEqual(makeActionBind('search'), null, 'Search cannot be bound');
const scrubbed = scrubLegacySearchBinds([
  { kind: 'action', id: 'search', command: 'look' },
  { kind: 'action', id: 'look', command: 'look' },
]);
assert.strictEqual(scrubbed[0], null, 'legacy Search=look bind scrubbed');
assert.strictEqual(scrubbed[1]?.id, 'look', 'Look bind preserved');

assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'look'), 'Look remains bindable');
assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'rest'), 'Rest remains bindable');
assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'talk'), 'Talk remains bindable');
assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'flee'), 'Flee remains bindable');
assert.ok(
  DEFAULT_HOTBAR_BINDS.every((b) => b === null),
  'hotbar default empty'
);
console.log('hudPrefs: Option C (room + chrome INV/MAP/SAY, no Search=look) OK');
