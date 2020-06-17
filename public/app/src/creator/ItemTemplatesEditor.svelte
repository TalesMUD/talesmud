<style>
  .sidelist {
    width: 20em;
  }
  textarea {
    color: white;
    margin-top: 1em;
  }
  input {
    color: white;
  }
  :global(input) {
    color: #fff;
  }
  .margininput {
    margin-left: 0.5em;
    margin-right: 0.5em;
  }
  input:disabled {
    color: white;
  }
  .no-padding {
    padding: 0;
  }
  .title {
    font-size: 2em;
  }

  .first_label {
    transform: translateX(-10px) translateY(-14px) scale(0.8);
  }

  .btn-small {
    margin-right: 0.5em;
    margin-left: 0.5em;
  }

  .collection-item {
    color: #333;
  }
  .collection {
    color: #333;
  }

  .materialize-textarea {
    border-bottom: none;
  }
  .header {
    font-size: 150%;
    font-weight: 600;
  }
  .active {
    color: #ccc;
  }
  label {
    color: #00796b;
  }
</style>

<script>
  import ItemsToolbar from "./ItemsToolbar.svelte";
  import { store } from "./ItemsEditorStore.js";
  import { PlusIcon } from "svelte-feather-icons";
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { createAuth, getAuth } from "../auth.js";
  import { v4 as uuidv4 } from "uuid";

  import axios from "axios";
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

  export let location;

  let itemQualities = [];
  let itemTypes = [];
  let itemSubTypes = [];
  let itemSlots = [];

  // create level array
  var levels = [];
  for (var i = 1; i <= 50; i++) {
    levels.push(i);
  }

  const {
    isLoading,
    isAuthenticated,
    login,
    logout,
    authToken,
    authError,
    userInfo,
  } = getAuth();

  $: state = {
    isLoading: $isLoading,
    isAuthenticated: $isAuthenticated,
    authError: $authError,
    userInfo: $userInfo ? $userInfo.name : null,
    authToken: $authToken.slice(0, 20),
  };

  const loadData = async (cb) => {
    if ($isLoading && !$isAuthenticated) return;
    getItemTemplates(
      $authToken,
      (itemTemplates) => {
        store.setItemTemplates(itemTemplates);
        if (cb) cb();
      },
      (err) => console.log(err)
    );
  };

  onMount(async () => {
    getItemQualities((q) => (itemQualities = q));
    getItemTypes((t) => (itemTypes = t));
    getItemSubTypes((st) => (itemSubTypes = st));
    getItemSlots((s) => (itemSlots = s));

    loadData(() => {
      selectItemTemplate($store.itemTemplates[0]);
    });
  });

  const newItem = () => {
    let item = {
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
    };

    selectItemTemplate(item);
  };

  const delItemTemplate = async () => {
    deleteItemTemplate(
      $authToken,
      $store.selectedItemTemplate.id,
      () => {
        console.log("delete successful.");
        loadData(() => {
          if ($store.itemTemplates) {
            store.setSelectedItemTemplate($store.itemTemplates[0]);
          }
        });
      },
      () => {
        console.log("create error.");
      }
    );
  };

  const create = async () => {
    createItemTemplate(
      $authToken,
      $store.selectedItemTemplate,
      (itemTemplate) => {
        console.log("create successful.");
        loadData();
        $store.selectedItemTemplate = itemTemplate;
      },
      () => {
        console.log("create error.");
      }
    );
  };

  const selectItemTemplate = (itemTemplate) => {
    store.setSelectedItemTemplate(itemTemplate, () => {
      var elems = document.querySelectorAll("select");
      var instances = M.FormSelect.init(elems, {});

      // second time to fix the selects
      setTimeout(function () {
        var elems = document.querySelectorAll("select");
        var instances = M.FormSelect.init(elems, {});
      }, 50);
    });

    /*
    var targets = document.querySelectorAll(".autocomplete");
    const options = {
      data: {},
      onAutocomplete: function (roomName) {
        console.log(roomName);
      },
    };

    $store.rooms.forEach((value) => {
      options.data[value.name] = null;
    });
    var targetInstances = M.Autocomplete.init(targets, options);*/
  };
  const update = () => {
    updateItemTemplate(
      $authToken,
      $store.selectedItemTemplate.id,
      $store.selectedItemTemplate,
      () => {
        console.log("update successful.");
        loadData();
      },
      () => {
        console.log("update error.");
      }
    );
  };
