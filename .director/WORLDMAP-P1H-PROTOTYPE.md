# World map P1h: local stamp prototypes

September 29, 2026. Local review only. The content sources are `tools/map_hires_art_p1h_a.py` and `tools/map_hires_art_p1h_b.py`; `generate_map_tiles.py --art-module` selects either while requiring an alternate output directory. Packaged sheets and review plates are in content `assets/map-tiles/prototypes/p1h-{a,b}/`.

| Sheet | PNG bytes | Delta vs P1g | Version | Direction |
| --- | ---: | ---: | --- | --- |
| P1g | 471,730 | — | `39069d368e7e` | Production baseline |
| P1h-A | 476,216 | +4,486 (+0.95%) | `7917352f32ed` | Denser tile courses, staggered mortar, bark and stump grain, leaf and needle marks, plank/rail marks |
| P1h-B | 488,044 | +16,314 (+3.46%) | `3460c8fbc7f5` | One-pixel warm lit edges and darker internal undersides; cooler stone and warmer timber |

Both sheets remain 288×9024 RGBA, 48px cells, six variants and the same row, blend and floor metadata as P1g. Their alpha channels match P1g byte for byte. Terrain, floor and blend generation are unchanged. A retains P1g's outlines by masking its added detail to the original stamp alpha. B shades only existing opaque pixels, with no external shadow or bloom.

At maximum zoom on this atlas fixture (10× fit scale, about 99.8px per cell; 220px cap unchanged), **A is the stronger candidate**. The roofs, lookout deck and bridge read as distinct materials while the familiar silhouette survives. B gives oak crowns and building edges modest extra depth and preserves the outline perfectly, but its effect is subtler in the full map; the extra edge treatment also costs more PNG bytes. At fit overview, both stay close to P1g. A's finer marks are best judged in the max zoom and review plate.

Captures at 1920×1080 are in `.director/ux-audit/after/`: `worldmap-p1h-{a,b}-{maxzoom,oldtown,overview}-1920x1080.png`. They came from isolated local previews on ports 8151 and 8152 using the existing P1g atlas fixture and a replacement sheet in each preview's own `/api/map-tiles/` folder. The browser reported no page errors and nearest-neighbor map draws. Bake and warm draw timings were not sampled; row count, source dimensions and runtime code are unchanged, so the remaining runtime risk is the larger PNG transfer/decode, especially for B.

Production verification: content P1g PNG MD5 `41f0924505f4381cb4956f6b36962ecc` and metadata version `39069d368e7e`; engine `terrainSheet.js` SHA-256 `7d68829ece31540202fa161c3d3a06ff71028b0be1a37cb326062dfde78d945c`. The production client files have no tracked diff and its worker query remains `worldmap-p1g`. No sync, VPS, Door or deployment was used.
