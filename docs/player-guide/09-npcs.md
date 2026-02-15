# NPC Interactions

Learn how to interact with non-player characters (NPCs) for quests, trading, and story progression.

## NPC Overview

**NPCs (Non-Player Characters)** are computer-controlled characters that populate the game world. They provide quests, sell items, give information, and bring the world to life.

## NPC Types

### Friendly NPCs

**Characteristics:**
- Won't attack you
- Often have quests
- Provide information
- Part of story

**Visual Indicators:**
- Blue "Talk" button
- Neutral or friendly name color
- May have special markers

**Examples:**
- Quest givers
- Town guards
- Story characters
- Trainers

### Merchants

**Characteristics:**
- Buy and sell items
- Fixed inventory
- Respawn items periodically
- Located in towns usually

**Visual Indicators:**
- Green "Trade" button
- Merchant type label
- Often in shop locations

**Services:**
- Sell equipment
- Sell consumables
- Buy your items
- Specialized goods

### Enemies

**Characteristics:**
- Hostile creatures
- Will attack or can be attacked
- Drop loot when defeated
- Provide XP

**Visual Indicators:**
- Red "Attack" button
- Enemy type label
- Health bars
- Level display

**Examples:**
- Wolves, bandits, goblins
- Dungeon monsters
- Boss enemies

## Talking to NPCs

### Initiating Conversation

**Method 1 - Button:**
1. Look at NPC entity card in Room Widget
2. Click blue "Talk" button

**Method 2 - Command:**
```
talk [npc name]
```

**Examples:**
```
talk guard
talk village elder
talk merchant
talk mysterious stranger
```

### Dialog System

When you talk to an NPC:

**Dialog Box Appears:**
- Semi-transparent overlay
- NPC icon or portrait
- NPC name at top
- Dialog text (what they say)
- Numbered response options

**Example:**
```
Village Elder:
"Wolves have been attacking travelers. Can you help?"

1. "I'll take care of the wolves."
2. "Tell me more about the problem."
3. "I'm not interested."
```

### Selecting Responses

**Type the number** of your chosen response:

```
1
```

**Response Effects:**
- Option 1: Usually accepts quest
- Option 2: Usually asks for more information
- Option 3: Usually declines or ends conversation

**Multiple Choice:**
- Each option leads to different outcome
- Some choices affect quest path
- Some choices just provide info
- Dead-end choices may close dialog

### Dialog Flow

**Conversation Structure:**

1. **Initial Greeting**
   - NPC introduces themselves or situation
   - Presents main options

2. **Information Gathering**
   - Select "Tell me more" options
   - Learn background
   - Get quest details

3. **Decision Point**
   - Accept quest
   - Decline quest
   - Ask more questions

4. **Conclusion**
   - Quest accepted and added to log
   - Dialog ends
   - NPC returns to idle state

### Ending Dialog

**Dialog Ends When:**
- You accept a quest
- You decline and close conversation
- You select final option
- NPC has nothing more to say

**After Dialog:**
- You return to normal game mode
- Can talk to NPC again if needed
- New dialog options may appear after quest progress

## NPC Information Display

### Entity Cards

NPCs are displayed as entity cards in the Room Widget:

**Information Shown:**
- **Name** - NPC's name
- **Type** - Friendly, Merchant, or Enemy
- **Level** - Their level (if applicable)
- **Health Bar** - Current HP (for enemies)

**Action Buttons:**
- **Talk** (Blue) - Friendly NPCs
- **Trade** (Green) - Merchants
- **Attack** (Red) - Enemies

### Health Bars (Enemies)

Enemy NPCs display health bars:

**Colors:**
- **Green** - Healthy (75-100% HP)
- **Orange/Yellow** - Wounded (25-75% HP)
- **Red** - Critical (0-25% HP)

**Usage:**
- Gauge combat difficulty
- Decide when to flee
- Track battle progress

## Merchant NPCs

### Finding Merchants

**Common Locations:**
- Town centers
- Market squares
- Trading posts
- Shops and stores
- Inns and taverns

**Identifying:**
- Labeled as "Merchant" or specific role
- Green "Trade" button
- Often stationary
- Shop-themed room descriptions

### Trading with Merchants

**Opening Trade:**

**Method 1 - Button:**
1. Click green "Trade" button on merchant card

**Method 2 - Command:**
```
list
```

**Trade Interface:**
- Merchant's inventory displayed
- Your inventory shown
- Gold totals visible
- Buy/Sell options

### Buying Items

**Steps:**
1. Open trade with merchant
2. Browse their inventory
3. Select item to buy
4. Confirm purchase (if you have enough gold)
5. Item added to your inventory
6. Gold deducted

**Command:**
```
buy [item name]
```

**Example:**
```
buy iron sword
buy health potion
```

### Selling Items

**Steps:**
1. Open trade with merchant
2. Select item from your inventory
3. Click "Sell" option
4. Confirm sale
5. Item removed from inventory
6. Gold added

**Command:**
```
sell [item name]
```

**Example:**
```
sell rusty dagger
sell wolf pelt
```

### Merchant Types

**General Merchants:**
- Buy and sell most items
- Balanced prices
- General inventory

**Specialized Merchants:**
- **Weaponsmith** - Weapons only
- **Armorer** - Armor only
- **Alchemist** - Potions and ingredients
- **Food Vendor** - Consumables
- **Magic Shop** - Scrolls and enchanted items

**Better Prices:**
- Sell items to appropriate specialist
- Weaponsmith pays more for weapons
- Alchemist pays more for potions

### Merchant Inventory

