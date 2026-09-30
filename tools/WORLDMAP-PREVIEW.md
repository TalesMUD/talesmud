# Continent map P1/P1b/P1c/P1d/P1e/P1f/P1g/P1j

The in-game atlas uses `pkg/worldmap/map_terrain.json`, embedded at Go build time. The content repository's `tools/build_public_map_data.py` serves the separate public lore map and does not supply room tiles. No public lore/spoiler export was changed.

The ordered classifier checks exact roomType/areaType, specific tags, room-name words, underground/indoor tags, exact area and area words, description/detail words, legacy biome/kind, then outdoor. Unknown ground defaults to grassland; the census counts fallback separately from deliberately classified grassland. Words match whole tokens; underscores and punctuation become spaces. Edit the JSON to change rules or area defaults, then rebuild Go. No room or character schema migration is needed.

Art is hand-authored through deterministic Pillow drawing code in the content repo. There are six native 48px variants for each of 14 terrains, fog/sea rows, and 40 transparent building/service/landmark/nature stamps plus four underground floor rows and 128 precomposed directional dither rows. City ground contains paving only. The sheet is 288×9024 RGBA, one cache-busted PNG request through `GET/HEAD /api/map-tiles/terrain-sheet.png`. The endpoint reads the embedded play assets. Keep the normal `public/mud-client/public/` → `pkg/webuiplay/dist/` copy in local build steps.

## Rebuild art

From the engine repository:

```sh
python3 ../talesmud-rpg-1/tools/generate_map_tiles.py --config pkg/worldmap/map_terrain.json
python3 tools/sync_map_tiles.py ../talesmud-rpg-1/assets/map-tiles
cd public/mud-client && npm run build
```

Pillow is needed only to rebuild art. The sync helper validates terrain/decoration row order, dimensions, fallback colors, default terrain, and the sheet content hash before writing the renderer manifest.

## Read-only room census and production-overlay preview

```sh
python3 tools/worldmap_snapshot.py --database talesmud.db --out /tmp/worldmap-rooms.json
# Or, for the larger canonical content export (needs PyYAML):
python3 tools/worldmap_snapshot.py --content-rooms ../talesmud-rpg-1/data/rooms --out /tmp/worldmap-rooms.json
go run ./cmd/map-preview -rooms /tmp/worldmap-rooms.json -current R0201 -fog R0416,R0417 > /tmp/worldmap-preview.json
node tools/preview_worldmap.mjs /tmp/worldmap-preview.json /tmp/worldmap-preview
python3 -m http.server 8140 --bind 127.0.0.1 --directory /tmp/worldmap-preview
```

Open `http://127.0.0.1:8140/`. This mounts the production Svelte overlay on a fully explored in-memory atlas, with the requested fog neighbors. It does not start the game, create a guest, write a database, or connect to production. The JSON includes counts and fallback IDs. `tools/worldmap_snapshot.py` exports only atlas inputs and opens SQLite with `mode=ro`.

For captures, install Puppeteer Core in a separate development environment, then set `PUPPETEER_MODULE` to its directory and `CHROMIUM_PATH` to your Chromium executable. Optionally set `WORLDMAP_CONTENT_JSON` to the census JSON for a larger-world performance check. Historical P1d checks use `node tools/capture_worldmap_p1d.cjs` against that commit; use `node tools/capture_worldmap_p1g.cjs` for the current art. The P1d tool runs against a local preview server (the P1 script retains its historical 8137 capture flow). It waits for terrain loading, `document.fonts.ready`, image decoding, and layout frames. It captures overview, Oldtown + Silverbrook zoom, a Lower close-up, minimum zoom and a Highlands close-up; exercises exterior selection, non-clickable filler, town filtering/interior selection, visible entrance switching, instanced interior marker, recenter, world fit, phone bounds, and Escape; and records changed-scene bake, marker-only snapshot, warm draw timings, and sheet requests in `.director/ux-audit/after/`.

## Layout and rendering

Embedded `pkg/worldmap/map_layout.json` configures compact zone centers (grid units), natural filler biome, town flags, minimum separation, and coast padding. Authored room coordinates describe local zone geometry and remain unchanged in storage. Unconfigured zones attach using inter-zone compass exits; disconnected areas get a nearby shelf position. Sorted rooms/exits/zones make repeated builds deterministic. Anonymous patches and broad bridges connect all surface zones; bridges add no room exits or roads. Actual authored outdoor water/shore/field/forest/swamp cells modify the natural ground locally.

