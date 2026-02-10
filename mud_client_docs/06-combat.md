# Combat System

Master the art of battle and emerge victorious against enemies throughout the game world.

## Combat Overview

Combat in TalesMUD is turn-based, with you and your enemies taking actions until one side is defeated or flees.

### Entering Combat

**Initiating Combat:**

**Method 1 - Button:**
1. Click red "Attack" button on enemy entity card in Room Widget

**Method 2 - Command:**
```
attack [enemy name]
```

**Example:**
```
attack goblin
attack wolf
attack bandit
```

**What Happens:**
- Combat state begins
- Your Character Widget glows red
- "IN COMBAT" badge appears
- Action Bar switches to combat commands
- Movement is restricted

### Combat Interface Changes

When combat starts, the interface adapts:

**Action Bar:**
- Standard commands (Look, Inventory, etc.) are replaced
- Combat commands appear: Attack, Defend, Flee, Status

**Character Widget:**
- Glows red to indicate combat
- Shows "IN COMBAT" status
- HP updates in real-time as damage is taken/healed

**Terminal:**
- Displays combat messages
- Shows attack results
- Reports damage dealt and taken
- Announces combat events

## Combat Commands

### Attack

**Command:** `attack`

**Effect:**
- Deal damage to the enemy
- Your primary offensive action
- Damage based on your weapon, Strength, and other factors

**When to Use:**
- Default action in most situations
- When enemy HP is high
- To finish off wounded enemies

### Defend

**Command:** `defend`

**Effect:**
- Take defensive stance
- Reduces damage from enemy attacks
- May slightly reduce your damage output

**When to Use:**
- When your HP is low
- Against powerful enemies
- To survive until help arrives or cooldowns refresh

### Flee

**Command:** `flee`

**Effect:**
- Attempt to escape combat
- May succeed or fail (chance-based)
- If successful, moves you to a random adjacent room
- If failed, enemy gets a free attack

**When to Use:**
- When HP is critically low
- Against enemies far above your level
- When outnumbered
- Strategic retreat to heal

### Status

**Command:** `status`

**Effect:**
- Displays current combat situation
- Shows your HP and enemy HP
- Lists active effects or buffs
- Provides tactical information

**When to Use:**
- Check enemy health
- Verify your HP
- See combat state details

## Combat Flow

### Turn-Based System

Combat proceeds in turns:

1. **Player Action Phase**
   - Choose your action (Attack/Defend/Flee)
   - System processes your action
   - Damage/effects applied

2. **Enemy Action Phase**
   - Enemy AI chooses action
   - Enemy attack processed
   - Damage/effects applied to you

3. **Status Update**
   - HP changes displayed
   - Combat messages shown
   - Turn counter increments

4. **Repeat**
   - Until victory, defeat, or flee

### Combat Messages

**Message Types:**

**Combat Start:**
```
You engage the goblin in combat!
```

**Attack Messages:**
```
You strike the goblin for 15 damage!
The goblin attacks you for 8 damage!
```

**Critical Hits:**
```
Critical hit! You deal 30 damage!
```

**Misses:**
```
You miss the goblin!
The goblin's attack misses you!
```

**Death:**
```
You have defeated the goblin!
The goblin has fallen!
```

### Victory Conditions

**You Win When:**
- Enemy HP reaches 0
- Enemy flees successfully
- Special victory conditions met

**You Lose When:**
- Your HP reaches 0
- Fatal game mechanic triggered

**Stalemate:**
- You flee successfully
- Combat times out (rare)

## Combat Statistics

### Damage Calculation

Your damage output depends on:

**Weapon Damage:**
- Base weapon damage value
- Weapon quality multiplier
- Weapon type effectiveness

**Character Stats:**
- Strength (for physical weapons)
- Intelligence (for magic weapons)
- Agility (affects critical hits)

