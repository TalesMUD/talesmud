<script>
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import { getAuth } from "../auth.js";
  import {
    getLootTables,
    getLootTable,
    createLootTable,
    updateLootTable,
    deleteLootTable,
    rollLootTable,
  } from "../api/loottables.js";
  import { getItemTemplates } from "../api/items.js";
  import { itemTemplateColumns, lootTableColumns } from "./tableColumns.js";

  const { isAuthenticated, authToken } = getAuth();
  const store = createStore();

  let itemTemplates = [];
  let rollN = 1000;
  let rollSeed = "";
  let rollLevel = 1;
  let rollBoss = false;
  let rollResult = null;
  let rollError = "";
  let rolling = false;

  const config = {
    title: "Loot tables",
    entityType: "loottable",
    subtitle: "Drop chances for item templates. Roll ×N previews frequency and does not grant items.",
    listTitle: "Loot tables",
    columns: lootTableColumns,
    hideDetails: true,
    labels: {
      create: "Create Loot Table",
      update: "Update Loot Table",
      delete: "Delete",
    },
    get: getLootTables,
    getElement: getLootTable,
    create: createLootTable,
    update: updateLootTable,
    delete: deleteLootTable,
    new: (select) => {
      select({
        id: uuidv4(),
        name: "New loot table",
        description: "",
        entries: [],
        goldMultiplier: 1,
        dropBonus: 0,
        isNew: true,
      });
    },
  };

  function loadTemplates() {
    if (!$isAuthenticated || !$authToken) return;
    getItemTemplates($authToken, [], (rows) => {
      itemTemplates = rows || [];
    }, () => {
      itemTemplates = [];
    });
  }

  $: if ($isAuthenticated && $authToken) loadTemplates();

  function entries() {
    const rows = $store.selectedElement?.entries;
    return Array.isArray(rows) ? rows : [];
  }

  function setEntries(next) {
    if (!$store.selectedElement) return;
    $store.selectedElement.entries = next;
  }

  function addEntry() {
    setEntries([
      ...entries(),
      {
        itemTemplateId: "",
        dropChance: 0.1,
        minQuantity: 1,
        maxQuantity: 1,
        guaranteed: false,
        bossOnly: false,
      },
    ]);
  }

  function removeEntry(index) {
    setEntries(entries().filter((_, i) => i !== index));
  }

  function percent(frequency) {
    const value = Number(frequency);
    if (!Number.isFinite(value)) return "—";
    return `${(value * 100).toFixed(1)}%`;
  }

  async function rollPreview() {
    const table = $store.selectedElement;
    if (!table?.id || table.isNew || !$authToken) return;
    rolling = true;
    rollError = "";
    const params = {
      n: rollN || 1000,
      playerLevel: rollLevel || 1,
      boss: rollBoss ? "1" : "0",
    };
    if (String(rollSeed).trim() !== "") params.seed = String(rollSeed).trim();
    try {
      rollResult = await rollLootTable($authToken, table.id, params);
      if (rollResult?.seed != null) rollSeed = String(rollResult.seed);
    } catch (err) {
      rollResult = null;
      rollError = err?.response?.data?.error || "Roll failed.";
    } finally {
      rolling = false;
    }
  }

  let rollFor = "";
  $: {
    const id = $store.selectedElement?.id || "";
    if (id !== rollFor) {
      rollFor = id;
      rollResult = null;
      rollError = "";
    }
  }
</script>

