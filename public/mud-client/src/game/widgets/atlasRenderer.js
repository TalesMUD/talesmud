const BIOME = {
  meadow: {
    fill: 'rgba(132, 184, 92, 0.14)',
    wash: 'rgba(100, 150, 80, 0.12)',
    ink: '#86efac',
    path: '#6b8f5e',
    tile: '#d8ccb0',
    tileEdge: '#a89878',
  },
  forest: {
    fill: 'rgba(52, 120, 72, 0.16)',
    wash: 'rgba(40, 90, 60, 0.14)',
    ink: '#34d399',
    path: '#4a7a58',
    tile: '#c8d4bc',
    tileEdge: '#7a9a72',
  },
  water: {
    fill: 'rgba(56, 132, 176, 0.16)',
    wash: 'rgba(40, 100, 140, 0.12)',
    ink: '#7dd3fc',
    path: '#5a8aaa',
    tile: '#c8d8e4',
    tileEdge: '#6a8aa8',
  },
  dungeon: {
    fill: 'rgba(148, 92, 64, 0.16)',
    wash: 'rgba(100, 70, 50, 0.14)',
    ink: '#fdba74',
    path: '#8a6a52',
    tile: '#b8a898',
    tileEdge: '#7a6858',
  },
  settlement: {
    fill: 'rgba(196, 148, 72, 0.16)',
    wash: 'rgba(160, 120, 60, 0.14)',
    ink: '#fcd34d',
    path: '#a89060',
    tile: '#e0d0b0',
    tileEdge: '#a89068',
  },
  wild: {
    fill: 'rgba(120, 130, 150, 0.12)',
    wash: 'rgba(90, 100, 120, 0.1)',
    ink: '#cbd5e1',
    path: '#7a8494',
    tile: '#d0ccc4',
    tileEdge: '#9a9488',
  },
};

const COMPASS_DIRS = new Set([
  'north', 'south', 'east', 'west',
  'northeast', 'northwest', 'southeast', 'southwest',
  'ne', 'nw', 'se', 'sw',
]);

function biomeOf(key) {
  return BIOME[key] || BIOME.wild;
}

function gridKey(place) {
  return `${Math.round(place.x)},${Math.round(place.y)}`;
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
  const pad = Math.max(36, Math.min(w, h) * 0.12);
  const fit = Math.min((w - pad * 2) / spanX, (h - pad * 2) / spanY);
  const tileStep = Math.max(34, Math.min(fit * userScale, 88));
  return {
    tileStep,
    ox: (minX + maxX) / 2,
    oy: (minY + maxY) / 2,
    panX,
    panY,
  };
}

function projectGrid(gx, gy, cam, w, h) {
  return {
    px: w / 2 + cam.panX + (gx - cam.ox) * cam.tileStep,
    py: h / 2 + cam.panY + (gy - cam.oy) * cam.tileStep,
  };
}

function projectPlace(place, cam, w, h) {
  return projectGrid(Math.round(place.x), Math.round(place.y), cam, w, h);
}

function isGridLink(path, a, b) {
  if (!a || !b) return false;
  if (path.kind === 'passage' || path.kind === 'hidden') return false;
  const d = (path.dir || '').toLowerCase();
  if (d === 'up' || d === 'down' || path.kind === 'stair') return false;
  if (!COMPASS_DIRS.has(d)) return false;
  const dx = Math.round(b.x - a.x);
  const dy = Math.round(b.y - a.y);
  return Math.abs(dx) <= 1 && Math.abs(dy) <= 1 && (dx !== 0 || dy !== 0);
}

function tileHalf(tileStep) {
  return tileStep * 0.38;
}

