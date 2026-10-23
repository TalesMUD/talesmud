import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import {
  DESKTOP_GROUND_MIN_WIDTH,
  boxesIntersect,
  placeGroundList,
} from './groundListLayout.js';

const read = (name) => readFileSync(new URL(name, import.meta.url), 'utf8');

/** Enemy art along the bottom of the left room column (14 of 24). */
function bottomEnemies(viewportWidth, viewportHeight, count) {
  const roomRight = Math.floor(viewportWidth * 14 / 24);
  const cardW = 128;
  const cardH = 140;
  const gap = 8;
  const boxes = [];
  let x = 12;
  let y = viewportHeight - 240;
  for (let index = 0; index < count; index += 1) {
    if (x + cardW > roomRight - 4) {
      x = 12;
      y -= cardH + gap;
    }
    boxes.push({ x, y, width: cardW, height: cardH });
    x += cardW + gap;
  }
  return boxes;
}

/** The pre-fix modal: 440px, centered. It cuts through a full enemy row. */
function centeredList(viewportWidth, viewportHeight) {
  const width = Math.min(440, viewportWidth);
  const height = Math.min(560, Math.floor(viewportHeight * 0.72));
  return {
    x: Math.floor((viewportWidth - width) / 2),
    y: Math.floor((viewportHeight - height) / 2),
    width,
    height,
  };
}

test('desktop ground list misses enemy art at 1280, 1440, and 1920', () => {
  const viewports = [[1280, 800], [1440, 900], [1920, 1080]];
  for (const [width, height] of viewports) {
    const centered = centeredList(width, height);
    const roomRight = Math.floor(width * 14 / 24);
    const overlappingRow = [];
    for (let x = 12; x + 128 <= roomRight; x += 136) {
      overlappingRow.push({ x, y: centered.y + 24, width: 128, height: 150 });
    }
    assert.ok(overlappingRow.length >= 3, `${width} row`);
    assert.equal(
      overlappingRow.some((box) => boxesIntersect(centered, box)),
      true,
      `${width} centered should hit`,
    );
    const dockedRow = placeGroundList(width, height, overlappingRow, 900);
    assert.ok(dockedRow, `${width} dock`);
    for (const box of overlappingRow) {
      assert.equal(boxesIntersect(dockedRow, box), false, `${width} full row`);
    }
    for (const count of [1, 3, 6, 8]) {
      const enemies = bottomEnemies(width, height, count);
      const placed = placeGroundList(width, height, enemies, 900);
      assert.ok(placed, `${width}x${height} x${count}`);
      assert.equal(placed.width, 380);
      for (const box of enemies) {
        assert.equal(boxesIntersect(placed, box), false, `${width}x${height} enemies=${count}`);
      }
    }
  }
});

test('a long item name does not change the desktop list box', () => {
  const enemies = bottomEnemies(1440, 900, 6);
  const shortList = placeGroundList(1440, 900, enemies);
  const longList = placeGroundList(1440, 900, enemies, 4000);
  assert.deepEqual(longList, shortList);
  assert.equal(shortList.width, 380);
});

test('an enemy covering the top-right dock still leaves a clear list', () => {
  const blocked = placeGroundList(1440, 900, [], 560);
  const enemies = [{ ...blocked }];
  const placed = placeGroundList(1440, 900, enemies);
  assert.ok(placed);
  assert.equal(boxesIntersect(placed, enemies[0]), false);
});

test('phone and tablet keep the existing ground list placement', () => {
  const enemies = bottomEnemies(1024, 768, 6);
  assert.equal(placeGroundList(1024, 768, enemies), null);
  assert.equal(placeGroundList(375, 812, bottomEnemies(375, 812, 3)), null);
  assert.equal(placeGroundList(DESKTOP_GROUND_MIN_WIDTH - 1, 800, []), null);
  assert.ok(placeGroundList(DESKTOP_GROUND_MIN_WIDTH, 800, []));
});

test('the pickup panel uses the desktop ground-list placement', () => {
  const panel = read('ui/PickupPanel.svelte');
  assert.match(panel, /placeGroundList/);
  assert.match(panel, /ENEMY_ART_SELECTOR/);
  assert.match(panel, /data-ground-list/);
  assert.match(panel, /desktop-dock/);
});
