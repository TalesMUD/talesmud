# Veilspan public world map (`/map`)

Fullscreen SVG + JSON map for **veilspan.com/map** (marketing / overview — not the in-game Cartographer).

Chrome matches the landing design system (abyss bg, Cinzel/Cormorant/Fira, green glow chips, scanline/vignette).

## Asset paths

HTML uses `<base href="/map/">` plus absolute `/map/...` links and `?v=map10` cache-bust.
Engine `landing.go` redirects `/map` → `/map/` so CSS/JS never resolve as SPA HTML.

## Views (WoW-style)

- **Continent** — `assets/world/world-plate.jpg` (2048²). Town/city icons sit on painted settlements. Click a zone or town to enter its map.
- **Zone** — `assets/zones/zXX-plate.jpg`. POI icons (inn, bindstone, dungeon, boss, town, road) on the zone plate. Back, zoom-out past min, or Escape returns to the continent.

## Data flow

1. Loremaster edits `zones/map/world-map-lore.v1.json` (may include `internal[]`).
2. Geometry lives in `zones/map/map-geometry.v1.json` (continent pixels + per-zone POI coords).
3. Build PUBLIC payload (strips `internal`):

```bash
python3 tools/build_public_map_data.py
```

4. Ships as `landing/map/map-data.json` — **never** contains `internal`.

## Imagine art drop-in

| Slot | Path | Notes |
|------|------|-------|
| World plate | `assets/world/world-plate.jpg` | 2048² continent art; SVG viewBox matches |
| Zone plates | `assets/zones/zXX-plate.jpg` | Full-bleed map after clicking a zone |
| Icons | `assets/icons/{inn,bindstone,dungeon,boss,town,road}.png` | POI markers |

## Engine serving

Landing middleware serves files under `LANDING_PATH`. `/map` redirects to `/map/`.
