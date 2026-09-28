import assert from 'node:assert/strict';
import { get } from 'svelte/store';
import { layoutStore } from './LayoutStore.js';
import { createClient } from '../Client.js';
import { createStore } from '../MUDXPlusStore.js';
import { LAYOUT_STORAGE_KEY } from './layoutTemplates.js';

globalThis.window = { innerWidth: 1920, innerHeight: 1080 };
const storage = new Map();
globalThis.localStorage = { getItem: k => storage.get(k), setItem: (k, v) => storage.set(k, v) };
const widgets = () => structuredClone(get(layoutStore).widgets);
const savedWidgets = () => JSON.parse(storage.get(LAYOUT_STORAGE_KEY)).widgets;
const geometry = items => items.map(w => ({
  id: w.id, x: w[24].x, y: w[24].y, w: w[24].w, h: w[24].h,
  tabs: w.tabs, activeTabIndex: w.activeTabIndex,
}));

for (const kind of ['desktop', 'wide', 'compact']) {
  layoutStore.applyPreset(kind);
  const normal = widgets();
  layoutStore.syncCombatFocus('active');
  assert.equal(get(layoutStore).focusId, 'battle-stage');
  assert.deepEqual(widgets(), normal, 'cover preserves terminal and tab geometry');
  layoutStore.syncCombatFocus('active'); // repeated starts / ally joins
  layoutStore.saveToStorage();
  assert.ok(savedWidgets().every(w => w.id !== 'battle-stage'));
  const template = layoutStore.saveAsTemplate(`Combat ${kind}`);
  assert.deepEqual(JSON.parse(JSON.stringify(get(layoutStore).templates.find(t => t.id === template.id).widgets)), savedWidgets());
  layoutStore.syncCombatFocus('ending');
  assert.equal(get(layoutStore).focusId, 'battle-stage', 'outcome retains cover');
  layoutStore.syncCombatFocus('idle'); // dismiss, timeout, defeat, or combatLeave
  assert.equal(get(layoutStore).focusId, null);
  assert.deepEqual(widgets(), normal, 'previous arrangement restored');
  // The same cover also restores on a direct leave without an outcome.
  layoutStore.syncCombatFocus('active');
  layoutStore.syncCombatFocus('idle');
  assert.deepEqual(widgets(), normal);
}

layoutStore.applyPreset('desktop');
const normal = widgets();
layoutStore.toggleFocus('room-1');
const manual = widgets();
layoutStore.syncCombatFocus('active');
layoutStore.enterEditMode();
layoutStore.exitEditMode(true);
assert.equal(get(layoutStore).focusId, 'battle-stage', 'saving does not end combat cover');
assert.equal(savedWidgets().find(w => w.id === 'room-1').w, normal.find(w => w.id === 'room-1').w);
layoutStore.toggleFocus('battle-stage');
assert.equal(get(layoutStore).focusId, 'battle-stage', 'Escape cannot discard an active combat cover');
layoutStore.syncCombatFocus('idle');
assert.equal(get(layoutStore).focusId, 'room-1');
assert.deepEqual(geometry(widgets()), geometry(manual), 'prior manual focus restored');
layoutStore.toggleFocus('room-1');
assert.deepEqual(widgets().map(w => w[24]), normal.map(w => w[24]));

// Resize while covered is temporary; saved rows survive, restored view fits.
layoutStore.saveToStorage();
layoutStore.loadFromStorage();
layoutStore.saveToStorage();
const saved = savedWidgets();
layoutStore.syncCombatFocus('active');
window.innerHeight = 768;
layoutStore.onViewportResize();
layoutStore.saveToStorage();
assert.deepEqual(savedWidgets(), saved);
layoutStore.syncCombatFocus('idle');
assert.ok(widgets().every(w => w[24].y + w[24].h <= 17));
console.log('combatFocus_test ok');

// Drive the actual client handlers, including the joiner's combatStart and leave.
{
  const play = createStore();
  const unsubscribe = play.subscribe(state => layoutStore.syncCombatFocus(state.combatPhase));
  let receive;
  const client = createClient(() => {}, () => {}, play);
  client.setWSClient({ addEventListener(type, handler) { if (type === 'message') receive = handler; } });
  const message = msg => receive({ data: JSON.stringify(msg) });
  message({ type: 'combatStart', enemies: [{ id: 'rat', hp: 10, maxHp: 10 }], players: [{ id: 'self', hp: 20, maxHp: 20 }], message: 'Joined combat.' });
  assert.equal(get(layoutStore).focusId, 'battle-stage');
  message({ type: 'combatLeave', message: 'You leave combat.' });
  assert.equal(get(play).combatPhase, 'idle');
  assert.equal(get(layoutStore).focusId, null);
  unsubscribe();
}
console.log('combat client enter/leave ok');
