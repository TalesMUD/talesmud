const path = require('node:path');
const fs = require('node:fs');
const assert = require('node:assert/strict');
const puppeteer = require('/home/clawd/clawd/node_modules/puppeteer-core');
(async () => {
  const out = path.resolve(__dirname, '../.director/ux-audit/after');
  const base = process.env.FOG1_PREVIEW_URL || 'http://127.0.0.1:8155/';
  const browser = await puppeteer.launch({
    executablePath: '/usr/bin/google-chrome-stable',
    headless: 'new',
    args: ['--no-sandbox', '--disable-dev-shm-usage'],
  });
  try {
    const page = await browser.newPage();
    const errors = [];
    page.on('pageerror', e => errors.push(e.message));
    await page.setViewport({ width: 1920, height: 1080 });
    await page.goto(base, { waitUntil: 'networkidle0', timeout: 90000 });
    await page.waitForFunction(() => window.__mapPreview && window.__mapPreview.ready, { timeout: 90000 });
    await page.evaluate(async () => {
      await document.fonts.ready;
      await window.__mapPreview.waitForMapScenes();
      for (let i = 0; i < 4; i++) await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)));
    });
    const fit = await page.$('[title="Fit world"]');
    if (fit) {
      await fit.click();
      await page.evaluate(() => window.__mapPreview.waitForMapScenes());
      await new Promise(r => setTimeout(r, 1800));
    }
    await page.screenshot({ path: path.join(out, 'fog1-overview-1920x1080.png') });
    const canvas = await page.$('canvas');
    const box = await canvas.boundingBox();
    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
    for (let i = 0; i < 26; i++) await page.mouse.wheel({ deltaY: -100 });
    await page.evaluate(() => window.__mapPreview.waitForMapScenes());
    await new Promise(r => setTimeout(r, 2000));
    await page.screenshot({ path: path.join(out, 'fog1-maxzoom-1920x1080.png') });
    const fogCells = await page.evaluate(() => (window.__mapPreview.snapshot.atlas.landscape || []).filter(c => c.terrain === 'fog').length);
    fs.writeFileSync(path.join(out, 'fog1-smoke.json'), JSON.stringify({
      errors, fogCells,
      screenshots: ['fog1-overview-1920x1080.png', 'fog1-maxzoom-1920x1080.png'],
      note: 'Local preview atlas with expanded fog cells to demonstrate soft volumetric overlay',
    }, null, 2));
    assert.equal(errors.length, 0, errors.join('; '));
    console.log(JSON.stringify({ errors, fogCells }, null, 2));
  } finally { await browser.close(); }
})().catch(e => { console.error(e); process.exitCode = 1; });
