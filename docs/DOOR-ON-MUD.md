# Door on the MUD engine

Design for running a second world on the shared TalesMUD engine. Its client may be specific to that world. Every rule, resource, combat behavior, and piece of character state lives in the shared engine and stays available to an unconfigured server.

Written 2026-09-26. Work happens on `feat/door-on-mud` in this worktree. Nothing here merges into `engine-june`.

## The rule

A second world differs from an unconfigured server only by:

1. Config: ruleset and game-mode YAML toggles and parameters.
2. World content: pack YAML and Lua.
3. Its client UI.

The Door UI is a view. It reads real rooms, NPCs, merchants, services, combat, and resources, and it sends normal engine commands. It does not own game rules or state.

World-specific names stay out of engine Go code and engine defaults. They live in a separate world pack and in that world's client.

Every new engine behavior is useful without that pack and is default-off or default-unchanged. Tests prove the unconfigured path when the new file is absent and when it is present with the shipped defaults.

## Decision tiers

Marcus, 2026-09-26: where code does not work, work around it with a Lua script in the pack. The order for every system below is:

1. An existing engine feature plus YAML config.
2. A Lua script in a separate world pack.
3. A small generic Lua API or hook (a getter, a setter, or an event). Engine code, so no world-specific names. Sandboxed, tested, and listed in this document.
4. A new generic Go primitive only when Lua cannot own it: persistence across restart, the combat core, auth, or transport.

Prices, news, service chatter, daily flavor, forest events, trainer speech, and a later dragon or prestige script stay at tier 2. When a rule can be "the engine exposes a hook and Lua does the rule," it stops at tier 3. The map records the tier each item landed on, and why a lower tier was not enough.

`config/combat_balance.yaml` stays the owner of difficulty multipliers, named overrides, `level_gap`, threat colors, `reward_scale` (including `first_kill_bonus`), and `class_balance`. The new ruleset file sits beside it and must not repeat those keys. The loader rejects a ruleset document that contains any of them.

## What already exists

| Need | Already in engine-june |
| --- | --- |
| Fights | Tactical combat (`pkg/mudserver/game/combat`): d20 vs AC, skills, threat, party assist. Autofire is a 5s decision window; a queued action resolves on the next beat (`kickWaitingPlayerTurnLocked`). |
| Level-gap rewards | `reward_scale` multiplies **base** XP and gold after the split is planned, using the highest level among recipients. Boss first-kill bonus is `Character.FirstBossKills`. |
| Level curve | `pkg/mudserver/game/leveling`: cumulative XP table for levels 1–50, cap 50, per-character `MaxLevelCap` (guests). |
| Level-up checks | Combat victory, quest turn-in (`GrantQuestRewards`), exploration grants (`commands/xp_grants.go`), and character select. |
| Death today | `processCombatDefeat` in `game_combat.go`: 10% of XP, 1 on-hand gold, armor durability, 50% HP, relocate to `BoundRoomID` when set. Orange-or-worse attack warning stays in `commands/attack.go` and is not part of death or victory. |
| Private rooms | `pkg/instances`: per-character clone of an authored graph, destroyed when empty. Party Follow does not cross instances (`PullPartyFollowers` only when `allow=true`). |
| Shops | `MerchantTrait` buy/sell. |
| Rest | `rest` sets a faster-regen flag. It does not heal to full and does not bind. |
| People online | `who`. |
| Script hooks | Room on-enter, room actions (`response`, `response_room`, `script`), item on-use, quest hooks. The global events registry has no `Dispatch` callers. `EnemyTrait` OnDeath / OnAggro / OnFlee are not executed. |

Not in this engine: a refilling resource store, a trainer/healer/banker/inn role, a ruleset file, turn-based pacing, procedural instance generation, local Argon2id auth, or the ANSI door client. Those exist on `feat/door-mode-p0` (`pkg/daily`, `pkg/gamemode`, `pkg/authlocal`, `pkg/door`, `public/door`) and are the porting source. `pkg/door` game logic is not ported.

## System map

