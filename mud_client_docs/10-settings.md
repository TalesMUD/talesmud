# Settings & Customization

Customize your TalesMUD experience with various settings and layout options.

## Accessing Settings

**Opening Settings:**
1. Look for Settings icon/button (usually gear icon)
2. Click to open Settings modal
3. Browse settings tabs
4. Make changes
5. Settings save automatically

## Settings Categories

### General Settings

Audio and general game preferences.

#### Sound Enabled

**Toggle:** On/Off

**Effect:**
- Master audio switch
- Enables/disables all game sounds
- Affects music and sound effects

**When to Disable:**
- Playing in public places
- Need quiet environment
- Audio distracting

#### Music Volume

**Range:** 0-100

**Effect:**
- Controls background music volume
- Independent from sound effects
- 0 = muted, 100 = full volume

**Recommended:**
- 30-50 for subtle background
- 70+ for immersive experience
- 0 if you prefer your own music

#### SFX Volume

**Range:** 0-100

**Effect:**
- Controls sound effects volume
- Combat sounds, UI sounds, etc.
- Independent from music

**Recommended:**
- 50-70 for balanced experience
- Higher for important audio cues
- Lower if sound effects are distracting

**Note:** Audio system framework exists but may not be fully implemented.

### Interface Settings

Visual and UI customization options.

#### Parchment Background

**Toggle:** On/Off

**Effect:**
- Adds parchment texture to room descriptions
- Enhances fantasy atmosphere
- Slight readability trade-off

**When to Enable:**
- You prefer thematic UI
- Enjoy immersive visuals
- Like the aesthetic

**When to Disable:**
- Prefer clean, minimal interface
- Need maximum readability
- Performance concerns

#### Compact Mode

**Toggle:** On/Off

**Effect:**
- Reduces spacing and padding in UI
- Shows more information on screen
- Tighter, denser layout

**When to Enable:**
- Small screen
- Want to see more at once
- Prefer information-dense interface

**When to Disable:**
- Large screen with plenty of space
- Prefer comfortable spacing
- Accessibility needs

#### Room Text Overlay

**Toggle:** On/Off (Always on for mobile)

**Effect:**
- Displays game messages as overlays on room image
- Messages fade after a few seconds
- Supports bold text formatting

**Features:**
- Auto-fade (configurable duration)
- Slide-in animation
- Appears on top of room background
- Good for important messages

**When to Enable:**
- Want cinematic feel
- Prefer messages on visuals
- Like animated notifications

**When to Disable:**
- Prefer all text in terminal
- Find overlays distracting
- Terminal-only playstyle

**Mobile Note:** Always enabled on mobile for better readability.

## Settings Persistence

**Automatic Saving:**
- All settings saved to browser localStorage
- Persist across sessions
- Version controlled (settings_v1)
- No manual save needed

**Storage Key:**
- `talesmud_settings_v1`
- Browser-specific
- Per-device storage

**Clearing Settings:**
- Clear browser data to reset
- Settings will return to defaults
- Can reconfigure anytime

## Layout Customization

### Edit Mode

**Entering Edit Mode:**
1. Click "Edit Layout" button
2. Interface switches to edit mode
3. Widgets become draggable and resizable
4. Edit toolbar appears

**In Edit Mode:**
- Drag widgets by title bars
- Resize by dragging corners/edges
- Add/remove widgets
- Rearrange interface

**Exiting Edit Mode:**
1. Click "Save Layout" button
2. Or click "Cancel" to discard changes
3. Layout saved to localStorage
4. Return to normal mode

### Moving Widgets

**Drag and Drop:**
1. Enter edit mode
2. Click and hold widget title bar
3. Drag to desired position
4. Release to place
5. Widget snaps to grid

**Tips:**
- Widgets can overlap (adjust as needed)
- Maintain important widgets visible
- Keep Action Bar accessible
- Room or Terminal should be prominent

### Resizing Widgets

**Resize Controls:**
1. Enter edit mode
2. Hover over widget corners or edges
3. Cursor changes to resize indicator
4. Click and drag to resize
5. Release when desired size reached

