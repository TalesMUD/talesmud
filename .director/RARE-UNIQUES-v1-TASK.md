# RARE UNIQUES v1 — Crown 2026-10-09 (Veilspan daily director)
Marcus backlog item 11: Diablo-style rare unique boss-only drops (HK first, never guaranteed).

## Engine (generic, neutral — no Veilspan names in engine code)
- Loot table entries can carry `rarity: unique` + `chance` + `boss_only: true`.
- Unique ownership cap: a character may hold at most 1 of each unique (bag+equipped+bank); roll skipped if already owned.
- On drop: room-wide gold "UNIQUE" announcement chip (reuse boss-sig gold chip style) + item card gets a unique frame on web.
- Fix the two stale tests on engine-june: TestPartyCreateSetsLeaderAndRichMembers (class kit names) and TestGapMatrixTargets (tune or rescope bounds honestly; explain).
## Content (talesmud-rpg-1)
- Hollow Knight: Unmarked Vigil Blade ITM0265 moved onto this unique path (keep 10%).
- Add 2 new uniques: Burrow Brute and Orc War-Chanter, 5–8% each, stats sane for their level band, short punchy blurbs (veilspan-short-room-voice style).
## Done when
- go test ./... green; sim of 200 kills per boss shows drop rate within ±3pp and never a duplicate.
- Committed [grokbot] on engine-june + content; deployed via normal path; https://veilspan.com/play/?v= bumped to uniques1 and returns 200; guest login OK.
- Screenshot of the unique drop chip in .director/ux-audit/after/uniques1-*.png.
- End with: UNIQUES DONE <engine sha> <content sha>.
No Cursor cloud. No ad-hoc VPS edits outside deploy path.