Each row is one system. The first number is the tier that owns it. Later numbers are the pieces under it. Door names, prices, and chatter are pack content.

| System | Tier | Why | Default when unset |
| --- | --- | --- | --- |
| Daily walks | **4**, then **3** `tales.resources`, **1** allowance YAML, **2** the pack script | The balance has to survive restart on a calendar or interval clock. Lua cannot own that table. The script spends a key and prints the line. | No keys. An unknown key writes nothing. |
| Daily flavor | **2** | A room-action `response` or on-enter script. The new-day pass does not run scripts: select still has no script event. | No lines. |
| Level-scaled monsters | **1** | Authored `EnemyTrait`, `level_gap`, and threat already scale a fight. The generator's level filter is part of the procedural row. | Authored rooms and spawners unchanged. |
| Forest random events | **2** | An on-enter or room-action script. The generator does not roll story events. | No script, no event. |
| 12 levels and trainer gate | **4**, then **1** `level_cap` / `level_up_mode`, **3** `applyLevels`, **2** the trainer script | Combat victory and quest turn-in apply levels in Go today. The mode has to intercept those sites or XP never banks. Price and speech are the script. | `auto`, cap 50. Levels still apply immediately. |
| Master challenges | **2**, on top of **1** | A normal `attack` on a boss NPC. `first_kill_bonus` in `combat_balance.yaml` already pays the first kill. No duel command. | Bosses stay authored content. |
| Death | **4**, then **1** the percents and respawn mode | Defeat has no script hook, and the default path must match today's losses with no content installed. `ApplyDeath` is called only from `processCombatDefeat`. It is not wired through victory or `reward_scale`. The orange attack warning stays in `attack.go`. | 10% XP, 1 on-hand gold, bindpoint, 50% HP, armor damage. |
| Healer | **2**, plus **3** `addGold` | `tales.characters.heal` already fills HP. The script checks coin and debits with `addGold`. No healer command. | No script, no charge. |
| Bank | **2**, plus **3** `addGold` and existing **1** `setFlag` | Death percent reads on-hand `Gold` only. The script moves coin into a character flag. No `BankGold` field. | No script. Flag absent. |
| Weapon and armor shops | **1** | `MerchantTrait` buy and sell. | Unchanged. |
| Inn | **2**, plus **3** `setBind` | The script calls `setFlag(id, "resting", true)`, which is what `rest` already stores, and `setBind` for the room id defeat already reads. | `rest` unchanged. Nothing binds unless a script calls `setBind`. |
| Gems | **2** | A character flag or a normal item. A counter does not need a column. | Unchanged. |
| News and ledger | **2** | A room-action `response` or script. Lines are content. | No new command. |
| Player list | **1** | `who`. The Door view renders that reply. | Unchanged. |
| New-day full heal | **1** | `new_day.full_heal` on character select. Select has no script event, so a YAML switch has to work with no pack script. Resource refill stays on the store's clock. | `false`. Select does not heal. |
| Dragon endgame | **2** | A later pack script. No engine type until a script is actually short of an API. Not written in Phase 1. | Off. |
| Prestige reset | **3** `setProgress`, then **2** the pack script | Level and XP have to change without a combat victory. `setProgress` does only that. The script decides when a win earns the reset and what hit points to keep. | The call is unused. Level still comes from XP. |
| PvP | **4**, not built | Hits are the combat core. Lua cannot resolve them. No flag was added that pretends a duel exists. | Impossible, as today. |
| Global event dispatch | **2** | Room and on-enter scripts already carry a flavor event. The registry still has no `Dispatch` callers, and OnDeath / OnAggro / OnFlee stay unexecuted. Calling that registry would be tier 4 and was not done. | Unchanged. |
| Turn-based fights | **4**, then **1** `combat.pacing` | This is the combat core. `auto` must stay the current 5s window and manual kick. | `auto`. |
| Procedural dungeon | **4**, then **3** `tales.instances.generate`, **2** the room-action script | Exit rewriting, cleanup, and follow blocking are the instance manager. The script pays the resource first and prints the line. The manager does not read the calendar and does not pull followers. | Existing `Enter` graph clone unchanged. |
| Local accounts | **4** | Password hashing, sessions, and the reset outbox are auth. Lua cannot own them. | Auth0. Local routes are not mounted. |
| Text client | **4** | The socket, the frame, and the key-to-command map are transport. The page reads rooms, NPCs, merchants, combat, and resources and sends engine commands. It keeps no prices, XP, or fight results. | Classic `/play`. The hook is not installed. |