</script>

<ItemsToolbar create="{newItem}" />

<div class="row">

  <div class="col s3">
    <div class="collection">
      {#each $store.itemTemplates as item}
        <a
          href="#!"
          class="collection-item"
          on:click="{selectItemTemplate(item)}"
        >
          {#if item.slot}
            <span class="new badge" data-badge-caption="">
              {item.quality} {item.subType}
            </span>
          {/if}
          {item.name}
        </a>
      {/each}
    </div>
  </div>

  {#if $store.selectedItemTemplate}
    <div class="col s9">

      <div class="card-panel cyan darken-4">

        <div class="row">

          <span class="header">{$store.selectedItemTemplate.name}</span>

          {#if $store.selectedItemTemplate.isNew}
            <button
              on:click="{() => create()}"
              class="waves-effect waves-light btn-small green right"
            >
              Create
            </button>
          {:else}
            <button
              on:click="{() => update()}"
              class="waves-effect waves-light btn-small green right"
            >
              Update
            </button>
            <button
              on:click="{() => delItemTemplate()}"
              class="waves-effect waves-light btn-small red right"
            >
              Delete
            </button>
          {/if}
        </div>

        <div id="general"></div>

        <div class="row">
          <div class="no-padding input-field col s6">
            <input
              placeholder="Name"
              id="itemTemplate_name"
              type="text"
              bind:value="{$store.selectedItemTemplate.name}"
            />
            <label class="first_label" for="itemTemplate_name">Name</label>
          </div>

          <div class="margininput input-field col s1">
            <select bind:value="{$store.selectedItemTemplate.level}" on:change>
              {#each levels as lvl}
                <option value="{lvl}">{lvl}</option>
              {/each}
            </select>
            <label>Level</label>
          </div>

          {#if $store.selectedItemTemplate.isNew}
            <div class="input-field col s4">
              <input
                placeholder="ID"
                id="itemTemplate_id"
                type="text"
                bind:value="{$store.selectedItemTemplate.id}"
              />
              <label class="active" for="itemTemplate_id">ID</label>
            </div>
          {:else}
            <div class="input-field col s4">
              <input
                placeholder="ID"
                id="itemTemplate_id"
                type="text"
                bind:value="{$store.selectedItemTemplate.id}"
                disabled
              />
              <label class="active" for="itemTemplate_id">ID</label>
            </div>
          {/if}

        </div>

        <div class="row">
          <div class="input-field">
            <input
              placeholder="Item Description"
              id="itemTemplate_description"
              type="text"
              class="materialize-textarea"
              bind:value="{$store.selectedItemTemplate.description}"
            />
            <label class="active" for="itemTemplate_description">
              Description
            </label>
          </div>
        </div>

        <div class="row">
          <div class="input-field">
            <input
              placeholder="Item Details"
              id="itemTemplate_detail"
              type="text"
              class="materialize-textarea"
              bind:value="{$store.selectedItemTemplate.detail}"
            />
            <label class="active" for="itemTemplate_detail">
              Detail (look)
            </label>
          </div>
        </div>

        <div class="row">

          <div class="margininput input-field col s5">
            <select bind:value="{$store.selectedItemTemplate.type}" on:change>
              <option value="" disabled selected>Item Type</option>
              {#each itemTypes as type}
                <option value="{type}">{type.capitalize()}</option>
              {/each}
            </select>
            <label>Select Item Type</label>
          </div>

          <div class="margininput input-field col s5">
            <select
              bind:value="{$store.selectedItemTemplate.subType}"
              on:change
            >
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
            <select
              bind:value="{$store.selectedItemTemplate.quality}"
              on:change
            >
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
              bind:value="{$store.selectedItemTemplate.slot}"
              on:change
            >
              <option value="" selected>Item Slot</option>
              {#each itemSlots as slot}
                <option class="select" value="{slot}">
                  {slot.capitalize()}
                </option>
              {/each}
            </select>
            <label>Select Item Slot</label>
          </div>

        </div>

      </div>
    </div>
  {/if}

</div>
