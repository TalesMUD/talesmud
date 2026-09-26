# Viewport audit (B1)

Captured 2026-09-24 against live Veilspan `http://127.0.0.1:8010/play/` (`?v=a3reward`) with a fresh guest in the Awakening Chamber. Tooling: headless Chromium via puppeteer-core (`/tmp/b1pw/audit.js`). Screenshots are in `.director/ux-audit/before/`.

The starter room has no enemy, so BattleStage was not on screen. Combat layout is called out from the same grid the room uses.

## Screenshots

| File | What it shows |
| --- | --- |
| `before/00-landing-1366x768.png` | Welcome (Sign up / Log in / Play as Guest) |
| `before/room-1366x768.png` | Room + terminal + hotbar + action bar |
| `before/room-1920x1080.png` | Same layout, more empty space under the bars |
| `before/room-2560x1440.png` | Same pattern, wider |
| `before/room-3440x1440.png` | Layout is a short band; most of the ultrawide is empty background |
| `before/room-mobile-390x844.png` | Phone shell (cold start at 390×844) |
| `before/inventory-1366x768.png` | Inventory overlay |
| `before/edit-1366x768.png` | Guest account menu (no Edit Layout) |
| `before/combat-1366x768.png` | Same as the room; no fight started |

## Findings

### 1. Desktop layout does not fill the viewport — high
The default grid is a fixed number of 40px rows (room/terminal height 12, hotbar 2, action bar 3). On 1366×768 there is already a dark band under the action bar. On 1920×1080 and especially 3440×1440 the playable UI is a strip across the upper third, with a large empty blurred background below and wide side margins. Widgets are not clamped or restretched when the window is larger than the preset.

### 2. Terminal text is clipped on the right — high
At 1366×768 the room description in the terminal is cut mid-word (`fo`, `fadin`). The panel does not wrap or scroll that line into view. The same clip shows up, milder, at 1920.

### 3. Hotbar is disconnected from the panels — medium
Eight slots float in the gap between the room/terminal and the action bar, not attached to either widget. Most slots are empty dashed boxes. On a tall screen that gap becomes a hole.

### 4. Account chip collides with panels — medium
The name chip (Nomad / Outsider, class, level) sits in the top-right corner, outside the terminal frame. On the inventory overlay it covers the modal's top-right corner, next to the close button.

### 5. Inventory has two headers — medium
The overlay draws an "Inventory" bar (gold, Craft, close) and the widget inside draws a second "INVENTORY" bar. Padding, type, and icon style do not match the room panel or the terminal. This is the chrome inconsistency B4 should collapse to one header.

### 6. Guests cannot open edit layout — high
The guest account menu is only Create Account, Settings, and End Session. Edit Layout is not in that menu, so resize handles, reset, and undo are unreachable for a guest. The saved-layout tools need a visible entry that does not depend on a full account.

### 7. Narrow viewport swap drops the desktop session — high
Resizing a live desktop session down to 390px replaced the play UI with the welcome screen, and widening it again did not return to the room. A cold start at 390×844 does show the phone shell, so mobile itself renders. Crossing the breakpoint should keep the guest session.

### 8. Phone room copy is clipped — medium
At 390×844 the room description ends in `whe…` over the hotbar. "Tap to expand" sits on top of that text. The hotbar and the stacked action buttons compete with the description. Button heights are uneven (direction keys shorter than EXAMINE).

### 9. No character sheet in the default desktop layout — low
Character stats are only the corner chip. Inventory is an overlay, not a grid widget. A viewport preset should still be able to open both without overlapping the chip.

### 10. BattleStage not exercised here — note
Awakening Chamber has no hostile NPC, so this pass has no combat screenshot. B4 should still give BattleStage the same header/padding rules as the other widgets when that screen is re-shot.

## What B2–B4 should do

