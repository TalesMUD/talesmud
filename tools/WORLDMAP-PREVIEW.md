# Terrain map P1: local review

The in-game atlas uses `pkg/worldmap/map_terrain.json`, embedded at Go build time. The content repository's `tools/build_public_map_data.py` serves the separate public lore map and does not supply room tiles. No public lore/spoiler export was changed.

The ordered classifier checks exact roomType/areaType, specific tags, room-name words, underground/indoor tags, exact area and area words, description/detail words, legacy biome/kind, then outdoor. Unknown ground defaults to grassland; the census counts fallback separately from deliberately classified grassland. Words match whole tokens; underscores and punctuation become spaces. Edit the JSON to change rules or area defaults, then rebuild Go. No room or character schema migration is needed.

Art is hand-authored through deterministic Pillow drawing code in the content repo. There are three 32px variants for each of 14 terrains, plus fog and sea rows. The sheet is 96×512, one cache-busted PNG request through `GET/HEAD /api/map-tiles/terrain-sheet.png`. The endpoint reads the embedded play assets. Keep the normal `public/mud-client/public/` → `pkg/webuiplay/dist/` copy in local build steps.

## Rebuild art

From the engine repository:

```sh
python3 ../talesmud-rpg-1/tools/generate_map_tiles.py --config pkg/worldmap/map_terrain.json
python3 tools/sync_map_tiles.py ../talesmud-rpg-1/assets/map-tiles
cd public/mud-client && npm run build
```

Pillow is needed only to rebuild art. The sync helper validates row order, default terrain, and the sheet content hash before writing the renderer manifest.

## Read-only room census and production-overlay preview

```sh
python3 tools/worldmap_snapshot.py --database talesmud.db --out /tmp/worldmap-rooms.json
# Or, for the larger canonical content export (needs PyYAML):
python3 tools/worldmap_snapshot.py --content-rooms ../talesmud-rpg-1/data/rooms --out /tmp/worldmap-rooms.json
go run ./cmd/map-preview -rooms /tmp/worldmap-rooms.json -current R0201 -fog R0416,R0417 > /tmp/worldmap-preview.json
node tools/preview_worldmap.mjs /tmp/worldmap-preview.json /tmp/worldmap-preview
python3 -m http.server 8137 --bind 127.0.0.1 --directory /tmp/worldmap-preview
```

Open `http://127.0.0.1:8137/`. This mounts the production Svelte overlay on a fully explored in-memory atlas, with the requested fog neighbors. It does not start the game, create a guest, write a database, or connect to production. The JSON includes counts and fallback IDs. `tools/worldmap_snapshot.py` exports only atlas inputs and opens SQLite with `mode=ro`.

For captures, install Puppeteer Core in a separate development environment, then set `PUPPETEER_MODULE` to its directory and `CHROMIUM_PATH` to your Chromium executable. Optionally set `WORLDMAP_CONTENT_JSON` to the census JSON for a larger-world performance check. Run `node tools/capture_worldmap.cjs` while the local preview server is up. It waits for terrain loading, `document.fonts.ready`, image decoding, and layout frames. It exercises tile selection, zoom/pan, recenter, world fit, layer switching, phone bounds, and Escape, and records draw timings and sheet requests in `.director/ux-audit/after/`.

## Boundaries of this pass

Tiles touch at grid boundaries. Short (one or two missing cells) compass-exit gaps within a zone receive terrain, provided both ends are charted and on the same layer. Occupied cells, fog, hidden passages, and inter-zone gaps are never filled. Existing authored coordinates and atlas layout are retained. The classifier can misinterpret evocative prose; prefer structured terrain fields/tags or adjust the mapping. Snow and desert art exists even when a source snapshot has no rooms in those terrains.

POI overlays, zone plates, blended alpha edges, external coastlines, and road/compass/legend framing are deferred to later map passes. The existing layout compiler's map iteration can vary zone packing between compilations, and some content zones have overlapping authored coordinate extents. These are existing layout limitations, not changed by terrain rendering. Screenshot framing is calculated from the loaded atlas rather than hardcoded geography.
