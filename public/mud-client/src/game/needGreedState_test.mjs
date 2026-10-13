import assert from 'assert';
import { normalizeNeedGreed } from './needGreedState.js';

assert.strictEqual(normalizeNeedGreed(null), null);
assert.strictEqual(normalizeNeedGreed({ pending: false }), null);

const model = normalizeNeedGreed({
  pending: true,
  roundId: 'ng-1',
  itemName: 'Rat Tooth',
  quality: 'magic',
  quantity: 1,
  expiresAt: 1700000030,
});
assert.deepStrictEqual(model, {
  pending: true,
  roundId: 'ng-1',
  itemName: 'Rat Tooth',
  quality: 'magic',
  quantity: 1,
  itemLabel: 'Rat Tooth',
  expiresAt: 1700000030,
});

const stacked = normalizeNeedGreed({
  pending: true,
  itemName: 'Bone Chip',
  rarity: 'rare',
  quantity: 3,
  expiresAt: 99,
});
assert.strictEqual(stacked.itemLabel, 'Bone Chip (x3)');
assert.strictEqual(stacked.quality, 'rare');
assert.strictEqual(stacked.expiresAt, 99);
assert.ok(stacked.pending);

console.log('needGreedState: normalize helpers OK');
