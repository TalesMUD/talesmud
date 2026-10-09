<script>
  import { onMount } from "svelte";
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { exportDrift, getDrift, getHealth, putMute } from "../api/health.js";

  const { isAuthenticated, authToken } = getAuth();
  const hitCap = 200;

  let report = null;
  let drift = null;
  let loading = false;
  let error = "";
  let query = "";
  let open = {};
  let muting = "";
  let loaded = false;

  const tabByType = {
    room: "rooms",
    npc: "npcs",
    item: "item-templates",
    dialog: "dialogs",
    quest: "quests",
    script: "scripts",
    skill: "skills",
    charactertemplate: "character-templates",
  };

  function normType(entityType) {
    return String(entityType || "").toLowerCase().replaceAll("_", "");
  }

  function entityPath(entityType, id, related) {
    const kind = normType(entityType);
    if (!id || kind === "loottable") return "";
    if (kind === "quest") {
      return `/creator/quests/debug?id=${encodeURIComponent(id)}`;
    }
    if (kind === "spawner") {
      const room = (related || []).find((item) => normType(item.type) === "room" && item.id);
      return room ? `/creator/rooms?id=${encodeURIComponent(room.id)}` : "";
    }
    const tab = tabByType[kind];
    return tab ? `/creator/${tab}?id=${encodeURIComponent(id)}` : "";
  }

  function openEntity(event, entityType, id, related) {
    const path = entityPath(entityType, id, related);
    if (!path) return;
    event.preventDefault();
    navigateTo(path);
  }

  function shortCommit(commit) {
    if (!commit) return "none";
    return commit.slice(0, 7);
  }

  function when(value) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "";
    return date.toLocaleString();
  }

  $: rules = report?.rules || [];
  $: needle = query.trim().toLowerCase();
  $: visibleRules = rules.filter((rule) => {
    if (!needle) return true;
    const blob = `${rule.id} ${rule.title} ${rule.summary}`.toLowerCase();
    if (blob.includes(needle)) return true;
    return (rule.hits || []).some((hit) => hitText(hit).includes(needle));
  });

  function hitText(hit) {
    const related = (hit.related || []).map((item) => `${item.type} ${item.id}`).join(" ");
    return `${hit.entityId} ${hit.entityName || ""} ${hit.field || ""} ${hit.message || ""} ${related}`.toLowerCase();
  }

  function shownHits(rule) {
    const hits = rule.hits || [];
    const matched = needle ? hits.filter((hit) => hitText(hit).includes(needle) || `${rule.title} ${rule.id}`.toLowerCase().includes(needle)) : hits;
    return matched;
  }

  async function load() {
    if (!$isAuthenticated || !$authToken) return;
    loading = true;
    error = "";
    try {
      const [health, driftReport] = await Promise.all([
        getHealth($authToken),
        getDrift($authToken).catch(() => null),
      ]);
      report = health;
      drift = driftReport;
      loaded = true;
    } catch (err) {
      error = err?.response?.data?.error || "Content health is unavailable.";
    } finally {
      loading = false;
    }
  }

  async function toggleMute(rule) {
    if (!$authToken || muting) return;
    muting = rule.id;
    try {
      await putMute($authToken, rule.id, !rule.muted);
      await load();
    } catch (err) {
      error = err?.response?.data?.error || "Could not update the mute list.";
    } finally {
      muting = "";
    }
  }

  async function downloadDrift(row) {
    if (!$authToken) return;
    try {
      const blob = await exportDrift($authToken, row.type, row.id);
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${row.type}-${row.id}.yaml`;
      link.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      error = err?.response?.data?.error || "Could not export that entity.";
    }
  }

  function toggle(id) {
    open = { ...open, [id]: !open[id] };
  }

  function mark(rule) {
    if (rule.muted) return "–";
    if ((rule.hits || []).length === 0) return "✓";
    if (rule.severity === "warning") return "!";
    if (rule.severity === "info") return "i";
    return "✕";
  }

  function markClass(rule) {
    if (rule.muted || (rule.hits || []).length === 0) return "text-slate-500";
    if (rule.severity === "warning" || rule.severity === "info") return "text-amber-300";
    return "text-red-300";
  }

  onMount(load);

  $: if ($isAuthenticated && $authToken && !loaded && !loading) {
    load();
  }
</script>

<div class="flex h-full min-h-0 flex-col">
  <div class="px-6 pt-5 pb-3 flex-shrink-0">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
      <div class="space-y-1">
        <div class="label-caps">Operate · Health</div>
        <h1 class="text-2xl font-bold tracking-tight">Content health</h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm">
          {#if report}
            Last run {when(report.generatedAt)} · content <span class="font-mono text-slate-300">{shortCommit(report.contentCommit)}</span>
          {:else}
            Checks the live world against the engine rules.
          {/if}
          Rules are engine-generic. Content packs add their own.
        </p>
      </div>
      <button class="btn btn-primary" type="button" on:click={load} disabled={loading}>
        <span class="material-symbols-outlined text-sm">refresh</span>
        {loading ? "Running…" : "Run now"}
      </button>
    </div>
  </div>

  <div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-6 pb-6 thin-scrollbar">
    {#if error}
      <div class="rounded-md border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-100">{error}</div>
    {/if}

    <div class="grid grid-cols-2 lg:grid-cols-5 gap-3">
      <div class="card p-4">
        <div class="label-caps">Blocking</div>
        <div class="mt-1 text-3xl font-bold text-red-300">{report?.summary?.errors || 0}</div>
        <div class="text-xs text-slate-500">player-visible bugs</div>
      </div>
      <div class="card p-4">
        <div class="label-caps">Warnings</div>
        <div class="mt-1 text-3xl font-bold text-amber-300">{report?.summary?.warnings || 0}</div>
        <div class="text-xs text-slate-500">should fix</div>
      </div>
      <div class="card p-4">
        <div class="label-caps">Muted</div>
        <div class="mt-1 text-3xl font-bold text-slate-400">{report?.summary?.muted || 0}</div>
        <div class="text-xs text-slate-500">hits hidden from the counts</div>
      </div>
      <div class="card p-4">
        <div class="label-caps">Drift</div>
        <div class="mt-1 text-3xl font-bold text-amber-300">{report?.summary?.drift || 0}</div>
        <div class="text-xs text-slate-500">DB differs from the last import</div>
      </div>
      <div class="card p-4">
        <div class="label-caps">Live anomalies</div>
        <div class="mt-1 text-3xl font-bold text-amber-300">{report?.summary?.liveAnomalies || 0}</div>
        <div class="text-xs text-slate-500">characters and instance rooms</div>
      </div>
    </div>

    <div class="flex flex-col xl:flex-row gap-4 items-start">
      <div class="card p-4 flex-1 min-w-0 w-full">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="label-caps">Rules</div>
          <input class="input-base sm:max-w-xs" type="search" placeholder="Search rules and hits…" bind:value={query} />
        </div>
        <div class="mt-3 divide-y divide-slate-800">
          {#if loading && !report}
            <div class="py-8 text-center text-sm text-slate-500">Checking content health…</div>
          {:else if visibleRules.length === 0}
            <div class="py-8 text-center text-sm text-slate-500">No rules match.</div>
          {:else}
            {#each visibleRules as rule (rule.id)}
              {@const hits = shownHits(rule)}
              <div class="py-3">
                <div class="flex items-start gap-3">
                  <button class="mt-0.5 w-6 text-left text-lg font-bold {markClass(rule)}" type="button" on:click={() => toggle(rule.id)} aria-label="Toggle {rule.title}">
                    {mark(rule)}
                  </button>
                  <button class="min-w-0 flex-1 text-left" type="button" on:click={() => toggle(rule.id)}>
                    <div class="font-semibold">{rule.title}</div>
                    <div class="text-xs text-slate-500 truncate">{rule.summary}</div>
                  </button>
                  <div class="text-lg font-bold tabular-nums {markClass(rule)}">{hits.length}</div>
                  <button class="btn btn-outline px-3 py-1" type="button" disabled={muting === rule.id} on:click={() => toggleMute(rule)}>
                    {rule.muted ? "Unmute" : "Mute"}
                  </button>
                </div>
                {#if open[rule.id]}
                  <div class="mt-2 ml-9 space-y-1">
                    {#if hits.length === 0}
                      <div class="text-xs text-slate-500">No hits.</div>
                    {:else}
                      {#each hits.slice(0, hitCap) as hit, index (`${hit.entityId}-${hit.field}-${index}`)}
                        <div class="text-sm text-slate-300">
                          {#if entityPath(hit.entityType, hit.entityId, hit.related)}
                            <a class="font-mono text-emerald-300 hover:underline" href={entityPath(hit.entityType, hit.entityId, hit.related)} on:click={(event) => openEntity(event, hit.entityType, hit.entityId, hit.related)}>{hit.entityId}</a>
                          {:else}
                            <span class="font-mono text-slate-400">{hit.entityId}</span>
                          {/if}
                          {#if hit.entityName}<span class="text-slate-400"> {hit.entityName}</span>{/if}
                          <span class="text-slate-500"> — {hit.message}</span>
                          {#each hit.related || [] as related}
                            {#if entityPath(related.type, related.id)}
                              <a class="ml-1 font-mono text-emerald-300 hover:underline" href={entityPath(related.type, related.id)} on:click={(event) => openEntity(event, related.type, related.id)}>{related.id}</a>
                            {/if}
                          {/each}
                        </div>
                      {/each}
                      {#if hits.length > hitCap}
                        <div class="text-xs text-slate-500">Showing {hitCap} of {hits.length} hits. Narrow the search to see the rest.</div>
                      {/if}
                      {#if hits[0]?.fixHint}
                        <div class="text-xs text-slate-500">Fix: {hits[0].fixHint}</div>
                      {/if}
                    {/if}
                  </div>
                {/if}
              </div>
            {/each}
          {/if}
        </div>
      </div>

      <div class="w-full xl:w-96 space-y-4">
        <div class="card p-4">
          <div class="flex items-center justify-between gap-2">
            <div class="label-caps">Drift · DB since last import</div>
            <a class="text-xs text-emerald-300 hover:underline" href="/creator/drift" on:click={(event) => { event.preventDefault(); navigateTo("/creator/drift"); }}>Drift page</a>
          </div>
          {#if drift?.importedAt}
            <p class="mt-1 text-xs text-slate-500">Imported {when(drift.importedAt)} · {shortCommit(drift.contentCommit)}</p>
          {/if}
          {#if (drift?.changes || []).length === 0}
            <p class="mt-3 text-sm text-slate-400">No drifted entities.</p>
          {:else}
            <div class="mt-3 space-y-2">
              {#each drift.changes as row (`${row.type}-${row.id}-${row.kind}`)}
                <div class="text-sm">
                  <div class="flex items-center justify-between gap-2">
                    <div>
                      {#if entityPath(row.type, row.id)}
                        <a class="font-mono text-emerald-300 hover:underline" href={entityPath(row.type, row.id)} on:click={(event) => openEntity(event, row.type, row.id)}>{row.id}</a>
                      {:else}
                        <span class="font-mono text-slate-300">{row.id}</span>
                      {/if}
                      <span class="text-slate-500"> {row.kind}</span>
                      {#if row.name}<span class="text-slate-400"> {row.name}</span>{/if}
                    </div>
                    <button class="btn btn-outline px-2 py-1" type="button" on:click={() => downloadDrift(row)}>Export</button>
                  </div>
                  {#if row.fields?.length}
                    <div class="text-xs text-slate-500 truncate">{row.fields.map((field) => field.path).slice(0, 4).join(", ")}</div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
          <p class="mt-3 text-xs text-slate-500">The next content import overwrites these edits.</p>
        </div>

        <div class="card p-4">
          <div class="label-caps">Live anomalies</div>
          {#if (report?.liveAnomalies || []).length === 0}
            <p class="mt-3 text-sm text-slate-400">No live anomalies.</p>
          {:else}
            <div class="mt-3 space-y-2">
              {#each report.liveAnomalies as anomaly, index (`${anomaly.entityId}-${index}`)}
                <p class="text-sm text-slate-300">
                  <span class="text-amber-300">●</span>
                  {#if entityPath(anomaly.entityType, anomaly.entityId)}
                    <a class="font-mono text-emerald-300 hover:underline" href={entityPath(anomaly.entityType, anomaly.entityId)} on:click={(event) => openEntity(event, anomaly.entityType, anomaly.entityId)}>{anomaly.entityId}</a>
                  {/if}
                  {anomaly.message}
                </p>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>
