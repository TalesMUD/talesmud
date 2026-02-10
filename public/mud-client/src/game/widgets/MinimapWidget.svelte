<script>
  import { onDestroy } from 'svelte';

  export let store = null;
  export let sendMessage = null;

  let canvas;
  let container;
  let resizeObserver;
  let tooltip = { visible: false, text: '', x: 0, y: 0 };

  // Rendering constants
  const ROOM_SIZE = 18;
  const CELL_SPACING = 40;
  const LINE_WIDTH = 2;
  const MAX_VISIBLE_DISTANCE = 5;
  const ROOM_RADIUS = 3; // border-radius for rounded squares

  // Colors
  const COLOR_CURRENT = '#f59e0b';
  const COLOR_CURRENT_FILL = '#2a2000';
  const COLOR_ROOM_BORDER = '#6b7280';
  const COLOR_ROOM_FILL = '#1f2937';
  const COLOR_LINE = '#4b5563';
  const COLOR_TARGET = '#22d3ee';
  const COLOR_TARGET_FILL = '#002a2e';
  const COLOR_PATH = '#0ea5e9';

  // Opacity per BFS distance
  function getOpacity(distance) {
    if (distance === 0) return 1.0;
    if (distance === 1) return 0.85;
    if (distance === 2) return 0.65;
    if (distance === 3) return 0.45;
    if (distance === 4) return 0.25;
    return 0;
  }

  // Reactive store data
  let visitedRooms = {};
  let currentRoomId = null;

  // Travel state
  let travelTargetId = null;
  let travelPath = []; // [{roomId, direction}] — remaining steps
  let isTraveling = false;
  let travelPathRoomIds = new Set(); // room IDs on the travel path (for rendering)

  $: if (store) {
    visitedRooms = $store.visitedRooms || {};
    const newRoomId = $store.currentRoomId;
    if (newRoomId !== currentRoomId) {
      const prevRoomId = currentRoomId;
      currentRoomId = newRoomId;
      // Advance travel when room changes
      if (isTraveling && newRoomId) {
        advanceTravel(newRoomId);
      }
    }
  }

  // Recompute and redraw whenever store data changes
  $: if (canvas && currentRoomId) {
    draw(visitedRooms, currentRoomId);
  }

  // Also redraw when travel state changes
  $: if (canvas && currentRoomId && (travelTargetId || travelTargetId === null)) {
    draw(visitedRooms, currentRoomId);
  }

  // BFS to calculate distance from current room through visited rooms
  function computeDistances(rooms, startId) {
    const distances = {};
    if (!startId || !rooms[startId]) return distances;

    const queue = [startId];
    distances[startId] = 0;

    while (queue.length > 0) {
      const id = queue.shift();
      const room = rooms[id];
      if (!room) continue;

      for (const exit of room.cardinalExits || []) {
        if (!distances.hasOwnProperty(exit.targetId) && rooms[exit.targetId]) {
          distances[exit.targetId] = distances[id] + 1;
          queue.push(exit.targetId);
        }
      }
    }

    return distances;
  }

  // BFS pathfinding: returns array of {roomId, direction} steps from startId to targetId
  function findPath(rooms, startId, targetId) {
    if (!startId || !targetId || !rooms[startId] || !rooms[targetId]) return null;
    if (startId === targetId) return [];

    const queue = [startId];
    const visited = new Set([startId]);
    const parent = {}; // parent[childId] = {parentId, direction}

    while (queue.length > 0) {
      const id = queue.shift();
      const room = rooms[id];
      if (!room) continue;

      for (const exit of room.cardinalExits || []) {
        if (!visited.has(exit.targetId) && rooms[exit.targetId]) {
          visited.add(exit.targetId);
          parent[exit.targetId] = { parentId: id, direction: exit.dir };
          queue.push(exit.targetId);

          if (exit.targetId === targetId) {
            // Reconstruct path
            const path = [];
            let cur = targetId;
            while (parent[cur]) {
              path.unshift({ roomId: cur, direction: parent[cur].direction });
              cur = parent[cur].parentId;
            }
            return path;
          }
        }
      }
    }

    return null; // unreachable
  }

  function startTravel(targetId) {
    if (!currentRoomId || targetId === currentRoomId) return;

    const path = findPath(visitedRooms, currentRoomId, targetId);
    if (!path || path.length === 0) return;

    travelTargetId = targetId;
    travelPath = path;
    isTraveling = true;
    travelPathRoomIds = new Set(path.map(s => s.roomId));

    // Send first step
    sendNextStep();
  }

  function cancelTravel() {
    travelTargetId = null;
    travelPath = [];
    isTraveling = false;
    travelPathRoomIds = new Set();
  }

  function sendNextStep() {
    if (!isTraveling || travelPath.length === 0) {
      cancelTravel();
      return;
    }
    const step = travelPath[0];
    if (sendMessage) {
      sendMessage(step.direction);
    }
  }

  function advanceTravel(newRoomId) {
    if (!isTraveling || travelPath.length === 0) {
      cancelTravel();
      return;
    }

    // Check if we arrived at the expected next room
    if (travelPath[0].roomId === newRoomId) {
      travelPath = travelPath.slice(1);
      travelPathRoomIds = new Set(travelPath.map(s => s.roomId));

      if (travelPath.length === 0) {
        // Arrived at destination
        cancelTravel();
      } else {
        // Small delay between steps for visual feedback
        setTimeout(() => sendNextStep(), 150);
      }
    } else {
      // Unexpected room — cancel travel
      cancelTravel();
    }
  }

  function handleCanvasClick(e) {
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;

    let found = null;
    for (const entry of roomsByPixel) {
      if (mx >= entry.x && mx <= entry.x + entry.w &&
          my >= entry.y && my <= entry.y + entry.h) {
        found = entry;
        break;
      }
    }

    if (!found) return;

    const clickedId = found.room.id;
    if (clickedId === currentRoomId) return;

    // If clicking current target, cancel travel
    if (clickedId === travelTargetId) {
      cancelTravel();
      return;
    }

    // Start new travel (cancels any existing)
    cancelTravel();
    startTravel(clickedId);
  }

  function handleClearMap() {
    cancelTravel();
    if (store) {
      store.clearVisitedRooms();
    }
  }

  // Build a lookup from coords string to room for hit testing
  let roomsByPixel = []; // [{x, y, w, h, room}] for hit testing

  function draw(rooms, curId) {
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;

    // Work in CSS pixel space
    const w = canvas.width / dpr;
    const h = canvas.height / dpr;

    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.imageSmoothingEnabled = false;
    ctx.clearRect(0, 0, w, h);

    const currentRoom = rooms[curId];
    if (!currentRoom || !currentRoom.coords) return;

    const currentZ = currentRoom.coords.z;

    // Filter to same z-level and rooms with coords
    const visibleRoomIds = Object.keys(rooms).filter(id => {
      const r = rooms[id];
      return r.coords && r.coords.z === currentZ;
    });

    if (visibleRoomIds.length === 0) return;

    // BFS distances
    const distances = computeDistances(rooms, curId);

    // Filter to rooms within visible distance
    const drawableIds = visibleRoomIds.filter(id =>
      distances.hasOwnProperty(id) && distances[id] <= MAX_VISIBLE_DISTANCE
    );

    if (drawableIds.length === 0) return;

    // Center viewport on current room — snap to integer pixels
    const cx = currentRoom.coords.x;
    const cy = currentRoom.coords.y;
    const centerX = Math.round(w / 2);
    const centerY = Math.round(h / 2);

    // Convert world coords to pixel coords — fixed spacing, Y flipped (north=up)
    function toPixel(worldX, worldY) {
      return {
        px: Math.round(centerX + (worldX - cx) * CELL_SPACING),
        py: Math.round(centerY + (worldY - cy) * CELL_SPACING),
      };
    }

    // Only draw rooms that fit within the canvas (with margin for room size)
    const margin = ROOM_SIZE / 2 + 2;
    const visibleDrawableIds = drawableIds.filter(id => {
      const room = rooms[id];
      const { px, py } = toPixel(room.coords.x, room.coords.y);
      return px - margin >= 0 && px + margin <= w && py - margin >= 0 && py + margin <= h;
    });

    // Reset hit-test data
    roomsByPixel = [];

    // Draw connections first (below rooms)
    for (const id of visibleDrawableIds) {
      const room = rooms[id];
      const dist = distances[id];
      const { px: x1, py: y1 } = toPixel(room.coords.x, room.coords.y);

      for (const exit of room.cardinalExits || []) {
        const target = rooms[exit.targetId];
        if (!target || !target.coords || target.coords.z !== currentZ) continue;
        const targetDist = distances[exit.targetId];
        if (targetDist === undefined || targetDist > MAX_VISIBLE_DISTANCE) continue;

        // Use the max distance of the two endpoints for line opacity
        const lineDist = Math.max(dist, targetDist);
        const opacity = getOpacity(lineDist);
        if (opacity <= 0) continue;

        const { px: x2, py: y2 } = toPixel(target.coords.x, target.coords.y);

        // Highlight path connections
        const isPathConnection = travelPathRoomIds.has(id) || travelPathRoomIds.has(exit.targetId);

        ctx.save();
        ctx.globalAlpha = isPathConnection ? 1.0 : opacity;
        ctx.strokeStyle = isPathConnection ? COLOR_PATH : COLOR_LINE;
        ctx.lineWidth = isPathConnection ? 3 : LINE_WIDTH;
        ctx.beginPath();
        // For even lineWidth, draw on integer coords; for odd, offset by 0.5
        const lw = isPathConnection ? 3 : LINE_WIDTH;
        const offset = lw % 2 === 0 ? 0 : 0.5;
        ctx.moveTo(x1 + offset, y1 + offset);
        ctx.lineTo(x2 + offset, y2 + offset);
        ctx.stroke();
        ctx.restore();
      }
    }

    // Draw rooms
    for (const id of visibleDrawableIds) {
      const room = rooms[id];
      const dist = distances[id];
      const opacity = getOpacity(dist);
      if (opacity <= 0) continue;

      const { px, py } = toPixel(room.coords.x, room.coords.y);
      const half = ROOM_SIZE / 2;
      const rx = px - half;
      const ry = py - half;
      const isCurrent = id === curId;
      const isTarget = id === travelTargetId;
      const isOnPath = travelPathRoomIds.has(id);

      ctx.save();
      ctx.globalAlpha = (isTarget || isOnPath) ? 1.0 : opacity;

      // Glow for current room or travel target
      if (isCurrent) {
        ctx.shadowColor = COLOR_CURRENT;
        ctx.shadowBlur = 12;
      } else if (isTarget) {
        ctx.shadowColor = COLOR_TARGET;
        ctx.shadowBlur = 12;
      }

      // Room fill
      let fillColor = COLOR_ROOM_FILL;
      if (isCurrent) fillColor = COLOR_CURRENT_FILL;
      else if (isTarget) fillColor = COLOR_TARGET_FILL;
      ctx.fillStyle = fillColor;
      roundRect(ctx, rx, ry, ROOM_SIZE, ROOM_SIZE, ROOM_RADIUS);
      ctx.fill();

      // Room border
      ctx.shadowBlur = 0;
      let borderColor = COLOR_ROOM_BORDER;
      if (isCurrent) borderColor = COLOR_CURRENT;
      else if (isTarget) borderColor = COLOR_TARGET;
      else if (isOnPath) borderColor = COLOR_PATH;
      ctx.strokeStyle = borderColor;

      const thick = isCurrent || isTarget;
      ctx.lineWidth = thick ? 2 : 1;
      if (!thick) {
        roundRect(ctx, rx + 0.5, ry + 0.5, ROOM_SIZE - 1, ROOM_SIZE - 1, ROOM_RADIUS);
      } else {
        roundRect(ctx, rx, ry, ROOM_SIZE, ROOM_SIZE, ROOM_RADIUS);
      }
      ctx.stroke();

      ctx.restore();

      // Store for hit testing
      roomsByPixel.push({
        x: rx,
        y: ry,
        w: ROOM_SIZE,
        h: ROOM_SIZE,
        room,
      });
    }

    // Restore the DPR transform
    ctx.restore();
  }

  // Rounded rect helper
  function roundRect(ctx, x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.lineTo(x + w - r, y);
    ctx.quadraticCurveTo(x + w, y, x + w, y + r);
    ctx.lineTo(x + w, y + h - r);
    ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
    ctx.lineTo(x + r, y + h);
    ctx.quadraticCurveTo(x, y + h, x, y + h - r);
    ctx.lineTo(x, y + r);
    ctx.quadraticCurveTo(x, y, x + r, y);
    ctx.closePath();
  }

  function handleMouseMove(e) {
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;

    let found = null;
    for (const entry of roomsByPixel) {
      if (mx >= entry.x && mx <= entry.x + entry.w &&
          my >= entry.y && my <= entry.y + entry.h) {
        found = entry;
        break;
      }
    }

    if (found) {
      const roomName = found.room.name || found.room.id;
      const isTarget = found.room.id === travelTargetId;
      const isCurrent = found.room.id === currentRoomId;
      let text = roomName;
      if (isCurrent) text += ' (you are here)';
      else if (isTarget && isTraveling) text += ' (traveling...)';
      else if (isTarget) text += ' (target)';
      else text += ' (click to travel)';
      tooltip = {
        visible: true,
        text,
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
      };
    } else {
      tooltip = { ...tooltip, visible: false };
    }
  }

  function handleMouseLeave() {
    tooltip = { ...tooltip, visible: false };
  }

  function resizeCanvas() {
    if (!canvas || !container) return;
    const rect = container.getBoundingClientRect();
    if (rect.width === 0 || rect.height === 0) return;
    const dpr = window.devicePixelRatio || 1;
    const pixelW = Math.round(rect.width * dpr);
    const pixelH = Math.round(rect.height * dpr);
    canvas.width = pixelW;
    canvas.height = pixelH;
    canvas.style.width = (pixelW / dpr) + 'px';
    canvas.style.height = (pixelH / dpr) + 'px';
    draw(visitedRooms, currentRoomId);
  }

  // Re-observe whenever the container element appears/changes ({#if} toggle)
  let observedContainer = null;
  $: if (container !== observedContainer) {
    if (resizeObserver) resizeObserver.disconnect();
    observedContainer = container;
    if (container) {
      resizeObserver = new ResizeObserver(() => resizeCanvas());
      resizeObserver.observe(container);
      resizeCanvas();
    }
  }

  onDestroy(() => {
    if (resizeObserver) resizeObserver.disconnect();
    cancelTravel();
  });
