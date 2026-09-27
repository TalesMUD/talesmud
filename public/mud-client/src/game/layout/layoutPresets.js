/** Viewport presets for the 24-column widget grid (row height 40, horizontal gap 8). */

export const GRID_COLS = 24;
export const ROW_HEIGHT = 40;
export const GRID_GAP = 8;
/** Spell slots live inside the action bar, so the dock is one widget this tall. */
export const ACTION_DOCK_H = 4;

const ACTION_H = ACTION_DOCK_H;

/** 52px top band + 8px bottom padding inside the play shell. */
export const VIEWPORT_CHROME_PX = 60;

/**
 * How many 40px rows fit in the play shell.
 * Leaves at least 8px so a border or a subpixel does not spill into page scroll.
 */
export function viewportRows(heightPx) {
  const height = Number(heightPx) || 800;
  const budget = Math.max(ROW_HEIGHT * 8, height - VIEWPORT_CHROME_PX);
  let rows = Math.floor(budget / ROW_HEIGHT);
  if (rows * ROW_HEIGHT > budget - 8) rows -= 1;
  return Math.max(10, Math.min(36, rows));
}

/** Lowest occupied row (y + h) in a layout. */
export function layoutBottom(widgets) {
  let bottom = 0;
  if (!Array.isArray(widgets)) return 0;
  for (const w of widgets) {
    const y = Math.round(Number(w.y) || 0);
    const h = Math.round(Number(w.h) || 0);
    if (y + h > bottom) bottom = y + h;
  }
  return bottom;
}

/**
 * Scale a layout that is taller than `rows` so it ends on that row.
 * A full-width action bar that already sits on the bottom keeps a readable
 * height and stays pinned there. Shorter layouts are returned unchanged.
 * x and w are left alone. Tab lists are copied with the widget.
 */
export function fitWidgetsToRows(widgets, rows) {
  if (!Array.isArray(widgets)) return [];
  const target = Math.max(6, Math.round(Number(rows) || 0));
  const copy = widgets.map((w) => ({ ...w }));
  const bottom = layoutBottom(copy);
  if (bottom <= target) return copy;

  const action = copy.find((w) => w.widgetType === 'actionbar' && w.visible !== false);
  let pin = false;
  let dockSrc = 0;
  let dockH = 0;
  if (action) {
    const ay = Math.round(Number(action.y) || 0);
    const ah = Math.max(2, Math.round(Number(action.h) || 2));
    const aw = Math.round(Number(action.w) || 0);
    if (ay + ah >= bottom && aw >= GRID_COLS) {
      pin = true;
      dockSrc = ah;
      dockH = Math.min(ah, ACTION_DOCK_H);
      if (dockH < 3 && ah >= 3 && target >= 8) dockH = 3;
      if (target - dockH < 4) dockH = Math.max(2, target - 4);
    }
  }

  const srcSpan = Math.max(1, bottom - (pin ? dockSrc : 0));
  const destSpan = Math.max(1, target - (pin ? dockH : 0));
  const mapped = copy.map((w) => {
    if (pin && w === action) {
      return { ...w, y: target - dockH, h: dockH };
    }
    const y = Math.max(0, Math.round(Number(w.y) || 0));
    const h = Math.max(1, Math.round(Number(w.h) || 1));
    const y1 = Math.min(y, srcSpan);
    const y2 = Math.min(y + h, srcSpan);
    let ny = Math.round((y1 * destSpan) / srcSpan);
    let nh = Math.round((y2 * destSpan) / srcSpan) - ny;
    if (nh < 1) nh = 1;
    if (ny >= destSpan) ny = destSpan - 1;
    if (ny + nh > destSpan) nh = destSpan - ny;
    if (nh < 1) nh = 1;
    return { ...w, y: ny, h: nh };
  });
  return mapped.map((w) => {
    let y = Math.max(0, Math.round(Number(w.y) || 0));
    let h = Math.max(1, Math.round(Number(w.h) || 1));
    if (y >= target) {
      y = target - 1;
      h = 1;
    } else if (y + h > target) {
      h = target - y;
    }
    return { ...w, y, h };
  });
}

/**
 * Right-hand column heights for the desktop preset. Sum is `body`.
 * The character tab gets about two thirds so attributes and combat stats
 * fit at 1080p without the panel scrolling.
 */
function columnSplit(body) {
  const total = Math.max(6, body);
  let sheet = Math.max(4, Math.round(total * 0.66));
  let tools = Math.max(2, Math.round(total * 0.18));
  let inv = total - sheet - tools;
  if (inv < 2) {
    const fromTools = Math.min(2 - inv, Math.max(0, tools - 2));
    tools -= fromTools;
    inv += fromTools;
  }
  if (inv < 2) {
    const fromSheet = Math.min(2 - inv, Math.max(0, sheet - 3));
    sheet -= fromSheet;
    inv += fromSheet;
  }
  inv = total - sheet - tools;
  return { sheet, tools, inv };
}

