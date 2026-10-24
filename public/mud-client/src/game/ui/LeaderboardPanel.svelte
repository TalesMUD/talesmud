<script>
  import { leaderboardView } from '../leaderboardView.js';

  export let store = null;
  export let embedded = false;
  export let onClose = null;

  let status = 'loading';
  let view = { rows: [], yourRank: null, limit: 25 };
  let loadedFor = null;

  $: self = store ? ($store?.character?.id || $store?.character?.name || '') : '';
  $: if (typeof window !== 'undefined' && self !== loadedFor) {
    loadedFor = self;
    load(self);
  }

  async function load(who) {
    status = 'loading';
    try {
      const res = await fetch('/api/leaderboard?limit=25&self=' + encodeURIComponent(who || ''));
      if (!res.ok) throw new Error(String(res.status));
      view = leaderboardView(await res.json());
      status = 'ready';
    } catch {
      status = 'error';
    }
  }

  function close() {
    if (onClose) onClose();
  }
</script>

<div class="leaderboard-panel" class:embedded>
  <div class="board-header">
    <div class="board-title">Leaderboard</div>
    {#if !embedded}
      <button class="board-close" type="button" aria-label="Close leaderboard" on:click={close}>×</button>
    {/if}
  </div>

  {#if status === 'loading'}
    <p class="board-note">Loading ranks…</p>
  {:else if status === 'error'}
    <p class="board-note">The leaderboard is unavailable.</p>
    <button class="board-retry" type="button" on:click={() => load(self)}>Try again</button>
  {:else if view.rows.length === 0}
    <p class="board-note">No ranked characters yet.</p>
  {:else}
    <div class="board-list" role="table" aria-label="Leaderboard">
      {#each view.rows as row (row.rank + '-' + row.name)}
        <div class="rank-row" class:yours={row.yours} role="row">
          <span class="rank-num">{row.rank}</span>
          <span class="rank-who">
            <span class="rank-name">{row.name}</span>
            <span class="rank-meta">{row.race} {row.className}</span>
          </span>
          <span class="rank-level">L{row.level}</span>
          <span class="rank-score">{row.itemScore}</span>
        </div>
      {/each}
    </div>
  {/if}

  {#if view.yourRank}
    <div class="your-rank" class:outside={view.yourRank.outside}>
      <span class="rank-num">{view.yourRank.rank}</span>
      <span class="rank-who">
        <span class="rank-name">Your rank</span>
        <span class="rank-meta">{view.yourRank.name}</span>
      </span>
      <span class="rank-level">L{view.yourRank.level}</span>
      <span class="rank-score">{view.yourRank.itemScore}</span>
    </div>
  {:else if status === 'ready'}
    <p class="board-note yours-note">Your character is not ranked.</p>
  {/if}
</div>

<style>
  .leaderboard-panel {
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: #151312;
    color: #e7e5e4;
    border: 0;
    border-radius: 16px;
    box-shadow: none;
  }

  .leaderboard-panel.embedded {
    position: absolute;
    inset: 0;
    background: transparent;
    border-radius: 0;
  }

  .leaderboard-panel.embedded .board-header {
    display: none;
  }

  .board-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 8px 8px 8px 16px;
  }

  .board-title {
    font-size: 16px;
    font-weight: 650;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .board-close,
  .board-retry {
    min-width: 44px;
    min-height: 44px;
    border: 0;
    border-radius: 12px;
    background: #1c1a18;
    color: #e7e5e4;
    cursor: pointer;
    font-size: 20px;
  }

  .board-retry {
    margin: 0 16px 16px;
    font-size: 14px;
    padding: 0 16px;
  }

  .board-note {
    margin: 0;
    padding: 8px 16px 16px;
    color: #a8a29e;
  }

  .board-list {
    overflow: auto;
    min-height: 0;
    padding: 0 8px 8px;
  }

  .rank-row,
  .your-rank {
    display: grid;
    grid-template-columns: 36px 1fr auto auto;
    gap: 8px;
    align-items: center;
    min-height: 44px;
    padding: 4px 8px;
    border-radius: 12px;
  }

  .rank-row.yours,
  .your-rank {
    background: #1c1a18;
  }

  .rank-num,
  .rank-level,
  .rank-score {
    font-variant-numeric: tabular-nums;
  }

  .rank-who {
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .rank-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .rank-meta {
    color: #a8a29e;
    font-size: 13px;
  }

  .rank-score {
    font-weight: 650;
  }

  .yours-note {
    padding-top: 0;
  }
</style>
