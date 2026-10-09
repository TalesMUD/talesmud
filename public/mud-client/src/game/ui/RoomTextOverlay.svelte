<script>
  import { onDestroy } from 'svelte';
  import { overlayStore } from './overlayStore.js';

  // Always show action/system reaction toasts on the room hero art (desktop + mobile).
  // This is the room reaction overlay — not the command log.

  let timers = new Map();

  $: scheduleTimers($overlayStore);

  function scheduleTimers(messages) {
    for (const msg of messages) {
      // Stack accel (or an early startFade) may mark fading before our display timer fires.
      if (msg.fading) {
        const existing = timers.get(msg.id);
        if (existing && existing.removeTimer) continue;
        if (existing && existing.displayTimer) clearTimeout(existing.displayTimer);
        const removeTimer = setTimeout(() => {
          overlayStore.removeMessage(msg.id);
          timers.delete(msg.id);
        }, msg.fadeOutDuration || 350);
        timers.set(msg.id, { displayTimer: null, removeTimer });
        continue;
      }

      if (timers.has(msg.id)) continue;

      const displayTimer = setTimeout(() => {
        overlayStore.startFade(msg.id);

        const removeTimer = setTimeout(() => {
          overlayStore.removeMessage(msg.id);
          timers.delete(msg.id);
        }, msg.fadeOutDuration);

        const entry = timers.get(msg.id);
        if (entry) entry.removeTimer = removeTimer;
      }, msg.displayDuration);

      timers.set(msg.id, { displayTimer, removeTimer: null });
    }

    // Drop timers for messages that were replaced/removed (e.g. new examine card).
    for (const id of [...timers.keys()]) {
      if (!messages.some(m => m.id === id)) {
        const t = timers.get(id);
        if (t) {
          if (t.displayTimer) clearTimeout(t.displayTimer);
          if (t.removeTimer) clearTimeout(t.removeTimer);
        }
        timers.delete(id);
      }
    }
  }

  function formatText(text) {
    const escaped = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return escaped.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  }

  function dismiss(id) {
    const t = timers.get(id);
    if (t) {
      clearTimeout(t.displayTimer);
      if (t.removeTimer) clearTimeout(t.removeTimer);
      timers.delete(id);
    }
    overlayStore.removeMessage(id);
  }

  function isUniqueExamine(examine) {
    if (!examine || !examine.details) return false;
    return examine.details.some((row) => row.label === 'Mark' && String(row.value).toLowerCase() === 'unique');
  }

  function qualityClass(value) {
    const v = String(value || '').toLowerCase();
    if (v === 'magic') return 'q-magic';
    if (v === 'rare') return 'q-rare';
    if (v === 'legendary') return 'q-legendary';
    if (v === 'mythic') return 'q-mythic';
    return 'q-normal';
  }

  onDestroy(() => {
    for (const [, t] of timers) {
      clearTimeout(t.displayTimer);
      if (t.removeTimer) clearTimeout(t.removeTimer);
    }
    timers.clear();
  });
</script>

