const BIOME = {
  meadow: {
    wash: 'rgba(100, 140, 80, 0.11)',
    path: '#5a7048',
    pathFog: 'rgba(90, 112, 72, 0.35)',
    tile: '#e4d8bc',
    tileEdge: '#9a8868',
    ink: '#6b8f5e',
  },
  forest: {
    wash: 'rgba(48, 96, 64, 0.12)',
    path: '#456850',
    pathFog: 'rgba(68, 96, 72, 0.35)',
    tile: '#d6e0c8',
    tileEdge: '#6a8860',
    ink: '#4a7a58',
  },
  water: {
    wash: 'rgba(48, 108, 148, 0.12)',
    path: '#4a7088',
    pathFog: 'rgba(72, 104, 128, 0.35)',
    tile: '#ccdce8',
    tileEdge: '#5a7898',
    ink: '#5a8aaa',
  },
  dungeon: {
    wash: 'rgba(96, 68, 48, 0.14)',
    path: '#6a5040',
    pathFog: 'rgba(88, 68, 52, 0.38)',
    tile: '#c4b4a4',
    tileEdge: '#6a5848',
    ink: '#8a6a52',
  },
  settlement: {
    wash: 'rgba(148, 112, 56, 0.13)',
    path: '#8a7048',
    pathFog: 'rgba(120, 96, 60, 0.35)',
    tile: '#ead8b4',
    tileEdge: '#a88858',
    ink: '#a89060',
  },
  wild: {
    wash: 'rgba(96, 104, 120, 0.1)',
    path: '#6a7280',
    pathFog: 'rgba(96, 104, 116, 0.32)',
    tile: '#d8d4cc',
    tileEdge: '#8a8478',
    ink: '#7a8494',
  },
};

// Muted tints derived from Creator GridWorldEditor area palette
const AREA_TINTS = [
  'rgba(180, 90, 80, 0.14)',
  'rgba(80, 120, 160, 0.14)',
  'rgba(80, 140, 90, 0.14)',
  'rgba(180, 130, 60, 0.14)',
  'rgba(130, 90, 160, 0.14)',
  'rgba(70, 140, 130, 0.14)',
  'rgba(180, 110, 60, 0.14)',
  'rgba(60, 110, 150, 0.14)',
  'rgba(70, 130, 80, 0.14)',
  'rgba(150, 70, 70, 0.14)',
];

const COMPASS_DIRS = new Set([
  'north', 'south', 'east', 'west',
  'northeast', 'northwest', 'southeast', 'southwest',
  'ne', 'nw', 'se', 'sw',
]);

const CARDINAL_ONLY = new Set(['north', 'south', 'east', 'west']);

function biomeOf(key) {
  return BIOME[key] || BIOME.wild;
}

function hashString(str) {
  let h = 0;
  for (let i = 0; i < str.length; i++) {
    h = ((h << 5) - h) + str.charCodeAt(i);
    h |= 0;
  }
  return Math.abs(h);
}

function areaTint(area) {
  if (!area) return 'rgba(120, 110, 90, 0.08)';
  return AREA_TINTS[hashString(area) % AREA_TINTS.length];
}

function computeCamera(places, w, h, panX, panY, userScale) {
  let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
  for (const p of places) {
    const gx = Math.round(p.x);
    const gy = Math.round(p.y);
    if (gx < minX) minX = gx;
    if (gx > maxX) maxX = gx;
    if (gy < minY) minY = gy;
    if (gy > maxY) maxY = gy;
  }
  if (!isFinite(minX)) {
    minX = maxX = minY = maxY = 0;
  }
  const spanX = Math.max(1, maxX - minX + 1);
  const spanY = Math.max(1, maxY - minY + 1);
  const pad = Math.max(40, Math.min(w, h) * 0.14);
  const fit = Math.min((w - pad * 2) / spanX, (h - pad * 2) / spanY);
  const tileStep = Math.max(38, Math.min(fit * userScale, 92));
  return {
    tileStep,
    ox: (minX + maxX) / 2,
    oy: (minY + maxY) / 2,
    panX,
    panY,
  };
}

/** World coords: north decreases Y. Screen: north at top (canvas Y down). */
function projectGrid(gx, gy, cam, w, h) {
  return {
    px: w / 2 + cam.panX + (gx - cam.ox) * cam.tileStep,
    py: h / 2 + cam.panY + (gy - cam.oy) * cam.tileStep,
  };
}

