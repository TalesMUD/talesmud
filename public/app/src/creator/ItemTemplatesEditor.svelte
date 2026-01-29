<script>
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";

  import { getAuth } from "../auth.js";
  const { isAuthenticated, authToken } = getAuth();

  import {
    getItem,
    getItemTemplates,
    createItem,
    updateItem,
    deleteItem,
    getItemTypes,
    getItemSubTypes,
    getItemSlots,
    getItemQualities,
  } from "../api/items.js";
  import { getScripts } from "../api/scripts.js";

  const store = createStore();
  const scriptsValueHelp = writable([]);
  const itemQualitiesStore = writable([]);
  const itemTypesStore = writable([]);
  const itemSubTypesStore = writable([]);
  const itemSlotsStore = writable([]);
  let newPropertyName = "";
  let newAttributeName = "";

  let levels = [];
  for (let i = 1; i <= 50; i += 1) {
    levels.push(i);
  }

  // Load scripts for OnUse dropdown
  const loadScripts = () => {
    if (!$isAuthenticated || !$authToken) return;
    getScripts(
      $authToken,
      [],
      (scripts) => scriptsValueHelp.set(scripts || []),
      (err) => console.error("Failed to load scripts:", err)
    );
  };

  // Load item metadata immediately (these are public endpoints, no auth needed)
  const loadItemMetadata = () => {
    console.log("Loading item metadata...");
    getItemQualities(
      (q) => {
        console.log("Received qualities:", q);
        itemQualitiesStore.set(q || []);
      },
      (err) => console.error("Failed to load item qualities:", err)
    );
    getItemTypes(
      (t) => {
        console.log("Received types:", t);
        itemTypesStore.set(t || []);
      },
      (err) => console.error("Failed to load item types:", err)
    );
    getItemSubTypes(
      (st) => {
        console.log("Received subtypes:", st);
        itemSubTypesStore.set(st || []);
      },
      (err) => console.error("Failed to load item subtypes:", err)
    );
    getItemSlots(
      (s) => {
        console.log("Received slots:", s);
        itemSlotsStore.set(s || []);
      },
      (err) => console.error("Failed to load item slots:", err)
    );
  };

  // Call immediately at module level
  loadItemMetadata();

  const config = {
    title: "Item Template Editor",
    subtitle: "Design base attributes and script behaviors for world items.",
    listTitle: "Templates",
    labels: {
      create: "Save Template",
      update: "Save Template",
      delete: "Delete",
    },
    get: getItemTemplates,
    getElement: getItem,
    create: createItem,
    update: updateItem,
    delete: deleteItem,
  };

  const createNewTemplate = () => {
    config.new((element) => {
      store.setSelectedElement(element);
    });
  };

  config.extraActions = [
    {
      label: "New Item",
      icon: "add_box",
      variant: "btn-outline",
      onClick: createNewTemplate,
    },
  ];

  config.new = (select) => {
    select({
      id: uuidv4(),
      name: "Unnamed Item",
      description: "",
      detail: "",
      type: "",
      subType: "",
      slot: "inventory",
      quality: "normal",
      level: 1,
      properties: {},
      attributes: {},
      noPickup: false,
      tags: [],
      isTemplate: true,
      isNew: true,
      // Consumable/stacking fields
      consumable: false,
      stackable: false,
      quantity: 1,
      maxStack: 1,
      onUseScriptId: "",
    });
  };

  const addAttribute = () => {
    store.update((state) => {
      if (!state.selectedElement.attributes) {
        state.selectedElement.attributes = {};
      }
      if (newAttributeName) {
        state.selectedElement.attributes[newAttributeName] = "value";
        newAttributeName = "";
      }
      return state;
    });
  };

  const addProperty = () => {
    store.update((state) => {
      if (!state.selectedElement.properties) {
        state.selectedElement.properties = {};
      }
      if (newPropertyName) {
        state.selectedElement.properties[newPropertyName] = "value";
        newPropertyName = "";
      }
      return state;
    });
  };

  onMount(async () => {
    console.log("ItemTemplatesEditor: onMount called");
    loadScripts();
  });
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="template-level">Level</label>
        <select id="template-level" class="input-base" bind:value={$store.selectedElement.level}>
          {#each levels as lvl}
            <option value={lvl}>{lvl}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="template-type">Item Type</label>
        <select id="template-type" class="input-base" bind:value={$store.selectedElement.type}>
          <option value="" disabled selected>Item Type</option>
          {#each $itemTypesStore as type}
            <option value={type}>{type}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="template-subtype">Item Subtype</label>
        <select id="template-subtype" class="input-base" bind:value={$store.selectedElement.subType}>
          <option value="" selected>Item Subtype</option>
          {#each $itemSubTypesStore as subType}
            <option value={subType}>{subType}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="template-quality">Item Quality</label>
        <select id="template-quality" class="input-base" bind:value={$store.selectedElement.quality}>
          <option value="" disabled selected>Item Quality</option>
          {#each $itemQualitiesStore as quality}
            <option value={quality}>{quality}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="template-slot">Slot</label>
        <select id="template-slot" class="input-base" bind:value={$store.selectedElement.slot}>
          {#each $itemSlotsStore as slot}
            <option value={slot}>{slot}</option>
          {/each}
        </select>
      </div>
    </div>

    <!-- Consumable & Stacking Settings -->
    <div class="card p-4 space-y-4">
      <div class="label-caps text-primary">Consumable & Stacking</div>
      <div class="flex flex-wrap items-center gap-6">
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            class="w-4 h-4 rounded border-slate-600 bg-slate-800 text-primary focus:ring-primary focus:ring-offset-0"
            bind:checked={$store.selectedElement.consumable}
          />
          <span class="text-sm text-slate-300">Consumable</span>
          <span class="text-[10px] text-slate-500">(Removed/decremented on use)</span>
        </label>
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            class="w-4 h-4 rounded border-slate-600 bg-slate-800 text-primary focus:ring-primary focus:ring-offset-0"
            bind:checked={$store.selectedElement.stackable}
          />
          <span class="text-sm text-slate-300">Stackable</span>
        </label>
      </div>
      {#if $store.selectedElement.stackable}
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1.5">
            <label class="label-caps" for="template-quantity">Default Quantity</label>
            <input
              id="template-quantity"
              type="number"
              class="input-base"
              min="1"
              bind:value={$store.selectedElement.quantity}
            />
          </div>
          <div class="space-y-1.5">
            <label class="label-caps" for="template-maxstack">Max Stack</label>
            <input
              id="template-maxstack"
              type="number"
              class="input-base"
              min="1"
              bind:value={$store.selectedElement.maxStack}
            />
          </div>
        </div>
      {/if}
    </div>

    <!-- On Use Script -->
    <div class="space-y-1.5">
      <label class="label-caps" for="template-onuse-script">On Use Script</label>
      <select id="template-onuse-script" class="input-base" bind:value={$store.selectedElement.onUseScriptId}>
        <option value="">None</option>
        {#each $scriptsValueHelp as script}
          <option value={script.id}>{script.name}</option>
        {/each}
      </select>
      <p class="text-[10px] text-slate-400">
        Lua script executed when item is used. Context: ctx.item, ctx.character, ctx.room
      </p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="card p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div class="label-caps text-primary">Attributes</div>
          <button class="text-xs text-primary hover:underline" type="button" on:click={addAttribute}>
            + Add
          </button>
        </div>
        <div class="flex gap-2">
          <input class="input-base text-xs" placeholder="Attribute name" bind:value={newAttributeName} />
        </div>
        <div class="space-y-2">
          {#each Object.entries($store.selectedElement.attributes || {}) as [key, value]}
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono text-slate-400">{key}</span>
              <input class="input-base text-xs flex-1" bind:value={$store.selectedElement.attributes[key]} />
            </div>
          {/each}
        </div>
      </div>

      <div class="card p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div class="label-caps text-primary">Properties</div>
          <button class="text-xs text-primary hover:underline" type="button" on:click={addProperty}>
            + Add
          </button>
        </div>
        <div class="flex gap-2">
          <input class="input-base text-xs" placeholder="Property name" bind:value={newPropertyName} />
        </div>
        <div class="space-y-2">
          {#each Object.entries($store.selectedElement.properties || {}) as [key, value]}
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono text-slate-400">{key}</span>
              <input class="input-base text-xs flex-1" bind:value={$store.selectedElement.properties[key]} />
            </div>
          {/each}
        </div>
      </div>
    </div>
  </div>
</CRUDEditor>
