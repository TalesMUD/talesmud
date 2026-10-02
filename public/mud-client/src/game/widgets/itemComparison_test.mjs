import assert from 'node:assert/strict';
import { comparisonItems, comparisonRows, isClearUpgrade, itemUsabilityReason } from './itemComparison.js';

const warrior = { level: 5, class: { id: 'warrior', armorType: 'Plate' } };
const rogue = { level: 5, class: { id: 'rogue', armorType: 'Leather' } };
const sword = { id: 'old', name: 'Rusty Sword', slot: 'main_hand', attributes: { damage: 3 } };
const better = { id: 'new', name: 'Steel Sword', slot: 'main_hand', attributes: { damage: 6 } };
assert.equal(isClearUpgrade(better, { main_hand: sword }, warrior), true);
assert.equal(comparisonRows(better, { main_hand: sword }, warrior).find((row) => row.key === 'damage').delta, 3);
assert.equal(isClearUpgrade(sword, { main_hand: better }, warrior), false);
assert.equal(itemUsabilityReason({ tags: ['class:warrior'] }, rogue), 'Requires warrior class');
assert.equal(itemUsabilityReason({ level: 7 }, warrior), 'Requires level 7');
assert.equal(itemUsabilityReason({ type: 'armor', properties: { armorWeight: 'plate' } }, rogue), 'Requires plate armor training');
assert.equal(isClearUpgrade({ ...better, level: 7 }, { main_hand: sword }, warrior), false);
const rings = { ring1: { id: 'strong', attributes: { damage: 4 } }, ring2: { id: 'weak', attributes: { damage: 1 } } };
assert.equal(comparisonItems({ slot: 'ring1' }, rings, warrior)[0].id, 'weak');
assert.deepEqual(comparisonItems({ slot: 'ring1' }, { ring1: rings.ring1 }, warrior), []);
const twoHand = { slot: 'main_hand', subType: 'twohandsword', attributes: { damage: 9 } };
assert.equal(comparisonItems(twoHand, { main_hand: sword, off_hand: { id: 'shield', attributes: { armor: 2 } } }, warrior).length, 2);
assert.equal(comparisonRows(twoHand, { main_hand: sword, off_hand: { id: 'shield', attributes: { armor: 2 } } }, warrior).find((row) => row.key === 'damage').delta, 6);
console.log('itemComparison_test ok');

import { itemIsUsable, itemOffersUseOn, itemIsFireStarter, itemCmdName } from './itemComparison.js';

assert.equal(itemIsFireStarter({ name: 'Flint and Steel', subType: 'tool', tags: ['tool'] }), true);
assert.equal(itemOffersUseOn({ name: 'Flint and Steel', subType: 'tool', tags: ['utility'] }), true);
assert.equal(itemIsUsable({ name: 'Flint and Steel', subType: 'tool', tags: ['tool'] }), true);
assert.equal(itemIsUsable({ type: 'consumable', name: 'Potion' }), true);
assert.equal(itemIsUsable({ name: 'Rusty Dagger', type: 'weapon' }), false);
assert.equal(itemCmdName({ name: 'Dusty Torch', instanceSuffix: 'a1' }), 'Dusty Torch-a1');
console.log('item use-on helpers ok');
