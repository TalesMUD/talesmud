# WORLDMAP P1 — "WoW meets old-school RPG" world map, pass 1 (terrain tiles)

Requested by Marcus 2026-09-28 20:32 Berlin. Commits must contain `[grokbot]`. Branch `engine-june` (+ content repo `~/dev/talesmud-rpg-1` for art). Door (:8020, /home/atla/dev/aethermoor-door-mud) is never touched.

## Goal
The full world map overlay (`public/mud-client/src/game/ui/MapOverviewOverlay.svelte`, data from `tools/build_public_map_data.py` / map overview store) currently draws each room as a small square with a generic icon on flat dark zone boxes. Marcus wants it to look like a painted old-school RPG overworld (Zelda/Dragon Quest/Warcraft II style overworld maps, the "Aetheria overworld" look): each room tile shows the TERRAIN of that room (forest, grassland, farmland, city/street, castle/keep, dungeon/cave, swamp/bog, mountain, snow, desert/sand, water/shore, ruins, interior/tavern), tiles blend into a continuous landscape per zone, and later passes add POI artwork overlays (castle, farmhouse, cave mouth, tower) and zone banners.

## This pass (P1) — terrain tiles only
1. **Terrain classification.** Derive a terrain key for every room from existing data (roomType, areaType, area, tags, biome/kind already in the map data, room name/description keywords as fallback). Put the mapping in one data file (e.g. `config/map_terrain.yaml` or a JSON next to the map data) with a documented default. Report the resulting counts per terrain across the live world; aim for <5% falling back to default.
2. **Terrain tile art.** Generate a consistent pixel-art top-down tile set (one base tile per terrain, 2–3 variants each to avoid repetition, plus transparent edge/shore blends if feasible) in the style of classic 16-bit overworld maps: bright readable colors, top-down, slight 3/4 trees and roofs, no text. Use the existing Gemini pipeline in `~/dev/talesmud-rpg-1/tools` (see generate_images.py / asset_pipeline) or hand-authored pixel tiles; keep them small (e.g. 32 or 48 px, crisp `image-rendering: pixelated`). Serve like other art (`/api/...` static) and cache-bust.
3. **Rendering.** In the overview map, draw each room as its terrain tile instead of the square+icon. Neighbouring rooms in a zone touch so the zone reads as one landmass (fill small gaps between adjacent rooms with the dominant terrain of the pair, or draw connecting path strips along exits). Keep: current-room marker (make it a clear "you are here" glow), selected-room side panel, zone labels, click/select, zoom/pan, fog of unexplored rooms (unexplored = dimmed/parchment, not revealed art). The Minimap widget should use the same tiles if cheap; otherwise leave it.
4. **Background.** Replace the flat dark void between zones with a subtle painted backdrop (deep sea or dark parchment texture). Full framing is P3; keep it simple here.
5. **Performance.** The map must stay smooth with the full world (several hundred rooms): prefer a canvas or a single sprite-sheet over hundreds of separate <img> requests.

## Do NOT deploy in P1
Build and test locally only (local server on a spare port, not :8010). Marcus wants to see a mockup first.
- Screenshots: `.director/ux-audit/after/worldmap-p1-overview-1920x1080.png` (zoomed out, several zones visible incl. Oldtown / Silverbrook Vale / Ashenvale Woods), `worldmap-p1-zoom-1920x1080.png` (zoomed in on Oldtown + Silverbrook Vale), and the terrain sheet `worldmap-p1-tiles.png`. Wait for `document.fonts.ready` and images loaded before capture.
- Commit to engine-june / content repo, push, but do NOT touch the VPS.
- Append a "WORLDMAP P1" section to `.director/CORE-FOCUS-PROGRESS.md` (SHAs, terrain counts, screenshots, residuals).

## Later passes (not now)
P2: POI artwork overlays (castle for Oldtown keep, farmhouse, cave mouths, towers, docks), zone name banners in an ornate plate. P3: world framing (sea/coastline around landmasses, roads along exits, compass, legend), then deploy.
