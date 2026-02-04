<script>
  import { onDestroy } from 'svelte';
  import { overlayStore } from './overlayStore.js';
  import { settingsStore } from '../SettingsStore.js';
  import { mobileStore } from '../mobile/mobileStore.js';

  const { isMobile } = mobileStore;

  // Mobile: always show. Desktop: respect setting.
  $: overlayEnabled = $isMobile || $settingsStore.interface?.roomTextOverlay;

  // Track timers per message so we can clean up
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
    // Escape HTML first, then convert **bold** markers to <strong> tags
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
  .room-text-overlay {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 50;
    display: flex;
    flex-direction: column;
    gap: 0.25em;
    padding: 0.5em 0.75em;
    pointer-events: none;
    max-height: 60%;
    overflow: hidden;
  }

  .overlay-message {
    background: rgba(0, 0, 0, 0.75);
    color: #e5e7eb;
    padding: 0.5em 0.75em;
    border-radius: 6px;
    font-size: 0.9em;
    line-height: 1.4;
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    border-left: 2px solid rgba(245, 158, 11, 0.5);
    opacity: 1;
    transition: opacity var(--fade-duration) ease-out;
    animation: overlaySlideIn 0.2s ease-out;
    white-space: pre-wrap;
  }

  .overlay-message :global(strong) {
    color: #f0e6d3;
    font-weight: 700;
  }

  .overlay-message.fading {
    opacity: 0;
  }

  @keyframes overlaySlideIn {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @media screen and (max-width: 768px) {
    .overlay-message {
      font-size: 0.85em;
      padding: 0.4em 0.6em;
    }
  }
</style>

{#if overlayEnabled && $overlayStore.length > 0}
  <div class="room-text-overlay">
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
