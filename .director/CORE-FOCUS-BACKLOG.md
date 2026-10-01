# Core Focus Backlog (Marcus, 2026-09-24)

For the coming days the daily director works ONLY on two tracks, alternating:
1. Player-vs-Enemy combat — BattleStage screens, UX, balancing, rewards. Fun and rewarding; bosses and higher-level enemies genuinely challenging, beatable with good gear.
2. Browser UX — grid/widget layout system, polish, features.

Parked: instant dungeons, more party features, new zones/content, Door.

## Track 1 — first topic: Level gap matters (challenge + rewards)
Finding: level barely affects combat. Damage = ATK - DEF/2 (engine.go CalculateDamage), NPC mods = level/4, XP = 15*enemyLevel+5 regardless of player level.
- A1 Level-gap combat math: hit/crit/damage modifiers from level difference, capped; gear can close ~2-3 levels.
- A2 Threat UI: con colors (grey/green/yellow/orange/red/skull) on enemy nameplates in room + BattleStage, "much stronger" warning before engaging.
- A3 Reward scaling: XP/gold scaled by gap (grey trickle, red bonus), boss first-kill bonus; show breakdown on victory screen.
- A4 Sim harness targets in simutil: win-rate table by gap (-3..+5) for solo L-appropriate gear vs good gear; tune combat_balance.yaml.
Next topics: boss mechanics (phases/telegraphs/enrage), victory screen & loot reveal, hit feedback juice.

## Track 2 — first topic: Layout system that just works
- B1 Viewport audit: screenshots at 1366x768, 1920x1080, 2560x1440, 3440x1440 + mobile; list overlaps, dead space, clipped widgets.
- B2 Smart defaults: viewport-aware default presets (auto-picked on first load, re-pickable), no widget ever off-screen.
- B3 Edit-mode ergonomics: clear resize handles, snap/ghost preview, undo/reset, lock layout.
- B4 Unified widget chrome: consistent header (title, collapse, maximize/pop-out), consistent padding/typography.
Next topics: keyboard shortcuts, contextual widgets (combat auto-focus), settings panel, performance.

## Queued: A5 — Class balance pass (Marcus, 2026-09-24)
From the A4 gap matrix (`docs/COMBAT-BALANCE.md`):
- Warriors beat same-level bosses too easily (~75%, ceiling 65%).
- Rogues are too weak against good-gear +3 bosses (~21%, target ~50%).
- Cloth mages lose most elite and boss fights; they miss the melee bands.
Goal: every class lands in the same win-rate bands per gap and gear tier. Next combat-track slice after B4.

## Queued: B13 — Combat contextual layout (2026-09-28 director)
Backlog next UX topic after B9 shortcuts: contextual widgets (combat auto-focus). When combat starts, focus BattleStage; on end, restore. See CORE-FOCUS-B13-TASK.md.

## Queued: B14 — Settings panel (2026-09-30 director)
Backlog next UX topic after B13: settings panel players can trust (reduced motion, combat auto-focus, inventory mode, gold chrome). See CORE-FOCUS-B14-TASK.md.

## Done (recent)
- A9 boss phases — accepted 2026-09-30 (code `9026bba`, smoke `152eadd`; live under later tip `?v=worldmap-p1j`).
- WORLDMAP P1e–P1j — accepted 2026-09-30 as live EXTRA (`?v=worldmap-p1j` Kenney-inspired craft sheet).
- B13 combat contextual layout — accepted 2026-09-29.
- WORLDMAP P1–P1d — accepted 2026-09-29.

## Queued: A10 — Focus target + threat re-warn (2026-10-01 director)
Combat track after B14. Closes A2 residual (no re-warn on mid-fight target switch). See CORE-FOCUS-A10-TASK.md.

## Done (2026-10-01 accept)
- B14 Settings panel — accepted 2026-10-01.
- Overnight UX EXTRA (buffport1) — accepted 2026-10-01.
