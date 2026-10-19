# Quest System

Complete quests to earn experience, gold, and powerful rewards while uncovering the stories of the game world.

## Quest Overview

Quests are missions given by NPCs that involve objectives to complete. They're the primary way to progress through the game's story and earn substantial rewards.

## Finding Quests

### Quest Givers

**Where to Find Them:**
- Town NPCs (often marked with special indicators)
- Travelers on roads
- Important characters in story areas
- Hidden NPCs in remote locations

**Identifying Quest Givers:**
- NPCs with special markers or symbols
- NPCs mentioned in area descriptions
- Characters who greet you when you enter
- Try talking to everyone!

### Starting a Quest

**Steps:**

1. **Talk to the NPC**
   - Click blue "Talk" button on NPC card
   - Or type: `talk [npc name]`

2. **Read the Dialog**
   - NPC explains the situation
   - Describes what needs to be done
   - May provide background story

3. **Accept the Quest**
   - Select the "Accept" dialog option (usually option 1 or 2)
   - Type the number of the option

4. **Confirmation**
   - Quest accepted notification appears
   - Quest added to your Quest Log
   - Objectives listed

**Example:**
```
talk village elder
> 1 (to accept the quest)
```

## Quest Log

### Viewing Your Quests

**Methods:**
1. **Quest Log Widget** - Always-visible panel showing all quests
2. **Quest Log Overlay** - Press Q or click quest icon to open full screen view
3. **Quest Notifications** - Toast notifications appear when quests update

### Quest Log Features

#### Search Quests 🔍
Find quests quickly by typing in the search bar:
- Searches quest names, descriptions, and objectives
- Real-time filtering as you type
- Click × to clear search
- Case-insensitive matching

**Example:**
- Type "goblin" to find all goblin-related quests
- Type "collect" to find collection objectives

#### Filter Quests 📂
Organize your quest log with filters:

**Category Filter:**
- All Types (default)
- Main - Story quests
- Side - Optional quests
- Daily - Repeatable daily quests

**Visibility Toggles:**
- ☑ Completed - Show/hide completed quests
- ☐ Abandoned - Show/hide abandoned quests

#### Sort Quests 📊
Change quest order with sort options:
- **Sort: Status** - Groups by Active → Completed → Abandoned
- **Sort: Name** - Alphabetical order
- **Sort: Level** - Lowest to highest level
- **Sort: Type** - Groups by category

#### Pin Important Quests 📌
Keep track of priority quests:
- Click "Pin Quest" button in quest details
- Pinned quests appear in special "📌 Pinned" section at top
- Pin up to 5 quests at once
- Pinned quests have golden pulsing indicator
- Click "Unpin Quest" to remove

**Why Pin Quests?**
- Focus on main story quests
- Track time-limited quests
- Remember which quests to complete first
- Quick access to important objectives

### Quest Log Sections

#### 📌 Pinned Quests (if any)
- Your priority quests
- Always at the top
- Special golden highlighting
- Quick access to objectives

#### Active Quests
- Quests you're currently working on
- Can have multiple active at once
- Shows progress on objectives (X/Y format)
- Click to expand for full details
- Pin button to mark as priority
- Abandon button to drop quest

#### Completed Quests
- Quests you've finished
- Historical record
- Click to expand and review
- Shows completion date
- Displays rewards earned

### Quest Information

Each quest displays:

**Quest Name:**
- Title of the quest
- Often hints at the objective

**Quest Badges:**
- **Level Badge** - Shows recommended level (e.g., "L5")
- **Category Badge** - Quest type (Main/Side/Daily)
  - 🟠 Main - Amber/Orange color
  - 🔵 Side - Blue color
  - 🟣 Daily - Purple color

**Description:**
- Background story
- Why this quest matters
- Context for objectives
- Expandable in quest details

**Objectives:**
- List of tasks to complete
- Progress tracking (e.g., "2/5 completed")
- Real-time updates as you progress
- ✓ Checkmarks when finished
- Multiple objectives per quest

**Rewards Preview:**
- 💫 Experience Points (XP)
- 💰 Gold
- 🎁 Items (count shown)
- Visible before completing quest

