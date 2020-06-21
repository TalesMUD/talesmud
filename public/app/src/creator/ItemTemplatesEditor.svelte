<style>

</style>

<script>
	import Toolbar from './Toolbar.svelte';
  import Sprites from "./../game/Sprites.svelte";
  import { onMount } from "svelte";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { v4 as uuidv4 } from "uuid";

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

  const config = {
    title: "Manage Item Templates",
    actions: [
      {
        icon: "add",
        name: "Item from Template",
        color: "",
        fnc: () => {
          console.log(
            "Create new template for id " + $store.selectedElement.id
          );
        },
      },
    ],
    get: getItemTemplates,
    getElement: getItemTemplate,
    create: createItemTemplate,
    update: updateItemTemplate,
    delete: deleteItemTemplate,
    refreshUI: () => {
      var elems = document.querySelectorAll("select");
      var instances = M.FormSelect.init(elems, {});

      // second time to fix the selects
      setTimeout(function () {
        var elems = document.querySelectorAll("select");
        var instances = M.FormSelect.init(elems, {});
      }, 50);
    },

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
  var levels = [];
  for (var i = 1; i <= 50; i++) {
    levels.push(i);
  }

  onMount(async () => {
    getItemQualities((q) => (itemQualities = q));
    getItemTypes((t) => (itemTypes = t));
    getItemSubTypes((st) => (itemSubTypes = st));
    getItemSlots((s) => (itemSlots = s));
  });
  /////////

  const addAttribute = () => {
store.update((state) => {
      if (state.selectedElement.attributes == null) {
        state.selectedElement.attributes = {};
      }

      state.selectedElement.exits.push({
        name: "new_attribg",
        descrvalueiption: "todo",
        target: "select target",
      });
      return state;
    });
    config.refreshUI();

};
  const addProperty = () => {
    console.log("New attribute");
  };
  const attributesToolbar = {
    title: "Attributes",
    small: true,
    actions: [
      {
        icon: "add",
        fnc: () => {
          addAttribute();
        },
      },
    ],
  };
  const propertiesToolbar = {
    title: "Properties",
    small: true,
    actions: [
      {
        icon: "add",
        fnc: () => {
          addProperty();
        },
      },
    ],
  };

</script>

<CRUDEditor store="{store}" config="{config}">

  <span slot="hero" class="col s1 valign-wrapper">
    <Sprites item="weapon" />
  </span>

  <div slot="content">
    <div class="row">

      <div class="margininput input-field col s1">
        <select bind:value="{$store.selectedElement.level}" on:change>
          {#each levels as lvl}
            <option value="{lvl}">{lvl}</option>
          {/each}
        </select>
        <label>Level</label>
      </div>

      <div class="margininput input-field col s5">
        <select bind:value="{$store.selectedElement.type}" on:change>
          <option value="" disabled selected>Item Type</option>
          {#each itemTypes as type}
            <option value="{type}">{type.capitalize()}</option>
          {/each}
        </select>
        <label>Select Item Type</label>
      </div>

      <div class="margininput input-field col s5">
        <select bind:value="{$store.selectedElement.subType}" on:change>
          <option value="" selected>Item Subtype</option>
          {#each itemSubTypes as subType}
            <option value="{subType}">{subType.capitalize()}</option>
          {/each}
        </select>
        <label>Select Item Sub Type</label>
      </div>

    </div>
    <div class="row">

      <div class="margininput input-field col s5">
        <select bind:value="{$store.selectedElement.quality}" on:change>
          <option value="" disabled selected>Item Quality</option>
          {#each itemQualities as quality}
            <option value="{quality}">{quality.capitalize()}</option>
          {/each}

        </select>
        <label>Select Item Quality</label>
      </div>

      <div class="margininput input-field col s5">
        <select
          class="margininput"
          bind:value="{$store.selectedElement.slot}"
          on:change
        >
          <option value="" selected>Item Slot</option>
          {#each itemSlots as slot}
            <option class="select" value="{slot}">{slot.capitalize()}</option>
          {/each}
        </select>
        <label>Select Item Slot</label>
      </div>

    </div>
  </div>

  <div slot="extensions">

    <Toolbar toolbar="{attributesToolbar}" />

    <div class="card-panel blue-grey darken-3">
      <table>
        <thead>
          <tr>
            <th>Attribute</th>
            <th>Value</th>
          </tr>
        </thead>
        <tbody>
          {#each Object.entries($store.selectedElement.attributes) as [key, value]}
            <tr>
              <td>{key}</td>
              <td>{value}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <Toolbar toolbar="{propertiesToolbar}" />

    <div class="card-panel blue-grey darken-3">
      <table>
        <thead>
          <tr>
            <th>Property</th>
            <th>Value</th>
          </tr>
        </thead>
        <tbody>
          {#each Object.entries($store.selectedElement.properties) as [key, value]}
            <tr>
              <td>{key}</td>
              <td>{value}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

  </div>
</CRUDEditor>
