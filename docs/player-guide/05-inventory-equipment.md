# Inventory & Equipment

Complete guide to managing items, gear, and your character's belongings.

## Inventory System

Your **inventory** holds all items you're carrying, whether equipped or not.

### Viewing Inventory

**Methods:**
1. **Inventory Widget** - Visual panel with item icons
2. **`inventory` command** - List in terminal
3. **Inventory button** - In Action Bar

### Inventory Organization

Items are organized into categories:

#### Equipment
- Weapons (swords, bows, staves)
- Armor (helmets, chest pieces, boots)
- Accessories (rings, amulets, belts)

#### Consumables
- Healing potions
- Mana potions
- Food and drink
- Temporary buff items

#### Quest Items
- Items needed for quests
- Cannot be dropped or sold
- Automatically removed when quest completes

#### Other
- Crafting materials
- Collectibles
- Currency items
- Miscellaneous items

### View Modes

**Grid View:**
- Shows item thumbnails/icons
- Visual representation
- Easier to identify items at a glance
- Best for players who prefer visual interfaces

**List View:**
- Compact text list
- Shows more items at once
- Item names and key stats
- Best for fast scanning

**Toggle:** Click the view mode button in the Inventory Widget.

### Inventory Capacity

- **Limited Space** - You can only carry so much
- **Weight/Slot System** - May use weight or slot limits
- **Overencumbered** - Too many items may slow you down
- **Managing Space:**
  - Drop unneeded items
  - Sell to merchants
  - Use/consume items
  - Store in banks (if available)

## Item Types

### Weapons

**Melee Weapons:**
- Swords, axes, maces, daggers
- Equipped in Main Hand or Off Hand
- Primary attribute: Damage
- Secondary: Strength, Critical, Speed

**Ranged Weapons:**
- Bows, crossbows
- Equipped in Main Hand
- Primary attribute: Damage
- Secondary: Agility, Critical

**Magic Weapons:**
- Staves, wands, orbs
- Equipped in Main Hand
- Primary attribute: Magic Damage
- Secondary: Intelligence, Mana

### Armor

**Heavy Armor:**
- High defense rating
- Worn by Warriors, Paladins
- May reduce speed/agility
- Slots: Head, Chest, Hands, Legs, Boots

**Light Armor:**
- Moderate defense
- Worn by Rogues, Rangers
- Agility bonuses
- Allows mobility

**Cloth Armor:**
- Low defense, high magic stats
- Worn by Mages, Clerics
- Intelligence/Wisdom bonuses
- Mana regeneration

### Accessories

**Rings:**
- Two ring slots
- Various attribute bonuses
- Magic effects
- Stackable bonuses

**Amulets/Necklaces:**
- One neck slot
- Powerful unique effects
- Often quest rewards or rare drops

### Consumables

**Potions:**
- Instant or over-time effects
- Restore HP or Mana
- Temporary attribute buffs
- Single use

**Food:**
- Restores HP gradually
- May provide temporary buffs
- Stackable

**Scrolls:**
- One-time spell effects
- Don't require magic ability
- Consumed on use

## Item Quality

Items have quality ratings that indicate rarity and power:

### Quality Tiers

| Quality | Color | Description |
|---------|-------|-------------|
| **Normal** | Gray | Common items, basic stats |
| **Magic** | Blue | Uncommon, improved stats |
| **Rare** | Purple | Rare finds, strong bonuses |
| **Legendary** | Orange | Very rare, excellent stats |
| **Mythic** | Red | Extremely rare, best-in-slot |

### Quality Indicators

**Visual Cues:**
- **Border Color** - Colored border around item icon
- **Text Color** - Item name displayed in quality color
- **Glow Effects** - Higher quality items may glow

**Stat Differences:**
- Higher quality = Better attributes
- More attribute bonuses on high-quality items
- Unique effects on Legendary/Mythic items

## Item Attributes

Items can have various stat bonuses:

### Offensive Attributes

| Attribute | Effect |
|-----------|--------|
| **Damage** | Increases weapon damage |
| **Strength** | Boosts physical damage |
| **Critical** | Increases critical hit chance |
| **Speed** | Faster attack rate |

