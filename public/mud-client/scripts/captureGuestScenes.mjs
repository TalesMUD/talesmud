import fs from 'node:fs';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { createRequire } from 'node:module';

const require = createRequire(process.env.UX_PLAYWRIGHT_PACKAGE || new URL('../package.json', import.meta.url));
const { chromium } = require('playwright');
const args = process.argv.slice(2);
const option = (name, fallback) => {
  const index = args.indexOf(name);
  return index === -1 ? fallback : args[index + 1];
};
const base = option('--base', '');
const out = path.resolve(option('--out', '/tmp/guest-scenes'));
const requested = option('--widths', '375,1024,1440').split(',').map(Number);
const sizes = [{ width: 375, height: 812 }, { width: 1024, height: 768 }, { width: 1440, height: 900 }, { width: 1920, height: 1080 }].filter(size => requested.includes(size.width));
if (!base || !sizes.length) throw new Error('Pass --base and valid --widths');
const manifest = { base, mode: 'live-guest-gameplay', startedAt: new Date().toISOString(), sizes: {} };
const browser = await chromium.launch({ headless: true, args: ['--no-sandbox', '--disable-dev-shm-usage'] });
fs.mkdirSync(out, { recursive: true });

async function command(page, message) {
  await page.evaluate(message => {
    if (window.__captureSocket?.readyState !== 1) throw new Error('Guest socket is closed');
    window.__captureSocket.send(JSON.stringify({ type: 'message', message }));
  }, message);
}

async function roomId(page) {
  return page.evaluate(() => window.__captureRoom || '');
}

async function walk(page, direction, expected) {
  await command(page, direction);
  await page.waitForFunction(expected => window.__captureRoom === expected, expected, { timeout: 12000 });
}

async function capture(page, width, scene, detail) {
  const dir = path.join(out, String(width));
  fs.mkdirSync(dir, { recursive: true });
  await page.screenshot({ path: path.join(dir, `${scene}.png`), animations: 'disabled' });
  manifest.sizes[width][scene] = { how: 'live', detail };
  console.log(`${width}/${scene} captured through guest gameplay`);
}

try {
  for (const size of sizes) {
    const context = await browser.newContext({ viewport: size, deviceScaleFactor: size.width === 375 ? 2 : 1, hasTouch: size.width === 375, isMobile: size.width === 375, reducedMotion: 'reduce' });
    const page = await context.newPage();
    const errors = [];
    const assets = [];
    const reads = [];
    manifest.sizes[size.width] = { assets };
    page.on('pageerror', error => errors.push(error.message));
    page.on('response', response => {
      const url = new URL(response.url());
      if (['bundle.js', 'extra.css', 'w2-screens.css', 'w3-screens.css'].includes(url.pathname.split('/').at(-1))) {
        reads.push(response.body().then(body => assets.push({ path: url.pathname, status: response.status(), sha256: createHash('sha256').update(body).digest('hex') })));
      }
    });
    await context.addInitScript(() => {
      const Original = window.WebSocket;
      window.WebSocket = class extends Original {
        constructor(...args) {
          super(...args);
          window.__captureSocket = this;
          this.addEventListener('message', event => {
            try {
              const frame = JSON.parse(event.data);
              if (frame.type === 'enterRoom' || frame.type === 'roomUpdate') window.__captureRoom = frame.room?.id || frame.room?.ID || '';
            } catch {}
          });
        }
      };
    });
    try {
      const response = await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 45000 });
      if (response.status() !== 200) throw new Error('Play page did not return HTTP 200');
      await page.locator('.btn-welcome.guest').click();
      for (let i = 0; i < 40; i += 1) {
        if (await page.locator('button.guest-play').count() || await roomId(page)) break;
        await page.waitForTimeout(200);
      }
      if (await page.locator('button.guest-play').count()) {
        const choice = page.locator('button.guest-choice').first();
        if (await choice.count()) await choice.click();
        const race = page.locator('.guest-race-grid button.guest-choice').first();
        if (await race.count()) await race.click();
        await page.locator('button.guest-play').click();
      }
      await page.waitForFunction(() => !!window.__captureRoom, null, { timeout: 30000 });
      if (await roomId(page) !== 'R0001') throw new Error('Guest did not start in tutorial room R0001');
      await walk(page, 'north', 'R0002');
      await walk(page, 'west', 'R0003');
      await command(page, 'speak to Spectral Warden');
      await page.waitForSelector('.dialog-overlay');
      await page.waitForTimeout(400);
      await capture(page, size.width, 'dialog', 'R0003, Spectral Warden');
      await page.locator('.dialog-dismiss').click();
      await page.waitForSelector('.dialog-overlay', { state: 'detached' });
      await walk(page, 'east', 'R0002');
      await walk(page, 'south', 'R0001');
      await walk(page, 'east', 'R0005');
      await page.waitForSelector('.entity-card.enemy', { timeout: 45000 });
      await command(page, 'attack catacomb rat');
      await page.waitForSelector('.battle-stage');
      await page.waitForTimeout(600);
      await capture(page, size.width, 'combat', 'R0005, catacomb rat');
      const deadline = Date.now() + 120000;
      let finished = false;
      while (Date.now() < deadline) {
        if (await page.locator('button.outcome-continue').isVisible().catch(() => false)) {
          await page.waitForTimeout(250);
          await capture(page, size.width, 'outcome', 'Result of the same live tutorial fight');
          finished = true;
          break;
        }
        await command(page, 'attack');
        await page.waitForTimeout(1200);
      }
      if (!finished) throw new Error('Tutorial fight did not finish within 120 seconds');
      if (errors.length) throw new Error(`Client raised ${errors.length} runtime errors`);
      await page.locator('button.outcome-continue').click();
    } catch (error) {
      manifest.sizes[size.width].failure = error.message;
      process.exitCode = 1;
      console.error(`${size.width} capture failed: ${error.message}`);
    } finally {
      await Promise.allSettled(reads);
      manifest.sizes[size.width].runtimeErrors = errors;
      await context.close();
      fs.writeFileSync(path.join(out, 'manifest.json'), `${JSON.stringify(manifest, null, 2)}\n`);
    }
  }
} finally {
  await browser.close();
  manifest.finishedAt = new Date().toISOString();
  fs.writeFileSync(path.join(out, 'manifest.json'), `${JSON.stringify(manifest, null, 2)}\n`);
}
