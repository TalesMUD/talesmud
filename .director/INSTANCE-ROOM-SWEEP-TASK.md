# Instance room copies leak across restart — startup/shutdown sweep + relocation (2026-10-08)

Worktree `~/dev/talesmud-hooks2` (branch `hooks-v2` tracking `origin/engine-june`, tip `9d76dde`). Push `git push origin hooks-v2:engine-june`, fast-forward only (rebase on reject, never force). `[grokbot]` in every commit. Never touch Door :8020 / door-* processes. No public master merge. No firewall/port changes. Finish by ~18:30 Berlin.

## Problem (already investigated)
`pkg/instances/manager.go` `Enter`/`Generate` persist room copies (`CloneID` = `<template>~<instanceID>`) into the sqlite rooms table via `roomsSvc.Import`. Cleanup (`NoteLeave`, `DestroyCharacterInstance`, `Expire`, `destroyLocked`) depends on the in-memory `Manager` maps. After a process restart the maps are empty: leftover `~` rows stay in the rooms table forever (only a content `--import` clears them, because it drops the table), and `session_start.go` uses `IsClone()` (map lookup) so a character saved in a `~` room is NOT recognised as being in a clone and logs into an orphan copy. NPC copies use `<npcID>~<cloneRoom>` (room_instances.go cloneNPCs) — check whether those are persisted anywhere (npcs table / NPC manager) and include them in the sweep if so.

## Do (generic, framework-neutral — no Veilspan ids in core)
1. **Startup sweep**, run once during game init before the websocket accepts players: delete every room whose id contains the instance marker (`~`; centralise the marker/parse in `pkg/instances`, e.g. `IsCloneID`, reuse `TemplateIDFromClone`), plus any other instance-generated persisted rows tied to them (NPC copies, spawners, ground items, etc. — only if they are actually stored). Log a summary (count deleted).
2. **Relocation**: every character whose saved `currentRoomID` is a `~` copy, or a room that no longer exists, is moved:
   - first choice: the instance's entrance room = the hub/"way in" room for that template group (the non-instance room whose exit leads into the template room's instance graph — `CollectGraph`/hub logic already knows hubs; derive it from the template id, walking the template graph back to a non-instance neighbour),
   - else a configured return/exit room for the group if the engine has such a concept (procedural specs?),
   - else the world start room (`service.ResolveStartRoomID`), else the existing `fallbackRoom`.
   Clear stale combat flags as session_start already does. Persist the update. On that character's next login send one short line: `You find yourself back at <room name>.` (store a pending notice on the character or in memory keyed by character id — whatever is simplest and survives until login; if in-memory, note that the sweep runs at startup so memory is fine).
   Also harden `session_start.go`: treat a `~` id (not just `IsClone`) as "in clone", so a copy that somehow survives still relocates.
3. **Graceful shutdown**: if there is a clean shutdown hook (SIGTERM path), run the same room sweep (destroy live instances) there too — only if cheap and safe; characters still get relocated on next startup by step 2.
4. **Tests**: leftover copies removed; non-instance rooms (including authored `instance`-tagged templates without `~`) untouched; characters in a copy relocated to the hub/entrance; character in a missing room relocated to start room; login line delivered once; sweep idempotent (second run deletes 0, moves 0); NPC copy rows (if persisted) removed. `go test -count=1 -timeout 20m ./pkg/...` green (needs gitignored `pkg/webuiplay/dist` present, as before), plus `-race` on touched packages.

## Deploy (same path as previous ships — see `.director/CORE-FOCUS-PROGRESS.md` "Aggro on sight")
VPS `ssh veilspan-vps`, `/home/atla/dev/talesmud`, engine ff pull (leave portrait dirt unstaged), build `bin/tales.next` (Go 1.24.12, CGO_ENABLED=1, GOAMD64=v1), keep the running binary as `bin/tales.prev-16ed721` (sha256 590300ab…), record new sha256, SIGTERM talesmud MainPID only (Restart=always). No content import needed.

## Live check
`/play/` 200, `POST /api/guest` 200, `/api/server-info` 200. If practical: throwaway guest → enter a cellar instance (e.g. stand in R0211 Oldtown sewer grate and go `down` into R0241, or set `currentRoomID` to a hub like R0211 while disconnected first, then walk in), confirm a `~` room row exists in sqlite, disconnect abruptly or restart talesmud (SIGTERM MainPID only) while the guest is inside, confirm after restart: zero `~` rows, guest's `currentRoomID` = the hub/entrance, and on reconnect the line "You find yourself back at …" arrives once. Restore the guest to R0001 afterwards. Read-only sqlite checks otherwise.

## Report
Append "Instance room sweep" to `.director/CORE-FOCUS-PROGRESS.md` (SHAs, sweep rules, relocation order, tests, deploy sha256s/PIDs, live check, residuals), commit, push, then print `SWEEP DONE` or `SWEEP BLOCKED: <reason>`.
