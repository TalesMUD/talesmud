# Veilspan TUI

Status: spec only. No implementation in this change.
Base: `engine-june` `99492b3` (character switch blocked mid-fight).
Date: 2026-10-11. Owner ask: Marcus. Revision 2: installable client only.

This document specifies Veilspan TUI, a standalone terminal client for the Veilspan world. It does not specify a replacement for the Door text client on `door.veilspan.com:2223`.

SSH :2222 (`pkg/sshgate` + textline line mode) stays exactly as-is and is out of scope for this client.

## 1. Executive summary

Veilspan TUI is one Veilspan-branded Go binary. UI is Bubble Tea v2, Lip Gloss v2, and Bubbles v2. It dials `wss://veilspan.com/ws` (`GET /ws` in `pkg/server/server.go`) and the existing HTTP APIs (guest, characters, classes, art). It is a protocol client: the model consumes the structs that `wsTransport.Send` already JSON-encodes (`pkg/mudserver/transport.go`). It does not scrape ANSI, and it does not import `pkg/mudserver/game`. The engine must not import Bubble Tea, Lip Gloss, or this client.

Screen-reader mode is a local flag, `--a11y`, or `:a11y on` saved in the config file. It is a second presentation of the same events: linear lines, no alternate screen, no cursor-addressed redraw. It is not an SSH environment variable.

Rust plus Ratatui is the stronger pixel-art toolkit. For a standalone client the gap is narrower than it was when this spec also scored an in-process SSH session. Go is still the recommendation, for the reasons in section 1.2. Those reasons are shared Go fixtures, the team that already ships this engine, and a Bubble Tea v2 renderer that is enough on a local terminal. They are not "`sshgate` is already Go."

### 1.1 Where it lives

Put the client in its own repo, `atla/veilspan-tui`, the same way the Flutter client lives in `atla/veilspan-client`. Do not add `cmd/veilspan-tui` to talesmud.

- Engine neutrality. Bubble Tea never becomes an import of `pkg/mudserver`, `pkg/server`, or `pkg/entities`. A separate module makes that a build fact.
- Release cadence. The terminal client ships when its UI or install story changes. It does not wait on, or force, an `engine-june` binary.
- Branding. The binary is Veilspan. The OSS engine stays a generic MUD. Door and the line-mode SSH client stay in talesmud, untouched.
- The play client and Flutter already talk to this world from outside the engine process. This client is the third one.

The wire contract stays in sync without importing talesmud (that import would pull the game):

- Source of truth stays `pkg/mudserver/game/messages/responses.go` on the engine.
- An engine test marshals one fixture per constructor in section 2.3 and fails if the committed JSON drifts.
- `atla/veilspan-tui` CI checks out those fixtures at a pinned engine commit and fails if `internal/wire` cannot decode them.
- `proto` (section 2.5 and section 7.3) is the compatibility number. Additive fields do not bump it.
- A later extract of a tiny module (for example `atla/veilspan-wire`) that both sides import is optional. It is not required for v1. v1 is fixtures plus `proto`.

Binary name: `veilspan`. Module path: `github.com/atla/veilspan-tui`.

### 1.2 Recommendation

| Choice | Decision |
| --- | --- |
| Language | Go |
| Repo | `atla/veilspan-tui`. Not a command inside talesmud. |
| UI | `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2` |
| Renderer | Bubble Tea v2 Cursed renderer (cell diff, mode 2026). Enough for a local terminal. |
| Transport | Dial `GET /ws`. Send `{message, type:"message"}`. HTTP for everything in section 6 and section 4. |
| Wire | Existing JSON. New fields only as additive messages. Fixtures generated from the engine structs. |
| Art | Kitty, then sixel, then iTerm2 inline, then half-block truecolor, then ASCII. Half-block is the MVP picture. |

### 1.3 Weighted comparison

Standalone client only. Weights are 1–5 by how much they matter. Scores are 1–5. Higher is better. Weighted total is the sum of weight times score.

The earlier table scored about Go 200 to Rust 158 on weights that summed to 40. Most of that gap was rows that only exist if the TUI is also an SSH session: zero-install on `:2222`, SSH bandwidth, one binary that is also the SSH runner, and a line-mode fallback beside the TUI. Those rows are gone. What is left:

