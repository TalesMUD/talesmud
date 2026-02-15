# User Interface Guide

TalesMUD features a modern, customizable interface with multiple display panels (widgets) that show different aspects of the game.

## Widget System

Widgets are movable, resizable panels that display game information. You can customize their position and size to create your perfect layout.

### Available Widgets

#### Room Widget
Displays your current location with rich visual information.

**Features:**
- **Room Image** - Pixelated background artwork for each location
- **Room Name** - Centered with decorative styling
- **Room Description** - Detailed text describing your surroundings
- **Parchment Style** - Optional background texture (enable in Settings)
- **NPCs/Entities** - Cards showing creatures and people in the room
- **Entity Details** - Name, type, level, and health bars

**Entity Actions:**
- **Talk** (Blue button) - Converse with friendly NPCs
- **Trade** (Green button) - Buy/sell with merchants
- **Attack** (Red button) - Engage enemies in combat

#### Terminal Widget (Classic)
Traditional MUD terminal using xterm.js rendering.

**Features:**
- Classic monospaced terminal display
- Command history (Up/Down arrow keys)
- Auto-complete suggestions
- Font size controls (S/M/L buttons)
- Line wrapping for long text
- Cursor blink indicator

**Best for:** Players who prefer traditional MUD gameplay

#### Terminal X Widget (Enhanced)
Modern terminal with enhanced visual styling.

**Features:**
- Syntax highlighting (cyan exits, bright room names)
- Custom color classification
- Improved readability
- Font size controls
- Same command capabilities as classic terminal

**Best for:** Players who want enhanced visuals while keeping terminal play

#### Action Bar Widget
Quick-access command interface.

**Features:**
- **Direction Buttons** - N, E, S, W, Up, Down (only available exits shown)
- **Standard Commands** - Look, Inventory, Equipment, Character
- **Combat Commands** - Attack, Defend, Flee (when in combat)
- **More Menu** - Additional commands (Who, Help, Drop, Use, etc.)
- **Room Actions** - Context-specific actions (Pull, Push, etc.)
- **Pickup Menu** - Select items from the ground

**Layout:**
- Automatically adapts based on game state
- Combat mode replaces standard commands with combat options
- Merchant present adds "List" button
- Room actions appear inline or in "Actions" menu

#### Character Widget
Displays your character's vital statistics.

**Information Shown:**
- **Name** - Your character's name
- **Race** - Character race (Human, Elf, Dwarf, etc.)
- **Class** - Character class (Warrior, Mage, Rogue, etc.)
- **Level** - Current level
- **HP Bar** - Health points (current/max) with visual bar
- **XP Bar** - Experience points (current/next level) with progress bar
- **Gold** - Amount of currency (formatted: 1.5K, 2.3M, etc.)
- **Attributes** - Grid showing all character stats (Strength, Agility, etc.)

**Visual Indicators:**
- **Red Glow** - When in combat
- **Combat Badge** - Shows "IN COMBAT" status

#### Inventory Widget
Manage items you're carrying.

**Features:**
- **Category Organization** - Equipment, Consumables, Quest Items, Other
- **Collapsible Categories** - Click to expand/collapse
- **View Modes:**
  - **Grid View** - Thumbnails with item icons
  - **List View** - Compact text list
- **Item Details** - Click any item to see full stats
- **Item Actions** - Equip, Use, Drop buttons
- **Item Quality Colors:**
  - Gray - Normal
  - Blue - Magic
  - Purple - Rare
  - Orange - Legendary
  - Red - Mythic

**Item Information:**
- Name and quality
- Type (Weapon, Armor, Consumable, etc.)
- Attributes (Damage, Defense, Strength, etc.)
- Level requirement
- Equipment slot
- Stack quantity (for stackable items)

#### Equipment Widget
Shows gear currently equipped on your character.

**Equipment Slots:**
- **Head** - Helmets, hats, crowns
- **Neck** - Amulets, necklaces
- **Chest** - Armor, robes, shirts
- **Main Hand** - Primary weapons, tools
- **Off Hand** - Shields, secondary weapons
- **Hands** - Gloves, gauntlets
- **Legs** - Pants, leggings, leg armor
- **Boots** - Footwear
- **Ring 1** - First ring slot
- **Ring 2** - Second ring slot

