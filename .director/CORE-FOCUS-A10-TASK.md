# CORE FOCUS — A10 Focus target + threat re-warn (combat)

Director lock: 2026-10-01 ~10:10 Europe/Berlin.
Crown: live demo / local `:8010`. Repo `~/dev/talesmud-june`, branch `engine-june`.
Worker: Codex on clawdbot tmux `codex-a10` (Grok weekly limit — use Codex per standing rule).
Prior crown B14 Settings ACCEPTED; overnight UX EXTRA (map travel / quest areas / equip / buff chips / turn-ins) ACCEPTED live `?v=buffport1` tip `2c8d37e`.

## Rules
- Commits MUST include `[grokbot]` in the message.
- NEVER Cursor cloud. Never touch `:8020` / tales-door / the door host.
- Deploy via normal VPS path only (ff `engine-june`, copy play client into `pkg/webuiplay/dist`, `go build -o bin/tales.next ./cmd/tales`, rename over `bin/tales`, SIGTERM only talesmud MainPID; Restart=always). No cowboy prod source edits.
- Keep OSS engine product-neutral (no private world pack / product hardcoding).
- Append results to `.director/CORE-FOCUS-PROGRESS.md` under an A10 section.
- Cache-bust client to `?v=a10focus` (and matching stylesheet/worker queries if touched).

## Why
A2 threat colors warn on engage, but switching targets mid-fight does not re-warn (A2 residual). Multi-enemy fights (swarms, boss + adds) need a clear focus target so attack/skills and threat read stay honest after overnight portrait buff chips.

## Done-when
1. **Focus target**: In an active combat with 2+ hostiles, the local player can select a focus target (click enemy nameplate / BattleStage portrait). Focused enemy has a clear gold highlight distinct from buff chips. Selection persists until death, flee, combat end, or another select.
2. **Default aim**: Basic attack and targeted hostile skills use the focus target when set and still valid in the instance; otherwise fall back to current engine default. Room `attack <name>` still works and should set focus to that NPC when join/engage succeeds.
3. **Threat re-warn on switch**: When focus changes to a different living hostile, re-evaluate level-gap threat tier (A2 colors). If the new target is orange/red/skull (or the existing "much stronger" band), show the same class of warning again (room/BattleStage toast or banner — reuse A2 copy/colors; do not spam on every click of the same target).
4. **Single-hostile fights**: No new UI clutter; focus can auto-set to the only hostile. Reduced motion honored for any new highlight animation.
5. **Tests**: Node (and Go if engine aim/threat touched) covering focus select, invalidation on death, re-warn on switch to skull/red, no re-warn on re-click same target. `npm run build` green. Existing combat/layout tests stay green.
6. **Docs**: FEATURES.md / PROJECT.md note Focus target + in-combat threat re-warn (A10).
7. **Ship**: Commit + push `engine-june`, deploy the live demo `:8010` only, live smoke `/` `/play/` `POST /api/guest` `/api/server-info` 200 with `?v=a10focus`. Screenshot `.director/ux-audit/after/a10-focus-1920x1080.png` (two-hostile or controlled payload OK if natural multi-hostile is awkward). Door untouched.
8. Progress entry with SHA, levers, residuals.

## Out of scope
New zones, party auto-pull, need/greed loot, Flutter (web first; note Flutter parity residual), DSA armor-weight, audio Settings, Door, Google OAuth.

## Report back
SHA(s), live `?v=`, how focus is stored/sent, re-warn copy, smoke notes, residuals.
