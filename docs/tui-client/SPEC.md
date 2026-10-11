# Veilspan TUI client

Status: spec only. No implementation in this change.
Base: `engine-june` `99492b3` (character switch blocked mid-fight).
Date: 2026-10-11. Owner ask: Marcus.

This document specifies a terminal client for the classic Veilspan world. It does not specify a replacement for the Door text client on `door.veilspan.com:2223`.

## 1. Executive summary

Build one Go program, `cmd/veilspan-tui`, on Bubble Tea v2, Lip Gloss v2, and Bubbles v2. The same `tea.Model` runs two ways:

1. Server-side, on an already-admitted SSH session on the existing `:2222` listener, when the session has a PTY and a usable size. Zero install. `ssh -t -p 2222 veilspan.com` lands in the TUI.
2. As a downloadable binary on the player's machine. It dials the existing game socket `GET /ws` (`pkg/server/server.go`) and speaks the same JSON.

Both paths are protocol clients. The model consumes the structs that `wsTransport.Send` already JSON-encodes (`pkg/mudserver/transport.go`). It does not scrape ANSI, and it does not import `pkg/mudserver/game`. Core packages must not import Bubble Tea, Lip Gloss, Wish, or the TUI module.

The line-based SSH mode stays. `pkg/sshgate` plus `pkg/presentation/textline` remain the classic client. The TUI is the default when the PTY is at least 80x24 and `TERM` is not `dumb`. A small window stays on the line editor. Opting out on a large window needs `VEILSPAN_CLASSIC=1`, which `acceptEnv` does not allow yet (section 1.4). The downloadable binary takes `--classic` and `--a11y` as local flags. Screen-reader mode is a third presentation of the same events: linear lines, no alternate screen, no cursor-addressed redraw.

Rust plus Ratatui is the stronger pixel-art toolkit and a proven SSH-game stack. It is the wrong default here. The engine, the socket types, `sshgate`, portraits, item art, and the Door ANSI helpers are Go. A second language would fork the contract the web client and the Flutter client already share.

### 1.1 Recommendation

| Choice | Decision |
| --- | --- |
| Language | Go |
| UI | `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2` |
| Renderer | Bubble Tea v2 Cursed renderer (cell diff, mode 2026). This is the bandwidth plan for SSH. |
| SSH server | Keep `pkg/sshgate` (gliderlabs/ssh). Do not stand up a second listener. |
| Wish | Reference and optional helper for "one program per session", not a replacement lobby. Auth, device code, guest caps, and key linking stay in `sshgate`. |
| Wire | Existing JSON. New fields only as additive messages. |
| Art | Kitty, then sixel, then iTerm2 inline, then half-block truecolor, then ASCII. Half-block is the MVP picture. |
| Classic SSH | Unchanged `textline` editor. |

### 1.2 Weighted comparison

Weights are 1–5 by how much they matter for this product. Scores are 1–5. Higher is better. Weighted total is sum of weight times score.

| Criterion | W | Go + Bubble Tea v2 + existing sshgate | Rust + Ratatui 0.30.2 + russh |
| --- | --- | --- | --- |
| Shares the Go engine and JSON types without a second decoder stack | 5 | 5 | 2 |
| Zero-install over the SSH port that already exists | 5 | 5 (session handoff after admit) | 3 (russh can serve, but it replaces or duplicates `sshgate`) |
| Same binary as a downloadable websocket client | 4 | 5 | 4 |
| Terminal graphics (sixel, kitty, iTerm2, half-blocks) | 4 | 3 (`charmbracelet/x` has sixel and kitty encoders; no Ratatui-image equivalent) | 5 (`ratatui-image` 11.1.0 stable, 12.0.0-rc.2 on 2026-10-05) |
| SSH bandwidth / redraw diffing | 4 | 5 (Cursed renderer, mode 2026; Wish v2 notes order-of-magnitude savings) | 4 (immediate mode, app owns the diff) |
| Fits the "core stays framework-neutral" rule | 5 | 5 if the TUI lives in `cmd/` and a leaf package | 5 (separate process) but the in-process adapter is then a Go bridge anyway |
| Team velocity on this repo | 4 | 5 | 2 |
| Snapshot tests | 3 | 4 (`teatest`) | 5 (mature golden tests) |
| Screen reader / classic fallback beside a TUI | 3 | 4 (classic path already shipped) | 3 |
| Risk of fighting Door's ANSI client | 3 | 4 if we do not import `doorview` | 5 (fully separate) |

Weighted: Go about 200, Rust about 158 (weights sum 40; Go 4.9, Rust 3.7 average). Graphics is the one row where Rust wins cleanly. Pay that down with half-block first and a small Go encoder for kitty/sixel, not with a second client.