**Features:**
- Visual grid showing all slots
- Empty slots clearly indicated
- Click equipped item to see details
- Unequip button on item details popup

#### Quest Log Widget
Track your quest progress.

**Features:**
- **Active Quests** - Quests you're currently working on
- **Completed Quests** - Quests you've finished
- **Quest Details:**
  - Quest name and description
  - List of objectives
  - Progress tracking (e.g., "2/5 wolves defeated")
  - Checkmarks for completed objectives
- **Expandable View** - Click quest to expand/collapse details

**Quest Notifications:**
- **Quest Accepted** - Toast notification when starting a quest
- **Quest Progress** - Updates when objectives advance
- **Quest Completed** - Notification with reward details

#### Minimap Widget
Visual map of explored areas.

**Features:**
- **Room Grid** - Canvas-based map showing room positions
- **Current Location** - Highlighted in distinct color
- **Visited Rooms** - Shows all rooms you've explored
- **Floor-Based** - Only displays current Z-level (floor)
- **Room Connections** - Lines showing exits between rooms
- **Distance Fade** - Rooms farther away are more transparent
- **Hover Tooltips** - Mouse over rooms to see names
- **Auto-Updates** - Tracks as you explore

**Limitations:**
- Resets when you reload the page (session storage)
- Only tracks rooms with cardinal directions
- Shows up to 5 rooms away

## Layout Customization

### Edit Mode

To customize your layout:

1. Click **Edit Layout** button
2. **Drag widgets** by their title bars to move them
3. **Resize widgets** by dragging their corners/edges
4. **Add widgets** using the widget menu
5. **Remove widgets** by clicking the X button
6. Click **Save Layout** to persist changes

### Default Layout

The standard layout includes:
- **Left:** Room Widget
- **Right:** Terminal Widget
- **Bottom:** Action Bar Widget
- **Side Panels:** Character, Inventory, Equipment, Quest Log, Minimap

### Layout Tips

- **Keep Action Bar visible** - Essential for navigation
- **Terminal + Room side-by-side** - Classic MUD experience
- **Stack vertically on smaller screens** - Better for narrow displays
- **Hide unused widgets** - Reduce clutter
- **Save multiple layouts** - Create different setups for different activities

## Room Text Overlay

An optional feature that displays game messages as overlays on the room image.

**Features:**
- Messages appear over the room image
- Auto-fade after a few seconds
- Supports bold formatting (`**text**`)
- Slide-in animation
- Always enabled on mobile

**Enable/Disable:** Settings → Interface → Room Text Overlay

## Terminal Features

### Command History

- Press **Up Arrow** - Previous command
- Press **Down Arrow** - Next command
- Cycles through your recent commands

### Auto-Complete

Start typing common commands and press Tab:
- `nor` + Tab → `north`
- `sou` + Tab → `south`
- `say` + Tab → `say `

### Font Sizing

Both Terminal and Terminal X widgets have S/M/L buttons to adjust font size:
- **S** - Small (compact, more lines visible)
- **M** - Medium (balanced)
- **L** - Large (easier to read)

Your preference is saved automatically.

## Widget Descriptions

### When to Use Each Widget

| Widget | Use When | Skip When |
|--------|----------|-----------|
| Room | Always recommended | You prefer pure text |
| Terminal/TerminalX | You like typing commands | You only use buttons |
| Action Bar | Always recommended | - |
| Character | You want quick stat visibility | You check stats rarely |
| Inventory | Managing items frequently | You use terminal commands |
| Equipment | Optimizing gear setup | You rarely change equipment |
| Quest Log | Tracking objectives | You're not questing |
| Minimap | Navigating complex areas | You know the world well |

## Mobile Interface Differences

On mobile devices, the interface adapts:

- **Tab-based layout** - Switch between Room and Console tabs
- **Bottom sheets** - Character, Inventory, Equipment in drawers
- **Touch-optimized buttons** - Larger hit targets (44px minimum)
- **Vertical stacking** - Widgets stack instead of side-by-side
- **Room text overlay** - Always enabled
- **Safe area padding** - Respects device notches

See the [Mobile Play Guide](11-mobile.md) for full details.

## Next Steps

- Learn all available commands in the [Commands Reference](03-commands.md)
- Understand your character stats in [Character System](04-character.md)
- Master item management in [Inventory & Equipment](05-inventory-equipment.md)
