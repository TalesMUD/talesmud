<script>
  import { onMount } from "svelte";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { v4 as uuidv4 } from "uuid";

  import {
    getItem,
    getItems,
    createItem,
    updateItem,
    deleteItem,
    getItemQualities,
    getItemTypes,
    getItemSubTypes,
    getItemSlots,
  } from "../api/items.js";

  const config = {
    title: "Manage Items",
    subtitle: "Create and update live items in the world.",
    listTitle: "Items",
    labels: {
      create: "Create Item",
      update: "Update Item",
      delete: "Delete",
    },
    get: getItems,
    getElement: getItem,
    create: createItem,
    update: updateItem,
    delete: deleteItem,
    new: (select) => {
      select({
        id: uuidv4(),
        name: "Unnamed Item",
        description: "",
        detail: "",
        type: "",
        slot: "inventory",
        quality: "normal",
        level: 1,
        properties: new Map(),
        attributes: new Map(),
        noPickup: false,
        tags: [],
        isNew: true,
      });
    },
    badge: (element) => {
      return element.quality + " " + element.subType;
    },
  };
  // create store outside of the component to use it in the slot..
  const store = createStore();

  ///////// ADDITIONAL DATA
  // additional data
  let itemQualities = [];
  let itemTypes = [];
  let itemSubTypes = [];
  let itemSlots = [];
  // create level array
  let levels = [];
  for (let i = 1; i <= 50; i++) {
    levels.push(i);
  }

  onMount(async () => {
    getItemQualities((q) => (itemQualities = q));
    getItemTypes((t) => (itemTypes = t));
    getItemSubTypes((st) => (itemSubTypes = st));
    getItemSlots((s) => (itemSlots = s));
  });
  /////////
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="item-level">Level</label>
        <select id="item-level" class="input-base" bind:value={$store.selectedElement.level}>
          {#each levels as lvl}
            <option value={lvl}>{lvl}</option>
          {/each}
        </select>
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="item-type">Item Type</label>
        <select id="item-type" class="input-base" bind:value={$store.selectedElement.type}>
          <option value="" disabled selected>Item Type</option>
          {#each itemTypes as type}
            <option value={type}>{type}</option>
          {/each}
        </select>
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="item-subtype">Item Subtype</label>
        <select id="item-subtype" class="input-base" bind:value={$store.selectedElement.subType}>
          <option value="" selected>Item Subtype</option>
          {#each itemSubTypes as subType}
            <option value={subType}>{subType}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="item-quality">Item Quality</label>
        <select id="item-quality" class="input-base" bind:value={$store.selectedElement.quality}>
          <option value="" disabled selected>Item Quality</option>
          {#each itemQualities as quality}
            <option value={quality}>{quality}</option>
          {/each}
        </select>
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="item-slot">Item Slot</label>
        <select id="item-slot" class="input-base" bind:value={$store.selectedElement.slot}>
          <option value="" selected>Item Slot</option>
          {#each itemSlots as slot}
            <option value={slot}>{slot}</option>
          {/each}
        </select>
      </div>
    </div>
  </div>
</CRUDEditor>
