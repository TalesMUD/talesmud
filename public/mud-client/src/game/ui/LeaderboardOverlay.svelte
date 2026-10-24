<script>
  import LeaderboardPanel from './LeaderboardPanel.svelte';

  export let store = null;

  $: open = !!(store && $store.leaderboardOverlayOpen);

  function close() {
    if (store && store.closeLeaderboard) store.closeLeaderboard();
  }

  function onKeydown(event) {
    if (open && event.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={onKeydown} />

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="leaderboard-overlay" role="presentation" on:click={close}>
    <div class="leaderboard-frame gf-panel" role="dialog" aria-modal="true" aria-label="Leaderboard" on:click|stopPropagation>
      <LeaderboardPanel {store} onClose={close} />
    </div>
  </div>
{/if}

<style>
  .leaderboard-overlay {
    position: fixed;
    inset: 0;
    z-index: 90;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgba(0, 0, 0, 0.45);
  }

  .leaderboard-frame {
    width: min(520px, 100%);
    max-height: min(80vh, 640px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: #151312;
    border: 0;
    border-radius: 16px;
    box-shadow: none;
  }

  @media (max-width: 768px) {
    .leaderboard-overlay {
      align-items: flex-end;
      padding: 0;
    }

    .leaderboard-frame {
      width: 100%;
      max-height: min(92dvh, 720px);
      border-radius: 16px 16px 0 0;
    }
  }
</style>
