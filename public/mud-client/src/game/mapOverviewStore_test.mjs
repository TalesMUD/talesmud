
import assert from 'assert';

// Lightweight unit check that map overview setters return a new object identity.
// Mirrors MUDXPlusStore open/close/setMapOverviewOpen (immutable spread).

function applyOpen(state) {
  return { ...state, mapOverviewOpen: true };
}
function applyClose(state) {
  return { ...state, mapOverviewOpen: false };
}
function applySet(state, open) {
  return { ...state, mapOverviewOpen: !!open };
}

const base = { mapOverviewOpen: false, atlas: { places: [] }, inventoryOverlayOpen: false };
const opened = applyOpen(base);
assert.notStrictEqual(opened, base);
assert.strictEqual(opened.mapOverviewOpen, true);
assert.strictEqual(base.mapOverviewOpen, false);
assert.strictEqual(opened.atlas, base.atlas);

const closed = applyClose(opened);
assert.notStrictEqual(closed, opened);
assert.strictEqual(closed.mapOverviewOpen, false);

const toggled = applySet(closed, true);
assert.notStrictEqual(toggled, closed);
assert.strictEqual(toggled.mapOverviewOpen, true);

console.log('mapOverviewStore_test: ok');
