<script>
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { getAuth } from "../auth.js";
  import {
    getSkills,
    getSkill,
    createSkill,
    updateSkill,
    deleteSkill,
  } from "../api/skills.js";
  import { skillColumns } from "./tableColumns.js";

  const { isAuthenticated, authToken } = getAuth();
  const store = createStore();

  const classOptions = ["warrior", "rogue", "mage", "cleric", "ranger", "druid"];
  const resourceTypes = [
    { value: "mana", label: "Mana" },
    { value: "cooldown", label: "Cooldown" },
  ];
  const targetTypes = [
    { value: "enemy", label: "Enemy" },
    { value: "self", label: "Self" },
    { value: "all_enemies", label: "All Enemies" },
  ];
  const effectTypes = [
    { value: "damage", label: "Damage" },
    { value: "heal", label: "Heal" },
    { value: "buff", label: "Buff" },
    { value: "debuff", label: "Debuff" },
    { value: "dot", label: "DoT" },
    { value: "hot", label: "HoT" },
  ];
  const scalingAttrs = ["STR", "DEX", "INT", "WIS"];

  const config = {
    title: "Manage Skills",
    subtitle: "Create and configure combat skills and spells.",
    listTitle: "Skills",
    columns: skillColumns,
    labels: {
      create: "Create Skill",
      update: "Update Skill",
      delete: "Delete",
    },
    get: getSkills,
    getElement: getSkill,
    create: createSkill,
    update: updateSkill,
    delete: deleteSkill,
  };

  config.new = (select) => {
    select({
      id: uuidv4(),
      name: "New Skill",
      description: "Skill description...",
      classIds: [],
      levelRequired: 1,
      resourceType: "cooldown",
      manaCost: 0,
      cooldownRounds: 2,
      target: "enemy",
      effect: "damage",
      scalingAttr: "STR",
      basePower: 10,
      scalingFactor: 1.0,
      duration: 0,
      buffStat: "",
      buffPercent: 0,
      ignoresDefense: false,
      hitCount: 1,
      secondaryEffect: "",
      secondaryBasePower: 0,
      secondaryScaling: 0,
      secondaryTarget: "",
      isNew: true,
    });
  };

  function toggleClass(classId) {
    if (!$store.selectedElement) return;
    if (!$store.selectedElement.classIds) {
      $store.selectedElement.classIds = [];
    }
    const idx = $store.selectedElement.classIds.indexOf(classId);
    if (idx >= 0) {
      $store.selectedElement.classIds = $store.selectedElement.classIds.filter(c => c !== classId);
    } else {
      $store.selectedElement.classIds = [...$store.selectedElement.classIds, classId];
    }
  }

  $: el = $store.selectedElement;
  $: showMana = el?.resourceType === "mana";
  $: showCooldown = el?.resourceType === "cooldown";
  $: showBuffFields = el?.effect === "buff" || el?.effect === "debuff";
  $: showDuration = el?.effect === "buff" || el?.effect === "debuff" || el?.effect === "dot" || el?.effect === "hot";
</script>

