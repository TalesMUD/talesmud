<style>
  .world-container {
    width: 100%;
    height: 80vh;
    position: relative;
    background: #1a1a1a;
    border-radius: 8px;
    overflow: hidden;
  }

  .room-panel {
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    width: 320px;
    background: #2c2c2c;
    border-left: 2px solid #444;
    padding: 20px;
    overflow-y: auto;
    z-index: 10;
    box-shadow: -4px 0 12px rgba(0, 0, 0, 0.5);
    transform: translateX(100%);
    transition: transform 0.3s ease;
  }

  .room-panel.visible {
    transform: translateX(0);
  }

  .room-panel h3 {
    margin-top: 0;
    color: #00bcd4;
    font-size: 18px;
    border-bottom: 2px solid #444;
    padding-bottom: 10px;
  }

  .room-description {
    color: #ccc;
    font-size: 14px;
    line-height: 1.6;
    margin-bottom: 20px;
  }

  .exits-section {
    margin-top: 20px;
  }

  .exits-section h4 {
    color: #00bcd4;
    font-size: 14px;
    margin-bottom: 10px;
  }

  .exit-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .exit-item {
    padding: 10px;
    margin-bottom: 8px;
    background: #3a3a3a;
    border-radius: 4px;
    border-left: 3px solid #00bcd4;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .exit-item:hover {
    background: #454545;
    transform: translateX(4px);
  }

  .exit-item.cardinal {
    border-left-color: #4caf50;
  }

  .exit-item.hidden {
    border-left-color: #ff9800;
  }

  .exit-name {
    font-weight: 600;
    color: #fff;
    font-size: 13px;
  }

  .exit-description {
    font-size: 11px;
    color: #aaa;
    margin-top: 4px;
  }

  .room-info {
    margin-bottom: 15px;
  }

  .room-info-label {
    font-size: 11px;
    color: #888;
    text-transform: uppercase;
    margin-bottom: 4px;
  }

  .room-info-value {
    font-size: 13px;
    color: #fff;
    padding: 6px 10px;
    background: #3a3a3a;
    border-radius: 4px;
  }

  .close-panel {
    position: absolute;
    top: 10px;
    right: 10px;
    background: transparent;
    border: none;
    color: #888;
    font-size: 24px;
    cursor: pointer;
    padding: 5px 10px;
    line-height: 1;
  }

  .close-panel:hover {
    color: #fff;
  }

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 80vh;
    color: #00bcd4;
    font-size: 18px;
  }

  :global(.svelte-flow) {
    background: #1a1a1a;
  }

  :global(.svelte-flow__edge-path) {
    stroke: #555;
    stroke-width: 2;
  }

  :global(.svelte-flow__edge.selected .svelte-flow__edge-path) {
    stroke: #00bcd4;
  }

  :global(.svelte-flow__edge-text) {
    fill: #aaa;
    font-size: 11px;
  }
</style>

<script>
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import { getWorldGraph } from "../api/world.js";
  import {
    SvelteFlow,
    Background,
    Controls,
    MiniMap,
    BackgroundVariant,
  } from "@xyflow/svelte";
  import RoomNode from "./RoomNode.svelte";

  export let location;

  const {
    isLoading,
    isAuthenticated,
    authToken,
    authError,
    userInfo,
  } = getAuth();

  let nodes = writable([]);
  let edges = writable([]);
  let selectedRoom = null;
  let showPanel = false;
  let loading = true;
  let error = null;

  const nodeTypes = {
    room: RoomNode,
  };

  // Transform backend graph data to Svelte Flow format
  function transformGraphData(graphData) {
    const flowNodes = graphData.nodes.map((node) => ({
      id: node.id,
      type: "room",
      position: { x: node.x, y: node.y },
      data: {
        label: node.name,
        room: {
          id: node.id,
          name: node.name,
          description: node.description,
          area: node.area,
          areaType: node.areaType,
          roomType: node.roomType,
          z: node.z,
        },
      },
    }));

    // Only show cardinal direction edges initially
    const flowEdges = graphData.edges
      .filter((edge) => edge.isCardinal)
      .map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        label: edge.label,
        type: "smoothstep",
        animated: false,
        style: "stroke: #555; stroke-width: 2px;",
      }));

    // Store all exits in node data for panel display
    const exitsByRoom = {};
    graphData.edges.forEach((edge) => {
      if (!exitsByRoom[edge.source]) {
        exitsByRoom[edge.source] = [];
      }
      exitsByRoom[edge.source].push({
        name: edge.label,
        target: edge.target,
        isCardinal: edge.isCardinal,
        isHidden: edge.isHidden,
        exitType: edge.exitType,
      });
    });

    // Add exits to node data
    flowNodes.forEach((node) => {
      node.data.exits = exitsByRoom[node.id] || [];
    });

    return { nodes: flowNodes, edges: flowEdges };
  }

  function onNodeClick(event) {
    const nodeId = event.detail.node.id;
    const node = $nodes.find((n) => n.id === nodeId);
    if (node) {
      selectedRoom = node.data;
      showPanel = true;
    }
  }

  function closePanel() {
    showPanel = false;
    selectedRoom = null;
  }

  function onExitClick(exit) {
    // Find and select the target room
    const targetNode = $nodes.find((n) => n.id === exit.target);
    if (targetNode) {
      selectedRoom = targetNode.data;
      showPanel = true;
    }
  }

  $: {
    if (!$isLoading && $isAuthenticated && $authToken) {
      getWorldGraph(
        $authToken,
        (graphData) => {
          const { nodes: flowNodes, edges: flowEdges } =
            transformGraphData(graphData);
          nodes.set(flowNodes);
          edges.set(flowEdges);
          loading = false;
        },
        (err) => {
          console.error("Error loading world graph:", err);
          error = "Failed to load world map";
          loading = false;
        }
      );
    }
  }

  onMount(async () => {});
