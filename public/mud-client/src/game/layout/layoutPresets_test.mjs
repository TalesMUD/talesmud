import assert from 'assert';
import {
  ACTION_DOCK_H,
  GRID_COLS,
  ROW_HEIGHT,
  foldDockedHotbar,
  presetWidgets,
  viewportRows,
} from './layoutPresets.js';

function byType(widgets, type) {
  return widgets.find((w) => w.widgetType === type);
}

for (const height of [768, 1080, 1440]) {
  const rows = viewportRows(height);
  const stack = rows * ROW_HEIGHT;
  assert.ok(stack <= height - 64, `${height} stack ${stack} overflows the grid`);
  assert.ok(stack > height - 64 - ROW_HEIGHT, `${height} leaves a tall empty band`);
  const desktop = presetWidgets('desktop', height);
  assert.equal(byType(desktop, 'hotbar'), undefined);
  const action = byType(desktop, 'actionbar');
  assert.equal(action.x, 0);
  assert.equal(action.w, GRID_COLS);
  assert.equal(action.h, ACTION_DOCK_H);
  assert.equal(action.y + action.h, rows);
  const room = byType(desktop, 'room');
  const term = byType(desktop, 'terminal');
  assert.equal(room.y, 0);
  assert.equal(term.y, 0);
  assert.equal(room.h, action.y);
  assert.ok(room.y + room.h <= action.y);
}

{
  const compact = presetWidgets('compact', 900);
  const room = byType(compact, 'room');
  const term = byType(compact, 'terminal');
  assert.ok(room.h > 0 && term.h > 0);
  assert.equal(term.y, room.h);
  assert.equal(byType(compact, 'hotbar'), undefined);
}

{
  const folded = foldDockedHotbar([
    { id: 'room-1', widgetType: 'room', x: 0, y: 0, w: 12, h: 12, visible: true },
    { id: 'terminal-1', widgetType: 'terminal', x: 12, y: 0, w: 12, h: 12, visible: true },
    { id: 'hotbar-1', widgetType: 'hotbar', x: 0, y: 12, w: 24, h: 2, visible: true },
    { id: 'actionbar-1', widgetType: 'actionbar', x: 0, y: 14, w: 24, h: 3, visible: true },
  ]);
  assert.equal(byType(folded, 'hotbar'), undefined);
  const action = byType(folded, 'actionbar');
  assert.equal(action.y, 12);
  assert.equal(action.h, 5);
  assert.equal(byType(folded, 'room').h, 12);
}

{
  const custom = [
    { id: 'hotbar-1', widgetType: 'hotbar', x: 0, y: 0, w: 8, h: 2, visible: true },
    { id: 'actionbar-1', widgetType: 'actionbar', x: 0, y: 10, w: 24, h: 3, visible: true },
  ];
  const folded = foldDockedHotbar(custom);
  assert.equal(folded.length, 2);
  assert.equal(byType(folded, 'hotbar').w, 8);
}

console.log('layoutPresets_test ok');
