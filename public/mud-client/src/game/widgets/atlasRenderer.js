import { TERRAIN_SHEET } from './terrainSheet.js';

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

const COMPASS_DIRS = new Set([
  'north', 'south', 'east', 'west',
  'northeast', 'northwest', 'southeast', 'southwest',
  'ne', 'nw', 'se', 'sw',
]);

const CARDINAL_ONLY = new Set(['north', 'south', 'east', 'west']);

/** Exit names that must never be painted as canvas text (ticks only). */
const DIR_LABEL_BLOCKLIST = new Set([
  ...COMPASS_DIRS,
  'up', 'down', 'u', 'd', 'n', 's', 'e', 'w',
  'entrance', 'exit', 'outside', 'inside', 'in', 'out',
  'upward', 'upwards', 'downward', 'downwards',
  'ascend', 'descend',
]);

export function isBlockedDirLabel(dir) {
  const d = String(dir || '').toLowerCase().trim();
  return !d || DIR_LABEL_BLOCKLIST.has(d);
}

/** Match atlas place id to live room id (incl. R0215~instance clones). */
export function isCurrentPlace(placeId, currentRoomId) {
  if (!placeId || !currentRoomId) return false;
  if (placeId === currentRoomId) return true;
  const i = String(currentRoomId).indexOf('~');
  if (i > 0 && placeId === currentRoomId.slice(0, i)) return true;
  return false;
}

/** Label LOD from zoom: area | near | all */
export function labelLodForScale(userScale) {
  const s = Number(userScale) || 1;
  if (s < 0.85) return 'area';
  if (s < 1.2) return 'near';
  return 'all';
}

export function adjacentPlaceIds(paths, currentRoomId, places) {
  const ids = new Set();
  if (!currentRoomId) return ids;
  const hereKeys = new Set([currentRoomId]);
  const tilde = String(currentRoomId).indexOf('~');
  if (tilde > 0) hereKeys.add(currentRoomId.slice(0, tilde));
  for (const p of places || []) {
    if (isCurrentPlace(p.id, currentRoomId)) hereKeys.add(p.id);
  }
  for (const path of paths || []) {
    if (hereKeys.has(path.from)) ids.add(path.to);
    if (hereKeys.has(path.to)) ids.add(path.from);
  }
  return ids;
}

function boxesOverlap(a, b) {
  return !(a.x + a.w < b.x || b.x + b.w < a.x || a.y + a.h < b.y || b.y + b.h < a.y);
}

/**
 * Place room labels with AABB collision avoidance.
 * Higher-priority candidates (earlier) win; force=true always keeps a slot.
 */
export function layoutRoomLabels(candidates, measure) {
  const placed = [];
  const boxes = [];
  for (const c of candidates) {
    const m = measure(c.text, c.font);
    const w = Math.max(8, m.w);
    const h = Math.max(8, m.h);
    const attempts = [
      { x: c.px - w / 2, y: c.py },
      { x: c.px - w / 2, y: c.py + h + 2 },
      { x: c.px - w / 2, y: c.py - h - 6 },
      { x: c.px + 8, y: c.py },
      { x: c.px - w - 8, y: c.py },
    ];
    let chosen = null;
    for (const a of attempts) {
      const box = { x: a.x - 2, y: a.y - 2, w: w + 4, h: h + 4 };
      if (!boxes.some((b) => boxesOverlap(b, box))) {
        chosen = { ...c, x: a.x, y: a.y, w, h };
        boxes.push(box);
        break;
      }
    }
    if (!chosen && c.force) {
      const a = attempts[0];
      chosen = { ...c, x: a.x, y: a.y, w, h };
      boxes.push({ x: a.x - 2, y: a.y - 2, w: w + 4, h: h + 4 });
    }
    if (chosen) placed.push(chosen);
  }
  return placed;
}

function biomeOf(key) {
  return BIOME[key] || BIOME.wild;
}

const tileImages = Object.create(null);
let youPortraitImage = null;
let youPortraitSrc = '';
let tilesReady = false;
const tileWaiters = new Set();

function notifyTilesReady() {
  tilesReady = true;
  for (const fn of tileWaiters) fn();
}

