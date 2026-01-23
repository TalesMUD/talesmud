<script>
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";

  export let config;
  export let store;

  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  let filterInput = "";
  let hasLoadedData = false;

  const loadData = async (cb) => {
    if (!$isAuthenticated || !$authToken) return;
    config.get(
      $authToken,
      $store.filters,
      (all) => {
        store.setElements(all);
        hasLoadedData = true;
        if (cb) cb();
      },
      (err) => console.log(err)
    );
  };

  // Reactively load data when auth becomes available
  $: if ($isAuthenticated && $authToken && !hasLoadedData) {
    loadData(() => {
      selectElement($store.elements[0]);
    });
  }

  onMount(async () => {
    // Initial load attempt (may be skipped if auth not ready)
    loadData(() => {
      selectElement($store.elements[0]);
    });
  });

  const isDraft = (isnew) => {
    if (isnew === true) return "border border-dashed border-amber-400/60";
    return "";
  };

  const deleteElement = async () => {
    config.delete(
      $authToken,
      $store.selectedElement.id,
      () => {
        loadData(() => {
          if ($store.elements) {
            store.setSelectedElement($store.elements[0]);
          }
        });
      },
      () => {
        console.log("delete error.");
      }
    );
  };

  const create = async () => {
    config.create(
      $authToken,
      $store.selectedElement,
      (element) => {
        loadData();
        selectElement(element);
      },
      () => {
        console.log("create error.");
      }
    );
  };

  const selectElement = (element) => {
    if (!element) return;
    if (config.beforeSelect) {
      config.beforeSelect(element);
    }

    store.setSelectedElement(element, () => {
      refreshUI();
    });
  };

  const addFilter = (filter) => {
    if (filter.includes(":")) {
      let keyval = filter.split(":");
      store.addFilter(keyval[0], keyval[1]);
      filterInput = "";

      loadData(() => {
        selectElement($store.elements[0]);
      });
    }
  };

  const removeFilter = (filter) => {
    if (filter.includes(":")) {
      let keyval = filter.split(":");
      store.removeFilter(keyval[0]);

      loadData(() => {
        selectElement($store.elements[0]);
      });
    }
  };

  const refreshUI = () => {
    if (config.refreshUI) {
      config.refreshUI();
    }
  };

  const update = () => {
    config.update(
      $authToken,
      $store.selectedElement.id,
      $store.selectedElement,
      () => {
        loadData();
      },
      () => {
        console.log("update error.");
      }
    );
  };

  const labels = {
    create: config?.labels?.create || "Create",
    update: config?.labels?.update || "Update",
    delete: config?.labels?.delete || "Delete",
  };
</script>

