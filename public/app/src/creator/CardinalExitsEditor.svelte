<script>
  import { createEventDispatcher } from "svelte";
  import { CARDINAL_DIRECTIONS } from "./WorldEditorStore.js";

  export let exits = [];
  export let roomsValueHelp = [];

  const dispatch = createEventDispatcher();

  // Reactive map of exits by direction for proper updates when exits prop changes
  $: exitsByDirection = CARDINAL_DIRECTIONS.reduce((map, dir) => {
    const exit = exits?.find(e => e.name && e.name.toLowerCase() === dir.toLowerCase());
    map[dir] = exit || null;
    return map;
  }, {});

  // Get the exit for a direction if it exists
  function getExitForDirection(direction) {
    return exitsByDirection[direction];
  }

  // Update or create exit for a direction
  function handleTargetChange(direction, targetId) {
    const existingExit = getExitForDirection(direction);

    if (targetId === "") {
      // Remove the exit
      if (existingExit) {
        dispatch("removeExit", direction);
      }
    } else if (existingExit) {
      // Update existing exit
      dispatch("updateExit", { direction, target: targetId });
    } else {
      // Create new exit
      dispatch("addExit", {
        name: direction,
        description: "",
        target: targetId,
        exitType: "direction",
        hidden: false,
      });
    }
  }

  // Direction labels and icons
  const directionConfig = {
    north: { label: "North", icon: "north" },
    east: { label: "East", icon: "east" },
    south: { label: "South", icon: "south" },
    west: { label: "West", icon: "west" },
  };
</script>

<div class="cardinal-exits-grid">
  {#each CARDINAL_DIRECTIONS as direction (direction)}
    {@const exit = exitsByDirection[direction]}
    {@const config = directionConfig[direction]}
    <div class="exit-slot" class:has-exit={exit}>
      <div class="exit-header">
        <span class="material-symbols-outlined direction-icon">{config.icon}</span>
        <span class="direction-label">{config.label}</span>
      </div>
      <select
        class="exit-select"
        value={exit ? exit.target : ""}
        on:change={(e) => handleTargetChange(direction, e.target.value)}
      >
        <option value="">No exit</option>
        {#each roomsValueHelp as room}
          <option value={room.id}>{room.name}</option>
        {/each}
      </select>
    </div>
  {/each}
</div>

<style>
  .cardinal-exits-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .exit-slot {
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    padding: 8px;
    transition: all 0.2s ease;
  }

  .exit-slot.has-exit {
    border-color: #4caf50;
    background: rgba(76, 175, 80, 0.1);
  }

  .exit-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 6px;
  }

  .direction-icon {
    font-size: 16px;
    color: #888;
  }

  .exit-slot.has-exit .direction-icon {
    color: #4caf50;
  }

  .direction-label {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: #aaa;
  }

  .exit-slot.has-exit .direction-label {
    color: #4caf50;
  }

  .exit-select {
    width: 100%;
    padding: 6px 8px;
    font-size: 12px;
    background: #2a2a2a;
    border: 1px solid #3a3a3a;
    border-radius: 4px;
    color: #fff;
    cursor: pointer;
  }

  .exit-select:focus {
    outline: none;
    border-color: #00bcd4;
  }

  .exit-select option {
    background: #2a2a2a;
    color: #fff;
  }
</style>
