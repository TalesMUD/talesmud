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
  import Toolbar from "./Toolbar.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { PlusIcon } from "svelte-feather-icons";
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { createAuth, getAuth } from "../auth.js";

  export let config;
  export let store;

  const { isAuthenticated, authToken } = getAuth();

  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  const loadData = async (cb) => {
    if (!$isAuthenticated) return;
    config.get(
      $authToken,
      (all) => {
        store.setElements(all);
        if (cb) cb();
      },
      (err) => console.log(err)
    );
  };

  onMount(async () => {
    loadData(() => {
      selectElement($store.elements[0]);
    });
  });

  const deleteElement = async () => {
    config.delete(
      $authToken,
      $store.selectedElement.id,
      () => {
        loadData(() => {
          if ($store.elements) {
            store.setSelectedElement($store.elements[0]);
          }
        });
      },
      () => {
        console.log("create error.");
      }
    );
  };

  const create = async () => {
    config.create(
      $authToken,
      $store.selectedElement,
      (element) => {
        loadData();
        selectElement(element);
      },
      () => {
        console.log("create error.");
      }
    );
  };

  const selectElement = (element) => {
    store.setSelectedElement(element, () => {
      refreshUI();
    });
  };

  const refreshUI = () => {
    var elems = document.querySelectorAll("select");
    var instances = M.FormSelect.init(elems, {});
    // second time to fix the selects
    setTimeout(function () {
      var elems = document.querySelectorAll("select");
      var instances = M.FormSelect.init(elems, {});
    }, 50);

    if (config.refreshUI) {
      config.refreshUI();
    }
  };

  const update = () => {
    config.udate(
      $authToken,
      $store.selectedElement.id,
      $store.selectedElement,
      () => {
        console.log("update successful.");
        loadData();
      },
      () => {
        console.log("update error.");
      }
    );
  };

  const toolbarConfig = {
    title: config.title,
    actions: [
      {
        name: "Create Item Template",
        color: "green",
        fnc: () => config.new(selectElement),
      },
      ...config.actions, // add extra ctions
    ],
  };
</script>

<Toolbar toolbar="{toolbarConfig}" />

<div class="row">

  <!-- START: ELEMENT LIST (Master)-->
  <div class="col s3">
    <div class="collection">
      {#each $store.elements as element}
        <a
          href="#!"
          class="collection-item"
          on:click="{selectElement(element)}"
        >
          {#if element.slot}
            <span class="new badge" data-badge-caption="">
              {config.badge(element)}
            </span>
          {/if}
          {element.name}
        </a>
      {/each}
    </div>
  </div>
  <!-- END: ELEMENT LIST -->

  <!-- START: OBJECT PAGE (Detail)-->
  {#if $store.selectedElement}
    <div class="col s9">
      <div class="card-panel cyan darken-4">
        <div class="row">

          <span class="header">{$store.selectedElement.name}</span>

          {#if $store.selectedElement.isNew}
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
              on:click="{() => deleteElement()}"
              class="waves-effect waves-light btn-small red right"
            >
              Delete
            </button>
          {/if}
        </div>

        <div class="row">
          <div class="no-padding input-field col s6">
            <input
              placeholder="Name"
              id="itemTemplate_name"
              type="text"
              bind:value="{$store.selectedElement.name}"
            />
            <label class="first_label" for="itemTemplate_name">Name</label>
          </div>

          {#if $store.selectedElement.isNew}
            <div class="input-field col s4">
              <input
                placeholder="ID"
                id="itemTemplate_id"
                type="text"
                bind:value="{$store.selectedElement.id}"
              />
              <label class="active" for="itemTemplate_id">ID</label>
            </div>
          {:else}
            <div class="input-field col s4">
              <input
                placeholder="ID"
                id="itemTemplate_id"
                type="text"
                bind:value="{$store.selectedElement.id}"
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
              bind:value="{$store.selectedElement.description}"
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
              bind:value="{$store.selectedElement.detail}"
            />
            <label class="active" for="itemTemplate_detail">
              Detail (look)
            </label>
          </div>
        </div>

        <slot element="{$store.selectedElement}" />

      </div>
    </div>
  {/if}

</div>
