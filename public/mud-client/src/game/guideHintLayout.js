// Placement for the next-step chip. Pure so a test can pin the 375 layout.

export function hintBlocked(state) {
  if (!state) return true;
  if (state.inCombat) return true;
  if (state.combatPhase && state.combatPhase !== "idle") return true;
  if (state.combatOutcome) return true;
  if (state.dialogActive) return true;
  if (state.shop) return true;
  if (state.needGreed) return true;
  return false;
}

function boxOf(rect) {
  if (!rect) return null;
  const top = Number(rect.top);
  const left = Number(rect.left);
  const width = Number(rect.width);
  const height = Number(rect.height);
  if (![top, left, width, height].every(Number.isFinite)) return null;
  if (width <= 0 || height <= 0) return null;
  return { top, left, width, height, right: left + width, bottom: top + height };
}

function intersects(a, b) {
  if (!a || !b) return false;
  return a.left < b.right && a.right > b.left && a.top < b.bottom && a.bottom > b.top;
}

// hintRect sits the chip just above the highest hotbar and rejects any overlap
// with the room name or a hotbar. Null means do not show it.
export function hintRect({ vw, vh, hotbars, roomNames, width, height }) {
  vw = Number(vw);
  vh = Number(vh);
  if (!Number.isFinite(vw) || !Number.isFinite(vh) || vw < 80 || vh < 80) return null;
  const w = Math.min(Number(width) || 320, vw - 16);
  const h = Number(height) || 48;
  if (w < 80 || h < 24) return null;

  const bars = (hotbars || []).map(boxOf).filter(Boolean);
  const names = (roomNames || []).map(boxOf).filter(Boolean);
  const anchor = bars.reduce((min, rect) => (!min || rect.top < min.top ? rect : min), null);

  let top = vh - h - 16;
  if (anchor) top = anchor.top - 8 - h;
  let left = Math.round((vw - w) / 2);
  if (left < 8) left = 8;
  if (left + w > vw - 8) left = Math.max(8, vw - 8 - w);

  const rect = boxOf({ top, left, width: w, height: h });
  if (!rect) return null;
  if (rect.top < 8 || rect.bottom > vh - 4) return null;
  if (bars.some((bar) => intersects(rect, bar))) return null;
  if (names.some((name) => intersects(rect, name))) return null;
  return { top: rect.top, left: rect.left, width: rect.width, height: rect.height };
}
