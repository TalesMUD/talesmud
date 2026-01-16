# TalesMUD — Game Design & MVP Backlog

This document describes (1) what TalesMUD already supports and (2) the missing **MUD MVP features** to implement next. It is written to be “implementation-ready” (epic/story style) so you can hand sections to Claude Code.

## Table of Contents

- [1. MVP definition](#1-mvp-definition)
- [2. Current features (implemented)](#2-current-features-implemented)
- [3. Design principles](#3-design-principles)
- [4. MVP backlog epics](#4-mvp-backlog-epics)
  - [EPIC A — Enemy NPCs](#epic-a--enemy-npcs)
  - [EPIC B — Combat system (player vs enemy)](#epic-b--combat-system-player-vs-enemy)
  - [EPIC C — Combat instances (ad-hoc rooms)](#epic-c--combat-instances-ad-hoc-rooms)
  - [EPIC D — Items in rooms, containers, and loot drops](#epic-d--items-in-rooms-containers-and-loot-drops)
  - [EPIC E — Inventory & equipment management](#epic-e--inventory--equipment-management)
  - [EPIC F — Merchants & trading](#epic-f--merchants--trading)
- [5. Command reference (MVP)](#5-command-reference-mvp)
- [6. Data model additions (MVP)](#6-data-model-additions-mvp)
- [7. Message/UX notes (terminal + overlay)](#7-messageux-notes-terminal--overlay)

---

## 1. MVP definition

**Minimal viable “Multi-User Dungeon” for TalesMUD** means:

1) **A persistent world** of rooms that players can traverse.
2) **Multiplayer presence + communication** (who is here, chat/emotes).
3) **Interactables**: items on the ground, containers (chests), and NPCs.
4) **Core RPG loop**: fight enemies → get loot → equip upgrades → survive stronger enemies.
5) **Economy loop (basic)**: sell loot, buy gear.

**Explicit MVP constraints (to keep scope tight):**

- PvE only (no PvP).
- 1v1 combat only (one player vs one enemy) for first iteration.
- No crafting, no quests, no parties, no skills/spells beyond basic weapon attacks.
- No complex AI; enemies can be “reactive” (fight back) and optionally roam later.

---

## 2. Current features (implemented)

This section is a concise “what exists today” snapshot.

### 2.1 World & movement

- Rooms exist as persisted entities with:
  - Named exits (directional + custom exit names), room actions, and metadata (mood/background).
  - Room coordinates for map view (X/Y/Z).
  - Lists of characters and NPCs present.

- Commands already cover:
  - Move (`north/south/east/west` + aliases), `look`, dynamic exits, and room actions.

### 2.2 Multiplayer

- WebSocket-based real-time command input and message broadcasting.
- Players can:
  - See each other in rooms.
  - Use room/global chat and emotes (e.g., `scream`, `shrug`).
  - Use `who` to list online players.

### 2.3 Characters

- Character creation + selection flow exists (templates, races/classes).
- Core attributes (STR/DEX/CON/INT/WIS/CHA), HP, XP/Level, and “all-time stats” exist.

### 2.4 Items (foundational)

- **Item Templates** exist and are the canonical way to define an item “base”. A template should describe:
  - The **base identity**: name, type (weapon/armor/consumable/currency/etc.), level requirement, base price, description, slot compatibility, etc.
  - **Quality / variation rules**: when a template becomes a real item in the world, it can roll into different “qualities” (e.g., poor/common/rare) that affect stats via ranges or multipliers.
  - **Stat ranges**: templates define what stats can vary and how (e.g., `damageMin`/`damageMax` ranges, `armor` range, etc.).
  - **Random properties (“affixes”)**: templates define what extra properties can roll on drop (eligible affix pool, roll counts, weights, constraints like “no duplicate affixes”, etc.).
- **Item Instances** are the actual items that exist in the world:
  - They are created by instantiating an item template (e.g., loot drops, merchant stock creation, admin/world placement).
  - On creation they **roll** their quality + variable stats + random properties as defined by the template.
  - They then live somewhere concrete: in a **room** (on the ground), in a **player inventory/equipment**, or held by an **NPC** (merchant stock, carried loot, handed out, etc.).
- Item entity also supports:
  - Properties/attributes map for extensible stats (for both template defaults and instance-rolled values).
  - Containers with nested items (`Items`, `MaxItems`, `Closed`, `Locked`).

### 2.5 NPCs & dialogs (branch work)

- NPC entity exists and is trait-composed (DialogTrait / MerchantTrait / EnemyTrait placeholders).
- Dialog system supports branching trees, templating, once-only options, idle dialogs, etc.

---

## 3. Design principles

### 3.1 “Text-first, UI-enhanced”

The terminal remains the source of truth; UI overlays (inventory panels, dialogs, merchant lists) should be optional enhancements.

### 3.2 Deterministic + explainable combat

Damage, hit chance, and armor effects must be simple and visible (“You hit for 4 (2 blocked)”).

### 3.3 Keep persistence minimal

- Persist only what players care about long-term: character stats, items, equipment, world state.
- Combat sessions/instances can be **ephemeral in memory** (cleaned up on timeout).

### 3.4 Small surface area for scripting

Expose clear “hooks” (onAttack, onDeath, onLootOpen, onBuy, onSell) rather than letting scripts mutate anything.

---

## 4. MVP backlog epics

### EPIC A — Enemy NPCs

**Goal:** Add enemy NPC definitions that can be placed/spawned in rooms and can participate in combat.

#### A1 — Scope

- Enemy NPCs are a specialization of NPC using `EnemyTrait`.
- Enemies are visible in rooms (like other NPCs) and can be targeted by name.
- Enemies can:
  - Take damage / die.
  - Fight back when attacked.
  - Drop loot on death.

**Out of scope (MVP):** roaming AI, group combat, threat tables, spellcasting.

#### A2 — User stories

- **As a player**, I can `look` and see enemies in a room.
- **As a player**, I can `attack rat` to start a fight.
- **As a player**, when the rat dies I see a loot drop and can pick it up.
- **As a world builder**, I can create/edit enemy NPCs in the editor and place them in rooms.

#### A3 — Data model (EnemyTrait)

Add (or finalize) `EnemyTrait` fields:

- `Aggro` (bool) — optional for later, default false.
- `Level` (int32) — already on NPC; used for scaling.
- `Stats` (struct) — optional overrides; otherwise derive from level.
  - `AttackPower`, `Defense`, `Accuracy`, `Evasion`, `CritChance` (keep minimal; start with AttackPower + Defense).
- `WeaponProfile` (enum/string): `melee`, `ranged` (mostly for flavor for now).
- `LootTableID` (string) OR inline loot table definition.
- `XPReward` (int32).

#### A4 — Editor support

- NPC editor gains an “Enemy” panel when EnemyTrait exists.
- Minimal fields: HP, Level, AttackPower, Defense, XPReward, LootTable.

#### A5 — Acceptance criteria

- Enemy NPC can be created, persisted, and placed in a room.
- `look` displays enemy list with HP bar/percent.
- Enemy can be targeted reliably even with multiple NPCs (see naming rules in EPIC B).

---

### EPIC B — Combat system (player vs enemy)

**Goal:** Provide deterministic, command-driven PvE combat driven by equipped weapons/armor.

#### B1 — Combat model

**Round-based, tick-driven** combat is the simplest:

- Combat runs in discrete **rounds** (e.g., every 2 seconds) inside a combat instance.
- Players act by commands (`attack`, `shoot`, `flee`, `use`, `equip`, etc.).
- Each round:
  1) Apply player action if any.
  2) Enemy performs action (usually basic attack).
  3) Check death, rewards, cleanup.

#### B2 — Commands & parsing (combat)

- `attack <enemy>` — melee attack using equipped melee weapon, else bare hands.
- `shoot <enemy>` — ranged attack using equipped ranged weapon.
- `flee` — attempt to leave combat instance back to origin room.
- `combat` — show combat status (HP, enemy HP, round timer).

**Target resolution rules:**

- Match by exact NPC name first (case-insensitive).
- If multiple matches, allow suffix numbers: `attack rat 2`.
- If still ambiguous, return a disambiguation list.

#### B3 — Equipment influence

**Weapon (equipped):**

- `WeaponDamageMin`, `WeaponDamageMax`
- `WeaponType`: `melee` or `ranged`
- Optional: `AccuracyBonus`, `CritChanceBonus`

**Armor (equipped):**

- `ArmorValue` (int) — reduces damage or increases defense.

**Character base stats used:**

- STR: melee damage modifier.
- DEX: hit chance modifier.
- CON: max HP modifier (optional; can already be in character).

#### B4 — Damage formulas (MVP)

Keep it readable and tweakable:

- **Hit chance** (0–95%):
  - `base = 70%`
  - `+ (attacker.DEX - defender.Level) * 2%`
  - clamp to [5%, 95%]

- **Damage roll**:
  - `weaponRoll = rand(min,max)` (or fixed average for deterministic tests)
  - `statBonus = floor(STR / 5)` for melee, `floor(DEX / 5)` for ranged
  - `raw = weaponRoll + statBonus`

- **Mitigation**:
  - `blocked = defender.ArmorValue` (sum of equipped armor)
  - `final = max(1, raw - blocked)`

**Messaging:**

- “You attack Rat and hit for 6 (2 blocked). Rat HP: 9/15.”
- “Rat hits you for 3. Your HP: 17/20.”

#### B5 — Combat state

Create an in-memory `CombatSession` (ephemeral; safe to lose on restart):

- `SessionID`
- `OriginRoomID` (where combat started)
- `CombatRoomID` (instance room id)
- `PlayerCharacterIDs[]` (MVP: one entry)
- `EnemyNPCIDs[]` (MVP: one entry)
- `StartedAt`, `LastActionAt`
- `PendingPlayerAction` (struct, MVP) OR `PendingPlayerActions` (map keyed by `PlayerCharacterID`, future)
- `RoundInterval` (duration)
- `State`: `active`, `ended`

**Future-proofing requirement:** the combat state must support \(N\) players vs \(M\) enemies so a single player can fight multiple enemies, and later a group of players can fight multiple enemies. MVP can still enforce 1v1 at the command layer while the session data model supports arrays.

#### B6 — End conditions

- **Enemy death**:
  - Award XP.
  - Generate loot drops (EPIC D).
  - Return player to origin room.
- **Player death** (MVP option A):
  - Respawn player in a safe room (e.g., “Inn”) with partial HP.
  - Optional death penalty: lose some currency or XP.
- **Flee**:
  - Succeeds immediately (MVP) OR with chance (later).
  - Return player to origin room.

#### B7 — Acceptance criteria

- Players can start combat with `attack <enemy>`.
- Combat runs without blocking the global game loop and works with multiple simultaneous fights.
- Equipped weapons/armor change outcomes measurably.
- Flee works and cleans up the session.

---

### EPIC C — Combat instances (ad-hoc rooms)

**Goal:** When combat starts, move the player + enemy into an isolated “instance” so combat can run without being interrupted by other room traffic.

#### C1 — Why rooms instead of a separate “battle UI”

Rooms already provide:

- Audience scoping (messages to “room”) and presence lists.
- Existing `look` and room UI.

Using a combat-room instance keeps the engine consistent.

#### C2 — Instance room requirements

- A combat room is created on-demand:
  - `id = "combat:<uuid>"`
  - `name = "Battle: <player> vs <enemy>"`
  - `tags = ["combat", "instance"]`
  - `meta.background` can be inherited from origin room or set to a combat vignette.

- It must store:
  - `OriginRoomID`
  - `ReturnRoomID` (typically origin)
  - Participant IDs

**Persistence:**

- MVP: do **not** persist to MongoDB; keep instances in memory.
- If the server restarts, sessions end and players are returned to their last persisted room.

#### C3 — Teleport/return logic

- On combat start:
  - Remove player character from origin room roster.
  - Remove enemy NPC from origin room roster (or mark as “engaged”).
  - Add both to combat room roster.
  - Broadcast to origin room: “<player> engages <enemy> and disappears into combat!”

- On combat end:
  - Move player back to origin room.
  - If enemy died: do not return enemy.
  - If enemy fled (future): return enemy to origin.

#### C4 — “Leave combat room” / flee

- `flee` triggers exit from combat room and session cleanup.
- Also support `leave` as alias in instance rooms.

#### C5 — Cleanup and timeouts

- If no action for N seconds (e.g., 60): end session and return player.
- If player disconnects: end session and return player.

#### C6 — Acceptance criteria

- Multiple combats can happen in parallel without cross-talk.
- Other players can continue playing in origin rooms.
- Instance rooms do not pollute persistent world data.

---

### EPIC D — Items in rooms, containers, and loot drops

**Goal:** Allow items to exist in rooms, in chests/containers, and as loot generated by enemy deaths.

#### D0 — Item templates vs item instances (core rule)

- **Templates define rolls; instances are what players interact with.**
- Loot tables (and merchant stock) should reference **ItemTemplateID**; when an item is produced, the system creates an **Item Instance** by applying the template’s quality/stat/affix roll rules.
- When an instance moves (picked up, dropped, traded), it is the same rolled item moving between “owners/locations” (player inventory ↔ room ground ↔ NPC inventory/stock ↔ containers).

#### D1 — Room items

- Room has `Items` list (already in room entity); implement interaction commands:
  - `get <item>` / `take <item>`
  - `drop <item>`
  - `look <item>`

- Name resolution:
  - Match by exact item name first, then partial.
  - Support “index” disambiguation: `get potion 2`.

#### D2 — Containers / chests

Use existing item container support:

- Commands:
  - `open <container>` — sets `Closed=false` if not locked.
  - `close <container>`
  - `unlock <container>` — MVP can be stubbed; depends on keys.
  - `loot <container>` — lists contents.
  - `take <item> from <container>`
  - `put <item> into <container>` (optional MVP)

Behavior:

- If container is closed: `loot` shows “It’s closed.”
- If locked: “It’s locked.”

#### D3 — Enemy loot drops

On enemy death:

- Generate one “loot pile” container item in the *origin room*:
  - Name: “Loot of <enemy>”
  - Auto-open, disappears after TTL (optional).
- Populate it by a loot table.

#### D4 — Loot tables

MVP approach: store loot table inline on EnemyTrait or in a new collection `loottables`.

Example schema:

- `entries[]` each has:
  - `ItemTemplateID`
  - `MinQty`, `MaxQty`
  - `DropChance` (0–1)

Optional: a “roll N times” or “pick 1 of these” mode.

**Important:** loot table entries reference templates; the generated loot is always **instances** with rolled quality/stats/affixes per the template.

#### D5 — Acceptance criteria

- Items can be spawned in rooms and picked up/dropped.
- Containers work (open/loot/take-from).
- Enemy death creates loot based on table and is accessible after the fight.

---

### EPIC E — Inventory & equipment management

**Goal:** Make inventory actually usable: list items, equip/unequip, basic capacity, and ensure equipped gear influences combat.

#### E1 — Inventory UI and commands

- `inventory` / `i` prints:
  - Currency
  - Equipped summary
  - Bag items

- New commands:
  - `equip <item>`
  - `unequip <slot>` OR `unequip <item>`
  - `wear <item>` (alias of equip)
  - `remove <item>` (alias of unequip)
  - `compare <item>` — compare stats vs currently equipped item in same slot (optional MVP)

#### E2 — Slots

Leverage the existing slot model (10 slots already exist in character entity):

Typical slots:

- Head, Chest, Legs, Feet
- Hands, Ring, Amulet
- MainHand, OffHand
- Ranged (or MainHand can cover)

Ensure:

- Equip validation: armor can only go into matching slot.
- Weapon rules:
  - MainHand always allowed for melee weapons.
  - Ranged slot for bows.

#### E3 — Capacity (optional MVP)

- MVP simplest: no weight, only “max item count” in inventory.
- Later: add weight and encumbrance.

#### E4 — Item stat encoding

Standardize item property keys so combat can consume them without special casing:

- Weapon:
  - `damageMin`, `damageMax`, `weaponType` (melee|ranged)
- Armor:
  - `armor`

#### E5 — Acceptance criteria

- Equip/unequip updates character state and persists.
- Combat uses the equipped weapon/armor.
- Inventory rendering is stable and readable.

---

### EPIC F — Merchants & trading

**Goal:** Let players buy and sell items with NPC merchants.

#### F1 — Merchant interaction model

Two viable MVP options:

**Option 1 (recommended):** `trade <merchant>` opens a “merchant session” similar to dialog.

- Player enters a merchant state, commands are interpreted as trading commands until `exit`.

**Option 2:** Integrate into dialog options (“Show me your wares”).

- Requires dialog nodes to trigger actions.

Start with Option 1, then bridge into dialog later.

#### F2 — Merchant inventory

Merchant NPC has:

- `Stock[]` (item template IDs + quantities)
- `BuybackEnabled` (bool)
- Price multipliers:
  - `SellToPlayerMultiplier` (e.g., 1.0)
  - `BuyFromPlayerMultiplier` (e.g., 0.5)

Pricing rules:

- Base price stored on item template (e.g., `price` property).

**Note:** merchants reference templates for what they sell, but the items the player receives are **instances** (created from templates, rolling quality/stats/affixes if the template allows it).

#### F3 — Currency

For MVP:

- Treat currency as an integer `Gold` on character.
- Optionally also support an actual “currency item” type later.

#### F4 — Trading commands

Within merchant session:

- `list` — list merchant stock with prices.
- `buy <index|name> [qty]`
- `sell <index|name> [qty]`
- `inspect <index|name>` — show details.
- `exit` — leave merchant session.

Also allow global shortcuts:

- `buy <item> from <merchant>`
- `sell <item> to <merchant>`

#### F5 — Acceptance criteria

- Player can buy items (currency decreases, item added).
- Player can sell items (item removed, currency increases).
- Prices are consistent and based on item template base price.
- Merchant stock persists.

---

## 5. Command reference (MVP)

### World & social

- `n/s/e/w` — move
- `look [target]` — inspect room, player, NPC, or item
- `who` — online players
- `say <text>` — room chat (recommended addition if not present)
- `scream <text>` — room/global (existing)
- `emote <text>` / `shrug` — emotes

### Items & inventory

- `inventory` / `i`
- `get|take <item> [<index>]`
- `drop <item> [<index>]`
- `equip <item>`
- `unequip <slot|item>`
- `open|close <container>`
- `loot <container>`
- `take <item> from <container>`

### Combat

- `attack <enemy> [<index>]`
- `shoot <enemy> [<index>]`
- `combat`
- `flee`

### Merchants

- `trade <merchant>`
- `buy ...`, `sell ...`, `list`, `exit` (inside trade)

---

## 6. Data model additions (MVP)

### 6.1 New entities / collections

- `loottables` (optional; can be inline on EnemyTrait initially)
- `merchantstocks` (optional; can be inline on MerchantTrait initially)

### 6.2 New in-memory state

- `combatSessions` map keyed by player character ID (or session ID)
- `instanceRooms` map keyed by room id
- `merchantSessions` map keyed by player character ID

---

## 7. Message/UX notes (terminal + overlay)

### 7.1 Combat UX

- Always print a compact status line after each round.
- Add a `MessageTypeCombatUpdate` (optional) so frontend can render a combat panel.

### 7.2 Inventory UX

- Terminal:
  - number items for index targeting
  - show slot icons/labels
- Overlay:
  - keep the existing inventory panel but add equip buttons later.

### 7.3 Merchant UX

- Terminal list formatted as:
  - `1) Iron Sword — 12g`
  - `2) Leather Cap — 8g`


---

## 8. Implementation stories (epic breakdown, JIRA-style)

This section turns each epic into concrete “stories” with technical notes aligned to the current architecture (Game loop + CommandProcessor/RoomProcessor + Services + Mongo). See `ARCHITECTURE.md` for the existing layering and message flow.

### EPIC A stories — Enemy NPCs

#### Story A-1: Finalize EnemyTrait data model

**As a world builder,** I can create an NPC with EnemyTrait fields so it can participate in combat.

**Implementation notes (backend):**

- Entity: `pkg/entities/npcs/NPC` already exists; ensure `EnemyTrait` is not `nil` when NPC is an enemy.
- Define a minimal stat model; keep it “flat” and serializable:
  - `AttackPower int32`
  - `Defense int32`
  - `Accuracy int32` (optional MVP; can be derived)
  - `Evasion int32` (optional MVP)
  - `XPReward int32`
  - `LootTableID string` (or inline table)
- Add `State` for enemies engaged in combat:
  - `EngagedSessionID string` (empty when free)
  - OR keep this only in memory to avoid persistence churn.

**Acceptance criteria:**

- Enemy NPCs can be created/updated via existing NPC CRUD endpoints.
- Enemy NPCs serialize to Mongo and load back with EnemyTrait intact.

**Test ideas:**

- Unit: JSON/YAML serialization round-trip for enemy NPC.

#### Story A-2: Place/spawn enemies in rooms

**As a player,** I see enemies listed in `look` and in the room UI.

**Implementation notes:**

- Rooms already have `NPCs` presence; ensure the room render path includes enemy NPCs.
- Optional MVP: allow a room to have a spawn list (template IDs + count) and spawn them on server start.

**Acceptance criteria:**

- A room shows “NPCs here:” including enemies.

#### Story A-3: Enemy death removes enemy from the world and produces rewards

**As a player,** killing an enemy removes it and spawns loot.

**Implementation notes:**

- On enemy death, remove NPC from its room roster (origin room or instance room depending on EPIC C).
- Trigger loot generation (EPIC D) and XP award.

---

### EPIC B stories — Combat system

#### Story B-1: Add combat session manager (in-memory)

**As the game engine,** I can track multiple independent combat sessions and advance them on a short tick.

**Implementation notes:**

- Add a `CombatManager` owned by `Game`:
  - `sessions map[string]*CombatSession` keyed by `PlayerCharacterID` (MVP 1v1) OR session ID (recommended once sessions can contain multiple players).
  - `instanceRooms map[string]*rooms.Room` (if you keep combat rooms in memory; see EPIC C).

**Tick model (recommended):**

- Keep all state mutations in the main game loop goroutine.
- Create a `time.Ticker` (e.g., 1s) in the MUD server that posts `CombatTick` events into `Game.onMessageReceived`.
- The game loop handles the tick event by:
  - iterating active sessions,
  - applying queued actions,
  - emitting messages.

This avoids locking and keeps concurrency simple.

**Acceptance criteria:**

- 20+ simultaneous sessions do not block normal commands.
- Sessions auto-timeout and clean up.

#### Story B-2: Implement `attack <enemy>` command

**As a player,** I can start combat by attacking an enemy in my current room.

**Implementation notes:**

- Add a new command under `pkg/mudserver/game/commands/` (global or room command).
- Flow:
  1) Resolve target enemy NPC in current room.
  2) Validate: enemy not already engaged.
  3) Create combat session + combat instance room (EPIC C).
  4) Move player + enemy to instance.
  5) Emit messages:
     - to player: combat start + status
     - to origin room: “X engages Y …”

