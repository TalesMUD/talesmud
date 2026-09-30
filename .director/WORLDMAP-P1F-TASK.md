# WORLDMAP P1f — deeper zoom (DEPLOY after screenshots)

Marcus 2026-09-29 12:22 Berlin on live `?v=worldmap-p1e`: wants to zoom in at least 50–100% further. KEEP aesthetic. Commits `[grokbot]`.

## Done-when
1. Raise max zoom so closest view is ~1.5×–2× closer than current max (pixel tiles larger / more detail). Keep min zoom / continent-fit as P1d. Wheel + pinch + buttons if any all respect the new clamp.
2. Nearest-neighbor / crisp pixels at high zoom (no blurry upscale).
3. Screenshots: `.director/ux-audit/after/worldmap-p1f-maxzoom-1920x1080.png` (Oldtown at new max), `worldmap-p1f-overview-1920x1080.png`. Append WORLDMAP P1f to CORE-FOCUS-PROGRESS.md. Commit + push.
4. Then DEPLOY to VPS (same rules as P1e): build client+binary with new `?v=worldmap-p1f`, backup bin/tales.prev-<sha>, swap on deploy-host via deploy SSH, SIGTERM MainPID of talesmud.service, never touch Door. Verify play serves new ?v=.
