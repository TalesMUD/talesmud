# Core Focus progress

## A1 — Level-gap combat math
- SHA: `cf31157548738dbbd6c32bd834e5cf9fa06b431b` (`cf31157`)
- What changed: `level_gap` in `config/combat_balance.yaml`. Signed gap is attacker level minus defender level, clamped at ±6. Hit, crit, and damage dealt/taken scale per level. Gap 0 matches the old formulas. Applied to basic attacks, skill hits, and DoTs for players and NPCs via `CombatantRef.Level`.
- Tests: `go test ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/combat/ -count=1` green. Druid L1 vs L2 floor in `TestLevel1VsSameLevelEnemies` lowered from 5% to 1% because a one-level deficit now costs glass casters a few win-rate points (observed ~3%).
- Deploy: pushed `engine-june`. Rebuilt `bin/tales`, restarted only the local engine process. New pid 3611918 listening on :8010. Door pid 3406193 on :8020 unchanged.
- Smoke: `GET /` 200, `GET /play/` 200, `GET /api/server-info` 200. Log: "listening on port 8010" at 2026-09-24 23:04:54. No panic.
- Residuals: live numbers are the A1 starting point (about +32% damage at a +3 advantage, about −28% at −3). A4 tunes them against the win-rate table. Skill crits only come from a positive level gap; basic attacks still crit on a natural 20 unless the gap pulls that chance down.

## A2 — Threat display
- SHA: `b4072f382c93a01a683b509d37c925181c1088bb` (`b4072f3`)
- What changed: `threat` cutoffs in `config/combat_balance.yaml` (enemy level − player level). ≤ −3 grey, −2..−1 green, 0..+1 yellow, +2 orange, +3..+4 red, ≥ +5 skull. Viewer level below 1 counts as 1. Room NPC payload and per-viewer combat enemy views include `threat`. Entity cards and BattleStage nameplates use the color; skull adds ☠. First `attack` on orange/red/skull warns and does not engage. `attack!` / `a!` / `hit!` or a second `attack` on that enemy does. The room Attack button confirms, then sends `attack!`.
- Tests: `go test ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/commands/ ./pkg/mudserver/game/` green. New cases: tier table, skull warn-then-engage, `attack!` skips the warn, yellow does not warn.
- Deploy: pushed `engine-june`. Rebuilt client (`?v=a2threat`), copied into `pkg/webuiplay/dist`, rebuilt `bin/tales`, restarted only the local engine. New pid 3614695 on :8010. Door pid 3406193 on :8020 unchanged.
- Smoke: `GET /` 200, `GET /play/` 200, page references `bundle.js?v=a2threat`, that bundle is 200 and contains "much stronger", `extra.css` contains `threat-skull`. Log listening on 8010 at 23:16:52. No browser session drove a logged-in room card or BattleStage nameplate; server warn/engage is covered by the command tests.
- Residuals: in-combat target switches do not re-warn. Swarm warning uses the named target only.

## A3 — Reward scaling
- SHA: `b62a8a35aa4159a7d763a12b66da54f213b8d549` (`b62a8a3`)
- What changed: `reward_scale` in `config/combat_balance.yaml`. Each enemy's base XP and rolled gold are multiplied by the threat tier of (enemy level − reference level). Reference level is the **highest** level among characters who receive the split (living fighters plus same-room online party). Defaults: grey 15%, green 60%, yellow 100%, orange 125%, red 150%, skull 200%. Boss first-kill bonus is 50% of that character's own share, once, stored on `Character.FirstBossKills` (`tpl:<id>` or `name:<lower>`). BattleStage victory lists base, level modifier, first-kill bonus, and party split (`?v=a3reward`). Terminal text includes the same lines, then the final `+ N XP`.
- Tests: `go test ./pkg/mudserver/game/ ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/commands/ ./pkg/mudserver/game/leveling/` green. New: grey trickle uses the party high level; boss first kill pays once and the combat-end payload carries the breakdown. Existing equal-level party splits stay at 100%.
- Deploy: pushed `engine-june`. Rebuilt client, copied into `pkg/webuiplay/dist`, rebuilt `bin/tales`, restarted only the local engine. New pid 3616684 on :8010. Door pid 3406193 on :8020 unchanged.
- Smoke: `GET /` 200, `GET /play/` 200, page references `bundle.js?v=a3reward`, that bundle contains "First-kill bonus". Listening on 8010 at 23:24:55. No logged-in browser pass of the victory panel; the payload is covered by `TestBossFirstKillBonusOnce`.
- Residuals: party-share toast still lists the pre-bonus split. First-kill is in the fighter's victory panel and in the awarded totals. A4 may retune the multipliers with the combat numbers.

