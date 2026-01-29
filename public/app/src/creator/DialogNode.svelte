<script>
  import { createEventDispatcher } from "svelte";

  export let node;
  export let x;
  export let y;
  export let width = 200;
  export let height = 100;
  export let selected = false;
  export let isRoot = false;
  export let isExit = false;
  export let isAnswer = false;
  export let hasOptions = false;
  export let hasBackRef = false;
  export let backRefTargets = [];

  const dispatch = createEventDispatcher();

  // Split text into multiple lines (max 3 lines)
  function splitText(text, maxCharsPerLine = 28, maxLines = 3) {
    if (!text) return ["(no text)"];

    const words = text.split(' ');
    const lines = [];
    let currentLine = '';

    for (const word of words) {
      if (lines.length >= maxLines) break;

      const testLine = currentLine ? currentLine + ' ' + word : word;

      if (testLine.length <= maxCharsPerLine) {
        currentLine = testLine;
      } else {
        if (currentLine) {
          lines.push(currentLine);
          currentLine = word;
        } else {
          // Single word longer than max - truncate it
          lines.push(word.slice(0, maxCharsPerLine - 2) + '..');
          currentLine = '';
        }
      }
    }

    // Add remaining text
    if (currentLine && lines.length < maxLines) {
      lines.push(currentLine);
    }

    // Add ellipsis if text was truncated
    if (words.length > 0 && lines.length === maxLines) {
      const remainingWords = words.slice(words.join(' ').indexOf(lines.join(' ').split(' ').pop()) + 1);
      if (remainingWords.length > 0 || currentLine) {
        const lastLine = lines[maxLines - 1];
        if (!lastLine.endsWith('...') && !lastLine.endsWith('..')) {
          lines[maxLines - 1] = lastLine.slice(0, maxCharsPerLine - 3) + '...';
        }
      }
    }

    return lines.length > 0 ? lines : ["(no text)"];
  }

  // Format ID for display (truncate if too long)
  function formatId(id) {
    if (!id) return "unnamed";
    if (id.length <= 18) return id;
    return id.slice(0, 16) + '..';
  }

  function handleClick(event) {
    event.stopPropagation();
    dispatch("select", { node });
  }

  // Get node color based on type
  function getNodeColor() {
    if (isRoot) return { fill: "#1a472a", stroke: "#4caf50" }; // Green for root
    if (isExit) return { fill: "#4a1a1a", stroke: "#f44336" }; // Red for exit
    if (isAnswer) return { fill: "#1a3a4a", stroke: "#2196f3" }; // Blue for answer
    if (hasOptions) return { fill: "#3a3a1a", stroke: "#ff9800" }; // Orange for branching
    return { fill: "#2c3e50", stroke: "#00bcd4" }; // Default cyan
  }

  $: colors = getNodeColor();
  $: textLines = splitText(node?.text);
  $: displayId = formatId(node?.nodeId || node?.id);
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<g
  class="dialog-node"
  class:selected
  class:is-root={isRoot}
  class:is-exit={isExit}
  transform="translate({x}, {y})"
  on:click={handleClick}
  on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleClick(e); }}
  role="button"
  tabindex="0"
