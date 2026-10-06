# WORLDMAP P1g — higher-res polished map assets (NO deploy until Marcus approves)

Marcus 2026-09-29 13:31 Berlin on live max-zoom (`?v=worldmap-p1f`): likes direction, wants all map assets more polished and a bit higher-res. KEEP palette / style / roads-under-stamps / zoom clamps. Commits `[grokbot]`. No VPS/Door.

## Done-when
1. **Raise tile/stamp resolution** from current 32px (or whatever the sheet uses) to at least **48px or 64px** per cell for terrain + prop stamps, with matching sheet + runtime sync. Crisp nearest-neighbor only.
2. **Polish every visible stamp** at max zoom: trees (species, bark, canopy volume), bandit/lookout towers, gates, bridges/creek crossings, roofs/walls, rocks, stumps, flowers, farmland, windmills, mountain peaks, water foam — denser shading, not chunky geometric blobs.
3. Room-driven variants still work. Perf: keep bake/draw within ~P1e budgets or document deltas.
4. Screenshots: `.director/ux-audit/after/worldmap-p1g-maxzoom-1920x1080.png` (same Oldtown Road / Bandit Lookout / Meadows area Marcus showed), `worldmap-p1g-oldtown-1920x1080.png`, `worldmap-p1g-overview-1920x1080.png`. Append WORLDMAP P1g to CORE-FOCUS-PROGRESS.md. Commit + push both repos. **No deploy.**
