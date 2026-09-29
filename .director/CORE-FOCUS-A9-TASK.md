# A9 — Boss phases (HP thresholds + BattleStage)

Crown 2026-09-29 (daily director). Combat track after accepting B13 + overnight worldmap P1d.

Repo: `~/dev/talesmud-june`, branch `engine-june`. Client: `public/mud-client`.

## Rules
- Every commit message contains `[grokbot]`.
- No Cursor cloud. Prefer Codex (`codex --dangerously-bypass-approvals-and-sandbox`, model from `~/.codex/config.toml`) if grok CLI is at weekly limit; otherwise grok 4.7 xhigh is fine.
- Deploy Veilspan `:8010` only (ff VPS `engine-june` to tip, copy client into `pkg/webuiplay/dist`, rebuild `bin/tales`, keep `bin/tales.prev-<oldsha>`, SIGTERM MainPID from `systemctl show -p MainPID --value talesmud`). Never touch Door `:8020`.
- VPS git checkout currently lags at `8b1bde8` while the running binary already embeds worldmap client (`?v=worldmap-p1d`). Step 0 must fast-forward the VPS checkout to origin tip before/with deploy so git and binary match.
- Bump cache-bust only if the client changes (e.g. `?v=a9phases`).
- Append a section to `.director/CORE-FOCUS-PROGRESS.md` (SHA, what changed, tests, deploy, smoke, residuals) and push.
- Keep OSS engine Veilspan-neutral (numbers in `config/combat_balance.yaml` / generic code).

## Step 0 — Align VPS git with live tip
1. On clawdbot: `git fetch && git status -sb`; tip should include worldmap P1d (`44cac0c` or newer) and B13.
2. On VPS: `git fetch origin engine-june && git pull --ff-only origin engine-june` (resolve only if ff fails — do not reset through Marcus data).
3. Confirm checkout SHA matches origin tip before rebuilding.

## Goal
Named bosses feel like fights with chapters, not a single HP bar. When a boss crosses configured HP thresholds, combat enters a new phase with a clear BattleStage beat and (optionally) different telegraph/enrage behavior. Trash/elites unchanged unless config says otherwise.

## Done when
1. `config/combat_balance.yaml` gains a config-driven `boss_mechanics.phases` (or sibling) list: at least two phases after the opening (example bands ~100–66%, ~66–33%, ~33–0%), each with a short label and optional overrides for telegraph label / enrage multipliers / damage dealt. Prefer data over hardcoded boss names.
2. Engine transitions when current HP crosses a threshold (once per threshold, no thrash). Emits a clear combat event/message the client can show (phase enter). Existing A6 telegraph + enrage still work; document how they compose with phases (e.g. enrage only in final phase, or keep current rules).
3. BattleStage shows a phase banner or nameplate pip on transition (and current phase is visible during the fight). Respect `prefers-reduced-motion`. Group fights (A8) and B13 focus/restore stay intact.
4. At least one live content boss (or scaled boss path used by Catacomb / existing boss tier) actually uses the phases in smoke — not only a unit-test fixture.
5. Tests: Go coverage for threshold transitions, no double-fire, trash/elite unchanged; client test or small node test for phase banner if UI changed. `go test` for combat/balance packages green; `npm run build` green if client touched.
6. Smoke: guest (or two) fights a real boss-tier enemy; screenshot phase mid-fight at 1920×1080 into `.director/ux-audit/after/a9-phase-*.png`. Live `/play/` 200, guest 200, Door untouched.
7. Progress entry + push.

## Out of scope
New zones/content rooms, Door, party features, full add-summon AI system, DSA armor-weight (still unsigned), settings panel, further worldmap polish.

Start now. Read A6 + Round 2 + B13 sections in CORE-FOCUS-PROGRESS.md for deploy/smoke patterns.
