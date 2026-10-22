import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(process.env.UX_PLAYWRIGHT_PACKAGE || new URL('../package.json', import.meta.url));
const { chromium } = require('playwright');
const base = process.env.UX_BASE || 'http://127.0.0.1:8016/play/?v=phone1';
const outDir = process.env.UX_OUT || '/tmp/phone-panels-proof';
const sizes = [
  { id: '375x812', width: 375, height: 812, deviceScaleFactor: 3, hasTouch: true, isMobile: true, touch: true },
  { id: '768x1024', width: 768, height: 1024, deviceScaleFactor: 2, hasTouch: true, isMobile: true, touch: true },
  { id: '812x375', width: 812, height: 375, deviceScaleFactor: 2, hasTouch: true, isMobile: true, touch: true },
  { id: '1440x900', width: 1440, height: 900, deviceScaleFactor: 1, hasTouch: false, isMobile: false, touch: false },
].filter((size) => !process.env.UX_ONLY || size.id === process.env.UX_ONLY);

function channel(hex, index) {
  const raw = hex.replace('#', '');
  const n = parseInt(raw.slice(index * 2, index * 2 + 2), 16) / 255;
  return n <= 0.03928 ? n / 12.92 : ((n + 0.055) / 1.055) ** 2.4;
}
function contrast(fg, bg) {
  const lum = (hex) => 0.2126 * channel(hex, 0) + 0.7152 * channel(hex, 1) + 0.0722 * channel(hex, 2);
  const a = lum(fg);
  const b = lum(bg);
  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
}

const report = { base, sizes: {}, failures: [] };
function fail(size, label, detail) {
  report.failures.push({ size, label, detail });
  console.log('FAIL', size, label, JSON.stringify(detail));
}

const browser = await chromium.launch({ headless: true, args: ['--no-sandbox', '--disable-dev-shm-usage'] });

async function enter(page) {
  await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.locator('button.btn-welcome.guest').click();
  await page.waitForSelector('.hotbar-widget .slot', { timeout: 15000 });
  await page.waitForTimeout(400);
}

async function inject(page, messages) {
  await page.evaluate((messages) => {
    for (const message of messages) {
      window.__auditSocket.dispatchEvent(new MessageEvent('message', { data: JSON.stringify(message) }));
    }
  }, messages);
  await page.waitForTimeout(150);
}

async function armSendSpy(page) {
  await page.evaluate(() => {
    window.__sent = [];
    const socket = window.__auditSocket;
    if (socket.__phoneSpy) return;
    const orig = socket.send.bind(socket);
    socket.send = (data) => {
      window.__sent.push(String(data));
      return orig(data);
    };
    socket.__phoneSpy = true;
  });
}

async function sent(page) {
  return page.evaluate(() => window.__sent || []);
}

async function measure(page, selector) {
  const locator = page.locator(selector).first();
  if (!(await locator.count())) return null;
  await locator.scrollIntoViewIfNeeded().catch(() => {});
  return locator.evaluate((el) => {
    const r = el.getBoundingClientRect();
    const x = Math.min(Math.max(r.x + r.width / 2, 0), window.innerWidth - 1);
    const y = Math.min(Math.max(r.y + r.height / 2, 0), window.innerHeight - 1);
    const top = document.elementFromPoint(x, y);
    const cs = getComputedStyle(el);
    return {
      width: r.width,
      height: r.height,
      top: r.top,
      left: r.left,
      hit: !!top && (top === el || el.contains(top)),
      fontSize: parseFloat(cs.fontSize),
      lineHeight: cs.lineHeight,
      color: cs.color,
      background: cs.backgroundColor,
    };
  });
}

function rgbHex(color) {
  const match = String(color || '').match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
  if (!match) return null;
  return '#' + [1, 2, 3].map((i) => Number(match[i]).toString(16).padStart(2, '0')).join('');
}

const fixtureMessages = [
  {
    type: 'characterUpdate',
    currentHitPoints: 30,
    maxHitPoints: 30,
    level: 1,
    gold: 15,
    equippedSkills: ['rogue_slip'],
    maxSkillSlots: 4,
    attributes: [{ name: 'Strength', value: 14 }],
  },
  {
    type: 'inventoryUpdate',
    inventory: { items: [{ id: 'layout-helm', name: 'Layout helm', type: 'armor', slot: 'head', quality: 'uncommon', quantity: 1 }] },
    equippedItems: { main_hand: { id: 'layout-sword', name: 'Layout blade', type: 'weapon', slot: 'main_hand', quality: 'rare' } },
    gold: 15,
  },
  {
    type: 'questLog',
    quests: [{
      questId: 'layout-errand',
      questName: 'Layout errand',
      description: 'Carry a note across the yard.',
      status: 'active',
      category: 'side',
      area: 'Layout yard',
      turnInRoomId: 'layout-room',
      turnInNpcName: 'Training guide',
      objectives: [{ description: 'Speak to the guide', current: 0, required: 1, completed: false }],
      rewards: { xp: 10, gold: 2, itemTemplateIds: [] },
    }],
  },
];

