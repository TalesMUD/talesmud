# Combat balance — level gap table

Player level is 10. Gap is enemy level minus player level. Trash / elite / boss use the level-scaled bodies in `CreateScaledEnemy` (easy / hard / boss). Appropriate gear is class starter plus a small per-level bump. Good gear is about 2.2× that. Hit, crit, and the level gap come from `config/combat_balance.yaml` `level_gap`. `class_balance` then scales each class's damage dealt and taken.

Targets: level-appropriate trash at gap 0 wins about 95%+, elites 75–90%, bosses 50–65%. Bosses at +3 with good gear about 50%. +5 with appropriate gear stays under 15%.

## Warrior — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 71% | 54% |
| trash | rounds | 4.8 | 5.0 | 5.8 | 7.0 | 7.0 | 7.2 | 8.6 | 9.0 | 10.1 |
| trash | HP left | 98% | 98% | 95% | 90% | 85% | 83% | 67% | 55% | 45% |
| elite | win% | 100% | 100% | 100% | 96% | 88% | 71% | 12% | 0% | 0% |
| elite | rounds | 11.0 | 12.4 | 14.3 | 17.4 | 18.1 | 19.8 | 17.5 | 15.4 | 9.3 |
| elite | HP left | 94% | 90% | 85% | 70% | 44% | 37% | 38% | — | — |
| boss | win% | 100% | 100% | 83% | 50% | 21% | 0% | 0% | 0% | 0% |
| boss | rounds | 21.2 | 25.8 | 28.8 | 28.2 | 21.6 | 21.3 | 12.7 | 10.8 | 9.9 |
| boss | HP left | 85% | 71% | 51% | 41% | 28% | — | — | — | — |

## Warrior — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% |
| trash | rounds | 4.0 | 5.0 | 5.1 | 5.3 | 5.2 | 6.7 | 7.5 | 8.5 | 9.7 |
| trash | HP left | 99% | 98% | 98% | 96% | 97% | 93% | 93% | 90% | 83% |
| elite | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 88% | 79% |
| elite | rounds | 8.7 | 11.0 | 11.6 | 13.8 | 14.5 | 16.3 | 21.4 | 25.1 | 30.0 |
| elite | HP left | 97% | 94% | 84% | 86% | 82% | 75% | 74% | 58% | 65% |
| boss | win% | 100% | 100% | 100% | 96% | 88% | 71% | 58% | 4% | 0% |
| boss | rounds | 17.2 | 19.9 | 23.2 | 28.8 | 29.9 | 32.9 | 41.1 | 55.4 | 45.2 |
| boss | HP left | 95% | 82% | 80% | 61% | 50% | 49% | 35% | 8% | — |

## Rogue — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 96% | 100% | 100% | 100% | 100% | 62% |
| trash | rounds | 4.0 | 4.3 | 4.1 | 5.0 | 3.0 | 3.1 | 3.1 | 4.4 | 3.5 |
| trash | HP left | 97% | 96% | 94% | 79% | 88% | 78% | 83% | 61% | 83% |
| elite | win% | 100% | 96% | 100% | 79% | 92% | 83% | 38% | 21% | 8% |
| elite | rounds | 7.9 | 8.7 | 9.9 | 11.5 | 5.8 | 6.9 | 5.8 | 5.5 | 5.1 |
| elite | HP left | 93% | 87% | 90% | 70% | 80% | 58% | 60% | 54% | 38% |
| boss | win% | 100% | 96% | 75% | 62% | 50% | 21% | 25% | 0% | 0% |
| boss | rounds | 14.3 | 16.0 | 16.6 | 18.5 | 9.0 | 9.2 | 8.1 | 5.4 | 5.5 |
| boss | HP left | 73% | 56% | 64% | 28% | 78% | 64% | 79% | — | — |

