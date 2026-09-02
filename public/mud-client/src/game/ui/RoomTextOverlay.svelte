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
  /* Centered on the room hero art — not parked on the description text. */
  .room-text-overlay {
    position: absolute;
    inset: 0;
    z-index: 55;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.35em;
    padding: 0.7em 1em;
    pointer-events: none;
    overflow: hidden;
  }

  .overlay-message {
    background: rgba(8, 10, 14, 0.88);
    color: #f3f4f6;
    padding: 0.4em 0.75em;
    border-radius: 8px;
    font-size: clamp(0.72rem, 1.15vw, 0.88rem);
    font-weight: 500;
    line-height: 1.25;
    letter-spacing: 0;
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1px solid rgba(249, 115, 22, 0.65);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
    opacity: 1;
    transition: opacity var(--fade-duration) ease-out;
    animation: overlayPopIn 0.22s ease-out;
    white-space: pre-wrap;
    word-break: break-word;
    overflow-wrap: anywhere;
    text-align: center;
    max-width: min(78%, 26rem);
    width: max-content;
    max-height: min(38%, 10rem);
    overflow-y: auto;
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
      padding: 0.55em 0.65em;
    }
    .overlay-message {
      font-size: clamp(0.7rem, 2.6vw, 0.84rem);
      line-height: 1.22;
      padding: 0.35em 0.65em;
      max-width: 86%;
      max-height: min(36%, 9rem);
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
        {@html formatText(msg.text)}
      </div>
    {/each}
  </div>
{/if}
