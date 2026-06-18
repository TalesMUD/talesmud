<script>
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import { getWorldDiagnosticsAsync } from "../api/validation.js";

  const { isAuthenticated, authToken } = getAuth();

  let diagnostics = null;
  let loading = false;
  let error = "";
  let severityFilter = "";
  let entityTypeFilter = "";
  let textFilter = "";
  let loaded = false;

  $: issues = diagnostics?.issues || [];
  $: entityTypes = [...new Set(issues.map((issue) => issue.entityType).filter(Boolean))].sort();
  $: filteredIssues = issues.filter((issue) => {
    if (severityFilter && issue.severity !== severityFilter) return false;
    if (entityTypeFilter && issue.entityType !== entityTypeFilter) return false;
    if (textFilter) {
      const haystack = `${issue.message} ${issue.entityId} ${issue.field} ${issue.refId} ${issue.code}`.toLowerCase();
      if (!haystack.includes(textFilter.toLowerCase())) return false;
    }
    return true;
  });

  async function loadDiagnostics() {
    if (!$isAuthenticated || !$authToken) return;
    loading = true;
    error = "";
    try {
      diagnostics = await getWorldDiagnosticsAsync($authToken);
      loaded = true;
    } catch (err) {
      error = err?.response?.data?.error || "World diagnostics are unavailable.";
    } finally {
      loading = false;
    }
  }

  onMount(loadDiagnostics);

  $: if ($isAuthenticated && $authToken && !loaded && !loading) {
    loadDiagnostics();
  }
</script>

<div class="flex flex-col h-[calc(100vh-128px)]">
  <div class="px-6 pt-5 pb-3 flex-shrink-0">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
      <div class="space-y-1">
        <h1 class="text-2xl font-bold tracking-tight">World Health</h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm">Find broken references before they reach players.</p>
      </div>
      <button class="btn btn-outline" type="button" on:click={loadDiagnostics} disabled={loading}>
        <span class="material-symbols-outlined text-sm">refresh</span>
        Refresh
      </button>
    </div>
  </div>

  <div class="flex-1 overflow-y-auto px-6 pb-5 thin-scrollbar space-y-4">
    {#if error}
      <div class="rounded-md border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-100">{error}</div>
    {/if}

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="card p-4">
        <div class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Errors</div>
        <div class="mt-1 text-3xl font-bold text-red-300">{diagnostics?.errors || 0}</div>
      </div>
      <div class="card p-4">
        <div class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Warnings</div>
        <div class="mt-1 text-3xl font-bold text-amber-300">{diagnostics?.warnings || 0}</div>
      </div>
      <div class="card p-4">
        <div class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Status</div>
        <div class="mt-2 inline-flex items-center gap-2 text-sm font-semibold {diagnostics?.valid ? 'text-emerald-300' : 'text-red-300'}">
          <span class="material-symbols-outlined text-base">{diagnostics?.valid ? "check_circle" : "error"}</span>
          {diagnostics?.valid ? "Ready" : "Needs attention"}
        </div>
      </div>
    </div>

    <div class="card p-4 space-y-3">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <select class="input-base" bind:value={severityFilter}>
          <option value="">All severities</option>
          <option value="error">Errors</option>
          <option value="warning">Warnings</option>
        </select>
        <select class="input-base" bind:value={entityTypeFilter}>
          <option value="">All entity types</option>
          {#each entityTypes as type}
            <option value={type}>{type}</option>
          {/each}
        </select>
        <input class="input-base" type="search" placeholder="Filter issues..." bind:value={textFilter} />
      </div>
    </div>

    <div class="card overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-slate-700">
            <th class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-500">Severity</th>
            <th class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-500">Entity</th>
            <th class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-500">Field</th>
            <th class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-500">Issue</th>
          </tr>
        </thead>
        <tbody>
          {#if loading}
            <tr><td colspan="4" class="px-4 py-10 text-center text-xs text-slate-500">Checking world health...</td></tr>
          {:else if filteredIssues.length === 0}
            <tr><td colspan="4" class="px-4 py-10 text-center text-xs text-slate-500">No issues match the current filters.</td></tr>
          {:else}
            {#each filteredIssues as issue}
              <tr class="border-b border-slate-800/50">
                <td class="px-4 py-2.5 text-xs">
                  <span class="inline-flex items-center gap-1 font-semibold {issue.severity === 'error' ? 'text-red-300' : 'text-amber-300'}">
                    <span class="material-symbols-outlined text-sm">{issue.severity === "error" ? "error" : "warning"}</span>
                    {issue.severity}
                  </span>
                </td>
                <td class="px-4 py-2.5 text-xs font-mono text-slate-300">{issue.entityType}:{issue.entityId}</td>
                <td class="px-4 py-2.5 text-xs font-mono text-slate-400">{issue.field || "-"}</td>
                <td class="px-4 py-2.5 text-xs text-slate-200">
                  <div>{issue.message}</div>
                  {#if issue.refId}
                    <div class="mt-1 font-mono text-[10px] text-slate-500">{issue.refType}:{issue.refId}</div>
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</div>

<style>
  .thin-scrollbar::-webkit-scrollbar {
    width: 6px;
  }
  .thin-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .thin-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(148, 163, 184, 0.2);
    border-radius: 3px;
  }
</style>
