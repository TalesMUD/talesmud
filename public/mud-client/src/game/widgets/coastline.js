import { makeCanvas } from './mapArt.js';
let lastCoastKey,lastCoast;
const BAYER=[0,8,2,10,12,4,14,6,3,11,1,9,15,7,13,5];
// Palette/threshold tables are shared constants, independent of atlas data.
const {bandTable,maskTable}=(()=>{
  const bandBytes=new Uint8ClampedArray(8192*4),maskBytes=new Uint8ClampedArray(4096*4);
  for(let v=18;v<256;v++)for(let t=0;t<16;t++) {
    const threshold=(t+.5)/16,ground=Math.max(0,Math.min(1,(v-132)/26));
    if(ground>=threshold)maskBytes.set([255,255,255,255],(v*16+t)*4);
    let color;
    if(v>=123)color=[207,185,130];
    else if(v>=112&&threshold>.48)color=[178,212,191];
    else {const depth=Math.max(0,Math.min(1,(v-25)/85));color=[23+Math.round(depth*49),57+Math.round(depth*85),71+Math.round(depth*84)]}
    const alpha=Math.min(255,(v-18)*6);
    bandBytes.set([...color,alpha],(v*16+t)*4);
    bandBytes.set(v<180?[116,106,81,alpha]:[...color,alpha],(4096+v*16+t)*4);
  }
  return {bandTable:new Uint32Array(bandBytes.buffer),maskTable:new Uint32Array(maskBytes.buffer)};
})();
// A blurred binary field rejects narrow cell spurs and rounds diagonal shores.
// Thresholds make beach/foam/shallow bands follow the same organic contour.
export function smoothCoast(model) {
  const {bounds,cells,byCell}=model,scale=12;
  const signature=JSON.stringify([bounds,cells.map(c=>[c.x,c.y,c.terrain==='fog'])]);
  if(signature===lastCoastKey)return lastCoast;
  const w=(bounds.maxX-bounds.minX+1)*scale,h=(bounds.maxY-bounds.minY+1)*scale;
  const raw=makeCanvas(w,h),rc=raw.getContext('2d');rc.fillStyle='#fff';
  for(const c of cells)rc.fillRect((c.x-bounds.minX)*scale,(c.y-bounds.minY)*scale,scale,scale);
  const field=makeCanvas(w,h),fc=field.getContext('2d',{willReadFrequently:true});
  fc.filter='blur(7.5px)';fc.drawImage(raw,0,0);fc.filter='none';
  const alpha=fc.getImageData(0,0,w,h).data;
  const mask=makeCanvas(w,h),mc=mask.getContext('2d'),mi=mc.createImageData(w,h);
  const bands=makeCanvas(w,h),bc=bands.getContext('2d'),bi=bc.createImageData(w,h);
  const fogGrid=new Uint8Array(w*h);
  const findFog=(gx,gy)=>{
    const direct=byCell.get(`${gx}:${gy}`);if(direct)return direct.terrain==='fog';
    for(let r=1;r<=2;r++)for(const [dx,dy] of [[-r,0],[r,0],[0,-r],[0,r],[-r,-r],[r,-r],[-r,r],[r,r]]){
      const c=byCell.get(`${gx+dx}:${gy+dy}`);if(c)return c.terrain==='fog';
    }
    return true;
  };
  for(let gy=0;gy<h/scale;gy++)for(let gx=0;gx<w/scale;gx++)if(findFog(gx+bounds.minX,gy+bounds.minY)) {
    for(let yy=0;yy<scale;yy++)fogGrid.fill(1,(gy*scale+yy)*w+gx*scale,(gy*scale+yy)*w+(gx+1)*scale);
  }
  // RGBA lookup tables keep the sampled contour loop free of allocations and
  // repeated spatial lookups. Uint32 views copy complete pixels in one store.
  const bandPixels=new Uint32Array(bi.data.buffer),maskPixels=new Uint32Array(mi.data.buffer),tileFlags=new Uint8Array(w/scale*h/scale);
  for(let y=0;y<h;y++)for(let x=0;x<w;x++) {
    const i=y*w+x,v=alpha[i*4+3],t=BAYER[(y%4)*4+x%4],key=v*16+t;
    // Soft fog overlay owns unexplored look; coast bands stay organic (no Bayer fog checker).
    maskPixels[i]=maskTable[key];bandPixels[i]=bandTable[key];
    // Broken foam glints follow the existing band, never charted ground or fog.
    if(!fogGrid[i]&&v>=112&&v<123&&((x*17+y*31)%11<3))bandPixels[i]=bandTable[v*16+15];
    tileFlags[Math.floor(y/scale)*(w/scale)+Math.floor(x/scale)]|=maskPixels[i]?1:2;
  }
  mc.putImageData(mi,0,0);bc.putImageData(bi,0,0);
  // Scale these twelve samples per cell only when compositing, avoiding two
  // unnecessary multi-megapixel intermediate canvases.
  const tileKinds=new Map();
  for(let gy=0;gy<h/scale;gy++)for(let gx=0;gx<w/scale;gx++) {
    const flag=tileFlags[gy*w/scale+gx];
    tileKinds.set(`${gx+bounds.minX}:${gy+bounds.minY}`,flag===1?'solid':flag===2?'empty':'edge');
  }
  // Trace the threshold contour once. A compact vector clip replaces hundreds
  // of per-tile alpha-compositing operations while the depth bands stay sampled.
  const edges=new Map(),stride=w+1;
  const point=(x,y)=>y*stride+x;
  const land=(x,y)=>x>=0&&x<w&&y>=0&&y<h&&alpha[(y*w+x)*4+3]>=145;
  for(let y=0;y<h;y++)for(let x=0;x<w;x++)if(land(x,y)) {
    if(!land(x,y-1))edges.set(point(x,y),point(x+1,y));
    if(!land(x+1,y))edges.set(point(x+1,y),point(x+1,y+1));
    if(!land(x,y+1))edges.set(point(x+1,y+1),point(x,y+1));
    if(!land(x-1,y))edges.set(point(x,y+1),point(x,y));
  }
  const path=new Path2D();
  while(edges.size) {
    const first=edges.keys().next().value,loop=[];let key=first;
    while(edges.has(key)) {
      loop.push({x:key%stride*32/scale,y:Math.floor(key/stride)*32/scale});
      const next=edges.get(key);edges.delete(key);key=next;if(key===first)break;
    }
    const corners=loop.filter((p,i)=>{const a=loop[(i+loop.length-1)%loop.length],b=loop[(i+1)%loop.length];return (p.x-a.x)*(b.y-p.y)!==(p.y-a.y)*(b.x-p.x)});
    if(corners.length<4)continue;
    const start=corners[0],last=corners.at(-1);path.moveTo((start.x+last.x)/2,(start.y+last.y)/2);
    corners.forEach((p,i)=>{const next=corners[(i+1)%corners.length];path.quadraticCurveTo(p.x,p.y,(p.x+next.x)/2,(p.y+next.y)/2)});path.closePath();
  }
  lastCoastKey=signature;lastCoast={mask,bands,tileKinds,path,scale:32/scale};
  return lastCoast;
}