function projectPlace(place, cam, w, h) {
  return projectGrid(Math.round(place.x), Math.round(place.y), cam, w, h);
}

function tileHalf(tileStep) {
  return tileStep * 0.4;
}

function worldDelta(a, b) {
  return {
    dx: Math.round(b.x - a.x),
    dy: Math.round(b.y - a.y),
  };
}

function isGridLink(path, a, b) {
  if (!a || !b) return false;
  if (path.kind === 'passage' || path.kind === 'hidden') return false;
  const d = (path.dir || '').toLowerCase();
  if (d === 'up' || d === 'down' || path.kind === 'stair') return false;
  if (!COMPASS_DIRS.has(d)) return false;
  const { dx, dy } = worldDelta(a, b);
  return Math.abs(dx) <= 1 && Math.abs(dy) <= 1 && (dx !== 0 || dy !== 0);
}

function isCrossArea(a, b) {
  return !!(a.area && b.area && a.area !== b.area);
}

function tileEdgePoint(px, py, half, worldDx, worldDy) {
  if (worldDx === 0 && worldDy === 0) return { px, py };
  if (Math.abs(worldDx) >= Math.abs(worldDy)) {
    return { px: px + (worldDx > 0 ? half : -half), py };
  }
  return { px, py: py + (worldDy > 0 ? half : -half) };
}

function roundRect(ctx, x, y, w, h, r) {
  const rr = Math.min(r, w / 2, h / 2);
  ctx.beginPath();
  ctx.moveTo(x + rr, y);
  ctx.lineTo(x + w - rr, y);
  ctx.quadraticCurveTo(x + w, y, x + w, y + rr);
  ctx.lineTo(x + w, y + h - rr);
  ctx.quadraticCurveTo(x + w, y + h, x + w - rr, y + h);
  ctx.lineTo(x + rr, y + h);
  ctx.quadraticCurveTo(x, y + h, x, y + h - rr);
  ctx.lineTo(x, y + rr);
  ctx.quadraticCurveTo(x, y, x + rr, y);
  ctx.closePath();
}

function shadeColor(hex, percent) {
  if (!hex || hex[0] !== '#') return hex;
  const num = parseInt(hex.slice(1), 16);
  const r = Math.min(255, Math.max(0, (num >> 16) + percent));
  const g = Math.min(255, Math.max(0, ((num >> 8) & 0xff) + percent));
  const b = Math.min(255, Math.max(0, (num & 0xff) + percent));
  return `rgb(${r},${g},${b})`;
}

function drawParchmentBg(ctx, w, h, maximized) {
  const base = ctx.createLinearGradient(0, 0, w, h);
  base.addColorStop(0, '#2a2218');
  base.addColorStop(0.5, '#1e1812');
  base.addColorStop(1, '#14100c');
  ctx.fillStyle = base;
  ctx.fillRect(0, 0, w, h);

  ctx.fillStyle = 'rgba(180, 150, 110, 0.025)';
  for (let i = 0; i < 120; i++) {
    const x = (i * 97) % w;
    const y = (i * 53) % h;
    ctx.fillRect(x, y, 1, 1);
  }

  if (maximized) {
    drawCompassRose(ctx, w - 36, 36, 14);
  }
}

function drawCompassRose(ctx, cx, cy, r) {
  ctx.save();
  ctx.translate(cx, cy);
  ctx.strokeStyle = 'rgba(200, 170, 120, 0.35)';
  ctx.fillStyle = 'rgba(200, 170, 120, 0.12)';
  ctx.lineWidth = 1;
  for (let i = 0; i < 4; i++) {
    ctx.rotate(Math.PI / 2);
    ctx.beginPath();
    ctx.moveTo(0, -r);
    ctx.lineTo(r * 0.28, -r * 0.28);
    ctx.lineTo(0, 0);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
  }
  ctx.beginPath();
  ctx.arc(0, 0, r * 0.18, 0, Math.PI * 2);
  ctx.fill();
  ctx.font = '600 8px Georgia, serif';
  ctx.fillStyle = 'rgba(220, 190, 140, 0.55)';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('N', 0, -r - 8);
  ctx.restore();
}

