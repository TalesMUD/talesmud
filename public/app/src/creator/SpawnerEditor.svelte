<script>
  export let spawner;
  export let npcTemplates = [];
  export let onDelete;
  export let onSave;
  export let isNew = false;

  // Parse duration string to seconds for display
  function parseDurationToSeconds(duration) {
    if (!duration) return 60;
    if (typeof duration === "number") return duration / 1000000000; // nanoseconds to seconds

    const str = String(duration);
    // Handle Go duration format (e.g., "30s", "1m", "1h30m")
    let total = 0;
    const hourMatch = str.match(/(\d+)h/);
    const minMatch = str.match(/(\d+)m(?!s)/);
    const secMatch = str.match(/(\d+)s/);

    if (hourMatch) total += parseInt(hourMatch[1]) * 3600;
    if (minMatch) total += parseInt(minMatch[1]) * 60;
    if (secMatch) total += parseInt(secMatch[1]);

    return total || 60;
  }

  // Format seconds to display string
  function formatDuration(seconds) {
    if (seconds >= 3600) {
      const h = Math.floor(seconds / 3600);
      const m = Math.floor((seconds % 3600) / 60);
      const s = seconds % 60;
      let result = `${h}h`;
      if (m > 0) result += `${m}m`;
      if (s > 0) result += `${s}s`;
      return result;
    } else if (seconds >= 60) {
      const m = Math.floor(seconds / 60);
      const s = seconds % 60;
      return s > 0 ? `${m}m${s}s` : `${m}m`;
    }
    return `${seconds}s`;
  }

  // Convert seconds to Go nanoseconds for API
  function secondsToNanoseconds(seconds) {
    return seconds * 1000000000;
  }

  // Local state for duration inputs (in seconds for easier editing)
  let spawnIntervalSeconds = parseDurationToSeconds(spawner.spawnInterval);
  let respawnOverrideSeconds = spawner.respawnTimeOverride
    ? parseDurationToSeconds(spawner.respawnTimeOverride)
    : null;
  let useRespawnOverride = spawner.respawnTimeOverride != null;

  // Update spawner when local values change
  $: spawner.spawnInterval = secondsToNanoseconds(spawnIntervalSeconds);
  $: spawner.respawnTimeOverride = useRespawnOverride
    ? secondsToNanoseconds(respawnOverrideSeconds || 60)
    : null;

  function toggleRespawnOverride() {
    useRespawnOverride = !useRespawnOverride;
    if (useRespawnOverride && respawnOverrideSeconds === null) {
      respawnOverrideSeconds = 60;
    }
  }

  function getTemplateName(templateId) {
    const template = npcTemplates.find(t => t.id === templateId);
    return template ? template.name : "Unknown Template";
  }
</script>

<div class="p-4 rounded-lg border border-slate-200 dark:border-slate-700/50 bg-slate-50 dark:bg-slate-900/50 space-y-3">
  <div class="flex items-center justify-between">
    <div class="text-xs font-bold uppercase tracking-wider text-slate-400">
      {#if spawner.name}
        {spawner.name}
      {:else if spawner.templateId}
        {getTemplateName(spawner.templateId)} Spawner
      {:else}
        New Spawner
      {/if}
    </div>
    <div class="flex items-center gap-2">
      {#if isNew}
        <button
          class="text-xs text-primary hover:underline"
          type="button"
          on:click={() => onSave(spawner)}
        >
          Save
        </button>
      {/if}
      <button
        class="text-xs text-accent-red hover:underline"
        type="button"
        on:click={() => onDelete(spawner)}
      >
        {isNew ? "Cancel" : "Delete"}
      </button>
    </div>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <div class="space-y-1.5">
      <label class="label-caps" for={`spawner-name-${spawner.id}`}>Name (Optional)</label>
      <input
        class="input-base text-xs"
        id={`spawner-name-${spawner.id}`}
        type="text"
        placeholder="e.g., Forest Goblin Spawner"
        bind:value={spawner.name}
      />
    </div>
    <div class="space-y-1.5">
      <label class="label-caps" for={`spawner-template-${spawner.id}`}>NPC Template</label>
      <select
        class="input-base text-xs"
        id={`spawner-template-${spawner.id}`}
        bind:value={spawner.templateId}
      >
        <option value="" disabled>Select NPC template</option>
        {#each npcTemplates as template}
          <option value={template.id}>
            {template.name} (Lvl {template.level || 1})
          </option>
        {/each}
      </select>
    </div>
  </div>

  <div class="grid grid-cols-3 gap-4">
    <div class="space-y-1.5">
      <label class="label-caps" for={`spawner-max-${spawner.id}`}>Max Instances</label>
      <input
        class="input-base text-xs text-center"
        id={`spawner-max-${spawner.id}`}
        type="number"
        min="1"
        max="100"
        bind:value={spawner.maxInstances}
      />
      <p class="text-[9px] text-slate-500">Max alive at once</p>
    </div>
    <div class="space-y-1.5">
      <label class="label-caps" for={`spawner-initial-${spawner.id}`}>Initial Count</label>
      <input
        class="input-base text-xs text-center"
        id={`spawner-initial-${spawner.id}`}
        type="number"
        min="0"
        max="100"
        bind:value={spawner.initialCount}
      />
      <p class="text-[9px] text-slate-500">Spawn on load</p>
    </div>
    <div class="space-y-1.5">
      <label class="label-caps" for={`spawner-interval-${spawner.id}`}>Spawn Interval</label>
      <div class="flex items-center gap-1">
        <input
          class="input-base text-xs text-center flex-1"
          id={`spawner-interval-${spawner.id}`}
          type="number"
          min="1"
          bind:value={spawnIntervalSeconds}
        />
        <span class="text-[10px] text-slate-500">sec</span>
      </div>
      <p class="text-[9px] text-slate-500">{formatDuration(spawnIntervalSeconds)}</p>
    </div>
  </div>

  <div class="pt-2 border-t border-slate-200 dark:border-slate-700/50">
    <div class="flex items-center justify-between mb-2">
      <label class="label-caps flex items-center gap-2">
        <input
          type="checkbox"
          class="rounded border-slate-300 dark:border-slate-600"
          checked={useRespawnOverride}
          on:change={toggleRespawnOverride}
        />
        Override Respawn Time
      </label>
    </div>
    {#if useRespawnOverride}
      <div class="flex items-center gap-2">
        <input
          class="input-base text-xs text-center w-24"
          type="number"
          min="1"
          bind:value={respawnOverrideSeconds}
        />
        <span class="text-[10px] text-slate-500">seconds ({formatDuration(respawnOverrideSeconds || 60)})</span>
      </div>
      <p class="text-[9px] text-slate-500 mt-1">Overrides the NPC template's respawn time</p>
    {/if}
  </div>
</div>
