# Responsive audit of veilspan.com/play — 2026-10-14

Live play client `?v=needgreed1` (cache tag confirmed on every viewport). Throwaway guests only. Viewports: phone 375×812 (DPR 3, touch), tablet 768×1024 (touch), desktop 1440×900, bonus phone landscape 812×375. Headless Chromium. No game code was changed and nothing was deployed.

Shots: `~/dev/briefs/UX-AUDIT-2026-10-14/shots/<size>/<screen>.png`. Contact sheets: `shots/contact-phone.png`, `shots/contact-tablet.png`, `shots/contact-desktop.png`, `shots/contact-phone-landscape.png`.

Walk, same on every size: guest class picker, Awakening Chamber (R0001), Collapsed Corridor pickup, Spectral Warden dialog and flint examine in Torch Alcove, rune puzzle, Rat Nest ground loot, then a two-guest party fight against the Catacomb Rats. The leader attacked; the guest stayed in the party in the room. After the fight, on to Mira Thornwood at the Weary Wanderer (R0203). Guests: phone Wayfarer_6273 (Ward) with Stranger_f647; tablet Outsider_4ba3 (Rune Hand) with Wayfarer_e7cc; desktop Drifter_5bf5 with Vagrant_3daa; landscape Rover_6599 (Alley) with Pilgrim_1b7c.

Phone and tablet both get the mobile shell (`mobileStore.js`, `MOBILE_BREAKPOINT` 768, so width ≤ 768). Landscape is 812px wide, so `mobile=false` and `kindForWidth` returns compact (width under 1100).

A resource 404 showed up in the console on a few screens (room, map, inventory, party invite). It did not change the UI. Headless Chromium does not show an OS keyboard, so keyboard-over-input was not observed.

## Top 5

1. **UX-02** Dialog and shop are trapped inside the room image. Phone and tablet hide most dialog options and clip the shop. Desktop clips the last dialog option. Tue 20.
2. **UX-01** Phone hotbar slots 1 and 9 are clipped, and the page does not scroll to them. The combat dock shows the same cut. Wed 21.
3. **UX-03** The victory or defeat card covers Need/Greed while the 30s roll timer runs, and quest toasts stack on top. Tue 20.
4. **UX-04** The mobile header and Room/Console tab bar paint over inventory and the party invite form. Tue 20.
5. **UX-05** The phone shell has no character, equipment, skills, or quest-log panel. Stats and Equip print to the terminal. Thu 22.

UX-08 (phone landscape crushed, and the home link blocks NPC cards) is the next one down. It needs its own card.

12 issues: P0=3, P1=6, P2=3.

## Coverage

OK means the screen was reached and none of the issues below is that cell's fault. Issue means a listed issue is visible there. Not reached is explained under the table.

| Screen | Phone | Tablet | Desktop | Landscape |
| --- | --- | --- | --- | --- |
| Landing / guest start | OK | OK | OK | OK |
| Character select | OK | OK | OK | OK |
| Room | issue | issue | OK | issue |
| Hotbar | issue | OK | OK | OK |
| Console / terminal | issue | issue | OK | issue |
| Toast | OK | OK | OK | OK |
| Inventory | issue | issue | OK | OK |
| Map (full) | OK | OK | OK | OK |
| Minimap | not reached | not reached | not reached | not reached |
| Character sheet | issue | issue | OK | not reached |
| Equipment | issue | issue | OK | not reached |
| Skills / spells | issue | issue | OK | not reached |
| Quest log | issue | issue | OK | not reached |
| Overflow / pin sheet | issue | OK | OK | OK |
| Settings | issue | OK | OK | OK |
| Shortcuts (`?`) | OK | OK | OK | OK |
| Party, empty | OK | OK | OK | OK |
| Pickup | OK | OK | OK | OK |
| NPC dialog | issue | issue | issue | not reached |
| Examine card | issue | OK | OK | not reached |
| Rat Nest | OK | OK | OK | OK |
| Combat | issue | OK | OK | OK |
| Combat outcome | issue | issue | issue | issue |
| Party invite popup | OK | OK | OK | OK |
| Party with a member | issue | issue | OK | OK |
| Need / Greed | issue | not reached | issue | issue |
| Shop | issue | issue | OK | not reached |
| Loading | not reached | not reached | not reached | not reached |

Character select is the guest class/race picker. The signed-in creation wizard is not on the guest path.

Not reached:

