import { TERRAIN_SHEET } from './terrainSheet.js';
import { surfaceGroups, outdoorRoads } from './surfaceAtlas.js';

export const LAND_CELL_SIZE = 32;
const rasterCache = new WeakMap();
const recentRasters = new Map(); // Two scenes: marker-only snapshots reuse pixels.
const blends = new Map(), masks = new Map();
const bayer = [0,8,2,10,12,4,14,6,3,11,1,9,15,7,13,5];
const directions = [[-1,0],[1,0],[0,-1],[0,1],[-1,-1],[1,-1],[-1,1],[1,1]];
const hash = text => { let n=0; for (const c of String(text)) n=(n*31+c.charCodeAt(0))|0; return n>>>0; };
const makeCanvas = (w,h) => { const c=document.createElement('canvas');c.width=w;c.height=h;return c; };
const sheetReady = sheet => sheet && sheet.complete && sheet.naturalWidth > 0;

export function landscapeModel(atlas) {
  const cells = atlas.landscape || [];
  const byCell = new Map(cells.map(c => [`${c.x}:${c.y}`,c]));
  const groups = surfaceGroups(atlas.places, 'overworld');
  if (!cells.length) return null;
  const xs=cells.map(c=>c.x),ys=cells.map(c=>c.y);
  return { cells, byCell, groups, roads:outdoorRoads(atlas.places,atlas.paths),
    bounds:{minX:Math.min(...xs),maxX:Math.max(...xs),minY:Math.min(...ys),maxY:Math.max(...ys)} };
}
function drawSprite(ctx,sheet,key,seed,x,y,size,decoration=false) {
  const row=(decoration?TERRAIN_SHEET.decorations:TERRAIN_SHEET.rows)?.[key];
  if (row == null) return;
  if (sheetReady(sheet)) {
    ctx.imageSmoothingEnabled=false;
    ctx.drawImage(sheet,(hash(seed)%3)*32,row*32,32,32,x,y,size,size);
  } else if (!decoration) {
    ctx.fillStyle=TERRAIN_SHEET.colors?.[key] || '#649c48';ctx.fillRect(x,y,size,size);
  }
}
function edgeMask(dx,dy,width=10,strong=false) {
  const key=`${dx}:${dy}:${width}:${strong}`;
  if (masks.has(key)) return masks.get(key);
  const canvas=makeCanvas(32,32),ctx=canvas.getContext('2d'),im=ctx.createImageData(32,32);
  for(let y=0;y<32;y++)for(let x=0;x<32;x++){
    const xx=dx<0?x:31-x,yy=dy<0?y:31-y;
    const d=dx&&dy?Math.max(xx,yy):dx?xx:yy;
    const alpha=Math.max(0,(strong?1:.55)*(1-d/width));
    if ((bayer[(y%4)*4+x%4]+.5)/16<=alpha) im.data[(y*32+x)*4+3]=255;
  }
  ctx.putImageData(im,0,0);masks.set(key,canvas);return canvas;
}
function blendSprite(sheet,key,variant,dx,dy,strong=false,width=10) {
  const id=`${key}:${variant}:${dx}:${dy}:${strong}:${width}:${sheetReady(sheet)}`;
  if(blends.has(id))return blends.get(id);
  const canvas=makeCanvas(32,32),ctx=canvas.getContext('2d');
  if(sheetReady(sheet))ctx.drawImage(sheet,variant*32,TERRAIN_SHEET.rows[key]*32,32,32,0,0,32,32);
  else{ctx.fillStyle=TERRAIN_SHEET.colors[key];ctx.fillRect(0,0,32,32)}
  ctx.globalCompositeOperation='destination-in';ctx.drawImage(edgeMask(dx,dy,width,strong),0,0);
  blends.set(id,canvas);return canvas;
}
function coastMask(empty) {
  const key='coast:'+empty.join('');if(masks.has(key))return masks.get(key);
  const canvas=makeCanvas(32,32),ctx=canvas.getContext('2d'),im=ctx.createImageData(32,32);
  for(let y=0;y<32;y++)for(let x=0;x<32;x++){
    let alpha=1;
    const distances=[x,31-x,y,31-y];
    for(let i=0;i<4;i++)if(empty[i])alpha=Math.min(alpha,(distances[i]+.5)/3);
    // Round headlands at pairs of sea-facing edges, with a dithered fringe.
    for(const [a,b,cx,cy] of [[0,2,9,9],[1,2,22,9],[0,3,9,22],[1,3,22,22]]){
      if(!empty[a]||!empty[b])continue;
      if(distances[a]<9&&distances[b]<9)alpha=Math.min(alpha,(10-Math.hypot(x-cx,y-cy))/2);
    }
    if((bayer[(y%4)*4+x%4]+.5)/16<=alpha)im.data[(y*32+x)*4+3]=255;
  }
  ctx.putImageData(im,0,0);masks.set(key,canvas);return canvas;
}

