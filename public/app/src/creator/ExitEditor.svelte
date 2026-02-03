<script>
  export let exit;
  export let valueHelp;
  export let deleteExit;
</script>

<div class="grid grid-cols-12 gap-3 items-center">
  <div class="col-span-3">
    <input
      class="input-base text-xs"
      placeholder="Direction"
      bind:value={exit.name}
    />
  </div>
  <div class="col-span-3">
    <input
      class="input-base text-xs"
      placeholder="Description"
      bind:value={exit.description}
    />
  </div>
  <div class="col-span-4">
    <select class="input-base text-xs" bind:value={exit.target}>
      <option value="" disabled>Select target</option>
      {#if $valueHelp}
        {#each $valueHelp as room}
          <option value={room.id}>
            {room.name} ({room.id})
          </option>
        {/each}
      {/if}
    </select>
  </div>
  <button
    class="col-span-1 flex items-center justify-center"
    type="button"
    title={exit.hidden ? "Hidden exit (click to make visible)" : "Visible exit (click to hide)"}
    on:click={() => exit.hidden = !exit.hidden}
  >
    <span
      class="material-symbols-outlined text-sm {exit.hidden ? 'text-amber-400' : 'text-slate-500 hover:text-slate-300'}"
    >
      {exit.hidden ? 'visibility_off' : 'visibility'}
    </span>
  </button>
  <button
    class="col-span-1 text-slate-400 hover:text-accent-red"
    type="button"
    on:click={() => deleteExit(exit)}
  >
    <span class="material-symbols-outlined text-sm">close</span>
  </button>
</div>
