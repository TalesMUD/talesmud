import assert from 'assert';
import {
  adjacentPlaceIds,
  computeCamera,
  isBlockedDirLabel,
  isCurrentPlace,
  labelLodForScale,
  layoutDistance,
  layoutRoomLabels,
  panToCenterPlace,
  placesWithinBfsDepth,
  selectFramingPlaces,
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


const places = [
  { id: 'A', x: 0, y: 0 },
  { id: 'B', x: 4, y: 0 },
  { id: 'C', x: 0, y: 4 },
];
const here = places[0];
const pan = panToCenterPlace(places, here, 800, 600, 1);
assert.ok(typeof pan.panX === 'number' && typeof pan.panY === 'number');
// Here is at map min corner vs centroid → pan should push it toward center (positive for x/y in this layout)
assert.ok(pan.panX > 0 || pan.panY > 0, 'recenter offset should move corner place toward center');

// Framing: distant separateAreas must not dominate the camera when you are local.
const oldtown = [
  { id: 'O1', x: 0, y: 0 },
  { id: 'O2', x: 1, y: 0 },
  { id: 'O3', x: 0, y: 1 },
];
const meadow = [
  { id: 'M1', x: 80, y: 0 },
  { id: 'M2', x: 81, y: 0 },
  { id: 'M3', x: 80, y: 1 },
];
const world = [...oldtown, ...meadow];
const hereMeadow = meadow[0];
assert.strictEqual(layoutDistance(hereMeadow, meadow[1]), 1);
assert.ok(layoutDistance(hereMeadow, oldtown[0]) > 10);

const framed = selectFramingPlaces(world, hereMeadow, { maxDist: 10, minCount: 3 });
assert.ok(framed.every((p) => p.id.startsWith('M')), 'frame meadow cluster only');
assert.ok(framed.length >= 3);

const camAll = computeCamera(world, 800, 600, 0, 0, 1, null);
const camFocus = computeCamera(world, 800, 600, 0, 0, 1, hereMeadow);
assert.ok(camFocus.tileStep > camAll.tileStep, 'nearby frame yields larger tiles than world-fit');
assert.ok(Math.abs(camFocus.ox - 80.5) < 2, 'camera origin near meadow');

const panMeadow = panToCenterPlace(world, hereMeadow, 800, 600, 1);
// With framing, ox≈meadow → pan to center M1 should be small
assert.ok(Math.abs(panMeadow.panX) < 200, 'recenter on meadow stays local, not Oldtown offset');

const bfs = placesWithinBfsDepth(
  world,
  [
    { from: 'M1', to: 'M2', dir: 'east' },
    { from: 'M2', to: 'M3', dir: 'south' },
    { from: 'O1', to: 'O2', dir: 'east' },
  ],
  'M1',
  2
);
assert.ok(bfs.some((p) => p.id === 'M1'));
assert.ok(bfs.some((p) => p.id === 'M2'));
assert.ok(!bfs.some((p) => p.id === 'O1'), 'BFS neighborhood excludes far Oldtown');

console.log('atlasRenderer: LOD + dir labels + you-marker + framing helpers OK');