## Rogue — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% |
| trash | rounds | 3.0 | 3.8 | 4.0 | 3.8 | 3.0 | 3.1 | 3.1 | 3.2 | 4.8 |
| trash | HP left | 99% | 98% | 96% | 91% | 95% | 98% | 95% | 97% | 79% |
| elite | win% | 100% | 100% | 100% | 92% | 100% | 100% | 100% | 92% | 88% |
| elite | rounds | 6.5 | 7.9 | 8.0 | 8.8 | 6.0 | 5.9 | 6.8 | 8.5 | 10.8 |
| elite | HP left | 94% | 89% | 82% | 83% | 88% | 83% | 90% | 76% | 58% |
| boss | win% | 100% | 92% | 88% | 79% | 96% | 83% | 46% | 33% | 12% |
| boss | rounds | 11.8 | 13.0 | 15.3 | 16.4 | 9.3 | 11.7 | 9.9 | 11.7 | 10.5 |
| boss | HP left | 84% | 73% | 73% | 66% | 79% | 53% | 100% | 100% | 100% |

## Ranger — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 88% | 79% | 42% | 12% |
| trash | rounds | 4.0 | 4.0 | 5.0 | 5.0 | 6.0 | 6.8 | 6.9 | 6.7 | 7.0 |
| trash | HP left | 99% | 96% | 93% | 95% | 79% | 67% | 54% | 49% | 61% |
| elite | win% | 100% | 100% | 92% | 92% | 71% | 21% | 0% | 0% | 0% |
| elite | rounds | 7.8 | 9.1 | 9.8 | 13.3 | 14.2 | 10.7 | 9.1 | 9.5 | 7.7 |
| elite | HP left | 92% | 88% | 84% | 53% | 30% | 57% | — | — | — |
| boss | win% | 92% | 92% | 67% | 25% | 12% | 4% | 0% | 0% | 0% |
| boss | rounds | 14.5 | 19.2 | 20.5 | 16.7 | 15.4 | 14.5 | 10.5 | 5.4 | 5.0 |
| boss | HP left | 79% | 62% | 47% | 58% | 32% | 8% | — | — | — |

## Ranger — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 88% | 96% |
| trash | rounds | 3.0 | 3.9 | 4.0 | 5.2 | 5.2 | 5.7 | 7.2 | 7.3 | 8.8 |
| trash | HP left | 100% | 98% | 97% | 93% | 93% | 86% | 86% | 73% | 78% |
| elite | win% | 100% | 100% | 100% | 100% | 100% | 88% | 79% | 75% | 67% |
| elite | rounds | 6.5 | 8.1 | 9.1 | 10.7 | 12.7 | 14.2 | 15.4 | 19.7 | 24.8 |
| elite | HP left | 95% | 87% | 84% | 83% | 81% | 67% | 68% | 70% | 63% |
| boss | win% | 100% | 96% | 96% | 75% | 50% | 67% | 33% | 21% | 4% |
| boss | rounds | 13.2 | 14.1 | 17.6 | 18.5 | 20.1 | 27.7 | 24.0 | 31.4 | 46.9 |
| boss | HP left | 89% | 77% | 64% | 69% | 58% | 59% | 50% | 47% | 1% |

## Mage — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 88% | 92% | 75% | 12% |
| trash | rounds | 3.0 | 3.0 | 3.0 | 3.0 | 3.1 | 4.1 | 4.1 | 4.2 | 3.5 |
| trash | HP left | 91% | 86% | 86% | 75% | 72% | 65% | 48% | 32% | 39% |
| elite | win% | 100% | 100% | 100% | 92% | 79% | 71% | 8% | 4% | 0% |
| elite | rounds | 4.0 | 4.9 | 5.0 | 5.9 | 5.8 | 6.1 | 4.8 | 4.6 | 5.0 |
| elite | HP left | 86% | 74% | 74% | 44% | 48% | 32% | 27% | 23% | — |
| boss | win% | 100% | 88% | 67% | 58% | 17% | 0% | 0% | 0% | 0% |
| boss | rounds | 6.9 | 6.8 | 7.2 | 8.6 | 6.8 | 6.5 | 6.7 | 3.3 | 2.9 |
| boss | HP left | 75% | 57% | 35% | 34% | 36% | — | — | — | — |

