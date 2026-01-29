<script>
  import { createEventDispatcher } from "svelte";

  export let room;
  export let x;
  export let y;
  export let width = 180;
  export let height = 80;
  export let selected = false;
  export let dragging = false;
  export let isTemporary = false;
  export let areaColor = "#888"; // Area color for border

  const dispatch = createEventDispatcher();

  // Check if a cardinal exit exists
  function hasExit(direction) {
    return room?.exits?.some(
      (e) => e.name?.toLowerCase() === direction.toLowerCase()
    );
  }

  function handleMouseDown(event) {
    if (event.button === 0 && !isTemporary) {
      event.stopPropagation();
      dispatch("dragstart", { room, event });
    }
  }

  function handleClick(event) {
    event.stopPropagation();
    if (!isTemporary) {
      dispatch("select", { room });
    }
  }

  // Handle exit handle click for creating new rooms
  function handleExitMouseDown(event, direction) {
    event.stopPropagation();
    if (!hasExit(direction) && !isTemporary) {
      dispatch("exitdragstart", { room, direction, event });
    }
  }

  // Truncate name to fit
  function truncateName(name, maxChars = 24) {
    if (!name) return 'Unnamed';
    if (name.length <= maxChars) return name;
    return name.substring(0, maxChars - 1) + '…';
  }

  // Truncate description for preview
  function truncateDescription(desc, maxChars = 32) {
    if (!desc) return '';
    if (desc.length <= maxChars) return desc;
    return desc.substring(0, maxChars - 1) + '…';
  }

  // Format room ID for display (show last 8 chars if UUID)
  function formatRoomId(id) {
    if (!id) return '';
    if (id.length > 12) {
      return id.substring(0, 8).toUpperCase();
    }
    return id.toUpperCase();
  }

  $: displayName = truncateName(room?.name, 24);
  $: displayDescription = truncateDescription(room?.description, 32);
  $: roomId = formatRoomId(room?.id);

  // Exit handle positions (at edges)
  $: exitHandles = {
    north: { cx: 0, cy: -height/2 },
    south: { cx: 0, cy: height/2 },
    east: { cx: width/2, cy: 0 },
    west: { cx: -width/2, cy: 0 },
  };

  // Non-cardinal exits for badges
  const CARDINAL = ['north', 'south', 'east', 'west'];
  $: specialExits = (room?.exits || []).filter(
    (e) => e.name && !CARDINAL.includes(e.name.toLowerCase())
  );

  function getPortalIcon(exitName) {
    const name = exitName.toLowerCase();
    if (name === 'up') return '↑';
    if (name === 'down') return '↓';
    if (name === 'enter') return '▶';
    if (name === 'portal') return '●';
    if (name === 'northeast') return '↗';
    if (name === 'northwest') return '↖';
    if (name === 'southeast') return '↘';
    if (name === 'southwest') return '↙';
    return '◆';
  }

  function handlePortalClick(event, exit) {
    event.stopPropagation();
    if (exit.target) {
      dispatch("portalclick", { targetId: exit.target, exitName: exit.name });
    }
  }

  // Lighten color for card fill
  function lightenColor(hex, percent = 90) {
    // Return a very light tint of the area color
    return `${hex}15`; // 15 = ~8% opacity in hex
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<g
  class="room-tile"
  class:selected
  class:dragging
  class:temporary={isTemporary}
  transform="translate({x}, {y})"
  on:mousedown={handleMouseDown}
  on:click={handleClick}
  on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleClick(e); }}
  role="button"
  tabindex="0"