- B2: pick a preset from the viewport on first load (no saved layout). Clamp every widget on load and resize so nothing is off-screen or zero-sized. Keep a saved layout; do not wipe it. Presets should use the vertical space instead of a fixed 12-row room.
- B3: Edit Layout has to be obvious (including for guests). Resize handles, a drop ghost, multi-step undo, reset-to-default, and a lock toggle. Fix the breakpoint session drop if it is an edit-mode or layout remount bug.
- B4: one widget header (title, collapse, maximize). One padding, font, and scrollbar. Inventory should not stack two titles. Re-shoot this folder after and add before/after notes here.

## After B4 (2026-09-25)

Re-shot against live Veilspan `http://127.0.0.1:8010/play/` (`?v=b4chrome`) with a fresh guest. Files are in `.director/ux-audit/after/`. Same sizes as the before set.

| File | What it shows |
| --- | --- |
| `after/room-1366x768.png` | Room scene title, Terminal header, chip in the top band |
| `after/room-1920x1080.png` | Same layout, grid fills the height |
| `after/room-2560x1440.png` | Same |
| `after/room-3440x1440.png` | Same, grid across the width |
| `after/room-mobile-390x844.png` | Shorter hero, description, hint, then hotbar and actions |
| `after/inventory-1366x768.png` | One Inventory header (gold, Craft, close) |
| `after/edit-1366x768.png` | Edit mode with gold corners, Undo, Lock |
| `after/combat-1366x768.png` | North from the chamber: Collapsed Corridor. No hostile, so BattleStage stayed closed |

Measured grid box (guest, no saved layout):

| Viewport | Grid top | Grid height | Grid bottom | Chip overlaps terminal |
| --- | --- | --- | --- | --- |
| 1366×768 | 60 | 704 | 764 | no |
| 1920×1080 | 60 | 1016 | 1076 | no |
| 2560×1440 | 60 | 1376 | 1436 | no |
| 3440×1440 | 60 | 1376 | 1436 | no |

The xterm screen ends about 34px inside the terminal panel, and long lines wrap onto the next row. At 1366 the room description in the terminal continues as `ht drifts` / `isper` / `uries` on the following line.

### What changed since the before shots

- The account chip and the TalesMUD link sit in a band above the grid. On the inventory overlay the chip is under the modal, so the close button is clear.
- Inventory has one title. Gold, Craft, and close are on that bar. The item count stays on the toolbar inside the panel.
- Terminal, inventory, equipment, character, quest log, and tab containers use one header: Cinzel title, collapse, and focus. The room keeps its scene title and gets the same two buttons. Hotbar and action bar stay chrome-free so a two-row bar is not eaten by a title. Scrollbars in panels, the terminal, and the inventory body use the same thin gold thumb.
- Phone: the hero is shorter. The description box is 101px with overflow hidden (scroll height 144). "TAP TO EXPAND" starts 6px below that box. The hotbar and action buttons follow in the column.
- Edit mode still shows the gold corners, Undo, and Lock. The chip no longer covers the terminal's remove button.

### After B5 (2026-09-26)

Re-shot the desktop room viewports against public `https://veilspan.com/play/` (`?v=b5wrap`). Files replaced: `after/room-1366x768.png`, `after/room-1920x1080.png`, `after/room-2560x1440.png`, `after/room-3440x1440.png`.

The terminal keeps `light`, `centuries`, and `whisper` on one line. The spell bar is the top strip of the action-bar frame (8 slots, empty dashes still visible), not a separate row. On 1366 the dock bottom is 732 in a 768 viewport (36px under it). On 1080 and 1440 the gap under the dock is 28px.

### Still open

- Default layout still has no character sheet. Stats stay on the corner chip.
- BattleStage uses the same header type, padding, and gold rule. This pass did not open a fight: north of the Awakening Chamber is an empty corridor.
- Focusing a panel parks the others at 2×2 under it until focus is toggled again. Saving while focused writes that arrangement.
- The top-right account chip, Edit Layout, and the party/friends icons still overlap. That is B6.