**Edge cases:**

- If player is already in combat → show current combat status.
- If enemy not found → suggest candidates.

**Acceptance criteria:**

- `attack rat` starts combat and prints a clear start banner.

#### Story B-3: Implement `shoot <enemy>` command

Same as B-2 but validates ranged weapon.

**MVP simplification:**

- If no ranged weapon equipped, return: “You have nothing to shoot with.”
- Ignore ammo for MVP.

#### Story B-4: Implement combat round resolution

**As a player,** my actions and the enemy’s actions resolve each round with readable output.

**Implementation notes:**

- Store `PendingPlayerAction` per session:
  - `Type: attack|shoot|use|none`
  - `TargetNPCID`
  - `SubmittedAt`
- On each round tick:
  - Apply player action (if any).
  - Apply enemy action.
  - Emit `MessageTypeDefault` lines for each step.

**Acceptance criteria:**

- Combat output is deterministic enough to test (seed RNG or inject RNG).

#### Story B-5: Implement `flee` command

**As a player,** I can leave combat and return to the origin room.

**Implementation notes:**

- `flee` validates player is in combat.
- On success:
  - end session,
  - move player back,
  - decide enemy handling (MVP: enemy returns to origin room and is “reset” to full HP OR stays where it is).

**Acceptance criteria:**