function drawGrid(ctx, cam, w, h, places) {
  if (!places.length) return;
  let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
  for (const p of places) {
    const gx = Math.round(p.x);
    const gy = Math.round(p.y);
    if (gx < minX) minX = gx;
    if (gx > maxX) maxX = gx;
    if (gy < minY) minY = gy;
    if (gy > maxY) maxY = gy;
  }
  minX -= 1;
  maxX += 1;
  minY -= 1;
  maxY += 1;
  const half = tileHalf(cam.tileStep);
  ctx.strokeStyle = 'rgba(160, 140, 100, 0.06)';
  ctx.lineWidth = 1;
  for (let gx = minX; gx <= maxX; gx++) {
    const top = projectGrid(gx, minY, cam, w, h);
    const bottom = projectGrid(gx, maxY, cam, w, h);
    ctx.beginPath();
    ctx.moveTo(top.px - half, top.py - half);
    ctx.lineTo(bottom.px - half, bottom.py + half);
    ctx.stroke();
  }
  for (let gy = minY; gy <= maxY; gy++) {
    const left = projectGrid(minX, gy, cam, w, h);
    const right = projectGrid(maxX, gy, cam, w, h);
    ctx.beginPath();
    ctx.moveTo(left.px - half, left.py + half);
    ctx.lineTo(right.px + half, right.py + half);
    ctx.stroke();
  }
}

function drawAreaCells(ctx, places, cam, w, h, showLabels) {
  const byArea = new Map();
  for (const p of places) {
    if (!p.discovered || !p.area) continue;
    if (!byArea.has(p.area)) byArea.set(p.area, []);
    byArea.get(p.area).push(p);
  }
  const half = tileHalf(cam.tileStep);
  const cell = cam.tileStep * 0.92;
  for (const [area, rooms] of byArea) {
    const tint = areaTint(area);
    let labelPos = null;
    for (const p of rooms) {
      const { px, py } = projectPlace(p, cam, w, h);
      ctx.fillStyle = tint;
      roundRect(ctx, px - cell / 2, py - cell / 2, cell, cell, 5);
      ctx.fill();
      if (!labelPos) labelPos = { px: px - cell / 2 + 6, py: py - cell / 2 + 10 };
    }
    if (showLabels && labelPos && rooms.length > 1) {
      const name = rooms[0].areaName || area;
      ctx.font = 'italic 600 9px Georgia, serif';
      ctx.fillStyle = 'rgba(220, 200, 160, 0.45)';
      ctx.textAlign = 'left';
      ctx.textBaseline = 'middle';
      ctx.fillText(name, labelPos.px, labelPos.py);
    }
  }
}

function drawRegionWash(ctx, region, cam, w, h) {
  const pts = (region.hull || []).map(([x, y]) => projectGrid(Math.round(x), Math.round(y), cam, w, h));
  if (!pts.length) return;
  const biome = biomeOf(region.biome);
  let minPx = Infinity, maxPx = -Infinity, minPy = Infinity, maxPy = -Infinity;
  for (const p of pts) {
    if (p.px < minPx) minPx = p.px;
    if (p.px > maxPx) maxPx = p.px;
    if (p.py < minPy) minPy = p.py;
    if (p.py > maxPy) maxPy = p.py;
  }
  const pad = cam.tileStep * 0.5;
  ctx.fillStyle = biome.wash;
  roundRect(ctx, minPx - pad, minPy - pad, maxPx - minPx + pad * 2, maxPy - minPy + pad * 2, 12);
  ctx.fill();
}

function drawCorridor(ctx, a, b, pa, pb, path, cam, onTravel) {
  const half = tileHalf(cam.tileStep);
  const { dx, dy } = worldDelta(a, b);
  const from = tileEdgePoint(pa.px, pa.py, half, dx, dy);
  const to = tileEdgePoint(pb.px, pb.py, half, -dx, -dy);
  const biome = biomeOf(a.biome);
  const cross = isCrossArea(a, b);
  const bothKnown = a.discovered && b.discovered;

  ctx.beginPath();
  ctx.moveTo(from.px, from.py);
  ctx.lineTo(to.px, to.py);
  ctx.strokeStyle = onTravel ? '#38bdf8' : cross ? 'rgba(200, 110, 70, 0.75)' : biome.path;
  ctx.lineWidth = onTravel ? 3.5 : cross ? 2 : path.kind === 'road' ? 2.8 : 2.2;
  ctx.globalAlpha = bothKnown ? 0.9 : 0.38;
  ctx.setLineDash(cross ? [7, 5] : []);
  ctx.lineCap = 'round';
  ctx.stroke();
  ctx.setLineDash([]);
  ctx.globalAlpha = 1;
}