Overworld interiors project onto their graph-nearest same-area exterior anchor (lexical tie breaks), including upstairs rooms. If no exterior exists, the isolated building remains one selectable anchor. Outdoor positive Z is elevation; negative depth or subterranean context puts rooms on Lower. The full atlas still contains actual interior IDs for intel/exits/travel. The side panel lists discovered interiors only, and visible Lower entrances do not disclose unexplored target names. Hidden entrances require exit revelation. Filler `{x,y,terrain}` cells are anonymous, non-clickable, and fogged until nearby ground or their surface area is explored.

The shared renderer caches native-resolution landscape pixels: eight-neighbor ordered-dither terrain transitions, masked coastal shallows, town paving/walls, 40 transparent building/service/landmark/nature stamps plus four underground floor rows, and dirt paths along real charted outdoor compass exits. Pan/zoom reuse the canvas. Marker-only atlas replacements reuse a bounded two-scene cache; all fog/art inputs participate in its key. Lower uses its own cached cave/crypt/cellar/sewer floors and rock-sided actual-exit corridors, perimeter walls, torches, and stairs. Upper retains room tiles and short same-zone connectors. World fit uses continent bounds, while minimap/recenter frame exterior groups. Stored coordinates, exits, and gameplay travel are unchanged.

## Boundaries of this pass

Local read-only SQLite and canonical YAML snapshots are used, never VPS data or live player sessions. The decorative continent is schematic: compact offsets resolve inconsistent authored coordinate origins and exit cycles, and dirt roads can cross the simplified natural ground. Terrain/layer/name inference can misinterpret evocative prose; prefer structured tags/coordinates or adjust mapping defaults. A changed explored scene requires a new bitmap bake; warm pan/zoom and marker-only snapshots reuse pixels. The browser report separates these costs rather than treating warm draw time as a device-independent frame-rate guarantee.

Additional landmark art, ornate zone banners, compass/legend ornaments, and VPS deployment remain outside this pass. P1b resolves P1's separate-island packing, missing continent/coast/road blending, and overworld interior floor tiles. Snow/desert sprites remain available even when a snapshot contains neither biome.

## P1c art hints and polish checks

`pkg/worldmap/art.go` derives `mapFeatures`, `artSeed`, and Lower `undergroundStyle` from existing room customization. Explicit tags/types/names/service action names precede description/detail fallbacks. Reveal omits hints for unknown rooms. The snapshot helper copies action names only, excluding scripts, parameters, and responses. Sprite source is the content repo's `generate_map_tiles.py` plus `map_hires_art.py` (row definitions remain in `map_polish_art.py`); room inputs/discovery storage do not change.

P1c capture checks all P1b interactions plus fog-sensitive cache pixels, every surface room center inside the smooth coastline, single-cell spur removal, specialized stamp kinds, bridge crossings, grouped ridges/canopies, four underground styles, torch/stair markers, and minimum zoom. Reports separate first/changed-scene bakes, marker-only replacements, and median/p95 warm overview/close/Lower draws. The viewport zoom threshold chooses the cached detailed or overview scene; far zoom uses fixed-size town glyphs. Low-rate ambient accents respect reduced motion and pause while hidden or closed.

## P1d framing, detail and bake checks

Current capture output is `worldmap-p1d-{overview,zoom,lower,far,highlands}-1920x1080.png`, plus browser and layout JSON under `.director/ux-audit/after/`. Minimum scale is 1 and shares the Fit world frame: 78% of map height, limited by available width on narrow screens. Existing local recenter and maximum scale behavior remain. Highlands capture frames Thornfield Highlands and Ironspine Foothills.

P1d traces angular walls from quarter-cell courtyard/street footprints, with corner towers and incoming road gates. Relief depends on mountain depth: low rock foothills surround taller central peaks and only the highest variant carries a cap. Disclosed zone context chooses oaks, pines or marsh dead trees; rooms add town/farm/highland/plateau props. Offshore rocks and coastal cliffs are decorative, with no room/travel IDs. Drop shadows on buildings/trees are removed.