- `flee` always works in MVP.

---

### EPIC C stories — Combat instances

#### Story C-1: Create in-memory instance rooms

**As the engine,** I can create an ad-hoc room that behaves like a normal room but is not persisted.

**Implementation notes:**

- Represent an instance room with the existing `rooms.Room` structure.
- Keep a `Game.instanceRooms` registry.
- Adjust room lookup logic:
  - If `CurrentRoomID` starts with `combat:` OR exists in `instanceRooms`, load from memory.
  - Else load from `RoomsService`/Mongo.

**Acceptance criteria:**

- `look` works in the instance room and lists both participants.

#### Story C-2: Move character + NPC into the instance

**As a player,** starting combat moves me into a combat room.

**Implementation notes:**

- Reuse the existing room membership update helpers (the code already removes offline chars, syncs rooms, etc.).
- Ensure both the character’s `CurrentRoomID` and the room’s `Characters` list update consistently.

**Acceptance criteria:**

- After `attack`, `CurrentRoomID` is the combat room.

#### Story C-3: Cleanup instance room on combat end

**As the engine,** I clean up memory and restore world state.

**Implementation notes:**

- Remove instance room from registry.
- Return player to `OriginRoomID`.
- If enemy survives, return it too (or reset and respawn).

---

### EPIC D stories — Items, containers, and loot

