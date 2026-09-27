import assert from 'assert';
import {
  ACTION_DOCK_H,
  GRID_COLS,
  ROW_HEIGHT,
  VIEWPORT_CHROME_PX,
  fitWidgetsToRows,
  foldDockedHotbar,
  layoutBottom,
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

for (const height of [730, 768, 945, 1080, 1440]) {
  const rows = viewportRows(height);
  const stack = rows * ROW_HEIGHT;
  assert.ok(stack <= height - VIEWPORT_CHROME_PX, `${height} stack ${stack} overflows the grid`);
  assert.ok(stack > height - VIEWPORT_CHROME_PX - ROW_HEIGHT - 8, `${height} leaves a tall empty band`);
  for (const kind of ['desktop', 'wide']) {
    const layout = presetWidgets(kind, height);
    assert.equal(byType(layout, 'hotbar'), undefined);
    assert.equal(byType(layout, 'character'), undefined);
    assert.equal(byType(layout, 'equipment'), undefined);
    assert.equal(byType(layout, 'terminal'), undefined);
    const action = byType(layout, 'actionbar');
    assert.equal(action.x, 0);
    assert.equal(action.w, GRID_COLS);
    assert.equal(action.h, ACTION_DOCK_H);
    assert.equal(action.y + action.h, rows);
    const room = byType(layout, 'room');
    const sheet = layout.find((w) => w.id === 'sheet-1');
    const tools = layout.find((w) => w.id === 'tools-1');
    const inv = byType(layout, 'inventory');
    assert.equal(room.x, 0);
    assert.equal(room.w, 14);
    assert.equal(room.h, action.y);
    assert.equal(sheet.widgetType, 'tabcontainer');
    assert.equal(sheet.activeTabIndex, 0);
    assert.equal(sheet.tabs[0].widgetType, 'character');
    assert.equal(sheet.tabs[1].widgetType, 'equipment');
    assert.equal(tools.tabs[0].widgetType, 'terminal');
    assert.equal(tools.tabs[1].widgetType, 'questlog');
    assert.equal(tools.tabs[2].widgetType, 'minimap');
    assert.equal(sheet.x, 14);
    assert.equal(tools.x, 14);
    assert.equal(inv.x, 14);
    assert.equal(sheet.y, 0);
    assert.equal(tools.y, sheet.h);
    assert.equal(inv.y, tools.y + tools.h);
    assert.equal(inv.y + inv.h, action.y);
    assert.ok(sheet.h >= tools.h, `${kind}@${height} sheet shorter than the tool tabs`);
    assertOnGrid(layout, rows, `${kind}@${height}`);
  }
}

{
  // Saved Gimli-style stack: sheet on the left, room in the middle, two tab
  // columns on the right, action bar under a layout that is taller than 1080p.
  const gimli = [
    { id: 'equipment-1', widgetType: 'equipment', x: 0, y: 0, w: 6, h: 14, visible: true },
    { id: 'character-1', widgetType: 'character', x: 0, y: 14, w: 6, h: 14, visible: true },
    { id: 'room-1', widgetType: 'room', x: 6, y: 0, w: 10, h: 28, visible: true },
    {
      id: 'tabs-a',
      widgetType: 'tabcontainer',
      x: 16,
      y: 0,
      w: 8,
      h: 14,
      visible: true,
      tabs: [{ widgetType: 'questlog', id: 'questlog-1' }],
      activeTabIndex: 0,
    },
    {
      id: 'tabs-b',
      widgetType: 'tabcontainer',
      x: 16,
      y: 14,
      w: 8,
      h: 14,
      visible: true,
      tabs: [{ widgetType: 'terminal', id: 'terminal-1' }],
      activeTabIndex: 0,
    },
    { id: 'actionbar-1', widgetType: 'actionbar', x: 0, y: 28, w: 24, h: 4, visible: true },
  ];
  assert.equal(layoutBottom(gimli), 32);
  for (const height of [730, 768, 945, 1080, 1440]) {
    const rows = viewportRows(height);
    const fitted = fitWidgetsToRows(gimli, rows);
    const bottom = layoutBottom(fitted);
    assert.ok(bottom <= rows, `${height} fitted bottom ${bottom} > ${rows}`);
    assert.ok(bottom * ROW_HEIGHT <= height - VIEWPORT_CHROME_PX, `${height} pixels overflow`);
    const action = byType(fitted, 'actionbar');
    assert.equal(action.y + action.h, bottom);
    assert.ok(action.h >= 3, `${height} action bar shrank to ${action.h}`);
    assert.equal(action.w, GRID_COLS);
    assert.equal(byType(fitted, 'room').x, 6);
    assert.equal(fitted.find((w) => w.id === 'tabs-a').tabs[0].widgetType, 'questlog');
    for (const w of fitted) {
      assert.ok(w.y + w.h <= bottom, `${height} ${w.id} past the fitted bottom`);
    }
  }
  const short = fitWidgetsToRows(
    [{ id: 'room-1', widgetType: 'room', x: 0, y: 0, w: 24, h: 8, visible: true }],
    20,
  );
  assert.equal(short[0].h, 8);
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
