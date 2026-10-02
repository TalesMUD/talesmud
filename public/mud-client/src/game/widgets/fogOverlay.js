import { hash, makeCanvas } from './mapArt.js';

/** Separable box blur for OffscreenCanvas (CSS filter is unreliable in workers). */
function boxBlurAlpha(src, passes = 2) {
  const w = src.width, h = src.height;
  const ctx = src.getContext('2d', { willReadFrequently: true });
  let img = ctx.getImageData(0, 0, w, h);
  const radius = 4;
  for (let pass = 0; pass < passes; pass++) {
    const srcData = img.data;
    const tmp = new Uint8ClampedArray(srcData.length);
    for (let y = 0; y < h; y++) {
      for (let x = 0; x < w; x++) {
        let a = 0, n = 0;
        for (let k = -radius; k <= radius; k++) {
          const xx = x + k;
          if (xx < 0 || xx >= w) continue;
          a += srcData[(y * w + xx) * 4 + 3];
          n++;
        }
        const i = (y * w + x) * 4;
        tmp[i] = tmp[i + 1] = tmp[i + 2] = 255;
        tmp[i + 3] = (a / n) | 0;
      }
    }
    const out = new Uint8ClampedArray(srcData.length);
    for (let y = 0; y < h; y++) {
      for (let x = 0; x < w; x++) {
        let a = 0, n = 0;
        for (let k = -radius; k <= radius; k++) {
          const yy = y + k;
          if (yy < 0 || yy >= h) continue;
          a += tmp[(yy * w + x) * 4 + 3];
          n++;
        }
        const i = (y * w + x) * 4;
        out[i] = out[i + 1] = out[i + 2] = 255;
        out[i + 3] = (a / n) | 0;
      }
    }
    img = ctx.createImageData(w, h);
    img.data.set(out);
  }
  ctx.putImageData(img, 0, 0);
  return src;
}

/**
 * Soft volumetric fog overlay for unexplored landscape cells.
 * Feathered explore boundary via blurred mask; cloud washes instead of per-tile fog stamps.
 */
