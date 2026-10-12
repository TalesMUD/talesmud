<script>
  import { createEventDispatcher } from "svelte";
  import { isNavActive } from "./navState.js";

  export let groups = [];
  export let pathname = "";
  export let collapsed = false;
  export let folded = {};
  export let idPrefix = "creator-group";

  const dispatch = createEventDispatcher();

  function groupDomId(id) {
    return `${idPrefix}-${id}`;
  }
</script>

<div class="flex flex-col gap-3 px-2 pb-4">
  {#each groups as group (group.id)}
    {@const isFolded = folded[group.id] === true}
    <div>
      <button
        type="button"
        class="flex w-full items-center gap-2 rounded-md px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-slate-500 hover:text-slate-300 dark:text-slate-500"
        class:justify-center={collapsed}
        aria-expanded={!isFolded}
        aria-controls={groupDomId(group.id)}
        title={group.label}
        aria-label={group.label}
        on:click={() => dispatch("toggleGroup", group.id)}
      >
        {#if collapsed}
          <span aria-hidden="true">{group.label.slice(0, 1)}</span>
        {:else}
          <span class="flex-1 text-left">{group.label}</span>
          <span class="material-symbols-outlined text-sm" aria-hidden="true">
            {isFolded ? "chevron_right" : "expand_more"}
          </span>
        {/if}
      </button>
      {#if !isFolded}
        <div id={groupDomId(group.id)} role="group" aria-label={group.label} class="mt-0.5 flex flex-col gap-0.5">
          {#each group.items as item (item.id)}
            {@const active = isNavActive(pathname, item)}
            <a
              class="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-100 {active
                ? 'bg-emerald-500/15 text-primary dark:text-emerald-300'
                : ''}"
              class:justify-center={collapsed}
              href={item.href}
              aria-current={active ? "page" : undefined}
              title={item.title || item.label}
              aria-label={item.label}
              on:click={() => dispatch("navigate")}
            >
              <span class="material-symbols-outlined text-[20px]" aria-hidden="true">{item.icon}</span>
              {#if !collapsed}
                <span class="truncate">{item.label}</span>
              {/if}
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/each}
</div>
