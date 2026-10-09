import assert from "node:assert/strict";
import test from "node:test";
import { fitViewBox, roomPixel } from "./mapFit.js";

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
  assert.ok(box.y < 0, "north-up tiles sit above the origin");
});

test("a negative-Y zone stays inside the view", () => {
  const coords = [
    { x: 4, y: -2 },
    { x: 6, y: -10 },
  ];
  const box = fitViewBox(coords, SCALE, 2);
  for (const point of coords) contains(box, point);
});

test("missing coordinates are skipped", () => {
  const box = fitViewBox([null, { x: 1, y: 1 }, {}], SCALE, 2);
  contains(box, { x: 1, y: 1 });
});

test("no coordinates keeps the default view", () => {
  assert.deepEqual(fitViewBox([], SCALE, 2), { x: -600, y: -400, width: 1200, height: 800 });
});
