<script>
  import { writable } from "svelte/store";
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import { getAuth } from "../auth.js";
  import {
    getQuests,
    getQuest,
    createQuest,
    updateQuest,
    deleteQuest,
  } from "../api/quests.js";
  import { getNPCs } from "../api/npcs.js";
  import { getItemTemplates } from "../api/items.js";
  import { getRoomsValueHelp } from "../api/rooms.js";
  import { getScripts } from "../api/scripts.js";
  import { getDialogs } from "../api/dialogs.js";
  import {
    questColumns, npcColumns, itemTemplateColumns,
    roomColumns, scriptColumns, dialogColumns,
  } from "./tableColumns.js";
  import { uniqueValues } from "./fieldSuggestions.js";

  const { isAuthenticated, authToken } = getAuth();

  const store = createStore();
  const npcList = writable([]);
  const itemTemplateList = writable([]);
  const roomList = writable([]);
  const scriptList = writable([]);
  const dialogList = writable([]);
  const questList = writable([]);
  let loadedTypes = {};

  function loadEntityType(type) {
    if (loadedTypes[type]) return;
    if (!$isAuthenticated || !$authToken) return;
    loadedTypes[type] = true;

    if (type === "npcs") {
      getNPCs($authToken, [], (data) => npcList.set(data || []), () => {});
    } else if (type === "itemTemplates") {
      getItemTemplates($authToken, [], (data) => itemTemplateList.set(data || []), () => {});
    } else if (type === "rooms") {
      getRoomsValueHelp($authToken, (data) => roomList.set(data || []), () => {});
    } else if (type === "scripts") {
      getScripts($authToken, [], (data) => scriptList.set(data || []), () => {});
    } else if (type === "dialogs") {
      getDialogs($authToken, [], (data) => dialogList.set(data || []), () => {});
    } else if (type === "quests") {
      getQuests($authToken, [], (data) => questList.set(data || []), () => {});
    }
  }

  $: if ($isAuthenticated && $authToken) {
    loadEntityType("npcs");
    loadEntityType("itemTemplates");
    loadEntityType("rooms");
    loadEntityType("scripts");
    loadEntityType("dialogs");
    loadEntityType("quests");
  }

  // Populate area column filter options dynamically
  const columns = questColumns;
  $: {
    const areaCol = columns.find((c) => c.key === "area");
    if (areaCol && $store.elements.length) {
      areaCol.options = uniqueValues($store.elements, "area");
    }
  }
  $: areaSuggestions = uniqueValues($store.elements, "area");

  const objectiveTypes = [
    { value: "kill", label: "Kill" },
    { value: "collect", label: "Collect" },
    { value: "deliver", label: "Deliver" },
    { value: "visit", label: "Visit Room" },
    { value: "talk", label: "Talk to NPC" },
    { value: "custom", label: "Custom (Lua)" },
  ];

  const config = {
    title: "Manage Quests",
    entityType: "quest",
    subtitle: "Create and configure quests with objectives and rewards.",
    listTitle: "Quests",
    columns: questColumns,
    labels: {
      create: "Create Quest",
      update: "Update Quest",
      delete: "Delete",
    },
    get: getQuests,
    getElement: getQuest,
    create: createQuest,
    update: updateQuest,
    delete: deleteQuest,
  };

  config.new = (select) => {
    select({
      id: uuidv4(),
      name: "New Quest",
      description: "Quest description...",
      category: "side",
      area: "",
      level: 1,
      repeatable: false,
      source: { type: "npc", npcId: "" },
      objectives: [],
      rewards: { xp: 0, gold: 0, itemTemplateIds: [] },
      requiredQuestIds: [],
      requiredLevel: 0,
      acceptDialogText: "",
      progressDialogText: "",
      completeDialogText: "",
      isNew: true,
    });
  };

  function addObjective() {
    if (!$store.selectedElement) return;
    if (!$store.selectedElement.objectives) {
      $store.selectedElement.objectives = [];
    }
    $store.selectedElement.objectives = [
      ...$store.selectedElement.objectives,
      {
        id: "obj" + ($store.selectedElement.objectives.length + 1),
        type: "kill",
        description: "",
        targetId: "",
        targetName: "",
        amount: 1,
        deliverToNpcId: "",
        deliverToNpcName: "",
        dialogNodeId: "",
        checkScriptId: "",
        order: $store.selectedElement.objectives.length,
      },
    ];
  }

  function removeObjective(index) {
    $store.selectedElement.objectives = $store.selectedElement.objectives.filter(
      (_, i) => i !== index
    );
  }

  function addRewardItem() {
    if (!$store.selectedElement.rewards) {
      $store.selectedElement.rewards = { xp: 0, gold: 0, itemTemplateIds: [] };
    }
    if (!$store.selectedElement.rewards.itemTemplateIds) {
      $store.selectedElement.rewards.itemTemplateIds = [];
    }
    $store.selectedElement.rewards.itemTemplateIds = [
      ...$store.selectedElement.rewards.itemTemplateIds,
      "",
    ];
  }

  function removeRewardItem(index) {
    $store.selectedElement.rewards.itemTemplateIds =
      $store.selectedElement.rewards.itemTemplateIds.filter((_, i) => i !== index);
  }

  function addRequiredQuest() {
    if (!$store.selectedElement.requiredQuestIds) {
      $store.selectedElement.requiredQuestIds = [];
    }
    $store.selectedElement.requiredQuestIds = [
      ...$store.selectedElement.requiredQuestIds,
      "",
    ];
  }

  function removeRequiredQuest(index) {
    $store.selectedElement.requiredQuestIds =
      $store.selectedElement.requiredQuestIds.filter((_, i) => i !== index);
  }
