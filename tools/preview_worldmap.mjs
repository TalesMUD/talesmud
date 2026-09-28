// Build a local, read-only preview of the production overlay. No game service.
// Usage: node tools/preview_worldmap.mjs <map-preview JSON> <output directory>
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const client = path.join(root, 'public/mud-client');
const require = createRequire(path.join(client, 'package.json'));
const { rollup } = require('rollup');
const svelte = require('rollup-plugin-svelte');
const resolve = require('@rollup/plugin-node-resolve').default;
const css = require('rollup-plugin-css-only');
const [source, output] = process.argv.slice(2);
if (!source || !output) throw new Error('Provide map-preview JSON and output directory');
await fs.mkdir(output, { recursive: true });
const entry = path.join(client, '.worldmap-preview-entry.mjs');
process.chdir(client);
try {
  await fs.writeFile(entry, `
import { writable } from 'svelte/store';
import Overlay from './src/game/ui/MapOverviewOverlay.svelte';
import { onMapTilesReady, paintAtlas } from './src/game/widgets/atlasRenderer.js';
import { surfaceGroups, groupForRoom } from './src/game/widgets/surfaceAtlas.js';
import { landscapeModel, continentRaster } from './src/game/widgets/continentRenderer.js';
document.fonts.load('24px \"Material Icons\"').then(()=>document.documentElement.classList.add('material-icons-ready'));
const snapshot = await (await fetch('./atlas.json')).json();
const store = writable({atlas:snapshot.atlas,currentRoomId:snapshot.atlas.currentRoomId,mapOverviewOpen:true,atlasLayer:'overworld'});
store.selectMapPlace = id => store.update(s=>({...s,mapSelectedId:id}));
store.closeMapOverview = () => store.update(s=>({...s,mapOverviewOpen:false}));
store.setAtlasLayer = id => store.update(s=>({...s,atlasLayer:id}));
new Overlay({target:document.getElementById('app'),props:{store}});
window.__mapPreview = { store, snapshot, paintAtlas, surfaceGroups, groupForRoom, landscapeModel, continentRaster, ready:false };
onMapTilesReady(()=>{window.__mapPreview.ready=true});
`);
  const bundle = await rollup({ input:entry, plugins:[svelte({emitCss:true}), css({output:'preview.css'}), resolve({browser:true,dedupe:['svelte']})] });
  await bundle.write({file:path.join(output,'preview.js'),format:'es'});
  await bundle.close();
} finally { await fs.rm(entry,{force:true}); }
await fs.copyFile(source,path.join(output,'atlas.json'));
for (const f of ['global.css','icons.css']) await fs.copyFile(path.join(client,'public',f),path.join(output,f));
for (const folder of ['fonts']) await fs.cp(path.join(client,'public',folder),path.join(output,folder),{recursive:true});
await fs.mkdir(path.join(output,'api/map-tiles'),{recursive:true});
await fs.copyFile(path.join(client,'public/map-tiles/terrain-sheet.png'),path.join(output,'api/map-tiles/terrain-sheet.png'));
await fs.writeFile(path.join(output,'index.html'), `<!doctype html><html><head><meta charset="utf-8"><title>Veilspan terrain map — local continent review</title><link rel="stylesheet" href="global.css"><link rel="stylesheet" href="icons.css"><link rel="stylesheet" href="preview.css"><style>body{margin:0;background:#091820}.review{position:fixed;left:24px;top:24px;color:#d5c08e;font:14px Georgia,serif;letter-spacing:.15em}</style></head><body><div class="review">VEILSPAN · CARTOGRAPHER / LOCAL TERRAIN REVIEW</div><div id="app"></div><script type="module" src="preview.js"></script></body></html>`);
console.log(`Preview built in ${output}. Serve locally, e.g. python3 -m http.server 8138 --bind 127.0.0.1 --directory ${output}`);