## Ruleset file

Path: `config/ruleset.yaml`, next to `config/combat_balance.yaml`.

Absent file and the shipped file below are the same behavior. The loader records which one it used so tests can prove both.

```yaml
progression:
  level_cap: 50
  level_up_mode: auto    # auto | trainer
  # xp_required:         # omit = leveling.CalculateXPRequired
  #   1: 0
  #   2: 40
  # base_xp_by_enemy_level:   # omit = 15*level+5 when EnemyTrait.XPReward is 0
  #   1: 20

death:
  xp_loss_percent: 10
  gold_loss_flat: 1          # used when gold_loss_percent is 0
  gold_loss_percent: 0       # percent of on-hand gold; ignores the flat amount when > 0
  respawn: bindpoint         # bindpoint | next_reset
  respawn_hp_percent: 50
  damage_armor: true

new_day:
  full_heal: false
  timezone: UTC

resources: {}
  # forest_walks:
  #   allowance: 25
  #   reset: calendar        # calendar | interval
  #   timezone: Europe/Berlin
  #   interval: 24h          # only for reset: interval

combat:
  pacing: auto               # auto | turn_based
```

Rejected keys if present: `difficulty_multipliers`, `named_overrides`, `level_gap`, `threat`, `reward_scale`, `class_balance`, `first_kill_bonus`.

### How XP meets reward_scale

`EnemyTrait.XPReward`, when non-zero, is the base. When it is zero, the base is `base_xp_by_enemy_level[level]` if that table has the level, otherwise `15*level+5`. That integer is what victory already stores as base XP. `applyRewardScale` and the first-kill bonus run after it, exactly as they do now. The table never writes a final award and never reads `reward_scale`.

Quest XP is not an enemy base. It stays a flat quest reward. It does call the same level-up hook, so `trainer` mode banks quest levels too. Quest XP is not multiplied by `reward_scale`.

`xp_required` replaces the cumulative curve when present. Missing levels fall back to `CalculateXPRequired`. `level_cap` is the global cap passed into `GetEffectiveMaxLevel`. A lower `MaxLevelCap` on the character (guests) still wins.

### Level-up hook

One function, `ruleset.MaybeLevelUp(char)`:

- `auto`: current `CheckLevelUp` + `ApplyLevelUp`.
- `trainer`: returns nil. XP stays above the threshold. A pack script calls `tales.characters.applyLevels`, which applies whatever `CheckLevelUp` now reports. In `auto` mode that call is a no-op when victory already applied the levels.

Call sites, all of them: combat victory, `GrantQuestRewards`, `commands/xp_grants.go`, `commands/select_character.go`. Select in `auto` still catches up. Select in `trainer` does not silently train.

### Death hook

`ApplyDeath(policy, char) Outcome` computes XP lost, gold lost, respawn room (empty when `next_reset` or unbound), and the HP to set. `processCombatDefeat` applies the outcome, optionally damages armor, calls the existing `RelocateCharacter` when a room is returned, and sends the defeat message.

`next_reset` sets `Character.AwaitingReset`, leaves the character in the death room at the policy HP (0 unless `respawn_hp_percent` says otherwise), and does not relocate. The new-day pass clears `AwaitingReset` when it runs. With the default policy this path never runs.

The orange attack warning in `attack.go` stays. Death does not consult threat color.

### New-day pass

`ruleset.ApplyNewDay(char, now) bool` on successful character select, and again from `Game.ApplySessionStart` when a text-client session connects with a character already chosen. When `full_heal` is false, it returns false and writes nothing. When true, it compares `now` in `new_day.timezone` to `Character.LastResetDay`. On a new day it fills HP and mana, clears `AwaitingReset`, and stores the day. Resource balances refill when `ApplySessionStart` calls `Get` for each configured key, because the store refills on a new period.