### 1.3 What "both ways" means

```
+------------------+     JSON events      +----------------------+
| tea.Model        |<--------------------| session adapter      |
| panels, keys     |  command lines       |                      |
|                  |-------------------->|                      |
+------------------+                      +----------+-----------+
                                                     |
                         +---------------------------+---------------------------+
                         |                                                       |
              downloadable binary                                    SSH after admit
              dial wss://host/ws?access_token=                       Transport.Send(v)
              gorilla/websocket                                      AttachExternal input()
                         |                                                       |
                         +---------------------------+---------------------------+
                                                     |
                                            pkg/mudserver  (unchanged contract)
                                            web /play and Flutter use this too
```

Server-side, the adapter is a `mudserver.Transport` (`pkg/mudserver/transport.go`). `Send(v any)` JSON-encodes the same value `wsTransport` writes. `AttachExternal` (`pkg/mudserver/mudserver.go`) already delivers command lines. The TUI sends lines through that callback (`input(text, false)`). It does not call game methods.

A loopback dial to `/ws` is the conformance test that the adapter and the socket match. It is not the production SSH path. An extra hop on the same host adds latency to the 5 second decision window and breaks when the process is bound to a socket the TUI cannot see.

### 1.4 Coexistence with line mode and with Door

Admission does not change. Guest, device code, and linked-key login stay in `pkg/sshgate/lobby.go` and `pkg/sshgate/session.go`. A shell with no PTY is already rejected (`session.go` `shell` request). `exec` is refused. `acceptEnv` allows only `LANG`, `LC_ALL`, `LC_CTYPE`, and `TALES_CHARSET`. So `ssh host classic` does not work today, and a client `SendEnv VEILSPAN_CLASSIC` is dropped until the allowlist grows.

After admit, `classicLoop` is what runs now. The TUI choice replaces that call:

| Condition | Session |
| --- | --- |
| `TERM=dumb`, or cols < 80, or rows < 24 (pty-req and window-change already parse size) | Classic `textline` editor. `ssh.mud.history` defaults to 20 (`docs/ssh-access.md`). |
| `VEILSPAN_CLASSIC=1` once `acceptEnv` allows that name | Classic, even on a large terminal. |
| `VEILSPAN_A11Y=1` once allowed | Linear mode. Same events, no alternate screen. |
| Otherwise | Bubble Tea on that channel. |

M5 adds those two names to `acceptEnv`, sanitized the same way as `TALES_CHARSET`. OpenSSH only sends them if the player's config has `SendEnv`. Players who do nothing get the TUI on a normal 80x24-or-larger PTY. Ctrl-C / Ctrl-D keep their classic meaning inside the line editor. Inside the TUI, Ctrl-C asks for quit and then closes the channel (section 7).

Door (`:2223`, `gamemode.ANSI()`, `pkg/presentation/doorview`) is a different product. `doorview.View` holds `*game.Game` and paints from live engine objects. The Veilspan TUI must not do that, and must not import `doorview` (that import would pull the game into the client). Reuse, by copying or by a later leaf extract:

- `pkg/presentation/ansi`: frame sanitizer. Door frames drop OSC, DCS, C1, and bidi controls (`docs/ssh-access.md`).
- `pkg/presentation/termout`: charset selection for CP437 art.
- The idea in `doorview` of mapping a key to an engine command string. The map itself stays Door's.

Do not share a screen layout, a color theme, or a binary with Doorkeeper.

### 1.5 Library facts (checked 2026-10-11)

