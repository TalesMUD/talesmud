import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { formatAvailableAgain, optionLabel, repeatLayoutAt, repeatTag } from './questRepeat.js';

const here = dirname(fileURLToPath(import.meta.url));

assert.equal(repeatTag({ repeatReset: 'daily', repeats: true }), 'Daily');
assert.equal(repeatTag({ repeatReset: 'cooldown', repeats: true }), 'Repeats');
assert.equal(repeatTag({ repeats: true }), 'Repeats');
assert.equal(repeatTag({ repeatable: true }), '');

const now = new Date('2026-10-17T10:00:00Z');
assert.equal(
  formatAvailableAgain('2026-10-17T15:12:00Z', now),
  'Available again in 5h 12m',
);
assert.equal(formatAvailableAgain('2026-10-17T09:00:00Z', now), '');
assert.equal(formatAvailableAgain('', now), '');

assert.equal(
  optionLabel({ text: '[Quest] Alley Contract — Available again in 5h 12m', nextAvailableAt: '2026-10-18T04:00:00Z' }),
  '[Quest] Alley Contract',
);
assert.equal(optionLabel({ text: '[Quest] Alley Contract' }), '[Quest] Alley Contract');

const phone = repeatLayoutAt(375);
assert.equal(phone.tagFits, true);
assert.equal(phone.timerFits, true);
assert.equal(phone.position, 'flow');
assert.equal(phone.coversRoomName, false);
assert.equal(phone.coversDialogChrome, false);
assert.equal(phone.coversHotbar, false);
assert.equal(phone.coversNeedGreed, false);

const tablet = repeatLayoutAt(768);
const desk = repeatLayoutAt(1440);
assert.equal(tablet.tagFits && desk.tagFits, true);

const mark = readFileSync(join(here, 'ui/RepeatMark.svelte'), 'utf8');
const dialog = readFileSync(join(here, 'ui/DialogOverlay.svelte'), 'utf8');
const log = readFileSync(join(here, 'ui/QuestLog.svelte'), 'utf8');
const widget = readFileSync(join(here, 'widgets/QuestLogWidget.svelte'), 'utf8');

assert.equal(/position:\s*(fixed|absolute)/.test(mark), false);
assert.match(mark, /overflow-wrap:\s*anywhere/);
assert.match(mark, /flex-wrap:\s*wrap/);
assert.match(dialog, /RepeatMark/);
assert.match(dialog, /optionLabel/);
assert.match(dialog, /overflow-y:\s*auto/);
assert.match(dialog, /min-width:\s*0/);
assert.match(log, /RepeatMark/);
assert.match(widget, /RepeatMark/);

console.log('questRepeat_test ok');