#### Story D-1: Implement room item pickup/drop

**As a player,** I can pick up items from the room and drop them back.

**Commands:**

- `get|take <item>`
- `drop <item>`

**Implementation notes:**

- Update both:
  - Character inventory (add/remove item references or embedded items depending on how inventory is represented).
  - Room `Items` list.
- Persist both changes.

**Acceptance criteria:**

- Items move between room and character inventory and survive reload.

#### Story D-2: Implement container interactions

**As a player,** I can `open` a chest, `loot` it, and take items from it.

**Implementation notes:**

- Containers are items with nested `Items` list.
- Decide where container lives:
  - In room `Items`, or in inventory.
- Ensure “take from container” updates both the container contents and the destination inventory.

**Acceptance criteria:**

- Closed containers block looting.

#### Story D-3: Loot generation on enemy death

**As a player,** I get loot after defeating an enemy.

**Implementation notes:**

- On death, create a “loot pile” container item (unlocked/open).
- Insert into origin room’s `Items`.
- Generate contents by loot table:
  - Use `ItemTemplatesService` to instantiate items.

**Acceptance criteria:**

- Loot appears consistently and is collectible.

---

### EPIC E stories — Inventory & equipment

#### Story E-1: Implement `equip` and `unequip`

**As a player,** I can equip gear and it affects stats.

