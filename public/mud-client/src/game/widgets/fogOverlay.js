import { hash, makeCanvas } from './mapArt.js';

/**
 * Soft volumetric fog overlay for unexplored landscape cells.
 * Feathered explore boundary via blurred mask; cloud washes instead of per-tile fog stamps.
 * Returns a canvas in logical map units (same space as continentRaster ground), or null.
 */
export function buildSoftFogOverlay(cells, bounds, logicalW, logicalH, options = {}) {
  const blurPx = options.blurPx ?? 16;
  const workScale = options.workScale ?? 4;
  const opacity = options.opacity ?? 0.88;
  const mw = Math.max(1, Math.ceil(logicalW / workScale));
  const mh = Math.max(1, Math.ceil(logicalH / workScale));
  const cell = 32 / workScale;

  const mask = makeCanvas(mw, mh);
  const mc = mask.getContext('2d');
  mc.clearRect(0, 0, mw, mh);
  mc.fillStyle = '#fff';
  let fogCount = 0;
  for (const c of cells || []) {
    if (c.terrain !== 'fog') continue;
    fogCount++;
    // Slight overfill so blur fills cell seams without a checker.
    mc.fillRect((c.x - bounds.minX) * cell - 0.25, (c.y - bounds.minY) * cell - 0.25, cell + 0.5, cell + 0.5);
  }
  if (!fogCount) return null;

  const soft = makeCanvas(mw, mh);
  const sc = soft.getContext('2d');
  sc.filter = `blur(${Math.max(2.5, blurPx / workScale)}px)`;
  sc.drawImage(mask, 0, 0);
  sc.filter = 'none';

  const fog = makeCanvas(logicalW, logicalH);
  const fc = fog.getContext('2d');

  // Deep cool haze — thick enough to hide undisclosed art, soft enough for clouds.
  fc.fillStyle = `rgba(10, 18, 28, ${opacity})`;
  fc.fillRect(0, 0, logicalW, logicalH);

  const blobs = [];
  for (const c of cells) {
    if (c.terrain !== 'fog') continue;
    const n = hash(`${c.x}:${c.y}:fogcloud`);
    if (n % 5 > 2) continue;
    blobs.push({
      cx: (c.x - bounds.minX + 0.5) * 32 + ((n % 7) - 3),
      cy: (c.y - bounds.minY + 0.5) * 32 + (((n >> 3) % 7) - 3),
      rx: 20 + (n % 18),
      ry: 15 + ((n >> 4) % 16),
      a: 0.12 + (n % 9) * 0.016,
      rot: ((n % 100) / 100) * 1.1,
    });
  }
  for (let i = 0; i < Math.min(56, Math.ceil(fogCount / 5)); i++) {
    const n = hash(`drift:${i}:${bounds.minX}:${bounds.minY}:${fogCount}`);
    const cx = (n % Math.max(1, logicalW));
    const cy = ((n >> 9) % Math.max(1, logicalH));
    const gx = Math.floor(cx / 32) + bounds.minX;
    const gy = Math.floor(cy / 32) + bounds.minY;
    const nearFog = cells.some(c => c.terrain === 'fog' && Math.abs(c.x - gx) <= 3 && Math.abs(c.y - gy) <= 3);
    if (!nearFog) continue;
    blobs.push({
      cx, cy,
      rx: 36 + (n % 55),
      ry: 26 + ((n >> 5) % 42),
      a: 0.09 + (n % 8) * 0.012,
      rot: ((n % 80) / 80) * 0.9,
    });
  }

  for (const b of blobs) {
    const rad = Math.max(b.rx, b.ry);
    const g = fc.createRadialGradient(b.cx, b.cy, 0, b.cx, b.cy, rad);
    g.addColorStop(0, `rgba(92, 112, 128, ${b.a})`);
    g.addColorStop(0.4, `rgba(40, 58, 74, ${b.a * 0.9})`);
    g.addColorStop(1, 'rgba(8, 14, 22, 0)');
    fc.fillStyle = g;
    fc.beginPath();
    fc.ellipse(b.cx, b.cy, b.rx, b.ry, b.rot, 0, Math.PI * 2);
    fc.fill();
  }

  // Soft pearl wisps — cloud volume, not tile noise.
  fc.globalCompositeOperation = 'lighter';
  for (const c of cells) {
    if (c.terrain !== 'fog') continue;
    const n = hash(`${c.x}:${c.y}:wisp`);
    if (n % 10 !== 0) continue;
    const cx = (c.x - bounds.minX + 0.5) * 32;
    const cy = (c.y - bounds.minY + 0.5) * 32;
    const g = fc.createRadialGradient(cx, cy, 0, cx, cy, 26);
    g.addColorStop(0, 'rgba(180, 200, 220, 0.06)');
    g.addColorStop(1, 'rgba(180, 200, 220, 0)');
    fc.fillStyle = g;
    fc.fillRect(cx - 28, cy - 28, 56, 56);
  }
  fc.globalCompositeOperation = 'source-over';

  // Feather at explore boundary.
  fc.globalCompositeOperation = 'destination-in';
  fc.imageSmoothingEnabled = true;
  fc.drawImage(soft, 0, 0, logicalW, logicalH);
  fc.globalCompositeOperation = 'source-over';
  return fog;
}

/** Soft fog wash over undiscovered underground rooms (no fog tile stamps). */
export function paintUndergroundSoftFog(ctx, places, native, size = 32) {
  const fogged = (places || []).filter(p => !p.discovered);
  if (!fogged.length) return;
  const pad = size * 1.9;
  const xs = fogged.map(p => native(p.x, p.y).x);
  const ys = fogged.map(p => native(p.x, p.y).y);
  const minX = Math.floor(Math.min(...xs) - pad);
  const minY = Math.floor(Math.min(...ys) - pad);
  const maxX = Math.ceil(Math.max(...xs) + pad);
  const maxY = Math.ceil(Math.max(...ys) + pad);
  const w = Math.max(1, maxX - minX);
  const h = Math.max(1, maxY - minY);
  const mask = makeCanvas(w, h);
  const mc = mask.getContext('2d');
  mc.fillStyle = '#fff';
  for (const p of fogged) {
    const at = native(p.x, p.y);
    mc.beginPath();
    mc.arc(at.x - minX, at.y - minY, size * 0.75, 0, Math.PI * 2);
    mc.fill();
  }
  const soft = makeCanvas(w, h);
  const sc = soft.getContext('2d');
  sc.filter = 'blur(11px)';
  sc.drawImage(mask, 0, 0);
  sc.filter = 'none';

  const fog = makeCanvas(w, h);
  const fc = fog.getContext('2d');
  fc.fillStyle = 'rgba(12, 20, 26, 0.9)';
  fc.fillRect(0, 0, w, h);
  for (const p of fogged) {
    const at = native(p.x, p.y);
    const n = hash(`${p.id}:ugfog`);
    const g = fc.createRadialGradient(at.x - minX, at.y - minY, 2, at.x - minX, at.y - minY, size * 1.15);
    g.addColorStop(0, `rgba(72, 90, 98, ${0.22 + (n % 5) * 0.03})`);
    g.addColorStop(1, 'rgba(12, 20, 26, 0)');
    fc.fillStyle = g;
    fc.fillRect(at.x - minX - size, at.y - minY - size, size * 2, size * 2);
  }
  fc.globalCompositeOperation = 'destination-in';
  fc.imageSmoothingEnabled = true;
  fc.drawImage(soft, 0, 0);
  ctx.drawImage(fog, minX, minY);
}
