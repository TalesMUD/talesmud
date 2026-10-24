import test from 'node:test';
import assert from 'node:assert/strict';
import { scoreEquipped, scorePieces } from './itemScore.js';

test('empty gear scores 0', () => {
  assert.equal(scoreEquipped(null), 0);
  assert.equal(scoreEquipped({}), 0);
  assert.equal(scorePieces([]), 0);
});

test('rarity order is monotonic at the same level', () => {
  const order = ['normal', 'magic', 'rare', 'legendary', 'mythic'];
  let prev = -1;
  for (const quality of order) {
    const got = scorePieces([{ id: 'a', level: 4, quality }]);
    assert.ok(got > prev, quality);
    prev = got;
  }
});

test('a two-hand weapon stored in both slots counts once', () => {
  const sword = { id: 'sword', level: 5, quality: 'rare', attributes: { damage: 8 } };
  const once = scorePieces([sword]);
  const twice = scoreEquipped({ main_hand: sword, off_hand: sword });
  assert.equal(twice, once);
  const withShield = scoreEquipped({
    main_hand: sword,
    off_hand: { id: 'shield', level: 5, quality: 'normal', attributes: { armor: 2 } },
  });
  assert.ok(withShield > once);
});

test('unknown and broken items do not crash and do not invent a score', () => {
  const got = scorePieces([
    null,
    { quality: 'mythic', level: 40 },
    { id: ' ', quality: 'mythic' },
    {
      id: 'ok',
      quality: 'not-a-quality',
      level: 0,
      attributes: { damage: 'fast', useMessage: 'glow', defense: -3, armor: '4', notes: true },
    },
  ]);
  assert.equal(got, 14);
  assert.equal(scoreEquipped({ head: null, chest: { name: 'ghost' }, inventory: { id: 'bag', quality: 'mythic', level: 9 } }), 0);
});

test('level and stat budget use the same integers as the engine default', () => {
  const low = scorePieces([{ id: 'a', level: 1, quality: 'normal', attributes: { damage: 1 } }]);
  const high = scorePieces([{ id: 'a', level: 3, quality: 'normal', attributes: { damage: 1 } }]);
  assert.equal(low, 21);
  assert.ok(high > low);
});
