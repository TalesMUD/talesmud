// Usage: PUPPETEER_MODULE=/path/to/puppeteer-core node tools/capture_worldmap.cjs
// Requires the read-only preview on 127.0.0.1:8137 and optional full content JSON.
const fs=require('node:fs'), path=require('node:path'), assert=require('node:assert/strict');
const puppeteer=require(process.env.PUPPETEER_MODULE||'puppeteer-core');
const root=path.resolve(__dirname,'..'), out=path.join(root,'.director/ux-audit/after');
const settle=page=>page.evaluate(async()=>{await document.fonts.ready;await Promise.all([...document.images].map(i=>i.decode().catch(()=>{})));await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)))});
(async()=>{
 const browser=await puppeteer.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium-browser',headless:'new',args:['--no-sandbox','--disable-dev-shm-usage']});
 try{
  const page=await browser.newPage(), errors=[], sheets=[];
  page.on('pageerror',e=>errors.push(e.message));page.on('response',r=>{if(r.url().includes('/api/map-tiles/'))sheets.push({url:r.url(),status:r.status()})});
  await page.setViewport({width:1920,height:1080});await page.goto('http://127.0.0.1:8137/',{waitUntil:'networkidle0'});
  await page.waitForFunction(()=>window.__mapPreview?.ready);await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),'Sampletown Gate');
  assert.equal(await page.$eval('[title="Fit world"] i',e=>getComputedStyle(e).opacity),'1');
  await page.screenshot({path:path.join(out,'worldmap-p1-overview-1920x1080.png')});
  const bounds=await page.$eval('canvas',e=>{const r=e.getBoundingClientRect();return{x:r.x,y:r.y,w:r.width,h:r.height}});
  await page.mouse.move(bounds.x+bounds.w/2,bounds.y+bounds.h/2);
  const pan=await page.evaluate(({w,h})=>{
   const ps=window.__mapPreview.snapshot.atlas.places.filter(p=>p.layer==='overworld');
   const xs=ps.map(p=>Math.round(p.x)),ys=ps.map(p=>Math.round(p.y));
   const ox=(Math.min(...xs)+Math.max(...xs))/2,oy=(Math.min(...ys)+Math.max(...ys))/2;
   const pad=Math.max(24,Math.min(w,h)*.07);
   const fit=Math.min((w-2*pad)/(Math.max(...xs)-Math.min(...xs)+1),(h-2*pad)/(Math.max(...ys)-Math.min(...ys)+1));
   const near=ps.filter(p=>['Z02_sample_town','Z04_sample_woods'].includes(p.area));
   const nx=near.map(p=>p.x),ny=near.map(p=>p.y);
   const cx=(Math.min(...nx)+Math.max(...nx))/2,cy=(Math.min(...ny)+Math.max(...ny))/2;
   const targetStep=Math.min(w*.8/(Math.max(...nx)-Math.min(...nx)+1),h*.8/(Math.max(...ny)-Math.min(...ny)+1),80);
   const wheels=Math.max(0,Math.round(Math.log(targetStep/fit)/Math.log(1.12)));
   const step=Math.min(110,fit*1.12**wheels);
   return{x:(ox-cx)*step,y:(oy-cy)*step,step,cx,cy,wheels};
  },bounds);
  for(let i=0;i<pan.wheels;i++)await page.mouse.wheel({deltaY:-100});await settle(page);
  const sx=bounds.x+bounds.w*.7,sy=bounds.y+bounds.h*.5;
  await page.mouse.move(sx,sy);await page.mouse.down();await page.mouse.move(sx+pan.x,sy+pan.y,{steps:12});await page.mouse.up();await settle(page);
  await page.screenshot({path:path.join(out,'worldmap-p1-zoom-1920x1080.png')});
  // Select a charted place using its rendered tile, exercising real hit testing.
  const target=await page.evaluate(()=>window.__mapPreview.snapshot.atlas.places.find(p=>p.id==='R0501'));
  assert.ok(target);await page.mouse.click(bounds.x+bounds.w/2+(target.x-pan.cx)*pan.step,bounds.y+bounds.h/2+(target.y-pan.cy)*pan.step);await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),target.name);
  await page.click('[title="Recenter on you"]');await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),target.name,'recenter preserves selection');
  await page.click('[title="Fit world"]');await settle(page);
  await page.click('.layer-tab');await settle(page);
  assert.equal(await page.$eval('.layer-tab.active',e=>e.textContent),'Lower');
  await page.click('.layer-tab:last-child');await settle(page);
  // Time the production canvas painter against the 351-room full content snapshot.
  const stress=process.env.WORLDMAP_CONTENT_JSON?JSON.parse(fs.readFileSync(process.env.WORLDMAP_CONTENT_JSON,'utf8')):null;
  const performanceResult=await page.evaluate(stress=>{
   const atlas=stress?.atlas||window.__mapPreview.snapshot.atlas;
   const canvas=document.createElement('canvas');canvas.width=1205;canvas.height=819;const ctx=canvas.getContext('2d');
   const samples=[];
   for(let i=0;i<90;i++){
    const layer=atlas.layers[i%atlas.layers.length].id;
    const start=performance.now();window.__mapPreview.paintAtlas(ctx,{w:1205,h:819,atlas,activeLayer:layer,visiblePlaces:atlas.places.filter(p=>p.layer===layer),visibleRegions:atlas.regions.filter(r=>r.layer===layer),currentRoomId:atlas.currentRoomId,maximized:true,panX:i%8,panY:0,userScale:1,frameWorld:true});
    if(i>=10)samples.push(performance.now()-start);
   }
   samples.sort((a,b)=>a-b);return{rooms:atlas.places.length,frames:samples.length,medianMs:samples[Math.floor(samples.length/2)],p95Ms:samples[Math.floor(samples.length*.95)],maxMs:samples.at(-1)};
  },stress);
  assert.equal(sheets.length,1,'one sheet request for all terrain/layers');assert.equal(sheets[0].status,200);assert.deepEqual(errors,[]);
  assert.ok(performanceResult.p95Ms<16.7,'canvas painter should fit a 60fps frame on this machine');
  await page.setViewport({width:390,height:844});await settle(page);
  const phone=await page.$eval('#map-overview-overlay',e=>{const r=e.getBoundingClientRect();return{w:r.width,h:r.height,scroll:document.scrollingElement.scrollHeight}});
  assert.equal(phone.w,390);assert.equal(phone.h,844);
  await page.keyboard.press('Escape');assert.equal(await page.$('#map-overview-overlay'),null);
  const report={source:'Local read-only production-overlay preview; fully explored fixture with fog neighbors. No VPS access.',sheets,errors,performance:performanceResult,phone,checks:['fonts/images ready','overview/zoom','click selects Sample Woods room','recenter','world fit','layer switching','phone bounds','Escape closes']};
  fs.writeFileSync(path.join(out,'worldmap-p1-browser-checks.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report,null,2));
 }finally{await browser.close()}
})().catch(e=>{console.error(e);process.exitCode=1});
