<script>
  import { onMount } from "svelte";
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { getAuth } from "../auth.js";

  import {
    getCharacterTemplates,
    getCharacterTemplate,
    createCharacterTemplate,
    updateCharacterTemplate,
    deleteCharacterTemplate,
    seedCharacterTemplates,
  } from "../api/character-templates.js";

  import { getItemTemplates } from "../api/item-templates.js";

  const store = createStore();
  const { isAuthenticated, authToken } = getAuth();

  let newAttributeName = "";
  let seedStatus = "";
  let itemTemplates = [];
  let itemTemplatesLoaded = false;

  // Archetype options
  const archetypes = ["warrior", "rogue", "mage", "cleric", "ranger", "druid", "custom"];

  // Levels 1-20
  const levels = [];
  for (let i = 1; i <= 20; i++) {
    levels.push(i);
  }

  // Slots for starting items
  const itemSlots = [
    "inventory", "main_hand", "off_hand", "head", "chest", "legs", "boots", "hands", "neck", "ring1", "ring2"
  ];

  const config = {
    title: "Character Template Editor",
    subtitle: "Define character archetypes with stats, backstory, and starting gear.",
    listTitle: "Templates",
    labels: {
      create: "Save Template",
      update: "Save Template",
      delete: "Delete",
    },
    get: getCharacterTemplates,
    getElement: getCharacterTemplate,
    create: createCharacterTemplate,
    update: updateCharacterTemplate,
    delete: deleteCharacterTemplate,
    badge: (element) => `${element.archetype || "custom"} · Lvl ${element.level || 1}`,
  };

  const createNewTemplate = () => {
    config.new((element) => {
      store.setSelectedElement(element);
    });
  };

  const doSeed = () => {
    seedStatus = "Seeding...";
    seedCharacterTemplates(
      $authToken,
      (result) => {
        if (result.status === "seeded") {
          seedStatus = `Seeded ${result.characterTemplates} character templates and ${result.itemTemplatesCreated} item templates!`;
        } else {
          seedStatus = `Skipped: ${result.message}`;
        }
        // Reload data
        config.get($authToken, null, (all) => store.setElements(all), console.error);
        // Reload item templates too
        loadItemTemplates();
      },
      (err) => {
        seedStatus = "Error seeding: " + (err.response?.data?.error || err.message);
      }
    );
  };

  config.extraActions = [
    {
      label: "Seed Defaults",
      icon: "download",
      variant: "btn-outline",
      onClick: doSeed,
    },
    {
      label: "New Template",
      icon: "add_box",
      variant: "btn-outline",
      onClick: createNewTemplate,
    },
  ];

  config.new = (select) => {
    select({
      id: uuidv4(),
      name: "New Template",
      description: "",
      backstory: "",
      originArea: "",
      archetype: "warrior",
      level: 1,
      currentHitPoints: 20,
      maxHitPoints: 20,
      race: { id: "human", name: "Human", description: "", heritage: "" },
      class: { id: "warrior", name: "Warrior", description: "", armorType: "Plate", combatType: "Melee" },
      attributes: [
        { name: "Strength", short: "str", value: 10 },
        { name: "Dexterity", short: "dex", value: 10 },
        { name: "Intelligence", short: "int", value: 10 },
        { name: "Wisdom", short: "wis", value: 10 },
        { name: "Stamina", short: "sta", value: 10 },
      ],
      startingItems: [],
      source: "custom",
      isNew: true,
    });
  };

  const addAttribute = () => {
    store.update((state) => {
      if (!state.selectedElement.attributes) {
        state.selectedElement.attributes = [];
      }
      if (newAttributeName) {
        state.selectedElement.attributes.push({
          name: newAttributeName,
          short: newAttributeName.slice(0, 3).toLowerCase(),
          value: 10,
        });
        newAttributeName = "";
      }
      return state;
    });
  };

  const removeAttribute = (index) => {
    store.update((state) => {
      state.selectedElement.attributes.splice(index, 1);
      return state;
    });
  };

  const addStartingItem = () => {
    store.update((state) => {
      if (!state.selectedElement.startingItems) {
        state.selectedElement.startingItems = [];
      }
      state.selectedElement.startingItems.push({
        slot: "main_hand",
        itemTemplateId: "",
        itemTemplateName: "",
      });
      return state;
    });
  };

  const removeStartingItem = (index) => {
    store.update((state) => {
      state.selectedElement.startingItems.splice(index, 1);
      return state;
    });
  };

  // When an item template is selected, update both ID and name
  const onItemTemplateSelect = (index, templateId) => {
    store.update((state) => {
      const template = itemTemplates.find(t => t.id === templateId);
      state.selectedElement.startingItems[index].itemTemplateId = templateId;
      state.selectedElement.startingItems[index].itemTemplateName = template?.name || "";
      return state;
    });
  };

  // Get item template name by ID
  const getItemTemplateName = (id) => {
    if (!id) return "(none)";
    const template = itemTemplates.find(t => t.id === id);
    return template?.name || id;
  };

  // Load item templates
  const loadItemTemplates = () => {
    if (!$authToken) return;
    getItemTemplates(
      $authToken,
      null,
      (templates) => {
        itemTemplates = templates || [];
        itemTemplatesLoaded = true;
      },
      (err) => {
        console.error("Failed to load item templates:", err);
        itemTemplates = [];
        itemTemplatesLoaded = true;
      }
    );
  };

  // Reactive: load item templates when auth becomes available
  $: if ($isAuthenticated && $authToken && !itemTemplatesLoaded) {
    loadItemTemplates();
  }

  onMount(async () => {
    // Initial load attempt
    loadItemTemplates();
  });
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
    {#if seedStatus}
      <div class="text-sm text-primary bg-primary/10 px-3 py-2 rounded">{seedStatus}</div>
    {/if}

    <!-- Basic Info -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="ct_archetype">Archetype</label>
        <select id="ct_archetype" class="input-base" bind:value={$store.selectedElement.archetype}>
          {#each archetypes as arch}
            <option value={arch}>{arch}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="ct_level">Level</label>
        <select id="ct_level" class="input-base" bind:value={$store.selectedElement.level}>
          {#each levels as lvl}
            <option value={lvl}>{lvl}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="ct_source">Source</label>
        <input id="ct_source" class="input-base" bind:value={$store.selectedElement.source} placeholder="system / custom" />
      </div>
    </div>

    <!-- Backstory & Origin -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="ct_backstory">Backstory</label>
        <textarea id="ct_backstory" rows="3" class="input-base" bind:value={$store.selectedElement.backstory} placeholder="Character background lore…"></textarea>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="ct_origin">Origin Area</label>
        <input id="ct_origin" class="input-base" bind:value={$store.selectedElement.originArea} placeholder="e.g. Oldtown, Forest Edge" />
      </div>
    </div>

    <!-- HP -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="ct_hp_current">Current HP</label>
        <input id="ct_hp_current" type="number" class="input-base" bind:value={$store.selectedElement.currentHitPoints} />
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="ct_hp_max">Max HP</label>
        <input id="ct_hp_max" type="number" class="input-base" bind:value={$store.selectedElement.maxHitPoints} />
      </div>
    </div>

    <!-- Race -->
    <div class="card p-4 space-y-3">
      <div class="label-caps text-primary">Race</div>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="ct_race_id">ID</label>
          <input id="ct_race_id" class="input-base text-xs" bind:value={$store.selectedElement.race.id} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="ct_race_name">Name</label>
          <input id="ct_race_name" class="input-base text-xs" bind:value={$store.selectedElement.race.name} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="ct_race_heritage">Heritage</label>
          <input id="ct_race_heritage" class="input-base text-xs" bind:value={$store.selectedElement.race.heritage} />
        </div>
      </div>
    </div>

    <!-- Class -->
    <div class="card p-4 space-y-3">
      <div class="label-caps text-primary">Class</div>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="ct_class_id">ID</label>
          <input id="ct_class_id" class="input-base text-xs" bind:value={$store.selectedElement.class.id} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="ct_class_name">Name</label>
          <input id="ct_class_name" class="input-base text-xs" bind:value={$store.selectedElement.class.name} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="ct_class_armor">Armor Type</label>
          <select id="ct_class_armor" class="input-base text-xs" bind:value={$store.selectedElement.class.armorType}>
            <option value="Cloth">Cloth</option>
            <option value="Leather">Leather</option>
            <option value="Plate">Plate</option>
          </select>
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="ct_class_combat">Combat Type</label>
          <select id="ct_class_combat" class="input-base text-xs" bind:value={$store.selectedElement.class.combatType}>
            <option value="Melee">Melee</option>
            <option value="Ranged">Ranged</option>
            <option value="Magic">Magic</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Attributes -->
    <div class="card p-4 space-y-3">
      <div class="flex items-center justify-between">
        <div class="label-caps text-primary">Attributes</div>
        <button class="text-xs text-primary hover:underline" type="button" on:click={addAttribute}>
          + Add
        </button>
      </div>
      <div class="flex gap-2 mb-2">
        <input class="input-base text-xs" placeholder="Attribute name" bind:value={newAttributeName} />
      </div>
      <div class="space-y-2">
        {#if $store.selectedElement.attributes}
          {#each $store.selectedElement.attributes as attr, i}
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono text-slate-400 w-24">{attr.name}</span>
              <span class="text-xs font-mono text-slate-500 w-12">({attr.short})</span>
              <input class="input-base text-xs flex-1" type="number" bind:value={$store.selectedElement.attributes[i].value} />
              <button class="text-xs text-red-400 hover:text-red-300" type="button" on:click={() => removeAttribute(i)}>
                ✕
              </button>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Starting Items -->
    <div class="card p-4 space-y-3">
      <div class="flex items-center justify-between">
        <div class="label-caps text-primary">Starting Items</div>
        <button class="text-xs text-primary hover:underline" type="button" on:click={addStartingItem}>
          + Add
        </button>
      </div>

      {#if itemTemplates.length === 0 && itemTemplatesLoaded}
        <p class="text-xs text-amber-400">
          No item templates found. Click "Seed Defaults" to create starter items, or create them in the Item Templates editor.
        </p>
      {/if}

      <div class="space-y-2">
        {#if $store.selectedElement.startingItems}
          {#each $store.selectedElement.startingItems as item, i}
            <div class="flex items-center gap-2">
              <!-- Slot dropdown -->
              <label class="sr-only" for={`slot_${i}`}>Slot</label>
              <select id={`slot_${i}`} class="input-base text-xs w-28" bind:value={$store.selectedElement.startingItems[i].slot}>
                {#each itemSlots as slot}
                  <option value={slot}>{slot}</option>
                {/each}
              </select>

              <!-- Item Template searchable dropdown -->
              <label class="sr-only" for={`item_${i}`}>Item Template</label>
              <select
                id={`item_${i}`}
                class="input-base text-xs flex-1"
                value={item.itemTemplateId || ""}
                on:change={(e) => onItemTemplateSelect(i, e.target.value)}
              >
                <option value="">-- Select Item Template --</option>
                {#each itemTemplates as template}
                  <option value={template.id}>
                    {template.name} ({template.type || "item"})
                  </option>
                {/each}
              </select>

              <button class="text-xs text-red-400 hover:text-red-300" type="button" on:click={() => removeStartingItem(i)}>
                ✕
              </button>
            </div>
          {/each}
        {/if}
      </div>

      <p class="text-xs text-slate-500">
        Select item templates from the dropdown. When a character is created from this template, actual items will be generated from these templates.
      </p>
    </div>
  </div>
</CRUDEditor>