- Loading, all four sizes. The welcome screen replaced it before a frame was visible.
- Minimap, all four. Phone and tablet have no minimap widget; the Map button opens the full Cartographer overlay, which did open. Desktop and landscape: the minimap tab is not in the active preset, including at 1440×900.
- Landscape character, equipment, quest log. The compact preset has no those tabs.
- Landscape skills. Skills sit on the character sheet, and that tab is absent, so skills were not opened.
- Landscape dialog and examine. The Spectral Warden card was on screen. `<a href="/" class="back-link">` intercepted the click (UX-08).
- Landscape shop. The walk reached R0203 and Mira's card was on screen. The same back link intercepted the click.
- Tablet Need/Greed. The party fight ended in defeat (respawn R0001 at 9/19 HP) before a drop, so no loot roll opened. Phone, desktop, and landscape each got a Rat Tail roll.

## Issues

### P0

#### UX-01 — Phone hotbar clips slots 1 and 9

- Sizes: phone. Tablet fits (nine slots, 55×55, first left 113, last right 655, `.slots` 560/560). Landscape fits (nine 36×36 slots inside 812). Desktop fits.
- Screen: room, hotbar, combat dock.
- What's wrong: at 375px the compact slot is `clamp(40px, 7.2vw, 56px)`, so it sits on the 40px floor. Nine 40×40 slots: slot 1 is left −15 / right 25, slot 9 is left 350 / right 390. Viewport is 375. `.slots` scrollWidth 382 vs clientWidth 359. The document scrollWidth stays 375 and `overflowX` is false, and `.slots` does not set `overflow-x`, so the clipped slots cannot be scrolled into view. `phone/combat.png` shows the same nine-slot row inside the battle dock, with the end slots cut at the dock edges. The narrow `.combat-hotbar` rule does set `overflow-x: auto`, so that row may scroll, but the resting frame does not show both ends as full targets.
- Shots: `shots/phone/room.png`, `shots/phone/hotbar.png`, `shots/phone/combat.png`.
- Where: `public/mud-client/src/game/widgets/HotbarWidget.svelte` (`.slots` around line 215, `.slot` width at line 226). Combat row: `public/mud-client/src/game/ui/BattleStage.svelte` (`.combat-hotbar` / `.hb-slot`, narrow rule around line 3742).
- Effort: M. Sprint: Wed 21 touch + hotbar.

#### UX-02 — Dialog and shop are mounted in the room image

- Sizes: phone, tablet, desktop (dialog). Shop is phone and tablet. Desktop shop fits four wares and is usable.
- Screen: NPC dialog, shop.
- What's wrong: `MobileRoomView` mounts `DialogOverlay`, `ShopOverlay`, and `RecipesOverlay` inside `.room-image-section`, and the overlays are `position: absolute; inset: 0` on that box. Phone dialog box is top 48, bottom 308, height 260, width 375 (viewport 812). Option 1 is partly inside (top 251). Options 2–6, through "I should go." at top 509, sit outside the box. There is no close control; the server ends the dialog, and Esc does not (dialog is not in the shortcut panel order). Phone shop clips the second ware (Weak Health Potion) while the action bar stays live underneath. Tablet dialog box is top 48, bottom 348, height 300, width 768. Options 1–2 are inside; option 3 starts at top 351, below the image. Tablet shop fits a 2×2 of four wares, and the room controls stay live under it. Desktop dialog box is top 61, bottom 566, height 505, width 808. Options 1–5 are visible. Option 6 "I should go." is top 560 / bottom 613, past the room panel. Recipes uses the same mount and was not opened.
- Shots: `shots/phone/dialog.png`, `shots/phone/shop.png`, `shots/tablet/dialog.png`, `shots/tablet/shop.png`, `shots/desktop/dialog.png`. Desktop shop: `shots/desktop/shop.png`.
- Where: `public/mud-client/src/game/mobile/MobileRoomView.svelte` (mounts around lines 318–334). `public/mud-client/src/game/ui/DialogOverlay.svelte` (absolute inset, z-index 100, `overflow-y: auto`). `public/mud-client/src/game/ui/ShopOverlay.svelte` (absolute inset, z-index 120, `overflow: hidden`). On desktop the dialog is still sized to the room column rather than the viewport.
- Effort: M. Sprint: Tue 20 top fixes.

#### UX-03 — Victory card covers Need/Greed while the timer runs

