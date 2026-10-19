import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createVisibleLifetime, notificationsBlocked } from './notificationPresentation.js';

const read = path => readFileSync(new URL(path, import.meta.url), 'utf8');
const css = read('../../../public/global.css');
const layer = name => Number(css.match(new RegExp(`--layer-${name}:\\s*(\\d+)`))[1]);

test('page layers keep chrome under inventory, party, and sheets, and rolls above combat', () => {
  for (const above of ['inventory', 'social', 'invite', 'sheet', 'dialog', 'shop', 'roll']) {
    assert.ok(layer(above) > layer('chrome'), above);
  }
  assert.ok(layer('battle') > layer('moment'));
  assert.ok(layer('roll') > layer('battle'));
  const uses = {
    '../mobile/MobileHeader.svelte': 'chrome', '../mobile/MobileTabBar.svelte': 'chrome',
    'InventoryOverlay.svelte': 'inventory', 'PartyOverlay.svelte': 'social',
    '../mobile/BottomSheet.svelte': 'sheet', 'BattleStage.svelte': 'battle',
    'NeedGreedOverlay.svelte': 'roll', 'QuestNotifications.svelte': 'moment',
  };
  for (const [file, name] of Object.entries(uses)) assert.ok(read(file).includes(`var(--layer-${name})`), file);
});

test('combat and ending outcome reserve the toast lane until rolls resolve', () => {
  for (const state of [
    {inCombat: true}, {combatPhase: 'active'}, {combatPhase: 'ending'},
    {combatOutcome: 'defeat'}, {combatPhase: 'idle', needGreed: {pending: true}},
  ]) assert.equal(notificationsBlocked(state), true);
  assert.equal(notificationsBlocked({combatPhase: 'idle', needGreed: {pending: false}}), false);
  assert.match(read('QuestNotifications.svelte'), /blocked \? \[\]/);
  assert.match(read('RoomTextOverlay.svelte'), /notificationsBlocked/);
});

function fakeClock() {
  let time = 0;
  let nextId = 0;
  const timers = new Map();
  return {
    now: () => time,
    setTimeout(fn, duration) { const id = ++nextId; timers.set(id, {at:time+duration,fn}); return id; },
    clearTimeout(id) { timers.delete(id); },
    advance(ms) {
      const end = time + ms;
      while (true) {
        const ready = [...timers].filter(([,t]) => t.at <= end).sort((a,b) => a[1].at-b[1].at)[0];
        if (!ready) break;
        time = ready[1].at;
        timers.delete(ready[0]);
        ready[1].fn();
      }
      time = end;
    },
  };
}

test('hidden notifications retain reading time and pauses retain the remaining time', () => {
  const clock = fakeClock();
  const expired = [];
  const lifetime = createVisibleLifetime(n => expired.push(n), clock);
  const queued = Array.from({length: 5}, (_,id) => ({id}));
  clock.advance(30000);
  assert.deepEqual(expired, []);
  const first = lifetime(null, {notification:queued[0], duration:5000});
  clock.advance(2000);
  first.destroy();
  clock.advance(30000);
  assert.deepEqual(expired, []);
  const resumed = lifetime(null, {notification:queued[0], duration:5000});
  clock.advance(2999);
  assert.deepEqual(expired, []);
  clock.advance(1);
  assert.deepEqual(expired, [queued[0]]);
  resumed.destroy();
  for (const notification of queued.slice(1)) {
    const shown = lifetime(null, {notification, duration:5000});
    clock.advance(5000);
    shown.destroy();
  }
  clock.advance(30000);
  assert.deepEqual(expired, queued);
});

test('short landscape shares the mobile query and keeps the home link out of the room', () => {
  const query = '(max-width: 768px), (max-width: 1024px) and (max-height: 500px) and (orientation: landscape)';
  assert.ok(read('../mobile/mobileStore.js').includes(query));
  const app = read('../../App.svelte');
  assert.ok(app.includes(`@media ${query}`));
  assert.match(app.slice(app.indexOf(`@media ${query}`)), /\.back-link\s*\{\s*display:\s*none/);
});

test('motion preference removes decorative notification and sheet animations', () => {
  for (const file of ['QuestNotifications.svelte', 'RoomTextOverlay.svelte', '../mobile/BottomSheet.svelte']) {
    assert.match(read(file), /@media \(prefers-reduced-motion: reduce\)/);
  }
});
