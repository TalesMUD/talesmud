# TalesMUD Project Documentation

## Overview

TalesMUD is a browser-based Multi-User Dungeon (MUD) framework built with Go and Svelte. It provides a complete platform for creating and playing text-based multiplayer adventure games, featuring real-time WebSocket communication, a web-based content editor, and persistent game state via SQLite.

**Repository:** [github.com/TalesMUD/talesmud](https://github.com/TalesMUD/talesmud)

## Documentation Index

- **Architecture:** `ARCHITECTURE.md`
- **Core Systems & Features:** `FEATURES.md` (comprehensive reference for all systems, data structures, and APIs)
- **Game design + MVP backlog:** `docs/design/GAME_DESIGN.md`
- **Door on the shared engine:** `docs/DOOR-ON-MUD.md`
- **Ruleset profile:** `config/ruleset.yaml` (level cap, level-up mode, death, new day, resource keys, combat pacing, regen for out of combat, resting, and in combat). Combat math stays in `config/combat_balance.yaml`.
- **Scripting system:** `docs/design/SCRIPTING.md`
- **World map implementation:** `docs/design/WORLD_MAP_IMPLEMENTATION.md`
- **Quest authoring guide:** `docs/design/QUEST_AUTHORING.md`
- **Player guide:** `docs/player-guide/`
- **Development docs:** `docs/development/`

## MVP Roadmap (next up)

Planned epics (see `game-design/GAME_DESIGN.md`):

- Enemy NPCs + combat
- Combat instances (ad-hoc rooms)
- Items/loot/containers
- Inventory + equip/unequip
- Merchants/trading

## Features

### Core Game Features

- **Room-Based World System**
  - Interconnected rooms with customizable exits (directional, named, teleport)
  - Hidden/secret exits (toggleable visibility in editor)
  - Room actions for custom player interactions (respond, broadcast, run script)
  - Room action names match case-insensitively and beat global `examine`/`take`/`use` when they collide
  - Movement exits match case-insensitively; room presence fan-out does not block the game loop on SQLite
  - Response actions send the narrative `response` text, not the help `description`
  - Action descriptions shown in room text ("You can:" section)
  - Visual backgrounds and mood settings
  - Coordinate-based world mapping (X, Y, Z grid) plus a compiled atlas: authored area-local coords and compass exits preserve room geometry; compact configurable zone centers form one continent with decorative biome ground; per-character fog of war at `GET /api/characters/:id/map`
  - Dynamic item and NPC spawning
  - Unique NPCs auto-spawn into their assigned room on server start via `CurrentRoomID`
  - Per-character friends list (`friend add/remove/list`); HUD overlay with online flags + whisper; guests refused
  - Guest-public NPC, enemy, and player portraits (`/api/portraits/:filename`). Player art covers Human, Dwarf, and Elf across Warrior, Rogue, Mage/Wizard, Ranger/Hunter, Cleric, and Druid; missing art falls back to a class silhouette.
  - Merchant shop overlay in the room widget (structured `shop` WS message with item stats/description; click inspects, explicit Buy/Sell confirm; WoW-style compare-to-equipped deltas on buy inspect; dialog Trade inject)
  - Player chrome Map: Cartographer overlay (desktop ~80% + intel rail; phone full-bleed + bottom intel sheet); tap select with Travel primary (Inspect optional); compact Map tab has no duplicate Map/Open Map header; 14 terrain types with six native 48px pixel-art variants plus room-driven roof/keep/service/landmark stamps and four underground floor styles on one sprite sheet; dithered biomes, smooth organic beaches/foam/depth bands and offshore rocks/coastal cliffs, mixed foothills and central peaks, biome-specific oaks/pines/marsh trees, crisp town glyphs, quiet dirt roads batched beneath relief, canopies, buildings and opaque town paving, water bridges along charted outdoor exits, street-shaped town paving with angular footprint walls, corner towers and road gates, roofs and town/farm props with filterable interior selection, parchment fog, textured sea, glowing you-marker, Fit world / recenter and zoom/pan, with up to 10× fit scale and 220px tiles (twice the previous closest view), using crisp nearest-neighbor sampling. World fit and minimum zoom share a continent frame at roughly 70–80% of the map height on desktop. The minimap shares cached overview/close landscapes and the underground renderer; 48px close detail bakes on demand in an OffscreenCanvas worker where supported, with a synchronous fallback. The distant overview retains a compact 32px cache; finer leaves, bark, timber lookout towers, roof tiles, masonry, bridge planks, crop rows and props remain crisp at maximum zoom. Lower uses 48px themed floors, bones/barrels and dim cluster light, with no long void-crossing connectors. The full map has low-rate water/smoke accents that respect reduced motion; trees and buildings have no drop shadows. Outdoor elevations stay on Overworld, interiors share exterior anchors, underground rooms stay on Lower. Terrain rules/defaults live in `pkg/worldmap/map_terrain.json`; compact centers/ground/town flags in `pkg/worldmap/map_layout.json`; read-only preview instructions are in `tools/WORLDMAP-PREVIEW.md`
  - Play client WS: single-flight socket gate; close 4001 (session replaced) does not auto-reconnect; `/play` JS/CSS served no-cache. An optional SSH listener (`ssh.enabled`, default off) uses the same session replace across web and SSH. See `docs/ssh-access.md`.
- SSH device confirm: classic `/activate` redirects to `/play/?activate=<code>`. The play client looks the code up and confirms only after a click. Account menus list, add, and revoke SSH keys. Guests do not see those controls.
  - Action bar Option C: room dirs + room actions + Shop; fixed INV/MAP/SAY chrome; **Recipes** pin seeded by default; optional Look/Rest/… via ⋯
  - Gathering & crafting v1 (no professions): room GATHER chips + recipes/craft; R0209 CRAFT/RECIPES chips; R0102 first-gather hint
  - Spell Bar / Hotbar: docked on the desktop action bar (a moved hotbar widget stays separate); nine slots; skills + consumables; Rest seeded on empty/default bar (slot 7); Look/Talk/Flee bindable; no Search=look. Keys 1–9 fire those slots, Tab cycles combat targets, Escape closes the top panel, and `?` opens the shortcut list. None of those fire while a command or other text field is focused.
  - Desktop and wide presets keep the room on the left. Character and Equipment share one tab (Character open), Terminal, Quest Log, and Map share another (Terminal open), and Inventory sits under those tabs. The play shell is the window height, and a saved layout taller than the window is scaled so the action bar stays on screen. Invalid saved layouts fall back to the viewport preset. Saving a layout while a widget is focused stores the arrangement from before that expansion.
  - Play terminal wraps room lines on word boundaries and reflows that scrollback when the panel is resized
  - Play header: one row for Edit Layout, Party, Friends, and the account chip, with a gold menu aligned to the chip
  - Material Icons are served locally and preloaded; icon ligatures remain hidden until the font loads, so a font failure leaves empty icons.
  - OOC Resting chip on character HP / mobile header while `Flags.resting`; clears on combat or rest end
  - BattleStage: arena art clipped to the fight band; Attack/Defend/Items/Flee + hotbar share one dock strip; FF-style plates; queue chip centered in the dock. Layout B is the default fight view (party left, enemies right, corner detail frames, slim over-sprite HP, existing race/class and enemy portraits). Other players show on a compact left ally strip. Classic cards remain a Settings opt-out (`?battleLayout=classic`). Every fighter in a group appears: the player's large card and up to four compact ally cards with live HP/MP, turn, down/fled, hit FX, and join banner. Damage numbers float over the struck sprite in white with a dark outline, holding full opacity before fading. A missing portrait or item icon swaps once to a class, enemy, or generic silhouette instead of a broken image.
  - Combat start/join focuses the existing BattleStage cover by default. The local Settings panel can turn auto-focus off; a manual Open BattleStage button remains available during combat. Dismissing victory/defeat, outcome timeout, or leaving combat restores the previous panel focus. Layout and template saves keep the normal arrangement; command input and text fields keep their typing focus.
  - Settings persist locally: theme, parchment room descriptions, reduced motion (system/on/off), combat auto-focus, Battle layout B (default on; off selects Classic), and inventory opening as an overlay or layout widget. Audio controls are held until game sound exists; inactive Compact Mode and Room Text Overlay controls are hidden.
  - Combat self and allies share one row on desktop and a stacked section on phones. Joins show one banner; combat prose stays in the log. Defeated enemy sprites remain visible in grey with a Defeated label.
  - In multi-enemy combat, clicking a BattleStage enemy or pressing Tab changes the gold focus highlight and server aim without spending a turn. Basic attacks and hostile skills use that living target; switching to an orange, red, or skull foe repeats the level-gap warning once per change. A dead target falls back to a living foe.
  - Say chrome opens a message popup, then sends `say <text>`
  - Inventory chrome opens overlay by default (preference: overlay | on-screen widget)
  - Equipment paper-doll: square slots around portrait (head/neck/chest/hands | legs/boots/ring1/ring2; main_hand + off_hand under); compact ATK/DEF strip
  - Clicking equipped gear opens a detail card; Unequip is an explicit action. Inventory item cards show stats, value, available weight, usability requirements, and differences from worn gear. Clear class-relevant upgrades get a green badge. Equip, Use, Sell, and Drop remain explicit buttons. A unique item card has a gold frame and a UNIQUE mark.
  - A loot entry can be `rarity: unique` with a `chance` and `boss_only`. The roll is skipped when every victory recipient already holds that template (bag, equipped, nested container; no separate item bank). That drop announces `UNIQUE:` as a gold room chip. A template `unique` flag caps ownership and does not send the chip.
  - Room action/system reaction toast: centered on room hero art; LOOK-sized padding (no half-cut last line)
  - Quest Accepted / Complete: Veilspan moment cards (centered); open Talk dialog refreshes `[Quest]` → `[In Progress]`
  - Quest log Turn In for anywhere-ready quests; otherwise Turn in: NPC hint
  - WoW-style private cellar instances: `type: instance` exits, or a shared-room exit into a room tagged `instance`, clone a small room graph per character; town hub stays shared; empty copies are destroyed. Startup deletes leftover `~` room rows and moves characters to the hub, the copy's return exit, the start room, or the bind room, then tells them once on the next login

- **Character System**
  - Full RPG character creation with races and classes
  - Six-attribute system (STR, DEX, CON, INT, WIS, CHA)
  - Equipment system with 10 equipment slots
  - Equipped armor takes durability damage on death (not deleted); `repair` at a merchant restores it
  - Inventory management
  - Experience and leveling with flattened early-game XP curve (piecewise formula: gentle L2-5, transitional L6-15, steeper L16+)
  - Exploration XP: awards 5 XP per new room discovered, 15 XP for first room in a new area/zone
  - **Distributable Attribute Points**: 2 points per level-up for players to allocate freely into STR, DEX, INT, WIS, or STA
  - Class-based attribute caps prevent degenerate builds (e.g., warriors cap INT at 5, wizards cap STR at 5)
  - Terminal command `spend <attr> [amount]` to allocate points; `spend` with no args shows status table with current values, spent/cap per attribute
  - Character widget shows unspent points badge and interactive "+" buttons on each attribute when points are available
  - **Derived Combat Stats Display**: Character widget shows computed ATK (weapon damage + STR modifier), DEF (total armor from equipment), and MP/RND (mana regen per combat round, caster classes only). These update live when equipment or attributes change.
  - Existing characters receive retroactive points on login ((level - 1) * 2)
  - Server-side room/area discovery tracking per character, used by the discovered-world atlas (web + mobile JSON)
  - All-time statistics tracking (including rooms discovered)
  - Mana system for caster classes (Mage, Cleric, Druid) with level and INT scaling
  - Mana regeneration: `regen.out_of_combat` defaults to 2% HP and 5% mana every 10s, `regen.resting` to 10% HP and 15% mana every 10s, and `regen.in_combat` to 0.5% HP and 1% mana every 10s. Intervals follow the server clock. Per-round combat mana stays `CalculateManaRegen` (1+WISMod) and is separate from `regen.in_combat`.
  - Mana potions (Small/Medium/Large) as consumable items

- **Skills & Spells System**
  - Database-stored skills, editable via Creator UI (Skills tab)
  - Multi-class support: skills can be assigned to multiple classes (e.g., Heal for Cleric and Druid)
  - 29 default abilities across 6 classes, seeded on first run
  - Two resource types: mana-based (casters) and cooldown-based (physical classes)
  - Equippable skill slots (1-4 per class, level-gated progression)
  - Skill management: equip/unequip outside combat, locked during combat
  - Skill effects: damage, heal, buff, debuff, DoT, HoT, stun, multi-hit, AoE
  - Status effect system: buffs, debuffs, DoTs, HoTs with duration tracking
  - Attribute-scaled damage: STR (warrior), DEX (rogue/ranger), INT (mage/druid), WIS (cleric)
  - Mana shield absorption mechanic
  - Skill cooldown tracking per combat instance
  - Combat commands clear stale character combat flags when no live combat instance exists
  - In-combat commands: `cast <skill> [target]`, numeric shortcuts `1`-`4`
  - Management commands: `skills`, `skills equip <name>`, `skills unequip <name>`
  - YAML import/export for skills data

- **Item System**
  - Multiple item types: Currency, Consumable, Armor, Weapon, Collectible, Quest, Crafting Material
  - Quality tiers: Normal, Magic, Rare, Legendary, Mythic
  - Item templates for reusable definitions
  - Stackable consumables and partial drops persist reduced quantities consistently between character inventory and item instances
  - Container support with nested items

- **Quest System**
  - Data-driven quest definitions with multiple objective types: Kill, Collect, Deliver, Visit, Talk, Custom (Lua)
  - Quest progress tracking per character with persistent state
  - NPC dialog integration: automatic quest offer/turn-in options injected into NPC conversations, including quest-only NPCs without full dialog trees
  - Real-time quest log WebSocket updates include quest definition details, objectives, and rewards after dialog quest actions
  - Quest rewards: XP, Gold, and item grants on completion
  - Quest prerequisites: required quest completions and level requirements
  - Repeatable quests support
  - Quest categories and area labels for filtering and organizing regional quest lines
  - QuestTracker: automatic progress updates from game events (NPC kills, item pickups with stack quantities, room entries, dialog nodes, NPC delivery checks)
  - Accepting a collect quest pre-fills progress from matching items already in the character inventory, including stack quantities
  - Delivery objectives require and consume matching inventory items before progress is granted
  - Player quest log shows enriched objective descriptions, ready-to-turn-in state, and quest notifications for accept/progress/ready/complete events
  - Lua scripting API (`tales.quests`) for custom quest logic
  - Creator UI: full quest editor with objectives, rewards, prerequisites, dialog text configuration, validation feedback, and player flow preview
  - Player commands: `quests`/`ql` (quest log), `quest <name>` (details), `abandon <name>` (abandon quest)

- **Guest Mode (Play as Guest)**
  - Anonymous 30-minute demo sessions without Auth0 registration
  - "Play as guest" button on the logged-out welcome screen, under Continue with X, Continue with Google, and Email and password
  - Random character with random class from system templates
  - Spawns in `ServerSettings.StartRoomID` (default `R0001` when that room exists)
  - Auto-grants `source.type: auto` quests for the start room's zone (Z00 catacombs: QST0001–QST0004)
  - Entering a new area grants that zone's auto quests (Z01 meadows: QST010*)
  - Full starter items equipped automatically
  - Per-character level cap of 5 for guest characters
  - Full chat access during session
  - 5-minute warning before session expiry
  - Auto-deletion of guest user + character after session ends or disconnect (5-min grace period for reconnection)
  - Server-configurable: `GuestsAllowed` (default: true), `MaxGuestAccounts` (default: 20)
  - IP-based rate limiting (10 guest sessions per IP per hour)
  - HMAC-SHA256 guest tokens (separate from Auth0 JWTs), signed with `GUEST_SECRET` env var
  - Background cleanup goroutine removes expired guest accounts every 5 minutes

- **New Player Onboarding**
  - Phase-based flow: Welcome Screen, Nickname Setup, Character Creation Wizard, Game
  - Logged-out players see **Continue with X**, **Continue with Google**, **Email and password**, and **Play as guest**. X and Google pass `connection` (`twitter`, `google-oauth2`) so Auth0 skips the universal login password form. Email is the only button that opens that form
  - Logout clears the Auth0 session and this tab's guest token, then returns to that choice. A guest reload still resumes the guest session
  - The account menu has **Switch character**, which opens a picker of every character from `/api/my-characters` and sends `sc <name>`. Guests get the same X, Google, and email choices. Signed-in players get **Log out**
  - A signed-in player with more than one character sees that picker once per login (the server still enters on `lastCharacter`)
  - First-time users prompted to choose a display name/nickname
  - Three-step character creation wizard: Choose Template, Name Character, Confirm & Create
  - Signed-in create equips that template's starting items: a fresh copy of each named starter, not the template itself. Guests use the same equip path. Characters already in the world are left unchanged.
  - Automatic phase detection from user profile and character data
  - Guest users skip onboarding (character auto-created server-side)

- **Room Text Overlay**
  - Game text (combat, actions, player messages) displayed as translucent overlay on room image
  - Auto-dismiss with duration scaled by text length (2-4 seconds)
  - Smooth fade-in/fade-out animations
  - Stacks up to 5 messages during rapid sequences (e.g. combat)
  - Always enabled on mobile; optional toggle for desktop in Settings > Interface

- **Multiplayer**
  - Real-time player interactions via WebSocket
  - Players see each other in rooms
  - Global and room-based chat
  - Private tells/whispers and minimal party chat
  - Party flow: create parties, invite online players, accept/decline invites, list members, leave party, and send party chat
  - Party Combat Assist v1: same-room players join in-progress fights via `attack <npc>`; party members get a join nudge; XP/gold split among living joiners
  - Party Loot & XP Share v1: victory gold and XP also split equally with online party members in the killer's room (leftover to the engager; out-of-room and offline members excluded; items stay on the ground)
  - Party Follow v1: `party follow` / `party unfollow` trail the party leader through normal exits (`RelocateCharacter`); combat, teleports, portals, and private instances do not pull followers
  - Party UI: HUD launch next to Friends, PartyOverlay (create/invite/list/say/leave), centered Accept/Decline invite popup with countdown/timeout; guests can party like signed-in players
  - Emote system
  - Live session-based player presence tracking for room UI, chat routing, `who`, and silent room presence refreshes
  - Reconnect-aware client state with visible connecting/reconnecting status and automatic reconnect attempts
  - In-game character switcher for changing active characters without leaving the play UI

### Content Creation

- **Web-Based Editor**
  - Full-width filterable data tables for browsing all entity types (Rooms, Items, Item Templates, NPCs, Dialogs, Quests, Skills, Scripts, Character Templates)
  - Per-column filtering (text search, enum dropdowns) with instant client-side filtering and sorting
  - Side-by-side master-detail layout: data table + edit form shown together, closeable to full-width table view
  - **Entity Selection Modal**: All entity ID selectors (rooms, NPCs, items, scripts, dialogs, quests, character template starting items) use a centered modal dialog with a full filterable DataTable instead of simple dropdowns. This scales to hundreds of entries with per-column search, sort, and filter support. Components: `EntitySelectButton` (inline trigger) + `EntitySelectModal` (table dialog). **UI Guideline: Never use `<select>` dropdowns for entity ID references. Always use `EntitySelectButton` with the appropriate column definitions from `tableColumns.js`.**
  - Room editor with exit, action, spawner, items, and NPC resident configuration. An Inspector tab shows exits in and out (hidden exits name the script that reveals them), NPCs and spawners, items, on-enter and action scripts, visit quests, and reachability from the start room. Live characters, NPC instances, and instance copies come from the existing live endpoints. Admin teleport and instance cleanup use the shared confirm dialog. Inbound references stay on the Referenced by panel under the form.
  - Item and item template management with attributes and properties
  - NPC editor with behavior controls for state, spawn room, wander radius, patrol paths, idle chatter, enemy traits, and merchant traits. The enemy tab edits combat stats, behaviour, loot, and Lua hooks. Imported enemies keep unscaled content base stats beside the effective stats combat uses. `GET /api/balance/enemy-scaling` supplies the tier and named-override factors. An unknown tier warns and applies no scaling. The detail can switch from the form to an Inspector: type, level, difficulty, content base versus effective stats, loot chances, dialogs, hooks, spawners, and kill/talk/deliver quests. The loot table id opens `/creator/loot-tables`. Live instances come from `GET /api/live/npcs`. Admin heal, respawn, despawn, and end combat use the shared confirm dialog. Inbound references stay on the Referenced by panel under the form. NPC type is a table badge; the row dot is reserved for validation issues. Tables show full entity IDs with a copy button.
  - Loot tables (`/creator/loot-tables?id=`) list and edit drop entries: item template, chance, quantity, boss-only, and guaranteed. `POST /api/loot-tables/:id/roll?n=1000&seed=` previews drop frequency. The preview is seeded, creates no items, and is not an audit write. Inbound references stay on the Referenced by panel.
  - Spawners (`/creator/spawners?id=`) list room, template, max, respawn time, and the live count from `GET /api/live/npcs`. Filters cover template, room, zone (the room area), and unique NPCs that have no spawner. The editor uses the existing spawner API. Room and NPC links open those inspectors (`?view=inspector`). Both editors also open the inspector for `?tab=inspector`.
  - Lua script editor with syntax highlighting and integrated test runner
  - Dialog tree editor with options and alternate texts
  - Quest editor with validation, player flow preview, and a quest debugger for the step chain, reachability, prerequisites, and live character progress
  - Character template editor with archetype selection and starting gear
  - Skills editor with multi-class assignment, resource types, effects, and secondary effects
  - World map visualization (GridWorldEditor). Fit view frames tiles with north up (smaller Y toward the top), so a zone filter such as Z02 stays on the canvas. A Reachability layer colors rooms and lists unreachable islands from `GET /api/world/reachability`. Each island reason is its own line, and room and script ids link to the room inspector (`?view=inspector`) and the script editor. Level, aggro, spawners, players, quests, missing art, and live instance copies are separate toggles on `GET /api/world/overlays`, saved in `localStorage`. Player names are admin only. Overlay badges stay about 12px on screen at any zoom. A quest badge reads Q×N and keeps the quest ids in the tooltip. Badges that do not fit the tile at that size collapse into +N. The tile itself is unchanged.
  - World Health diagnostics for broken cross-system references across rooms, NPCs, items, loot tables, quests, dialogs, scripts, spawners, and character template starting gear. The message column stays visible, and room, NPC, dialog, quest, item, and script IDs link to `?id=` on that editor.
  - Content health: one rule report for reachability, reveal scripts, quests that cannot complete, bosses without spawners, unknown difficulty tiers, unreferenced scripts, and pack rules from `data/rules`. An item from a completable quest reward counts as obtainable when that quest's prerequisites can come first. Reachability starts at the configured start room, then the engine default. The content commit is that folder's own git HEAD, or `CONTENT_COMMIT` / `.content-commit`, or `unknown`. Live health warns with `deploy-tree-dirty` when the server checkout has tracked changes outside `import/`. The Creator Health page and `tales -check` share the content rules. See `docs/content-health.md`.
  - Creator quality validation: inline warnings/errors, broken-reference detection, save blocking for invalid references, and a world health diagnostics tab. Dialog node IDs may repeat when the same node is linked again; a duplicate is an error only when two definitions of that node disagree. Opening `/creator/<tab>?id=<entityId>` selects that entity. Deletes that call the API, plus room, room-item, special-exit, and dialog-node deletes, ask in a confirm dialog first. Merchant stock `maxQuantity` of -1 is unlimited.
  - Live ops on the Players tab (admin only): teleport, give and take items, end a fight, quest step changes, re-grant a starter kit, and instance cleanup. Creators do not see the Players tab. NPC heal, respawn, and despawn, and room teleport and instance cleanup, share `OpsButtons`, which stays hidden unless the user is an admin. Every op confirms first, runs on the game command loop, and can offer Undo from the audit log when the change is reversible.
  - Audit log: creator CRUD writes and live ops, with before/after JSON and admin undo. A red LIVE badge appears in the top bar when `ADMIN_ENV_LABEL` is set.
  - Drift page (`/creator/drift`, Operate): entities changed since the last import, with each field's before and after value. Export YAML downloads one entity. Export all downloads every added or changed entity as one concatenated YAML file. Removed entities are listed and are not exported.
  - Creator navigation is a left sidebar (World, Actors, Narrative, Systems, Operate). World includes Spawners. Actors includes Loot tables. It collapses to icons with Ctrl+B or Cmd+B, folds each group on its own, and becomes an off-canvas drawer below 1024px. The top-bar search field opens the creator search palette. Below 1024px the text links and that search field are hidden so the LIVE badge stays on one row. Classes have no editor route, so they are not in the sidebar. `/creator/items` stays reachable and is not listed.
  - Creator search (`GET /api/search`): Ctrl+K or Cmd+K opens a palette. Results are grouped by type, and Enter opens the matching Creator editor. The shortcut is ignored while typing in a field or the script editor. Rooms, NPCs, items, dialogs, quests, and scripts show a collapsed "Referenced by (n)" panel under the form. Deleting one of those entities names the inbound references.
  - Preview/test tools for dialogs, quests, rooms, merchants, and Lua scripts
  - CRUD operations with live preview

- **Scripting System**
  - Lua-based scripting via gopher-lua (primary)
  - Room action scripts (type "script" triggers Lua execution with room.action context)
  - Room on-enter scripts
  - Item behavior scripts
  - NPC behavior scripts
  - Quest scripting support
  - Game API: messaging, inventory checks (`hasItem`, `hasEquipped`), character flags (`getFlag`/`setFlag`), item rewards (`giveItem`), per-character exit reveals (`revealExit`)

### AI / LLM Integration

- **Groq API Integration**
  - Reusable Groq LLM client (`pkg/service/groq/`) for AI-powered text generation
  - Character creation: AI-generated names and descriptions based on selected template (archetype, race, class, backstory)
  - Protected API endpoint: `POST /api/generate/character`
  - Graceful degradation when `GROQ_API_KEY` is not configured
  - Designed for extension to other AI features (NPC dialogue, room descriptions, quest generation)

### Technical Features

- **Authentication & Authorization**
  - Auth0 OAuth2 integration
  - JWT-based API protection
  - Guest mode with HMAC-SHA256 tokens (no Auth0 required)
  - Dual auth middleware: tries guest token first, falls back to Auth0 JWT
  - Frontend session state avoids logging or retaining auth token excerpts
  - Basic auth for legacy admin endpoints (export/import), with explicit credentials required and insecure release defaults rejected
  - Session management
  - Three-tier role system: MUD Admin, MUD Creator, Player
  - MUD Admin configured via `MUD_ADMIN_OAUTHID` env var (has full access)
  - MUD Creators can modify game content (Creator area)
  - Players can play the game and access only their own characters and quest progress
  - User ban system (bans by Reference ID and email)

- **User Management (Admin Only)**
  - View all registered users with ID, Name, Nickname, Email, Access Level
  - Promote players to Creator role or demote Creators to Player
  - Ban/unban players with double-confirmation modal
  - Banned users are blocked from all authenticated endpoints

- **Optional Landing Page**
  - Serve a static landing page at `/` from the OS filesystem via `LANDING_PATH`
  - When disabled (default), the main app SPA is served at `/` as usual
  - Auth0 callbacks (`?code=` / `?error=`) pass through to the main SPA automatically
  - Static assets (images, CSS) in the landing directory are served alongside `index.html`

- **Data Persistence**
  - SQLite for all game data
  - World export/import functionality
  - YAML/JSON data file support
  - Optional per-character refilling resources (`character_resources`). Keys come from `config/ruleset.yaml`. The shipped file lists none, so a default server never spends a balance. Scripts read them through `tales.resources`.
  - Ruleset death penalties, level-up mode (`auto` or `trainer`), and an optional dawn heal. Defaults match the previous 10% XP loss, 1 gold, bind-point respawn, cap 50, and immediate level-up. A successful flee or slip is not a defeat: the escaper keeps hit points, gold, and the room they reached. A mixed party penalizes only the players who died.
  - Procedural private instances (`tales.instances.generate`): a per-character room line from templates, level-filtered encounters, cleanup on leave or timeout. Existing cellar instances and Party Follow are unchanged. Generated exits do not pull followers.
  - Optional text client at `/door` when `presentation: door_tui`. Classic mode does not mount that path. The page title, subtitle, and token key come from the game-mode file, with generic TalesMUD defaults. It paints rooms, exits, actions, NPCs, resources, combat status, and recent command replies on the same screen. The header shows experience beside gold. The fight log stays for the whole fight and clears on any key after it, including Enter. A bare extra attack on the victory screen is not sent. The status line stays pinned above the log. The stats key paints hit points, gold, worn gear, and a gems count. A compass letter uses an open exit before a menu bind, and open east and west exits are listed. The index heading uses the configured title. A one-shot reply belongs to the key that produced it. It sends normal commands. A repeated attack for a swing that is already queued does not print a second line. A pack `keymap.yaml` can bind keys per room or area. A new character picks a numbered path, then sex. Reconnect applies the new-day pass. `-config` can point a second process at its own port and database. Classic play is unchanged when no config is set.

## Technology Stack

### Backend
| Component | Technology |
|-----------|------------|
| Language | Go 1.18 |
| HTTP Framework | Gin |
| Database | SQLite |
| WebSocket | Gorilla WebSocket |
| Authentication | Auth0 JWT |
| Scripting | Otto (JavaScript VM) |
| AI / LLM | Groq API (llama-3.3-70b) |
| Logging | Logrus |

### Frontend
| Component | Technology |
|-----------|------------|
| Framework | Svelte 3.59 |
| UI Library | Materialize CSS |
| Terminal | xterm.js |
| HTTP Client | Axios |
| Router | yrv |
| Build Tool | Rollup |

### Infrastructure
| Component | Technology |
|-----------|------------|
| Container | Docker |
| Orchestration | Docker Compose |
| CI/CD | GitHub Actions |
| Database | SQLite |

## Project Structure

```
talesmud/
├── cmd/                    # Application entry points
│   ├── tales/              # Main server
│   └── dialog_sandbox/     # Dialog testing tool
├── pkg/                    # Go packages
│   ├── entities/           # Data models (characters, rooms, items, NPCs, dialogs)
│   ├── mudserver/          # Game server (WebSocket, game loop, commands)
│   ├── server/             # HTTP API server
│   ├── service/            # Business logic layer
│   ├── repository/         # Data access layer
│   ├── db/                 # Database utilities (SQLite)
│   └── scripts/            # Script execution engine
├── public/                 # Frontend
│   └── app/
│       └── src/            # Svelte source
│           ├── game/       # Game client
│           ├── creator/    # Content editor
│           ├── characters/ # Character management
│           └── api/        # API clients
├── api/                    # API test files & sample data
├── data/                   # Sample game data
└── bin/                    # Compiled binaries
```

## Game Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `north`, `south`, `east`, `west` | `n`, `s`, `e`, `w` | Move between rooms |
| `look` | `l` | Examine current room |
| `inventory` | `i` | Display inventory |
| `selectcharacter` | `sc` | Select active character |
| `listcharacters` | `lc` | List your characters |
| `newcharacter` | `nc` | Create new character |
| `who` | - | List online players |
| `party create` | `p create` | Create a party with your current character |
| `party invite <player>` | `p invite <player>` | Invite an online player to a party |
| `party accept` / `party decline` | - | Respond to a pending party invite |
| `party list` | - | List party members |
| `party leave` | - | Leave the current party |
| `party say <message>` / `party <message>` | `p say <message>` / `p <message>` | Send party chat |
| `scream` | - | Broadcast to room |
| `shrug` | - | Emote action |
| `help` | `h` | Show help |
| `attack` | `a`, `hit` | Attack a target / switch combat target |
| `defend` | `d`, `guard` | Queue defensive stance for next combat turn |
| `flee` | `run`, `escape` | Queue a flee attempt. Success ends as escaped, with no death penalty |
| `status` | `cs`, `combat` | Show combat status |
| `cast` | `spell` | Use a skill in combat: cast \<skill\> [target] |
| `skills` | `spells`, `abilities` | Manage skills: skills [equip\|unequip] [name] |
| `quests` | `ql`, `questlog` | Show quest log |
| `quest` | - | Show quest details: quest [name] |
| `abandon` | - | Abandon a quest: abandon [name] |
| `spend` | - | Spend attribute points: spend \<attr\> [amount] |
| `pickup` | `get`, `take` | Pick up an item from the room (room-placed catalog items copy per character and stay for the next guest; dropped loot is still taken) |
| `drop` | - | Drop an item to the room (blocked for bound items) |
| `destroy` | `discard` | Destroy an item from inventory |
| `examine` | `inspect` | Examine an item in detail |
| `use` | `eat`, `drink`, `consume` | Use an item (`use flint on torch` for item-on-item); Inventory Use on… picks a target |
| `equip` | `wear` | Equip an item |
| `unequip` | `remove` | Unequip an item |
| `equipment` | `eq`, `gear` | Show equipped items |
| `list` | `shop`, `trade` | List merchant inventory |
| `buy` | - | Buy from merchant; stackable quantities fit in a single stack when possible |
| `sell` | - | Sell to merchant (blocked for bound items and unsupported item types) |
| `value` | `price` | Check item sell price |

## Current Development Status

### Branch: NPCs (Active Development)

The NPCs branch represents the latest development work, focusing on NPC systems and player-NPC interactions.

#### Completed Features

1. **NPC Entity System**
   - Core NPC data structure mirroring player characters
   - Trait-based composition (DialogTrait, MerchantTrait, EnemyTrait)
   - Room integration with NPC presence tracking
   - Health, level, and class systems
   - Runtime NPC behavior loop for idle, patrol, dead/respawn, and combat states
   - Deterministic patrol paths and bounded wandering from spawn rooms

2. **Dialog Engine**
   - Full dialog tree system with branching conversations
   - State management tracking visited dialogs
   - Template rendering with dynamic variables ({{PLAYER}}, {{NPC}}, {{TIME}})
   - Conditional option display based on conversation history
   - Alternate text variations for natural dialogue
   - Ordered responses (different text on repeated visits)
   - Dialog sandbox for testing conversations

3. **Dialog Features**
   - Interactive dialogs (triggered by player interaction)
   - Idle dialogs (ambient NPC chatter with timeout)
   - Show-once options
   - Dialog exit markers
   - YAML serialization for dialog definitions

4. **Auto-Attack Combat System**
   - Automatic combat rounds (players and NPCs auto-attack each turn)
   - Turn-order initiative system with auto-processing
   - Players can queue special actions between auto-attacks: target switch, defend, flee
   - Combat starts with `attack`/`kill` and proceeds automatically. Enemies with `aggroOnSight` start the same way after `combat.aggro_on_sight` grace (default 2.5s, gap 5, 15s reaggro cooldown; `enabled: false` turns it off). The sight line is ordinary text, then one fight, including swarm pack and the party assist nudge.
   - Level-gap modifiers (`config/combat_balance.yaml` `level_gap`): hit, crit, and damage dealt/taken scale with attacker level minus defender level, clamped (default ±6). Equal levels are unchanged. Applies to basic attacks and skills for players and NPCs.
   - Class balance (`class_balance`): per-class damage dealt and taken, plus an uphill `behind_dealt` multiplier capped at 1.15. The world-pack class catalog wins when it has a row. `config/combat_balance.yaml` is the fallback. Ward takes hits at 1.00. A soak class's Slam starts its cooldown only when the swing hits. `wizard` uses the mage row. Ranger and hunter share the rogue row.
   - Boss mechanics (`boss_mechanics`): bosses and elites telegraph blows; bosses enrage on a round count or HP percent. Bosses also progress through Opening, Escalation (66% HP), and Last Stand (33% HP), with a transition banner and persistent BattleStage phase label. YAML controls tiers, bands, labels, and optional damage/enrage overrides.
   - Threat colors from the same gap (grey through skull) on room enemy names and battle nameplates. Orange or worse asks once before `attack`; `attack!` or a second `attack` engages.
   - Victory XP and gold scale by that tier against the highest level in the reward split. Bosses pay a one-time first-kill bonus per character. The battle victory panel shows base, level modifier, and a first-kill bonus when one was paid, then reveals each drop in its rarity color and calls out a level-up. Defeat lists XP, gold, battered armor, and the room you wake in. The panel dismisses on click, Enter, or Escape and leaves the terminal usable. Hits float a number over the struck nameplate (a crit is larger, a miss reads "miss"); crits and Crushing Blow flash harder. `prefers-reduced-motion` turns those animations off.
   - No turn timeouts or AFK mechanics needed
   - Enemy `attackSpeed` is attacks per round. 0 or omitted keeps one swing on the old beat. `onAggroScript`, `onDeathScript`, `onFleeScript`, and `onLowHealthScript` run once per fight in the sandboxed Lua runner. Low health defaults to 30% of max HP and does not run when the blow kills. A script can heal an ally, apply an existing buff or debuff, or `tales.combat.summon` up to 3 enemy-template adds per fight. Adds match template stats, grant no loot or XP, and leave when the fight ends. A death or flee drops that combatant from the turn order and keeps the turn index in range, so the fight continues and ends promptly when the last enemy dies. Hook room lines stay ordinary system messages and add `style` `combatEvent`, `hook`, and `source`. A summon during a hook flush sends the live roster on `combatStatus` at once. The play client shows those lines as a gold combat-log chip (`?v=uniques1`). A `rarity: unique` drop uses the same chip with hook `unique`.

5. **NPC Behavior and Quest Interaction**
   - NPC update loop handles idle wandering, ordered patrol paths, respawn cleanup, and idle chatter cooldowns
   - `talk` and `speak` open NPC dialogs and inject quest offer/progress/turn-in options
   - Quest-giving NPCs without a main dialog open quest-only conversations for numbered quest choices
   - MUD client NPC cards show enemy, merchant, quest giver, dialog, idle chatter, and current state badges
   - Creator NPC editor exposes behavior controls and uses modal entity selectors for patrol and room references

### Recent Commits (NPCs Branch)

| Commit | Description |
|--------|-------------|
| `b17856d` | Fixed Svelte issues |
| `6b621a2` | Huge improvements on player and NPC interaction |
| `29674d5` | New work on Dialogs |
| `b54c92e` | Further work on dialogs |
| `b5ae2c3` | More progress on dialog logic |

## Configuration

### Environment Variables (.env)

```bash
# Server Configuration
GIN_MODE=debug
PORT=8010

# Optional comma-separated CORS origins in addition to the built-in production
# domains and local dev origins for the Creator and MUD clients.
CORS_ALLOWED_ORIGINS=

# SQLite database path
SQLITE_PATH=./talesmud.db

# Auth0
AUTH0_AUDIENCE=http://talesofapirate.com/dnd/api
AUTH0_DOMAIN=https://owndnd.eu.auth0.com/
AUTH0_WK_JWKS=https://owndnd.eu.auth0.com/.well-known/jwks.json
AUTH_ENABLED=false

# Admin (basic auth for legacy export/import)
# Leave either value blank to disable these endpoints.
# admin/admin is rejected in release mode.
ADMIN_USER=admin
ADMIN_PASSWORD=changeme

# MUD Admin OAuth ID (Auth0 sub claim, e.g. "twitter|16651340")
# The user with this OAuth ID gets full admin access
MUD_ADMIN_OAUTHID=

# Guest mode secret key for signing guest JWTs (HMAC-SHA256)
# If not set, a random key is generated at startup (guest tokens won't survive server restart)
GUEST_SECRET=

# SSH listener (also settable in the game-mode ssh: block). All default off.
# These enable the listener and set its public address. They do not name a user or a key.
# SSH_ENABLED=false
# SSH_LISTEN=127.0.0.1:2222
# SSH_HOST_KEY_PATH=/var/lib/talesmud/ssh/host_ed25519
# SSH_PUBLIC_HOST=veilspan.com
# SSH_PUBLIC_PORT=2222
# SSH_ACTIVATE_URL=https://veilspan.com/activate
# SSH_GUEST_ENABLED=false
# SSH_DEVICE_ENABLED=false

# Creator live badge. Empty locally. Set ADMIN_ENV_LABEL=LIVE on a production admin.
# ADMIN_ENV_HOST overrides the request host shown after the label.
ADMIN_ENV_LABEL=
ADMIN_ENV_HOST=

# Optional landing page (path to directory with index.html + static assets)
# LANDING_PATH=./public/landing

# AI Generation (Groq API) — used for character name/description generation
# Get a key at https://console.groq.com
GROQ_API_KEY=
```

## Building & Running

### Prerequisites
- Go 1.26+ (toolchain go1.26.9)
- Node.js (for frontend build)

### Build Commands

```bash
# Build everything
make build

# Build frontend only
make build-frontend

# Build backend only
make build-backend

# Backend-only builds and server runs prepare fallback embedded frontend assets
# if pkg/webui/dist or pkg/webuiplay/dist has not been generated yet. Run
# `make build` to embed freshly rebuilt frontend bundles.

# Run the server
make run-server

# Run a second process with its own port and database
# (example: presentation door_tui, auth local, port 8030)
./bin/tales -config config/gamemode.yaml

# Forwarded client IPs are trusted only from loopback unless this is set
# TRUSTED_PROXIES=127.0.0.1,::1

# Run the server with SQLite (single binary + embedded frontend)
DB_DRIVER=sqlite SQLITE_PATH=./talesmud.db ./bin/tales

# Run frontend dev server
make run-frontend

# Run dialog sandbox
make run-dialogs-sandbox
```

### Docker Deployment

```bash
# Start with Docker Compose
docker-compose up -d
```

### Data Import

Import world data into SQLite:

```bash
go run cmd/migrate/main.go -input export.json -sqlite talesmud.db
```

## API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `GET /api/templates/characters` - Character creation templates
- `GET /api/room-of-the-day` - Featured room
- `POST /api/guest` - Create guest session (returns HMAC token)
- `GET /api/server-info` - Public server info: `serverName`, `envLabel` (`ADMIN_ENV_LABEL`), and `host` (`ADMIN_ENV_HOST` or the request host)
- `GET /api/ssh/info` - SSH listener status. `{enabled:false}` when SSH is off. When it is on, the body includes host, port, full host-key fingerprints, guest flag, and activate URL
- `GET /activate` - Device-code page. Local auth serves a sign-in form. Classic auth redirects to `/play/?activate=<code>`

### Protected Endpoints (Require Auth - Player Level)
- `GET /api/characters`, `POST /api/newcharacter` - Character management; direct character object access is owner/admin only
- `POST /api/generate/character` - AI-powered character name/description generation
- `GET /api/rooms`, `GET /api/items`, `GET /api/skills` - Read game data
- `GET /api/user`, `PUT /api/user` - User profile. A guest may read it. PUT returns 403 and leaves name, email, nickname, picture, role, and the guest flag unchanged
- `GET/POST /api/ssh/keys`, `DELETE /api/ssh/keys/:id` - List, paste, and revoke the signed-in account's SSH public keys. Guests and banned accounts are refused. 404 while the key store is off
- `POST /api/ssh/device/lookup|confirm|deny` - Confirm an SSH device code. Lookup returns a csrf nonce. Confirm and deny send it back. Nothing is confirmed automatically

### Protected Endpoints (Player Level - Quests)
- `GET /api/quests` - List all quest definitions
- `GET /api/quests/:id` - Get quest by ID
- `GET /api/quest-progress/:characterId` - Get character quest log for own/admin character
- `POST /api/quest-progress/:characterId/accept/:questId` - Accept quest for own/admin character
- `POST /api/quest-progress/:characterId/abandon/:questId` - Abandon quest for own/admin character
- `POST /api/quest-progress/:characterId/complete/:questId` - Complete a ready quest for own/admin character

### Creator Endpoints (Require Creator or Admin Role)
- `POST/PUT/DELETE /api/rooms` - Room management
- `GET /api/rooms/:id/inspect` - Static room inspector (creator or admin). An instance-copy id resolves to its template. Live rows stay on the live endpoints. Reachability uses the world index.
- `POST/PUT/DELETE /api/items` - Item management
- `POST/PUT/DELETE /api/scripts` - Script management
- `POST/PUT/DELETE /api/npcs` - NPC management. When `enemyTrait.baseStats` is set, create and update recompute effective HP, attack, and defense from the content base and the combat balance table.
- `GET /api/npcs/:id/inspect` - Static NPC inspector (creator or admin). An instance id resolves to its template. Live rows stay on `GET /api/live/npcs`.
- `GET /api/balance/enemy-scaling` - Read-only difficulty tiers and named overrides (`hp`, `attack`, `defense`)
- `POST/PUT/DELETE /api/dialogs` - Dialog management
- `POST/PUT/DELETE /api/quests` - Quest management
- `GET /api/quests/:id/debug` - Quest debugger: step chain, where each objective can be satisfied, reachability, prerequisite graph, offer and turn-in, rewards, and character progress. Each character row includes the open `objectiveId`. `ops.questStep.implemented` is true. The Creator page follows `?id=` on load and on navigation. An admin resets that open step with `POST /api/ops/quest-step`.
- `POST/PUT/DELETE /api/skills` - Skill management
- `GET /api/world/validation` - World Health diagnostics
- `GET /api/diagnostics/world` - World health diagnostics across rooms, NPCs, dialogs, quests, loot, items, and scripts
- `GET /api/world/reachability?from=` - World-map reachability from the start room, or from `from` when that room exists. Same walk as content health. Unknown `from` is 404.
- `GET /api/world/overlays` - World-map layers for a creator or admin. A room is included when it has a level band, an aggressive enemy, a spawner, an online player, a quest target, missing art, or a live instance copy. Level band is the NPC level range in the room, or the zone range when the room has none. The band is the midpoint: 0–4, 5–9, 10–14, 15–19, 20–29, 30+. Aggressive count is each `aggroOnSight` resident plus each aggro spawner's max instances. Spawners include the respawn time (override, then delay, then the template). Quest ids are rooms named by a visit, kill, talk, or deliver objective. Missing art is an empty background image. Copies is the number of live `~` rooms for that template. `playerNames` is present only for an admin; a creator still gets `players`. `live` is false when the game is not running, and the static layers still return.
- `GET /api/search?q=&types=&limit=` - Creator search over rooms, NPCs, items, dialogs (including node text), quests, scripts (name and body), loot tables, spawners, skills, and character templates. Exact id ranks above an id prefix, then an id fragment, then a name, then other text. Each hit includes `type`, `id`, `name`, a snippet, and a Creator path. Loot tables open `/creator/loot-tables?id=`. Spawners open `/creator/spawners?id=`. `limit` defaults to 25 and caps at 100. A blank query returns no hits.
- `POST /api/loot-tables/:id/roll?n=&seed=&playerLevel=&boss=` - Creator or admin. Seeded drop-frequency preview for one loot table. `n` defaults to 1000 and caps at 10000. It does not create items, change the table, or write an audit row. `POST /api/loottables/:id/roll` remains the single roll that creates item instances.
- `GET /api/refs/:type/:id` - Inbound and outbound references for one entity, grouped by type, with the field and a short reason. Unknown type is 400. A missing id is 404. The index is cached until a successful creator POST, PUT, PATCH, or DELETE.
- `GET /api/health` - Content-health report (summary, rules, live anomalies)
- `PUT /api/health/mute` - Mute or unmute a content-health rule (`{"ruleId","muted"}`)
- `GET /api/health/drift` - Entities changed since the last import baseline. The Creator Drift page lists them with before/after field values.
- `GET /api/health/drift/export?type=&id=` - Importer-format YAML for one drifted entity. Export all on that page concatenates one response per added or changed entity.
- `POST /api/validate/:entityType` - Validate a draft Creator entity before save
- `POST /api/preview/dialog`, `/api/preview/quest`, `/api/preview/room`, `/api/preview/merchant` - Preview/test draft content with validation issues
- `PUT /api/settings` - Server settings
- `GET /api/audit?entityType=&entityId=&limit=` - Audit log of creator writes and live ops
- `GET /api/live/npcs?templateId=&roomId=` - Running NPC instances, including room name, who they are fighting, and `deadUntil` when a corpse has a respawn time
- `GET /api/live/instances` - Instance room copies

### Admin API Endpoints (Require Admin Role)
- `GET /api/admin/users` - List all users
- `PUT /api/admin/users/:id/role` - Change user role
- `POST /api/admin/users/:id/ban` - Ban user
- `POST /api/admin/users/:id/unban` - Unban user
- `GET /api/live/characters` - Online characters and anyone seen in the last 30 days (`online`, `guest`, `inCombat`, `zone`, `all=1`). A creator receives 403.
- `GET /api/live/characters/:id` - Inventory, equipment, quest log, revealed exits, and current combat. A creator receives 403.
- `POST /api/audit/:id/undo` - Restore a creator write, or run a live op's inverse. Refuses when the row is not undoable, already undone, or the entity changed.
- `POST /api/ops/:action` - Live op on the game command loop. Body must include `confirm: true`. Actions: `teleport`, `give-item`, `take-item`, `npc-heal`, `npc-respawn`, `npc-despawn`, `end-combat`, `instance-cleanup`, `quest-step`, `regrant-starter-kit`.

### Legacy Admin Endpoints (Basic Auth)
- `GET /admin/export` - Export world data; requires explicit `ADMIN_USER` and `ADMIN_PASSWORD`
- `POST /admin/import` - Import world data; validates JSON before replacing stored data
- `GET /admin/world` - World map rendering

### WebSocket
- `GET /ws` - Game connection (authenticated)

## File Statistics

| Category | Count |
|----------|-------|
| Go source files | 86 |
| Svelte components | ~319 |
| JavaScript files | 23 |
| Total backend code | ~484KB |
| Total frontend code | ~344KB |

## License

See LICENSE file for details.

## Contributing

This project is actively developed. The NPCs branch contains the latest work on NPC systems and dialog interactions.

### Development Workflow
1. Fork the repository
2. Create a feature branch from `NPCs` (current active branch)
3. Make changes following existing patterns
4. Test with dialog sandbox for NPC-related changes
5. Submit pull request

## Related Resources

- [MUD Wikipedia](https://en.wikipedia.org/wiki/MUD) - Background on Multi-User Dungeons
- [Go Documentation](https://golang.org/doc/) - Go language reference
- [Svelte Tutorial](https://svelte.dev/tutorial) - Svelte framework guide
- [SQLite Documentation](https://www.sqlite.org/docs.html) - Database documentation
