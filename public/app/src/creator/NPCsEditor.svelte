<script>
  import { writable } from "svelte/store";
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import MerchantPreviewModal from "./MerchantPreviewModal.svelte";
  import EnemyTraitPanel from "./EnemyTraitPanel.svelte";
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import NPCInspector from "./NPCInspector.svelte";
  import { getAuth } from "../auth.js";

  import {
    getNPC,
    getNPCs,
    createNPC,
    updateNPC,
    deleteNPC,
  } from "../api/npcs.js";
  import { getDialogs } from "../api/dialogs.js";
  import { getRoomsValueHelp } from "../api/rooms.js";
  import { getScripts } from "../api/scripts.js";
  import { getItemTemplates } from "../api/items.js";
  import { getLootTables } from "../api/loottables.js";
  import { getEnemyScaling } from "../api/balance.js";
  import { npcColumns, dialogColumns, roomColumns } from "./tableColumns.js";
  import { knownRaces, knownClasses } from "./fieldSuggestions.js";
  import { previewMerchant } from "../api/previews.js";

  // Clone columns so we can populate dynamic dropdown options
  const columns = npcColumns.map((c) => ({ ...c }));
  const raceCol = columns.find((c) => c.key === "race.name");
  if (raceCol) raceCol.options = knownRaces;
  const classCol = columns.find((c) => c.key === "class.name");
  if (classCol) classCol.options = knownClasses;

  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
  };

  const dialogsValueHelp = writable([]);
  const roomsValueHelp = writable([]);
  const store = createStore();
  let hasLoadedDialogs = false;
  let hasLoadedRooms = false;
  let hasLoadedEnemyRefs = false;
  let scripts = [];
  let lootTables = [];
  let itemTemplates = [];
  let enemyScaling = null;
  let showMerchantPreview = false;
  let merchantPreview = null;

  let levels = [];
  for (let i = 1; i <= 50; i += 1) levels.push(i);

  const races = [
    { id: "human", name: "Human", description: "The common race" },
    { id: "elf", name: "Elf", description: "Splendid forestwalkers" },
    { id: "construct", name: "Construct", description: "Built, not born" },
    { id: "dwarf", name: "Dwarf", description: "Small, but fierce" },
    { id: "halfling", name: "Halfling", description: "Small and nimble" },
    { id: "orc", name: "Orc", description: "Strong and fierce" },
    { id: "goblin", name: "Goblin", description: "Cunning and sneaky" },
    { id: "undead", name: "Undead", description: "Risen from death" },
    { id: "demon", name: "Demon", description: "Infernal beings" },
    { id: "beast", name: "Beast", description: "Wild creatures" },
  ];

  const classes = [
    { id: "warrior", name: "Warrior", description: "Melee warrior" },
    { id: "ranger", name: "Ranger", description: "Ranged combatant" },
    { id: "wizard", name: "Wizard", description: "Master of elements" },
    { id: "rogue", name: "Rogue", description: "Stealthy and deadly" },
    { id: "cleric", name: "Cleric", description: "Divine healer" },
    { id: "merchant", name: "Merchant", description: "Trader of goods" },
    { id: "guard", name: "Guard", description: "Protector of the realm" },
    { id: "commoner", name: "Commoner", description: "Simple folk" },
  ];

  const getRaceById = (id) => races.find((r) => r.id === id) || races[0];
  const getClassById = (id) => classes.find((c) => c.id === id) || classes[0];

  let selectedRaceId = "human";
  let selectedClassId = "commoner";

  // Tab state for traits section
  let activeTraitTab = "behavior";
  let detailView = "edit";
  let detailViewFor = "";
  let inspectorRequested = null;

  $: {
    const id = $store.selectedElement?.id || "";
    if (id !== detailViewFor) {
      detailViewFor = id;
      if (inspectorRequested === null && typeof window !== "undefined") {
        inspectorRequested = new URLSearchParams(window.location.search).get("view") === "inspector";
      }
      detailView = inspectorRequested ? "inspector" : "edit";
      inspectorRequested = false;
    }
  }

  const config = {
    title: "Manage NPCs",
    entityType: "npc",
    subtitle: "Configure NPC profiles, traits, and dialog bindings.",
    listTitle: "NPCs",
    columns: columns,
    labels: {
      create: "Create NPC",
      update: "Update NPC",
      delete: "Delete",
    },
    get: getNPCs,
    getElement: getNPC,
    create: createNPC,
    update: updateNPC,
    delete: deleteNPC,
    beforeSelect: (element) => {
      if (!element.race) element.race = getRaceById("human");
      if (!element.class) element.class = getClassById("commoner");
      ensureBehaviorDefaults(element);
      selectedRaceId = element.race.id || "human";
      selectedClassId = element.class.id || "commoner";
    },
    new: (select) => {
      selectedRaceId = "human";
      selectedClassId = "commoner";
      select({
        id: uuidv4(),
        name: "New NPC",
        description: "",
        race: getRaceById("human"),
        class: getClassById("commoner"),
        level: 1,
        currentHitPoints: 10,
        maxHitPoints: 10,
        dialogID: "",
        idleDialogID: "",
        idleDialogTimeout: 0,
        currentRoomID: "",
        spawnRoomID: "",
        state: "idle",
        wanderRadius: 0,
        patrolPath: [],
        enemyTrait: null,
        merchantTrait: null,
        isTemplate: true,  // Default to template for spawning
        isNew: true,
      });
    },
    badge: (element) => (element.race ? element.race.name : ""),
    icon: (element) => {
      if (!element.isTemplate) {
        return { name: "star", color: "#f59e0b", title: "Unique NPC" };
      }
      return null;
    },
  };

  const runMerchantPreview = () => {
    if (!$isAuthenticated || !$authToken || !$store.selectedElement) return;
    previewMerchant(
      $authToken,
      $store.selectedElement,
      (preview) => {
        merchantPreview = preview;
        showMerchantPreview = true;
      },
      (err) => {
        console.error("Failed to preview merchant:", err);
        alert("Failed to preview merchant. Please try again.");
      }
    );
  };

  config.extraActions = [
    {
      label: "Preview Merchant",
      icon: "storefront",
      variant: "btn-outline",
      onClick: runMerchantPreview,
    },
  ];

  const toggleEnemyTrait = () => {
    store.update((state) => {
      state.selectedElement.enemyTrait = state.selectedElement.enemyTrait
        ? null
        : {
            creatureType: "beast",
            combatStyle: "melee",
            difficulty: "normal",
            attackPower: 5,
            defense: 2,
            attackSpeed: 0,
            aggroRadius: 0,
            aggroOnSight: false,
            callForHelp: false,
            fleeThreshold: 0,
            xpReward: 10,
            goldDrop: { min: 0, max: 0 },
            lootTableId: "",
            guaranteedLoot: [],
            maxDrops: 0,
            onAggroScript: "",
            onDeathScript: "",
            onFleeScript: "",
            onLowHealthScript: "",
            lowHealthThreshold: 0,
          };
      return state;
    });
  };

  const toggleMerchantTrait = () => {
    store.update((state) => {
      state.selectedElement.merchantTrait = state.selectedElement.merchantTrait
        ? null
        : {
            buyPriceModifier: 1.0,
            sellPriceModifier: 0.5,
            inventory: [],
          };
      return state;
    });
  };

  const onRaceChange = () => {
    store.update((state) => {
      state.selectedElement.race = getRaceById(selectedRaceId);
      return state;
    });
  };

  const onClassChange = () => {
    store.update((state) => {
      state.selectedElement.class = getClassById(selectedClassId);
      return state;
    });
  };

  const ensureBehaviorDefaults = (element) => {
    if (!element) return;
    if (!element.state) element.state = "idle";
    if (!Array.isArray(element.patrolPath)) element.patrolPath = [];
    if (element.wanderRadius === undefined || element.wanderRadius === null) element.wanderRadius = 0;
    if (element.idleDialogTimeout === undefined || element.idleDialogTimeout === null) element.idleDialogTimeout = 0;
  };

  const durationToSeconds = (duration) => Math.floor((Number(duration) || 0) / 1000000000);
  const secondsToDuration = (seconds) => Math.max(0, Number(seconds) || 0) * 1000000000;

  const setIdleDialogTimeoutSeconds = (seconds) => {
    store.update((state) => {
      state.selectedElement.idleDialogTimeout = secondsToDuration(seconds);
      return state;
    });
  };

  const addPatrolStop = () => {
    store.update((state) => {
      ensureBehaviorDefaults(state.selectedElement);
      state.selectedElement.patrolPath = [...state.selectedElement.patrolPath, ""];
      if (state.selectedElement.state === "idle") {
        state.selectedElement.state = "patrol";
      }
      return state;
    });
  };

  const updatePatrolStop = (index, roomID) => {
    store.update((state) => {
      ensureBehaviorDefaults(state.selectedElement);
      state.selectedElement.patrolPath[index] = roomID;
      return state;
    });
  };

  const removePatrolStop = (index) => {
    store.update((state) => {
      ensureBehaviorDefaults(state.selectedElement);
      state.selectedElement.patrolPath = state.selectedElement.patrolPath.filter((_, i) => i !== index);
      if (state.selectedElement.patrolPath.length === 0 && state.selectedElement.state === "patrol") {
        state.selectedElement.state = "idle";
      }
      return state;
    });
  };

  const loadDialogs = () => {
    if (hasLoadedDialogs) return;
    if (!$isAuthenticated || !$authToken) return;

    getDialogs(
      $authToken,
      [],
      (dialogs) => {
        dialogsValueHelp.set(dialogs || []);
        hasLoadedDialogs = true;
      },
      (err) => {
        console.log("Failed to load dialogs for NPCs editor:", err);
      }
    );
  };

  const loadRooms = () => {
    if (hasLoadedRooms) return;
    if (!$isAuthenticated || !$authToken) return;

    getRoomsValueHelp(
      $authToken,
      (rooms) => {
        roomsValueHelp.set(rooms || []);
        hasLoadedRooms = true;
      },
      (err) => {
        console.log("Failed to load rooms for NPCs editor:", err);
      }
    );
  };

  const loadEnemyRefs = () => {
    if (hasLoadedEnemyRefs) return;
    if (!$isAuthenticated || !$authToken) return;
    hasLoadedEnemyRefs = true;
    getScripts(
      $authToken,
      [],
      (all) => {
        scripts = all || [];
      },
      (err) => console.log("Failed to load scripts for enemy editor:", err)
    );
    getLootTables(
      $authToken,
      [],
      (all) => {
        lootTables = all || [];
      },
      (err) => console.log("Failed to load loot tables for enemy editor:", err)
    );
    getItemTemplates(
      $authToken,
      [],
      (all) => {
        itemTemplates = all || [];
      },
      (err) => console.log("Failed to load item templates for enemy editor:", err)
    );
    getEnemyScaling(
      $authToken,
      (data) => {
        enemyScaling = data;
      },
      (err) => console.log("Failed to load enemy scaling:", err)
    );
  };

  // Load dialogs and rooms once auth token becomes available
  $: if ($isAuthenticated && $authToken && !hasLoadedDialogs) {
    loadDialogs();
  }
  $: if ($isAuthenticated && $authToken && !hasLoadedRooms) {
    loadRooms();
  }
  $: if ($isAuthenticated && $authToken && !hasLoadedEnemyRefs) {
    loadEnemyRefs();
  }
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
    <div class="flex items-center gap-1 border-b border-slate-200 dark:border-slate-700">
      <button type="button" class="tab-btn" class:active={detailView === "edit"} on:click={() => detailView = "edit"}>
        Edit
      </button>
      <button type="button" class="tab-btn" class:active={detailView === "inspector"} on:click={() => detailView = "inspector"}>
        Inspector
      </button>
    </div>

    {#if detailView === "inspector"}
      <NPCInspector npcId={$store.selectedElement.id} isNew={!!$store.selectedElement.isNew} />
    {:else}
    <!-- NPC Type Configuration -->
    <div class="p-4 rounded-lg bg-slate-800/50 border border-slate-700/50 space-y-3">
      <h3 class="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-2">
        <span class="material-symbols-outlined text-base">category</span>
        NPC Type
      </h3>
      <div class="flex flex-wrap items-center gap-4">
        <label class="flex items-center gap-2 cursor-pointer p-2 rounded-lg border transition-all {$store.selectedElement.isTemplate ? 'border-primary bg-primary/10' : 'border-slate-700 hover:border-slate-600'}">
          <input
            type="radio"
            name="npcType"
            class="text-primary"
            checked={$store.selectedElement.isTemplate}
            on:change={() => $store.selectedElement.isTemplate = true}
          />
          <div>
            <span class="text-sm font-medium">Template</span>
            <p class="text-[10px] text-slate-500">Blueprint for spawning multiple instances</p>
          </div>
        </label>
        <label class="flex items-center gap-2 cursor-pointer p-2 rounded-lg border transition-all {!$store.selectedElement.isTemplate ? 'border-amber-500 bg-amber-500/10' : 'border-slate-700 hover:border-slate-600'}">
          <input
            type="radio"
            name="npcType"
            class="text-amber-500"
            checked={!$store.selectedElement.isTemplate}
            on:change={() => $store.selectedElement.isTemplate = false}
          />
          <div>
            <span class="text-sm font-medium">Unique</span>
            <p class="text-[10px] text-slate-500">Single NPC, cannot be spawned</p>
          </div>
        </label>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="npc-level">Level</label>
        <select id="npc-level" class="input-base" bind:value={$store.selectedElement.level}>
          {#each levels as lvl}
            <option value={lvl}>{lvl}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="npc-race">Race</label>
        <select id="npc-race" class="input-base" bind:value={selectedRaceId} on:change={onRaceChange}>
          {#each races as race}
            <option value={race.id}>{race.name}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="npc-class">Class</label>
        <select id="npc-class" class="input-base" bind:value={selectedClassId} on:change={onClassChange}>
          {#each classes as cls}
            <option value={cls.id}>{cls.name}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <div class="label-caps">Dialog</div>
        <EntitySelectButton
          value={$store.selectedElement.dialogID}
          elements={$dialogsValueHelp || []}
          columns={dialogColumns}
          title="Select Dialog"
          placeholder="None"
          on:change={(e) => $store.selectedElement.dialogID = e.detail}
        />
      </div>
      <div class="space-y-1.5">
        <div class="label-caps">Idle Dialog</div>
        <EntitySelectButton
          value={$store.selectedElement.idleDialogID}
          elements={$dialogsValueHelp || []}
          columns={dialogColumns}
          title="Select Idle Dialog"
          placeholder="None"
          on:change={(e) => $store.selectedElement.idleDialogID = e.detail}
        />
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <div class="label-caps">Current Room</div>
        <EntitySelectButton
          value={$store.selectedElement.currentRoomID}
          elements={$roomsValueHelp || []}
          columns={roomColumns}
          title="Select Current Room"
          placeholder="None"
          on:change={(e) => $store.selectedElement.currentRoomID = e.detail}
        />
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="npc-hp-current">Hit Points</label>
        <div class="grid grid-cols-2 gap-2">
          <input id="npc-hp-current" class="input-base text-center" bind:value={$store.selectedElement.currentHitPoints} type="number" />
          <input
            id="npc-hp-max"
            class="input-base text-center"
            bind:value={$store.selectedElement.maxHitPoints}
            type="number"
            readonly={!!$store.selectedElement.enemyTrait?.baseStats}
            title={$store.selectedElement.enemyTrait?.baseStats ? "Effective max after scaling. Edit the content base on the Enemy tab." : "Max hit points"}
          />
        </div>
        {#if $store.selectedElement.enemyTrait?.baseStats}
          <p class="text-[10px] text-slate-500">Current HP, then effective max. Edit the content base on the Enemy tab.</p>
        {/if}
      </div>
    </div>

    {/if}
  </div>

  <div slot="extensions" class="space-y-4">
    {#if detailView !== "inspector"}
    <!-- Tabbed Navigation for Traits -->
    <div class="flex items-center gap-1 border-b border-slate-200 dark:border-slate-700">
      <button
        type="button"
        class="tab-btn"
        class:active={activeTraitTab === "behavior"}
        on:click={() => activeTraitTab = "behavior"}
      >
        <span class="material-symbols-outlined text-base">route</span>
        Behavior
        {#if ($store.selectedElement.wanderRadius || 0) > 0 || ($store.selectedElement.patrolPath || []).length > 0 || $store.selectedElement.idleDialogID}
          <span class="tab-badge active-badge">ON</span>
        {/if}
      </button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTraitTab === "enemy"}
        on:click={() => activeTraitTab = "enemy"}
      >
        <span class="material-symbols-outlined text-base">swords</span>
        Enemy Trait
        {#if $store.selectedElement.enemyTrait}
          <span class="tab-badge active-badge">ON</span>
        {/if}
      </button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTraitTab === "merchant"}
        on:click={() => activeTraitTab = "merchant"}
      >
        <span class="material-symbols-outlined text-base">storefront</span>
        Merchant Trait
        {#if $store.selectedElement.merchantTrait}
          <span class="tab-badge active-badge">ON</span>
        {/if}
      </button>
    </div>

    <!-- Tab Content -->
    <div class="tab-content">
      {#if activeTraitTab === "behavior"}
        <div class="card p-6 space-y-5">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">route</span>
              Movement and Idle Behavior
            </h3>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="space-y-1.5">
              <label class="label-caps" for="npc-state">State</label>
              <select id="npc-state" class="input-base text-xs" bind:value={$store.selectedElement.state}>
                <option value="idle">Idle</option>
                <option value="patrol">Patrol</option>
                <option value="combat">Combat</option>
                <option value="fleeing">Fleeing</option>
              </select>
            </div>
            <div class="space-y-1.5">
              <label class="label-caps" for="npc-wander-radius">Wander Radius</label>
              <input
                id="npc-wander-radius"
                class="input-base text-xs text-center"
                type="number"
                min="0"
                step="1"
                bind:value={$store.selectedElement.wanderRadius}
              />
            </div>
            <div class="space-y-1.5">
              <label class="label-caps" for="npc-idle-timeout">Idle Chatter Seconds</label>
              <input
                id="npc-idle-timeout"
                class="input-base text-xs text-center"
                type="number"
                min="0"
                step="5"
                value={durationToSeconds($store.selectedElement.idleDialogTimeout)}
                on:input={(e) => setIdleDialogTimeoutSeconds(e.currentTarget.value)}
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <div class="label-caps">Spawn Room</div>
              <EntitySelectButton
                value={$store.selectedElement.spawnRoomID}
                elements={$roomsValueHelp || []}
                columns={roomColumns}
                title="Select Spawn Room"
                placeholder="None"
                on:change={(e) => $store.selectedElement.spawnRoomID = e.detail}
              />
            </div>
            <div class="space-y-1.5">
              <div class="label-caps">Idle Chatter Dialog</div>
              <EntitySelectButton
                value={$store.selectedElement.idleDialogID}
                elements={$dialogsValueHelp || []}
                columns={dialogColumns}
                title="Select Idle Dialog"
                placeholder="None"
                on:change={(e) => $store.selectedElement.idleDialogID = e.detail}
              />
            </div>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h4 class="label-caps">Patrol Path</h4>
              <button class="text-xs text-primary hover:underline" type="button" on:click={addPatrolStop}>
                + Add Room
              </button>
            </div>

            {#if ($store.selectedElement.patrolPath || []).length > 0}
              <div class="space-y-2">
                {#each $store.selectedElement.patrolPath as roomID, index}
                  <div class="grid grid-cols-[32px_1fr_auto] gap-2 items-center">
                    <span class="text-xs font-bold text-slate-500 text-center">{index + 1}</span>
                    <EntitySelectButton
                      value={roomID}
                      elements={$roomsValueHelp || []}
                      columns={roomColumns}
                      title="Select Patrol Room"
                      placeholder="Select room..."
                      on:change={(e) => updatePatrolStop(index, e.detail)}
                    />
                    <button
                      type="button"
                      class="text-xs text-red-400 hover:text-red-300 px-2"
                      on:click={() => removePatrolStop(index)}
                    >
                      Remove
                    </button>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="p-4 rounded-lg bg-slate-800/30 border border-slate-700/50 text-center">
                <span class="material-symbols-outlined text-3xl text-slate-600 mb-2">route</span>
                <p class="text-xs text-slate-500 dark:text-slate-400">
                  No patrol route configured. Add rooms to make this NPC follow a loop.
                </p>
              </div>
            {/if}
          </div>
        </div>
      {:else if activeTraitTab === "enemy"}
        <div class="card p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">swords</span>
              Enemy Trait
            </h3>
            <button class="text-xs text-primary hover:underline" type="button" on:click={toggleEnemyTrait}>
              {#if $store.selectedElement.enemyTrait}Remove Trait{:else}+ Enable Trait{/if}
            </button>
          </div>

          {#if $store.selectedElement.enemyTrait}
            <EnemyTraitPanel
              npc={$store.selectedElement}
              {store}
              {scripts}
              {lootTables}
              {itemTemplates}
              scaling={enemyScaling}
            />
          {:else}
            <div class="p-4 rounded-lg bg-slate-800/30 border border-slate-700/50 text-center">
              <span class="material-symbols-outlined text-3xl text-slate-600 mb-2">swords</span>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                No enemy trait configured. Enable this trait to make the NPC hostile and define combat stats.
              </p>
            </div>
          {/if}
        </div>
      {:else if activeTraitTab === "merchant"}
        <div class="card p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">storefront</span>
              Merchant Trait
            </h3>
            <button class="text-xs text-primary hover:underline" type="button" on:click={toggleMerchantTrait}>
              {#if $store.selectedElement.merchantTrait}Remove Trait{:else}+ Enable Trait{/if}
            </button>
          </div>

          {#if $store.selectedElement.merchantTrait}
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Configure trading behavior for this NPC when players buy or sell items.
            </p>
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-1.5">
                <label class="label-caps" for="merchant-buy">Buy Price Modifier</label>
                <input id="merchant-buy" class="input-base text-xs text-center" type="number" step="0.1" min="0" bind:value={$store.selectedElement.merchantTrait.buyPriceModifier} />
                <p class="text-[9px] text-slate-500">Multiplier when NPC buys from player (0.5 = 50% of base price)</p>
              </div>
              <div class="space-y-1.5">
                <label class="label-caps" for="merchant-sell">Sell Price Modifier</label>
                <input id="merchant-sell" class="input-base text-xs text-center" type="number" step="0.1" min="0" bind:value={$store.selectedElement.merchantTrait.sellPriceModifier} />
                <p class="text-[9px] text-slate-500">Multiplier when NPC sells to player (1.0 = base price)</p>
              </div>
            </div>

            <div class="p-3 rounded-lg bg-slate-800/30 border border-slate-700/50">
              <p class="text-[10px] text-slate-500 uppercase font-bold mb-2">Inventory Management</p>
              <p class="text-xs text-slate-400">
                Merchant inventory is managed separately. Configure items for sale through the item system.
              </p>
            </div>
          {:else}
            <div class="p-4 rounded-lg bg-slate-800/30 border border-slate-700/50 text-center">
              <span class="material-symbols-outlined text-3xl text-slate-600 mb-2">storefront</span>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                No merchant trait configured. Enable this trait to allow the NPC to buy and sell items.
              </p>
            </div>
          {/if}
        </div>
      {/if}
    </div>
    {/if}
  </div>
</CRUDEditor>

<MerchantPreviewModal
  open={showMerchantPreview}
  preview={merchantPreview}
  on:close={() => showMerchantPreview = false}
/>

<style>
  /* Tab Styles */
  .tab-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 16px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #64748b;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-bottom: -1px;
  }

  .tab-btn:hover {
    color: #94a3b8;
  }

  .tab-btn.active {
    color: #00bcd4;
    border-bottom-color: #00bcd4;
  }

  .tab-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    font-size: 10px;
    font-weight: 700;
    background: rgba(100, 116, 139, 0.3);
    border-radius: 9px;
  }

  .tab-btn.active .tab-badge {
    background: rgba(0, 188, 212, 0.2);
    color: #00bcd4;
  }

  .tab-badge.active-badge {
    background: rgba(34, 197, 94, 0.2);
    color: #22c55e;
  }

  .tab-content {
    min-height: 200px;
  }
</style>
