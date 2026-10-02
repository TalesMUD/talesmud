import assert from 'node:assert/strict';
import { createCanvas, ImageData } from 'canvas';
globalThis.ImageData = ImageData;

// mapArt.makeCanvas prefers OffscreenCanvas when present.
globalThis.OffscreenCanvas = class {
  constructor(w, h) {
    return createCanvas(w, h);
  }
};

const { buildSoftFogOverlay, paintUndergroundSoftFog } = await import('./fogOverlay.js');
const { makeCanvas } = await import('./mapArt.js');

const bounds = { minX: 0, maxX: 4, minY: 0, maxY: 4 };
const cells = [
  { x: 0, y: 0, terrain: 'grassland' },
  { x: 1, y: 0, terrain: 'grassland' },
  { x: 2, y: 0, terrain: 'fog' },
  { x: 3, y: 0, terrain: 'fog' },
  { x: 4, y: 0, terrain: 'fog' },
  { x: 2, y: 1, terrain: 'fog' },
  { x: 3, y: 1, terrain: 'grassland' },
];
const w = (bounds.maxX - bounds.minX + 1) * 32;
const h = (bounds.maxY - bounds.minY + 1) * 32;
const fog = buildSoftFogOverlay(cells, bounds, w, h);
assert.ok(fog, 'fog overlay builds when fog cells exist');
assert.equal(fog.width, w);
assert.equal(fog.height, h);

const fc = fog.getContext('2d');
const fogPx = fc.getImageData(2 * 32 + 16, 16, 1, 1).data;
const clearPx = fc.getImageData(16, 16, 1, 1).data;
assert.ok(fogPx[3] > 160, `fog cell alpha should be thick, got ${fogPx[3]}`);
assert.ok(clearPx[3] < 100, `explored cell should stay mostly clear at feather, got ${clearPx[3]}`);

assert.equal(buildSoftFogOverlay(cells.filter(c => c.terrain !== 'fog'), bounds, w, h), null);

const ug = makeCanvas(200, 200);
const uctx = ug.getContext('2d');
paintUndergroundSoftFog(
  uctx,
  [{ id: 'a', x: 1, y: 1, discovered: false }, { id: 'b', x: 3, y: 1, discovered: true }],
  (x, y) => ({ x: x * 32 + 16, y: y * 32 + 16 }),
  32,
);
const mid = uctx.getImageData(1 * 32 + 16, 1 * 32 + 16, 1, 1).data;
assert.ok(mid[3] > 100, 'undiscovered underground soft fog paints opacity');

console.log('fogOverlay: soft mask, feathered explore edge, underground wash OK');