**Implementation notes:**

- Validate slot compatibility.
- On equip:
  - remove item from inventory,
  - put into `EquippedItems[slot]`.
- On unequip:
  - move back into inventory.

**Acceptance criteria:**

- Equipped item no longer appears in bag list.

#### Story E-2: Compute derived combat stats from equipment

**As the combat engine,** I can compute armor and weapon profile from equipped items.

**Implementation notes:**

- Add helper functions:
  - `GetEquippedWeapon(character, weaponType)`
  - `GetTotalArmor(character)`
- Keep stat keys normalized (`damageMin`, `damageMax`, `armor`).

**Acceptance criteria:**

- Changing armor changes incoming damage.

#### Story E-3: Improve inventory rendering

**As a player,** `inventory` output is easy to scan.

**Implementation notes:**

- Number items.
- Show slot summary.
- Show currency.

---

### EPIC F stories — Merchants

#### Story F-1: MerchantTrait + stock definition

**As a world builder,** I can define merchant stock and pricing.

**Implementation notes:**

- Add `MerchantTrait` fields:
  - `Stock[]` (template ID + qty)
  - multipliers
- Prefer item templates as stock source.

#### Story F-2: Implement `trade <merchant>` session

**As a player,** I can enter a merchant session and buy/sell.

**Implementation notes:**

