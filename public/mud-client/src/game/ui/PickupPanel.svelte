<script>
  import { createEventDispatcher, onDestroy, tick } from 'svelte';
  import { itemArtSrc, onItemArtError } from '../itemArtSrc.js';
  import { currentTopPanel, isTextEntry, prefersReducedMotion } from '../keyboardShortcuts.js';
  import {
    groupGroundItems,
    pickupAllCommands,
    pickupEntryCount,
    pickupKeyAction,
    pickupRowCommands,
    shouldClosePickup,
    showPickupAll,
  } from '../pickupPanel.js';
  import { portal } from '../portal.js';
  import { settingsStore } from '../SettingsStore.js';

  export let open = false;
  export let items = [];
  export let sendMessage = () => {};
  export let removeItem = () => {};

  const dispatch = createEventDispatcher();

  let dialogEl;
  let wasOpen = false;
  let sentIds = [];
  let closedForEmpty = false;
  let listening = false;

  $: rows = groupGroundItems(items);
  $: count = pickupEntryCount(items);
  $: pickAllVisible = showPickupAll(items);
  $: reduceMotion = prefersReducedMotion($settingsStore?.interface?.reducedMotion);

  $: pruneSent(items);

  function liveIds(list) {
    const ids = new Set();
    groupGroundItems(list).forEach((row) => {
      row.ids.forEach((id) => ids.add(id));
    });
    return ids;
  }

  function pruneSent(list) {
    const live = liveIds(list);
    const next = sentIds.filter((id) => live.has(id));
    if (next.length !== sentIds.length) sentIds = next;
  }

  function close() {
    dispatch('close');
  }

  function take(commands) {
    if (!commands.length) return;
    const next = new Set(sentIds);
    for (const cmd of commands) {
      if (!cmd || !cmd.id || next.has(cmd.id)) continue;
      next.add(cmd.id);
      sendMessage(cmd.command);
      removeItem(cmd.id);
    }
    sentIds = [...next];
  }

  function pickRow(row) {
    take(pickupRowCommands(items, row, sentIds));
  }

  function pickAll() {
    take(pickupAllCommands(items, sentIds));
  }

  function onKey(event) {
    if (!open) return;
    const top = currentTopPanel();
    if (top && top !== 'pickup') return;
    const action = pickupKeyAction(event, rows.length);
    if (!action) return;
    event.preventDefault();
    event.stopPropagation();
    if (action.type === 'close') close();
    else if (action.type === 'all') pickAll();
    else if (action.type === 'row' && rows[action.index]) pickRow(rows[action.index]);
  }

  function bindKeys() {
    if (listening || typeof window === 'undefined') return;
    window.addEventListener('keydown', onKey, true);
    listening = true;
  }

  function unbindKeys() {
    if (!listening || typeof window === 'undefined') return;
    window.removeEventListener('keydown', onKey, true);
    listening = false;
  }

  async function focusDialog() {
    await tick();
    if (!open || !dialogEl || typeof document === 'undefined') return;
    if (isTextEntry(document.activeElement)) return;
    dialogEl.focus();
  }

  $: if (open) bindKeys();
  else unbindKeys();

  $: if (open && !wasOpen) {
    wasOpen = true;
    focusDialog();
  } else if (!open) {
    wasOpen = false;
  }

  $: if (open && shouldClosePickup(items)) {
    if (!closedForEmpty) {
      closedForEmpty = true;
      close();
    }
  } else {
    closedForEmpty = false;
  }

  onDestroy(unbindKeys);
</script>

