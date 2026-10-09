# TalesMUD Core Systems & Features Reference

**Purpose**: This document catalogs ALL core systems, data structures, features, and APIs available in TalesMUD. It is designed to provide complete context for AI agents reworking game content to leverage the latest capabilities.

**Last Updated**: 2026-06-18

---

## Table of Contents

1. [Data Structures (Entities)](#data-structures-entities)
2. [Room System](#room-system)
3. [Item System](#item-system)
4. [Character System](#character-system)
5. [Multiplayer Session & Social System](#multiplayer-session--social-system)
6. [NPC System](#npc-system)
7. [Combat System](#combat-system)
8. [Skills & Spells System](#skills--spells-system)
9. [Quest System](#quest-system)
10. [Dialog System](#dialog-system)
11. [Scripting System (Lua API)](#scripting-system-lua-api)
12. [Creator UI Capabilities](#creator-ui-capabilities)
13. [Recent Features & Best Practices](#recent-features--best-practices)
14. [Discovered World Atlas](#discovered-world-atlas)
15. [Game Client Tab Container Widget](#game-client-tab-container-widget)

---

## Data Structures (Entities)

### Base Entity
All entities inherit from:
```go
type Entity struct {
    ID string `json:"id"`  // UUID
}
```

### Common Traits
Reusable trait composition pattern:
- `BelongsUser` - User ownership tracking (`BelongsUserID`)
- `CurrentRoom` - Location tracking (`CurrentRoomID`)
- `LookAt` - Inspection text (`Detail` field)

---

## Room System

### Room Entity Structure
```go
type Room struct {
    *entities.Entity
    traits.LookAt

    Name, Description string
    RoomType, Area, AreaType string
    Tags []string

    // Scripting Hooks
    OnEnterScriptID string  // Lua script executed when player enters room

    Actions *Actions  // Custom player interactions
    Exits   *Exits    // Room connections

    // Live Data (runtime state)
    Items      *Items       // Item IDs in room
    Characters *Characters  // Character IDs in room
    NPCs       *NPCs        // NPC resident IDs (not instances!)

    // Grid Positioning (optional)
    Coords *struct{X, Y, Z int32}

    // Client Meta
    Meta *struct{
        Mood       string  // Ambient mood setting
        Background string  // Background image ID
    }

    CanBind bool  // Allow /bind command for respawn point
}
```

### Room Actions
Players can interact with rooms via custom actions:
```go
type Action struct {
    Name        string  // Command trigger (e.g., "examine statue")
    Description string  // Shown in "You can:" list
    Response    string  // Reply to player (type: response)
    Type        RoomActionType  // "response", "response_room", "script"
    ScriptId    string  // Lua script to execute (type: script)
    Params      map[string]interface{}
}
```

**Action Types**:
- `response` - Send message to player only
- `response_room` - Broadcast message to all in room
- `script` - Execute Lua script with context: `ctx.room`, `ctx.character`, `ctx.action`

### Room Exits
```go
type Exit struct {
    Name        string  // Exit command (e.g., "north", "trapdoor")
    Description string  // Shown in exits list
    Type        RoomExitType  // "normal", "direction", "teleport"
    Hidden      bool    // Not shown until revealed (per-character)
    Target      string  // Target room ID
    Params      map[string]interface{}
}
```

**Exit Types**:
- `direction` - Standard cardinal directions (n/s/e/w/u/d)
- `normal` - Named exits (e.g., "door", "gate")
- `teleport` - Instant transport to distant location

### Per-Character Hidden Exit Reveals
Hidden exits can be revealed on a **per-character basis** using scripting:
```lua
-- In a room action script:
tales.game.revealExit(roomID, "secret-passage", characterID)
```

- Reveal is tracked on `Character.RevealedExits` map (persisted)
- Other players do NOT see the exit unless they also reveal it
- Client receives automatic room update showing new exit

### Room Discovery & Exploration XP
```go
// On Character entity:
DiscoveredRooms map[string]bool  // Room IDs
DiscoveredAreas map[string]bool  // Area names
```

**Exploration XP Rewards**:
- **5 XP** per new room discovered (grant path is currently gated; discovery itself still records)
- **15 XP** for first room in a new area/zone

**Atlas API**: `GET /api/characters/:id/map` returns the character's fog-of-war atlas (places, paths, area hulls, overworld/lower/upper layers). Layout preserves authored area-local geometry and compass exits, translates zones onto compact configurable centers, and fills anonymous ground between zones into one connected continent. Hidden exits stay off the map until `revealExit`.

### Terrain atlas and continent (Worldmap P1/P1b/P1c/P1d/P1e/P1f/P1g/P1j)
Discovered atlas places carry `terrain`: grassland, forest, farmland, city, castle, dungeon, swamp, mountain, snow, desert, water, shore, ruins, or interior. Unexplored neighbors carry `terrain: "fog"`; their art remains hidden. `pkg/worldmap/map_terrain.json` owns ordered aliases, area defaults, and the unknown-ground default (grassland). RoomType/areaType and specific tags take priority, followed by name, indoor/underground context, area, then descriptive/legacy fallbacks. Classification does not add persisted entity fields or scripting APIs.

The overview and minimap share a compact cached 32px overview with precomposed directional dither sprites from the native 48px art sheet and 48px close detail baked on demand with six stable variants per terrain, nearest-neighbor sampling, ordered-dither edge/corner transitions, a smoothed organic contour with sand/foam/depth bands, mixed rock foothills/taller central peaks and biome-specific oak/pine/dead-tree clumps, and quieter dirt roads/bridges following charted outdoor compass exits. The road network batches unique exit segments into one path beneath mountains, tree canopies, town paving, walls, buildings and props; full-opacity town paving hides dirt inside streets and thinner strokes reduce far-zoom clutter. Towns render street-shaped paving, varied red/brown/blue roofs, angular wall polygons following courtyard/street footprints, corner towers, incoming road gatehouses, keeps, and small hamlet clusters. Crisp town/keep/village glyphs remain readable at far zoom and select their owning real room. P1g redrew all terrain and decoration rows at 48px. **P1j** ships the imagegen craft atlas at the same 48px / 288×9024 layout (seamless terrains, chroma/alpha stamps, recomposed dither blends). Shore bands use twelve samples per cell with broken foam glints; image scaling uses nearest-neighbor sampling. P1e roads-under-stamps, P1f zoom clamps (1–10, 220px cap), and A9 remain intact. Close zoom retains small pixel trees; overview uses larger biome-specific clumps and varied ridge chains, with snow caps only on the highest peak variants. Towns/farms add wells, carts, lanterns, fences and one windmill per farm zone; plateau/highland rooms add mesa edges/outcrops. Seas use varied wave texture, offshore rocks and highland coastal cliffs. World fit and minimum zoom share a frame filling roughly 70–80% of desktop map height (width permitting). Wheel and pinch share a 1–10 scale clamp; the tile-size cap is 220px, so the closest view is twice P1e’s maximum in both world-fit and local recenter frames. High zoom keeps nearest-neighbor pixel sampling. Visible underground passages have cave or mine entrances. Shops, taverns, houses, halls, and upstairs rooms share exterior anchors instead of separate overworld floor tiles. Selecting a town/building offers a filterable list of discovered interiors; the selected interior keeps its actual room ID, intel, exits, and Travel action. Visible entrance choices switch to Lower, without disclosing unexplored names or hidden exits.

Derived presentation fields are `Place.mapRole` (`surface`, `interior`, `underground`), `surfaceRoomId` (interior anchor), `town`, and `entrances` (visible Lower target IDs). Discovered rooms additionally carry `mapFeatures` (derived decoration keys), `artSeed` (deterministic customization seed), and `undergroundStyle` (`cave`, `crypt`, `cellar`, `sewer` on Lower); fog rooms omit all three. `PlayerMap.landscape` contains decorative `{x,y,terrain}` cells: no room IDs, names, hit targets, or travel destinations. Ground reveals near discovered surface/interior rooms or across a fully charted surface area; other ground is fogged. Terrain/fog, coordinates, grouping, towns, entrances, features, art seeds, and underground style updates repaint; stale fog cannot replace charted ground. Outdoor positive Z represents elevation on Overworld; subterranean rooms stay Lower. Other untagged above-ground floors can still use Upper. Gold current-room glow (including instanced interiors), zone labels, selection/intel, travel, layer tabs, zoom/pan, Fit world, and local recenter remain available.

Room art follows tags, room/area types, names, service action names, descriptions/detail fallbacks, and bind context. Forge roofs have chimney smoke; shrines spires; taverns/shops signs; farms fields/farmhouses; guards/gates towers; ruins broken walls; graveyards headstones; docks piers; mines timber entrances; magic sites glowing stones. Reeds, stumps, flowers, rocks, and tree species vary deterministically. Hints do not export source descriptions, script IDs, or service parameters. Outdoor water rooms retain their own terrain even beside forests/swamps. Art changes require no new authored entity fields or Lua API.

Lower uses dedicated cave/crypt/cellar/sewer floors, merged adjacent rooms, actual-exit rock-sided corridors, perimeter stone walls, torch light, and disclosed stair/entrance marks against dark void. Unknown rooms keep fog and expose no art hints. The full overlay has low-rate water shimmer and smoke; hidden/closed views pause it, reduced motion fixes the phase and suppresses its timer, and teardown clears it. Cached overview/close/Lower scenes preserve marker-only reuse and fog isolation.

P1d Lower uses rough cave rock, crypt paving/bones, sewer water channels, and wooden cellar floors/barrels with dim cluster light. Only short, aligned known compass exits get visible tunnels; longer/ambiguous links remain usable navigation exits without grey lines across void. Tree/building drop shadows are removed; only restrained mountain face shading remains. On supported browsers an OffscreenCanvas worker builds scene bitmaps while the main thread remains responsive; only one job runs and stale queued exploration work is replaced. Close LOD builds on demand; worker/canvas fallback preserves map behavior if unavailable. Unexplored overworld ground uses a soft volumetric fog overlay (blurred explore-boundary mask + cloud wash) instead of per-tile fog stamps; You / Turn-in / Selected markers paint above it. Authored room coords that treat +Y as north are flipped at compile so Cartographer north paints upward (`?v=mapnorth1`). Cached fog ground is painted once, and real room centers win hit testing over neighboring decorative props.

`GET/HEAD /api/map-tiles/terrain-sheet.png` serves the 288×9024 RGBA sheet with a content-hash query version; live client JS/CSS uses `?v=creamtimber1` (sheet still content-hash busted). Production art is the P1j imagegen craft atlas (content `assets/map-tiles`, version `5caee700810a` (cream-timber settlement stamp remapper on P1j; prior `c8169d17ef81`)); Pillow generators (`tools/generate_map_tiles.py`, `tools/map_hires_art.py`) remain for regenerating the older P1g look only. `pkg/worldmap/map_layout.json` configures zone centers, natural ground, town flags, separation, and coast padding; unconfigured zones attach using exits. The layout does not mutate stored coordinates or gameplay topology. See `tools/WORLDMAP-PREVIEW.md`. Ornate banners and map framing ornaments remain later work.

### NPC / enemy portraits
Room presence sends `portrait` URLs (`/api/portraits/{templateOrId}.png`). Import copies `assets/images/sprites/{npcs,enemies}/` into `uploads/portraits/`. Sprites are 512px full-figure art; the original NPC/enemy cards clip a 48px square around the body (`object-fit: cover` + zoom). Missing files fall back to hashed `img/avatars/{1-14}p.png`. Component CSS lives in `public/mud-client/public/extra.css` and must be deployed with `bundle.js`.

### Player portraits
The equipment paper doll, combat card, and party roster use 512px transparent sprites at `/api/portraits/player-<race>-<class>.png`. The set covers Human, Dwarf, Elf × Warrior, Rogue, Mage, Ranger, Cleric, Druid. Stored `elve` maps to `elf`, `wizard` to `mage`, and `hunter` to `ranger`. The server supplies combat and party portrait URLs; the client derives the equipment URL from character race/class. An unavailable combination or failed image uses a class silhouette.

### Group combat cards
BattleStage shows the local player and all other combat players (up to the five-player party cap). Compact ally cards display portrait, class, level, live HP/MP, current turn, and down/fled state. WoW-style buff/debuff icon rectangles sit above/beside combat portraits (self, allies, enemies) with remaining rounds from statusEffects. Combat action snapshots carry participant type, class, mana, and status; a join sends the roster to existing fighters immediately and triggers a short join banner. Healing and buffs currently target self or enemies only; ally-card clicks do not queue unsupported commands.

**Battle layout B (default):** room-field combat for players with no saved preference. Party stands on the left — the player on a gold marker, and, when other players are in the fight, a compact left ally strip (existing portrait, name, slim HP). Enemies stand on the right. Detail frames stay top-left (you: name, HP/MP, buffs) and top-right (focused foe). Sprites use existing race/class and enemy portraits (`/api/portraits/player-<race>-<class>.png` and `/api/portraits/{template}.png`); a class or enemy silhouette is only the failed-load fallback. Over each sprite, only a slim HP bar. Mid-fight Resolving/Waiting pills and turn-name chips stay hidden; the round chip and your-turn countdown remain. Full hotbar 1–9 + Flee dock + room vignette unchanged. Classic cards remain in Settings → Gameplay (turn Battle layout B off), `localStorage.talesmud_battle_layout_b=0`, or `?battleLayout=classic` / `?battlepoc=0`. `?battleLayout=b` / `?battlepoc=1` forces layout B. URL wins over localStorage; an explicit `0`/`1` wins over Settings. Client cache `?v=battleb1`.

The self card and ally cards share a desktop row and stack on phones, with horizontally scrollable phone allies. Only structured actions generate an action banner; join prose is not repeated in stage banners. Defeated enemies keep a grey sprite and Defeated label. Damage and healing numbers sit over player/enemy sprites with a dark outline and hold full opacity before fading; reduced motion disables their animation. Material Icons use a preloaded local WOFF2 font and remain hidden if it fails to load.

WebSocket connects go through a process-wide gate (`websocketGate.js`): one CONNECTING/OPEN/CLOSING socket, no reactive `ws=null` reconnect, and close code 4001 (session replaced) does not auto-reconnect. The Map overview is an Inventory-style body-portal panel (`map-panel`, never Materialize's `.modal`) with explicit pixel size so the canvas fills the stage. `/play` JS/CSS/HTML is served `Cache-Control: no-cache` plus `?v=` on asset URLs so deploys are not stuck behind a cached `bundle.js`.

### Instanced cellars
Exits with `type: instance` or `instance: true`, or a normal directional exit from a shared/hub room into a room tagged `instance`/`instanced`, create a private copy of the destination plus rooms reachable without walking back into the hub. Two guests share the town room (e.g. The Weary Wanderer `R0203`) and get different cellar IDs (`R0215~aabbccdd`). Hidden cellar wings (R0230+) are cloned with the entrance. When the last occupant leaves, clones and copied NPCs are destroyed.

Copy ids use `~` (`pkg/instances.IsCloneID`). The in-memory manager is empty after a restart, so before the process accepts players it deletes every persisted room id that contains `~`. NPC copies are memory-only. Ground item ids are stored on the room document, so deleting the copy removes that placement and leaves the item catalog. An NPC row whose id contains `~`, or a spawner whose id or room id contains `~`, is deleted too. Characters saved in a copy, or in a room id that no longer exists, move to that template's hub (the non-instance room whose exit enters the graph), else the copy's return exit, else the world start room, else the bind room. An empty saved room is left empty. Stale combat flags are cleared on the characters who move. The next login in that process sends `You find yourself back at <room name>.` once. A later login does not repeat it. A live copy still known to the manager is left alone. An id that contains `~` but is missing from that map still counts as a copy for disconnect release and for the login room check. There is no SIGTERM hook; the next start runs the same sweep.

### Armor durability
Dying chips equipped armor (default 4 hits to broken). Defense scales with remaining durability. Items are never deleted. `repair` at a merchant restores them.

### Merchant prices
Each NPC `MerchantTrait` has its own `buyMultiplier` / `sellMultiplier` / `priceOverride` on the same item template, so overlapping stock can cost differently per shop.

---

## Item System

### Item Entity Structure
```go
type Item struct {
    *entities.Entity
    traits.LookAt

    // Template/Instance Pattern (same as NPCs)
    IsTemplate     bool    // True if blueprint
    TemplateID     string  // Source template ID (for instances)
    InstanceSuffix string  // Unique suffix (e.g., "abc123")

    Name, Description string
    Type    ItemType     // "currency", "consumable", "armor", "weapon", "collectible", "quest", "crafting_material"
    SubType ItemSubType  // "sword", "twohandsword", "axe", "spear", "shield"
    Slot    ItemSlot     // Equipment slot or "inventory"
    Quality ItemQuality  // "normal", "magic", "rare", "legendary", "mythic"
    Level   int32

    // Custom Properties
    Properties map[string]interface{}  // Arbitrary key-value data
    Attributes map[string]interface{}  // Stats (e.g., "healthRestore": 50)

    // Container System
    Closed   bool
    Locked   bool
    LockedBy string  // Key item ID
    Items    Items   // Nested items
    MaxItems int32

    // Interaction Flags
    NoPickup           bool    // Cannot be picked up
    CopyOnPickup       bool    // Create personal copy instead of removing from room
    BoundToCharacterID string  // Item is bound to this character (cannot drop/sell/trade)

    // Stacking & Economy
    Stackable bool
    Quantity  int32
    MaxStack  int32
    BasePrice int64  // Gold value

    // Scripting & Consumption
    OnUseScriptID string  // Lua script executed when item is used
    Consumable    bool    // Remove/decrement on use

    // Metadata
    Tags      []string
    Created   time.Time
    CreatedBy string
    Meta      *struct{Img string}
}
```

### Item Types & Slots
Stackable item quantities are kept consistent when consumed or partially dropped: the character inventory and backing item instance are both updated.

`unique: true` on an item template means a character can hold at most one copy. The count covers the bag, equipped gear, and items nested in a container. There is no separate item bank. Pickup of another is refused with "You already have the <name>." Extra copies already in the bag are trimmed to one on that attempt. The ground drop is left for someone else.

A loot entry may set `rarity: unique`, `chance` (a non-zero value wins over `dropChance` at import), and `boss_only: true`. Boss-only entries roll only when the dead NPC's difficulty is `boss`. The unique roll is skipped when every victory recipient already holds that template. A drop from a `rarity: unique` entry sends a room message `UNIQUE: <item name>` with `style: combatEvent`, `hook: unique`, and `source` set to the NPC's display name. The play client frames that victory row and the item card. A template `unique: true` flag still caps ownership and frames the item card on examine. It does not send the room chip on its own.

**Item Types**:
- `currency` - Gold, tokens
- `consumable` - Potions, food, scrolls
- `armor` - Wearable protection
- `weapon` - Swords, axes, staves
- `collectible` - Quest items, trophies
- `quest` - Quest-specific items
- `crafting_material` - Components for gathering/crafting recipes

### Gathering & Crafting (v1 — no professions)
- **Everyone** can gather and craft. No skill ranks, no profession unlocks.
- **Gathering**: room actions `GATHER …` / `FORAGE …` / `HARVEST …` (scripted nodes with deplete flags) plus commands `gather` / `forage` / `harvest [target]`.
- **Crafting**: `recipes` lists all recipes; bare `craft` opens the same list; `craft <recipe>` consumes mats and grants output. Client seeds a **Recipes** action-bar pin; R0209 also exposes **CRAFT** / **RECIPES** room chips.
- Optional **stations**: recipe `station: forge` requires room tag `forge` or `crafting`; `campfire` requires `campfire`/`kitchen`/`hearth`. Food/leather recipes craft anywhere.
- Recipe YAML lives in content `data/recipes/` (loaded from `import/mvp-rpg-1/data/recipes` at runtime; seed fallback in engine).

**Equipment Slots**:
- `head`, `chest`, `legs`, `boots`, `hands`, `neck`, `ring1`, `ring2`
- `main_hand`, `off_hand`
- Special: Two-handed weapons occupy both hand slots

### CopyOnPickup Feature
When `Item.CopyOnPickup = true`, `Item.IsTemplate = true`, or the item is a world catalog ID placed in a room (no `templateId` / instance suffix):
1. **First pickup**: Creates a personal instance from the blueprint, adds to inventory
2. **Instance is bound**: `BoundToCharacterID` is set on the instance, making it character-bound
3. **Room remains unchanged** - the world item stays for other players and later guests
4. **Character tracking**: Marks item as collected via `Character.Flags["collected_item:<templateID>"]`
5. **Subsequent pickups**: Player sees "You have already collected X"
6. **Client behavior**: Item is hidden in room display for that character
World import sets `copyOnPickup` on every item referenced by a room so tutorial caches (torch, flint, dagger, fragments) cannot be consumed by the first guest. Dropped loot instances (templateId set, not a template) are still taken from the room.

Room presence and other room-audience messages are fanned out from in-memory sessions, not a SQLite room reload on the send goroutine. Reloading there can deadlock the game loop after an on-enter script, so a follow-up `north` never leaves the room. Unique NPCs with only `spawnRoomId` (empty `currentRoomID`) are still placed on server start. Spawners refill to `initialCount` immediately; `spawnInterval` only throttles extra instances up to `maxInstances`.

**Bound item restrictions** (enforced on items with `BoundToCharacterID`):
- Cannot be dropped (`drop` command rejects with message to use `destroy`)
- Cannot be sold to merchants
- Cannot be traded
- `destroy` command removes from inventory, clears collected flag, and refreshes room

**Helper methods**: `Item.IsBound()`, `Item.IsOwnedBy(characterID)`

**Use case**: Quest items, story artifacts that all players should find

### Item Usage & Consumables
```go
// Built-in effects via Attributes:
Attributes: {
    "healthRestore": 50,      // Restore 50 HP
    "manaRestore": 30,        // Restore 30 mana
    "useMessage": "You feel refreshed"  // Custom message
}
```

**Usage flow**:
1. Player uses item: `use health potion` or `use flint on torch` (optional `on <target>`)
2. System checks `Attributes` for built-in effects
3. If `OnUseScriptID` is set, executes Lua script (preferred over built-in torch lighting)
4. Fire-starters without a script auto-target a carried `light_source` on bare `use`; otherwise soft-hint `use <item> on <target>`
5. If `Consumable = true`, decrements quantity or removes item

World YAML may set `onUseScript` or `onUseScriptId` (importer accepts both). Inventory UI shows **Use** for usable items and **Use on…** for tools/fire-starters to pick a second inventory item without Terminal X.

### Item Template/Instance Pattern
```go
// Create instance from template:
instance, err := itemsService.CreateInstanceFromTemplate(templateID)

// Check if item is an instance:
if item.IsInstance() {  // Has TemplateID and InstanceSuffix
    // ...
}
```

---

## Character System

### Character Entity Structure
```go
type Character struct {
    *entities.Entity
    traits.BelongsUser
    traits.CurrentRoom

    Name, Description string
    Race  Race   // Human, Elf, Dwarf, Halfling, Orc
    Class Class  // Warrior, Mage, Rogue, Cleric, Ranger, Druid

    // Core Stats
    CurrentHitPoints, MaxHitPoints int32
    CurrentMana, MaxMana int32  // Casters only
    XP, Level int32
    Gold int64
    MaxLevelCap int32  // Per-character level cap (0 = use global MaxLevel)

    // Attribute System
    Attributes Attributes  // []{Name, Short, Value} - STR, DEX, CON, INT, WIS, CHA

    // Distributable Attribute Points (2 per level)
    UnspentAttributePoints int32
    SpentAttributePoints   map[string]int32  // Tracks total spent per attribute

    // Inventory & Equipment
    Inventory     items.Inventory
    EquippedItems map[items.ItemSlot]*items.Item
    EquippedSkills []string  // Skill IDs (max 4, level-gated)

    // Combat State
    InCombat         bool
    CombatInstanceID string
    BoundRoomID      string  // Respawn location (set via /bind)

    // Per-Character Game State
    Flags map[string]interface{}  // Script-settable flags
    RevealedExits map[string][]string  // roomID → exit names
    DiscoveredRooms map[string]bool
    DiscoveredAreas map[string]bool
    FriendIDs []string  // Other character IDs on this character's friends list

    // All-Time Statistics
    AllTimeStats struct {
        PlayersKilled   int32
        GoldCollected   int32
        QuestsCompleted int32
        RoomsDiscovered int32
    }
}
```

### Attribute System
Six core attributes (D&D-style):
- **STR** (Strength) - Melee damage, warrior scaling
- **DEX** (Dexterity) - Dodge, rogue/ranger scaling, initiative
- **CON** (Constitution) - HP bonus
- **INT** (Intelligence) - Mage/Druid spell power, mana capacity
- **WIS** (Wisdom) - Cleric spell power, mana regeneration
- **CHA** (Charisma) - Social interactions (future)

**Attribute Modifiers**:
```go
modifier = (value - 10) / 2
// Example: STR 14 → +2 modifier, STR 8 → -1 modifier
```

### Distributable Attribute Points
Players earn **2 points per level-up** to spend freely:
```bash
# Terminal commands:
spend          # Show status table with current values, spent/cap
spend str 3    # Allocate 3 points to STR
spend dex      # Allocate 1 point to DEX
```

**Class-based Caps** prevent degenerate builds:
- Warriors: INT capped at 5
- Mages: STR capped at 5
- Each class has primary/secondary attribute caps defined

**Retroactive Points**: Existing characters receive `(level - 1) * 2` points on login

### Character Methods (Combat-Related)
```go
// Attribute access
GetAttribute(short string) int32          // Get raw value
GetAttributeModifier(short string) int    // Get (value-10)/2
GetSTRMod(), GetDEXMod(), GetCONMod() int // Specific modifiers
GetINTMod(), GetWISMod() int

// Combat calculations
GetWeaponDamage() int32        // Main hand damage (1 if unarmed)
GetArmorDefense() int32        // Total from equipped armor
CalculateMaxMana() int32       // Caster: 20 + Level*5 + INTMod*4
CalculateManaRegen() int32     // Per combat round: 1 + WISMod (min 1). Separate from regen.in_combat.
```

### Character Flags System
Arbitrary key-value storage for script state:
```lua
-- Set flag:
tales.game.setFlag(characterID, "puzzle_solved_statue", true)

-- Get flag:
local solved = tales.game.getFlag(characterID, "puzzle_solved_statue")
```

**Common uses**: Quest state, puzzle progress, secret discoveries

`LastResetDay` records the calendar day of a new-day heal. `AwaitingReset` is set when the death policy's respawn mode is `next_reset`. Both stay empty on existing characters.

### Ruleset profile

`config/ruleset.yaml` sits beside `config/combat_balance.yaml` and must not repeat its keys (`level_gap`, `threat`, `reward_scale`, `class_balance`, difficulty multipliers, named overrides). The shipped file matches current play: level cap 50, automatic level-up, 10% XP loss and 1 on-hand gold on defeat, respawn at the bind room with half HP, no dawn heal, no resource keys, combat pacing `auto`, `combat.safe_room: stay`, `combat.disconnect: continue`, and `combat.bare_attack: ask`.

An enemy's authored XP reward is the base. When that reward is 0, `progression.base_xp_by_enemy_level` supplies the base, and otherwise the built-in `15*level+5` curve does. `reward_scale` multiplies that base afterward. `level_up_mode: trainer` banks combat, quest, exploration, and select catch-up until `tales.characters.applyLevels`. Quest XP is not multiplied by `reward_scale`.

Death math is `ruleset.ApplyDeath`, called from defeat only. A fight is a defeat when nobody is still fighting and at least one player is actually dead. When every player has fled or slipped, the fight ends as fled: no XP loss, no gold loss, no armor damage, and no respawn move. A slip that already took one exit stays in that room. In a mixed party only the dead take the penalty. Anyone who fled gets the escaped notice and keeps their hit points, gold, and room. Post-combat cleanup still clears `InCombat` and syncs the escaper's combat hit points.

`combat.pacing: auto` keeps the 5 second decision window and resolves a queued action on the next beat. `turn_based` leaves that window open until the player sends a command. NPCs still take their own turns afterward. The default file is `auto`. During a fight, a bare `attack` queues a swing on the current target or the first living enemy so a turn-based round advances. Outside combat, `combat.bare_attack: ask` (the default) still answers "Attack whom?". `first_hostile` starts the fight against the first hostile in the room.

`combat.disconnect: continue` (the default) leaves a dropped connection in the fight and does not move the character. `release` ends that fight as a flee: no gold loss, no XP loss, and no death flag. `combat.safe_room` is `stay` (default), `bind`, or `start`, and applies only when disconnect is `release`. `stay` leaves the character in a real room. `bind` and `start` move them. A room id containing `~` is treated as an instance even when the in-memory map has lost it, so release still moves that character. A generated instance that times out still moves its occupant to the return room and ends the fight without a defeat. On the next enter, a saved room that no longer exists is replaced by the bind room, then the start room. An orphan `~` id uses the startup sweep's room order instead.

Passive regeneration is the `regen` block in the ruleset, and a world game-mode file may carry the same block. `out_of_combat`, `resting`, and `in_combat` each have HP and mana pools with `enabled`, `percent`, `flat`, and `interval_seconds`. Missing keys keep out of combat at 2% HP and 5% mana, resting at 10% HP and 15% mana, and in combat at 0.5% HP and 1% mana, every 10 seconds, flat 0. The gain is `int(max * percent / 100) + flat`, at least 1 when the pool is active. A pool is active when `enabled` is true and percent or flat is positive. `enabled: false`, or percent and flat both 0, grants nothing and is not due. An explicit `interval_seconds` below 1 is rejected.

Intervals follow one server clock, not the character. The first tick after login lands anywhere from 0 to interval−1 seconds in. A very small enabled interval saves the character and sends one websocket update per regenerating player per interval. Turning every pool off also skips the fully-rested and in-combat resting cleanup on that clock. `applyRegeneration` and `InterruptRest` still clear the resting flag.

Off: `enabled: false` on that pool (`regen.out_of_combat.hp`, `regen.resting.hp`, `regen.in_combat.hp`, and the matching mana pool). Slow: `percent: 0.5` and `interval_seconds: 60` out of combat, `percent: 1` and `interval_seconds: 60` while resting, `percent: 0.1` and `interval_seconds: 30` in combat.

`CalculateManaRegen` (1 + WISMod per combat round, minimum 1) is separate from `regen.in_combat`.

### Refilling resources

Per-character balances live in the `character_resources` table. A key grants uses only after something configures an allowance (calendar day in a timezone, or a fixed interval). Inside a period, raising the allowance or a modifier does not give the extra uses back; the next period refills to the new amount. An empty catalog, which is the process default, answers every key as not configured and writes no row.

Another world can use a key for a daily gathering node or a delve ticket, spent from a room-action script. No content ships a key, so play is unchanged. Entering play calls `Get` for each configured key, so a new period is refilled even before a script reads it. A one-word room action accepts a trailing argument (`deposit 20`); the script sees it as `ctx.args`. Multi-word action names stay exact.

### Procedural instances

`tales.instances.generate(characterID, playerLevel, spec)` builds a private line of up to 20 rooms from a template pool. Encounters whose level band contains `playerLevel` are returned as a spawn plan and placed when a game is attached. A second character gets a different copy. The same character cannot hold two instances. Leaving to a non-clone room destroys the line, and a timeout (default 30 minutes) destroys only these generated instances. Authored cellar graphs are unchanged. Every generated exit is marked so Party Follow does not cross it. The generator does not spend a resource and does not move the character; the room-action script does. A timeout still ends a fight in that copy without a defeat penalty and moves the character to the return room before the copy is deleted. A disconnect does that only when `combat.disconnect` is `release`.

### CopyOnPickup Tracking
```go
// Character methods:
HasCollectedCopyItem(templateID string) bool
MarkCollectedCopyItem(templateID string)

// Stored in Flags["collected_copy_items"] as array of template IDs
```

---

## Multiplayer Session & Social System

### Live Session Presence
The game engine keeps an in-memory live session registry keyed by connected
user ID. Each live session tracks the selected character, character name,
current room, and last-seen timestamp.

Uses:
- Room player lists in `enterRoom` and `roomUpdate` WebSocket messages
- Room chat and room-based message fan-out
- `who` command output
- `tell`/`whisper` target lookup
- Passive regeneration ticks
- Periodic cleanup of stale `Room.Characters` entries

`User.IsOnline` remains persisted for administrative/status visibility, but
live routing and room presence use the session registry so stale database flags
do not make disconnected players appear reachable.

### Character Selection Lifecycle
On WebSocket connect, the game attaches the user's last character if available
or auto-selects the first owned character. Selecting another character with
`sc <name>` removes the previous character from its room, updates
`User.LastCharacter`, refreshes the live session, sends `characterSelected`,
then sends the current room, character stats, inventory, and quest log.
Character switch, room movement, and disconnect paths emit silent room presence
refreshes so other clients update player lists without waiting for a full room
render.

### Room Presence WebSocket Message
```json
{
  "type": "roomPresence",
  "roomId": "room-uuid",
  "players": [
    { "id": "character-uuid", "name": "Aster", "isYou": false }
  ]
}
```

The server broadcasts this to rooms with online active characters only. The
client marks `isYou` locally based on the currently selected character because a
single broadcast is shared by multiple users.

### Party Commands

`party create`, `party invite <player>`, `party accept`, `party decline`,
`party leave`, `party kick <player>` (leader), `party promote <player>` (leader),
`party list`, `party follow`, `party unfollow`, `party say <message>` (or `party <message>`).

Party membership is persisted in the existing `Party` entity (SQLite JSON),
including `leaderCharacterId`. Creator is leader. Soft/hard cap: **5** members
(`entities.MaxPartySize`) enforced on invite and accept. Pending invites
remain in-memory on the game server and expire after **45 seconds** (both sides
get a clear timeout notice + `party_invite{pending:false}`). Guests may invite,
accept, decline, follow, and chat in parties (Friends still refuse guests).

Structured WebSocket payloads:
`party` `{inParty,partyId,partyName,leaderId,maxMembers,members[{id,name,online,level,class,portrait,isLeader}]}`
and `party_invite` `{pending,inviterName,partyId,expiresAt}` (clear with `pending:false`).

Client (Party invite popup + guest parties, cache-bust `?v=partyinvite1`): HUD Party
button opens a gold-bordered panel — party name title + `N/M members · K online`
subtitle; member rows with avatar/initial, You/Leader badges, class · level, online
pill; sticky action bar (Say primary, Invite secondary, Leave danger+confirm);
party-say strip (~8 lines); Create/Invite empty state; mobile bottom-sheet. A
pending invite opens a centered Accept/Decline popup with a countdown (not only a
terminal line). Friends rows use matching **Invite** outline. Leader sees Kick on
other members.

Room players overlay and Friends rows can invite online players. Guests see the
Party button and full party UI.

### Party Combat Assist (v1)
Same-room players can join an in-progress fight by `attack <npc>` on an enemy
already in combat. Joiners receive `combatStart` (BattleStage), initiative is
rolled, and turn order is rebuilt (current actor preserved). Cross-room join
is refused. Party membership is **not** required to join; when combat starts,
same-room online party members get a `[Party] … Type 'attack <enemy>' to join`
nudge. Out of scope: need/greed item assignment, auto-pull without attack,
following anyone except the party leader, Flutter.

### Party Loot & XP Share (v1)
On victory, gold and XP split equally among the living combatants **and**
online party members standing in the killer's room (persisted `CurrentRoomID`).
The killer is the first living combatant (the player who engaged). Offline
members and members in another room get nothing. A solo victor, or a fight
where no two recipients share a party, is unchanged: each living combatant
gets `total/n` and any remainder is dropped.

When two or more recipients are in the same party, leftover gold and XP
(`total % n`) go to the killer so nothing is discarded. The victory combat log
lists each recipient (`PARTY SHARE`), and every recipient gets a one-line
`[Party] Equal split…` toast. Item drops stay on the ground (no need/greed).
Quest kill credit stays with living combatants only.

### Party Follow
`party follow` starts following the current party leader. `party unfollow` stops.
While following, a normal exit walk by the leader (`TakeExit`, exit type empty /
`normal` / `direction`, arriving at the authored target) moves followers who are
standing in the room the leader just left. Anyone left behind walks a path of
those same ordinary exits (up to 12 rooms) toward the leader: a late `party follow`,
the end of combat, or a reconnect. Followers in combat stay put until the fight
ends, then catch up. A disconnect keeps the flag; an offline body is not moved,
and the other side is told the follow is still on. Teleports, portals, bindstones,
script relocations, hidden exits (except the step just taken with the leader),
and private-instance crossings are not used for the chase. Guests, characters who
are not in a party, the leader, and anyone already in combat cannot start following.
Leaving the party, being kicked, the leader leaving, or a leadership change clears
the follow flag. The flag is in-memory on the game server. The party payload
includes `following` for the recipient. The reply is a `[Party] You are following <leader>`
line (party strip and room toast). Still out of scope: auto-join combat without
`attack`, item need/greed, and following a member who is not the leader.

### Friends (v1)
Per-character friends list stored as `Character.FriendIDs` (character UUIDs) in
the SQLite character JSON blob. Names are resolved at display time.

```bash
friend add <name>       # Add by live session name, or exact offline character name
friend remove <name>    # Remove from your list
friend list / friends   # Show Online/Offline from the live session registry
```

Rules: no self-add; idempotent add; guests are refused politely (and cannot be
added). Open add-by-name (not a mutual request flow). Presence: friends who
are watching you get a short "came online / went offline" line. The `friends`
WebSocket message carries `{id,name,online}` for the play client overlay.

Client: HUD group button (next to the character switcher) opens a Friends
panel — list, Tell (whisper), Remove, add-by-name. Room players overlay can
add the other player. Guests hide the HUD entry.

### Client Session UX
The MUD client exposes connection state in `MUDXPlusStore`:
`disconnected`, `connecting`, `connected`, and `reconnecting`.
`Game.svelte` owns reconnect scheduling and retries automatically after socket
close. `CharacterSwitcher.svelte` shows the active character and connection state.
**Switch character** in the account menu (desktop chip and phone header) opens
`CharacterPicker.svelte`, which lists every character from `/api/my-characters`.
Choosing one sends `sc <name>`. A signed-in player with more than one character
sees that picker once per login; the server still enters on `lastCharacter`.
Guests get **Continue with X**, **Continue with Google**, and **Email and password**
(Auth0 `loginWithRedirect`; X and Google pass `connection` so the password form
is not the default). **Log out** clears the Auth0 session and this tab's guest token and
returns to the welcome choice instead of restoring a guest. `Client.js` handles `roomPresence` messages and
updates `MUDXPlusStore.players` without changing the room description.

---

## NPC System

### NPC Entity Structure
```go
type NPC struct {
    *entities.Entity
    traits.BelongsUser
    traits.CurrentRoom

    Name, Description string
    Race  Race
    Class Class

    CurrentHitPoints, MaxHitPoints int32
    Level int32

    // Template/Instance Pattern
    IsTemplate     bool    // Blueprint vs spawned instance
    TemplateID     string  // Source template for instances
    InstanceSuffix string  // Unique ID suffix

    // Behavior Configuration
    SpawnRoomID  string         // Where NPC respawns
    RespawnTime  time.Duration  // Time to respawn (0 = no respawn)
    WanderRadius int            // Rooms to wander from spawn
    PatrolPath   []string       // Room IDs for patrol route

    // State Machine
    IsDead    bool
    DeathTime time.Time
    State     string  // "idle", "combat", "patrol", "dead", "fleeing"

    // Combat State
    InCombat         bool
    CombatInstanceID string

    // Behavior Traits
    EnemyTrait    *EnemyTrait
    MerchantTrait *MerchantTrait

    // Dialog References
    DialogID          string          // Main dialog tree
    IdleDialogID      string          // Ambient chatter
    IdleDialogTimeout time.Duration   // Idle dialog frequency
    LastIdleDialog    time.Time       // Last ambient chatter trigger
}
```

### NPC Behavior Loop
NPCs are processed by the game update loop every 10 seconds:
- `idle` NPCs with `WanderRadius > 0` move through visible exits while staying within that many room hops from `SpawnRoomID`.
- `patrol` NPCs follow `PatrolPath` as a looping ordered list of room IDs. If the current room is not in the path, the NPC moves to the first patrol room.
- NPCs with `IdleDialogID` and `IdleDialogTimeout` broadcast ambient chatter to their current room when the cooldown has elapsed.
- Dead spawned instances are removed for spawner replacement; dead unique NPCs respawn at `SpawnRoomID` after `RespawnTime`.
- An enemy with `aggroOnSight` engages a player who enters its room, and engages players already there when it spawns, respawns, or walks in. `config/ruleset.yaml` `combat.aggro_on_sight` defaults to on, 2.5s grace, level gap 5, and 15s reaggro cooldown. That quiet period also covers other aggressive NPCs still in the room. `enabled: false` stops it. The sight line is ordinary text. The fight then uses the same path as `attack`.
- NPC movement sends silent room updates so clients refresh NPC presence without reprinting the room description.

Room NPC payloads sent to the MUD client include `isEnemy`, `isMerchant`, `isQuestGiver`, `hasDialog`, `hasIdleDialog`, and `state` so the UI can show interaction badges without duplicating backend lookup rules.

### Enemy Trait
```go
type EnemyTrait struct {
    // Classification
    CreatureType string  // "beast", "humanoid", "undead", "elemental", "construct", "demon", "dragon", "aberration"
    CombatStyle  string  // "melee", "ranged", "magic", "swarm", "brute", "agile"
    Difficulty   string  // "trivial", "easy", "normal", "hard", "boss"

    // Combat Stats (base values, modified by difficulty multipliers)
    AttackPower  int32
    Defense      int32
    AttackSpeed  float64 // attacks per round; 0 or omitted is one swing and the old beat

    // AI Behavior
    AggroRadius   int     // Detection range in rooms (0 = passive)
    AggroOnSight  bool    // same-room engage after combat.aggro_on_sight grace
    CallForHelp   bool    // Alert nearby enemies
    FleeThreshold float64 // HP % to flee (e.g., 0.2 = flee at 20% HP)

    // Rewards
    XPReward       int64
    GoldDrop       Range{Min, Max int64}
    LootTableID    string    // Reference to loot table
    GuaranteedLoot []string  // Item template IDs that always drop
    MaxDrops       int32     // Max items from loot table (0 = unlimited)

    // Event Scripts (once per fight, sandboxed; errors are logged and swallowed)
    OnAggroScript      string  // when this NPC enters the fight
    OnDeathScript      string  // when this NPC dies, before loot and XP
    OnFleeScript       string  // when this NPC first chooses to flee
    OnLowHealthScript  string  // first time HP drops below LowHealthThreshold while still alive
    LowHealthThreshold float64 // fraction of max HP; 0 or unset = 0.30; clamp to (0, 1)
}
```

### Merchant Trait
```go
type MerchantTrait struct {
    MerchantType   string  // "general", "blacksmith", "alchemist"
    Inventory      []MerchantItem
    RestockMinutes int32
    LastRestock    time.Time
    BuyMultiplier  float64  // Price when buying (1.0 = normal)
    SellMultiplier float64  // Price when selling (0.5 = half)
    AcceptedTypes  []string // Item types merchant will buy (empty = all)
    RejectedTags   []string // Tags preventing buy (e.g., "soulbound", "quest")
}

type MerchantItem struct {
    ItemTemplateID string
    BasePrice      int64   // Override item base price
    PriceOverride  int64   // Force specific price (ignores multipliers)
    Quantity       int32   // Current stock (-1 = unlimited)
    MaxQuantity    int32   // Max after restock
    RequiredLevel  int32   // Player level requirement
}
```

Merchant commands are available in rooms with merchant NPCs:
- `list`, `shop`, or `trade` (exact key — no NPC name) opens a structured `shop` WS payload for the client overlay and refreshes after buy/sell
- `buy <item> [quantity]` purchases stock; stackable quantities can fit in one inventory stack
- `sell <item> [quantity]` sells accepted, unbound inventory items
- `value <item>` / `price <item>` checks the merchant's sell price

Web/Flutter clients inject a **Trade / Shop** dialog option when talking to a merchant (`isMerchant`), open a room-widget shop overlay (2-column horizontal buy/sell rows with icon + name + price, paginated), and send bare `list` (not `trade <name>`). Shop/inventory icons use `/api/item-art/{templateId}.png` (never `meta.img` prose prompts), then generic PNG, then local SVG.

Trading is blocked while the character is in combat. Merchant stock can restock lazily when a player interacts after the configured interval.

### NPC Spawner System
```go
type NPCSpawner struct {
    *entities.Entity

    TemplateID    string         // NPC template to spawn
    RoomID        string         // Where to spawn
    MaxInstances  int            // Max alive at once
    SpawnInterval time.Duration  // Time between spawns
    InitialCount  int            // Spawn on world load

    RespawnTimeOverride *time.Duration  // Override template respawn
}
```

**Spawner Behavior** (automated by server):
- Checks every 5 seconds
- Spawns instances up to `MaxInstances`
- Tracks live instances via `NPCInstanceManager`

### NPC Resident System
Two methods for placing NPCs:
1. **Via spawners** - Dynamic, respawning instances
2. **Via Room.NPCs list** - Static residents
3. **Via NPC.CurrentRoomID** - Unique NPCs auto-spawn into their assigned room on server start

**Important**: `Room.NPCs` contains template/unique NPC IDs, NOT instance IDs

---

## Combat System

### Joining in-progress fights (Party Combat Assist v1)
Same-room players may `attack <npc>` to join an active combat instance that
NPC is already in (`JoinCombat`). Cross-room joins are refused. Living joiners
share victory XP/gold. Online party members in the killer's room are added to
that equal split (Party Loot & XP Share v1); leftover goes to the engager when
the split is a party share. Loot remains room drops.

### Combat Instance Model
```go
type CombatInstance struct {
    ID              string
    OriginRoomID    string  // Room where combat started

    Players         []CombatantRef
    Enemies         []CombatantRef

    TurnOrder       []CombatantRef  // Initiative-sorted
    CurrentTurnIdx  int
    TurnStartTime   time.Time
    Round           int

    State           CombatState  // "pending", "active", "victory", "defeat", "fled", "timeout"
    Log             []CombatLogEntry
}
```

### Combatant Reference
```go
type CombatantRef struct {
    ID, Name        string
    Type            CombatantType  // "player" or "npc"
    Initiative      int
    IsAlive, HasFled bool

    // Core Stats
    Level           int32  // Snapshot used for level-gap modifiers
    MaxHP, CurrentHP int32
    AttackPower, Defense int32
    STRMod, DEXMod, CONMod, INTMod, WISMod int
    DefenseBonus    int32  // From defend action

    // Mana & Skills
    MaxMana, CurrentMana, ManaRegen int32
    EquippedSkills []string
    SkillCooldowns map[string]int  // SkillID → rounds remaining
    StatusEffects  []StatusEffect
    QueuedSkillID  string
}
```

### Status Effects
```go
type StatusEffect struct {
    ID, SkillID, Name string
    Type    string    // "buff", "debuff", "dot", "hot", "stun"
    Stat    string    // "attack", "defense", "dodge"
    Value   int32     // Flat modifier
    Percent float64   // Percentage modifier (0.30 = +30%)
    Duration int      // Rounds remaining
    SourceID string   // Caster ID
}
```

### Combat Flow
```
1. INITIATION
   - Player: attack <npc>
   - NPC: aggroOnSight, same room, after combat.aggro_on_sight grace (AggroRadius is not a leash)
   - Create CombatInstance, roll initiative (1d20 + DEX mod)

2. TURN ORDER
   - Sorted by initiative (highest first)
   - Each combatant takes action in sequence
   - Round increments after all combatants act

3. PLAYER ACTIONS (60-second timer per turn)
   - attack <target> - Melee/ranged attack
   - cast <skill> [target] - Use skill (mana/cooldown)
   - defend - +50% defense until next turn
   - flee - Attempt escape (50% + DEX bonus)
   - timeout - Auto-defend after 60 seconds

4. NPC ACTIONS (automated AI)
   - If HP < FleeThreshold → attempt flee
   - Otherwise → attack weakest player

5. ROUND START EFFECTS
   - Process DoT/HoT ticks
   - Regenerate mana (1 + WISMod per round)
   - Decrement cooldowns
   - Check stun effects

6. RESOLUTION
  - Victory (all enemies dead) → XP, gold, loot
  - Defeat (all players dead) → 10% XP loss, 1 gold loss, respawn at bind point
  - Fled (all players escaped) → NPCs reset to idle
  - Timeout (30 minutes) → Combat ends, no rewards
```
If a character remains flagged as in combat after the runtime combat instance is gone, combat commands clear the stale flag and return the normal not-in-combat response.

### Combat Commands
```bash
attack <target>    # Attack enemy or switch target
a <target>         # Alias

cast <skill> [target]  # Use skill
spell <skill>          # Alias
1, 2, 3, 4             # Quick-cast slot shortcuts

defend             # Defensive stance (+50% defense)
d, guard           # Aliases

flee               # Attempt to escape
run, escape        # Aliases

status             # Show combat status
cs, combat         # Aliases
```

### Difficulty Balance System
**Config File**: `config/combat_balance.yaml`

```yaml
difficulty_multipliers:
  trivial:
    hp: 3.75
    attack: 4.0
    defense: 1.0
  easy:
    hp: 2.9
    attack: 1.67
    defense: 1.0
  normal:
    hp: 2.67
    attack: 1.5
    defense: 2.0
  hard:
    hp: 1.25
    attack: 1.18
    defense: 1.25
  boss:
    hp: 2.0
    attack: 1.5
    defense: 1.67
```

**How it works**:
- Enemy base stats (HP, Attack, Defense) are defined on `EnemyTrait`
- On combat initiation, stats are multiplied by difficulty tier
- Allows fine-tuning balance without editing all NPCs

### Level-gap modifiers
**Config**: `level_gap` in `config/combat_balance.yaml` (defaults in `pkg/mudserver/game/balance/level_gap.go`).

Signed gap = attacker level − defender level, clamped (default ±6). Gap 0 leaves hit, crit, and damage unchanged.

| Knob | Per level of attacker advantage | Where it applies |
| --- | --- | --- |
| `hit_chance` | +3.5% (about +1 on a d20 every one or two levels) | Basic attacks add it to the d20 roll. Natural 1 always misses, natural 20 always hits. Skills have no armor class: a negative delta is a per-hit miss chance. |
| `crit_chance` | +1% | Added to the 5% natural-20 base on basic attacks. Skills have no base crit; only a positive delta can crit (2×). |
| `damage_dealt`, `damage_taken` | +3.5% and +2% | Multiplied together after defense and before crit, then clamped (default 0.40–1.80). Same multiplier on basic attacks, skill hits, and DoT ticks (scaled when the DoT is applied). |

Players and NPCs both use `CombatantRef.Level`, copied from the character or NPC at combat start. A missed damage skill does not apply its secondary effect or Shield Bash stun. Numbers stay in config so the engine stays world-neutral. Feel targets: an enemy 3 levels up is hard in level-appropriate gear and fair in good gear; +5 is a skull fight; −3 or lower feels trivial.

### Class balance
**Config**: `class_balance` in `config/combat_balance.yaml` (defaults in `pkg/mudserver/game/balance/class_balance.go`).

After the level-gap multiplier and before a crit, `damage_dealt` scales hits that class lands and `damage_taken` scales hits that class receives. `behind_dealt` multiplies `damage_dealt` again when that class is the lower level. Class id `wizard` uses the `mage` row. A missing class or a multiplier of 1 leaves that side unchanged. A loaded class pack overrides this file. Ward's pack row takes damage at 1.00, so a soak hit is not larger than the unscaled blow, and the template's stamina is what grows the later hit-point pool. Grit is still gained from a connecting hit. A soak class's Slam prints a miss and starts its cooldown only when the swing hits. The level-10 gap table uses this so warrior, rogue, ranger, and mage share one band: at-level bosses about 50–65%, and a good-gear boss three levels up about 50%.

### Boss telegraph and enrage
**Config**: `boss_mechanics` in `config/combat_balance.yaml`.

Bosses and elites (`hard`) spend `telegraph_turns` actions winding up `telegraph_label` before that hit lands. BattleStage shows a banner and pulses the nameplate for that window (`telegraph_ms`). The resolving hit carries `ability` (Crushing Blow) so the nameplate flash and the floating number are heavier than a normal crit. Bosses enrage after `enrage_after_rounds` or at `enrage_below_hp`, gain an Enraged badge, hit for `enrage_damage`, and stop starting new wind-ups. Trash does not wind up. Elites do not enrage. A miss floats the word "miss". `prefers-reduced-motion` leaves the number in place and skips the flash.

Boss phases are configured by `boss_mechanics.phase_tiers` (defaults to bosses only) and `phases`, an ordered list of `label` / `below_hp` bands. The opening band must be 1.0; later thresholds must descend and remain above zero. Current defaults are Opening (100%), Escalation (66%), and Last Stand (33%). Omit the list to disable phases; explicitly add `hard` to opt elites in. Phase state is per enemy and per encounter (`bossPhase`, `bossPhaseLabel`, `bossPhaseCount`), advances once at or below each threshold on attacks, skills, or DoT, and never rolls back after healing. Large nonlethal hits emit every crossed threshold; lethal hits do not announce a phase.

A phase can override `telegraph_label`, `damage_dealt`, and `enrage_damage`; omitted values inherit the global behavior (phase damage defaults to 1). Phase damage multiplies after class/level scaling and before crits. An enrage override replaces the global enrage multiplier. A6 enrage still triggers at round 16 or 30% HP in any phase and cancels/skips wind-ups. A wind-up already in progress retains its original ability label across a phase transition unless enrage cancels it. Each phase entry sends a structured `combatAction` with action/result/fxId `phase-enter`, human text, and current combatant snapshots to all living participants. BattleStage shows a four-second phase banner and a persistent phase number/name under each boss nameplate; reduced motion disables the arrival animation. Late joiners see the current phase without replaying a transition. Group roster and contextual focus/restore behavior are preserved.

### Threat colors
`threat` in `config/combat_balance.yaml` maps `(enemyLevel - playerLevel)` to `grey / green / yellow / orange / red / skull` (defaults: ≤ −3 grey, −2..−1 green, 0..+1 yellow, +2 orange, +3..+4 red, ≥ +5 skull). The tier is on the room NPC payload (`threat`) and on combat enemy views, computed for the viewer. Room cards and BattleStage nameplates use that color; skull enemies also show ☠. `attack` on orange, red, or skull warns once ("X is much stronger than you") and does not engage. `attack!` or a second `attack` on that enemy does. The room Attack button confirms, then sends `attack!`.

In combat, the gold nameplate and portrait ring mark the local player's focus target. Clicking a living enemy portrait/nameplate or cycling with Tab sends `focus <enemy ID>` to update the engine's auto attack aim without queuing an attack. `attack <name>` also changes focus and queues an attack; the BattleStage Attack button sends the exact enemy ID so duplicate names are unambiguous. The combat start payload identifies the engaged enemy, including for players joining an existing fight. Basic attacks and untargeted hostile skill commands use the focused living enemy; an invalid or defeated focus falls back to the first living enemy. Selecting a different orange, red, or skull foe shows "X is much stronger than you." in its threat color for 3.2 seconds; clicking the same foe does not repeat it. The only hostile is selected automatically, and ending combat clears focus. The warning banner has no motion.

### Equipment and inventory item cards

Clicking or pressing Enter on an equipped paper-doll slot opens the shared item card. Right-click opens it too. The item stays equipped until the Unequip button is pressed; Escape, Close, or clicking outside closes the card. Inventory tiles and rows open the same card with explicit Equip, Use, Use on… (tools/fire-starters), Examine, Sell, and Drop actions. Use on… picks another inventory item and sends `use A on B`; bare Use still works (flint lights a carried torch). The card shows art, rarity, slot, stats, description, value, and weight when the item supplies it. Inventory cards compare stat differences with equipped gear; rings use the weaker worn ring, an empty ring slot counts as a full gain, and two-handed weapons compare with both hand slots. A green arrow marks a clear class-relevant, usable upgrade. Class tags, item level, and explicit armor-weight metadata can block equipping; the card explains the requirement and disables Equip. The `equip` command enforces the same requirements, including when typed in the terminal.

### Viewport layout presets
Combat start/join promotes the existing full-screen BattleStage cover using the layout focus/save contract. The widgets and terminal remain mounted with their prior geometry and active tabs. Victory/defeat dismissal, outcome timeout, and combatLeave restore the previous arrangement, including a manually focused panel; viewport fitting resumes on restore. Save and Save as template retain the normal arrangement during combat. Keyboard focus moves to the stage only when no command input or other text field is active, and returns to the prior control when the stage closes unless the player has focused another field. The combat stage fits Compact and phone viewports without document scrolling.

The Settings panel opens from the account chip and Escape closes it. `interface.battleLayoutB` defaults to true (Layout B). Turning it off selects Classic cards and stores `talesmud_battle_layout_b=0`; a missing key is not an opt-out. `interface.combatAutoFocus` defaults to true; turning it off keeps the room layout visible on combat start/join and provides an explicit Open BattleStage control during active combat. That control uses the same transient focus/restore behavior. `interface.reducedMotion` defaults to `system`; `on` suppresses BattleStage phase, hit, ally, and loot animations plus map ambience, while `off` permits them even when the OS requests reduced motion. `interface.inventoryOpenMode` chooses the existing overlay or layout widget and applies immediately. These choices persist in local storage, not server settings. The General tab labels audio as coming soon because no game audio path consumes the stored sound fields. The old Compact Mode and Room Text Overlay fields remain readable for existing local settings but their controls are hidden because they have no active presentation consumer.

With no usable saved layout, the play client picks Compact (under 1100px wide, room stacked over the terminal), Desktop, or Wide from the window size, and sizes the grid so the widgets fit the viewport height with no page scroll. Empty, unknown, malformed, or wholly hidden saved grids fall back to that preset. Desktop and Wide put the room on the left. On the right, Character and Equipment share one tab container (Character open), Terminal, Quest Log, and Map share another (Terminal open), and Inventory sits under those tabs, with the action bar across the bottom. Nothing in that preset crosses the 24-column grid or another widget. The room scene shrinks to share the panel with the description, and a long description scrolls inside the room. Panel padding and the character sheet are tight enough that attributes and combat stats fit in the Character tab at 1080p. Compact stays a stack and does not add the sheet. A saved layout that is taller than the window is scaled down for display (the action bar stays on the bottom row); Save still writes the player's unscaled rows. Compact and phone may still scroll. The spell bar is docked in the top of the action bar instead of a separate row and has nine slots. Keys 1–9 fire those slots, Tab cycles living combat targets, Escape closes the top open panel (shortcut list, dialogs, map, inventory, the battle outcome, the account menu, a focused widget, then edit mode), and `?` opens the shortcut list. Those keys do nothing while the command line or any text field is focused. Focusing a widget still covers the grid, and Save stores the arrangement from before that cover. Resize reflows that preset. A saved layout is kept, clamped back onto the 24-column grid (minimum 2×2, nothing past the right edge), and scaled vertically when it is taller than the window. A saved full-width spell bar that sits directly on the action bar is folded into that dock on load; a spell bar placed somewhere else stays its own widget and can still be moved in edit mode. Edit mode can switch Compact / Desktop / Wide without deleting a saved layout until Save. Guests open the same editor from the account menu. The toolbar has multi-step Undo, Reset, and a Lock toggle that keeps edit mode open but stops dragging and resizing. Corner handles stay visible while the layout is unlocked, and a gold ghost shows where a widget will land. A guest token in this tab is restored after a reload, so crossing into a mobile-emulation reload does not dump the session back to the welcome screen. Panels share one header (title, collapse, focus) in the same type and padding; the inventory overlay keeps a single title. A tab container uses that same single row: the tabs sit on the left, and collapse and focus stay on the right. Extra tabs scroll sideways and, past three, also open from a More menu. There is no separate "Tab container" title. The account chip, Edit Layout, and the Party and Friends buttons share one header row in that top band: same height, gold border, and gold hover. Edit Layout is its own button until the window is under 1100px wide, where it moves into the account menu. The menu is gold, lines up with the chip's right edge, and closes on Escape or an outside click. On a phone the same menu hangs from the account button in the room header and includes Switch character, Log in / Save progress for guests, and Log out for a signed-in player. Terminal lines wrap on word boundaries inside the panel; a token longer than the row may still break. Resizing the terminal reflows that scrollback.

### Reward scaling
`reward_scale` in `config/combat_balance.yaml` multiplies each enemy's base XP and gold by that threat tier. The reference level is the **highest** level among characters who receive the victory split (living fighters plus same-room online party), so a high-level member greys out the whole award. Defaults: grey 15%, green 60%, yellow 100%, orange 125%, red 150%, skull 200%. A boss's first kill for a character adds `first_kill_bonus` (default 50%) of that character's own share of the boss, once, stored on `Character.FirstBossKills` (`tpl:<templateId>` or `name:<lower name>`). A missing item icon swaps once to `/api/item-art/generic-<type>.png`, then the default generic, then a built-in silhouette. A missing enemy or NPC portrait swaps once to a built-in silhouette (enemies darker, friendly NPCs gold). A player with no portrait file, including a guest on the battle card, uses a class silhouette. Those stand-ins are data URIs, so a failed image cannot loop or stay as a broken icon. `combatEnd` still carries `outcome` and the human `message`. Victory adds `rewards` (base, level modifier, first-kill, share) plus `loot` (`name`, `quality`, `quantity`) and `levelUp` (`oldLevel`, `newLevel`) when a level was gained. Defeat adds `defeat` (`xpLost`, `goldLost`, `armor`, `respawnRoom`, `hp`, `maxHp`). Older clients ignore the extra fields. BattleStage shows that breakdown, reveals each drop one at a time in its rarity color, and calls out the new level. The terminal still gets the full text. The panel sits over the room, dismisses on click, Enter, or Escape, and does not take pointer events away from the terminal. Typing in the command line keeps Enter.

---

## Skills & Spells System

### Skill Entity Structure
```go
type Skill struct {
    *entities.Entity

    Name, Description string
    ClassIDs       []string  // Multi-class: ["warrior"], ["cleric", "druid"]
    LevelRequired  int32

    // Resource System
    ResourceType   ResourceType  // "mana" (casters) or "cooldown" (physical)
    ManaCost       int32         // Mana cost per use
    CooldownRounds int           // Rounds before reuse

    // Targeting & Effects
    Target         TargetType    // "enemy", "self", "all_enemies"
    Effect         EffectType    // "damage", "heal", "buff", "debuff", "dot", "hot"
    ScalingAttr    string        // "STR", "DEX", "INT", "WIS"
    BasePower      int32
    ScalingFactor  float64
    Duration       int           // Rounds (0 = instant)

    // Buff/Debuff Details
    BuffStat       string        // "ATK", "DEF", "STR", "DEX"
    BuffPercent    float64       // 0.30 = +30%

    // Special Mechanics
    IgnoresDefense bool
    HitCount       int           // Multi-hit attacks

    // Secondary Effect (optional, for hybrid skills)
    SecondaryEffect, SecondaryTarget, etc.
}
```

### Skill Storage & Cache
- **Database-backed** with in-memory cache
- `LoadFromDB()` at server startup
- `RefreshCache()` after CRUD operations
- Combat engine reads from cache (zero service threading)

### Skill Registry Functions
```go
SkillByID(id) *Skill                      // Look up by ID
SkillsForClass(classID) []*Skill          // All skills for class
AvailableSkills(classID, level) []*Skill  // Skills at level
MaxSkillSlots(classID, level) int         // Equippable slots
IsCasterClass(classID) bool               // Uses mana
```

### Skill Slot Progression
A class with `hotbar_cap` in the class catalog always has that many slots (the signed pack uses 4). Cleric and druid stay on the caster curve (L1=2, L15=3, L30=4). Ranger stays on the physical curve (L1=1, L10=2, L20=3, L30=4). Stored ids such as hitch still resolve through the catalog.

### Class kit
Class names, blurbs, race lists, portraits, and kit skills live in the world pack under `data/classes/*.yaml`. The engine loads them into `pkg/classkit` at startup and on import. `GET /api/classes` is the public payload the play client reads (`?v=classkit1`). With no pack, the catalog is the generic Warrior / Rogue / Mage sample. `ClassKit()` builds skill rows from that catalog. `SkillsForClass` drops legacy seed rows for a kit class. Cleric, ranger, and druid stay in the static client catalog. Old ids such as Fireball and Power Strike remain display fallbacks for a saved hotbar and are not offered as Available. `config/combat_balance.yaml` `class_balance` is only the numeric fallback when the catalog has no row.

### Skill Management Commands
```bash
skills                  # List available and equipped skills
skills equip <name>     # Equip skill to next slot
skills unequip <name>   # Remove skill from slots
```

**Combat Restrictions**: Cannot equip/unequip during combat

### Default Skills Seeding
An empty database seeds `SeedSkills()` (legacy class rows plus `ClassKit`). An existing database upserts the class kit on startup. Kit classes only see kit skills. Cleric, ranger, and druid still use their seeded rows.

---

## Quest System

### Quest Entity Structure
```go
type Quest struct {
    *entities.Entity

    Name, Description string
    Category    string  // "main", "side", "daily"
    Area        string  // Optional region/zone label
    Level       int32
    Repeatable  bool

    Source      QuestSource  // "npc", "item", "auto", "script"
    Objectives  []Objective
    Rewards     Reward

    RequiredQuestIDs []string  // Prerequisite quests
    RequiredLevel    int32

    // NPC Dialog Integration
    AcceptDialogText, ProgressDialogText, CompleteDialogText string
    OnCompleteScriptID string  // Lua script on completion
}
```

### Quest Source Types
```go
type QuestSource struct {
    Type   string  // "npc", "item", "auto", "script"
    NPCID  string  // Quest-giving NPC
    ItemID string  // Quest-triggering item
}
```

### Objective Types
```go
type Objective struct {
    ID          string
    Type        ObjectiveType  // "kill", "collect", "deliver", "visit", "talk", "custom"
    Description string

    // Type-specific fields
    TargetID        string  // NPC/item/room target depending on type
    TargetName      string
    Amount          int32
    DeliverToNPCID  string  // For deliver
    DialogNodeID    string  // Optional for talk
    CheckScriptID   string  // For custom (Lua)
    Order           int32
}
```

### Quest Progress (Per-Character)
```go
type QuestProgress struct {
    *entities.Entity

    CharacterID string
    QuestID     string
    Status      QuestStatus  // "available", "active", "completed", "failed", "abandoned"
    Objectives  []ObjectiveProgress
    AcceptedAt, CompletedAt time.Time
}
```

### Quest Rewards
```go
type Reward struct {
    XP    int64
    Gold  int64
    Items []RewardItem  // Item template IDs with quantities
}
```

### Quest Definition Validation
Quest create/update validates impossible definitions before persistence:
- Required name, description, source type, and at least one objective
- Required source references for NPC/item sources
- Required objective IDs, descriptions, target fields, and custom scripts
- Duplicate objective IDs, negative objective amounts, negative rewards
- Empty reward/prerequisite IDs and self-prerequisites
- Missing referenced NPCs, item templates, rooms, scripts, and prerequisite quests

### Quest Tracker (Automatic Progress)
The `QuestTracker` listens to game events and updates objectives:

| Event | Source | Objectives Updated |
|-------|--------|-------------------|
| NPC killed | Combat victory | Kill objectives (matches template ID) |
| Item pickup | Pickup command | Collect objectives (matches template ID and stack quantity) |
| Room enter | Room navigation | Visit objectives (matches room ID) |
| Dialog node | Talk command | Talk objectives (matches NPC + node) |
| Talk to NPC | Talk command | Deliver objectives (requires and consumes matching item quantity) |

`QuestsService.ApplyQuestEvent` provides the same normalized event matching for service-level workflows and tests, returning progress or ready-to-turn-in results. `QuestsService.TurnInQuest` validates the turn-in NPC, completes the quest, grants rewards, and returns the reward summary in one call.

### Quest Dialog Integration
When talking to a quest-source NPC:
1. **Offer option** - If quest is available and prerequisites met
2. **Progress option** - If quest is active but not complete
3. **Turn-in option** - If all objectives complete

Dialog options are **automatically injected** into NPC conversations.
Quest-source NPCs do not need a full dialog tree: if an NPC has quest options but no `DialogID`, the server opens a quest-only conversation and numeric selection accepts, checks, or turns in the quest.

### Quest Commands
```bash
quests          # Show quest log (active quests + progress)
ql, questlog    # Aliases

quest <name>    # Show quest details

abandon <name>  # Abandon active quest
```

### Quest UI Features

#### Quest Log Widget
The quest log widget provides a comprehensive quest management interface:

**Core Features:**
- **Real-time Quest Search**
  - Search across quest names, descriptions, and objectives
  - Case-insensitive, instant filtering
  - Clear button (×) to reset search

- **Category Filtering**
  - Filter by: All Types, Main, Side, Daily
  - Toggle visibility: Completed quests, Abandoned quests
  - Helps focus on relevant quests

- **Sort Options**
- Sort by Status (Ready → Active → Completed → Abandoned)
  - Sort by Name (alphabetical)
  - Sort by Level (ascending)
  - Sort by Category (Main/Side/Daily)

- **Quest Pinning System**
  - Pin up to 5 priority quests
  - Pinned section always at top
  - Golden pulsing indicator for pinned quests
  - Persisted in localStorage
  - Pin/Unpin buttons in quest details

**Quest Display:**
- Quest name with level badge (e.g., "L5")
- Ready-to-turn-in badge and highlighted row when all active objectives are complete
- Category badge with color coding:
  - 🟠 Main quests (amber/orange #f59e0b)
  - 🔵 Side quests (blue #3b82f6)
  - 🟣 Daily quests (purple #8b5cf6)
- Expandable quest details
- Objective progress (X/Y format) with checkmarks
- Objective descriptions are populated from quest definitions in both REST quest-log refreshes and WebSocket quest-log messages
- Rewards preview (XP, Gold, Items)
- Abandon button for active quests
- Pin/Unpin button for tracking

**Quest Sections:**
1. **📌 Pinned** - Priority quests (if any)
2. **Active** - In-progress quests
3. **Completed** - Finished quests with completion dates
4. **Abandoned** - Dropped quests (when visible)
5. **Failed** - Failed quests (when visible)

#### Quest History & Statistics Panel

Click the 📊 button in quest log header to access:

**Quest Statistics:**
- Total quests encountered
- Active quest count
- Completed quest count
- Completion rate percentage
- Total XP earned from quests
- Total gold earned from quests

**Category Breakdown:**
- Main quests completed
- Side quests completed
- Daily quests completed

**Quest Achievements:**
8 built-in achievements that unlock automatically:

| Achievement | Requirement |
|-------------|-------------|
| First Steps | Complete 1 quest |
| Quest Novice | Complete 5 quests |
| Quest Veteran | Complete 10 quests |
| Quest Master | Complete 25 quests |
| Story Seeker | Complete 5 main quests |
| Side Quest Hero | Complete 10 side quests |
| Daily Devotee | Complete 5 daily quests |
| Completionist | 100% completion rate (min 5 quests) |

**Achievement Display:**
- 🏆 Unlocked achievements (gold border, highlighted)
- 🔒 Locked achievements (grayed out, collapsible)
- Shows name and description
- Real-time progress tracking

#### Quest Notifications

Enhanced notification system with interactions:

**Notification Types:**
- **Quest Accepted** / **Quest Complete** — Veilspan centered moment cards (amber/brass), not top-right chips
- **Quest Progress** (blue border) - shows the changed objective and current/required counts
- **Quest Ready** (yellow border) - shown when all objectives are complete and the quest can be turned in

**Interactions:**
- **Click to View** - Opens quest in quest log
- **Dismiss Button (×)** - Manual dismiss with slide-out animation
- **Hover Effects** - Highlights notification
- **Auto-dismiss** - Corner toasts after 5s; accepted banner after 10s

**Features:**
- Unique notification IDs
- Notification queue capped to the latest 4 visible events
- Slide-in and slide-out animations
- Smooth transitions
- Accepting a quest refreshes the open Talk dialog so `[Quest]` becomes `[In Progress]` without closing Talk
- Quest log: READY + turnInAnywhere shows **Turn In** (`complete <name>`); otherwise a **Turn in: &lt;NPC&gt;** hint

### Spell Bar / Hotbar
- 8 square slots docked on top of the desktop action bar (a placed hotbar widget still floats on its own). Mobile keeps its own strip
- Bind equipped combat skills (`cast` / combat-only) or inventory consumables (`use`)
- Look / Talk / Flee are bindable actions but **not** seeded by default
- Rest is seeded on an empty/default hotbar (slot 7). Customized binds are never overwritten
- Out of combat, a **Resting** chip shows near HP while `Flags.resting` is true; combat or movement clears it
- Search is not a look alias and is not offered as a hotbar action
- Binds persist in `talesmud_settings_v1` (`interface.hotbarBinds`); empty equipped list shows "spellbook empty"
- Does not use or overload the respawn `bind` command

### Quest API Endpoints

#### Get Quest Log with Full Details
```
GET /api/quest-progress/:characterId
```

Requires the authenticated user to own `characterId`, unless the user is an admin.

**Response includes:**
- Quest progress (status, objectives)
- Quest definition (name, description, category, level, objective descriptions)
- Rewards (XP, gold, item template IDs)
- Timestamps (acceptedAt, completedAt)

**Response Format:**
```json
[
  {
    "questId": "quest-uuid",
    "questName": "The Wolf Problem",
    "status": "active",
    "description": "Wolves have been attacking travelers...",
    "category": "main",
    "level": 3,
    "objectives": [
      {
        "objectiveId": "obj-1",
        "description": "Defeat wolves",
        "current": 3,
        "required": 5,
        "completed": false
      }
    ],
    "rewards": {
      "xp": 500,
      "gold": 50,
      "itemTemplateIds": ["item-template-1"]
    },
    "acceptedAt": "2026-02-16T10:30:00Z",
    "completedAt": null
  }
]
```

### Quest WebSocket Messages

#### Quest Log Message
Sent on character selection and quest updates:
```json
{
  "type": "questLog",
  "quests": [ /* array of QuestLogEntry */ ]
}
```
Dialog-driven quest accepts and completions send the same enriched quest entries as character selection, including quest name, description, category, level, objective descriptions, rewards, and timestamps when available.

#### Quest Update Messages
```json
{
  "type": "questAccepted",
  "questId": "quest-uuid",
  "questName": "The Wolf Problem",
  "message": "Quest accepted!"
}

{
  "type": "questProgress",
  "questId": "quest-uuid",
  "questName": "The Wolf Problem",
  "objectives": [ /* updated objectives */ ],
  "changedObjective": {
    "objectiveId": "obj-1",
    "description": "Defeat wolves",
    "current": 5,
    "required": 5,
    "completed": true
  }
}

{
  "type": "questReady",
  "questId": "quest-uuid",
  "questName": "The Wolf Problem",
  "objectives": [ /* all objectives complete */ ],
  "changedObjective": { /* last changed objective */ }
}

{
  "type": "questCompleted",
  "questId": "quest-uuid",
  "questName": "The Wolf Problem",
  "message": "Quest completed!"
}
```

### Quest Client Store

The MUD client stores quest data in `MUDXPlusStore`:

**State:**
```javascript
{
  quests: [],              // Full quest log with details
  questNotifications: [],  // Active notifications
  pinnedQuests: []        // Stored in localStorage
}
```

**Methods:**
- `updateQuests(questLog)` - Update full quest list
- `addQuestNotification(notification)` - Add notification with auto-dismiss
- Quest notifications auto-dismiss after 5 seconds and keep only the newest 4 visible entries

**LocalStorage:**
- `pinnedQuests` - JSON array of quest IDs (max 5)
- Persists across sessions
- Shared between widget and overlay

### Quest Tracker Implementation

Automatic progress tracking on game events:

**OnNPCKilled(characterID, userID, deadNPC):**
- Checks all active quests for kill objectives
- Matches NPC template ID
- Increments objective counter
- Sends progress update message

**OnItemPickup(characterID, userID, item):**
- Checks collect objectives
- Matches item template ID
- Increments by picked-up stack quantity when present
- Sends progress update

When a collect quest is accepted, matching items already in inventory initialize objective progress. Stackable quantities count toward the initial objective amount.

**OnRoomEnter(characterID, userID, room):**
- Checks visit objectives
- Matches room ID
- Marks objective complete
- Sends progress update

**OnDialogNode(characterID, userID, npcID, dialogID, nodeID):**
- Checks talk objectives
- Matches NPC and optional dialog node
- Marks complete
- Sends progress update

**OnTalkToNPC(characterID, userID, npc):**
- Checks deliver objectives
- Verifies player has required item quantity
- Consumes matching items from inventory
- Marks objective complete
- Sends progress update

### Quest Completion Flow

1. **Check Objectives** - All must be complete (checked server-side)
2. **Mark Complete** - Update quest progress status
3. **Grant Rewards**
   - Add XP to character
   - Add gold to character
   - Create items from templates
   - Add to character inventory
4. **Send Notification** - WebSocket message with rewards
5. **Update Quest Log** - Client refreshes via API call
6. **Update Achievements** - Calculated client-side on quest log update
```

### Quest Scripting API
```lua
-- Check quest status
local status = tales.quests.getStatus(characterID, questID)

-- Accept quest
tales.quests.accept(characterID, questID)

-- Complete quest
tales.quests.complete(characterID, questID)

-- Update objective progress
tales.quests.updateProgress(characterID, questID, objectiveIndex, count)

-- Grant quest items
tales.quests.grantItems(characterID, questID)

-- Abandon quest
tales.quests.abandon(characterID, questID)

-- Check if quest can be accepted
local canAccept = tales.quests.canAccept(characterID, questID)
```

---

## Dialog System

### Dialog Entity Structure
```go
type Dialog struct {
    ID   string
    Text string  // Primary speech

    // Text Variations
    AlternateTexts []string  // Random variations
    OrderedTexts   *bool     // Sequential on repeat

    // Dialog Flow
    Options []Dialog  // Player choices (branching)
    Answer  *Dialog   // Auto-response (linear)

    // Conditional Display
    RequiresVisitedDialogs []string  // Prerequisites
    ShowOnlyOnce           *bool     // One-time option
    IsDialogExit           *bool     // End conversation
}
```

### Dialog State (Per-Conversation)
```go
type DialogState struct {
    CurrentDialogID  string
    DialogVisited    map[string]int  // Node ID → visit count
    Context          map[string]string  // Static variables
    DynamicContext   map[string]func() string  // Runtime variables
}
```

### Dialog Variable Substitution
```
{{PLAYER}}   → Character name
{{NPC}}      → NPC name
{{TIME}}     → Current time
{{CUSTOM}}   → Script-set context variables
```

### Dialog Features
- **Branching** - Multiple player choices via `Options`
- **Linear** - Automatic progression via `Answer`
- **Conditional** - Options shown only if prerequisites visited
- **One-time** - `ShowOnlyOnce` options disappear after first selection
- **Variations** - Random or sequential alternate texts
- **Exit markers** - Return to "main" node on completion

### Dialog API
```lua
-- Get dialog
local dialog = tales.dialogs.get(dialogID)

-- Get conversation state
local conv = tales.dialogs.getConversation(characterID, npcID)

-- Set context variable
tales.dialogs.setContext(conversationID, "playerClass", "warrior")

-- Get context variable
local class = tales.dialogs.getContext(conversationID, "playerClass")

-- Check if visited
local visited = tales.dialogs.hasVisited(conversationID, nodeID)

-- Get visit count
local count = tales.dialogs.getVisitCount(conversationID, nodeID)
```

---

## Scripting System (Lua API)

### Lua Runner Architecture
- **Language**: Lua 5.1 (via gopher-lua)
- **VM Pool**: Reusable Lua states for performance
- **Timeout**: 5-second execution limit
- **Sandbox**: Restricted environment (no file I/O, no network)

### Script Types
- `item` - Item creation, behavior, OnUse handlers
- `room` - Room actions, OnEnter handlers
- `npc` - NPC behavior, aggro/death/flee events
- `quest` - Quest logic, custom objectives
- `event` - Event handlers (player.enter_room, npc.death, etc.)
- `custom` - General purpose

### Context Variable (`ctx`)
Scripts receive context data via global `ctx`:
```lua
-- Room enter script:
ctx.eventType   -- "player.enter_room"
ctx.room        -- Room entity
ctx.character   -- Character entity
ctx.user        -- User entity

-- Item use script:
ctx.eventType   -- "item.use"
ctx.item        -- Item entity
ctx.character   -- Character entity
ctx.room        -- Room entity (if in room)

-- Room action script:
ctx.eventType   -- "room.action"
ctx.room        -- Room entity
ctx.character   -- Character entity
ctx.action      -- Action entity
```

### tales.game Module

All functions below are called as `tales.game.functionName(...)`. Every parameter is **required** unless noted otherwise. Functions that perform writes return `bool` (true on success, false on failure). Missing or wrong-type arguments cause a **Lua error that aborts the script**.

#### Messaging
```lua
tales.game.msgToRoom(roomID, message)
-- Send message to ALL players in a room. Returns bool.
-- roomID: string (room entity ID, e.g., "R0004")
-- message: string

tales.game.msgToCharacter(characterID, message)
-- Send message to a specific character (via their user's socket). Returns bool.
-- characterID: string (character entity ID)
-- message: string

tales.game.msgToUser(userID, message)
-- Send message to a specific user by user ID. Returns bool.
-- userID: string (user entity ID — NOT character ID)
-- message: string

tales.game.broadcast(message)
-- Send message to ALL connected players globally. Returns bool.
-- message: string

tales.game.msgToRoomExcept(roomID, message, excludeCharacterID)
-- Send message to all players in room EXCEPT the specified character. Returns bool.
-- roomID: string
-- message: string
-- excludeCharacterID: string (character entity ID to exclude)
```

#### Logging
```lua
tales.game.log(level, message)
-- Write to server log. Returns nothing.
-- level: string — "debug", "info", "warn", "error"
-- message: string
```

#### Inventory & Equipment Checks
```lua
tales.game.hasItem(characterID, itemID)
-- Check if character has an item in inventory. Returns bool.
-- Checks by exact item ID first, then by TemplateID (so you can pass "ITM0001").
-- characterID: string
-- itemID: string (item instance ID or template ID)

tales.game.hasEquipped(characterID, slotName)
-- Check if character has an item equipped in a specific slot. Returns bool.
-- characterID: string
-- slotName: string — "head", "chest", "legs", "boots", "hands", "neck",
--   "ring1", "ring2", "main_hand", "off_hand"
```

#### Character Flags (Per-Character State)
```lua
tales.game.getFlag(characterID, flagName)
-- Get a flag value from a character. Returns the value or nil if not set.
-- characterID: string
-- flagName: string
-- Return types: bool, number, string, or nil

tales.game.setFlag(characterID, flagName, value)
-- Set a flag on a character. Persisted to database immediately. Returns bool.
-- characterID: string
-- flagName: string
-- value: bool, number, string, or nil (nil deletes the flag)
```

#### Hidden Exit Reveals (Per-Character)

**CRITICAL**: Note the different parameter orders between `revealExit` and `hasRevealedExit`.

```lua
tales.game.revealExit(roomID, exitName, characterID)
-- Reveal a hidden exit for a specific character. Returns bool.
-- The exit must already exist on the room with hidden=true.
-- The reveal is stored on the character (Character.RevealedExits), not the room.
-- Sends a silent room update to the client (no full re-render).
-- Idempotent: safe to call multiple times (returns true if already revealed).
-- ALL THREE PARAMETERS ARE REQUIRED.
-- roomID: string (room entity ID where the exit is defined)
-- exitName: string (exit name, case-insensitive match, e.g., "north")
-- characterID: string (character entity ID to reveal the exit for)

tales.game.hasRevealedExit(characterID, roomID, exitName)
-- Check if a character has revealed a specific hidden exit. Returns bool.
-- NOTE: Parameter order is (characterID, roomID, exitName) — different from revealExit!
-- characterID: string
-- roomID: string
-- exitName: string (case-insensitive match)
```

#### Item Rewards
```lua
tales.game.giveItem(characterID, templateID)
-- Create an item instance from a template and add it to the character's inventory.
-- Returns bool. Persisted to database immediately.
-- characterID: string
-- templateID: string (item template ID, e.g., "ITM0022")
```

#### CopyOnPickup Item Tracking
```lua
tales.game.hasCollectedItem(characterID, templateID)
-- Check if a character has already collected a CopyOnPickup item. Returns bool.
-- characterID: string
-- templateID: string (item template ID)

tales.game.resetCollectedItem(characterID, templateID)
-- Reset the collected flag for a CopyOnPickup item, allowing re-collection. Returns bool.
-- characterID: string
-- templateID: string (item template ID)
```

### tales.combat Module (enemy hooks)

Callable from sandboxed scripts. Failures return a zero result and do not abort combat.

```lua
tales.combat.healNpc(npcID, amount)       -- HP restored (0 if none)
tales.combat.applyEffect(targetID, effectID) -- existing buff or debuff only
tales.combat.summon(templateId, count)    -- enemy-template adds spawned (0 on failure)
```

`summon` uses the running enemy-hook fight and room. `count` clamps to 1..3 and to whatever remains of a fight cap of 3. Adds copy template stats, are not charged to a spawner, grant no loot, gold, XP, or quest credit, and are removed when the fight ends (win, lose, flee, or timeout). Each add may run its own `onAggroScript` once. When any combatant dies or flees, they leave the turn order and the current index stays inside the remaining order. If that index is past the end, the next tick wraps into the following round and ends the fight when it is over, so an add dying cannot stall the fight until the idle timeout.

### tales.items Module
```lua
-- Get item
local item = tales.items.get(itemID)

-- Find by name
local items = tales.items.findByName("sword")

-- Templates
local template = tales.items.getTemplate(templateID)
local templates = tales.items.findTemplates("sword")

-- Create instance
local newItem = tales.items.createFromTemplate(templateID)

-- Delete
local success = tales.items.delete(itemID)
```

### tales.rooms Module
```lua
-- Get room
local room = tales.rooms.get(roomID)

-- Find rooms
local rooms = tales.rooms.findByName("tavern")
local rooms = tales.rooms.findByArea("dungeon")
local allRooms = tales.rooms.getAll()

-- Room contents
local characters = tales.rooms.getCharacters(roomID)
local npcs = tales.rooms.getNPCs(roomID)
local items = tales.rooms.getItems(roomID)
```

### tales.characters Module
```lua
-- Get character
local char = tales.characters.get(characterID)

-- Find characters
local chars = tales.characters.findByName("hero")
local allChars = tales.characters.getAll()

-- Get character's room
local room = tales.characters.getRoom(characterID)

-- Character operations
tales.characters.damage(characterID, amount)
tales.characters.heal(characterID, amount)
tales.characters.teleport(characterID, roomID)
tales.characters.giveXP(characterID, amount)

-- Signed gold change. A debit that would go below zero is refused and returns false.
tales.characters.addGold(characterID, delta)

-- Set the respawn room. An empty room id clears it. Unknown rooms return false.
tales.characters.setBind(characterID, roomID)

-- Apply levels the current XP can already buy. Returns how many levels were gained.
tales.characters.applyLevels(characterID)

-- Set level and XP directly. Level clamps to 1..the effective cap. Optional maxHP replaces hit points.
-- Class, skills, inventory, gold, and flags are left alone.
tales.characters.setProgress(characterID, level, xp, maxHP)

-- Read-only top list. n defaults to 12 and is capped at 50.
-- sortKey "xp" orders by experience. Any other key orders by level, then experience.
local rows = tales.characters.top(n, sortKey) -- rows[i].name, .level, .xp
```

### tales.resources Module

Balances for configured keys. An unknown key returns `ok=false` and does not create a row.

```lua
local allowance, remaining, ok = tales.resources.get(characterID, key)
local remaining, ok = tales.resources.consume(characterID, key, n)
```

`consume` with `n <= 0` returns the current remaining and `ok=true` when the key exists. Spending more than `remaining` leaves the balance unchanged and returns `ok=false`.

### tales.npcs Module
```lua
-- Get NPC
local npc = tales.npcs.get(npcID)

-- Find NPCs
local npcs = tales.npcs.findByName("guard")
local npcs = tales.npcs.findInRoom(roomID)
local allNPCs = tales.npcs.getAll()

-- NPC operations
tales.npcs.damage(npcID, amount)
tales.npcs.heal(npcID, amount)
tales.npcs.moveTo(npcID, roomID)

-- NPC checks
local isDead = tales.npcs.isDead(npcID)
local isEnemy = tales.npcs.isEnemy(npcID)
local isMerchant = tales.npcs.isMerchant(npcID)

-- Template & instance operations
local templates = tales.npcs.getTemplates()
local isTemplate = tales.npcs.isTemplate(npcID)
local instance = tales.npcs.spawnFromTemplate(templateID, roomID)
local inst = tales.npcs.getInstance(instanceID)
local instances = tales.npcs.getInstancesInRoom(roomID)

-- Instance operations
tales.npcs.kill(instanceID)
tales.npcs.setState(instanceID, state)  -- "idle", "combat", "patrol", "fleeing"
local state = tales.npcs.getState(instanceID)
local died = tales.npcs.damageInstance(instanceID, amount)
tales.npcs.healInstance(instanceID, amount)
tales.npcs.moveInstance(instanceID, roomID)

-- Delete
tales.npcs.delete(npcID)
```

### tales.dialogs Module
```lua
-- Get dialog
local dialog = tales.dialogs.get(dialogID)

-- Find dialogs
local dialogs = tales.dialogs.findByName("greeting")
local allDialogs = tales.dialogs.getAll()

-- Conversation management
local conv = tales.dialogs.getConversation(characterID, npcID)
tales.dialogs.setContext(conversationID, "key", "value")
local value = tales.dialogs.getContext(conversationID, "key")

-- Visit tracking
local visited = tales.dialogs.hasVisited(conversationID, nodeID)
local count = tales.dialogs.getVisitCount(conversationID, nodeID)
```

### tales.quests Module
```lua
-- Quest status
local status = tales.quests.getStatus(characterID, questID)

-- Quest operations
tales.quests.accept(characterID, questID)
tales.quests.complete(characterID, questID)
tales.quests.abandon(characterID, questID)

-- Progress tracking
tales.quests.updateProgress(characterID, questID, objectiveIndex, count)

-- Rewards
tales.quests.grantItems(characterID, questID)

-- Checks
local canAccept = tales.quests.canAccept(characterID, questID)
```

### tales.utils Module
```lua
-- Random numbers
local num = tales.utils.random(1, 100)  -- Inclusive
local f = tales.utils.randomFloat()     -- 0.0-1.0

-- UUID
local id = tales.utils.uuid()

-- Time
local now = tales.utils.now()           -- Unix timestamp
local nowMs = tales.utils.nowMs()       -- Milliseconds
local str = tales.utils.formatTime(timestamp)

-- Dice rolling
local result = tales.utils.roll("2d6+3")

-- Probability
local success = tales.utils.chance(25)  -- 25% chance → true/false

-- Array utilities
local picked = tales.utils.pick({"sword", "axe", "spear"})
local shuffled = tales.utils.shuffle(myArray)

-- Math utilities
local clamped = tales.utils.clamp(value, 0, 100)
local lerped = tales.utils.lerp(0, 100, 0.5)  -- Returns 50
```

### Script Execution Contexts

Each script type receives a `ctx` global table with different fields. Access fields as `ctx.fieldName`. Object fields expose entity properties (e.g., `ctx.character.ID`, `ctx.room.Name`).

**Room OnEnter Script** (`type: room`, assigned via room's `onEnterScript`):
- Triggered when player enters room (walk, teleport, or character select)
- Context variables:
  - `ctx.eventType` — string: `"player.enter_room"`
  - `ctx.room` — Room entity object (access `.ID`, `.Name`, etc.)
  - `ctx.toRoom` — Room entity object (same as `ctx.room`)
  - `ctx.character` — Character entity object (access `.ID`, `.Name`, etc.)
  - `ctx.user` — User entity object
- **Getting room ID**: Use `ctx.room.ID`
- **Getting character ID**: Use `ctx.character.ID` or store as `local charID = character.ID`

**Room Action Script** (`type: custom`, assigned via room action's `scriptId`):
- Triggered by player room action (e.g., `EXAMINE RUNES`, `PRESS CIRCLE`)
- Context variables:
  - `ctx.eventType` — string: `"room.action"`
  - `ctx.room` — Room entity object
  - `ctx.roomID` — string: room ID (convenience shortcut for `ctx.room.ID`)
  - `ctx.character` — Character entity object
  - `ctx.characterID` — string: character ID (convenience shortcut for `ctx.character.ID`)
  - `ctx.action` — string: action name (e.g., `"PRESS CIRCLE"`)
  - `ctx.params` — table: action params from YAML (e.g., `{action="press", symbol="circle"}`)
- **Note**: Room action scripts get BOTH object and string ID shortcuts (`ctx.roomID` and `ctx.room.ID` both work)

**Item OnUse Script** (`type: item`, assigned via item's `onUseScriptID`):
- Triggered by `use <item>` command
- Context variables:
  - `ctx.eventType` — string: `"item.use"`
  - `ctx.item` — Item entity object
  - `ctx.character` — Character entity object
  - `ctx.room` — Room entity object (if character is in a room)

**NPC Event Scripts** (`type: npc`):
- `OnAggroScript`, `OnDeathScript`, `OnFleeScript`, `OnLowHealthScript`
- `LowHealthThreshold` is a fraction of max HP. `0` or unset means `0.30`. A killing blow from above the line does not run the low-health script.
- Context varies by event type. Enemy hooks set `ctx.hook`, `ctx.roomId`, `ctx.npc`, `ctx.opponents`, and `ctx.allies`.
- `tales.combat.summon(templateId, count)` spawns enemy-template adds into the current fight. Count clamps to 1..3 and to the remaining fight cap of 3. Adds grant no loot, gold, XP, or quest credit and are removed when the fight ends. Unknown template, non-enemy, no fight, or a full cap returns 0.
- Room lines from `tales.game.msgToRoom` and `tales.game.msgToRoomExcept` during an enemy hook stay `type: "message"` and `username: "SYSTEM"`. They also carry `style: "combatEvent"`, `hook` (`onAggro`, `onLowHealth`, `onDeath`, `onFlee`), and `source` (that NPC's display name). Other scripts omit those fields. A hook flush that summoned adds sends `combatStatus` with the live `combatants` roster immediately.
- The play client renders `style === "combatEvent"` as a gold chip in the BattleStage combat log (icon per hook, `source` as a small label) and still prints the plain line. Hook `unique` is the room announcement for a one-per-character drop, sent by the loot path rather than a Lua hook. Cache-bust for that client is `?v=uniques1`.

### Context Variable Quick Reference

| Variable | OnEnter (room) | Action (custom) | Item Use |
|----------|:-:|:-:|:-:|
| `ctx.eventType` | `"player.enter_room"` | `"room.action"` | `"item.use"` |
| `ctx.room` | Room object | Room object | Room object |
| `ctx.roomID` | — | string shortcut | — |
| `ctx.character` | Character object | Character object | Character object |
| `ctx.characterID` | — | string shortcut | — |
| `ctx.user` | User object | — | — |
| `ctx.action` | — | action name string | — |
| `ctx.params` | — | action params table | — |
| `ctx.item` | — | — | Item object |

**Important**: In OnEnter scripts, you must use `ctx.room.ID` and `ctx.character.ID` (no string shortcuts). In Action scripts, you can use either `ctx.roomID` or `ctx.room.ID`.

---

## Creator UI Capabilities

### Data Table System
All entity editors use a unified **filterable, sortable data table**:
- **Per-column filtering** - Text search, enum dropdowns
- **Client-side filtering** - Instant, no API calls
- **Sorting** - Click column headers (asc → desc → none)
- **Master-detail layout** - Table + edit form side-by-side
- **Full-width mode** - Close detail panel to maximize table view

### Entity Select Modal
**MANDATORY UI GUIDELINE**:
- **NEVER use `<select>` dropdowns for entity ID references**
- **ALWAYS use `EntitySelectButton` + `EntitySelectModal`**
- Provides filterable DataTable for selecting rooms, NPCs, items, scripts, dialogs, quests, and loot tables
- Scales to hundreds of entries with search and filter

```svelte
<EntitySelectButton
  value={roomID}
  elements={rooms}
  columns={roomColumns}
  title="Select Room"
  placeholder="Select a room..."
  on:change={(e) => roomID = e.detail}
/>
```

### Creator Validation And Diagnostics
Creator editors share backend validation rules from `pkg/service/validation`:
- Inline validation panels show errors and warnings for the selected draft entity.
- Save/update requests for rooms, items, NPCs, NPC spawners, dialogs, loot tables, quests, and scripts reject error-severity broken references before data is stored.
- World Health runs cross-world diagnostics and reports structured issues with `severity`, `entityType`, `entityId`, `field`, `code`, and `message`. The message column is visible and wraps. Room, NPC, dialog, quest, item, and script IDs link to that editor with `?id=`.
- Content health (`GET /api/health`, `tales -check`) adds reachability islands, reveal scripts that name a missing exit, hidden exits with no revealer, quest objectives that cannot complete, bosses with no spawner, unknown difficulty tiers, unreferenced scripts, dangling exits, missing room items, and pack rules. A quest reward counts as an obtainable item when that quest can complete and its prerequisites can precede the quest that needs the item. Reachability starts at `startRoomID` when that room exists, otherwise the engine default. The content commit is that folder's own git HEAD, or `CONTENT_COMMIT` / `.content-commit`, or `unknown`. Live health warns with `deploy-tree-dirty` when the server checkout has tracked changes outside `import/`. `tales -check` does not run that rule. Hits can be muted. Drift compares the database with the last import. See `docs/content-health.md`.
- Dialog validation allows a node ID to be reached more than once, including shared branches and cycles. It reports `duplicate_dialog_node` only when two definitions of the same node disagree. Option edges that point at a node are not a second definition.
- Merchant stock with `maxQuantity` -1 is unlimited. Only values below -1 are warned.
- Deletes that call a DELETE API open a confirm dialog (entity type, name, and ID) before the request. The same dialog covers map room delete, saved spawner delete, room item removal, special exit removal, dialog-graph option and answer removal, and bulk deletion of deprecated scripts. Unsaved form rows (patrol stops, quest objectives, cardinal exits, room actions, alternate texts) stay immediate.
- Data tables show the full entity ID, with a tooltip and a copy button. The row dot marks validation errors and warnings. NPC type (Enemy, Merchant, both, Neutral) is a text badge.
- Opening a creator editor URL with `?id=<entityId>` selects that entity after the list loads. The dialog graph honors the same query.
- Ctrl+K or Cmd+K opens a search palette over rooms, NPCs, items, dialogs (including node text), quests, scripts, loot tables, spawners, skills, and character templates. Results stay grouped by type. Enter opens that entity's Creator path. The shortcut is ignored inside an input, a textarea, or the script editor. The top-bar search box (`data-creator-search`) opens the same palette and shares its query.
- Rooms, NPCs, items, dialogs, quests, and scripts show a collapsed "Referenced by (n)" panel under the form. Deleting one of those entities warns when something still references it.
- Data table rows surface warning/error indicators when stored entities have diagnostics. An editor-specific row indicator is not used for NPC type.
- Preview/test tools validate draft dialogs, quests, rooms, merchants, and Lua scripts before content is published into the world.

### Creator navigation
The Creator screens sit in a left sidebar instead of a horizontal tab row. Groups are World (Rooms, World / Zones, Health), Actors (NPCs, Character Templates, Items), Narrative (Dialogs, Dialog Graph, Quests), Systems (Skills, Scripts, Settings), and Operate (Players, Audit log). Players stays admin-only. Audit log stays visible to creators. `/creator` still opens Rooms. `?id=` links are unchanged. `/creator/quests/debug` stays on the Quests item.

The sidebar collapses to an icon rail from its toggle or Ctrl+B / Cmd+B. The shortcut is ignored in inputs, textareas, selects, contenteditable editors (the script editor), and CodeMirror. Collapse and per-group folds persist in `localStorage` (`tales.creator.nav.v1`). The group that contains the current page opens on navigation. Below 1024px the rail is an off-canvas drawer: the top bar shows a menu button, and the backdrop, Escape, or a navigation click closes it.

The top bar keeps the app links, a search box (`data-creator-search`) that opens the search palette, and the red LIVE badge when `ADMIN_ENV_LABEL` is set. Below 1024px the text links and the search box are hidden so the badge stays on one row. Ctrl+K still opens the palette.

### Creator screens
1. **Rooms** - Full room editor (exits, actions, spawners, NPCs, items, scripts) plus an Inspector tab
2. **Items** - Live item instances at `/creator/items`. That route stays available and is not in the sidebar.
3. **Item Templates** - Reusable item blueprints. The sidebar entry is labeled Items.
4. **NPCs** - NPC templates and unique NPCs. The detail switches between the form and an Inspector.
5. **Dialogs** - Dialog tree editor
6. **Quests** - Quest editor (objectives, rewards, prerequisites)
7. **Skills** - Skill/spell editor (multi-class, effects)
8. **Scripts** - Lua script editor with syntax highlighting
9. **Character Templates** - Archetype editor with modal item-template selection for starting gear
10. **World Map** - Grid-based world visualization. Fit view uses the same axis as the tiles: smaller Y is north and sits toward the top, matching authored room coordinates, so a zone filter stays on the canvas. The Reachability layer colors rooms from `GET /api/world/reachability`: green when reachable from the start room, indigo for an instance template that is reachable, and red for an unreachable island. Hidden exits are dashed and one-way exits are dotted while that layer is on. Clicking an island frames those rooms and shows the same reason content health uses.
11. **World Health** - Content-health report (reachability, reveal scripts, impossible quests, bosses, pack rules, drift) plus cross-system diagnostics for broken entity references and suspicious content values, including character template starting item references
12. **Players** - Admin only. Live characters (online, guest, zone, in combat; optional older rows), a side drawer, and confirmed ops: teleport (room picker), give item (template picker), take item (that character's inventory), end combat, quest complete/reset/abandon, and re-grant starter kit. Instance copies and cleanup sit on the same page. Creators do not see this page.
13. **Audit log** - Filterable creator and ops history, before/after JSON, and Undo for an admin

### Live Ops And Audit
- `ADMIN_ENV_LABEL` (optional `ADMIN_ENV_HOST`) feeds `GET /api/server-info`. A non-empty label shows a red `LABEL · host` badge in the top bar. Local servers leave the label empty.
- Live ops and live reads run on the game command loop through `Game.Call`. The HTTP handler does not edit an online character behind that loop.
- Every op body must include `confirm: true`. The Players page and `OpsButtons` open the shared confirm dialog before the request. A success toast offers Undo only when the audit row is undoable.
- `OpsButtons` modes are `character` (end combat, re-grant starter kit), `quest` (complete quest, abandon, reset quest, mark one step done, reset one step), `npc` (heal, respawn, despawn, end combat), and `room` (teleport here, clean up one instance copy). The NPC inspector embeds the npc mode. The room inspector embeds the room mode. The component renders only when the signed-in user is an admin. Room teleport does not pass `force`.
- Teleport moves online and offline characters through `RelocateCharacter`. A fight blocks the move unless `force` aborts it first. Undo teleports back when the character is still in the destination and not in a new fight.
- Give and take follow unique-item rules and tell an online player. Undo of a give removes the added pieces. Undo of a take puts those pieces back.
- `end-combat` aborts the fight: no rewards, penalties, or healing. It is not undoable. `instance-cleanup` relocates players in a chosen copy, or deletes only empty copies when `allEmpty` is set. It does not abort a fight in the copy and is not undoable.
- `quest-step` `op=complete` with `objectiveId` advances that objective and does not grant rewards. Without `objectiveId` it completes the quest and grants rewards. Undo restores the previous progress row and does not remove XP, gold, or items already granted.
- `regrant-starter-kit` adds missing class-kit or character-template starter pieces. It does not replace equipped gear. A class with no starter list is a skipped, non-undoable success.
- NPC heal restores full HP. Respawn creates or revives an instance only when it is not already alive. Despawn refuses an NPC who is in combat. A persisted unique is marked dead rather than deleted as content.
- Audit undo of a creator update or delete restores the stored before JSON, or recreates a deleted row. It returns 409 when the entity changed since the audited write.

### Room Editor Features
- **Exit management** - Add/edit/delete exits, toggle hidden
- **Action management** - Custom interactions (response, broadcast, script)
- **NPC residents** - Assign unique NPCs to room
- **Item placement** - Add items to room
- **Spawner configuration** - NPC spawner setup
- **OnEnter script** - Lua script triggered on room entry
- **Background & mood** - Visual settings
- **Coordinates** - Grid positioning (X, Y, Z)
- **Bind point** - Allow `/bind` for respawn
- **Inspector** - Inspector tab on the room detail. It calls `GET /api/rooms/:id/inspect` for exits in and out (a hidden exit lists the script that reveals it, or says none does), resident and spawn-room NPCs, spawners (template, max, respawn time), items, the on-enter script, actions, and quests with a visit objective for this room. Reachability is the world-index result from the start room. An instance-copy id resolves to its template. Live characters in this room or a copy of it come from `GET /api/live/characters?all=1` and are admin only. Live NPC instances come from `GET /api/live/npcs`. Instance copies come from `GET /api/live/instances`. Admin ops are teleport into this room or a chosen copy, and instance cleanup for that copy, each behind the shared confirm dialog and the audit toast. Creators see the live NPC and copy rows and a note that those ops are admin only. A `data-backlinks-slot` is left for the inbound-reference panel.

### NPC Editor Features
- **Template vs. Unique** - Toggle `IsTemplate` flag
- **Basic info** - Name, description, race, class, level
- **Enemy trait** - Grouped into combat stats, behaviour, loot and rewards, and Lua hooks. Fields: creature type, combat style, difficulty, attack, defense, attack speed, XP, aggro radius, aggro on sight, call for help, flee threshold (shown as a percent), gold range, loot table, guaranteed item templates, max drops, and on-aggro, on-death, on-flee, and on-low-health scripts with the low-health line as a percent and resulting HP. Help text states the engine default when a field is unset. Loot tables, item templates, and scripts use `EntitySelectButton`.
- **Content base vs effective stats** - Imported enemies store unscaled `enemyTrait.baseStats` (`maxHitPoints`, `attackPower`, `defense`) beside the effective stats written to `maxHitPoints`, `attackPower`, and `defense`. The enemy tab labels which values the inputs edit. Effective stats are read-only and named with the tier. A named override replaces that tier's factors. An unknown tier shows `unknown difficulty tier — no scaling applied` and keeps the base numbers. Editor-created enemies have no content base; their inputs edit the stored combat stats. Saving recomputes effective stats on the server when a content base is present, and a wounded NPC keeps its current HP. YAML export writes the content base when it is present so a round trip does not scale twice.
- **Merchant trait** - Inventory, pricing, restock, accepted items
- **Dialog assignment** - Main dialog, idle dialog
- **Behavior** - State, spawn room, wander radius, patrol path, idle chatter dialog, idle chatter timeout, respawn time
- **Resident placement** - Assign `CurrentRoomID` for auto-spawn
- **Inspector** - Edit / Inspector on the NPC detail. Inspector calls `GET /api/npcs/:id/inspect` for the template summary: type, level, difficulty, content base versus effective stats (same multipliers as `GET /api/balance/enemy-scaling`; a named override replaces the tier; an unknown tier keeps the base), loot-table chances, dialog and idle-dialog links, on-aggro / on-death / on-flee / on-low-health scripts, spawners (room, respawn time, max), and quests with a kill, talk, or deliver objective for this NPC. A running instance id resolves to its template. Live instances come from `GET /api/live/npcs?templateId=` (room, HP, opponent names, dead-until). Admin ops on that list are heal, respawn, despawn, and end combat, each behind the shared confirm dialog and the audit toast. Creators see the live rows and a note that ops are admin only. A `data-backlinks-slot` is left for the inbound-reference panel.

### Quest Editor Features
- **Objectives** - Add kill/collect/deliver/visit/talk/custom objectives
- **Area** - Optional regional label with dynamic suggestions and table filtering
- **Rewards** - XP, gold, item grants
- **Prerequisites** - Required quests, level requirement
- **Dialog integration** - Accept/progress/complete text
- **Repeatable** - Daily/recurring quests
- **OnComplete script** - Lua script on quest completion
- **Validation panel** - Flags missing source references, objective targets, duplicate IDs, invalid rewards, missing scripts, self-prerequisites, and unresolved entity IDs before save
- **Player flow preview** - Summarizes offer source, objective lines, ready turn-in target, and rewards for quick testing while authoring
- **Quest debugger** - `GET /api/quests/:id/debug` and `/creator/quests/debug?id=`. The id comes from the URL on load and whenever that query changes, including Debug quest and Health links. Shows the step chain, where each objective can be satisfied, whether that place is reachable, and a green, amber, or red verdict using the same completability check as content health. Also shows prerequisites, follow-on quests, the offer and turn-in, rewards, and characters currently on the quest. Each character row includes the open `objectiveId`. Reset quest step is admin-only. It confirms, then POSTs `/api/ops/quest-step` with `op=reset` for that objective. The toast shows the summary and can undo.

### Skill Editor Features
- **Multi-class assignment** - Skills can belong to multiple classes
- **Resource type** - Mana-based or cooldown-based
- **Effect system** - Damage, heal, buff, debuff, DoT, HoT, stun
- **Scaling** - Attribute-based damage/healing (STR/DEX/INT/WIS)
- **Secondary effects** - Hybrid skills (e.g., damage + debuff)

### Script Editor Features
- **Syntax highlighting** - Lua code highlighting
- **Test runner** - Execute scripts with test context
- **Shared validation** - Static Lua checks are shared with import validation and Creator save/preview flows
- **Type categorization** - item, room, npc, quest, event, custom

---

## Recent Features & Best Practices

### Per-Character Hidden Exit Reveals
**Feature**: Hidden exits can be revealed individually per character via scripting.

**Best Practice — Revealing an exit in an action script** (`type: custom`):
```lua
-- In a puzzle room action script (has ctx.roomID and ctx.characterID shortcuts):
local charID = ctx.character.ID
if tales.game.hasItem(charID, "bronze-key-template-id") then
    tales.game.msgToCharacter(charID, "The key fits! A secret passage opens.")
    -- ALL THREE PARAMS REQUIRED: roomID, exitName, characterID
    tales.game.revealExit(ctx.roomID, "secret-passage", charID)
else
    tales.game.msgToCharacter(charID, "You need a bronze key to unlock this.")
end
```

**Best Practice — State reconciliation in OnEnter script** (`type: room`):

When a puzzle sets a flag AND reveals an exit, the exit reveal could fail independently. Always reconcile state in the room's OnEnter script to ensure consistency:
```lua
-- In the room's onEnterScript (type: room — uses ctx.room.ID, NOT ctx.roomID):
local charID = ctx.character.ID
local puzzleSolved = tales.game.getFlag(charID, "puzzle_solved")
if puzzleSolved then
    -- Reconcile: ensure the exit is actually revealed for this character
    if not tales.game.hasRevealedExit(charID, ctx.room.ID, "north") then
        tales.game.revealExit(ctx.room.ID, "north", charID)
    end
    tales.game.msgToCharacter(charID, "The passage stands open.")
end
```

**Storage**: `Character.RevealedExits` map (roomID → exit names), persisted to DB
**Client**: Automatic silent room update shows new exit without re-rendering description
**Idempotent**: Calling `revealExit` on an already-revealed exit is a no-op (returns true)

**IMPORTANT — Parameter order difference**:
- `revealExit(roomID, exitName, characterID)` — room first
- `hasRevealedExit(characterID, roomID, exitName)` — character first

### CopyOnPickup Items
**Feature**: Items that create personal copies instead of removing from room. Instances are bound to the collecting character. Room-placed catalog items are treated as copy-on-pickup even if the YAML flag was omitted.

**Best Practice**:
- Use for **quest items** that all players should find
- Use for **story artifacts** that don't deplete
- Set `Item.CopyOnPickup = true` in Creator UI (import also sets this for any item listed on a room)
- System tracks via `Character.Flags["collected_item:<templateID>"]`
- Instances are automatically bound (`BoundToCharacterID`) — cannot be dropped, sold, or traded
- Players use `destroy` command to discard bound items (clears collected flag, allowing re-pickup)

**Example**: Ancient scroll in library - all players can read it, but each gets a bound copy

### Item Consumables with Lua Scripts
**Feature**: Items can have both data-driven effects AND custom Lua logic.

**Best Practice**:
```go
// In Creator UI - Item Attributes:
{
  "healthRestore": 50,
  "useMessage": "You feel refreshed"
}

// Set OnUseScriptID for additional effects:
OnUseScriptID: "potion-buff-script"
```

**Script Example**:
```lua
-- potion-buff-script (type: item — ctx.character is a Character object)
local charID = ctx.character.ID
tales.game.msgToCharacter(charID, "A warm glow surrounds you.")
-- Could add temporary buff, quest progress, etc.
```

### Distributable Attribute Points
**Feature**: Players earn 2 points per level to spend on attributes.

**Best Practice**:
- Design quests/challenges that reward specific builds
- Use class caps to prevent degenerate builds (warriors can't max INT)
- Show unspent points badge in character widget
- Remind players to spend points at level-up

### Combat Balance via YAML Config
**Feature**: Difficulty multipliers loaded from `config/combat_balance.yaml`.

**Best Practice**:
- Set base stats conservatively on NPCs
- Use difficulty tier to scale for different zones
- Trivial: Tutorial enemies
- Easy: Early-game
- Normal: Mid-game
- Hard: Late-game
- Boss: Unique encounters

**Tuning**: Adjust YAML file without editing individual NPCs

### Quest Dialog Auto-Injection
**Feature**: Quest offer/progress/complete options auto-injected into NPC dialogs.

**Best Practice**:
- Design NPC dialog trees WITHOUT quest options
- System automatically adds quest options at runtime
- Set `AcceptDialogText`, `ProgressDialogText`, `CompleteDialogText` on Quest
- Quest-giver NPC should have general greeting dialog
- Use the Quest editor validation/preview panel before saving to confirm the offer, progress, ready turn-in, and reward flow
- Quest-source NPCs without a main dialog still open a quest-only conversation so numbered quest choices can accept, show progress, or turn in quests.

### Exploration XP System
**Feature**: Automatic XP rewards for discovering rooms/areas.

**Best Practice**:
- Design areas with hidden rooms for bonus XP
- Use revealed exits to reward thorough exploration
- Track via `Character.DiscoveredRooms` and `DiscoveredAreas`
- **5 XP** per room, **15 XP** for first room in new area

### Room OnEnter Scripts
**Feature**: Lua scripts executed when player enters room. Script `type` must be `room`.

**Best Practice**:
```lua
-- OnEnter scripts use ctx.character.ID (uppercase .ID, not .id)
local character = ctx.character
if not character then return end
local charID = character.ID

-- First-visit message:
if not tales.game.getFlag(charID, "visited_dungeon") then
    tales.game.msgToCharacter(charID, "You enter a dark, foreboding dungeon.")
    tales.game.setFlag(charID, "visited_dungeon", true)
end

-- State reconciliation for hidden exits (if this room has puzzle-revealed exits):
local puzzleSolved = tales.game.getFlag(charID, "dungeon_puzzle_solved")
if puzzleSolved and not tales.game.hasRevealedExit(charID, ctx.room.ID, "north") then
    tales.game.revealExit(ctx.room.ID, "north", charID)
end
```

**Context**: `ctx.room` (Room object), `ctx.character` (Character object), `ctx.user` (User object)
**Note**: OnEnter scripts do NOT have `ctx.roomID` or `ctx.characterID` string shortcuts — use `ctx.room.ID` and `ctx.character.ID` instead.

### Multi-Class Skills
**Feature**: Skills can be assigned to multiple classes.

**Best Practice**:
- Healing skills for both Cleric and Druid
- Defensive buffs for Warrior, Ranger, Rogue
- Set `ClassIDs: ["cleric", "druid"]` in Creator UI
- Registry handles lookups automatically

### NPC Auto-Spawn via CurrentRoomID
**Feature**: Unique NPCs auto-spawn into their assigned room on server start.

**Best Practice**:
- Use for quest-givers, merchants, named NPCs
- Set `NPC.CurrentRoomID` in Creator UI
- Do NOT add to spawners
- Do NOT add to `Room.NPCs` list (system handles it)

### Item Template/Instance Pattern
**Feature**: Items follow same template/instance pattern as NPCs.

**Best Practice**:
- Create item **templates** with `IsTemplate = true`
- Use `CreateInstanceFromTemplate()` for loot, rewards
- Instances track `TemplateID` for quest matching
- Use `CopyOnPickup` for shared quest items

---

## Summary of Key Patterns

### Entity ID References
- **Rooms** - Reference other rooms via `Exit.Target`
- **Items** - Reference templates via `TemplateID`, scripts via `OnUseScriptID`
- **NPCs** - Reference templates via `TemplateID`, dialogs via `DialogID`, scripts via `OnAggroScript`/`OnDeathScript`/`OnFleeScript`/`OnLowHealthScript`
- **Quests** - Reference NPCs via `Source.NPCID`, scripts via `OnCompleteScriptID`
- **Skills** - Reference classes via `ClassIDs` array

### Per-Character State
- **Flags** - Arbitrary script data (`Character.Flags`)
- **Revealed Exits** - Hidden exit discoveries (`Character.RevealedExits`)
- **Discovered Rooms** - Exploration tracking (`Character.DiscoveredRooms`)
- **Collected Copy Items** - CopyOnPickup tracking (`Character.Flags["collected_copy_items"]`)
- **Quest Progress** - Active quests (`QuestProgress` entity)
- **Equipped Skills** - Active skill slots (`Character.EquippedSkills`)

### Template/Instance Pattern
**Used by**: Items, NPCs

**Fields**:
- `IsTemplate` - True if blueprint
- `TemplateID` - Source template ID (for instances)
- `InstanceSuffix` - Unique suffix (e.g., "abc123")

**Creation**:
```go
instance, err := service.CreateInstanceFromTemplate(templateID)
```

### Scripting Hooks
**Room**:
- `OnEnterScriptID` - Player enters room

**Item**:
- `OnUseScriptID` - Player uses item

**NPC**:
- `OnAggroScript` - Enemy aggros
- `OnDeathScript` - Enemy dies
- `OnFleeScript` - Enemy flees
- `OnLowHealthScript` - Enemy HP first drops below `LowHealthThreshold` while still alive (`tales.combat.summon` may add up to 3 enemy-template copies to that fight; they grant no rewards and leave when the fight ends)

**Quest**:
- `OnCompleteScriptID` - Quest completed

**Room Action**:
- `Action.ScriptId` - Player triggers action

---

## Discovered World Atlas

### Overview
The atlas is a per-character fog-of-war map. The server compiles a **stable world layout** from room exits (and optional `coords`), then reveals only rooms this character has entered plus unnamed fog neighbors through visible exits. Web and mobile clients render the same JSON.

Area-local authored coordinates and compass exits define geography. Compact zone translations and anonymous filler ground make one overworld continent with blended biomes, coastal sea, dirt paths, and towns. Interiors remain real selectable rooms grouped under exterior anchors; decorative ground never becomes a room.

### Server
- `Character.DiscoveredRooms` / `DiscoveredAreas` persist on enter (`worldmap.MarkOn` during `TakeExit` and character select)
- `GET /api/characters/:id/map` (owner or admin) returns `PlayerMap`
- Layout package: `pkg/worldmap` — `Compile(rooms)` then `Reveal(world, character)`
- Layers: semantic `overworld` for outdoors (including positive elevation) and anchored interiors, `lower` for subterranean context/negative depth, `upper` for other positive floors
- Hidden exits do not appear until the character has revealed them
- Optional room `coords` define area-local geometry. Embedded `map_layout.json` translates zones to compact centers; no room YAML migration is required.

### Payload
```json
{
  "characterId": "...",
  "currentRoomId": "R0102",
  "currentLayer": "overworld",
  "layers": [{"id": "overworld", "name": "Overworld", "kind": "overworld"}],
  "places": [{"id": "R0102", "name": "Wildflower Field", "x": 2, "y": -1, "layer": "overworld", "biome": "meadow", "terrain": "grassland", "mapRole": "surface", "kind": "wild", "discovered": true, "canTravel": true}],
  "paths": [{"from": "R0101", "to": "R0102", "dir": "north", "kind": "trail", "layer": "overworld"}],
  "regions": [{"id": "Z01_meadows_forest_path:overworld", "name": "Meadows Forest Path", "hull": [[1, -2], [3, -2], [3, 0]], "biome": "meadow"}],
  "landscape": [{"x": 2, "y": -1, "terrain": "grassland"}]
}
```
Fog neighbors are places with `discovered: false`, empty `name`, and `kind: "uncharted"`.

### Client
- Map widget (player-facing name; same `minimap` widget slot / atlas protocol) receives the atlas over WebSocket on enter, and can also fetch `GET /api/characters/:id/map`
- Action-bar **Map** chrome / Expand always opens a real fullscreen Map overlay (dimmed play surface, Esc/X close) via `MapOverviewOverlay` portaled to `document.body` — Inventory-style centered panel (`#map-overview-overlay` / `.map-panel`), not clipped to the Map widget and not Materialize `.modal`
- Area names: drawn over landscape regions (gold/cream + dark stroke, font scales with zoom); uses `region.name` / `place.areaName` only — never invents labels. Room-name LOD unchanged: mid = current + adjacent; zoomed in = more room names (collision-aware). Compass/vertical exit words are never painted (exit ticks only)
- Cartographer overlay fills ~80% of the viewport on desktop (side intel rail). On phone (≤768px) it is full-bleed / safe-area; intel is a bottom sheet (peek summary + Travel, expand for exits/residents). Tap selects; Travel is a primary button on select (no double-tap required; Inspect remains optional). Pinch-zoom and pan keep scale. Compact Map tab selection shows Travel immediately (Inspect optional); the inner Map title / Open Map chrome is removed so the MAP tab is the only header. During Travel the camera smoothly follows the player; when Travel completes in the Cartographer overlay, the overlay fades out and closes (compact Map tab stays open). Active quests in the Quest Log group by area with collapse/expand (current area expanded). Equipment paper-doll uses larger slots and a denser column layout. BattleStage shows WoW-style buff/debuff icon rows with remaining rounds from combatant `statusEffects`.
- Exterior elevations and upstairs interiors stay on Overworld; dungeons, crypts, cellars, and sewers use Lower. Oldtown stays north of Meadows on the configured continent. Interior selection retains the actual room ID.
- Title stays **Map**. Layer tabs (Overworld/Lower/Upper) only when `atlas.layers` has more than one entry. Compact optional widget opens fullscreen; Map chrome pin is primary
- Compact minimap and fullscreen MapOverviewOverlay share `atlasRenderer.paintAtlas`, surface grouping, and the cached blended continent. Lower has a dedicated torch-lit rock/floor/corridor scene; Upper retains terrain tiles. Fog hides uncharted art; one gold marker follows the current room or its exterior anchor.
- The widget auto-fits discovered places into its panel and keeps that fit (canvas is out of flow so it cannot resize the widget)
- Layer tabs, pan, wheel zoom, click-to-travel along discovered paths
- Desktop/mobile action bars (Option C): room-only dirs + room actions + Shop when a merchant is present; fixed INV / MAP / SAY chrome; **Recipes** seeded by default for crafting discoverability; optional Look/Rest/… pins via ⋯; layout revision migrates legacy Look/pin clutter and seeds Recipes onto rev-2 bars
- Inventory chrome opens a popup overlay by default; preference can switch to on-screen widget / mobile sheet
- Room action/system reaction toasts render large and centered on the room hero art (not the command log)
- Same JSON is the contract for a future mobile renderer

### Key Files
- `pkg/worldmap/` — layout, biomes, hulls, discovery, reveal
- `pkg/server/handler/charactermap.go` — REST endpoint
- `public/mud-client/src/game/widgets/MinimapWidget.svelte` — local Map renderer + fullscreen overlay host
- `public/mud-client/src/game/widgets/atlasRenderer.js` — layer framing, hit targets, label LOD, single you-marker
- `public/mud-client/src/game/widgets/continentRenderer.js` / `coastline.js` / `mapArt.js` / `surfaceAtlas.js` — cached zoom scenes, organic shores, room art, town/interior grouping, roads/bridges
- `public/mud-client/src/game/widgets/undergroundRenderer.js` — rock-sided corridors, themed floors, torches, stairs
- `public/mud-client/public/map-tiles/terrain-sheet.png` — shared 48px terrain and transparent building sprites
- `public/mud-client/src/game/hudPrefs.js` — Option C action-bar chrome/pins and hotbar helpers. Kit skills come from `GET /api/classes` (`?v=classkit1`)

---

## Game Client Tab Container Widget

### Overview
The Tab Container widget allows players to group multiple widgets into a single resizable container with switchable tabs. This reduces layout clutter, especially on smaller screens (e.g., Character + Inventory + Equipment in one tab group).

### Features
- **Multiple tab containers**: Unlimited tab containers can exist on screen simultaneously
- **Tab switching**: Click tabs to switch between contained widgets; active tab is highlighted amber
- **State preservation**: All tab content is kept mounted (via CSS visibility) — scroll positions, terminal buffers, and widget state are preserved when switching tabs
- **Tab labels**: Automatically derived from widget registry (icon + name)
- **Add/remove tabs**: Available in edit mode only — "+" button opens a dropdown of available widget types, "x" button removes individual tabs
- **No nesting**: Tab containers cannot be placed inside other tab containers (enforced at registry, store, and UI levels)
- **maxInstances enforcement**: Widgets inside tab containers count toward their `maxInstances` limit (e.g., you can't add Character both as a standalone widget and inside a tab)
- **Persistence**: Tab configuration (which widgets, active tab) persists to localStorage alongside layout data

### Data Model
Tab containers store extra fields on the widget layout item:
```javascript
{
  id: 'tabcontainer-1707000000000',
  widgetType: 'tabcontainer',
  tabs: [
    { widgetType: 'character', id: 'tab-character-001' },
    { widgetType: 'inventory', id: 'tab-inventory-002' }
  ],
  activeTabIndex: 0
}
```

### Key Files
- `public/mud-client/src/game/widgets/TabContainerWidget.svelte` — The tab container component
- `public/mud-client/src/game/layout/WidgetComponents.js` — Shared component map used by both WidgetGrid and TabContainerWidget
- `public/mud-client/src/game/layout/WidgetRegistry.js` — Registry entry (`tabcontainer` type, `layout` category)
- `public/mud-client/src/game/layout/LayoutStore.js` — Tab management methods (`addTabToContainer`, `removeTabFromContainer`, `setActiveTab`)

### LayoutStore Tab Methods
```javascript
layoutStore.addTabToContainer(containerId, widgetType)  // Add a widget as a new tab
layoutStore.removeTabFromContainer(containerId, tabIndex) // Remove a tab by index
layoutStore.setActiveTab(containerId, tabIndex)           // Switch active tab
```

---

## Guest Mode System

### Overview
Guest mode allows anonymous visitors to play TalesMUD for 30 minutes without Auth0 registration. Designed for embedding a live demo on the website.

### Guest Service (`pkg/service/guest.go`)
```go
type GuestService interface {
    CreateGuestSession(remoteIP string) (token string, err error)
    ValidateGuestToken(tokenStr string) (userID string, err error)
    CleanupExpiredGuests()
    StartCleanupLoop()
}
```

### Guest Session Lifecycle
1. Client calls `POST /api/guest` (public endpoint)
2. Server checks `ServerSettings.GuestsAllowed` and `MaxGuestAccounts`
3. IP rate limit checked (10 per hour per IP)
4. Random character created from system template presets with full starter items. Signed-in create uses the same equip path: each `StartingItems` name becomes a new item copy (`IsTemplate` false) in the listed slot. Presets take those names from the class catalog `starting_items`. An id that is not on the roster falls back to the stored template's names. Existing characters are not updated.
5. Character spawned in `ServerSettings.StartRoomID` (default `R0001` if that room exists); auto quests for that zone are granted
5a. Entering a room grants auto-source quests for that room's area (Z01 meadows: QST010*) so they fire after leaving Z00
5b. Lua `tales.game.giveItem` notifies collect-quest progress (foraging, script rewards)
5c. Quest YAML `onAcceptScriptId` runs after a dialog accept (Z01 Wren reveals the creek burrow)
6. Character `MaxLevelCap` set to 5
7. User created with `IsGuest=true`, `GuestExpiresAt=now+30min`
8. HMAC-SHA256 token signed with `GUEST_SECRET` env var
9. Token returned to client, stored in `sessionStorage`
10. On WebSocket connect: timeout goroutines start (5-min warning + expiry)
11. On disconnect: 5-min grace period before deleting guest data
12. Background cleanup loop removes expired guests every 5 minutes

### Server Configuration
```go
// In ServerSettings:
GuestsAllowed    bool  // Enable/disable guest mode (default: true)
MaxGuestAccounts int   // Max concurrent guests (default: 20, 0 = unlimited)
StartRoomID      string // New/guest spawn room (default "R0001")
```

### User Entity Guest Fields
```go
// In User:
IsGuest        bool       // Marks temporary guest account
GuestExpiresAt time.Time  // When guest session expires
```

### Per-Character Level Cap
```go
// In Character:
MaxLevelCap int32  // 0 = use global MaxLevel, otherwise per-character cap

// Helper method:
func (c *Character) GetEffectiveMaxLevel(globalMax int32) int32
```

The leveling system (`CheckLevelUp`, `ApplyLevelUp`) respects `MaxLevelCap` automatically.

### Frontend Guest Flow
- `WelcomeScreen.svelte` — logged-out choice: "Continue with X", "Continue with Google", "Email and password", and "Play as guest"
- `App.svelte` — `handleGuestPlay()` stores token in sessionStorage, skips onboarding. Logout clears that token so the next load is the welcome choice
- Account menu — guests see Continue with X, Continue with Google, Email and password, and End Session; a signed-in player sees "Switch character" and "Log out"
- `api/guest.js` — `createGuestSession()` API client

### Authentication
- Guest HMAC tokens are validated before Auth0 JWTs in `AuthMiddleware`
- Token claims: `sub` (RefID), `uid` (user entity ID), `exp` (30min), `guest: true`
- If `GUEST_SECRET` is not set, a random key is generated at startup
- Optional local username/password sessions (Argon2id) when a game-mode file sets `auth: local`. Classic servers leave this off. API responses omit the password hash. Login attempts are limited per client address. `X-Forwarded-For` is trusted only from loopback unless `trusted_proxies` or `TRUSTED_PROXIES` says otherwise.
- `presentation: door_tui` serves `public/door` and `GET /api/door/config` (title, subtitle, token key). Classic mode does not mount `/door`. The page paints live rooms, exits, actions, NPCs, resources, and combat status, plus the last few command replies. Keys and typed lines become engine commands. A world pack may add `keymap.yaml` (per room, per area, and a combat overlay) and `character_paths.yaml`. `d` stays down. With no character selected, the page asks for a name, then a numbered path, then sex. An existing name is selected. A new character uses the pack path when that file is present, and otherwise a numbered system template. Reconnect runs the new-day pass without another select.

---

## Document Maintenance

**Update this file when**:
- Adding new entity fields
- Adding new Lua API functions
- Adding new game systems
- Adding new Creator UI features
- Changing data structures

**Keep in sync with**:
- `PROJECT.md` - High-level features
- `ARCHITECTURE.md` - System architecture
- `docs/design/SCRIPTING.md` - Scripting documentation
- Creator UI column definitions (`tableColumns.js`)

---

**End of Document**
