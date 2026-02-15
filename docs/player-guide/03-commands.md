# Commands Reference

Complete reference of all commands available in TalesMUD.

## Movement Commands

### Cardinal Directions

Move in the four main directions:

| Command | Shortcut | Description |
|---------|----------|-------------|
| `north` | `n` | Move north |
| `south` | `s` | Move south |
| `east` | `e` | Move east |
| `west` | `w` | Move west |

**Note:** Only available exits will appear as buttons in the Action Bar.

### Vertical Movement

Move up and down:

| Command | Shortcut | Description |
|---------|----------|-------------|
| `up` | `u` | Go up (stairs, ladders, etc.) |
| `down` | `d` | Go down |

### Special Exits

Some rooms have unique exits with custom names (e.g., "enter portal", "climb tree"). These will appear as buttons when available.

## Information Commands

View game state and character info:

| Command | Shortcut | Description |
|---------|----------|-------------|
| `look` | `l` | Examine the current room |
| `inventory` | `i` | View your inventory |
| `equipment` | `eq` | View equipped gear |
| `character` | `c` | View character stats |
| `who` | - | List players currently online |
| `help` | `?` | Display help information |

## Interaction Commands

### Looking and Examining

| Command | Description | Example |
|---------|-------------|---------|
| `look` | Re-examine the current room | `look` |
| `examine [target]` | Look closely at something | `examine statue` |

### Item Commands

| Command | Description | Example |
|---------|-------------|---------|
| `pickup [item]` | Pick up an item from the ground | `pickup sword` |
| `drop [item]` | Drop an item from inventory | `drop torch` |
| `use [item]` | Use or consume an item | `use potion` |
| `equip [item]` | Equip an item from inventory | `equip helmet` |
| `unequip [item]` | Remove equipped item | `unequip shield` |

**Tip:** You can also use interactive buttons on item cards instead of typing these commands.

### NPC Interaction

| Command | Description | Example |
|---------|-------------|---------|
| `talk [npc]` | Start conversation with NPC | `talk guard` |
| `[number]` | Select dialog option | `1` or `2` |

During conversations, you'll be presented with numbered options. Type the number to select your response.

## Combat Commands

These commands are available during combat encounters:

| Command | Description |
|---------|-------------|
| `attack` | Attack the enemy |
| `defend` | Take defensive stance (reduce damage) |
| `flee` | Attempt to escape combat |
| `status` | Check current combat status |

**Combat Flow:**
1. Initiate combat by attacking an enemy
2. Action Bar switches to combat commands
3. Choose your action each turn
4. Combat ends when enemy is defeated, you die, or you flee successfully

## Social Commands (Emotes)

Express yourself with emotes:

| Command | Description |
|---------|-------------|
| `shrug` | Shrug your shoulders |
| `scream` | Let out a scream |

**Note:** More emotes may be available - check the "More" menu.

## Merchant Commands

When near a merchant NPC:

| Command | Description |
|---------|-------------|
| `list` | View merchant's inventory |
| `buy [item]` | Purchase an item |
| `sell [item]` | Sell an item from inventory |

**Note:** The `list` button appears automatically when a merchant is present.

## Utility Commands

| Command | Description |
|---------|-------------|
| `bind` | Set your bind point (recall location) |

## Room-Specific Actions

Some rooms have special actions you can perform:

| Command | Example Rooms |
|---------|---------------|
| `pull [object]` | Lever, rope, chain |
| `push [object]` | Button, door, block |
| `read [object]` | Sign, book, scroll |
| `open [object]` | Chest, door, container |
| `close [object]` | Door, container |
| `unlock [object]` | Locked door, chest |

**Note:** Available actions appear as buttons in the Action Bar or in the "Actions" menu.

## Command Syntax

### Basic Rules

1. **Case insensitive** - `NORTH`, `North`, and `north` all work
2. **Shortcuts available** - Most commands have short versions
3. **Spaces separate arguments** - `pickup rusty sword`
4. **Tab completion** - Start typing and press Tab

### Target Names

When targeting items or NPCs:

- Use the exact name: `talk merchant`
- Multi-word names need full text: `pickup iron sword`
- Some commands accept partial matches: `attack gob` might work for "goblin"

### Examples

```
look                    # Examine room
look at statue         # Alternative syntax
inventory              # View items
i                      # Same as inventory
pickup sword           # Pick up item
equip iron helmet      # Equip specific item
talk to guard          # Talk to NPC
1                      # Select first dialog option
attack wolf            # Start combat
flee                   # Escape combat
```

## Button vs Text Commands

Many actions can be performed two ways:

### Using Buttons
- Click direction buttons (N, E, S, W)
- Click command buttons (Look, Inventory, etc.)
- Click action buttons on entity/item cards
- Click "More" for additional commands

### Using Text
- Type commands in the terminal
- Faster for experienced players
- Access to command history (Up/Down arrows)
- Auto-complete support

**Recommendation:** Use buttons while learning, switch to text commands for speed.

## Command Availability

### Always Available
- Movement (north, south, east, west, up, down)
- Information (look, inventory, equipment, character)
- Item management (pickup, drop, use, equip)
- Social (who, help)

### Context-Dependent
- **Combat commands** - Only during combat
- **Merchant commands** - Only near merchants
- **Room actions** - Only in specific rooms
- **Dialog options** - Only during NPC conversations

### Disabled During
- **Movement** - While in combat (must flee first)
- **Standard commands** - Replaced by combat commands during fights
- **Some actions** - While in dialog with NPC

## Special Command Notes

### Look Command
- Refreshes room information
- Shows current entities and items
- Updates exit availability
- Useful after combat or item pickups

### Inventory vs Equipment
- **Inventory** - Items you're carrying (not equipped)
- **Equipment** - Items currently worn/wielded
- Some items are in both (equipped items still "in" inventory)

### Use Command
- Consumables (potions, food) - Consumed and removed
- Tools (keys, torches) - Used but kept
- Quest items - May trigger quest events
- Effect varies by item type

## Advanced Command Tips

### Command History
- **Up Arrow** - Cycle to previous command
- **Down Arrow** - Cycle to next command
- Repeats recent commands quickly
- Saved per terminal session

### Auto-Complete
- Type partial command + Tab
- Works for common commands
- Suggests completions
- Saves typing time

### Batch Actions
Most actions happen one at a time. To perform multiple actions:
```
pickup sword
equip sword
look
```

Enter each command separately.

## Getting Help

### In-Game Help
- Type `help` for general help
- Type `help [command]` for specific command help (if supported)
- Check the "More" menu for less common commands

### This Documentation
- Return to this reference anytime
- Each section has more detailed command explanations
- See specific system guides for advanced tactics

## Command Categories Summary

| Category | Commands |
|----------|----------|
| **Movement** | north, south, east, west, up, down, [special exits] |
| **Information** | look, inventory, equipment, character, who, help |
| **Items** | pickup, drop, use, equip, unequip |
| **Combat** | attack, defend, flee, status |
| **NPC** | talk, [dialog numbers] |
| **Merchant** | list, buy, sell |
| **Social** | shrug, scream |
| **Utility** | bind, examine |
| **Room Actions** | pull, push, read, open, close, unlock |

## Next Steps

- Learn about [Character System](04-character.md) to understand stats
- Read [Combat System](06-combat.md) for fighting strategies
- Check [Inventory & Equipment](05-inventory-equipment.md) for item management
