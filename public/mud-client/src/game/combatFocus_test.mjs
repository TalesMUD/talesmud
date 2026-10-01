import assert from 'node:assert/strict';
import { createStore } from './MUDXPlusStore.js';
import { livingFocus } from './combatFocus.js';

const store = createStore();
let state;
store.subscribe((next) => { state = next; });
const enemies = [
  { id: 'rat', name: 'Rat', type: 'npc', hp: 12, maxHp: 12, threat: 'yellow' },
  { id: 'ogre', name: 'Ogre', type: 'npc', hp: 30, maxHp: 30, threat: 'red' },
  { id: 'lich', name: 'Lich', type: 'npc', hp: 40, maxHp: 40, threat: 'skull' },
];
store.beginCombat(enemies, [], 'Fight', 'rat');
assert.equal(state.combatTargetId, 'rat');
assert.equal(state.combatThreatWarning, null);
store.setCombatTarget('ogre');
assert.equal(state.combatTargetId, 'ogre');
assert.match(state.combatThreatWarning.text, /Ogre is much stronger than you/);
assert.equal(state.combatThreatWarning.tier, 'red');
const warning = state.combatThreatWarning;
store.setCombatTarget('ogre');
assert.equal(state.combatThreatWarning, warning, 're-click must not re-warn');
store.setCombatTarget('lich');
assert.equal(state.combatThreatWarning.tier, 'skull');
assert.match(state.combatThreatWarning.text, /Lich is much stronger than you/);
store.applyCombatAction({ action: 'attack', targetId: 'lich', remainingHp: 0 });
assert.equal(state.combatTargetId, 'rat', 'dead focus falls back to a living hostile');
assert.equal(state.combatThreatWarning, null);
assert.equal(livingFocus(state.combatEnemies, 'lich').id, 'rat');
store.applyCombatAction({ action: 'focus', targetId: 'ogre' });
assert.equal(state.combatTargetId, 'ogre', 'room attack syncs focus');
assert.equal(state.combatThreatWarning.tier, 'red');
store.endCombat('victory', 'Done');
assert.equal(state.combatTargetId, null);
assert.equal(state.combatThreatWarning, null);
console.log('combatFocus_test ok');
