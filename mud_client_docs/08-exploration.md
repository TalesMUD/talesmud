# Exploration & Minimap

Navigate the vast game world effectively using the minimap and exploration tools.

## Exploration Basics

### Moving Through the World

**Cardinal Directions:**
- **North** - Primary direction
- **South** - Opposite of north
- **East** - Right on compass
- **West** - Left on compass

**Vertical Movement:**
- **Up** - Stairs, ladders, slopes ascending
- **Down** - Descending paths

**Special Exits:**
- Custom named exits (e.g., "enter portal", "climb tree")
- Appear as buttons when available

### Room Structure

Each location (room) contains:

**Room Name:**
- Centered at top of Room Widget
- Distinctive title for the location
- Helps identify where you are

**Room Description:**
- Detailed text about surroundings
- Often contains clues
- Mentions notable features
- Describes atmosphere

**Room Image:**
- Visual representation (pixelated background)
- Sets the mood
- Unique to each area

**Entities:**
- NPCs present in room
- Enemies that may attack
- Merchants for trading

**Items:**
- Objects on the ground
- Can be picked up
- Quest items or loot

**Exits:**
- Available directions to travel
- Shown as buttons in Action Bar
- Only functional exits displayed

## Using the Minimap

The Minimap Widget is your navigation tool for tracking explored areas.

### Minimap Features

**What It Shows:**
- Rooms you've visited
- Your current location (highlighted)
- Room connections (exit lines)
- Room layout on current floor

**What It Doesn't Show:**
- Unexplored rooms
- Rooms on different floors/levels
- Enemy positions
- NPCs or items

### Understanding the Minimap

**Room Display:**
- **Current Room** - Bright/distinct color highlight
- **Visited Rooms** - Shown with normal colors
- **Nearby Rooms** - Full opacity
- **Distant Rooms** - Faded (up to 5 rooms away shown)

**Connections:**
- Lines between rooms show exits
- Cardinal directions create grid pattern
- Special exits may not appear on minimap

**Floor-Based View:**
- Only shows current Z-level (floor)
- Going up/down clears the view
- Minimap shows new floor layout

### Minimap Coordinates

The minimap automatically calculates room positions:

**How It Works:**
- Starting room is set as origin (0, 0, 0)
- North increases Y coordinate
- South decreases Y coordinate
- East increases X coordinate
- West decreases X coordinate
- Up increases Z coordinate
- Down decreases Z coordinate

**Inference:**
- System calculates room positions from exits
- Creates grid-based map
- Handles complex layouts

### Minimap Hover Information

**Mouse Over Rooms:**
- Shows room name in tooltip
- Quick reference without traveling
- Helps plan routes

### Minimap Limitations

**Important Notes:**

1. **Session-Based**
   - Resets when you reload the page
   - Uses browser session storage
   - Not permanently saved

2. **Cardinal Directions Only**
   - Special exits don't create map connections
   - Complex areas may not map perfectly
   - Teleportation breaks mapping

3. **Current Floor Only**
   - Multi-level dungeons require mental tracking
   - Minimap clears when changing floors

4. **Visual Only**
   - Doesn't provide navigation commands
   - Reference tool, not autopilot

## Exploration Strategies

### Systematic Exploration

**Methodical Approach:**

1. **Pick a Direction**
   - Choose north, south, east, or west
   - Stick with it until dead end

2. **Map Thoroughly**
   - Visit every room
   - Check all exits
   - Note important locations

3. **Mark Key Locations**
   - Towns (merchants, quest givers)
   - Dangerous areas (high-level enemies)
   - Quest objectives
   - Useful resources

4. **Return to Hub**
   - Use main town as base
   - Venture out from there
   - Return to resupply

**Pattern:**
```
Start → North → North → East → South → West → Return
```

### Using Breadcrumbs

**Mental Mapping:**
- Remember the path you took
- Count turns
- Note distinctive rooms
- Use room names as landmarks