### Resource store

`pkg/resources.Store` on the process SQLite DB (`character_resources` table, created if missing).

- Identity: `characterID|key`.
- Period key: `YYYY-MM-DD` in the allowance timezone, or the interval bucket start.
- `Get` refills `remaining` to `allowance` when the period key changes. Inside a period, a config edit does not grant more uses.
- `Consume` fails with `ErrExhausted` and does not write when `n` exceeds `remaining`.
- `Modifier` is `func(characterID, key string, allowance int) int`. Zero modifiers means the config allowance. A content script can register one to add a boon. Modifiers run before the period check so a changed allowance still does not refill mid-period; they only affect the next refill and the displayed allowance.

No game command calls the store. A pack script calls `tales.resources.consume` (slice 1a exposes it; with no configured key the call returns exhausted and changes nothing). An empty `resources` map is the unconfigured state. The Go modifier hook stays for tests and for a boon registered by engine code; content uses the YAML allowance.

### Combat pacing

Read by the combat controller only as a branch around the existing decision window.

`auto` (default): the block in `processAllTurnsLocked` is unchanged. Five-second window, then auto-attack. A queued action still kicks the waiting turn. Balance tests (`TestGapMatrixTargets`, `TestCombatDuration`, `TestLevel1*`, `TestBosses`) do not go through this branch; formulas stay put.

`turn_based`: a living player's turn sets the phase to waiting and does not arm a deadline. The ticker does not auto-attack and does not resolve that turn. Queueing an action still kicks, the player's turn resolves, and later combatants (including NPCs) take their turns under the existing beat. NPC turns that are already current still resolve; the mode does not reorder initiative. It only refuses to invent a player action. A bare `attack` during a fight queues that kick against the current target, or the first living enemy. Outside combat, `combat.bare_attack` defaults to `ask` ("Attack whom?"). `first_hostile` starts the fight against the first hostile in the room.

`combat.disconnect` defaults to `continue`: closing the session does not end the fight and does not move the character. `release` ends it as a flee, so gold, XP, and the death flag are untouched. `combat.safe_room` (`stay` by default, or `bind` or `start`) applies only in that release path. A generated instance that times out still moves its occupant to the return room before the copy is deleted. A missing room on the next enter uses the bind room, then the start room.

### Procedural instances

`Manager.Generate(rooms, characterID, playerLevel, spec)`:

- Picks `spec.Count` templates from `spec.TemplateIDs` (with replacement if the pool is smaller).
- Links them in a line. The last room's exit returns to `spec.ReturnRoomID`. The first room is the entry.
- Spawns NPC templates from `spec.Encounters` whose `[minLevel, maxLevel]` contains `playerLevel`, weighted. Templates outside the band are skipped. An empty band match spawns nothing.
- Registers the clones on the existing per-character map. A second character does not join this copy.
- `NoteLeave` to a non-clone destroys it, as today. A timeout (spec, default 30 minutes) sweeps only instances this generator created, after the occupant is moved to the return room and any fight is ended without a defeat. Authored graph instances are not on that list.
- The generator does not read the clock calendar and does not call Party Follow.

Entry is an existing `type: script` room action. The script consumes a resource, then calls `tales.instances.generate`. Params live in the script and in the action's `params` map, which room actions already pass into the script context. Exhausted budget: the script does not call generate. Followers are not pulled (`allow=false` is already how instance crossings work). There is no new action type.

Existing exits with `instance: true` still call `Enter` and clone the authored graph.

### Lua additions

Small, generic, no world names. Each is a no-op or a pure read when the caller passes nothing new, and each has an ordinary content caller in mind (a toll script, a shrine, a mentor, a daily node, a delve).

