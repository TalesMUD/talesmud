<script>
  import { onDestroy, tick } from 'svelte';
  import { portal } from '../portal.js';

  export let store = null;
  export let sendMessage = null;

  let needBtn;
  let nowSec = Math.floor(Date.now() / 1000);
  let countdownTimer = null;
  let escHandler = null;

  $: roll = (store && $store && $store.needGreed) || null;
  $: pending = !!(roll && roll.pending);
  $: secondsLeft = pending && roll.expiresAt
    ? Math.max(0, roll.expiresAt - nowSec)
    : 0;
  $: quality = (roll && roll.quality) || 'normal';
  $: label = (roll && roll.itemLabel) || (roll && roll.itemName) || 'an item';

  $: if (pending) {
    if (!countdownTimer && typeof window !== 'undefined') {
      nowSec = Math.floor(Date.now() / 1000);
      countdownTimer = setInterval(() => {
        nowSec = Math.floor(Date.now() / 1000);
      }, 1000);
    }
    if (!escHandler) {
      escHandler = (e) => {
        if (e.key === 'Escape') {
          e.preventDefault();
          choose('pass');
        }
      };
      if (typeof window !== 'undefined') window.addEventListener('keydown', escHandler);
    }
    tick().then(() => {
      if (needBtn && typeof needBtn.focus === 'function') needBtn.focus();
    });
  } else {
    if (countdownTimer) {
      clearInterval(countdownTimer);
      countdownTimer = null;
    }
    if (escHandler) {
      if (typeof window !== 'undefined') window.removeEventListener('keydown', escHandler);
      escHandler = null;
    }
  }

  function choose(choice) {
    if (!sendMessage || !pending) return;
    sendMessage(choice);
  }

  onDestroy(() => {
    if (escHandler && typeof window !== 'undefined') window.removeEventListener('keydown', escHandler);
    if (countdownTimer) clearInterval(countdownTimer);
  });
</script>

<style>
  .ng-overlay {
    position: fixed;
    inset: 0;
    z-index: var(--layer-roll);
    background: transparent;
    pointer-events: none;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding:
      max(0.75rem, env(safe-area-inset-top, 0px))
      max(0.75rem, env(safe-area-inset-right, 0px))
      max(0.75rem, env(safe-area-inset-bottom, 0px))
      max(0.75rem, env(safe-area-inset-left, 0px));
    box-sizing: border-box;
  }
  .ng-panel {
    pointer-events: auto;
    width: min(400px, 100%);
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 1.25em 1.35em 1.15em;
    background: var(--social-panel-bg, rgba(12, 16, 24, 0.97));
    border: 1px solid var(--social-border, rgba(212, 175, 55, 0.4));
    border-radius: 12px;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
  }
  .ng-kicker {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #fbbf24;
    display: flex;
    align-items: center;
    gap: 0.35em;
  }
  .ng-kicker i { font-size: 1.15em; }
  .ng-item {
    margin: 0;
    font-size: 17px;
    font-weight: 700;
    line-height: 1.35;
  }
  .ng-item.rarity-normal { color: #e5e7eb; }
  .ng-item.rarity-magic { color: #60a5fa; }
  .ng-item.rarity-rare { color: #facc15; }
  .ng-item.rarity-legendary { color: #fb923c; }
  .ng-item.rarity-mythic { color: #e879f9; }
  .ng-item.rarity-unique { color: #f6d77e; }
  .ng-timer {
    color: #9ca3af;
    font-size: 12px;
    font-variant-numeric: tabular-nums;
  }
  .ng-actions {
    display: flex;
    gap: 8px;
    margin-top: 4px;
  }
  .ng-act {
    flex: 1;
    min-height: 44px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 12px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: rgba(255,255,255,0.05);
    color: #e5e7eb;
  }
  .ng-act.need {
    background: rgba(34, 197, 94, 0.18);
    border-color: rgba(34, 197, 94, 0.45);
    color: #86efac;
  }
  .ng-act.greed {
    background: rgba(251, 191, 36, 0.16);
    border-color: rgba(251, 191, 36, 0.45);
    color: #fbbf24;
  }
  .ng-act.pass {
    border-color: rgba(248, 113, 113, 0.4);
    color: #fca5a5;
    background: transparent;
  }
  @media (min-width: 900px) and (min-height: 700px) {
    .ng-overlay {
      align-items: center;
    }
  }
  @media (max-width: 1024px) and (max-height: 500px) and (orientation: landscape) {
    .ng-overlay { justify-content: flex-end; }
    .ng-panel { width: calc(50vw - 1.5rem); }
  }
</style>

{#if pending}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div
    class="ng-overlay"
    use:portal
    role="dialog"
    aria-modal="true"
    aria-label="Loot roll"
  >
    <div class="ng-panel" on:click|stopPropagation>
      <div class="ng-kicker"><i class="material-icons" aria-hidden="true">casino</i> Loot Roll</div>
      <p class="ng-item rarity-{quality}">{label}</p>
      {#if roll.expiresAt}
        <div class="ng-timer" aria-live="polite">
          {secondsLeft > 0 ? `Expires in ${secondsLeft}s` : 'Expiring…'}
        </div>
      {/if}
      <div class="ng-actions">
        <button class="ng-act need" type="button" bind:this={needBtn} on:click={() => choose('need')}>Need</button>
        <button class="ng-act greed" type="button" on:click={() => choose('greed')}>Greed</button>
        <button class="ng-act pass" type="button" on:click={() => choose('pass')}>Pass</button>
      </div>
    </div>
  </div>
{/if}
