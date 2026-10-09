// Overlay badges stay a fixed screen size. The tile stays in map units.
// svgPerPx is SVG user units per CSS pixel (viewBox size / viewport).
// Multiply the returned screen-pixel boxes by scale when drawing them
// inside the tile. scale is 0 when the tile is too small for a 12px chip.

export const BADGE_SCREEN_PX = 12;

const GAP_PX = 2;
const INSET_PX = 2;

export function badgeWidthPx(text) {
  const n = [...String(text ?? "")].length;
  return Math.max(16, 6 + n * 7);
}

function rowWidth(items) {
  if (!items.length) return 0;
  const pills = items.reduce((sum, item) => sum + badgeWidthPx(item.text), 0);
  return pills + GAP_PX * (items.length - 1);
}

function mergeBadge(badges) {
  const title = badges.map((badge) => badge.title).filter(Boolean).join("\n");
  return { kind: "more", text: `+${badges.length}`, title };
}

function fits(items, available) {
  return rowWidth(items) <= available;
}

// Keep as many leading badges as fit at full size. Collapse the rest into
// one +N chip. Never shrink a chip to make the row fit.
function pack(list, available) {
  const shown = [];
  let index = 0;
  while (index < list.length) {
    const rest = list.slice(index);
    if (fits([...shown, ...rest], available)) return [...shown, ...rest];
    const next = rest[0];
    const after = rest.slice(1);
    if (after.length > 0 && fits([...shown, next], available) && fits([...shown, next, mergeBadge(after)], available)) {
      shown.push(next);
      index += 1;
      continue;
    }
    const merged = mergeBadge(rest);
    if (fits([...shown, merged], available)) return [...shown, merged];
    return shown;
  }
  return shown;
}

export function layoutOverlayBadges(badges, { tileWidth, tileHeight, svgPerPx } = {}) {
  const list = Array.isArray(badges) ? badges.filter((badge) => badge && badge.text != null && badge.text !== "") : [];
  const px = Number(svgPerPx);
  const tw = Number(tileWidth);
  const th = Number(tileHeight);
  if (!list.length || !(px > 0) || !(tw > 0) || !(th > 0)) {
    return { scale: 0, badges: [] };
  }
  const tilePxW = tw / px;
  const tilePxH = th / px;
  if (tilePxH < BADGE_SCREEN_PX + INSET_PX || tilePxW < badgeWidthPx("1") + INSET_PX) {
    return { scale: 0, badges: [] };
  }
  const chosen = pack(list, tilePxW - INSET_PX);
  let cursor = -INSET_PX;
  const placed = chosen.map((badge) => {
    const width = badgeWidthPx(badge.text);
    cursor -= width;
    const box = {
      ...badge,
      x: cursor + width / 2,
      y: INSET_PX + BADGE_SCREEN_PX / 2,
      width,
      height: BADGE_SCREEN_PX,
    };
    cursor -= GAP_PX;
    return box;
  });
  return { scale: px, badges: placed };
}