</script>

<style>
  .minimap-container {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
    background: #0d1117;
    border-radius: 8px;
  }

  .minimap-header {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    padding: 6px 10px;
    font-size: 11px;
    font-weight: 600;
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    display: flex;
    align-items: center;
    gap: 6px;
    z-index: 2;
    pointer-events: none;
  }

  .minimap-header i {
    font-size: 14px;
    color: #6b7280;
  }

  .header-spacer {
    flex: 1;
  }

  .header-btn {
    pointer-events: auto;
    cursor: pointer;
    background: none;
    border: none;
    color: #4b5563;
    padding: 2px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s, background 0.15s;
    font-size: 14px;
  }

  .header-btn:hover {
    color: #e5e7eb;
    background: rgba(255, 255, 255, 0.1);
  }

  .header-btn.cancel {
    color: #ef4444;
  }

  .header-btn.cancel:hover {
    background: rgba(239, 68, 68, 0.15);
  }

  .travel-indicator {
    pointer-events: auto;
    font-size: 10px;
    font-weight: 600;
    color: #22d3ee;
    letter-spacing: 0;
    text-transform: none;
  }

  .canvas-wrap {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
  }

  canvas {
    display: block;
    cursor: pointer;
  }

  .tooltip {
    position: absolute;
    background: rgba(0, 0, 0, 0.85);
    color: #e5e7eb;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    pointer-events: none;
    white-space: nowrap;
    z-index: 10;
    border: 1px solid #374151;
    transform: translate(-50%, -100%);
    margin-top: -8px;
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #4b5563;
    font-size: 12px;
    text-align: center;
    padding: 16px;
  }

  .empty-state i {
    font-size: 32px;
    margin-bottom: 8px;
    display: block;
    color: #374151;
  }