<style>
  /* Centered on the room hero art — leave the bottom lane for Pip / hotbar. */
  .room-text-overlay {
    position: absolute;
    inset: 0;
    z-index: 55;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.55em;
    padding: 1em 1.25em 5em;
    pointer-events: none;
    overflow: hidden;
    box-sizing: border-box;
  }

  /*
   * Midpoint type (e4d6aa1) kept. Cards grow wide/tall enough for a normal LOOK
   * paragraph; overflow-y only engages for unusually long blobs.
   */
  .overlay-message {
    box-sizing: border-box;
    background: rgba(8, 10, 14, 0.88);
    color: #f3f4f6;
    border-radius: 10px;
    font-size: clamp(0.88rem, 1.68vw, 1.10rem);
    font-weight: 500;
    line-height: 1.45;
    letter-spacing: 0;
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1.5px solid rgba(249, 115, 22, 0.65);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
    opacity: 1;
    transition: opacity var(--fade-duration) ease-out;
    animation: overlayPopIn 0.22s ease-out;
    text-align: center;
    max-width: min(90%, 36rem);
    width: max-content;
    max-height: min(56%, 22rem);
    overflow-x: hidden;
    overflow-y: auto;
    padding: 0;
  }

  /* Soft mood / ambiance toasts — not orange alert chrome. */
  .overlay-message.ambiance {
    background: rgba(12, 14, 18, 0.72);
    color: #f5e6c8;
    font-style: italic;
    font-weight: 400;
    border: 1px solid rgba(212, 175, 110, 0.42);
    box-shadow:
      0 6px 20px rgba(0, 0, 0, 0.35),
      0 0 18px rgba(212, 175, 110, 0.08);
    animation: overlayMoodIn 0.45s ease-out;
  }

  .overlay-message.ambiance :global(strong) {
    color: #f0d9a0;
    font-weight: 600;
    font-style: italic;
  }

  /* Examine item card — Veilspan dark/gold, matches ItemDetailCard / Settings. */
  .overlay-message.examine {
    pointer-events: auto;
    text-align: left;
    width: min(92%, 28rem);
    max-width: min(92%, 28rem);
    max-height: min(70%, 28rem);
    background: rgba(12, 14, 18, 0.94);
    color: #e7e2d8;
    border: 1px solid rgba(212, 175, 55, 0.55);
    border-radius: 12px;
    box-shadow:
      0 18px 48px rgba(0, 0, 0, 0.55),
      0 0 24px rgba(212, 175, 55, 0.08);
    font-size: 0.92rem;
    font-weight: 400;
    animation: overlayExamineIn 0.28s ease-out;
  }

  .overlay-message.examine.examine-unique {
    border: 2px solid #facc15;
    box-shadow:
      0 0 0 1px #7a5a16,
      0 18px 48px rgba(0, 0, 0, 0.55),
      0 0 22px rgba(250, 204, 21, 0.35);
  }

  .examine-card {
    box-sizing: border-box;
    padding: 1rem 1.15rem 1.1rem;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    min-height: 0;
  }

  .examine-header {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .examine-title {
    flex: 1;
    min-width: 0;
    margin: 0;
    font-family: Cinzel, "Times New Roman", serif;
    font-size: clamp(1.05rem, 2.2vw, 1.28rem);
    font-weight: 600;
    letter-spacing: 0.04em;
    color: #f3e4bd;
    line-height: 1.25;
  }

  .examine-close {
    flex: none;
    margin: -0.15rem -0.25rem 0 0;
    padding: 0.15rem 0.45rem;
    border: 1px solid rgba(212, 175, 55, 0.35);
    border-radius: 6px;
    background: rgba(30, 26, 20, 0.7);
    color: #d4c4a0;
    font-size: 1.15rem;
    line-height: 1;
    cursor: pointer;
    pointer-events: auto;
  }

  .examine-close:hover {
    border-color: rgba(212, 175, 55, 0.7);
    color: #ffe3a4;
  }

  .examine-blurb {
    margin: 0;
    font-size: 0.9rem;
    line-height: 1.45;
    color: #d1c8b8;
  }

  .examine-lore {
    margin: 0;
    max-height: 11rem;
    overflow-y: auto;
    padding-right: 0.25rem;
    font-size: 0.86rem;
    line-height: 1.5;
    color: #b9b0a0;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .examine-lore::-webkit-scrollbar {
    width: 6px;
  }
  .examine-lore::-webkit-scrollbar-thumb {
    background: rgba(212, 175, 55, 0.35);
    border-radius: 4px;
  }

  .examine-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    margin-top: 0.15rem;
  }

  .examine-chip {
    display: inline-flex;
    align-items: baseline;
    gap: 0.35rem;
    padding: 0.28rem 0.55rem;
    border-radius: 999px;
    border: 1px solid rgba(74, 61, 37, 0.9);
    background: rgba(28, 24, 18, 0.9);
    font-size: 0.72rem;
    letter-spacing: 0.02em;
    color: #cfc3a8;
  }

  .examine-chip-label {
    text-transform: uppercase;
    font-size: 0.62rem;
    letter-spacing: 0.08em;
    color: #9a8f78;
  }

  .examine-chip-value {
    color: #f0e5d1;
    font-weight: 600;
  }

  .examine-chip.q-magic .examine-chip-value { color: #4ade80; }
  .examine-chip.q-rare .examine-chip-value { color: #60a5fa; }
  .examine-chip.q-legendary .examine-chip-value { color: #c084fc; }
  .examine-chip.q-mythic .examine-chip-value { color: #fbbf24; }

  .examine-rows {
    display: grid;
    gap: 0.28rem;
    border-top: 1px solid rgba(74, 61, 37, 0.55);
    padding-top: 0.55rem;
  }

  .examine-row {
    display: flex;
    justify-content: space-between;
    gap: 0.75rem;
    font-size: 0.78rem;
  }

  .examine-row-label {
    color: #9a8f78;
  }

  .examine-row-value {
    color: #e7e2d8;
    font-weight: 600;
    text-align: right;
  }

  .overlay-message-inner {
    box-sizing: border-box;
    padding: 0.95em 1.25em 1.15em;
    white-space: pre-wrap;
    word-break: break-word;
    overflow-wrap: anywhere;
  }

  .overlay-message :global(strong) {
    color: #fde68a;
    font-weight: 700;
  }

  .overlay-message.fading {
    opacity: 0;
  }

  @keyframes overlayPopIn {
    from {
      opacity: 0;
      transform: scale(0.96) translateY(6px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  @keyframes overlayMoodIn {
    from {
      opacity: 0;
      transform: scale(0.98) translateY(4px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  @keyframes overlayExamineIn {
    from {
      opacity: 0;
      transform: scale(0.97) translateY(8px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  @media screen and (max-width: 768px) {
    .room-text-overlay {
      padding: 0.4em 0.55em 5.5em;
      gap: 0.3em;
      justify-content: flex-start;
    }
    .overlay-message {
      font-size: clamp(0.72rem, 2.6vw, 0.88rem);
      line-height: 1.3;
      max-width: 96%;
      max-height: min(28%, 7.5rem);
      border-radius: 8px;
    }
    .overlay-message.examine {
      max-width: 96%;
      width: 96%;
      max-height: min(52%, 18rem);
      font-size: 0.84rem;
    }
    .examine-card {
      padding: 0.75rem 0.85rem 0.85rem;
      gap: 0.5rem;
    }
    .examine-lore {
      max-height: 7.5rem;
    }
    .overlay-message-inner {
      padding: 0.45em 0.7em 0.55em;
    }
  }
</style>

{#if $overlayStore.length > 0}
  <div class="room-text-overlay" aria-live="polite">
    {#each $overlayStore as msg (msg.id)}
      <div
        class="overlay-message"
        class:ambiance={msg.kind === 'ambiance'}
        class:examine={msg.kind === 'examine' && msg.examine}
        class:examine-unique={msg.kind === 'examine' && isUniqueExamine(msg.examine)}
        class:fading={msg.fading}
        style="--fade-duration: {msg.fadeOutDuration}ms"
        role={msg.kind === 'examine' ? 'dialog' : undefined}
        aria-label={msg.examine ? msg.examine.title : undefined}
      >
        {#if msg.kind === 'examine' && msg.examine}
          <div class="examine-card">
            <div class="examine-header">
              <h2 class="examine-title">{msg.examine.title}</h2>
              <button
                type="button"
                class="examine-close"
                aria-label="Close examine"
                on:click={() => dismiss(msg.id)}
              >×</button>
            </div>

            {#if msg.examine.blurb}
              <p class="examine-blurb">{msg.examine.blurb}</p>
            {/if}

            {#if msg.examine.lore}
              <div class="examine-lore">{msg.examine.lore}</div>
            {/if}

            {#if msg.examine.details.length}
              <div class="examine-chips">
                {#each msg.examine.details as row}
                  <span
                    class="examine-chip"
                    class:q-magic={row.label === 'Quality' && qualityClass(row.value) === 'q-magic'}
                    class:q-rare={row.label === 'Quality' && qualityClass(row.value) === 'q-rare'}
                    class:q-legendary={row.label === 'Quality' && qualityClass(row.value) === 'q-legendary'}
                    class:q-mythic={row.label === 'Quality' && qualityClass(row.value) === 'q-mythic'}
                  >
                    <span class="examine-chip-label">{row.label}</span>
                    <span class="examine-chip-value">{row.value}</span>
                  </span>
                {/each}
              </div>
            {/if}

            {#if msg.examine.attributes.length || msg.examine.properties.length}
              <div class="examine-rows">
                {#each [...msg.examine.attributes, ...msg.examine.properties] as row}
                  <div class="examine-row">
                    <span class="examine-row-label">{row.label}</span>
                    <span class="examine-row-value">{row.value}</span>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {:else}
          <div class="overlay-message-inner">
            {@html formatText(msg.text)}
          </div>
        {/if}
      </div>
    {/each}
  </div>
{/if}
