// World-map fit math.
// World coords: north decreases Y, the same axis as authored rooms and the play atlas.
// Screen: north is toward the top (SVG y grows downward).

const EMPTY = { x: -600, y: -400, width: 1200, height: 800 };

export function roomPixel(coords, gridScale) {
  return {
    x: Number(coords.x) * gridScale,
    y: Number(coords.y) * gridScale,
  };
}

// Inverse of roomPixel. A drag toward the top of the screen decreases Y.
export function pixelToCoords(px, py, gridScale) {
  return {
    x: Math.round(px / gridScale),
    y: Math.round(py / gridScale),
  };
}

// fitViewBox frames the given coordinates, including padding in grid units.
// Rooms with no coordinates are skipped. An empty list returns the default view.
export function fitViewBox(coords, gridScale, padding = 2) {
  const points = [];
  for (const entry of coords || []) {
    if (!entry || !Number.isFinite(Number(entry.x)) || !Number.isFinite(Number(entry.y))) {
      continue;
    }
    points.push({ x: Number(entry.x), y: Number(entry.y) });
  }
  if (points.length === 0 || !Number.isFinite(gridScale) || gridScale <= 0) {
    return { ...EMPTY };
  }

  let minX = points[0].x;
  let maxX = points[0].x;
  let minY = points[0].y;
  let maxY = points[0].y;
  for (const point of points) {
    if (point.x < minX) minX = point.x;
    if (point.x > maxX) maxX = point.x;
    if (point.y < minY) minY = point.y;
    if (point.y > maxY) maxY = point.y;
  }

  return {
    x: (minX - padding) * gridScale,
    y: (minY - padding) * gridScale,
    width: Math.max((maxX - minX + padding * 2) * gridScale, 600),
    height: Math.max((maxY - minY + padding * 2) * gridScale, 400),
  };
}