| Library | State | Source |
| --- | --- | --- |
| Bubble Tea v2.0.0 | Stable 2026-02-24. Cursed renderer, `tea.View`, mode 2026 sync updates, Kitty keyboard protocol, Lip Gloss no longer touches I/O. | https://charm.land/blog/v2.md (post date 23 Feb 2026). Patch v2.0.10 on 2026-09-24 skips capability queries when input is disabled: https://github.com/charmbracelet/bubbletea/releases (via newreleases, retrieved 2026-10-11). Latest is v2.1.0 (2026-10-08): OSC 7501 program status, 0x0 tty size falls back to defaults, shutdown deadlock fix. Target v2.1.x. |
| Lip Gloss v2.0.0 | Stable 2026-02-24. Pure styling. Import `charm.land/lipgloss/v2`. | https://github.com/charmbracelet/lipgloss/releases/tag/v2.0.0 |
| Bubbles v2.0.0 | Stable 2026-02-24. Import `charm.land/bubbles/v2/...`. Light/dark styles are explicit. Wish must read the client background; do not use Lip Gloss `compat` on SSH. | https://github.com/charmbracelet/bubbles/releases (v2 notes, retrieved 2026-10-11) |
| Wish v2.0.0 | Stable 2026-03-09. Built on Bubble Tea v2. Import `charm.land/wish/v2`. | https://github.com/charmbracelet/wish/releases (v2.0.0 notes, retrieved 2026-10-11) |
| Ultraviolet | Cell diff renderer under Bubble Tea v2. Kitty keyboard, SGR mouse, Windows console input. | https://github.com/charmbracelet/ultraviolet (retrieved 2026-10-11) |
| charmbracelet/x | Experimental. `ansi` package gained sixel (`DCS`) and kitty graphics helpers (PR #380, merged Feb 2025). | https://github.com/charmbracelet/x/pull/380 |
| Ratatui | 0.30.2 on 2026-06-19. Backend split; `ratatui-crossterm` 0.1.2. | https://crates.io/crates/ratatui |
| ratatui-image | 11.1.0 stable 2026-09-17. 12.0.0-rc.2 on 2026-10-05. Sixel, kitty, iTerm2, half-blocks. | https://lib.rs/crates/ratatui-image |
| russh | 0.64.1 on 2026-10-05. Used by SSH-delivered Ratatui games (kartoffels). | https://crates.io/crates/russh/0.64.1 |

Not verified: any Bubble Tea release after v2.0.10, and whether `charmbracelet/x` sixel is the API we should pin. Pin versions in `go.mod` when implementation starts.

### 1.6 Lessons from other terminals

| Work | Take |
| --- | --- |
| terminal.shop, Charm soft-serve | Wish/Bubble Tea over SSH is a product, not a demo. Capability detection and resize are the hard parts. |
| Brogue CE, Caves of Qud (curses) | A fixed grid, a small verb set, and a message log beat a free-form terminal. 80 columns is the floor. |
| Cataclysm DDA | Dense panels work when every panel has a key and a one-line help row. |
| TinTin++, Blightmud | Scriptable triggers are how experts escape bad UIs. We do not need them in MVP. We do need the command line to stay first-class so muscle memory from classic SSH still works. |
| Mudlet | GUI maps and triggers set the expectation. Our atlas message is the structured map. Do not redraw a pixel map by scraping room text. |
| Evennia web | Confirms the split we already have: server sends structured state, client renders. |
| `ssh tetris` / SSH toys | Fun and brittle. They skip auth, resize, and reconnect. We already have those problems solved for the line client. Do not throw them out. |

## 2. Architecture

### 2.1 Modules

| Path | Role | May import |
| --- | --- | --- |
| `cmd/veilspan-tui` | `main`. Flags, SSH-channel vs stdio, websocket dial. | `pkg/tui`, Bubble Tea |
| `pkg/tui` | Model, views, keys, art fallback, themes. | `pkg/wire` (new, leaf), Bubble Tea. Not `mudserver`, not `doorview`, not `game`. |
| `pkg/wire` | JSON structs that match today's socket. Decode only. | stdlib |
| `pkg/sshgate` | Admit, then choose classic vs TUI. | May construct the TUI adapter. Must not import Bubble Tea types in the lobby. The session file can depend on a small `tui.Runner` interface defined in `pkg/sshgate` and implemented in `cmd` or `pkg/tui`. |
| `pkg/mudserver` | Unchanged contract. | Must not import `pkg/tui`. |

`pkg/wire` is allowed to duplicate the JSON tags from `pkg/mudserver/game/messages/responses.go`. A later extract that moves those structs into `pkg/wire` and has `messages` alias them is optional and is the only core edit this spec wants. Until that extract, the TUI carries its own structs. Drift is caught by the fixture test in section 7.

`pkg/bus` (`pkg/bus/bus.go`) is an in-process event bus. It is not the player socket. The TUI does not subscribe to it.

### 2.2 Directions

| Direction | Today | TUI |
| --- | --- | --- |
| Client to server | One JSON object per command. `messages.IncomingMessage` is `{message, type, key}` (`pkg/mudserver/game/messages/incoming.go`). Web play sends `{message, type:"message"}` (`public/mud-client/src/game/Client.js`). Door sends `{type:"door_key", key}` (`public/door/door.js`). | SSH adapter calls `input(line, false)`. Downloadable client sends `{message, type:"message"}`. No new client verb in MVP. Panels format the same command strings the keyboard already sends (`attack`, `1`, `party accept`). |
| Server to client | `Transport.Send` of a struct with `type`. Audience fields are `json:"-"` and are not on the wire. | Adapter marshals that value. Unknown `type` is ignored, except a non-empty `message` is appended to the log (Flutter already does this fallback in `veilspan-client/lib/src/game/game_view_state.dart`). |

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

These are the gaps that make a panel lie or fall back to prose. Priority matches section 9.

| Gap | Why it hurts | Proposal | Priority |
| --- | --- | --- | --- |
| No `hello` / protocol version / capability | TUI cannot say "I understand combatEnd.loot" vs an older process | Client may send `{type:"hello", proto:1}` with an empty `message` once. Classic `handleInput` ignores an empty line (`text != ""` before enqueue) but still runs `ConnectUserSession` and a user update, same as any other read (`mudserver.go`). Do not put `hello` in `message` or it runs as a command. A server reply `hello` is new and optional. `door_key` is the only incoming `type` with special handling. | v1 |
| `aggroOnSight` not on `RoomNPC` | `EnemyTrait.AggroOnSight` exists (`pkg/entities/npcs/enemy_trait.go`) and the scheduler runs (`game_aggro.go`). The room card does not say so. Play client in this tree has no Aggressive label. | Add `aggroOnSight` bool on `RoomNPC`. Until then the TUI has only the prose line when the fight starts. | v1 |
| Examine is prose | `examine` / `inspect` reply with `message` text (`commands/examine.go`). Web parses it (`parseExamineOverlay.js`). | `examineResult` with name, slot, quality, stats, `effects[]` using `items.FormatEffect` ("On hit — Name: text"). | v1 |
| Who, say, tell, emote are prose | `who`, `say`, `tell`/`whisper`/`pm`, `scream`, `shrug` are `message` lines. | Keep prose in the log. A `who` snapshot can wait. | later |
| Skills list is not a message | `skills` prints text. Equipped ids arrive on `characterUpdate.equippedSkills`. Definitions are not on the wire. Web seeds a local catalog (`hudPrefs.js`) and also reads `GET /api/classes` (`classkit.ClientSkill`). | TUI loads `GET /api/classes` for names, tooltips, cooldown, target, once-per-fight. Do not hardcode Fenwatch skill text. | MVP (HTTP, not a new socket type) |
| Hotbar layout is client-local | Engine registers shortcuts `1`–`4` only (`commandprocessor.go`). `skills.HotbarCap` is 4 (`pkg/entities/skills/classkit.go`). Web shows 9 slots in `localStorage`, mixing skills, items, and commands (`hudPrefs.js`). | TUI shows `maxSkillSlots` from `characterUpdate` (4 for kit classes). Keys 1–4 send `1`..`4` in combat. Extra web slots are not protocol. Item-on-hotbar waits. | MVP |
| Inventory is a full dump, not a delta | Fine at current bag sizes. | Keep the dump. Do not invent a delta. | — |
| No art cell size on the room | `room.meta.background` is an id. Bytes live at portrait and item-art URLs. | Section 4. No new socket type for MVP. Optional later: `art` message with a content hash and cell-size variants. | later |
| `quit` does not disconnect | There is no `quit` command. Classic SSH leaves only when `textline` sets `Quit` on confirmed Ctrl-C / Ctrl-D (`pkg/presentation/textline/editor.go`, `sshgate/session.go` `onBytes`). | TUI quit closes the SSH channel or the websocket. Do not send the word `quit` and hope. Optional later: server `session` message `{op:"bye"}` so a remote kick and a local quit look the same. Close `4001` already covers "session moved". | MVP behavior, later message |
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
- A breaking change (rename, or a new required field) bumps `proto` and is not allowed on `:2222` until both play clients tolerate it.
- The TUI may send one empty-message `hello` and continues if the server never answers. Absence of a reply means "classic server, proto 1 as documented here". The empty read is not a command. It does touch presence, so do not send it on a timer.

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

`quit` typed at the classic prompt does not drop SSH. Ctrl-C and Ctrl-D do, after `Quit? [Y/N]`. The TUI's quit path calls `Transport.Close` / channel close, the same end state as a confirmed classic quit.

## 3. Screens

Coordinates are terminal cells. The minimum is 80x24. Below that, the session stays classic (section 1.4). Resize reflows on the next frame. No panel stores a pixel position.

Breakpoints:

| Size | Layout |
| --- | --- |
| 80–109 cols, or rows < 30 | One column. Art up to 16x6. Log at least 4 rows. Panels replace the room (inventory, quests, map) instead of sitting beside it. |
| 110–159 cols and rows >= 30 | Art beside description. Log under both. Hotbar is one row. |
| 160–199 cols | Art, description, and a 28-col minimap. Log gets the extra rows. A 160x48 screen is this split plus 24 more log rows. |
| >= 200 cols and >= 48 rows | The 160 layout plus a right inspector (examine, quest pin, or who). 200x60 is that inspector at full height. |
| rows < 24 | Refuse the TUI. Classic. |

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

Log is a ring (section 7). `combatEvent` lines are prefixed and kept when the fight UI opens.

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

1. Kitty graphics protocol (Unicode placeholders, RGBA). Best over SSH when the terminal answers the query. Ghostty, Kitty, and WezTerm advertise it. WezTerm's ratatui-image notes say kitty placement is not the bug-free path there; prefer iTerm2 or sixel on WezTerm if the kitty query fails.
2. Sixel (DCS). foot, xterm (`-ti 340`), Windows Terminal, mlterm. 256-color palette. Re-send on every redraw is expensive. Cache the DCS blob for the cell size.
3. iTerm2 inline (OSC 1337). iTerm2, and some other terminals. Large base64. Avoid on slow SSH if sixel or kitty worked.
4. Half-block (`▀` upper, `▄` lower) in truecolor. Two pixels per cell. This is the MVP and the SSH-safe default when queries time out.
5. Quarter-block or ASCII ramp (` .:-=+*#%@`) when color is 16 or mono.

Detection, once per session, budget 150ms after the first prompt, then give up. Over SSH the round trip may be 80ms. Do not block the first room paint on the query. Paint half-blocks, upgrade if a reply arrives.

Sequence:

1. Honor `VEILSPAN_ART=half|sixel|kitty|iterm|off` and `NO_COLOR` before any query.
2. Read `TERM`, `COLORTERM`, `TERM_PROGRAM`. Hints only. `TERM=xterm-kitty` suggests kitty. `COLORTERM=truecolor` allows half-block color. Never treat a hint as proof of sixel.
3. Send DA1 (primary device attributes). A reply containing parameter 4 means sixel. Timeout 150ms.
4. Send kitty graphics query (`a=q`). A valid response means kitty. Timeout 150ms. Run this only if DA1 did not already settle the choice and the user did not force one.
5. iTerm2: no reliable query. Use it only when `TERM_PROGRAM=iTerm.app` or the user forced it.
6. tmux: if `TERM` starts with `tmux` or `screen`, do not send sixel or kitty unless the user forced it. tmux needs `allow-passthrough` and still corrupts DCS. Half-block is the default inside a multiplexer. Document that in the help line.

Server-side pre-render (SSH mode only): the process already has the PNG bytes. The adapter may cache, per asset and per `(cols, rows, protocol)`:

- half-block cell grid
- sixel blob at the cell pixel size (use the terminal's reported cell size if XTWINOPS / CSI 16 t answered; else assume 8x16 and accept a blurry first frame)

The downloadable client does this locally. The game server does not grow a graphics stack. Cache cap: 32 images, drop least recently used on resize. A resize invalidates cell-sized blobs and keeps the PNG.

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

History: 100 local lines, not the SSH `mud.history` of 20. Up/Down walk it. Ctrl-R is optional later.

Mouse: SGR mouse when the terminal reports it (Bubble Tea v2 / Ultraviolet). Click an exit, an NPC, a foe, a dialog row, a shop row. Wheel scrolls the log. Mouse off with `--no-mouse` because some SSH clients paste as mouse. Selection for copy uses the terminal's shift-drag. The TUI does not capture shift-drag.

### 5.3 Accessibility

`VEILSPAN_A11Y=1` or `:a11y on`:

- No alternate screen.
- No cursor-addressed full redraw. Print a line when something changes.
- Room enter prints: name, mood, exits, NPC names, one-line description.
- Combat prints: "your turn, N seconds", then the skill names, then results as the `message` field already provides.
- Dialog prints options as `1 text`.
- HP changes print only when the integer HP changes, not every frame.
- Bell optional on your turn (`BEL`), off by default.

This mode can run inside Bubble Tea's inline mode or as a thin printer over the same event stream. Prefer the printer so a screen reader is not fighting the Cursed renderer. Same adapter, different consumer.

### 5.4 Themes and keyboards

Themes: `veil` (default, dim gold on near-black, 16-color fallback uses yellow/white), `mono`, `paper` (dark on light). Lip Gloss v2 has no `AdaptiveColor`. Pick the theme from the detected background. On SSH, read the background the Wish/Bubble Tea query reports. If the query times out, assume dark.

Kitty keyboard protocol: Bubble Tea v2 already decodes it. Bind actions to keys, not to bytes. Do not assume Ctrl-C is the byte `0x03` on every terminal. Still treat raw `0x03` as quit-confirm when the enhanced protocol is off, because classic SSH sends that.

Layout: physical keys. Do not localize n/s/e/w. Command words stay English because the server matches English verbs (`commandprocessor.go`).

## 6. Auth

| Path | Behavior |
| --- | --- |
| Guest, SSH | `ssh -t -p 2222 -l guest host` when `ssh.guest.enabled` and the process allows guests (`docs/ssh-access.md`). The TUI starts already admitted. No token on disk. Guest cap and 30 minute default session stay in `sshgate`. |
| Guest, downloadable | `POST /api/guest`, store the HMAC token in the user config with mode `0600`. Send it as `access_token`. Logout deletes the file. |
| Device code | Unchanged lobby in `sshgate`. Code like `BCDF-GHJK`, URL `/activate`. Classic Auth0 redirects `/activate?code=` into `/play/?activate=`. The TUI does not paint its own activate screen. The lobby paints, then hands off a live session. |
| Linked key | `ssh.keys.enabled` must be on. This tree's default is off (`docs/ssh-access.md`, `SSHKeysConfig.Enabled`). The brief says production key auth is off pending root. Not re-checked against the live host in this spec. When it is on, a linked key skips the lobby and the TUI starts. Web list/revoke stays `GET/DELETE /api/ssh/keys`. `POST /api/ssh/keys` does not link a pasted key. |
| Downloadable account | Auth0 device flow or a pasted bearer token. Store in the config file, mode `0600`. Do not put the token in the process list: prefer `Authorization` on the dial if the websocket stack allows it; the server also accepts `?access_token=` (`auth.go`). If the URL form is required, document that it can land in proxy logs. The server already redacts `access_token` in its HTTP log (`docs/ssh-access.md`). |
| Character select | After connect, if `createCharacter` arrives, open the creator. Else `GET /api/my-characters`. One character: `sc` that name. Several: picker. Remember the last choice locally. Never `sc` while `inCombat`. |
| Session replaced | Close code 4001, or SSH text `Session moved to another client.` Show it and exit. Do not reconnect. |

Guests do not get Party and Friends as account features on the web header. The TUI hides those panels when the account is a guest. If a `party` message arrives anyway, show it.

## 7. Performance, reconnect, quit, tests

### 7.1 Frames

- Cap at 20 fps while idle-animating a timer, 8 fps is enough. The decision timer repaints once a second, plus once when it crosses each whole second. Do not repaint at 60.
- Combat log and room log are a ring of 500 lines. SSH classic history stays 20 for the line editor only.
- The Cursed renderer sends diffs. Do not also send a full screen on a timer.
- Art blobs are not inside the cell diff. Send graphics only when the image id or the cell rect changes.
- Bandwidth budget over SSH: a room change should be one diff under 8 KB plus at most one cached image. If the image is cold, half-block first.

### 7.2 Resize

`tea.WindowSizeMsg` rebuilds the layout. Invalidate art cell caches. If the new size is under 80x24, offer a one-line "window too small, Ctrl-C to quit or resize" and pause input other than quit. Do not fall through to classic mid-session. Classic is chosen at admit.

### 7.3 Reconnect

Downloadable client: on socket close other than 4001, retry with backoff 1s, 2s, 4s, max 15s, for 2 minutes, then show the error. On success, the server sends a fresh `enterRoom` / `characterSelected` as it does for any new socket. A newer login kicks this one with 4001.

SSH: the TCP session is the connection. There is no second reconnect. Idle warning and idle drop stay in `classicLoop` (30 minutes, warn at 1 minute, `docs/ssh-access.md`). The TUI runner must honor the same timers. Guest max session stays. Do not paint over the idle warning with a frame. Print it as a log line and keep the UI.

Resume: there is no client-side combat replay. After a new socket, show whatever the server sends. The v1 gap (combat snapshot on attach) is in section 2.4.

### 7.4 Quit

| Gesture | Result |
| --- | --- |
| Ctrl-C, or `quit` / `exit` after confirm | Close the SSH channel or the websocket. The game sees a disconnect and runs `handleUserQuit` (`game_handlers.go`). |
| Close the terminal | Same. |
| `party leave`, `flee` | Game commands. They do not disconnect. |

Do not send the text `quit` to the command processor. It is not registered, so today it does nothing useful and does not drop the link.

### 7.5 Tests

| Test | What |
| --- | --- |
| Golden | `teatest` at 80x24 and 160x48 for exploration, combat (your turn, victory with `unique`, defeat), dialog, shop, invite. Feed recorded JSON. Compare the view string. |
| Contract | Fixtures checked in from `responses.go` constructors (`NewEnterRoomMessage`, `NewCombatEndMessage`, ...). `pkg/wire` unmarshals them. A Go test in the engine module marshals one of each type. The TUI module reads the same files. |
| Commands | Table test: key to command string, including "no `sc` while inCombat" as a UI gate plus the server string still available. |
| Timer | `deadlineMs` in the past paints `0s` and does not go negative. A deadline 5s ahead paints `5s` then `4s` without a full-screen snapshot change except the timer cells. |
| Fuzz | The line editor is not ours. Fuzz the JSON decoder and the layout splitter (cols, rows, random message order). |
| Resize | 80x24 view, then 40x12, stays on the too-small screen and does not panic. |
| CI | `go test ./pkg/tui/... ./pkg/wire/...` on the engine repo. No SSH to production. No dependency on a live websocket. |

## 8. Milestones

Agent-sized. Hours are focused implementation time, not calendar time. Each one is shippable on the branch and does not deploy.

### M0 — Wire fixtures and empty model (3–5 h)

- `pkg/wire` decodes every type in section 2.3 from fixtures.
- `cmd/veilspan-tui` paints "not connected" at 80x24.
- Core `go test` packages that already pass still pass. No import from `pkg/mudserver` to `pkg/tui`.

Done when: `go test ./pkg/wire` is green and `grep` shows no `bubbletea` import under `pkg/mudserver`, `pkg/server`, `pkg/entities`.

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

### M3 — Downloadable socket (4–6 h)

- Dial `/ws` with a token flag.
- Send `{type:"message", message:<line>}`.
- Handle 4001 by exiting.
- Reconnect backoff for other closes.
- `hello` sent once. Ignore the lack of a reply.

Done when: against a local `go test` server or a recorded httptest upgrade, one `enterRoom` reaches the model. Do not point this at production.

### M4 — Panels (6–8 h)

- Dialog, shop, inventory (effects), quests, party invite accept/decline, character HTTP picker.
- `sc` blocked in the UI while `inCombat`.

Done when: golden tests for dialog and invite exist, and accept sends `party accept`.

### M5 — SSH handoff (6–8 h)

- After `sshgate` admit, size and `TERM` select the runner. No PTY is still rejected, as today.
- `acceptEnv` gains `VEILSPAN_CLASSIC` and `VEILSPAN_A11Y` only. `exec` stays refused.
- Small terminals and `TERM=dumb` stay on `textline`.
- Quit closes the channel.
- Idle and guest max timers still fire.

Done when: `pkg/sshgate` tests cover "small window stays classic", "VEILSPAN_CLASSIC stays classic", "a large PTY runs the runner interface", and "exec is still refused", with a fake runner. No live port.

### M6 — Art and a11y (4–6 h)

- Half-block from a PNG in the fixture.
- Query timeouts fall back.
- A11y printer emits one line per `enterRoom` and does not emit cursor addressing.

Done when: `NO_COLOR=1` golden output has no ANSI color.

### M7 — Gaps worth a small engine change (separate, only if Marcus wants)

- `aggroOnSight` on `RoomNPC`.
- `examineResult`.
- `combatStatus` on attach when already in combat.
- Server `hello` reply.

Each is its own commit on the engine, with a client fixture update. Not part of the TUI branch's first merge.

Acceptance for the whole spec: a player can explore, fight, talk, and quit over SSH in the TUI, and a second binary can do the same over `/ws`, and classic SSH still passes the existing `sshgate` tests.

## 9. Feature translation matrix

Every row is a feature that exists in this tree (engine command, socket message, play client, Flutter client, or creator app). Priority is when the TUI should grow it. Protocol is whether a structured message exists on `99492b3`.

`partial` means some fields exist and the panel still depends on prose or on a second HTTP call.

### 9.1 Rooms, movement, look

| Feature | TUI | Keys | Terminal vs web | Pri | Protocol |
| --- | --- | --- | --- | --- | --- |
| Room enter, description | Room panel + log | `l`, `r` | Text description is the native form. Web art is richer until half-block. | MVP | yes `enterRoom` |
| Exits, hidden exits | Exit row | n/s/e/w and the exit name | Single-key move is better than a compass click. Hidden exits already stripped server-side. | MVP | yes, on `room.exits` |
| Room actions | Hint row | type the action name | Same as classic. | MVP | partial (`room.actions` on the room object; result is `message`) |
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

### 9.2 Combat

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

### 9.3 Talk, quests, economy

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

### 9.4 Character, social, body

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
| Guests | Hide party/friends, show login hint | — | Same as web header. | MVP | no guest flag on the socket (know it from how you logged in) |
| Notifications | Log + modal for invite and level-up | Esc | No toast stack. One modal. | MVP | partial |

### 9.5 Client chrome the web has and the TUI should not copy

| Web / Flutter surface | TUI decision | Pri |
| --- | --- | --- |
| Widget grid, edit layout (`layout/WidgetRegistry.js`) | Fixed breakpoints. No drag grid. | never |
| Nine-slot local hotbar with commands and items (`hudPrefs.js`) | Four server slots. | never as protocol |
| CRT terminal widget | The whole app is the terminal. | never |
| Settings modal (motion, scale) | `:set` and a local file. Reduced motion = no timer blink. | v1 |
| Shortcut sheet | `?` | MVP |
| Account menu, Auth0 buttons | SSH lobby or the downloadable login. Not inside the 80-col frame. | MVP |
| Phone bottom sheet | 80-col panels already are sheets. | — |
| Flutter battle FX | Skip. `fxId` may color the log word. | later |
| Reconnect overlay | Section 7.3 | MVP |

### 9.6 Door, and admin later

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
| World health, drift, validation, search palette | A future `cmd/veilspan-ops` could call the health endpoints. Not the game TUI. | HTTP |
| Live ops (`OpsButtons.svelte`, `RunOp`) | Dangerous (teleport, give, end combat). If ever, a separate binary, creator auth, confirm. | HTTP `RunOp` |
| Users, guests stats | No. | HTTP |
| Audit log | No. | HTTP |

## 10. Features with no structured protocol yet

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
14. Quit / disconnect (no message; the link drops).
15. `aggroOnSight` on the room NPC (trait exists, payload does not).
16. Combat snapshot for a client that connects mid-fight.
17. Hotbar item binds and slots 5–9 (web local storage only).
18. Settings, themes, layout.
19. Guest-vs-account as a socket field.
20. Admin, creator, world health, drift, search, loot editor, spawners, ops.
21. `combatLeave` (web handler only; server does not send it).
22. Door key events (`door_key`) — Door's, not this client's.

Section 2.4 is the subset worth adding to the socket. This list is the full "do not scrape, do not fake" set.

## 11. Risks and open questions

### 11.1 Risks

| Risk | Mitigation |
| --- | --- |
| TUI code imports `game` "just for one field" and core stops being framework-neutral | `pkg/tui` is banned from that import. CI grep in M0. |
| SSH TUI and classic editor both read the channel | Choose once at admit. Do not switch mid-session. `exec` stays refused. The classic opt-out is an allowlisted env var, not a remote command. |
| Bubble Tea capability queries leak escape codes into a classic session | Only start the program after the choice. v2.0.10 skips queries when input is disabled. Keep input enabled only on the TUI path. |
| Graphics DCS corrupts the next frame inside tmux | Default half-block when `TERM` is tmux/screen. |
| 5s timer feels late because of SSH lag | Display server `deadlineMs`, not a local 5s. |
| Player hits `q` and expects quit, gets quests | Help line and `?`. Quit is Ctrl-C and the words `quit`/`exit` with confirm. |
| `sc` during a fight | UI lock plus the server sentence from `99492b3`. |
| Door and this TUI diverge on ANSI safety | Copy the sanitize rules. Do not share the view. |
| Fixture drift from `responses.go` | Contract test marshals engine structs. |
| Wish rewrite of `sshgate` | Do not. The lobby is audited. The TUI is a runner behind it. |

### 11.2 Open questions for Marcus

1. Is the default on a large PTY the TUI, with `SendEnv VEILSPAN_CLASSIC=1` as the escape, or the reverse until golden tests exist? `exec` cannot carry this. It is refused in `session.go`.
2. May M7 add `aggroOnSight` on `RoomNPC` and an `examineResult` message on `engine-june`, or do those wait?
3. Downloadable auth: is a pasted Auth0 bearer acceptable for the first binary, or do you want a device-code flow that does not reuse the SSH lobby?
4. Should screen-reader mode be selected from the SSH lobby (a third question) or only from `VEILSPAN_A11Y`?
5. Art: is half-block enough for the first public SSH default, with kitty/sixel behind `:art`?
6. Confirm production still has `ssh.keys.enabled` false. This spec did not log into the host. The repo default is false.
7. Rigger and Ward are in `pkg/classkit/testdata`. Is the live pack's `GET /api/classes` the same five? The TUI will trust the HTTP list either way.
8. Is a separate ops TUI in scope this quarter, or is the note in section 9.6 enough?

### 11.3 Not verified

- The live process on veilspan.com (ports, key auth, which class pack is loaded). Docs and this tree only.
- The Doorkeeper worktree's private UI beyond `pkg/presentation/doorview` and `public/door` in this commit. `~/dev/talesmud-door-mud` exists and was not used as a second source of truth.
- Bubble Tea v2.1.0 (2026-10-08) was checked only from its release notes, not built against.
- Exact HTTP path that serves `room.meta.background` bytes.
- Whether Flutter on disk outside this repo handles `party_invite`. The `game_view_state.dart` switch in `~/dev/veilspan-client` does not.

## 12. Implementation note for docs

When a later change implements this, update `PROJECT.md` (player-facing SSH/TUI behavior), `ARCHITECTURE.md` (new `cmd/veilspan-tui` and `pkg/tui` boundary), and `FEATURES.md` (new socket fields, if any). This spec does not change those files because no behavior changed.