</style>

<div class="minimap-container">
  <div class="minimap-header">
    <i class="material-icons">map</i>
    Minimap
    {#if isTraveling}
      <span class="travel-indicator">Traveling...</span>
    {/if}
    <span class="header-spacer"></span>
    {#if isTraveling}
      <button class="header-btn cancel" title="Cancel travel" on:click={cancelTravel}>
        <i class="material-icons" style="font-size: inherit">close</i>
      </button>
    {/if}
    {#if currentRoomId && Object.keys(visitedRooms).length > 0}
      <button class="header-btn" title="Clear map data" on:click={handleClearMap}>
        <i class="material-icons" style="font-size: inherit">delete_sweep</i>
      </button>
    {/if}
  </div>

  {#if currentRoomId && Object.keys(visitedRooms).length > 0}
    <div class="canvas-wrap" bind:this={container}>
      <canvas
        bind:this={canvas}
        on:mousemove={handleMouseMove}
        on:mouseleave={handleMouseLeave}
        on:click={handleCanvasClick}
      ></canvas>
      {#if tooltip.visible}
        <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">
          {tooltip.text}
        </div>
      {/if}
    </div>
  {:else}
    <div class="empty-state">
      <div>
        <i class="material-icons">explore</i>
        Explore to reveal the map
      </div>
    </div>
  {/if}
</div>