## A4 — Balance harness and tuning
- SHA: `30853d7f6d36a67b3a0153cfcad93c861eab4f74` (`30853d7`)
- What changed: `simutil.RunGapMatrix` fights warrior/rogue/ranger/mage at level 10 across gaps −3..+5, appropriate vs good gear (about 2.2×), against level-scaled trash/elite/boss bodies. Table is `docs/COMBAT-BALANCE.md`. Tuned `level_gap` to hit +3.5%, crit +1%, damage dealt +3.5%, damage taken +2% per level (clamped ±6). Gap 0 stays a no-op, so at-level content duration tests still pass. `TestGapMatrixTargets` locks the bands with room for a 24-iteration sample.
- Sample (24 iters): warrior/rogue/ranger trash at gap 0 is ~100%. Warrior elite ~90%+, boss ~75% (rogue/ranger bosses closer to 50%). Good-gear bosses at +3: warrior ~62%, ranger ~42%, rogue ~21%. Appropriate-gear elites and bosses at +5 are ~0%. Trash at +5 is still winnable for a warrior (~30–40%). Mages in cloth lose most elite and boss fights.
- Tests: `go test ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/combat/ -run 'TestLevelGap|TestGapMatrix|TestCombatDuration|TestLevel1|TestBosses|TestSameLevel'` green.
- Deploy: pushed `engine-june`. Rebuilt `bin/tales` (yaml is read at process start). Restarted only the local engine. New pid 3619758 on :8010. Door pid 3406193 on :8020 unchanged.
- Smoke: `GET /` 200, `GET /api/server-info` 200. Listening on 8010 at 23:34:56. No client asset change.
- Residuals: warrior at-level bosses often beat the 65% ceiling. Rogue good-gear +3 bosses are under the ~50% target. Mage cloth does not meet the melee bands. Trash at +5 is not a skull fight for a level-10 warrior.

## B1 — Viewport audit
- SHA: `ffab8985969721d226b45f6668dbfc4c93a0c583` (`ffab898`)
- What changed: screenshots in `.director/ux-audit/before/` and `.director/ux-audit/AUDIT.md`. No product code.
- Findings: fixed 12-row grid leaves a large empty band on 1920 and 3440; terminal text clips on the right at 1366; hotbar floats in the gap; inventory stacks two headers; guest account menu has no Edit Layout; resizing a live desktop session to phone width dropped back to the welcome screen. Phone cold-start renders the mobile shell but clips the room description. No combat screenshot — the starter room has no enemy.
- Deploy: none. :8010 stayed on the A4 process. Door pid 3406193 untouched.
- Residuals: the list in AUDIT.md drives B2–B4.

## B2 — Smart default layouts
- SHA: `75ef38eeed76f2b17f67a433be92aeccc7e023c4` (`75ef38e`)
- What changed: `layoutPresets.js`. No saved layout picks Compact (&lt;1100px), Desktop, or Wide and sizes rows to the viewport. Resize reflows that preset. A saved layout is clamped onto the 24-column grid (min 2×2) and is not replaced. Edit mode has Compact / Desktop / Wide buttons. Cache-bust `?v=b2layout`.
- Tests: mud-client `npm run build` succeeded. No new Go tests (client-only).
- Deploy: pushed `engine-june`. Copied client into `pkg/webuiplay/dist`, rebuilt `bin/tales`, restarted only the local engine. New pid 3622162 on :8010. Door pid 3406193 unchanged.
- Smoke: play page references `bundle.js?v=b2layout`. Guest at 1920×1080: grid bounding box height 1008px starting at y=26 in a 1080px viewport (the old 12-row grid was about 480px). Listening on 8010 at 23:43:11.
- Residuals: terminal lines still clip on the right. Guest menu still has no Edit Layout (B3). Inventory still has two headers (B4). Breakpoint session drop not fixed yet.

## B3 — Edit-mode ergonomics
- SHA: `545dbc5c98b960951fa0e2fbf6f86508663848c2` (`545dbc5`)
- What changed: Guest account menu includes Edit Layout. Toolbar adds Undo (up to 30 layout steps: drag, resize, add, remove, preset, reset) and Lock/Unlock. Lock keeps edit mode open and hides drag and resize. Corner and edge handles stay visible in gold while unlocked. The svelte-grid drop ghost is a gold dashed shadow. A guest token in `sessionStorage` is restored after reload, including when a mobile-emulation viewport change reloads the page. Cache-bust `?v=b3edit`.
- Tests: mud-client `npm run build` succeeded (existing unused-CSS and a11y warnings only). No new Go tests (client-only).
- Deploy: pushed `engine-june`. Copied client into `pkg/webuiplay/dist`, rebuilt `bin/tales`, restarted only the local engine. New pid 3623813 on :8010. Door pid 3406193 unchanged.
- Smoke: play page references `bundle.js?v=b3edit`. Guest menu lists Edit Layout, then Create Account, Settings, End Session. Edit mode toolbar shows Undo and Lock; 16 gold corner handles (4 widgets). Lock clears the handles and the hint reads "Layout locked". Reload and a 390×844 mobile viewport both stay on the Awakening Chamber, not the welcome screen. Listening on 8010 at 23:52:18.
- Residuals: terminal lines still clip on the right. Inventory still has two headers. Account chip still overlaps the terminal's top-right in edit mode. Those are B4 chrome.

