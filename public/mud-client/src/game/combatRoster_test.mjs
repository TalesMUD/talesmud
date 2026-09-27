import assert from 'node:assert/strict';
import { normalizeCombatant, mergeCombatantSnapshots } from './MUDXPlusStore.js';

const first = normalizeCombatant({ id: 'a', type: 'player', name: 'Aryn', hp: 20, maxHp: 25, classId: 'warrior', mana: 5, maxMana: 8, isAlive: true });
const enemy = normalizeCombatant({ id: 'rat', type: 'npc', name: 'Rat', hp: 9, maxHp: 9, isAlive: true });
const joined = normalizeCombatant({ id: 'b', type: 'player', name: 'Bran', hp: 15, maxHp: 20, classId: 'mage', mana: 12, maxMana: 14, isAlive: true });
const roster = mergeCombatantSnapshots([enemy], [first], [first, joined, enemy]);
assert.deepEqual(roster.players.map((c) => c.id), ['a', 'b']);
assert.deepEqual(roster.enemies.map((c) => c.id), ['rat']);
assert.equal(roster.players[1].classId, 'mage');
assert.equal(roster.players[1].mana, 12);
const hit = mergeCombatantSnapshots(roster.enemies, roster.players, [normalizeCombatant({ ...joined, hp: 3, mana: 8, isAlive: true })]);
assert.equal(hit.players[1].hp, 3);
assert.equal(hit.players[1].mana, 8);
const fled = mergeCombatantSnapshots(hit.enemies, hit.players, [normalizeCombatant({ ...joined, hasFled: true })]);
assert.equal(fled.players[1].hasFled, true);
console.log('combatRoster_test ok');
