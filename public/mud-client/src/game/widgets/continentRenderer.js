import { TERRAIN_SHEET } from './terrainSheet.js';
import { surfaceGroups, outdoorRoads } from './surfaceAtlas.js';
import { smoothCoast } from './coastline.js';
import { hash, makeCanvas, freezeCanvas, sheetReady, drawMapSprite, knownFeatures, roofKind, roadWaterCrossings } from './mapArt.js';
import { townFortifications, mountainDepth, reliefAt, forestSpecies } from './mapDetails.js';
export const LAND_CELL_SIZE = 32;
const rasterCache = new WeakMap(), recentRasters = new Map(), townLayers = new Map();
const directions = [[-1,0],[1,0],[0,-1],[0,1],[-1,-1],[1,-1],[-1,1],[1,1]];
const sprite=(ctx,sheet,key,seed,x,y,size)=>drawMapSprite(ctx,sheet,key,seed,x,y,size);

export function landscapeModel(atlas) {
  const cells=atlas.landscape||[];if(!cells.length)return null;
  const byCell=new Map(cells.map(c=>[`${c.x}:${c.y}`,c]));
  const groups=surfaceGroups(atlas.places,'overworld');
  const xs=cells.map(c=>c.x),ys=cells.map(c=>c.y);
  return {cells,byCell,groups,roads:outdoorRoads(atlas.places,atlas.paths),
    bounds:{minX:Math.min(...xs)-2,maxX:Math.max(...xs)+2,minY:Math.min(...ys)-2,maxY:Math.max(...ys)+2}};
}
function paintTown(ctx,sheet,town,roads,native) {
  const points=town.map(p=>native(p.x,p.y)),ids=new Set(town.map(p=>p.id));
  const x=Math.floor(Math.min(...points.map(p=>p.x))-32),y=Math.floor(Math.min(...points.map(p=>p.y))-32);
  const w=Math.ceil(Math.max(...points.map(p=>p.x))-x+32),h=Math.ceil(Math.max(...points.map(p=>p.y))-y+32);
  const pavement=makeCanvas(w,h),pc=pavement.getContext('2d');
  const mask=makeCanvas(w,h),mc=mask.getContext('2d');
  mc.strokeStyle=mc.fillStyle='#fff';mc.lineCap=mc.lineJoin='round';mc.lineWidth=19;
  for(const {a,b} of roads.filter(({a,b})=>ids.has(a.id)&&ids.has(b.id))) {
    const p=native(a.x,a.y),q=native(b.x,b.y);mc.beginPath();mc.moveTo(p.x-x,p.y-y);mc.lineTo(q.x-x,q.y-y);mc.stroke();
  }
  for(const p of points){mc.beginPath();mc.arc(p.x-x,p.y-y,13,0,Math.PI*2);mc.fill()}
  for(let yy=0;yy<h;yy+=32)for(let xx=0;xx<w;xx+=32)drawMapSprite(pc,sheet,'city',`${town[0].area}:${xx}:${yy}`,xx,yy,32,'rows');
  pc.globalCompositeOperation='destination-in';pc.filter='blur(3px)';pc.drawImage(mask,0,0);pc.filter='none';
  ctx.drawImage(pavement,x,y);
  const walls=town.length>3?townFortifications(town,roads):{loops:[],towers:[],gates:[]};
  for(const loop of walls.loops) {
    ctx.beginPath();loop.forEach((p,i)=>{const at=native(p.x,p.y);if(i)ctx.lineTo(at.x,at.y);else ctx.moveTo(at.x,at.y)});ctx.closePath();
    ctx.strokeStyle='#536b65';ctx.lineWidth=6;ctx.stroke();
    ctx.strokeStyle='#b5bba0';ctx.lineWidth=3;ctx.stroke();
    ctx.setLineDash([3,5]);ctx.strokeStyle='#d6cfac';ctx.lineWidth=1;ctx.stroke();ctx.setLineDash([]);
  }
  return walls;
}