<CRUDEditor {config} {store}>
  <div slot="content" class="space-y-6">
    {#if $store.selectedElement}
      <div class="space-y-1.5">
        <label class="label-caps" for="loot-description">Description</label>
        <textarea id="loot-description" rows="2" class="input-base" bind:value={$store.selectedElement.description}></textarea>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="loot-gold">Gold multiplier</label>
          <input id="loot-gold" type="number" min="0" step="0.1" class="input-base" bind:value={$store.selectedElement.goldMultiplier} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="loot-bonus">Drop bonus</label>
          <input id="loot-bonus" type="number" step="0.01" class="input-base" bind:value={$store.selectedElement.dropBonus} />
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <div class="label-caps">Entries</div>
          <button class="btn btn-outline" type="button" on:click={addEntry}>Add entry</button>
        </div>
        {#if entries().length === 0}
          <p class="text-sm text-slate-500">No drops yet.</p>
        {/if}
        {#each entries() as entry, index}
          <div class="rounded-lg border border-slate-200 dark:border-slate-700 p-3 space-y-3">
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 space-y-1.5">
                <div class="label-caps">Item</div>
                <EntitySelectButton
                  value={entry.itemTemplateId}
                  elements={itemTemplates}
                  columns={itemTemplateColumns}
                  title="Select item template"
                  placeholder="Select an item..."
                  on:change={(e) => entry.itemTemplateId = e.detail}
                />
              </div>
              <button class="text-xs text-accent-red hover:underline" type="button" on:click={() => removeEntry(index)}>Remove</button>
            </div>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div class="space-y-1.5">
                <label class="label-caps" for={`loot-chance-${index}`}>Chance</label>
                <input id={`loot-chance-${index}`} type="number" min="0" max="1" step="0.01" class="input-base" bind:value={entry.dropChance} />
              </div>
              <div class="space-y-1.5">
                <label class="label-caps" for={`loot-min-${index}`}>Min qty</label>
                <input id={`loot-min-${index}`} type="number" min="1" class="input-base" bind:value={entry.minQuantity} />
              </div>
              <div class="space-y-1.5">
                <label class="label-caps" for={`loot-max-${index}`}>Max qty</label>
                <input id={`loot-max-${index}`} type="number" min="1" class="input-base" bind:value={entry.maxQuantity} />
              </div>
              <div class="space-y-2 pt-5">
                <label class="text-sm flex items-center gap-2" for={`loot-guaranteed-${index}`}>
                  <input id={`loot-guaranteed-${index}`} type="checkbox" bind:checked={entry.guaranteed} />
                  Guaranteed
                </label>
                <label class="text-sm flex items-center gap-2" for={`loot-boss-${index}`}>
                  <input id={`loot-boss-${index}`} type="checkbox" bind:checked={entry.bossOnly} />
                  Boss only
                </label>
              </div>
            </div>
          </div>
        {/each}
      </div>

      <div class="rounded-lg border border-slate-200 dark:border-slate-700 p-4 space-y-3" data-loot-roll>
        <div class="label-caps">Roll preview</div>
        {#if $store.selectedElement.isNew}
          <p class="text-sm text-slate-500">Save the table before rolling.</p>
        {:else}
          <div class="flex flex-wrap items-end gap-3">
            <div class="space-y-1.5">
              <label class="label-caps" for="loot-roll-n">Rolls</label>
              <input id="loot-roll-n" type="number" min="1" max="10000" class="input-base w-28" bind:value={rollN} />
            </div>
            <div class="space-y-1.5">
              <label class="label-caps" for="loot-roll-seed">Seed</label>
              <input id="loot-roll-seed" type="text" class="input-base w-36" placeholder="random" bind:value={rollSeed} />
            </div>
            <div class="space-y-1.5">
              <label class="label-caps" for="loot-roll-level">Player level</label>
              <input id="loot-roll-level" type="number" min="1" class="input-base w-24" bind:value={rollLevel} />
            </div>
            <label class="text-sm flex items-center gap-2 pb-2" for="loot-roll-boss">
              <input id="loot-roll-boss" type="checkbox" bind:checked={rollBoss} />
              Boss
            </label>
            <button class="btn btn-outline" type="button" disabled={rolling} on:click={rollPreview}>
              {rolling ? "Rolling…" : `Roll ×${rollN || 1000}`}
            </button>
          </div>
          {#if rollError}
            <p class="text-sm text-amber-300">{rollError}</p>
          {/if}
          {#if rollResult}
            <p class="text-xs text-slate-500">Seed {rollResult.seed}. {rollResult.n} rolls. Nothing was saved.</p>
            <table class="w-full text-sm">
              <thead>
                <tr class="text-left text-[10px] uppercase tracking-wide text-slate-500">
                  <th class="py-1 font-medium">Item</th>
                  <th class="py-1 font-medium">Frequency</th>
                  <th class="py-1 font-medium">Hits</th>
                  <th class="py-1 font-medium">Quantity</th>
                </tr>
              </thead>
              <tbody>
                {#each rollResult.drops || [] as drop}
                  <tr class="border-t border-slate-800">
                    <td class="py-1.5">
                      {drop.name || drop.itemTemplateId || "—"}
                      {#if drop.itemTemplateId}
                        <span class="font-mono text-[10px] text-slate-500">{drop.itemTemplateId}</span>
                      {/if}
                    </td>
                    <td class="py-1.5">{percent(drop.frequency)}</td>
                    <td class="py-1.5">{drop.hits}</td>
                    <td class="py-1.5">{drop.quantity}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          {/if}
        {/if}
      </div>
    {/if}
  </div>
</CRUDEditor>
