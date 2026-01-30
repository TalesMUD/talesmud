<script>
  import { onMount, onDestroy, createEventDispatcher, tick } from 'svelte';

  export let onTerminalReady = () => {};
  export let onInput = () => {};

  const dispatch = createEventDispatcher();

  let outputEl;
  let inputEl;
  let lines = [];
  let inputText = '';
  let history = [];
  let historyIndex = -1;
  let tempInput = '';

  const MAX_LINES = 1000;
  let lineCounter = 0;

  // ── Font size configuration ──

  const FONT_SIZES = [
    { label: 'S', size: '0.78rem', key: 'small' },
    { label: 'M', size: '0.92rem', key: 'medium' },
    { label: 'L', size: '1.08rem', key: 'large' },
  ];
  const STORAGE_KEY = 'talesmud_termx_fontsize';

  let fontSizeIndex = 0;

  function loadFontSize() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const idx = FONT_SIZES.findIndex(f => f.key === stored);
        if (idx >= 0) fontSizeIndex = idx;
      }
    } catch (e) { /* ignore */ }
  }

  function cycleFontSize() {
    fontSizeIndex = (fontSizeIndex + 1) % FONT_SIZES.length;
    try {
      localStorage.setItem(STORAGE_KEY, FONT_SIZES[fontSizeIndex].key);
    } catch (e) { /* ignore */ }
  }

  $: currentFontSize = FONT_SIZES[fontSizeIndex].size;
  $: currentFontLabel = FONT_SIZES[fontSizeIndex].label;

  // ── Line classification: detect content type and assign color class ──

  function classifyLine(text) {
    if (!text || !text.trim()) return '';
    const trimmed = text.trim();

    // Room name: [Something] on its own line
    if (/^\[.+\]$/.test(trimmed)) return 'tx-bright';

    // Exit bullet: + [direction]
    if (/^\+\s*\[/.test(trimmed)) return 'tx-cyan';

    // Exit header
    if (trimmed.startsWith('- The visible exits')) return 'tx-cyan';

    // Exits inline: "Exits: [north] [east]"
    if (/^Exits?:/i.test(trimmed)) return 'tx-cyan';

    // Room entities
    if (trimmed.startsWith('- In the room:')) return 'tx-muted';

    // System / feedback
    if (trimmed === 'You look around ...') return 'tx-muted';
    if (trimmed.startsWith('Connected to')) return 'tx-muted';
    if (trimmed === 'Connection Closed.') return 'tx-muted';
    if (trimmed === 'reconnecting ...') return 'tx-muted';
    if (trimmed === 'Please log in to connect to the game.') return 'tx-amber';

    return '';
  }

  // ── Escape HTML then convert **bold** markers to highlighted spans ──

  function escapeHtml(text) {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function renderLine(text) {
    if (!text) return '\u00A0';
    let html = escapeHtml(text);
    // Convert **bold** to bright highlighted spans
    html = html.replace(/\*\*(.+?)\*\*/g, '<span class="tx-highlight">$1</span>');
    return html;
  }

  // ── Add a line to the output buffer ──

  function addLine(text, cls = '') {
    lineCounter++;
    const autoClass = cls || classifyLine(text);
    lines.push({ html: renderLine(text), cls: autoClass, id: lineCounter });
    if (lines.length > MAX_LINES) {
      lines = lines.slice(lines.length - MAX_LINES);
    } else {
      lines = lines; // trigger Svelte reactivity
    }
  }

  async function scrollToBottom() {
    await tick();
    if (outputEl) {
      outputEl.scrollTop = outputEl.scrollHeight;
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter') {
      e.preventDefault();
      const cmd = inputText;
      if (cmd.trim()) {
        addLine('> ' + cmd, 'tx-cmd');
        history = [...history, cmd];
        historyIndex = -1;
        tempInput = '';
        onInput(cmd);
        scrollToBottom();
      }
      inputText = '';
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (history.length > 0) {
        if (historyIndex === -1) {
          tempInput = inputText;
          historyIndex = history.length - 1;
        } else if (historyIndex > 0) {
          historyIndex--;
        }
        inputText = history[historyIndex];
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex !== -1) {
        if (historyIndex < history.length - 1) {
          historyIndex++;
          inputText = history[historyIndex];
        } else {
          historyIndex = -1;
          inputText = tempInput;
        }
      }
    }
  }

  function handleContainerClick() {
    if (inputEl) inputEl.focus();
  }

  onMount(() => {
    loadFontSize();

    // Load Fira Code font if not already present
    if (!document.querySelector('link[href*="Fira+Code"]')) {
      const link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = 'https://fonts.googleapis.com/css2?family=Fira+Code:wght@300;400;500&display=swap';
      document.head.appendChild(link);
    }

    // Create the renderer function that receives game output
    const renderer = (data) => {
      if (data == null) return;
      const parts = String(data).split('\n');
      parts.forEach(part => addLine(part));
      scrollToBottom();
    };

    // Notify parent that terminal is ready
    onTerminalReady(null, renderer);
    dispatch('ready', { renderer });

    // Auto-focus the input
    if (inputEl) inputEl.focus();
  });
</script>

<style>
  /* ═══════════════════════════════════════════════════
     TERMINAL X — Veilspan-style MUD terminal widget
     ═══════════════════════════════════════════════════ */

  .tx-window {
    background: #0c0c0c;
    border: 1px solid rgba(61, 220, 132, 0.08);
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    position: relative;
    font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
  }

  /* ── Scanline overlay ── */
  .tx-scanlines {
    position: absolute;
    inset: 0;
    z-index: 10;
    pointer-events: none;
    background: repeating-linear-gradient(
      0deg,
      transparent,
      transparent 2px,
      rgba(0, 0, 0, 0.03) 2px,
      rgba(0, 0, 0, 0.03) 4px
    );
  }

  /* ── Vignette overlay ── */
  .tx-vignette {
    position: absolute;
    inset: 0;
    z-index: 9;
    pointer-events: none;
    background: radial-gradient(ellipse at center, transparent 50%, rgba(0, 0, 0, 0.4) 100%);
  }

  /* ── Title bar (macOS style) ── */
  .tx-titlebar {
    background: #111820;
    padding: 0.6rem 1rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    border-bottom: 1px solid rgba(61, 220, 132, 0.08);
    z-index: 1;
    flex-shrink: 0;
  }

  .tx-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .tx-dot-red { background: #5c1a1a; }
  .tx-dot-amber { background: #5c4419; }
  .tx-dot-green { background: #1a5c38; }

  .tx-title {
    font-size: 0.6rem;
    color: #3d3a36;
    letter-spacing: 0.1em;
    margin-left: 0.5rem;
    flex: 1;
  }

  /* Font size toggle button in title bar */
  .tx-fontsize-btn {
    font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
    font-size: 0.55rem;
    letter-spacing: 0.08em;
    color: #3d3a36;
    background: transparent;
    border: 1px solid rgba(61, 220, 132, 0.12);
    padding: 0.15rem 0.45rem;
    cursor: pointer;
    transition: color 0.2s, border-color 0.2s;
    line-height: 1;
  }

  .tx-fontsize-btn:hover {
    color: #7ae8a4;
    border-color: rgba(61, 220, 132, 0.3);
  }

  /* ── Terminal body (scrollable output) ── */
  .tx-body {
    flex: 1;
    padding: 1.5rem;
    padding-bottom: 0;
    line-height: 1.9;
    color: #ccc8c2;
    overflow-y: auto;
    overflow-x: hidden;
    z-index: 1;
    min-height: 0;

    /* Thin green-tinted scrollbar */
    scrollbar-width: thin;
    scrollbar-color: #1a5c38 #0c0c0c;
  }

  .tx-body::-webkit-scrollbar {
    width: 6px;
  }

  .tx-body::-webkit-scrollbar-track {
    background: #0c0c0c;
  }

  .tx-body::-webkit-scrollbar-thumb {
    background: #1a5c38;
    border-radius: 3px;
  }

  /* ── Output lines ── */
  .tx-line {
    display: block;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  /* ── Semantic color classes ── */

  /* User command echo */
  .tx-cmd {
    color: #7ae8a4;
  }

  /* Room name — bright white */
  .tx-bright {
    color: #e8e6e3;
    font-weight: 500;
  }

  /* Exits — cyan */
  .tx-cyan {
    color: #8be9fd;
  }

  /* System / entity info — muted */
  .tx-muted {
    color: #6b6660;
  }

  /* Amber — warnings / special */
  .tx-amber {
    color: #f0c674;
  }

  /* Inline **bold** highlight */
  .tx-line :global(.tx-highlight) {
    color: #e8e6e3;
    font-weight: 500;
  }

  /* ── Input area — blends seamlessly with body ── */
  .tx-input-area {
    display: flex;
    align-items: center;
    padding: 0.75rem 1.5rem 1.25rem;
    z-index: 1;
    flex-shrink: 0;
    background: #0c0c0c;
  }

  .tx-prompt {
    color: #7ae8a4;
    margin-right: 0.5rem;
    user-select: none;
    flex-shrink: 0;
  }

  /*
   * Input field — aggressively reset all browser & Materialize CSS
   * focus styles to prevent colored underlines / outlines / shadows.
   */
  .tx-input,
  .tx-input:focus,
  .tx-input:active,
  .tx-input:hover {
    flex: 1;
    background: transparent !important;
    border: none !important;
    border-bottom: none !important;
    outline: none !important;
    box-shadow: none !important;
    -webkit-box-shadow: none !important;
    color: #7ae8a4;
    font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
    line-height: 1.9;
    caret-color: #3ddc84;
    padding: 0 !important;
    margin: 0 !important;
    height: auto !important;
    width: 100%;
    -webkit-appearance: none;
    -moz-appearance: none;
    appearance: none;
  }

  .tx-input::placeholder {
    color: #3d3a36;
  }

  /* Kill Materialize's .input-field wrappers if ancestors apply them */
  .tx-input-area :global(input[type="text"]:focus) {
    border-bottom: none !important;
    box-shadow: none !important;
  }
</style>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div class="tx-window" on:click={handleContainerClick}>
  <!-- CRT overlays -->
  <div class="tx-scanlines"></div>
  <div class="tx-vignette"></div>

  <!-- macOS-style title bar -->
  <div class="tx-titlebar">
    <div class="tx-dot tx-dot-red"></div>
    <div class="tx-dot tx-dot-amber"></div>
    <div class="tx-dot tx-dot-green"></div>
    <span class="tx-title">talesmud &mdash; terminal session</span>
    <button class="tx-fontsize-btn" on:click|stopPropagation={cycleFontSize} title="Font size: {FONT_SIZES[fontSizeIndex].key}">
      {currentFontLabel}
    </button>
  </div>

  <!-- Scrollable output (font-size driven by reactive var) -->
  <div class="tx-body" style="font-size: {currentFontSize};" bind:this={outputEl}>
    {#each lines as line (line.id)}
      <span class="tx-line {line.cls}">{@html line.html}</span>
    {/each}
  </div>

  <!-- Command input (font-size matches body) -->
  <div class="tx-input-area" style="font-size: {currentFontSize};">
    <span class="tx-prompt">&gt;</span>
    <input
      type="text"
      class="tx-input"
      style="font-size: {currentFontSize};"
      bind:this={inputEl}
      bind:value={inputText}
      on:keydown={handleKeydown}
      placeholder="_"
      spellcheck="false"
      autocomplete="off"
    />
  </div>
</div>
