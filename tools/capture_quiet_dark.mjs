import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
const require = createRequire(process.env.PLAYWRIGHT_PACKAGE || '/home/clawd/apps/cozytown-arcade/package.json');
const { chromium } = require('playwright');
const phase = process.env.QUIET_PHASE || 'before';
const url = process.env.QUIET_URL || 'https://veilspan.com/play/?v=phone1';
const root = process.env.QUIET_SHOTS || '/home/clawd/dev/briefs/quiet-dark-shots';
const sizes = [[1024, 768], [375, 812], [1440, 900]].filter(([w]) => !process.env.QUIET_WIDTH || Number(process.env.QUIET_WIDTH) === w);
const browser = await chromium.launch({ headless: true, args: ['--no-sandbox', '--disable-dev-shm-usage'] });
const results = [];
let activePage;
try {
  for (const [width, height] of sizes) {
    const context = await browser.newContext({ viewport: { width, height }, deviceScaleFactor: 1, hasTouch: width <= 1024, reducedMotion: 'reduce' });
    const page = await context.newPage();
    activePage = page;
    page.setDefaultTimeout(20000);
    const dir = path.join(root, phase, 'web', `${width}x${height}`);
    fs.mkdirSync(dir, { recursive: true });
    const result = { width, height, errors: [] };
    page.on('pageerror', error => { result.errors.push(error.message); console.error(error.message); });
    await context.addInitScript(() => {
      const Native = window.WebSocket;
      window.WebSocket = class extends Native {
        constructor(...args) {
          super(...args);
          window.__quietSocket = this;
          window.__quietFrames = [];
          this.addEventListener('message', event => {
            try { const frame = JSON.parse(event.data); window.__quietFrames.push(frame); } catch {}
          });
        }
      };
    });
    if (process.env.QUIET_LOCAL) {
      await context.addInitScript(() => localStorage.setItem('talesmud_settings_v1', JSON.stringify({version:1, interface:{theme:'quiet-dark',themeExplicit:true}})));
      await page.route('**/play/**', async route => {
        const name = new URL(route.request().url()).pathname.replace(/^\/play\//, '') || 'index.html';
        const file = path.resolve('public/mud-client/public', name);
        if (file.startsWith(path.resolve('public/mud-client/public') + path.sep) && fs.existsSync(file) && fs.statSync(file).isFile()) await route.fulfill({ path: file });
        else await route.continue();
      });
    }
    let replayCommand;
    if (process.env.QUIET_REPLAY) {
      await context.addInitScript(() => sessionStorage.setItem('talesmud_guest_token', 'isolated-visual-replay'));
      const frames = JSON.parse(fs.readFileSync(process.env.QUIET_REPLAY, 'utf8'));
      result.source = 'replay of captured live guest frames, isolated from live combat';
      await page.route('**/api/guest', route => route.fulfill({ json: { token: 'isolated-visual-replay', expiresIn: 1800 } }));
      await page.route('**/api/server-info', route => route.fulfill({json:{serverName:'Screenshot replay',defaultTheme:'quiet-dark'}}));
      await page.routeWebSocket(() => true, socket => {
        const initial = frames.slice(0, frames.findIndex(f => f.type === 'enterRoom' && f.room.id === 'R0005'));
        const deliver = async frame => {
          await page.evaluate(frame => { window.__quietFrames ||= []; window.__quietFrames.push(frame); }, frame);
          socket.send(JSON.stringify(frame));
        };
        setTimeout(async () => { for (const frame of initial) await deliver(frame); }, 100);
        replayCommand = async command => {
          if (command === 'east') await deliver(frames.find(f => f.type === 'enterRoom' && f.room.id === 'R0005'));
          if (command.startsWith('attack')) {
            const combat = structuredClone(frames.find(f => f.type === 'combatStart'));
            const self = frames.find(f => f.type === 'characterSelected').character;
            combat.players = combat.players.filter(p => p.id === self.id);
            await deliver(combat);
          }
          if (command === 'flee') await deliver({ type: 'combatLeave' });
        };
        socket.onMessage(raw => replayCommand(JSON.parse(raw).message || ''));
      });
    }
    const response = await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 45000 });
    result.status = response.status();
    if (!process.env.QUIET_REPLAY) await page.locator('.btn-welcome.guest').click();
    await page.waitForFunction(() => window.__quietFrames?.some(f => f.type === 'enterRoom') || document.querySelector('.guest-play'));
    if (await page.locator('.guest-play').count()) {
      await page.locator('.guest-choice').first().click();
      if (await page.locator('.guest-race-grid .guest-choice').count()) await page.locator('.guest-race-grid .guest-choice').first().click();
      await page.locator('.guest-play').click();
    }
    await page.waitForFunction(() => window.__quietFrames?.some(f => f.type === 'enterRoom'));
    await page.waitForTimeout(900);
    if (await page.locator('.dialog-close').count()) await page.locator('.dialog-close').first().evaluate(el => el.click());
    await page.evaluate(() => document.fonts.ready);
    const shot = async name => {
      await page.screenshot({ path: path.join(dir, `${name}.png`), animations: 'disabled' });
      if (phase === 'after') {
        result.geometry ||= {};
        result.geometry[name] = await page.evaluate(() => [...document.querySelectorAll('button, [role=button], .hotbar-widget .slot')].filter(el => {
          const r = el.getBoundingClientRect();
          return r.width && r.height && r.bottom > 0 && r.top < innerHeight && getComputedStyle(el).visibility !== 'hidden';
        }).map(el => {
          const r = el.getBoundingClientRect();
          return { text: (el.getAttribute('aria-label') || el.textContent).trim().slice(0, 50), width: r.width, height: r.height };
        }));
      }
    };
    if (process.env.QUIET_LOCAL) await page.evaluate(() => { document.body.dataset.theme = 'quiet-dark'; });
    result.theme = await page.locator('body').getAttribute('data-theme');
    await shot('main');
    await shot('hotbar');
    if (await page.locator('[data-sheet="character"]').count()) {
      for (const id of ['character', 'equipment', 'quests']) {
        await page.locator(`[data-sheet="${id}"]`).evaluate(el => el.click());
        await page.locator('.bottom-sheet').waitFor();
        await shot(id);
        if (id === 'character') await shot('bottom-sheet');
        await page.locator('.sheet-close').evaluate(el => el.click());
      }
      await page.locator('.mobile-action-bar button').filter({ hasText: /Inv/ }).first().evaluate(el => el.click());
      await shot('inventory');
      if (await page.locator('.inv-close').count()) await page.locator('.inv-close').evaluate(el => el.click());
      await page.keyboard.press('Escape');
    } else {
      await shot('panels');
    }
    result.room = await page.evaluate(() => window.__quietFrames.find(f => f.type === 'enterRoom')?.room?.id);
    const send = async command => process.env.QUIET_REPLAY ? replayCommand(command) : page.evaluate(command => window.__quietSocket.send(JSON.stringify({ type: 'message', message: command })), command);
    const here = async () => page.evaluate(() => [...window.__quietFrames].reverse().find(f => f.type === 'enterRoom' || f.type === 'roomUpdate')?.room?.id);
    const route = { R0001: 'east', R0002: 'south', R0003: 'east', R0004: 'south', R0006: 'south' };
    for (let moves = 0; await here() !== 'R0005' && moves < 8; moves++) {
      const before = await here();
      if (!route[before]) throw new Error(`No tutorial route from ${before}`);
      await send(route[before]);
      await page.waitForFunction(before => [...window.__quietFrames].reverse().find(f => f.type === 'enterRoom')?.room?.id !== before, before);
      await page.waitForTimeout(250);
    }
    await page.waitForFunction(() => [...document.querySelectorAll('.entity-card.enemy, .battle-stage')].some(el => el.getBoundingClientRect().width > 0), null, { timeout: 90000 });
    await shot('enemies');
    if (!await page.locator('.battle-stage').count()) await send('attack catacomb rat');
    await page.locator('.battle-stage').waitFor();
    await page.waitForTimeout(300);
    await shot('combat');
    fs.writeFileSync(path.join(dir, 'guest-frames.json'), JSON.stringify(await page.evaluate(() => window.__quietFrames.filter(f => ['enterRoom', 'characterSelected', 'characterUpdate', 'inventoryUpdate', 'questLog', 'combatStart'].includes(f.type))), null, 2));
    result.combatSource = process.env.QUIET_REPLAY ? 'captured live combat payload replay' : 'live catacomb rat';
    result.finalTheme = await page.locator('body').getAttribute('data-theme');
    result.styles = await page.evaluate(() => Object.fromEntries(['.game-panel', '.enemy-card', '.enemy-sprite-wrap', '.bottom-sheet', '.hotbar-widget .slot', '.combat-hotbar .hb-slot', '.enemy-strip', '.quiet-enemy-label', '.enemy-sprite'].map(selector => {
      const el = document.querySelector(selector);
      if (!el) return [selector, null];
      const s = getComputedStyle(el);
      const r = el.getBoundingClientRect(); return [selector, { background: s.backgroundColor, border: s.border, shadow: s.boxShadow, display: s.display, x:r.x,y:r.y,width:r.width,height:r.height,opacity:s.opacity,transform:s.transform }];
    })));
    result.controls = await page.evaluate(() => {
      const selectors = '.mobile-header button, .mobile-tab-bar button, .mobile-action-bar button, .hotbar .slot';
      return [...document.querySelectorAll(selectors)].filter(el => el.getBoundingClientRect().width && el.getBoundingClientRect().height).map(el => {
        const r = el.getBoundingClientRect(), s = getComputedStyle(el);
        return { text: (el.getAttribute('aria-label') || el.textContent).trim().slice(0, 45), width: r.width, height: r.height, font: s.fontSize, color: s.color, background: s.backgroundColor };
      });
    });
    fs.writeFileSync(path.join(dir, 'metrics.json'), JSON.stringify(result, null, 2));
    results.push(result);
    await send('flee');
    await page.unrouteAll({ behavior: 'ignoreErrors' });
    await context.close();
    console.log(`${phase} ${width}x${height} ${result.theme}`);
  }
} catch (error) {
  if (activePage && !activePage.isClosed()) { await activePage.screenshot({path: path.join(root, 'capture-error.png')}); console.error(await activePage.evaluate(() => ({title:document.title, text:document.body.innerText.slice(0,500), frames:window.__quietFrames?.map(f=>f.type)}))); }
  throw error;
} finally {
  await browser.close();
  fs.writeFileSync(path.join(root, `${phase}-web.json`), JSON.stringify(results, null, 2));
}
