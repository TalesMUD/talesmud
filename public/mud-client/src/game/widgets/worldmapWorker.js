import { continentRaster } from './continentRenderer.js';
let sheet;
self.onmessage=async ({data})=>{
  if(data.type==='sheet'){sheet=data.sheet;return}
  if(data.type!=='scene')return;
  try {
    const scene=continentRaster(data.atlas,sheet);
    const source=data.close?scene.closeCanvas():scene.canvas;
    const bitmap=await createImageBitmap(source);
    if(data.close){self.postMessage({id:data.id,bitmap,bakeMs:scene.closeBakeMs},[bitmap]);return}
    const coast=scene.coastMask.getContext('2d').getImageData(0,0,scene.coastMask.width,scene.coastMask.height);
    const message={id:data.id,bitmap,coast:{width:coast.width,height:coast.height,pixels:coast.data}};
    for(const key of ['bakeMs','buildings','glyphs','ridges','ambience','bridges','stamps','fortifications','coastDetails'])message[key]=scene[key];
    self.postMessage(message,[bitmap,coast.data.buffer]);
  } catch(error){self.postMessage({id:data.id,error:String(error)})}
};
