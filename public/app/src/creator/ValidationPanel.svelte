<script>
  export let result = null;
  export let loading = false;
  export let unavailable = "";

  $: issues = result?.issues || [];
  $: errors = issues.filter((issue) => issue.severity === "error");
  $: warnings = issues.filter((issue) => issue.severity === "warning");
</script>

{#if loading}
  <div class="rounded-md border border-slate-700 bg-slate-900/50 px-3 py-2 text-xs text-slate-400">
    Checking content...
  </div>
{:else if unavailable}
  <div class="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
    {unavailable}
  </div>
{:else if issues.length}
  <div class="space-y-2">
    {#each errors as issue}
      <div class="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-xs text-red-100">
        <div class="flex items-start gap-2">
          <span class="material-symbols-outlined mt-0.5 text-sm text-red-300">error</span>
          <div class="min-w-0">
            <div class="font-semibold leading-snug">{issue.message}</div>
            <div class="mt-1 font-mono text-[10px] text-red-200/70">
              {issue.field || issue.code}{#if issue.refId} -> {issue.refId}{/if}
            </div>
          </div>
        </div>
      </div>
    {/each}
    {#each warnings as issue}
      <div class="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
        <div class="flex items-start gap-2">
          <span class="material-symbols-outlined mt-0.5 text-sm text-amber-300">warning</span>
          <div class="min-w-0">
            <div class="font-semibold leading-snug">{issue.message}</div>
            <div class="mt-1 font-mono text-[10px] text-amber-200/70">
              {issue.field || issue.code}{#if issue.refId} -> {issue.refId}{/if}
            </div>
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}
