# WORLDMAP P1i — full sheet craft rebuild (pixel-art detail, NOT outlines)

Marcus correction on P1h-C (2026-09-29 Berlin): Kenney/town-map feedback was about **pixel-art craft / detail style**, NOT outlines as the main feature.

> "Naaaa my feedback was primarily about the pixel art style not the outline. We need to update all assets for the map with improved ones."

Also earlier: colors/lighting should match darker room/NPC/enemy art; keep our palette; authored zone layout.

## Label
**P1i** — full sheet rebuild of **ALL** map assets (14 terrains + fog/sea + all transparent stamps + Lower floors), not an outline pass on P1g silhouettes.

## Style goals (from Kenney village + walled-town refs under `.director/p1i-refs/`)
- Chunky readable 3/4 stamps: wood plank siding, stone block masonry, roof tile/shingle patterns, wells, fences, denser props
- Trees with clear volume (lollipop oaks / conical pines) and soft top-left cell shading
- Inhabited density on towns
- **NO** heavy black outline treatment as the signature (subtle edge darkening OK if needed for readability)
- Darker dusk midtones matching darker room/NPC/enemy art — not bright Kenney candy

## Constraints
- clawdbot; engine `talesmud-june` `engine-june`; content `talesmud-rpg-1` `main`
- Codex preferred; no cloud agents; no Grok CLI
- **No deploy**; live stays `?v=worldmap-p1g`
- Keep 48px (or justify 64), 6 variants, roads-under-stamps, zoom clamps, A9
- `[grokbot]` commits OK for prototype/local
- Do not overwrite production `assets/map-tiles/terrain-sheet.{png,json}` or client map-tiles permanently
- Do not message Marcus

## Deliverables
1. `.director/WORLDMAP-P1I-TASK.md` (this) + progress notes in CORE-FOCUS-PROGRESS.md or P1I md
2. New art module `tools/map_hires_art_p1i.py` + regenerated sheet under `assets/map-tiles/prototypes/p1i/`
3. Screenshots: maxzoom, oldtown, overview 1920×1080 + terrain-review under `.director/ux-audit/after/worldmap-p1i-*`
4. Optionally refresh layout mocks to match new craft (not outline-heavy)

Return paths, judgment vs P1g/C, size delta, SHAs.