function startTileLoad() {
  if (typeof Image === 'undefined') { tilesReady = true; return; }
  const img = new Image();
  img.onload = notifyTilesReady;
  img.onerror = notifyTilesReady;
  img.src = `/api/map-tiles/terrain-sheet.png?v=${TERRAIN_SHEET.version}`;
  tileImages.sheet = img;
}
startTileLoad();

export function setYouPortrait(url) {
  const next = String(url || '');
  if (next === youPortraitSrc) return;
  youPortraitSrc = next;
  youPortraitImage = null;
  if (!next || typeof Image === 'undefined') return;
  const img = new Image();
  img.onload = () => { youPortraitImage = img; notifyTilesReady(); };
  img.src = next;
}

export function onMapTilesReady(fn) {
  if (typeof fn !== 'function') return () => {};
  tileWaiters.add(fn);
  if (tilesReady) fn();
  return () => tileWaiters.delete(fn);
}

function tileImageReady(img) {
  return !!(img && img.complete && img.naturalWidth > 0);
}

export function tileKeyFor(place) {
  if (!place || !place.discovered || place.kind === 'uncharted') return 'fog';
  return Object.hasOwn(TERRAIN_SHEET.rows, place.terrain) && !['fog', 'sea'].includes(place.terrain)
    ? place.terrain : TERRAIN_SHEET.default;
}

function drawTerrain(ctx, key, seed, x, y, size) {
  const img = tileImages.sheet;
  const { tileSize, variants, rows } = TERRAIN_SHEET;
  if (tileImageReady(img)) {
    ctx.imageSmoothingEnabled = false;
    ctx.drawImage(img, (hashString(seed) % variants) * tileSize, rows[key] * tileSize,
      tileSize, tileSize, x, y, size, size);
  } else {
    ctx.fillStyle = key === 'fog' ? '#746a51' : key === 'sea' ? '#173947' : '#649c48';
    ctx.fillRect(x, y, size, size);
  }
}

function hashString(str) {
  let h = 0;
  for (let i = 0; i < str.length; i++) {
    h = ((h << 5) - h) + str.charCodeAt(i);
    h |= 0;
  }
  return Math.abs(h);
}

/** Chebyshev distance in layout units between two places. */
export function layoutDistance(a, b) {
  if (!a || !b) return Infinity;
  return Math.max(Math.abs(Math.round(a.x) - Math.round(b.x)), Math.abs(Math.round(a.y) - Math.round(b.y)));
}

/**
 * Prefer framing you + nearby rooms so distant separateAreas (Oldtown vs Meadow)
 * do not shrink the camera to a world-fit. Falls back to all places when sparse.
 * @param {object[]} places
 * @param {object|null} focus
 * @param {{ maxDist?: number, minCount?: number }} [opts]
 */
export function selectFramingPlaces(places, focus, opts = {}) {
  const list = places || [];
  if (!list.length) return [];
  if (!focus) return list;
  const maxDist = opts.maxDist != null ? opts.maxDist : 10;
  const minCount = opts.minCount != null ? opts.minCount : 3;
  const near = list.filter((p) => layoutDistance(p, focus) <= maxDist);
  // Always include focus; expand radius once if too few neighbors.
  if (near.length >= minCount) return near;
  const wider = list.filter((p) => layoutDistance(p, focus) <= maxDist * 1.6);
  if (wider.length >= 2) return wider;
  return list;
}

/**
 * BFS neighborhood by atlas path edges (depth ≤ maxDepth).
 * Useful when layout coords are sparse but graph connectivity is dense.
 */