**Equipment Bonuses:**
- All equipped items with damage bonuses
- Set bonuses (if applicable)
- Temporary buff effects

### Defense Calculation

Damage you take is reduced by:

**Armor:**
- Equipped armor pieces
- Total armor rating
- Armor type effectiveness

**Character Stats:**
- Constitution (damage reduction)
- Agility (dodge chance)

**Active Effects:**
- Defend command bonus
- Defensive buff potions
- Shield equipped

### Critical Hits

**Chance:**
- Base critical hit chance (varies by class)
- Agility/Dexterity bonus
- Critical attribute from equipment
- Weapon critical modifiers

**Effect:**
- Deals bonus damage (often 2x or more)
- Special visual/text indicator
- May apply special effects

## Combat Strategy

### Basic Tactics

**Against Weak Enemies:**
- Spam attack
- Don't waste potions
- Quick victories

**Against Equal Enemies:**
- Attack primarily
- Defend if HP drops below 50%
- Use potions if HP critical
- Consider flee if losing badly

**Against Strong Enemies:**
- Check their level first
- Use defend more often
- Keep HP above 50%
- Flee if outmatched
- Come back at higher level

### Health Management

**During Combat:**
- Monitor HP bar constantly
- Use healing potions when HP drops below 30%
- Don't wait until the last second
- Better to "waste" a potion than die

**HP Thresholds:**
- **75-100%** - Safe, attack freely
- **50-75%** - Cautious, watch carefully
- **25-50%** - Danger zone, use potions or defend
- **0-25%** - Critical, heal immediately or flee

### Potion Usage

**When to Use Healing Potions:**
- HP below 30-40% against tough enemies
- HP below 50% against bosses
- Before final attacks in close fights
- After combat if no enemies nearby

**How to Use in Combat:**
```
use health potion
use healing draught
use small potion
```

**Tip:** Keep potions in easy-to-access inventory slots.

### Knowing When to Flee

**Flee If:**
- Enemy level is 3+ levels above you
- Your HP drops below 20% early in fight
- Enemy has barely taken damage
- You're surprised by enemy strength
- Better to retreat and prepare

**Don't Flee If:**
- Enemy is nearly dead
- You have healing potions
- Enemy is fleeing too
- Victory is certain

## Enemy Types

### Regular Enemies

**Characteristics:**
- Standard HP and damage
- Basic AI behavior
- Fair loot drops
- Level-appropriate challenge

**Examples:** Wolves, bandits, goblins, spiders

### Elite Enemies

**Characteristics:**
- Higher HP than normal
- Increased damage output
- Better loot
- May have special abilities

**Indicators:**
- Special name formatting
- Higher level display
- Elite badge/marker

### Bosses

**Characteristics:**
- Massive HP pools
- Powerful attacks
- Unique mechanics
- Exceptional loot and XP

**Preparation:**
- Stock up on potions
- Ensure best equipment
- Consider grouping (if multiplayer)
- Study boss patterns

### Aggressive vs Passive

**Aggressive Enemies:**
- Attack you on sight
- Marked with red coloring
- Cannot avoid combat
- Must defeat or flee

**Passive Enemies:**
- Don't attack unless provoked
- Neutral coloring
- You must initiate combat
- Can be avoided

## Combat Rewards

### Experience Points

**XP Gain:**
- Based on enemy level
- Higher-level enemies give more XP
- Defeating enemies your level or higher is most efficient
- Boss kills grant bonus XP

**XP Penalties:**
- Enemies far below your level give reduced XP
- May give no XP if too low-level

### Loot Drops

**What Enemies Drop:**
- Gold/currency
- Equipment (weapons, armor)
- Consumables (potions)
- Crafting materials
- Quest items (if needed)

**Loot Quality:**
- Based on enemy type and level
- Elites drop better loot
- Bosses drop rare/legendary items
- Random chance involved

**Auto-Pickup vs Manual:**
- Some items auto-pickup (gold)
- Equipment appears on ground
- Use pickup command/button to collect

