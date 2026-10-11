import assert from 'node:assert/strict';
import { get } from 'svelte/store';
import { createStore } from './MUDXPlusStore.js';

const store = createStore();
store.applyCombatStatus({
  round: 3, actorId: 'hero', actorName: 'Hero', deadlineMs: 1234,
  targetId: 'rat', decisionDeadlineMs: 1234, nextActionAtMs: 1234,
  queuedAction: 'defend', skillCooldowns: { brace: 2 },
  combatants: [
    { id: 'hero', type: 'player', name: 'Hero', hp: 10, maxHp: 20, isAlive: true },
    { id: 'rat', type: 'npc', name: 'Rat', hp: 9, maxHp: 9, isAlive: true },
  ],
});
let state = get(store);
assert.equal(state.inCombat, true);
assert.equal(state.combatPhase, 'active');
assert.deepEqual(state.combatEnemies.map(c => c.id), ['rat']);
assert.deepEqual(state.combatPlayers.map(c => c.id), ['hero']);
assert.deepEqual(state.combatTurn, { actorId: 'hero', actorName: 'Hero', round: 3, deadlineMs: 1234 });
assert.equal(state.combatTargetId, 'rat');
assert.equal(state.combatDecisionDeadlineMs, 1234);
assert.equal(state.combatQueuedAction, 'defend');
assert.deepEqual(state.combatSkillCooldowns, { brace: 2 });
store.applyCombatStatus({ skillCooldowns: {} });
state = get(store);
assert.equal(state.combatTurn.actorId, 'hero');
assert.equal(state.combatTargetId, 'rat');
assert.deepEqual(state.combatSkillCooldowns, {});
console.log('combatSnapshot_test ok');
