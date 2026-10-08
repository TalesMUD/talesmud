# Boss signatures v1 — one hook mechanic per zone boss (crown 2026-10-08, daily director)

Crown = SYSTEM-into-CONTENT: last night shipped enemy combat hooks v2 (onLowHealth + `tales.combat.summon`, cap 3), hook combat cards (`style:"combatEvent"`), and aggro-on-sight (bosses passive, no aggressive enemies below L10 — content 65bba1c). Today the 11 other zone bosses get a signature beat so each boss fight feels different, not just bigger numbers. Engine stays framework-neutral; Veilspan specifics live only in content Lua/YAML.

## Rules
- Engine worktree `~/dev/talesmud-hooks2` (branch `hooks-v2`, `git pull --ff-only` first), push `git push origin hooks-v2:engine-june` ff only, never force. Do NOT commit from `~/dev/talesmud-june` (portrait dirt).
- Content `~/dev/talesmud-rpg-1` main; stage only files you change (untracked `.director` dirt stays).
- Every commit contains `[grokbot]`. No Cursor cloud. No public master merge. Never touch Door `:8020` / door-* processes. No firewall/port changes.
- If grok hits its usage limit, stop cleanly and print `USAGE_EXHAUSTED` (never half-deployed).
- Bosses stay passive: do NOT set `aggroOnSight` on any boss. `tests/test_enemy_aggro_rule.py` must stay green.

## Step 0 — pre-flight (residual from last night)
`go test -count=1 -timeout 20m ./pkg/...` on engine-june tip (full suite was not re-run after 72bcd9c). Red → fix first, own commit.

## Step 1 — signatures (content)
Bosses: ENM0011 Burrow Brute L4, ENM0017 Mire Hag L14, ENM0024 Bandit Captain Rask L19, ENM0028 Lord Edric Vayne L20, ENM0036 Ironjaw L30, ENM0042 Blighted Hart, ENM0048 Dynamo Core, ENM0054 Thornfield Colossus, ENM0060 Slagfiend, ENM0066 Ley Wraith, ENM0072 Corrupted Dynamo (all L30). ENM0009 Hollow Knight already has SCR0207 — leave it.
1. Give each boss exactly ONE signature, using only existing hook APIs (`onAggro` / `onLowHealth` / `onDeath` / `onFlee`, `tales.combat.summon`, `tales.combat.healNpc`, `tales.combat.applyEffect`, `tales.game.msgToRoom`). Mix the kinds across the roster: roughly a third summon adds (existing low-level templates from that zone; never invent templates unless none fit), a third a one-time self-heal or ward/shield, a third a debuff/DoT or enrage effect on the party. Fit each to the boss's zone lore (read the zone LORE.md / boss yaml); Mire Hag should tie in to the existing fetish/antidote work if it fits.
2. Align with A9 phases: set `lowHealthThreshold: 0.33` where the beat should land with the Last Stand banner; otherwise leave the 0.30 default. Don't duplicate the existing round-16/30% enrage.
3. One room line per signature, Veilspan clipped voice: one short sentence, concrete nouns, what the boss DOES, no metaphor stacks, no AI tells (hush, resonate, tapestry, ancient song…). Example of the bar: "The Knight's runes flare. Something answers from the drains."
4. Next free SCR ids; validator + import dry-run clean (known warnings only R1910/R1934 + elite-difficulty).

## Step 2 — balance
Run the existing sim harness (simutil / balance tooling used in A4/A5/HK-S0) for each touched boss before vs after: solo appropriate-gear and the party band used before. Signatures may make fights harder but must not push a same-level geared party below the existing target band; tune summon count, heal %, or threshold until in band. Record a before/after win-rate table for Quest Master.

## Step 3 — deploy + live proof
Normal VPS path (content ff, `deploy.sh --dry-run --skip-export --no-assets`, import; engine rebuild only if engine changed, keep `bin/tales.prev-<oldsha>` + sha256s, SIGTERM talesmud MainPID only). Smoke `/play/` 200 (`?v=hooks2` unless client touched), `POST /api/guest` 200, `/api/server-info` 200.
Live proof using the A9 prepared-disposable-guest method (`tools/capture_boss_phases.cjs` pattern; fresh guests only, restore their fields after): fight the Hollow Knight AND one new summon-type boss to low health and capture the raw websocket frames showing `style:"combatEvent"`, `hook:"onLowHealth"`, `source`, plus the summoned adds arriving in `combatStatus.combatants`. Screenshot the BattleStage gold chip at 1920×1080 to `.director/ux-audit/after/boss-sig-chip-1920x1080.png` (last night's chip was never seen in a browser). Close all fights/sockets.

## Report
Append "Boss signatures v1" to `.director/CORE-FOCUS-PROGRESS.md` (SHAs, per-boss table: hook, mechanic, threshold, line, SCR id, sim before/after; tests; deploy sha256s/PIDs; smoke frames; residuals), commit this brief + report `[grokbot]`, push, then print `BOSSSIG DONE` or `BOSSSIG BLOCKED: <reason>`.

## Out of scope
New zones/rooms, new art, aggro radius/leash, Flutter, Door, DSA armor-weight, engine API changes beyond a bug fix the work exposes.
