import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { hintBlocked, hintRect } from "./guideHintLayout.js";

const here = dirname(fileURLToPath(import.meta.url));
const chip = readFileSync(join(here, "ui/GuideHint.svelte"), "utf8");

test("hint stays inside a 375 viewport and misses the room name and hotbar", () => {
  const roomName = { top: 52, left: 12, width: 220, height: 28 };
  const hotbar = { top: 700, left: 8, width: 359, height: 64 };
  const rect = hintRect({
    vw: 375,
    vh: 812,
    hotbars: [hotbar],
    roomNames: [roomName],
    width: 340,
    height: 48,
  });
  assert.ok(rect, "expected a chip");
  assert.ok(rect.top >= 8);
  assert.ok(rect.left >= 8);
  assert.ok(rect.left + rect.width <= 375 - 8);
  assert.ok(rect.top + rect.height <= 812 - 4);
  assert.ok(rect.top + rect.height <= hotbar.top);
  assert.ok(rect.top >= roomName.top + roomName.height);
});

test("hint hides when it would cover the room name or the hotbar", () => {
  const crushed = hintRect({
    vw: 375,
    vh: 812,
    hotbars: [{ top: 80, left: 0, width: 375, height: 700 }],
    roomNames: [{ top: 8, left: 0, width: 375, height: 70 }],
    width: 340,
    height: 48,
  });
  assert.equal(crushed, null);
});

test("hint hides during combat, dialog, shop, and need or greed", () => {
  assert.equal(hintBlocked({ guideHint: { text: "Head east." }, inCombat: true }), true);
  assert.equal(hintBlocked({ combatPhase: "ending" }), true);
  assert.equal(hintBlocked({ dialogActive: true }), true);
  assert.equal(hintBlocked({ shop: { name: "stall" } }), true);
  assert.equal(hintBlocked({ needGreed: { pending: true } }), true);
  for (const field of ['inventoryOverlayOpen', 'partyOverlayOpen', 'friendsOverlayOpen', 'mapOverviewOpen', 'recipes']) {
    assert.equal(hintBlocked({[field]: true}), true, field);
  }
  assert.equal(hintBlocked({ combatPhase: "idle" }), false);
});

test("chip is portaled and stays under battle, dialogs, and need or greed", () => {
  assert.equal(chip.includes("use:portal"), true);
  assert.match(chip, /class="guide-hint"/);
  assert.match(chip, /z-index:\s*var\(--layer-guide\)/);
  assert.equal(chip.includes("R0001"), false);
  assert.equal(chip.includes("ENM0001"), false);
  assert.equal(chip.includes("QST0002"), false);
  assert.equal(chip.includes("guide dismiss"), true);
});
