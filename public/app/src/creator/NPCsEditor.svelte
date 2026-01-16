<script>
  import { onMount } from "svelte";
  import { writable } from "svelte/store";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { v4 as uuidv4 } from "uuid";
  import Toolbar from "./Toolbar.svelte";

  import { getAuth } from "../auth.js";
  const { isAuthenticated, authToken } = getAuth();

  import {
    getNPC,
    getNPCs,
    createNPC,
    updateNPC,
    deleteNPC,
  } from "../api/npcs.js";

  import { getDialogs } from "../api/dialogs.js";

  // Value help for dialogs
  const dialogsValueHelp = writable([]);

  // Levels array
  let levels = [];
  for (let i = 1; i <= 50; i++) {
    levels.push(i);
  }

  // Race and class options (matching Go struct format)
  const races = [
    { id: "human", name: "Human", description: "The common race", heritage: "Big cities" },
    { id: "elve", name: "Elve", description: "Splendid forestwalkers", heritage: "Near the forest" },
    { id: "dwarf", name: "Dwarf", description: "Small, but dont underestimate them", heritage: "Deep below the mountains" },
    { id: "halfling", name: "Halfling", description: "Small and nimble", heritage: "Peaceful villages" },
    { id: "orc", name: "Orc", description: "Strong and fierce", heritage: "Tribal lands" },
    { id: "goblin", name: "Goblin", description: "Cunning and sneaky", heritage: "Dark caves" },
    { id: "undead", name: "Undead", description: "Risen from death", heritage: "Crypts and graveyards" },
    { id: "demon", name: "Demon", description: "Infernal beings", heritage: "The underworld" },
    { id: "beast", name: "Beast", description: "Wild creatures", heritage: "The wilderness" },
  ];
  const classes = [
    { id: "warrior", name: "Warrior", description: "Strong plate wearing melee warrior", armorType: "Plate", combatType: "Melee" },
    { id: "ranger", name: "Ranger", description: "Quick bow wielding ranged combatant", armorType: "Leather", combatType: "Ranged" },
    { id: "wizard", name: "Wizard", description: "Master of the elements", armorType: "Cloth", combatType: "Magic" },
    { id: "rogue", name: "Rogue", description: "Stealthy and deadly", armorType: "Leather", combatType: "Melee" },
    { id: "cleric", name: "Cleric", description: "Divine healer and protector", armorType: "Plate", combatType: "Magic" },
    { id: "merchant", name: "Merchant", description: "Trader of goods", armorType: "Cloth", combatType: "Melee" },
    { id: "guard", name: "Guard", description: "Protector of the realm", armorType: "Plate", combatType: "Melee" },
    { id: "commoner", name: "Commoner", description: "Simple folk", armorType: "Cloth", combatType: "Melee" },
  ];

  // Helper to get race/class by id
  const getRaceById = (id) => races.find(r => r.id === id) || races[0];
  const getClassById = (id) => classes.find(c => c.id === id) || classes[0];

  // Reactive selected race/class ids for the dropdowns
  let selectedRaceId = "human";
  let selectedClassId = "commoner";

  const config = {
    title: "Manage NPCs",
    actions: [],
    get: getNPCs,
    getElement: getNPC,
    create: createNPC,
    update: updateNPC,
    delete: deleteNPC,
    beforeSelect: (element) => {
      if (element.enemyTrait === undefined) {
        element.enemyTrait = null;
      }
      if (element.merchantTrait === undefined) {
        element.merchantTrait = null;
      }
      // Initialize race/class if not set
      if (!element.race) {
        element.race = getRaceById("human");
      }
      if (!element.class) {
        element.class = getClassById("commoner");
      }
      // Update the dropdown selection based on loaded element
      selectedRaceId = element.race.id || "human";
      selectedClassId = element.class.id || "commoner";
    },
    refreshUI: () => {
      var elems = document.querySelectorAll("select");
      var instances = M.FormSelect.init(elems, {});

      setTimeout(function () {
        var elems = document.querySelectorAll("select");
        var instances = M.FormSelect.init(elems, {});
      }, 50);

      M.updateTextFields();
      var elems2 = document.querySelectorAll(".collapsible");
      if (elems2 != undefined) {
        var instances = M.Collapsible.init(elems2, {});
      }

      var textareas = document.querySelectorAll(".materialize-textarea");
      textareas.forEach((e) => {
        M.textareaAutoResize(e);
      });

      // trigger valuehelp updates
      dialogsValueHelp.set($dialogsValueHelp);
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

    badge: (element) => {
      return element.race ? element.race.name : "";
    },
  };

  const store = createStore();

  // Toggle enemy trait
  const toggleEnemyTrait = () => {
    store.update((state) => {
      if (state.selectedElement.enemyTrait) {
        state.selectedElement.enemyTrait = null;
      } else {
        state.selectedElement.enemyTrait = {
          aggressive: false,
          attackPower: 5,
          defense: 2,
          experienceReward: 10,
        };
      }
      return state;
    });
    config.refreshUI();
  };

  // Toggle merchant trait
  const toggleMerchantTrait = () => {
    store.update((state) => {
      if (state.selectedElement.merchantTrait) {
        state.selectedElement.merchantTrait = null;
      } else {
        state.selectedElement.merchantTrait = {
          buyPriceModifier: 1.0,
          sellPriceModifier: 0.5,
          inventory: [],
        };
      }
      return state;
    });
    config.refreshUI();
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

  // Reactive toolbar configs
  $: enemyToolbar = {
    title: "Enemy Trait",
    small: true,
    actions: [
      {
        icon: ($store.selectedElement && $store.selectedElement.enemyTrait) ? "remove" : "add",
        fnc: () => toggleEnemyTrait(),
      },
    ],
  };

  $: merchantToolbar = {
    title: "Merchant Trait",
    small: true,
    actions: [
      {
        icon: ($store.selectedElement && $store.selectedElement.merchantTrait) ? "remove" : "add",
        fnc: () => toggleMerchantTrait(),
      },
    ],
  };
</script>

<CRUDEditor store="{store}" config="{config}">
  <div slot="content">
    <div class="row">
      <div class="margininput input-field col s2">
        <select bind:value="{$store.selectedElement.level}" on:change>
          {#each levels as lvl}
            <option value="{lvl}">{lvl}</option>
          {/each}
        </select>
        <label>Level</label>
      </div>

      <div class="margininput input-field col s5">
        <select bind:value="{selectedRaceId}" on:change="{() => { $store.selectedElement.race = getRaceById(selectedRaceId); }}">
          <option value="" disabled>Race</option>
          {#each races as race}
            <option value="{race.id}">{race.name}</option>
          {/each}
        </select>
        <label>Race</label>
      </div>

      <div class="margininput input-field col s5">
        <select bind:value="{selectedClassId}" on:change="{() => { $store.selectedElement.class = getClassById(selectedClassId); }}">
          <option value="" disabled>Class</option>
          {#each classes as cls}
            <option value="{cls.id}">{cls.name}</option>
          {/each}
        </select>
        <label>Class</label>
      </div>
    </div>

    <div class="row">
      <div class="no-padding input-field col s3">
        <input
          placeholder="Current HP"
          id="currentHitPoints"
          type="number"
          bind:value="{$store.selectedElement.currentHitPoints}"
        />
        <label class="active first_label" for="currentHitPoints">Current HP</label>
      </div>

      <div class="input-field col s3">
        <input
          placeholder="Max HP"
          id="maxHitPoints"
          type="number"
          bind:value="{$store.selectedElement.maxHitPoints}"
        />
        <label class="active" for="maxHitPoints">Max HP</label>
      </div>

      <div class="input-field col s6">
        <select bind:value="{$store.selectedElement.dialogID}" on:change>
          <option value="">No Dialog</option>
          {#each $dialogsValueHelp as dialog}
            <option value="{dialog.id}">{dialog.name}</option>
          {/each}
        </select>
        <label>Main Dialog</label>
      </div>
    </div>

    <div class="row">
      <div class="input-field col s6">
        <select bind:value="{$store.selectedElement.idleDialogID}" on:change>
          <option value="">No Idle Dialog</option>
          {#each $dialogsValueHelp as dialog}
            <option value="{dialog.id}">{dialog.name}</option>
          {/each}
        </select>
        <label>Idle Dialog</label>
      </div>

      <div class="input-field col s6">
        <input
          placeholder="Current Room ID"
          id="currentRoomID"
          type="text"
          bind:value="{$store.selectedElement.currentRoomID}"
        />
        <label class="active" for="currentRoomID">Current Room ID</label>
      </div>
    </div>
  </div>

  <div slot="extensions">
    <Toolbar toolbar="{enemyToolbar}" />

    {#if $store.selectedElement && $store.selectedElement.enemyTrait}
      <div class="card-panel blue-grey darken-3">
        <div class="row">
          <div class="col s3">
            <label>
              <input
                type="checkbox"
                bind:checked="{$store.selectedElement.enemyTrait.aggressive}"
              />
              <span>Aggressive</span>
            </label>
          </div>

          <div class="input-field col s3">
            <input
              placeholder="Attack Power"
              id="attackPower"
              type="number"
              bind:value="{$store.selectedElement.enemyTrait.attackPower}"
            />
            <label class="active" for="attackPower">Attack Power</label>
          </div>

          <div class="input-field col s3">
            <input
              placeholder="Defense"
              id="defense"
              type="number"
              bind:value="{$store.selectedElement.enemyTrait.defense}"
            />
            <label class="active" for="defense">Defense</label>
          </div>

          <div class="input-field col s3">
            <input
              placeholder="XP Reward"
              id="experienceReward"
              type="number"
              bind:value="{$store.selectedElement.enemyTrait.experienceReward}"
            />
            <label class="active" for="experienceReward">XP Reward</label>
          </div>
        </div>
      </div>
    {/if}

    <Toolbar toolbar="{merchantToolbar}" />

    {#if $store.selectedElement && $store.selectedElement.merchantTrait}
      <div class="card-panel blue-grey darken-3">
        <div class="row">
          <div class="input-field col s6">
            <input
              placeholder="Buy Price Modifier"
              id="buyPriceModifier"
              type="number"
              step="0.1"
              bind:value="{$store.selectedElement.merchantTrait.buyPriceModifier}"
            />
            <label class="active" for="buyPriceModifier">Buy Price Modifier</label>
          </div>

          <div class="input-field col s6">
            <input
              placeholder="Sell Price Modifier"
              id="sellPriceModifier"
              type="number"
              step="0.1"
              bind:value="{$store.selectedElement.merchantTrait.sellPriceModifier}"
            />
            <label class="active" for="sellPriceModifier">Sell Price Modifier</label>
          </div>
        </div>
      </div>
    {/if}
  </div>
</CRUDEditor>
