<script>
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import {
    getDialogs,
    getDialog,
    updateDialog,
  } from "../api/dialogs.js";
  import DialogNode from "./DialogNode.svelte";
  import DialogDetailPanel from "./DialogDetailPanel.svelte";

  const { isLoading, isAuthenticated, authToken } = getAuth();

  // Layout constants
  const NODE_WIDTH = 220;
  const NODE_HEIGHT = 110;
  const LEVEL_SPACING_Y = 170;
  const NODE_SPACING_X = 260;

  // State
  let dialogs = [];
  let loading = true;
  let error = null;

  // Editor state
  let selectedDialogId = null;
  let selectedDialog = null;
  let selectedNode = null;
  let saving = false;

  // Graph data
  let graphNodes = [];
  let graphEdges = [];

  // Viewport state
  let viewBox = { x: -400, y: -50, width: 1200, height: 800 };
  let svgElement;
  let containerElement;

  // Pan state
  let isPanning = false;
  let panStart = { x: 0, y: 0, viewX: 0, viewY: 0 };

  // Load dialogs
  async function loadDialogs() {
    if (!$authToken) return;

    loading = true;
    error = null;

    return new Promise((resolve, reject) => {
      getDialogs(
        $authToken,
        null,
        (data) => {
          dialogs = data || [];
          loading = false;
          resolve();
        },
        (err) => {
          console.error("Error loading dialogs:", err);
          error = "Failed to load dialogs";
          loading = false;
          reject(err);
        }
      );
    });
  }

  // Load a single dialog with full tree
  async function loadDialog(id) {
    if (!$authToken || !id) return;

    return new Promise((resolve, reject) => {
      getDialog(
        $authToken,
        id,
        (data) => {
          selectedDialog = data;
          selectedNode = data;
          buildGraph(data);
          resolve(data);
        },
        (err) => {
          console.error("Error loading dialog:", err);
          reject(err);
        }
      );
    });
  }

  // Build graph from dialog tree
  function buildGraph(dialog) {
    if (!dialog) {
      graphNodes = [];
      graphEdges = [];
      return;
    }

    const nodes = [];
    const edges = [];
    const positions = new Map();

    // Calculate positions using tree layout
    function layoutNode(node, level, indexInLevel, parentX = 0, isAnswer = false) {
      if (!node) return;

      const nodeId = node.nodeId || node.id || `node_${nodes.length}`;

      // Calculate children count for this level
      const childCount = (node.options?.length || 0) + (node.answer ? 1 : 0);
      const totalWidth = childCount > 0 ? (childCount - 1) * NODE_SPACING_X : 0;
      const startX = parentX - totalWidth / 2;

      // Position this node
      const x = level === 0 ? 0 : parentX;
      const y = level * LEVEL_SPACING_Y;

      positions.set(nodeId, { x, y });

      nodes.push({
        id: nodeId,
        data: node,
        x,
        y,
        isRoot: level === 0,
        isExit: node.is_dialog_exit,
        isAnswer: isAnswer,
        hasOptions: (node.options?.length || 0) > 0,
      });

      // Process answer node first (centered below)
      if (node.answer) {
        const answerId = node.answer.nodeId || `answer_${nodeId}`;
        const answerX = childCount > 1 ? startX + (childCount - 1) * NODE_SPACING_X / 2 : x;

        edges.push({
          id: `edge_${nodeId}_${answerId}`,
          source: nodeId,
          target: answerId,
          sourceX: x,
          sourceY: y + NODE_HEIGHT / 2,
          targetX: answerX,
          targetY: y + LEVEL_SPACING_Y - NODE_HEIGHT / 2,
          isAnswer: true,
        });

        layoutNode(node.answer, level + 1, 0, answerX, true);
      }

      // Process options
      if (node.options?.length > 0) {
        const optionStartIndex = node.answer ? 1 : 0;
        node.options.forEach((option, i) => {
          const optionId = option.nodeId || `option_${nodeId}_${i}`;
          const optionX = startX + (i + optionStartIndex) * NODE_SPACING_X;

          edges.push({
            id: `edge_${nodeId}_${optionId}`,
            source: nodeId,
            target: optionId,
            sourceX: x,
            sourceY: y + NODE_HEIGHT / 2,
            targetX: optionX,
            targetY: y + LEVEL_SPACING_Y - NODE_HEIGHT / 2,
            isAnswer: false,
            optionIndex: i + 1,
          });

          layoutNode(option, level + 1, i, optionX, false);
        });
      }
    }

    layoutNode(dialog, 0, 0, 0);

    graphNodes = nodes;
    graphEdges = edges;

    // Auto-fit view after building
    setTimeout(fitView, 100);
  }

  // Handle dialog selection from list
  function handleDialogSelect(dialog) {
    selectedDialogId = dialog.id;
    loadDialog(dialog.id);
  }

  // Handle node selection in graph
  function handleNodeSelect(event) {
    const { node } = event.detail;
    selectedNode = node;
  }

  // Handle node selection from panel
  function handlePanelNodeSelect(event) {
    const node = event.detail;
    selectedNode = node;
    // Rebuild graph to reflect any structural changes
    buildGraph(selectedDialog);
  }

  // Handle save
  async function handleSave(event) {
    const dialogData = event.detail;
    if (!$authToken || !dialogData?.id) return;

    saving = true;

    return new Promise((resolve, reject) => {
      updateDialog(
        $authToken,
        dialogData.id,
        dialogData,
        async () => {
          await loadDialogs();
          buildGraph(dialogData);
          saving = false;
          resolve();
        },
        (err) => {
          console.error("Error saving dialog:", err);
          saving = false;
          reject(err);
        }
      );
    });
  }

  // Handle close panel
  function handleClosePanel() {
    selectedDialogId = null;
    selectedDialog = null;
    selectedNode = null;
    graphNodes = [];
    graphEdges = [];
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

  // Handle pan start
  function handleMouseDown(event) {
    if (event.button === 1 || (event.button === 0 && event.ctrlKey)) {
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

  // Handle pan move
  function handleMouseMove(event) {
    if (isPanning && containerElement) {
      const dx = (event.clientX - panStart.x) * (viewBox.width / containerElement.clientWidth);
      const dy = (event.clientY - panStart.y) * (viewBox.height / containerElement.clientHeight);

      viewBox = {
        ...viewBox,
        x: panStart.viewX - dx,
        y: panStart.viewY - dy,
      };
    }
  }

  // Handle pan end
  function handleMouseUp() {
    isPanning = false;
  }

  // Handle zoom
  function handleWheel(event) {
    event.preventDefault();
    const mousePos = getMouseSVGPosition(event);

    const zoomFactor = event.deltaY > 0 ? 1.1 : 0.9;
    const newWidth = viewBox.width * zoomFactor;
    const newHeight = viewBox.height * zoomFactor;

    if (newWidth < 400 || newWidth > 4000) return;

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
    if (graphNodes.length === 0) {
      viewBox = { x: -400, y: -50, width: 1200, height: 800 };
      return;
    }

    const minX = Math.min(...graphNodes.map(n => n.x));
    const maxX = Math.max(...graphNodes.map(n => n.x));
    const minY = Math.min(...graphNodes.map(n => n.y));
    const maxY = Math.max(...graphNodes.map(n => n.y));

    const padding = 100;
    const width = Math.max(maxX - minX + NODE_WIDTH + padding * 2, 800);
    const height = Math.max(maxY - minY + NODE_HEIGHT + padding * 2, 600);

    viewBox = {
      x: minX - NODE_WIDTH / 2 - padding,
      y: minY - NODE_HEIGHT / 2 - padding,
      width,
      height,
    };
  }

  // Handle pane click (deselect node but keep dialog)
  function handlePaneClick(event) {
    if (event.target === svgElement || event.target.classList.contains('grid-background')) {
      selectedNode = selectedDialog;
    }
  }

  // Generate grid lines
  function getGridLines() {
    const lines = [];
    const gridSize = 50;
    const startX = Math.floor(viewBox.x / gridSize) * gridSize;
    const startY = Math.floor(viewBox.y / gridSize) * gridSize;
    const endX = viewBox.x + viewBox.width;
    const endY = viewBox.y + viewBox.height;

    for (let x = startX; x <= endX; x += gridSize) {
      lines.push({ x1: x, y1: viewBox.y - 100, x2: x, y2: endY + 100, isMajor: x === 0 });
    }

    for (let y = startY; y <= endY; y += gridSize) {
      lines.push({ x1: viewBox.x - 100, y1: y, x2: endX + 100, y2: y, isMajor: y === 0 });
    }

    return lines;
  }

  $: gridLines = getGridLines(viewBox);

  // Auto-load when auth ready
  let hasLoaded = false;
  $: if (!$isLoading && $isAuthenticated && $authToken && !hasLoaded) {
    hasLoaded = true;
    loadDialogs();
  }
</script>

<div class="row">
  <div class="header">
    <h4>Visual Dialog Editor</h4>
    <div class="header-actions">
      <span class="header-hint">
        Select a dialog from the list to visualize its tree structure.
        <strong style="color: #00bcd4; margin-left: 8px;">
          [{dialogs.length} dialogs]
        </strong>
      </span>
      <button class="icon-btn" on:click={fitView} title="Fit view">
        <span class="material-symbols-outlined">fit_screen</span>
      </button>
    </div>
  </div>

  {#if $isAuthenticated}
    {#if loading}
      <div class="loading">Loading dialogs...</div>
    {:else if error}
      <div class="loading" style="color: #f44336;">{error}</div>
    {:else}
      <div class="editor-container" bind:this={containerElement}>
        <!-- Dialog List Sidebar -->
        <div class="dialog-list">
          <div class="list-header">
            <span class="list-title">Dialogs</span>
          </div>
          <div class="list-content">
            {#each dialogs as dialog}
              <button
                class="dialog-item"
                class:selected={selectedDialogId === dialog.id}
                on:click={() => handleDialogSelect(dialog)}
              >
                <span class="dialog-name">{dialog.name || dialog.nodeId || "Unnamed"}</span>
                {#if dialog.options?.length > 0}
                  <span class="options-badge">{dialog.options.length}</span>
                {/if}
              </button>
            {/each}
            {#if dialogs.length === 0}
              <p class="empty-list">No dialogs created yet.</p>
            {/if}
          </div>
        </div>

        <!-- Graph Area -->
        <div class="graph-area">
          {#if selectedDialog}
            <svg
              bind:this={svgElement}
              class="graph-svg"
              viewBox="{viewBox.x} {viewBox.y} {viewBox.width} {viewBox.height}"
              on:mousedown={handleMouseDown}
              on:mousemove={handleMouseMove}
              on:mouseup={handleMouseUp}
              on:mouseleave={handleMouseUp}
              on:wheel={handleWheel}
              on:click={handlePaneClick}
            >
              <!-- Grid -->
              <g class="grid-lines">
                {#each gridLines as line}
                  <line
                    x1={line.x1}
                    y1={line.y1}
                    x2={line.x2}
                    y2={line.y2}
                    stroke={line.isMajor ? '#3a3a3a' : '#2a2a2a'}
                    stroke-width={line.isMajor ? 2 : 1}
                    class="grid-background"
                  />
                {/each}
              </g>

              <!-- Edges -->
              <g class="edges">
                {#each graphEdges as edge}
                  <!-- Curved path from source to target -->
                  {@const midY = (edge.sourceY + edge.targetY) / 2}
                  <path
                    d="M {edge.sourceX} {edge.sourceY}
                       C {edge.sourceX} {midY},
                         {edge.targetX} {midY},
                         {edge.targetX} {edge.targetY}"
                    fill="none"
                    stroke={edge.isAnswer ? '#2196f3' : '#ff9800'}
                    stroke-width="2"
                    class="edge-path"
                  />
                  <!-- Option number label -->
                  {#if edge.optionIndex}
                    <circle
                      cx={(edge.sourceX + edge.targetX) / 2}
                      cy={midY}
                      r="12"
                      fill="#1a1a1a"
                      stroke="#ff9800"
                      stroke-width="2"
                    />
                    <text
                      x={(edge.sourceX + edge.targetX) / 2}
                      y={midY}
                      text-anchor="middle"
                      dominant-baseline="middle"
                      class="edge-label"
                    >
                      {edge.optionIndex}
                    </text>
                  {/if}
                  {#if edge.isAnswer}
                    <text
                      x={(edge.sourceX + edge.targetX) / 2}
                      y={midY - 16}
                      text-anchor="middle"
                      class="answer-label"
                    >
                      AUTO
                    </text>
                  {/if}
                {/each}
              </g>

              <!-- Nodes -->
              <g class="nodes">
                {#each graphNodes as node (node.id)}
                  <DialogNode
                    node={node.data}
                    x={node.x}
                    y={node.y}
                    width={NODE_WIDTH}
                    height={NODE_HEIGHT}
                    selected={selectedNode === node.data || selectedNode?.nodeId === node.data?.nodeId}
                    isRoot={node.isRoot}
                    isExit={node.isExit}
                    isAnswer={node.isAnswer}
                    hasOptions={node.hasOptions}
                    on:select={handleNodeSelect}
                  />
                {/each}
              </g>
            </svg>

            <!-- Zoom controls -->
            <div class="controls">
              <button on:click={() => viewBox = { ...viewBox, width: viewBox.width * 0.8, height: viewBox.height * 0.8 }}>
                <span class="material-symbols-outlined">add</span>
              </button>
              <button on:click={() => viewBox = { ...viewBox, width: viewBox.width * 1.2, height: viewBox.height * 1.2 }}>
                <span class="material-symbols-outlined">remove</span>
              </button>
              <button on:click={fitView}>
                <span class="material-symbols-outlined">fit_screen</span>
              </button>
            </div>

            <!-- Legend -->
            <div class="legend">
              <div class="legend-item">
                <span class="legend-color" style="background: #4caf50;"></span>
                <span>Root Node</span>
              </div>
              <div class="legend-item">
                <span class="legend-color" style="background: #ff9800;"></span>
                <span>Branch</span>
              </div>
              <div class="legend-item">
                <span class="legend-color" style="background: #2196f3;"></span>
                <span>Answer</span>
              </div>
              <div class="legend-item">
                <span class="legend-color" style="background: #f44336;"></span>
                <span>Exit</span>
              </div>
            </div>
          {:else}
            <div class="empty-state">
              <span class="material-symbols-outlined" style="font-size: 64px; color: #333;">account_tree</span>
              <p>Select a dialog from the list to view its structure</p>
            </div>
          {/if}
        </div>

        <!-- Detail Panel -->
        <DialogDetailPanel
          dialog={selectedDialog}
          selectedNode={selectedNode}
          {saving}
          on:save={handleSave}
          on:close={handleClosePanel}
          on:selectNode={handlePanelNodeSelect}
        />
      </div>
    {/if}
  {:else}
    <p>Please log in to view dialogs.</p>
  {/if}
</div>

<style>
  .editor-container {
    display: flex;
    height: 80vh;
    background: #1a1a1a;
    border-radius: 8px;
    overflow: hidden;
  }

  .dialog-list {
    width: 220px;
    min-width: 220px;
    background: #1e1e1e;
    border-right: 1px solid #333;
    display: flex;
    flex-direction: column;
  }

  .list-header {
    padding: 16px;
    border-bottom: 1px solid #333;
  }

  .list-title {
    font-size: 12px;
    font-weight: 600;
    color: #00bcd4;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .list-content {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }

  .dialog-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 10px 12px;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 6px;
    color: #aaa;
    font-size: 13px;
    text-align: left;
    cursor: pointer;
    transition: all 0.15s ease;
    margin-bottom: 4px;
  }

  .dialog-item:hover {
    background: #2a2a2a;
    color: #fff;
  }

  .dialog-item.selected {
    background: #2a3a4a;
    border-color: #00bcd4;
    color: #fff;
  }

  .dialog-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .options-badge {
    font-size: 10px;
    padding: 2px 6px;
    background: #ff9800;
    border-radius: 10px;
    color: #000;
    font-weight: 600;
  }

  .empty-list {
    font-size: 12px;
    color: #555;
    text-align: center;
    padding: 20px;
  }

  .graph-area {
    flex: 1;
    position: relative;
    min-width: 0;
    overflow: hidden;
  }

  .graph-svg {
    width: 100%;
    height: 100%;
    background: #1a1a1a;
    cursor: grab;
  }

  .graph-svg:active {
    cursor: grabbing;
  }

  .grid-lines line {
    pointer-events: none;
  }

  .edge-path {
    pointer-events: none;
  }

  .edge-label {
    fill: #ff9800;
    font-size: 11px;
    font-weight: 700;
    pointer-events: none;
  }

  .answer-label {
    fill: #2196f3;
    font-size: 9px;
    font-weight: 700;
    pointer-events: none;
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

  .legend {
    position: absolute;
    right: 16px;
    bottom: 16px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    background: rgba(30, 30, 30, 0.9);
    border: 1px solid #333;
    border-radius: 6px;
    padding: 12px;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 11px;
    color: #888;
  }

  .legend-color {
    width: 12px;
    height: 12px;
    border-radius: 3px;
  }

  .empty-state {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    text-align: center;
    color: #555;
  }

  .empty-state p {
    margin: 16px 0 0;
    font-size: 14px;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .header h4 {
    margin: 0;
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

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 80vh;
    color: #00bcd4;
    font-size: 18px;
  }
</style>