**Example:**
```
Town → North → Forest Path → East → River Crossing → North → Cave Entrance
```

To return: South → West → South

### Efficient Pathfinding

**Shortest Routes:**
- Use minimap to visualize paths
- Count steps between locations
- Find shortcuts
- Avoid unnecessary detours

**Quick Navigation:**
- Learn major routes
- Use cardinal directions for speed
- Type directions faster than clicking buttons

**Example:**
Fast path from Town to Cave:
```
n
n
e
n
```

## Navigation Tips

### Avoiding Getting Lost

**Best Practices:**

1. **Use the Minimap**
   - Check frequently
   - Reference current location
   - Plan return route

2. **Remember Key Rooms**
   - "Four-way intersection"
   - "Ancient oak tree"
   - "Stone bridge"
   - These are landmarks

3. **Look Command**
   - Type `look` to refresh room info
   - Re-read descriptions
   - Confirm location

4. **Bind Point**
   - Set bind in safe locations
   - Provides emergency return (via death)
   - Not ideal but works in crisis

### When You're Lost

**Solutions:**

1. **Check Minimap**
   - Find your position
   - Trace path back to known location
   - Look for connecting rooms

2. **Systematic Search**
   - Try all exits
   - Map current area
   - Find familiar room

3. **Ask for Help**
   - Use `who` to see other players
   - Ask in chat (if multiplayer)
   - Request directions

4. **Return to Bind**
   - Last resort: die intentionally
   - Respawn at bind point
   - Lose progress but not lost anymore

**Prevention:**
- Don't explore when low HP
- Stay aware of path taken
- Use minimap constantly

## Area Types

### Towns and Cities

**Characteristics:**
- Safe zones (no combat)
- Merchants for buying/selling
- Quest NPCs
- Inns for binding
- Multiple exits to other areas

**Key Services:**
- Equipment shops
- Potion vendors
- Quest givers
- Trainers (if class system exists)

**Navigation:**
- Usually central hub
- Multiple connections to wilderness
- Easy to find (near starting area)

### Wilderness

**Characteristics:**
- Open areas between towns
- Mix of safe and dangerous rooms
- Random enemy encounters
- Natural features (forests, plains, hills)

**Dangers:**
- Wandering enemies
- Can get lost easily
- Weather/environmental effects (if implemented)

**Tips:**
- Stay on paths when possible
- Watch for enemy levels
- Keep track of direction

### Dungeons

**Characteristics:**
- Enclosed spaces (caves, ruins, castles)
- Higher enemy density
- Valuable loot
- Boss encounters
- Multiple floors/levels

**Challenges:**
- Easy to get lost
- All enemies often aggressive
- Minimap resets per floor
- May need keys or quest items

**Preparation:**
- Stock up on potions
- Ensure good equipment
- Set bind point nearby
- Map carefully

### Special Areas

**Quest Locations:**
- Specific to quest objectives
- May be hidden or hard to find
- Often contain unique items or NPCs

**Hidden Areas:**
- Secret rooms or passages
- Special commands to access
- Valuable rewards
- Easter eggs

## Environmental Clues

### Reading Room Descriptions

**Look For:**

**Exits Mentioned:**
```
"A path leads north into the forest. To the east,
you see a cave entrance."
```
→ North and East exits available

**Hidden Exits:**
```
"Behind the waterfall, you notice a gap."
```
→ May need special command: `enter gap`

**Dangers:**
```
"The air feels cold and hostile here."
```
→ Warning: dangerous area ahead

**Quest Clues:**
```
"An old man sits by the fire."
```
→ Potential quest giver

### Special Room Actions

Some rooms have interactive elements:

**Common Actions:**
- `pull lever` - Activate mechanism
- `push button` - Trigger event
- `read sign` - Get information
- `examine statue` - Discover secrets
- `open door` - Access new area

**Finding Them:**
- Read room descriptions carefully
- Look for mentioned objects
- Experiment with commands
- Check Action Bar for room actions