>
  <!-- Selection highlight / hover glow -->
  {#if selected}
    <rect
      x={-width/2 - 4}
      y={-height/2 - 4}
      width={width + 8}
      height={height + 8}
      rx="8"
      ry="8"
      class="selection-glow"
      style="stroke: {areaColor}; filter: drop-shadow(0 0 6px {areaColor}66);"
    />
  {/if}

  <!-- Room card background (light colored) -->
  <rect
    x={-width/2}
    y={-height/2}
    {width}
    {height}
    rx="6"
    ry="6"
    class="room-rect"
    style="stroke: {areaColor}; fill: #1a1a2a;"
  />

  <!-- Room ID (top left, monospace) -->
  <text
    x={-width/2 + 8}
    y={-height/2 + 14}
    class="room-id"
    text-anchor="start"
    dominant-baseline="middle"
  >
    {roomId}
  </text>

  <!-- Room name (centered, bold) -->
  <text
    x="0"
    y={-4}
    class="room-name"
    text-anchor="middle"
    dominant-baseline="middle"
  >
    {displayName}
  </text>

  <!-- Description snippet (below name, smaller, muted) -->
  {#if displayDescription}
    <text
      x="0"
      y={16}
      class="room-description"
      text-anchor="middle"
      dominant-baseline="middle"
    >
      {displayDescription}
    </text>
  {/if}

  <!-- Exit connection points (at edges) -->
  {#if !isTemporary}
    {#each Object.entries(exitHandles) as [direction, pos]}
      {@const connected = hasExit(direction)}
      <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
      <circle
        cx={pos.cx}
        cy={pos.cy}
        r={connected ? 5 : 4}
        class="exit-dot"
        class:connected
        class:available={!connected}
        on:mousedown={(e) => handleExitMouseDown(e, direction)}
        on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleExitMouseDown(e, direction); }}
        role="button"
        tabindex="0"
        style={connected ? `fill: ${areaColor};` : ''}
      />
    {/each}
  {/if}

  <!-- Portal badges for non-cardinal exits (bottom left) -->
  {#if !isTemporary && specialExits.length > 0}
    {#each specialExits as exit, i}
      {@const badgeX = -width/2 + 16 + i * 22}
      {@const badgeY = height/2 - 14}
      <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
      <g
        class="portal-badge"
        transform="translate({badgeX}, {badgeY})"
        on:click={(e) => handlePortalClick(e, exit)}
        on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') handlePortalClick(e, exit); }}
        on:mousedown={(e) => e.stopPropagation()}
        role="button"
        tabindex="0"
      >
        <circle r="9" class="portal-badge-bg" />
        <text
          x="0"
          y="1"
          text-anchor="middle"
          dominant-baseline="middle"
          class="portal-badge-icon"
        >
          {getPortalIcon(exit.name)}
        </text>
        <title>{exit.name}</title>
      </g>
    {/each}
  {/if}
</g>

<style>
  .room-tile {
    cursor: pointer;
    user-select: none;
  }

  .room-tile.temporary {
    cursor: default;
    opacity: 0.7;
  }

  .room-tile:focus {
    outline: none;
  }

  .room-rect {
    fill: #1a1a2a;
    stroke-width: 2;
    transition: filter 0.15s, fill 0.15s;
  }

  .room-tile.temporary .room-rect {
    fill: rgba(76, 175, 80, 0.15);
    stroke: #4caf50;
    stroke-dasharray: 4 2;
  }

  .room-tile:hover:not(.temporary) .room-rect {
    filter: drop-shadow(0 0 8px rgba(100, 150, 255, 0.4));
    fill: #222238;
  }

  .room-tile.selected .room-rect {
    stroke-width: 2.5;
    fill: #252540;
  }

  .room-tile.dragging .room-rect {
    stroke-dasharray: 4 2;
    opacity: 0.85;
  }

  .selection-glow {
    fill: none;
    stroke-width: 2;
    stroke-opacity: 0.6;
    pointer-events: none;
  }

  .room-id {
    fill: #666;
    font-size: 10px;
    font-family: 'SF Mono', 'Consolas', 'Monaco', monospace;
    pointer-events: none;
  }

  .room-name {
    fill: #e8e8e8;
    font-size: 12px;
    font-weight: 600;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    pointer-events: none;
  }

  .room-description {
    fill: #888;
    font-size: 9px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    pointer-events: none;
  }

  .exit-dot {
    transition: all 0.15s ease;
  }

  .exit-dot.connected {
    cursor: default;
  }

  .exit-dot.available {
    fill: #3a3a5a;
    cursor: crosshair;
  }

  .exit-dot.available:hover {
    fill: #4caf50;
    r: 6;
  }

  /* Portal badges */
  .portal-badge {
    cursor: pointer;
  }

  .portal-badge-bg {
    fill: #2a2010;
    stroke: #b08050;
    stroke-width: 1.5;
    transition: fill 0.15s, stroke 0.15s;
  }

  .portal-badge:hover .portal-badge-bg {
    fill: #4a3020;
    stroke: #d4a060;
  }

  .portal-badge-icon {
    fill: #c9a066;
    font-size: 10px;
    pointer-events: none;
  }

  .portal-badge:hover .portal-badge-icon {
    fill: #e8c080;
  }
</style>
