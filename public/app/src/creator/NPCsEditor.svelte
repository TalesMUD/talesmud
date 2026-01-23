<script>
  import { writable } from "svelte/store";
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { getAuth } from "../auth.js";

  import {
    getNPC,
    getNPCs,
    createNPC,
    updateNPC,
    deleteNPC,
  } from "../api/npcs.js";
  import { getDialogs } from "../api/dialogs.js";

  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  const dialogsValueHelp = writable([]);
  const store = createStore();
  let hasLoadedDialogs = false;

  let levels = [];
  for (let i = 1; i <= 50; i += 1) levels.push(i);

  const races = [
    { id: "human", name: "Human", description: "The common race" },
    { id: "elve", name: "Elve", description: "Splendid forestwalkers" },
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

  // CreatureType for enemies - what the creature fundamentally IS
  const creatureTypes = [
    { id: "beast", name: "Beast", description: "Animals, insects, natural creatures" },
    { id: "humanoid", name: "Humanoid", description: "Goblins, orcs, bandits - uses Race/Class" },
    { id: "undead", name: "Undead", description: "Skeletons, zombies, ghosts" },
    { id: "elemental", name: "Elemental", description: "Fire, water, earth, air beings" },
    { id: "construct", name: "Construct", description: "Golems, animated objects" },
    { id: "demon", name: "Demon", description: "Demons, devils, otherworldly beings" },
    { id: "dragon", name: "Dragon", description: "Dragons and dragonkin" },
    { id: "aberration", name: "Aberration", description: "Unnatural, eldritch creatures" },
  ];

  // CombatStyle for enemies - HOW the creature fights
  const combatStyles = [
    { id: "melee", name: "Melee", description: "Close-range physical attacks" },
    { id: "ranged", name: "Ranged", description: "Bows, thrown weapons, spitting" },
    { id: "magic", name: "Magic", description: "Spells and magical attacks" },
    { id: "swarm", name: "Swarm", description: "Overwhelm with numbers" },
    { id: "brute", name: "Brute", description: "Heavy, slow, powerful attacks" },
    { id: "agile", name: "Agile", description: "Fast, evasive, hit-and-run" },
  ];

  // Difficulty levels for enemies
  const difficulties = [
    { id: "trivial", name: "Trivial" },
    { id: "easy", name: "Easy" },
    { id: "normal", name: "Normal" },
    { id: "hard", name: "Hard" },
    { id: "boss", name: "Boss" },
  ];

  const getRaceById = (id) => races.find((r) => r.id === id) || races[0];
  const getClassById = (id) => classes.find((c) => c.id === id) || classes[0];

  let selectedRaceId = "human";
  let selectedClassId = "commoner";

  // Tab state for traits section
  let activeTraitTab = "enemy";

  const config = {
    title: "Manage NPCs",
    subtitle: "Configure NPC profiles, traits, and dialog bindings.",
    listTitle: "NPCs",
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
        currentRoomID: "",
        enemyTrait: null,
        merchantTrait: null,
        isTemplate: true,  // Default to template for spawning
        isNew: true,
      });
    },
    badge: (element) => (element.race ? element.race.name : ""),
  };

  const toggleEnemyTrait = () => {
    store.update((state) => {
      state.selectedElement.enemyTrait = state.selectedElement.enemyTrait
        ? null
        : {
            creatureType: "beast",
            combatStyle: "melee",
            difficulty: "normal",
            aggroOnSight: false,
            attackPower: 5,
            defense: 2,
            xpReward: 10,
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

  // Load dialogs once auth token becomes available
  $: if ($isAuthenticated && $authToken && !hasLoadedDialogs) {
    loadDialogs();
  }
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
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
        <label class="label-caps">Level</label>
        <select class="input-base" bind:value={$store.selectedElement.level}>
          {#each levels as lvl}
            <option value={lvl}>{lvl}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps">Race</label>
        <select class="input-base" bind:value={selectedRaceId} on:change={onRaceChange}>
          {#each races as race}
            <option value={race.id}>{race.name}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps">Class</label>
        <select class="input-base" bind:value={selectedClassId} on:change={onClassChange}>
          {#each classes as cls}
            <option value={cls.id}>{cls.name}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps">Dialog ID</label>
        <select class="input-base" bind:value={$store.selectedElement.dialogID}>
          <option value="">None</option>
          {#if $dialogsValueHelp}
            {#each $dialogsValueHelp as dialog}
              <option value={dialog.id}>
                {dialog.name} ({dialog.id})
              </option>
            {/each}
          {/if}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps">Idle Dialog ID</label>
        <select class="input-base" bind:value={$store.selectedElement.idleDialogID}>
          <option value="">None</option>
          {#if $dialogsValueHelp}
            {#each $dialogsValueHelp as dialog}
              <option value={dialog.id}>
                {dialog.name} ({dialog.id})
              </option>
            {/each}
          {/if}
        </select>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps">Current Room ID</label>
        <input class="input-base" bind:value={$store.selectedElement.currentRoomID} />
      </div>
      <div class="space-y-1.5">
        <label class="label-caps">Hit Points</label>
        <div class="grid grid-cols-2 gap-2">
          <input class="input-base text-center" bind:value={$store.selectedElement.currentHitPoints} type="number" />
          <input class="input-base text-center" bind:value={$store.selectedElement.maxHitPoints} type="number" />
        </div>
      </div>
    </div>

  </div>

  <div slot="extensions" class="space-y-4">
    <!-- Tabbed Navigation for Traits -->
    <div class="flex items-center gap-1 border-b border-slate-200 dark:border-slate-700">
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
      {#if activeTraitTab === "enemy"}
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
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Configure combat behavior for this NPC when engaged in battle.
            </p>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
              <div class="space-y-1.5">
                <label class="label-caps">Creature Type</label>
                <select class="input-base text-xs" bind:value={$store.selectedElement.enemyTrait.creatureType}>
                  {#each creatureTypes as ct}
                    <option value={ct.id}>{ct.name}</option>
                  {/each}
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="label-caps">Combat Style</label>
                <select class="input-base text-xs" bind:value={$store.selectedElement.enemyTrait.combatStyle}>
                  {#each combatStyles as cs}
                    <option value={cs.id}>{cs.name}</option>
                  {/each}
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="label-caps">Difficulty</label>
                <select class="input-base text-xs" bind:value={$store.selectedElement.enemyTrait.difficulty}>
                  {#each difficulties as d}
                    <option value={d.id}>{d.name}</option>
                  {/each}
                </select>
              </div>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <div class="space-y-1.5">
                <label class="label-caps">Attack Power</label>
                <input class="input-base text-xs text-center" type="number" bind:value={$store.selectedElement.enemyTrait.attackPower} />
              </div>
              <div class="space-y-1.5">
                <label class="label-caps">Defense</label>
                <input class="input-base text-xs text-center" type="number" bind:value={$store.selectedElement.enemyTrait.defense} />
              </div>
              <div class="space-y-1.5">
                <label class="label-caps">XP Reward</label>
                <input class="input-base text-xs text-center" type="number" bind:value={$store.selectedElement.enemyTrait.xpReward} />
              </div>
            </div>

            <div class="flex items-center gap-3 pt-2 border-t border-slate-200 dark:border-slate-700/50">
              <label class="flex items-center gap-2 text-xs cursor-pointer">
                <input
                  type="checkbox"
                  class="rounded border-slate-300 dark:border-slate-600"
                  bind:checked={$store.selectedElement.enemyTrait.aggroOnSight}
                />
                <span class="label-caps">Aggressive (attacks on sight)</span>
              </label>
            </div>
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
                <label class="label-caps">Buy Price Modifier</label>
                <input class="input-base text-xs text-center" type="number" step="0.1" min="0" bind:value={$store.selectedElement.merchantTrait.buyPriceModifier} />
                <p class="text-[9px] text-slate-500">Multiplier when NPC buys from player (0.5 = 50% of base price)</p>
              </div>
              <div class="space-y-1.5">
                <label class="label-caps">Sell Price Modifier</label>
                <input class="input-base text-xs text-center" type="number" step="0.1" min="0" bind:value={$store.selectedElement.merchantTrait.sellPriceModifier} />
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
  </div>
</CRUDEditor>

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
