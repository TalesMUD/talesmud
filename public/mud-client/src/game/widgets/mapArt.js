import { TERRAIN_SHEET } from './terrainSheet.js';
export const hash = text => { let n=0; for (const c of String(text)) n=(n*31+c.charCodeAt(0))|0; return n>>>0; };
export const makeCanvas = (w,h) => { if(typeof OffscreenCanvas!=='undefined')return new OffscreenCanvas(w,h);const c=document.createElement('canvas');c.width=w;c.height=h;return c; };
export const freezeCanvas = canvas => typeof canvas.transferToImageBitmap==='function'?canvas.transferToImageBitmap():canvas;
export const sheetReady = sheet => sheet && (sheet.complete && sheet.naturalWidth > 0 || typeof sheet.close==='function' && sheet.width > 0);
export function drawMapSprite(ctx,sheet,key,seed,x,y,size,section='decorations',variant=null) {
  const row=TERRAIN_SHEET[section]?.[key];
  if(row==null)return;
  if(sheetReady(sheet)) {
    ctx.imageSmoothingEnabled=false;
    ctx.drawImage(sheet,(variant??hash(seed)%TERRAIN_SHEET.variants)*32,row*32,32,32,x,y,size,size);
  } else if(section!=='decorations') {
    ctx.fillStyle=TERRAIN_SHEET.colors?.[key]||'#485b5c';ctx.fillRect(x,y,size,size);
  }
}
export function knownFeatures(group) {
  return [...new Set(group.members.filter(p=>p.discovered).flatMap(p=>p.mapFeatures||[]))];
}
export function roofKind(room) {
  return ['keep','forge','shrine','tavern','shop','tower','farm'].find(k=>(room?.mapFeatures||[]).includes(k))||'roof';
}
export function roadWaterCrossings(roads,byCell) {
  const seen=new Set(),out=[];
  for(const {a,b} of roads) {
    const steps=Math.max(1,Math.ceil(Math.hypot(b.x-a.x,b.y-a.y)*3));
    for(let i=0;i<=steps;i++) {
      const t=i/steps,x=a.x+(b.x-a.x)*t,y=a.y+(b.y-a.y)*t;
      const cell=byCell.get(`${Math.round(x)}:${Math.round(y)}`),key=cell&&`${cell.x}:${cell.y}`;
      if(cell?.terrain==='water'&&!seen.has(key)) {
        seen.add(key);out.push({x:cell.x,y:cell.y,angle:Math.atan2(b.y-a.y,b.x-a.x),from:a.id,to:b.id});
      }
    }
  }
  return out;
}
export function roundedHull(points,padding=24) {
  // Convex envelope of the street graph plus room courtyards; walls follow
  // street geography instead of a pale ellipse painted across the district.
  const pts=points.flatMap(p=>[[-padding,-padding],[padding,-padding],[padding,padding],[-padding,padding]].map(([x,y])=>({x:p.x+x,y:p.y+y})));
  pts.sort((a,b)=>a.x-b.x||a.y-b.y);
  const cross=(a,b,c)=>(b.x-a.x)*(c.y-a.y)-(b.y-a.y)*(c.x-a.x);
  const half=list=>{const h=[];for(const p of list){while(h.length>1&&cross(h.at(-2),h.at(-1),p)<=0)h.pop();h.push(p)}return h};
  return [...half(pts).slice(0,-1),...half([...pts].reverse()).slice(0,-1)];
}
export function hullPath(ctx,points) {
  if(!points.length)return;
  const first=points[0],last=points.at(-1);
  ctx.beginPath();ctx.moveTo((first.x+last.x)/2,(first.y+last.y)/2);
  for(let i=0;i<points.length;i++){const p=points[i],next=points[(i+1)%points.length];ctx.quadraticCurveTo(p.x,p.y,(p.x+next.x)/2,(p.y+next.y)/2)}
  ctx.closePath();
}