export function placesWithinBfsDepth(places, paths, focusId, maxDepth = 4) {
  const list = places || [];
  if (!focusId || !list.length) return list;
  const byId = new Map(list.map((p) => [p.id, p]));
  if (!byId.has(focusId)) {
    // template match for instanced rooms
    for (const p of list) {
      if (isCurrentPlace(p.id, focusId)) {
        focusId = p.id;
        break;
      }
    }
  }
  if (!byId.has(focusId)) return list;
  const adj = new Map();
  for (const path of paths || []) {
    if (!byId.has(path.from) || !byId.has(path.to)) continue;
    if (!adj.has(path.from)) adj.set(path.from, []);
    if (!adj.has(path.to)) adj.set(path.to, []);
    adj.get(path.from).push(path.to);
    adj.get(path.to).push(path.from);
  }
  const out = new Map();
  const queue = [{ id: focusId, depth: 0 }];
  out.set(focusId, byId.get(focusId));
  while (queue.length) {
    const { id, depth } = queue.shift();
    if (depth >= maxDepth) continue;
    for (const next of adj.get(id) || []) {
      if (out.has(next)) continue;
      out.set(next, byId.get(next));
      queue.push({ id: next, depth: depth + 1 });
    }
  }
  return out.size >= 2 ? [...out.values()] : list;
}