Authoring precomposes directional dither masks into the sheet. Runtime uses a cached vector coast clip, shared shore palette tables, a bounded town/road/prop overlay cache and completed offscreen bitmaps; close detail bakes only on first close zoom. The browser report separately measures first scene, changed scene, first close draw and marker-only updates, plus 80-frame median/p95 overview/close/Lower draws. On supported browsers `worldmap-worker.js` builds scenes off the main thread; a bounded signature cache and one running/one queued job keep stale exploration work isolated. Assertions enforce P1b’s 162 ms main-thread bake benchmark and 16.7 ms warm draw budget on this review machine, and verify timers continue firing during worker work. Reports separately disclose worker CPU bake time and elapsed completion time; these are not counted as main-thread blocking. Capability or worker-load failures fall back to the synchronous painter and can take longer to bake. These remain local measurements, not guarantees for all devices.

Lower paints only short aligned known compass tunnels. Longer/misaligned/vertical links remain navigation exits without lines across void; stairs stay as room-owned entrance glyphs. Crypt bones, cellar barrels/wood, sewer channels and rough caves retain their own floors, with dim ambient cluster light. Historical P1/P1b/P1c captures describe their corresponding commits; use the P1d tool with current code.

## P1e quieter roads

Roads are a single batched network beneath terrain relief, trees, buildings and props. Town paving is fully opaque, so the dirt paths do not show through street tiles. Each discovered outdoor exit is included once and thin strokes keep overview junctions clear. Capture the four 1920×1080 overview, Oldtown/Silverbrook zoom, Lower and minimum-zoom views with `tools/capture_worldmap_p1e.cjs`; the local tool waits for fonts, art and worker scenes before saving.

## P1f deeper zoom

Scale now spans 1–10 with a 220px tile cap, twice the prior maximum; minimum zoom and continent fit remain P1d’s frame. Wheel and pinch use the shared clamp, and nearest-neighbor sampling keeps enlarged pixels crisp. Run `tools/capture_worldmap_p1f.cjs` against the local production-overlay preview to save overview and Oldtown at maximum zoom, and check wheel/pinch limits, Fit world, local recenter, pixel sampling and page errors. Evidence is under `.director/ux-audit/after/worldmap-p1f-*`.


## P1j production atlas

A production TalesMUD deploy can use the imagegen craft sheet (content version `c8169d17ef81`, `?v=worldmap-p1j`). Promote by copying `assets/map-tiles/prototypes/p1j/{terrain-sheet.png,terrain-review.png}` and a production-trimmed `terrain-sheet.json` (drop prototype-only keys) into `assets/map-tiles/`, then:

```sh
python3 tools/sync_map_tiles.py ../talesmud-rpg-1/assets/map-tiles
# bump client cache queries to ?v=worldmap-p1j
cd public/mud-client && npm run build
# copy public/ → pkg/webuiplay/dist/
```

Do not run `generate_map_tiles.py` for this promotion — that rebuilds Pillow P1g art and overwrites the imagegen atlas. P1e roads-under-stamps, P1f zoom clamps, A9, and 48px sheet dimensions are unchanged.

## P1g native art and performance comparison

Native 48px terrain, all 40 stamp rows and Lower floors retain six room-driven variants. The detailed and Lower scenes keep native resolution; the overview stays at 32px per cell after measurements showed repeated 48px overview downscales cost too much. Logical geography stays at 32 units per cell, and projection derives the displayed bitmap resolution. Coastal fields use twelve samples per cell, and every bitmap scale uses nearest-neighbor. Outdoor towers select braced wooden lookout variants; town defensive towers select stone variants. Roads remain under all stamps and opaque town paving.

Serve the current preview on :8144 and a retained P1e preview of the same 351-room fixture on :8143, then run `PUPPETEER_MODULE=/path/to/puppeteer-core node tools/capture_worldmap_p1g.cjs`. Override `WORLDMAP_PREVIEW_URL` and `WORLDMAP_BASELINE_URL` if needed. It captures `worldmap-p1g-{maxzoom,oldtown,overview}-1920x1080.png`, validates unchanged zoom clamps/fit, native detailed/Lower cells, disclosed stamp roles and coastline bounds, then records matched first/changed/close bake and 80-frame warm draw measurements. `maxzoom` frames Oldtown Road Sign, Bandit Lookout and Meadows; `oldtown` frames the town. The report documents bake, asset-size and memory deltas rather than treating off-thread bake latency as main-thread blocking. This pass is local review only, with no VPS access or deployment.
