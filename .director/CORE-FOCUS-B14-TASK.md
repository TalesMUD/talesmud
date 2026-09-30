# B14 — Settings panel (play chrome + real prefs)

Crown 2026-09-30 (daily director). Browser UX track after accepting A9 boss phases + overnight worldmap P1e–P1j.

Repo: `~/dev/talesmud-june`, branch `engine-june`. Client: `public/mud-client`.

## Rules
- Every commit message contains `[grokbot]`.
- No Cursor cloud. Prefer Codex (`codex --dangerously-bypass-approvals-and-sandbox`, model from `~/.codex/config.toml`) if grok CLI is at weekly limit; otherwise grok 4.7 xhigh is fine. Standing rule: if grok hits usage limit mid-run, switch to Codex immediately.
- Deploy local `:8010` only (ff VPS `engine-june` to tip, copy client into `pkg/webuiplay/dist`, rebuild `bin/tales`, keep `bin/tales.prev-<oldsha>`, SIGTERM MainPID from `systemctl show -p MainPID --value talesmud`). Never touch Door `:8020`.
- Cache-bust client to `?v=b14settings` (and matching stylesheet/worker queries if touched).
- Append a section to `.director/CORE-FOCUS-PROGRESS.md` (SHA, what changed, tests, deploy, smoke, residuals) and push.
- Keep OSS engine product-neutral.

## Context
Settings already exist (`SettingsStore.js`, `SettingsModal.svelte`, chip menu → Settings in `CharacterSwitcher.svelte`) but the panel is half-wired: theme/parchment/compact/room overlay toggles, audio toggles that may not drive real audio, and no controls for the prefs players actually need after B9–B13/A9 (reduced motion, combat auto-focus, inventory open mode). Marcus plays at 1920×1080; modal must fit without page scroll. Gold the live demo chrome should match B6 header language (not leftover blue Material accents).

## Goal
Ship a Settings panel players can trust: open from the account chip (and Escape/`?` shortcut list already naming settings), persist real gameplay/accessibility prefs, and match the gold chrome theme.

## Done when
1. **Accessibility — reduced motion**: setting `interface.reducedMotion` = `system` | `on` | `off`. `prefersReducedMotion()` (and BattleStage/phase/hit juice) honor the override, not only OS media query. Default `system`.
2. **Combat — auto-focus BattleStage**: setting `interface.combatAutoFocus` (bool, default true). When false, B13 focus/restore does not auto-cover on combat start/join; manual focus and layout save behavior otherwise unchanged. Document in FEATURES.md.
3. **Inventory open mode**: expose existing `inventoryOpenMode` (`overlay` | `widget`) in the Interface tab with clear labels; changing it persists and applies without reload.
4. **Chrome polish**: Settings modal uses gold accents consistent with B6 (not blue primary). Escape closes when open (keyboardShortcuts already tracks `settings` overlay). Fits inside 1920×1080 and 1366×768 viewports without document scroll; phone usable.
5. **Honesty on audio**: either wire `soundEnabled` / volume sliders to whatever audio path exists, or disable/hide non-functional controls with a short “coming soon” note — do not leave dead toggles that look live.
6. **Tests**: node tests for reduced-motion override + combatAutoFocus gate (extend `hudPrefs_test.mjs` / `keyboardShortcuts_test.mjs` / `combatFocus_test.mjs` as needed). `npm run build` green. Go tests only if engine touched (prefer client-only).
7. **Smoke**: guest opens Settings from chip at 1920×1080; screenshot `.director/ux-audit/after/b14-settings-1920x1080.png`. Toggle reduced motion + combat auto-focus, reload, confirm persistence. Live `/play/` 200 with `?v=b14settings`, guest POST 200, Door untouched.
8. Progress entry + push.

## Out of scope
New themes beyond existing dark-fantasy/clean-hud, full sound system redesign, DSA armor-weight, Door, new zones/content, further worldmap art, party features, settings sync to server account.

Start now. Read B6/B13/A9 deploy+smoke patterns in CORE-FOCUS-PROGRESS.md.
