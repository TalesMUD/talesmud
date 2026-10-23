/** Desktop widths where the ground list must stay off enemy art. */
export const DESKTOP_GROUND_MIN_WIDTH = 1280;
export const GROUND_LIST_WIDTH = 380;
export const GROUND_LIST_TOP = 72;
export const GROUND_LIST_RIGHT = 16;
export const GROUND_LIST_MAX_HEIGHT = 560;
export const GROUND_LIST_MIN_HEIGHT = 120;

/** Sprites in the room and in combat. The ground list must miss these boxes. */
export const ENEMY_ART_SELECTOR = '.entity-card.enemy .entity-bg, .battle-stage .enemy-sprite';

export function boxesIntersect(a, b) {
  if (!a || !b) return false;
  const aw = Number(a.width) || 0;
  const ah = Number(a.height) || 0;
  const bw = Number(b.width) || 0;
  const bh = Number(b.height) || 0;
  if (aw <= 0 || ah <= 0 || bw <= 0 || bh <= 0) return false;
  return a.x < b.x + bw && a.x + aw > b.x && a.y < b.y + bh && a.y + ah > b.y;
}

function listWidth(viewportWidth) {
  return Math.min(GROUND_LIST_WIDTH, Math.floor(viewportWidth * 0.32));
}

function maxListHeight(viewportHeight, listHeight) {
  const requested = Number(listHeight) > 0 ? Number(listHeight) : GROUND_LIST_MAX_HEIGHT;
  return Math.min(GROUND_LIST_MAX_HEIGHT, Math.floor(viewportHeight * 0.72), requested);
}

function fits(rect, boxes, viewportWidth, viewportHeight) {
  if (!rect || rect.width <= 0 || rect.height < GROUND_LIST_MIN_HEIGHT) return false;
  if (rect.x < 0 || rect.y < 0) return false;
  if (rect.x + rect.width > viewportWidth || rect.y + rect.height > viewportHeight) return false;
  return !boxes.some((box) => boxesIntersect(rect, box));
}

/**
 * Place the ground list on desktop so its box misses every enemy-art box.
 * Below 1280 the caller keeps the phone sheet and the tablet modal.
 * The preferred spot is the top-right, clear of the left-hand room art.
 */
export function placeGroundList(viewportWidth, viewportHeight, enemyBoxes = [], listHeight) {
  const widthPx = Math.floor(Number(viewportWidth) || 0);
  const heightPx = Math.floor(Number(viewportHeight) || 0);
  if (widthPx < DESKTOP_GROUND_MIN_WIDTH || heightPx <= 0) return null;
  const width = listWidth(widthPx);
  const height = maxListHeight(heightPx, listHeight);
  const boxes = (enemyBoxes || []).filter((box) => box && box.width > 0 && box.height > 0);
  const rightX = widthPx - GROUND_LIST_RIGHT - width;
  const candidates = [{ x: rightX, y: GROUND_LIST_TOP, width, height }];
  for (const box of boxes) {
    const below = Math.floor(box.y + box.height + 8);
    candidates.push({
      x: rightX,
      y: below,
      width,
      height: Math.min(height, heightPx - below - GROUND_LIST_RIGHT),
    });
  }
  const tops = boxes.map((box) => box.y).filter((y) => y > GROUND_LIST_TOP + GROUND_LIST_MIN_HEIGHT);
  if (tops.length) {
    const gap = Math.floor(Math.min(...tops) - GROUND_LIST_TOP - 8);
    candidates.push({ x: rightX, y: GROUND_LIST_TOP, width, height: gap });
  }
  candidates.push({ x: GROUND_LIST_RIGHT, y: GROUND_LIST_TOP, width, height });
  for (const candidate of candidates) {
    if (fits(candidate, boxes, widthPx, heightPx)) return candidate;
  }
  return null;
}

export function groundListStyle(rect) {
  if (!rect) return '';
  return `position:fixed;left:${rect.x}px;top:${rect.y}px;width:${rect.width}px;max-height:${rect.height}px;margin:0;`;
}
