import assert from "node:assert/strict";
import test from "node:test";
import { BADGE_SCREEN_PX, layoutOverlayBadges } from "./badgeLayout.js";

// A fitted zone (about 39 rooms on a 1600px window) makes a 180×80 tile
// roughly 45–60px wide. Badges stay 12px; they do not shrink with the map.
const TILE = { tileWidth: 180, tileHeight: 80 };

const badges = [
  { kind: "aggro", text: "2", title: "2 aggressive" },
  { kind: "spawner", text: "S", title: "Rat hole 5m0s" },
  { kind: "players", text: "1", title: "1 online" },
  { kind: "quest", text: "Q×2", title: "Q1, Q2" },
  { kind: "copies", text: "~3", title: "3 live instance copies" },
];

function layout(svgPerPx) {
  return layoutOverlayBadges(badges, { ...TILE, svgPerPx });
}

function insideTile(result, svgPerPx) {
  const tilePxW = TILE.tileWidth / svgPerPx;
  for (const badge of result.badges) {
    assert.equal(badge.height, BADGE_SCREEN_PX);
    assert.ok(badge.x + badge.width / 2 <= 0, "badge hangs off the right edge");
    assert.ok(badge.x - badge.width / 2 >= -tilePxW, "badge is wider than the tile");
  }
}

test("badges stay 12px when the tile is large", () => {
  const result = layout(1);
  assert.equal(result.scale, 1);
  assert.deepEqual(result.badges.map((badge) => badge.kind), ["aggro", "spawner", "players", "quest", "copies"]);
  insideTile(result, 1);
});

test("a fitted zone counter-scales and merges the overflow", () => {
  const result = layout(3);
  assert.equal(result.scale, 3);
  assert.ok(result.badges.length < badges.length);
  assert.equal(result.badges.at(-1).kind, "more");
  assert.match(result.badges.at(-1).text, /^\+\d+$/);
  assert.match(result.badges.at(-1).title, /Q1, Q2/);
  assert.match(result.badges.at(-1).title, /3 live instance copies/);
  insideTile(result, 3);
});

test("zooming in does not grow the badge past 12px", () => {
  const result = layout(0.25);
  assert.equal(result.scale, 0.25);
  assert.equal(result.badges.length, badges.length);
  insideTile(result, 0.25);
});

test("a tile shorter than the badge hides the row", () => {
  const result = layout(8);
  assert.equal(result.scale, 0);
  assert.deepEqual(result.badges, []);
});

test("an empty list draws nothing", () => {
  assert.deepEqual(layoutOverlayBadges([], { ...TILE, svgPerPx: 1 }), { scale: 0, badges: [] });
});