</script>

<div class="row">
  <h4>Interactive World Map</h4>

  {#if $isAuthenticated}
    {#if loading}
      <div class="loading">Loading world map...</div>
    {:else if error}
      <div class="loading" style="color: #f44336;">{error}</div>
    {:else}
      <div class="world-container">
        <SvelteFlow
          {nodes}
          {edges}
          {nodeTypes}
          fitView
          on:nodeclick={onNodeClick}
        >
          <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
          <Controls />
          <MiniMap nodeColor="#4a4a4a" />
        </SvelteFlow>

        <div class="room-panel" class:visible={showPanel}>
          <button class="close-panel" on:click={closePanel}>&times;</button>

          {#if selectedRoom}
            <h3>{selectedRoom.label}</h3>

            {#if selectedRoom.room && selectedRoom.room.description}
              <div class="room-description">
                {selectedRoom.room.description}
              </div>
            {/if}

            {#if selectedRoom.room && selectedRoom.room.area}
              <div class="room-info">
                <div class="room-info-label">Area</div>
                <div class="room-info-value">{selectedRoom.room.area}</div>
              </div>
            {/if}

            {#if selectedRoom.room && selectedRoom.room.roomType}
              <div class="room-info">
                <div class="room-info-label">Room Type</div>
                <div class="room-info-value">{selectedRoom.room.roomType}</div>
              </div>
            {/if}

            <div class="exits-section">
              <h4>All Exits</h4>
              {#if selectedRoom.exits && selectedRoom.exits.length > 0}
                <ul class="exit-list">
                  {#each selectedRoom.exits as exit}
                    <li
                      class="exit-item"
                      class:cardinal={exit.isCardinal}
                      class:hidden={exit.isHidden}
                      on:click={() => onExitClick(exit)}
                    >
                      <div class="exit-name">
                        {exit.name}
                        {#if exit.isHidden}
                          <span style="color: #ff9800;">🔒</span>
                        {/if}
                      </div>
                      {#if exit.exitType && exit.exitType !== "direction"}
                        <div class="exit-description">
                          Type: {exit.exitType}
                        </div>
                      {/if}
                    </li>
                  {/each}
                </ul>
              {:else}
                <p style="color: #888; font-size: 13px;">No exits available</p>
              {/if}
            </div>

            <div class="exits-section">
              <h4>Cardinal Directions</h4>
              <ul class="exit-list">
                {#each selectedRoom.exits.filter((e) => e.isCardinal) as exit}
                  <li
                    class="exit-item cardinal"
                    on:click={() => onExitClick(exit)}
                  >
                    <div class="exit-name">{exit.name}</div>
                  </li>
                {:else}
                  <p style="color: #888; font-size: 13px;">
                    No cardinal exits
                  </p>
                {/each}
              </ul>
            </div>

            <div class="exits-section">
              <h4>Special Exits</h4>
              <ul class="exit-list">
                {#each selectedRoom.exits.filter((e) => !e.isCardinal) as exit}
                  <li
                    class="exit-item"
                    class:hidden={exit.isHidden}
                    on:click={() => onExitClick(exit)}
                  >
                    <div class="exit-name">
                      {exit.name}
                      {#if exit.isHidden}
                        <span style="color: #ff9800;">🔒</span>
                      {/if}
                    </div>
                    {#if exit.exitType && exit.exitType !== "direction"}
                      <div class="exit-description">Type: {exit.exitType}</div>
                    {/if}
                  </li>
                {:else}
                  <p style="color: #888; font-size: 13px;">
                    No special exits
                  </p>
                {/each}
              </ul>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  {:else}
    <p>Please log in to view the world map.</p>
  {/if}
</div>