{#if !$isAuthenticated}
  <div class="px-6 py-12 text-center text-sm text-slate-500 dark:text-slate-400">
    Please log in to access Creator tools.
  </div>
{:else}
  <div class="flex h-[calc(100vh-128px)]">
    <aside class="w-72 border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-col">
      <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
        <h2 class="text-sm font-semibold">{config.listTitle || config.title}</h2>
        <div class="flex items-center gap-1">
          <button
            class="btn btn-ghost p-1.5"
            type="button"
            on:click={() => store.toggleFilter()}
          >
            <span class="material-symbols-outlined text-lg">filter_alt</span>
          </button>
          <button
            class="btn btn-ghost p-1.5"
            type="button"
            on:click={() => config.new(selectElement)}
          >
            <span class="material-symbols-outlined text-lg">add</span>
          </button>
        </div>
      </div>
      {#if $store.filterActive}
        <div class="p-3 border-b border-slate-200 dark:border-slate-800">
          <div class="flex items-center gap-2">
            <input
              class="input-base text-xs"
              placeholder="key:value"
              bind:value={filterInput}
            />
            <button
              class="btn btn-outline px-3 py-2 text-xs"
              type="button"
              on:click={() => addFilter(filterInput)}
            >
              Add
            </button>
          </div>
          {#if $store.filters.length}
            <div class="mt-2 flex flex-wrap gap-2">
              {#each $store.filters as filter}
                <button
                  class="text-xs rounded-full border border-slate-200 px-2 py-1 text-slate-500 hover:text-primary dark:border-slate-700"
                  type="button"
                  on:click={() => removeFilter(`${filter.key}:${filter.val}`)}
                >
                  {filter.key}:{filter.val}
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
      <div class="flex-1 overflow-y-auto p-2 space-y-1">
        {#each $store.elements as element}
          <button
            type="button"
            class={`w-full text-left px-3 py-2.5 rounded-md transition-colors ${
              element === $store.selectedElement
                ? "bg-primary/10 text-primary border-l-2 border-primary"
                : "hover:bg-slate-100 dark:hover:bg-slate-800"
            }`}
            on:click={() => selectElement(element)}
          >
            <div class="text-sm font-medium flex items-center gap-2">
              <span>{element.name}</span>
              {#if element.isNew}
                <span class="text-[10px] uppercase tracking-wider bg-primary/20 text-primary px-2 py-0.5 rounded">
                  Draft
                </span>
              {/if}
            </div>
            <div class="text-[10px] text-slate-400 dark:text-slate-500 font-mono">
              {config.badge ? config.badge(element) : `ID: ${element.id}`}
            </div>
          </button>
        {/each}
      </div>
    </aside>

    <section class="flex-1 overflow-y-auto">
      <div class="p-8 pb-4">
        <div class="max-w-5xl mx-auto">
          <div class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-8">
            <div class="space-y-1">
              <h1 class="text-3xl font-bold tracking-tight">{config.title}</h1>
              {#if config.subtitle}
                <p class="text-slate-500 dark:text-slate-400 text-sm">
                  {config.subtitle}
                </p>
              {/if}
            </div>
            <div class="flex items-center gap-3">
              {#if config.extraActions}
                {#each config.extraActions as action}
                  <button
                    class={`btn ${action.variant || "btn-outline"}`}
                    type="button"
                    on:click={() => action.onClick()}
                  >
                    {#if action.icon}
                      <span class="material-symbols-outlined text-sm">{action.icon}</span>
                    {/if}
                    {action.label}
                  </button>
                {/each}
              {/if}
              {#if $store.selectedElement}
                {#if $store.selectedElement.isNew}
                  <button class="btn btn-primary" type="button" on:click={() => create()}>
                    <span class="material-symbols-outlined text-sm">add_box</span>
                    {labels.create}
                  </button>
                {:else}
                  <button class="btn btn-danger" type="button" on:click={() => deleteElement()}>
                    <span class="material-symbols-outlined text-sm">delete</span>
                    {labels.delete}
                  </button>
                  <button class="btn btn-primary" type="button" on:click={() => update()}>
                    <span class="material-symbols-outlined text-sm">save</span>
                    {labels.update}
                  </button>
                {/if}
              {/if}
            </div>
          </div>

          {#if $store.selectedElement}
            <div class={`card p-6 space-y-6 ${isDraft($store.selectedElement.isNew)} relative`}>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="md:col-span-2 space-y-4">
                  <div class="space-y-1.5">
                    <label class="label-caps" for="element_name">Name</label>
                    <input
                      id="element_name"
                      type="text"
                      class="input-base text-lg font-semibold"
                      bind:value={$store.selectedElement.name}
                    />
                  </div>
                  {#if !config.hideDetails}
                    <div class="space-y-1.5">
                      <label class="label-caps" for="element_description">Description</label>
                      <textarea
                        id="element_description"
                        rows="3"
                        class="input-base"
                        bind:value={$store.selectedElement.description}
                      />
                    </div>
                    <div class="space-y-1.5">
                      <label class="label-caps" for="element_detail">Details</label>
                      <textarea
                        id="element_detail"
                        rows="3"
                        class="input-base"
                        bind:value={$store.selectedElement.detail}
                      />
                    </div>
                  {/if}
                </div>
                <div class="space-y-4">
                  <div class="space-y-1.5">
                    <label class="label-caps" for="element_id">ID</label>
                    <input
                      id="element_id"
                      type="text"
                      class="input-base font-mono text-xs"
                      bind:value={$store.selectedElement.id}
                      disabled={!$store.selectedElement.isNew}
                    />
                  </div>
                  <slot name="hero" />
                </div>
              </div>

              <slot name="content" />
            </div>

            <div class="mt-6">
              <slot name="extensions" />
            </div>
          {:else}
            <div class="text-sm text-slate-500 dark:text-slate-400">
              Select an entry to begin editing.
            </div>
          {/if}
        </div>
      </div>
    </section>
  </div>
{/if}
