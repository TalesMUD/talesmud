<script>
  import { tick } from "svelte";
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { searchContent } from "../api/search.js";
  import { groupHits, isTypingTarget, typeLabel } from "./searchGroups.js";

  // The top-bar input (data-creator-search) is wired after the nav rebase.
  let open = false;
  let query = "";
  let hits = [];
  let active = 0;
  let loading = false;
  let note = "";
  let timer;
  let requestGen = 0;
  let inputEl;

  const { isAuthenticated, authToken } = getAuth();
  $: groups = groupHits(hits);

  function onWindowKey(event) {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") {
      if (isTypingTarget(event.target)) return;
      event.preventDefault();
      openPalette();
      return;
    }
    if (!open) return;
    if (event.key === "Escape") {
      event.preventDefault();
      close();
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      active = Math.min(active + 1, Math.max(hits.length - 1, 0));
      revealActive();
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      active = Math.max(active - 1, 0);
      revealActive();
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      choose(hits[active]);
    }
  }

  function openPalette() {
    open = true;
    tick().then(() => inputEl && inputEl.focus());
    if (query.trim()) scheduleSearch();
  }

  function close() {
    open = false;
    clearTimeout(timer);
  }

  function scheduleSearch() {
    clearTimeout(timer);
    timer = setTimeout(runSearch, 150);
  }

  async function runSearch() {
    const q = query.trim();
    const gen = ++requestGen;
    if (!q) {
      hits = [];
      note = "";
      loading = false;
      active = 0;
      return;
    }
    if (!$isAuthenticated || !$authToken) {
      hits = [];
      note = "Log in to search.";
      loading = false;
      return;
    }
    loading = true;
    note = "";
    try {
      const data = await searchContent($authToken, q);
      if (gen !== requestGen) return;
      hits = Array.isArray(data?.hits) ? data.hits : [];
      active = 0;
      if (hits.length === 0) note = "No matches.";
    } catch {
      if (gen !== requestGen) return;
      hits = [];
      note = "Search failed.";
    } finally {
      if (gen === requestGen) loading = false;
    }
  }

  function revealActive() {
    tick().then(() => {
      const row = document.querySelector(`[data-search-index="${active}"]`);
      if (row && row.scrollIntoView) row.scrollIntoView({ block: "nearest" });
    });
  }

  function choose(hit) {
    if (!hit?.path) return;
    close();
    navigateTo(hit.path);
  }
</script>

<svelte:window on:keydown={onWindowKey} />

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="sp-backdrop" on:click|self={close}>
    <div class="sp-card" role="dialog" aria-modal="true" aria-label="Search the world">
      <input
        bind:this={inputEl}
        bind:value={query}
        on:input={scheduleSearch}
        class="input-base"
        type="search"
        placeholder="Search rooms, NPCs, items, quests…"
        aria-label="Search the world"
      />
      <div class="sp-results">
        {#if loading}
          <p class="sp-note">Searching…</p>
        {/if}
        {#each groups as group}
          <div class="sp-type">{typeLabel(group.type)}</div>
          {#each group.hits as hit}
            <button
              class="sp-hit"
              class:sp-active={hit.index === active}
              type="button"
              data-search-index={hit.index}
              on:mouseenter={() => (active = hit.index)}
              on:click={() => choose(hit)}
            >
              <span class="sp-name">{hit.name || hit.id}</span>
              <span class="sp-id">{hit.id}</span>
              {#if hit.snippet && hit.snippet !== hit.name && hit.snippet !== hit.id}
                <span class="sp-snippet">{hit.snippet}</span>
              {/if}
            </button>
          {/each}
        {/each}
        {#if note && !loading}
          <p class="sp-note">{note}</p>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .sp-backdrop {
    position: fixed;
    inset: 0;
    z-index: 240;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 12vh;
    background: rgba(15, 23, 42, 0.45);
  }
  .sp-card {
    width: min(640px, calc(100vw - 2rem));
    border-radius: 12px;
    background: white;
    border: 1px solid rgb(226 232 240);
    box-shadow: 0 20px 50px rgba(15, 23, 42, 0.25);
    padding: 0.75rem;
  }
  :global(.dark) .sp-card {
    background: rgb(15 23 42);
    border-color: rgb(51 65 85);
  }
  .sp-results {
    max-height: 50vh;
    overflow: auto;
    margin-top: 0.5rem;
  }
  .sp-type {
    padding: 0.4rem 0.5rem 0.15rem;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: rgb(100 116 139);
  }
  .sp-hit {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.75rem;
    width: 100%;
    text-align: left;
    border: 0;
    background: transparent;
    border-radius: 8px;
    padding: 0.45rem 0.5rem;
    cursor: pointer;
  }
  .sp-active {
    background: rgba(99, 102, 241, 0.12);
  }
  .sp-name {
    font-weight: 600;
  }
  .sp-id,
  .sp-snippet {
    font-family: ui-monospace, monospace;
    font-size: 12px;
    color: rgb(100 116 139);
  }
  .sp-note {
    padding: 0.6rem 0.5rem;
    color: rgb(100 116 139);
    font-size: 13px;
  }
</style>
