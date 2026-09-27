import assert from 'assert';
import {
  ACTION_DOCK_H,
  GRID_COLS,
  ROW_HEIGHT,
  foldDockedHotbar,
  presetWidgets,
  viewportRows,
  widgetsOverlap,
  widgetsToPersist,
} from './layoutPresets.js';

function byType(widgets, type) {
  return widgets.find((w) => w.widgetType === type);
}

function assertOnGrid(widgets, rows, label) {
  for (let i = 0; i < widgets.length; i++) {
    const w = widgets[i];
    assert.ok(w.x >= 0 && w.y >= 0, `${label} ${w.id} negative`);
    assert.ok(w.w >= 2 && w.h >= 2, `${label} ${w.id} too small`);
    assert.ok(w.x + w.w <= GRID_COLS, `${label} ${w.id} past the right edge`);
    assert.ok(w.y + w.h <= rows, `${label} ${w.id} past the bottom`);
    for (let j = i + 1; j < widgets.length; j++) {
      assert.ok(!widgetsOverlap(w, widgets[j]), `${label} ${w.id} overlaps ${widgets[j].id}`);
    }
  }
}

for (const height of [768, 1080, 1440]) {
  const rows = viewportRows(height);
  const stack = rows * ROW_HEIGHT;
  assert.ok(stack <= height - 64, `${height} stack ${stack} overflows the grid`);
  assert.ok(stack > height - 64 - ROW_HEIGHT, `${height} leaves a tall empty band`);
  for (const kind of ['desktop', 'wide']) {
    const layout = presetWidgets(kind, height);
    assert.equal(byType(layout, 'hotbar'), undefined);
    const action = byType(layout, 'actionbar');
    assert.equal(action.x, 0);
    assert.equal(action.w, GRID_COLS);
    assert.equal(action.h, ACTION_DOCK_H);
    assert.equal(action.y + action.h, rows);
    const room = byType(layout, 'room');
    const term = byType(layout, 'terminal');
    const sheet = byType(layout, 'character');
    const gear = byType(layout, 'equipment');
    assert.equal(room.y, 0);
    assert.equal(term.y, 0);
    assert.equal(room.x, 0);
    assert.equal(term.x, 9);
    assert.equal(sheet.x, 18);
    assert.equal(gear.x, 18);
    assert.equal(room.h, action.y);
    assert.equal(sheet.y + sheet.h, gear.y);
    assert.equal(gear.y + gear.h, action.y);
    assertOnGrid(layout, rows, `${kind}@${height}`);
  }
}

{
  const normal = [
    { id: 'room-1', widgetType: 'room', x: 0, y: 0, w: 9, h: 13 },
    { id: 'terminal-1', widgetType: 'terminal', x: 9, y: 0, w: 9, h: 13 },
  ];
  const expanded = [
    { id: 'room-1', widgetType: 'room', x: 0, y: 0, w: 24, h: 17 },
    { id: 'terminal-1', widgetType: 'terminal', x: 0, y: 0, w: 2, h: 2 },
  ];
  const saved = widgetsToPersist(expanded, 'room-1', normal);
  assert.equal(saved[0].w, 9);
  assert.equal(saved[1].w, 9);
  const plain = widgetsToPersist(normal, null, expanded);
  assert.equal(plain[0].w, 9);
}

{
  const compact = presetWidgets('compact', 900);
  const room = byType(compact, 'room');
  const term = byType(compact, 'terminal');
  assert.ok(room.h > 0 && term.h > 0);
  assert.equal(term.y, room.h);
  assert.equal(byType(compact, 'hotbar'), undefined);
  assert.equal(byType(compact, 'character'), undefined);
  assert.equal(byType(compact, 'equipment'), undefined);
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
