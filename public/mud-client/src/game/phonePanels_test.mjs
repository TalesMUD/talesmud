import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import {
  PHONE_SHEET_LAUNCHERS,
  TOUCH_TARGET_PX,
  contrastRatio,
  floorPhoneFontPx,
  GILDED_PAIRS,
  hotbarSlotLayout,
  sheetIdForCommand,
} from './mobile/phoneSheets.js';

const read = (path) => readFileSync(new URL(path, import.meta.url), 'utf8');
const css = read('../../public/global.css');
const layer = (name) => Number(css.match(new RegExp(`--layer-${name}:\\s*(\\d+)`))[1]);

test('phone sheet launchers route bare panel commands and leave quest details as text', () => {
  assert.deepEqual(PHONE_SHEET_LAUNCHERS.map((row) => row.id), ['character', 'equipment', 'skills', 'quests']);
  assert.equal(sheetIdForCommand('character'), 'character');
  assert.equal(sheetIdForCommand('stats'), 'character');
  assert.equal(sheetIdForCommand('equipment'), 'equipment');
  assert.equal(sheetIdForCommand('equip'), 'equipment');
  assert.equal(sheetIdForCommand('skills'), 'skills');
  assert.equal(sheetIdForCommand('spells'), 'skills');
  assert.equal(sheetIdForCommand('quests'), 'quests');
  assert.equal(sheetIdForCommand('ql'), 'quests');
  assert.equal(sheetIdForCommand('quest'), 'quests');
  assert.equal(sheetIdForCommand('quest find the well'), null);
  assert.equal(sheetIdForCommand('attack'), null);
  assert.equal(sheetIdForCommand(''), null);
});

test('phone sheets mount the desktop character, equipment, and quest views', () => {
  const bar = read('mobile/MobileActionBar.svelte');
  const layout = read('mobile/MobileLayout.svelte');
  const quests = read('widgets/QuestLogWidget.svelte');
  assert.match(bar, /data-sheet=\{launcher\.id\}/);
  assert.match(bar, /openPhoneSheet\(sheet\)/);
  assert.match(bar, /sheetIdForCommand\(pin\.name \|\| pin\.id\)/);
  assert.match(layout, /initialTab="stats"/);
  assert.match(layout, /initialTab="skills"/);
  assert.match(layout, /<EquipmentWidget/);
  assert.match(layout, /<QuestLogWidget[\s\S]*embedded=\{true\}/);
  assert.match(quests, /class="quest-popup gf-panel"/);
  assert.match(quests, /Turn-in room/);
  assert.match(quests, /!embedded && isQuestOpen/);
});

test('phone touch targets and the sheet band stay at least 44px without covering chrome', () => {
  const sheet = read('mobile/BottomSheet.svelte');
  const hotbar = read('widgets/HotbarWidget.svelte');
  const header = read('mobile/MobileHeader.svelte');
  const players = read('ui/PlayersOverlay.svelte');
  const terminal = read('widgets/TerminalXWidget.svelte');
  const card = read('widgets/ItemDetailCard.svelte');
  assert.match(sheet, /--sheet-top/);
  assert.match(sheet, /--sheet-bottom/);
  assert.match(sheet, /min-width: 44px/);
  assert.match(sheet, /min-height: 44px/);
  assert.match(sheet, /class:roomy/);
  assert.match(hotbar, /min-width: 44px/);
  assert.match(hotbar, /flex-wrap: wrap/);
  assert.match(header, /width: 44px/);
  assert.match(header, /min-height: 44px/);
  assert.match(players, /min-height: 44px/);
  assert.match(terminal, /floorPhoneFontPx/);
  assert.match(terminal, /min-height: 44px !important/);
  assert.match(card, /calc\(var\(--layer-sheet\) \+ 40\)/);
  assert.equal(TOUCH_TARGET_PX, 44);
});

test('phone hotbar slots wrap at 375 and stay on one row in landscape', () => {
  const phone = hotbarSlotLayout(375, 9);
  assert.equal(phone.perRow, 7);
  assert.equal(phone.rows, 2);
  assert.ok(phone.width >= 44);
  assert.ok(phone.fits);
  const landscape = hotbarSlotLayout(812, 9);
  assert.equal(landscape.perRow, 9);
  assert.equal(landscape.rows, 1);
  assert.ok(landscape.fits);
});

test('phone console body text floors at 14px and larger steps stay larger', () => {
  assert.equal(floorPhoneFontPx('0.78rem', true), '14px');
  assert.equal(floorPhoneFontPx('0.92rem', true), '0.92rem');
  assert.equal(floorPhoneFontPx('1.08rem', true), '1.08rem');
  assert.equal(floorPhoneFontPx('0.78rem', false), '0.78rem');
  assert.equal(floorPhoneFontPx('13px', true), '14px');
});

test('gilded gold and green clear 4.5:1 on the dark panels', () => {
  for (const pair of GILDED_PAIRS) {
    const ratio = contrastRatio(pair.fg, pair.bg);
    assert.ok(ratio >= 4.5, `${pair.role} ${ratio.toFixed(2)}`);
  }
});

test('sheet and toast layers stay in the existing ladder', () => {
  for (const above of ['inventory', 'social', 'invite', 'sheet', 'dialog', 'shop', 'roll']) {
    assert.ok(layer(above) > layer('chrome'), above);
  }
  assert.ok(layer('sheet') > layer('inventory'));
  assert.ok(layer('battle') > layer('moment'));
  assert.ok(layer('roll') > layer('battle'));
  assert.ok(layer('dialog') > layer('sheet'));
  const index = read('../../public/index.html');
  assert.match(index, /bundle\.js\?v=v056/);
  const tags = new Set([...index.matchAll(/\?v=([\w.-]+)/g)].map((m) => m[1]));
  assert.deepEqual([...tags], ['v056']);
});
