# Gilded Frame theme pack

Engine theme id: `gilded`. The play client turns it on with `data-theme="gilded"` on `<body>`. Dark Fantasy stays the engine default. A deployment sets `ui.default_theme` in its ruleset. The public server-info field is `defaultTheme`. The client setting `auto` follows that value. A theme the player picks in Settings wins.

## Files

- `tokens.css` defines `--gf-*` and the Marcellus `@font-face`.
- `components.css` is the primitive contract below. It is scoped to `[data-theme="gilded"]`.
- `chrome.css` maps the look onto the existing play chrome without moving layout regions.
- `textures/` holds our generated stone, leather, and slot images.
- `fonts/Marcellus-Regular.woff2` plus the source TTF. License: `fonts/OFL.txt` (SIL OFL 1.1).

## Provenance

The textures in `textures/` and the SVG icons in `slots/` are original works, released under this repository's MIT license. The Marcellus font is under the SIL Open Font License 1.1; see `fonts/OFL.txt`.

`index.html` loads `tokens.css` before `themes.css`, then `components.css` and `chrome.css` after `global.css`, all with the same `?v=` cache tag as the other play CSS.

## Primitives

| Class | Role |
|---|---|
| `.gf-panel` | Textured body, bronze double bevel, inner vignette. |
| `.gf-titlebar` | Centered Marcellus title, placed as the first child of a panel. |
| `.gf-subtitle` | Gold line under the title (level and class). |
| `.gf-close` | Red square, gold ×. A direct child of `.gf-panel` sits at the panel's top-right. The button draws the × itself. |
| `.gf-plaque` | Leather section header. Put the label as the element's text. |
| `.gf-stat` | One stat band. Odd rows use row A. |
| `.gf-stat__label` | Gold label. The colon is added for you. Pass `Health`, not `Health:`. |
| `.gf-stat__value` | Right-aligned value. Base color is warm white. |
| `.gf-stat--buff` / `.gf-stat--debuff` | Green or red value. |
| `.gf-slot` | 40×40 recessed slot, bronze frame, four corner studs. |
| `.gf-slot--md` / `.gf-slot--lg` | 48px and 56px. |
| `data-quality` | Inner border. Engine names: `normal` white, `magic` green, `rare` blue, `legendary` purple, `mythic` and `unique` orange. Gilded names: `poor`, `common`, `uncommon`, `rare`, `epic` (purple). |
| `data-slot` | Empty silhouette: `helm`, `neck`, `shoulder`, `back`, `chest`, `shirt`, `tabard`, `wrist`, `hands`, `waist`, `legs`, `feet`, `ring`, `trinket`, `main-hand` (or `mainhand`), `off-hand` (or `offhand`), `ranged`, `bag`. Anything else, including a missing attribute, is a neutral rune. |
| `data-filled="true"` or `.gf-slot--filled` | Hides the empty silhouette so item art can sit in the slot. |
| `.gf-slot__count` | Bottom-right count. No pill. |
| `.gf-slot--selected` or `aria-selected="true"` | Gold double rim. Hover and focus use the bright bronze rim. |
| `.gf-iconbtn` | 32px square. The hit box extends past the border so the touch target is at least 44px. |
| `.gf-iconbtn--active` or `aria-pressed="true"` | Bright rim. |
| `.gf-medallion` | Circular portrait frame. Put an `img` inside. |
| `.gf-tab` | Square side tab. `.gf-tab--active` or `aria-selected="true"` uses a bright rim and a 2px outward shift. Reduced motion drops the shift. |
| `.gf-scroll` | Thin bronze scrollbar on the element that scrolls. |
| `.gf-btn.gf-btn--primary` | Leather plaque, gold text. |
| `.gf-btn.gf-btn--danger` | Red fill. |
| `.gf-btn.gf-btn--secondary` | Panel fill, warm text. |
| `.gf-bar.gf-bar--hp` / `--mana` / `--xp` | Recessed track. Put `.gf-bar__fill` inside for the meter and `.gf-bar__label` for the centered numbers. |
| `.gf-input` | Recessed field. Focus uses the bright bronze rim. |

Frames on existing widgets stay inside the current boxes. Do not grow padding or border width on those widgets to fit a primitive. Drop a primitive inside the box, or let `chrome.css` paint with shadows.