export function buildSoftFogOverlay(cells, bounds, logicalW, logicalH, options = {}) {
  const workScale = options.workScale ?? 2;
  const blurPasses = options.blurPasses ?? 5;
  const blurRadius = options.blurRadius ?? 4;
  const opacity = options.opacity ?? 0.86;
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
    mc.fillRect((c.x - bounds.minX) * cell - 1.2, (c.y - bounds.minY) * cell - 1.2, cell + 2.4, cell + 2.4);
  }
  if (!fogCount) return null;

  const soft = makeCanvas(mw, mh);
  const sc = soft.getContext('2d');
  try {
    sc.filter = `blur(${Math.max(3, 10 / workScale)}px)`;
    sc.drawImage(mask, 0, 0);
    sc.filter = 'none';
  } catch (_) {
    sc.drawImage(mask, 0, 0);
  }
  boxBlurAlpha(soft, 5);

  const fog = makeCanvas(logicalW, logicalH);
  const fc = fog.getContext('2d');
  fc.fillStyle = `rgba(8, 16, 26, ${opacity})`;
  fc.fillRect(0, 0, logicalW, logicalH);

  const blobs = [];
  for (const c of cells) {
    if (c.terrain !== 'fog') continue;
    const n = hash(`${c.x}:${c.y}:fogcloud`);
    if (n % 4 > 2) continue;
    blobs.push({
      cx: (c.x - bounds.minX + 0.5) * 32 + ((n % 7) - 3),
      cy: (c.y - bounds.minY + 0.5) * 32 + (((n >> 3) % 7) - 3),
      rx: Math.max(8, 22 + (n % 20)),
      ry: Math.max(8, 16 + ((n >> 4) % 18)),
      a: 0.14 + (n % 9) * 0.016,
      rot: ((n % 100) / 100) * 1.1,
    });
  }
  for (let i = 0; i < Math.min(64, Math.ceil(fogCount / 4)); i++) {
    const n = hash(`drift:${i}:${bounds.minX}:${bounds.minY}:${fogCount}`);
    const cx = (n % Math.max(1, logicalW));
    const cy = ((n >> 9) % Math.max(1, logicalH));
    const gx = Math.floor(cx / 32) + bounds.minX;
    const gy = Math.floor(cy / 32) + bounds.minY;
    const nearFog = cells.some(c => c.terrain === 'fog' && Math.abs(c.x - gx) <= 3 && Math.abs(c.y - gy) <= 3);
    if (!nearFog) continue;
    blobs.push({
      cx, cy,
      rx: Math.max(12, 40 + (n % 55)),
      ry: Math.max(12, 28 + ((n >> 5) % 42)),
      a: 0.1 + (n % 8) * 0.012,
      rot: ((n % 80) / 80) * 0.9,
    });
  }

  for (const b of blobs) {
    const rx = Math.max(1, b.rx);
    const ry = Math.max(1, b.ry);
    const rad = Math.max(rx, ry);
    const g = fc.createRadialGradient(b.cx, b.cy, 0, b.cx, b.cy, rad);
    g.addColorStop(0, `rgba(96, 118, 136, ${b.a})`);
    g.addColorStop(0.4, `rgba(36, 54, 72, ${b.a * 0.9})`);
    g.addColorStop(1, 'rgba(8, 14, 22, 0)');
    fc.fillStyle = g;
    fc.beginPath();
    fc.ellipse(b.cx, b.cy, rx, ry, b.rot || 0, 0, Math.PI * 2);
    fc.fill();
  }

  fc.globalCompositeOperation = 'lighter';
  for (const c of cells) {
    if (c.terrain !== 'fog') continue;
    const n = hash(`${c.x}:${c.y}:wisp`);
    if (n % 9 !== 0) continue;
    const cx = (c.x - bounds.minX + 0.5) * 32;
    const cy = (c.y - bounds.minY + 0.5) * 32;
    const g = fc.createRadialGradient(cx, cy, 0, cx, cy, 30);
    g.addColorStop(0, 'rgba(186, 206, 224, 0.07)');
    g.addColorStop(1, 'rgba(186, 206, 224, 0)');
    fc.fillStyle = g;
    fc.fillRect(cx - 32, cy - 32, 64, 64);
  }
  fc.globalCompositeOperation = 'source-over';

  fc.globalCompositeOperation = 'destination-in';
  fc.imageSmoothingEnabled = true;
  fc.drawImage(soft, 0, 0, logicalW, logicalH);
  fc.globalCompositeOperation = 'source-over';
  return fog;
}

export function paintUndergroundSoftFog(ctx, places, native, size = 32) {
  const fogged = (places || []).filter(p => !p.discovered);
  if (!fogged.length) return;
  const pad = size * 2.2;
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
    mc.arc(at.x - minX, at.y - minY, size * 0.85, 0, Math.PI * 2);
    mc.fill();
  }
  const soft = makeCanvas(w, h);
  const sc = soft.getContext('2d');
  try {
    sc.filter = 'blur(12px)';
    sc.drawImage(mask, 0, 0);
    sc.filter = 'none';
  } catch (_) {
    sc.drawImage(mask, 0, 0);
  }
  boxBlurAlpha(soft, 5);

  const fog = makeCanvas(w, h);
  const fc = fog.getContext('2d');
  fc.fillStyle = 'rgba(10, 18, 24, 0.92)';
  fc.fillRect(0, 0, w, h);
  for (const p of fogged) {
    const at = native(p.x, p.y);
    const n = hash(`${p.id}:ugfog`);
    const g = fc.createRadialGradient(at.x - minX, at.y - minY, 2, at.x - minX, at.y - minY, size * 1.25);
    g.addColorStop(0, `rgba(72, 90, 98, ${0.24 + (n % 5) * 0.03})`);
    g.addColorStop(1, 'rgba(10, 18, 24, 0)');
    fc.fillStyle = g;
    fc.fillRect(at.x - minX - size, at.y - minY - size, size * 2, size * 2);
  }
  fc.globalCompositeOperation = 'destination-in';
  fc.imageSmoothingEnabled = true;
  fc.drawImage(soft, 0, 0);
  ctx.drawImage(fog, minX, minY);
}
