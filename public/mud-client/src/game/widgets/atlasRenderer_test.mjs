import assert from 'assert';
import {
  adjacentPlaceIds,
  isBlockedDirLabel,
  isCurrentPlace,
  labelLodForScale,
  layoutRoomLabels,
} from './atlasRenderer.js';

assert.strictEqual(labelLodForScale(0.5), 'area');
assert.strictEqual(labelLodForScale(1.0), 'near');
assert.strictEqual(labelLodForScale(1.5), 'all');

assert.ok(isBlockedDirLabel('south'));
assert.ok(isBlockedDirLabel('EAST'));
assert.ok(isBlockedDirLabel('up'));
assert.ok(isBlockedDirLabel('down'));
assert.ok(isBlockedDirLabel('outside'));
assert.ok(isBlockedDirLabel('entrance'));
assert.ok(!isBlockedDirLabel('residence'));
assert.ok(!isBlockedDirLabel('cellar stair'));

assert.ok(isCurrentPlace('R0215', 'R0215'));
assert.ok(isCurrentPlace('R0215', 'R0215~guest-a'));
assert.ok(!isCurrentPlace('R0215', 'R0203'));
assert.ok(!isCurrentPlace('R0215~x', 'R0215'));

const near = adjacentPlaceIds(
  [
    { from: 'R0215', to: 'R0203', dir: 'up' },
    { from: 'R0215', to: 'R0230', dir: 'deeper' },
    { from: 'R0101', to: 'R0102', dir: 'north' },
  ],
  'R0215~guest-a',
  [{ id: 'R0215' }]
);
assert.ok(near.has('R0203'));
assert.ok(near.has('R0230'));
assert.ok(!near.has('R0102'));

const measure = (text) => ({ w: text.length * 6, h: 12 });
const placed = layoutRoomLabels(
  [
    { text: 'The Weary Wanderer', px: 100, py: 100, font: '10px serif', force: true },
    { text: 'Guest Room', px: 105, py: 102, font: '9px serif', force: false },
    { text: "Galdric's Room", px: 108, py: 104, font: '9px serif', force: false },
  ],
  measure
);
assert.ok(placed.length >= 1, 'current label always placed');
assert.ok(
  placed.every((a, i) =>
    placed.slice(i + 1).every((b) => {
      const A = { x: a.x, y: a.y, w: a.w, h: a.h };
      const B = { x: b.x, y: b.y, w: b.w, h: b.h };
      return A.x + A.w < B.x || B.x + B.w < A.x || A.y + A.h < B.y || B.y + B.h < A.y;
    })
  ),
  'placed labels do not overlap'
);

console.log('atlasRenderer: LOD + dir labels + you-marker helpers OK');