| Criterion | W | Go + Bubble Tea v2 | Rust + Ratatui 0.30.2 |
| --- | --- | --- | --- |
| Shares the Go response structs. Fixtures are marshalled in Go. The client still does not import the engine module. | 5 | 5 | 2 (a second decoder, kept honest only by the JSON fixtures) |
| Team velocity. The engine, the play server, and the fixture test are Go. | 4 | 5 | 2 |
| Terminal graphics (sixel, kitty, iTerm2, half-blocks) | 4 | 3 (`charmbracelet/x` has sixel and kitty encoders; no Ratatui-image equivalent) | 5 (`ratatui-image` 11.1.0 stable, 12.0.0-rc.2 on 2026-10-05) |
| Local redraw: cell diff, resize, mouse, Kitty keyboard | 3 | 4 (Cursed renderer, mode 2026, Ultraviolet) | 4 (immediate mode; the app owns the diff; Ratatui's widgets are the more proven set) |
| Core stays framework-neutral | 4 | 5 (separate repo; talesmud never imports Bubble Tea) | 5 (separate repo) |
| Snapshot tests | 3 | 4 (`teatest`) | 5 (mature golden tests) |

Weights sum to 23. Go totals 101 (about 4.4). Rust totals 85 (about 3.7). The gap is 16 points. Rust still wins graphics and snapshot tests. Local redraw and engine neutrality are a tie once nobody is hosting the TUI inside the game process. Go stays the choice because the fixtures and the people building them are Go, and Bubble Tea v2's renderer is enough for this UI. Half-block first still pays down the graphics gap. A Rust client would be the right call if pictures were the product. They are a fallback ladder on top of a JSON protocol.

### 1.4 Process shape

```
+------------------+     JSON events       +----------------------+
| tea.Model        |<----------------------| websocket + HTTP     |
| panels, keys     |  {message, type}      |                      |
+------------------+---------------------->|                      |
                                            +----------+-----------+
                                                       |
                                            wss://veilspan.com/ws
                                            https://veilspan.com/api/...
                                                       |
                                            pkg/mudserver (unchanged)
                                            web /play and Flutter use this too
```

There is no in-process `Transport` adapter and no `AttachExternal` path. The client is a process on the player's machine. A loopback `httptest` websocket is the conformance test in CI. It is not a production mode.

### 1.5 Door stays a different product

Door (`:2223`, `gamemode.ANSI()`, `pkg/presentation/doorview`) is a different product. `doorview.View` holds `*game.Game` and paints from live engine objects. Veilspan TUI must not do that, and must not import `doorview` (that import would pull the game into the client). Reuse, by copying into the client repo:

- The frame sanitizer rules from `pkg/presentation/ansi`. Door frames drop OSC, DCS, C1, and bidi controls (`docs/ssh-access.md`).
- Charset selection from `pkg/presentation/termout` when a CP437 glyph is the fallback.
- The idea of mapping a key to an engine command string. The Door map itself stays Door's.

Do not share a screen layout, a color theme, or a binary with Doorkeeper.

### 1.6 Library facts (checked 2026-10-11)

| Library | State | Source |
| --- | --- | --- |
| Bubble Tea v2.0.0 | Stable 2026-02-24. Cursed renderer, `tea.View`, mode 2026 sync updates, Kitty keyboard protocol, Lip Gloss no longer touches I/O. | https://charm.land/blog/v2.md (post date 23 Feb 2026). Patch v2.0.10 on 2026-09-24 skips capability queries when input is disabled: https://github.com/charmbracelet/bubbletea/releases (via newreleases, retrieved 2026-10-11). Latest is v2.1.0 (2026-10-08): OSC 7501 program status, 0x0 tty size falls back to defaults, shutdown deadlock fix. Target v2.1.x. |
| Lip Gloss v2.0.0 | Stable 2026-02-24. Pure styling. Import `charm.land/lipgloss/v2`. | https://github.com/charmbracelet/lipgloss/releases/tag/v2.0.0 |
| Bubbles v2.0.0 | Stable 2026-02-24. Import `charm.land/bubbles/v2/...`. Light and dark styles are explicit. There is no `AdaptiveColor`. Detect the background yourself. If the query times out, assume dark. | https://github.com/charmbracelet/bubbles/releases (v2 notes, retrieved 2026-10-11) |
| Ultraviolet | Cell diff renderer under Bubble Tea v2. Kitty keyboard, SGR mouse, Windows console input. | https://github.com/charmbracelet/ultraviolet (retrieved 2026-10-11) |
| charmbracelet/x | Experimental. `ansi` package gained sixel (`DCS`) and kitty graphics helpers (PR #380, merged Feb 2025). | https://github.com/charmbracelet/x/pull/380 |
| Ratatui | 0.30.2 on 2026-06-19. Backend split; `ratatui-crossterm` 0.1.2. Cited for the comparison only. | https://crates.io/crates/ratatui |
| ratatui-image | 11.1.0 stable 2026-09-17. 12.0.0-rc.2 on 2026-10-05. Sixel, kitty, iTerm2, half-blocks. Cited for the comparison only. | https://lib.rs/crates/ratatui-image |

Not verified: building against Bubble Tea v2.1.0 (the notes were read; the module was not compiled), and whether `charmbracelet/x` sixel is the API to pin. Pin versions in the client `go.mod` when implementation starts.

### 1.7 Lessons from other terminals

| Work | Take |
| --- | --- |
| terminal.shop, Charm soft-serve | Capability detection and resize are the hard parts of a full-screen terminal program. Use those lessons in this local binary. Do not host this TUI on `:2222`. |
| Brogue CE, Caves of Qud (curses) | A fixed grid, a small verb set, and a message log beat a free-form terminal. 80 columns is the floor. |
| Cataclysm DDA | Dense panels work when every panel has a key and a one-line help row. |
| TinTin++, Blightmud | Scriptable triggers are how experts escape bad UIs. We do not need them in MVP. The command line stays first-class so typed verbs from the line client still work. |
| Mudlet | GUI maps and triggers set the expectation. Our atlas message is the structured map. Do not redraw a pixel map by scraping room text. |
| Evennia web | Confirms the split we already have: server sends structured state, client renders. |
| Full-screen SSH toys | They skip auth, resize, and reconnect. This client has to do all three against `/ws`. It does not extend the line-mode session. |

## 2. Architecture

### 2.1 Modules

Repo `atla/veilspan-tui`:

| Path | Role | May import |
| --- | --- | --- |
| `cmd/veilspan` | `main`. Flags, config dir, websocket dial, subcommands `update` and `logout`. | `internal/tui`, `internal/auth`, Bubble Tea |
| `internal/tui` | Model, views, keys, art fallback, themes. | `internal/wire`, Bubble Tea. Not talesmud. |
| `internal/wire` | JSON structs that match today's socket. Decode only. | stdlib |
| `internal/auth` | Guest POST, device-code start/poll client, token file. | stdlib |
| `internal/update` | GitHub releases check. No network unless the check is enabled. | stdlib |

In talesmud, add no TUI package. The v1 engine touch is the fixture test that writes JSON (section 8.5). The device-code HTTP pair in section 6 lands only when Marcus wants account login. `pkg/mudserver` must not import Bubble Tea. `pkg/bus` (`pkg/bus/bus.go`) is an in-process event bus. It is not the player socket. The TUI does not subscribe to it.

`internal/wire` duplicates the JSON tags from `responses.go`. Drift is caught by the fixture test in section 8.5.

### 2.2 Directions

| Direction | Today | TUI |
| --- | --- | --- |
| Client to server | One JSON object per command. `messages.IncomingMessage` is `{message, type, key}` (`pkg/mudserver/game/messages/incoming.go`). Web play sends `{message, type:"message"}` (`public/mud-client/src/game/Client.js`). Door sends `{type:"door_key", key}` (`public/door/door.js`). | Send `{message, type:"message"}`. No new client verb in MVP. Panels format the same command strings the keyboard already sends (`attack`, `1`, `party accept`). |
| Server to client | `Transport.Send` of a struct with `type`. Audience fields are `json:"-"` and are not on the wire. | Decode that JSON. Unknown `type` is ignored, except a non-empty `message` is appended to the log (Flutter already does this fallback in `veilspan-client/lib/src/game/game_view_state.dart`). |

Auth on the socket is `AuthMiddleware`: `access_token` query or `Authorization` header. Guest HMAC is tried before Auth0 (`pkg/server/auth.go`). Close code `4001` means "session moved". Clients must not reconnect (`pkg/mudserver/mudserver.go`, `closeSessionReplaced`).

### 2.3 Messages that exist today

Server to client. Names are the `type` string in `pkg/mudserver/game/messages/messagetypes.go`. Payloads are the structs in `responses.go`. "Wire" is what a client can rely on. Audience is never on the wire.

| type | When | Fields the TUI uses |
| --- | --- | --- |
| `message` | Say, tells, room lines, command errors, combat prose, enemy-hook cards | `username`, `message`. Optional `style:"combatEvent"`, `hook` (`onAggro`, `onLowHealth`, `onDeath`, `onFlee`), `source`. |
| `ping` | Every 60s, `sendUserPings` | Ignore. Do not print. |
| `enterRoom` | You enter a room. Also the look refresh. | `message` (rendered description), `room` (`rooms.Room`: `name`, `description`, `area`, `tags`, `exits[]` of `{name, description, type, hidden, instance, target}`, `actions[]`, `meta.mood`, `meta.background`, `canBind`), `npcs[]`, `items[]`, `players[]`. |
| `roomUpdate` | Silent refresh of the same shape | Same as `enterRoom` without replacing the log. Web: `Client.js` `roomUpdate`. |
| `roomPresence` | Players only | `roomId`, `players[]` of `{id, name, isYou}`. `isYou` is false on this broadcast. The client marks itself. |
| `createCharacter` | Account has no character | `message` text only. No template list. |
| `characterSelected` | After `sc <name>` | `character` (full `characters.Character`), `xpForNextLevel`. |
| `characterUpdate` | Vitals, combat flag, skills | `currentHitPoints`, `maxHitPoints`, `currentMana`, `maxMana`, `xp`, `xpForNextLevel`, `level`, `gold`, `inCombat`, `resting`, `attributes`, `equippedSkills`, `maxSkillSlots`, `unspentAttributePoints`, `spentAttributePoints`, `attackPower`, `attackAttr`, `weaponDamage`, `attackMod`, `defense`, `manaRegen`. |
| `levelUp` | Levels gained | `oldLevel`, `newLevel`, `levelsGained`, `hpGained`, `manaGained`, `attributeGains`, `attributePointsGained`, `unspentAttributePoints`, `maxHitPoints`, `maxMana`, plus `message` ASCII fallback. |
| `inventoryUpdate` | Full bag, not a delta | `inventory`, `equippedItems`, `gold`. Shapes are the entity structs. Effect lines are `items.ItemEffect` `{trigger, name, text}` when the item carries them. |
| `dialog` | NPC options | `npcName`, `npcText`, `options[]` `{index, text}`, `conversationID`, `message` (same text). |
| `dialogEnd` | Conversation over | `message`, `username` (npc name). |
| `combatStart` | Fight UI | `enemies[]`, `players[]` as `CombatantView`, `targetId`, `message`. |
| `combatTurn` | Decision window | `actorId`, `actorName`, `round`, `deadlineMs` (unix ms), plus `CombatQueueState`. |
| `combatAction` | One resolution | `actorId`, `actorName`, `targetId`, `targetIds`, `action`, `ability`, `result` (`hit`, `miss`, `crit`, `block`, `fled`, `defended`, `dodged`, `cast`), `damage`, `heal`, `remainingHp`, `maxHp`, `fxId`, `combatants[]`, queue state, `message`. |
| `combatStatus` | Cooldowns / roster without a blow | `round`, `combatants[]`, queue state. |
| `combatEnd` | Fight over | `outcome` (`victory`, `defeat`, `fled`, `timeout`), `rewards` (`RewardBreakdown`), `loot[]` (`name`, `quality`, `quantity`, `unique`), `levelUp` callout, `defeat` (`xpLost`, `goldLost`, `respawnRoom`, `respawnRoomId`, `armor`, `hp`, `maxHp`), `message`. |
| `questAccepted`, `questProgress`, `questReady`, `questCompleted`, `questAbandoned` | One quest | `questName`, `questId`, `status`, `objectives[]` `{objectiveId, description, current, required, completed}`. |
| `questLog` | Full log | `quests[]` as `QuestLogEntry` (category, area, level, `readyToTurnIn`, turn-in npc/room, rewards). |
| `atlas` | Fog-of-war map | `atlas` is `worldmap.PlayerMap`: `characterId`, `currentRoomId`, `currentLayer`, `layers`, `places`, `paths`, `regions`, `landscape`. Place has `x,y`, `danger`, `exits`, `discovered`, `canTravel`, `mapFeatures`, `artSeed`. |
| `shop` | Merchant overlay | `merchantName`, `merchantId`, `gold`, `stock[]` (price, qty `-1` = unlimited, `effects[]`, quality, slot), `acceptedTypes`, `rejectedTags`, `sellMultiplier`. |
| `recipes` | Craft list | `recipes[]` (`canCraft`, `stationOk`, ingredients with `have` / `haveEnough`, output), `roomTags`. |
| `friends` | Roster | `friends[]` `{id, name, online}`. |
| `party` | Roster | `inParty`, `partyId`, `partyName`, `leaderId`, `maxMembers`, `members[]` `{id, name, online, level, class, portrait, isLeader}`, `following`. |
| `party_invite` | Accept / Decline popup | `pending`, `inviterName`, `partyId`, `expiresAt` (unix seconds). TTL is 45s (`def.PartyInviteTTL` in `pkg/mudserver/game/def/gamectrl.go`). |

`CombatantView` (`responses.go`): `id`, `type`, `name`, `portrait`, `hp`, `maxHp`, `mana`, `maxMana`, `classId`, `isAlive`, `hasFled`, `level`, `threat`, `telegraph`, `enraged`, `bossPhase`, `bossPhaseLabel`, `bossPhaseCount`.

`CombatQueueState`: `queuedAction`, `queuedSkillId`, `queuedTargetId`, `skillCooldowns` (map, always present), `nextActionAtMs`, `decisionDeadlineMs`.

`RoomNPC` on `enterRoom` (`responses.go` and `pkg/mudserver/game/util/game_utils.go`): `id`, `name`, `displayName`, `isEnemy`, `isMerchant`, `isQuestGiver`, `hasDialog`, `hasIdleDialog`, `currentHp`, `maxHp`, `level`, `threat` (enemy only, relative to the viewer), `state`, `portrait`, `templateId`.

Room items on the wire are `util.RoomItem`: id, name, no-pickup, copy-on-pickup. Not the full item.

Player combat decision window defaults to 5 seconds (`CombatConfig.DecisionWindowSeconds` in `pkg/mudserver/game/combat/engine.go`). `deadlineMs` is absolute unix time. The TUI counts down from that, not from local "5". A queued action during the window resolves immediately (`docs/combat-battle-stage.md`). Empty queue becomes auto Attack. `TurnBeatMs` default 1000, `ReactionMs` 400. The combat loop tick is 1 second. Do not animate faster than the events.

`style:"combatEvent"` lines are ordinary `message` frames from enemy hooks (`pkg/mudserver/game/enemy_hooks.go`): `onAggro` at fight start, `onLowHealth` once when an enemy first crosses the threshold, `onDeath`, `onFlee`. Summons are world changes (new NPCs / combatants on the next `combatStatus` or `combatAction.combatants`), not a separate type. `docs/DOOR-ON-MUD.md` still says OnDeath / OnAggro / OnFlee are not executed. That sentence is stale on this commit. Trust `enemy_hooks.go`.

Web `Client.js` also switches on `combatLeave`. No Go constant emits it. Ignore it. Do not invent it.

Flutter (`game_view_state.dart`) handles the same set minus `party`, `party_invite`, `recipes`, `levelUp`, and the quest progress variants beyond accepted/completed. The TUI follows the Go list, not the Flutter subset.

### 2.4 Messages that are missing

These are the gaps that make a panel lie or fall back to prose. Priority matches section 10.

| Gap | Why it hurts | Proposal | Priority |
| --- | --- | --- | --- |
| No `hello` / protocol version / capability | TUI cannot say "I understand combatEnd.loot" vs an older process | Client sends `{type:"hello", proto:1, client:"<semver>"}` with an empty `message` once. `handleInput` ignores an empty line (`text != ""` before enqueue) but still runs `ConnectUserSession` and a user update, same as any other read (`mudserver.go`). Do not put `hello` in `message` or it runs as a command. A server reply with `ok` / `upgrade-recommended` / `too-old` is new (section 7.3). `door_key` is the only incoming `type` with special handling today. | v1 |
| `aggroOnSight` not on `RoomNPC` | `EnemyTrait.AggroOnSight` exists (`pkg/entities/npcs/enemy_trait.go`) and the scheduler runs (`game_aggro.go`). The room card does not say so. Play client in this tree has no Aggressive label. | Add `aggroOnSight` bool on `RoomNPC`. Until then the TUI has only the prose line when the fight starts. | v1 |
| Examine is prose | `examine` / `inspect` reply with `message` text (`commands/examine.go`). Web parses it (`parseExamineOverlay.js`). | `examineResult` with name, slot, quality, stats, `effects[]` using `items.FormatEffect` ("On hit — Name: text"). | v1 |
| Who, say, tell, emote are prose | `who`, `say`, `tell`/`whisper`/`pm`, `scream`, `shrug` are `message` lines. | Keep prose in the log. A `who` snapshot can wait. | later |
| Skills list is not a message | `skills` prints text. Equipped ids arrive on `characterUpdate.equippedSkills`. Definitions are not on the wire. Web seeds a local catalog (`hudPrefs.js`) and also reads `GET /api/classes` (`classkit.ClientSkill`). | TUI loads `GET /api/classes` for names, tooltips, cooldown, target, once-per-fight. Do not hardcode Fenwatch skill text. | MVP (HTTP, not a new socket type) |
| Hotbar layout is client-local | Engine registers shortcuts `1`–`4` only (`commandprocessor.go`). `skills.HotbarCap` is 4 (`pkg/entities/skills/classkit.go`). Web shows 9 slots in `localStorage`, mixing skills, items, and commands (`hudPrefs.js`). | TUI shows `maxSkillSlots` from `characterUpdate` (4 for kit classes). Keys 1–4 send `1`..`4` in combat. Extra web slots are not protocol. Item-on-hotbar waits. | MVP |
| Inventory is a full dump, not a delta | Fine at current bag sizes. | Keep the dump. Do not invent a delta. | — |
| No art cell size on the room | `room.meta.background` is an id. Bytes live at portrait and item-art URLs. | Section 4. No new socket type for MVP. Optional later: `art` message with a content hash and cell-size variants. | later |
| `quit` does not disconnect | There is no `quit` command on the server. | Quit closes the websocket and exits the process. Do not send the word `quit`. The server sees a dropped socket and runs `handleUserQuit` (`game_handlers.go`). Optional later: server `session` message `{op:"bye"}` so a remote kick and a local quit look the same. Close `4001` already covers "session moved". | MVP behavior, later message |
| No combat snapshot on reconnect | A rejoining client waits for the next `combatTurn` / `combatStatus`. | On attach, if `characterUpdate.inCombat`, server sends one `combatStatus` with combatants. | v1 |
| Room NPC portrait is a path | `portrait` is a URL path from `pkg/portraits` (`/api/portraits/<id>.png`). | HTTP fetch, then section 4. | MVP |
| Party HP is not in `party` | Overlay has level and class, not vitals. Combat roster has HP only while `inCombat`. | Use `combatStart.players` during a fight. Out of combat, do not invent HP. | — |
| Shop sell list is the bag | Selling uses `sell <item>`. Stock is structured. The player's sellable rows are `inventoryUpdate`. | Join them in the panel. | MVP |
| Craft errors are prose | `Client.js` pattern-matches "you craft" / "missing materials". | Keep that match for MVP. A `craftResult` can wait. | later |
| Character list is HTTP | `GET /api/my-characters`, `POST /api/characters` with `{templateId, name, description, race}` (`pkg/server/dto/createcharacter.go`). Socket `lc` / `nc` print text. `sc` selects. | TUI uses HTTP for the picker and `sc` for enter. | MVP |
| Settings, layout, themes | Web `SettingsStore`, layout key `talesmud_layout_v1`. | Local file. Not protocol. | MVP |
| Admin / creator | HTTP under the creator app (`public/app`). Not on the game socket. | Out of the game TUI. Section 9. | later |

### 2.5 Versioning

- `proto` starts at 1 and means "the table in 2.3".
- Servers add fields as `omitempty`. Clients ignore unknown fields and unknown `type` values.
- A breaking change (rename, or a new required field) bumps `proto` and is not allowed on the public websocket until the play client and Flutter tolerate it.
- The TUI sends one empty-message `hello` (with `proto` and `client`) and continues if the server never answers. Absence of a reply means "this server, proto 1 as documented here". The empty read is not a command. It does touch presence, so do not send it on a timer. Reply statuses, when they exist, are in section 7.3.

### 2.6 Commands the line still has to type

Registered in `pkg/mudserver/game/commands/commandprocessor.go` and `roomprocessor.go`. The TUI binds keys to these strings. It does not add verbs.

Movement and room: `n/north`, `s/south`, `e/east`, `w/west`, any other exit name, `l/look`, `r/room`. Dynamic room actions by name. `craft` / `recipes` room actions open the recipe list.

Chat and social: `say`, `tell`/`whisper`/`pm`, `party`/`p` with `create|invite|accept|decline|leave|kick|promote|list|say|follow|unfollow` or a bare message, `scream`, `shrug`, `who`, `friend`/`friends`.

Character: `sc`/`selectcharacter`, `lc`/`listcharacters`, `nc`/`newcharacter`, `character`/`char`/`stats`, `spend`, `h`/`help`.

Items: `pickup`/`get`/`take`, `drop`, `destroy`/`discard`, `examine`/`inspect`, `use`/`eat`/`drink`/`consume`, `equip`/`wear`, `unequip`/`remove`, `equipment`/`eq`/`gear`, `list`/`shop`/`trade`, `buy`, `sell`, `value`/`price`, `repair`.

Combat: `attack`/`a`/`hit` and the `!` forms for a much stronger foe, `focus`, `defend`/`d`/`guard`, `flee`/`run`/`escape`, `status`/`cs`/`combat`, `cast`, `bolt`, `rig`, `skills`/`abilities` (and silent `spells`/`spell`), `1` `2` `3` `4` in combat only.

Quests and rest: `quests`/`ql`/`questlog`, `quest`, `abandon`, `complete`/`turnin`, `bind`, `rest`, `gather`/`forage`/`harvest`, `recipes`/`recipe`, `craft`.

Dialog numbers are eaten by `DialogSelectCommand` before room exits (`roomprocessor.go`).

Ops (`pkg/mudserver/game/ops.go`) are HTTP admin actions, not player commands: `teleport`, `give-item`, `take-item`, `npc-heal`, `npc-respawn`, `npc-despawn`, `end-combat`, `instance-cleanup`, `quest-step`, `regrant-starter-kit`.

### 2.7 Session rules already fixed in the engine

Character switch mid-fight is refused. `select_character.go` replies `You can't switch characters during a fight. Flee or finish the fight first.` when the active character `IsPlayerInCombat` and the requested name differs. The bug was fight traffic routed per account while commands ran per character. The TUI greys Switch while `characterUpdate.inCombat` is true, and it still sends `sc` so the server stays the authority.

The word `quit` is not a registered command, so sending it does nothing useful. Veilspan TUI closes the websocket and exits the process. The server treats the dropped socket as a disconnect (`handleUserQuit`).

## 3. Screens

Coordinates are terminal cells. The minimum is 80x24. Below that, the binary shows a one-line "window too small" screen and accepts only quit or a resize. It does not switch to another client. Resize reflows on the next frame. No panel stores a pixel position.

Breakpoints:

| Size | Layout |
| --- | --- |
| 80–109 cols, or rows 24–29 | One column. Art up to 16x6. Log at least 4 rows. Panels replace the room (inventory, quests, map) instead of sitting beside it. |
| 110–159 cols and rows >= 30 | Art beside description. Log under both. Hotbar is one row. |
| 160–199 cols | Art, description, and a 28-col minimap. Log gets the extra rows. A 160x48 screen is this split plus 24 more log rows. |
| >= 200 cols and >= 48 rows | The 160 layout plus a right inspector (examine, quest pin, or who). 200x60 is that inspector at full height. |
| rows < 24 or cols < 80 | Too-small screen. Quit or resize. |

Combat replaces exploration. Dialog, shop, party invite, and level-up are modal layers in that order: invite and level-up outrank dialog; dialog outranks shop; Escape closes the top layer. This matches the play client's "top panel wins" idea (`keyboardShortcuts.js` `PANEL_ORDER`) without copying its widget grid.

Focus: the command line owns printable keys. Pane keys (F1–F6, Ctrl shortcuts) work even while the line is focused, except when a modal wants digits. In dialog, digits select options and are not typed as commands.

Color: truecolor when the detected profile says so, else 256, else 16, else mono. `NO_COLOR` forces mono and disables art protocols (text and half-block outlines only). Threat colors use the engine names grey, green, yellow, orange, red, skull (`RoomNPC.threat`). Unique loot uses a single bright marker `UNIQUE`, same idea as the web gold diamond (`combatEnd.loot[].unique`).

Mockups below are ASCII, exact width, 24 rows. Names are illustrative. Blank rows are the log's spare cells at that size. On a 160x48 or 200x60 terminal those rows fill with the log. They are not a second layout.

### 3.1 Exploration, 80x24

```
+- explore 80x24 --------------------------------------------------------------+
| Cinder Lane                    Fenwatch L12   HP 42/60  MP 11/18             |
| Ashmarket, night. Mood: cinder. Resting. Gold 18.                            |
| +--------------+ The lane is a trench of cold cinders. A shutter             |
| | ####  art    | ticks. Mira waits under the lamp. A cellar rat              |
| | #..#  14x5   | watches the drain.                                          |
| | ####  half   |                                                             |
| +--------------+ Exits  N yard   E alley   D cellar                          |
| Here  Mira (talk, quest)   Rat #1 L10 enemy                                  |
| You   Bran (from the east)                                                   |
| Ground  cinder shard, torn bill     Actions  look, gather                    |
| Map          . N .                                                           |
| 3x3          W * E      [*]=you                                              |
|              . S .                                                           |
| -------------------- log ---------------------                               |
| Bran arrives from the east.                                                  |
| The rat's eyes catch the lamp.                                               |
| [combat event] Rat #1 bristles. (onAggro)                                    |
| > _                                                                          |
| 1 Brace  2 Slam  3 Stand  4 --     Tab targets   ? help                      |
|                                                                              |
|                                                                              |
|                                                                              |
+------------------------------------------------------------------------------+
```

Header reads `characterUpdate` and the room name from `enterRoom.room.name`. Art uses `room.meta.background` when set, else a terrain glyph from the atlas biome if `atlas` has arrived, else a hash of the room id drawn as blocks. Exits are `room.exits` with `hidden` omitted unless the character has revealed them. The server already strips hidden exits in `NewEnterRoomMessage` via `util.StripHiddenExits`. The client does not second-guess that.

NPC row: `displayName`, markers from `hasDialog`, `isMerchant`, `isQuestGiver`, `isEnemy`, `level`, `threat`. `AGGRESSIVE` appears only after `aggroOnSight` is on the payload (section 2.4). Until then, omit the word.

Log is a ring (section 8). `combatEvent` lines are prefixed and kept when the fight UI opens.

### 3.2 Exploration, 160 columns

Top 24 rows of a 160x48 screen. Rows 25–48 continue the log. A 200-column screen adds the inspector to the right of this block. The sample at the end of this section is the 200-column rule, not a full playfield.

```
+- explore 160 cols, top 24 of 48. Rows 25-48 are more log. ---------------------------------------------------------------------------------------------------+
| Cinder Lane                      Fenwatch L12   HP 42/60 MP 11/18  Gold 18  resting                                                                          |
| Ashmarket · night · mood=cinder                 layer overworld · room R0144                                                                                 |
| +--------------------+  The lane is a trench of cold cinders. A shutter ticks. Mira waits under the lamp.                                                    |
| | ####    art 20x6  |  A cellar rat watches the drain. Mood line: cinder (room.meta.mood).                                                                   |
| | ##  ##  half/sixel|  Exits  [N] yard   [E] alley   [D] cellar          Hidden exits stay off this list.                                                    |
| | #    #            |  Here   Mira (talk, quest)     Rat #1  L10  enemy  AGGRESSIVE once the flag is on the wire                                             |
| | ##  ##            |  Also   Bran (player, east)                                                                                                            |
| | ####              |  Ground cinder shard, torn bill          Actions: look, gather                                                                         |
| +--------------------+  Pins   quest: The Lamp Shard  0/1 bring the shard to Mira                                                                            |
|  . N .     danger: low  Map danger from atlas place.danger. Current room marked *.                                                                           |
| ------------------------------------------------------------ log ------------------------------------------------------------------------------------------  |
| Bran arrives from the east.                                                                                                                                  |
| The rat's eyes catch the lamp.                                                                                                                               |
| [combatEvent] Rat #1 bristles.  hook=onAggro  source=Rat #1                                                                                                  |
| > _                                                                                                                                                          |
| 1 Brace  2 Slam(cd2)  3 Stand  4 empty     F1 quests  F2 bag  F3 map  ? keys                                                                                 |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
+--------------------------------------------------------------------------------------------------------------------------------------------------------------+
```

### 3.3 Combat, 80x24

Opened by `combatStart`. Closed by `combatEnd`. Party on the left, foes on the right, last blow in the middle. This is the play client's Layout B idea (allies as a compact card, your turn countdown visible) squeezed into cells. See `docs/combat-battle-stage.md`.

```
+- combat 80x24 ---------------------------------------------------------------+
| ROUND 4            YOUR TURN  3.2s          target Rat #1                    |
| +-- party --------+---------------------------+-- foes --------+             |
| | You   HP ###### |                           | Rat #1 ######  |             |
| | 42/60 MP ###    |   (arena: last blow)      | L10  HP 12/28  |             |
| | Bran  HP ####   |   Slam hits Rat #1 for 6  | telegraph: bite|             |
| | ally  30/40     |                           | phase -        |             |
| +-----------------+---------------------------+----------------+             |
| Event  Rat #1 is bleeding.  (onLowHealth)                                    |
| Log    You brace. The next blow glances.                                     |
|        Rat #1 bites Bran for 4.                                              |
| Bar    A attack   D defend   F flee   I item   Tab target                    |
| Hotbar 1 Brace*  2 Slam cd2  3 Stand  4 --                                   |
|        * = keeps the swing     queued: Slam -> Rat #1                        |
| Timer  [########..]  3.2s then auto Attack on Rat #1                         |
| > _                                                                          |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
+------------------------------------------------------------------------------+
```

Timer uses `deadlineMs`. Cooldown badges use `skillCooldowns` from `combatTurn`, `combatAction`, or `combatStatus`. Telegraph, enrage, and boss phase come from `CombatantView`. A summon shows up as a new row on the next combatant list. Do not keep a row the server dropped.

`combatEnd.outcome`:

| Outcome | Panel |
| --- | --- |
| `victory` | Rewards (`base`, level mod, share, first-kill, party size) and loot. A row with `unique:true` is marked `UNIQUE`. |
| `defeat` | `xpLost`, `goldLost`, armor names, respawn room, HP you wake with. |
| `fled` | One line plus the room you land in (`enterRoom` follows). |
| `timeout` | Same chrome as defeat or fled, using `message` for the reason. |

The log hides while the victory panel is up, matching the play client. Dismiss with Enter. The next `enterRoom` or a key returns to exploration.

Targeting: Tab cycles living foes from `combatants`. `focus <name>` is the command. `attack` with no target uses the server's auto target. `attack!` is a confirmation bind (Shift-A) when the last error said the foe is much stronger (`commands/attack.go`).

Rigger: `bolt` and `rig` are commands, shown on the hotbar only when `GET /api/classes` marks those skills `oncePerFight`. They forfeit the swing (`commandprocessor.go` help text).

### 3.4 Combat, 160 columns

```
+- combat 160 cols, top 24 of 48. Extra rows are combat log. --------------------------------------------------------------------------------------------------+
| ROUND 4     YOUR TURN  3.2s     target Rat #1     queued Slam     phase waitingPlayer                                                                        |
| party                         arena                                      foes                                                                                |
| You   HP ######  42/60       Slam hits Rat #1 for 6.              Rat #1  ######  12/28                                                                      |
|       MP ###     11/18       fx: slash                            L10  threat orange                                                                         |
| Bran  HP ####    30/40                                            telegraph: bite                                                                            |
|       Fenwatch L14           Event card                           bossPhase 0                                                                                |
| (compact ally card)          onLowHealth                          enraged no                                                                                 |
|                              Rat #1 is bleeding.                                                                                                             |
|                                                                                                                                                              |
|                                                                                                                                                              |
| A attack   D defend   F flee   I use item   Tab / Shift-Tab cycle living foes   1-4 skills                                                                   |
| 1 Brace (keeps swing)   2 Slam cd 2   3 Stand   4 --     timer [########..] 3.2s -> auto attack                                                              |
| Victory/defeat replaces this frame: rewards, UNIQUE loot row, or defeat xp/gold/respawn room.                                                                |
| > _                                                                                                                                                          |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
|                                                                                                                                                              |
+--------------------------------------------------------------------------------------------------------------------------------------------------------------+
```

### 3.5 Dialog, 80x24

Driven only by `dialog` / `dialogEnd`. Options use the server `index` (1-based). Sending that digit is `DialogSelectCommand`. Do not send the option text.

```
+- dialog 80x24 ---------------------------------------------------------------+
| Talk  Mira the lamp-keeper                                                   |
| ----------------------------------------------------------------             |
| "The cellar is open if you still want the shard. I will not                  |
|  go down there again."                                                       |
| ----------------------------------------------------------------             |
|  1  I'll take the job.                                                       |
|  2  What is in the cellar?                                                   |
|  3  Show me what you sell.     (opens shop)                                  |
|  4  Goodbye.                                                                 |
| ----------------------------------------------------------------             |
| Keys  1-9 pick   Esc / 0 leave   Up/Down move   Enter confirm                |
| Room stays behind this panel. Exits are frozen until dialogEnd.              |
| > _     (a number is a dialog choice, not a skill)                           |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
+------------------------------------------------------------------------------+
```

If a `shop` message arrives, the shop panel replaces the dialog (web and Flutter both clear dialog on shop). Shop rows are `stock[]`. Keys: `b` buys the highlighted row (`buy <name>`), `s` sells the highlighted bag row (`sell <name>`), numbers jump. Quantity `-1` prints as `inf`. Effect lines under the row use `effects[]`.

Recipes are the same pattern (`recipes`): station not ok, missing counts, and `canCraft` are server fields. Enter sends `craft <key>`.

### 3.6 Inventory and equipment, 80x24

`inventoryUpdate` is the whole picture. Effect lines render with `items.FormatEffect` order: label, name, text. Triggers: `onHit` "On hit", `onEquip` "While equipped", `onUse` "On use". A script with no authored text is the fallback "Special effect" (`pkg/entities/items/effects.go`). The server attaches that fallback on shop stock. For the bag, show `effects` if present and do not guess a script name the payload omitted.

```
+- inventory 80x24 ------------------------------------------------------------+
| Inventory                     Gold 18          Equipped                      |
| Bag                           Wear             Slot                          |
|  1 cinder shard   x3  junk    [weapon] ash knife   +2 dmg                    |
|  2 torn bill          quest   [chest ] road coat    3 def                    |
|  3 lamp oil       x1  use     [legs  ] --                                    |
|  4 --                          [feet  ] --                                   |
|                                [ring  ] --                                   |
| ----------------------------------------------------------------             |
| ash knife          weapon, normal, L1                                        |
| On hit -- Nick: a small bleed.                                               |
| While equipped -- none                                                       |
| Compare vs empty hand: +2 attack, +0 defense                                 |
| Keys  j/k move  e equip  d drop  u use  x examine  q close                   |
| > equip 1                                                                    |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
+------------------------------------------------------------------------------+
```

Item art is a 4-cell thumbnail when the protocol is available (`/api/item-art/<template>.png`, else `generic-<bucket>.png` from `pkg/itemart`). Quality colors: normal, magic, rare, legendary, mythic.

### 3.7 Quests, 80x24

`questLog` fills the list. `questProgress` / `questReady` update one row and append a log line. Turn-in sends `complete <name>`.

```
+- quests 80x24 ---------------------------------------------------------------+
| Quest log                     filter: active                                 |
| * The Lamp Shard        active   turn in: Mira, Cinder Lane                  |
|   1/1  Enter the cellar                          done                        |
|   0/1  Bring the shard to Mira                   open                        |
|   Reward  40 xp, 8 gold, item tpl-lamp-shard                                 |
|   The Cellar Rat        ready    turn in anywhere                            |
|   3/3  Defeat cellar rats                        done                        |
|   Completed (2 hidden, press c)                                              |
| ----------------------------------------------------------------             |
| Keys  j/k  Enter detail  a abandon  t turn-in  c completed  q close          |
| Turn-in still sends: complete The Lamp Shard                                 |
| > _                                                                          |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
+------------------------------------------------------------------------------+
```

### 3.8 Party and invite, 80x24

`party` for the roster. `following` is this character trailing the leader (`PartyMessage`). Follow sends `party follow`. Instances do not pull followers (`docs/DOOR-ON-MUD.md`). The panel can say "follow stops at a private instance" only as static help, not because the server sent a flag.

Invite is a modal from `party_invite` while `pending` is true. `A` sends `party accept`. `X` sends `party decline`. Expiry is `expiresAt`. When `pending` is false, close the modal.

```
+- party 80x24 ----------------------------------------------------------------+
| Party  Cinder Watch            leader Bran     2/5     following             |
|  * Bran     L14 Fenwatch   online    leader                                  |
|    You      L12 Fenwatch   online    following the leader                    |
|  (empty)                                                                     |
| Keys  i invite  l leave  f follow  u unfollow  k kick  o promote             |
| Chat  party say <text>     or  P then the line                               |
| ----------------------------------------------------------------             |
| +-------------- invite ---------------+                                      |
| | Bran invites you.  31s              |                                      |
| |  [A]ccept     [X] decline     Esc   |                                      |
| +-------------------------------------+                                      |
| Invite rides party_invite. expiresAt is unix seconds.                        |
| > _                                                                          |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
+------------------------------------------------------------------------------+
```

Friends (`friends`) is the same list pattern. No invite popup. `f` adds by the command line `friend add <name>`.

Who stays a log page (`who`) until a structured message exists.

### 3.9 Character create and select, 80x24

Picker data is `GET /api/my-characters` plus `GET /api/classes`. Create is `POST /api/characters`. Enter world is `sc <name>`. The socket `createCharacter` message only means "you have none"; it does not list classes.

Class cards come from the live catalog. The repo sample is Fenwatch (warrior), Alley (rogue), Rune Hand (wizard), Ward, Rigger (`pkg/classkit/testdata/classes/*.yaml`, asserted by `catalog_test.go`). Races and blurbs are on `ClientClass`. Do not hardcode the five if `/api/classes` returns a different list. Rigger's race in that sample is `construct`.

Switching character while `inCombat` is disabled in the picker. The server still refuses `sc`.

```
+- create 80x24 ---------------------------------------------------------------+
| Who walks the lane?            GET /api/classes   POST /api/characters       |
| 1 Fenwatch    plate, STR    human, dwarf     Brace / Slam / Stand            |
|   You stand in the door until they do not.                                   |
| 2 Alley       leather, DEX  human, dwarf, elf   Slip / Nick / Smoke          |
| 3 Rune Hand   cloth, INT    human, elf       Inscribe / Sear / Glyph         |
| 4 Ward        plate, STA    human, dwarf     Guard / ...                     |
| 5 Rigger      construct     bolt, rig, overload. Once per fight.             |
| Name  ____________     Race  [human]                                         |
| Origin blurb comes from the template. Hotbar cap 4.                          |
| Enter creates. Esc cancels. Guests can create; login saves later.            |
| > _                                                                          |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
+------------------------------------------------------------------------------+
```

Level-up uses `levelUp`: old, new, HP, mana, attribute gains, unspent points. `spend <attr> <n>` from that card. Dismiss does not spend.

### 3.10 200-column rule

```
+- 200-col rule ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| 200 cols: room art | description+exits | log | inspector (examine / quest pin / who)                                                                                                                 |
| Breakpoint >=200 and rows>=48 opens the inspector. Below that the inspector is a panel (x).                                                                                                          |
| This row is exactly 200 columns including the frame.                                                                                                                                                 |
|                                                                                                                                                                                                      |
|                                                                                                                                                                                                      |
|                                                                                                                                                                                                      |
+------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
```

### 3.11 Map

Minimap is a 9x5 slice of `atlas.places` around `currentRoomId` on `currentLayer`. `danger` tints the cell. `canTravel` false is a fog cell. Landscape cells are decoration and are not exits (`pkg/worldmap/types.go`). Full map is F3: a scrollable place list plus a block of discovered rooms. Click or Enter on a neighbor exit sends that exit's direction, not a teleport. `canTravel` does not mean "walk there from anywhere".

## 4. Art

Sources, all HTTP, guest-readable:

| Art | URL | Code |
| --- | --- | --- |
| NPC portrait | `/api/portraits/<template>.png` | `pkg/portraits.URL`. Instances use the template id. |
| Player portrait | `/api/portraits/player-<race>-<class>.png` | `ForPlayer`. Races human, dwarf, elf. Classes warrior, rogue, mage, ranger, cleric, druid, ward. Rigger and unknown pairs return empty. Wizard maps to mage, hunter to ranger. |
| Item | `/api/item-art/<template>.png` | `pkg/itemart.URL` |
| Item fallback | `/api/item-art/generic-<bucket>.png` | Buckets: weapon, armor, quest, currency, consumable, junk (collectible and crafting_material), torch, default. |
| Room background | `room.meta.background` id | Play client calls `setBackground`. The file path is that id under the existing background route. Confirm the URL when implementing. Not re-verified against nginx in this spec. |
| Map tiles | `/api/map-tiles/<file>` | `server.go` serves embedded play assets. Optional for the TUI. A block map does not need them. |

Fallback chain, first that works:

1. Kitty graphics protocol (Unicode placeholders, RGBA). Best when the local terminal answers the query. Ghostty, Kitty, and WezTerm advertise it. WezTerm's ratatui-image notes say kitty placement is not the bug-free path there; prefer iTerm2 or sixel on WezTerm if the kitty query fails.
2. Sixel (DCS). foot, xterm (`-ti 340`), Windows Terminal, mlterm. 256-color palette. Re-send on every redraw is expensive. Cache the DCS blob for the cell size.
3. iTerm2 inline (OSC 1337). iTerm2, and some other terminals. Large base64. Prefer sixel or kitty when those queries succeeded.
4. Half-block (`▀` upper, `▄` lower) in truecolor. Two pixels per cell. This is the MVP default when queries time out.
5. Quarter-block or ASCII ramp (` .:-=+*#%@`) when color is 16 or mono.

Detection, once per session, budget 150ms after the first prompt, then give up. A local query is usually a few milliseconds. Do not block the first room paint on the query. Paint half-blocks, upgrade if a reply arrives. The client fetches PNGs over HTTPS and renders them here. The game server does not grow a graphics stack.

Sequence:

1. Honor `VEILSPAN_ART=half|sixel|kitty|iterm|off` and `NO_COLOR` before any query.
2. Read `TERM`, `COLORTERM`, `TERM_PROGRAM`. Hints only. `TERM=xterm-kitty` suggests kitty. `COLORTERM=truecolor` allows half-block color. Never treat a hint as proof of sixel.
3. Send DA1 (primary device attributes). A reply containing parameter 4 means sixel. Timeout 150ms.
4. Send kitty graphics query (`a=q`). A valid response means kitty. Timeout 150ms. Run this only if DA1 did not already settle the choice and the user did not force one.
5. iTerm2: no reliable query. Use it only when `TERM_PROGRAM=iTerm.app` or the user forced it.
6. tmux: if `TERM` starts with `tmux` or `screen`, do not send sixel or kitty unless the user forced it. tmux needs `allow-passthrough` and still corrupts DCS. Half-block is the default inside a multiplexer. Document that in the help line.

Local cache, per asset and per `(cols, rows, protocol)`:

- half-block cell grid
- sixel blob at the cell pixel size (use the terminal's reported cell size if XTWINOPS / CSI 16 t answered; else assume 8x16 and accept a blurry first frame)

Cache cap: 32 images, drop least recently used. A resize invalidates cell-sized blobs and keeps the PNG.

User override is a command in the TUI (`:art half`) stored in the local config. It applies on the next paint.

## 5. Input

### 5.1 Keys

When the command line is empty and no modal is open:

| Key | Action | Sends |
| --- | --- | --- |
| Arrows or n/s/e/w | Move if that exit exists | exit name |
| Enter on a highlighted NPC | Talk or attack if enemy and not in dialog | `talk <name>` or `attack <name>` |
| l | Look | `look` |
| a | Attack current target | `attack` |
| Shift-A | Confirm a stronger foe | `attack!` |
| d | Defend in combat, else drop is not bound | `defend` only in combat |
| f | Flee in combat | `flee` |
| 1–4 | Skill slot in combat | `1`..`4` |
| Tab | Cycle target in combat, else complete | — |
| i | Inventory | local panel |
| q | Quests. Not quit. | local panel |
| p | Party | local panel |
| m | Map | local panel |
| c | Character sheet | local panel |
| r | Rest | `rest` |
| ? | Help overlay of this table | local |
| Ctrl-C | Quit confirm | closes the session on yes |
| Esc | Close top modal, else clear the line | — |
| : | Focus the command line | — |
| Up/Down | History when the line is focused | — |

`q` is quests because `quit` is not a game command and muscle memory from other MUDs would disconnect people by surprise. Quit is Ctrl-C, or the words `quit` and `exit` on an empty exploration screen after a confirm. Those words are not sent to the server.

In combat, letters a/d/f/1–4 are the fight. Movement keys are ignored while `inCombat` except through the command line (`flee` is the way out).

### 5.2 Command line

Bubbles `textarea` is the wrong widget (multiline). Use a single-line text input. Completion on Tab, in order: command names from section 2.6, exit names in this room, NPC `displayName`, item names in bag and room, skill names from `/api/classes`. Completion replaces the token. It does not send.

History: 100 local lines. Up/Down walk it. Ctrl-R is optional later.

Mouse: SGR mouse when the terminal reports it (Bubble Tea v2 / Ultraviolet). Click an exit, an NPC, a foe, a dialog row, a shop row. Wheel scrolls the log. `--no-mouse` turns it off, because some terminals report a paste as mouse clicks. Selection for copy uses the terminal's shift-drag. The TUI does not capture shift-drag.

### 5.3 Accessibility

`--a11y`, or `:a11y on` saved in the config file:

- No alternate screen.
- No cursor-addressed full redraw. Print a line when something changes.
- Room enter prints: name, mood, exits, NPC names, one-line description.
- Combat prints: "your turn, N seconds", then the skill names, then results as the `message` field already provides.
- Dialog prints options as `1 text`.
- HP changes print only when the integer HP changes, not every frame.
- Bell optional on your turn (`BEL`), off by default.

This mode can run inside Bubble Tea's inline mode or as a thin printer over the same event stream. Prefer the printer so a screen reader is not fighting the Cursed renderer. Same events, different consumer.

### 5.4 Themes and keyboards

Themes: `veil` (default, dim gold on near-black, 16-color fallback uses yellow/white), `mono`, `paper` (dark on light). Lip Gloss v2 has no `AdaptiveColor`. Pick the theme from the background Bubble Tea reports. If the query times out, assume dark.

Kitty keyboard protocol: Bubble Tea v2 already decodes it. Bind actions to keys, not to bytes. Do not assume Ctrl-C is the byte `0x03` on every terminal. Still treat raw `0x03` as quit-confirm when the enhanced keyboard protocol is off.

Layout: physical keys. Do not localize n/s/e/w. Command words stay English because the server matches English verbs (`commandprocessor.go`).

## 6. Auth

Default login is the device-code flow against `/activate`. Guest is the other path. A pasted bearer is a developer escape, not the screen players see.

### 6.1 What the server already does

Checked on `99492b3`.

`pkg/devicecode.Store` is in-memory, shared today by the SSH lobby and the activate API, and lost on process restart. `Begin(ip, mode, client, fps)` opens one code. The comment on `Begin` says it opens a code for an SSH lobby, and that the display string is the only copy of the code. It returns an internal id, a formatted user code (shape `BCDF-GHJK`), and an expiry. Defaults when the config is unset (`store.go`): TTL 10 minutes, `MaxPendingPerIP` 3, `MaxPending` 100. `Take(id)` returns the confirmed `userRef` (an account id) once and marks the code consumed. It does not return a websocket token. Nothing in `pkg/server` exposes `Begin` or `Take` over HTTP.

The browser half is public and authenticated as follows (`pkg/server/server.go`, `pkg/server/handler/sshaccess.go`, `sshactivate.go`):

| Call | Who | Result |
| --- | --- | --- |
| `GET /activate?code=` | The human's browser | On a classic Auth0 process, redirect to `/play/?activate=<code>` unless embedded `activate.html` is present. A local-auth process serves its own confirm form. |
| `POST /api/ssh/device/lookup` `{user_code}` | Already signed-in account. Origin check. CSRF issued here. | `ip`, `created`, `mode`, `client_version`, masked `key`, `csrf`. Never the raw code. |
| `POST /api/ssh/device/confirm` `{user_code, csrf}` | Same account | `{ok:true}`. Binds that account's `userRef` onto the pending code. Does not return a token. |
| `POST /api/ssh/device/deny` | Same account | `{ok:true}`. |

Guests are rejected on lookup (`ssh_flow_test.go`). The TUI does not call lookup, confirm, or deny. The browser does, after the player is logged into play.

Guest login that this client can use today: `POST /api/guest` with an optional body `{templateId, race}` (`GuestHandler.CreateGuestSession`). Response `{token, expiresIn: 1800}`. The token is HMAC HS256. Claims include `sub`, `uid`, `exp`, `iat`, and `guest: true` (`pkg/service/guest.go`). `GuestExpiresAt` is 30 minutes. `AuthMiddleware` tries that HMAC first (`pkg/server/auth.go`). A guest claim on a user row that is not a guest is `"invalid guest token"`. There is no guest refresh route.

`pkg/authlocal.Service.Login` issues a token from a username and password. It has no `Refresh`. The server does not mint Auth0 access tokens. The only refresh in `pkg/server/auth.go` is the JWKS cache (one hour). That is not a user-token refresh. Auth0 refresh lives in the browser SDK.

### 6.2 What a non-SSH client can reuse

- The code alphabet and the `BCDF-GHJK` display format.
- `GET /activate` and the play redirect. The human confirms in a browser that is already logged into play.
- The authenticated lookup, confirm, and deny routes, unchanged.

### 6.3 What is missing

A public start and a public poll. Names here are the contract, not a promise that the paths are final:

- `POST /api/device/start` calls `Begin` and returns `{user_code, verification_uri, verification_uri_complete, expires_in, device_code, interval}`. `device_code` is the poll secret (the store id, or a separate secret). It must not equal `user_code`. `verification_uri_complete` is `https://veilspan.com/activate?code=<user_code>`.
- `POST /api/device/poll` takes that secret. While the code is pending it answers `authorization_pending` (and `slow_down` if the client polls faster than `interval`). On confirm, the handler calls `Take` and mints a token `AuthMiddleware` will accept on `GET /ws`.

That token cannot be a guest HMAC. Poll must issue a server-signed session token for the confirmed account, on a validation path `AuthMiddleware` already trusts (the local issuer, or a new issuer added beside it). Returning `userRef` alone is not enough to open `/ws`.

Until those two routes exist, the binary plays as a guest, or accepts `--token` for tests. It must not pretend the SSH lobby handed it a session.

### 6.4 Token file, refresh, logout

Store tokens under `os.UserConfigDir()/veilspan/`. Directory mode `0700`. File `tokens.json` mode `0600`. An OS keyring is optional later. The file is the v1 store.

Do not put the token on the process command line. Prefer `Authorization: Bearer` on the websocket dial. The server also accepts `?access_token=` (`auth.go`). The query form can land in proxy logs. The server redacts `access_token` in its HTTP log. Use the header.

One current token is enough for v1. The file may later hold `{accounts:[{label, token, guest}], current}`. Logout deletes the current token from the file and closes the socket. There is no server logout for a guest HMAC. Multiple accounts are optional and not an MVP gate.

Refresh, as of this commit:

| Token | Refresh today | What v1 does |
| --- | --- | --- |
| Guest HMAC | None. 30 minutes, then it dies. | On expiry or 401, offer a new guest or device login. Do not call a refresh URL. |
| Account token from poll | The route does not exist, so no refresh exists. | The minted token should carry `exp`. v1 stores it and does not refresh it. On 401, run device code again. A refresh route is a later engine change if `exp` is short. |
| Pasted Auth0 bearer | No game-server refresh. The browser SDK is not in this process. | Use it until `/ws` rejects it, then ask for device code or a new paste. |

### 6.5 After the socket is up

If `createCharacter` arrives, open the creator. Else `GET /api/my-characters`. One character: `sc` that name. Several: the picker. Remember the last choice in the config file. Never send `sc` while `inCombat`.

Close code `4001` (`closeSessionReplaced`) means the session moved. Show it and exit. Do not reconnect.

Guests do not get Party and Friends as account features on the web header. The TUI hides those panels when the stored token is a guest token. If a `party` message arrives anyway, show it.

## 7. Install and distribution

The binary is `veilspan`, from `github.com/atla/veilspan-tui`. Releases are GitHub releases of that repo. talesmud does not embed the TUI and does not grow an install script.

### 7.1 Release build

GoReleaser, `CGO_ENABLED=0`:

| Target | Archives |
| --- | --- |
| linux/amd64, linux/arm64 | `.tar.gz` |
| darwin/amd64, darwin/arm64 | `.tar.gz` |
| windows/amd64, windows/arm64 | `.zip` |

Each release publishes `checksums.txt` (sha256). Sign the checksum file (cosign or minisign). macOS: Developer ID, notarize with `notarytool`, staple the ticket. Windows: Authenticode when a cert exists. The first tag may ship unsigned if the signing accounts in section 12.2 are still open. The notes for that tag say so. An SBOM (Syft via GoReleaser) is optional and is not a v1 gate.

Homebrew: a formula in a tap repo (section 12.2), pointed at the release tarball, `sha256` taken from `checksums.txt`. Scoop: a bucket manifest with the Windows amd64 and arm64 zips. winget: a manifest whose package id is decided with the publisher. Those three files can live in the client repo until the tap, bucket, and winget repo exist. Publishing them is not the same commit as writing them.

`curl | sh`, checksum verified: a script shipped as a release asset. It picks OS and arch, downloads the archive and `checksums.txt`, checks sha256, and only then unpacks to `~/.local/bin/veilspan` (or `/usr/local/bin` when run as root). If the checksum file is missing or the hash does not match, the script exits and leaves the old binary in place.

`go install github.com/atla/veilspan-tui/cmd/veilspan@latest` is the toolchain path. It skips notarization. Document that next to the script.

### 7.2 Self-update

On startup, unless `--no-update-check` or config `update_check: false`, GET the GitHub releases API for the latest tag. A 403 or 404 skips the check. No other telemetry (section 12.2).

If the tag is newer, print one line: `veilspan X.Y.Z is available. Run: veilspan update`. Do not download yet.

`veilspan update` downloads the archive and `checksums.txt`, verifies sha256, then asks `Replace <path>? [y/N]`. Default is no. A no leaves the binary untouched. If the running path is owned by Homebrew, Scoop, or winget, print the package-manager command and do not replace the file.

### 7.3 Version handshake

Once per socket, after the upgrade, send `{type:"hello", proto:1, client:"<semver>", message:""}`. The empty `message` is required so the line is not a command (section 2.5). Do not repeat it on a timer.

The server does not reply today. Silence means proto 1, keep playing.

The reply to add with the other engine gaps:

| `status` | Client |
| --- | --- |
| `ok` | Play. |
| `upgrade-recommended` | Print `message` in the log. Play. |
| `too-old` | Print `message`, close the socket, exit non-zero. Do not enter the world. |

The reply also carries `proto` and `minProto`. The client speaks proto 1 only. If the server says the client proto is below `minProto`, that is `too-old`.

Compatibility: servers add fields as `omitempty`. Clients ignore unknown fields and unknown `type` values. A non-empty `message` on an unknown type still goes to the log. Removing a field, renaming one, or making a field required bumps `proto`. That bump waits until the play client and Flutter tolerate it (section 2.5).

### 7.4 What the player runs

`veilspan` with no args: load the token file, or offer guest vs device code, then dial `wss://veilspan.com/ws`. `--host` overrides the host for a dev server. `--token` uses a pasted bearer and does not write it unless `--save` is set.

## 8. Performance, reconnect, quit, tests

### 8.1 Frames

- Cap at 20 fps while idle-animating a timer, 8 fps is enough. The decision timer repaints once a second, plus once when it crosses each whole second. Do not repaint at 60.
- Combat log and room log are a ring of 500 lines.
- The Cursed renderer sends diffs. Do not also send a full screen on a timer.
- Art blobs are not inside the cell diff. Send graphics only when the image id or the cell rect changes.
- A room change is one cell diff plus at most one image. If the image is cold, half-block first.

### 8.2 Resize

`tea.WindowSizeMsg` rebuilds the layout. Invalidate art cell caches. If the new size is under 80x24, show one line, "window too small, Ctrl-C to quit or resize", and pause input other than quit. Stay on that screen until the terminal is large enough again.

### 8.3 Reconnect

On socket close other than 4001, retry with backoff 1s, 2s, 4s, max 15s, for 2 minutes, then show the error. On success, the server sends a fresh `enterRoom` / `characterSelected` as it does for any new socket. A newer login kicks this one with 4001.

A guest token dies after 30 minutes. When the socket closes because the token is expired, say so and offer a new guest or device login. Do not spin the backoff on a 401.

Resume: there is no client-side combat replay. After a new socket, show whatever the server sends. The v1 gap (combat snapshot on attach) is in section 2.4.

### 8.4 Quit

| Gesture | Result |
| --- | --- |
| Ctrl-C, or `quit` / `exit` after confirm | Close the websocket and exit the process. The game sees a disconnect and runs `handleUserQuit` (`game_handlers.go`). |
| Close the terminal | Same. |
| `party leave`, `flee` | Game commands. They do not disconnect. |

Do not send the text `quit` to the command processor. It is not registered, so today it does nothing useful and does not drop the link.

### 8.5 Tests

| Test | What |
| --- | --- |
| Golden | `teatest` at 80x24 and 160x48 for exploration, combat (your turn, victory with `unique`, defeat), dialog, shop, invite. Feed recorded JSON. Compare the view string. |
| Contract | Fixtures from `responses.go` constructors (`NewEnterRoomMessage`, `NewCombatEndMessage`, ...). The engine test fails if the committed JSON changes without an update. `internal/wire` in the client repo decodes the same files. |
| Commands | Table test: key to command string, including "no `sc` while inCombat" as a UI gate plus the server string still available. |
| Timer | `deadlineMs` in the past paints `0s` and does not go negative. A deadline 5s ahead paints `5s` then `4s` without a full-screen snapshot change except the timer cells. |
| Fuzz | Fuzz the JSON decoder and the layout splitter (cols, rows, random message order). |
| Resize | 80x24 view, then 40x12, stays on the too-small screen and does not panic. |
| CI | `go test` in `atla/veilspan-tui`. The engine fixture test runs in talesmud. No production websocket. |

## 9. Milestones

Agent-sized. Hours are focused implementation time for one agent, not calendar time and not a measured benchmark. Each client milestone lands in `atla/veilspan-tui` and does not deploy. The engine milestones are separate commits on talesmud.

### M0 — Wire fixtures and empty model (3–5 h)

- The engine test marshals every type in section 2.3 to JSON fixtures.
- `internal/wire` decodes those files. The client `go.mod` does not require `github.com/talesmud/talesmud`.
- `cmd/veilspan` paints "not connected" at 80x24.

Done when: both tests are green, and a grep of talesmud shows no `bubbletea` import under `pkg/mudserver`, `pkg/server`, or `pkg/entities`.

### M1 — Replay exploration (4–6 h)

- Feed a recorded `enterRoom`, `roomUpdate`, `roomPresence`, `message`.
- 80x24 exploration golden test matches section 3.1's regions (header, exits, log, prompt).
- Half-block placeholder for missing art. No network.

Done when: the golden test fails if the exit row disappears.

### M2 — Combat replay (4–6 h)

- `combatStart`, `combatTurn` with `deadlineMs`, `combatAction`, `combatStatus`, `combatEnd` victory with `unique:true`, defeat with `defeat`.
- Timer is a function of a fake clock.
- Keys 1–4 and `a`/`d`/`f` produce the command strings and do not send them yet.

Done when: a fixture with `loot[].unique` shows `UNIQUE`, and a deadline in 3.2s shows `3s` on the second-aligned clock.

### M3 — Auth (4–6 h)

- Guest: `POST /api/guest`, store the token at `os.UserConfigDir()/veilspan/` mode `0600`, logout deletes it.
- Device-code client: print `user_code` and the activate URL, poll with the interval. The poll test uses an httptest fake that returns `{token}` or `authorization_pending`. It does not claim the engine routes exist.
- `--token` still works and is not written unless asked.

Done when: a unit test round-trips a `0600` token file, guest JSON parses, and the poll client stays up on `authorization_pending` and then accepts a token. Section 6.3 stays the list of engine routes this milestone does not add.

### M4 — Websocket (4–6 h)

- Dial `/ws` with the stored token on `Authorization`.
- Send `{type:"message", message:<line>}`.
- Send `hello` once with `proto` and `client`. Silence is proto 1.
- Close `4001` exits. Other closes use the backoff in section 8.3.
- Quit closes the socket and exits. It does not send `quit`.

Done when: an httptest upgrade delivers one `enterRoom`, and a `4001` test exits. Do not point this at production.

### M5 — Panels (6–8 h)

- Dialog, shop, inventory (effects), quests, party invite accept/decline, character HTTP picker.
- `sc` blocked in the UI while `inCombat`.

Done when: golden tests for dialog and invite exist, and accept sends `party accept`.

### M6 — Art and a11y (4–6 h)

- Half-block from a PNG in the fixture.
- Query timeouts fall back to half-block.
- `--a11y` printer emits one line per `enterRoom` and does not emit cursor addressing.

Done when: `NO_COLOR=1` golden output has no ANSI color.

### M7 — Install and release (6–8 h)

- GoReleaser config for linux, darwin, and windows, each on amd64 and arm64, plus `checksums.txt`.
- The install script refuses a bad checksum in a test.
- `veilspan update` asks before replace and does nothing on the default answer.
- Homebrew, Scoop, and winget manifest files are in the repo. Publishing them waits on section 12.2.

Done when: `goreleaser release --snapshot` produces the six archives and the checksum file, the script test fails closed on a bad hash, and declining the update prompt leaves the binary in place. This milestone does not publish a tap.

### M8 — Engine gaps (separate, only if Marcus wants)

- Public device start and poll that return a non-guest token usable on `/ws` (section 6.3).
- `aggroOnSight` on `RoomNPC`.
- `examineResult`.
- `combatStatus` on attach when already in combat.
- Server `hello` reply with `ok` / `upgrade-recommended` / `too-old`.

Each is its own commit on the engine, with a client fixture update when the JSON changes. Not part of the client's first tag. Account login in the binary stays on the httptest fake until the poll route is real.

Acceptance for the whole spec: a player can install `veilspan`, play as a guest (and, once M8's poll exists, with a device code), explore, fight, talk, and quit by closing the websocket and exiting. Line-mode SSH on `:2222` is unchanged and is not part of this acceptance.

## 10. Feature translation matrix

Every row is a feature that exists in this tree (engine command, socket message, play client, Flutter client, or creator app). Priority is when the TUI should grow it. Protocol is whether a structured message exists on `99492b3`.

`partial` means some fields exist and the panel still depends on prose or on a second HTTP call.

### 10.1 Rooms, movement, look

| Feature | TUI | Keys | Terminal vs web | Pri | Protocol |
| --- | --- | --- | --- | --- | --- |
| Room enter, description | Room panel + log | `l`, `r` | Text description is the native form. Web art is richer until half-block. | MVP | yes `enterRoom` |
| Exits, hidden exits | Exit row | n/s/e/w and the exit name | Single-key move is better than a compass click. Hidden exits already stripped server-side. | MVP | yes, on `room.exits` |
| Room actions | Hint row | type the action name | Same typed name as the play client. | MVP | partial (`room.actions` on the room object; result is `message`) |
| Mood | Header | — | One word. Web can tint the whole scene. | MVP | partial `room.meta.mood` |
| Room art `meta.background` | Art cell | `:art` | Degrades to blocks. Better only when kitty/sixel works. | v1 | partial (id only, bytes via HTTP) |
| Room items | Ground row | `get <name>` | Names, not sprites. | MVP | partial `items[]` id/name/noPickup |
| Other players | Here row | — | Same data as web. | MVP | yes `players` / `roomPresence` |
| NPCs, threat, state | Here row | `t` talk, `a` attack | Threat word is enough. No sprite scale. | MVP | yes `npcs[]` except aggro flag |
| Aggressive (L10 content, `AggroOnSight`) | `AGGRESSIVE` tag once the field exists | — | Web does not show it on this commit either. Engine still aggros (`game_aggro.go`) after grace (`AggroGrace`, default 2.5s) unless the player is beyond `AggroMaxLevelGap`. | v1 | no on the room card |
| Examine | Overlay | `x` | Web parses prose (`parseExamineOverlay.js`). Terminal can show the same prose immediately. | MVP prose, v1 struct | no |
| Minimap / atlas | 3x3 then F3 | `m` | Block map is honestly good. Pixel terrain tiles are not. | v1 | yes `atlas` |
| Follow leader | Status on party | `party follow` | Same rule: no follow into an instance. | v1 | partial `party.following` |
| Instances | Log only | the exit | No special chrome. | later | no (exit `instance` bool is on the exit) |

### 10.2 Combat

| Feature | TUI | Keys | Terminal vs web | Pri | Protocol |
| --- | --- | --- | --- | --- | --- |
| Start, roster, portraits | Combat frame | — | Portraits become blocks or kitty. HP bars are native. | MVP | yes `combatStart` |
| Rounds, whose turn | Header | — | Clearer than the web pill that play now hides. | MVP | yes `combatTurn` |
| 5s decision timer | Countdown | — | `deadlineMs` is absolute. Better than a local guess. | MVP | yes |
| Auto-attack on timeout | Hint line | — | Server does it. Client must not send a second attack when the timer hits zero. | MVP | yes (absence of queue) |
| Attack, defend, flee, focus | Bar | a, d, f, Tab | Same commands. `focus` exists (`commands/focus.go`). | MVP | partial (commands in, result via `combatAction`) |
| Skills hotbar 1–4 | Slots | 1–4 | Engine cap is 4. Web's 9 local slots do not all exist on the server. | MVP | partial `equippedSkills`, `maxSkillSlots`, cooldowns on queue state |
| Skill definitions | Tooltip | — | HTTP `GET /api/classes`. | MVP | no on the socket |
| Party vs foes | Two columns | Tab | Layout B allies as a compact card. | MVP | yes `players` / `enemies` |
| Event cards, hooks | One event row | — | `style=combatEvent`. A card is a log line with a label. Web animates. | MVP | partial (the line, not a duration) |
| Summons | New foe row | — | Appears on the next combatant list. | MVP | partial |
| Boss phase, telegraph, enrage | Foe row | — | Text chips. No screen shake. | v1 | yes on `CombatantView` |
| Unique drops | Victory row `UNIQUE` | Enter dismisses | Same signal as the web diamond (`loot[].unique`). | MVP | yes `combatEnd` |
| Victory rewards, defeat, flee | End panel | Enter | Numbers are structured. Good fit. | MVP | yes |
| Loot onto the ground | Ground row after `roomUpdate` | `get` | Names. | MVP | partial |
| Rigger bolt / rig | Extra binds if the class has them | `bolt`, `rig` | Once per fight, forfeits the swing. | v1 | partial (commands; once-per-fight is class HTTP) |
| Orange-or-worse confirm | Shift-A | `attack!` | Keep the warning. Do not auto-confirm. | MVP | no (prose error) |
| In-combat regen tick | Bar tick | — | `characterUpdate` when the server sends it. Regen rates live in ruleset, not in the client. | v1 | partial |
| Switch character mid-fight | Disabled | — | Server refuses. UI matches. | MVP | partial (`inCombat` plus prose refusal) |

### 10.3 Talk, quests, economy

| Feature | TUI | Keys | Terminal vs web | Pri | Protocol |
| --- | --- | --- | --- | --- | --- |
| Dialog options | Modal | 1–9, Esc | Better than typing the sentence. Index is the command. | MVP | yes `dialog` |
| Dialog end | Close modal | — | | MVP | yes `dialogEnd` |
| Quest log, progress, ready, abandon, turn-in | Panel | F2 or q, t, a | Structured. | MVP | yes |
| Quest accept banner | Log line | — | | MVP | yes `questAccepted` |
| Shop buy/sell by row | Panel | b, s | Stock is structured. Sell side is the bag. | MVP | yes `shop` + inventory |
| Repair, value | Command line | `repair`, `value` | Prose result. | later | no |
| Craft, recipes, stations | Panel | F4, Enter | `canCraft` / `stationOk` / `have`. Errors still prose. | v1 | partial |
| Gather | Command | `gather` | Prose. | later | no |
| Item effects | Detail pane | x | `FormatEffect` lines. Native to text. | MVP | partial (on shop stock and on items inside `inventoryUpdate` when the entity has `effects`) |

### 10.4 Character, social, body

| Feature | TUI | Keys | Terminal vs web | Pri | Protocol |
| --- | --- | --- | --- | --- | --- |
| Classes, create, select | Picker | arrows, Enter | HTTP plus `sc`. Kit names from `/api/classes`, not hardcoded. | MVP | partial (`characterSelected`, `createCharacter` text, HTTP for the form) |
| Level-up card, spend | Modal | `spend str 1` | Structured `levelUp`. | v1 | yes |
| Stats, attributes, derived combat numbers | Sheet | c | `characterUpdate` has attack, defense, regen. | MVP | yes |
| Rest | Header flag | r | `resting` bool. Does not fill HP. | MVP | partial (flag plus later HP updates) |
| Bind, respawn | Command, defeat panel | `bind` | Defeat names the room. Bind success is prose. | v1 | partial |
| Who | Log | `who` | Prose. | later | no |
| Say, tell, scream, shrug | Log | `say` | Prose with `username`. Color the name. | MVP | partial |
| Party roster, follow | Panel | p | Structured. | MVP | yes `party` |
| Party invite popup | Modal | A / X | 45s `expiresAt`. | MVP | yes `party_invite` |
| Friends | Panel | — | Structured list. Add is prose. | v1 | partial |
| Guests | Hide party/friends, show a login hint | — | Same as the web header. Guest means the stored token is a guest HMAC. | MVP | no guest flag on the socket |
| Notifications | Log + modal for invite and level-up | Esc | No toast stack. One modal. | MVP | partial |

### 10.5 Client chrome the web has and the TUI should not copy

| Web / Flutter surface | TUI decision | Pri |
| --- | --- | --- |
| Widget grid, edit layout (`layout/WidgetRegistry.js`) | Fixed breakpoints. No drag grid. | never |
| Nine-slot local hotbar with commands and items (`hudPrefs.js`) | Four server slots. | never as protocol |
| CRT terminal widget | The whole app is the terminal. | never |
| Settings modal (motion, scale) | `:set` and a local file. Reduced motion = no timer blink. | v1 |
| Shortcut sheet | `?` | MVP |
| Account menu, Auth0 buttons | Device code and guest, before the 80-col frame (section 6). | MVP |
| Phone bottom sheet | 80-col panels already are sheets. | — |
| Flutter battle FX | Skip. `fxId` may color the log word. | later |
| Reconnect overlay | Section 8.3 | MVP |

### 10.6 Door, and admin later

Door features that stay Door's: CP437 splash, activate screen templates, `door_key` single-key commands, `doorview` phase/draft UI, resource/daily lines that are pack text. The Veilspan TUI does not grow a Door layout.

Admin worth a later ops TUI, not this client. All of these live in `public/app` and HTTP, not on `/ws`:

| Creator surface | Later ops TUI idea | Protocol |
| --- | --- | --- |
| Rooms, grid world, exits (`RoomsEditor`, `GridWorldEditor`) | Out of scope. Editing YAML in a TUI is a different product. | HTTP |
| NPC editor, enemy trait, inspector | Read-only `aggroOnSight` is the player-facing slice. | HTTP |
| Dialog graph | No. | HTTP |
| Items, loot tables, spawners | No. | HTTP |
| Quests editor, quest debug | No. | HTTP |
| Scripts | No. | HTTP |
| World health, drift, validation, search palette | A future separate ops binary could call the health endpoints. Not this client. | HTTP |
| Live ops (`OpsButtons.svelte`, `RunOp`) | Dangerous (teleport, give, end combat). If ever, a separate binary, creator auth, confirm. | HTTP `RunOp` |
| Users, guests stats | No. | HTTP |
| Audit log | No. | HTTP |

## 11. Features with no structured protocol yet

Text or HTTP only, on this commit. The TUI must not pretend these are socket messages.

1. Examine / inspect body (`examine`).
2. Who list (`who`).
3. Say, tell, whisper, pm, scream, shrug bodies (a `message` with `username` is all there is).
4. Help text (`help`).
5. Bind success (`bind`).
6. Repair and value (`repair`, `value`).
7. Gather results (`gather` / `forage` / `harvest`).
8. Craft success and failure (prose; the recipe list itself is structured).
9. Friend add/remove confirmations (`friends` roster is structured; the verbs are not).
10. Skill list command output (`skills`). Definitions are HTTP `GET /api/classes`.
11. Character list and create form (`lc`, `nc` prose). Use `GET /api/my-characters` and `POST /api/characters`.
12. `createCharacter` socket message (one sentence, no roster).
13. Attack warning that requires `attack!`.
14. Quit / disconnect. The client closes the websocket and exits. No quit command is sent.
15. `aggroOnSight` on the room NPC (trait exists, payload does not).
16. Combat snapshot for a client that connects mid-fight.
17. Hotbar item binds and slots 5–9 (web local storage only).
18. Settings, themes, layout.
19. Guest-vs-account as a socket field. The client uses the token it stored (guest HMAC vs account token).
20. Admin, creator, world health, drift, search, loot editor, spawners, ops.
21. `combatLeave` (web handler only; server does not send it).
22. Door key events (`door_key`) — Door's, not this client's.

Section 2.4 is the subset worth adding to the socket. This list is the full "do not scrape, do not fake" set.

## 12. Risks and open questions

### 12.1 Risks

| Risk | Mitigation |
| --- | --- |
| The client imports talesmud "just for one struct" and pulls the game | `go.mod` does not require the engine module. CI in M0. Fixtures are JSON. |
| Account login is blocked on two routes that do not exist yet | Guest and `--token` work without them. M3 tests the poll client against a fake. M8 is the engine change. Do not call `Begin` from the SSH lobby and pretend that is the binary. |
| `veilspan update` replaces a binary the player did not mean to replace | Checksum first. Ask. Default is no. If Homebrew, Scoop, or winget owns the path, print their command and do not write the file. |
| Graphics DCS corrupts the next frame inside tmux | Default half-block when `TERM` is tmux/screen. |
| The 5s timer feels late on a slow link | Display server `deadlineMs`, not a local 5s. |
| Player hits `q` and expects quit, gets quests | Help line and `?`. Quit is Ctrl-C and the words `quit`/`exit` with confirm. Those words are not sent. |
| `sc` during a fight | UI lock plus the server sentence from `99492b3`. |
| Door and this TUI diverge on ANSI safety | Copy the sanitize rules. Do not share the view. |
| Fixture drift between talesmud and `atla/veilspan-tui` | Engine test marshals `responses.go`. Client CI decodes the same files at a pinned commit. |

### 12.2 Open questions for Marcus

1. Repo name and owner: is `atla/veilspan-tui` (module `github.com/atla/veilspan-tui`) the one to create?
2. May the engine gain a public device start and poll that returns a non-guest token usable on `/ws`, and may that plus the other M8 fields land on `engine-june`?
3. Who holds the Apple Developer account for signing and notarization, and is there a Windows Authenticode certificate?
4. Homebrew tap repo name and owner? Same question for the Scoop bucket and the winget publisher.
5. Confirm telemetry is none. This spec sends no analytics. The only startup request besides the game is the GitHub releases check, and `--no-update-check` turns that off.
6. Rigger and Ward are in `pkg/classkit/testdata`. Is the live pack's `GET /api/classes` the same five? The TUI trusts the HTTP list either way.
7. Is a separate ops TUI in scope this quarter, or is the note in section 10.6 enough?

### 12.3 Not verified

- The live process on veilspan.com (which class pack is loaded). Docs and this tree only.
- The Doorkeeper worktree's private UI beyond `pkg/presentation/doorview` and `public/door` in this commit. `~/dev/talesmud-door-mud` exists and was not used as a second source of truth.
- Bubble Tea v2.1.0 (2026-10-08) was checked only from its release notes, not built against.
- Exact HTTP path that serves `room.meta.background` bytes.
- Whether Flutter on disk outside this repo handles `party_invite`. The `game_view_state.dart` switch in `~/dev/veilspan-client` does not.

## 13. Implementation note for docs

This spec does not change `PROJECT.md`, `ARCHITECTURE.md`, or `FEATURES.md`, because no behavior changed. When an engine change from M8 lands, update those three for the new routes or socket fields. The TUI repo gets its own README. Do not add `cmd/veilspan` to talesmud.
