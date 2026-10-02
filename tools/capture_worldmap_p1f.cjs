// Capture and verify the production overlay locally; see WORLDMAP-PREVIEW.md.
// PUPPETEER_MODULE=/path/to/puppeteer-core node tools/capture_worldmap_p1f.cjs
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || 'puppeteer-core');
const out = path.resolve(__dirname, '../.director/ux-audit/after');
const settle = page => page.evaluate(async () => {
  await document.fonts.ready;
  await document.fonts.load('24px "Material Icons"');
  await Promise.all([...document.images].map(i => i.decode().catch(() => {})));
  for (let i = 0; i < 3; i++) {
    await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)));
    await window.__mapPreview.waitForMapScenes();
  }
});
(async () => {
  const browser = await puppeteer.launch({executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser', headless: 'new', args: ['--no-sandbox', '--disable-dev-shm-usage']});
  try {
    const page = await browser.newPage(), errors = [], sheets = [], workers = [];
    page.on('pageerror', e => errors.push(e.message));
    page.on('workercreated', w => workers.push({url: w.url()}));
    page.on('response', r => { if (r.url().includes('/api/map-tiles/')) sheets.push({url: r.url(), status: r.status()}); });
    // Observe actual overlay raster draws without changing the production component.
    await page.evaluateOnNewDocument(() => {
      const draw = CanvasRenderingContext2D.prototype.drawImage;
      CanvasRenderingContext2D.prototype.drawImage = function (...args) {
        if (this.canvas.isConnected && this.canvas.closest('.stage') && args.length === 9 && args[0].width > 192) {
          window.__mapDraw = {tileStep: args[7] / args[3] * 32, smoothing: this.imageSmoothingEnabled};
        }
        return draw.apply(this, args);
      };
    });
    await page.setViewport({width: 1920, height: 1080});
    await page.goto(process.env.WORLDMAP_PREVIEW_URL || 'http://127.0.0.1:8142/', {waitUntil: 'networkidle0'});
    await page.waitForFunction(() => window.__mapPreview?.ready);
    await settle(page);
    const read = () => page.evaluate(() => ({...window.__mapDraw}));
    const save = kind => page.screenshot({path: path.join(out, `worldmap-p1f-${kind}-1920x1080.png`)});
    await page.click('[title="Fit world"]'); await settle(page);
    const overview = await read();
    await page.mouse.move(30, 30); await save('overview');
    const bounds = await page.$eval('canvas', e => {const r = e.getBoundingClientRect(); return {x: r.x, y: r.y, w: r.width, h: r.height};});
    const wheel = async (delta, count) => {
      await page.mouse.move(bounds.x + bounds.w / 2, bounds.y + bounds.h / 2);
      for (let i = 0; i < count; i++) await page.mouse.wheel({deltaY: delta});
      await settle(page);
    };
    await wheel(-100, 30);
    const maximum = await read(), oldMaximum = Math.min(110, overview.tileStep * 5);
    assert.ok(Math.abs(maximum.tileStep - Math.min(220, overview.tileStep * 10)) < .01, `wheel reaches the new scale/tile clamp: ${JSON.stringify({overview, maximum})}`);
    assert.ok(Math.abs(maximum.tileStep / oldMaximum - 2) < .01, 'closest view is twice the previous max');
    assert.equal(maximum.smoothing, false, 'enlarged landscape uses nearest-neighbor sampling');
    await wheel(-100, 5); assert.deepEqual(await read(), maximum, 'wheel cannot exceed maximum');
    const pan = await page.evaluate(step => {
      const a = window.__mapPreview.snapshot.atlas, b = window.__mapPreview.landscapeModel(a).bounds;
      const town = window.__mapPreview.surfaceGroups(a.places).filter(p => p.area === 'Z02_sample_town');
      const xs = town.map(p => p.x), ys = town.map(p => p.y);
      return {x: ((b.minX + b.maxX) / 2 - (Math.min(...xs) + Math.max(...xs)) / 2) * step,
        y: ((b.minY + b.maxY) / 2 - (Math.min(...ys) + Math.max(...ys)) / 2) * step};
    }, maximum.tileStep);
    const sx = bounds.x + bounds.w * .55, sy = bounds.y + bounds.h * .65;
    await page.mouse.move(sx, sy); await page.mouse.down();
    await page.mouse.move(sx + pan.x, sy + pan.y, {steps: 12}); await page.mouse.up(); await settle(page);
    await page.mouse.move(30, 30); await save('maxzoom');
    await page.click('[title="Recenter on you"]'); await settle(page);
    await wheel(-100, 30);
    const recenterMaximum = await read();
    assert.ok(Math.abs(recenterMaximum.tileStep - 220) < .01, 'local recenter respects the enlarged tile cap');
    await page.click('[title="Fit world"]'); await settle(page);
    assert.deepEqual(await read(), overview, 'Fit world restores the original continent frame');
    await wheel(100, 30); const minimum = await read();
    assert.deepEqual(minimum, overview, 'minimum zoom retains the Fit world frame');
    // Exercise the same pointer handlers used by touch pinch; over/undershoot both bounds.
    const pinch = async (start, end) => {
      await page.evaluate(({start, end}) => {
        const canvas = document.querySelector('canvas'), r = canvas.getBoundingClientRect(), x = r.x + r.width / 2, y = r.y + r.height / 2;
        const send = (type, id, dx) => canvas.dispatchEvent(new PointerEvent(type, {bubbles: true, pointerType: 'touch', pointerId: id, clientX: x + dx, clientY: y}));
        send('pointerdown', 101, -start / 2); send('pointerdown', 102, start / 2);
        send('pointermove', 101, -end / 2); send('pointermove', 102, end / 2);
        send('pointerup', 101, -end / 2); send('pointerup', 102, end / 2);
      }, {start, end});
      await settle(page); return read();
    };
    const pinchMaximum = await pinch(20, 400); assert.deepEqual(pinchMaximum, maximum, 'pinch shares wheel max clamp');
    const pinchMinimum = await pinch(400, 10); assert.deepEqual(pinchMinimum, overview, 'pinch shares wheel min clamp');
    for (const worker of workers) worker.status = await page.evaluate(async url => (await fetch(url, {cache: 'force-cache'})).status, worker.url);
    assert.equal(errors.length, 0, 'no page errors');
    assert.ok(workers.length && workers.every(w => w.status === 200), 'worker loads');
    assert.ok(sheets.length && sheets.every(s => s.status === 200), 'terrain art loads');
    const report = {source: 'Local read-only production-overlay preview; fonts, art and worker scenes settled before screenshots.', viewport: {width: 1920, height: 1080}, overview, maximum, oldMaximum, maxZoomRatio: maximum.tileStep / oldMaximum, recenterMaximum, minimum, pinchMaximum, pinchMinimum, workers, sheets, errors, screenshots: ['overview', 'maxzoom'].map(n => `worldmap-p1f-${n}-1920x1080.png`)};
    fs.writeFileSync(path.join(out, 'worldmap-p1f-browser-checks.json'), JSON.stringify(report, null, 2) + '\n');
    console.log(JSON.stringify(report, null, 2));
  } finally { await browser.close(); }
})().catch(e => {console.error(e); process.exitCode = 1;});
