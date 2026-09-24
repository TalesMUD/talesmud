/** Viewport presets for the 24-column widget grid (row height 40, gap 8). */

export const GRID_COLS = 24;
export const ROW_HEIGHT = 40;
export const GRID_GAP = 8;

const HOTBAR_H = 2;
const ACTION_H = 3;

/** How many rows fit in a viewport height, including gaps. */
export function viewportRows(heightPx) {
  const height = Number(heightPx) || 800;
  /* 96px keeps the grid under the account-chip band and the bottom padding. */
  const usable = Math.max(ROW_HEIGHT * 10, height - 96);
  const rows = Math.floor((usable + GRID_GAP) / (ROW_HEIGHT + GRID_GAP));
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
  const body = Math.max(6, rows - HOTBAR_H - ACTION_H);
  const bars = [
    { id: 'hotbar-1', widgetType: 'hotbar', x: 0, y: body, w: GRID_COLS, h: HOTBAR_H, visible: true },
    { id: 'actionbar-1', widgetType: 'actionbar', x: 0, y: body + HOTBAR_H, w: GRID_COLS, h: ACTION_H, visible: true },
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
