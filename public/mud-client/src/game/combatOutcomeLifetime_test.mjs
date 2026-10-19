import test from 'node:test';
import assert from 'node:assert/strict';
import {get} from 'svelte/store';
import {createStore} from './MUDXPlusStore.js';

test('quest arrivals survive a blocked lane beyond the old queue cap and expiry', context => {
  context.mock.timers.enable({apis:['setTimeout']});
  const store = createStore();
  store.beginCombat([],[],'Fight');
  const notices = Array.from({length:6},(_,id) => ({id,type:'completed'}));
  for (const notice of notices) store.addQuestNotification(notice);
  context.mock.timers.tick(30000);
  assert.deepEqual(get(store).questNotifications,notices);
  store.clearCombat();
  assert.deepEqual(get(store).questNotifications,notices);
});

test('a dismissed outcome cannot expire the result of the next fight', context => {
  context.mock.timers.enable({apis:['setTimeout']});
  const store = createStore();
  store.beginCombat([],[],'First fight');
  store.endCombat('victory','First result');
  context.mock.timers.tick(11000);
  store.dismissCombat();
  store.beginCombat([],[],'Second fight');
  store.endCombat('defeat','Second result');
  context.mock.timers.tick(1000);
  assert.equal(get(store).combatPhase,'ending');
  assert.equal(get(store).combatOutcome,'defeat');
  context.mock.timers.tick(11000);
  assert.equal(get(store).combatPhase,'idle');
});
