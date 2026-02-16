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
  const ROOM_RADIUS = 3;

  // Colors
  const COLOR_CURRENT = '#f59e0b';
  const COLOR_CURRENT_FILL = '#2a2000';
  const COLOR_NEARBY_BORDER = '#9ca3af';
  const COLOR_NEARBY_FILL = '#1f2937';
  const COLOR_ROOM_BORDER = '#6b7280';
  const COLOR_ROOM_FILL = '#1a1f2b';
  const COLOR_LINE = '#4b5563';
  const COLOR_LINE_DIM = '#374151';
  const COLOR_TARGET = '#22d3ee';
  const COLOR_TARGET_FILL = '#002a2e';
  const COLOR_PATH = '#0ea5e9';

  // Reactive store data
  let visitedRooms = {};
  let currentRoomId = null;

  // Travel state
  let travelTargetId = null;
  let travelPath = [];
  let isTraveling = false;
  let travelPathRoomIds = new Set();

  // Pan state
  let panOffsetX = 0;
  let panOffsetY = 0;
  let isPanning = false;
  let panStartX = 0;
  let panStartY = 0;
  let panStartOffsetX = 0;
  let panStartOffsetY = 0;
  let didDrag = false;

  // Maximize state
  let maximized = false;

  // Portal action — moves element to document.body so it escapes widget grid stacking contexts
  function portal(node) {
    document.body.appendChild(node);
    return {
      destroy() {
        if (node.parentNode) {
          node.parentNode.removeChild(node);
        }
      }
    };
  }

  $: if (store) {
    visitedRooms = $store.visitedRooms || {};
    const newRoomId = $store.currentRoomId;
    if (newRoomId !== currentRoomId) {
      currentRoomId = newRoomId;
      panOffsetX = 0;
      panOffsetY = 0;
      if (isTraveling && newRoomId) {
        advanceTravel(newRoomId);
      }
    }
  }

  $: if (canvas && currentRoomId) {
    draw(visitedRooms, currentRoomId);
  }

  $: if (canvas && currentRoomId && (travelTargetId || travelTargetId === null)) {
    draw(visitedRooms, currentRoomId);
  }

  $: if (canvas && currentRoomId && (panOffsetX !== undefined || panOffsetY !== undefined)) {
    draw(visitedRooms, currentRoomId);
  }

  // BFS distance from current room
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

  // BFS pathfinding
  function findPath(rooms, startId, targetId) {
    if (!startId || !targetId || !rooms[startId] || !rooms[targetId]) return null;
    if (startId === targetId) return [];

    const queue = [startId];
    const visited = new Set([startId]);
    const parent = {};

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

    return null;
  }

  function startTravel(targetId) {
    if (!currentRoomId || targetId === currentRoomId) return;

    const path = findPath(visitedRooms, currentRoomId, targetId);
    if (!path || path.length === 0) return;

    travelTargetId = targetId;
    travelPath = path;
    isTraveling = true;
    travelPathRoomIds = new Set(path.map(s => s.roomId));

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

    if (travelPath[0].roomId === newRoomId) {
      travelPath = travelPath.slice(1);
      travelPathRoomIds = new Set(travelPath.map(s => s.roomId));

      if (travelPath.length === 0) {
        cancelTravel();
      } else {
        setTimeout(() => sendNextStep(), 150);
      }
    } else {
      cancelTravel();
    }
  }

  // --- Pan handlers ---

  function handlePointerDown(e) {
    if (!canvas) return;
    isPanning = true;
    didDrag = false;
    panStartX = e.clientX;
    panStartY = e.clientY;
    panStartOffsetX = panOffsetX;
    panStartOffsetY = panOffsetY;
    canvas.setPointerCapture(e.pointerId);
  }

  function handlePointerMove(e) {
    if (!canvas) return;

    if (isPanning) {
      const dx = e.clientX - panStartX;
      const dy = e.clientY - panStartY;
      if (Math.abs(dx) > 3 || Math.abs(dy) > 3) {
        didDrag = true;
      }
      panOffsetX = panStartOffsetX + dx;
      panOffsetY = panStartOffsetY + dy;
      return;
    }

    // Tooltip on hover
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
      const isCurrent = found.room.id === currentRoomId;
      const isTarget = found.room.id === travelTargetId;
      let text = roomName;
      if (isCurrent) text += ' (you are here)';
      else if (isTarget && isTraveling) text += ' (traveling...)';
      else if (isTarget) text += ' (target)';
      else text += ' (click to travel)';
      tooltip = { visible: true, text, x: e.clientX - rect.left, y: e.clientY - rect.top };
    } else {
      tooltip = { ...tooltip, visible: false };
    }
  }

  function handlePointerUp(e) {
    if (!canvas) return;
    canvas.releasePointerCapture(e.pointerId);

    if (isPanning && !didDrag) {
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
        const clickedId = found.room.id;
        if (clickedId !== currentRoomId) {
          if (clickedId === travelTargetId) {
            cancelTravel();
          } else {
            cancelTravel();
            startTravel(clickedId);
          }
        }
      }
    }

    isPanning = false;
    didDrag = false;
  }

  function handleMouseLeave() {
    tooltip = { ...tooltip, visible: false };
  }

  function recenter() {
    panOffsetX = 0;
    panOffsetY = 0;
  }

  function toggleMaximize() {
    maximized = !maximized;
    // After Svelte swaps the DOM (container/canvas change via bind:this),
    // the ResizeObserver reactive block picks it up automatically.
    // But we need a small delay for the portal to mount.
    setTimeout(() => resizeCanvas(), 30);
  }

  function handleBackdropClick(e) {
    // Close modal when clicking the dark backdrop area (not the modal content)
    if (e.target === e.currentTarget) {
      toggleMaximize();
    }
  }

  function handleBackdropKeydown(e) {
    if (e.key === 'Escape') {
      toggleMaximize();
    }
  }

  function handleClearMap() {
    cancelTravel();
    if (store) {
      store.clearVisitedRooms();
    }
  }

  // Hit-test data
  let roomsByPixel = [];

  function draw(rooms, curId) {
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;

    const w = canvas.width / dpr;
    const h = canvas.height / dpr;

    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.imageSmoothingEnabled = false;
    ctx.clearRect(0, 0, w, h);

    const currentRoom = rooms[curId];
    if (!currentRoom || !currentRoom.coords) {
      ctx.restore();
      return;
    }

    const currentZ = currentRoom.coords.z;

    const visibleRoomIds = Object.keys(rooms).filter(id => {
      const r = rooms[id];
      return r.coords && r.coords.z === currentZ;
    });

    if (visibleRoomIds.length === 0) {
      ctx.restore();
      return;
    }

    const distances = computeDistances(rooms, curId);

    const cx = currentRoom.coords.x;
    const cy = currentRoom.coords.y;
    const centerX = Math.round(w / 2) + panOffsetX;
    const centerY = Math.round(h / 2) + panOffsetY;

    function toPixel(worldX, worldY) {
      return {
        px: Math.round(centerX + (worldX - cx) * CELL_SPACING),
        py: Math.round(centerY + (worldY - cy) * CELL_SPACING),
      };
    }

    const cullMargin = CELL_SPACING + ROOM_SIZE;
    const drawableIds = visibleRoomIds.filter(id => {
      const room = rooms[id];
      const { px, py } = toPixel(room.coords.x, room.coords.y);
      return px + cullMargin >= 0 && px - cullMargin <= w &&
             py + cullMargin >= 0 && py - cullMargin <= h;
    });

    if (drawableIds.length === 0) {
      ctx.restore();
      return;
    }

    const drawableSet = new Set(drawableIds);

    roomsByPixel = [];

    // Draw connections
    for (const id of drawableIds) {
      const room = rooms[id];
      const dist = distances[id];
      const { px: x1, py: y1 } = toPixel(room.coords.x, room.coords.y);

      for (const exit of room.cardinalExits || []) {
        const target = rooms[exit.targetId];
        if (!target || !target.coords || target.coords.z !== currentZ) continue;
        if (!drawableSet.has(exit.targetId)) continue;

        const isPathConnection = travelPathRoomIds.has(id) || travelPathRoomIds.has(exit.targetId);
        const isNearby = dist !== undefined && dist <= 2;

        const { px: x2, py: y2 } = toPixel(target.coords.x, target.coords.y);

        ctx.save();
        ctx.globalAlpha = 1.0;
        ctx.strokeStyle = isPathConnection ? COLOR_PATH : (isNearby ? COLOR_LINE : COLOR_LINE_DIM);
        ctx.lineWidth = isPathConnection ? 3 : LINE_WIDTH;
        ctx.beginPath();
        const lw = isPathConnection ? 3 : LINE_WIDTH;
        const offset = lw % 2 === 0 ? 0 : 0.5;
        ctx.moveTo(x1 + offset, y1 + offset);
        ctx.lineTo(x2 + offset, y2 + offset);
        ctx.stroke();
        ctx.restore();
      }
    }

    // Draw rooms
    for (const id of drawableIds) {
      const room = rooms[id];
      const dist = distances[id];

      const { px, py } = toPixel(room.coords.x, room.coords.y);
      const half = ROOM_SIZE / 2;
      const rx = px - half;
      const ry = py - half;
      const isCurrent = id === curId;
      const isTarget = id === travelTargetId;
      const isOnPath = travelPathRoomIds.has(id);
      const isNearby = !isCurrent && !isTarget && dist !== undefined && dist <= 2;

      ctx.save();
      ctx.globalAlpha = 1.0;

      if (isCurrent) {
        ctx.shadowColor = COLOR_CURRENT;
        ctx.shadowBlur = 12;
      } else if (isTarget) {
        ctx.shadowColor = COLOR_TARGET;
        ctx.shadowBlur = 12;
      }

      let fillColor = COLOR_ROOM_FILL;
      if (isCurrent) fillColor = COLOR_CURRENT_FILL;
      else if (isTarget) fillColor = COLOR_TARGET_FILL;
      else if (isNearby) fillColor = COLOR_NEARBY_FILL;
      ctx.fillStyle = fillColor;
      roundRect(ctx, rx, ry, ROOM_SIZE, ROOM_SIZE, ROOM_RADIUS);
      ctx.fill();

      ctx.shadowBlur = 0;
      let borderColor = COLOR_ROOM_BORDER;
      if (isCurrent) borderColor = COLOR_CURRENT;
      else if (isTarget) borderColor = COLOR_TARGET;
      else if (isOnPath) borderColor = COLOR_PATH;
      else if (isNearby) borderColor = COLOR_NEARBY_BORDER;
      ctx.strokeStyle = borderColor;

      const thick = isCurrent || isTarget;
      ctx.lineWidth = thick ? 2 : 1;
      if (!thick) {
        roundRect(ctx, rx + 0.5, ry + 0.5, ROOM_SIZE - 1, ROOM_SIZE - 1, ROOM_RADIUS);
      } else {
        roundRect(ctx, rx, ry, ROOM_SIZE, ROOM_SIZE, ROOM_RADIUS);
      }
      ctx.stroke();

      // Up/down indicators
      const hasUp = (room.cardinalExits || []).some(e => e.dir === 'up');
      const hasDown = (room.cardinalExits || []).some(e => e.dir === 'down');
      if (hasUp || hasDown) {
        ctx.save();
        ctx.font = `bold ${Math.round(ROOM_SIZE * 0.55)}px sans-serif`;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillStyle = '#a78bfa';
        if (hasUp) {
          ctx.fillText('\u25B2', px, ry - 4);
        }
        if (hasDown) {
          ctx.fillText('\u25BC', px, ry + ROOM_SIZE + 5);
        }
        ctx.restore();
      }

      ctx.restore();

      roomsByPixel.push({ x: rx, y: ry, w: ROOM_SIZE, h: ROOM_SIZE, room });
    }

    // Z-level indicator
    ctx.save();
    ctx.font = '11px monospace';
    ctx.textAlign = 'left';
    ctx.textBaseline = 'top';
    ctx.fillStyle = '#6b7280';
    const zLabel = currentZ === 0 ? 'Ground' : currentZ > 0 ? `Floor +${currentZ}` : `Floor ${currentZ}`;
    ctx.fillText(zLabel, 6, 6);
    ctx.restore();

    ctx.restore();
  }

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
    maximized = false;
  });
