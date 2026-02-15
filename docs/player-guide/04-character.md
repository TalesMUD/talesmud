# Character System

Understanding your character's stats, progression, and capabilities.

## Character Information

### Basic Information

Your character has the following core attributes:

| Field | Description |
|-------|-------------|
| **Name** | Your character's unique name |
| **Race** | Your character's race (Human, Elf, Dwarf, Orc, etc.) |
| **Class** | Your character's class (Warrior, Mage, Rogue, Cleric, etc.) |
| **Level** | Current experience level (1-∞) |

### Viewing Your Character

Use these methods to view your character:

1. **Character Widget** - Always-visible panel (if enabled)
2. **`character` command** - Type in terminal
3. **Character button** - Click in Action Bar

## Vital Statistics

### Hit Points (HP)

**Health points** represent your life force.

- **Current HP** - How much health you have now
- **Max HP** - Your maximum health capacity
- **HP Bar** - Visual representation (green = healthy, red = critical)

**When HP reaches 0:**
- Your character dies
- You may respawn at your bind point
- Possible penalties (varies by game rules)

**Restoring HP:**
- Use healing potions
- Rest in safe locations
- Visit healers
- Level up (restores to max)
- Some equipment provides regeneration

### Experience Points (XP)

**Experience** measures your progression toward the next level.

- **Current XP** - Total experience gained
- **XP to Next Level** - Experience needed to level up
- **XP Bar** - Visual progress indicator

**Gaining XP:**
- Defeating enemies in combat
- Completing quests
- Discovering new locations
- Solving puzzles
- Special achievements

### Level

Your **character level** indicates overall power and progression.

**Benefits of Leveling Up:**
- Increased maximum HP
- Higher stats/attributes
- Access to better equipment
- New abilities (class-dependent)
- Increased carrying capacity
- Stronger in combat

**Level Display:**
- Shown on Character Widget
- Visible to other players
- Determines equipment requirements

### Gold

**Currency** used throughout the game world.

- **Display Format:**
  - Under 1,000: Shows exact amount (e.g., "523 Gold")
  - 1,000+: Shows abbreviated (e.g., "1.5K Gold")
  - 1,000,000+: Shows abbreviated (e.g., "2.3M Gold")

**Earning Gold:**
- Defeating enemies (loot drops)
- Completing quests (rewards)
- Selling items to merchants
- Finding treasure chests
- Trading with other players

**Spending Gold:**
- Buying equipment from merchants
- Purchasing consumables (potions, food)
- Paying for services (training, repairs)
- Unlocking special features

## Attributes

**Attributes** are your character's core stats that affect various aspects of gameplay.

### Common Attributes

| Attribute | Affects |
|-----------|---------|
| **Strength** | Physical damage, carrying capacity |
| **Agility** | Attack speed, dodge chance, critical hit chance |
| **Intelligence** | Magic damage, mana pool, spell effectiveness |
| **Constitution** | Maximum HP, health regeneration |
| **Wisdom** | Magic resistance, mana regeneration |
| **Dexterity** | Accuracy, ranged damage, evasion |
| **Charisma** | Merchant prices, NPC reactions, leadership |
| **Luck** | Critical hit chance, loot quality, random events |

**Note:** Actual attributes may vary. Check your Character Widget to see which attributes your character has.

### How Attributes Work

**Base Attributes:**
- Set by race and class during character creation
- Increase when you level up

**Equipment Bonuses:**
- Weapons, armor, and accessories add attribute points
- Bonuses shown on item details
- Stacks with base attributes

**Temporary Buffs:**
- Potions, spells, and abilities can temporarily boost attributes
- Effects wear off after duration
- May stack with equipment

### Viewing Attributes

The Character Widget shows all attributes in a grid:
- Attribute name
- Current value (including all bonuses)
- Color-coded or formatted display

## Character Progression

### Leveling Process

1. **Gain Experience** - Defeat enemies, complete quests
2. **Fill XP Bar** - Progress toward next level
3. **Level Up** - Automatic when XP requirement is met
4. **Stat Increases** - HP, attributes automatically improve
5. **Notification** - You'll be informed of the level up

### Level Milestones

