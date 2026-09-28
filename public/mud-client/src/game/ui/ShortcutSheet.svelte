<script>
  import { cheatSheetOpen } from '../uiChrome.js';

  const rows = [
    ['1 – 9', 'Fire hotbar slots'],
    ['Tab', 'Cycle combat targets'],
    ['Escape', 'Close the top panel'],
    ['?', 'Open or close this list'],
  ];

  function close() {
    cheatSheetOpen.set(false);
  }
</script>

{#if $cheatSheetOpen}
  <div class="sheet-backdrop" on:click={close} role="presentation">
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="sheet" role="dialog" aria-label="Keyboard shortcuts" on:click|stopPropagation>
      <div class="sheet-title">Shortcuts</div>
      <ul>
        {#each rows as row}
          <li><kbd>{row[0]}</kbd><span>{row[1]}</span></li>
        {/each}
      </ul>
      <p class="hint">These keys stay quiet while you are typing a command.</p>
      <button type="button" on:click={close}>Close</button>
    </div>
  </div>
{/if}

<style>
  .sheet-backdrop {
    position: fixed;
    inset: 0;
    z-index: 9500;
    display: grid;
    place-items: center;
    background: rgba(0, 0, 0, 0.45);
    pointer-events: auto;
  }
  .sheet {
    width: min(28rem, calc(100vw - 2rem));
    padding: 1.1rem 1.2rem 1rem;
    border: 1.5px solid rgba(212, 164, 74, 0.75);
    border-radius: 12px;
    background: linear-gradient(180deg, rgba(28, 20, 10, 0.97), rgba(8, 6, 4, 0.97));
    color: #f5e6c0;
    font-family: 'Cinzel', Georgia, serif;
    box-shadow: 0 18px 48px rgba(0, 0, 0, 0.55);
  }
  .sheet-title {
    letter-spacing: 0.14em;
    text-transform: uppercase;
    margin-bottom: 0.7rem;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.35rem 0;
    border-top: 1px solid rgba(212, 164, 74, 0.25);
    font-family: system-ui, sans-serif;
    font-size: 0.92rem;
  }
  kbd {
    color: #fbbf24;
    font-family: inherit;
  }
  .hint {
    margin: 0.7rem 0;
    font-family: system-ui, sans-serif;
    font-size: 0.8rem;
    color: #d1d5db;
  }
  button {
    width: 100%;
    padding: 0.55rem 0.8rem;
    border-radius: 8px;
    border: 1px solid #e8c878;
    background: rgba(48, 34, 14, 0.98);
    color: #f5e6c0;
    font-family: inherit;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    cursor: pointer;
  }
</style>