- Add `MerchantSession` stored in `Game` keyed by player character.
- While in session:
  - route commands to `MerchantProcessor` (similar to RoomProcessor).
- Emit formatted lists.

**Acceptance criteria:**

- `trade bob` then `list`, `buy 1`, `sell 2`, `exit` works.

#### Story F-3: Integrate dialog option hook (post-MVP)

**As a player,** I can reach trading via dialog.

**Implementation notes:**

- Add an optional `Action` field on dialog nodes (e.g., `openMerchant`, `startQuest`).
- Execute as a safe whitelist.


---

## 9. Extra MVP “must-haves” (small but important)

These are not big epics, but they tend to be required to make PvE + loot feel stable.

### 9.1 Enemy respawn / room repopulation

- After an enemy is killed, repopulate it after a delay (e.g., 60–300s) or on server restart.
- Keep respawn rules per room: `spawnTemplateID`, `count`, `respawnSeconds`.

### 9.2 Death + respawn rules

- Define a single respawn room (“Inn”, “Town Square”).
- On death: restore to X% HP, optionally apply a light penalty (currency drop, durability later).

### 9.3 Persistence + restart safety

- If the server restarts mid-combat, characters should load back into their last persisted non-instance room.
- Instance rooms and combat sessions should not persist.

### 9.4 Basic anti-spam / rate limiting (chat + commands)

- Per-player command cooldown (e.g., 3–5 cmds/sec) to avoid log/CPU abuse.
- Chat throttling for `scream`.

### 9.5 Targeting ergonomics

- A consistent “index targeting” helper used across NPCs and items:
  - `rat`, `rat 2`, `rusty sword`, `rusty sword 3`.
- Always show numbered lists in `look` and `inventory` so targeting is obvious.

### 9.6 Observability (dev only)

- Add `debug combat` command for admins to dump active sessions.
- Add structured logs for: combat start/end, loot generation, trade operations.
