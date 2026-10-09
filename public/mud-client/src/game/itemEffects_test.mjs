import assert from 'assert';
import { itemEffectLines, sortLootUniqueFirst, uniqueDropName, SPECIAL_EFFECT_TEXT } from './itemEffects.js';

assert.deepStrictEqual(itemEffectLines(null), []);
assert.deepStrictEqual(itemEffectLines({ name: 'Rock' }), []);

const blade = itemEffectLines({
  onHitScriptId: 'SCR0265',
  effects: [{ trigger: 'onHit', name: 'Vigil Burn', text: '2 damage each turn for 3 turns.' }, { trigger: 'onHit' }],
});
assert.deepStrictEqual(blade, [{ trigger: 'onHit', label: 'On hit', name: 'Vigil Burn', text: '2 damage each turn for 3 turns.' }]);

const bare = itemEffectLines({ onHitScriptId: 'SCR9', onUseScriptId: 'SCR8' });
assert.strictEqual(bare.length, 2);
assert.strictEqual(bare[0].label, 'On hit');
assert.strictEqual(bare[0].text, SPECIAL_EFFECT_TEXT);
assert.strictEqual(bare[1].label, 'On use');

const equip = itemEffectLines({ effects: [{ trigger: 'onEquip', text: 'Glows.' }] });
assert.strictEqual(equip[0].label, 'While equipped');

const loot = sortLootUniqueFirst([{ name: 'Silver Mark' }, { name: 'Potion' }, { name: 'Vigil Blade', unique: true }, { name: 'Boots' }]);
assert.deepStrictEqual(loot.map((i) => i.name), ['Vigil Blade', 'Silver Mark', 'Potion', 'Boots']);
assert.deepStrictEqual(sortLootUniqueFirst(null), []);

assert.strictEqual(uniqueDropName('UNIQUE: Unmarked Vigil Blade'), 'Unmarked Vigil Blade');
assert.strictEqual(uniqueDropName('Unmarked Vigil Blade'), 'Unmarked Vigil Blade');

console.log('itemEffects ok');
