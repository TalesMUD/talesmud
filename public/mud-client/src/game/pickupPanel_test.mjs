import assert from 'assert';
import { itemArtSrc } from './itemArtSrc.js';
import {
  groupGroundItems,
  groundItemHint,
  isCurrencyItem,
  pickupAllCommands,
  pickupEntryCount,
  pickupKeyAction,
  pickupRowCommands,
  shouldClosePickup,
  showPickupAll,
} from './pickupPanel.js';

const plain = { tagName: 'DIV', isContentEditable: false, closest: () => null };
const typing = { tagName: 'INPUT', isContentEditable: false, closest: () => null };

const ground = [
  { id: 'm1', name: 'Silver Mark' },
  { id: 'm2', name: 'silver mark' },
  { id: 'pot', name: 'Weak Health Potion', type: 'consumable', attributes: { healthRestore: 25 } },
  { id: 'herb', name: 'Meadow Herbs', type: 'crafting_material' },
  { id: 'blade', name: 'Vigil Blade', unique: true, quality: 'rare' },
  { id: 'rock', name: 'Loose Rock', noPickup: true },
  { id: 'torch', name: 'Dusty Torch', type: 'consumable', quantity: 3 },
  { id: 'coin', name: 'Gold', type: 'currency', quantity: 12 },
];

assert.equal(isCurrencyItem({ name: 'Silver Mark' }), true);
assert.equal(isCurrencyItem({ name: 'gold' }), true);
assert.equal(isCurrencyItem({ name: 'Gold Coin' }), true);
assert.equal(isCurrencyItem({ type: 'currency', name: 'Old Token' }), true);
assert.equal(isCurrencyItem({ name: 'Thornback Claw' }), false);
assert.equal(isCurrencyItem({ name: 'Golden Thorn' }), false);
assert.equal(isCurrencyItem({ name: 'Marigold' }), false);

assert.equal(groundItemHint({ name: 'Thornback Claw' }), '');
assert.equal(
  groundItemHint({ name: 'Weak Health Potion', type: 'consumable', attributes: { healthRestore: 25 } }),
  'Consumable · Heals 25 HP'
);
assert.equal(groundItemHint({ name: 'Meadow Herbs', type: 'crafting_material' }), 'Material');
assert.equal(groundItemHint({ name: 'Silver Mark' }), 'Currency');
assert.equal(
  groundItemHint({ name: 'Fang', type: 'weapon', effects: [{ trigger: 'onHit', text: 'Bleeds' }] }),
  'Weapon · Bleeds'
);

const rows = groupGroundItems(ground);
assert.deepEqual(rows.map((row) => row.name), [
  'Silver Mark',
  'Weak Health Potion',
  'Meadow Herbs',
  'Vigil Blade',
  'Dusty Torch',
  'Gold',
]);
assert.equal(rows[0].count, 2);
assert.equal(rows[0].currency, true);
assert.equal(rows[0].hint, 'Currency');
assert.equal(rows[0].badge, '1');
assert.deepEqual(rows[0].ids, ['m1', 'm2']);
assert.equal(itemArtSrc(rows[0].artItem), '/api/item-art/generic-currency.png');
assert.equal(rows[1].hint, 'Consumable · Heals 25 HP');
assert.equal(rows[1].rarity, 'normal');
assert.equal(rows[2].hint, 'Material');
assert.equal(rows[3].rarity, 'unique');
assert.equal(rows[3].unique, true);
assert.equal(rows[4].count, 3);
assert.equal(rows[4].ids.length, 1);
assert.equal(rows[5].count, 12);
assert.equal(rows[5].currency, true);
assert.equal(pickupEntryCount(ground), 7);
assert.equal(showPickupAll(ground), true);
assert.ok(!rows.some((row) => row.name === 'Loose Rock'));

const templateArt = groupGroundItems([{ id: 'uuid-1', templateId: 'ITM0015', name: 'Silver Mark', type: 'currency' }]);
assert.equal(itemArtSrc(templateArt[0].artItem), '/api/item-art/ITM0015.png');
const instanceArt = groupGroundItems([{ id: 'ITM0001~abcd', name: 'Dusty Torch', type: 'consumable' }]);
assert.equal(itemArtSrc(instanceArt[0].artItem), '/api/item-art/ITM0001.png');

const one = pickupRowCommands(ground, rows[0], []);
assert.equal(one.length, 1);
assert.equal(one[0].command, 'pickup Silver Mark');
const second = pickupRowCommands(ground, rows[0], [one[0].id]);
assert.equal(second.length, 1);
assert.notEqual(second[0].id, one[0].id);
assert.equal(second[0].command, 'pickup silver mark');
assert.deepEqual(pickupRowCommands(ground, rows[0], ['m1', 'm2']), []);

const all = pickupAllCommands(ground, ['m2', 'gone']);
assert.deepEqual(all.map((cmd) => cmd.id), ['m1', 'pot', 'herb', 'blade', 'torch', 'coin']);
assert.equal(all.filter((cmd) => cmd.id === 'm2').length, 0);
const stillThere = ground.filter((item) => item.id !== 'm2' && item.id !== 'rock');
const again = pickupAllCommands(stillThere, ['m2']);
assert.deepEqual(again.map((cmd) => cmd.id), ['m1', 'pot', 'herb', 'blade', 'torch', 'coin']);
assert.equal(pickupAllCommands(stillThere, again.map((cmd) => cmd.id)).length, 0);

assert.equal(showPickupAll([{ id: 'a', name: 'Silver Mark', quantity: 4 }]), false);
assert.equal(groupGroundItems([{ id: 'a', name: 'Silver Mark', quantity: 4 }])[0].count, 4);
assert.equal(pickupAllCommands([{ id: 'a', name: 'Silver Mark', quantity: 4 }], []).length, 1);

assert.equal(shouldClosePickup([]), true);
assert.equal(shouldClosePickup([{ id: 'rock', name: 'Loose Rock', noPickup: true }]), true);
assert.equal(shouldClosePickup([{ id: 'a', name: 'Silver Mark' }]), false);
assert.equal(shouldClosePickup(null), true);

assert.deepEqual(pickupKeyAction({ key: '1', target: plain }, rows.length), { type: 'row', index: 0 });
assert.deepEqual(pickupKeyAction({ key: '6', target: plain }, rows.length), { type: 'row', index: 5 });
assert.equal(pickupKeyAction({ key: '7', target: plain }, rows.length), null);
assert.equal(pickupKeyAction({ key: '9', target: plain }, 2), null);
assert.equal(pickupKeyAction({ key: '0', target: plain }, 3), null);
assert.deepEqual(pickupKeyAction({ key: 'Enter', target: plain }, 1), { type: 'all' });
assert.deepEqual(pickupKeyAction({ key: 'Escape', target: plain }, 1), { type: 'close' });
assert.equal(pickupKeyAction({ key: '1', target: typing }, 3), null);
assert.equal(pickupKeyAction({ key: 'Enter', target: typing }, 3), null);
assert.equal(pickupKeyAction({ key: '1', target: plain, metaKey: true }, 3), null);
assert.equal(pickupKeyAction({ key: 'k', target: plain }, 3), null);

const many = groupGroundItems(Array.from({ length: 11 }, (_, i) => ({ id: `i${i}`, name: `Item ${i}` })));
assert.equal(many[8].badge, '9');
assert.equal(many[9].badge, '');
assert.equal(pickupKeyAction({ key: '9', target: plain }, many.length).index, 8);

console.log('pickupPanel_test ok');
