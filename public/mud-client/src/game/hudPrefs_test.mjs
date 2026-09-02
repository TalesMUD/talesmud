import assert from 'assert';
import {
  DEFAULT_ACTION_BAR_PINS,
  DEFAULT_INVENTORY_OPEN_MODE,
  INVENTORY_OPEN_OVERLAY,
  INVENTORY_OPEN_WIDGET,
  PINNABLE_COMMANDS,
  commandForPin,
  normalizeActionBarPins,
  normalizeInventoryOpenMode,
  resolvePinnedCommands,
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

console.log('hudPrefs: pin persist helpers + inventory open mode OK');