- Sizes: phone, tablet, desktop, landscape.
- Screen: combat outcome, Need/Greed, quest toasts.
- What's wrong: the battle stage is `position: fixed; inset: 0; z-index: 9000`. The outcome panel is `position: absolute; inset: 0` with a 62% black backdrop and z-index 5 inside that stage, so it takes the clicks. Need/Greed is z-index 146, under the stage. Quest toasts (`.veilspan-card`, z-index 1200) sit above Need/Greed and under the stage. On the phone victory shot the card covers Need, Greed and Pass show through the backdrop and cannot be clicked, and Pest Control toasts cover the reward text. After Continue, the Rat Tail roll is usable (Need / Greed / Pass, countdown visible) with the same toasts stacked on it. The script had to press Continue before a clean roll shot on phone, desktop, and landscape. Desktop shows the victory card on the left and the roll in the center, both visible, with toasts stacked on the right. Tablet's outcome is a defeat card floating over a still-live room (exits and hotbar bright) after a respawn to R0001 at 9/19 HP. The 30s roll timer runs while the outcome card is up.
- Shots: `shots/phone/combat-after.png`, `shots/phone/need-greed.png`, `shots/desktop/combat-after.png`, `shots/desktop/need-greed.png`, `shots/tablet/combat-after.png`, `shots/phone-landscape/combat-after.png`, `shots/phone-landscape/need-greed.png`.
- Where: `public/mud-client/src/game/ui/BattleStage.svelte` (stage z-index 9000 around line 1419, `.outcome-panel` around line 3072). `public/mud-client/src/game/ui/NeedGreedOverlay.svelte` (z-index 146). `public/mud-client/src/game/ui/QuestNotifications.svelte` (z-index 1200).
- Effort: M. Sprint: Tue 20 top fixes.

### P1

#### UX-04 — Header and tab bar cover inventory and party

- Sizes: phone, tablet (both `mobile=true`).
- Screen: inventory, party sheet.
- What's wrong: inventory is `position: fixed; inset: 0; z-index: 120`. The mobile header and the Room/Console tab bar are z-index 1001, so they paint over the sheet. On the phone the inventory title sits under the header. The party invite field is cut to "Invite player by nam", and the Invite button is cut by the tab bar. On the tablet the inventory title is under the header, and the party Invite button's bottom border is clipped by the tab bar. The party sheet's mobile rule (`max-width: 768px`) pins the sheet to the bottom with a safe-area inset and no tab-bar inset. Esc still closes inventory.
- Shots: `shots/phone/inventory.png`, `shots/phone/party.png`, `shots/tablet/inventory.png`, `shots/tablet/party.png`.
- Where: `public/mud-client/src/game/ui/InventoryOverlay.svelte` (z-index 120). `public/mud-client/src/game/mobile/MobileHeader.svelte` (z-index 1001). `public/mud-client/src/game/mobile/MobileTabBar.svelte` (z-index 1001). `public/mud-client/src/game/ui/PartyOverlay.svelte` (z-index 130, mobile block around line 502).
- Effort: M. Sprint: Tue 20 top fixes.

#### UX-05 — No character, equipment, skills, or quest panel on the phone shell

- Sizes: phone, tablet.
- Screen: character, equipment, skills, quest log.
- What's wrong: Stats and Equip pins send the terminal commands `character` and `equipment`. The console prints the text and the bottom sheet does not open. `skills` / `spells` and `quests` also print to the console. `openBottomSheet` is only called for inventory. Desktop at 1440 has Character, Equipment, Skills (Slip, with Add to Hotbar), and a Quest Log tab in the tools column, and those screens are usable. Landscape compact omits the tabs (see the matrix).
- Shots: `shots/phone/character.png`, `shots/phone/equipment.png`, `shots/phone/skills.png`, `shots/phone/quests.png`, and the tablet equivalents. Desktop contrast: `shots/desktop/character.png`, `shots/desktop/skills.png`, `shots/desktop/equipment.png`, `shots/desktop/quests.png`.
- Where: `public/mud-client/src/game/mobile/MobileActionBar.svelte` (`openBottomSheet('inventory')` only). `public/mud-client/src/game/mobile/mobileStore.js`. Quest log widget is the desktop tools tab, not a phone sheet.
- Effort: L. Sprint: Thu 22 panels + type.

#### UX-06 — Tap targets under 44px on touch sizes

- Sizes: phone and tablet. Desktop is not a touch size in this pass.
- Screen: header, hotbar, players badge, console.
- What's wrong, measured from the live DOM: account button `.acct-btn` is 36×36 on phone and tablet. Phone hotbar slots are 40×40 (tablet slots are 55×55). Player badge is 44×24. Console input `.tx-input` is 21px tall (319×21 on the phone, 712×21 on the tablet). Font-size steppers `.tx-fontsize-btn` are 20×14. The Room/Console tab bar itself is 56px tall.
- Shots: `shots/phone/room.png`, `shots/phone/console.png`, `shots/phone/hotbar.png`, `shots/tablet/console.png`.
- Where: `MobileHeader.svelte` (`.acct-btn` 36×36, lines 59–60). `HotbarWidget.svelte` (40px floor). Console input and font stepper in the terminal widget styles.
- Effort: M. Sprint: Wed 21 touch + hotbar.

#### UX-07 — Type under 12px on the phone

