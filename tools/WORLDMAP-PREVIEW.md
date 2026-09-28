# Continent map P1/P1b: local review

The in-game atlas uses `pkg/worldmap/map_terrain.json`, embedded at Go build time. The content repository's `tools/build_public_map_data.py` serves the separate public lore map and does not supply room tiles. No public lore/spoiler export was changed.

The ordered classifier checks exact roomType/areaType, specific tags, room-name words, underground/indoor tags, exact area and area words, description/detail words, legacy biome/kind, then outdoor. Unknown ground defaults to grassland; the census counts fallback separately from deliberately classified grassland. Words match whole tokens; underscores and punctuation become spaces. Edit the JSON to change rules or area defaults, then rebuild Go. No room or character schema migration is needed.

Art is hand-authored through deterministic Pillow drawing code in the content repo. There are three 32px variants for each of 14 terrains, fog/sea rows, and transparent roof/keep/cave stamps. City ground contains paving only. The sheet is 96×608 RGBA, one cache-busted PNG request through `GET/HEAD /api/map-tiles/terrain-sheet.png`. The endpoint reads the embedded play assets. Keep the normal `public/mud-client/public/` → `pkg/webuiplay/dist/` copy in local build steps.

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
python3 -m http.server 8138 --bind 127.0.0.1 --directory /tmp/worldmap-preview
```

Open `http://127.0.0.1:8138/`. This mounts the production Svelte overlay on a fully explored in-memory atlas, with the requested fog neighbors. It does not start the game, create a guest, write a database, or connect to production. The JSON includes counts and fallback IDs. `tools/worldmap_snapshot.py` exports only atlas inputs and opens SQLite with `mode=ro`.

For captures, install Puppeteer Core in a separate development environment, then set `PUPPETEER_MODULE` to its directory and `CHROMIUM_PATH` to your Chromium executable. Optionally set `WORLDMAP_CONTENT_JSON` to the census JSON for a larger-world performance check. Run `node tools/capture_worldmap_p1b.cjs` (the P1 script retains its historical 8137 capture flow) while the local preview server is up. It waits for terrain loading, `document.fonts.ready`, image decoding, and layout frames. It captures overview, Oldtown + Silverbrook zoom, and a Lower close-up; exercises exterior selection, non-clickable filler, town filtering/interior selection, visible entrance switching, instanced interior marker, recenter, world fit, phone bounds, and Escape; and records changed-scene bake, marker-only snapshot, warm draw timings, and sheet requests in `.director/ux-audit/after/`.

## Layout and rendering

Embedded `pkg/worldmap/map_layout.json` configures compact zone centers (grid units), natural filler biome, town flags, minimum separation, and coast padding. Authored room coordinates describe local zone geometry and remain unchanged in storage. Unconfigured zones attach using inter-zone compass exits; disconnected areas get a nearby shelf position. Sorted rooms/exits/zones make repeated builds deterministic. Anonymous patches and broad bridges connect all surface zones; bridges add no room exits or roads. Actual authored outdoor water/shore/field/forest/swamp cells modify the natural ground locally.

Overworld interiors project onto their graph-nearest same-area exterior anchor (lexical tie breaks), including upstairs rooms. If no exterior exists, the isolated building remains one selectable anchor. Outdoor positive Z is elevation; negative depth or subterranean context puts rooms on Lower. The full atlas still contains actual interior IDs for intel/exits/travel. The side panel lists discovered interiors only, and visible Lower entrances do not disclose unexplored target names. Hidden entrances require exit revelation. Filler `{x,y,terrain}` cells are anonymous, non-clickable, and fogged until nearby ground or their surface area is explored.

The shared renderer caches native-resolution landscape pixels: eight-neighbor ordered-dither terrain transitions, masked coastal shallows, town paving/walls, transparent roof/keep/cave stamps, and dirt paths along real charted outdoor compass exits. Pan/zoom reuse the canvas. Marker-only atlas replacements reuse a bounded two-scene cache; all fog/art inputs participate in its key. Lower/Upper retain room tiles and short same-zone connectors. World fit uses continent bounds, while minimap/recenter frame exterior groups. Stored coordinates, exits, and gameplay travel are unchanged.

## Boundaries of this pass

Local read-only SQLite and canonical YAML snapshots are used, never VPS data or live player sessions. The decorative continent is schematic: compact offsets resolve inconsistent authored coordinate origins and exit cycles, and dirt roads can cross the simplified natural ground. Terrain/layer/name inference can misinterpret evocative prose; prefer structured tags/coordinates or adjust mapping defaults. A changed explored scene requires a new bitmap bake; warm pan/zoom and marker-only snapshots reuse pixels. The browser report separates these costs rather than treating warm draw time as a device-independent frame-rate guarantee.

Additional POI art, ornate zone banners, compass/legend ornaments, and VPS deployment remain outside this pass. P1b resolves P1's separate-island packing, missing continent/coast/road blending, and overworld interior floor tiles. Snow/desert sprites remain available even when a snapshot contains neither biome.
