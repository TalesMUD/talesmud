// Soft fog Cartographer capture on local :8010.
const path = require('node:path');
const fs = require('node:fs');
const assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || '/home/clawd/clawd/node_modules/puppeteer-core');

(async () => {
  const base = process.env.FOG1_BASE_URL || 'http://127.0.0.1:8010';
  const outDir = path.resolve(__dirname, '../.director/ux-audit/after');
  fs.mkdirSync(outDir, { recursive: true });
  const overviewPath = path.join(outDir, 'fog1-overview-1920x1080.png');
  const zoomPath = path.join(outDir, 'fog1-maxzoom-1920x1080.png');
  const browser = await puppeteer.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser',
    headless: 'new',
    args: ['--no-sandbox', '--disable-dev-shm-usage'],
  });
  const page = await browser.newPage();
  const errors = [];
  try {
    await page.setViewport({ width: 1920, height: 1080 });
    page.on('pageerror', (e) => errors.push(e.message));
    const response = await page.goto(base + '/play/?v=fog1', { waitUntil: 'domcontentloaded', timeout: 60000 });
    assert.equal(response.status(), 200);
    await page.waitForSelector('.btn-welcome.guest', { timeout: 30000 });
    await page.click('.btn-welcome.guest');
    await page.waitForSelector('.switcher-button, .room-widget, .minimap-widget', { timeout: 45000 });
    // Wait for character + atlas
    await page.waitForFunction(() => {
      const s = window.__muxStore || window.muxStore;
      return !!(s && (s.atlas || s.get?.()?.atlas || s.subscribe));
    }, { timeout: 45000 }).catch(() => {});
    // Open cartographer via store API
    const opened = await page.evaluate(() => {
      const candidates = [
        window.__muxStore,
        window.muxStore,
        window.store,
      ].filter(Boolean);
      // Svelte store might be on Game
      const fromDoc = document.querySelector('[data-store]');
      for (const s of candidates) {
        if (s.openMapOverview) { s.openMapOverview(); return 'openMapOverview'; }
        if (s.setMapOverviewOpen) { s.setMapOverviewOpen(true); return 'setMapOverviewOpen'; }
      }
      // Try clicking minimap expand / Map button
      const btn = [...document.querySelectorAll('button, [role="button"], .chip, a')].find(el =>
        /cartograph|world map|^map$/i.test((el.textContent || '').trim()) ||
        /map|explore|fullscreen/i.test(el.getAttribute('title') || '') ||
        el.className?.toString?.().includes('map-expand')
      );
      if (btn) { btn.click(); return 'click:' + (btn.title || btn.textContent || '').slice(0, 40); }
      return null;
    });
    // Fallback: press M or look for map in hotbar
    if (!opened) {
      await page.keyboard.press('m');
      await new Promise(r => setTimeout(r, 800));
    }
    await page.waitForSelector('#map-overview-overlay, .map-overview, .cartographer', { timeout: 20000 });
    await page.waitForFunction(() => {
      const ov = document.getElementById('map-overview-overlay');
      if (!ov) return false;
      const canvas = ov.querySelector('canvas');
      return !!(canvas && canvas.width > 100);
    }, { timeout: 45000 });
    // Let fog bake
    await new Promise(r => setTimeout(r, 2500));
    await page.evaluate(async () => {
      await document.fonts.ready;
      await Promise.all([...document.images].map(i => i.decode().catch(() => {})));
    });
    // Fit world
    const fit = await page.$('[title="Fit world"]');
    if (fit) { await fit.click(); await new Promise(r => setTimeout(r, 1500)); }
    await page.screenshot({ path: overviewPath, fullPage: false });

    // Zoom in toward explore edge
    const canvas = await page.$('#map-overview-overlay canvas, .map-overview canvas, canvas');
    const box = await canvas.boundingBox();
    if (box) {
      await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      for (let i = 0; i < 24; i++) await page.mouse.wheel({ deltaY: -120 });
      await new Promise(r => setTimeout(r, 2000));
    }
    await page.screenshot({ path: zoomPath, fullPage: false });

    // Sample paintAtlas soft fog via canvas read if available
    const fogCheck = await page.evaluate(() => {
      const ov = document.getElementById('map-overview-overlay');
      const canvas = ov?.querySelector('canvas');
      if (!canvas) return { ok: false, reason: 'no canvas' };
      const ctx = canvas.getContext('2d');
      // Heuristic: scan a strip for non-binary alpha/edge softness by looking at nearby dark pixels variance
      const { width: w, height: h } = canvas;
      const data = ctx.getImageData(0, 0, w, h).data;
      let dark = 0, mid = 0, bright = 0;
      for (let i = 0; i < data.length; i += 16) {
        const a = data[i + 3], lum = (data[i] + data[i + 1] + data[i + 2]) / 3;
        if (a < 10) continue;
        if (lum < 40) dark++;
        else if (lum < 90) mid++;
        else bright++;
      }
      return { ok: true, dark, mid, bright, w, h, v: document.querySelector('script[src*="fog1"]') ? 'fog1' : 'other' };
    });

    const smoke = {
      opened,
      fogCheck,
      errors,
      screenshots: [overviewPath, zoomPath],
      playV: await page.evaluate(() => [...document.querySelectorAll('script[src],link[href]')].map(e => e.src || e.href).filter(u => /fog1|bundle/.test(u))),
    };
    fs.writeFileSync(path.join(outDir, 'fog1-smoke.json'), JSON.stringify(smoke, null, 2));
    assert.equal(errors.length, 0, 'page errors: ' + errors.join('; '));
    console.log(JSON.stringify(smoke, null, 2));
  } finally {
    await browser.close();
  }
})().catch((e) => { console.error(e); process.exitCode = 1; });
