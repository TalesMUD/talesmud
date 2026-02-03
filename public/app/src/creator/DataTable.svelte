<script>
  import { createEventDispatcher } from "svelte";
  import DataTableFilterBar from "./DataTableFilterBar.svelte";

  export let columns = [];
  export let elements = [];
  export let selectedElement = null;
  export let filterValues = {};
  export let sortKey = "";
  export let sortDir = "asc";
  export let compact = false;
  /** Optional function (element) => { color, title } | null for a colored row indicator dot */
  export let rowIndicator = null;

  const dispatch = createEventDispatcher();

  // When compact, hide low-priority columns (priority > 2)
  $: visibleColumns = compact
    ? columns.filter((c) => (c.priority ?? 1) <= 2)
    : columns;

  function getNestedValue(obj, path) {
    return path.split(".").reduce((acc, part) => acc?.[part], obj) ?? "";
  }

  function getCellValue(el, col) {
    if (col.accessor) return col.accessor(el);
    return getNestedValue(el, col.key);
  }

  function getSortValue(el, col) {
    if (col.sortAccessor) return col.sortAccessor(el);
    return getCellValue(el, col);
  }

  // Client-side filtering
  $: filteredElements = elements.filter((el) => {
    return Object.entries(filterValues).every(([key, val]) => {
      if (!val && val !== 0) return true;
      const col = columns.find((c) => c.key === key);
      if (!col) return true;
      const cellValue = getCellValue(el, col);
      const cellStr = String(cellValue).toLowerCase();
      const filterStr = String(val).toLowerCase();

      if (col.type === "select" || col.type === "boolean") {
        return cellStr === filterStr;
      }
      if (col.type === "number") {
        return String(cellValue) === String(val);
      }
      return cellStr.includes(filterStr);
    });
  });

  // Client-side sorting
  $: sortedElements = (() => {
    if (!sortKey) return filteredElements;
    const col = columns.find((c) => c.key === sortKey);
    if (!col || col.sortable === false) return filteredElements;
    return [...filteredElements].sort((a, b) => {
      const aVal = getSortValue(a, col);
      const bVal = getSortValue(b, col);
      let cmp = 0;
      if (typeof aVal === "number" && typeof bVal === "number") {
        cmp = aVal - bVal;
      } else {
        cmp = String(aVal).localeCompare(String(bVal));
      }
      return sortDir === "desc" ? -cmp : cmp;
    });
  })();

  function toggleSort(key) {
    const col = columns.find((c) => c.key === key);
    if (col?.sortable === false) return;
    if (sortKey === key) {
      if (sortDir === "asc") {
        dispatch("sort", { key, dir: "desc" });
      } else {
        dispatch("sort", { key: "", dir: "asc" });
      }
    } else {
      dispatch("sort", { key, dir: "asc" });
    }
  }

  function selectRow(element) {
    dispatch("select", element);
  }
</script>

<div class="space-y-4">
  <DataTableFilterBar
    columns={visibleColumns}
    {filterValues}
    totalCount={elements.length}
    filteredCount={sortedElements.length}
    on:filterChange
    on:resetFilters
  />

  <div class="card overflow-hidden">
    <div class="overflow-x-auto thin-scrollbar">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-slate-700">
            {#each visibleColumns as col (col.key)}
              <th
                class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-500 select-none whitespace-nowrap"
                class:cursor-pointer={col.sortable !== false}
                class:hover:text-slate-300={col.sortable !== false}
                style={col.width ? `width: ${col.width}px; min-width: ${col.width}px` : ""}
                on:click={() => toggleSort(col.key)}
              >
                <span class="inline-flex items-center gap-1">
                  {col.label}
                  {#if sortKey === col.key}
                    <span class="material-symbols-outlined text-primary" style="font-size: 14px">
                      {sortDir === "asc" ? "arrow_upward" : "arrow_downward"}
                    </span>
                  {/if}
                </span>
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#if sortedElements.length === 0}
            <tr>
              <td colspan={visibleColumns.length} class="px-4 py-10 text-center text-xs text-slate-500">
                {#if elements.length === 0}
                  No entries found.
                {:else}
                  No results match your filters.
                {/if}
              </td>
            </tr>
          {:else}
            {#each sortedElements as element (element.id)}
              <tr
                class="border-b border-slate-800/50 cursor-pointer transition-colors
                  {element === selectedElement
                    ? 'bg-primary/10 border-l-2 border-l-primary'
                    : 'hover:bg-slate-800/40'}"
                on:click={() => selectRow(element)}
              >
                {#each visibleColumns as col (col.key)}
                  <td
                    class="px-4 py-2.5 text-xs whitespace-nowrap {col.mono ? 'font-mono text-slate-400' : 'text-slate-300'}"
                  >
                    {#if col.key === "name"}
                      {@const indicator = rowIndicator ? rowIndicator(element) : null}
                      {#if indicator}
                        <span
                          class="inline-block w-2 h-2 rounded-full mr-2 flex-shrink-0"
                          style="background-color: {indicator.color}"
                          title={indicator.title || ""}
                        ></span>
                      {/if}
                      <span class="font-medium text-slate-200">
                        {getCellValue(element, col)}
                      </span>
                      {#if element.isNew}
                        <span class="ml-1.5 text-[9px] uppercase tracking-wider bg-primary/20 text-primary px-1.5 py-0.5 rounded">
                          Draft
                        </span>
                      {/if}
                    {:else}
                      {getCellValue(element, col)}
                    {/if}
                  </td>
                {/each}
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
    height: 6px;
    width: 6px;
  }
  .thin-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .thin-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(148, 163, 184, 0.2);
    border-radius: 3px;
  }
  .thin-scrollbar::-webkit-scrollbar-thumb:hover {
    background: rgba(148, 163, 184, 0.35);
  }
</style>
