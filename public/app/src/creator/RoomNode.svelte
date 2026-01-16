<script>
  import { Handle, Position } from "@xyflow/svelte";

  export let data;
  export let selected = false;

  // Get area-based color
  function getAreaColor(areaType) {
    const colors = {
      forest: "#2d5016",
      town: "#4a4a4a",
      dungeon: "#1a1a2e",
      cave: "#3d2817",
      water: "#1e3a5f",
      mountain: "#5a4a3a",
      default: "#2c3e50",
    };
    return colors[areaType] || colors.default;
  }

  $: backgroundColor = getAreaColor(data.room && data.room.areaType);
  $: borderColor = selected ? "#00bcd4" : "#555";
</script>

<style>
  .room-node {
    padding: 12px 16px;
    border-radius: 8px;
    border: 2px solid;
    background: var(--bg-color);
    border-color: var(--border-color);
    min-width: 150px;
    max-width: 200px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .room-node:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    transform: translateY(-2px);
  }

  .room-name {
    font-weight: 600;
    font-size: 14px;
    color: #fff;
    margin-bottom: 4px;
    word-wrap: break-word;
  }

  .room-area {
    font-size: 11px;
    color: #aaa;
    font-style: italic;
  }

  .room-type {
    font-size: 10px;
    color: #888;
    margin-top: 2px;
  }
</style>

<div
  class="room-node"
  style="--bg-color: {backgroundColor}; --border-color: {borderColor}"
>
  <Handle type="target" position={Position.Top} />
  <Handle type="target" position={Position.Left} />
  <Handle type="target" position={Position.Right} />
  <Handle type="target" position={Position.Bottom} />

  <div class="room-name">{data.label}</div>
  {#if data.room && data.room.area}
    <div class="room-area">{data.room.area}</div>
  {/if}
  {#if data.room && data.room.roomType}
    <div class="room-type">{data.room.roomType}</div>
  {/if}

  <Handle type="source" position={Position.Top} />
  <Handle type="source" position={Position.Left} />
  <Handle type="source" position={Position.Right} />
  <Handle type="source" position={Position.Bottom} />
</div>
