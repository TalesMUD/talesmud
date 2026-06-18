<script>
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import { getWorldValidationAsync } from "../api/world.js";

  const { isAuthenticated, authToken } = getAuth();

  let loading = false;
  let error = "";
  let report = { issues: [], errorCount: 0, warningCount: 0 };
  let selectedSeverity = "all";
  let selectedSystem = "all";

  $: issues = report?.issues || [];
  $: systems = Array.from(new Set(issues.map((issue) => issue.system))).sort();
  $: filteredIssues = issues.filter((issue) => {
    const severityMatch = selectedSeverity === "all" || issue.severity === selectedSeverity;
    const systemMatch = selectedSystem === "all" || issue.system === selectedSystem;
    return severityMatch && systemMatch;
  });
  $: groupedCounts = systems.map((system) => ({
    system,
    errors: issues.filter((issue) => issue.system === system && issue.severity === "error").length,
    warnings: issues.filter((issue) => issue.system === system && issue.severity === "warning").length,
  }));

  async function loadValidation() {
    if (!$authToken) return;
    loading = true;
    error = "";
    try {
      report = await getWorldValidationAsync($authToken);
    } catch (err) {
      console.error("Failed to load world validation:", err);
      error = err?.response?.data?.error || "Failed to load world health.";
    } finally {
      loading = false;
    }
  }

  function severityClass(severity) {
    if (severity === "error") {
      return "bg-red-100 text-red-700 border-red-200 dark:bg-red-950/40 dark:text-red-300 dark:border-red-900";
    }
    return "bg-amber-100 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900";
  }

  onMount(() => {
    loadValidation();
  });
</script>

{#if !$isAuthenticated}
  <div class="px-6 py-12 text-center text-sm text-slate-500 dark:text-slate-400">
    Please log in to access Creator tools.
  </div>
{:else}
  <div class="min-h-[calc(100vh-128px)] overflow-y-auto">
    <div class="p-8 space-y-6">
      <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 class="text-3xl font-bold tracking-tight">World Health</h1>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
            Validate cross-system references before players hit broken content.
          </p>
        </div>
        <button class="btn btn-primary" type="button" on:click={loadValidation} disabled={loading}>
          <span class="material-symbols-outlined text-sm">refresh</span>
          {loading ? "Checking..." : "Run Check"}
        </button>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="card p-5">
          <div class="text-xs uppercase tracking-widest text-slate-400">Errors</div>
          <div class="text-3xl font-bold text-red-500 mt-2">{report.errorCount || 0}</div>
        </div>
        <div class="card p-5">
          <div class="text-xs uppercase tracking-widest text-slate-400">Warnings</div>
          <div class="text-3xl font-bold text-amber-500 mt-2">{report.warningCount || 0}</div>
        </div>
        <div class="card p-5">
          <div class="text-xs uppercase tracking-widest text-slate-400">Systems</div>
          <div class="text-3xl font-bold text-slate-800 dark:text-slate-100 mt-2">{systems.length}</div>
        </div>
      </div>

      {#if error}
        <div class="border border-red-200 dark:border-red-900 bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-300 px-4 py-3 text-sm">
          {error}
        </div>
      {/if}

      <div class="flex flex-wrap gap-3 items-center">
        <select class="input-base max-w-[180px]" bind:value={selectedSeverity}>
          <option value="all">All severities</option>
          <option value="error">Errors</option>
          <option value="warning">Warnings</option>
        </select>
        <select class="input-base max-w-[220px]" bind:value={selectedSystem}>
          <option value="all">All systems</option>
          {#each systems as system}
            <option value={system}>{system}</option>
          {/each}
        </select>
      </div>

      {#if groupedCounts.length > 0}
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
          {#each groupedCounts as row}
            <button
              type="button"
              class="border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-left px-4 py-3 hover:border-primary transition-colors"
              on:click={() => (selectedSystem = row.system)}
            >
              <div class="text-sm font-semibold text-slate-800 dark:text-slate-100">{row.system}</div>
              <div class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {row.errors} errors · {row.warnings} warnings
              </div>
            </button>
          {/each}
        </div>
      {/if}

      <div class="border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden">
        <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-4">
          <h2 class="text-sm font-semibold uppercase tracking-widest text-slate-500 dark:text-slate-400">
            Diagnostics
          </h2>
          <span class="text-xs text-slate-400">{filteredIssues.length} shown</span>
        </div>

        {#if loading}
          <div class="p-8 text-center text-sm text-slate-500 dark:text-slate-400">Checking world health...</div>
        {:else if filteredIssues.length === 0}
          <div class="p-8 text-center text-sm text-slate-500 dark:text-slate-400">
            No diagnostics match the selected filters.
          </div>
        {:else}
          <div class="overflow-x-auto">
            <table class="min-w-full text-sm">
              <thead class="bg-slate-50 dark:bg-slate-950 text-slate-500 dark:text-slate-400">
                <tr>
                  <th class="text-left px-4 py-3 font-semibold">Severity</th>
                  <th class="text-left px-4 py-3 font-semibold">System</th>
                  <th class="text-left px-4 py-3 font-semibold">Entity</th>
                  <th class="text-left px-4 py-3 font-semibold">Field</th>
                  <th class="text-left px-4 py-3 font-semibold">Message</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                {#each filteredIssues as issue}
                  <tr class="hover:bg-slate-50 dark:hover:bg-slate-950/50">
                    <td class="px-4 py-3">
                      <span class={`inline-flex items-center border px-2 py-0.5 text-xs font-semibold ${severityClass(issue.severity)}`}>
                        {issue.severity}
                      </span>
                    </td>
                    <td class="px-4 py-3 font-mono text-xs text-slate-600 dark:text-slate-300">{issue.system}</td>
                    <td class="px-4 py-3">
                      <div class="font-medium text-slate-800 dark:text-slate-100">{issue.entityType}</div>
                      <div class="font-mono text-xs text-slate-400">{issue.entityId || "-"}</div>
                    </td>
                    <td class="px-4 py-3 font-mono text-xs text-slate-500 dark:text-slate-400">{issue.field || "-"}</td>
                    <td class="px-4 py-3 text-slate-700 dark:text-slate-200">{issue.message}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
