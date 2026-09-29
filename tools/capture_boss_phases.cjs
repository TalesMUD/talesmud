// Live A9 smoke: only fresh disposable guests are prepared through their owner API.
// Existing boss content, rooms, other characters, and Door are never edited.
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || '/home/clawd/clawd/node_modules/puppeteer-core');
const base = process.env.A9_BASE_URL || 'https://veilspan.com';
const bossName = process.env.A9_BOSS_NAME || 'Burrow Brute';
const roomId = process.env.A9_ROOM_ID || 'R0112';
const guestLevel = Number(process.env.A9_GUEST_LEVEL || 4);
const guestStrength = Number(process.env.A9_GUEST_STRENGTH || 18);
const out = path.resolve(__dirname, '../.director/ux-audit/after');
const sleep = ms => new Promise(r => setTimeout(r, ms));
(async () => {
 const report = { base, preparation: `Fresh owner-authorized guest records only: level ${guestLevel} warrior, 250 HP, STR ${guestStrength}, room ${roomId}. Existing live ${bossName} is unchanged.`, guests: [], phases: [], errors: [] };
 const api = async (url, token, method = 'GET', body) => {
  const r = await fetch(base + '/api/' + url, { method, headers: { ...(token ? {Authorization: 'Bearer ' + token} : {}), 'Content-Type': 'application/json' }, ...(body ? {body: JSON.stringify(body)} : {}) });
  const text = await r.text(); assert.equal(r.status, 200, `${method} ${url}: ${r.status} ${text.slice(0,100)}`); return text ? JSON.parse(text) : null;
 };
 const browser = await puppeteer.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser', headless: 'new', args: ['--no-sandbox', '--disable-dev-shm-usage'] });
 const geometry = page => page.evaluate(() => [...document.querySelectorAll('.svlt-grid-item')].map(e => { const r = e.getBoundingClientRect(); return [r.x,r.y,r.width,r.height]; }));
 const guests = [], pages = [];
 try {
  for (const reduced of [false, true]) {
   const {token} = await api('guest', null, 'POST');
   const [character] = await api('my-characters', token); assert.ok(character?.id);
   const prepared = { ...character, currentRoomID: roomId, level:guestLevel, class:{id:'warrior',name:'Warrior'}, maxHitPoints:250, currentHitPoints:250, attributes:(character.attributes || []).map(a => ({...a,value:a.short.toUpperCase()==='STR' ? guestStrength : 14})) };
   await api('characters/' + character.id, token, 'PUT', prepared);
   guests.push({token, original:character}); report.guests.push({id:character.id,name:character.name,reducedMotion:reduced,guestStatus:200});
   const page = await browser.newPage(); pages.push(page); await page.setViewport({width:1920,height:1080});
   if (reduced) await page.emulateMediaFeatures([{name:'prefers-reduced-motion',value:'reduce'}]);
   page.on('pageerror', e => report.errors.push(e.message));
   await page.evaluateOnNewDocument(token => {
    sessionStorage.setItem('talesmud_guest_token',token);
    const Native = WebSocket; window.__events = [];
    window.WebSocket = class extends Native { constructor(...args) { super(...args); window.__socket=this; this.addEventListener('message', e => { try { window.__events.push(JSON.parse(e.data)); } catch {} }); } };
   }, token);
   await page.goto(base + '/play/', {waitUntil:'domcontentloaded'});
   await page.waitForFunction(roomId => window.__events.some(e => e.type==='enterRoom' && e.room.id===roomId), {timeout:25000},roomId);
   await page.evaluate(async () => { await document.fonts.ready; });
   await page.waitForFunction(() => document.documentElement.classList.contains('material-icons-ready'));
   console.log('GUEST_READY',character.name, roomId,reduced);
  }
  const baseline = await geometry(pages[0]); assert.ok(baseline.length);
  const command = (i,message) => pages[i].evaluate(message => window.__socket.send(JSON.stringify({type:'message',message})),message);
  await command(0,'attack! '+bossName); await pages[0].waitForSelector('.battle-stage',{timeout:15000});
  await pages[1].focus('.xterm-helper-textarea'); await command(1,'attack! '+bossName);
  await Promise.all(pages.map(p => p.waitForSelector('.ally-card',{timeout:15000})));
  // Queue normal player attacks during each decision window, like pressing Attack.
  await Promise.all(pages.map(p => p.evaluate(bossName => {
   window.__a9AutoQueue = setInterval(() => {
    if (document.querySelector('.battle-stage:not(.ending)')) window.__socket.send(JSON.stringify({type:'message',message:'attack '+bossName}));
   }, 1000);
  },bossName)));
  for (const phase of [2,3]) {
   await pages[0].bringToFront();
   await pages[0].waitForFunction(phase => window.__events.some(e => e.type==='combatAction' && e.action==='phase-enter' && e.combatants.some(c=>c.bossPhase===phase)), {timeout:150000,polling:100},phase);
   for (let i=0;i<pages.length;i++) {
    const page=pages[i]; await page.bringToFront();
    await page.waitForSelector('.phase-banner',{timeout:2500});
    const check = await page.evaluate(() => {
     const rect = e => {const r=e.getBoundingClientRect();return {x:r.x,y:r.y,right:r.right,bottom:r.bottom};};
     const self=rect(document.querySelector('.player-panel')), allies=[...document.querySelectorAll('.ally-card')].map(rect);
     return { width:innerWidth,height:innerHeight,scroll:document.scrollingElement.scrollHeight, caption:document.querySelector('.boss-phase-label')?.textContent, banner:document.querySelector('.phase-banner')?.textContent, animation:getComputedStyle(document.querySelector('.phase-banner')).animationName, allies:allies.length, overlap:allies.some(r=>r.x<self.right&&r.right>self.x&&r.y<self.bottom&&r.bottom>self.y), focus:document.activeElement?.className, font:document.fonts.check('24px "Material Icons"') };
    });
    console.log('PHASE_CHECK',JSON.stringify({phase,guest:i,...check}));
    await page.screenshot({path:path.join(out,`a9-phase-${phase}${i ? '-reduced-motion' : ''}-1920x1080.png`)});
    assert.equal(check.width,1920); assert.equal(check.height,1080); assert.ok(check.scroll <= check.height); assert.equal(check.overlap,false); assert.equal(check.allies,1); assert.ok(check.caption.includes(`Phase ${phase} / 3`)); assert.ok(check.banner.includes(`phase ${phase}:`));
    if(i===1) { assert.equal(check.animation,'none'); assert.ok(check.focus.includes('xterm-helper-textarea')); }
    report.phases.push({phase,guest:i,...check});
    console.log('PHASE',JSON.stringify({phase,guest:i,...check}));
   }
  }
  await pages[0].bringToFront(); await pages[0].waitForFunction(() => window.__events.some(e=>e.type==='combatEnd'),{timeout:150000,polling:250});
  report.outcome = await pages[0].evaluate(() => window.__events.find(e=>e.type==='combatEnd')?.outcome); assert.equal(report.outcome,'victory');
  report.events = await Promise.all(pages.map(p => p.evaluate(() => window.__events.filter(e=>e.type==='combatAction' && e.action==='phase-enter').map(e=>({text:e.message,boss:e.combatants.find(c=>c.bossPhase),action:e.action})))));
  for (const events of report.events) { assert.equal(events.length,2); assert.equal(events[0].boss.bossPhase,2); assert.equal(events[1].boss.bossPhase,3); }
  await pages[0].click('.outcome-continue'); await pages[0].waitForSelector('.battle-stage',{hidden:true}); await sleep(300);
  assert.deepEqual(await geometry(pages[0]),baseline); report.restored = true;
  await pages[0].screenshot({path:path.join(out,'a9-phase-restored-1920x1080.png')});
  assert.deepEqual(report.errors,[]); report.ok=true;
 } finally {
  await browser.close();
  // Restore only the prepared guests' original fields after disconnect releases combat.
  await sleep(1200);
  for (const guest of guests) {
   try { const current = await api('characters/' + guest.original.id, guest.token); await api('characters/' + guest.original.id, guest.token,'PUT',{...current,currentRoomID:guest.original.currentRoomID,level:guest.original.level,class:guest.original.class,maxHitPoints:guest.original.maxHitPoints,currentHitPoints:guest.original.currentHitPoints,attributes:guest.original.attributes}); report.guests.find(g=>g.id===guest.original.id).restored=true; }
   catch(e) { report.errors.push('Guest restoration: '+e.message); }
  }
  fs.mkdirSync(out,{recursive:true});fs.writeFileSync(path.join(out,'a9-phase-smoke.json'),JSON.stringify(report,null,2));
 }
})().catch(e=>{console.error(e);process.exitCode=1;});
