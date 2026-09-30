# WORLDMAP P1j — imagegen craft sheet prototype

Prototype built 2026-09-30 (Europe/Berlin). **Promoted to production** as P1j (`?v=worldmap-p1j`, sheet `c8169d17ef81`) after Marcus approved “Ship P1j imagegen sheet to veilspan.com”.

## Tool used

Primary: **Codex built-in `image_gen`** (ChatGPT login on clawdbot via `codex exec`). Platform `OPENAI_API_KEY` / `scripts/image_gen.py` returned credit-exhausted (429). Grok Build / Imagine CLI returned 402 (usage balance exhausted).

Secondary fill: **Gemini `gemini-2.5-flash-image`** for five mid-priority stamp plates during the Codex wave; three of those (07/12/13) were **re-generated with Codex** after Marcus feedback and chroma issues. Remaining Gemini rows: `torch`, `reeds`, `stump`, `flowers`, `hill`, `mesa`, `outcrop`, `cliff` (plate-09 + plate-11).

### Plate → tool

| Plates | Rows (4×6 cells) | Tool |
| --- | --- | --- |
| 00–06, 08, 10, 14 | terrains, primary town/tree/bridge/underground | Codex `image_gen` |
| 07, 12, 13 | mine/magic/ridge/canopy; well/cart/fence/windmill; lantern/barrel/bones/offshore | Codex `image_gen` (stamp fix pass) |
| 09, 11 | torch/reeds/stump/flowers; hill/mesa/outcrop/cliff | Gemini 2.5 Flash Image |

## Process

1. Slice P1i authored row order (60 rows) into 15 plates of 4 kinds × 6 variants.
2. Generate 3:2 plates with Kenney craft density, Veilspan dusk palette, **no heavy black outlines**, soft top-left light. Stamps prompted on **magenta `#FF00FF`** (Codex often returned native alpha instead; assembler accepts both).
3. **Marcus mid-run correction:** terrains must be **seamless tileable** (not postcard chips). Assembler `assemble_p1j.py` applies wrap-offset + center crossfade `make_seamless`, harder edge match, then nearest-neighbor snap to **48px**. Engine **directional dither blend rows** recomposed from the seamless ground variants (same Bayer mask path as `generate_map_tiles.py`).
4. Stamps: corner flood-fill + magenta/pink key; reject full-bleed framed scenes; preserve transparent overlays.
5. Atlas layout unchanged: **288×9024**, 48px cells, 6 variants, same row order so `sync_map_tiles` validation still applies if promoted later.
6. Local preview `/tmp/worldmap-p1j-preview` on `127.0.0.1:8155` (P1i preview tree with swapped sheet + patched `TERRAIN_SHEET`); capture via `tools/capture_worldmap_p1j.cjs`.

## Artifacts

- Content sheet: `talesmud-rpg-1/assets/map-tiles/prototypes/p1j/{terrain-sheet.png,terrain-sheet.json,terrain-review.png,assemble_p1j.py}`
- Raw plates + seamless 2×2 checks: `.../p1j/_raw/`
- Screenshots: `.director/ux-audit/after/worldmap-p1j-{maxzoom,oldtown,overview}-1920x1080.png` + `worldmap-p1j-terrain-review.png`
- Box copies for parent: `/workspace/ux-audit/worldmap-p1j-*.png`

## Measurement

| | bytes | notes |
| --- | ---: | --- |
| P1g production | 471,730 | live |
| P1i procedural craft | 551,122 | prior prototype |
| **P1j imagegen** | **1,389,081** | +837,959 vs P1i (+152%); denser photographic-pixel detail |

- P1j SHA-256: `c8169d17ef813582c8c6e96c4b749025da2e268f26db6a7d760f7c737412b2ee`
- Manifest version: `c8169d17ef81`
- Capture: zero page errors; `imageSmoothingEnabled=false`; world-fit max cell step **99.815625px** (unchanged clamps).

## Judgment

Vs P1i procedural stamps: buildings/trees/props read as authored Kenney-like craft with dusk grading instead of Pillow brushes. Vs postcard risk: terrain cells now report **lr/tb edge MSE 0** after `make_seamless`; 2×2 wrap checks under `_raw/seamless-*-2x2.png`. Biome-to-biome morph still uses engine blend overlays (not painted transition tiles). Stamps are chroma/alpha cutouts, not full-bleed scenes.

Residual visual risk: Gemini plate-09/11 props may sit slightly off Codex style; optional Codex regen of those two plates. Imagegen atlas is ~2.5× P1i weight — worker bake cost not remeasured. Production client/sheet untouched.

## Residuals / TODO

- Optional: Codex-replace Gemini rows (plate-09, plate-11) for full style lock.
- Optional: regenerate any terrain plate that still vignettes after seamless post (none flagged by edge MSE).
- Promoted via `sync_map_tiles` into production; live client cache `?v=worldmap-p1j`.
