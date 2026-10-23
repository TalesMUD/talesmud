import assert from 'assert';
import { readFileSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const here = dirname(fileURLToPath(import.meta.url));
const root = join(here, '../../..');
const dialog = readFileSync(join(here, 'DialogOverlay.svelte'), 'utf8');
const shop = readFileSync(join(here, 'ShopOverlay.svelte'), 'utf8');
const need = readFileSync(join(here, 'NeedGreedOverlay.svelte'), 'utf8');
const hotbar = readFileSync(join(here, '../widgets/HotbarWidget.svelte'), 'utf8');
const battle = readFileSync(join(here, 'BattleStage.svelte'), 'utf8');
const index = readFileSync(join(root, 'public/index.html'), 'utf8');

function blockAfter(source, marker) {
  const at = source.indexOf(marker);
  assert.notEqual(at, -1, marker);
  return source.slice(at, at + 700);
}

assert.equal(dialog.includes('use:portal'), true);
assert.match(dialog, /\.dialog-overlay\s*\{[^}]*position:\s*fixed/s);
assert.match(dialog, /z-index:\s*var\(--layer-dialog\)/);
assert.match(dialog, /safe-area-inset-bottom/);
assert.match(dialog, /max-width:\s*768px|min-width:\s*769px/);
assert.match(dialog, /max-height:\s*500px[\s\S]*overflow-y:\s*auto/);
assert.match(dialog, /aria-label="Close"/);
assert.equal(dialog.includes('Leave conversation'), false);
assert.equal(dialog.includes('dialogclose'), false);

assert.equal(shop.includes('use:portal'), true);
assert.match(shop, /\.shop-overlay\s*\{[^}]*position:\s*fixed/s);
assert.match(shop, /z-index:\s*var\(--layer-shop\)/);
assert.match(shop, /safe-area-inset-bottom/);
assert.match(shop, /max-height:\s*500px/);
assert.match(shop, /aria-label="Close shop"/);

const compactAt = hotbar.indexOf('.hotbar-widget.compact .slot {');
assert.notEqual(compactAt, -1);
const compactEnd = hotbar.indexOf('\n  }', compactAt);
const compact = hotbar.slice(compactAt, compactEnd);
assert.match(compact, /min-width:\s*44px/);
assert.match(compact, /max-width:\s*56px/);
assert.match(compact, /flex:\s*1 1 44px/);
assert.match(hotbar, /\.hotbar-widget\.compact \.slots \{[^}]*flex-wrap:\s*wrap/s);
assert.equal(compact.includes('clamp(40px'), false);

const narrowAt = battle.indexOf('@media (max-width: 768px)');
assert.notEqual(narrowAt, -1);
const narrow = battle.slice(narrowAt);
const hot = blockAfter(narrow, '.hb-slot {');
assert.match(hot, /min-width:\s*0/);
assert.match(hot, /max-width:\s*52px/);
assert.equal(/min-width:\s*44px/.test(hot), false);
assert.match(blockAfter(narrow, '.combat-hotbar {'), /overflow:\s*hidden/);

assert.equal(need.includes('use:portal'), true);
assert.match(need, /z-index:\s*var\(--layer-roll\)/);
assert.match(need, /pointer-events:\s*none/);
assert.match(need, /\.ng-panel\s*\{[^}]*pointer-events:\s*auto/s);
assert.match(battle, /yield-roll/);
assert.match(battle, /lootRollPending/);

assert.match(index, /bundle\.js\?v=v056/);
assert.equal(index.includes('needgreed1'), false);

console.log('uxP0Layout_test ok');