function drawPortalLink(ctx, fromPx, fromPy, toPx, toPy, label, color, discovered, cross) {
  const dx = toPx - fromPx;
  const dy = toPy - fromPy;
  const len = Math.hypot(dx, dy) || 1;
  const nx = dx / len;
  const ny = dy / len;
  const stub = 16;
  const portalX = fromPx + nx * (stub + 8);
  const portalY = fromPy + ny * (stub + 8);

  ctx.beginPath();
  ctx.moveTo(fromPx, fromPy);
  ctx.lineTo(portalX, portalY);
  ctx.strokeStyle = cross ? 'rgba(200, 110, 70, 0.7)' : color;
  ctx.lineWidth = 1.6;
  ctx.globalAlpha = discovered ? 0.8 : 0.35;
  ctx.setLineDash([4, 4]);
  ctx.stroke();
  ctx.setLineDash([]);
  ctx.globalAlpha = 1;

  ctx.beginPath();
  roundRect(ctx, portalX - 7, portalY - 5, 14, 10, 2);
  ctx.fillStyle = discovered ? 'rgba(36, 28, 20, 0.9)' : 'rgba(24, 20, 16, 0.65)';
  ctx.fill();
  ctx.strokeStyle = cross ? 'rgba(200, 110, 70, 0.85)' : color;
  ctx.lineWidth = 1.2;
  ctx.stroke();

  if (label) {
    const text = label.length > 9 ? label.slice(0, 8) + '…' : label;
    ctx.font = '600 7px sans-serif';
    ctx.fillStyle = discovered ? 'rgba(232, 220, 190, 0.9)' : 'rgba(148, 140, 128, 0.6)';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'bottom';
    ctx.fillText(text, portalX, portalY - 7);
  }
}

function drawKindGlyph(ctx, place, px, py, size) {
  const s = size * 0.2;
  ctx.fillStyle = 'rgba(48, 40, 30, 0.35)';
  if (place.kind === 'settlement' || place.landmark) {
    ctx.fillRect(px - s, py - s * 0.5, s * 2, s * 1.2);
    ctx.beginPath();
    ctx.moveTo(px, py - s * 1.1);
    ctx.lineTo(px + s * 0.9, py - s * 0.15);
    ctx.lineTo(px - s * 0.9, py - s * 0.15);
    ctx.closePath();
    ctx.fill();
  } else if (place.kind === 'dungeon') {
    ctx.beginPath();
    ctx.moveTo(px, py - s);
    ctx.lineTo(px + s, py + s * 0.2);
    ctx.lineTo(px, py + s);
    ctx.lineTo(px - s, py + s * 0.2);
    ctx.closePath();
    ctx.fill();
  }
}

function drawTile(ctx, place, px, py, tileStep, opts) {
  const half = tileHalf(tileStep);
  const x = px - half;
  const y = py - half;
  const size = half * 2;
  const biome = biomeOf(place.biome);

  if (!place.discovered || place.kind === 'uncharted') {
    ctx.strokeStyle = 'rgba(148, 130, 100, 0.22)';
    ctx.lineWidth = 1;
    ctx.setLineDash([4, 4]);
    roundRect(ctx, x, y, size, size, 3);
    ctx.stroke();
    ctx.setLineDash([]);
    ctx.fillStyle = 'rgba(24, 20, 16, 0.45)';
    roundRect(ctx, x + 2, y + 2, size - 4, size - 4, 2);
    ctx.fill();
    return half;
  }

  const grad = ctx.createLinearGradient(x, y, x + size, y + size);
  grad.addColorStop(0, shadeColor(biome.tile, 8));
  grad.addColorStop(0.45, biome.tile);
  grad.addColorStop(1, shadeColor(biome.tile, -22));
  ctx.fillStyle = grad;
  roundRect(ctx, x, y, size, size, 5);
  ctx.fill();

  ctx.strokeStyle = 'rgba(255, 248, 230, 0.08)';
  ctx.lineWidth = 1;
  roundRect(ctx, x + 2, y + 2, size - 4, size - 4, 4);
  ctx.stroke();

  ctx.strokeStyle = biome.tileEdge;
  ctx.lineWidth = place.current ? 2.8 : 1.4;
  if (place.current) {
    ctx.strokeStyle = '#d4a030';
    ctx.shadowColor = 'rgba(212, 160, 48, 0.5)';
    ctx.shadowBlur = 10;
  }
  roundRect(ctx, x, y, size, size, 5);
  ctx.stroke();
  ctx.shadowBlur = 0;

  if (place.current) {
    ctx.beginPath();
    ctx.arc(px, py - half + 6, 3, 0, Math.PI * 2);
    ctx.fillStyle = 'rgba(212, 160, 48, 0.85)';
    ctx.fill();
  }

  drawKindGlyph(ctx, place, px, py, size);

  if (place.id === opts.travelTargetId) {
    ctx.strokeStyle = '#22d3ee';
    ctx.lineWidth = 2;
    roundRect(ctx, x - 2, y - 2, size + 4, size + 4, 6);
    ctx.stroke();
  }

  return half;
}

