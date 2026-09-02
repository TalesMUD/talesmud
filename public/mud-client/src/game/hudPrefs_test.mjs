import assert from 'assert';
import {
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
  normalizeActionBarPins,
  normalizeHotbarBinds,
  normalizeInventoryOpenMode,
  resolveHotbarActivation,
  resolvePinnedCommands,
  makeActionBind,
  skillDisplayName,
  skillGenericArtUrl,
  togglePin,
} from './hudPrefs.js';

const sayPin = PINNABLE_COMMANDS.find((c) => c.id === 'say');
assert.ok(sayPin, 'Say is in PINNABLE_COMMANDS catalog');
assert.strictEqual(sayPin.kind, 'say', 'Say uses kind say');
assert.ok(
  !DEFAULT_ACTION_BAR_PINS.includes('say'),
  'Say is default OFF the action bar'
);
assert.strictEqual(
  commandForPin(sayPin),
  null,
  'commandForPin must not emit bare say'
);

assert.deepStrictEqual(
  normalizeActionBarPins(null),
  DEFAULT_ACTION_BAR_PINS,
  'null pins → defaults'
);
assert.deepStrictEqual(
  normalizeActionBarPins([]),
  DEFAULT_ACTION_BAR_PINS,
  'empty pins → defaults'
);
assert.deepStrictEqual(
  normalizeActionBarPins(['look', 'inv', 'map', 'look', 'nope']),
  ['look', 'inv', 'map'],
  'dedupe + drop unknown'
);
assert.deepStrictEqual(
  normalizeActionBarPins(['WHO', ' Help ']),
  ['who', 'help'],
  'normalize case/whitespace'
);

const toggledOff = togglePin(['look', 'inv', 'map'], 'map');
assert.deepStrictEqual(toggledOff, ['look', 'inv'], 'unpin map');
const toggledOn = togglePin(['look', 'inv'], 'who');
assert.deepStrictEqual(toggledOn, ['look', 'inv', 'who'], 'pin who');
const sayPinned = togglePin(['look', 'inv', 'map'], 'say');
assert.deepStrictEqual(sayPinned, ['look', 'inv', 'map', 'say'], 'can pin Say');
assert.strictEqual(
  resolvePinnedCommands(sayPinned).find((c) => c.id === 'say')?.kind,
  'say',
  'pinned Say resolves with kind say'
);
assert.deepStrictEqual(
  togglePin(['look'], 'look'),
  DEFAULT_ACTION_BAR_PINS,
  'cannot clear last pin — reset to defaults'
);

const resolved = resolvePinnedCommands(['look', 'inv', 'map']);
assert.strictEqual(resolved.length, 3);
assert.strictEqual(resolved[0].id, 'look');
assert.strictEqual(resolved[1].kind, 'inventory');
assert.strictEqual(resolved[2].kind, 'map');

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

console.log('hudPrefs: pins + Say + hotbar binds/combat gate OK');

assert.strictEqual(skillGenericArtUrl('mage_fireball'), '/api/item-art/generic-spell-fire.png');
assert.strictEqual(skillGenericArtUrl('Fireball'), '/api/item-art/generic-spell-fire.png');
assert.strictEqual(skillGenericArtUrl('cleric_heal'), '/api/item-art/generic-spell-heal.png');
assert.strictEqual(skillGenericArtUrl('ranger_aimed_shot'), '/api/item-art/generic-action-ranged.png');

const lookBind = makeActionBind('look');
assert.deepStrictEqual(lookBind, { kind: 'action', id: 'look', name: 'Look', command: 'look' });
const lookAct = resolveHotbarActivation(lookBind, { inCombat: false, inventory: [] });
assert.strictEqual(lookAct.ok, true);
assert.strictEqual(lookAct.command, 'look');

console.log('hudPrefs: skill generic art + action binds OK');
