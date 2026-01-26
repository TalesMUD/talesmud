<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { writable } from 'svelte/store';
  import xterm from 'xterm';
  import { FitAddon } from 'xterm-addon-fit';
  import LocalEchoController from '../echo/LocalEchoController';
  import '../../../node_modules/xterm/css/xterm.css';

  export let onTerminalReady = () => {};
  export let onInput = () => {};

  const dispatch = createEventDispatcher();

  let terminalContainer;
  let term;
  let fitAddon;
  let localEcho;
  let resizeObserver;

  function getTerminalFontSize() {
    const width = window.innerWidth;
    if (width >= 3000) return 26;
    if (width >= 2600) return 24;
    if (width >= 2200) return 22;
    if (width >= 1800) return 20;
    if (width >= 1600) return 18;
    if (width >= 1400) return 17;
    return 15;
  }

  function debounce(fn, ms) {
    let timeout;
    return (...args) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => fn(...args), ms);
    };
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
    const fontSize = getTerminalFontSize();
    term = new xterm.Terminal({
      fontSize: fontSize,
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Consolas', monospace"
    });

    fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.setOption('cursorBlink', true);
    term.setOption('convertEol', true);

    term.open(terminalContainer);
    fitAddon.fit();

    localEcho = new LocalEchoController(term);
    localEcho.addAutocompleteHandler(autocompleteCommonCommands);

    // Create renderer function
    const renderer = (data) => {
      localEcho.clearInput();
      term.writeln(data);
    };

    // Notify parent that terminal is ready
    onTerminalReady(term, renderer);
    dispatch('ready', { term, renderer });

    readLine();

    // Handle resize with debounce
    const handleResize = debounce(() => {
      const newFontSize = getTerminalFontSize();
      term.setOption('fontSize', newFontSize);
      fitAddon.fit();
    }, 100);

    window.addEventListener('resize', handleResize);

    // ResizeObserver for container resize (widget resize)
    resizeObserver = new ResizeObserver(debounce(() => {
      if (fitAddon) {
        fitAddon.fit();
      }
    }, 100));
    resizeObserver.observe(terminalContainer);

    return () => {
      window.removeEventListener('resize', handleResize);
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
    if (localEcho) {
      localEcho.clearInput();
    }
    if (term) {
      term.writeln(text);
    }
  }

  export function getTerminal() {
    return term;
  }

  export function fit() {
    if (fitAddon) {
      fitAddon.fit();
    }
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
  }

  .terminal-inner :global(.xterm-screen) {
    background: transparent !important;
  }
</style>

<div class="terminal-widget">
  <div class="terminal-container">
    <div class="terminal-inner" bind:this={terminalContainer}></div>
  </div>
</div>
