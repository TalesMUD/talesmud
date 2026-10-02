import assert from 'assert';
import { parseExamineText, cleanDetailValue, isExamineOverlayText } from './parseExamineOverlay.js';

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

console.log('parseExamineOverlay_test: ok');
