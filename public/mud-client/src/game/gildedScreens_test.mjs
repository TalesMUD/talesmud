import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { itemQualityKey, qualityColor } from './qualityColor.js';

const here = dirname(fileURLToPath(import.meta.url));
const css = readFileSync(join(here, '../../public/themes/gilded/w2-screens.css'), 'utf8');
const equipment = readFileSync(join(here, 'widgets/EquipmentWidget.svelte'), 'utf8');
const inventory = readFileSync(join(here, 'widgets/InventoryWidget.svelte'), 'utf8');
const card = readFileSync(join(here, 'widgets/ItemDetailCard.svelte'), 'utf8');
const need = readFileSync(join(here, 'ui/NeedGreedOverlay.svelte'), 'utf8');
const index = readFileSync(join(here, '../../public/index.html'), 'utf8');

assert.match(index, /bundle\.js\?v=wowf1/);
assert.match(index, /themes\/gilded\/w2-screens\.css/);
assert.equal(css.includes('Veilspan'), false);
assert.equal(css.includes('WoW'), false);
assert.equal(/grid-template-columns:\s*1fr/.test(css), false);
assert.match(css, /body\[data-theme="gilded"\] \.welcome-screen \.btn-welcome\.guest \{[^}]*color:\s*#ffd24a/);
assert.match(css, /body\[data-theme="gilded"\] \.welcome-screen \.btn-welcome\.primary \{[^}]*background:\s*#c79a45/);
assert.match(css, /body\[data-theme="gilded"\] \.welcome-screen \.btn-welcome\.primary \{[^}]*color:\s*#1a1408/);

for (const hex of ['#0b0908', '#15110e', '#ffd24a', '#3ee04a', '#ff4a3d', '#1eff00', '#0070dd', '#a335ee', '#ff8000']) {
  assert.ok(css.includes(hex), hex);
}

assert.match(equipment, /data-slot=\{slot\.key\}/);
assert.match(equipment, /data-quality=\{itemQualityKey\(item\)\}/);
assert.match(inventory, /data-quality=\{itemQualityKey\(item\)\}/);
assert.match(card, /item-card-unique/);
assert.match(card, /UNIQUE/);
assert.match(need, /rarity-unique/);
assert.match(need, /class="ng-act need"/);

assert.equal(qualityColor('magic'), 'var(--color-quality-magic, #22c55e)');
assert.equal(qualityColor('unique', 'card'), 'var(--gf-q-unique, #fde68a)');
assert.equal(qualityColor('nope'), 'var(--color-quality-normal, #9ca3af)');
assert.equal(itemQualityKey(null), 'empty');
assert.equal(itemQualityKey({ unique: true, quality: 'rare' }), 'unique');
assert.equal(itemQualityKey({ quality: 'magic' }), 'magic');
assert.equal(itemQualityKey({}), 'normal');

console.log('gildedScreens_test ok');
