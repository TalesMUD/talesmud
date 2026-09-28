# WORLDMAP P1c — overview polish (still NO deploy)

Marcus reviewed P1b (engine 3be4e99 / content 6cd445b) on 2026-09-28 23:02 Berlin and chose "polish, then landmarks". Commits must contain `[grokbot]`. Do NOT touch the VPS or Door (:8010/:8020). Build on P1b; keep everything it got right (zoom view looks great: blending, roads, walled Oldtown, Silverbrook farmland, gold marker, side panel, interior list, fog, perf).

Problems visible in `.director/ux-audit/after/worldmap-p1b-overview-1920x1080.png` / `-lower-`:
- Towns read as pale flat ovals pasted onto the land at overview zoom.
- Mountain areas become grey speckle noise when zoomed out.
- Coastline is grid-stepped / spindly (thin land strips, stair-step shores).
- Lower layer still looks like P1 plain blocks.

## Done-when
1. **Coastline.** Smooth, organic shore: no single-cell spindly peninsulas, no stair-step diagonals. Use corner/edge shore autotiles or a smoothed mask (e.g. marching squares / blurred-threshold) with a sand beach band and shallow-water gradient into deep sea. The continent silhouette should look hand-drawn, with a few bays/capes, not a rectangle of pixels.
2. **Towns at overview.** Replace the pale ovals with real town art that reads at every zoom: clustered rooftops (red/brown/blue roofs), wall ring with gatehouse, keep/castle stamp for castle-type towns, and a small village cluster for hamlets. Blend town ground into surrounding terrain (no flat halo). At far zoom, fall back to a crisp town icon (like a map glyph) rather than a blurred blob.
3. **Mountains / hills.** Readable at overview zoom: grouped mountain ridges with light/shadow faces (Zelda/Warcraft II style peaks), not per-cell noise. Consider a zoom-dependent LOD: at far zoom draw larger mountain sprites or a simplified shaded ridge pattern. Same idea for dense forest (canopy clumps, not speckle).
4. **Lower layer.** Give underground its own tile set in the same style: cave floor, rock walls around rooms, dark void between clusters, torch-lit crypt/cellar/sewer variants, stair/entrance markers. No plain coloured squares.
5. Keep all P1b behaviour, tests, and perf (report the same median/p95 figures).
6. Screenshots (fonts + images loaded): `.director/ux-audit/after/worldmap-p1c-overview-1920x1080.png`, `worldmap-p1c-zoom-1920x1080.png` (Oldtown + Silverbrook Vale), `worldmap-p1c-lower-1920x1080.png`, plus `worldmap-p1c-far-1920x1080.png` (whole continent at minimum zoom). Commit + push both repos, append "WORLDMAP P1c" to CORE-FOCUS-PROGRESS.md. No deploy.
