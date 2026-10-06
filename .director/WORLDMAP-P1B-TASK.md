# WORLDMAP P1b — continent, blending, towns (still NO deploy)

Marcus reviewed P1 (688b6a0 / content 1177a4c) on 2026-09-28 21:19 Berlin and approved this follow-up. Commits must contain `[grokbot]`. Do NOT touch the VPS or Door.

P1 problem: map still reads as a checkerboard of separate squares; each zone is a tiny island in a big empty sea; interior rooms show as brown crate tiles. Target look: one painted overworld continent (Zelda / Dragon Quest / Warcraft II / "Aetheria overworld").

## Done-when
1. **One continent.** Lay the zones out close together on a shared landmass (use existing zone/world coordinates and inter-zone exits for relative placement; add a small layout config for zone offsets if needed). Fill the land between zones with generated filler terrain (grassland/forest/hills, following the neighbouring zones' biomes) so there is continuous land with a coastline around the whole continent and sea only at the edges (plus a few inlets/lakes). Filler is decorative only, not clickable rooms.
2. **Blending.** No hard square edges: autotile transitions between terrains (grass↔forest, grass↔sand↔water shore, grass↔swamp, etc.) via edge/corner blend tiles or a canvas soft-mask/dither pass. Roads: draw dirt paths along exits between outdoor rooms.
3. **Towns.** Interior rooms (shops, taverns, houses, halls) are not drawn as their own tiles on the overworld. A town/city zone renders as one cluster of streets + rooftops + walls (castle tiles for keeps), with interior rooms still selectable via the town tile they belong to (side panel lists them). Dungeons: overworld shows a cave/dungeon entrance tile; underground rooms stay on the Lower layer.
4. Keep: you-are-here glow, select + side panel, zone labels, zoom/pan, fog for unexplored, performance (canvas / sprite sheet).
5. Screenshots (fonts + images loaded): `.director/ux-audit/after/worldmap-p1b-overview-1920x1080.png`, `worldmap-p1b-zoom-1920x1080.png` (Oldtown + Silverbrook Vale), `worldmap-p1b-lower-1920x1080.png` (Lower layer). Commit + push, append "WORLDMAP P1b" to CORE-FOCUS-PROGRESS.md. No deploy.