// Cache the whole landscape at native sprite resolution. Camera changes only
// blit this bitmap; ordered-dither edges, coast masks, towns and roads are baked.
export function continentRaster(atlas,sheet) {
  const ready=!!sheetReady(sheet),prev=rasterCache.get(atlas);
  if(prev && prev.ready===ready)return prev;
  const model=landscapeModel(atlas);if(!model)return null;
  const {bounds,byCell,cells,groups,roads}=model;
  // Ignore current-room/selection flags, but include every disclosure and art
  // input. Never reuse a charted scene for another character's fogged scene.
  const signature=JSON.stringify([ready,cells,groups.map(g=>[
    g.id,g.x,g.y,g.town,g.area,g.discovered,g.tags,
    g.members.map(p=>[p.id,p.discovered,p.mapRole,p.terrain,p.name,p.tags,p.entrances])
  ]),roads.map(({a,b})=>[a.id,b.id,a.x,a.y,b.x,b.y])]);
  if(recentRasters.has(signature)) {
    const cached=recentRasters.get(signature);
    recentRasters.delete(signature);recentRasters.set(signature,cached);
    rasterCache.set(atlas,cached);return cached;
  }
  const canvas=makeCanvas((bounds.maxX-bounds.minX+1)*32,(bounds.maxY-bounds.minY+1)*32);
  const ctx=canvas.getContext('2d');ctx.imageSmoothingEnabled=false;
  const tileCanvas=makeCanvas(32,32),tc=tileCanvas.getContext('2d');
  const native=(x,y)=>({x:(x-bounds.minX+.5)*32,y:(y-bounds.minY+.5)*32});
  for(const cell of cells){
    const key=cell.terrain;
    const neighbours=directions.map(([dx,dy])=>byCell.get(`${cell.x+dx}:${cell.y+dy}`));
    const empty=neighbours.slice(0,4).map(p=>!p);
    const targetX=(cell.x-bounds.minX)*32,targetY=(cell.y-bounds.minY)*32;
    // Most cells are uniform ground: one sheet blit, no intermediate mask.
    if(!empty.some(Boolean) && (key==='fog' || neighbours.every(p=>!p||p.terrain===key))) {
      drawSprite(ctx,sheet,key,`${cell.x}:${cell.y}`,targetX,targetY,32);continue;
    }
    tc.clearRect(0,0,32,32);
    drawSprite(tc,sheet,key,`${cell.x}:${cell.y}`,0,0,32);
    for(let n=0;n<directions.length;n++){
      const [dx,dy]=directions[n],neighbour=neighbours[n];
      if(key==='fog')continue; // Never transfer revealed art into a fog cell.
      if(neighbour && neighbour.terrain!==key){
        const other=neighbour.terrain;
        tc.drawImage(blendSprite(sheet,other,hash(`${cell.x+dx}:${cell.y+dy}`)%3,dx,dy),0,0);
      }else if(!neighbour && !(dx&&dy)){
        tc.drawImage(blendSprite(sheet,'shore',0,dx,dy,true,9),0,0);
        tc.drawImage(blendSprite(sheet,'water',1,dx,dy,true,3),0,0);
      }
    }
    if(empty.some(Boolean)){
      tc.globalCompositeOperation='destination-in';tc.drawImage(coastMask(empty),0,0);tc.globalCompositeOperation='source-over';
    }
    ctx.drawImage(tileCanvas,targetX,targetY);
  }
  // Town districts: continuous paving and a shared wall, rather than floors
  // for every shop/tavern. Village anchors get a smaller open hamlet.
  const townAreas=new Map();
  for(const group of groups){if(!group.town || !group.discovered)continue;if(!townAreas.has(group.area))townAreas.set(group.area,[]);townAreas.get(group.area).push(group)}
  for(const town of townAreas.values()){
    const xs=town.map(p=>p.x),ys=town.map(p=>p.y);
    const minX=Math.min(...xs),maxX=Math.max(...xs),minY=Math.min(...ys),maxY=Math.max(...ys);
    const {x,y}=native((minX+maxX)/2,(minY+maxY)/2);
    const rx=(maxX-minX+2.4)*16,ry=(maxY-minY+2.4)*16;
    ctx.save();ctx.beginPath();ctx.ellipse(x,y,rx,ry,0,0,Math.PI*2);ctx.clip();
    for(let yy=Math.floor(minY-1);yy<=maxY+1;yy++)for(let xx=Math.floor(minX-1);xx<=maxX+1;xx++){
      const c=byCell.get(`${xx}:${yy}`);if(!c||c.terrain==='fog')continue;
      drawSprite(ctx,sheet,'city',`${xx}:${yy}`, (xx-bounds.minX)*32,(yy-bounds.minY)*32,32);
    }
    ctx.restore();
    if(town.length>3){
      ctx.beginPath();ctx.ellipse(x,y,rx-1,ry-1,0,0,Math.PI*2);ctx.strokeStyle='#51665d';ctx.lineWidth=4;ctx.stroke();ctx.strokeStyle='#c3c5aa';ctx.lineWidth=2;ctx.stroke();
      // Break walls at the actual gates so dirt roads can reach the streets.
      for(const gate of town.filter(p=>p.tags?.includes('gate'))){const g=native(gate.x,gate.y);ctx.fillStyle='#b5a78b';ctx.fillRect(g.x-4,g.y-5,8,10)}
    }
  }
  // Dirt roads follow actual, charted outdoor exits. They add no topology.
  for(const {a,b} of roads){
    const from=native(a.x,a.y),to=native(b.x,b.y);
    ctx.beginPath();ctx.moveTo(from.x,from.y);
    const dx=to.x-from.x,dy=to.y-from.y;
    ctx.quadraticCurveTo((from.x+to.x)/2-dy*.06,(from.y+to.y)/2+dx*.06,to.x,to.y);
    ctx.lineCap='round';ctx.strokeStyle='#735f40';ctx.lineWidth=5;ctx.stroke();ctx.strokeStyle='#c7ad75';ctx.lineWidth=3;ctx.stroke();
  }
  const buildings=[];
  for(const group of groups){
    const known=group.members.filter(p=>p.discovered);
    if(!group.discovered&&!known.length)continue;
    const at=native(group.x,group.y),interiors=known.filter(p=>p.mapRole==='interior');
    const keep=known.some(p=>p.terrain==='castle' || /\b(keep|castle|fortress|citadel)\b/i.test([p.name,...(p.tags||[])].join(' ')));
    const roofCount=Math.min(5,Math.max(interiors.length,group.town?3:0));
    for(let i=0;i<roofCount;i++){
      const offsets=[[-.65,-.6],[.7,-.5],[-.7,.55],[.65,.7],[0,-1.1]];
      const [ox,oy]=offsets[i],size=keep&&i===0?37:27;
      const type=keep&&i===0?'keep':'roof';const x=at.x+ox*32,y=at.y+oy*32;
      drawSprite(ctx,sheet,type,`${group.id}:${i}`,x-size/2,y-size/2,size,true);
      buildings.push({surfaceId:group.id,x:group.x+ox,y:group.y+oy,half:size/64});
    }
    if(keep&&!roofCount){drawSprite(ctx,sheet,'keep',group.id,at.x-20,at.y-22,40,true)}
    const entrances=known.flatMap(p=>p.entrances||[]);
    if(entrances.length){drawSprite(ctx,sheet,'cave',group.id,at.x-13,at.y-14,26,true)}
  }
  const result={canvas,bounds,buildings,ready,model};
  rasterCache.set(atlas,result);recentRasters.set(signature,result);
  if(recentRasters.size>2)recentRasters.delete(recentRasters.keys().next().value);
  return result;
}
