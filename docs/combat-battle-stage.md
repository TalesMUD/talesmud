# Combat Battle Stage

**Status:** C0 locked · **C1 done** · **C2 done** · **C3 done** (mobile thumb dock, 2026-09-06) — on engine-june  
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
| C4 | Flutter parity |
| C5 | FX pack |
| C6 | Balance duration pass |

## Mocks

See `docs/combat-mocks/` (desktop + mobile).
