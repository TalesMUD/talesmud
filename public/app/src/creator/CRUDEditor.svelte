<script>
  import { onDestroy, onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import { getWorldDiagnosticsAsync, validateEntityAsync } from "../api/validation.js";
  import DataTable from "./DataTable.svelte";
  import ValidationPanel from "./ValidationPanel.svelte";

  export let config;
  export let store;

  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
  };

  let hasLoadedData = false;
  let validationResult = null;
  let validationLoading = false;
  let validationUnavailable = "";
  let validationTimer;
  let issuesByEntity = {};

  $: hasValidationErrors = (validationResult?.errors || 0) > 0;

  const loadData = async (cb) => {
    if (!$isAuthenticated || !$authToken) return;
    config.get(
      $authToken,
      $store.filters,
      (all) => {
        store.setElements(all);
        hasLoadedData = true;
        loadDiagnostics();
        if (cb) cb();
      },
      (err) => console.log(err)
    );
  };

  const loadDiagnostics = async () => {
    if (!config.entityType || !$isAuthenticated || !$authToken) return;
    try {
      const diagnostics = await getWorldDiagnosticsAsync($authToken);
      const grouped = {};
      for (const issue of diagnostics?.issues || []) {
        if (issue.entityType !== config.entityType || !issue.entityId) continue;
        if (!grouped[issue.entityId]) grouped[issue.entityId] = [];
        grouped[issue.entityId].push(issue);
      }
      issuesByEntity = grouped;
    } catch (err) {
      issuesByEntity = {};
    }
  };

  // Reactively load data when auth becomes available
  $: if ($isAuthenticated && $authToken && !hasLoadedData) {
    loadData();
  }

  onMount(async () => {
    loadData();
  });

  onDestroy(() => {
    if (validationTimer) clearTimeout(validationTimer);
  });

  const extractValidationResult = (err) => {
    const data = err?.response?.data;
    if (data?.issues) return data;
    return null;
  };

  const runValidation = async (entity = $store.selectedElement) => {
    if (!config.entityType || !$isAuthenticated || !$authToken || !entity) {
      validationResult = null;
      validationUnavailable = "";
      return null;
    }

    validationLoading = true;
    validationUnavailable = "";
    try {
      validationResult = await validateEntityAsync($authToken, config.entityType, entity);
      return validationResult;
    } catch (err) {
      const backendResult = extractValidationResult(err);
      if (backendResult) {
        validationResult = backendResult;
        return backendResult;
      }
      validationUnavailable = "Validation is unavailable right now.";
      return null;
    } finally {
      validationLoading = false;
    }
  };

  const scheduleValidation = () => {
    if (!config.entityType || !$store.selectedElement) return;
    if (validationTimer) clearTimeout(validationTimer);
    validationTimer = setTimeout(() => {
      runValidation();
    }, 350);
  };

  $: if (config.entityType && $isAuthenticated && $authToken && $store.selectedElement) {
    scheduleValidation();
  }

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
          store.closeDetail();
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
    const result = await runValidation();
    if (result?.errors > 0) return;
    config.create(
      $authToken,
      $store.selectedElement,
      (element) => {
        loadData();
        selectElement(element);
      },
      (err) => {
        const backendResult = extractValidationResult(err);
        if (backendResult) {
          validationResult = backendResult;
        } else {
          console.log("create error.");
        }
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
    store.openDetail();
  };

  const closeDetail = () => {
    store.closeDetail();
  };

  const refreshUI = () => {
    if (config.refreshUI) {
      config.refreshUI();
    }
  };

  const update = () => {
    runValidation().then((result) => {
      if (result?.errors > 0) return;

    config.update(
      $authToken,
      $store.selectedElement.id,
      $store.selectedElement,
      () => {
        loadData();
      },
      (err) => {
        const backendResult = extractValidationResult(err);
        if (backendResult) {
          validationResult = backendResult;
        } else {
          console.log("update error.");
        }
      }
    );
    });
  };

  const issueIndicator = (element) => {
    const issues = issuesByEntity[element.id] || [];
    if (issues.some((issue) => issue.severity === "error")) {
      return { color: "#ef4444", title: "Has validation errors" };
    }
    if (issues.length) {
      return { color: "#f59e0b", title: "Has validation warnings" };
    }
    return null;
  };

  const labels = {
    create: config?.labels?.create || "Create",
    update: config?.labels?.update || "Update",
    delete: config?.labels?.delete || "Delete",
  };

  // Navigate to previous/next element relative to current selection
  const selectPrevious = () => {
    const idx = $store.elements.indexOf($store.selectedElement);
    if (idx > 0) selectElement($store.elements[idx - 1]);
  };

  const selectNext = () => {
    const idx = $store.elements.indexOf($store.selectedElement);
    if (idx < $store.elements.length - 1) selectElement($store.elements[idx + 1]);
  };

  const newElement = () => {
    if (config.new) {
      config.new((element) => {
        if (config.beforeSelect) {
          config.beforeSelect(element);
        }
        store.setSelectedElement(element);
        store.openDetail();
      });
    }
  };
</script>

{#if !$isAuthenticated}
  <div class="px-6 py-12 text-center text-sm text-slate-500 dark:text-slate-400">
    Please log in to access Creator tools.
  </div>
{:else}
  <div class="flex flex-col h-[calc(100vh-128px)]">
    <!-- Header: Title + Action Buttons -->
    <div class="px-6 pt-5 pb-3 flex-shrink-0">
      <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div class="space-y-1">
          <h1 class="text-2xl font-bold tracking-tight">{config.title}</h1>
          {#if config.subtitle}
            <p class="text-slate-500 dark:text-slate-400 text-sm">{config.subtitle}</p>
          {/if}
        </div>
        <div class="flex items-center gap-3 flex-shrink-0">
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
          <button class="btn btn-outline" type="button" on:click={newElement}>
            <span class="material-symbols-outlined text-sm">add</span>
            New
          </button>
            {#if $store.selectedElement && $store.detailOpen}
              {#if $store.selectedElement.isNew}
              <button class="btn btn-primary" type="button" disabled={hasValidationErrors} on:click={() => create()}>
                <span class="material-symbols-outlined text-sm">add_box</span>
                {labels.create}
              </button>
            {:else}
              <button class="btn btn-danger" type="button" on:click={() => deleteElement()}>
                <span class="material-symbols-outlined text-sm">delete</span>
                {labels.delete}
              </button>
              <button class="btn btn-primary" type="button" disabled={hasValidationErrors} on:click={() => update()}>
                <span class="material-symbols-outlined text-sm">save</span>
                {labels.update}
              </button>
            {/if}
          {/if}
        </div>
      </div>
    </div>

    <!-- Main content: side-by-side table + detail -->
    <div class="flex-1 flex overflow-hidden px-6 pb-5 gap-5">
      <!-- Data Table Panel -->
      <div
        class="flex-shrink-0 overflow-y-auto transition-all duration-200 thin-scrollbar"
        style={$store.detailOpen ? "width: 44%; min-width: 380px;" : "width: 100%;"}
      >
        <DataTable
          columns={config.columns || []}
          elements={$store.elements}
          selectedElement={$store.detailOpen ? $store.selectedElement : null}
          filterValues={$store.tableFilterValues}
          sortKey={$store.sortKey}
          sortDir={$store.sortDir}
          compact={$store.detailOpen}
          rowIndicator={config.rowIndicator || issueIndicator}
          on:select={(e) => selectElement(e.detail)}
          on:sort={(e) => store.setSort(e.detail.key, e.detail.dir)}
          on:filterChange={(e) => store.setTableFilter(e.detail.key, e.detail.value)}
          on:resetFilters={() => store.resetTableFilters()}
        />
      </div>

      <!-- Detail Panel (visible when an element is selected and detail is open) -->
      {#if $store.detailOpen && $store.selectedElement}
        <div class="flex-1 overflow-y-auto min-w-[400px] thin-scrollbar">
          <!-- Detail header bar -->
          <div class="card px-4 py-2.5 mb-4 flex items-center justify-between">
            <div class="flex items-center gap-3 min-w-0">
              <button
                class="btn btn-ghost p-1.5 flex-shrink-0"
                type="button"
                on:click={closeDetail}
                title="Close detail"
              >
                <span class="material-symbols-outlined text-lg">close</span>
              </button>
              <div class="min-w-0">
                <div class="text-sm font-semibold truncate">{$store.selectedElement.name}</div>
                <div class="text-[10px] text-slate-400 font-mono truncate">{$store.selectedElement.id}</div>
              </div>
              {#if $store.selectedElement.isNew}
                <span class="text-[10px] uppercase tracking-wider bg-primary/20 text-primary px-2 py-0.5 rounded flex-shrink-0">
                  Draft
                </span>
              {/if}
            </div>
            <div class="flex items-center gap-1 flex-shrink-0">
              <button class="btn btn-ghost p-1" type="button" title="Previous" on:click={selectPrevious}>
                <span class="material-symbols-outlined text-sm">chevron_left</span>
              </button>
              <button class="btn btn-ghost p-1" type="button" title="Next" on:click={selectNext}>
                <span class="material-symbols-outlined text-sm">chevron_right</span>
              </button>
            </div>
          </div>

          <!-- Detail form -->
          <div class={`card p-6 space-y-6 ${isDraft($store.selectedElement.isNew)} relative`}>
            {#if config.entityType}
              <ValidationPanel
                result={validationResult}
                loading={validationLoading}
                unavailable={validationUnavailable}
              />
            {/if}

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
        </div>
      {/if}
    </div>
  </div>
{/if}

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