Different levels may unlock:
- **Level 5** - New equipment tier available
- **Level 10** - Advanced areas accessible
- **Level 20** - Elite equipment unlocked
- **Level 30+** - Endgame content available

**Note:** Specific milestones depend on game design.

### Power Curve

As you level up:
- Enemies in starter areas become trivial
- New areas with tougher enemies become accessible
- Equipment requirements increase
- Quest difficulty scales
- Rewards improve

## Character States

### Normal State
- Full access to all commands
- Can move freely
- Can interact with NPCs and objects
- Stats display normally

### In Combat
- Character Widget glows red
- "IN COMBAT" badge shown
- Movement restricted (must flee first)
- Combat commands replace standard commands
- HP changes visible in real-time

### Dead
- HP at 0
- Cannot perform most actions
- May need to respawn
- Possible penalties apply

### Bind Point

Your **bind point** is where you respawn after death.

**Setting Bind Point:**
- Use `bind` command in a safe location
- Some locations (inns, temples) allow binding
- Starting area is default bind point

**Benefits:**
- Quick return to familiar location
- Strategic positioning near quest areas
- Safety net for exploration

## Character Display

### Character Widget Features

The Character Widget shows:
- Portrait or icon (if available)
- Name, race, class, level
- HP bar with current/max values
- XP bar with progress percentage
- Gold amount (formatted)
- Attributes grid
- Combat status indicator

### HP Bar Colors

| Color | Meaning |
|-------|---------|
| **Green** | Healthy (75-100% HP) |
| **Yellow** | Wounded (25-75% HP) |
| **Red** | Critical (0-25% HP) |

### XP Bar

- Fills from left to right
- Shows progress to next level
- Resets upon leveling up
- Percentage display available

## Character Optimization

### Maximizing Stats

**Through Equipment:**
- Equip items with attribute bonuses
- Prioritize attributes for your class
  - Warriors: Strength, Constitution
  - Mages: Intelligence, Wisdom
  - Rogues: Agility, Dexterity
- Match gear to your playstyle

**Through Leveling:**
- Complete quests for extra XP
- Fight enemies at or above your level
- Explore for discovery XP
- Efficient progression path

### Survival Tips

**Keep HP High:**
- Carry healing potions
- Rest after combat
- Don't overextend in dangerous areas
- Watch HP bar during fights

**Manage Gold:**
- Save for important upgrades
- Sell unneeded items
- Complete quests for gold rewards
- Don't buy every item you see

**Level Efficiently:**
- Focus on quests (better XP than grinding)
- Fight appropriate-level enemies
- Explore new areas
- Complete objectives systematically

## Character Comparison

### Checking Other Characters

When viewing other players or NPCs, you can see:
- Their name and level
- Their race and class (usually)
- Their equipped items (sometimes)

**Level Differences:**
- Much higher level = Very dangerous enemy or strong player
- Similar level = Balanced encounter
- Much lower level = Easy opponent or newer player

## Stats Summary

### Essential Stats to Watch

| Stat | Why Important | Monitor |
|------|---------------|---------|
| **HP** | Survival | Always |
| **XP** | Progression | After each battle/quest |
| **Gold** | Equipment upgrades | Before shopping |
| **Level** | Overall power | After XP gains |
| **Attributes** | Combat effectiveness | When equipping items |

### Quick Reference Command

Type `character` or `c` anytime to review all stats.

## Advanced Character Topics

### Character Building

**Race Selection:**
- Affects starting attributes
- May have racial abilities
- Cosmetic differences in descriptions

**Class Selection:**
- Determines playstyle
- Affects available equipment
- May unlock class-specific abilities

**Hybrid Builds:**
- Balance multiple attributes
- Flexibility vs. specialization trade-off
- Requires more strategic equipment choices

### Min-Maxing

For competitive players:
- Focus all resources on key attributes
- Optimize every equipment slot
- Efficient quest pathing for maximum XP
- Strategic combat choices

**Note:** Not necessary for casual play!

## Next Steps

- Learn about [Inventory & Equipment](05-inventory-equipment.md) to gear up your character
- Master [Combat System](06-combat.md) to level up effectively
- Check [Quest System](07-quests.md) for efficient XP and gold gains
