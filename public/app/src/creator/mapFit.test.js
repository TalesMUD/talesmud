import assert from "node:assert/strict";
import test from "node:test";
import { fitViewBox, pixelToCoords, roomPixel } from "./mapFit.js";

const SCALE = 220;

function contains(box, coords) {
  const pixel = roomPixel(coords, SCALE);
  assert.ok(pixel.x >= box.x && pixel.x <= box.x + box.width, `x ${coords.x} outside ${box.x}`);
  assert.ok(pixel.y >= box.y && pixel.y <= box.y + box.height, `y ${coords.y} pixel ${pixel.y} outside ${box.y}`);
}

test("a positive-Y zone stays inside the view", () => {
  const coords = [
    { x: 0, y: 5 },
    { x: 2, y: 8 },
    { x: -1, y: 3 },
  ];
  const box = fitViewBox(coords, SCALE, 2);
  for (const point of coords) contains(box, point);
  assert.ok(box.y > 0, "positive Y sits below the origin");
});

test("a negative-Y zone stays inside the view", () => {
  const coords = [
    { x: 4, y: -2 },
    { x: 6, y: -10 },
  ];
  const box = fitViewBox(coords, SCALE, 2);
  for (const point of coords) contains(box, point);
  assert.ok(box.y < 0, "negative Y sits above the origin");
});

test("north is above south", () => {
  const south = roomPixel({ x: 0, y: 6 }, SCALE);
  const north = roomPixel({ x: 0, y: 5 }, SCALE);
  assert.ok(north.y < south.y);
  const box = fitViewBox([{ x: 0, y: 6 }, { x: 0, y: 5 }, { x: 0, y: -1 }], SCALE, 2);
  contains(box, { x: 0, y: 6 });
  contains(box, { x: 0, y: -1 });
});

test("dragging toward the top decreases Y", () => {
  const pixel = roomPixel({ x: 2, y: 5 }, SCALE);
  assert.deepEqual(pixelToCoords(pixel.x, pixel.y - SCALE, SCALE), { x: 2, y: 4 });
});

test("missing coordinates are skipped", () => {
  const box = fitViewBox([null, { x: 1, y: 1 }, {}], SCALE, 2);
  contains(box, { x: 1, y: 1 });
});

test("no coordinates keeps the default view", () => {
  assert.deepEqual(fitViewBox([], SCALE, 2), { x: -600, y: -400, width: 1200, height: 800 });
});