**Example Quest Display:**
```
📌 Quest: The Wolf Problem [L3] [Main]
Description: Wolves have been attacking travelers.
Help protect the road by dealing with them.

Objectives:
✓ Talk to Village Elder (1/1)
⬜ Defeat wolves (3/5)
⬜ Return to Village Elder (0/1)

Rewards:
💫 500 XP
💰 50 Gold
🎁 1 item(s)

[Pin Quest] [Abandon Quest]
```

### Quest History & Statistics 📊

Click the 📊 button in the quest log header to view your quest statistics:

**Quest Summary:**
- Total Quests - All quests you've encountered
- Active - Currently in progress
- Completed - Successfully finished
- Completion Rate - Your success percentage

**Total Rewards Earned:**
- 💫 Cumulative XP from all completed quests
- 💰 Total gold earned from quests
- Tracks your overall quest progress

**Category Breakdown:**
- Main quests completed
- Side quests completed
- Daily quests completed

**Quest Achievements:** See [Quest Achievements](#quest-achievements) section below

### Quest Achievements 🏆

Unlock achievements by completing quests:

#### Progression Achievements
1. **🏆 First Steps** - Complete your first quest
2. **🏆 Quest Novice** - Complete 5 quests
3. **🏆 Quest Veteran** - Complete 10 quests
4. **🏆 Quest Master** - Complete 25 quests

#### Category Achievements
5. **🏆 Story Seeker** - Complete 5 main quests
6. **🏆 Side Quest Hero** - Complete 10 side quests
7. **🏆 Daily Devotee** - Complete 5 daily quests

#### Special Achievements
8. **🏆 Completionist** - Achieve 100% completion rate (minimum 5 quests)

**Achievement Display:**
- Unlocked achievements shown with 🏆 gold trophy
- Locked achievements shown with 🔒 and grayed out
- Click "Locked Achievements" to see what you can unlock next
- Achievements tracked automatically
- View in Quest History panel (📊 button)

## Quest Categories

Quests are organized into three main categories:

### Main Quests 🟠
- **Badge Color:** Amber/Orange
- **Purpose:** Advance the main story
- **Rewards:** Best rewards, significant XP
- **Recommended:** Do these first
- **Impact:** Unlock new areas and features

### Side Quests 🔵
- **Badge Color:** Blue
- **Purpose:** Optional content, side stories
- **Rewards:** Good rewards, moderate XP
- **Recommended:** Do when available
- **Impact:** Extra gear, gold, and lore

### Daily Quests 🟣
- **Badge Color:** Purple
- **Purpose:** Repeatable quests for daily rewards
- **Rewards:** Consistent XP and gold
- **Recommended:** Do daily for steady progress
- **Impact:** Regular income and experience

**Filtering by Category:**
- Use category dropdown in Quest Log
- Select Main/Side/Daily to filter
- Helps focus on specific quest types
- "All Types" shows everything

## Quest Types

### Kill Quests

**Objective:** Defeat a certain number of enemies

**Example:**
- "Defeat 5 wolves"
- "Kill 10 bandits"
- "Eliminate the goblin chief"

**Tips:**
- Track progress in Quest Log
- Enemies must be in specific area (usually)
- Bosses often count for more

### Collection Quests

**Objective:** Gather specific items

**Example:**
- "Collect 10 wolf pelts"
- "Gather 5 rare herbs"
- "Find 3 ancient artifacts"

**Tips:**
- Items may drop from enemies
- Items may be found in specific locations
- Check ground items carefully

### Delivery Quests

**Objective:** Take item to another NPC

**Example:**
- "Deliver letter to the mayor"
- "Bring medicine to the healer"

**Tips:**
- Item appears in inventory (quest item)
- Travel to specified location
- Talk to target NPC to deliver

### Exploration Quests

**Objective:** Discover new locations

**Example:**
- "Find the hidden cave"
- "Explore the ancient ruins"
- "Reach the mountain peak"

**Tips:**
- Use minimap to track progress
- Read room descriptions for clues
- May unlock new areas

### Dialog Quests

**Objective:** Talk to specific NPCs

**Example:**
- "Speak with the three elders"
- "Gather information from townspeople"

**Tips:**
- NPCs may be in different locations
- Dialog options matter
- Pay attention to names

### Compound Quests

**Objective:** Multiple different tasks

**Example:**
- "Defeat 5 bandits AND collect their insignia"
- "Find the merchant AND escort him to town"

**Tips:**
- All objectives must be completed
- Order may matter
- Check Quest Log frequently

## Quest Objectives

### Progress Tracking

**Counters:**
- Shows current vs required (e.g., "3/10")
- Updates automatically when you make progress
- Visual checkmarks when complete

**Real-Time Updates:**
- Quest Log updates immediately
- Notifications appear for major milestones
- No need to check manually constantly

### Multiple Objectives

Some quests have sequential objectives:

**Example:**
1. "Talk to the guard" (must complete first)
2. "Defeat bandits" (unlocks after step 1)
3. "Return to guard" (unlocks after step 2)

**Others have parallel objectives:**
- "Defeat 5 wolves AND collect 5 pelts"
- Both can be done simultaneously
- Quest completes when all are done

## Completing Quests

### Finishing Objectives

**When all objectives show checkmarks:**
1. Quest is ready to turn in
2. Return to the quest giver (usually)
3. Talk to them again
4. Select completion dialog option

**Example:**
```
talk village elder
> 1 (complete quest)
```

### Quest Rewards

**Common Rewards:**

**Experience Points:**
- Large XP bonus
- Often equivalent to many enemy kills
- Scales with quest level
- May grant level up!

**Gold:**
- Currency reward
- Buy better equipment
- Save for expensive items

**Items:**
- Equipment (weapons, armor)
- Consumables (potions)
- Quest-specific unique items
- May be rare or legendary quality

**Other:**
- Reputation increases (if system exists)
- Unlock new quests
- Access to new areas
- Story progression

### Reward Notification

When you complete a quest:

**Toast Notification Appears:**
```
Quest Completed: The Wolf Problem

Rewards:
- 500 XP
- 50 Gold
- Iron Sword
```

**Automatic:**
- Rewards added to character
- XP applied
- Gold added
- Items go to inventory

## Quest Notifications

Quest notifications appear in the top-right corner when quests update.

### Notification Types

**Quest Accepted:**
```
Quest Name
Quest accepted
```
- Appears when you accept a quest
- Border: Amber/orange
- Auto-dismisses after 5 seconds

**Quest Progress:**
```
Quest Name
Objective complete: [objective description]
```
- Shows when you complete an objective
- Border: Blue
- Updates on major milestones
- Auto-dismisses after 5 seconds

**Quest Completed:**
```
Quest Name
Quest completed!
```
- Appears when all objectives are done
- Border: Green
- Celebration moment
- Auto-dismisses after 5 seconds

### Notification Interactions

**Click to View:**
- Click anywhere on a notification
- Opens quest in Quest Log
- Automatically dismisses notification
- Quick way to check quest details

**Manual Dismiss:**
- Click the × button on the right
- Immediately removes notification
- Smooth slide-out animation
- No need to wait for auto-dismiss

**Hover Effects:**
- Notifications highlight when you hover
- Shows they're interactive
- Slight movement animation indicates clickability

## Quest Strategies

### Efficient Questing

**Tips for Fast Progression:**

1. **Accept Multiple Quests**
   - Take all available quests in an area
   - Complete objectives simultaneously
   - Return to turn in together

2. **Group Similar Quests**
   - "Kill wolves" + "Collect wolf pelts" = Same enemies
   - Do both at once for efficiency

3. **Read Carefully**
   - Objectives sometimes overlap
   - Locations may be near each other
   - Plan efficient routes

4. **Use the Minimap**
   - Track where you've been
   - Avoid getting lost
   - Find your way back to quest givers

### Quest Prioritization

**Using Quest Pinning:**
1. **Pin Main Story Quests** - Keep story progression visible
2. **Pin Time-Limited Quests** - Daily quests or events
3. **Pin Gear Upgrade Quests** - Equipment reward quests
4. **Pin Quest Chain Ends** - Track multi-part quest progress
5. **Maximum 5 Pins** - Choose wisely

**Filtering for Focus:**
- Filter by "Main" to focus on story
- Filter by "Side" for completion runs
- Filter by "Daily" for daily routine
- Sort by "Level" to find appropriate quests

**Do First:**
- Pinned quests (your priorities)
- Main story quests (🟠 badge, best rewards)
- Quests at your level
- Quests with equipment rewards
- Quests that unlock new areas

**Do Later:**
- Side quests below your level (less XP)
- Collection quests (can be tedious)
- Delivery quests to far locations (save for travel)

**Stack and Complete:**
- Use search to find similar quests (e.g., search "wolf")
- Take all quests in a town
- Pin the most important ones
- Go to quest areas
- Complete all objectives
- Return to town
- Turn in all at once

### Quest Chains

Some quests are part of chains:

**Example:**
1. "The Wolf Problem" (Part 1)
2. "The Alpha Wolf" (Part 2) - Unlocks after Part 1
3. "The Cure" (Part 3) - Unlocks after Part 2

**Benefits:**
- Tell longer stories
- Better rewards at chain end
- Progressive difficulty
- Character development

**Tips:**
- Complete chains in order
- Don't abandon mid-chain
- Final quest usually gives best reward

## Quest Management

### Active Quest Limit

Some games limit active quests:
- May have maximum (e.g., 10 active)
- Abandon quests to free space (if needed)
- Prioritize important quests

### Abandoning Quests

If you need to drop a quest:

**How to Abandon:**
1. Open the quest in Quest Log
2. Click to expand quest details
3. Click red "Abandon Quest" button
4. Confirm the abandonment dialog

**What Happens:**
- Quest removed from active list
- Moves to "Abandoned" section (if "Show Abandoned" is checked)
- Progress is lost
- Can usually re-accept from original NPC
- Quest items may be removed from inventory

**When to Abandon:**
- Quest too difficult for your current level
- Not interested in the rewards
- Accidentally accepted wrong quest
- Stuck and can't complete
- Want to focus on other quests

**Tips:**
- Don't abandon quest chains mid-way
- Check rewards before abandoning
- Consider coming back when higher level instead
- Some quests may have time limits

### Quest Items

Items for quests appear in inventory:

**Characteristics:**
- Marked as "Quest Item"
- Cannot be dropped or sold
- Automatically used when appropriate
- Removed when quest completes or is abandoned

**Storage:**
- Don't count toward inventory limit (usually)
- Separate category in inventory
- Clearly marked

## Quest Difficulty

### Level Recommendations

Quests often have recommended levels:

**At Your Level:**
- Balanced challenge
- Appropriate XP reward
- Good difficulty

**Above Your Level:**
- More difficult
- May need better gear
- Consider grouping or leveling first
- Better rewards

**Below Your Level:**
- Very easy
- Reduced XP (or none)
- Quick completion
- Mainly for story/completion

### Quest Colors (If Implemented)

Some games color-code quests:

- **Gray** - Trivial, no XP
- **Green** - Easy, low XP
- **Yellow** - At level, normal XP
- **Orange** - Hard, good XP
- **Red** - Very hard, great XP

## Story and Lore

### Reading Quest Text

**Why It Matters:**
- Understand game world
- Enjoy narrative
- Find clues for objectives
- Immersion

**When to Skip:**
- Repeating quests you've done
- Pure grind/collection quests
- If you're just focused on mechanics

### Dialog Choices

Some quests have dialog choices:

**May Affect:**
- Quest outcomes
- Rewards received
- NPC reactions
- Story branches
- Reputation

**Tip:** Choose thoughtfully if you care about story!

## Quest Tips and Tricks

### General Tips

1. **Talk to everyone** - Many NPCs have quests
2. **Use Search** - Type keywords to find specific quests quickly
3. **Pin Important Quests** - Keep priorities visible at top
4. **Filter by Category** - Focus on Main/Side/Daily as needed
5. **Check Quest History** - Track your overall progress and achievements
6. **Read Rewards Before Starting** - Know what you'll earn
7. **Sort by Level** - Find quests appropriate for your character
8. **Click Notifications** - Quick way to open quest details
9. **Group quests by area** - Use search to find location-based quests
10. **Keep inventory space** - Room for quest items and rewards

### Advanced Quest Features

**Quest Search Tips:**
- Search by enemy type: "goblin", "wolf", "bandit"
- Search by action: "collect", "defeat", "deliver"
- Search by location: "marsh", "forest", "cave"
- Search by item: "herb", "sword", "letter"

**Quest History Features:**
- Track completion rate percentage
- See total XP and gold earned from quests
- View category breakdown (Main/Side/Daily)
- Monitor achievement progress
- Review completed quest count

**Quest Organization:**
- Pin up to 5 most important quests
- Filter out completed quests when focusing on active
- Show abandoned quests to see what you've dropped
- Sort by different criteria for different workflows

### Common Mistakes

**Avoid:**
- Accepting quests without reading
- Ignoring Quest Log
- Not checking quest levels
- Abandoning quests needlessly
- Missing quest turn-ins
- Selling quest items (usually can't, but be careful)

### Maximizing Rewards

**Best Practices:**
- Do quests at or slightly above your level
- Complete quest chains fully
- Look for quests with equipment rewards
- Stack quests before completing
- Don't outlevel quest areas (reduced XP)

## Quest-Related Commands

| Command | Action |
|---------|--------|
| `talk [npc]` | Start quest conversation |
| `[number]` | Select dialog/quest option |
| `abandon [quest name]` | Abandon an active quest |
| Open Quest Log Widget | Click quest icon or view always-visible widget |
| Click 📊 button | Open Quest History & Statistics panel |
| Type in search bar | Filter quests by keywords |
| Click notification | Open quest details from notification |

## Quest Log Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| **Q** | Toggle Quest Log overlay (if implemented) |
| **Esc** | Close Quest Log overlay |
| **Click quest** | Expand/collapse quest details |
| **Click 📌 pin button** | Pin/unpin quest |
| **Click × dismiss** | Dismiss notification |

## Troubleshooting

### Quest Won't Complete

**Check:**
- All objectives have checkmarks (✓)
- You're talking to correct NPC
- You're selecting right dialog option
- Quest items still in inventory
- Quest hasn't been abandoned

**Quick Fix:**
- Expand quest in Quest Log
- Verify all objectives show (X/X) completion
- Look for completion requirements in description

### Lost Track of Quest

**Solutions:**
1. **Use Quest Search** - Type quest name or keywords
2. **Check Quest Log** - Open and expand the quest
3. **Read Description** - Quest details explain objectives
4. **Pin the Quest** - Keep it visible at top
5. **Check Objectives** - See what's remaining

**Filter Tips:**
- Make sure quest isn't filtered out
- Check "Show Completed" if looking for finished quest
- Check "Show Abandoned" if you dropped it

### Can't Find Quest Objective

**Tips:**
- Read objective description carefully
- Use search to find similar completed quests for hints
- Check quest category badge (Main quests often in story areas)
- Expand quest to see full description
- Use minimap to navigate to quest areas

### Too Many Active Quests

**Solutions:**
1. **Use Filters** - Hide completed quests
2. **Pin Important Ones** - Keep 5 priorities visible
3. **Abandon Unnecessary** - Drop quests you won't do
4. **Sort by Level** - Focus on appropriate difficulty
5. **Use Categories** - Filter Main/Side/Daily

### Quest Search Not Finding Quest

**Check:**
- Spelling of search terms
- Try searching for part of the name
- Search for objective keywords instead
- Clear search and browse manually
- Quest might be in filtered category

### Achievement Not Unlocking

**Verify:**
- Check exact requirements in Quest History (📊)
- Some achievements need specific quest categories
- Completionist needs minimum 5 quests
- Progress may be delayed, refresh Quest History

## Next Steps

- Level up faster with [Combat System](06-combat.md)
- Earn gold for better gear in [Inventory & Equipment](05-inventory-equipment.md)
- Navigate efficiently with [Exploration & Minimap](08-exploration.md)
- Talk to NPCs using [NPC Interactions](09-npcs.md)

**Happy questing!**
