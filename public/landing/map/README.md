# Veilspan public world map (`/map`)

Fullscreen SVG + JSON map for **veilspan.com/map** (marketing / overview — not the in-game Cartographer).

Chrome matches the landing design system (abyss bg, Cinzel/Cormorant/Fira, green glow chips, scanline/vignette).

## Asset paths

HTML uses `<base href="/map/">` plus absolute `/map/...` links and `?v=map2` cache-bust.
Engine `landing.go` redirects `/map` → `/map/` so CSS/JS never resolve as SPA HTML.

## Data flow

1. Loremaster edits `zones/map/world-map-lore.v1.json` (may include `internal[]`).
2. Geometry lives in `zones/map/map-geometry.v1.json`.
3. Build PUBLIC payload (strips `internal`):

```bash
python3 tools/build_public_map_data.py
```

4. Ships as `landing/map/map-data.json` — **never** contains `internal`.
5. Optional future: `cities[]` (capitals ≫ Oldtown ≫ towns) for LOD city markers.

## Imagine art drop-in

| Slot | Path | Notes |
|------|------|-------|
| World plate | `assets/world/world-plate.jpg` | Full continent art (relabel may refresh) |
| Zone plates | `assets/zones/z00-plate.webp` … | Inside zone ellipse when present |
| Icons | `assets/icons/{inn,bindstone,dungeon,boss,town,road}.png` | POI markers |

## LOD

- **Lands / continent** — plate + landmass washes  
- **Zones** — zone names, levels, connection lines  
- **POIs** — inns, bosses, hubs (live zones only)

Auto LOD follows zoom; toolbar chips force a layer.