{#if open && count > 0}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div
    class="pickup-scrim"
    class:reduced={reduceMotion}
    use:portal
    on:click={close}
  >
    <div
      class="pickup-panel"
      class:reduced={reduceMotion}
      role="dialog"
      aria-modal="true"
      aria-labelledby="pickup-panel-title"
      tabindex="-1"
      bind:this={dialogEl}
      on:click|stopPropagation
    >
      <header class="pickup-header">
        <i class="material-icons pickup-hand" aria-hidden="true">back_hand</i>
        <h2 id="pickup-panel-title">Pick up</h2>
        <span class="pickup-total">{count}</span>
        <button class="pickup-close" type="button" aria-label="Close" on:click={close}>
          <i class="material-icons" aria-hidden="true">close</i>
        </button>
      </header>

      <div class="pickup-list">
        {#each rows as row (row.key)}
          <button
            class="pickup-row"
            class:is-currency={row.currency}
            class:is-unique={row.unique}
            type="button"
            aria-keyshortcuts={row.badge || undefined}
            on:click={() => pickRow(row)}
          >
            {#if row.badge}
              <span class="pickup-badge" aria-hidden="true">{row.badge}</span>
            {:else}
              <span class="pickup-badge pickup-badge-empty" aria-hidden="true"></span>
            {/if}
            <img
              class="pickup-art"
              src={itemArtSrc(row.artItem)}
              alt=""
              on:error={(event) => onItemArtError(event, row.artItem)}
            />
            <span class="pickup-copy">
              <span class="pickup-name rarity-{row.rarity}">
                {row.name}
                {#if row.unique}<span class="pickup-unique-mark">Unique</span>{/if}
              </span>
              {#if row.hint}<span class="pickup-hint">{row.hint}</span>{/if}
            </span>
            {#if row.currency}
              <span class="pickup-amount">{row.count}</span>
            {:else if row.count > 1}
              <span class="pickup-stack">×{row.count}</span>
            {/if}
          </button>
        {/each}
      </div>

      {#if pickAllVisible}
        <footer class="pickup-footer">
          <button class="pickup-all" type="button" on:click={pickAll}>
            Pick up all ({count})
          </button>
        </footer>
      {/if}
    </div>
  </div>
{/if}

<style>
  .pickup-scrim {
    position: fixed;
    inset: 0;
    z-index: 10060;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgba(0, 0, 0, 0.72);
    animation: pickup-fade 0.15s ease-out;
  }

  .pickup-panel {
    display: flex;
    flex-direction: column;
    width: min(440px, 100%);
    max-height: min(72vh, 560px);
    background: #14171f;
    color: #e5e7eb;
    border: 1px solid rgba(201, 162, 90, 0.65);
    border-radius: 14px;
    box-shadow: 0 22px 60px rgba(0, 0, 0, 0.75), 0 0 0 1px rgba(0, 0, 0, 0.6);
    outline: none;
    overflow: hidden;
    animation: pickup-in 0.18s ease-out;
  }

  .pickup-header {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 0 0 auto;
    padding: 12px 12px 10px 14px;
    border-bottom: 1px solid rgba(201, 162, 90, 0.28);
    background: #14171f;
  }

  .pickup-hand {
    color: #e8c25a;
    font-size: 22px;
  }

  h2 {
    margin: 0;
    font-family: var(--font-display, Georgia, serif);
    font-size: 1.15rem;
    font-weight: 650;
    letter-spacing: 0.01em;
    text-transform: none;
    color: #f3e6c8;
  }

  .pickup-total {
    min-width: 22px;
    height: 22px;
    padding: 0 6px;
    border-radius: 999px;
    background: #2a2418;
    border: 1px solid rgba(232, 194, 90, 0.45);
    color: #f6d77e;
    font-size: 12px;
    font-weight: 700;
    line-height: 20px;
    text-align: center;
  }

  .pickup-close {
    margin-left: auto;
    width: 36px;
    height: 36px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.16);
    background: #1b2030;
    color: #d5dbe5;
    cursor: pointer;
  }

  .pickup-close:hover,
  .pickup-close:focus-visible {
    background: #222839;
    border-color: #e8c25a;
    outline: 2px solid #e8c25a;
    outline-offset: 1px;
  }

  .pickup-list {
    overflow-y: auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
    background: #14171f;
  }

  .pickup-row {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    min-height: 52px;
    margin: 0;
    padding: 8px 10px;
    text-align: left;
    text-transform: none;
    letter-spacing: 0;
    cursor: pointer;
    color: #e5e7eb;
    background: #1b2028;
    border: 1px solid #30363d;
    border-radius: 10px;
  }

  .pickup-row:hover,
  .pickup-row:focus-visible {
    background: #241e16;
    border-color: #e8c25a;
    outline: 2px solid #e8c25a;
    outline-offset: 1px;
  }

  .pickup-row.is-currency {
    background: #1c1810;
    border-color: #8a6d32;
  }

  .pickup-row.is-unique {
    background: linear-gradient(90deg, rgba(232, 194, 90, 0.16), #1b2028 42%);
    border-color: rgba(232, 194, 90, 0.7);
  }

  .pickup-badge {
    flex: 0 0 22px;
    width: 22px;
    height: 22px;
    border-radius: 6px;
    background: #2a241c;
    color: #e8c25a;
    font-size: 12px;
    font-weight: 800;
    line-height: 22px;
    text-align: center;
  }

  .pickup-badge-empty {
    background: transparent;
  }

  .pickup-art {
    flex: 0 0 40px;
    width: 40px;
    height: 40px;
    object-fit: contain;
    image-rendering: pixelated;
    background: #080b0f;
    border: 1px solid #3d3423;
    border-radius: 8px;
  }

  .pickup-copy {
    min-width: 0;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .pickup-name {
    font-size: 15px;
    font-weight: 650;
    line-height: 1.25;
    text-transform: none;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pickup-unique-mark {
    display: inline-block;
    margin-left: 6px;
    padding: 0 6px;
    border: 1px solid rgba(250, 204, 21, 0.75);
    border-radius: 999px;
    color: #fde68a;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    vertical-align: 1px;
  }

  .pickup-hint {
    color: #a9a397;
    font-size: 12px;
    line-height: 1.3;
    text-transform: none;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pickup-stack,
  .pickup-amount {
    flex: 0 0 auto;
    font-size: 14px;
    font-weight: 750;
    font-variant-numeric: tabular-nums;
  }

  .pickup-stack {
    color: #d5dbe5;
  }

  .pickup-amount {
    color: #f6d77e;
  }

  .pickup-footer {
    flex: 0 0 auto;
    padding: 10px 12px 12px;
    border-top: 1px solid rgba(201, 162, 90, 0.28);
    background: #14171f;
  }

  .pickup-all {
    width: 100%;
    min-height: 48px;
    margin: 0;
    padding: 10px 14px;
    border-radius: 10px;
    border: 1px solid #b38c3f;
    background: #3d3018;
    color: #ffe3a4;
    font-size: 15px;
    font-weight: 750;
    letter-spacing: 0.01em;
    text-transform: none;
    cursor: pointer;
  }

  .pickup-all:hover,
  .pickup-all:focus-visible {
    background: #4c3c1e;
    border-color: #e8c25a;
    outline: 2px solid #e8c25a;
    outline-offset: 1px;
  }

  .rarity-normal { color: #e5e7eb; }
  .rarity-magic { color: #60a5fa; }
  .rarity-rare { color: #facc15; }
  .rarity-legendary { color: #fb923c; }
  .rarity-mythic { color: #e879f9; }
  .rarity-unique { color: #f6d77e; }

  @media (max-width: 600px) {
    .pickup-scrim {
      align-items: flex-end;
      padding: 0;
    }

    .pickup-panel {
      width: 100%;
      max-width: none;
      max-height: min(78vh, 640px);
      border-radius: 16px 16px 0 0;
      border-bottom: none;
      padding-bottom: env(safe-area-inset-bottom);
    }

    .pickup-row,
    .pickup-all,
    .pickup-close {
      min-height: 48px;
    }

    .pickup-list {
      max-height: none;
    }
  }

  @keyframes pickup-fade {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes pickup-in {
    from { opacity: 0; transform: translateY(8px); }
    to { opacity: 1; transform: none; }
  }

  .reduced,
  .reduced .pickup-panel {
    animation: none;
  }

  @media (prefers-reduced-motion: reduce) {
    .pickup-scrim,
    .pickup-panel {
      animation: none;
    }
  }
</style>