/** Shape id from width. Height is applied separately so the grid fills the window. */
export function kindForWidth(widthPx) {
  const width = Number(widthPx) || 1366;
  if (width < 1100) return 'compact';
  if (width >= 2200) return 'wide';
  return 'desktop';
}

export function presetLabel(kind) {
  if (kind === 'compact') return 'Compact';
  if (kind === 'wide') return 'Wide';
  return 'Desktop';
}

/**
 * Widgets for a named preset. Compact stacks room over terminal.
 * Desktop and wide keep the room large on the left. The right column is
 * Character/Equipment tabs, Terminal/Quest/Map tabs, then Inventory.
 */
export function presetWidgets(kind, heightPx) {
  const rows = viewportRows(heightPx);
  const body = Math.max(6, rows - ACTION_H);
  const bars = [
    { id: 'actionbar-1', widgetType: 'actionbar', x: 0, y: body, w: GRID_COLS, h: ACTION_H, visible: true },
  ];
  if (kind === 'compact') {
    const top = Math.max(4, Math.floor(body * 0.55));
    return [
      { id: 'room-1', widgetType: 'room', x: 0, y: 0, w: GRID_COLS, h: top, visible: true },
      { id: 'terminal-1', widgetType: 'terminal', x: 0, y: top, w: GRID_COLS, h: Math.max(4, body - top), visible: true },
      ...bars,
    ];
  }
  const { sheet, tools, inv } = columnSplit(body);
  const side = 10;
  const roomW = GRID_COLS - side;
  return [
    { id: 'room-1', widgetType: 'room', x: 0, y: 0, w: roomW, h: body, visible: true },
    {
      id: 'sheet-1',
      widgetType: 'tabcontainer',
      x: roomW,
      y: 0,
      w: side,
      h: sheet,
      visible: true,
      activeTabIndex: 0,
      tabs: [
        { widgetType: 'character', id: 'character-1' },
        { widgetType: 'equipment', id: 'equipment-1' },
      ],
    },
    {
      id: 'tools-1',
      widgetType: 'tabcontainer',
      x: roomW,
      y: sheet,
      w: side,
      h: tools,
      visible: true,
      activeTabIndex: 0,
      tabs: [
        { widgetType: 'terminal', id: 'terminal-1' },
        { widgetType: 'questlog', id: 'questlog-1' },
        { widgetType: 'minimap', id: 'minimap-1' },
      ],
    },
    { id: 'inventory-1', widgetType: 'inventory', x: roomW, y: sheet + tools, w: side, h: inv, visible: true },
    ...bars,
  ];
}

/** True when two grid rects share any cell. */
export function widgetsOverlap(a, b) {
  if (!a || !b) return false;
  return a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y;
}

/**
 * A focused widget is expanded over the grid. Persist the arrangement from
 * before that expansion so Save does not store the 24-column cover.
 */
export function widgetsToPersist(live, focusId, focusSnapshot) {
  if (focusId && Array.isArray(focusSnapshot)) {
    return focusSnapshot.map((w) => ({ ...w }));
  }
  return Array.isArray(live) ? live.map((w) => ({ ...w })) : [];
}

/** Pull every widget back onto the 24-column grid. Minimum size is 2×2. Does not reorder a saved layout. */
export function clampWidgets(widgets) {
  if (!Array.isArray(widgets)) return [];
  return widgets.map((w) => {
    let x = Math.round(Number(w.x) || 0);
    let y = Math.round(Number(w.y) || 0);
    let width = Math.round(Number(w.w) || 2);
    let height = Math.round(Number(w.h) || 2);
    if (width < 2) width = 2;
    if (height < 2) height = 2;
    if (width > GRID_COLS) width = GRID_COLS;
    if (x < 0) x = 0;
    if (y < 0) y = 0;
    if (x + width > GRID_COLS) x = Math.max(0, GRID_COLS - width);
    return { ...w, x, y, w: width, h: height };
  });
}

/**
 * A full-width spell bar sitting on a full-width action bar is the old
 * floating strip. Fold those rows into the action bar so the slots render
 * inside that dock. A hotbar the player moved elsewhere stays put.
 */
export function foldDockedHotbar(widgets) {
  if (!Array.isArray(widgets)) return [];
  const hot = widgets.find((w) => w.widgetType === 'hotbar' && w.visible !== false);
  const act = widgets.find((w) => w.widgetType === 'actionbar' && w.visible !== false);
  if (!hot || !act) return widgets.slice();
  const hx = Math.round(Number(hot.x) || 0);
  const hy = Math.round(Number(hot.y) || 0);
  const hw = Math.round(Number(hot.w) || 0);
  const hh = Math.round(Number(hot.h) || 0);
  const ax = Math.round(Number(act.x) || 0);
  const ay = Math.round(Number(act.y) || 0);
  const aw = Math.round(Number(act.w) || 0);
  const ah = Math.round(Number(act.h) || 0);
  if (hx !== ax || hw !== aw || hw < GRID_COLS || hy + hh !== ay) return widgets.slice();
  return widgets
    .filter((w) => w !== hot)
    .map((w) => (w === act ? { ...w, y: hy, h: hh + ah } : w));
}