</script>

<CRUDEditor {store} {config}>
  <div slot="content" class="space-y-6">
    <!-- General Settings -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="space-y-1.5">
        <label class="label-caps" for="quest-category">Category</label>
        <select
          id="quest-category"
          class="input-base"
          bind:value={$store.selectedElement.category}
        >
          <option value="">None</option>
          <option value="main">Main</option>
          <option value="side">Side</option>
          <option value="daily">Daily</option>
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="quest-area">Area</label>
        <input
          id="quest-area"
          type="text"
          class="input-base"
          list="quest-area-suggestions"
          bind:value={$store.selectedElement.area}
          placeholder="e.g. Gloomfen Marsh"
        />
        <datalist id="quest-area-suggestions">
          {#each areaSuggestions as val}
            <option value={val} />
          {/each}
        </datalist>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="quest-level">Level</label>
        <input
          id="quest-level"
          type="number"
          class="input-base"
          bind:value={$store.selectedElement.level}
          min="0"
        />
      </div>
      <div class="flex items-end gap-3 pb-1">
        <label class="flex items-center gap-2 text-sm cursor-pointer">
          <input
            type="checkbox"
            class="w-4 h-4 rounded border-slate-600 bg-slate-800 text-primary"
            bind:checked={$store.selectedElement.repeatable}
          />
          Repeatable
        </label>
      </div>
    </div>

    <!-- Source -->
    <div class="space-y-3">
      <h3
        class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
      >
        Quest Source
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="source-type">Source Type</label>
          <select
            id="source-type"
            class="input-base"
            bind:value={$store.selectedElement.source.type}
          >
            <option value="npc">NPC</option>
            <option value="item">Item</option>
            <option value="auto">Auto</option>
            <option value="script">Script</option>
          </select>
        </div>
        {#if $store.selectedElement.source.type === "npc"}
          <div class="space-y-1.5 md:col-span-2">
            <div class="label-caps">Source NPC</div>
            <EntitySelectButton
              value={$store.selectedElement.source.npcId}
              elements={$npcList}
              columns={npcColumns}
              title="Select Source NPC"
              placeholder="Select NPC..."
              on:change={(e) => $store.selectedElement.source.npcId = e.detail}
            />
          </div>
        {/if}
        {#if $store.selectedElement.source.type === "item"}
          <div class="space-y-1.5 md:col-span-2">
            <div class="label-caps">Source Item</div>
            <EntitySelectButton
              value={$store.selectedElement.source.itemId}
              elements={$itemTemplateList}
              columns={itemTemplateColumns}
              title="Select Source Item"
              placeholder="Select item template..."
              on:change={(e) => $store.selectedElement.source.itemId = e.detail}
            />
          </div>
        {/if}
      </div>
    </div>

    <!-- Objectives -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h3
          class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
        >
          Objectives ({$store.selectedElement.objectives?.length || 0})
        </h3>
        <button
          class="btn btn-outline text-xs"
          type="button"
          on:click={addObjective}
        >
          <span class="material-symbols-outlined text-sm">add</span>
          Add Objective
        </button>
      </div>

      {#if $store.selectedElement.objectives}
        {#each $store.selectedElement.objectives as obj, i}
          <div
            class="card p-4 space-y-3 border border-slate-700/50"
          >
            <div class="flex items-center justify-between">
              <span class="text-xs font-mono text-slate-500"
                >#{i + 1} - {obj.id}</span
              >
              <button
                class="btn btn-ghost text-xs text-red-400"
                type="button"
                on:click={() => removeObjective(i)}
              >
                <span class="material-symbols-outlined text-sm">delete</span>
              </button>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div class="space-y-1.5">
                <label class="label-caps" for="obj-id-{i}">ID</label>
                <input
                  id="obj-id-{i}"
                  type="text"
                  class="input-base font-mono text-xs"
                  bind:value={obj.id}
                />
              </div>
              <div class="space-y-1.5">
                <label class="label-caps" for="obj-type-{i}">Type</label>
                <select id="obj-type-{i}" class="input-base" bind:value={obj.type}>
                  {#each objectiveTypes as ot}
                    <option value={ot.value}>{ot.label}</option>
                  {/each}
                </select>
              </div>
              <div class="space-y-1.5 md:col-span-2">
                <label class="label-caps" for="obj-desc-{i}">Description</label>
                <input
                  id="obj-desc-{i}"
                  type="text"
                  class="input-base"
                  bind:value={obj.description}
                  placeholder="Kill 10 Marsh Rats"
                />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              {#if obj.type === "kill" || obj.type === "collect" || obj.type === "visit" || obj.type === "talk"}
                <div class="space-y-1.5">
                  <div class="label-caps">Target</div>
                  {#if obj.type === "visit"}
                    <EntitySelectButton
                      value={obj.targetId}
                      elements={$roomList}
                      columns={roomColumns}
                      title="Select Target Room"
                      placeholder="Select room..."
                      on:change={(e) => obj.targetId = e.detail}
                    />
                  {:else if obj.type === "kill"}
                    <EntitySelectButton
                      value={obj.targetId}
                      elements={$npcList}
                      columns={npcColumns}
                      title="Select Target NPC"
                      placeholder="Select NPC template..."
                      on:change={(e) => obj.targetId = e.detail}
                    />
                  {:else if obj.type === "collect"}
                    <EntitySelectButton
                      value={obj.targetId}
                      elements={$itemTemplateList}
                      columns={itemTemplateColumns}
                      title="Select Target Item"
                      placeholder="Select item template..."
                      on:change={(e) => obj.targetId = e.detail}
                    />
                  {:else if obj.type === "talk"}
                    <EntitySelectButton
                      value={obj.targetId}
                      elements={$npcList}
                      columns={npcColumns}
                      title="Select Target NPC"
                      placeholder="Select NPC..."
                      on:change={(e) => obj.targetId = e.detail}
                    />
                  {/if}
                </div>
                <div class="space-y-1.5">
                  <label class="label-caps" for="obj-targetname-{i}">Target Name</label>
                  <input
                    id="obj-targetname-{i}"
                    type="text"
                    class="input-base"
                    bind:value={obj.targetName}
                    placeholder="Display name"
                  />
                </div>
              {/if}
              {#if obj.type === "kill" || obj.type === "collect"}
                <div class="space-y-1.5">
                  <label class="label-caps" for="obj-amount-{i}">Amount</label>
                  <input
                    id="obj-amount-{i}"
                    type="number"
                    class="input-base"
                    bind:value={obj.amount}
                    min="1"
                  />
                </div>
              {/if}
            </div>

            {#if obj.type === "deliver"}
              <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                <div class="space-y-1.5">
                  <div class="label-caps">Item</div>
                  <EntitySelectButton
                    value={obj.targetId}
                    elements={$itemTemplateList}
                    columns={itemTemplateColumns}
                    title="Select Deliver Item"
                    placeholder="Select item template..."
                    on:change={(e) => obj.targetId = e.detail}
                  />
                </div>
                <div class="space-y-1.5">
                  <div class="label-caps">Deliver To NPC</div>
                  <EntitySelectButton
                    value={obj.deliverToNpcId}
                    elements={$npcList}
                    columns={npcColumns}
                    title="Select Deliver-To NPC"
                    placeholder="Select NPC..."
                    on:change={(e) => obj.deliverToNpcId = e.detail}
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="label-caps" for="obj-npcname-{i}">NPC Name</label>
                  <input
                    id="obj-npcname-{i}"
                    type="text"
                    class="input-base"
                    bind:value={obj.deliverToNpcName}
                  />
                </div>
              </div>
            {/if}

            {#if obj.type === "talk"}
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div class="space-y-1.5">
                  <div class="label-caps">Dialog Node (optional)</div>
                  <EntitySelectButton
                    value={obj.dialogNodeId}
                    elements={$dialogList}
                    columns={dialogColumns}
                    title="Select Dialog Node"
                    placeholder="Select dialog..."
                    on:change={(e) => obj.dialogNodeId = e.detail}
                  />
                </div>
              </div>
            {/if}

            {#if obj.type === "custom"}
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div class="space-y-1.5">
                  <div class="label-caps">Check Script</div>
                  <EntitySelectButton
                    value={obj.checkScriptId}
                    elements={$scriptList}
                    columns={scriptColumns}
                    title="Select Check Script"
                    placeholder="Select Lua script..."
                    on:change={(e) => obj.checkScriptId = e.detail}
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="label-caps" for="obj-custom-amount-{i}">Amount</label>
                  <input
                    id="obj-custom-amount-{i}"
                    type="number"
                    class="input-base"
                    bind:value={obj.amount}
                    min="1"
                  />
                </div>
              </div>
            {/if}
          </div>
        {/each}
      {/if}
    </div>

    <!-- Rewards -->
    <div class="space-y-3">
      <h3
        class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
      >
        Rewards
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="reward-xp">XP</label>
          <input
            id="reward-xp"
            type="number"
            class="input-base"
            bind:value={$store.selectedElement.rewards.xp}
            min="0"
          />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="reward-gold">Gold</label>
          <input
            id="reward-gold"
            type="number"
            class="input-base"
            bind:value={$store.selectedElement.rewards.gold}
            min="0"
          />
        </div>
      </div>

      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <div class="label-caps">Reward Items</div>
          <button
            class="btn btn-ghost text-xs"
            type="button"
            on:click={addRewardItem}
          >
            <span class="material-symbols-outlined text-sm">add</span>
            Add Item
          </button>
        </div>
        {#if $store.selectedElement.rewards?.itemTemplateIds}
          {#each $store.selectedElement.rewards.itemTemplateIds as itemId, i}
            <div class="flex items-center gap-2">
              <div class="flex-1">
                <EntitySelectButton
                  value={$store.selectedElement.rewards.itemTemplateIds[i]}
                  elements={$itemTemplateList}
                  columns={itemTemplateColumns}
                  title="Select Reward Item"
                  placeholder="Select item template..."
                  allowNone={false}
                  on:change={(e) => $store.selectedElement.rewards.itemTemplateIds[i] = e.detail}
                />
              </div>
              <button
                class="btn btn-ghost text-xs text-red-400"
                type="button"
                on:click={() => removeRewardItem(i)}
              >
                <span class="material-symbols-outlined text-sm">close</span>
              </button>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Prerequisites -->
    <div class="space-y-3">
      <h3
        class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
      >
        Prerequisites
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="req-level">Required Level</label>
          <input
            id="req-level"
            type="number"
            class="input-base"
            bind:value={$store.selectedElement.requiredLevel}
            min="0"
          />
        </div>
      </div>

      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <div class="label-caps">Required Quests</div>
          <button
            class="btn btn-ghost text-xs"
            type="button"
            on:click={addRequiredQuest}
          >
            <span class="material-symbols-outlined text-sm">add</span>
            Add Quest
          </button>
        </div>
        {#if $store.selectedElement.requiredQuestIds}
          {#each $store.selectedElement.requiredQuestIds as reqId, i}
            <div class="flex items-center gap-2">
              <div class="flex-1">
                <EntitySelectButton
                  value={$store.selectedElement.requiredQuestIds[i]}
                  elements={$questList}
                  columns={questColumns}
                  title="Select Required Quest"
                  placeholder="Select quest..."
                  allowNone={false}
                  on:change={(e) => $store.selectedElement.requiredQuestIds[i] = e.detail}
                />
              </div>
              <button
                class="btn btn-ghost text-xs text-red-400"
                type="button"
                on:click={() => removeRequiredQuest(i)}
              >
                <span class="material-symbols-outlined text-sm">close</span>
              </button>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Dialog Text -->
    <div class="space-y-3">
      <h3
        class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
      >
        Dialog Text
      </h3>
      <div class="space-y-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="accept-text">Accept Dialog Text</label>
          <textarea
            id="accept-text"
            rows="2"
            class="input-base"
            bind:value={$store.selectedElement.acceptDialogText}
            placeholder="NPC text when offering quest (leave empty for auto)"
          />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="progress-text"
            >Progress Dialog Text</label
          >
          <textarea
            id="progress-text"
            rows="2"
            class="input-base"
            bind:value={$store.selectedElement.progressDialogText}
            placeholder="NPC text while quest is in progress (leave empty for auto)"
          />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="complete-text"
            >Complete Dialog Text</label
          >
          <textarea
            id="complete-text"
            rows="2"
            class="input-base"
            bind:value={$store.selectedElement.completeDialogText}
            placeholder="NPC text on quest turn-in (leave empty for auto)"
          />
        </div>
      </div>
    </div>

    <!-- On Complete Script -->
    <div class="space-y-1.5">
      <div class="label-caps">On Complete Script</div>
      <EntitySelectButton
        value={$store.selectedElement.onCompleteScriptId}
        elements={$scriptList}
        columns={scriptColumns}
        title="Select On-Complete Script"
        placeholder="None"
        on:change={(e) => $store.selectedElement.onCompleteScriptId = e.detail}
      />
    </div>
  </div>
</CRUDEditor>