<CRUDEditor {config} {store}>
  {#if $store.selectedElement}
    <div class="space-y-6">

      <!-- Basic Info -->
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Name</label>
          <input type="text" bind:value={$store.selectedElement.name}
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Level Required</label>
          <input type="number" bind:value={$store.selectedElement.levelRequired} min="1"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
        </div>
      </div>

      <div>
        <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Description</label>
        <textarea bind:value={$store.selectedElement.description} rows="2"
          class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm"></textarea>
      </div>

      <!-- Classes (multi-select checkboxes) -->
      <div>
        <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-2">Classes</label>
        <div class="flex flex-wrap gap-3">
          {#each classOptions as cls}
            <label class="inline-flex items-center gap-1.5 text-sm cursor-pointer">
              <input type="checkbox"
                checked={$store.selectedElement.classIds?.includes(cls)}
                on:change={() => toggleClass(cls)}
                class="rounded border-slate-300 dark:border-slate-600" />
              <span class="capitalize">{cls}</span>
            </label>
          {/each}
        </div>
      </div>

      <!-- Resource Type -->
      <div class="grid grid-cols-3 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Resource Type</label>
          <select bind:value={$store.selectedElement.resourceType}
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm">
            {#each resourceTypes as rt}
              <option value={rt.value}>{rt.label}</option>
            {/each}
          </select>
        </div>
        {#if showMana}
          <div>
            <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Mana Cost</label>
            <input type="number" bind:value={$store.selectedElement.manaCost} min="0"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
          </div>
        {/if}
        {#if showCooldown}
          <div>
            <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Cooldown (rounds)</label>
            <input type="number" bind:value={$store.selectedElement.cooldownRounds} min="0"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
          </div>
        {/if}
      </div>

      <!-- Targeting & Effect -->
      <div class="grid grid-cols-3 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Target</label>
          <select bind:value={$store.selectedElement.target}
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm">
            {#each targetTypes as t}
              <option value={t.value}>{t.label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Effect</label>
          <select bind:value={$store.selectedElement.effect}
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm">
            {#each effectTypes as e}
              <option value={e.value}>{e.label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Scaling Attribute</label>
          <select bind:value={$store.selectedElement.scalingAttr}
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm">
            {#each scalingAttrs as attr}
              <option value={attr}>{attr}</option>
            {/each}
          </select>
        </div>
      </div>

      <!-- Power & Scaling -->
      <div class="grid grid-cols-3 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Base Power</label>
          <input type="number" bind:value={$store.selectedElement.basePower} min="0"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Scaling Factor</label>
          <input type="number" bind:value={$store.selectedElement.scalingFactor} min="0" step="0.1"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Hit Count</label>
          <input type="number" bind:value={$store.selectedElement.hitCount} min="1"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
        </div>
      </div>

      <!-- Duration & Buff Fields -->
      {#if showDuration}
        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Duration (rounds)</label>
            <input type="number" bind:value={$store.selectedElement.duration} min="0"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
          </div>
          {#if showBuffFields}
            <div>
              <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Buff Stat</label>
              <select bind:value={$store.selectedElement.buffStat}
                class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm">
                <option value="">None</option>
                <option value="ATK">ATK</option>
                <option value="DEF">DEF</option>
                <option value="STR">STR</option>
                <option value="DEX">DEX</option>
                <option value="INT">INT</option>
                <option value="WIS">WIS</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Buff %</label>
              <input type="number" bind:value={$store.selectedElement.buffPercent} step="0.05"
                class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
            </div>
          {/if}
        </div>
      {/if}

      <!-- Special Flags -->
      <div class="flex items-center gap-6">
        <label class="inline-flex items-center gap-1.5 text-sm cursor-pointer">
          <input type="checkbox" bind:checked={$store.selectedElement.ignoresDefense}
            class="rounded border-slate-300 dark:border-slate-600" />
          <span>Ignores Defense</span>
        </label>
      </div>

      <!-- Secondary Effect (collapsible) -->
      <details class="border border-slate-200 dark:border-slate-700 rounded-lg p-4">
        <summary class="text-xs font-semibold text-slate-500 dark:text-slate-400 cursor-pointer">Secondary Effect</summary>
        <div class="grid grid-cols-2 gap-4 mt-3">
          <div>
            <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Effect</label>
            <select bind:value={$store.selectedElement.secondaryEffect}
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm">
              <option value="">None</option>
              {#each effectTypes as e}
                <option value={e.value}>{e.label}</option>
              {/each}
            </select>
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Target</label>
            <select bind:value={$store.selectedElement.secondaryTarget}
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm">
              <option value="">Same as primary</option>
              {#each targetTypes as t}
                <option value={t.value}>{t.label}</option>
              {/each}
            </select>
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Base Power</label>
            <input type="number" bind:value={$store.selectedElement.secondaryBasePower} min="0"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Scaling Factor</label>
            <input type="number" bind:value={$store.selectedElement.secondaryScaling} min="0" step="0.1"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-sm" />
          </div>
        </div>
      </details>

    </div>
  {/if}
</CRUDEditor>
