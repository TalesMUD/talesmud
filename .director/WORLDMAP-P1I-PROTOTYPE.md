# WORLDMAP P1i — full sheet craft prototype

Local review, 2026-09-29. **No deploy.** Production remains P1g at `?v=worldmap-p1g` with content hash `39069d368e7e`.

## Art direction and judgment

P1i reauthors all 60 authored rows: 14 surface terrains, fog and sea, 40 transparent decoration kinds, and four Lower floors. Six deterministic variants remain in every row. The 128 directional blend rows are recomposed from the new ground. The generator stays at native 48px, producing one 288×9024 RGBA atlas.

Terrain now uses dusk graded palettes, denser grass and crop marks, block courses in settlements and dungeons, water ripples, shore glints, and distinct cave, crypt, cellar and sewer detail. The metadata fallback colors match those new ground colors. Stamps use plank siding, shingle courses, staggered mortar, timber and stone props, yard details, and top-left lit oak and pine volumes. The P1h-C silhouette ink function was removed; form is separated by material values and internal shading.

Visual review of the local 1920×1080 captures: compared with P1g, the ground no longer appears bright against darker buildings, and town walls, roofs, trees and the well read with more material detail at closest zoom. Compared with P1h-C, tree crowns and roofs have a softer edge and no repeated near-black rim. Oldtown remains legible amid dense overlapping stamps. The restrained Veilspan hues remain recognizable; P1i is still a colorful map at full zoom, especially grassland. Existing room placement determines the repeated tree clusters and town density, so the sheet alone cannot vary those patterns.

## Artifacts and measurement

- Source: content `tools/map_hires_art_p1i.py`; generator opt-in: `--art-module map_hires_art_p1i` with alternate-output guard.
- Atlas, manifest and review: content `assets/map-tiles/prototypes/p1i/`.
- Screenshots: `.director/ux-audit/after/worldmap-p1i-{maxzoom,oldtown,overview}-1920x1080.png`; full review plate: `worldmap-p1i-terrain-review.png` in the same folder.
- Local preview: `/tmp/worldmap-p1i-preview`, served on `127.0.0.1:8154`; capture helper: `tools/capture_worldmap_p1i.cjs`.
- P1g PNG: 471,730 bytes. P1i PNG: 551,122 bytes, **+79,392 bytes (+16.8%)**. P1i SHA-256: `60a66e52904cab549788dadf209b423eea5e6a7dec92150ba7335ab0fd545e89`; manifest version `60a66e52904c`.

The review script confirmed 16 ground/fog/sea rows, 40 transparent stamp rows, four Lower rows, all six distinct variants in ground and Lower rows, and partial alpha in every stamp cell. Every authored row differs from both P1g and P1h-C. Sample RGB means moved from P1g to P1i: grassland `(102,158,73)` to `(83,128,59)`, forest `(74,118,62)` to `(60,95,50)`, city `(164,153,121)` to `(137,128,101)`. The preview reported zero page errors and disabled image smoothing; its world-fit maximum cell step remained 99.815625px.

Production `assets/map-tiles/terrain-sheet.png` still has MD5 `41f0924505f4381cb4956f6b36962ecc`, size 471,730 bytes and version `39069d368e7e`. Client `terrainSheet.js` was restored after the isolated preview build. No client, runtime, zoom, roads, A9, VPS or Door files were changed for this prototype.

## Residuals

The local captures use the read-only preview fixture, not a live character state. The denser P1i atlas adds 79 KB; worker bake cost was not remeasured. Room labels and repeated tree clusters still come from map layout and renderer rules. Aethermoor composite mocks were not refreshed because the direct screenshots show the generated sheet on actual layout without extra staged stamp placements.