function drawParchmentBg(ctx, w, h) {
  ctx.fillStyle = '#1c1814';
  ctx.fillRect(0, 0, w, h);
  const grad = ctx.createRadialGradient(w * 0.5, h * 0.45, 0, w * 0.5, h * 0.45, Math.max(w, h) * 0.7);
  grad.addColorStop(0, 'rgba(48, 40, 32, 0.35)');
  grad.addColorStop(1, 'rgba(12, 10, 8, 0.9)');
  ctx.fillStyle = grad;
  ctx.fillRect(0, 0, w, h);
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
  ctx.strokeStyle = 'rgba(148, 130, 100, 0.08)';
  ctx.lineWidth = 1;
  for (let gx = minX; gx <= maxX; gx++) {
    const top = projectGrid(gx, minY, cam, w, h);
    const bottom = projectGrid(gx, maxY, cam, w, h);
    ctx.beginPath();
    ctx.moveTo(top.px, top.py);
    ctx.lineTo(bottom.px, bottom.py);
    ctx.stroke();
  }
  for (let gy = minY; gy <= maxY; gy++) {
    const left = projectGrid(minX, gy, cam, w, h);
    const right = projectGrid(maxX, gy, cam, w, h);
    ctx.beginPath();
    ctx.moveTo(left.px, left.py);
    ctx.lineTo(right.px, right.py);
    ctx.stroke();
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
  const pad = cam.tileStep * 0.55;
  const rx = minPx - pad;
  const ry = minPy - pad;
  const rw = maxPx - minPx + pad * 2;
  const rh = maxPy - minPy + pad * 2;
  ctx.fillStyle = biome.wash;
  roundRect(ctx, rx, ry, rw, rh, 10);
  ctx.fill();
  ctx.strokeStyle = biome.fill.replace('0.14', '0.28').replace('0.16', '0.28').replace('0.12', '0.22');
  ctx.lineWidth = 1;
  ctx.stroke();
  if (region.name) {
    ctx.font = '600 10px Georgia, serif';
    ctx.fillStyle = 'rgba(226, 232, 240, 0.55)';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(region.name, rx + rw / 2, ry + 12);
  }
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

function drawGridCorridor(ctx, ax, ay, bx, by, color, width, dashed, alpha) {
  ctx.beginPath();
  ctx.moveTo(ax, ay);
  ctx.lineTo(bx, by);
  ctx.strokeStyle = color;
  ctx.lineWidth = width;
  ctx.globalAlpha = alpha;
  ctx.setLineDash(dashed ? [5, 4] : []);
  ctx.stroke();
  ctx.setLineDash([]);
  ctx.globalAlpha = 1;
}

function drawPortalLink(ctx, fromPx, fromPy, toPx, toPy, label, color, discovered) {
  const dx = toPx - fromPx;
  const dy = toPy - fromPy;
  const len = Math.hypot(dx, dy) || 1;
  const nx = dx / len;
  const ny = dy / len;
  const stub = 14;
  const endX = fromPx + nx * stub;
  const endY = fromPy + ny * stub;
  const portalX = fromPx + nx * (stub + 10);
  const portalY = fromPy + ny * (stub + 10);

  ctx.beginPath();
  ctx.moveTo(fromPx, fromPy);
  ctx.lineTo(endX, endY);
  ctx.strokeStyle = color;
  ctx.lineWidth = 1.5;
  ctx.globalAlpha = discovered ? 0.75 : 0.35;
  ctx.setLineDash([4, 3]);
  ctx.stroke();
  ctx.setLineDash([]);
  ctx.globalAlpha = 1;

  ctx.beginPath();
  ctx.arc(portalX, portalY, 5, 0, Math.PI * 2);
  ctx.fillStyle = discovered ? 'rgba(30, 25, 20, 0.85)' : 'rgba(20, 18, 16, 0.6)';
  ctx.fill();
  ctx.strokeStyle = color;
  ctx.lineWidth = 1.2;
  ctx.stroke();

  if (label && discovered) {
    const text = label.length > 10 ? label.slice(0, 9) + '…' : label;
    ctx.font = '600 8px sans-serif';
    ctx.fillStyle = 'rgba(226, 232, 240, 0.85)';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(text, portalX, portalY - 10);
  }
}

function drawKindGlyph(ctx, place, px, py, size) {
  const s = size * 0.22;
  ctx.fillStyle = 'rgba(40, 35, 28, 0.45)';
  if (place.kind === 'settlement') {
    ctx.fillRect(px - s, py - s * 0.6, s * 2, s * 1.4);
    ctx.beginPath();
    ctx.moveTo(px, py - s * 1.2);
    ctx.lineTo(px + s, py - s * 0.2);
    ctx.lineTo(px - s, py - s * 0.2);
    ctx.closePath();
    ctx.fill();
  } else if (place.kind === 'dungeon') {
    ctx.beginPath();
    ctx.moveTo(px, py - s);
    ctx.lineTo(px + s, py);
    ctx.lineTo(px, py + s);
    ctx.lineTo(px - s, py);
    ctx.closePath();
    ctx.fill();
  } else if (place.kind === 'water') {
    ctx.beginPath();
    ctx.arc(px, py, s, 0, Math.PI * 2);
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
    ctx.strokeStyle = 'rgba(148, 163, 184, 0.28)';
    ctx.lineWidth = 1;
    ctx.setLineDash([3, 3]);
    roundRect(ctx, x, y, size, size, 3);
    ctx.stroke();
    ctx.setLineDash([]);
    ctx.fillStyle = 'rgba(30, 28, 24, 0.35)';
    roundRect(ctx, x + 2, y + 2, size - 4, size - 4, 2);
    ctx.fill();
    return half;
  }

  const grad = ctx.createLinearGradient(x, y, x + size, y + size);
  grad.addColorStop(0, biome.tile);
  grad.addColorStop(1, shadeColor(biome.tile, -18));
  ctx.fillStyle = grad;
  roundRect(ctx, x, y, size, size, 4);
  ctx.fill();

  ctx.strokeStyle = biome.tileEdge;
  ctx.lineWidth = place.current ? 2.5 : 1.2;
  if (place.current) {
    ctx.strokeStyle = '#fbbf24';
    ctx.shadowColor = 'rgba(251, 191, 36, 0.45)';
    ctx.shadowBlur = 8;
  }
  roundRect(ctx, x, y, size, size, 4);
  ctx.stroke();
  ctx.shadowBlur = 0;

  drawKindGlyph(ctx, place, px, py, size);

  if (place.id === opts.travelTargetId) {
    ctx.strokeStyle = '#22d3ee';
    ctx.lineWidth = 2;
    roundRect(ctx, x - 2, y - 2, size + 4, size + 4, 5);
    ctx.stroke();
  }

  return half;
}

function shadeColor(hex, percent) {
  if (!hex || hex[0] !== '#') return hex;
  const num = parseInt(hex.slice(1), 16);
  const r = Math.min(255, Math.max(0, (num >> 16) + percent));
  const g = Math.min(255, Math.max(0, ((num >> 8) & 0xff) + percent));
  const b = Math.min(255, Math.max(0, (num & 0xff) + percent));
  return `rgb(${r},${g},${b})`;
}

/**
 * Paint RPG-style grid atlas onto a canvas context.
 * Returns { hits: [{ px, py, r, place }] } for pointer interaction.
 */
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
  drawParchmentBg(ctx, w, h);

  if (!visiblePlaces.length) {
    ctx.fillStyle = '#94a3b8';
    ctx.font = '12px Georgia, serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(currentRoomId ? 'Charting this floor…' : 'Walk to fill your atlas', w / 2, h / 2);
    return { hits: [] };
  }

  const cam = computeCamera(visiblePlaces, w, h, panX, panY, userScale);
  const byId = {};
  for (const p of atlas.places || []) byId[p.id] = p;

  drawGrid(ctx, cam, w, h, visiblePlaces);

  for (const region of visibleRegions) {
    drawRegionWash(ctx, region, cam, w, h);
  }

  const layerPaths = (atlas.paths || []).filter((path) => {
    const a = byId[path.from];
    const b = byId[path.to];
    return a && b && (a.layer === activeLayer || b.layer === activeLayer);
  });

  for (const path of layerPaths) {
    const a = byId[path.from];
    const b = byId[path.to];
    if (!a.discovered) continue;
    const pa = projectPlace(a, cam, w, h);
    const pb = projectPlace(b, cam, w, h);
    const biome = biomeOf(a.biome);
    const onTravel =
      travelPathRoomIds.has(path.from) && travelPathRoomIds.has(path.to) ||
      travelPathRoomIds.has(path.to) && path.from === currentRoomId;

    if (isGridLink(path, a, b)) {
      const bothKnown = a.discovered && b.discovered;
      drawGridCorridor(
        ctx,
        pa.px,
        pa.py,
        pb.px,
        pb.py,
        onTravel ? '#38bdf8' : biome.path,
        onTravel ? 3 : path.kind === 'road' ? 2.5 : 2,
        false,
        bothKnown ? 0.85 : 0.35,
      );
    }
  }

  for (const path of layerPaths) {
    const a = byId[path.from];
    const b = byId[path.to];
    if (!a.discovered || isGridLink(path, a, b)) continue;
    const pa = projectPlace(a, cam, w, h);
    const pb = projectPlace(b, cam, w, h);
    const biome = biomeOf(a.biome);
    drawPortalLink(ctx, pa.px, pa.py, pb.px, pb.py, path.dir, biome.ink, b.discovered);
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
    if (!place.discovered || !place.name) continue;
    const showName = place.current || place.landmark || maximized || visiblePlaces.length <= 10;
    if (!showName) continue;
    const { px, py } = projectPlace(place, cam, w, h);
    const half = tileHalf(cam.tileStep);
    ctx.font = place.current ? '700 10px Georgia, serif' : '600 9px sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    ctx.lineWidth = 3;
    ctx.strokeStyle = 'rgba(8, 12, 18, 0.85)';
    ctx.strokeText(place.name, px, py + half + 3);
    ctx.fillStyle = place.current ? '#fbbf24' : '#e8e0d4';
    ctx.fillText(place.name, px, py + half + 3);
  }

  return { hits };
}

export { BIOME, projectPlace, computeCamera };