</script>

<style>
  /* --- Inline widget styles --- */
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
    cursor: grab;
    touch-action: none;
  }

  canvas:active {
    cursor: grabbing;
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

  .maximized-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #4b5563;
    font-size: 12px;
    text-align: center;
  }

  /* --- Modal overlay styles (portaled to body) --- */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 10000;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .modal-panel {
    width: 100%;
    height: 100%;
    max-width: 100%;
    max-height: 100%;
    background: #0d1117;
    border: 1px solid #30363d;
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
  }

  .modal-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    border-bottom: 1px solid #21262d;
    font-size: 13px;
    font-weight: 600;
    color: #c9d1d9;
    flex-shrink: 0;
  }

  .modal-header i {
    font-size: 16px;
    color: #8b949e;
  }

  .modal-header .header-spacer {
    flex: 1;
  }

  .modal-canvas-wrap {
    flex: 1;
    position: relative;
    overflow: hidden;
  }
</style>

<!-- Maximized modal — portaled to document.body -->
{#if maximized}
  <div class="modal-backdrop" use:portal on:click={handleBackdropClick} on:keydown={handleBackdropKeydown}>
    <div class="modal-panel">
      <div class="modal-header">
        <i class="material-icons">map</i>
        Map
        {#if isTraveling}
          <span class="travel-indicator">Traveling...</span>
        {/if}
        <span class="header-spacer"></span>
        {#if panOffsetX !== 0 || panOffsetY !== 0}
          <button class="header-btn" title="Re-center on current room" on:click={recenter}>
            <i class="material-icons" style="font-size: inherit">my_location</i>
          </button>
        {/if}
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
        <button class="header-btn" title="Close map" on:click={toggleMaximize}>
          <i class="material-icons" style="font-size: inherit">close</i>
        </button>
      </div>
      <div class="modal-canvas-wrap" bind:this={container}>
        <canvas
          bind:this={canvas}
          on:pointerdown={handlePointerDown}
          on:pointermove={handlePointerMove}
          on:pointerup={handlePointerUp}
          on:pointerleave={handleMouseLeave}
        ></canvas>
        {#if tooltip.visible}
          <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">
            {tooltip.text}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Inline widget (always present in the grid slot) -->
<div class="minimap-container">
  <div class="minimap-header">
    <button class="header-btn" title="Expand map" on:click={toggleMaximize}>
      <i class="material-icons" style="font-size: inherit">open_in_full</i>
    </button>
    <i class="material-icons">map</i>
    Minimap
    {#if isTraveling && !maximized}
      <span class="travel-indicator">Traveling...</span>
    {/if}
    <span class="header-spacer"></span>
    {#if !maximized && (panOffsetX !== 0 || panOffsetY !== 0)}
      <button class="header-btn" title="Re-center on current room" on:click={recenter}>
        <i class="material-icons" style="font-size: inherit">my_location</i>
      </button>
    {/if}
    {#if isTraveling && !maximized}
      <button class="header-btn cancel" title="Cancel travel" on:click={cancelTravel}>
        <i class="material-icons" style="font-size: inherit">close</i>
      </button>
    {/if}
    {#if !maximized && currentRoomId && Object.keys(visitedRooms).length > 0}
      <button class="header-btn" title="Clear map data" on:click={handleClearMap}>
        <i class="material-icons" style="font-size: inherit">delete_sweep</i>
      </button>
    {/if}
  </div>

  {#if maximized}
    <div class="maximized-placeholder">
      Map is expanded
    </div>
  {:else if currentRoomId && Object.keys(visitedRooms).length > 0}
    <div class="canvas-wrap" bind:this={container}>
      <canvas
        bind:this={canvas}
        on:pointerdown={handlePointerDown}
        on:pointermove={handlePointerMove}
        on:pointerup={handlePointerUp}
        on:pointerleave={handleMouseLeave}
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