function paintRoadNetwork(ctx,roads,native,coastPath) {
  ctx.save();ctx.clip(coastPath);ctx.beginPath();
  const seen=new Set();
  for(const {a,b} of roads) {
    const key=[a.id,b.id].sort().join('|');if(seen.has(key))continue;seen.add(key);
    const from=native(a.x,a.y),to=native(b.x,b.y),dx=to.x-from.x,dy=to.y-from.y;
    ctx.moveTo(from.x,from.y);
    ctx.quadraticCurveTo((from.x+to.x)/2-dy*.06,(from.y+to.y)/2+dx*.06,to.x,to.y);
  }
  ctx.lineCap='round';ctx.lineJoin='round';
  // Each pass strokes the full compound network once. Shared exit junctions
  // therefore join cleanly without darkening from repeated per-road strokes.
  ctx.strokeStyle='rgba(37,48,34,.22)';ctx.lineWidth=3;ctx.stroke();
  ctx.strokeStyle='rgba(128,108,69,.72)';ctx.lineWidth=1.7;ctx.stroke();
  ctx.strokeStyle='rgba(199,173,117,.28)';ctx.lineWidth=.55;ctx.stroke();
  ctx.restore();
}

// Ground and town art are shared by the two LOD scenes. Detailed trees are
// baked only on first close zoom; pan, zoom and player movement reuse pixels.
export function continentRaster(atlas,sheet) {
  const ready=!!sheetReady(sheet),prev=rasterCache.get(atlas);if(prev&&prev.ready===ready)return prev;
  const bakeStart=performance.now();
  const model=landscapeModel(atlas);if(!model)return null;
  const {bounds,byCell,cells,groups,roads}=model;
  const artInputs=[ready,groups.map(g=>[
    g.id,g.x,g.y,g.town,g.area,g.discovered,g.tags,
    g.members.map(p=>[p.id,p.discovered,p.mapRole,p.terrain,p.name,p.tags,p.entrances,p.mapFeatures,p.artSeed])
  ]),roads.map(({a,b})=>[a.id,b.id,a.x,a.y,b.x,b.y])];
  const signature=JSON.stringify([cells,artInputs]);
  if(recentRasters.has(signature)) {
    const r=recentRasters.get(signature);recentRasters.delete(signature);recentRasters.set(signature,r);rasterCache.set(atlas,r);return r;
  }
  const w=(bounds.maxX-bounds.minX+1)*32,h=(bounds.maxY-bounds.minY+1)*32;
  const canvas=makeCanvas(w,h),sc=canvas.getContext('2d');
  const native=(x,y)=>({x:(x-bounds.minX+.5)*32,y:(y-bounds.minY+.5)*32});
  const coast=smoothCoast(model),ridges=[],ambience=[],coastDetails=[],tints=[];
  const depth=mountainDepth(cells,byCell),roomSeeds=new Map(groups.filter(g=>g.discovered).map(g=>[`${g.x}:${g.y}`,g.artSeed]));
  const areaPoints=new Map();
  for(const g of groups.filter(g=>g.discovered)) {
    const a=areaPoints.get(g.area)||{area:g.area,x:0,y:0,n:0};a.x+=g.x;a.y+=g.y;a.n++;areaPoints.set(g.area,a);
  }
  const areas=[...areaPoints.values()].map(a=>({...a,x:a.x/a.n,y:a.y/a.n}));
  const nearestArea=c=>areas.reduce((best,a)=>!best||Math.hypot(c.x-a.x,c.y-a.y)<Math.hypot(c.x-best.x,c.y-best.y)?a:best,null)?.area||'';
  sc.imageSmoothingEnabled=true;sc.drawImage(coast.bands,0,0,w,h);
  const paintGround=(c,blend=true,output=sc)=>{
    if(coast.tileKinds.get(`${c.x}:${c.y}`)==='empty')return;
    const x=(c.x-bounds.minX)*32,y=(c.y-bounds.minY)*32;
    drawMapSprite(output,sheet,c.terrain,roomSeeds.get(`${c.x}:${c.y}`)||`${c.terrain}:${c.x}:${c.y}`,x,y,32,'rows');
    if(blend&&c.terrain!=='fog'&&ready)for(const [dx,dy] of directions) {
      const n=byCell.get(`${c.x+dx}:${c.y+dy}`);if(!n||n.terrain===c.terrain)continue;
      const row=TERRAIN_SHEET.blends[n.terrain]?.[`${dx}:${dy}`];
      if(row!=null)output.drawImage(sheet,hash(`${n.terrain}:${c.x+dx}:${c.y+dy}`)%6*32,row*32,32,32,x,y,32,32);
    }
  };
  sc.save();sc.clip(coast.path);
  for(const c of cells) {
    const key=c.terrain,seed=roomSeeds.get(`${c.x}:${c.y}`)||`${key}:${c.x}:${c.y}`,x=(c.x-bounds.minX)*32,y=(c.y-bounds.minY)*32;
    paintGround(c);
    const neighbours=directions.map(([dx,dy])=>byCell.get(`${c.x+dx}:${c.y+dy}`));
    if(key!=='fog') {
      let first=null,second=null;
      for(const a of areas) {
        const d=(c.x-a.x)**2+(c.y-a.y)**2;
        if(!first||d<first.d){second=first;first={area:a.area,d}}
        else if(!second||d<second.d)second={area:a.area,d};
      }
      const edge=second?Math.max(0,1-(Math.sqrt(second.d)-Math.sqrt(first.d))/3):0;
      if(edge&&coast.tileKinds.get(`${c.x}:${c.y}`)==='solid'){const color=`rgba(${hash(first.area)%2?'198,167,86':'81,132,140'},${edge*.045})`;sc.fillStyle=color;sc.fillRect(x,y,32,32);tints.push({x,y,color})}
    }
    if(['mountain','snow'].includes(key)&&c.x%3===0&&c.y%3===0)ridges.push({x:c.x,y:c.y,...reliefAt(c,depth)});
    if(['forest','swamp'].includes(key)&&c.x%2===0&&c.y%2===0)ridges.push({x:c.x,y:c.y,kind:forestSpecies(nearestArea(c),key),size:key==='swamp'?42:66});
    if(key!=='fog'&&neighbours.some(n=>!n)&&hash(`${c.x}:${c.y}`)%11===0)coastDetails.push({x:c.x,y:c.y,kind:['mountain','snow'].includes(key)?'cliff':'offshore'});
    if(key==='water')ambience.push({x:c.x,y:c.y,kind:'water'});
  }
  const fringe=new Map();
  for(const c of cells)for(const [dx,dy] of directions) {
    const x=c.x+dx,y=c.y+dy,key=`${x}:${y}`;if(!byCell.has(key)&&!fringe.has(key))fringe.set(key,{x,y,terrain:c.terrain});
  }
  for(const c of fringe.values())paintGround(c,false);
  sc.restore();
  const detailStamp=(target,item,size)=>{
    const at=native(item.x,item.y),seed=roomSeeds.get(`${item.x}:${item.y}`)||`${item.kind}:${item.x}:${item.y}`;
    drawMapSprite(target,sheet,item.kind,seed,at.x-size/2,at.y-size*.57,size,'decorations',item.variant);
  };
  const paintDetail=(target,close)=>{
    target.save();target.clip(coast.path);
    const items=close?[...ridges.filter(r=>['ridge','hill'].includes(r.kind)),...cells.filter(c=>['forest','swamp'].includes(c.terrain)).map(c=>({...c,kind:forestSpecies(nearestArea(c),c.terrain),size:c.terrain==='swamp'?28:34}))]:ridges;
    for(const item of items)detailStamp(target,item,close&&['ridge','hill'].includes(item.kind)?item.size*.75:item.size);
    for(const c of coastDetails) {
      const at=native(c.x,c.y);
      if(c.kind==='cliff')detailStamp(target,{...c,variant:hash(`${c.x}:${c.y}`)%6},40);
      else continue;
    }
    for(const c of cells)if(c.terrain==='fog') {
      const x=(c.x-bounds.minX)*32,y=(c.y-bounds.minY)*32;target.clearRect(x,y,32,32);
      target.imageSmoothingEnabled=true;target.drawImage(coast.bands,x/4,y/4,8,8,x,y,32,32);paintGround(c,false,target);
    }
    target.restore();
    for(const c of coastDetails.filter(c=>c.kind==='offshore')) {
      const at=native(c.x,c.y);
      {
        const dx=!byCell.has(`${c.x-1}:${c.y}`)?-1:!byCell.has(`${c.x+1}:${c.y}`)?1:0;
        const dy=!byCell.has(`${c.x}:${c.y-1}`)?-1:!byCell.has(`${c.x}:${c.y+1}`)?1:0;
        sprite(target,sheet,c.kind,`${c.x}:${c.y}`,at.x+dx*28-12,at.y+dy*28-12,24);
      }
    }

  };
  // Roads sit on the ground. Relief and tree canopies, followed by town
  // paving and all room-owned structures, are painted over them.
  paintRoadNetwork(sc,roads,native,coast.path);
  paintDetail(sc,false);
  const bridges=roadWaterCrossings(roads,byCell),townKey=JSON.stringify([artInputs,bounds,bridges]);
  let art=townLayers.get(townKey);
  if(art)ambience.push(...art.smoke);
  else {
    const overlay=makeCanvas(w,h),ctx=overlay.getContext('2d');
    const townAreas=new Map(),glyphs=[],fortifications=[];
    for(const g of groups)if(g.town&&g.discovered){if(!townAreas.has(g.area))townAreas.set(g.area,[]);townAreas.get(g.area).push(g)}
    for(const town of townAreas.values()) {
      fortifications.push({area:town[0].area,...paintTown(ctx,sheet,town,roads,native)});
      const keep=town.some(g=>knownFeatures(g).includes('keep')||g.members.some(p=>p.discovered&&p.terrain==='castle'));
      glyphs.push({x:town.reduce((n,g)=>n+g.x,0)/town.length,y:town.reduce((n,g)=>n+g.y,0)/town.length,roomId:town[0].id,kind:keep?'keep':town.length>3?'town':'village'});
    }
    for(const b of bridges){const at=native(b.x,b.y);ctx.save();ctx.translate(at.x,at.y);ctx.rotate(b.angle);sprite(ctx,sheet,'bridge',`${b.from}:${b.to}`,-20,-20,40);ctx.restore()}
    const buildings=[],stamps=[],mills=new Set();
    const placeStamp=(group,kind,x,y,size,seed)=>{
      const at=native(x,y);sprite(ctx,sheet,kind,seed,at.x-size/2,at.y-size*.58,size);
      buildings.push({surfaceId:group.id,x,y,half:size/64});stamps.push({roomId:group.id,kind,x,y});
      if(kind==='forge')ambience.push({x:x+.23,y:y-.44,kind:'smoke'});
    };
    for(const group of groups) {
      const known=group.members.filter(p=>p.discovered);if(!known.length)continue;
      const interiors=known.filter(p=>p.mapRole==='interior'),features=knownFeatures(group);
      const count=Math.min(5,Math.max(interiors.length,group.town?2:0)),offsets=[[-.63,-.48],[.65,-.32],[-.64,.57],[.64,.67],[0,-1.05]];
      for(let i=0;i<count;i++){const [ox,oy]=offsets[i],room=interiors[i]||known[0],kind=roofKind(room);placeStamp(group,kind,group.x+ox,group.y+oy,kind==='keep'?41:29,room.artSeed||`${group.id}:${i}`)}
      if(features.includes('keep')&&!interiors.some(p=>roofKind(p)==='keep'))placeStamp(group,'keep',group.x-.12,group.y-.55,43,group.id);
      if(features.includes('tower')&&!group.town)placeStamp(group,'tower',group.x,group.y-.45,34,group.id);
      for(const kind of ['dock','ruins','graveyard','magic','farm','reeds','stump','flowers'])if(features.includes(kind)&&(!group.town||!['reeds','flowers'].includes(kind)))placeStamp(group,kind,group.x+(count ? .8 : 0),group.y+.32,kind==='dock'?37:30,`${group.id}:${kind}`);
      if(known.some(p=>(p.entrances||[]).length))placeStamp(group,features.includes('mine')?'mine':'cave',group.x,group.y,29,group.id);
      if(group.town&&hash(group.id)%4===0)placeStamp(group,'lantern',group.x-.65,group.y+.5,20,group.id);
      if(group.town&&hash(group.id)%7===0)placeStamp(group,hash(group.id)%2?'well':'cart',group.x+.55,group.y-.2,22,group.id);
      if(features.includes('farm')){
        if(!mills.has(group.area)){placeStamp(group,'windmill',group.x+.8,group.y-.8,36,group.id);mills.add(group.area)}
        placeStamp(group,'fence',group.x-.65,group.y+.4,24,group.id);
      }
      if(/plateau/i.test(group.area)&&hash(group.id)%3===0)placeStamp(group,'mesa',group.x+.65,group.y+.4,40,group.id);
      if(/highland|foothill/i.test(group.area)&&hash(group.id)%3===0)placeStamp(group,'outcrop',group.x+.65,group.y+.4,32,group.id);
    }
    for(const wall of fortifications) {
      for(const p of wall.towers){const at=native(p.x,p.y);sprite(ctx,sheet,'tower',`${wall.area}:${p.x}:${p.y}`,at.x-11,at.y-16,22)}
      for(const gate of wall.gates){const group=groups.find(g=>g.id===gate.roomId);if(group)placeStamp(group,'gatehouse',gate.x,gate.y,34,group.id)}
    }
    art={overlay,buildings,stamps,glyphs,fortifications,smoke:ambience.filter(a=>a.kind==='smoke')};
    townLayers.set(townKey,art);if(townLayers.size>2)townLayers.delete(townLayers.keys().next().value);
  }
  const {overlay,buildings,stamps,glyphs,fortifications}=art;
  sc.drawImage(overlay,0,0);
  let nearCanvas=null;
  const closeCanvas=()=>{
    if(nearCanvas)return nearCanvas;
    const start=performance.now();nearCanvas=makeCanvas(w,h);const target=nearCanvas.getContext('2d');
    target.imageSmoothingEnabled=true;target.drawImage(coast.bands,0,0,w,h);
    target.save();target.clip(coast.path);
    for(const c of cells)paintGround(c,true,target);for(const c of fringe.values())paintGround(c,false,target);
    for(const tint of tints){target.fillStyle=tint.color;target.fillRect(tint.x,tint.y,32,32)}
    target.restore();paintRoadNetwork(target,roads,native,coast.path);paintDetail(target,true);target.drawImage(overlay,0,0);
    nearCanvas=freezeCanvas(nearCanvas);result.closeBakeMs=performance.now()-start;return nearCanvas;
  };
  const result={canvas:freezeCanvas(canvas),closeCanvas,bounds,buildings,ready,model,glyphs,ridges,ambience,bridges,stamps,fortifications,coastDetails,coastMask:coast.mask,coastScale:4,bakeMs:performance.now()-bakeStart};
  rasterCache.set(atlas,result);recentRasters.set(signature,result);if(recentRasters.size>2)recentRasters.delete(recentRasters.keys().next().value);
  return result;
}
