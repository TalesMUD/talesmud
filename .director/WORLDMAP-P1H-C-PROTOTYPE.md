# Worldmap P1h-C — Kenney-detail structure, Veilspan palette

Local-only prototype, 2026-09-29. No deploy or production tile/client replacement.

## Art direction and implementation

- Re-authored stamp details in content `tools/map_hires_art_p1h_c.py` from the P1g code. The reference images informed structure and density; no Kenney pixels or assets were copied.
- A one-pixel dark olive/charcoal edge is inked **inside** major stamp alpha silhouettes. This protects the 48px footprint and leaves exterior transparency clean. There is no offset drop shadow or bloom.
- Oaks have a rounded, lobed crown over a short visible trunk; pines use three stepped conical tiers. Top-left lit facets and darker right/bottom facets provide cell shading.
- Houses have framed windows, doors, chimneys, scalloped roof courses and a lit top-left ridge. Town and village stamps add compact yard details. Gatehouse, keep and tower use staggered stone courses with darker mortar and shaded right faces. Fence posts and rails, a roofed stone well with water disk, carts and barrels use stronger silhouettes and material marks.
- Marcus follow-up: the first C pass still read too bright against Veilspan room, NPC and enemy art. This pass lowers foliage midtones toward olive dusk, sets tree shade near `#1e3328` and a soft top-left highlight near `#6a8258`, and darkens plaster, timber, stone faces, mortar and roof courses. Roof base hues still derive from `map_polish_art.ROOFS`, with a darker value and softer ridge light. Right and bottom faces are deeper while top-left light remains legible. The aim is darker Veilspan lighting, not Kenney candy colors or bright cartoony overworld greens.
- Ground terrain and underground floors remain P1g; their brighter grass is the principal remaining palette gap. Roads remain below stamps; atlas layout, zoom clamps and A9 code were not edited.

## Prototype and capture

- Tile size: **48px**, six variants. One-pixel inside outlines read at maximum zoom, so 64px is unnecessary.
- Source config: engine `pkg/worldmap/map_terrain.json`.
- Generate with `python3 tools/generate_map_tiles.py --config ../talesmud-june/pkg/worldmap/map_terrain.json --out /tmp/worldmap-p1h-c-art --art-module map_hires_art_p1h_c` from the content repo. The alternate-output guard rejects a prototype module targeting production `assets/map-tiles`.
- Sheet, metadata and review: content `assets/map-tiles/prototypes/p1h-c/`.
- Local preview: `/tmp/worldmap-p1h-c-preview/`, built from `/tmp/worldmap-p1g-preview/atlas.json` and served on `127.0.0.1:8153`. Its copied `api/map-tiles/terrain-sheet.{png,json}` was refreshed.
- Puppeteer captures at 1920×1080: `.director/ux-audit/after/worldmap-p1h-c-{maxzoom,oldtown,overview}-1920x1080.png`; review plate: `.director/ux-audit/after/worldmap-p1h-c-terrain-review.png`. Capture reported no page errors, maximum tile step 99.815625px and image smoothing disabled.
- Aethermoor composites: `.director/ux-audit/after/worldmap-p1h-c-aethermoor-mock-1-oldtown-meadows-1920x1080.png` and `.director/ux-audit/after/worldmap-p1h-c-aethermoor-mock-2-overview-1920x1080.png`. These use the real atlas capture and current sheet, with a further dusk grade and labeled detail panels; the grade is a mock treatment, not a live client change.

## Comparison

| Sheet | PNG bytes | Difference from P1g | Visual emphasis |
| --- | ---: | ---: | --- |
| P1g | 471,730 | baseline | Clean stamps, lighter detailing |
| P1h-A | 476,216 | +4,486 | Material courses within P1g alpha |
| P1h-B | 488,044 | +16,314 | Rim light and underside darkening on existing pixels |
| **P1h-C darker follow-up** | **461,741** | **−9,989 (−2.12%)** | Same C silhouettes and density, lower material values, softer top-left light and deeper underside |

The new sheet is 3,086 bytes larger than prior C (458,655 bytes; version `564fb637ba15`, MD5 `0afbf807ff841cb27111f24facd28df1`). Current C is version `bca0011165f6`, MD5 `85a814d36e7ef54b8c6544a384ead994`. Sampled opaque luminance fell from 77.2 to 58.9 for canopy, 78.7 to 59.3 for oak, 121.2 to 97.4 for keep and 103.1 to 84.2 for town; the sampled stamp alpha masks are identical to prior C.

P1h-C reads most distinctly at maximum zoom. P1h-A preserves the P1g silhouette most closely; P1h-B changes lighting without redrawing forms. C uses stronger contours and simpler crowns, so repeated forest rows look more regular and adjacent town buildings can visually merge at overview scale. The village/town yard accessories are intentionally tiny and are mainly visible at close zoom. The darker stamps now have stronger contrast against unchanged grassland; a future terrain pass could lower grass selectively if review finds that gap distracting.

## Integrity

- Production P1g `assets/map-tiles/terrain-sheet.png` MD5: `41f0924505f4381cb4956f6b36962ecc`; metadata version: `39069d368e7e` (both unchanged).
- P1h-C metadata version: `bca0011165f6`; PNG MD5: `85a814d36e7ef54b8c6544a384ead994`.
- No production `terrain-sheet.{png,json}`, client map tiles or `terrainSheet.js` were changed. This prototype has not been deployed.