## Mage — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 88% | 79% |
| trash | rounds | 3.0 | 3.0 | 3.0 | 3.0 | 3.2 | 4.1 | 4.2 | 4.6 | 5.7 |
| trash | HP left | 99% | 99% | 95% | 98% | 96% | 85% | 88% | 71% | 52% |
| elite | win% | 100% | 100% | 100% | 100% | 96% | 96% | 88% | 75% | 33% |
| elite | rounds | 4.0 | 4.9 | 5.0 | 6.0 | 6.0 | 6.3 | 7.3 | 8.1 | 7.8 |
| elite | HP left | 96% | 95% | 98% | 79% | 90% | 70% | 73% | 64% | 72% |
| boss | win% | 100% | 100% | 96% | 88% | 88% | 71% | 50% | 12% | 0% |
| boss | rounds | 6.8 | 6.9 | 8.0 | 8.5 | 10.1 | 11.2 | 10.4 | 9.8 | 8.1 |
| boss | HP left | 87% | 93% | 91% | 84% | 88% | 66% | 61% | 28% | — |

## How to read this sample

One 24-iteration run per cell (player level 10). Win% moves several points between runs. An 80-iteration check of the same config is the steadier read of the three cells A4 left open:

| Cell | A4 sample | 80-iteration check |
| --- | --- | --- |
| Warrior, appropriate gear, boss, gap 0 | 75% | 59% |
| Rogue, good gear, boss, gap +3 | 21% | 51% |
| Mage, appropriate gear, elite, gap 0 | 8% | 74% |
| Mage, appropriate gear, boss, gap 0 | 0% | 54% |
| Mage, good gear, boss, gap +3 | 4% | 52% |

`class_balance` in `config/combat_balance.yaml` scales damage after `level_gap` and before a crit. `behind_dealt` applies only when that class is the lower level, so an even boss and a fight three levels up can be tuned apart. The scaled boss body in `CreateScaledEnemy` is `220 + 23*level` hit points. Content bosses still use `CreateEnemy` and the duration bands.

- Warrior even-fight damage is unchanged, so at-level trash duration stays in the old windows. The thicker boss is what pulls an appropriate-gear at-level boss into the 50–65% band. `behind_dealt` 1.20 keeps a good-gear boss at +3 near 60%.
- Class rows were replaced, not stacked. Sentinel (warrior) 1.00 dealt / 0.90 taken, Cutpurse (rogue, and ranger/hunter weapons) 0.55 dealt × 2 swings / 1.15 taken, Runecaster (mage) 1.40 dealt / 1.25 taken, Ward (stored hitch id included) 0.95 dealt / 1.00 taken. The template keeps a 50-point budget and puts the spare points in stamina (20) so later levels have a soak pool; level-1 hit points stay on the 1.05 multiplier. `behind_dealt` is capped at 1.15 for every class. The old rogue 2.35 and mage 0.46 taken are gone.
- Signatures: Brace (Sentinel, once, halves the next hit), Slip (Cutpurse, once, the next swing misses), Inscribe (Runecaster basic, 4/round × 3, refresh, no stack; basic costs 0 mana), Guard (Ward, once, the next hit aimed at an ally hits you; guarding yourself stacks two Grit). A fight opens at 1 Grit. Slam is on the level-1 bar and scales with Grit. A level-1 basic swing adds +1 so the starter sword is a 7, not a 6. A missed Slam says it missed. Ward's Slam does not start its 4-round cooldown unless the swing hits. Other Slam skills still start their cooldown on a miss. Grit is still gained only from a connecting hit. No new trash flee table.
- Ranger damage dealt is 1.26× so an at-level boss is no longer a one-sided loss. Content boss fights still last at least 12 rounds.
- Appropriate-gear elites and bosses at +5 stay under 15%.

`level_gap` per level of attacker advantage is unchanged: hit +3.5%, crit +1%, damage dealt +3.5%, damage taken +2%, clamped at ±6. Gap 0 does not change the level-gap term.

## Enemy attackSpeed

