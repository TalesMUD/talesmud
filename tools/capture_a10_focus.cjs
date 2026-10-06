// Controlled two-hostile BattleStage payload on a fresh live guest.
const path = require('node:path');
const fs = require('node:fs');
const assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || '/home/clawd/clawd/node_modules/puppeteer-core');

(async () => {
  const base = process.env.A10_BASE_URL || 'https://veilspan.com';
  const out = path.resolve(__dirname, '../.director/ux-audit/after/a10-focus-1920x1080.png');
  const browser = await puppeteer.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser', headless: 'new', args: ['--no-sandbox', '--disable-dev-shm-usage'] });
  const page = await browser.newPage();
  const errors = [];
  try {
    await page.setViewport({ width: 1920, height: 1080 });
    page.on('pageerror', error => errors.push(error.message));
    await page.evaluateOnNewDocument(() => {
      const NativeSocket = WebSocket;
      window.WebSocket = class extends NativeSocket {
        constructor(...args) {
          super(...args);
          window.__a10Socket = this;
        }
        send(data) {
          try {
            const packet = JSON.parse(data);
            if (String(packet.message || '').startsWith('focus ')) return;
          } catch (_) { /* forward other payloads */ }
          return super.send(data);
        }
      };
    });
    const response = await page.goto(base + '/play/', { waitUntil: 'domcontentloaded' });
    assert.equal(response.status(), 200);
    await page.waitForSelector('.btn-welcome.guest', { timeout: 20000 });
    await page.click('.btn-welcome.guest');
    await page.waitForSelector('.switcher-button', { timeout: 25000 });
    await page.waitForFunction(() => !!window.__a10Socket, { timeout: 25000 });
    await page.evaluate(async () => {
      await document.fonts.ready;
      window.__a10Socket.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({
        type: 'combatStart', message: 'Controlled two-hostile focus preview', targetId: 'a10-rat',
        players: [{ id: 'a10-self', type: 'player', name: 'Guest', hp: 70, maxHp: 70, level: 4, isAlive: true }],
        enemies: [
          { id: 'a10-rat', type: 'npc', name: 'Crypt Rat', hp: 24, maxHp: 24, level: 4, threat: 'yellow', isAlive: true },
          { id: 'a10-wraith', type: 'npc', name: 'Skull Wraith', hp: 90, maxHp: 90, level: 9, threat: 'skull', isAlive: true },
        ],
      }) }));
    });
    await page.waitForSelector('.enemy-card:nth-of-type(2)', { timeout: 10000 });
    await page.click('.enemy-card:nth-of-type(2)');
    await page.waitForSelector('.focus-threat-warning', { timeout: 3000 });
    const check = await page.evaluate(() => ({
      focused: document.querySelector('.enemy-card[aria-pressed="true"]')?.textContent,
      warning: document.querySelector('.focus-threat-warning')?.textContent,
      height: document.scrollingElement.scrollHeight,
    }));
    assert.match(check.focused, /Skull Wraith.*FOCUS/);
    assert.match(check.warning, /Skull Wraith is much stronger than you/);
    assert.equal(check.height, 1080);
    assert.deepEqual(errors, []);
    fs.mkdirSync(path.dirname(out), { recursive: true });
    await page.screenshot({ path: out });
    console.log(JSON.stringify({ out, check, errors }));
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
