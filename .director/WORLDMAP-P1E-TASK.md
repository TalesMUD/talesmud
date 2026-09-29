# WORLDMAP P1e — quieter roads (still NO deploy until Marcus says)

Marcus 2026-09-29 10:23 Berlin on live `?v=worldmap-p1d`: map feels noisy, especially overlapping roads. Suggestion: draw roads below houses/trees. KEEP current aesthetic. Commits `[grokbot]`. No VPS/Door.

## Done-when
1. **Layer order.** Paint dirt roads / paths UNDER buildings, walls, rooftops, tree canopies, mountain stamps, and props (windmills, wells, lanterns). Terrain and coast still under roads. Marker/you-are-here and labels stay on top.
2. **Quieter junctions.** Where many exits meet, merge into a single path network (no stacked overlapping strokes). Soften or hide roads under dense forest canopy and inside town paving (town streets use paving tiles, not extra dirt overlays).
3. **Thinner/cleaner strokes** at overview zoom so far-zoom doesn't look like spaghetti; keep readable at close zoom.
4. Screenshots: `.director/ux-audit/after/worldmap-p1e-{overview,zoom,lower,far}-1920x1080.png`. Append "WORLDMAP P1e" to CORE-FOCUS-PROGRESS.md. Commit + push. No deploy.
