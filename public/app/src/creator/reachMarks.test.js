import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { edgeLook, islandPinned, roomMark } from "./reachMarks.js";

const report = {
  rooms: [
    { id: "R1", reachable: true, instance: false },
    { id: "R2", reachable: true, instance: true },
    { id: "R3", reachable: false, instance: false },
  ],
  islands: [{ roomIds: ["R3"], reason: "no exit" }],
};

function byId(rows) {
  return new Map(rows.map((row) => [row.id, row]));
}

test("room marks follow the loaded report", () => {
  const map = byId(report.rooms);
  assert.equal(roomMark({ id: "R1" }, false, report, map), "");
  assert.equal(roomMark({ id: "R1" }, true, null, map), "");
  assert.equal(roomMark({ id: "R9" }, true, report, map), "");
  assert.equal(roomMark({ id: "R1" }, true, report, map), "reachable");
  assert.equal(roomMark({ id: "R2" }, true, report, map), "instance");
  assert.equal(roomMark({ id: "R3" }, true, report, map), "unreachable");
});

test("an island pin needs the layer and the selected index", () => {
  assert.equal(islandPinned("R3", true, 0, report), true);
  assert.equal(islandPinned("R1", true, 0, report), false);
  assert.equal(islandPinned("R3", false, 0, report), false);
  assert.equal(islandPinned("R3", true, -1, report), false);
});

test("exits use reachability colors only while the layer is on", () => {
  const map = byId(report.rooms);
  const plain = { isCrossZone: false, isCardinal: true, isHidden: false, isBidirectional: true, sourceId: "R1", targetId: "R3" };
  assert.equal(edgeLook(plain, false, report, map).color, "#888");
  const lit = edgeLook(plain, true, report, map);
  assert.equal(lit.color, "#ef4444");
  assert.equal(lit.dash, "none");
  const hidden = edgeLook({ ...plain, isHidden: true, targetId: "R1" }, true, report, map);
  assert.equal(hidden.color, "#16a34a");
  assert.equal(hidden.dash, "5 4");
  const instance = edgeLook({ ...plain, targetId: "R2" }, true, report, map);
  assert.equal(instance.color, "#6366f1");
});

test("the map template passes reachability state into the tile helpers", () => {
  const source = readFileSync(new URL("./GridWorldEditor.svelte", import.meta.url), "utf8");
  assert.match(source, /edgeLook\(edge, reachOn, reachReport, reachById\)/);
  assert.match(source, /roomMark\(room, reachOn, reachReport, reachById\)/);
  assert.match(source, /islandPinned\(room\.id, reachOn, selectedIsland, reachReport\)/);
});
