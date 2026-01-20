<script>
  const entries = [
    {
      date: "12.06.2020",
      title: "Creator tools go live",
      body:
        "First minimalistic version live supporting room creation and character templates.",
    },
  ];
</script>

<div class="px-6 py-8">
  <div class="max-w-5xl mx-auto space-y-6">
    <div class="flex items-end justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">News</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          Updates from the TalesMUD team and release highlights.
        </p>
      </div>
    </div>

    <div class="space-y-4">
      {#each entries as entry}
        <article class="card p-6 space-y-2">
          <div class="text-xs font-mono text-slate-400">{entry.date}</div>
          <h2 class="text-lg font-semibold">{entry.title}</h2>
          <p class="text-sm text-slate-500 dark:text-slate-400">{entry.body}</p>
        </article>
      {/each}
    </div>
  </div>
</div>
