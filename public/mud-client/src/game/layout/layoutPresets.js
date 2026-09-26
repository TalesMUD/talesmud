/** Viewport presets for the 24-column widget grid (row height 40, horizontal gap 8). */

export const GRID_COLS = 24;
export const ROW_HEIGHT = 40;
export const GRID_GAP = 8;
/** Spell slots live inside the action bar, so the dock is one widget this tall. */
export const ACTION_DOCK_H = 4;

const ACTION_H = ACTION_DOCK_H;

/**
 * How many 40px rows fit under the account band.
 * 52px top padding + 12px bottom padding. The grid container is the rest,
 * and svelte-grid's content height is rows * ROW_HEIGHT.
 */
export function viewportRows(heightPx) {
  const height = Number(heightPx) || 800;
  const usable = Math.max(ROW_HEIGHT * 10, height - 64);
  const rows = Math.floor(usable / ROW_HEIGHT);
  return Math.max(12, Math.min(36, rows));
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
 * Desktop and wide are side by side; wide is the same split, just taller on big screens.
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
  return [
    { id: 'room-1', widgetType: 'room', x: 0, y: 0, w: 12, h: body, visible: true },
    { id: 'terminal-1', widgetType: 'terminal', x: 12, y: 0, w: 12, h: body, visible: true },
    ...bars,
  ];
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
