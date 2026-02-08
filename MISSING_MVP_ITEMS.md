# Missing MVP Items

Systems audit of what needs to be built or fleshed out before the game is MVP-ready. The core architecture is solid — no new major systems are needed, but several existing systems have gaps.

---

## P0 — Must Have (core loop is broken without these)

### Leveling System
**Status:** Missing — XP is tracked and awarded, `EventPlayerLevelUp` is defined, but no actual leveling logic exists.

**What's needed:**
- XP threshold curve per level (e.g., level 2 = 100 XP, level 3 = 250 XP, ...)
- Level-up detection when XP crosses threshold
- Stat increases on level-up (HP, attributes based on class)
- Level-up notification to player (message + UI update)
- Max level cap
- Level display in character stats and combat

**Files involved:** `pkg/entities/characters/character.go`, `pkg/mudserver/game/game_combat.go` (XP award), new leveling logic module

---

### HP Recovery
**Status:** Missing — after combat, players keep reduced HP with no way to recover except dying (respawn at 50% HP).

**What's needed:**
- Passive HP regen over time (slow, out-of-combat only)
- `rest` command for faster recovery while idle
- Healing consumable items (the `use` command exists but needs HP restore logic)
- Optional: healing from leveling up (full HP restore on level)

**Files involved:** `pkg/mudserver/game/game.go` (tick loop), `pkg/mudserver/game/commands/use.go`, new `rest` command

---

## P1 — Should Have (multiplayer and world feel incomplete without these)

### Basic Chat System
**Status:** Missing — only `scream` (room broadcast) and `shrug` (emote) exist. No direct communication between players.

**What's needed:**
- `say <message>` — speak to current room
- `whisper <player> <message>` or `tell <player> <message>` — private message to online player
- `yell <message>` — broadcast to adjacent rooms
- Chat message types in the message system for UI rendering

**Files involved:** `pkg/mudserver/game/commands/` (new command files), `pkg/mudserver/game/messages/messagetypes.go`

---

### NPC Idle Behavior
**Status:** Partial — wander, patrol, and idle dialog are all marked TODO in `game_npcs.go`. NPCs are static until interacted with.

**What's needed:**
- Wander behavior: NPCs randomly move to adjacent rooms within their wander radius
- Patrol behavior: NPCs follow their defined patrol path (ordered room list)
- Idle dialog: NPCs occasionally say their idle dialog lines in the room
- Arrival/departure messages when NPCs move between rooms

**Files involved:** `pkg/mudserver/game/game_npcs.go` (TODO lines ~56, ~62, ~76), `pkg/entities/npcs/npc.go`

---

## P2 — Nice to Have (improves depth but MVP works without it)

### Abilities / Skills
**Status:** Missing — combat is attack/defend/flee only. All classes play identically. A `Skill` entity exists but isn't integrated.

**What's needed:**
- Resource system (mana, stamina, or cooldown-based)
- 1-2 abilities per class to differentiate them (e.g., Warrior: Power Strike, Wizard: Fireball)
- Ability use in combat as an action type
- Damage/effect calculation for abilities
- UI for ability selection during combat turns
- Ability learning tied to level (unlock at certain levels)

**Files involved:** `pkg/entities/characters/class.go`, `pkg/mudserver/game/combat/engine.go`, new ability definitions, UI updates in `public/mud-client/`

---

## P3 — Defer (only needed if multiplayer co-op is in MVP scope)

### Party System
**Status:** Stub — `Party` struct exists with name + character list, but zero gameplay integration.

**What's needed:**
- Party formation commands: `invite`, `accept`, `leave`, `disband`
- Shared combat (party members join the same combat instance)
- XP splitting among party members
- Party chat channel
- Party member status visibility (HP, location)
- UI for party display

**Files involved:** `pkg/entities/party.go`, `pkg/service/parties.go`, `pkg/mudserver/game/game_combat.go`, new command files, UI updates

---

## Summary

| Priority | System | Effort | Impact |
|----------|--------|--------|--------|
| P0 | Leveling | Small-Medium | Enables the core progression loop |
| P0 | HP Recovery | Small | Makes combat sustainable |
| P1 | Basic Chat | Small | Enables multiplayer communication |
| P1 | NPC Idle Behavior | Medium | Makes the world feel alive |
| P2 | Abilities/Skills | Medium-Large | Differentiates classes, deepens combat |
| P3 | Party System | Large | Enables co-op play |
