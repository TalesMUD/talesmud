import { TERRAIN_SHEET } from './terrainSheet.js';
import { landscapeModel, continentRaster } from './continentRenderer.js';
import { makeCanvas, sheetReady } from './mapArt.js';

export function sceneSignature(atlas) {
  return JSON.stringify([atlas.landscape, (atlas.places||[]).filter(p=>p.layer==='overworld').map(p=>[
    p.id,p.x,p.y,p.area,p.discovered,p.mapRole,p.surfaceRoomId,p.town,p.terrain,p.name,p.tags,p.entrances,p.mapFeatures,p.artSeed
  ]),atlas.paths]);
}

// One running job and one replaceable queued job keep exploration updates from
// accumulating stale work. Completed scenes are always matched by disclosure.
export class SceneQueue {
  constructor(worker, initialize, onReady, decodeCoast=coast=>{const canvas=makeCanvas(coast.width,coast.height);canvas.getContext('2d').putImageData(new ImageData(coast.pixels,coast.width,coast.height),0,0);return canvas}) {
    this.worker=worker;this.initialize=initialize;this.onReady=onReady;this.decodeCoast=decodeCoast;
    this.cache=new Map();this.byAtlas=new WeakMap();this.sequence=0;this.running=null;this.queued=null;this.failed=false;
    worker.onmessage=e=>this.receive(e.data);
    worker.onerror=()=>{this.failed=true;this.running=this.queued=null;worker.terminate();onReady()};
    worker.onmessageerror=worker.onerror;
  }
  pending() { return !!(this.running||this.queued); }
  get(atlas) {
    const start=performance.now();let state=this.byAtlas.get(atlas);
    if(state&&!state.cancelled)return state;
    const key=sceneSignature(atlas);
    state=this.cache.get(key);
    if(state&&!state.cancelled){this.byAtlas.set(atlas,state);return state}
    const model=landscapeModel(atlas);if(!model)return null;
    state={key,atlas,model,bounds:model.bounds,requestedAt:performance.now(),pending:true,buildings:[],glyphs:[],ridges:[],ambience:[],requestMs:0,deliveryMs:0};
    state.closeCanvas=()=>{
      if(!state.nearCanvas&&!state.closePending&&!state.pending){state.closePending=true;this.enqueue(state,true)}
      return state.nearCanvas||state.canvas;
    };
    this.cache.set(key,state);this.byAtlas.set(atlas,state);this.enqueue(state,false);
    state.requestMs=performance.now()-start;
    while(this.cache.size>4){const old=this.cache.keys().next().value;this.cache.delete(old)}
    return state;
  }
  enqueue(state,close) {
    const job={id:++this.sequence,state,close,requestedAt:performance.now()};
    if(this.running) {
      if(this.queued){this.queued.state.cancelled=!this.queued.close;this.queued.state.closePending=false}
      this.queued=job;
    } else this.send(job);
  }
  send(job) {
    this.running=job;
    this.initialize.then(()=>{
      if(this.failed)return;
      const start=performance.now();this.worker.postMessage({type:'scene',id:job.id,atlas:job.state.atlas,close:job.close});job.dispatchMs=performance.now()-start;
    }).catch(()=>this.worker.onerror());
  }
  receive(message) {
    const job=this.running;if(!job||job.id!==message.id)return;
    if(message.error){this.worker.onerror();return}
    const start=performance.now(),state=job.state;
    if(job.close){state.nearCanvas=message.bitmap;state.closeBakeMs=message.bakeMs;state.closePending=false}
    else {
      state.canvas=message.bitmap;state.pending=false;state.cancelled=false;
      for(const key of ['cellSize','closeCellSize','coastScale','bakeMs','buildings','glyphs','ridges','ambience','bridges','stamps','fortifications','coastDetails'])state[key]=message[key];
      state.coastMask=this.decodeCoast(message.coast);
    }
    if(job.close){state.closeDispatchMs=job.dispatchMs;state.closeDeliveryMs=performance.now()-start;state.closeElapsedMs=performance.now()-job.requestedAt}
    else {state.dispatchMs=job.dispatchMs;state.deliveryMs=performance.now()-start;state.elapsedMs=performance.now()-state.requestedAt}
    this.running=null;
    if(this.queued){const next=this.queued;this.queued=null;this.send(next)}
    this.onReady();
  }
}

let queue=null;
const listeners=new Set();
export function onMapScenesReady(fn){listeners.add(fn);return()=>listeners.delete(fn)}
export const mapScenesPending=()=>!!queue?.pending();
export async function waitForMapScenes(){while(mapScenesPending())await new Promise(r=>setTimeout(r,10))}
const notify=()=>{for(const fn of listeners)fn()};
export function worldmapScene(atlas,sheet) {
  if(!sheetReady(sheet)&&typeof Worker!=='undefined'&&typeof OffscreenCanvas!=='undefined'&&!sheet?.complete){const model=landscapeModel(atlas);return model?{pending:true,bounds:model.bounds,buildings:[],glyphs:[],ridges:[],ambience:[]}:null}
  if(!sheetReady(sheet)||typeof Worker==='undefined'||typeof OffscreenCanvas==='undefined'||typeof createImageBitmap==='undefined')return continentRaster(atlas,sheet);
  if(!queue) {
    try {
      const worker=new Worker(new URL(`worldmap-worker.js?v=fog1-${TERRAIN_SHEET.version}`,document.baseURI));
      const initialize=createImageBitmap(sheet).then(bitmap=>worker.postMessage({type:'sheet',sheet:bitmap},[bitmap]));
      queue=new SceneQueue(worker,initialize,notify);
    } catch {return continentRaster(atlas,sheet)}
  }
  return queue.failed?continentRaster(atlas,sheet):queue.get(atlas);
}
