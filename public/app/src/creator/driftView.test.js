import assert from "node:assert/strict";
import test from "node:test";
import {
  canExportDrift,
  driftCounts,
  driftEntityPath,
  filterDrift,
  formatDriftValue,
  yamlBundle,
} from "./driftView.js";

const changes = [
  {
    type: "room",
    id: "R0001",
    name: "Gate",
    kind: "changed",
    fields: [{ path: "description", before: "Old yard", after: "New gate" }],
  },
  { type: "quest", id: "Q1", name: "Gone", kind: "removed" },
  { type: "item", id: "ITM9", name: "Lamp", kind: "added" },
];

test("drift links open the editor, including quests", () => {
  assert.equal(driftEntityPath("room", "R0001"), "/creator/rooms?id=R0001");
  assert.equal(driftEntityPath("quest", "QST0217"), "/creator/quests?id=QST0217");
  assert.equal(driftEntityPath("character_template", "CT1"), "/creator/character-templates?id=CT1");
  assert.equal(driftEntityPath("loot_table", "LT1"), "/creator/loot-tables?id=LT1");
  assert.equal(driftEntityPath("spawner", "SP1"), "/creator/spawners?id=SP1");
  assert.equal(driftEntityPath("room", ""), "");
});

test("removed entities are listed and not exported", () => {
  assert.equal(canExportDrift(changes[0]), true);
  assert.equal(canExportDrift(changes[1]), false);
  assert.equal(canExportDrift(changes[2]), true);
  assert.deepEqual(driftCounts(changes), { added: 1, changed: 1, removed: 1 });
});

test("filters match kind and before/after text", () => {
  assert.deepEqual(filterDrift(changes, { kind: "removed" }).map((row) => row.id), ["Q1"]);
  assert.deepEqual(filterDrift(changes, { query: "new gate" }).map((row) => row.id), ["R0001"]);
  assert.equal(formatDriftValue(""), "—");
  assert.equal(formatDriftValue({ a: 1 }), '{\n  "a": 1\n}');
});

test("export all is one labeled yaml document per entity", () => {
  const text = yamlBundle([
    { type: "room", id: "R0001", text: "id: R0001\nname: Gate\n" },
    { type: "item", id: "ITM9", text: "\n\nid: ITM9\n" },
    { type: "", id: "skip", text: "nope" },
  ]);
  assert.equal(text, "# room R0001\nid: R0001\nname: Gate\n---\n# item ITM9\nid: ITM9\n");
  assert.equal(yamlBundle([]), "");
});
