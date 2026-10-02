// Layout B default capture: fresh guest, no battle-layout localStorage, one ally.
const path = require('node:path');
const fs = require('node:fs');
const assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || '/home/clawd/clawd/node_modules/puppeteer-core');

(async () => {
  const base = process.env.BATTLEB_BASE_URL || 'http://127.0.0.1:8010';
  const publicDir = path.resolve(__dirname, '../public/mud-client/public');
  const out = path.resolve(__dirname, '../.director/ux-audit/after/b15-layoutb-1920x1080.png');
  const clientAssets = {
    '/play': 'index.html',
    '/play/': 'index.html',
    '/play/index.html': 'index.html',
    '/play/bundle.js': 'bundle.js',
    '/play/extra.css': 'extra.css',
    '/play/icons.css': 'icons.css',
    '/play/social-overlay.css': 'social-overlay.css',
    '/play/themes.css': 'themes.css',
    '/play/global.css': 'global.css',
  };
  const browser = await puppeteer.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser',
    headless: 'new',
    args: ['--no-sandbox', '--disable-dev-shm-usage'],
  });
  const page = await browser.newPage();
  const errors = [];
  try {
    await page.setViewport({ width: 1920, height: 1080 });
    page.on('pageerror', (error) => errors.push(error.message));
    await page.evaluateOnNewDocument(() => {
      const NativeSocket = WebSocket;
      window.WebSocket = class extends NativeSocket {
        constructor(...args) {
          super(...args);
          window.__bpSocket = this;
        }
      };
    });
    await page.setRequestInterception(true);
    page.on('request', (req) => {
      try {
        const url = new URL(req.url());
        const fileName = clientAssets[url.pathname];
        if (fileName && (url.hostname === '127.0.0.1' || url.hostname === 'localhost')) {
          const file = path.join(publicDir, fileName);
          const ext = path.extname(fileName);
          const type = ext === '.js' ? 'text/javascript' : (ext === '.css' ? 'text/css' : 'text/html; charset=utf-8');
          req.respond({ status: 200, contentType: type, body: fs.readFileSync(file) });
          return;
        }
      } catch (_) {}
      req.continue();
    });
    const response = await page.goto(base + '/play/?v=battleb1', { waitUntil: 'domcontentloaded' });
    assert.equal(response.status(), 200);
    await page.waitForSelector('.btn-welcome.guest', { timeout: 20000 });
    await page.click('.btn-welcome.guest');
    await page.waitForSelector('.switcher-button', { timeout: 25000 });
    await page.waitForFunction(() => !!window.__bpSocket, { timeout: 25000 });
    await page.evaluate(async () => {
      await document.fonts.ready;
      const send = (payload) => {
        window.__bpSocket.dispatchEvent(new MessageEvent('message', { data: JSON.stringify(payload) }));
      };
      send({
        type: 'combatStart',
        message: 'Layout B default preview',
        targetId: 'b15-rat',
        players: [
          {
            id: 'b15-self',
            type: 'player',
            name: 'Gimli',
            hp: 180,
            maxHp: 210,
            level: 8,
            classId: 'warrior',
            race: 'human',
            portrait: '/api/portraits/player-human-warrior.png',
            isAlive: true,
          },
          {
            id: 'b15-ally',
            type: 'player',
            name: 'Lyra',
            hp: 64,
            maxHp: 90,
            level: 6,
            classId: 'mage',
            race: 'elf',
            portrait: '/api/portraits/player-elf-mage.png',
            isAlive: true,
          },
        ],
        enemies: [
          {
            id: 'b15-rat',
            type: 'npc',
            name: 'Catacomb Rat',
            hp: 22,
            maxHp: 28,
            level: 2,
            threat: 'green',
            portrait: '/api/portraits/ENM0001.png',
            isAlive: true,
          },
        ],
      });
      send({
        type: 'combatTurn',
        actorId: 'b15-ally',
        actorName: 'Lyra',
        round: 2,
        message: "Round 2 — Lyra's turn.",
      });
    });
    await page.waitForSelector('.battle-stage.layout-b', { timeout: 10000 });
    await page.waitForSelector('.layout-b-frame.player-frame', { timeout: 5000 });
    await page.waitForSelector('.layout-b-frame.target-frame', { timeout: 5000 });
    await page.waitForSelector('.ally-card .ally-name', { timeout: 5000 });
    await page.waitForFunction(() => {
      const imgs = [...document.querySelectorAll('.battle-stage.layout-b img.player-sprite, .battle-stage.layout-b img.enemy-sprite, .battle-stage.layout-b .ally-portrait img')];
      return imgs.length >= 3 && imgs.every((img) => img.complete && img.naturalWidth > 32);
    }, { timeout: 8000 });
    const check = await page.evaluate(() => {
      const stage = document.querySelector('.battle-stage.layout-b');
      const playerFrame = document.querySelector('.layout-b-frame.player-frame');
      const targetFrame = document.querySelector('.layout-b-frame.target-frame');
      const foePlate = document.querySelector('.battle-stage.layout-b .foe-plate');
      const playerMeta = document.querySelector('.battle-stage.layout-b .player-panel .player-meta');
      const allyName = document.querySelector('.battle-stage.layout-b .ally-name');
      const allyCard = document.querySelector('.battle-stage.layout-b .ally-card');
      const player = document.querySelector('.battle-stage.layout-b .player-sprite');
      const enemy = document.querySelector('.battle-stage.layout-b .enemy-sprite');
      const turnChip = document.querySelector('.battle-stage.layout-b .turn-chip');
      const waitChip = document.querySelector('.battle-stage.layout-b .dock-status .queued-chip.wait');
      const waitingTimer = document.querySelector('.battle-stage.layout-b .decision-timer.waiting');
      const box = (el) => {
        if (!el) return null;
        const r = el.getBoundingClientRect();
        return { left: Math.round(r.left), top: Math.round(r.top), width: Math.round(r.width), height: Math.round(r.height) };
      };
      return {
        hasStage: !!stage,
        layoutKey: localStorage.getItem('talesmud_battle_layout_b'),
        playerFrameText: playerFrame?.textContent?.replace(/\s+/g, ' ').trim(),
        targetFrameText: targetFrame?.textContent?.replace(/\s+/g, ' ').trim(),
        allyText: allyCard?.textContent?.replace(/\s+/g, ' ').trim(),
        allyNameDisplay: allyName ? getComputedStyle(allyName).display : 'missing',
        foePlateDisplay: foePlate ? getComputedStyle(foePlate).display : 'missing',
        playerMetaDisplay: playerMeta ? getComputedStyle(playerMeta).display : 'missing',
        turnChipText: turnChip ? turnChip.textContent.replace(/\s+/g, ' ').trim() : '',
        waitDisplay: waitChip ? getComputedStyle(waitChip).display : 'none',
        waitingTimerDisplay: waitingTimer ? getComputedStyle(waitingTimer).display : 'none',
        playerSrc: player?.getAttribute('src') || '',
        enemySrc: enemy?.getAttribute('src') || '',
        allySrc: document.querySelector('.ally-portrait img')?.getAttribute('src') || '',
        allyBox: box(allyCard),
        playerBox: box(player),
        enemyBox: box(enemy),
        htmlHasBattleb1: !!document.querySelector('script[src*="battleb1"]'),
        height: document.scrollingElement.scrollHeight,
      };
    });
    assert.equal(check.hasStage, true);
    assert.equal(check.layoutKey, null);
    assert.match(check.playerFrameText, /Gimli/i);
    assert.match(check.targetFrameText, /Catacomb Rat/i);
    assert.match(check.allyText, /Lyra/i);
    assert.notEqual(check.allyNameDisplay, 'none');
    assert.equal(check.foePlateDisplay, 'none');
    assert.equal(check.playerMetaDisplay, 'none');
    assert.equal(check.turnChipText, '');
    assert.equal(check.waitDisplay, 'none');
    assert.equal(check.waitingTimerDisplay, 'none');
    assert.match(check.playerSrc, /player-human-warrior\.png/);
    assert.match(check.enemySrc, /ENM0001\.png/);
    assert.match(check.allySrc, /player-elf-mage\.png/);
    assert.ok(check.allyBox.left < check.playerBox.left, 'ally sits left of the player');
    assert.ok(check.playerBox.left < check.enemyBox.left, 'player sits left of the enemy');
    assert.ok(check.allyBox.width >= 80 && check.allyBox.height >= 40, 'ally strip is readable');
    assert.equal(check.htmlHasBattleb1, true);
    assert.equal(check.height, 1080);
    assert.deepEqual(errors, []);
    fs.mkdirSync(path.dirname(out), { recursive: true });
    await page.screenshot({ path: out });
    fs.writeFileSync(
      path.resolve(__dirname, '../.director/ux-audit/after/b15-smoke.json'),
      JSON.stringify({ out, check, errors }, null, 2)
    );
    console.log(JSON.stringify({ out, check, errors }));
  } finally {
    await browser.close();
  }
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
