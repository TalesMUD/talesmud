# Combat Battle Stage

**Status:** C0 locked · **C1–C3 done** · **C5–C6 done** · **C7 polish** (grid header/arena/dock/log, no absolute dock/log, log noise filter, short banner, 2026-09-07) — on engine-june  
**Follow-up:** Flutter FX (C4 parity) — not in this slice.
**Milestone:** [Combat Battle Stage](https://github.com/TalesMUD/talesmud/milestone/4)  
**Branch:** `engine-june` (do not merge to public `master` without Marcus okay)

## Backend truth

Combat is a logical `CombatInstance` with `OriginRoomID` — **not** a separate world room.
Players stay in the origin room; `InCombat` + WS combat messages drive the client.
Auto-attack + queued `attack` / `defend` / `flee` / `skill` already exist.
C1: `processAllTurns` resolves at most one combatant per tick, gated by authored `NextActionAt` / player `DecisionWindowSeconds`.

## Locked pillars

1. Enter combat → UI **takes over** (room HUD dims/hides; battle stage fills viewport)
2. Default = **auto-advance Attack** on locked target
3. Player can interrupt: Attack / Skills / Defend / Item / Flee (+ retarget)
4. Every hit/heal/miss is **visible** (HP bars, float numbers, sprite react)
5. Pacing is **authored**, not game-loop speed

## Layout (Classic — ship first)

Pokémon-style corners:
- **Enemies** upper-right strip (1–N), active target ringed, HP + name + status chips
- **Player** lower-left (portrait/bust or class sprite), HP/MP + status chips
- **Center** FX layer
- **Bottom dock** Attack · Skills · Defend · Items · Flee
- Optional thin collapsible log (not primary UX)

Later optional: **Party** mode (FF side-by-side) for large multi-combat.

### Surfaces
| Surface | Notes |
|---------|-------|
| Desktop web | Full-bleed stage over dimmed room / zone combat backdrop |
| Mobile web | Stacked: enemies top ~35% · FX · player bar · thumb dock |
| Flutter | Mirror mobile web; same WS event state machine |

## Pacing

- Actor turn beat: ~0.8–1.2s windup → resolve → ~0.4s reaction
- Player **decision window ~10s**; if no queue → auto Attack
- Queued actions apply on the player's next turn
- Duration targets (C6): trash 3–6 player turns, elites 8–15, bosses 15–30

## WS event protocol (C1)

Wire types (extend existing; terminal `message` prose kept):

| Event | Type string | Notes |
|-------|-------------|-------|
| Start | `combatStart` | enemies/players HP portraits |
| Turn | `combatTurn` | actorId, actorName, round, deadlineMs (player window) |
| Action | `combatAction` | actorId, targetId, action, result, damage, remainingHp, maxHp, fxId, combatants[] |
| End | `combatEnd` | outcome |

Conceptual aliases from the mock (`turnStart` / `actionResolved`) map to `combatTurn` / `combatAction`.
`hpDelta` / `statusApplied` / `skillFx` remain optional follow-ons; HP snapshots ride on `combatAction.combatants`.

Clients **animate from events**; they do not invent outcomes.

### C1 engine timings (`combat.CombatConfig`)

| Constant | Default | Role |
|----------|---------|------|
| `DecisionWindowSeconds` | 10 | Player turn wait; no `QueuedAction` → auto Attack |
| `TurnBeatMs` | 1000 | Authored windup / inter-turn beat |
| `ReactionMs` | 400 | Post-resolve reaction pause |
| `TurnTimeoutSeconds` | 60 | Legacy absolute turn timeout (kept; decision window is the player UX timer) |

`CombatController.processAllTurns` resolves **at most one** combatant per Update, gated by `NextActionAt` / `Phase` (`waitingPlayer` | `playingBeat` | `resolving`) on `CombatInstance`. Combat game-loop tick is **1s**.

## Ship order (one Now at a time)

| ID | Slice |
|----|-------|
| C0 | Spec + mocks (this doc) |
| C1 | Engine turn beats + structured WS events — **done on engine-june** |
| C2 | Svelte full-screen battle stage — **done on engine-june** |
| C3 | Mobile web thumb dock — **done on engine-june** |
| C4 | Flutter parity (incl. FX follow-up) |
| C5 | FX pack — **done on engine-june** (hit flash/shake, cast glow, miss puff, death dissolve, float dmg/heal; desktop+mobile) |
| C6 | Balance duration pass — **done on engine-june** |


## C6 — Balance duration pass

**Goal:** Fights last based on strength difference (not instant deletes, not endless sponges).

### Duration targets (player turns ≈ combat rounds in 1v1)

| Band | Examples | Player turns |
|------|----------|--------------|
| Trash | Catacomb Rat, Sewer Rat, Tunnel Mole | 3–6 |
| Elite | Meadow Wolf, Bandit, Thornback Bear | 8–15 |
| Boss | Burrow Brute, Hollow Knight | 15–30 |

Reference harness: **Warrior** with distributable attribute points auto-spent on primary, skills AI enabled.

### Measured (Warrior, after C6)

| Matchup | Before (approx) | After | Notes |
|---------|-----------------|-------|-------|
| L1 vs Catacomb Rat | ~7.2 rnds, 100% | **~4.3**, 100% | trash |
| L2 vs Sewer Rat | ~7.0 | **~4.1**, 100% | trash |
| L2 vs Meadow Wolf | ~13.1 | **~8.8**, 100% | elite |
| L3 vs Bandit | ~24.3 | **~12.3**, 100% | elite |
| L5 vs Thornback Bear | ~42, 40% win | **~11.6**, 100% | elite |
| L4 vs Burrow Brute | ~39, 18% win | **~20**, ~100% | boss |
| L6 vs Hollow Knight | ~34, **0%** win | **~21**, ~100% | boss (was unkillable) |

Before numbers: no attr spend, old `config/combat_balance.yaml` (boss ATK/HP too high → wipe or sponge).

### How to re-run sims

```bash
# From repo root (so config/combat_balance.yaml is found):
go test ./pkg/mudserver/game/combat/ -run TestCombatDurationTargets -v

# Broader balance matrix + CLI simulator:
go test ./pkg/mudserver/game/combat/ -run 'TestFullBalanceMatrix|TestBosses' -v
go run ./cmd/combat_simulator -class Warrior -level 6 -enemy 'Hollow Knight' -n 200
```

Tuning knobs (prefer these over rewriting `ProcessAttack`):

1. `config/combat_balance.yaml` — `difficulty_multipliers` + `named_overrides`
2. Defaults mirrored in `pkg/mudserver/game/balance/difficulty.go`
3. Content base stats in `talesmud-rpg-1` only if engine multipliers cannot separate two bosses on the same tier

`ApplyEnemyMultipliers` runs at **import** (`pkg/importer`) and in **sims** (`simutil.CreateEnemy`). Re-import NPCs after changing the YAML so the live DB picks up new finals.

### Content notes (talesmud-rpg-1)

Engine-only is enough for C6 duration bands. Optional content cleanups (not required to ship C6):

| NPC | Current | Suggested |
|-----|---------|-----------|
| Sewer Rat (`ENM0008`) | `difficulty: normal` | `easy` (trash tier) |
| Tunnel Mole (`ENM0010`) | `difficulty: normal` | `easy` (trash tier) |

Sims already treat those two as `easy`. Until content is retagged + re-imported, live DB keeps `normal` multipliers for them.

No Hollow Knight base-stat content patch needed — `named_overrides` for `"The Hollow Knight"` / `"Hollow Knight"` handle boss duration/survivability.

## Mocks

See `docs/combat-mocks/` (desktop + mobile).