### Defensive Attributes

| Attribute | Effect |
|-----------|--------|
| **Defense** | Reduces damage taken |
| **Armor** | Physical damage reduction |
| **Health** | Increases max HP |

### Magic Attributes

| Attribute | Effect |
|-----------|--------|
| **Intelligence** | Increases magic damage |
| **Mana** | Increases max mana pool |

### Other Attributes

| Attribute | Effect |
|-----------|--------|
| **Agility** | Improves dodge, crit chance |

## Equipment System

### Equipment Slots

Your character has the following equipment slots:

| Slot | Item Type | Examples |
|------|-----------|----------|
| **Head** | Helmets, hats, crowns | Iron Helmet, Wizard Hat |
| **Neck** | Amulets, necklaces | Amulet of Protection |
| **Chest** | Armor, robes | Leather Chestpiece, Mage Robe |
| **Main Hand** | Primary weapon | Longsword, Staff, Bow |
| **Off Hand** | Shield, secondary weapon | Wooden Shield, Dagger |
| **Hands** | Gloves, gauntlets | Leather Gloves |
| **Legs** | Pants, leg armor | Chain Leggings |
| **Boots** | Footwear | Iron Boots, Soft Shoes |
| **Ring 1** | Ring | Ring of Strength |
| **Ring 2** | Ring | Ring of Protection |

### Viewing Equipment

**Methods:**
1. **Equipment Widget** - Visual slot display
2. **`equipment` command** - List in terminal
3. **Equipment button** - In Action Bar

### Equipment Display

The Equipment Widget shows:
- All equipment slots in a grid
- Empty slots (gray/transparent)
- Equipped items with icons
- Item names and quality colors

## Managing Items

### Picking Up Items

When items are on the ground:

**Method 1 - Button:**
1. Look for "Pickup" button in Action Bar
2. Click to see list of available items
3. Select item to pick up

**Method 2 - Command:**
```
pickup [item name]
```

**Example:**
```
pickup rusty sword
pickup health potion
```

### Dropping Items

**Method 1 - Inventory Widget:**
1. Open Inventory Widget
2. Click on item
3. Click "Drop" button in item details popup

**Method 2 - Command:**
```
drop [item name]
```

**Note:** Some items cannot be dropped (quest items, bound items).

### Using Items

**Method 1 - Inventory Widget:**
1. Click item in inventory
2. Click "Use" button in popup

**Method 2 - Command:**
```
use [item name]
```

**What Happens:**
- **Consumables** - Item is consumed and effect applied
- **Tools** - Item is used (may stay in inventory)
- **Quest Items** - May trigger quest progression

### Equipping Items

**Method 1 - Inventory Widget:**
1. Click item in inventory
2. Click "Equip" button in popup

**Method 2 - Command:**
```
equip [item name]
```

**Requirements:**
- Item must match your class/race (if restricted)
- Must meet level requirement
- Must have correct slot available

**Auto-Swap:**
- If slot is occupied, previous item is unequipped
- Unequipped item stays in inventory

### Unequipping Items

**Method 1 - Equipment Widget:**
1. Click equipped item
2. Click "Unequip" button in popup

**Method 2 - Inventory Widget:**
1. Click equipped item (marked with icon)
2. Click "Unequip" button

**Method 3 - Command:**
```
unequip [item name]
```

**Result:**
- Item moves to inventory (remains unequipped)
- Slot becomes empty
- Stats/bonuses removed

## Item Details Popup

When you click any item, a popup shows:

### Information Displayed

- **Item Name** (with quality color)
- **Item Type** (Weapon, Armor, Consumable, etc.)
- **Quality** (Normal, Magic, Rare, etc.)
- **Equipment Slot** (if equippable)
- **Level Requirement** (if any)
- **Attributes** (all stat bonuses)
- **Description** (flavor text or effect description)

### Attribute Display

**Color Coding:**
- **Red Numbers** - Offensive stats (Damage, Strength, Critical)
- **Blue Numbers** - Defensive stats (Defense, Armor, Health)
- **Green Numbers** - Magic stats (Intelligence, Mana)

### Available Actions

