# Interactive World Map Implementation

## Overview
Replaced the static GraphViz-generated PNG world map with an interactive 2D grid-based map using Svelte Flow.

## Changes Made

### Backend Changes

#### 1. New Graph Data Endpoint (`pkg/server/handler/worldrenderer.go`)
- Added `RenderGraphData()` method that returns JSON graph data
- Created data structures:
  - `GraphNode`: Contains room ID, name, description, coordinates, area info
  - `GraphEdge`: Contains exit information with cardinal direction flags
  - `GraphData`: Complete graph structure with nodes and edges

#### 2. Position Calculation Algorithm
- Implemented `calculatePositions()` function that:
  - Uses existing `room.Coords` (X, Y, Z) where available
  - Auto-positions rooms without coordinates using BFS algorithm
  - Infers positions from cardinal direction exits (north/south/east/west)
  - Places unconnected rooms in a separate grid area
  - Scales coordinates by 100 pixels per grid unit for Svelte Flow

#### 3. API Route Registration (`pkg/server/server.go`)
- Added new endpoint: `GET /api/world/graph`
- Returns JSON instead of base64-encoded PNG

### Frontend Changes

#### 4. Package Dependencies (`public/app/package.json`)
- Added `@xyflow/svelte": "^0.1.19"` for interactive flow diagrams

#### 5. API Client (`public/app/src/api/world.js`)
- Added `getWorldGraph()` function to fetch JSON graph data
- Kept `getWorldMap()` for backwards compatibility

#### 6. Custom Room Node Component (`public/app/src/creator/RoomNode.svelte`)
- Custom Svelte component for rendering room nodes
- Features:
  - Color-coded by area type (forest, town, dungeon, cave, water, mountain)
  - Displays room name, area, and room type
  - Hover effects and selection highlighting
  - Multiple handles for connections (top, bottom, left, right)

#### 7. WorldEditor Component (`public/app/src/creator/WorldEditor.svelte`)
- Complete rewrite with Svelte Flow integration
- Features:
  - Interactive 2D grid map with zoom and pan
  - Dot grid background for visual grid structure
  - MiniMap for world overview
  - Controls for navigation (zoom, fit view)
  - Click-to-select room nodes
  - Sliding side panel showing room details

#### 8. Room Details Panel
- Displays when clicking any room node
- Shows:
  - Room name and description
  - Area and room type information
  - All exits organized by type:
    - Cardinal Directions (north/south/east/west) - shown on map
    - Special Exits (upstairs, downstairs, hidden paths, teleports) - shown in panel
  - Hidden exits marked with lock icon 🔒
  - Clickable exits to navigate between rooms

## Visual Design

### Map View
- Dark theme (#1a1a1a background)
- Cardinal direction exits shown as straight edges on the map
- Smooth zoom and pan controls
- Minimap in corner for navigation

### Room Nodes
- Color-coded by area type:
  - Forest: Dark green (#2d5016)
  - Town: Gray (#4a4a4a)
  - Dungeon: Dark blue-gray (#1a1a2e)
  - Cave: Brown (#3d2817)
  - Water: Blue (#1e3a5f)
  - Mountain: Brown-gray (#5a4a3a)
- Rounded rectangles with shadow effects
- Hover animations for better UX

### Side Panel
- Slides in from right when room is selected
- Organized sections for exits:
  - Cardinal directions (green border)
  - Special exits (cyan border)
  - Hidden exits (orange border)
- Click exits to navigate to connected rooms

## Technical Details

### Grid Layout
- Each grid unit = 100 pixels
- Rooms positioned based on coordinates or auto-calculated
- Cardinal directions follow standard compass layout:
  - North: Y + 1
  - South: Y - 1
  - East: X + 1
  - West: X - 1

### Exit Types
- **Cardinal**: north, south, east, west (shown as edges on map)
- **Non-Cardinal**: upstairs, downstairs, hidden paths, teleports (shown in panel only)
- **Hidden**: Special exits that may require discovery

## Usage

1. Navigate to the WORLD tab in the creator interface
2. The map will load automatically when authenticated
3. Use mouse to pan (drag) and zoom (scroll)
4. Click any room to view details and exits
5. Click exits in the panel to navigate to connected rooms
6. Use controls in bottom-left for zoom and fit view
7. Use minimap in bottom-right for overview navigation

## Future Enhancements

Possible improvements:
- Drag-and-drop to reposition rooms (save back to database)
- Filter by area or room type
- Search for specific rooms
- Toggle to show/hide non-cardinal exits as edges
- Multi-level support (Z-axis) with layer switching
- Room creation directly on the map
- Visual indicators for NPCs and items in rooms