export function paintAtlas(ctx, params) {
  const {
    w,
    h,
    atlas,
    activeLayer,
    visiblePlaces,
    visibleRegions,
    currentRoomId,
    maximized,
    panX,
    panY,
    userScale,
    travelPathRoomIds = new Set(),
    travelTargetId = null,
  } = params;

  ctx.clearRect(0, 0, w, h);
  drawParchmentBg(ctx, w, h, maximized);

  if (!visiblePlaces.length) {
    ctx.fillStyle = '#b8a888';
    ctx.font = '13px Georgia, serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(currentRoomId ? 'Charting this floor…' : 'Walk to fill your map', w / 2, h / 2);
    return { hits: [] };
  }

  const cam = computeCamera(visiblePlaces, w, h, panX, panY, userScale);
  const showRoomLabels = userScale > 1.08;
  const byId = {};
  for (const p of atlas.places || []) byId[p.id] = p;

  drawGrid(ctx, cam, w, h, visiblePlaces);

  for (const region of visibleRegions) {
    drawRegionWash(ctx, region, cam, w, h);
  }

  drawAreaCells(ctx, visiblePlaces, cam, w, h, showRoomLabels);

  const layerPaths = (atlas.paths || []).filter((path) => {
    const a = byId[path.from];
    const b = byId[path.to];
    return a && b && (a.layer === activeLayer || b.layer === activeLayer);
  });

  for (const path of layerPaths) {
    const a = byId[path.from];
    const b = byId[path.to];
    if (!a.discovered) continue;
    if (!isGridLink(path, a, b)) continue;
    const pa = projectPlace(a, cam, w, h);
    const pb = projectPlace(b, cam, w, h);
    const onTravel =
      travelPathRoomIds.has(path.from) && travelPathRoomIds.has(path.to) ||
      travelPathRoomIds.has(path.to) && path.from === currentRoomId;
    drawCorridor(ctx, a, b, pa, pb, path, cam, onTravel);
  }

  for (const path of layerPaths) {
    const a = byId[path.from];
    const b = byId[path.to];
    if (!a.discovered || isGridLink(path, a, b)) continue;
    const pa = projectPlace(a, cam, w, h);
    const pb = projectPlace(b, cam, w, h);
    const biome = biomeOf(a.biome);
    drawPortalLink(ctx, pa.px, pa.py, pb.px, pb.py, path.dir, biome.ink, b.discovered, isCrossArea(a, b));
  }

  const hits = [];
  const sorted = [...visiblePlaces].sort((a, b) => {
    if (a.discovered === b.discovered) return 0;
    return a.discovered ? 1 : -1;
  });

  for (const place of sorted) {
    const { px, py } = projectPlace(place, cam, w, h);
    const r = drawTile(ctx, place, px, py, cam.tileStep, { travelTargetId });
    hits.push({ px, py, r: r + 4, place });
  }

  for (const place of visiblePlaces) {
    if (!showRoomLabels || !place.discovered || !place.name) continue;
    const { px, py } = projectPlace(place, cam, w, h);
    const half = tileHalf(cam.tileStep);
    ctx.font = place.current ? '700 10px Georgia, serif' : '600 9px Georgia, serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    ctx.lineWidth = 3;
    ctx.strokeStyle = 'rgba(20, 14, 8, 0.8)';
    ctx.strokeText(place.name, px, py + half + 4);
    ctx.fillStyle = place.current ? '#e8c060' : '#e8dcc8';
    ctx.fillText(place.name, px, py + half + 4);
  }

  return { hits };
}

export { BIOME, projectPlace, computeCamera };
