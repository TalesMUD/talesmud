# Core Focus progress

## A1 — Level-gap combat math
- SHA: `cf31157548738dbbd6c32bd834e5cf9fa06b431b` (`cf31157`)
- What changed: `level_gap` in `config/combat_balance.yaml`. Signed gap is attacker level minus defender level, clamped at ±6. Hit, crit, and damage dealt/taken scale per level. Gap 0 matches the old formulas. Applied to basic attacks, skill hits, and DoTs for players and NPCs via `CombatantRef.Level`.
- Tests: `go test ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/combat/ -count=1` green. Druid L1 vs L2 floor in `TestLevel1VsSameLevelEnemies` lowered from 5% to 1% because a one-level deficit now costs glass casters a few win-rate points (observed ~3%).
- Deploy: pushed `engine-june`. Rebuilt `bin/tales`, restarted only the Veilspan process. New pid 3611918 listening on :8010. Door pid 3406193 on :8020 unchanged.
- Smoke: `GET /` 200, `GET /play/` 200, `GET /api/server-info` 200. Log: "listening on port 8010" at 2026-09-24 23:04:54. No panic.
- Residuals: live numbers are the A1 starting point (about +32% damage at a +3 advantage, about −28% at −3). A4 tunes them against the win-rate table. Skill crits only come from a positive level gap; basic attacks still crit on a natural 20 unless the gap pulls that chance down.