**Stock:**
- Fixed set of items
- May refresh periodically
- Some items always available
- Rare items may stock out

**Pricing:**
- Fixed buy prices
- Fixed sell prices (usually ~50% of buy price)
- May be affected by Charisma stat
- Special discounts (quests, reputation)

## Quest-Related NPCs

### Quest Givers

**Finding Them:**
- Marked with special indicators (often)
- Mentioned in quest text
- Located in key story locations
- Sometimes hidden

**Starting Quests:**
1. Talk to quest giver
2. Read their dialog
3. Accept quest (option 1 usually)
4. Quest added to log

### Quest Turn-In

**Completing Quests:**
1. Complete all objectives
2. Return to quest giver
3. Talk to them
4. Select completion option
5. Receive rewards

**Multiple Visits:**
- First visit: Accept quest
- During: May provide hints
- Final visit: Turn in quest

### Quest Dialog Options

**Common Options:**

**"Tell me about..."**
- Background information
- Quest hints
- Lore

**"What should I do?"**
- Reminder of objectives
- Hints on location
- Strategy tips

**"I'm ready" / "I accept"**
- Quest acceptance
- Start quest chain
- Commit to task

**"I've completed your task"**
- Quest turn-in
- Receive rewards
- Progress story

## NPC Behavior

### Stationary NPCs

**Characteristics:**
- Stay in one room
- Always available
- Predictable location
- Easy to find

**Examples:**
- Shop merchants
- Quest givers in towns
- Guards
- Innkeepers

### Wandering NPCs

**Characteristics:**
- Move between rooms
- May have patrol routes
- Harder to find
- More realistic

**Examples:**
- Traveling merchants
- Patrol guards
- Wandering adventurers

**Finding Them:**
- Check multiple rooms
- Wait for them to return
- Follow patrol route
- Ask other players

### Event-Based NPCs

**Characteristics:**
- Only appear during certain times
- Triggered by quest progress
- Temporary presence
- Special circumstances

**Examples:**
- Quest-specific characters
- Boss enemies
- Special event NPCs

## Dialog Tips and Strategies

### Getting Information

**Ask Questions:**
- Select "Tell me more" options
- Exhaust all dialog trees
- NPCs often have hidden information
- Some conversations unlock secrets

**Example Flow:**
```
1. "Who are you?"
   → Learn about NPC
2. "Tell me about this area"
   → Get location tips
3. "Do you have any work?"
   → Reveals quest
4. "I'll help"
   → Accept quest
```

### Quest Optimization

**Gather All Quests:**
- Talk to all NPCs in an area
- Accept multiple quests
- Complete them together
- Return to turn in

**Check Back:**
- NPCs may have new quests after you level
- Return periodically
- Story progression unlocks new dialog

### Dialog Choices Matter

**Be Careful:**
- Some choices affect outcomes
- Rude responses may anger NPCs
- Helpful choices may give bonuses
- Quest paths may diverge

**Read Carefully:**
- Understand what each option means
- Think about consequences
- Roleplay or min-max (your choice)

## Special NPC Interactions

### Innkeepers

**Services:**
- Set bind point (respawn location)
- Rent rooms (if implemented)
- Sell food and drink
- Provide local information

**Commands:**
```
bind
talk innkeeper
```

### Trainers

**Services:**
- Teach new abilities (if system exists)
- Improve skills
- Class specialization
- Cost gold usually

**Requirements:**
- Minimum level
- Class restrictions
- Quest completion

### Guards

**Behavior:**
- Protect towns
- Attack aggressors
- Provide information
- May give security quests

**Tips:**
- Don't attack in towns (guards retaliate)
- Ask for directions
- Some have patrol quests

## Common NPC Commands

| Command | Description | Example |
|---------|-------------|---------|
| `talk [npc]` | Start conversation | `talk guard` |
| `[number]` | Select dialog option | `1` |
| `list` | View merchant inventory | `list` |
| `buy [item]` | Purchase item | `buy sword` |
| `sell [item]` | Sell item | `sell pelt` |
| `attack [npc]` | Engage enemy NPC | `attack goblin` |

## Troubleshooting NPC Interactions

### "NPC won't talk to me"

**Possible Causes:**
- NPC is enemy type (use attack instead)
- Already in dialog with someone else
- Quest prerequisite not met
- NPC is event-based and not available

**Solutions:**
- Check NPC type (button color)
- Wait a moment and try again
- Complete prerequisite quests
- Check quest requirements

### "Dialog options don't appear"

**Solutions:**
- Dialog may be auto-progressing
- Wait for text to fully display
- Check terminal for options
- Try talking again

### "Can't find the NPC"

**Solutions:**
- NPC may be wandering (check nearby rooms)
- Wrong location (re-read quest)
- Event-based (not spawned yet)
- Ask other players

### "Merchant won't buy my items"

**Reasons:**
- Item is quest-bound
- Item cannot be sold
- Merchant doesn't buy that type
- Try different merchant

## NPC Etiquette (Multiplayer)

**If game has multiplayer:**

**Don't:**
- Kill quest-giving NPCs (if possible)
- Monopolize merchant time
- Spam talk commands
- Grief other players' interactions

**Do:**
- Wait your turn
- Share quest information
- Help others find NPCs
- Report NPC bugs

## Next Steps

- Accept quests from NPCs: [Quest System](07-quests.md)
- Trade for better gear: [Inventory & Equipment](05-inventory-equipment.md)
- Fight enemy NPCs: [Combat System](06-combat.md)
- Learn all commands: [Commands Reference](03-commands.md)

**Talk to everyone - you never know what adventures await!**
