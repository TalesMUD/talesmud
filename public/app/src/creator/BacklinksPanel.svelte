<script>
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { getRefs } from "../api/search.js";
  import { inboundCount, typeLabel } from "./searchGroups.js";

  export let type = "";
  export let id = "";

  let view = null;
  let note = "";
  let loading = false;
  let requestGen = 0;

  const { isAuthenticated, authToken } = getAuth();

  async function load(kind, entityId, token) {
    const gen = ++requestGen;
    if (!kind || !entityId || !token) {
      view = null;
      note = "";
      loading = false;
      return;
    }
    loading = true;
    note = "";
    try {
      const next = await getRefs(token, kind, entityId);
      if (gen !== requestGen) return;
      view = next;
    } catch (err) {
      if (gen !== requestGen) return;
      view = null;
      note = err?.response?.status === 404 ? "Not in the content index." : "References are unavailable.";
    } finally {
      if (gen === requestGen) loading = false;
    }
  }

  function openRef(event, path) {
    if (!path) return;
    event.preventDefault();
    navigateTo(path);
  }

  $: load(type, id, $isAuthenticated ? $authToken : "");
  $: count = view ? inboundCount(view) : null;
</script>

{#if type && id}
  <details class="bl-panel">
    <summary>
      Referenced by{count === null ? "" : ` (${count})`}
      {#if loading}
        <span class="bl-muted">…</span>
      {/if}
    </summary>
    {#if note}
      <p class="bl-note">{note}</p>
    {:else if view}
      {#if count === 0}
        <p class="bl-note">Nothing else references this.</p>
      {/if}
      {#each view.inbound || [] as group}
        <div class="bl-type">{typeLabel(group.type)}</div>
        {#each group.refs || [] as ref}
          <div class="bl-row">
            {#if ref.path}
              <a href={ref.path} on:click={(event) => openRef(event, ref.path)}>{ref.name || ref.id}</a>
            {:else}
              <span>{ref.name || ref.id}</span>
            {/if}
            <span class="bl-id">{ref.id}</span>
            {#if ref.reason}
              <span class="bl-reason">{ref.reason}</span>
            {/if}
          </div>
        {/each}
      {/each}
      {#if (view.outbound || []).length}
        <div class="bl-type">Links to</div>
        {#each view.outbound as group}
          {#each group.refs || [] as ref}
            <div class="bl-row">
              {#if ref.path}
                <a href={ref.path} on:click={(event) => openRef(event, ref.path)}>{typeLabel(ref.type)} {ref.name || ref.id}</a>
              {:else}
                <span>{typeLabel(ref.type)} {ref.name || ref.id}</span>
              {/if}
              {#if ref.reason}
                <span class="bl-reason">{ref.reason}</span>
              {/if}
            </div>
          {/each}
        {/each}
      {/if}
    {/if}
  </details>
{/if}

<style>
  .bl-panel {
    margin-top: 1.5rem;
    border: 1px solid rgb(226 232 240);
    border-radius: 12px;
    background: white;
    padding: 0.75rem 1rem;
  }
  :global(.dark) .bl-panel {
    background: rgb(15 23 42);
    border-color: rgb(51 65 85);
  }
  summary {
    cursor: pointer;
    font-size: 13px;
    font-weight: 600;
  }
  .bl-type {
    margin-top: 0.75rem;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: rgb(100 116 139);
  }
  .bl-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.75rem;
    padding: 0.25rem 0;
    font-size: 13px;
  }
  .bl-id,
  .bl-reason,
  .bl-muted,
  .bl-note {
    color: rgb(100 116 139);
  }
  .bl-id,
  .bl-reason {
    font-family: ui-monospace, monospace;
    font-size: 12px;
  }
  .bl-note {
    margin: 0.5rem 0 0;
    font-size: 13px;
  }
  a {
    color: inherit;
    text-decoration: underline;
  }
</style>