**Constraints:**
- Minimum size enforced
- Maximum size enforced
- Maintains aspect ratio (some widgets)
- Prevents too-small or too-large sizes

### Adding Widgets

**Add New Widget:**
1. Enter edit mode
2. Click "Add Widget" or widget menu
3. Select widget type from list
4. Widget appears in default position
5. Drag to desired location
6. Resize as needed

**Available Widgets:**
- Room Widget
- Terminal Widget
- Terminal X Widget
- Action Bar Widget
- Character Widget
- Inventory Widget
- Equipment Widget
- Quest Log Widget
- Minimap Widget

**Single-Instance Widgets:**
- Most widgets can only be added once
- Can't have two Room Widgets (pointless)
- Prevents duplicates

### Removing Widgets

**Delete Widget:**
1. Enter edit mode
2. Look for X or close button on widget
3. Click to remove
4. Widget disappears from layout
5. Can be re-added later

**Essential Widgets:**
- Action Bar (highly recommended)
- Room or Terminal (at least one!)
- Don't remove all navigation

### Default Layout

**Standard Configuration:**
- **Left Panel:** Room Widget
- **Right Panel:** Terminal Widget
- **Bottom:** Action Bar Widget
- **Side Panels:** Character, Inventory, Equipment, Quest Log, Minimap

**Resetting to Default:**
- Delete layout from localStorage
- Clear site data
- Reload page
- Default layout loads

### Layout Presets (Ideas)

**Terminal-Only Layout:**
- Large Terminal Widget (80% of screen)
- Action Bar at bottom
- Character/Inventory in small side panel

**Visual-First Layout:**
- Large Room Widget (60% of screen)
- Small Terminal X Widget (sidebar)
- Action Bar and Character visible
- Minimal terminal use

**Balanced Layout:**
- Room and Terminal side-by-side (50/50)
- Action Bar bottom
- All info widgets in panels

**Mobile-Optimized Layout:**
- Vertical stacking
- Full-width widgets
- Tab-based switching

## Terminal Customization

### Font Size (Classic Terminal)

**Options:** Small (S), Medium (M), Large (L)

**Buttons:** S, M, L in terminal header

**Effect:**
- Changes text size in xterm terminal
- Affects line count visible
- Improves readability

**Saved:** `talesmud_term_fontsize`

**Recommendations:**
- Small: For advanced users, see more text
- Medium: Balanced default
- Large: Better readability, accessibility

### Font Size (Terminal X)

**Options:** Small (S), Medium (M), Large (L)

**Buttons:** S, M, L in Terminal X header

**Effect:**
- Changes text size in custom terminal
- Independent from classic terminal
- Syntax highlighting maintained

**Saved:** `talesmud_termx_fontsize`

### Terminal Choice

**Classic Terminal:**
- Uses xterm.js
- Traditional monospace display
- Best for purists
- Standard command interface

**Terminal X:**
- Enhanced custom terminal
- Syntax highlighting
- Color-coded text
- Modern styling

**Both Available:**
- Can add both to layout
- Use whichever you prefer
- Or switch between them

## Inventory Customization

### View Mode

**Grid View:**
- Thumbnail/icon based
- Visual representation
- Easier to recognize items
- More space per item

**List View:**
- Compact text list
- See more items at once
- Faster scanning
- Less visual clutter

**Toggle:** Button in Inventory Widget header

**Saved:** `talesmud_inventory_view` (localStorage)

**Preference:**
- Use Grid for new players
- Use List for experienced players
- Try both and decide

### Category Collapse

**Collapsible Sections:**
- Equipment
- Consumables
- Quest Items
- Other

**Click to Collapse/Expand:**
- Hide categories you don't use
- Show only what you need
- Reduce scrolling

## Performance Settings

**Not Explicitly Configurable, but Tips:**

### Reduce Visual Overhead

**Disable:**
- Parchment backgrounds
- Room text overlays
- Unnecessary widgets

**Enable:**
- Compact mode
- Terminal-only layout

### Browser Optimization

**Clear Cache:**
- Remove old data
- Refresh resources
- Improve loading

