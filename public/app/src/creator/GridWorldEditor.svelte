<script>
  import { onMount } from "svelte";
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { getMinimalRoomsAsync } from "../api/world.js";
  import {
    getRoomsValueHelpAsync,
    createRoomAsync,
    updateRoomAsync,
    deleteRoomAsync,
    getRoomAsync,
  } from "../api/rooms.js";
  import GridRoomTile from "./GridRoomTile.svelte";
  import RoomEditorPanel from "./RoomEditorPanel.svelte";
  import { getOppositeDirection, CARDINAL_DIRECTIONS } from "./WorldEditorStore.js";

  // svelte-ignore unused-export-let
  export let location;

  const { isLoading, isAuthenticated, authToken } = getAuth();

  // Constants
  const GRID_SCALE = 220;      // Pixels per coordinate unit (larger for new card size)
  const ROOM_WIDTH = 180;      // Room tile width
  const ROOM_HEIGHT = 80;      // Room tile height
  const BG_COLOR = '#12121a';  // Dark parchment background
  const GRID_COLOR = '#1a1a28'; // Subtle grid lines

  // State
  let rooms = [];
  let allRooms = []; // All rooms including those without coords (for portal target lookups)
  let loading = true;
  let error = null;
  let roomsValueHelp = [];
  let portalNotification = null; // Toast notification for unplaced rooms

  // Z-level filter
  let selectedZLevel = "all"; // "all" or a number
  $: zLevels = getUniqueZLevels(rooms);
  $: filteredRooms = filterRoomsByZ(rooms, selectedZLevel);

  function getUniqueZLevels(roomsList) {
    const levels = new Set();
    roomsList.forEach(r => {
      if (r.coords?.z !== undefined) {
        levels.add(r.coords.z);
      }
    });
    return [...levels].sort((a, b) => b - a); // Sort descending (highest floor first)
  }

  function filterRoomsByZ(roomsList, zLevel) {
    if (zLevel === "all") return roomsList;
    const z = parseInt(zLevel);
    return roomsList.filter(r => r.coords?.z === z);
  }

  // Fit view when z-level changes
  let prevZLevel = selectedZLevel;
  $: if (selectedZLevel !== prevZLevel) {
    prevZLevel = selectedZLevel;
    setTimeout(fitView, 50);
  }

  // Editor state
  let selectedRoomId = null;
  let editingRoom = null;
  let isCreating = false;
  let saving = false;

  // Viewport state
  let viewBox = { x: -600, y: -400, width: 1200, height: 800 };
  let svgElement;
  let containerElement;

  // Drag state for moving rooms
  let dragging = null; // { roomId, startMouseX, startMouseY, startRoomX, startRoomY }

  // Drag state for creating new rooms from exits
  let exitDragging = null; // { sourceRoom, direction, currentX, currentY }

  // Pan state - left-click drag on background
  let isPanning = false;
  let panStart = { x: 0, y: 0, viewX: 0, viewY: 0 };

  // Warning state
  let adjacencyWarning = null;

  // Direction offsets for creating adjacent rooms
  const directionOffsets = {
    north: { x: 0, y: -1 },
    south: { x: 0, y: 1 },
    east: { x: 1, y: 0 },
    west: { x: -1, y: 0 },
  };

  // Area color palette
  const AREA_COLORS = [
    '#e74c3c', '#3498db', '#2ecc71', '#f39c12', '#9b59b6',
    '#1abc9c', '#e67e22', '#2980b9', '#27ae60', '#c0392b',
    '#8e44ad', '#16a085', '#d35400', '#2c3e50', '#f1c40f',
    '#7f8c8d', '#e91e63', '#00bcd4', '#ff5722', '#607d8b',
  ];

  function hashString(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = ((hash << 5) - hash) + str.charCodeAt(i);
      hash |= 0;
    }
    return Math.abs(hash);
  }

  function getAreaColor(areaName) {
    return AREA_COLORS[hashString(areaName) % AREA_COLORS.length];
  }

  // Compute area background regions (use filtered rooms)
  // Simple approach: one rectangle per room, sized to grid cell, so adjacent rooms merge visually
  $: areaGroups = computeAreaGroups(filteredRooms);

  function computeAreaGroups(roomsList) {
    const groups = new Map();

    roomsList.forEach((room) => {
      if (!room.area || !room.coords) return;
      const area = room.area;
      if (!groups.has(area)) {
        groups.set(area, { name: area, rooms: [], color: getAreaColor(area) });
      }
      groups.get(area).rooms.push(room);
    });

    const result = [];
    const cellW = GRID_SCALE; // Full grid cell width
    const cellH = GRID_SCALE; // Full grid cell height

    groups.forEach((group) => {
      const pixels = group.rooms.map(r => coordsToPixel(r.coords));

      // Each room gets a cell-sized rectangle
      const cells = pixels.map(p => ({
        x: p.x - cellW / 2,
        y: p.y - cellH / 2,
        w: cellW,
        h: cellH,
      }));

      // Bounding box for label positioning
      const xs = pixels.map(p => p.x);
      const ys = pixels.map(p => p.y);
      const minX = Math.min(...xs) - cellW / 2;
      const minY = Math.min(...ys) - cellH / 2;

      result.push({
        name: group.name,
        color: group.color,
        count: group.rooms.length,
        cells,
        labelPos: { x: minX + 10, y: minY + 18 },
        rooms: pixels.map((p, i) => ({
          id: group.rooms[i].id,
          x: p.x,
          y: p.y,
        })),
      });
    });

    return result;
  }

  // Get area color for a specific room
  function getRoomAreaColor(room) {
    if (!room?.area) return '#888';
    return getAreaColor(room.area);
  }

  // Compute edges from room exits (cardinal + non-cardinal, use filtered rooms)
  $: edges = computeEdges(filteredRooms);

  function computeEdges(roomsList) {
    const edgeList = [];
    const roomMap = new Map(roomsList.map(r => [r.id, r]));
    const processedPairs = new Set();

    roomsList.forEach((room) => {
      if (!room.exits || !room.coords) return;

      room.exits.forEach((exit) => {
        const exitName = exit.name?.toLowerCase();
        const isCardinal = CARDINAL_DIRECTIONS.includes(exitName);

        const targetRoom = roomMap.get(exit.target);
        if (!targetRoom?.coords) return;

        // Create unique pair key (include direction for non-cardinal to avoid collisions)
        const pairKey = isCardinal
          ? [room.id, exit.target].sort().join("-")
          : `${room.id}-${exit.target}-${exitName}`;
        if (processedPairs.has(pairKey)) return;
        processedPairs.add(pairKey);

        // Check if bidirectional
        const reverseExit = targetRoom.exits?.find(
          (e) =>
            e.target === room.id &&
            e.name?.toLowerCase() === getOppositeDirection(exitName)
        );
        const isBidirectional = !!reverseExit;

        // For non-cardinal bidirectional, also mark the reverse pair as processed
        if (!isCardinal && isBidirectional) {
          const reversePairKey = `${exit.target}-${room.id}-${getOppositeDirection(exitName)}`;
          processedPairs.add(reversePairKey);
        }

        // Check if cross-zone connection
        const isCrossZone = room.area && targetRoom.area && room.area !== targetRoom.area;

        edgeList.push({
          id: `edge-${pairKey}`,
          sourceId: room.id,
          targetId: exit.target,
          sourceCoords: room.coords,
          targetCoords: targetRoom.coords,
          direction: exitName,
          isBidirectional,
          isCardinal,
          isCrossZone,
          sourceArea: room.area,
          targetArea: targetRoom.area,
        });
      });
    });

    return edgeList;
  }

  // Load data
  async function loadRooms() {
    if (!$authToken) return;

    loading = true;
    error = null;

    try {
      roomsValueHelp = await getRoomsValueHelpAsync($authToken);
      const minimalData = await getMinimalRoomsAsync($authToken);

      if (minimalData?.rooms) {
        allRooms = minimalData.rooms;
        // Filter to only rooms with coordinates for map rendering
        rooms = minimalData.rooms.filter(r => r.coords != null);
        console.log(`Loaded ${rooms.length} rooms with coordinates (${allRooms.length} total)`);
      } else {
        allRooms = [];
        rooms = [];
      }
    } catch (err) {
      console.error("Error loading rooms:", err);
      error = "Failed to load world map";
    } finally {
      loading = false;
    }
  }

  // Convert grid coords to pixel position (center of room)
  // Negate Y so north (higher Y) renders at top of screen
  function coordsToPixel(coords) {
    return {
      x: coords.x * GRID_SCALE,
      y: -coords.y * GRID_SCALE,
    };
  }

  // Convert pixel position to grid coords (snap to nearest)
  // Negate Y to match world coordinates
  function pixelToCoords(px, py) {
    return {
      x: Math.round(px / GRID_SCALE),
      y: Math.round(-py / GRID_SCALE),
    };
  }

  // Get mouse position in SVG coordinates
  function getMouseSVGPosition(event) {
    if (!svgElement) return { x: 0, y: 0 };
    const CTM = svgElement.getScreenCTM();
    if (!CTM) return { x: 0, y: 0 };
    return {
      x: (event.clientX - CTM.e) / CTM.a,
      y: (event.clientY - CTM.f) / CTM.d,
    };
  }

  // Handle room selection
  async function handleRoomSelect(event) {
    const { room } = event.detail;
    selectedRoomId = room.id;

    try {
      const fullRoom = await getRoomAsync($authToken, room.id);
      editingRoom = fullRoom;
      isCreating = false;
    } catch (err) {
      console.error("Error fetching room:", err);
      editingRoom = { ...room, exits: room.exits || [], actions: [] };
      isCreating = false;
    }
  }

  // Handle room drag start (for moving)
  function handleRoomDragStart(event) {
    const { room, event: mouseEvent } = event.detail;
    const mousePos = getMouseSVGPosition(mouseEvent);
    const roomPixel = coordsToPixel(room.coords);

    dragging = {
      roomId: room.id,
      startMouseX: mousePos.x,
      startMouseY: mousePos.y,
      startRoomX: roomPixel.x,
      startRoomY: roomPixel.y,
      currentX: roomPixel.x,
      currentY: roomPixel.y,
      originalCoords: { ...room.coords },
    };

    selectedRoomId = room.id;
  }

  // Handle exit drag start (for creating new rooms)
  function handleExitDragStart(event) {
    const { room, direction, event: mouseEvent } = event.detail;
    const mousePos = getMouseSVGPosition(mouseEvent);
    const roomPixel = coordsToPixel(room.coords);

    // Calculate target position based on direction
    const offset = directionOffsets[direction];
    const targetX = roomPixel.x + offset.x * GRID_SCALE;
    const targetY = roomPixel.y + offset.y * GRID_SCALE;

    exitDragging = {
      sourceRoom: room,
      direction,
      startX: roomPixel.x,
      startY: roomPixel.y,
      currentX: mousePos.x,
      currentY: mousePos.y,
      targetX,
      targetY,
    };
  }

  // Handle mouse move
  function handleMouseMove(event) {
    if (dragging) {
      // Moving a room
      const mousePos = getMouseSVGPosition(event);
      const dx = mousePos.x - dragging.startMouseX;
      const dy = mousePos.y - dragging.startMouseY;

      // Calculate new center position
      const newCenterX = dragging.startRoomX + dx;
      const newCenterY = dragging.startRoomY + dy;

      // Snap center to grid intersection
      const snappedX = Math.round(newCenterX / GRID_SCALE) * GRID_SCALE;
      const snappedY = Math.round(newCenterY / GRID_SCALE) * GRID_SCALE;

      dragging.currentX = snappedX;
      dragging.currentY = snappedY;
      dragging = dragging; // Force reactivity
    } else if (exitDragging) {
      // Dragging from exit to create new room
      const mousePos = getMouseSVGPosition(event);
      exitDragging.currentX = mousePos.x;
      exitDragging.currentY = mousePos.y;
      exitDragging = exitDragging; // Force reactivity
    } else if (isPanning) {
      // Panning the view
      // With preserveAspectRatio="xMidYMid meet", the SVG scales uniformly
      // Use the max ratio to get the effective scale factor
      const scale = Math.max(viewBox.width / containerElement.clientWidth, viewBox.height / containerElement.clientHeight);
      const dx = (event.clientX - panStart.x) * scale;
      const dy = (event.clientY - panStart.y) * scale;

      viewBox = {
        ...viewBox,
        x: panStart.viewX - dx,
        y: panStart.viewY - dy,
      };
    }
  }

  // Handle mouse up
  async function handleMouseUp(event) {
    if (dragging) {
      // Finish moving a room
      const newCoords = pixelToCoords(dragging.currentX, dragging.currentY);
      const oldCoords = dragging.originalCoords;
      const roomId = dragging.roomId;

      dragging = null;

      // Check if position changed
      if (newCoords.x !== oldCoords.x || newCoords.y !== oldCoords.y) {
        // Update local state
        rooms = rooms.map(r => {
          if (r.id === roomId) {
            return {
              ...r,
              coords: { ...r.coords, x: newCoords.x, y: newCoords.y },
            };
          }
          return r;
        });

        // Check for adjacency breaks
        const room = rooms.find(r => r.id === roomId);
        if (room) {
          const breaks = checkAdjacencyBreaks(room, newCoords);
          if (breaks.length > 0) {
            const breakList = breaks.map(b => `${b.direction} -> ${b.targetName}`).join(", ");
            adjacencyWarning = {
              roomId,
              roomName: room.name,
              breaks,
              message: `Moving "${room.name}" breaks cardinal exits: ${breakList}`,
            };
          }
        }

        // Persist to backend
        try {
          const fullRoom = await getRoomAsync($authToken, roomId);
          if (fullRoom) {
            fullRoom.coords = { ...fullRoom.coords, x: newCoords.x, y: newCoords.y };
            await updateRoomAsync($authToken, roomId, fullRoom);
            console.log(`Updated room coords to (${newCoords.x}, ${newCoords.y})`);

            // Update editing room if selected
            if (editingRoom?.id === roomId) {
              editingRoom = { ...editingRoom, coords: fullRoom.coords };
            }
          }
        } catch (err) {
          console.error("Error updating room coordinates:", err);
        }
      }
    }

    if (exitDragging) {
      // Finish creating a new room from exit drag
      const { sourceRoom, direction } = exitDragging;
      exitDragging = null;

      // Start creating a new room at the target position
      startCreatingNewRoom(sourceRoom, direction);
    }

    if (isPanning) {
      isPanning = false;
    }
  }

  // Check adjacency breaks - only warns if the cardinal direction is wrong
  // (e.g. "east" exit points to a room that is west of or same-column as source)
  // Does NOT require exact ±1 offset, allowing gaps between zones
  function checkAdjacencyBreaks(room, newCoords) {
    const breaks = [];

    if (!room.exits) return breaks;

    room.exits.forEach((exit) => {
      const dir = exit.name?.toLowerCase();
      if (!CARDINAL_DIRECTIONS.includes(dir)) return;

      const targetRoom = rooms.find(r => r.id === exit.target);
      if (!targetRoom?.coords) return;

      const expectedOffset = directionOffsets[dir];
      if (!expectedOffset) return;

      const dx = targetRoom.coords.x - newCoords.x;
      const dy = targetRoom.coords.y - newCoords.y;

      // Check direction is correct (sign matches), not exact distance
      const directionWrong =
        (expectedOffset.x > 0 && dx <= 0) ||  // east but target is not to the right
        (expectedOffset.x < 0 && dx >= 0) ||  // west but target is not to the left
        (expectedOffset.y > 0 && dy <= 0) ||  // south but target is not below
        (expectedOffset.y < 0 && dy >= 0);    // north but target is not above

      if (directionWrong) {
        breaks.push({
          direction: dir,
          targetName: targetRoom.name,
        });
      }
    });

    return breaks;
  }

  // Handle SVG background click - start panning
  function handleSvgMouseDown(event) {
    // Only start panning if clicking on the SVG background (not on a room)
    if (event.target === svgElement || event.target.tagName === 'line' || event.target.classList?.contains('grid-background')) {
      if (event.button === 0) { // Left click
        event.preventDefault();
        isPanning = true;
        panStart = {
          x: event.clientX,
          y: event.clientY,
          viewX: viewBox.x,
          viewY: viewBox.y,
        };
      }
    }
  }

  // Handle pane click (deselect) - only if not panning
  function handlePaneClick(event) {
    // Only deselect if we didn't pan significantly
    if (!isPanning) {
      if (event.target === svgElement || event.target.tagName === 'line') {
        if (!isCreating) {
          selectedRoomId = null;
          editingRoom = null;
        }
      }
    }
  }

  // Handle zoom
  function handleWheel(event) {
    event.preventDefault();
    const mousePos = getMouseSVGPosition(event);

    const zoomFactor = event.deltaY > 0 ? 1.1 : 0.9;
    const newWidth = viewBox.width * zoomFactor;
    const newHeight = viewBox.height * zoomFactor;

    // Limit zoom (allow much larger zoom out for big maps, and closer zoom in)
    if (newWidth < 150 || newWidth > 20000) return;

    // Zoom centered on mouse position
    const mouseXRatio = (mousePos.x - viewBox.x) / viewBox.width;
    const mouseYRatio = (mousePos.y - viewBox.y) / viewBox.height;

    viewBox = {
      x: mousePos.x - newWidth * mouseXRatio,
      y: mousePos.y - newHeight * mouseYRatio,
      width: newWidth,
      height: newHeight,
    };
  }

  // Fit view to content (uses filtered rooms)
  function fitView() {
    if (filteredRooms.length === 0) {
      viewBox = { x: -600, y: -400, width: 1200, height: 800 };
      return;
    }

    const coords = filteredRooms.map(r => r.coords);
    const minX = Math.min(...coords.map(c => c.x));
    const maxX = Math.max(...coords.map(c => c.x));
    const minY = Math.min(...coords.map(c => c.y));
    const maxY = Math.max(...coords.map(c => c.y));

    const padding = 2; // Grid units
    const width = (maxX - minX + padding * 2) * GRID_SCALE;
    const height = (maxY - minY + padding * 2) * GRID_SCALE;

    viewBox = {
      x: (minX - padding) * GRID_SCALE,
      y: (minY - padding) * GRID_SCALE,
      width: Math.max(width, 600),
      height: Math.max(height, 400),
    };
  }

  // Panel handlers
  function handleCancel() {
    selectedRoomId = null;
    editingRoom = null;
    isCreating = false;
  }

  async function handleSave(event) {
    const roomData = event.detail;
    if (!$authToken || !roomData) return;

    saving = true;

    try {
      const { _sourceRoomId, _sourceDirection, isNew, ...apiRoomData } = roomData;

      if (apiRoomData.coords) {
        apiRoomData.coords.x = parseInt(apiRoomData.coords.x) || 0;
        apiRoomData.coords.y = parseInt(apiRoomData.coords.y) || 0;
        apiRoomData.coords.z = parseInt(apiRoomData.coords.z) || 0;
      }

      if (isCreating) {
        await createRoomAsync($authToken, apiRoomData);

        // If created via drag from exit, update source room to add exit
        if (_sourceRoomId && _sourceDirection) {
          const sourceRoom = await getRoomAsync($authToken, _sourceRoomId);
          if (sourceRoom) {
            const newExit = {
              name: _sourceDirection,
              description: "",
              target: apiRoomData.id,
              exitType: "direction",
              hidden: false,
            };
            sourceRoom.exits = [...(sourceRoom.exits || []), newExit];
            await updateRoomAsync($authToken, _sourceRoomId, sourceRoom);
          }
        }
      } else {
        await updateRoomAsync($authToken, roomData.id, apiRoomData);
      }

      await loadRooms();
      selectedRoomId = null;
      editingRoom = null;
      isCreating = false;
    } catch (err) {
      console.error("Error saving room:", err);
      alert("Failed to save room: " + (err.message || "Unknown error"));
    } finally {
      saving = false;
    }
  }

  async function handleDelete(event) {
    const roomId = event.detail;
    if (!$authToken || !roomId) return;

    saving = true;

    try {
      await deleteRoomAsync($authToken, roomId);
      await loadRooms();
      selectedRoomId = null;
      editingRoom = null;
      isCreating = false;
    } catch (err) {
      console.error("Error deleting room:", err);
      alert("Failed to delete room: " + (err.message || "Unknown error"));
    } finally {
      saving = false;
    }
  }

  function handleOpenFullEditor(event) {
    const roomId = event.detail;
    navigateTo(`/creator/rooms?id=${roomId}`);
  }

  // Create new room (from button or drag-from-exit)
  function startCreatingNewRoom(sourceRoom = null, direction = null) {
    let newCoords = { x: 0, y: 0, z: 0 };
    let exits = [];

    if (sourceRoom && direction) {
      const offset = directionOffsets[direction] || { x: 0, y: 0 };
      newCoords = {
        x: (sourceRoom.coords?.x || 0) + offset.x,
        y: (sourceRoom.coords?.y || 0) + offset.y,
        z: sourceRoom.coords?.z || 0,
      };

      // Add exit back to source room
      const oppositeDir = getOppositeDirection(direction);
      if (oppositeDir) {
        exits.push({
          name: oppositeDir,
          description: "",
          target: sourceRoom.id,
          exitType: "direction",
          hidden: false,
        });
      }
    }

    const newRoom = {
      id: crypto.randomUUID(),
      name: "New Room",
      description: "",
      detail: "",
      area: sourceRoom?.area || "",
      areaType: sourceRoom?.areaType || "",
      roomType: "",
      coords: newCoords,
      exits,
      actions: [],
      meta: { background: "" },
      isNew: true,
      _sourceRoomId: sourceRoom?.id,
      _sourceDirection: direction,
    };

    editingRoom = newRoom;
    isCreating = true;
  }

  function dismissWarning() {
    adjacencyWarning = null;
  }

  // Handle portal click from room tiles - pan to target room
  async function handlePortalClick(event) {
    const { targetId } = event.detail;
    if (!targetId) return;

    // Look up target in allRooms (includes unplaced rooms)
    const targetRoom = allRooms.find(r => r.id === targetId);
    if (!targetRoom) {
      showPortalNotification('Target room not found.');
      return;
    }

    if (targetRoom.coords != null) {
      // Pan to the target room
      panToRoom(targetRoom);

      // Select it
      selectedRoomId = targetRoom.id;
      try {
        const fullRoom = await getRoomAsync($authToken, targetRoom.id);
        editingRoom = fullRoom;
        isCreating = false;
      } catch (err) {
        console.error("Error fetching room:", err);
        editingRoom = { ...targetRoom, exits: targetRoom.exits || [], actions: [] };
        isCreating = false;
      }
    } else {
      // Room has no coordinates - show notification and open in editor panel
      showPortalNotification(`"${targetRoom.name}" has no map coordinates. Opening in editor panel.`);
      selectedRoomId = targetRoom.id;
      try {
        const fullRoom = await getRoomAsync($authToken, targetRoom.id);
        editingRoom = fullRoom;
        isCreating = false;
      } catch (err) {
        console.error("Error fetching room:", err);
        editingRoom = { ...targetRoom, exits: targetRoom.exits || [], actions: [] };
        isCreating = false;
      }
    }
  }

  function panToRoom(room) {
    if (!room?.coords) return;
    const pixel = coordsToPixel(room.coords);
    // Keep current zoom level, center on room
    viewBox = {
      ...viewBox,
      x: pixel.x - viewBox.width / 2,
      y: pixel.y - viewBox.height / 2,
    };
  }

  function showPortalNotification(message) {
    portalNotification = message;
    setTimeout(() => {
      portalNotification = null;
    }, 4000);
  }

  // Generate grid lines for visible area
  function getGridLines() {
    const lines = [];
    const startX = Math.floor(viewBox.x / GRID_SCALE) * GRID_SCALE;
    const startY = Math.floor(viewBox.y / GRID_SCALE) * GRID_SCALE;
    const endX = viewBox.x + viewBox.width;
    const endY = viewBox.y + viewBox.height;

    // Vertical lines
    for (let x = startX; x <= endX; x += GRID_SCALE) {
      const isMajor = x === 0;
      lines.push({ x1: x, y1: viewBox.y - 100, x2: x, y2: endY + 100, isMajor, isVertical: true });
    }

    // Horizontal lines
    for (let y = startY; y <= endY; y += GRID_SCALE) {
      const isMajor = y === 0;
      lines.push({ x1: viewBox.x - 100, y1: y, x2: endX + 100, y2: y, isMajor, isVertical: false });
    }

    return lines;
  }

  $: gridLines = getGridLines(viewBox);

  // Auto-load when auth ready
  let hasLoaded = false;
  $: if (!$isLoading && $isAuthenticated && $authToken && !hasLoaded) {
    hasLoaded = true;
    loadRooms().then(() => {
      setTimeout(fitView, 100);
    });
  }

  onMount(() => {
    console.log("GridWorldEditor mounted");
  });
