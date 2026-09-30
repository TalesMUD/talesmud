// B14 Settings browser smoke. B14_PLAY_URL may point to a local static preview or live /play/.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || '/home/clawd/clawd/node_modules/puppeteer-core');
const playUrl = process.env.B14_PLAY_URL || 'https://veilspan.com/play/';
const out = path.resolve(__dirname, '../.director/ux-audit/after');
const result = { playUrl, checks: {}, pageErrors: [] };
(async () => {
  fs.mkdirSync(out, { recursive: true });
  const browser = await puppeteer.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser', headless: 'new', args: ['--no-sandbox', '--disable-dev-shm-usage', ...(playUrl.includes('localhost:8137') ? ['--disable-web-security'] : [])] });
  const page = await browser.newPage();
  try {
    await page.setViewport({ width: 1920, height: 1080 });
    page.on('pageerror', e => result.pageErrors.push(e.message));
    page.on('response', r => { if (r.request().method() === 'POST' && r.url().endsWith('/api/guest')) result.guestPost = r.status(); });
    const response = await page.goto(playUrl, { waitUntil: 'domcontentloaded' });
    result.playStatus = response.status();
    await page.waitForSelector('.btn-welcome.guest', { timeout: 20000 });
    await page.click('.btn-welcome.guest');
    await page.waitForSelector('.switcher-button', { timeout: 25000 });
    await page.evaluate(async () => { await document.fonts.ready; });
    const openSettings = async () => {
      const phone = (await page.viewport()).width <= 600;
      if (phone) await page.waitForSelector('.mobile-header .acct-btn', { timeout: 10000 });
      await page.click(phone ? '.mobile-header .acct-btn' : '.switcher-button');
      await page.evaluate(() => {
        const button = [...document.querySelectorAll('.menu-item, .acct-menu button')].find(b => b.textContent.trim().endsWith('Settings'));
        if (!button) throw new Error('Settings missing from account chip menu');
        button.click();
      });
      await page.waitForSelector('.modal-container');
    };
    await openSettings();
    const metrics = () => page.evaluate(() => {
      const modal = document.querySelector('.modal-container').getBoundingClientRect();
      return { viewport: [innerWidth, innerHeight], pageHeight: document.scrollingElement.scrollHeight, modal: { x: modal.x, y: modal.y, right: modal.right, bottom: modal.bottom }, reducedMotion: JSON.parse(localStorage.getItem('talesmud_settings_v1') || '{}').interface?.reducedMotion || 'system', activeTab: document.querySelector('.tab-btn.active span')?.textContent.trim() };
    });
    result.checks.desktop = await metrics();
    assert.equal(result.checks.desktop.pageHeight, 1080);
    assert.ok(result.checks.desktop.modal.bottom <= 1080);
    assert.equal(result.checks.desktop.activeTab, 'Interface');
    result.checks.inactiveControlsHidden = await page.evaluate(() =>
      !document.querySelector('input[aria-label="Compact mode"], input[aria-label="Room text overlay"]'));
    assert.equal(result.checks.inactiveControlsHidden, true);
    await page.screenshot({ path: path.join(out, 'b14-settings-1920x1080.png') });
    await page.evaluate(() => [...document.querySelectorAll('.tab-btn')].find(b => b.textContent.includes('General')).click());
    result.checks.audioNote = await page.$eval('.note', e => e.textContent.trim());
    assert.ok(result.checks.audioNote.includes('Game audio is coming soon'));
    assert.equal(await page.$('.tab-content input'), null, 'audio has no dead controls');
    await page.evaluate(() => [...document.querySelectorAll('.tab-btn')].find(b => b.textContent.includes('Interface')).click());
    await page.evaluate(() => {
      [...document.querySelectorAll('[role="group"][aria-label="Reduced motion"] button')].find(b => b.textContent.trim() === 'On').click();
      document.querySelector('input[aria-label="Auto-focus BattleStage"]').click();
      [...document.querySelectorAll('[role="group"][aria-label="Inventory opens as"] button')].find(b => b.textContent.trim() === 'Layout widget').click();
    });
    result.checks.saved = await page.evaluate(() => JSON.parse(localStorage.getItem('talesmud_settings_v1')).interface);
    assert.equal(result.checks.saved.reducedMotion, 'on');
    assert.equal(result.checks.saved.combatAutoFocus, false);
    assert.equal(result.checks.saved.inventoryOpenMode, 'widget');
    await page.reload({ waitUntil: 'domcontentloaded' });
    await page.waitForSelector('.switcher-button', { timeout: 25000 });
    await openSettings();
    result.checks.reloaded = await page.evaluate(() => ({
      iface: JSON.parse(localStorage.getItem('talesmud_settings_v1')).interface,
      on: [...document.querySelectorAll('[role="group"][aria-label="Reduced motion"] button')].find(b => b.textContent.trim() === 'On')?.getAttribute('aria-pressed'),
      combat: document.querySelector('input[aria-label="Auto-focus BattleStage"]')?.checked,
      inventory: [...document.querySelectorAll('[role="group"][aria-label="Inventory opens as"] button')].find(b => b.textContent.trim() === 'Layout widget')?.getAttribute('aria-pressed'),
    }));
    assert.equal(result.checks.reloaded.iface.reducedMotion, 'on');
    assert.equal(result.checks.reloaded.iface.combatAutoFocus, false);
    assert.equal(result.checks.reloaded.iface.inventoryOpenMode, 'widget');
    assert.equal(result.checks.reloaded.on, 'true');
    assert.equal(result.checks.reloaded.combat, false);
    assert.equal(result.checks.reloaded.inventory, 'true');
    await page.keyboard.press('Escape');
    assert.equal(await page.$('.modal-container'), null, 'Escape closes Settings');
    for (const [name, width, height] of [['smallDesktop',1366,768],['phone',390,844]]) {
      await page.setViewport({ width, height });
      await openSettings();
      result.checks[name] = await metrics();
      assert.ok(result.checks[name].pageHeight <= height);
      assert.ok(result.checks[name].modal.bottom <= height);
      assert.ok(result.checks[name].modal.x >= 0);
      await page.keyboard.press('Escape');
    }
    assert.equal(result.playStatus, 200);
    assert.equal(result.guestPost, 200);
    assert.deepEqual(result.pageErrors, []);
    result.ok = true;
  } finally {
    await browser.close();
    fs.writeFileSync(path.join(out, 'b14-settings-smoke.json'), JSON.stringify(result, null, 2));
  }
  console.log(JSON.stringify(result, null, 2));
})().catch(e => { console.error(e); process.exitCode = 1; });
