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

  export let location;

  const { isLoading, isAuthenticated, authToken } = getAuth();

  // Constants
  const GRID_SCALE = 150;      // Pixels per coordinate unit
  const ROOM_WIDTH = 120;      // Room tile width
  const ROOM_HEIGHT = 80;      // Room tile height
  const GRID_COLOR = '#2a2a2a';
  const GRID_MAJOR_COLOR = '#3a3a3a';

  // State
  let rooms = [];
  let loading = true;
  let error = null;
  let roomsValueHelp = [];

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

  // Compute edges from room exits
  $: edges = computeEdges(rooms);

  function computeEdges(roomsList) {
    const edgeList = [];
    const roomMap = new Map(roomsList.map(r => [r.id, r]));
    const processedPairs = new Set();

    roomsList.forEach((room) => {
      if (!room.exits || !room.coords) return;

      room.exits.forEach((exit) => {
        const exitName = exit.name?.toLowerCase();
        const isCardinal = CARDINAL_DIRECTIONS.includes(exitName);
        if (!isCardinal) return;

        const targetRoom = roomMap.get(exit.target);
        if (!targetRoom?.coords) return;

        // Create unique pair key
        const pairKey = [room.id, exit.target].sort().join("-");
        if (processedPairs.has(pairKey)) return;
        processedPairs.add(pairKey);

        // Check if bidirectional
        const reverseExit = targetRoom.exits?.find(
          (e) =>
            e.target === room.id &&
            CARDINAL_DIRECTIONS.includes(e.name?.toLowerCase()) &&
            e.name?.toLowerCase() === getOppositeDirection(exitName)
        );
        const isBidirectional = !!reverseExit;

        edgeList.push({
          id: `edge-${pairKey}`,
          sourceId: room.id,
          targetId: exit.target,
          sourceCoords: room.coords,
          targetCoords: targetRoom.coords,
          direction: exitName,
          isBidirectional,
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
        // Filter to only rooms with coordinates
        rooms = minimalData.rooms.filter(r => r.coords != null);
        console.log(`Loaded ${rooms.length} rooms with coordinates`);
      } else {
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
  function coordsToPixel(coords) {
    return {
      x: coords.x * GRID_SCALE,
      y: coords.y * GRID_SCALE,
    };
  }

  // Convert pixel position to grid coords (snap to nearest)
  function pixelToCoords(px, py) {
    return {
      x: Math.round(px / GRID_SCALE),
      y: Math.round(py / GRID_SCALE),
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
      const dx = (event.clientX - panStart.x) * (viewBox.width / containerElement.clientWidth);
      const dy = (event.clientY - panStart.y) * (viewBox.height / containerElement.clientHeight);

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

  // Check adjacency breaks
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

      const actualOffsetX = targetRoom.coords.x - newCoords.x;
      const actualOffsetY = targetRoom.coords.y - newCoords.y;

      if (actualOffsetX !== expectedOffset.x || actualOffsetY !== expectedOffset.y) {
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

    // Limit zoom (allow much larger zoom out for big maps)
    if (newWidth < 400 || newWidth > 20000) return;

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

  // Fit view to content
  function fitView() {
    if (rooms.length === 0) {
      viewBox = { x: -600, y: -400, width: 1200, height: 800 };
      return;
    }

    const coords = rooms.map(r => r.coords);
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
      <span class="header-hint">
        Drag background to pan. Scroll to zoom. Drag rooms to move. Drag from empty exit handles to create connected rooms.
        <strong style="color: #00bcd4; margin-left: 8px;">
          [{rooms.length} rooms]
        </strong>
      </span>
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
              fill="#1a1a1a"
              class="grid-background"
            />

            <!-- Grid lines -->
            <g class="grid-lines">
              {#each gridLines as line}
                <line
                  x1={line.x1}
                  y1={line.y1}
                  x2={line.x2}
                  y2={line.y2}
                  stroke={line.isMajor ? GRID_MAJOR_COLOR : GRID_COLOR}
                  stroke-width={line.isMajor ? 2 : 1}
                />
              {/each}
            </g>

            <!-- Edges (connections between rooms) -->
            <g class="edges">
              {#each edges as edge}
                {@const source = coordsToPixel(edge.sourceCoords)}
                {@const target = coordsToPixel(edge.targetCoords)}
                <line
                  x1={source.x}
                  y1={source.y}
                  x2={target.x}
                  y2={target.y}
                  stroke={edge.isBidirectional ? '#00bcd4' : '#ff9800'}
                  stroke-width={edge.isBidirectional ? 3 : 2}
                  class="edge-line"
                />
                {#if !edge.isBidirectional}
                  <!-- Arrow for one-way exits -->
                  {@const midX = (source.x + target.x) / 2}
                  {@const midY = (source.y + target.y) / 2}
                  {@const angle = Math.atan2(target.y - source.y, target.x - source.x) * 180 / Math.PI}
                  <polygon
                    points="-8,-5 0,0 -8,5"
                    transform="translate({midX}, {midY}) rotate({angle})"
                    fill="#ff9800"
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
              {#each rooms as room (room.id)}
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
                  on:select={handleRoomSelect}
                  on:dragstart={handleRoomDragStart}
                  on:exitdragstart={handleExitDragStart}
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
                />
              {/if}
            </g>
          </svg>

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
          {#if rooms.length === 0 && !isCreating}
            <div class="empty-state">
              <p>No rooms in the world yet.</p>
              <button class="create-room-btn" on:click={() => startCreatingNewRoom()}>
                <span class="material-symbols-outlined" style="font-size: 18px;">add</span>
                Create Your First Room
              </button>
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
    background: #1a1a1a;
    border-radius: 8px;
    overflow: hidden;
    min-height: 0;
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
    background: #1a1a1a;
    cursor: grab;
    display: block;
  }

  .world-svg.panning {
    cursor: grabbing;
  }

  .grid-lines line {
    pointer-events: none;
  }

  .edge-line {
    pointer-events: none;
  }

  .drag-line {
    pointer-events: none;
  }

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 1;
    color: #00bcd4;
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

  .header-hint {
    font-size: 12px;
    color: #888;
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
    background: #252525;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    color: #aaa;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .icon-btn:hover {
    background: #333;
    color: #fff;
    border-color: #555;
  }

  .create-room-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    font-size: 13px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #000;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .create-room-btn:hover {
    background: #00a5bb;
  }

  .controls {
    position: absolute;
    left: 16px;
    bottom: 16px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: #252525;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    padding: 4px;
    z-index: 10;
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
    background: #333;
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
</style>