| Function | Behavior |
| --- | --- |
| `tales.characters.addGold(id, delta)` | Adds a signed amount. Refuses a debit that would go below 0 and changes nothing. Persists. |
| `tales.characters.setBind(id, roomID)` | Sets `BoundRoomID` when the room exists. Empty room id clears it. |
| `tales.characters.applyLevels(id)` | Applies every level the current XP can buy and returns how many were gained. Trainer mode banks XP until this call. Auto mode returns 0 when nothing is pending. |
| `tales.characters.setProgress(id, level, xp [, maxHP])` | Sets level and XP. Level clamps to 1..the effective cap. Negative XP becomes 0. A positive maxHP replaces max and current hit points. Class, skills, inventory, gold, and flags stay. |
| `tales.characters.top(n, sortKey)` | Read-only rows of name, level, and XP. `n` defaults to 12 and caps at 50. `sortKey` `"xp"` orders by experience. Any other key orders by level, then experience. |
| `tales.resources.get(characterID, key)` | Returns allowance and remaining for a configured key. Unknown key returns remaining 0 and `ok=false`. |
| `tales.resources.consume(characterID, key, n)` | Spends `n` against the configured allowance. Returns remaining, or `ok=false` when the key is missing or the balance is short. |
| `tales.instances.generate(characterID, playerLevel, spec)` | Spec is count, template room ids, return room, encounters `{id, minLevel, maxLevel, weight}`, timeout seconds. Returns the entry room id or fails without leaving clones. The caller script moves the character. |

`tales.characters.heal`, `damage`, `giveXP`, `teleport`, and `tales.game.setFlag` / `getFlag` stay as they are. `giveXP` still does not itself level; the ruleset hook does that on the combat and quest paths, and `applyLevels` is the explicit catch-up.

## Game-mode file

Path: `config/gamemode.yaml`, loaded only when the process is started with `-config`. Unset process keeps today's Auth0, classic client, `PORT`, and `SQLITE_PATH`.

```yaml
presentation: classic     # classic | door_tui
auth: auth0               # auth0 | local
port: ""                  # sets PORT when non-empty, so a second process can bind another port
sqlite_path: ""           # sets SQLITE_PATH when non-empty
world_pack: ""            # content folder the presentation may read for screen art
timezone: UTC
session_secret: ""        # prefer SESSION_SECRET in the environment
secret_path: data/session.key
outbox_path: data/auth-outbox.log
```

`presentation: door_tui` serves `public/door`, publishes `GET /api/door/config` for the page title, subtitle, and token key, and installs the view session hook. Classic mode does not mount `/door`. The shipped page defaults are "TalesMUD Door" and `talesmudDoorToken`. The page lists room exits and actions, and it keeps recent command replies inside the frame. A pack `keymap.yaml` binds keys by room, area, or combat; `d` stays down. A name typed at the character prompt selects that account's character, or asks for a numbered path and then sex. Paths come from `character_paths.yaml` in the world pack when that file exists, and otherwise from the system templates. Classic mode does not redirect `/` and does not install the hook. `/play` stays the classic client either way.

`auth: local` enables the Argon2id username/password routes from `pkg/authlocal`. Auth0 routes stay mounted for `auth: auth0`. Local auth is off unless the mode says so.

A second port means a second process, not a second listener inside the unconfigured server. That process runs with its own `-config`, port, and database. The default process is not modified to listen twice.

Environment overrides, all generic names: `PRESENTATION`, `RULESET` (path to the ruleset file, default `config/ruleset.yaml`), `AUTH_MODE`, `SESSION_SECRET`, `AUTH_OUTBOX_PATH`, `TRUSTED_PROXIES` (comma-separated; default `127.0.0.1,::1`). No engine default mentions a specific world. A game-mode file may set `trusted_proxies` instead of the environment variable. Forwarded client headers are ignored unless the peer is on that list.

## Door presentation

Port `public/door` (xterm page, hotkeys, login form) and the 80×25 ANSI frame builder.

The frame builder moves to `pkg/presentation/ansi`. It paints a title, body, prompt, and footer supplied by the caller. It does not know towns, prices, or combat math. The default title is empty; the Door pack supplies its title in the view's world-pack config, which is data, not an engine default.

