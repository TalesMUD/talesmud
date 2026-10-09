import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

// Layout B on phones: every foe (incl. mid-fight summons) shares one row
// under the HUD frames instead of wrapping into a column that climbs
// behind the top frames and the action toast.
const src = readFileSync(new URL('./ui/BattleStage.svelte', import.meta.url), 'utf8');
const blocks = src.split('@media (max-width: 768px)').slice(1);
const lb = blocks.reverse().find((b) => b.includes('.battle-stage.layout-b .enemy-strip.pack-swarm'));
assert.ok(lb, 'layout B phone block exists');
const strip = lb.slice(lb.indexOf('.battle-stage.layout-b .enemy-strip.pack-swarm'));
const rule = strip.slice(strip.indexOf('{') + 1, strip.indexOf('}'));
assert.match(rule, /flex-wrap:\s*nowrap/);
assert.match(rule, /max-width:\s*none/);
assert.doesNotMatch(rule, /220px/);
assert.doesNotMatch(rule, /bottom:\s*\d/);
assert.match(rule, /padding:\s*[\d.]+rem/, 'top padding clears the TL/TR frames');
assert.match(lb, /\.battle-stage\.layout-b \.action-banner-stack\s*\{\s*top:\s*50%/);
console.log('battleEnemyLane_test ok');
