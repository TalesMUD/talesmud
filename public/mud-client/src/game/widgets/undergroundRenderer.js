import { hash, makeCanvas, sheetReady, drawMapSprite } from './mapArt.js';
const cache=new WeakMap(),recent=new Map();
export function undergroundModel(atlas) {
  const places=(atlas.places||[]).filter(p=>p.layer==='lower');if(!places.length)return null;
  const byId=new Map(places.map(p=>[p.id,p])),byCell=new Map(places.map(p=>[`${p.x}:${p.y}`,p]));
  const seen=new Set(),corridors=[];
  for(const path of atlas.paths||[]) {
    const a=byId.get(path.from),b=byId.get(path.to);if(!a||!b||!a.discovered||!b.discovered||a.area!==b.area)continue;
    const key=[a.id,b.id].sort().join('|');if(seen.has(key)||['up','down'].includes(path.dir))continue;
    seen.add(key);corridors.push({a,b});
  }
  const xs=places.map(p=>p.x),ys=places.map(p=>p.y);
  return {places,byCell,corridors,bounds:{minX:Math.min(...xs)-2,maxX:Math.max(...xs)+2,minY:Math.min(...ys)-2,maxY:Math.max(...ys)+2}};
}
export function undergroundRaster(atlas,sheet) {
  const ready=!!sheetReady(sheet),prev=cache.get(atlas);if(prev&&prev.ready===ready)return prev;
  const model=undergroundModel(atlas);if(!model)return null;
  const {places,byCell,corridors,bounds}=model;
  const key=JSON.stringify([ready,places.map(p=>[p.id,p.x,p.y,p.discovered,p.undergroundStyle,p.artSeed,p.exits]),corridors.map(({a,b})=>[a.id,b.id])]);
  if(recent.has(key)){const r=recent.get(key);cache.set(atlas,r);return r}
  const canvas=makeCanvas((bounds.maxX-bounds.minX+1)*32,(bounds.maxY-bounds.minY+1)*32),ctx=canvas.getContext('2d');
  const native=(x,y)=>({x:(x-bounds.minX+.5)*32,y:(y-bounds.minY+.5)*32});
  const torches=[],stairs=[];
  // Rock-sided corridors follow actual exits. Isolated rooms remain isolated;
  // space between underground clusters is a dark void, never filler terrain.
  for(const {a,b} of corridors) {
    const p=native(a.x,a.y),q=native(b.x,b.y);
    ctx.beginPath();ctx.moveTo(p.x,p.y);ctx.lineTo(q.x,q.y);ctx.lineCap='round';
    ctx.strokeStyle='#182d32';ctx.lineWidth=25;ctx.stroke();ctx.strokeStyle='#758276';ctx.lineWidth=19;ctx.stroke();
    ctx.strokeStyle=a.undergroundStyle==='sewer'?'#3f6765':'#516164';ctx.lineWidth=13;ctx.stroke();ctx.strokeStyle='#6b7973';ctx.lineWidth=2;ctx.stroke();
  }
  // Draw the rock envelope first, then every floor, then perimeter details.
  // Adjacent floors merge; a later room's border never covers its neighbor.
  for(const p of places) {
    const at=native(p.x,p.y);
    ctx.save();ctx.shadowColor='#020c12';ctx.shadowBlur=7;ctx.shadowOffsetY=3;
    ctx.fillStyle=p.discovered?'#758479':'#363d3b';ctx.beginPath();ctx.roundRect(at.x-20,at.y-20,40,40,8);ctx.fill();ctx.restore();
  }
  for(const p of places) {
    const at=native(p.x,p.y),x=at.x-16,y=at.y-16;
    const open=(dx,dy)=>byCell.get(`${p.x+dx}:${p.y+dy}`)?.discovered;
    const corners=[!open(-1,0)&&!open(0,-1)?5:0,!open(1,0)&&!open(0,-1)?5:0,!open(1,0)&&!open(0,1)?5:0,!open(-1,0)&&!open(0,1)?5:0];
    ctx.save();ctx.beginPath();ctx.roundRect(x,y,32,32,corners);ctx.clip();
    drawMapSprite(ctx,sheet,p.discovered?(p.undergroundStyle||'cave'):'fog',p.artSeed||p.id,x,y,32,p.discovered?'underground':'rows');ctx.restore();
  }
  for(const p of places) {
    const at=native(p.x,p.y);
    if(!p.discovered)continue;
    // Exposed rock walls get irregular stones and a bright upper lip. Adjacent
    // rooms share open ground, avoiding separate bordered room rectangles.
    for(const [dx,dy] of [[-1,0],[1,0],[0,-1],[0,1]]) {
      if(byCell.has(`${p.x+dx}:${p.y+dy}`))continue;
      const rotation=dx===-1?-Math.PI/2:dx===1?Math.PI/2:dy===1?Math.PI:0;
      ctx.save();ctx.translate(at.x,at.y);ctx.rotate(rotation);
      for(let i=0;i<3;i++) {
        const sx=-16+i*11,offset=hash(`${p.id}:${dx}:${dy}:${i}`)%3;
        ctx.fillStyle=i%2?'#7e8a7b':'#617672';ctx.beginPath();ctx.moveTo(sx,-14-offset);ctx.lineTo(sx+4,-20-offset);ctx.lineTo(sx+10,-19);ctx.lineTo(sx+12,-13);ctx.lineTo(sx,-12);ctx.fill();
        ctx.strokeStyle='#a7ad94';ctx.lineWidth=1;ctx.beginPath();ctx.moveTo(sx+3,-18-offset);ctx.lineTo(sx+9,-18);ctx.stroke();
      }
      ctx.restore();
    }
    if(hash(p.id)%3!==0||p.undergroundStyle!=='cave') {
      const tx=p.x-.33,ty=p.y-.35,a=native(tx,ty);
      drawMapSprite(ctx,sheet,'torch',p.id,a.x-11,a.y-14,22);
      torches.push({x:tx,y:ty,roomId:p.id});
    }
    if((p.exits||[]).some(e=>e.vertical)) {
      drawMapSprite(ctx,sheet,'stairs',p.id,at.x-11,at.y-12,23);stairs.push({roomId:p.id,x:p.x,y:p.y});
    }
  }
  const result={canvas,bounds,ready,model,buildings:[],glyphs:[],ridges:[],ambience:[],torches,stairs};
  cache.set(atlas,result);recent.set(key,result);if(recent.size>2)recent.delete(recent.keys().next().value);
  return result;
}
