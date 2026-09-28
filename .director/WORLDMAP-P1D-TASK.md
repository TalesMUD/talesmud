# WORLDMAP P1d — second polish pass (still NO deploy)

Marcus reviewed P1c (engine 7d86b3c/3995045, content ef31d0b) on 2026-09-29 00:38 Berlin: "continue polishing before patching veilspan". He loves the direction: KEEP palette, pixel style, roads, blending, sprites. Commits must contain `[grokbot]`. Do NOT touch the VPS or Door (:8010/:8020). No deploy.

## Fix (from director review of P1c screenshots)
1. **Oldtown wall is still an oval ring.** Walls should follow the actual town footprint/streets: irregular polygon with corners, towers at corners, a gatehouse where roads enter. Same for Arthdor/Kaigrath/Veridane.
2. **Minimum zoom is too far.** At min zoom the continent is a speck in empty sea. Clamp min zoom so the continent fills ~70–80% of the viewport height; world-fit button should frame the continent the same way.
3. **Stray grey diagonal lines on Lower** next to Gloomfen Depths (looks like a connector/stair line drawn across the void). Remove or render as a proper tunnel/stairway.
4. **Mountains look snowy and uniform** at overview. Vary: brown/grey rock foothills at the edges, taller peaks toward the center, snow caps only on the highest; mix hills with mountains; shadowed ridge sides.
5. **Bake time regressed** (~375 ms cold / ~200 ms changed vs P1b ~160 ms). Bring back to ≤ P1b levels (e.g. chunked/incremental baking, off-main-thread OffscreenCanvas, or cache layers) without losing detail.

## Enhance (same aesthetic)
6. More room-driven life: forests with distinct species by biome (pines in highlands, oaks in woods, dead trees in marsh), marsh with pools and reeds, plateau with mesa/cliff edges, highlands with rocky outcrops, small details (wells, carts, fences, windmill at farms, lanterns in towns).
7. Water: slight deep-sea wave texture variation, a few small islets/rocks offshore, gentle coastal cliffs where highlands meet the sea.
8. Lower: vary per dungeon type (catacombs bone/crypt tiles, sewers water channels, caves rough rock, cellars wooden floors/barrels), plus dim ambient vignette between clusters.

## Deliver
Screenshots (fonts + images loaded) in `.director/ux-audit/after/`: `worldmap-p1d-overview-1920x1080.png`, `worldmap-p1d-zoom-1920x1080.png` (Oldtown + Silverbrook Vale), `worldmap-p1d-lower-1920x1080.png`, `worldmap-p1d-far-1920x1080.png` (min zoom), plus `worldmap-p1d-highlands-1920x1080.png` (zoom on Thornfield Highlands/Ironspine Foothills mountains). Keep all tests green, report perf, commit + push both repos, append "WORLDMAP P1d" to CORE-FOCUS-PROGRESS.md.
