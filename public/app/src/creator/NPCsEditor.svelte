<script>
  import { onMount } from "svelte";
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

  const getRaceById = (id) => races.find((r) => r.id === id) || races[0];
  const getClassById = (id) => classes.find((c) => c.id === id) || classes[0];

  let selectedRaceId = "human";
  let selectedClassId = "commoner";

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
            aggressive: false,
            attackPower: 5,
            defense: 2,
            experienceReward: 10,
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

  onMount(async () => {
    getDialogs(
      $authToken,
      [],
      (dialogs) => {
        dialogsValueHelp.set(dialogs || []);
      },
      () => {}
    );
  });
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
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

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="card p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div class="label-caps text-primary">Enemy Trait</div>
          <button class="text-xs text-primary hover:underline" type="button" on:click={toggleEnemyTrait}>
            {#if $store.selectedElement.enemyTrait}Remove{:else}Add{/if}
          </button>
        </div>
        {#if $store.selectedElement.enemyTrait}
          <div class="grid grid-cols-1 gap-2 text-xs text-slate-500 dark:text-slate-400">
            <label class="label-caps">Aggressive</label>
            <input type="checkbox" bind:checked={$store.selectedElement.enemyTrait.aggressive} />
            <label class="label-caps">Attack Power</label>
            <input class="input-base text-xs" type="number" bind:value={$store.selectedElement.enemyTrait.attackPower} />
            <label class="label-caps">Defense</label>
            <input class="input-base text-xs" type="number" bind:value={$store.selectedElement.enemyTrait.defense} />
            <label class="label-caps">Experience Reward</label>
            <input class="input-base text-xs" type="number" bind:value={$store.selectedElement.enemyTrait.experienceReward} />
          </div>
        {:else}
          <p class="text-xs text-slate-500 dark:text-slate-400">No enemy trait configured.</p>
        {/if}
      </div>

      <div class="card p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div class="label-caps text-primary">Merchant Trait</div>
          <button class="text-xs text-primary hover:underline" type="button" on:click={toggleMerchantTrait}>
            {#if $store.selectedElement.merchantTrait}Remove{:else}Add{/if}
          </button>
        </div>
        {#if $store.selectedElement.merchantTrait}
          <div class="grid grid-cols-1 gap-2 text-xs text-slate-500 dark:text-slate-400">
            <label class="label-caps">Buy Modifier</label>
            <input class="input-base text-xs" type="number" step="0.1" bind:value={$store.selectedElement.merchantTrait.buyPriceModifier} />
            <label class="label-caps">Sell Modifier</label>
            <input class="input-base text-xs" type="number" step="0.1" bind:value={$store.selectedElement.merchantTrait.sellPriceModifier} />
          </div>
        {:else}
          <p class="text-xs text-slate-500 dark:text-slate-400">No merchant trait configured.</p>
        {/if}
      </div>
    </div>
  </div>
</CRUDEditor>