The session hook (`pkg/presentation/doorview`) is a view:

- On connect, if the account has no selected character, the frame offers create/select and the hook calls the existing character services.
- Otherwise it loads the character's current room, NPCs, exits, resources, and combat phase from the engine and paints them.
- A hotkey or a typed line becomes an engine command or an existing room-action name (`north`, `attack`, `buy`, `look`, `who`, `rest`, or whatever the room authored). Service rooms expose those names as script actions. The view does not invent prices or outcomes.
- Combat results are the engine's combat messages, rendered into the next frame. The view does not roll hit, damage, XP, or death.

Screen art files live in the world pack (`screens/<id>.ans`) and are looked up by room id. Missing art is a text frame of the room name and description. The classic client never reads these files.

Local auth pages in that client post to `/api/auth/register`, `/api/auth/login`, `/api/auth/forgot`, `/api/auth/reset`. Tokens are the local session tokens. Forgot-password writes the outbox file and does not return the token in the response.

## Retiring pkg/door

| Old piece | Replacement |
| --- | --- |
| `pkg/door/combat.go` | Engine combat, NPC templates, `combat.pacing: turn_based`. |
| `pkg/door/hub.go` services and new-day heal | Pack scripts (`addGold`, `heal`, `setFlag`, `setBind`) and `ApplyNewDay`. |
| `pkg/door/dispatch.go` | `doorview` key-to-command map. Keys fire engine commands and room actions. |
| `pkg/door/screen.go` | `pkg/presentation/ansi`. |
| `pkg/door/pack.go` | Importer YAML world plus `ruleset.yaml` and `gamemode.yaml`. |
| `Character.Door` (`DoorProfile`) | On-hand `Gold`, a bank flag set by script, `LastResetDay`, `AwaitingReset`. Gems are a flag or an item. |
| `pkg/daily` forest-fights key | `pkg/resources` with a config key. |
| `pkg/gamemode` Door env names | `pkg/gamemode` with the generic names above. |
| `pkg/authlocal` | Ported as-is in behavior (Argon2id PHC, outbox mailer), generic paths. |
| `public/door` | Ported as the Door client. |

`pkg/door` is not copied into this branch.

## Phase plan

1. **0.** This document.
2. **1a.** `pkg/resources`, the SQLite table, and `tales.resources` get/consume. Store tests for calendar rollover, interval rollover, exhaust, modifier, and mid-period config edits. An unknown key does not create a balance. No default key is configured.
3. **1b.** `config/ruleset.yaml` and `pkg/ruleset`. XP base hook, level cap, level-up mode at all four call sites, `ApplyDeath` used only by defeat, new-day pass, `LastResetDay` and `AwaitingReset`, plus `addGold`, `setBind`, and `applyLevels`. Tests: defaults match current XP curve, fallback enemy XP, defeat losses, and auto level-up; a sample document rejects combat-balance keys; trainer mode banks combat and quest XP until `applyLevels`; `reward_scale` still multiplies a table base.
4. **1c.** `turn_based` branch. Default auto tests, including the named balance tests, stay green. A turn-based test shows the window does not fire and a queued action still resolves, with the NPC acting on a later turn. Progress note flagged for review before 1d.
5. **1d.** `Generate` plus `tales.instances.generate`. A scripted room action is the entry and the place that spends a resource. Existing instance and follow tests stay green. New tests: level filter, per-character clone, no follow across, cleanup on leave and timeout. The generator itself does not charge a resource.
6. **1e.** ANSI view, `public/door`, local auth, gamemode `-config` and port/sqlite override. Classic startup does not mount local auth and does not install the hook.
7. **Phase 2** (only if Phase 1 is green). A separate world pack, not this repository: importer YAML for town rooms, merchants, service scripts, and level-banded encounters, plus a game-mode file for the text client, local accounts, turn-based fights, a level cap, trainer level-up, and a daily resource.

## Docs kept in sync

Each implementation slice updates `PROJECT.md`, `ARCHITECTURE.md`, and `FEATURES.md` for the behavior it adds. This design stays the map; those files stay the current-state reference.
