import assert from 'assert';
import {
  parseExamineText,
  cleanDetailValue,
  isExamineOverlayText,
  examineResultToCard,
} from './parseExamineOverlay.js';

const medal = `=== Tarnished Medal ===

A corroded medal bearing an unfamiliar crest.

This medal was once precious—you can tell by its weight, the craftsmanship beneath the tarnish. The crest shows a symbol you don't recognize, yet it stirs something in your mind. A memory that isn't yours. When the three fragments are near each other, they seem to... hum.

--- Item Details ---
Type: Quest Item (artifact_fragment)
Quality: Normal`;

const parsed = parseExamineText(medal);
assert.ok(parsed);
assert.equal(parsed.title, 'Tarnished Medal');
assert.equal(parsed.blurb, 'A corroded medal bearing an unfamiliar crest.');
assert.match(parsed.lore, /once precious/);
assert.deepEqual(parsed.details, [
  { label: 'Type', value: 'Quest Item' },
  { label: 'Quality', value: 'Normal' },
]);
assert.ok(!parsed.details.some((d) => /artifact_fragment/.test(d.value)));

assert.equal(cleanDetailValue('Type', 'Quest Item (artifact_fragment)'), 'Quest Item');
assert.equal(cleanDetailValue('Type', 'Weapon (Sword)'), 'Weapon (Sword)');
assert.equal(isExamineOverlayText('Just a normal toast'), false);
assert.equal(isExamineOverlayText(medal), true);

const sword = `=== Iron Sword ===

A plain blade.

--- Item Details ---
Type: Weapon (Sword)
Quality: Normal
Equip Slot: Main Hand

--- Attributes ---
Damage: 5
Strength: 1`;

const s = parseExamineText(sword);
assert.equal(s.title, 'Iron Sword');
assert.equal(s.details.find((d) => d.label === 'Type').value, 'Weapon (Sword)');
assert.equal(s.attributes.find((a) => a.label === 'Damage').value, '5');

const blade = parseExamineText(`=== Unmarked Vigil Blade ===
An unmarked blade.

--- Item Details ---
Type: Weapon (Sword)

--- Effects ---
On hit — Vigil Burn: 2 damage at the start of each of the target's turns, for 3 turns.
On use — Special effect
`);
assert.deepEqual(blade.effects, [
  "On hit — Vigil Burn: 2 damage at the start of each of the target's turns, for 3 turns.",
  'On use — Special effect',
]);
assert.equal(blade.details.length, 1);
assert.deepEqual(parseExamineText(`=== Rock ===\nA rock.\n\n--- Item Details ---\nType: Junk`).effects, []);

// Structured examineResult → card model (preferred over prose parse).
const structured = examineResultToCard({
  type: 'examineResult',
  name: 'Blade',
  slot: 'main_hand',
  quality: 'magic',
  stats: { damage: 3, strength: 1 },
  value: 12,
  effects: ['On hit — Burn: 2 damage.', 'On use — Glow: Special effect'],
  description: 'A blue blade.',
  detail: 'Runes glow.',
});
assert.ok(structured);
assert.equal(structured.title, 'Blade');
assert.equal(structured.blurb, 'A blue blade.');
assert.equal(structured.lore, 'Runes glow.');
assert.deepEqual(structured.details, [
  { label: 'Quality', value: 'Magic' },
  { label: 'Equip Slot', value: 'Main Hand' },
  { label: 'Value', value: '12 gold' },
]);
assert.equal(structured.details.find((d) => d.label === 'Quality').value, 'Magic');
assert.ok(structured.attributes.some((a) => a.label === 'Damage' && a.value === '3'));
assert.deepEqual(structured.effects, [
  'On hit — Burn: 2 damage.',
  'On use — Glow: Special effect',
]);

// Empty effects omitted (empty array, not hidden inside flavour).
const noFx = examineResultToCard({
  name: 'Rock',
  quality: 'normal',
  description: 'A rock.',
  effects: [],
});
assert.ok(noFx);
assert.deepEqual(noFx.effects, []);
assert.equal(noFx.details.find((d) => d.label === 'Quality').value, 'Normal');

// Quality chrome keys for rare / legendary.
const rare = examineResultToCard({ name: 'Gem', quality: 'rare' });
assert.equal(rare.details.find((d) => d.label === 'Quality').value, 'Rare');
const legendary = examineResultToCard({ name: 'Crown', quality: 'legendary', effects: ['On use — Shine'] });
assert.equal(legendary.details.find((d) => d.label === 'Quality').value, 'Legendary');
assert.deepEqual(legendary.effects, ['On use — Shine']);

// No name → not an item card (NPC/prose-only path stays on parseExamineText).
assert.equal(examineResultToCard({ quality: 'magic', effects: ['x'] }), null);
assert.equal(examineResultToCard(null), null);
assert.equal(examineResultToCard({}), null);

console.log('parseExamineOverlay_test: ok');
