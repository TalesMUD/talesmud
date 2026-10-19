import fs from 'node:fs';
import {createHash} from 'node:crypto';
import path from 'node:path';
import {createRequire} from 'node:module';

const require = createRequire(process.env.UX_PLAYWRIGHT_PACKAGE || new URL('../package.json',import.meta.url));
const {chromium} = require('playwright');
const phase = process.env.UX_PHASE || 'local';
const base = process.env.UX_BASE || 'http://127.0.0.1:8016/play/?v=uxtop1';
const out = process.env.UX_OUT || '/tmp/ux-top-layout';
const failures = [];
const measurements = [];
const sizes = [
  {id:'phone',width:375,height:812,deviceScaleFactor:3,hasTouch:true,isMobile:true},
  {id:'tablet',width:768,height:1024,deviceScaleFactor:2,hasTouch:true,isMobile:true},
  {id:'desktop',width:1440,height:900,deviceScaleFactor:1},
  {id:'phone-landscape',width:812,height:375,deviceScaleFactor:2,hasTouch:true,isMobile:true},
  {id:'phone-landscape-wide',width:844,height:390,deviceScaleFactor:2,hasTouch:true,isMobile:true},
];
const browser = await chromium.launch({headless:true,args:['--no-sandbox','--disable-dev-shm-usage']});

