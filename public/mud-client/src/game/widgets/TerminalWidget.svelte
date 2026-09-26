<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { writable } from 'svelte/store';
  import xterm from 'xterm';
  import { FitAddon } from 'xterm-addon-fit';
  import LocalEchoController from '../echo/LocalEchoController';
  import { wrapAnsi } from '../echo/wrapAnsi.js';
  import '../../../node_modules/xterm/css/xterm.css';
  import { hexToRgb } from '../playerColors.js';

  export let onTerminalReady = () => {};
  export let onInput = () => {};

  const dispatch = createEventDispatcher();

  let terminalContainer;
  let term;
  let fitAddon;
  let localEcho;
  let resizeObserver;
  const rawLines = [];
  let paintedCols = 0;

  // ── Font size configuration (S / M / L) ──

  const FONT_SCALES = [
    { label: 'S', scale: 0.7, key: 'small' },
    { label: 'M', scale: 0.85, key: 'medium' },
    { label: 'L', scale: 1.0, key: 'large' },
  ];
  const STORAGE_KEY = 'talesmud_term_fontsize';

  let fontScaleIndex = 0;

  function loadFontScale() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const idx = FONT_SCALES.findIndex(f => f.key === stored);
        if (idx >= 0) fontScaleIndex = idx;
      }
    } catch (e) { /* ignore */ }
  }

  function cycleFontSize() {
    fontScaleIndex = (fontScaleIndex + 1) % FONT_SCALES.length;
    try {
      localStorage.setItem(STORAGE_KEY, FONT_SCALES[fontScaleIndex].key);
    } catch (e) { /* ignore */ }
    applyFontSize();
  }

  function getBaseTerminalFontSize() {
    const width = window.innerWidth;
    if (width >= 3000) return 26;
    if (width >= 2600) return 24;
    if (width >= 2200) return 22;
    if (width >= 1800) return 20;
    if (width >= 1600) return 18;
    if (width >= 1400) return 17;
    return 15;
  }

  function getScaledFontSize() {
    return Math.round(getBaseTerminalFontSize() * FONT_SCALES[fontScaleIndex].scale);
  }

  function fitTerm() {
    if (!fitAddon || !term) return;
    fitAddon.fit();
    const dims = fitAddon.proposeDimensions();
    if (!dims) return;
    // One column short of the measured width so the scrollbar and the panel
    // edge do not clip the last glyph.
    const cols = Math.max(2, dims.cols - 1);
    const rows = Math.max(1, dims.rows);
    if (term.cols !== cols || term.rows !== rows) term.resize(cols, rows);
  }

  function applyFontSize() {
    if (term) {
      term.setOption('fontSize', getScaledFontSize());
      fitTerm();
    }
  }

  $: currentFontLabel = FONT_SCALES[fontScaleIndex].label;

  function debounce(fn, ms) {
    let timeout;
    return (...args) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => fn(...args), ms);
    };
  }

  function restorePrompt() {
    if (!localEcho || !localEcho._active) return;
    localEcho._termSize = { cols: term.cols, rows: term.rows };
    localEcho.setInput(localEcho._input, false);
  }

  function writeWrapped(text) {
    const cols = Math.max(1, term.cols || 1);
    if (localEcho && localEcho._active) localEcho.clearInput();
    term.writeln(wrapAnsi(String(text ?? ''), cols));
    restorePrompt();
  }

  function repaint() {
    if (!term) return;
    const cols = Math.max(1, term.cols || 1);
    paintedCols = cols;
    const input = localEcho ? localEcho._input : '';
    const cursor = localEcho ? localEcho._cursor : 0;
    const active = !!(localEcho && localEcho._active);
    term.write('\x1b[3J\x1b[2J\x1b[H');
    for (const raw of rawLines) {
      term.writeln(wrapAnsi(raw, cols));
    }
    if (active) {
      localEcho._termSize = { cols: term.cols, rows: term.rows };
      localEcho._input = input;
      localEcho._cursor = Math.min(cursor, input.length);
      localEcho.setInput(input, false);
    }
  }

  function pushOutput(text) {
    const raw = text == null ? '' : String(text);
    rawLines.push(raw);
    if (rawLines.length > 500) rawLines.splice(0, rawLines.length - 500);
    const cols = Math.max(1, term.cols || 1);
    if (cols !== paintedCols) {
      repaint();
      return;
    }
    writeWrapped(raw);
  }

  function readLine() {
    localEcho
      .read('~$ ')
      .then((input) => {
        onInput(input);
        readLine();
      })
      .catch((error) => console.log(`Error reading: ${error}`));
  }

  function autocompleteCommonCommands(index, tokens) {
    if (index === 0) return ['north', 'east', 'south', 'west', 'say'];
    return [];
  }

  onMount(() => {
    loadFontScale();

    const fontSize = getScaledFontSize();
    term = new xterm.Terminal({
      fontSize: fontSize,
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Consolas', monospace"
    });

    fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.setOption('cursorBlink', true);
    term.setOption('convertEol', true);

    term.open(terminalContainer);
    fitTerm();
    paintedCols = term.cols || 0;

    localEcho = new LocalEchoController(term);
    localEcho.addAutocompleteHandler(autocompleteCommonCommands);
    term.onResize(() => {
      if (!rawLines.length) {
        paintedCols = term.cols || paintedCols;
        return;
      }
      if ((term.cols || 0) === paintedCols) return;
      repaint();
    });

    // Create renderer function
    const renderer = (data) => {
      if (typeof data === 'object' && data && data.username) {
        const { r, g, b } = hexToRgb(data.color);
        const coloredName = `\x1b[1;38;2;${r};${g};${b}m${data.username}\x1b[0m`;
        pushOutput(coloredName + ':  ' + data.message);
        return;
      }
      pushOutput(data);
    };

    // Notify parent that terminal is ready
    onTerminalReady(term, renderer);
    dispatch('ready', { term, renderer });

    readLine();

    // Handle resize with debounce
    const handleResize = debounce(() => {
      applyFontSize();
    }, 100);

    window.addEventListener('resize', handleResize);

    // ResizeObserver for container resize (widget resize)
    const onFont = () => {
      loadFontScale();
      applyFontSize();
    };
    window.addEventListener('talesmud-term-font', onFont);

    resizeObserver = new ResizeObserver(debounce(() => {
      fitTerm();
    }, 100));
    resizeObserver.observe(terminalContainer);

    return () => {
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('talesmud-term-font', onFont);
    };
  });

  onDestroy(() => {
    if (resizeObserver) {
      resizeObserver.disconnect();
    }
    if (term) {
      term.dispose();
    }
  });

  // Expose methods for external use
  export function writeln(text) {
    if (term) pushOutput(text);
  }

  export function getTerminal() {
    return term;
  }

  export function fit() {
    fitTerm();
  }
