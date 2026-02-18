<script>
  import { getAuth } from "../auth.js";
  import { getGuestStatsSummary, getGuestDailyStats } from "../api/gueststats.js";

  const { isAuthenticated, authToken } = getAuth();

  let summary = null;
  let loading = true;
  let error = null;

  function loadStats() {
    loading = true;
    error = null;
    getGuestStatsSummary(
      $authToken,
      (data) => {
        summary = data;
        loading = false;
      },
      (err) => {
        console.error("Failed to load guest stats:", err);
        error = "Failed to load guest statistics";
        loading = false;
      }
    );
  }

  $: if ($isAuthenticated && $authToken) {
    loadStats();
  }

  function formatDuration(seconds) {
    if (!seconds || seconds === 0) return "—";
    const m = Math.floor(seconds / 60);
    const s = Math.floor(seconds % 60);
    if (m === 0) return `${s}s`;
    return `${m}m ${s}s`;
  }

  function formatDate(dateStr) {
    const d = new Date(dateStr + "T00:00:00Z");
    return d.toLocaleDateString(undefined, { month: "short", day: "numeric", timeZone: "UTC" });
  }

  // Compute bar chart heights for the last 14 days
  $: chartDays = summary?.daily?.slice(-14) ?? [];
  $: chartMax = Math.max(1, ...chartDays.map((d) => d.count));
</script>

<div class="px-6 py-8">
  <div class="max-w-6xl mx-auto space-y-6">
    <div class="flex items-start justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Guest Statistics</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          Anonymous guest session analytics. No personal data is stored.
        </p>
      </div>
      <button class="btn btn-outline" on:click={loadStats} disabled={loading}>
        <span class="material-symbols-outlined text-base">refresh</span>
        Refresh
      </button>
    </div>

    {#if loading}
      <div class="card p-12 text-center">
        <p class="text-slate-400">Loading statistics...</p>
      </div>
    {:else if error}
      <div class="card p-12 text-center">
        <p class="text-red-400">{error}</p>
        <button class="btn btn-outline mt-4" on:click={loadStats}>Retry</button>
      </div>
    {:else if summary}
      <!-- Summary cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="card p-5">
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1">Total All-Time</p>
          <p class="text-3xl font-bold">{summary.totalAllTime}</p>
          <p class="text-xs text-slate-400 mt-1">guest sessions</p>
        </div>
        <div class="card p-5">
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1">Today</p>
          <p class="text-3xl font-bold">{summary.today}</p>
          <p class="text-xs text-slate-400 mt-1">sessions (UTC)</p>
        </div>
        <div class="card p-5">
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1">This Week</p>
          <p class="text-3xl font-bold">{summary.thisWeek}</p>
          <p class="text-xs text-slate-400 mt-1">last 7 days</p>
        </div>
        <div class="card p-5 relative">
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1">Active Now</p>
          <p class="text-3xl font-bold">{summary.activeNow}</p>
          <p class="text-xs text-slate-400 mt-1">live sessions</p>
          {#if summary.activeNow > 0}
            <span class="absolute top-4 right-4 w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
          {/if}
        </div>
      </div>

      <!-- Bar chart: last 14 days -->
      {#if chartDays.length > 0}
        <div class="card p-5">
          <h2 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-4">
            Sessions — Last 14 Days
          </h2>
          <div class="flex items-end gap-1 h-32">
            {#each chartDays as day}
              <div class="flex-1 flex flex-col items-center gap-1 group">
                <div
                  class="w-full rounded-t bg-primary/70 hover:bg-primary transition-colors relative"
                  style="height: {chartMax > 0 ? Math.max(4, (day.count / chartMax) * 100) : 4}%;"
                  title="{day.date}: {day.count} session{day.count !== 1 ? 's' : ''}"
                >
                  {#if day.count > 0}
                    <span class="absolute -top-5 left-1/2 -translate-x-1/2 text-[10px] font-bold opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                      {day.count}
                    </span>
                  {/if}
                </div>
                <span class="text-[9px] text-slate-400 rotate-45 origin-left mt-1 whitespace-nowrap hidden sm:block">
                  {formatDate(day.date)}
                </span>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Daily breakdown table: last 30 days -->
      <div class="card overflow-hidden">
        <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800">
          <h2 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">
            Daily Breakdown — Last 30 Days
          </h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800">
                <th class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">Date</th>
                <th class="px-4 py-3 text-right text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">Sessions</th>
                <th class="px-4 py-3 text-right text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">Avg Duration</th>
              </tr>
            </thead>
            <tbody>
              {#each [...summary.daily].reverse() as day}
                <tr class="border-b border-slate-100 dark:border-slate-800/60 hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors">
                  <td class="px-4 py-2.5 font-mono text-xs text-slate-600 dark:text-slate-300">
                    {day.date}
                  </td>
                  <td class="px-4 py-2.5 text-right font-semibold {day.count > 0 ? 'text-primary' : 'text-slate-400'}">
                    {day.count > 0 ? day.count : "—"}
                  </td>
                  <td class="px-4 py-2.5 text-right text-slate-500 dark:text-slate-400">
                    {formatDuration(day.avgDurationSeconds)}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>

      <p class="text-xs text-slate-400 text-center">
        All times are UTC. No IP addresses, locations, or personal data are stored.
      </p>
    {/if}
  </div>
</div>