function check(ok,size,label,detail) {
  measurements.push({size,label,ok,detail});
  if (!ok) failures.push({size,label,detail});
}
async function inject(page,messages) {
  await page.evaluate(messages => {
    for(const message of messages) window.__auditSocket.dispatchEvent(new MessageEvent('message',{data:JSON.stringify(message)}));
  },messages);
  await page.waitForTimeout(100);
}
async function targets(page,selector) {
  return page.locator(selector).evaluateAll(elements => elements.map(el => {
    const r = el.getBoundingClientRect();
    const points = [[r.x+r.width/2,r.y+r.height/2],[r.x+r.width*.25,r.y+r.height*.25],[r.x+r.width*.75,r.y+r.height*.75]];
    return {text:el.textContent.trim().slice(0,60),x:r.x,y:r.y,width:r.width,height:r.height,
      inView:r.x>=0 && r.y>=0 && r.right<=innerWidth+1 && r.bottom<=innerHeight+1,
      hits:points.map(([x,y])=>{ const top=document.elementFromPoint(x,y); return !!top && (top===el || el.contains(top)); })};
  }));
}
async function hitProof(page,size,label,selector,count) {
  const boxes=await targets(page,selector);
  check(boxes.length===count && boxes.every(b=>b.inView && b.hits.every(Boolean)),size,label,boxes);
}
async function shot(page,size,label) {
  const dir=path.join(out,'shots',phase,size);
  fs.mkdirSync(dir,{recursive:true});
  await page.screenshot({path:path.join(dir,`${label}.png`),animations:'disabled'});
}
async function noticesHidden(page,size,label) {
  const count=await page.locator('.quest-moment-layer, .quest-notifications, .room-text-overlay').count();
  check(count===0,size,label,{count});
}
async function closeOutcome(page) {
  try { await page.locator('button.outcome-continue').click({timeout:1500}); }
  catch(error) {
    if(phase==='before') await inject(page,[{type:'combatLeave'}]);
    else throw error;
  }
}
async function roomOverlays(page,id) {
  const room=page.locator('.mobile-room');
  if(await room.count()) await room.evaluate(el=>el.scrollTop=0);
  const npc=page.locator('.entity-card',{hasText:'Training guide'});
  await npc.scrollIntoViewIfNeeded();
  await hitProof(page,id,'npc-card','.entity-card.friendly',1);
  await npc.click();
  await page.waitForSelector('.dialog-overlay');
  const overlay=await page.locator('.dialog-overlay').boundingBox();
  const viewport=page.viewportSize();
  check(overlay.height>=viewport.height-1 && overlay.width>=viewport.width-1,id,'dialog-viewport',overlay);
  await page.locator('.dialog-option-btn').last().scrollIntoViewIfNeeded();
  await hitProof(page,id,'dialog-close','.dialog-dismiss',1);
  await shot(page,id,'dialog');
  await page.locator('.dialog-dismiss').click();
  await page.waitForSelector('.dialog-overlay',{state:'detached'});
  const merchant=page.locator('.entity-card',{hasText:'Training merchant'});
  await merchant.scrollIntoViewIfNeeded();
  await merchant.click();
  await page.waitForSelector('.shop-overlay');
  await hitProof(page,id,'shop-close','[aria-label="Close shop"]',1);
  await shot(page,id,'shop');
  await page.locator('[aria-label="Close shop"]').click();
}
async function panelProof(page,id) {
  await inject(page,[{type:'inventoryUpdate',inventory:{items:[{id:'layout-item',name:'Layout item',type:'misc',quantity:1,quality:'normal'}]},equippedItems:{},gold:15}]);
  await page.locator('button.btn').filter({hasText:/\binv\s*$/i}).first().click();
  await page.waitForSelector('.inv-overlay');
  await hitProof(page,id,'inventory-close','[aria-label="Close inventory"]',1);
  await hitProof(page,id,'inventory-first-row','.inv-overlay .item-slot, .inv-overlay .list-item',1);
  await shot(page,id,'inventory-populated');
  await page.keyboard.press('Escape');
  if(await page.locator('button[aria-label="Party"]').isVisible()) await page.locator('button[aria-label="Party"]').click();
  else {
    const room=page.locator('.mobile-room');
    if(await room.count()) await room.evaluate(el=>el.scrollTop=0);
    await page.locator('button.player-badge').click();
    await page.getByRole('button',{name:'Party',exact:true}).click();
  }
  await page.waitForSelector('.party-panel');
  await hitProof(page,id,'party-invite-input','input[aria-label="Invite player by name"]',1);
  await hitProof(page,id,'party-invite-button','.party-panel button.act.secondary',1);
  await shot(page,id,'party');
  await page.keyboard.press('Escape');
}
async function run(spec) {
  const {id,...options}=spec;
  const context=await browser.newContext({viewport:{width:spec.width,height:spec.height},...options,reducedMotion:'reduce'});
  const page=await context.newPage();
  page.setDefaultTimeout(10000);
  const errors=[];
  const assets=[];
  const reads=[];
  page.on('pageerror',error=>errors.push(error.message));
  page.on('response',response=>{
    const file=new URL(response.url()).pathname.split('/').at(-1);
    if(['bundle.js','extra.css','global.css'].includes(file)) reads.push(response.body().then(body=>assets.push({file,sha256:createHash('sha256').update(body).digest('hex')})));
  });
  await context.addInitScript(() => {
    const Original=window.WebSocket;
    window.WebSocket=class extends Original {
      constructor(...args) {super(...args); window.__auditSocket=this;}
    };
  });
  try {
    const response=await page.goto(base,{waitUntil:'domcontentloaded'});
    check(response.status()===200,id,'play-http',{status:response.status()});
    await page.locator('button.btn-welcome.guest').click();
    await page.waitForSelector('.hotbar-widget .slot');
    await page.waitForTimeout(700);
    await shot(page,id,'room');
    await panelProof(page,id);
    if(process.env.UX_FIXTURE==='1') await roomOverlays(page,id);
    const landscape=spec.width>768 && spec.height<500;
    if(landscape) {
      check(await page.locator('.mobile-layout').count()===1,id,'landscape-mobile-shell',{});
      check(!await page.locator('.back-link').isVisible(),id,'landscape-home-link-hidden',{});
      await page.locator('.hotbar-widget').scrollIntoViewIfNeeded();
      await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
      await hitProof(page,id,'landscape-hotbar','.hotbar-widget .slot',9);
      if(await page.locator('.mobile-layout').count()) {
      await page.locator('.mobile-tab-bar button',{hasText:'Console'}).click();
      await shot(page,id,'console');
      await hitProof(page,id,'landscape-room-tab','.mobile-tab-bar button',2);
      await page.locator('.mobile-tab-bar button',{hasText:'Room'}).click();
      }
    }
    if(await page.locator('.mobile-layout').count()) {
      await inject(page,[{type:'guideHint',text:'Next layout step',command:'look',label:'Look'}]);
      await page.locator('.mobile-tab-bar button',{hasText:'Console'}).click();
      await page.waitForTimeout(500);
      check(!await page.locator('.guide-hint').isVisible(),id,'guide-hidden-in-console',{});
      await hitProof(page,id,'console-tabs','.mobile-tab-bar button',2);
      await page.locator('.mobile-tab-bar button',{hasText:'Room'}).click();
    }
    const start={type:'combatStart',message:'Layout combat fixture',enemies:[{id:'layout-foe',name:'Training foe',hp:20,maxHp:20,type:'enemy'}],players:[]};
    await inject(page,[start]);
    const queued=Array.from({length:6},(_,i)=>({type:'questAccepted',questId:`layout-${i}`,questName:`Layout notice ${i}`,message:'Queued reward information'}));
    await inject(page,[...queued,{type:'message',message:'Queued room reward'}]);
    await noticesHidden(page,id,'notices-hidden-during-combat');
    await page.waitForTimeout(17000);
    await inject(page,[{type:'combatEnd',outcome:'victory',message:'Layout victory',rewards:{baseXp:20,xp:20,baseGold:5,gold:5},loot:Array.from({length:12},(_,i)=>({name:`Reward ${i}`,quantity:1}))},
      {type:'need_greed',pending:true,roundId:'layout-roll',itemName:'Training reward',expiresAt:Math.floor(Date.now()/1000)+30}]);
    await noticesHidden(page,id,'notices-hidden-during-outcome-and-roll');
    await hitProof(page,id,'victory-continue','button.outcome-continue',1);
    await hitProof(page,id,'need-greed-buttons','.ng-actions button',3);
    await shot(page,id,'toast-stress-victory');
    await inject(page,[{type:'need_greed',pending:false}]);
    await closeOutcome(page);
    check(await page.locator('.overlay-message').filter({hasText:'Queued room reward'}).count()>0,id,'room-loot-notice-resumes',{});
    const seen=[];
    for(let i=0;i<6;i++) {
      const title=page.locator('.veilspan-card .card-title');
      if(!await title.count()) break;
      seen.push(await title.textContent());
      await page.locator('.accept-dismiss').click();
    }
    check(queued.every(q=>seen.includes(q.questName)),id,'six-notices-survive-former-ttl',{seen});
    await inject(page,[start,{type:'combatEnd',outcome:'defeat',message:'Layout defeat',defeat:{xpLost:1,goldLost:1,hp:10,maxHp:20}}]);
    await hitProof(page,id,'defeat-continue','button.outcome-continue',1);
    await shot(page,id,'defeat');
    await closeOutcome(page);
  } catch(error) {
    check(false,id,'flow',{message:error.message});
    await shot(page,id,'fixture-error').catch(()=>{});
  } finally {
    await Promise.allSettled(reads);
    check(errors.length===0,id,'runtime-errors',{errors});
    check(assets.length>=2,id,'loaded-build-assets',{assets});
    await context.close();
    fs.mkdirSync(out,{recursive:true});
    fs.writeFileSync(path.join(out,`${phase}-${id}-fixture-proof.json`),JSON.stringify({failures:failures.filter(f=>f.size===id),measurements:measurements.filter(m=>m.size===id)},null,2));
    console.log(`${id} completed with ${failures.filter(f=>f.size===id).length} failures`);
  }
}
try {
  const wanted=process.env.UX_SIZES?.split(',');
  const selected=wanted?sizes.filter(s=>wanted.includes(s.id)):sizes;
  for(let i=0;i<selected.length;i+=2) await Promise.all(selected.slice(i,i+2).map(run));
} finally { await browser.close(); }
fs.writeFileSync(path.join(out,`${phase}-fixture-proof.json`),JSON.stringify({failures,measurements},null,2));
console.log(`${measurements.length} checks, ${failures.length} failures`);
if(failures.length) process.exitCode=1;