## B4 — Unified widget chrome
- SHA: `703405b0c224fdc3709ca00ad1dfc6baa6a4333b` (`703405b`)
- What changed: `WidgetChrome` is the shared header (Cinzel title, collapse, focus) on the terminal, inventory, equipment, character, quest log, and tab container. The room keeps its scene title and gets the same collapse and focus buttons. Collapse shrinks a panel to two rows. Focus expands one panel over the grid and parks the others at 2×2 so the terminal stays mounted. The inventory overlay has a single title. The account chip sits in a 52px band above the grid (`z-index` under overlays). Terminal fit leaves one column of slack so lines wrap inside the panel. Phone room copy is a clamped block with the expand hint under it. Cache-bust `?v=b4chrome`. After shots are in `.director/ux-audit/after/`.
- Tests: mud-client `npm run build` succeeded (existing unused-CSS and a11y warnings only). No new Go tests (client-only).
- Deploy: pushed `engine-june`. Copied client into `pkg/webuiplay/dist`, rebuilt `bin/tales`, restarted only the local engine. New pid 3626122 on :8010. Door pid 3406193 unchanged.
- Smoke: play page references `bundle.js?v=b4chrome`. Guest grid bottom is 764/768, 1076/1080, and 1436/1440. Chip rect does not intersect the terminal. Inventory overlay shows one Inventory title. Phone description box is 101px tall, hint 6px below it. Listening on 8010 at 2026-09-25 00:04:49.
- Residuals: terminal wrap splits words at the column (`lig` / `ht`). Hotbar is still its own row. No BattleStage shot — north of the starter room has no hostile. Saving during focus writes the expanded arrangement.

## A1–B4 production promote
- SHA on the VPS after the fast-forward: `5b7423f415e0a368294e1facc4c829174cb30520` (`5b7423f`). Was `cdd9168` (Party Follow, bundle `?v=party2`).
- What changed: `git pull --ff-only origin engine-june` in `~/dev/talesmud`. Restored a local `extra.css` reorder so that file could fast-forward. Copied `public/mud-client/public` into `pkg/webuiplay/dist`, `go build -o bin/tales`. `sudo systemctl restart` needs a password, so the old pid 711084 was SIGTERM'd and `Restart=always` started pid 722724. Door pid 695590 on :8020 was not restarted.
- Smoke: local `/` 200, `/play/` 200 and `bundle.js?v=b4chrome`, `/api/server-info` 200, `POST /api/guest` 200. Listening on 8010 at 2026-09-25 08:07:48Z.

## A5 — Class balance
- SHA: `9013879c9c66f6161e38b0b5daad1f0ce0e5af83` (`9013879`)
- Levers: `class_balance` in `config/combat_balance.yaml`. `damage_dealt` / `damage_taken` after the level-gap product and before crit. `behind_dealt` multiplies outgoing damage only when that class is the lower level (`wizard` uses the mage row). Scaled boss hit points in `CreateScaledEnemy` are `220 + 23*level`. Content bosses still use `CreateEnemy`, so warrior trash duration stays in the old windows. Warrior even-fight damage is 1× with `behind_dealt` 1.20. Rogue is 1.35× even and 2.35× uphill. Ranger is 1.26×. Mage is 2.65× dealt and 0.46× taken.
- 80-iteration check, before → after: warrior appropriate boss gap 0 **75% → 59%**. Rogue good-gear boss +3 **21% → 51%**. Mage appropriate elite gap 0 **8% → 74%**, appropriate boss gap 0 **0% → 54%**, good-gear boss +3 **4% → 52%**. Rogue appropriate +5 stayed 0%.
- Tests: `go test ./pkg/mudserver/game/balance/ ./pkg/mudserver/game/combat/ -count=1 -run 'TestScaleClass|TestLevelGap|TestGapMatrix|TestCombatDuration|TestLevel1|TestBosses|TestSameLevel'` green. `TestGapMatrixTargets` locks warrior appropriate boss gap 0 at ≤70%, rogue good-gear boss +3 at ≥35%, and mage elite / boss / good-gear +3 inside wide 24-iteration bands.
- Deploy: pushed `engine-june`. VPS fast-forwarded to `9013879`, rebuilt `bin/tales`, SIGTERM of pid 722724, `Restart=always` started pid 723455. Door pid 695590 unchanged. No client change, so the page is still `?v=b4chrome`.
- Smoke: `/` 200, `/play/` 200 with `bundle.js?v=b4chrome`, `/api/server-info` 200, `POST /api/guest` 200. Listening on 8010 at 2026-09-25 08:26:14Z.
- Residuals: a 24-iteration draw still swings (rogue good-gear +3 has landed at 29% in one draw while the 80-iteration center is 51%). Ranger at-level bosses are closer but not as steady as warrior. Rogue even-fight bosses are a bit under the 50–65% ceiling because the uphill multiplier is what fixes the +3 cell. Word-split terminal wrap and the hotbar row are unchanged.