Buttons shown depend on item and context:
- **Equip** - If item is equippable and not equipped
- **Unequip** - If item is currently equipped
- **Use** - If item is consumable or usable
- **Drop** - If item can be dropped

## Item Comparison

### Comparing Gear

When deciding between items:

**Check:**
1. **Quality** - Higher quality usually better
2. **Level Requirement** - Can you use it?
3. **Attributes** - Does it boost stats you need?
4. **Slot** - Does it replace something you have?

**Example Decision:**

Current: Normal Iron Sword (Damage +10, Strength +2)
New: Magic Steel Sword (Damage +15, Strength +3, Critical +5)

→ Magic Steel Sword is better in all ways!

### Best Practices

**For Warriors:**
- Prioritize Strength, Damage, Defense
- Heavy armor over light
- Weapons with high damage

**For Mages:**
- Prioritize Intelligence, Mana
- Cloth armor with magic bonuses
- Staves and wands

**For Rogues:**
- Prioritize Agility, Critical, Speed
- Light armor for mobility
- Fast weapons with high crit

## Selling and Trading

### Selling to Merchants

When near a merchant NPC:

1. Click "Trade" button on merchant card
2. View your inventory
3. Select items to sell
4. Confirm sale
5. Receive gold

**Tip:** Sell items you don't need to make space and earn gold.

### Item Value

- Higher quality items sell for more
- Equipment generally worth more than consumables
- Quest items typically cannot be sold

### Buying from Merchants

1. Click "Trade" or use `list` command
2. Browse merchant's inventory
3. Select item to buy
4. Confirm purchase (if you have enough gold)

## Stackable Items

Some items stack in a single inventory slot:

**Stackable:**
- Consumables (potions, food)
- Crafting materials
- Currency items
- Arrows/ammunition

**Display:**
- Shows quantity (e.g., "Health Potion x5")
- Uses one inventory slot regardless of quantity

**Using Stacked Items:**
- Uses one from the stack
- Quantity decreases
- Slot freed when last one is used

## Special Item Types

### Quest Items

- Marked with "Quest" type
- Cannot be dropped or sold
- Required for quest completion
- Automatically removed when no longer needed

### Bound Items

- Bound to your character
- Cannot be traded or dropped
- Usually powerful rewards
- Marked with "Bound" indicator

### Keys

- Used to unlock doors or chests
- May be consumed on use or reusable
- Quest-specific or general

## Inventory Management Tips

### Keep Your Inventory Clean

**Regular Maintenance:**
- Sell items you don't need
- Drop vendor trash in safe locations
- Use consumables rather than hoarding
- Store valuables in bank (if available)

### Prioritize Space

**What to Keep:**
- Current equipment upgrades
- Healing items (always!)
- Quest items (automatic)
- Valuable items for selling

**What to Drop/Sell:**
- Lower-quality equipment you've outgrown
- Excess consumables beyond stack needs
- Items below your level
- Duplicate equipment

### Organize by Category

The Inventory Widget automatically categorizes items:
- Collapse categories you don't need to view
- Expand categories you use frequently
- Toggle view mode for different situations

### Quick Equipment Swaps

**Situation-Based Gear:**
- Combat gear (high defense)
- Magic gear (high intelligence)
- Travel gear (high speed)

Keep alternate sets for different situations.

## Gold Management

### Earning Gold

**Best Sources:**
- Quest rewards
- Selling unneeded equipment
- Defeating enemies
- Finding treasure chests

### Spending Gold Wisely

**Priority Spending:**
1. **Equipment Upgrades** - Biggest power increase
2. **Healing Potions** - Survival essential
3. **Quest Requirements** - Progress blockers

**Avoid:**
- Buying low-quality items you'll replace soon
- Excessive consumables you won't use
- Cosmetic items (unless you really want them!)

### Gold Display

- Shows in Character Widget
- Formatted for readability (1.5K, 2.3M, etc.)
- Updates in real-time

## Next Steps

- Master [Combat System](06-combat.md) to use your gear effectively
- Complete [Quests](07-quests.md) to earn better equipment
- Learn [Commands](03-commands.md) for faster item management
