# Core Focus progress

## A1 — Level-gap combat math
- SHA: `cf31157548738dbbd6c32bd834e5cf9fa06b431b` (`cf31157`)
- What changed: `level_gap` in `config/combat_balance.yaml`. Signed gap is attacker level minus defender level, clamped at ±6. Hit, crit, and damage dealt/taken scale per level. Gap 0 matches the old formulas. Applied to basic attacks, skill hits, and DoTs for players and NPCs via `CombatantRef.Level`.
- Tests: `go test ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/combat/ -count=1` green. Druid L1 vs L2 floor in `TestLevel1VsSameLevelEnemies` lowered from 5% to 1% because a one-level deficit now costs glass casters a few win-rate points (observed ~3%).
- Deploy: pushed `engine-june`. Rebuilt `bin/tales`, restarted only the Veilspan process. New pid 3611918 listening on :8010. Door pid 3406193 on :8020 unchanged.
- Smoke: `GET /` 200, `GET /play/` 200, `GET /api/server-info` 200. Log: "listening on port 8010" at 2026-09-24 23:04:54. No panic.
- Residuals: live numbers are the A1 starting point (about +32% damage at a +3 advantage, about −28% at −3). A4 tunes them against the win-rate table. Skill crits only come from a positive level gap; basic attacks still crit on a natural 20 unless the gap pulls that chance down.

## A2 — Threat display
- SHA: `b4072f382c93a01a683b509d37c925181c1088bb` (`b4072f3`)
- What changed: `threat` cutoffs in `config/combat_balance.yaml` (enemy level − player level). ≤ −3 grey, −2..−1 green, 0..+1 yellow, +2 orange, +3..+4 red, ≥ +5 skull. Viewer level below 1 counts as 1. Room NPC payload and per-viewer combat enemy views include `threat`. Entity cards and BattleStage nameplates use the color; skull adds ☠. First `attack` on orange/red/skull warns and does not engage. `attack!` / `a!` / `hit!` or a second `attack` on that enemy does. The room Attack button confirms, then sends `attack!`.
- Tests: `go test ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/commands/ ./pkg/mudserver/game/` green. New cases: tier table, skull warn-then-engage, `attack!` skips the warn, yellow does not warn.
- Deploy: pushed `engine-june`. Rebuilt client (`?v=a2threat`), copied into `pkg/webuiplay/dist`, rebuilt `bin/tales`, restarted only Veilspan. New pid 3614695 on :8010. Door pid 3406193 on :8020 unchanged.
- Smoke: `GET /` 200, `GET /play/` 200, page references `bundle.js?v=a2threat`, that bundle is 200 and contains "much stronger", `extra.css` contains `threat-skull`. Log listening on 8010 at 23:16:52. No browser session drove a logged-in room card or BattleStage nameplate; server warn/engage is covered by the command tests.
- Residuals: in-combat target switches do not re-warn. Swarm warning uses the named target only.