function computeCamera(places, w, h, panX, panY, userScale, focus = null, paths = null) {
  const frame = focus
    ? (() => {
        const byDist = selectFramingPlaces(places, focus, { maxDist: 10, minCount: 3 });
        if (byDist.length >= 3 || !paths) return byDist;
        const byBfs = placesWithinBfsDepth(places, paths, focus.id, 4);
        return byBfs.length >= byDist.length ? byBfs : byDist;
      })()
    : places;
  const use = frame && frame.length ? frame : places;
  let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
  for (const p of use) {
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
  const pad = Math.max(24, Math.min(w, h) * 0.07);
  const fit = Math.min((w - pad * 2) / spanX, (h - pad * 2) / spanY);
  const tileStep = Math.max(3, Math.min(fit * userScale, MAP_TILE_STEP_MAX));
  return {
    tileStep,
    ox: (minX + maxX) / 2,
    oy: (minY + maxY) / 2,
    panX,
    panY,
  };
}

/** Pan offsets so `place` sits at the viewport center (keep userScale). */
export function panToCenterPlace(places, place, w, h, userScale, paths = null) {
  if (!place || !places || !places.length || w < 1 || h < 1) {
    return { panX: 0, panY: 0 };
  }
  const cam = computeCamera(places, w, h, 0, 0, userScale, place, paths);
  const gx = Math.round(place.x);
  const gy = Math.round(place.y);
  return {
    panX: -(gx - cam.ox) * cam.tileStep,
    panY: -(gy - cam.oy) * cam.tileStep,
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

export const MAP_SCALE_MIN = 0.12;
export const MAP_SCALE_MAX = 5;
export const MAP_TILE_STEP_MAX = 110;

export function clampMapScale(s) {
  const n = Number(s);
  if (!isFinite(n) || n <= 0) return 1;
  return Math.min(MAP_SCALE_MAX, Math.max(MAP_SCALE_MIN, n));
}

function tileHalf(tileStep) {
  return tileStep * 0.5;
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

function drawParchmentBg(ctx, w, h) {
  // One cached repeat pattern; no per-room DOM images or per-frame texture generation.
  if (tileImageReady(tileImages.sheet) && typeof document !== 'undefined') {
    if (!tileImages.seaPattern) {
      const texture = document.createElement('canvas');
      texture.width = texture.height = 96;
      const tc = texture.getContext('2d');
      drawTerrain(tc, 'sea', 'backdrop', 0, 0, 96);
      tileImages.seaPattern = ctx.createPattern(texture, 'repeat');
    }
    ctx.fillStyle = tileImages.seaPattern;
  } else ctx.fillStyle = '#173947';
  ctx.fillRect(0, 0, w, h);
  const wash = ctx.createRadialGradient(w / 2, h / 2, 0, w / 2, h / 2, Math.max(w, h) * 0.75);
  wash.addColorStop(0, 'rgba(65, 110, 119, 0.12)');
  wash.addColorStop(1, 'rgba(4, 17, 28, 0.5)');
  ctx.fillStyle = wash;
  ctx.fillRect(0, 0, w, h);
}

/** Readable area-name size that tracks map zoom (tileStep). */
function areaLabelFontSize(cam) {
  const step = cam && cam.tileStep ? cam.tileStep : 40;
  return Math.max(11, Math.min(18, Math.round(step * 0.3)));
}

/**
 * Clear area labels for tinted region/area groups.
 * Prefer atlas regions (same hulls as the wash rects); fall back to place.area clusters.
 * Always drawn (not LOD-gated) so overworld clusters stay named at default zoom.
 */
function drawAreaLabels(ctx, places, regions, cam, w, h) {
  const fontSize = areaLabelFontSize(cam);
  const font = `700 ${fontSize}px Georgia, serif`;
  const labelH = fontSize + 4;
  const candidates = [];
  const labeled = new Set();

  for (const region of regions || []) {
    const text = String(region.name || '').trim();
    if (!text) continue;
    const pts = (region.hull || []).map(([x, y]) => projectGrid(Math.round(x), Math.round(y), cam, w, h));
    if (!pts.length) continue;
    let minPx = Infinity, maxPx = -Infinity, minPy = Infinity, maxPy = -Infinity;
    for (const p of pts) {
      if (p.px < minPx) minPx = p.px;
      if (p.px > maxPx) maxPx = p.px;
      if (p.py < minPy) minPy = p.py;
      if (p.py > maxPy) maxPy = p.py;
    }
    const pad = cam.tileStep * 0.5;
    candidates.push({
      text,
      px: (minPx + maxPx) / 2,
      py: minPy - pad - labelH - 2,
      font,
      force: true,
      priority: -(region.places ? region.places.length : pts.length),
    });
    labeled.add(text.toLowerCase());
  }

  // Fallback for discovered place clusters that somehow lack a region hull.
  const byArea = new Map();
  for (const p of places || []) {
    if (!p.discovered || !p.area) continue;
    if (!byArea.has(p.area)) byArea.set(p.area, []);
    byArea.get(p.area).push(p);
  }
  const cell = cam.tileStep * 0.92;
  for (const [area, rooms] of byArea) {
    if (rooms.length < 1) continue;
    const text = String(rooms[0].areaName || '').trim();
    if (!text || labeled.has(text.toLowerCase())) continue;
    let minPx = Infinity, maxPx = -Infinity, minPy = Infinity, maxPy = -Infinity;
    for (const p of rooms) {
      const { px, py } = projectPlace(p, cam, w, h);
      if (px < minPx) minPx = px;
      if (px > maxPx) maxPx = px;
      if (py < minPy) minPy = py;
      if (py > maxPy) maxPy = py;
    }
    const pad = cell * 0.55;
    candidates.push({
      text,
      px: (minPx + maxPx) / 2,
      py: minPy - pad - labelH - 2,
      font,
      force: true,
      priority: -rooms.length,
    });
    labeled.add(text.toLowerCase());
  }

  if (!candidates.length) return;
  candidates.sort((a, b) => a.priority - b.priority);
  const measure = (text, f) => {
    ctx.font = f;
    const metrics = ctx.measureText(text);
    return { w: metrics.width, h: labelH };
  };
  const placed = layoutRoomLabels(candidates, measure);
  for (const lab of placed) {
    ctx.font = lab.font;
    ctx.textAlign = 'left';
    ctx.textBaseline = 'top';
    ctx.lineWidth = Math.max(3, Math.round(fontSize * 0.28));
    ctx.strokeStyle = 'rgba(16, 12, 6, 0.88)';
    ctx.strokeText(lab.text, lab.x, lab.y);
    ctx.fillStyle = '#f0d78c';
    ctx.fillText(lab.text, lab.x, lab.y);
  }
}

function drawCorridor(ctx, a, b, pa, pb, path, cam, onTravel) {
  const half = tileHalf(cam.tileStep);
  const { dx, dy } = worldDelta(a, b);
  const from = tileEdgePoint(pa.px, pa.py, half, dx, dy);
  const to = tileEdgePoint(pb.px, pb.py, half, -dx, -dy);
  const cross = isCrossArea(a, b);
  const bothKnown = a.discovered && b.discovered;

  ctx.beginPath();
  ctx.moveTo(from.px, from.py);
  ctx.lineTo(to.px, to.py);
  ctx.strokeStyle = onTravel ? '#5ee7ff' : cross ? 'rgba(210, 120, 70, 0.8)' : 'rgba(196, 168, 110, 0.72)';
  ctx.lineWidth = onTravel ? 2.4 : path.kind === 'road' ? 1.8 : 1.45;
  ctx.globalAlpha = bothKnown ? 0.95 : 0.4;
  ctx.setLineDash(onTravel ? [7, 4] : [4, 5]);
  ctx.lineCap = 'butt';
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

  // Exit tick on the portal end — never paint compass/vertical dir words.
  ctx.beginPath();
  roundRect(ctx, portalX - 5, portalY - 4, 10, 8, 2);
  ctx.fillStyle = discovered ? 'rgba(36, 28, 20, 0.9)' : 'rgba(24, 20, 16, 0.65)';
  ctx.fill();
  ctx.strokeStyle = cross ? 'rgba(200, 110, 70, 0.85)' : color;
  ctx.lineWidth = 1.2;
  ctx.stroke();

  if (label && !isBlockedDirLabel(label)) {
    const text = label.length > 9 ? label.slice(0, 8) + '…' : label;
    ctx.font = '600 7px sans-serif';
    ctx.fillStyle = discovered ? 'rgba(232, 220, 190, 0.9)' : 'rgba(148, 140, 128, 0.6)';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'bottom';
    ctx.fillText(text, portalX, portalY - 7);
  }
}

function drawSilhouette(ctx, px, py, size) {
  ctx.fillStyle = '#e8d5a8';
  ctx.beginPath();
  ctx.arc(px, py - size * 0.18, size * 0.16, 0, Math.PI * 2);
  ctx.fill();
  ctx.beginPath();
  ctx.moveTo(px, py - size * 0.04);
  ctx.quadraticCurveTo(px + size * 0.22, py + size * 0.32, px, py + size * 0.34);
  ctx.quadraticCurveTo(px - size * 0.22, py + size * 0.32, px, py - size * 0.04);
  ctx.fill();
}

function drawYouMarker(ctx, px, py, half, portraitImg) {
  ctx.save();
  ctx.beginPath();
  ctx.arc(px, py, Math.max(14, half * 0.85), 0, Math.PI * 2);
  ctx.shadowColor = '#ffdc78';
  ctx.shadowBlur = 20;
  ctx.strokeStyle = '#ffe69b';
  ctx.lineWidth = 3;
  ctx.stroke();
  ctx.restore();
  const size = Math.max(18, Math.min(26, half * 0.72));
  const x = px - size / 2;
  const y = py - size / 2;
  ctx.save();
  ctx.shadowColor = 'rgba(0, 0, 0, 0.55)';
  ctx.shadowBlur = 6;
  ctx.shadowOffsetY = 2;
  ctx.fillStyle = 'rgba(18, 14, 10, 0.88)';
  roundRect(ctx, x, y, size, size, 4);
  ctx.fill();
  ctx.shadowBlur = 0;
  ctx.shadowOffsetY = 0;
  ctx.strokeStyle = '#e0b84a';
  ctx.lineWidth = 1.6;
  roundRect(ctx, x, y, size, size, 4);
  ctx.stroke();
  const inset = 2.5;
  if (tileImageReady(portraitImg)) {
    ctx.save();
    roundRect(ctx, x + inset, y + inset, size - inset * 2, size - inset * 2, 3);
    ctx.clip();
    ctx.imageSmoothingEnabled = false;
    ctx.drawImage(portraitImg, x + inset, y + inset, size - inset * 2, size - inset * 2);
    ctx.restore();
  } else {
    drawSilhouette(ctx, px, py + 1, size);
  }
  ctx.restore();
}

function drawTile(ctx, place, px, py, tileStep, opts) {
  const half = tileHalf(tileStep);
  // Round shared boundaries independently to avoid subpixel seams while panning.
  const x = Math.round(px - half), y = Math.round(py - half);
  const size = Math.ceil(tileStep);
  drawTerrain(ctx, tileKeyFor(place), place.id, x, y, size);
  if (place.id === opts.travelTargetId) {
    ctx.strokeStyle = '#22d3ee'; ctx.lineWidth = 2;
    ctx.strokeRect(x, y, size, size);
  }
  if (opts.selected) drawCornerBrackets(ctx, x - 2, y - 2, size + 4, '#ffe29a');
  return half;
}

/** Short same-zone exit gaps become ground; never cross fog or another layer. */
export function terrainConnectors(places, paths) {
  const byId = new Map(places.map(p => [p.id, p]));
  const occupied = new Set(places.map(p => `${p.layer}:${Math.round(p.x)}:${Math.round(p.y)}`));
  const fill = new Map();
  for (const path of paths) {
    const a = byId.get(path.from), b = byId.get(path.to);
    if (!a || !b || !a.discovered || !b.discovered || a.area !== b.area || a.layer !== b.layer ||
      !COMPASS_DIRS.has(String(path.dir).toLowerCase()) || ['hidden', 'stair', 'passage'].includes(path.kind)) continue;
    const { dx, dy } = worldDelta(a, b);
    const distance = Math.max(Math.abs(dx), Math.abs(dy));
    if (distance < 2 || distance > 3 || (dx && dy && Math.abs(dx) !== Math.abs(dy))) continue;
    const dominant = a.id < b.id ? a : b;
    for (let i = 1; i < distance; i++) {
      const x = Math.round(a.x) + dx / distance * i, y = Math.round(a.y) + dy / distance * i;
      const key = `${a.layer}:${x}:${y}`;
      if (!occupied.has(key) && !fill.has(key)) fill.set(key, { ...dominant, id: key, x, y });
    }
  }
  return [...fill.values()];
}

function drawCornerBrackets(ctx, x, y, size, color) {
  const L = Math.max(5, size * 0.16);
  ctx.save();
  ctx.strokeStyle = color;
  ctx.lineWidth = Math.max(1.4, size * 0.035);
  ctx.lineCap = 'square';
  ctx.beginPath();
  ctx.moveTo(x, y + L); ctx.lineTo(x, y); ctx.lineTo(x + L, y);
  ctx.moveTo(x + size - L, y); ctx.lineTo(x + size, y); ctx.lineTo(x + size, y + L);
  ctx.moveTo(x, y + size - L); ctx.lineTo(x, y + size); ctx.lineTo(x + L, y + size);
  ctx.moveTo(x + size - L, y + size); ctx.lineTo(x + size, y + size); ctx.lineTo(x + size, y + size - L);
  ctx.stroke();
  ctx.restore();
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
    selectedId = null,
    frameWorld = false,
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

  const byId = {};
  for (const p of atlas.places || []) byId[p.id] = p;

  // Exactly one you-are-here place: prefer exact id, else template match.
  let herePlace = visiblePlaces.find((p) => p.id === currentRoomId) || null;
  if (!herePlace) {
    herePlace = visiblePlaces.find((p) => isCurrentPlace(p.id, currentRoomId)) || null;
  }
  const hereId = herePlace ? herePlace.id : null;

  // Frame you + nearby so separateAreas (Oldtown vs Meadow) follow the player.
  const cam = computeCamera(visiblePlaces, w, h, panX, panY, userScale, frameWorld ? null : herePlace, atlas.paths || []);
  const lod = labelLodForScale(userScale);
  const nearIds = lod === 'near'
    ? adjacentPlaceIds(atlas.paths || [], currentRoomId, visiblePlaces)
    : new Set();

  for (const place of terrainConnectors(visiblePlaces, atlas.paths || [])) {
    const { px, py } = projectPlace(place, cam, w, h);
    drawTile(ctx, place, px, py, cam.tileStep, {});
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
    if (!isGridLink(path, a, b)) continue;
    const pa = projectPlace(a, cam, w, h);
    const pb = projectPlace(b, cam, w, h);
    const onTravel =
      travelPathRoomIds.has(path.from) && travelPathRoomIds.has(path.to) ||
      travelPathRoomIds.has(path.to) && path.from === currentRoomId;
    if (onTravel || isCrossArea(a, b)) drawCorridor(ctx, a, b, pa, pb, path, cam, onTravel);
  }

  for (const path of layerPaths) {
    const a = byId[path.from];
    const b = byId[path.to];
    if (!a.discovered || isGridLink(path, a, b)) continue;
    if (a.layer === b.layer && a.area === b.area && COMPASS_DIRS.has(path.dir) && layoutDistance(a, b) <= 3) continue;
    const pa = projectPlace(a, cam, w, h);
    const pb = projectPlace(b, cam, w, h);
    const biome = biomeOf(a.biome);
    // Never paint compass/vertical dir strings — ticks only.
    if (cam.tileStep < 28) continue;
    const label = isBlockedDirLabel(path.dir) ? '' : path.dir;
    drawPortalLink(ctx, pa.px, pa.py, pb.px, pb.py, label, biome.ink, b.discovered, isCrossArea(a, b));
  }

  const hits = [];
  const sorted = [...visiblePlaces].sort((a, b) => {
    if (a.discovered === b.discovered) return 0;
    return a.discovered ? 1 : -1;
  });

  let herePx = null;
  let herePy = null;
  let hereHalf = 0;

  for (const place of sorted) {
    const { px, py } = projectPlace(place, cam, w, h);
    const isHere = place.id === hereId;
    const half = tileHalf(cam.tileStep);
    if (px + half < 0 || py + half < 0 || px - half > w || py - half > h) continue;
    const r = drawTile(ctx, place, px, py, cam.tileStep, {
      travelTargetId,
      isHere,
      selected: selectedId && place.id === selectedId,
    });
    if (px + r >= 0 && py + r >= 0 && px - r <= w && py - r <= h) hits.push({ px, py, r: r + 4, half: r, place: { ...place, current: isHere } });
    if (isHere) {
      herePx = px;
      herePy = py;
      hereHalf = r;
    }
  }

  // Soft trail: dim gold ring on the previous step along travel path (optional).
  if (travelPathRoomIds.size && hereId) {
    for (const id of travelPathRoomIds) {
      if (id === hereId) continue;
      const p = byId[id];
      if (!p || p.layer !== activeLayer) continue;
      const { px, py } = projectPlace(p, cam, w, h);
      const half = tileHalf(cam.tileStep);
      ctx.beginPath();
      ctx.arc(px, py, half * 0.35, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(212, 160, 48, 0.18)';
      ctx.fill();
    }
  }

  if (herePx != null) {
    drawYouMarker(ctx, herePx, herePy, hereHalf, youPortraitImage);
  }

  drawAreaLabels(ctx, visiblePlaces, visibleRegions, cam, w, h);

  if (cam.tileStep >= 20 && (lod === 'near' || lod === 'all')) {
    const candidates = [];
    for (const place of visiblePlaces) {
      if (!place.discovered || !place.name) continue;
      const isHere = place.id === hereId;
      if (!isHere && place.id !== selectedId && (cam.tileStep < 60 || lod === 'near' && !nearIds.has(place.id))) continue;
      const { px, py } = projectPlace(place, cam, w, h);
      const half = tileHalf(cam.tileStep);
      candidates.push({
        id: place.id,
        text: place.name,
        px,
        py: py + half + 4,
        font: isHere ? '700 10px Georgia, serif' : '600 9px Georgia, serif',
        fill: isHere ? '#e8c060' : '#e8dcc8',
        force: isHere,
        priority: isHere ? 0 : nearIds.has(place.id) ? 1 : 2,
      });
    }
    candidates.sort((a, b) => a.priority - b.priority);
    const measure = (text, font) => {
      ctx.font = font;
      const metrics = ctx.measureText(text);
      return { w: metrics.width, h: 12 };
    };
    const placed = layoutRoomLabels(candidates, measure);
    for (const lab of placed) {
      ctx.font = lab.font;
      ctx.textAlign = 'left';
      ctx.textBaseline = 'top';
      ctx.lineWidth = 3;
      ctx.strokeStyle = 'rgba(20, 14, 8, 0.8)';
      ctx.strokeText(lab.text, lab.x, lab.y);
      ctx.fillStyle = lab.fill;
      ctx.fillText(lab.text, lab.x, lab.y);
    }
  }

  return { hits };
}

export { BIOME, projectPlace, computeCamera, COMPASS_DIRS, DIR_LABEL_BLOCKLIST };
