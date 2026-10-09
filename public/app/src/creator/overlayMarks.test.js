import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { overlayView } from "./overlayMarks.js";

const layers = {
  level: true,
  aggro: true,
  spawners: true,
  players: true,
  quests: true,
  art: true,
  copies: true,
};

function byId(rows) {
  return new Map(rows.map((row) => [row.id, row]));
}

const report = byId([
  {
    id: "R1",
    levelMin: 2,
    levelMax: 6,
    levelBand: "1",
    levelSource: "zone",
    aggro: 2,
    spawners: [{ id: "SP1", name: "Rat hole", respawnTime: "5m0s" }],
    players: 1,
    playerNames: ["Ada"],
    quests: ["Q1", "Q2"],
    missingArt: true,
    copies: 3,
  },
  {
    id: "R2",
    players: 2,
    copies: 1,
    spawners: [{ id: "SP2", templateId: "RAT" }],
  },
]);

test("a tile stays plain until its layer is on", () => {
  const off = overlayView({ id: "R1" }, { level: false, aggro: false }, report);
  assert.equal(off.levelBand, "");
  assert.equal(off.badges.length, 0);
  assert.equal(off.missingArt, false);
  assert.equal(overlayView({ id: "R9" }, layers, report).title, "");
});

test("badges carry the hover text and hide names the payload left out", () => {
  const view = overlayView({ id: "R1" }, layers, report);
  assert.equal(view.levelBand, "1");
  assert.equal(view.missingArt, true);
  assert.match(view.title, /Levels 2–6 \(zone\)/);
  assert.match(view.title, /2 aggressive/);
  assert.match(view.title, /Rat hole 5m0s/);
  assert.match(view.title, /1 online: Ada/);
  assert.match(view.title, /Quests: Q1, Q2/);
  assert.match(view.title, /Missing art/);
  assert.match(view.title, /3 live instance copies/);
  const kinds = view.badges.map((badge) => badge.kind);
  assert.deepEqual(kinds, ["aggro", "spawner", "players", "quest", "copies"]);

  const counts = overlayView({ id: "R2" }, layers, report);
  assert.match(counts.title, /2 online/);
  assert.doesNotMatch(counts.title, /online:/);
  assert.match(counts.title, /RAT no respawn/);
  assert.match(counts.title, /1 live instance copy/);
});

test("the map template passes layer state into the overlay helper", () => {
  const source = readFileSync(new URL("./GridWorldEditor.svelte", import.meta.url), "utf8");
  assert.match(source, /overlayView\(room, mapLayers, overlayById\)/);
  assert.match(source, /reasonChain\(island\.reason, reachRoomIds\)/);
  assert.match(source, /readLayers\(localStorage\)/);
});
