<script>
  import { onDestroy } from 'svelte';
  import { overlayStore } from './overlayStore.js';

  // Always show action/system reaction toasts on the room hero art (desktop + mobile).
  // This is the room reaction overlay — not the command log.

  let timers = new Map();

  $: scheduleTimers($overlayStore);

  function scheduleTimers(messages) {
    for (const msg of messages) {
      if (timers.has(msg.id) || msg.fading) continue;

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
  }

  function formatText(text) {
    const escaped = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return escaped.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
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
  /* Centered on the room hero art — leave bottom alley for Pip / hotbar. */
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
    /* Outer padding is 0 — inner pad lives on the scrollport content so the
       last line + bottom pad stay fully visible (no half-cut glyphs). */
    padding: 0;
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
        class:fading={msg.fading}
        style="--fade-duration: {msg.fadeOutDuration}ms"
      >
        <div class="overlay-message-inner">
          {@html formatText(msg.text)}
        </div>
      </div>
    {/each}
  </div>
{/if}