</script>

<div class="world-editor-page">
  <div class="header">
    <h4>World Map</h4>
    <div class="header-actions">
      <!-- Z-Level Filter -->
      {#if zLevels.length > 1}
        <div class="z-filter">
          <label for="z-level-select">Floor:</label>
          <select id="z-level-select" bind:value={selectedZLevel}>
            <option value="all">All ({rooms.length})</option>
            {#each zLevels as z}
              <option value={z}>
                {z >= 0 ? `Level ${z}` : `Basement ${Math.abs(z)}`}
                ({rooms.filter(r => r.coords?.z === z).length})
              </option>
            {/each}
          </select>
        </div>
      {/if}

      <button class="icon-btn" on:click={fitView} title="Fit view">
        <span class="material-symbols-outlined">fit_screen</span>
      </button>
      <button class="create-room-btn" on:click={() => startCreatingNewRoom()}>
        <span class="material-symbols-outlined" style="font-size: 18px;">add</span>
        Create Room
      </button>
    </div>
  </div>

  {#if $isAuthenticated}
    {#if loading}
      <div class="loading">Loading world map...</div>
    {:else if error}
      <div class="loading" style="color: #f44336;">{error}</div>
    {:else}
      <div class="world-editor-container" bind:this={containerElement}>
        <div class="graph-area">
          <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
          <svg
            bind:this={svgElement}
            class="world-svg"
            class:panning={isPanning}
            viewBox="{viewBox.x} {viewBox.y} {viewBox.width} {viewBox.height}"
            preserveAspectRatio="xMidYMid meet"
            on:mousedown={handleSvgMouseDown}
            on:mousemove={handleMouseMove}
            on:mouseup={handleMouseUp}
            on:mouseleave={handleMouseUp}
            on:wheel={handleWheel}
            on:click={handlePaneClick}
          >
            <!-- Background rect for capturing clicks -->
            <rect
              x={viewBox.x - 10000}
              y={viewBox.y - 10000}
              width={viewBox.width + 20000}
              height={viewBox.height + 20000}
              fill={BG_COLOR}
              class="grid-background"
            />

            <!-- Area/Zone background regions (one rect per room, adjacent rooms merge) -->
            <g class="area-regions">
              {#each areaGroups as area}
                {#each area.cells as cell}
                  <rect
                    x={cell.x}
                    y={cell.y}
                    width={cell.w}
                    height={cell.h}
                    fill={area.color}
                    fill-opacity="0.08"
                    class="area-cell"
                  />
                {/each}
                <!-- Area label at top-left of region -->
                <text
                  x={area.labelPos.x}
                  y={area.labelPos.y}
                  fill={area.color}
                  fill-opacity="0.7"
                  font-size="14"
                  font-weight="bold"
                  font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"
                  class="area-label"
                >
                  {area.name}
                </text>
              {/each}
            </g>

            <!-- Edges (connections between rooms) -->
            <g class="edges">
              {#each edges as edge}
                {@const source = coordsToPixel(edge.sourceCoords)}
                {@const target = coordsToPixel(edge.targetCoords)}
                {@const edgeColor = edge.isCrossZone ? '#e06040' : (edge.isCardinal ? '#888' : '#b08050')}
                {@const dashArray = edge.isCrossZone ? '8,4' : (edge.isCardinal ? 'none' : '4,3')}
                {@const strokeWidth = edge.isCrossZone ? 2 : 1.5}
                <line
                  x1={source.x}
                  y1={source.y}
                  x2={target.x}
                  y2={target.y}
                  stroke={edgeColor}
                  stroke-width={strokeWidth}
                  stroke-dasharray={dashArray}
                  class="edge-line"
                />
                {#if !edge.isBidirectional}
                  <!-- Arrow for one-way exits -->
                  {@const midX = (source.x + target.x) / 2}
                  {@const midY = (source.y + target.y) / 2}
                  {@const angle = Math.atan2(target.y - source.y, target.x - source.x) * 180 / Math.PI}
                  <polygon
                    points="-7,-4 0,0 -7,4"
                    transform="translate({midX}, {midY}) rotate({angle})"
                    fill={edgeColor}
                  />
                {/if}
              {/each}
            </g>

            <!-- Drag line when creating new room from exit -->
            {#if exitDragging}
              <line
                x1={exitDragging.startX}
                y1={exitDragging.startY}
                x2={exitDragging.currentX}
                y2={exitDragging.currentY}
                stroke="#4caf50"
                stroke-width="3"
                stroke-dasharray="8 4"
                class="drag-line"
              />
              <!-- Ghost room at target position -->
              <rect
                x={exitDragging.targetX - ROOM_WIDTH/2}
                y={exitDragging.targetY - ROOM_HEIGHT/2}
                width={ROOM_WIDTH}
                height={ROOM_HEIGHT}
                rx="6"
                ry="6"
                fill="rgba(76, 175, 80, 0.2)"
                stroke="#4caf50"
                stroke-width="2"
                stroke-dasharray="4 2"
              />
            {/if}

            <!-- Room tiles -->
            <g class="rooms">
              {#each filteredRooms as room (room.id)}
                {@const pixel = dragging?.roomId === room.id
                  ? { x: dragging.currentX, y: dragging.currentY }
                  : coordsToPixel(room.coords)}
                <GridRoomTile
                  {room}
                  x={pixel.x}
                  y={pixel.y}
                  width={ROOM_WIDTH}
                  height={ROOM_HEIGHT}
                  selected={selectedRoomId === room.id}
                  dragging={dragging?.roomId === room.id}
                  areaColor={getRoomAreaColor(room)}
                  on:select={handleRoomSelect}
                  on:dragstart={handleRoomDragStart}
                  on:exitdragstart={handleExitDragStart}
                  on:portalclick={handlePortalClick}
                />
              {/each}

              <!-- Temporary room being created (shown on map while editing) -->
              {#if isCreating && editingRoom?.coords}
                {@const tempPixel = coordsToPixel(editingRoom.coords)}
                <GridRoomTile
                  room={editingRoom}
                  x={tempPixel.x}
                  y={tempPixel.y}
                  width={ROOM_WIDTH}
                  height={ROOM_HEIGHT}
                  selected={true}
                  isTemporary={true}
                  areaColor={getRoomAreaColor(editingRoom)}
                />
              {/if}
            </g>
          </svg>

          <!-- Info panel (top-left) -->
          <div class="info-panel">
            <div class="info-title">TalesMUD World Map</div>
            <div class="info-hint">Scroll to zoom · Drag to pan · Hover for details</div>
            <div class="info-stats">Rooms: {filteredRooms.length}</div>
          </div>

          <!-- Zoom controls -->
          <div class="controls">
            <button on:click={() => viewBox = { ...viewBox, width: viewBox.width * 0.8, height: viewBox.height * 0.8 }} title="Zoom in">
              <span class="material-symbols-outlined">add</span>
            </button>
            <button on:click={() => viewBox = { ...viewBox, width: viewBox.width * 1.2, height: viewBox.height * 1.2 }} title="Zoom out">
              <span class="material-symbols-outlined">remove</span>
            </button>
            <button on:click={fitView} title="Fit to view">
              <span class="material-symbols-outlined">fit_screen</span>
            </button>
          </div>

          <!-- Area legend -->
          <div class="area-legend">
            {#if areaGroups.length > 0}
              <div class="legend-section">
                <div class="legend-title">Zones</div>
                {#each areaGroups as area}
                  <div class="legend-item">
                    <span class="legend-swatch" style="background: {area.color};"></span>
                    <span class="legend-name">{area.name}</span>
                    <span class="legend-count">{area.count}</span>
                  </div>
                {/each}
              </div>
              <hr class="legend-divider" />
            {/if}
            <div class="legend-section">
              <div class="legend-item connection-type">
                <svg width="30" height="10"><line x1="0" y1="5" x2="30" y2="5" stroke="#888" stroke-width="1.5"/></svg>
                <span class="legend-name">Cardinal</span>
              </div>
              <div class="legend-item connection-type">
                <svg width="30" height="10"><line x1="0" y1="5" x2="30" y2="5" stroke="#e06040" stroke-width="2" stroke-dasharray="8,4"/></svg>
                <span class="legend-name">Cross-zone</span>
              </div>
              <div class="legend-item connection-type">
                <svg width="30" height="10"><line x1="0" y1="5" x2="30" y2="5" stroke="#b08050" stroke-width="1.5" stroke-dasharray="4,3"/></svg>
                <span class="legend-name">Special exit</span>
              </div>
            </div>
          </div>

          <!-- Portal notification toast -->
          {#if portalNotification}
            <div class="portal-notification">
              <span class="material-symbols-outlined" style="font-size: 16px;">info</span>
              {portalNotification}
            </div>
          {/if}

          <!-- Adjacency warning -->
          {#if adjacencyWarning}
            <div class="adjacency-warning">
              <div class="warning-content">
                <span class="material-symbols-outlined warning-icon">warning</span>
                <div class="warning-text">
                  <strong>Adjacency Warning</strong>
                  <p>{adjacencyWarning.message}</p>
                  <small>Cardinal exits may no longer point to adjacent rooms.</small>
                </div>
                <button class="dismiss-btn" on:click={dismissWarning}>
                  <span class="material-symbols-outlined">close</span>
                </button>
              </div>
            </div>
          {/if}

          <!-- Empty state -->
          {#if filteredRooms.length === 0 && !isCreating}
            <div class="empty-state">
              {#if rooms.length === 0}
                <p>No rooms in the world yet.</p>
                <button class="create-room-btn" on:click={() => startCreatingNewRoom()}>
                  <span class="material-symbols-outlined" style="font-size: 18px;">add</span>
                  Create Your First Room
                </button>
              {:else}
                <p>No rooms on this floor.</p>
                <button class="create-room-btn" on:click={() => selectedZLevel = "all"}>
                  <span class="material-symbols-outlined" style="font-size: 18px;">layers</span>
                  Show All Floors
                </button>
              {/if}
            </div>
          {/if}
        </div>

        <RoomEditorPanel
          room={editingRoom}
          {isCreating}
          {saving}
          {roomsValueHelp}
          on:save={handleSave}
          on:cancel={handleCancel}
          on:delete={handleDelete}
          on:openFullEditor={handleOpenFullEditor}
        />
      </div>
    {/if}
  {:else}
    <p>Please log in to view the world map.</p>
  {/if}
</div>

<style>
  .world-editor-page {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 120px);
    padding: 16px;
  }

  .world-editor-container {
    display: flex;
    flex: 1;
    background: #12121a;
    border-radius: 8px;
    overflow: hidden;
    min-height: 0;
    border: 1px solid #2a2a4a;
  }

  .graph-area {
    flex: 1;
    position: relative;
    min-width: 0;
    overflow: hidden;
  }

  .world-svg {
    width: 100%;
    height: 100%;
    background: #12121a;
    cursor: grab;
    display: block;
  }

  .world-svg.panning {
    cursor: grabbing;
  }

  .edge-line {
    pointer-events: none;
  }

  .drag-line {
    pointer-events: none;
  }

  .area-bounds {
    pointer-events: none;
  }

  .area-label {
    pointer-events: none;
    font-family: system-ui, -apple-system, sans-serif;
  }

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 1;
    color: #888;
    font-size: 18px;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    flex-shrink: 0;
  }

  .header h4 {
    margin: 0;
    color: #fff;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    background: #252540;
    border: 1px solid #3a3a5a;
    border-radius: 6px;
    color: #aaa;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .icon-btn:hover {
    background: #333355;
    color: #fff;
    border-color: #4a4a6a;
  }

  .create-room-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    font-size: 13px;
    font-weight: 500;
    background: #5c8d55;
    border: none;
    border-radius: 6px;
    color: #fff;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .create-room-btn:hover {
    background: #4a7a45;
  }

  .z-filter {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 12px;
    background: #252540;
    border: 1px solid #3a3a5a;
    border-radius: 6px;
  }

  .z-filter label {
    font-size: 12px;
    color: #aaa;
    white-space: nowrap;
  }

  .z-filter select {
    background: #1a1a2e;
    border: 1px solid #3a3a5a;
    border-radius: 4px;
    color: #fff;
    padding: 4px 8px;
    font-size: 12px;
    cursor: pointer;
    min-width: 120px;
  }

  .z-filter select:hover {
    border-color: #4a4a6a;
  }

  .z-filter select:focus {
    outline: none;
    border-color: #5c8d55;
  }

  /* Info panel (top-left) */
  .info-panel {
    position: absolute;
    top: 16px;
    left: 16px;
    background: rgba(20, 20, 35, 0.95);
    border: 1px solid #3a3a5a;
    border-radius: 8px;
    padding: 12px 16px;
    z-index: 10;
    box-shadow: 0 4px 12px rgba(0,0,0,0.4);
  }

  .info-title {
    font-size: 14px;
    font-weight: 700;
    color: #fff;
    margin-bottom: 4px;
  }

  .info-hint {
    font-size: 12px;
    color: #888;
    margin-bottom: 6px;
  }

  .info-stats {
    font-size: 11px;
    color: #666;
  }

  .controls {
    position: absolute;
    left: 16px;
    bottom: 16px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: rgba(20, 20, 35, 0.95);
    border: 1px solid #3a3a5a;
    border-radius: 6px;
    padding: 4px;
    z-index: 10;
    box-shadow: 0 4px 12px rgba(0,0,0,0.4);
  }

  .controls button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background: transparent;
    border: none;
    color: #aaa;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.15s ease;
  }

  .controls button:hover {
    background: #333355;
    color: #fff;
  }

  .empty-state {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    text-align: center;
    color: #666;
    z-index: 5;
  }

  .empty-state p {
    margin: 0 0 16px;
    font-size: 16px;
  }

  .adjacency-warning {
    position: absolute;
    top: 16px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 100;
    max-width: 500px;
  }

  .warning-content {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
    background: rgba(255, 152, 0, 0.15);
    border: 1px solid #ff9800;
    border-radius: 8px;
    backdrop-filter: blur(8px);
  }

  .warning-icon {
    color: #ff9800;
    font-size: 24px;
    flex-shrink: 0;
  }

  .warning-text {
    flex: 1;
  }

  .warning-text strong {
    display: block;
    color: #ff9800;
    font-size: 13px;
    margin-bottom: 4px;
  }

  .warning-text p {
    margin: 0 0 4px;
    font-size: 12px;
    color: #fff;
  }

  .warning-text small {
    font-size: 11px;
    color: #aaa;
  }

  .dismiss-btn {
    background: transparent;
    border: none;
    color: #888;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .dismiss-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
  }

  /* Area legend */
  .area-legend {
    position: absolute;
    top: 16px;
    right: 16px;
    background: rgba(20, 20, 35, 0.95);
    border: 1px solid #3a3a5a;
    border-radius: 8px;
    padding: 12px 14px;
    z-index: 10;
    max-height: 400px;
    overflow-y: auto;
    min-width: 150px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.4);
  }

  .legend-section {
    margin-bottom: 4px;
  }

  .legend-divider {
    border: none;
    border-top: 1px solid #3a3a5a;
    margin: 10px 0;
  }

  .legend-title {
    font-size: 11px;
    font-weight: 700;
    color: #ccc;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 3px 0;
    font-size: 12px;
    color: #aaa;
  }

  .legend-item.connection-type {
    gap: 6px;
  }

  .legend-item.connection-type svg {
    flex-shrink: 0;
  }

  .legend-swatch {
    width: 14px;
    height: 14px;
    border-radius: 3px;
    flex-shrink: 0;
  }

  .legend-name {
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .legend-count {
    font-size: 10px;
    color: #666;
    font-family: monospace;
  }

  /* Portal notification */
  .portal-notification {
    position: absolute;
    bottom: 16px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    background: rgba(30, 30, 50, 0.95);
    border: 1px solid #3a3a5a;
    border-radius: 8px;
    color: #ddd;
    font-size: 13px;
    z-index: 100;
    white-space: nowrap;
    backdrop-filter: blur(8px);
  }
</style>