try {
  for (const size of sizes) {
    const context = await browser.newContext({
      viewport: { width: size.width, height: size.height },
      deviceScaleFactor: size.deviceScaleFactor,
      hasTouch: size.hasTouch,
      isMobile: size.isMobile,
      reducedMotion: 'reduce',
    });
    const page = await context.newPage();
    page.setDefaultTimeout(12000);
    await context.addInitScript(() => {
      const Original = window.WebSocket;
      window.WebSocket = class extends Original {
        constructor(...args) {
          super(...args);
          window.__auditSocket = this;
        }
      };
    });
    const row = { touch: size.touch, checks: {} };
    report.sizes[size.id] = row;
    await enter(page);
    await armSendSpy(page);
    await inject(page, fixtureMessages);
    const mobile = (await page.locator('.mobile-layout').count()) > 0;
    row.mobile = mobile;
    if (size.touch && !mobile) fail(size.id, 'mobile-shell', 'touch size did not use the phone shell');
    if (!size.touch && mobile) fail(size.id, 'desktop-shell', 'desktop used the phone shell');

    if (mobile) {
      const hotbar = page.locator('.hotbar-widget');
      await hotbar.evaluate((el) => {
        const scroller = el.closest('.mobile-room');
        const tab = document.querySelector('.mobile-tab-bar');
        if (!scroller || !tab) return;
        const tabTop = tab.getBoundingClientRect().top;
        const rect = el.getBoundingClientRect();
        scroller.scrollTop += rect.bottom - (tabTop - 8);
      });
      const slots = await page.locator('.hotbar-widget .slot').evaluateAll((elements) => elements.map((el) => {
        const r = el.getBoundingClientRect();
        return { width: r.width, height: r.height, right: r.right };
      }));
      row.hotbar = { count: slots.length, slots };
      if (slots.length !== 9) fail(size.id, 'hotbar-count', slots.length);
      if (slots.some((slot) => slot.width < 44 || slot.height < 44)) fail(size.id, 'hotbar-44', slots);
      if (slots.some((slot) => slot.right > size.width + 1)) fail(size.id, 'hotbar-overflow', slots);

      const controls = {
        stats: '[data-sheet="character"]',
        equip: '[data-sheet="equipment"]',
        skills: '[data-sheet="skills"]',
        quests: '[data-sheet="quests"]',
        account: '.acct-btn',
        badge: 'button.player-badge',
      };
      for (const [name, selector] of Object.entries(controls)) {
        const box = await measure(page, selector);
        row.checks[name] = box;
        if (!box || box.width < 44 || box.height < 44 || !box.hit) fail(size.id, name, box);
      }

      await page.locator('.hotbar-widget').evaluate((el) => {
        const scroller = el.closest('.mobile-room');
        const tab = document.querySelector('.mobile-tab-bar');
        if (!scroller || !tab) return;
        const tabTop = tab.getBoundingClientRect().top;
        const rect = el.getBoundingClientRect();
        scroller.scrollTop += rect.bottom - (tabTop - 8);
      });
      await page.evaluate(() => new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve))));
      await page.evaluate(() => { window.__sent = []; });
      await page.locator('[data-sheet="character"]').click();
      await page.waitForSelector('.bottom-sheet');
      const statsSent = await sent(page);
      row.statsSent = statsSent;
      if (statsSent.some((line) => /"message"\s*:\s*"character"/i.test(line))) fail(size.id, 'stats-sent-text', statsSent);
      const chrome = await page.evaluate(() => {
        const header = document.querySelector('.mobile-header').getBoundingClientRect();
        const tab = document.querySelector('.mobile-tab-bar').getBoundingClientRect();
        const hot = document.querySelector('.tab-panel:not(.hidden) .hotbar-widget .slot');
        const hotRect = hot.getBoundingClientRect();
        const close = document.querySelector('.sheet-close').getBoundingClientRect();
        const point = (x, y) => {
          const el = document.elementFromPoint(x, y);
          return el ? (el.getAttribute('aria-label') || el.className || el.tagName) : null;
        };
        const owns = (x, y, selector) => {
          const el = document.elementFromPoint(x, y);
          const want = document.querySelector(selector);
          return !!el && !!want && (el === want || want.contains(el));
        };
        return {
          header: owns(header.x + header.width / 2, header.y + header.height / 2, '.mobile-header'),
          tab: owns(tab.x + 20, tab.y + tab.height / 2, '.mobile-tab-bar'),
          hotbar: owns(hotRect.x + hotRect.width / 2, hotRect.y + hotRect.height / 2, '.hotbar-widget'),
          close: owns(close.x + close.width / 2, close.y + close.height / 2, '.sheet-close'),
          backdrop: point(4, header.bottom + 3),
          closeBox: { width: close.width, height: close.height },
        };
      });
      row.chrome = chrome;
      if (!chrome.header) fail(size.id, 'header-hit', chrome);
      if (!chrome.tab) fail(size.id, 'tab-hit', chrome);
      if (!chrome.hotbar) fail(size.id, 'hotbar-hit', chrome);
      if (!chrome.close || chrome.closeBox.width < 44 || chrome.closeBox.height < 44) fail(size.id, 'sheet-close', chrome);
      if (!String(chrome.backdrop || '').includes('bottom-sheet-backdrop')) fail(size.id, 'backdrop-rim', chrome);

      await page.locator('.sheet-close').click();
      await page.waitForSelector('.bottom-sheet', { state: 'detached' });

      await page.locator('[data-sheet="equipment"]').click();
      await page.waitForSelector('[data-slot="main_hand"]');
      const quality = await page.locator('[data-slot="main_hand"]').getAttribute('data-quality');
      row.quality = quality;
      if (!quality || quality === 'normal') fail(size.id, 'quality-border', quality);
      await page.locator('[data-slot="main_hand"]').click();
      await page.waitForSelector('.item-card');
      const cardLayer = await page.locator('.item-card-backdrop').evaluate((el) => getComputedStyle(el).zIndex);
      row.itemCardZ = cardLayer;
      if (Number(cardLayer) < 2001) fail(size.id, 'item-card-z', cardLayer);
      await page.getByRole('button', { name: 'Unequip' }).click();
      const unequip = await sent(page);
      row.unequip = unequip.filter((line) => line.includes('unequip'));
      if (!row.unequip.length) fail(size.id, 'unequip', unequip);
      await page.locator('.sheet-close').click();

      await page.locator('[data-sheet="skills"]').click();
      await page.waitForSelector('text=Add to hotbar');
      const add = await measure(page, 'button.skill-btn');
      row.checks.addHotbar = add;
      if (!add || add.width < 44 || add.height < 44) fail(size.id, 'add-hotbar', add);
      await page.locator('button.skill-btn', { hasText: 'Add to hotbar' }).first().click();
      await page.locator('.sheet-close').click();

      await page.locator('[data-sheet="quests"]').click();
      await page.waitForSelector('button.quest-name');
      await page.evaluate(() => {
        const button = [...document.querySelectorAll('button.quest-name')].find((el) => el.textContent.includes('Layout errand'));
        const scroller = button.closest('.questlog-content');
        const delta = button.getBoundingClientRect().top - scroller.getBoundingClientRect().top;
        scroller.style.scrollBehavior = 'auto';
        scroller.style.overflowAnchor = 'none';
        scroller.scrollTop += delta;
      });
      const questHit = await page.locator('button.quest-name', { hasText: 'Layout errand' }).evaluate((el) => {
        const sheet = document.querySelector('.sheet-content').getBoundingClientRect();
        const hotbar = document.querySelector('.tab-panel:not(.hidden) .hotbar-widget');
        const limit = hotbar ? hotbar.getBoundingClientRect().top : sheet.bottom;
        const r = el.getBoundingClientRect();
        const left = Math.max(r.left, sheet.left);
        const right = Math.min(r.right, sheet.right);
        const topY = Math.max(r.top, sheet.top);
        const bottomY = Math.min(r.bottom, sheet.bottom, limit - 1);
        const width = right - left;
        const height = bottomY - topY;
        const x = left + Math.min(24, Math.max(width / 2, 1));
        const y = topY + Math.max(height / 2, 1);
        const top = height > 8 && width > 8 ? document.elementFromPoint(x, y) : null;
        const scroller = el.closest('.questlog-content');
        const widget = el.closest('.questlog-widget');
        const wr = widget ? widget.getBoundingClientRect() : null;
        return {
          width: r.width,
          height: r.height,
          visibleWidth: width,
          visibleHeight: height,
          hit: !!top && (top === el || el.contains(top)),
          top: top ? (top.className || top.tagName) : null,
          button: { top: r.top, bottom: r.bottom },
          sheet: { top: sheet.top, bottom: sheet.bottom },
          widget: wr ? { top: wr.top, bottom: wr.bottom, height: wr.height } : null,
          scroll: scroller ? { top: scroller.scrollTop, height: scroller.scrollHeight, client: scroller.clientHeight } : null,
          limit,
        };
      });
      row.checks.questRow = questHit;
      if (!questHit.hit || questHit.height < 44) fail(size.id, 'quest-row', questHit);
      await page.locator('button.quest-name', { hasText: 'Layout errand' }).evaluate((el) => el.click());
      await page.waitForSelector('.quest-popup');
      const popupText = await page.locator('.quest-popup').innerText();
      row.questPopup = popupText;
      for (const needle of ['Carry a note across the yard.', 'Speak to the guide', '(0/1)', 'Area: Layout yard', 'Turn-in room: layout-room', '10 XP', 'Turn in: Training guide']) {
        if (!popupText.includes(needle)) fail(size.id, 'quest-popup', needle);
      }
      const desc = await measure(page, '.quest-popup .quest-description');
      row.checks.questBody = desc;
      if (!desc || desc.fontSize < 14) fail(size.id, 'quest-font', desc);
      await page.locator('.quest-popup-close').evaluate((el) => el.click());
      await page.locator('.sheet-close').evaluate((el) => el.click());

      await page.locator('.mobile-tab-bar button', { hasText: 'Console' }).click();
      const input = await measure(page, '.tx-input');
      const stepper = await measure(page, '.tx-fontsize-btn');
      const body = await page.locator('.tx-body').evaluate((el) => {
        const cs = getComputedStyle(el);
        return { fontSize: parseFloat(cs.fontSize), lineHeight: cs.lineHeight, color: cs.color, background: cs.backgroundColor };
      });
      row.checks.consoleInput = input;
      row.checks.fontStepper = stepper;
      row.checks.consoleBody = body;
      if (!input || input.height < 44) fail(size.id, 'console-input', input);
      if (!stepper || stepper.width < 44 || stepper.height < 44) fail(size.id, 'font-step', stepper);
      if (!body || body.fontSize < 14) fail(size.id, 'console-font', body);
      await page.evaluate(() => { window.__sent = []; });
      for (const cmd of ['character', 'equipment', 'skills', 'quests', 'quest find the well']) {
        await page.locator('.tx-input').fill(cmd);
        await page.locator('.tx-input').press('Enter');
      }
      await page.waitForTimeout(200);
      const echo = await page.locator('.tx-body').innerText();
      row.consoleEcho = echo;
      if (!echo.includes('> character') || !echo.includes('> quest find the well')) fail(size.id, 'console-echo', echo);
      if ((await page.locator('.bottom-sheet').count()) !== 0) fail(size.id, 'console-opened-sheet', true);
      const consoleSent = await sent(page);
      row.consoleSent = consoleSent;
      for (const cmd of ['character', 'equipment', 'skills', 'quests']) {
        if (!consoleSent.some((line) => line.includes(`"message":"${cmd}"`))) fail(size.id, 'console-send-' + cmd, consoleSent);
      }
    } else {
      row.character = await page.locator('.character-widget').count();
      if (!row.character) fail(size.id, 'desktop-character', 0);
      const slot = await page.locator('.hotbar-widget .slot').first().boundingBox();
      row.desktopHotbar = slot;
    }

    const shotDir = path.join(outDir, 'shots', 'local', size.id);
    fs.mkdirSync(shotDir, { recursive: true });
    await page.screenshot({ path: path.join(shotDir, mobile ? 'console-or-room.png' : 'desktop.png'), animations: 'disabled' });
    await context.close();
    console.log('proof', size.id, 'failures', report.failures.filter((item) => item.size === size.id).length);
  }
} finally {
  await browser.close();
}

const gold = contrast('#ffd24a', '#15110e');
const green = contrast('#3ee04a', '#15110e');
report.contrast = { goldOnPanel: gold, greenOnPanel: green };
fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(path.join(outDir, 'geometry.json'), JSON.stringify(report, null, 2));
if (report.failures.length) {
  console.log(JSON.stringify(report.failures, null, 2));
  process.exit(1);
}
console.log('geometry ok', gold.toFixed(2), green.toFixed(2));
