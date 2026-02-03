<script>
  import { createEventDispatcher } from "svelte";

  export let columns = [];
  export let filterValues = {};
  export let totalCount = 0;
  export let filteredCount = 0;

  const dispatch = createEventDispatcher();

  $: filterableColumns = columns.filter(
    (c) => c.filterable !== false && c.type !== "computed"
  );

  $: hasActiveFilters = Object.values(filterValues).some((v) => v !== "" && v != null);

  function onInput(key, value) {
    dispatch("filterChange", { key, value });
  }

  function resetAll() {
    dispatch("resetFilters");
  }

  let debounceTimers = {};
  function onTextInput(key, e) {
    const value = e.target.value;
    clearTimeout(debounceTimers[key]);
    debounceTimers[key] = setTimeout(() => onInput(key, value), 200);
  }
</script>

<div class="flex flex-wrap items-end gap-3 px-4 py-3.5 rounded-lg bg-slate-800/40 border border-slate-700/50">
  {#each filterableColumns as col (col.key)}
    <div class="space-y-1 min-w-[100px]" style={col.width ? `max-width: ${col.width + 40}px` : ""}>
      <label class="text-[9px] font-bold uppercase tracking-wider text-slate-500" for={`filter-${col.key}`}>
        {col.label}
      </label>
      {#if col.type === "select"}
        <select
          id={`filter-${col.key}`}
          class="input-base text-xs py-1.5"
          value={filterValues[col.key] || ""}
          on:change={(e) => onInput(col.key, e.target.value)}
        >
          <option value="">All</option>
          {#each col.options || [] as opt}
            <option value={opt}>{opt}</option>
          {/each}
        </select>
      {:else if col.type === "boolean"}
        <select
          id={`filter-${col.key}`}
          class="input-base text-xs py-1.5"
          value={filterValues[col.key] || ""}
          on:change={(e) => onInput(col.key, e.target.value)}
        >
          <option value="">All</option>
          <option value="Yes">Yes</option>
          <option value="No">No</option>
        </select>
      {:else}
        <input
          id={`filter-${col.key}`}
          type={col.type === "number" ? "number" : "text"}
          class="input-base text-xs py-1.5"
          placeholder={col.label}
          value={filterValues[col.key] || ""}
          on:input={(e) => onTextInput(col.key, e)}
        />
      {/if}
    </div>
  {/each}

  <div class="flex items-center gap-3 ml-auto self-end pb-0.5">
    <span class="text-[10px] text-slate-500 whitespace-nowrap">
      {filteredCount} / {totalCount}
    </span>
    {#if hasActiveFilters}
      <button
        type="button"
        class="btn btn-ghost text-xs px-2 py-1.5 text-slate-400 hover:text-primary"
        on:click={resetAll}
      >
        <span class="material-symbols-outlined text-sm">filter_alt_off</span>
        Reset
      </button>
    {/if}
  </div>
</div>
