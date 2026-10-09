import assert from "node:assert/strict";
import test from "node:test";
import {
  decorateSpawner,
  durationLabel,
  filterSpawners,
  filterUniqueGaps,
  uniqueTemplatesWithoutSpawner,
} from "./spawnerList.js";

test("duration labels match Go-style seconds", () => {
  assert.equal(durationLabel(5 * 60 * 1e9), "5m0s");
  assert.equal(durationLabel("30m0s"), "30m0s");
  assert.equal(durationLabel(0), "");
  assert.equal(durationLabel(null), "");
});

test("spawners filter by template, room, and zone, and count the living", () => {
  const rooms = { R1: { id: "R1", name: "Yard", area: "Oldtown" }, R2: { id: "R2", name: "Pit", area: "Fen" } };
  const npcs = { WOLF: { id: "WOLF", name: "Wolf", respawnTime: 30e9 } };
  const live = [
    { templateId: "WOLF", roomId: "R1", dead: false },
    { templateId: "WOLF", roomId: "R1", dead: true },
    { templateId: "WOLF", roomId: "R2", dead: false },
  ];
  const rows = [
    decorateSpawner({ id: "S1", templateId: "WOLF", roomId: "R1", maxInstances: 2 }, { roomsById: rooms, npcsById: npcs, live }),
    decorateSpawner({ id: "S2", templateId: "WOLF", roomId: "R2", respawnTimeOverride: 600e9 }, { roomsById: rooms, npcsById: npcs, live }),
  ];
  assert.equal(rows[0].roomName, "Yard");
  assert.equal(rows[0].zone, "Oldtown");
  assert.equal(rows[0].templateName, "Wolf");
  assert.equal(rows[0].respawnLabel, "30s");
  assert.equal(rows[0].liveCount, 1);
  assert.equal(rows[1].respawnLabel, "10m0s");
  assert.deepEqual(filterSpawners(rows, { zone: "Fen" }).map((row) => row.id), ["S2"]);
  assert.deepEqual(filterSpawners(rows, { templateId: "WOLF", roomId: "R1" }).map((row) => row.id), ["S1"]);
});

test("unique templates without a spawner keep the named NPC and skip instances", () => {
  const npcs = [
    { id: "ENM0009", name: "The Hollow Knight", isTemplate: false, spawnRoomId: "R0228" },
    { id: "WOLF", name: "Wolf", isTemplate: true },
    { id: "WOLF~a", name: "Wolf", isTemplate: false, templateId: "WOLF" },
    { id: "GUARD", name: "Guard", isTemplate: false, spawnRoomId: "R1" },
  ];
  const spawners = [{ id: "S1", templateId: "GUARD", roomId: "R1" }];
  const gaps = uniqueTemplatesWithoutSpawner(npcs, spawners);
  assert.deepEqual(gaps.map((npc) => npc.id), ["ENM0009"]);
  const rooms = { R0228: { id: "R0228", area: "Sewers" } };
  assert.deepEqual(filterUniqueGaps(gaps, { zone: "Sewers", roomsById: rooms }).map((npc) => npc.id), ["ENM0009"]);
  assert.deepEqual(filterUniqueGaps(gaps, { roomId: "R1", roomsById: rooms }), []);
});
