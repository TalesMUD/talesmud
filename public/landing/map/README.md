# Veilspan public world map (`/map`)

Fullscreen SVG + JSON map for **veilspan.com/map** (marketing / overview — not the in-game Cartographer).

## Data flow

1. Loremaster edits `zones/map/world-map-lore.v1.json` (may include `internal[]`).
2. Geometry lives in `zones/map/map-geometry.v1.json`.
3. Build PUBLIC payload (strips `internal`):

```bash
python3 tools/build_public_map_data.py
```

4. Ships as `landing/map/map-data.json` — **never** contains `internal`.

## Imagine art drop-in

| Slot | Path | Notes |
|------|------|-------|
| World plate | `assets/world/world-plate.webp` | Full 1800×1580 parchment art; optional |
| Zone plates | `assets/zones/z00-plate.webp` … | Shown inside zone ellipse when present |
| Icons | `assets/icons/*.svg` | Future POI icon set |

Until art lands, CSS parchment + zone fills render.

## LOD

- **A / continent** — landmass washes, minimal labels  
- **B / zones** — zone names, levels, connection lines  
- **C / POIs** — inns, bosses, hubs (live zones only)

Auto LOD follows zoom; toolbar forces a layer.

## Engine serving

Landing middleware serves files under `LANDING_PATH`. Directory indexes (`/map/` → `map/index.html`) require the patched `landing.go` in the engine.