`EnemyTrait.AttackSpeed` is attacks per round. It does not change the turn beat (`TurnBeatMs` 1000 + `ReactionMs` 400). Players still use class swing counts.

- 0 or omitted: one swing when the enemy attacks, and the enemy never skips an attack action. Same tempo as before this field was honored.
- Positive values clamp to 0.25–3.
- 1.0 is one swing. Same count as 0.
- 2.0 is two swings on that attack action. 3 is three. The value is rounded, then clamped to 1–3 swings.
- Below 1 the enemy still has one swing, then holds. Period N is `round(1/speed)`, clamped to 2–4. The enemy swings when its attack-action index mod N is 0, and otherwise the log says it is slow to swing. 0.5 swings, holds, swings. 0.25 swings every fourth attack action. 0.8 uses a period of 2.
- A boss or hard telegraph is not an attack action. Extra swings apply when the blow lands.

`onAggroScript` runs once when that NPC enters a fight through `CombatController.InitiateCombat`. A player `attack` and aggro-on-sight both go through `BeginEngagement`, which calls that. `onDeathScript` runs once when that NPC dies, before loot and XP. `onFleeScript` runs once when it first chooses to flee. `onLowHealthScript` runs once when that NPC's HP first drops below `lowHealthThreshold` times max HP and the NPC is still alive. Unset or `0` means `0.30`. Other values outside `(0,1)` clamp into that range. A hit that goes from above the line to dead does not run it. Each call uses the existing Lua sandbox and its timeout. A script error is logged and swallowed. Helpers are `tales.game.msgToRoom`, `tales.combat.healNpc`, `tales.combat.applyEffect` (an existing buff or debuff id), and `tales.combat.summon(templateId, count)`.

`tales.combat.summon` spawns that many copies of an existing enemy template into the hook's fight, on the enemy side. Count clamps to 1..3 and to whatever remains of the fight cap of 3. Unknown templates, non-enemies, no active fight, and a full cap return 0 and log a warning. Adds use the normal template instance path, so HP and stats match a spawner copy, but they are not charged to a spawner. They grant no loot, gold, XP, or quest credit. Each add may run its own `onAggroScript` once. The cap stops a summon loop. Win, lose, flee, or timeout removes the adds from the room. Room lines sent while the hook runs stay `type` `message` / `username` `SYSTEM` and add `style` `combatEvent`, `hook`, and `source`. A flush that summoned adds also sends `combatStatus` with the live `combatants` roster before the next action.

## Aggro on sight

`combat.aggro_on_sight` in `config/ruleset.yaml` turns on same-room attacks for enemies whose `aggroOnSight` flag is already set. It does not edit those flags. `AggroRadius` is not a leash.

Shipped defaults: `enabled: true`, `grace_seconds: 2.5` (clamped to 0.5–10), `max_level_gap: 5`, `reaggro_cooldown_seconds: 15`. `enabled: false` is the kill switch. Gap `0` ignores level. Cooldown `0` allows another sighting immediately. An omitted block keeps these defaults, including enabled.

A grace watch is scheduled when a playing character enters a room (walk, relocate, login, teleport) and when an aggressive NPC arrives in a room that already has players. Summoned adds do not schedule one. Leaving during the grace cancels it and does not start the cooldown. The cooldown starts when a fight between that player and that NPC ends, including flee. Other aggroOnSight NPCs still standing in that room take the same quiet period, so a pair cannot hand the player to the one that did not swing. The timer only enqueues; the command loop re-checks and then uses the same engage path as `attack` (swarm pack, one `onAggro`, combat start, party nudge). The line `The <name> spots you.` is a normal message to that player and does not get `style`, `hook`, or `source`.

Skipped at schedule and again when the timer fires: ruleset off, NPC dead or not aggressive or already fighting, player dead, ghost (`AwaitingReset`), already in combat, gone from the room, disconnected, over the level gap, or still inside the cooldown. Guests are ordinary players. Several aggressive NPCs produce one fight; a swarm packs the room, and `CallForHelp` alone does not. Party members get the same assist nudge as a manual engage. They are not auto-joined.