**Close Other Tabs:**
- Free memory
- Better performance
- Smoother experience

**Update Browser:**
- Latest features
- Better rendering
- Bug fixes

## Accessibility Options

### Visual Accessibility

**Font Sizes:**
- Increase terminal font size (L)
- Better for visual impairment
- Reduce eye strain

**Contrast:**
- Dark theme by default (high contrast)
- Gold/amber accents visible
- Consider browser zoom

**Spacing:**
- Disable compact mode
- More comfortable spacing
- Easier to click targets

### Audio Accessibility

**Disable Sounds:**
- If audio distracts
- For sensory needs
- Personal preference

**Volume Control:**
- Adjust to comfortable levels
- Balance music and SFX
- Independent controls

## Advanced Customization

### Browser Developer Tools

**For Power Users:**

**Access localStorage:**
```javascript
// View current settings
localStorage.getItem('talesmud_settings_v1')

// View layout
localStorage.getItem('talesmud_layout_v1')

// Clear specific setting
localStorage.removeItem('talesmud_settings_v1')

// Clear all TalesMUD data
Object.keys(localStorage)
  .filter(key => key.startsWith('talesmud_'))
  .forEach(key => localStorage.removeItem(key))
```

**Warning:** Advanced users only! Can break configuration.

### Custom CSS (External)

**If using browser extensions:**
- Stylus, Stylish, etc.
- Custom styles possible
- Modify colors, fonts, spacing
- At your own risk

**Not officially supported.**

## Settings Tips

### Recommended Settings for New Players

**General:**
- Sound Enabled: On (if available)
- Music Volume: 40
- SFX Volume: 50

**Interface:**
- Parchment Background: On (immersive)
- Compact Mode: Off (comfortable)
- Room Text Overlay: On (helpful)

**Layout:**
- Default layout (Room + Terminal)
- Add Character Widget
- Add Quest Log Widget

### Recommended Settings for Experienced Players

**General:**
- Sound: Personal preference
- Volumes: As desired

**Interface:**
- Parchment Background: Off (efficiency)
- Compact Mode: On (more info)
- Room Text Overlay: Off (terminal only)

**Layout:**
- Custom layout (optimized for your flow)
- Larger terminal (typing efficiency)
- Minimap for navigation
- Inventory List view

### Recommended Settings for Mobile

**See:** [Mobile Play Guide](11-mobile.md)

**Interface:**
- Compact Mode: On
- Room Text Overlay: On (automatic)

**Layout:**
- Vertical stacking
- Tab-based interface
- Touch-optimized

## Troubleshooting Settings

### "Settings won't save"

**Possible Causes:**
- Browser blocking localStorage
- Private/incognito mode
- Browser storage full

**Solutions:**
- Check browser permissions
- Use normal browsing mode
- Clear old localStorage data
- Try different browser

### "Layout keeps resetting"

**Causes:**
- Browser clearing data automatically
- localStorage disabled
- Browser privacy settings

**Solutions:**
- Disable auto-clear
- Whitelist the site
- Adjust privacy settings

### "Interface looks broken"

**Solutions:**
- Reset to default layout
- Clear localStorage
- Reload page
- Check browser console for errors

### "Audio doesn't work"

**Check:**
- Sound Enabled toggle
- Volume sliders not at 0
- Browser allows audio
- Audio system implementation status

## Settings Summary

| Setting | Type | Default | Effect |
|---------|------|---------|--------|
| Sound Enabled | Toggle | On | Master audio switch |
| Music Volume | 0-100 | 50 | Background music level |
| SFX Volume | 0-100 | 50 | Sound effects level |
| Parchment Background | Toggle | On | Room description texture |
| Compact Mode | Toggle | Off | Dense UI layout |
| Room Text Overlay | Toggle | On | Message overlays on room |

## Next Steps

- Optimize for your device: [Mobile Play Guide](11-mobile.md)
- Customize layout for efficiency
- Experiment with settings to find your preference
- Return to [Getting Started](01-getting-started.md) with your optimized setup

**Make TalesMUD yours!**
