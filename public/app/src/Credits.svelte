<script>
  const credits = [
    {
      name: "32x32 fantasy portrait set",
      url: "https://opengameart.org/content/32x32-fantasy-portrait-set",
      author: "CobraLad",
    },
  ];
</script>

<div class="px-6 py-8">
  <div class="max-w-4xl mx-auto space-y-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Credits</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400">
        Asset attributions used across the app and backend.
      </p>
    </div>

    <div class="card divide-y divide-slate-200 dark:divide-slate-800">
      {#each credits as item}
        <div class="p-4 flex items-center justify-between">
          <div>
            <div class="text-sm font-medium">{item.name}</div>
            <div class="text-xs text-slate-500 dark:text-slate-400">
              by {item.author}
            </div>
          </div>
          <a class="text-sm text-primary hover:underline" href={item.url}>
            View source
          </a>
        </div>
      {/each}
    </div>
  </div>
</div>
