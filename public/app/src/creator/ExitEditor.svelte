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
  <div class="col-span-4">
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
    class="col-span-1 text-slate-400 hover:text-accent-red"
    type="button"
    on:click={() => deleteExit(exit)}
  >
    <span class="material-symbols-outlined text-sm">close</span>
  </button>
</div>
