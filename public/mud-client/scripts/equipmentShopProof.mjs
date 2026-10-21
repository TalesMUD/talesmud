import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const require = createRequire(process.env.UX_PLAYWRIGHT_PACKAGE || new URL('../package.json', import.meta.url));
const { chromium } = require('playwright');
const client = fileURLToPath(new URL('../', import.meta.url));
const out = process.env.UX_OUT || '/tmp/equipment-shop-proof';
const phase = process.env.UX_PHASE || 'after';
const port = Number(process.env.UX_PORT || 5117);
const base = `http://127.0.0.1:${port}/play/?v=wowf2`;
const measurements = [];
const failures = [];
const browser = await chromium.launch({ headless: true, args: ['--no-sandbox', '--disable-dev-shm-usage'] });

function check(theme, width, name, ok, boxes) {
  const row = { theme, width, name, ok, boxes };
  measurements.push(row);
  if (!ok) failures.push(row);
}

async function screenshot(page, theme, width, name) {
  const dir = path.join(out, phase, theme, String(width));
  fs.mkdirSync(dir, { recursive: true });
  await page.screenshot({ path: path.join(dir, `${name}.png`), animations: 'disabled' });
}

try {
  for (const theme of ['dark-fantasy', 'gilded']) {
    const server = spawn(process.execPath, [path.join(client, 'scripts/uxTopLocalServer.mjs')], {
      env: { ...process.env, UX_FIXTURE: '1', UX_PORT: String(port), UX_DEFAULT_THEME: theme, UX_PUBLIC: path.join(client, 'public') },
      stdio: 'ignore',
    });
    try {
      let ready = false;
      for (let i = 0; i < 100; i += 1) {
        try { ready = (await fetch(base)).ok; } catch {}
        if (ready) break;
        await new Promise(resolve => setTimeout(resolve, 100));
      }
      if (!ready) throw new Error('Fixture server did not start');
      for (const width of [375, 1024, 1440, 1920]) {
        const context = await browser.newContext({ viewport: { width, height: width === 375 ? 812 : width === 1920 ? 1080 : 900 }, reducedMotion: 'reduce' });
        const page = await context.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await context.addInitScript(() => {
          const Original = window.WebSocket;
          window.WebSocket = class extends Original {
            constructor(...args) { super(...args); window.__proofSocket = this; }
          };
        });
        try {
          await page.goto(base);
          await page.locator('.btn-welcome.guest').click();
          await page.waitForSelector('.hotbar-widget');
          await page.waitForTimeout(400);
          const equipmentTab = page.locator('.tab-bar button.tab').filter({ hasText: /Equipment/i });
          if (await equipmentTab.count()) {
            await equipmentTab.click();
            await page.waitForSelector('.equipment-widget');
            await screenshot(page, theme, width, 'equipment');
            const boxes = await page.evaluate(() => {
              const rect = selector => document.querySelector(selector).getBoundingClientRect().toJSON();
              return { doll: rect('.equipment-widget .doll'), portrait: rect('.equipment-widget .portrait-frame'), weapons: rect('.equipment-widget .weapon-row'), stats: rect('.equipment-widget .compact-stats') };
            });
            check(theme, width, 'stats-follow-entire-doll', boxes.stats.top >= Math.max(boxes.portrait.bottom, boxes.weapons.bottom) - 1, boxes);
            await page.locator('.compact-stats').scrollIntoViewIfNeeded();
            await screenshot(page, theme, width, 'equipment-stats');
            check(theme, width, 'stats-scroll-into-view', await page.locator('.compact-stats').evaluate(el => {
              const r = el.getBoundingClientRect();
              const p = el.parentElement.getBoundingClientRect();
              return r.top >= p.top - 1 && r.bottom <= p.bottom + 1;
            }), {});
          }
          await page.evaluate(() => window.__proofSocket.send(JSON.stringify({ type: 'message', message: 'list' })));
          await page.waitForSelector('.shop-overlay');
          const rows = await page.locator('.shop-row').evaluateAll(elements => elements.map(el => {
            const rect = selector => el.querySelector(selector).getBoundingClientRect().toJSON();
            return { icon: rect('.shop-icon'), name: rect('.shop-name'), price: rect('.shop-price') };
          }));
          check(theme, width, 'shop-price-below-name-and-beside-icon', rows.length > 0 && rows.every(row => row.price.top >= row.name.bottom - 1 && row.price.left >= row.icon.right - 1), rows);
          await screenshot(page, theme, width, 'shop');
          check(theme, width, 'runtime-errors', errors.length === 0, errors);
        } finally { await context.close(); }
      }
    } finally {
      const exited = new Promise(resolve => server.once('exit', resolve));
      server.kill();
      await exited;
    }
  }
} finally { await browser.close(); }

fs.mkdirSync(out, { recursive: true });
fs.writeFileSync(path.join(out, `${phase}-proof.json`), `${JSON.stringify({ mode: 'local-protocol-fixture', measurements, failures }, null, 2)}\n`);
console.log(`${measurements.length} layout checks, ${failures.length} failures`);
if (failures.length && phase !== 'before') process.exitCode = 1;