## Tracking Your Journey

### Manual Methods

**Mental Notes:**
- Remember important locations
- Recall quest destinations
- Track merchant locations

**External Tools:**
- Draw paper maps (old school!)
- Take screenshots
- Write notes in separate file

### In-Game Tracking

**Minimap:**
- Visual reference for explored areas
- Automatic position tracking
- Connection mapping

**Quest Log:**
- Quest locations mentioned
- Return destinations noted
- Progress tracking

**Bind Points:**
- Strategic save locations
- Quick return points

## Discovery and Secrets

### Finding Hidden Content

**Exploration Rewards:**
- Discovery XP (if implemented)
- Hidden treasure
- Secret quests
- Easter eggs
- Unique items

**How to Find:**
- Explore thoroughly
- Try unusual directions
- Examine everything
- Talk to all NPCs
- Read all signs and books
- Experiment with commands

### Off-the-Beaten-Path

**Remote Locations:**
- Far from main paths
- Require multiple direction changes
- Often have better rewards
- Fewer players find them

**Example:**
```
Town → North x5 → East x3 → Up → Secret Mountain Cave
```

## Travel Efficiency

### Speed Tips

**Fast Movement:**
- Type direction shortcuts: `n`, `s`, `e`, `w`
- Chain commands (if supported): `n; e; n`
- Use buttons for occasional travel
- Type for frequent routes

**Route Optimization:**
- Learn shortest paths
- Avoid unnecessary rooms
- Use special exits when available
- Plan multi-stop journeys

**Time Management:**
- Group quests by location
- Complete all objectives in area before leaving
- Return to towns less frequently (but stay supplied)

### Waypoint System (If Implemented)

Some games offer fast travel:
- Unlocked waypoints
- Instant travel between discovered locations
- May cost gold or have cooldown
- Saves significant time

**Check for:**
- Fast travel NPCs
- Teleportation devices
- Portal networks
- Recall spells

## Exploration Checklist

### New Area Checklist

When entering a new area:

- [ ] Read room description thoroughly
- [ ] Check available exits
- [ ] Look for NPCs (quest givers, merchants)
- [ ] Check for ground items
- [ ] Note enemy types and levels
- [ ] Set bind point if safe area
- [ ] Map systematically
- [ ] Look for special interactions

### Before Long Journey

- [ ] Set bind point in safe location
- [ ] Stock healing potions (10+)
- [ ] Check equipment condition
- [ ] Review quest objectives
- [ ] Plan route using minimap
- [ ] Ensure inventory space
- [ ] Note return path

## Exploration Commands

| Command | Description |
|---------|-------------|
| `look` | Examine current room |
| `examine [object]` | Look closely at something |
| `n`, `s`, `e`, `w` | Move in cardinal directions |
| `up`, `down` | Vertical movement |
| `bind` | Set respawn point |
| View Minimap Widget | Check explored areas |

## Troubleshooting Navigation

### "I can't find the quest location"

**Solutions:**
1. Re-read quest description
2. Check minimap for unexplored areas
3. Ask quest giver for hints (talk again)
4. Explore systematically
5. Look for special exits mentioned in descriptions

### "Minimap isn't showing correctly"

**Causes:**
- Special exits don't map
- Teleportation broke tracking
- Complex non-grid layout

**Solutions:**
- Use manual tracking
- Focus on room names
- Ask other players
- Explore multiple times to memorize

### "I keep going in circles"

**Solutions:**
1. Look at minimap
2. Choose unexplored direction
3. Count steps
4. Use different strategy (always turn right, etc.)
5. Mark rooms mentally ("been here")

## Next Steps

- Use navigation for [Quest System](07-quests.md) objectives
- Plan efficient [Combat System](06-combat.md) grinding routes
- Find merchants for [Inventory & Equipment](05-inventory-equipment.md) needs
- Talk to NPCs using [NPC Interactions](09-npcs.md) skills

**Adventure awaits! Start exploring!**
