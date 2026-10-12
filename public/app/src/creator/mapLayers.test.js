import assert from "node:assert/strict";
import test from "node:test";
import { defaultLayers, MAP_LAYER_KEY, overlayLayersOn, readLayers, writeLayers } from "./mapLayers.js";

function memory() {
  const data = new Map();
  return {
    getItem: (key) => (data.has(key) ? data.get(key) : null),
    setItem: (key, value) => data.set(key, String(value)),
  };
}

test("layers default off and ignore junk", () => {
  assert.deepEqual(readLayers(null), defaultLayers());
  const store = memory();
  store.setItem(MAP_LAYER_KEY, "{");
  assert.deepEqual(readLayers(store), defaultLayers());
  store.setItem(MAP_LAYER_KEY, JSON.stringify({ level: true, nope: true, aggro: "yes" }));
  const got = readLayers(store);
  assert.equal(got.level, true);
  assert.equal(got.aggro, false);
  assert.equal(got.nope, undefined);
});

test("writes only the known toggles", () => {
  const store = memory();
  writeLayers(store, { level: true, players: true, extra: true });
  const raw = JSON.parse(store.getItem(MAP_LAYER_KEY));
  assert.equal(raw.level, true);
  assert.equal(raw.players, true);
  assert.equal(raw.reachability, false);
  assert.equal(raw.extra, undefined);
  assert.equal(overlayLayersOn(readLayers(store)), true);
  assert.equal(overlayLayersOn(defaultLayers()), false);
});
