// Capture the production map overlay locally; see WORLDMAP-PREVIEW.md.
// PUPPETEER_MODULE=/path/to/puppeteer-core node tools/capture_worldmap_p1e.cjs
const fs=require('node:fs'),path=require('node:path');
const puppeteer=require(process.env.PUPPETEER_MODULE||'puppeteer-core');
const out=path.resolve(__dirname,'../.director/ux-audit/after');
const settle=page=>page.evaluate(async()=>{await document.fonts.ready;await document.fonts.load('24px "Material Icons"');await Promise.all([...document.images].map(i=>i.decode().catch(()=>{})));await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)));await window.__mapPreview.waitForMapScenes();await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)));await window.__mapPreview.waitForMapScenes();await new Promise(r=>requestAnimationFrame(r))});
(async()=>{
 const browser=await puppeteer.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium-browser',headless:'new',args:['--no-sandbox','--disable-dev-shm-usage']});
 try {
  const page=await browser.newPage(),errors=[],sheets=[],workers=[];
  page.on('workercreated',w=>workers.push({url:w.url()}));page.on('pageerror',e=>errors.push(e.message));
  page.on('response',r=>{if(r.url().includes('/api/map-tiles/'))sheets.push({url:r.url(),status:r.status()})});
  await page.setViewport({width:1920,height:1080});
  await page.goto(process.env.WORLDMAP_PREVIEW_URL||'http://127.0.0.1:8141/',{waitUntil:'networkidle0'});
  await page.waitForFunction(()=>window.__mapPreview?.ready);await settle(page);
  const save=kind=>page.screenshot({path:path.join(out,`worldmap-p1e-${kind}-1920x1080.png`)});
  await save('overview');
  const bounds=await page.$eval('canvas',e=>{const r=e.getBoundingClientRect();return{x:r.x,y:r.y,w:r.width,h:r.height}});
  await page.mouse.move(bounds.x+bounds.w/2,bounds.y+bounds.h/2);
  for(let i=0;i<30;i++)await page.mouse.wheel({deltaY:100});await settle(page);
  await page.mouse.move(30,30);await save('far');
  await page.click('[title="Fit world"]');await settle(page);
  const pan=await page.evaluate(({w,h})=>{
   const atlas=window.__mapPreview.snapshot.atlas,b=window.__mapPreview.landscapeModel(atlas).bounds;
   const ox=(b.minX+b.maxX)/2,oy=(b.minY+b.maxY)/2,pad=Math.max(24,Math.min(w,h)*.07);
   const fit=Math.min((w-2*pad)/(b.maxX-b.minX+1),h*.78/(b.maxY-b.minY+1));
   const near=window.__mapPreview.surfaceGroups(atlas.places).filter(p=>['Z02_sample_town','Z04_sample_woods'].includes(p.area));
   const xs=near.map(p=>p.x),ys=near.map(p=>p.y),cx=(Math.min(...xs)+Math.max(...xs))/2,cy=(Math.min(...ys)+Math.max(...ys))/2;
   const target=Math.min(w*.74/(Math.max(...xs)-Math.min(...xs)+3),h*.78/(Math.max(...ys)-Math.min(...ys)+3),72);
   const wheels=Math.max(0,Math.round(Math.log(target/fit)/Math.log(1.12))),scale=Math.min(5,1.12**wheels),step=Math.min(110,fit*scale);
   return{x:(ox-cx)*step,y:(oy-cy)*step,wheels};
  },bounds);
  await page.mouse.move(bounds.x+bounds.w/2,bounds.y+bounds.h/2);
  for(let i=0;i<pan.wheels;i++)await page.mouse.wheel({deltaY:-100});await settle(page);
  const sx=bounds.x+bounds.w*.55,sy=bounds.y+bounds.h*.65;
  await page.mouse.move(sx,sy);await page.mouse.down();await page.mouse.move(sx+pan.x,sy+pan.y,{steps:12});await page.mouse.up();await settle(page);
  await save('zoom');
  await page.click('[title="Fit world"]');await settle(page);
  const gate=await page.evaluate(({w,h})=>{
   const a=window.__mapPreview.snapshot.atlas,c=document.createElement('canvas');c.width=w;c.height=h;
   return window.__mapPreview.paintAtlas(c.getContext('2d'),{w,h,atlas:a,activeLayer:'overworld',visiblePlaces:a.places.filter(p=>p.layer==='overworld'),visibleRegions:a.regions.filter(r=>r.layer==='overworld'),currentRoomId:a.currentRoomId,maximized:true,panX:0,panY:0,userScale:1,frameWorld:true}).hits.find(h=>h.place.id==='R0201');
  },bounds);
  await page.mouse.click(bounds.x+gate.px,bounds.y+gate.py);await settle(page);
  await page.type('[aria-label="Filter buildings"]','Ground Floor');await settle(page);
  await page.click('[data-room-id="R0203"]');await settle(page);
  await page.click('[data-entrance-id="R0215"]');await settle(page);
  const lowerPan=await page.evaluate(({w,h})=>{
   const a=window.__mapPreview.snapshot.atlas,ps=a.places.filter(p=>p.layer==='lower'),b=window.__mapPreview.undergroundModel(a).bounds;
   const xs=ps.map(p=>p.x),ys=ps.map(p=>p.y),pad=Math.max(24,Math.min(w,h)*.07),ox=(b.minX+b.maxX)/2,oy=(b.minY+b.maxY)/2;
   const fit=Math.min((w-2*pad)/(b.maxX-b.minX+1),h*.78/(b.maxY-b.minY+1));
   const near=ps.filter(p=>['Z02_sample_town','Z00_sample_crypt','Z03_sample_marsh'].includes(p.area));
   const nx=near.map(p=>p.x),ny=near.map(p=>p.y),cx=(Math.min(...nx)+Math.max(...nx))/2,cy=(Math.min(...ny)+Math.max(...ny))/2;
   const target=Math.min(w*.65/(Math.max(...nx)-Math.min(...nx)+3),h*.75/(Math.max(...ny)-Math.min(...ny)+3),48);
   const wheels=Math.max(0,Math.round(Math.log(target/fit)/Math.log(1.12))),step=fit*Math.min(5,1.12**wheels);
   return{x:(ox-cx)*step,y:(oy-cy)*step,wheels};
  },bounds);
  await page.mouse.move(bounds.x+bounds.w/2,bounds.y+bounds.h/2);
  for(let i=0;i<lowerPan.wheels;i++)await page.mouse.wheel({deltaY:-100});await settle(page);
  await page.mouse.move(sx,sy);await page.mouse.down();await page.mouse.move(sx+lowerPan.x,sy+lowerPan.y,{steps:12});await page.mouse.up();await settle(page);
  await save('lower');
  const workerStatus=workers.length?await page.evaluate(async url=>(await fetch(url,{cache:'force-cache'})).status,workers[0].url):null;
  const report={source:'Local read-only 351-room map preview; fonts, sheet and worker scenes loaded.',workers:workers.map(w=>({...w,status:workerStatus})),sheets,errors,screenshots:['overview','zoom','lower','far'].map(n=>`worldmap-p1e-${n}-1920x1080.png`)};
  fs.writeFileSync(path.join(out,'worldmap-p1e-browser-checks.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report,null,2));
 }finally{await browser.close()}
})().catch(e=>{console.error(e);process.exitCode=1});
