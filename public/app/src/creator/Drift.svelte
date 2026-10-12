<script>
  import { onMount } from "svelte";
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { exportDrift, getDrift } from "../api/health.js";
  import {
    canExportDrift,
    driftCounts,
    driftEntityPath,
    filterDrift,
    formatDriftValue,
    yamlBundle,
  } from "./driftView.js";

  const { isAuthenticated, authToken } = getAuth();

  let report = null;
  let loading = false;
  let error = "";
  let kind = "all";
  let query = "";
  let openKey = "";
  let exportingId = "";
  let exportingAll = false;
  let loaded = false;

  $: changes = report?.changes || [];
  $: counts = driftCounts(changes);
  $: visible = filterDrift(changes, { kind, query });
  $: exportableCount = changes.filter(canExportDrift).length;

  function rowKey(row) {
    return `${row.type}\0${row.id}\0${row.kind}`;
  }

  function when(value) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "";
    return date.toLocaleString();
  }

  function shortCommit(commit) {
    if (!commit) return "none";
    return commit.slice(0, 7);
  }

  async function errorText(err, fallback) {
    const data = err?.response?.data;
    if (data && typeof data.text === "function") {
      try {
        const parsed = JSON.parse(await data.text());
        if (parsed?.error) return parsed.error;
      } catch {
        // The body was not JSON.
      }
    }
    return data?.error || fallback;
  }

  function saveBlob(filename, blob) {
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
  }

  async function load() {
    if (!$authToken) return;
    loading = true;
    error = "";
    try {
      report = await getDrift($authToken);
      loaded = true;
    } catch (err) {
      error = await errorText(err, "Drift is unavailable.");
      report = { changes: [] };
    } finally {
      loading = false;
    }
  }

  async function downloadOne(row) {
    if (!$authToken || !canExportDrift(row) || exportingId || exportingAll) return;
    exportingId = rowKey(row);
    error = "";
    try {
      const blob = await exportDrift($authToken, row.type, row.id);
      saveBlob(`${row.type}-${row.id}.yaml`, blob);
    } catch (err) {
      error = await errorText(err, "Could not export that entity.");
    } finally {
      exportingId = "";
    }
  }

  async function downloadAll() {
    if (!$authToken || exportingAll || exportingId || exportableCount === 0) return;
    exportingAll = true;
    error = "";
    const parts = [];
    const failed = [];
    try {
      for (const row of changes) {
        if (!canExportDrift(row)) continue;
        try {
          const blob = await exportDrift($authToken, row.type, row.id);
          parts.push({ type: row.type, id: row.id, text: await blob.text() });
        } catch {
          failed.push(`${row.type} ${row.id}`);
        }
      }
      if (parts.length === 0) {
        error = "Could not export any drifted entity.";
        return;
      }
      saveBlob("drift.yaml", new Blob([yamlBundle(parts)], { type: "application/yaml" }));
      if (failed.length > 0) {
        error = `Exported ${parts.length}. Failed: ${failed.join(", ")}.`;
      }
    } finally {
      exportingAll = false;
    }
  }

  function openEntity(event, row) {
    const path = driftEntityPath(row.type, row.id);
    if (!path) return;
    event.preventDefault();
    navigateTo(path);
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
        <div class="label-caps">Operate · Drift</div>
        <h1 class="text-2xl font-bold tracking-tight">Drift</h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm">
          Database edits since the last import. The next content import overwrites them.
        </p>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-outline" type="button" on:click={load} disabled={loading || exportingAll}>
          <span class="material-symbols-outlined text-sm">refresh</span>
          Refresh
        </button>
        <button
          class="btn btn-primary"
          type="button"
          on:click={downloadAll}
          disabled={exportingAll || exportingId || exportableCount === 0}
          title="Download one YAML file for every added or changed entity"
        >
          {exportingAll ? "Exporting…" : "Export all"}
        </button>
      </div>
    </div>
  </div>

  <div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-6 pb-5 thin-scrollbar">
    {#if error}
      <div class="rounded-md border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-100">{error}</div>
    {/if}

    {#if !$isAuthenticated && !report}
      <div class="card p-4 text-sm text-slate-400">Log in to see drift.</div>
    {:else}
      <div class="card p-4 space-y-3">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-sm text-slate-400">
          <div>
            {counts.changed} changed · {counts.added} added · {counts.removed} removed
            {#if report?.importedAt}
              <span class="text-slate-500"> · Imported {when(report.importedAt)} · {shortCommit(report.contentCommit)}</span>
            {/if}
          </div>
          <div class="text-xs text-slate-500">Export all skips removed entities. Each file is importer YAML.</div>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <select class="input-base" bind:value={kind} aria-label="Drift kind">
            <option value="all">All kinds</option>
            <option value="changed">Changed</option>
            <option value="added">Added</option>
            <option value="removed">Removed</option>
          </select>
          <input class="input-base" type="search" placeholder="Filter id, name, or field text…" bind:value={query} />
        </div>
      </div>

      <div class="card overflow-x-auto">
        <table class="w-full text-sm" style="min-width: 760px;">
          <thead>
            <tr class="text-left text-[10px] uppercase tracking-wider text-slate-500">
              <th class="px-3 py-2">Kind</th>
              <th class="px-3 py-2">Entity</th>
              <th class="px-3 py-2">Diff</th>
              <th class="px-3 py-2"></th>
            </tr>
          </thead>
          <tbody>
            {#if loading && !report}
              <tr><td class="px-3 py-6 text-slate-500" colspan="4">Loading drift…</td></tr>
            {:else if visible.length === 0}
              <tr><td class="px-3 py-6 text-slate-500" colspan="4">{changes.length === 0 ? "No drifted entities." : "No drifted entities match."}</td></tr>
            {:else}
              {#each visible as row (rowKey(row))}
                {@const path = driftEntityPath(row.type, row.id)}
                {@const key = rowKey(row)}
                <tr class="border-t border-slate-800 align-top">
                  <td class="px-3 py-2">
                    <span class={row.kind === "removed" ? "text-red-300" : row.kind === "added" ? "text-emerald-300" : "text-amber-300"}>{row.kind}</span>
                  </td>
                  <td class="px-3 py-2">
                    <div class="text-slate-500">{row.type}</div>
                    {#if path}
                      <a class="font-mono text-emerald-300 hover:underline" href={path} on:click={(event) => openEntity(event, row)}>{row.id}</a>
                    {:else}
                      <span class="font-mono text-slate-300">{row.id}</span>
                    {/if}
                    {#if row.name}<div class="text-slate-400">{row.name}</div>{/if}
                  </td>
                  <td class="px-3 py-2">
                    {#if row.kind === "added"}
                      <div class="text-slate-400">Added after the import. There is no before value.</div>
                    {:else if row.kind === "removed"}
                      <div class="text-slate-400">Removed from the database. There is no live YAML to export.</div>
                    {:else if !(row.fields || []).length}
                      <div class="text-slate-400">The entity changed, and no field diff was returned.</div>
                    {:else}
                      <button class="text-xs text-primary" type="button" on:click={() => (openKey = openKey === key ? "" : key)}>
                        {openKey === key ? "Hide diff" : `Before / after (${row.fields.length})`}
                      </button>
                      {#if openKey === key}
                        <div class="mt-2 space-y-2">
                          {#each row.fields as field, index (`${field.path}-${index}`)}
                            <div>
                              <div class="font-mono text-xs text-slate-500">{field.path}</div>
                              <div class="grid grid-cols-1 lg:grid-cols-2 gap-2 mt-1">
                                <pre class="drift-json">{formatDriftValue(field.before)}</pre>
                                <pre class="drift-json">{formatDriftValue(field.after)}</pre>
                              </div>
                            </div>
                          {/each}
                        </div>
                      {/if}
                    {/if}
                  </td>
                  <td class="px-3 py-2 whitespace-nowrap">
                    {#if canExportDrift(row)}
                      <button
                        class="btn btn-outline text-xs"
                        type="button"
                        disabled={exportingAll || exportingId === key}
                        on:click={() => downloadOne(row)}
                      >
                        {exportingId === key ? "Exporting…" : "Export YAML"}
                      </button>
                    {/if}
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>

<style>
  .drift-json {
    max-height: 240px;
    overflow: auto;
    margin: 0;
    padding: 8px;
    border-radius: 6px;
    background: #0f172a;
    color: #cbd5e1;
    font-size: 11px;
    line-height: 1.4;
    white-space: pre-wrap;
    word-break: break-word;
  }
</style>