</script>

<style>
  .terminal-widget {
    display: flex;
    flex-direction: column;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    overflow: hidden;
    height: 100%;
  }

  .terminal-toolbar {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 0.35rem 0.75rem;
    flex-shrink: 0;
    background: rgba(0, 0, 0, 0.3);
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .term-fontsize-btn {
    font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
    font-size: 0.6rem;
    letter-spacing: 0.08em;
    color: rgba(255, 255, 255, 0.3);
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 0.12rem 0.4rem;
    cursor: pointer;
    transition: color 0.2s, border-color 0.2s;
    line-height: 1;
    border-radius: 3px;
  }

  .term-fontsize-btn:hover {
    color: rgba(255, 255, 255, 0.7);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .terminal-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    padding: 0.75em;
  }

  .terminal-inner {
    flex: 1;
    background: transparent;
    min-height: 0;
  }

  .terminal-inner :global(.xterm) {
    height: 100% !important;
  }

  .terminal-inner :global(.xterm-viewport) {
    background: transparent !important;
    scrollbar-width: thin;
    scrollbar-color: rgba(180, 130, 60, 0.45) transparent;
  }

  .terminal-inner :global(.xterm-viewport::-webkit-scrollbar) {
    width: 8px;
  }

  .terminal-inner :global(.xterm-viewport::-webkit-scrollbar-thumb) {
    background: rgba(180, 130, 60, 0.45);
    border-radius: 4px;
  }

  .terminal-inner :global(.xterm-screen) {
    background: transparent !important;
  }
</style>

<div class="terminal-widget">
  <div class="terminal-toolbar">
    <button class="term-fontsize-btn" on:click={cycleFontSize} title="Font size: {FONT_SCALES[fontScaleIndex].key}">
      {currentFontLabel}
    </button>
  </div>
  <div class="terminal-container">
    <div class="terminal-inner" bind:this={terminalContainer}></div>
  </div>
</div>
