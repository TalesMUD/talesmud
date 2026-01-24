# MUD Web Client Redesign Proposals

## Current Issues

1. **Box-in-box effect**: Terminal sits inside room image container with hard borders, creating visual nesting
2. **Terminal constraints**: 640px max-width, 60% height with small font
3. **Container width**: Limited to 900px max
4. **No entity rendering**: No visual representation of NPCs/enemies/merchants

---

## Layout Option A: Side-by-Side Layout

```
┌─────────────────────────────────────────────────────────┐
│  [Room Image]              │  [Terminal - 60%+ width]   │
│  + Entity sprites          │  Larger font, full height  │
│  overlaid on image         │  No inner border           │
│                            │                            │
├────────────────────────────┼────────────────────────────┤
│  [Entity Cards: NPCs/Enemies/Merchants in this room]    │
└─────────────────────────────────────────────────────────┘
```

### Characteristics
- Clear separation between visual (image) and text (terminal)
- Room image on left with entity sprites overlaid
- Terminal takes majority of width on right
- Entity cards in horizontal strip below
- Good for widescreen displays

### Pros
- Clean visual hierarchy
- Easy to scan both image and text
- Entity panel has dedicated space

### Cons
- Image may feel disconnected from terminal narrative
- Less immersive feel

---

## Layout Option B: Overlay with Glass Morphism

```
┌─────────────────────────────────────────────────────────┐
│ ┌─────────────────────────────────────────────────────┐ │
│ │  [Full-width Room Image with gradient fade]         │ │
│ │  Entity sprites positioned on image                 │ │
│ │                                                     │ │
│ │  ╔═══════════════════════════════════════════════╗  │ │
│ │  ║ Terminal (glass/blur bg, no hard borders)     ║  │ │
│ │  ║ 50% larger, modern monospace font             ║  │ │
│ │  ╚═══════════════════════════════════════════════╝  │ │
│ └─────────────────────────────────────────────────────┘ │
│ [Quick action bar with animated entity cards]           │
└─────────────────────────────────────────────────────────┘
```

### Characteristics
- Evolution of current design, fixes the "box in box" feel
- Terminal uses glass-morphism (semi-transparent + backdrop blur)
- No hard borders, softer visual integration
- Entity sprites float on the room image
- Action bar below with entity interaction cards

### Pros
- Familiar layout for existing users
- Maintains immersion with room image visible
- Modern aesthetic with glass effects
- Smooth transition from current design

### Cons
- Text readability depends on background image
- May need careful contrast management

---

## Layout Option C: Full-Width Immersive

```
┌─────────────────────────────────────────────────────────┐
│ [Room Image as blurred background - full viewport]      │
│                                                         │
│   ┌──────────────────┐  ┌──────────────────────────┐   │
│   │ Entity Panel     │  │ Terminal (main focus)    │   │
│   │ - Merchant       │  │ Glass panel, large font  │   │
│   │ - Guard (NPC)    │  │ 1000px+ width            │   │
│   │ - Goblin (enemy) │  │ Smooth scroll            │   │
│   └──────────────────┘  │                          │   │
│                         └──────────────────────────┘   │
│                                                         │
│   [Exits]  [Actions]  [Inventory]  [Character]         │
└─────────────────────────────────────────────────────────┘
```

### Characteristics
- Room image becomes full-viewport blurred background
- Floating panels for entity list and terminal
- Entity panel on left shows all room occupants
- Terminal is the main focus, large and prominent
- Fixed action bar at bottom

### Pros
- Most modern and immersive feel
- Clear entity visibility with dedicated panel
- Maximum terminal real estate
- Great for combat scenarios (enemy list visible)

### Cons
- Loses the "framed" pixel art look
- More dramatic departure from current design
- Background blur may reduce pixel art appeal

---

## Shared Improvements (Apply to Any Option)

### Entity Rendering
- Pixel-art sprites or stylized cards for NPCs/enemies
- Health bars for enemies (during combat)
- Interaction icons: talk, attack, trade
- Animated idle states (subtle bobbing/breathing)
- Visual indicators for hostile vs friendly

### Terminal Enhancements
- **Glass-morphism background**: `backdrop-filter: blur(10px); background: rgba(0,0,0,0.7)`
- **Larger font**: 16-18px base, JetBrains Mono or Fira Code
- **Color-coded output**:
  - Combat messages: red/orange
  - Dialog/NPC speech: yellow/gold
  - System messages: cyan
  - Room descriptions: white
  - Items: green
- **Smooth typing effect** for narrative text
- **Auto-scroll with smooth animation**

### Animations
- Room transition: crossfade with slight zoom
- Entity entrance: fade-in + slide from edge
- Entity exit: fade-out
- Button hover: scale(1.05) + glow
- Panel expand/collapse: smooth height transition
- Combat hit effects: screen shake or flash

### UI Polish
- Consistent border-radius: 8-12px
- Subtle box-shadows instead of hard borders
- Hover states for all interactive elements
- Progress/health bars with gradient fills
- Tooltips for icons and entities
- Keyboard shortcuts displayed subtly

### Responsive Considerations
- 1000-1200px optimal width
- Collapse to single-column on mobile
- Touch-friendly button sizes (44px minimum)
- Swipe gestures for panel navigation on mobile

---

## Technical Implementation Notes

### CSS Variables for Theming
```css
:root {
  --terminal-bg: rgba(0, 0, 0, 0.75);
  --terminal-blur: 12px;
  --accent-color: #f59e0b;
  --text-primary: #e5e7eb;
  --text-muted: #9ca3af;
  --border-radius: 12px;
  --transition-fast: 150ms ease;
  --transition-normal: 300ms ease;
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
}
```

### Key Svelte Store Additions
```javascript
// Entity store for room occupants
{
  entities: [
    { id: 'npc_1', name: 'Innkeeper', type: 'npc', sprite: 'innkeeper.png' },
    { id: 'enemy_1', name: 'Goblin', type: 'enemy', health: 80, maxHealth: 100 },
    { id: 'merchant_1', name: 'Blacksmith', type: 'merchant', sprite: 'smith.png' }
  ]
}
```

### Animation Library Consideration
- Consider using `svelte/transition` built-ins
- Or lightweight library like `motion` (formerly Framer Motion)
- CSS animations for simple effects to minimize bundle size
