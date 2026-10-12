# SYSTEM crown 2026-10-12 — Structured examine cards v1

You are on clawdbot in `~/dev/talesmud-examine` branch `feat/examine-cards` (tracks `origin/engine-june` tip; live VPS is `5402f01` / `?v=pickup2`). Implement, test, commit `[grokbot]`, push ff to `engine-june`, deploy VPS the usual way. No Cursor cloud. No public master merge. Do not touch Door :8020.

## Why
Overnight TONIGHT-B shipped additive `examineResult` (stats, quality, value, description, explicit `effects[]` via `items.FormatEffect`) alongside unchanged examine prose. The play client still only builds the inspect overlay by parsing prose (`parseExamineOverlay.js`). Players do not reliably see script/magic effect lines from the structured payload. Close the loop.

## Done-when
1. **Client prefers `examineResult`**: when a websocket `examineResult` arrives for an item inspect, RoomTextOverlay (or the existing examine card) shows title, quality/rarity chrome, stats, value, description/detail, and each `effects[]` line as explicit effect text — not hidden in flavour. Prose `message`/`room` lines still print as today.
2. **Fallback**: if only prose arrives (old server / NPC examine), keep current `parseExamineText` path. NPC inspect stays prose-only (no fake item card).
3. **Cache-bust**: rebuild play client `?v=examine1` (bundle + linked CSS if touched).
4. **Tests**: mud-client unit tests for examineResult → card model (effects present, empty omit, quality chrome). Go suite still green if engine touched (prefer client-only).
5. **Docs**: FEATURES / structured-messages note that the play client consumes `examineResult`.
6. **Ship**: push `feat/examine-cards:engine-june` ff-only, deploy VPS binary only if Go changed else client dist sync + restart if required by embed path; smoke guest `examine` on an item with an effect (or unique/quest) and confirm overlay shows the effect line from structured data. Report SHA + `?v=` + smoke.

## Out of scope
Flutter (local pickup commits stay unpushed), new content, admin UI, Door, public master, inventing new item effects.

## Report
Print exactly `EXAMINE1 DONE <sha> ?v=examine1` or `EXAMINE1 BLOCKED: <reason>`.
