// Read-only production overlay preview. See WORLDMAP-PREVIEW.md for setup.
// PUPPETEER_MODULE=/path/to/puppeteer-core node tools/capture_worldmap_p1b.cjs
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const puppeteer=require(process.env.PUPPETEER_MODULE||'puppeteer-core');
const out=path.resolve(__dirname,'../.director/ux-audit/after');
const settle=page=>page.evaluate(async()=>{await document.fonts.ready;await Promise.all([...document.images].map(i=>i.decode().catch(()=>{})));await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)))});
(async()=>{
 const browser=await puppeteer.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium-browser',headless:'new',args:['--no-sandbox','--disable-dev-shm-usage']});
 try {
  const page=await browser.newPage(),errors=[],sheets=[];
  page.on('pageerror',e=>errors.push(e.message));page.on('response',r=>{if(r.url().includes('/api/map-tiles/'))sheets.push({url:r.url(),status:r.status()})});
  await page.setViewport({width:1920,height:1080});
  await page.goto(process.env.WORLDMAP_PREVIEW_URL||'http://127.0.0.1:8138/',{waitUntil:'networkidle0'});
  await page.waitForFunction(()=>window.__mapPreview?.ready);await settle(page);
  const screenshot=kind=>page.screenshot({path:path.join(out,`worldmap-p1b-${kind}-1920x1080.png`)});
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),'Sampletown Gate');
  await screenshot('overview');
  const bounds=await page.$eval('canvas',e=>{const r=e.getBoundingClientRect();return{x:r.x,y:r.y,w:r.width,h:r.height}});
  const pan=await page.evaluate(({w,h})=>{
   const atlas=window.__mapPreview.snapshot.atlas;
   const b=window.__mapPreview.landscapeModel(atlas).bounds;
   const ox=(b.minX+b.maxX)/2,oy=(b.minY+b.maxY)/2;
   const pad=Math.max(24,Math.min(w,h)*.07);
   const fit=Math.min((w-2*pad)/(b.maxX-b.minX+1),(h-2*pad)/(b.maxY-b.minY+1));
   const near=window.__mapPreview.surfaceGroups(atlas.places).filter(p=>['Z02_sample_town','Z04_sample_woods'].includes(p.area));
   const xs=near.map(p=>p.x),ys=near.map(p=>p.y);
   const cx=(Math.min(...xs)+Math.max(...xs))/2,cy=(Math.min(...ys)+Math.max(...ys))/2;
   const targetStep=Math.min(w*.74/(Math.max(...xs)-Math.min(...xs)+3),h*.78/(Math.max(...ys)-Math.min(...ys)+3),72);
   const wheels=Math.max(0,Math.round(Math.log(targetStep/fit)/Math.log(1.12)));
   const scale=1.12**wheels,step=Math.min(110,fit*scale);
   return{x:(ox-cx)*step,y:(oy-cy)*step,step,cx,cy,wheels,scale};
  },bounds);
  await page.mouse.move(bounds.x+bounds.w/2,bounds.y+bounds.h/2);
  for(let i=0;i<pan.wheels;i++)await page.mouse.wheel({deltaY:-100});await settle(page);
  const sx=bounds.x+bounds.w*.55,sy=bounds.y+bounds.h*.65;
  await page.mouse.move(sx,sy);await page.mouse.down();await page.mouse.move(sx+pan.x,sy+pan.y,{steps:12});await page.mouse.up();await settle(page);
  await screenshot('zoom');
  const target=await page.evaluate(()=>window.__mapPreview.snapshot.atlas.places.find(p=>p.id==='R0501'));
  await page.mouse.click(bounds.x+bounds.w/2+(target.x-pan.cx)*pan.step,bounds.y+bounds.h/2+(target.y-pan.cy)*pan.step);await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),target.name,'real exterior hit selects room');
  // Filler cells carry no hit target, even at close zoom.
  const filler=await page.evaluate(({w,h,scale,x,y})=>{
   const atlas=window.__mapPreview.snapshot.atlas,c=document.createElement('canvas');c.width=w;c.height=h;
   const result=window.__mapPreview.paintAtlas(c.getContext('2d'),{w,h,atlas,activeLayer:'overworld',visiblePlaces:atlas.places.filter(p=>p.layer==='overworld'),visibleRegions:atlas.regions.filter(r=>r.layer==='overworld'),currentRoomId:atlas.currentRoomId,maximized:true,panX:x,panY:y,userScale:scale,frameWorld:true});
   const hit=p=>result.hits.some(v=>Math.abs(v.px-p.px)<=v.half&&Math.abs(v.py-p.py)<=v.half);
   const cam=result.camera;
   const candidates=atlas.landscape.map(p=>({px:w/2+(p.x-cam.ox)*cam.tileStep+cam.panX,py:h/2+(p.y-cam.oy)*cam.tileStep+cam.panY}));
   return candidates.find(p=>p.px>30&&p.py>30&&p.px<w-30&&p.py<h-30&&!hit(p));
  },{...bounds,scale:pan.scale,x:pan.x,y:pan.y});
  assert.ok(filler,'decorative ground available outside rooms/buildings');
  await page.mouse.click(bounds.x+filler.px,bounds.y+filler.py);await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),target.name,'filler must not change selection');
  await page.click('[title="Recenter on you"]');await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),target.name,'recenter preserves selection');
  await page.click('[title="Fit world"]');await settle(page);
  // Return to the current gate through production canvas hit testing.
  const gate=await page.evaluate(({w,h})=>{
   const atlas=window.__mapPreview.snapshot.atlas,c=document.createElement('canvas');c.width=w;c.height=h;
   return window.__mapPreview.paintAtlas(c.getContext('2d'),{w,h,atlas,activeLayer:'overworld',visiblePlaces:atlas.places.filter(p=>p.layer==='overworld'),visibleRegions:atlas.regions.filter(r=>r.layer==='overworld'),currentRoomId:atlas.currentRoomId,maximized:true,panX:0,panY:0,userScale:1,frameWorld:true}).hits.find(h=>h.place.id==='R0201');
  },bounds);
  await page.mouse.click(bounds.x+gate.px,bounds.y+gate.py);await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),'Sampletown Gate');
  const interiorCount=await page.$$eval('.room-choice[data-room-id]',els=>els.length);assert.ok(interiorCount>6);
  await page.type('[aria-label="Filter buildings"]','Ground Floor');await settle(page);
  assert.equal(await page.$$eval('.room-choice[data-room-id]',els=>els.length),1);
  await page.click('[data-room-id="R0203"]');await settle(page);
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),'The Weary Wanderer - Ground Floor');
  assert.equal(await page.$eval('.layer-tab.active',e=>e.textContent),'Overworld');
  assert.ok(await page.$('[data-entrance-id="R0215"]'));
  await page.click('[data-entrance-id="R0215"]');await settle(page);
  assert.equal(await page.$eval('.layer-tab.active',e=>e.textContent),'Lower');
  assert.equal(await page.$eval('.intel-title',e=>e.textContent),'The Weary Wanderer - Cellar');
  const lowerPan=await page.evaluate(({w,h})=>{
   const ps=window.__mapPreview.snapshot.atlas.places.filter(p=>p.layer==='lower');
   const xs=ps.map(p=>p.x),ys=ps.map(p=>p.y),pad=Math.max(24,Math.min(w,h)*.07);
   const ox=(Math.min(...xs)+Math.max(...xs))/2,oy=(Math.min(...ys)+Math.max(...ys))/2;
   const fit=Math.min((w-2*pad)/(Math.max(...xs)-Math.min(...xs)+1),(h-2*pad)/(Math.max(...ys)-Math.min(...ys)+1));
   const near=ps.filter(p=>['Z02_sample_town','Z00_sample_crypt','Z03_sample_marsh'].includes(p.area));
   const nx=near.map(p=>p.x),ny=near.map(p=>p.y),cx=(Math.min(...nx)+Math.max(...nx))/2,cy=(Math.min(...ny)+Math.max(...ny))/2;
   const target=Math.min(w*.65/(Math.max(...nx)-Math.min(...nx)+3),h*.75/(Math.max(...ny)-Math.min(...ny)+3),48);
   const wheels=Math.max(0,Math.round(Math.log(target/fit)/Math.log(1.12))),step=fit*1.12**wheels;
   return{x:(ox-cx)*step,y:(oy-cy)*step,wheels};
  },bounds);
  await page.mouse.move(bounds.x+bounds.w/2,bounds.y+bounds.h/2);
  for(let i=0;i<lowerPan.wheels;i++)await page.mouse.wheel({deltaY:-100});await settle(page);
  await page.mouse.move(sx,sy);await page.mouse.down();await page.mouse.move(sx+lowerPan.x,sy+lowerPan.y,{steps:12});await page.mouse.up();await settle(page);
  await screenshot('lower');
  const performanceResult=await page.evaluate(()=>{
   const atlas=window.__mapPreview.snapshot.atlas,w=1205,h=819;
   const c=document.createElement('canvas');c.width=w;c.height=h;const ctx=c.getContext('2d');
   const params=(a,layer,i=0)=>({w,h,atlas:a,activeLayer:layer,visiblePlaces:a.places.filter(p=>p.layer===layer),visibleRegions:a.regions.filter(r=>r.layer===layer),currentRoomId:a.currentRoomId,maximized:true,panX:i%8,panY:0,userScale:1+(i%3)*.05,frameWorld:true});
   const cold=structuredClone(atlas);
   // Force a genuinely changed ground scene, then separately time a new atlas
   // whose only change is the player position (the common movement update).
   const cell=cold.landscape.find(p=>p.terrain!=='fog');cell.terrain=cell.terrain==='grassland'?'forest':'grassland';
   const start=performance.now();window.__mapPreview.paintAtlas(ctx,params(cold,'overworld'));const coldMs=performance.now()-start;
   const moved={...structuredClone(cold),currentRoomId:'R0203'},moveStart=performance.now();
   window.__mapPreview.paintAtlas(ctx,params(moved,'overworld'));const markerUpdateMs=performance.now()-moveStart;
   const measure=layer=>{const samples=[];for(let i=0;i<90;i++){const s=performance.now();window.__mapPreview.paintAtlas(ctx,params(cold,layer,i));if(i>=10)samples.push(performance.now()-s)}samples.sort((a,b)=>a-b);return{frames:samples.length,medianMs:samples[40],p95Ms:samples[76],maxMs:samples.at(-1)}};
   return{rooms:atlas.places.length,landCells:atlas.landscape.length,changedSceneBakeMs:coldMs,markerOnlySnapshotMs:markerUpdateMs,overworld:measure('overworld'),lower:measure('lower')};
  });
  for(const v of [performanceResult.overworld,performanceResult.lower])assert.ok(v.p95Ms<16.7,'cached canvas painter fits a 60fps draw budget here');
  // Interior-only discovery and instanced player position exercise the shared painter.
  const marker=await page.evaluate(()=>{
   const atlas=window.__mapPreview.snapshot.atlas,inn=atlas.places.find(p=>p.id==='R0203');
   const a={...atlas,places:atlas.places.map(p=>({...p,discovered:p.id===inn.id,name:p.id===inn.id?p.name:'',terrain:p.id===inn.id?p.terrain:'fog',entrances:[]})),landscape:atlas.landscape.map(c=>({...c,terrain:'fog'}))};
   const c=document.createElement('canvas');const r=window.__mapPreview.paintAtlas(c.getContext('2d'),{w:800,h:600,atlas:a,activeLayer:'overworld',visiblePlaces:a.places.filter(p=>p.layer==='overworld'),visibleRegions:[],currentRoomId:inn.id+'~player',maximized:true,panX:0,panY:0,userScale:1,frameWorld:true});
   const here=r.hits.filter(h=>h.place.current);return{ids:[...new Set(here.map(h=>h.place.id))],anchor:window.__mapPreview.groupForRoom(window.__mapPreview.surfaceGroups(a.places),inn.id+'~player').id};
  });
  assert.deepEqual(marker.ids,['R0203']);assert.equal(marker.anchor,'R0201');
  const cache=await page.evaluate(()=>{
   const make=terrain=>({places:[],paths:[],landscape:[{x:0,y:0,terrain}]});
   const a=make('grassland'),b={...make('grassland'),currentRoomId:'moved'},fog=make('fog');
   const raster=window.__mapPreview.continentRaster;
   const first=raster(a,null),same=raster(b,null),hidden=raster(fog,null);
   const pixel=ctx=>Array.from(ctx.getImageData(16,16,1,1).data);
   return{markerReusesCanvas:first.canvas===same.canvas,fogReusesCanvas:first.canvas===hidden.canvas,knownPixel:pixel(first.canvas.getContext('2d')),fogPixel:pixel(hidden.canvas.getContext('2d'))};
  });
  assert.ok(cache.markerReusesCanvas,'marker-only atlas snapshots reuse the raster');
  assert.equal(cache.fogReusesCanvas,false,'fog disclosure must change the raster cache key');
  assert.notDeepEqual(cache.knownPixel,cache.fogPixel,'fog scene cannot reuse known terrain pixels');

  assert.equal(sheets.length,1);assert.equal(sheets[0].status,200);assert.deepEqual(errors,[]);
  await page.setViewport({width:390,height:844});await settle(page);
  const phone=await page.$eval('#map-overview-overlay',e=>{const r=e.getBoundingClientRect();return{w:r.width,h:r.height,scroll:document.scrollingElement.scrollHeight}});
  assert.equal(phone.w,390);assert.equal(phone.h,844);assert.equal(phone.scroll,844);
  await page.keyboard.press('Escape');assert.equal(await page.$('#map-overview-overlay'),null);
  const report={source:'Local read-only 351-room content snapshot in production Svelte overlay; no VPS or game session.',sheets,errors,interiorCount,marker,cache,performance:performanceResult,phone,checks:['fonts/images loaded','continent overview','Sampletown + Sample Woods zoom/pan','exterior hit selection','decorative ground not clickable','recenter preserves selection','Fit world','town filter','interior selection','entrance switches Lower','instance marker on surface anchor','phone bounds','Escape']};
  fs.writeFileSync(path.join(out,'worldmap-p1b-browser-checks.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report,null,2));
 }finally{await browser.close()}
})().catch(e=>{console.error(e);process.exitCode=1});