- Sizes: phone. The tablet uses the same mobile console, so the same console sizes apply there. Measured on the phone.
- Screen: console, room, tabs, combat, map, pickup.
- What's wrong, computed font size: console section label "terminal" 8.4px; console body about 10.92px; Room/Console tab labels 10px; "tap to expand" 10px; action labels (EXAMINE ALTAR and the other room actions) 11px; enemy "Lv 1" 8px; map legend "Catacombs Intro" / "Safe" 10px; examine "Quality" / "Value" 8.68px.
- Shots: `shots/phone/console.png`, `shots/phone/room.png`, `shots/phone/combat.png`, `shots/phone/map.png`, `shots/phone/examine.png`.
- Where: `MobileTabBar.svelte` (tab label `font-size: 10px`, line 56). `MobileRoomView.svelte` ("tap to expand" and the pickup count badge at 10px). `MobileActionBar.svelte` (10px labels). Console body styles in the terminal widget.
- Effort: S/M. Sprint: Thu 22 panels + type.

#### UX-08 — Phone landscape is a crushed compact desktop, and the home link blocks NPC cards

- Sizes: phone landscape 812×375 only.
- Screen: room, console, dialog, examine, shop.
- What's wrong: width 812 is above the 768 mobile breakpoint and under 1100, so the compact preset stacks a room, a terminal, and an action bar into 375px of height. `mobile=false`. The room image is a thin strip. The terminal shows a single exit line. Room-action buttons are top 360 / bottom 396, past the 375px viewport, so the action bar is cut off. The nine hotbar slots do fit (top 309). The home link `a.back-link` is `position: absolute; left: 15px; top: 15px; z-index: 1000; padding: 1em`, and it is hidden only at `max-width: 768px`. It intercepted pointer events on the Spectral Warden card and on Mira Thornwood's card, so Talk, Examine, and Shop could not be clicked. Combat itself was usable: all nine slots, Flee, HP, and the log.
- Shots: `shots/phone-landscape/room.png`, `shots/phone-landscape/combat.png`. Dialog, examine, and shop have no landscape shot because of the intercept.
- Where: `public/mud-client/src/game/layout/layoutPresets.js` (`kindForWidth`, line 124). `public/mud-client/src/App.svelte` (`.back-link`, lines 41–47 and the 768px hide at line 63). `public/mud-client/src/game/mobile/mobileStore.js`.
- Effort: L. Sprint: new card.

#### UX-09 — Tablet portrait uses the phone shell and wastes the height

- Sizes: tablet 768×1024.
- Screen: room.
- What's wrong: width 768 is still mobile, so there is no character column and no minimap widget. A large empty band sits between the room text and the hotbar. The hotbar itself fits (see UX-01).
- Shot: `shots/tablet/room.png`.
- Where: `mobileStore.js` (`MOBILE_BREAKPOINT` 768). `layoutPresets.js` is not consulted while `mobile=true`.
- Effort: M. Sprint: new card (or Thu 22, if the tablet is treated as a panel-layout fix).

### P2

#### UX-10 — Settings shows developer override text

- Sizes: phone (the row was read there). The sentence lives in the shared modal.
- Screen: settings.
- What's wrong: under "Battle layout B" the description tells the player `?battleLayout=classic` and `localStorage talesmud_battle_layout_b=0`. The modal closes, and `.tab-content` scrolls (`overflow-y: auto`, container `max-height: calc(100dvh - 2rem)`), including on the landscape shot where the first viewport ends mid-list.
- Shot: `shots/phone/settings.png`.
- Where: `public/mud-client/src/game/ui/SettingsModal.svelte` (line 142).
- Effort: S. Sprint: Mon 26.

#### UX-11 — Examine card cuts the last line

- Sizes: phone.
- Screen: examine.
- What's wrong: the Flint and Steel card cuts the last line ("Essential for lighting") and the close control sits on the card edge. The NPC card overlaps the room image. The card is still readable enough to use.
- Shot: `shots/phone/examine.png`.
- Where: room examine card in `public/mud-client/src/game/ui/RoomTextOverlay.svelte` (the card is centered on the room hero).
- Effort: S. Sprint: Thu 22 panels + type.

#### UX-12 — Customize-pins sheet clips its footer

- Sizes: phone.
- Screen: overflow / pins.
- What's wrong: the customize-pins sheet draws the "Run command" footer against the action bar.
- Shot: `shots/phone/overflow.png`.
- Where: pin customizer sheet opened from `MobileActionBar.svelte`.
- Effort: S. Sprint: Thu 22 panels + type.

## Also seen, and left off the list

- Full map works on every size. The missing piece is the minimap widget, recorded as not reached above.
- Desktop shop, character, skills, equipment, inventory, and quest log were usable.
- Party invite popups opened on every size (two guests). Follow was not exercised.
- Empty party panel was reached on every size before create.
- Guest picker scrolls; later classes sit below the fold. Alley was reachable.
- One console 404 (resource load) on several screens. No player-facing fault came from it.
