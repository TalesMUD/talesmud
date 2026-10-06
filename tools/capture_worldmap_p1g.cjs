// Capture and verify the production overlay locally; see WORLDMAP-PREVIEW.md.
// PUPPETEER_MODULE=/path/to/puppeteer-core node tools/capture_worldmap_p1g.cjs
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const puppeteer = require(process.env.PUPPETEER_MODULE || 'puppeteer-core');
const out = path.resolve(__dirname, '../.director/ux-audit/after');
const settle = page => page.evaluate(async () => {
  await document.fonts.ready;
  await document.fonts.load('24px "Material Icons"');
  await Promise.all([...document.images].map(i => i.decode().catch(() => {})));
  for (let i = 0; i < 3; i++) {
    await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)));
    await window.__mapPreview.waitForMapScenes();
  }
});
const measure = (page, maxScale) => page.evaluate(async maxScale => {
  const p=window.__mapPreview,a=p.snapshot.atlas,w=1205,h=819;
  const c=document.createElement('canvas');c.width=w;c.height=h;const ctx=c.getContext('2d');
  const params=(atlas,layer,scale=1,i=0)=>({w,h,atlas,activeLayer:layer,visiblePlaces:atlas.places.filter(p=>p.layer===layer),visibleRegions:atlas.regions.filter(r=>r.layer===layer),currentRoomId:atlas.currentRoomId,userScale:scale,frameWorld:true,maximized:true,panX:i%8,panY:0});
  const first=p.paintAtlas(ctx,params(a,'overworld')).scene;
  const changed=structuredClone(a),cell=changed.landscape.find(p=>p.terrain==='grassland');cell.terrain='forest';
  let heartbeats=0;const timer=setInterval(()=>heartbeats++,5);
  const start=performance.now();p.paintAtlas(ctx,params(changed,'overworld'));const requestMs=performance.now()-start;
  await p.waitForMapScenes();clearInterval(timer);
  const apply=performance.now(),scene=p.paintAtlas(ctx,params(changed,'overworld')).scene;
  const changedMainMs=requestMs+(scene.dispatchMs||0)+(scene.deliveryMs||0)+performance.now()-apply;
  const moved={...structuredClone(changed),currentRoomId:'R0104'},ms=performance.now();p.paintAtlas(ctx,params(moved,'overworld'));const markerOnlyMs=performance.now()-ms;
  const closeStart=performance.now();p.paintAtlas(ctx,params(changed,'overworld',maxScale));const closeRequestMs=performance.now()-closeStart;
  await p.waitForMapScenes();const closeApply=performance.now();p.paintAtlas(ctx,params(changed,'overworld',maxScale));
  const closeMainMs=closeRequestMs+(scene.closeDispatchMs||0)+(scene.closeDeliveryMs||0)+performance.now()-closeApply;
  const sample=(layer,scale=1)=>{const times=[];for(let i=0;i<90;i++){const start=performance.now();p.paintAtlas(ctx,params(changed,layer,scale,i));if(i>=10)times.push(performance.now()-start)}times.sort((a,b)=>a-b);return {frames:80,medianMs:times[40],p95Ms:times[76],maxMs:times.at(-1)}};
  return {rooms:a.places.length,cells:a.landscape.length,heartbeats,firstWorkerBakeMs:first.bakeMs,firstMainMs:(first.requestMs||0)+(first.dispatchMs||0)+(first.deliveryMs||0)+(first.firstPaintMs||0),changedWorkerBakeMs:scene.bakeMs,changedMainMs,changedElapsedMs:scene.elapsedMs,closeWorkerBakeMs:scene.closeBakeMs,closeMainMs,closeElapsedMs:scene.closeElapsedMs,markerOnlyMs,overview:sample('overworld'),maximum:sample('overworld',maxScale),lower:sample('lower')};
},maxScale);
(async () => {
  const browser = await puppeteer.launch({executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser', headless: 'new', args: ['--no-sandbox', '--disable-dev-shm-usage']});
  try {
    const page = await browser.newPage(), errors = [], sheets = [], workers = [];
    page.on('pageerror', e => errors.push(e.message));
    page.on('workercreated', w => workers.push({url: w.url()}));
    page.on('response', r => { if (r.url().includes('/api/map-tiles/')) sheets.push({url: r.url(), status: r.status()}); });
    // Observe actual overlay raster draws without changing the production component.
    await page.evaluateOnNewDocument(() => {
      const draw = CanvasRenderingContext2D.prototype.drawImage;
      CanvasRenderingContext2D.prototype.drawImage = function (...args) {
        if (this.canvas.isConnected && this.canvas.closest('.stage') && args.length === 9 && args[0].width > 384) {
          const preview=window.__mapPreview,b=preview.landscapeModel(preview.snapshot.atlas).bounds;
          const sourceCell=args[0].width/(b.maxX-b.minX+1);
          window.__mapDraw = {tileStep: args[7] / args[3] * sourceCell, smoothing: this.imageSmoothingEnabled};
        }
        return draw.apply(this, args);
      };
    });
    await page.setViewport({width: 1920, height: 1080});
    await page.goto(process.env.WORLDMAP_PREVIEW_URL || 'http://127.0.0.1:8144/', {waitUntil: 'networkidle0'});
    await page.waitForFunction(() => window.__mapPreview?.ready);
    await settle(page);
    const manifest = await page.evaluate(() => window.__mapPreview.terrainSheet);
    assert.equal(manifest.tileSize, 48);
    const read = () => page.evaluate(() => ({...window.__mapDraw}));
    const save = kind => page.screenshot({path: path.join(out, `worldmap-p1g-${kind}-1920x1080.png`)});
    await page.click('[title="Fit world"]'); await settle(page);
    const overview = await read();
    await page.mouse.move(30, 30); await save('overview');
    const bounds = await page.$eval('canvas', e => {const r = e.getBoundingClientRect(); return {x: r.x, y: r.y, w: r.width, h: r.height};});
    const wheel = async (delta, count) => {
      await page.mouse.move(bounds.x + bounds.w / 2, bounds.y + bounds.h / 2);
      for (let i = 0; i < count; i++) await page.mouse.wheel({deltaY: delta});
      await settle(page);
    };
    await wheel(-100, 30);
    const maximum = await read(), oldMaximum = Math.min(110, overview.tileStep * 5);
    assert.ok(Math.abs(maximum.tileStep - Math.min(220, overview.tileStep * 10)) < .01, `wheel reaches the new scale/tile clamp: ${JSON.stringify({overview, maximum})}`);
    assert.ok(Math.abs(maximum.tileStep / oldMaximum - 2) < .01, 'closest view is twice the previous max');
    assert.equal(maximum.smoothing, false, 'enlarged landscape uses nearest-neighbor sampling');
    await wheel(-100, 5); assert.deepEqual(await read(), maximum, 'wheel cannot exceed maximum');
    const frame = async (area, kind) => {
      await page.click('[title="Fit world"]'); await settle(page); await wheel(-100, 30);
      const pan = await page.evaluate(({step, area}) => {
        const a = window.__mapPreview.snapshot.atlas, b = window.__mapPreview.landscapeModel(a).bounds;
        const places = window.__mapPreview.surfaceGroups(a.places).filter(p => p.area === area);
        const xs = places.map(p => p.x), ys = places.map(p => p.y);
        return {x: ((b.minX + b.maxX) / 2 - (Math.min(...xs) + Math.max(...xs)) / 2) * step,
          y: ((b.minY + b.maxY) / 2 - (Math.min(...ys) + Math.max(...ys)) / 2) * step};
      }, {step: maximum.tileStep, area});
      const sx = bounds.x + bounds.w * .55, sy = bounds.y + bounds.h * .65;
      await page.mouse.move(sx, sy); await page.mouse.down();
      await page.mouse.move(sx + pan.x, sy + pan.y, {steps: 12}); await page.mouse.up(); await settle(page);
      assert.deepEqual(await read(), maximum, 'pan preserves maximum zoom and nearest-neighbor sampling');
      await page.mouse.move(30, 30); await save(kind);
    };
    await frame('Z01_meadows_forest_path', 'maxzoom');
    await frame('Z02_oldtown', 'oldtown');
    await page.click('[title="Recenter on you"]'); await settle(page);
    await wheel(-100, 30);
    const recenterMaximum = await read();
    assert.ok(Math.abs(recenterMaximum.tileStep - 220) < .01, 'local recenter respects the enlarged tile cap');
    await page.click('[title="Fit world"]'); await settle(page);
    assert.deepEqual(await read(), overview, 'Fit world restores the original continent frame');
    await wheel(100, 30); const minimum = await read();
    assert.deepEqual(minimum, overview, 'minimum zoom retains the Fit world frame');
    // Exercise the same pointer handlers used by touch pinch; over/undershoot both bounds.
    const pinch = async (start, end) => {
      await page.evaluate(({start, end}) => {
        const canvas = document.querySelector('canvas'), r = canvas.getBoundingClientRect(), x = r.x + r.width / 2, y = r.y + r.height / 2;
        const send = (type, id, dx) => canvas.dispatchEvent(new PointerEvent(type, {bubbles: true, pointerType: 'touch', pointerId: id, clientX: x + dx, clientY: y}));
        send('pointerdown', 101, -start / 2); send('pointerdown', 102, start / 2);
        send('pointermove', 101, -end / 2); send('pointermove', 102, end / 2);
        send('pointerup', 101, -end / 2); send('pointerup', 102, end / 2);
      }, {start, end});
      await settle(page); return read();
    };
    const pinchMaximum = await pinch(20, 400); assert.deepEqual(pinchMaximum, maximum, 'pinch shares wheel max clamp');
    const pinchMinimum = await pinch(400, 10); assert.deepEqual(pinchMinimum, overview, 'pinch shares wheel min clamp');
    for (const worker of workers) worker.status = await page.evaluate(async url => (await fetch(url, {cache: 'force-cache'})).status, worker.url);
    assert.equal(errors.length, 0, 'no page errors');
    assert.ok(workers.length && workers.every(w => w.status === 200), 'worker loads');
    assert.ok(sheets.length && sheets.every(s => s.status === 200), 'terrain art loads');
    const performance = await measure(page, 10);
    const baselinePage = await browser.newPage();
    await baselinePage.setViewport({width:1920,height:1080});
    await baselinePage.goto(process.env.WORLDMAP_BASELINE_URL || 'http://127.0.0.1:8143/', {waitUntil:'networkidle0'});
    await baselinePage.waitForFunction(() => window.__mapPreview?.ready); await settle(baselinePage);
    const baselinePerformance = await measure(baselinePage, 5);
    const polish = await page.evaluate(() => {
      const a=window.__mapPreview.snapshot.atlas,c=document.createElement('canvas');c.width=1205;c.height=819;
      const paint=(atlas,layer)=>window.__mapPreview.paintAtlas(c.getContext('2d'),{w:1205,h:819,atlas,activeLayer:layer,visiblePlaces:atlas.places.filter(p=>p.layer===layer),visibleRegions:atlas.regions.filter(r=>r.layer===layer),currentRoomId:atlas.currentRoomId,userScale:10,frameWorld:true,maximized:true});
      const s=paint(a,'overworld').scene,l=paint(a,'lower').scene;
      const moved=paint({...a,currentRoomId:'R0104'},'overworld').scene;
      const missing=a.places.filter(p=>p.discovered&&p.layer==='overworld'&&p.mapRole!=='interior').filter(p=>s.coastMask.getContext('2d').getImageData((p.x-s.bounds.minX+.5)*32/s.coastScale,(p.y-s.bounds.minY+.5)*32/s.coastScale,1,1).data[3]<250).map(p=>p.id);
      return {cellSize:s.cellSize, closeCellSize:s.closeCellSize, lowerCellSize:l.cellSize, overviewRaster:{width:s.canvas.width,height:s.canvas.height}, closeRaster:{width:s.closeCanvas().width,height:s.closeCanvas().height}, markerReusesScene:moved===s, clippedRoomCenters:missing, bridges:s.bridges.length, lookout:s.stamps.filter(p=>p.roomId==='R0104'), features:s.stamps.reduce((o,p)=>(o[p.kind]=(o[p.kind]||0)+1,o),{})};
    });
    assert.equal(polish.cellSize,32); assert.equal(polish.closeCellSize,48); assert.equal(polish.lowerCellSize,48);
    assert.deepEqual(polish.clippedRoomCenters,[]); assert.ok(polish.markerReusesScene);
    assert.ok(polish.lookout.some(s=>s.kind==='tower') && polish.bridges>0);
    assert.ok(performance.heartbeats>2, 'worker bake leaves the main thread responsive');
    for(const timing of [performance.overview,performance.maximum,performance.lower])assert.ok(timing.p95Ms<16.7,'warm draw stays within one 60fps frame locally');
    console.log('Performance comparison:', JSON.stringify({performance,baselinePerformance}));
    const report = {manifest, performance, baselinePerformance, polish, source: 'Local read-only production-overlay preview; fonts, art and worker scenes settled before screenshots.', viewport: {width: 1920, height: 1080}, overview, maximum, oldMaximum, maxZoomRatio: maximum.tileStep / oldMaximum, recenterMaximum, minimum, pinchMaximum, pinchMinimum, workers, sheets, errors, screenshots: ['overview', 'maxzoom', 'oldtown'].map(n => `worldmap-p1g-${n}-1920x1080.png`)};
    fs.writeFileSync(path.join(out, 'worldmap-p1g-browser-checks.json'), JSON.stringify(report, null, 2) + '\n');
    console.log(JSON.stringify(report, null, 2));
  } finally { await browser.close(); }
})().catch(e => {console.error(e); process.exitCode = 1;});
