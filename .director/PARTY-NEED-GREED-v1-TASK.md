# SYSTEM crown 2026-10-13 — Party Need/Greed loot v1

You are on clawdbot in `~/dev/talesmud-needgreed` on branch `feat/need-greed` (tracks `origin/engine-june` tip `aa40a6f` / live VPS examine cards `?v=examine1`). Implement, test, commit `[grokbot]`, push ff to `engine-june`, deploy VPS the usual way. No Cursor cloud. No public master merge. Do not touch Door `:8020` / tales-door / door.veilspan.com.

Harness: **cursor-agent --yolo** (Grok weekly ~9% — do not use grok CLI for this crown).

## Why
Party gold+XP share, follow, combat assist, and invite Accept/Decline are live. Item drops from party kills are still free-for-all on the ground (explicit residual on Party Loot / Party Follow briefs). Close that loop with a small Need/Greed/Pass prompt so party play feels fair without inventing a full loot council.

## Done-when
1. **Trigger**: when a hostile dies and the fight has **2+ living party members** who are online and in the same room as the corpse (same eligibility band as today’s XP/gold share), each **item** drop from that kill (not gold) opens a Need / Greed / Pass round instead of instantly becoming free ground loot for whoever clicks first.
2. **Prompt**: every eligible member gets a popup (mirror party-invite UX: clear title, item name + rarity/quality chrome if already on the wire, Need / Greed / Pass, countdown). TTL **30s** from open. Commands also work: `need`, `greed`, `pass` (and `loot need|greed|pass` aliases if they do not collide).
3. **Resolve**: Need beats Greed beats Pass. Among the highest tier that anyone picked, pick one winner uniformly at random. Winner receives the item in inventory (or on the ground at their feet if inventory full — same as today’s grant path). Announce a short room/party line: who won and Need vs Greed. If **everyone Passes** or the TTL ends with no Need/Greed, the item stays on the corpse/ground as today’s free-for-all.
4. **Queue**: multiple drops from one kill resolve **one at a time** (no overlapping prompts). A member who disconnects, leaves the room, dies, or leaves the party mid-round is dropped from that round (counts as Pass).
5. **Solo / non-party**: unchanged — items drop to the ground exactly as today. One living party member in room also unchanged (no prompt).
6. **Cache-bust**: rebuild play client `?v=needgreed1` (bundle + CSS if touched).
7. **Tests**: Go covering eligibility, Need>Greed>Pass, all-pass → ground, timeout → ground, solo unchanged, disconnect=pass, sequential multi-drop. Client unit tests for the popup model (pending, expiresAt, item label). `go test -count=1 -timeout 20m ./pkg/...` green; client tests + `npm run build` green.
8. **Docs**: FEATURES / Party notes — Party Need/Greed v1; residual still out: master-looter, bind-on-pickup, cross-room rolls, Flutter.
9. **Ship**: `git push origin feat/need-greed:engine-june` ff-only. Deploy VPS usual path (ff, build `bin/tales.next` with the VPS Go toolchain, keep `bin/tales.prev-<oldsha>`, SIGTERM talesmud MainPID only; Door untouched). Smoke: two guests party, kill a rat/pack that drops an item (or seed a guaranteed drop in a throwaway fight), confirm Need/Greed popup, one Need wins, inventory has the item; second drop can Pass→ground. `/play/?v=needgreed1` 200, `POST /api/guest` 200.

## Out of scope
Flutter, admin UI, Door, public master, new zones, DSA armor-weight, changing gold/XP share math, auto-pull into combat, inventing new loot tables.

## Report
Append a "Party Need/Greed v1" section to `.director/CORE-FOCUS-PROGRESS.md` (SHAs, rules, tests, deploy sha256s/PIDs, smoke, residuals), commit `[grokbot]`, push, then print exactly:
`NEEDGREED1 DONE <sha> ?v=needgreed1`
or `NEEDGREED1 BLOCKED: <reason>`.