>
  <!-- Selection glow -->
  {#if selected}
    <rect
      x={-width/2 - 4}
      y={-height/2 - 4}
      width={width + 8}
      height={height + 8}
      rx="12"
      ry="12"
      class="selection-glow"
      style="stroke: {colors.stroke};"
    />
  {/if}

  <!-- Node rectangle -->
  <rect
    x={-width/2}
    y={-height/2}
    {width}
    {height}
    rx="8"
    ry="8"
    class="node-rect"
    style="fill: {colors.fill}; stroke: {colors.stroke};"
  />

  <!-- Type badge (top left) -->
  <g transform="translate({-width/2 + 8}, {-height/2 + 8})">
    {#if isRoot}
      <rect x="0" y="0" width="40" height="16" rx="3" fill="#4caf50" />
      <text x="20" y="12" class="badge-text" text-anchor="middle">ROOT</text>
    {:else if isExit}
      <rect x="0" y="0" width="32" height="16" rx="3" fill="#f44336" />
      <text x="16" y="12" class="badge-text" text-anchor="middle">EXIT</text>
    {:else if isAnswer}
      <rect x="0" y="0" width="52" height="16" rx="3" fill="#2196f3" />
      <text x="26" y="12" class="badge-text" text-anchor="middle">ANSWER</text>
    {/if}
  </g>

  <!-- Node ID (bold, top right) -->
  <text
    x={width/2 - 8}
    y={-height/2 + 16}
    class="node-id"
    text-anchor="end"
    dominant-baseline="middle"
  >
    {displayId}
  </text>

  <!-- Node text preview (up to 3 lines, starting higher) -->
  {#each textLines as line, i}
    <text
      x="0"
      y={-height/2 + 38 + i * 15}
      class="node-text"
      text-anchor="middle"
      dominant-baseline="middle"
    >
      {line}
    </text>
  {/each}

  <!-- Options count indicator (bottom right) -->
  {#if node?.options?.length > 0}
    <g transform="translate({width/2 - 16}, {height/2 - 16})">
      <circle cx="0" cy="0" r="12" fill="#ff9800" />
      <text x="0" y="1" class="options-count" text-anchor="middle" dominant-baseline="middle">
        {node.options.length}
      </text>
    </g>
  {/if}

  <!-- Answer indicator (bottom right, only if no options) -->
  {#if node?.answer && !node?.options?.length}
    <g transform="translate({width/2 - 16}, {height/2 - 16})">
      <circle cx="0" cy="0" r="10" fill="#2196f3" />
      <text x="0" y="1" class="answer-indicator" text-anchor="middle" dominant-baseline="middle">A</text>
    </g>
  {/if}

  <!-- Back-reference indicator (bottom left) -->
  {#if hasBackRef && backRefTargets.length > 0}
    <g transform="translate({-width/2 + 8}, {height/2 - 20})">
      <rect x="0" y="0" width="42" height="14" rx="3" fill="#9c27b0" />
      <text x="21" y="10" class="badge-text" text-anchor="middle">GOTO</text>
    </g>
  {/if}

  <!-- Bottom handle for children -->
  {#if hasOptions || node?.answer}
    <circle
      cx="0"
      cy={height/2}
      r="6"
      class="output-handle"
      style="fill: {colors.stroke};"
    />
  {/if}

  <!-- Top handle (except for root) -->
  {#if !isRoot}
    <circle
      cx="0"
      cy={-height/2}
      r="6"
      class="input-handle"
    />
  {/if}
</g>

<style>
  .dialog-node {
    cursor: pointer;
    user-select: none;
  }

  .dialog-node:focus {
    outline: none;
  }

  .node-rect {
    stroke-width: 2;
    transition: all 0.15s ease;
  }

  .dialog-node:hover .node-rect {
    filter: drop-shadow(0 4px 12px rgba(0, 188, 212, 0.4));
  }

  .dialog-node.selected .node-rect {
    stroke-width: 3;
  }

  .selection-glow {
    fill: none;
    stroke-width: 3;
    opacity: 0.5;
    pointer-events: none;
  }

  .node-id {
    fill: #fff;
    font-size: 11px;
    font-weight: 700;
    pointer-events: none;
  }

  .node-text {
    fill: #ccc;
    font-size: 11px;
    font-style: italic;
    pointer-events: none;
  }

  .badge-text {
    fill: #fff;
    font-size: 9px;
    font-weight: 700;
    pointer-events: none;
  }

  .options-count {
    fill: #000;
    font-size: 11px;
    font-weight: 700;
    pointer-events: none;
  }

  .answer-indicator {
    fill: #fff;
    font-size: 10px;
    font-weight: 700;
    pointer-events: none;
  }

  .output-handle, .input-handle {
    stroke: #1a1a1a;
    stroke-width: 2;
    pointer-events: none;
  }

  .input-handle {
    fill: #666;
  }
</style>