## Death and Respawn

### When You Die

**What Happens:**
1. Combat ends
2. Death message displayed
3. Respawn prompt appears
4. You return to bind point

**Penalties:**
- May lose some experience
- May drop some gold
- May have temporary stat debuff
- Equipment stays with you (usually)

### Respawning

**Respawn Location:**
- Your bind point (set with `bind` command)
- Default: Starting area
- Can be reset at inns or temples

**After Respawn:**
- Full HP restored
- Back in safe location
- Can re-equip and prepare
- Can return to challenge enemy again

## Advanced Combat

### Combat Buffs

**Temporary Boosts:**
- Strength potions (+damage)
- Defense elixirs (+armor)
- Speed tonics (+agility)
- Resistance potions (-damage taken)

**Duration:**
- Time-limited (e.g., 5 minutes)
- Effect shown in status
- Stack with equipment (usually)

### Equipment Swaps

**Situational Gear:**
- Fire resistance armor vs fire enemies
- Magic weapons vs magic-resistant foes
- Heavy armor vs physical enemies

**Mid-Combat Changes:**
- Usually not allowed
- Plan equipment before fight
- Swap between battles

### Combat Analytics

**Tracking Performance:**
- Damage per turn
- Average fight duration
- Potion usage rate
- Win/loss ratio

**Improvement:**
- Better equipment = more damage, less taken
- Higher stats = easier fights
- Better tactics = fewer potions used

## Enemy Health Bars

### Visual Indicators

Enemies display health bars above their entity cards:

**Colors:**
- **Green** - Healthy (75-100% HP)
- **Orange/Yellow** - Wounded (25-75% HP)
- **Red** - Critical (0-25% HP)

**Tactical Use:**
- Gauge when to flee
- Know when victory is near
- Decide whether to use potions

## Group Combat (Multiplayer)

If the game supports multiplayer combat:

**Fighting Together:**
- Combined damage
- Shared XP and loot
- Coordinated tactics (tank, healer, DPS roles)
- Communication essential

**Benefits:**
- Defeat stronger enemies
- Faster leveling
- Safer exploration
- Social gameplay

## Combat Tips and Tricks

### General Tips

1. **Always carry healing potions** - Minimum 5 at all times
2. **Fight enemies at your level** - Best XP and loot balance
3. **Upgrade weapons first** - Damage > Defense for efficiency
4. **Learn enemy patterns** - Some enemies telegraph attacks
5. **Rest between fights** - Let HP regenerate naturally
6. **Don't be greedy** - Flee if outmatched
7. **Save before bosses** - Set bind point nearby

### Common Mistakes

**Avoid These:**
- Fighting without potions
- Ignoring HP warnings
- Taking on enemies far above your level
- Forgetting to equip better gear
- Not fleeing when losing
- Using all potions on one fight

### Efficiency Tips

**Leveling Fast:**
- Fight enemies slightly above your level
- Complete combat quests
- Use all equipment slots
- Don't waste time on trivial enemies

**Staying Alive:**
- Keep HP above 50% between fights
- Rest in safe locations
- Stock up in towns before exploring
- Bind near dangerous areas

## Combat Commands Summary

| Command | Effect | Best Use |
|---------|--------|----------|
| `attack` | Deal damage | Primary offensive action |
| `defend` | Reduce damage taken | Low HP, strong enemy |
| `flee` | Escape combat | Losing badly, need to retreat |
| `status` | Check combat state | Verify HP, assess situation |
| `use [potion]` | Heal or buff | HP critical, need boost |

## Next Steps

- Optimize your loadout in [Inventory & Equipment](05-inventory-equipment.md)
- Level up efficiently with [Quest System](07-quests.md)
- Understand your stats in [Character System](04-character.md)
- Learn all commands in [Commands Reference](03-commands.md)

**Now go forth and conquer!**
