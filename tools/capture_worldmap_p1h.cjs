// Local-only P1h prototype capture. Does not touch production.
// Usage: WORLDMAP_PREVIEW_URL=http://127.0.0.1:8151/ WORLDMAP_P1H_LABEL=a \
//   PUPPETEER_MODULE=$HOME/dev/node_modules/puppeteer-core \
//   node tools/capture_worldmap_p1h.cjs
const fs = require('node:fs'), path = require('node:path');
const puppeteer = require(process.env.PUPPETEER_MODULE || 'puppeteer-core');
const label = process.env.WORLDMAP_P1H_LABEL || 'a';
const out = path.resolve(__dirname, '../.director/ux-audit/after');
fs.mkdirSync(out, { recursive: true });
const settle = page => page.evaluate(async () => {
  await document.fonts.ready;
  await document.fonts.load('24px "Material Icons"').catch(() => {});
  await Promise.all([...document.images].map(i => i.decode().catch(() => {})));
  for (let i = 0; i < 3; i++) {
    await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)));
    if (window.__mapPreview?.waitForMapScenes) await window.__mapPreview.waitForMapScenes();
  }
});
(async () => {
  const browser = await puppeteer.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser',
    headless: 'new',
    args: ['--no-sandbox', '--disable-dev-shm-usage']
  });
  try {
    const page = await browser.newPage();
    const errors = [];
    page.on('pageerror', e => errors.push(e.message));
    await page.evaluateOnNewDocument(() => {
      const draw = CanvasRenderingContext2D.prototype.drawImage;
      CanvasRenderingContext2D.prototype.drawImage = function (...args) {
        if (this.canvas.isConnected && this.canvas.closest('.stage') && args.length === 9 && args[0].width > 384) {
          const preview = window.__mapPreview, b = preview.landscapeModel(preview.snapshot.atlas).bounds;
          const sourceCell = args[0].width / (b.maxX - b.minX + 1);
          window.__mapDraw = { tileStep: args[7] / args[3] * sourceCell, smoothing: this.imageSmoothingEnabled };
        }
        return draw.apply(this, args);
      };
    });
    await page.setViewport({ width: 1920, height: 1080 });
    const url = process.env.WORLDMAP_PREVIEW_URL || 'http://127.0.0.1:8151/';
    await page.goto(url, { waitUntil: 'networkidle0', timeout: 120000 });
    await page.waitForFunction(() => window.__mapPreview?.ready, { timeout: 120000 });
    await settle(page);
    const save = kind => page.screenshot({ path: path.join(out, `worldmap-p1h-${label}-${kind}-1920x1080.png`) });
    await page.click('[title="Fit world"]'); await settle(page);
    await page.mouse.move(30, 30); await save('overview');
    const bounds = await page.$eval('canvas', e => { const r = e.getBoundingClientRect(); return { x: r.x, y: r.y, w: r.width, h: r.height }; });
    const wheel = async (delta, count) => {
      await page.mouse.move(bounds.x + bounds.w / 2, bounds.y + bounds.h / 2);
      for (let i = 0; i < count; i++) await page.mouse.wheel({ deltaY: delta });
      await settle(page);
    };
    const read = () => page.evaluate(() => ({ ...window.__mapDraw }));
    await wheel(-100, 30);
    const maximum = await read();
    const frame = async (area, kind) => {
      await page.click('[title="Fit world"]'); await settle(page); await wheel(-100, 30);
      const pan = await page.evaluate(({ step, area }) => {
        const a = window.__mapPreview.snapshot.atlas, b = window.__mapPreview.landscapeModel(a).bounds;
        const places = window.__mapPreview.surfaceGroups(a.places).filter(p => p.area === area);
        const xs = places.map(p => p.x), ys = places.map(p => p.y);
        return {
          x: ((b.minX + b.maxX) / 2 - (Math.min(...xs) + Math.max(...xs)) / 2) * step,
          y: ((b.minY + b.maxY) / 2 - (Math.min(...ys) + Math.max(...ys)) / 2) * step
        };
      }, { step: maximum.tileStep, area });
      const sx = bounds.x + bounds.w * .55, sy = bounds.y + bounds.h * .65;
      await page.mouse.move(sx, sy); await page.mouse.down();
      await page.mouse.move(sx + pan.x, sy + pan.y, { steps: 12 }); await page.mouse.up(); await settle(page);
      await page.mouse.move(30, 30); await save(kind);
    };
    await frame('Z01_sample_meadow', 'maxzoom');
    await frame('Z02_sample_town', 'sampletown');
    console.log(JSON.stringify({ label, errors, maximum, files: fs.readdirSync(out).filter(f => f.includes(`p1h-${label}`)) }, null, 2));
    if (errors.length) process.exitCode = 1;
  } finally {
    await browser.close();
  }
})().catch(err => { console.error(err); process.exit(1); });
