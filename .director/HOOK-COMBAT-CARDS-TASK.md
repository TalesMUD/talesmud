# Slice 3 (engine + web client) — tag enemy-hook combat lines + summon roster refresh — night 2026-10-07

Same rules/hard stop as ENEMY-COMBAT-HOOKS-v2-TASK.md (stop new work 00:45 Berlin or ~95% /usage; never half-deployed; Door :8020 untouched; no force push; no public master; commits `[grokbot]`). Worktree `~/dev/talesmud-hooks2`, push `hooks-v2:engine-june` ff only.

## Wire contract (FIXED — the Flutter worker is building against this in parallel; do not change names)
`MessageResponse` gains three optional JSON fields, all `omitempty`:
- `style` — `"combatEvent"` for lines emitted by enemy combat hook scripts.
- `hook` — the hook name: `onAggro` / `onLowHealth` / `onDeath` / `onFlee`.
- `source` — the display name of the NPC whose hook ran (e.g. `The Hollow Knight`).
`type` stays `"message"` and `username` stays `"SYSTEM"`, so older clients still show plain text. Nothing else changes for non-hook messages (fields omitted).

## Engine
1. While an enemy hook script runs, room messages it sends via `tales.game.msgToRoom` (and `msgToRoomExcept` if trivial) carry `style/hook/source` as above. Generic: no Veilspan names in Go. Implement via the hook execution context (e.g. per-run context on the runner / LState), not a global that can leak across concurrent scripts. Non-hook scripts (room/quest/item) are unchanged.
2. Summon roster refresh (v2 residual): after a hook flush that summoned adds, send the fight's combat status/roster update right away so clients see the adds without waiting for the next action. Reuse the existing combatStatus/roster message path.
3. Tests: hook msgToRoom carries the three fields; non-hook msgToRoom does not; JSON omits them when empty; summon triggers an immediate roster update.

## Web client (public/mud-client) — only if small
Render messages with `style === "combatEvent"` as a distinct compact system card in the log/BattleStage feed (e.g. Veilspan gold-edged chip with a small icon per hook: drum/aggro, low-health/summon, death, flee; `source` as a small label). Plain text fallback unchanged for everything else. node tests if there's a message-rendering test file; `npm run build` green; embed into `pkg/webuiplay/dist`; cache-bust to `?v=hooks2` (bundle + matching css/worker queries as previous ships did).

## Deploy + smoke
Same VPS path as v2 (prev binary kept as `bin/tales.prev-<oldsha>`, sha256s, SIGTERM talesmud MainPID only; no content import needed unless content changed). Smoke: `/play/` 200 with `?v=hooks2`, `POST /api/guest` 200, `/api/server-info` 200; if practical, a guest fight against the Z06 War-Chanter (R0628) and confirm the drum line arrives over the websocket with `style:"combatEvent"`, `hook:"onAggro"`, `source` set (capture the raw frame). Close fights/sockets; if one sticks, restart talesmud MainPID only.

## Report
Append "Hook combat cards (engine/web)" to `.director/CORE-FOCUS-PROGRESS.md`, commit, push, then print `CARDS DONE` or `CARDS BLOCKED: <reason>`.
