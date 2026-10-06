// Controlled layout-B BattleStage capture on local :8010.
const path = require('node:path');
const fs = require('node:fs');
const assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || '/home/clawd/clawd/node_modules/puppeteer-core');

(async () => {
  const base = process.env.BATTLEPOC_BASE_URL || 'http://127.0.0.1:8010';
  const out = path.resolve(__dirname, '../.director/ux-audit/after/battlepoc3-layoutb-1920x1080.png');
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
      try {
        localStorage.setItem('talesmud_battle_layout_b', '1');
      } catch (_) {}
      const NativeSocket = WebSocket;
      window.WebSocket = class extends NativeSocket {
        constructor(...args) {
          super(...args);
          window.__bpSocket = this;
        }
      };
    });
    const response = await page.goto(base + '/play/?battleLayout=b&v=battlepoc3', {
      waitUntil: 'domcontentloaded',
    });
    assert.equal(response.status(), 200);
    await page.waitForSelector('.btn-welcome.guest', { timeout: 20000 });
    await page.click('.btn-welcome.guest');
    await page.waitForSelector('.switcher-button', { timeout: 25000 });
    await page.waitForFunction(() => !!window.__bpSocket, { timeout: 25000 });
    await page.evaluate(async () => {
      await document.fonts.ready;
      window.__bpSocket.dispatchEvent(
        new MessageEvent('message', {
          data: JSON.stringify({
            type: 'combatStart',
            message: 'Layout B refine preview',
            targetId: 'bp2-wolf',
            players: [
              {
                id: 'bp2-self',
                type: 'player',
                name: 'Gimli',
                hp: 201,
                maxHp: 211,
                level: 8,
                classId: 'Warrior',
                isAlive: true,
                statusEffects: [
                  { id: 'focused', name: 'Focused', duration: 2, kind: 'buff' },
                ],
              },
            ],
            enemies: [
              {
                id: 'bp2-wolf',
                type: 'npc',
                name: 'Timber Wolf',
                hp: 115,
                maxHp: 120,
                level: 6,
                threat: 'yellow',
                isAlive: true,
                statusEffects: [
                  { id: 'bleed', name: 'Bleed', duration: 3, kind: 'debuff' },
                ],
              },
            ],
          }),
        })
      );
    });
    await page.waitForSelector('.battle-stage.layout-b', { timeout: 10000 });
    await page.waitForSelector('.layout-b-frame.player-frame', { timeout: 5000 });
    await page.waitForSelector('.layout-b-frame.target-frame', { timeout: 5000 });
    await page.waitForSelector('.sprite-hp', { timeout: 5000 });
    const check = await page.evaluate(() => {
      const stage = document.querySelector('.battle-stage.layout-b');
      const playerFrame = document.querySelector('.layout-b-frame.player-frame');
      const targetFrame = document.querySelector('.layout-b-frame.target-frame');
      const foePlate = document.querySelector('.battle-stage.layout-b .foe-plate');
      const playerMeta = document.querySelector('.battle-stage.layout-b .player-panel .player-meta');
      const spriteHp = document.querySelectorAll('.battle-stage.layout-b .sprite-hp');
      const spriteHpNames = document.querySelectorAll('.battle-stage.layout-b .sprite-hp-name');
      const waitChip = document.querySelector('.battle-stage.layout-b .dock-status .queued-chip.wait');
      const marker = document.querySelector('.battle-stage.layout-b .target-ring');
      const foeDisplay = foePlate ? getComputedStyle(foePlate).display : 'missing';
      const metaDisplay = playerMeta ? getComputedStyle(playerMeta).display : 'missing';
      const nameDisplay = spriteHpNames[0] ? getComputedStyle(spriteHpNames[0]).display : 'none';
      const waitDisplay = waitChip ? getComputedStyle(waitChip).display : 'none';
      const markerWidth = marker ? getComputedStyle(marker).width : '';
      return {
        hasStage: !!stage,
        playerFrameText: playerFrame?.textContent?.replace(/\s+/g, ' ').trim(),
        targetFrameText: targetFrame?.textContent?.replace(/\s+/g, ' ').trim(),
        foePlateDisplay: foeDisplay,
        playerMetaDisplay: metaDisplay,
        spriteHpCount: spriteHp.length,
        spriteHpNameDisplay: nameDisplay,
        dockWaitDisplay: waitDisplay,
        markerWidth,
        height: document.scrollingElement.scrollHeight,
        htmlHasBattlepoc3: !!document.querySelector('script[src*="battlepoc3"]'),
      };
    });
    assert.equal(check.hasStage, true);
    assert.match(check.playerFrameText, /Gimli/i);
    assert.match(check.targetFrameText, /Timber Wolf/i);
    assert.equal(check.foePlateDisplay, 'none');
    assert.equal(check.playerMetaDisplay, 'none');
    assert.ok(check.spriteHpCount >= 1);
    assert.equal(check.spriteHpNameDisplay, 'none');
    assert.equal(check.dockWaitDisplay, 'none');
    assert.equal(check.htmlHasBattlepoc3, true);
    assert.deepEqual(errors, []);
    fs.mkdirSync(path.dirname(out), { recursive: true });
    await page.screenshot({ path: out });
    fs.writeFileSync(
      path.resolve(__dirname, '../.director/ux-audit/after/battlepoc3-smoke.json'),
      JSON.stringify({ out, check, errors }, null, 2)
    );
    // Crop TL / TR / sprites for review
    const tl = await page.$('.layout-b-frame.player-frame');
    if (tl) await tl.screenshot({ path: path.resolve(__dirname, '../.director/ux-audit/after/battlepoc3-tl.png') });
    const tr = await page.$('.layout-b-frame.target-frame');
    if (tr) await tr.screenshot({ path: path.resolve(__dirname, '../.director/ux-audit/after/battlepoc3-tr.png') });
    const enemySprite = await page.$('.battle-stage.layout-b .enemy-sprite-wrap');
    if (enemySprite) await enemySprite.screenshot({ path: path.resolve(__dirname, '../.director/ux-audit/after/battlepoc3-enemy-sprite.png') });
    const playerSprite = await page.$('.battle-stage.layout-b .player-bust');
    if (playerSprite) await playerSprite.screenshot({ path: path.resolve(__dirname, '../.director/ux-audit/after/battlepoc3-player-sprite.png') });
    console.log(JSON.stringify({ out, check, errors }));
  } finally {
    await browser.close();
  }
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
