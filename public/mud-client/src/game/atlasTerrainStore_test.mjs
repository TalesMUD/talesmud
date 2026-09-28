import assert from 'node:assert/strict';
import fs from 'node:fs';
const source = fs.readFileSync(new URL('./MUDXPlusStore.js', import.meta.url), 'utf8');
function readFunction(name) {
  const start = source.indexOf(`function ${name}(`);
  const end = source.indexOf('\nfunction ', start + 1);
  return source.slice(start, end);
}
const {mergeAtlas,sameAtlasSnapshot} = new Function(`${readFunction('isAtlasCurrentPlace')}\n${readFunction('emptyAtlas')}\n${readFunction('mergeAtlas')}\n${readFunction('sameAtlasSnapshot')}\nreturn {mergeAtlas,sameAtlasSnapshot};`)();
const explored={places:[{id:'A',terrain:'forest',kind:'wild',discovered:true}],paths:[],regions:[],layers:[]};
const fog={places:[{id:'A',terrain:'fog',kind:'uncharted',discovered:false}],paths:[],regions:[],layers:[]};
const merged=mergeAtlas(explored,fog);
assert.equal(merged.places[0].terrain,'forest');
assert.equal(merged.places[0].kind,'wild');
assert.equal(merged.places[0].discovered,true);
const changed=mergeAtlas(explored,{...explored,places:[{...explored.places[0],terrain:'ruins'}]});
assert.equal(sameAtlasSnapshot(explored,changed),false,'terrain-only update must reach the canvas');
console.log('atlasTerrainStore: explored terrain survives fog snapshots; terrain updates repaint');
const ground={...explored,landscape:[{x:0,y:0,terrain:'forest'},{x:1,y:0,terrain:'fog'}]};
const mergedGround=mergeAtlas(ground,{...fog,landscape:[{x:0,y:0,terrain:'fog'},{x:1,y:0,terrain:'grassland'}]});
assert.deepEqual(mergedGround.landscape.map(p=>p.terrain),['forest','grassland']);
assert.equal(sameAtlasSnapshot(ground,mergedGround),false,'landscape disclosure repaints');
for(const change of [{x:2},{mapRole:'interior',surfaceRoomId:'B'},{town:true},{entrances:['C']}]) {
 assert.equal(sameAtlasSnapshot(explored,{...explored,places:[{...explored.places[0],...change}]}),false,'presentation updates repaint');
}
assert.equal(mergeAtlas(ground,{...explored,landscape:[{x:1,y:0,terrain:'forest'}]}).landscape.length,1,'layout changes drop obsolete filler cells');
console.log('atlasTerrainStore: landscape merge and surface presentation updates OK');
