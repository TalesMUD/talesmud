import { TERRAIN_SHEET } from './terrainSheet.js';
import { surfaceGroups, outdoorRoads } from './surfaceAtlas.js';
import { smoothCoast } from './coastline.js';
import { hash, makeCanvas, sheetReady, drawMapSprite, knownFeatures, roofKind, roadWaterCrossings, roundedHull, hullPath } from './mapArt.js';
export const LAND_CELL_SIZE = 32;
const rasterCache = new WeakMap(), recentRasters = new Map();
const blends = new Map(), masks = new Map(), composedTiles = new Map();
const bayer = [0,8,2,10,12,4,14,6,3,11,1,9,15,7,13,5];
const directions = [[-1,0],[1,0],[0,-1],[0,1],[-1,-1],[1,-1],[-1,1],[1,1]];
const drawSprite=(ctx,sheet,key,seed,x,y,size,decoration=false)=>drawMapSprite(ctx,sheet,key,seed,x,y,size,decoration?'decorations':'rows');

export function landscapeModel(atlas) {
  const cells = atlas.landscape || [];
  if (!cells.length) return null;
  const byCell = new Map(cells.map(c => [`${c.x}:${c.y}`,c]));
  const groups = surfaceGroups(atlas.places, 'overworld');
  const xs=cells.map(c=>c.x),ys=cells.map(c=>c.y);
  return {cells,byCell,groups,roads:outdoorRoads(atlas.places,atlas.paths),
    bounds:{minX:Math.min(...xs)-2,maxX:Math.max(...xs)+2,minY:Math.min(...ys)-2,maxY:Math.max(...ys)+2}};
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

function paintTown(ctx,sheet,town,roads,native,bounds) {
  const points=town.map(p=>native(p.x,p.y));
  const xs=points.map(p=>p.x),ys=points.map(p=>p.y);
  const x=Math.floor(Math.min(...xs)-32),y=Math.floor(Math.min(...ys)-32);
  const w=Math.ceil(Math.max(...xs)-x+32),h=Math.ceil(Math.max(...ys)-y+32);
  const pavement=makeCanvas(w,h),pc=pavement.getContext('2d');
  const pavingMask=makeCanvas(w,h),mc=pavingMask.getContext('2d');
  mc.strokeStyle=mc.fillStyle='#fff';mc.lineCap=mc.lineJoin='round';mc.lineWidth=19;
  for(const {a,b} of roads.filter(({a,b})=>a.area===town[0].area&&b.area===town[0].area)) {
    const p=native(a.x,a.y),q=native(b.x,b.y);mc.beginPath();mc.moveTo(p.x-x,p.y-y);mc.lineTo(q.x-x,q.y-y);mc.stroke();
  }
  for(const p of points){mc.beginPath();mc.arc(p.x-x,p.y-y,13,0,Math.PI*2);mc.fill()}
  for(let yy=0;yy<h;yy+=32)for(let xx=0;xx<w;xx+=32)drawSprite(pc,sheet,'city',`${town[0].area}:${xx}:${yy}`,xx,yy,32);
  pc.globalCompositeOperation='destination-in';pc.filter='blur(3px)';pc.drawImage(pavingMask,0,0);pc.filter='none';
  ctx.globalAlpha=.82;ctx.drawImage(pavement,x,y);ctx.globalAlpha=1;
  if(town.length>3) {
    const hull=roundedHull(points,27);hullPath(ctx,hull);
    ctx.shadowColor='rgba(18,36,31,.55)';ctx.shadowBlur=3;ctx.shadowOffsetY=3;
    ctx.strokeStyle='#536b65';ctx.lineWidth=6;ctx.stroke();ctx.shadowBlur=ctx.shadowOffsetY=0;
    ctx.strokeStyle='#b5bba0';ctx.lineWidth=3;ctx.stroke();
    ctx.setLineDash([3,5]);ctx.strokeStyle='#d6cfac';ctx.lineWidth=1;ctx.stroke();ctx.setLineDash([]);
  }
}

// Native scene cache. Camera/marker changes reuse the bitmap; the signature
// includes all disclosed room art and ground so fog never borrows known art.
export function continentRaster(atlas,sheet) {
  const ready=!!sheetReady(sheet),prev=rasterCache.get(atlas);
  if(prev&&prev.ready===ready)return prev;
  const bakeStart=performance.now();
  const model=landscapeModel(atlas);if(!model)return null;
  const {bounds,byCell,cells,groups,roads}=model;
  const signature=JSON.stringify([ready,cells,groups.map(g=>[
    g.id,g.x,g.y,g.town,g.area,g.discovered,g.tags,
    g.members.map(p=>[p.id,p.discovered,p.mapRole,p.terrain,p.name,p.tags,p.entrances,p.mapFeatures,p.artSeed])
  ]),roads.map(({a,b})=>[a.id,b.id,a.x,a.y,b.x,b.y])]);
  if(recentRasters.has(signature)) {
    const cached=recentRasters.get(signature);recentRasters.delete(signature);recentRasters.set(signature,cached);
    rasterCache.set(atlas,cached);return cached;
  }
  const w=(bounds.maxX-bounds.minX+1)*32,h=(bounds.maxY-bounds.minY+1)*32;
  const canvas=makeCanvas(w,h),sc=canvas.getContext('2d'),overlay=makeCanvas(w,h),ctx=overlay.getContext('2d');
  const ground=makeCanvas(w,h),gc=ground.getContext('2d');
  const tileCanvas=makeCanvas(32,32),tc=tileCanvas.getContext('2d');
  const native=(x,y)=>({x:(x-bounds.minX+.5)*32,y:(y-bounds.minY+.5)*32});
  const coast=smoothCoast(model),ridges=[],ambience=[];

  const roomSeeds=new Map(groups.filter(g=>g.discovered).map(g=>[`${g.x}:${g.y}`,g.artSeed]));
  const areaPoints=new Map();
  for(const g of groups.filter(g=>g.discovered)) {
    const a=areaPoints.get(g.area)||{area:g.area,x:0,y:0,n:0};a.x+=g.x;a.y+=g.y;a.n++;areaPoints.set(g.area,a);
  }
  const areas=[...areaPoints.values()].map(a=>({...a,x:a.x/a.n,y:a.y/a.n}));

  for(const cell of cells) {
    const key=cell.terrain,seed=roomSeeds.get(`${cell.x}:${cell.y}`)||`${key}:${cell.x}:${cell.y}`,targetX=(cell.x-bounds.minX)*32,targetY=(cell.y-bounds.minY)*32;
    const neighbours=directions.map(([dx,dy])=>byCell.get(`${cell.x+dx}:${cell.y+dy}`));
    if(key==='fog'||neighbours.every(p=>!p||p.terrain===key)) {
      drawSprite(gc,sheet,key,seed,targetX,targetY,32);
    } else {
      const tileKey=JSON.stringify([ready,key,hash(seed)%TERRAIN_SHEET.variants,
        neighbours.map((n,i)=>n?[n.terrain,hash(`${n.terrain}:${cell.x+directions[i][0]}:${cell.y+directions[i][1]}`)%TERRAIN_SHEET.variants]:null)]);
      let tile=composedTiles.get(tileKey);
      if(!tile) {
        tc.clearRect(0,0,32,32);drawSprite(tc,sheet,key,seed,0,0,32);
        for(let n=0;n<directions.length;n++) {
          const neighbour=neighbours[n];if(!neighbour||neighbour.terrain===key)continue;
          const [dx,dy]=directions[n];
          tc.drawImage(blendSprite(sheet,neighbour.terrain,hash(`${neighbour.terrain}:${cell.x+dx}:${cell.y+dy}`)%TERRAIN_SHEET.variants,dx,dy),0,0);
        }
        tile=makeCanvas(32,32);tile.getContext('2d').drawImage(tileCanvas,0,0);
        composedTiles.set(tileKey,tile);if(composedTiles.size>1024)composedTiles.delete(composedTiles.keys().next().value);
      }
      gc.drawImage(tile,targetX,targetY);
    }
    if(key!=='fog') {
      // A faint regional tint concentrates at the boundary between the two
      // nearest disclosed zone centers, with no invented names or boundaries.
      const distances=areas.map(a=>({area:a.area,d:Math.hypot(cell.x-a.x,cell.y-a.y)})).sort((a,b)=>a.d-b.d);
      const edge=distances.length>1?Math.max(0,1-(distances[1].d-distances[0].d)/3):0;
      if(edge) {
        const color=hash(distances[0].area)%2?'198,167,86':'81,132,140';
        gc.fillStyle=`rgba(${color},${edge*.045})`;gc.fillRect(targetX,targetY,32,32);
      }
    }
    if(['mountain','snow'].includes(key)&&cell.x%4===0&&cell.y%3===0) {

      ridges.push({x:cell.x+(hash(`${cell.x}:${cell.y}`)%5-2)*.15,y:cell.y+Math.sin(cell.x*.45)*.4,kind:'ridge'});
    }
    if(key==='forest'&&cell.x%2===0&&cell.y%2===0) {

      ridges.push({x:cell.x,y:cell.y,kind:'canopy'});
    }
    if(key==='water')ambience.push({x:cell.x,y:cell.y,kind:'water'});
  }
  // Continue terrain into contour-grown corner pixels. The mask trims this
  // extension; without it, concave beaches inherit hard raw-cell corners.
  const fringe=new Map();
  for(const c of cells)for(const [dx,dy] of directions) {
    const x=c.x+dx,y=c.y+dy,key=`${x}:${y}`;
    if(!byCell.has(key)&&!fringe.has(key))fringe.set(key,{x,y,terrain:c.terrain});
  }
  for(const c of fringe.values())drawSprite(gc,sheet,c.terrain,`fringe:${c.x}:${c.y}`,(c.x-bounds.minX)*32,(c.y-bounds.minY)*32,32);
  const detail=makeCanvas(w,h),dc=detail.getContext('2d');
  const sprite=makeCanvas(100,100),spriteContext=sprite.getContext('2d');
  const detailStamp=(target,kind,seed,x,y,size)=>{
    // Interior stamps need no full-scene mask blit. Only coastal/fog-border
    // stamps use a small local mask, avoiding several multi-megapixel copies.
    let edge=false;
    for(let gx=Math.floor(x/32)+bounds.minX;gx<=Math.floor((x+size)/32)+bounds.minX;gx++) {
      for(let gy=Math.floor(y/32)+bounds.minY;gy<=Math.floor((y+size)/32)+bounds.minY;gy++) {
        const cell=byCell.get(`${gx}:${gy}`);if(!cell||cell.terrain==='fog')edge=true;
      }
    }
    if(!edge){drawSprite(target,sheet,kind,seed,x,y,size,true);return}
    spriteContext.globalCompositeOperation='source-over';spriteContext.clearRect(0,0,100,100);
    drawSprite(spriteContext,sheet,kind,seed,0,0,size,true);
    spriteContext.globalCompositeOperation='destination-in';spriteContext.drawImage(coast.mask,x,y,size,size,0,0,size,size);
    spriteContext.globalCompositeOperation='source-over';target.drawImage(sprite,0,0,size,size,x,y,size,size);
  };
  const clearFog=target=>{for(const c of cells)if(c.terrain==='fog')target.clearRect((c.x-bounds.minX)*32,(c.y-bounds.minY)*32,32,32)};
  for(const ridge of ridges) {
    const at=native(ridge.x,ridge.y),size=ridge.kind==='ridge'?96:66;
    detailStamp(dc,ridge.kind,`${ridge.kind}:${ridge.x}:${ridge.y}`,at.x-size/2,at.y-size*.57,size);
  }
  clearFog(dc);
  const closeDetail=makeCanvas(w,h),nc=closeDetail.getContext('2d');
  for(const c of cells.filter(c=>c.terrain==='forest')) {
    const at=native(c.x,c.y),seed=roomSeeds.get(`${c.x}:${c.y}`)||`canopy:${c.x}:${c.y}`;
    detailStamp(nc,'canopy',seed,at.x-17,at.y-20,34);
  }
  for(const ridge of ridges.filter(r=>r.kind==='ridge')) {
    const at=native(ridge.x,ridge.y);detailStamp(nc,'ridge',`${ridge.x}:${ridge.y}`,at.x-34,at.y-40,68);
  }
  clearFog(nc);
  // Clip all surface art to the shared smooth contour, then place its depth
  // bands underneath. Rock/canopy shadows are authored into transparent stamps.
  gc.globalCompositeOperation='destination-in';gc.drawImage(coast.mask,0,0);gc.globalCompositeOperation='source-over';
  sc.drawImage(coast.bands,0,0);sc.drawImage(ground,0,0);
  const nearCanvas=makeCanvas(w,h),nearContext=nearCanvas.getContext('2d');nearContext.drawImage(canvas,0,0);
  sc.drawImage(detail,0,0);nearContext.drawImage(closeDetail,0,0);
  const townAreas=new Map(),glyphs=[];
  for(const group of groups) {
    if(!group.town||!group.discovered)continue;
    if(!townAreas.has(group.area))townAreas.set(group.area,[]);townAreas.get(group.area).push(group);
  }
  for(const town of townAreas.values()) {
    paintTown(ctx,sheet,town,roads,native,bounds);
    const keep=town.some(g=>knownFeatures(g).includes('keep')||g.members.some(p=>p.discovered&&p.terrain==='castle'));
    glyphs.push({x:town.reduce((n,g)=>n+g.x,0)/town.length,y:town.reduce((n,g)=>n+g.y,0)/town.length,roomId:town[0].id,kind:keep?'keep':town.length>3?'town':'village'});
  }
  for(const {a,b} of roads) {
    const from=native(a.x,a.y),to=native(b.x,b.y),dx=to.x-from.x,dy=to.y-from.y;
    ctx.beginPath();ctx.moveTo(from.x,from.y);ctx.quadraticCurveTo((from.x+to.x)/2-dy*.06,(from.y+to.y)/2+dx*.06,to.x,to.y);
    ctx.lineCap='round';ctx.strokeStyle='rgba(39,57,35,.28)';ctx.lineWidth=7;ctx.stroke();
    ctx.strokeStyle='#806c45';ctx.lineWidth=5;ctx.stroke();ctx.strokeStyle='#c7ad75';ctx.lineWidth=3;ctx.stroke();
  }
  const bridges=roadWaterCrossings(roads,byCell);
  for(const bridge of bridges) {
    const at=native(bridge.x,bridge.y);ctx.save();ctx.translate(at.x,at.y);ctx.rotate(bridge.angle);
    drawSprite(ctx,sheet,'bridge',`${bridge.from}:${bridge.to}`,-20,-20,40,true);ctx.restore();
  }
  const buildings=[],stamps=[];
  const placeStamp=(group,kind,x,y,size,seed)=>{
    const at=native(x,y);drawSprite(ctx,sheet,kind,seed,at.x-size/2,at.y-size*.58,size,true);
    buildings.push({surfaceId:group.id,x,y,half:size/64});stamps.push({roomId:group.id,kind,x,y});
    if(kind==='forge')ambience.push({x:x+.23,y:y-.44,kind:'smoke'});
  };
  for(const group of groups) {
    const known=group.members.filter(p=>p.discovered);if(!known.length)continue;
    const interiors=known.filter(p=>p.mapRole==='interior'),features=knownFeatures(group);
    const count=Math.min(5,Math.max(interiors.length,group.town?2:0));
    const offsets=[[-.63,-.48],[.65,-.32],[-.64,.57],[.64,.67],[0,-1.05]];
    for(let i=0;i<count;i++) {
      const [ox,oy]=offsets[i],room=interiors[i]||known[0],kind=roofKind(room),size=kind==='keep'?41:29;
      placeStamp(group,kind,group.x+ox,group.y+oy,size,room.artSeed||`${group.id}:${i}`);
    }
    if(features.includes('keep')&&!interiors.some(p=>roofKind(p)==='keep'))placeStamp(group,'keep',group.x-.12,group.y-.55,43,group.id);
    if(features.includes('tower')&&group.mapRole!=='interior')placeStamp(group,group.town?'gatehouse':'tower',group.x,group.y-.45,34,group.id);
    for(const kind of ['dock','ruins','graveyard','magic','farm','reeds','stump','flowers']) {
      if(features.includes(kind)&&(!group.town||!['reeds','flowers'].includes(kind)))placeStamp(group,kind,group.x+(count ? .8 : 0),group.y+.32,kind==='dock'?37:30,`${group.id}:${kind}`);
    }
    if(known.some(p=>(p.entrances||[]).length))placeStamp(group,features.includes('mine')?'mine':'cave',group.x,group.y,29,group.id);
  }
  sc.drawImage(overlay,0,0);nearContext.drawImage(overlay,0,0);
  const result={canvas,nearCanvas,bounds,buildings,ready,model,glyphs,ridges,ambience,bridges,stamps,coastMask:coast.mask,bakeMs:performance.now()-bakeStart};
  rasterCache.set(atlas,result);recentRasters.set(signature,result);
  if(recentRasters.size>2)recentRasters.delete(recentRasters.keys().next().value);
  return result;
}
