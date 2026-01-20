<script>
  import { createEventDispatcher } from "svelte";

  export let room;
  export let x;
  export let y;
  export let width = 120;
  export let height = 80;
  export let selected = false;
  export let dragging = false;
  export let isTemporary = false; // For rooms being created but not yet saved

  const dispatch = createEventDispatcher();

  // Check if a cardinal exit exists
  function hasExit(direction) {
    return room?.exits?.some(
      (e) => e.name?.toLowerCase() === direction.toLowerCase()
    );
  }

  function handleMouseDown(event) {
    if (event.button === 0 && !isTemporary) { // Left click only, not for temporary rooms
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
    // Only allow drag-to-create if this exit doesn't already exist and room is not temporary
    if (!hasExit(direction) && !isTemporary) {
      dispatch("exitdragstart", { room, direction, event });
    }
  }

  // Truncate name to fit in two lines
  function truncateName(name, maxChars = 24) {
    if (!name) return 'Unnamed';
    if (name.length <= maxChars) return name;
    return name.substring(0, maxChars - 2) + '...';
  }

  // Split name into two lines if needed
  function splitName(name) {
    const truncated = truncateName(name, 28);
    const words = truncated.split(' ');

    if (words.length === 1) {
      // Single long word - split at middle
      if (truncated.length > 14) {
        return [truncated.substring(0, 14), truncated.substring(14)];
      }
      return [truncated];
    }

    // Try to split words evenly across two lines
    let line1 = '';
    let line2 = '';
    let currentLine = 1;

    for (const word of words) {
      if (currentLine === 1) {
        if (line1.length + word.length + 1 <= 14) {
          line1 += (line1 ? ' ' : '') + word;
        } else {
          currentLine = 2;
          line2 = word;
        }
      } else {
        if (line2.length + word.length + 1 <= 14) {
          line2 += (line2 ? ' ' : '') + word;
        }
        // else: drop remaining words (truncated)
      }
    }

    return line2 ? [line1, line2] : [line1];
  }

  $: nameLines = splitName(room?.name);

  // Exit handle positions (OUTSIDE the rectangle)
  $: exitHandles = {
    north: { cx: 0, cy: -height/2 - 12 },
    south: { cx: 0, cy: height/2 + 12 },
    east: { cx: width/2 + 12, cy: 0 },
    west: { cx: -width/2 - 12, cy: 0 },
  };
</script>

<g
  class="room-tile"
  class:selected
  class:dragging
  class:temporary={isTemporary}
  transform="translate({x}, {y})"
  on:mousedown={handleMouseDown}
  on:click={handleClick}
  role="button"
  tabindex="0"
>
  <!-- Selection glow (rendered first/behind) -->
  {#if selected}
    <rect
      x={-width/2 - 4}
      y={-height/2 - 4}
      width={width + 8}
      height={height + 8}
      rx="10"
      ry="10"
      class="selection-glow"
    />
  {/if}

  <!-- Room rectangle -->
  <rect
    x={-width/2}
    y={-height/2}
    {width}
    {height}
    rx="6"
    ry="6"
    class="room-rect"
  />

  <!-- Room name (up to 2 lines) -->
  {#if nameLines.length === 1}
    <text
      x="0"
      y="-6"
      class="room-name"
      text-anchor="middle"
      dominant-baseline="middle"
    >
      {nameLines[0]}
    </text>
  {:else}
    <text
      x="0"
      y="-14"
      class="room-name"
      text-anchor="middle"
      dominant-baseline="middle"
    >
      {nameLines[0]}
    </text>
    <text
      x="0"
      y="2"
      class="room-name"
      text-anchor="middle"
      dominant-baseline="middle"
    >
      {nameLines[1]}
    </text>
  {/if}

  <!-- Area label -->
  {#if room?.area}
    <text
      x="0"
      y={nameLines.length === 1 ? 10 : 18}
      class="room-area"
      text-anchor="middle"
      dominant-baseline="middle"
    >
      {room.area}
    </text>
  {/if}

  <!-- Coordinates -->
  <text
    x="0"
    y={nameLines.length === 1 ? 26 : 32}
    class="room-coords"
    text-anchor="middle"
    dominant-baseline="middle"
  >
    ({room?.coords?.x ?? 0}, {room?.coords?.y ?? 0})
  </text>

  <!-- Exit handles (outside rectangle) - not shown for temporary rooms -->
  {#if !isTemporary}
    {#each Object.entries(exitHandles) as [direction, pos]}
      {@const connected = hasExit(direction)}
      <circle
        cx={pos.cx}
        cy={pos.cy}
        r="6"
        class="exit-handle"
        class:connected
        class:available={!connected}
        on:mousedown={(e) => handleExitMouseDown(e, direction)}
        role="button"
        tabindex="0"
      />
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
    opacity: 0.8;
  }

  .room-tile:focus {
    outline: none;
  }

  .room-rect {
    fill: #2c3e50;
    stroke: #00bcd4;
    stroke-width: 2;
    transition: fill 0.15s, stroke 0.15s;
  }

  .room-tile.temporary .room-rect {
    fill: rgba(76, 175, 80, 0.3);
    stroke: #4caf50;
    stroke-dasharray: 4 2;
  }

  .room-tile:hover:not(.temporary) .room-rect {
    stroke: #4dd0e1;
    filter: drop-shadow(0 4px 12px rgba(0, 188, 212, 0.3));
  }

  .room-tile.selected .room-rect {
    stroke: #4dd0e1;
    stroke-width: 2.5;
    fill: #34495e;
  }

  .room-tile.selected.temporary .room-rect {
    fill: rgba(76, 175, 80, 0.4);
    stroke: #4caf50;
  }

  .room-tile.dragging .room-rect {
    stroke: #4dd0e1;
    stroke-dasharray: 4 2;
    opacity: 0.8;
  }

  .selection-glow {
    fill: none;
    stroke: rgba(0, 188, 212, 0.4);
    stroke-width: 3;
    pointer-events: none;
  }

  .room-name {
    fill: white;
    font-size: 11px;
    font-weight: 600;
    pointer-events: none;
  }

  .room-area {
    fill: #aaa;
    font-size: 9px;
    pointer-events: none;
  }

  .room-coords {
    fill: #666;
    font-size: 8px;
    font-family: monospace;
    pointer-events: none;
  }

  .exit-handle {
    stroke-width: 2;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .exit-handle.connected {
    fill: #00bcd4;
    stroke: #00bcd4;
    cursor: default;
  }

  .exit-handle.available {
    fill: #1a1a1a;
    stroke: #555;
  }

  .exit-handle.available:hover {
    fill: #4caf50;
    stroke: #4caf50;
    cursor: crosshair;
  }
</style>
