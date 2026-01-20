<script>
  import { onMount } from "svelte";
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";

  import {
    getItemTemplate,
    getItemTemplates,
    createItemTemplate,
    updateItemTemplate,
    deleteItemTemplate,
    getItemTypes,
    getItemSubTypes,
    getItemSlots,
    getItemQualities,
  } from "../api/item-templates.js";

  const store = createStore();
  let newPropertyName = "";
  let newAttributeName = "";

  let itemQualities = [];
  let itemTypes = [];
  let itemSubTypes = [];
  let itemSlots = [];
  let levels = [];
  for (let i = 1; i <= 50; i += 1) {
    levels.push(i);
  }

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
    getElement: getItemTemplate,
    create: createItemTemplate,
    update: updateItemTemplate,
    delete: deleteItemTemplate,
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
      script: "",
      isNew: true,
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
    getItemQualities((q) => (itemQualities = q));
    getItemTypes((t) => (itemTypes = t));
    getItemSubTypes((st) => (itemSubTypes = st));
    getItemSlots((s) => (itemSlots = s));
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
        <label class="label-caps">Item Type</label>
        <select class="input-base" bind:value={$store.selectedElement.type}>
          <option value="" disabled selected>Item Type</option>
          {#each itemTypes as type}
            <option value={type}>{type}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps">Item Subtype</label>
        <select class="input-base" bind:value={$store.selectedElement.subType}>
          <option value="" selected>Item Subtype</option>
          {#each itemSubTypes as subType}
            <option value={subType}>{subType}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps">Item Quality</label>
        <select class="input-base" bind:value={$store.selectedElement.quality}>
          <option value="" disabled selected>Item Quality</option>
          {#each itemQualities as quality}
            <option value={quality}>{quality}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps">Slot</label>
        <select class="input-base" bind:value={$store.selectedElement.slot}>
          {#each itemSlots as slot}
            <option value={slot}>{slot}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="space-y-1.5">
      <label class="label-caps">Linked Script</label>
      <input class="input-base" bind:value={$store.selectedElement.script} />
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
