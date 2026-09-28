import { makeCanvas } from './mapArt.js';
let lastCoastKey,lastCoast;
const BAYER=[0,8,2,10,12,4,14,6,3,11,1,9,15,7,13,5];
// A blurred binary field rejects narrow cell spurs and rounds diagonal shores.
// Thresholds make beach/foam/shallow bands follow the same organic contour.
export function smoothCoast(model) {
  const {bounds,cells,byCell}=model,scale=8;
  const signature=JSON.stringify([bounds,cells.map(c=>[c.x,c.y,c.terrain==='fog'])]);
  if(signature===lastCoastKey)return lastCoast;
  const w=(bounds.maxX-bounds.minX+1)*scale,h=(bounds.maxY-bounds.minY+1)*scale;
  const raw=makeCanvas(w,h),rc=raw.getContext('2d');rc.fillStyle='#fff';
  for(const c of cells)rc.fillRect((c.x-bounds.minX)*scale,(c.y-bounds.minY)*scale,scale,scale);
  const field=makeCanvas(w,h),fc=field.getContext('2d',{willReadFrequently:true});
  fc.filter='blur(5px)';fc.drawImage(raw,0,0);fc.filter='none';
  const data=fc.getImageData(0,0,w,h).data;
  const mask=makeCanvas(w,h),mc=mask.getContext('2d'),mi=mc.createImageData(w,h);
  const bands=makeCanvas(w,h),bc=bands.getContext('2d'),bi=bc.createImageData(w,h);
  const fogCells=new Map();
  const findFog=(gx,gy)=>{
    const direct=byCell.get(`${gx}:${gy}`);if(direct)return direct.terrain==='fog';
    for(let r=1;r<=2;r++)for(const [dx,dy] of [[-r,0],[r,0],[0,-r],[0,r],[-r,-r],[r,-r],[-r,r],[r,r]]){
      const c=byCell.get(`${gx+dx}:${gy+dy}`);if(c)return c.terrain==='fog';
    }
    return true;
  };
  const nearestFog=(gx,gy)=>{
    const key=`${gx}:${gy}`;
    if(!fogCells.has(key))fogCells.set(key,findFog(gx,gy));
    return fogCells.get(key);
  };
  for(let y=0;y<h;y++)for(let x=0;x<w;x++) {
    const k=(y*w+x)*4,v=data[k+3];if(v<18)continue;
    const fog=v<180&&nearestFog(Math.floor(x/scale)+bounds.minX,Math.floor(y/scale)+bounds.minY);
    const threshold=(BAYER[(y%4)*4+x%4]+.5)/16;
    const ground=Math.max(0,Math.min(1,(v-132)/26));
    mi.data[k]=mi.data[k+1]=mi.data[k+2]=255;mi.data[k+3]=ground>=threshold?255:0;
    let color,alpha=Math.min(255,(v-18)*6);
    if(fog)color=[116,106,81];
    else if(v>=123)color=[207,185,130];
    else if(v>=112&&threshold>.48)color=[178,212,191];
    else {
      const t=Math.max(0,Math.min(1,(v-25)/85));
      color=[23+Math.round(t*49),57+Math.round(t*85),71+Math.round(t*84)];
    }
    [bi.data[k],bi.data[k+1],bi.data[k+2]]=color;bi.data[k+3]=alpha;
  }
  mc.putImageData(mi,0,0);bc.putImageData(bi,0,0);
  // Native pixel resolution keeps the original pixel art crisp; the contour
  // itself is sampled more finely than room cells, never as staircase tiles.
  const upscale=c=>{const out=makeCanvas(w*4,h*4),ctx=out.getContext('2d');ctx.imageSmoothingEnabled=true;ctx.drawImage(c,0,0,out.width,out.height);return out};
  lastCoastKey=signature;lastCoast={mask:upscale(mask),bands:upscale(bands)};
  return lastCoast;
}
