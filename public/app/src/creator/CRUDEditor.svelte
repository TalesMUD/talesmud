<style>
  .sidelist {
    width: 20em;
  }
  textarea {
    color: white;
    margin-top: 1em;
  }
  .search {
    color: white !important;
    padding: 0;
    margin: 0;
  }
  .not_stored {
    border: 1px dashed orange;
  }
  .card-panel {
    padding-bottom: 0;
    margin-bottom: 0;
  }
  .newbadge {
    background-color: #26a69a;
    border-radius: 5px;
    color: #d8dee9;
    padding: 3px;
    width: auto;
    text-overflow: ellipsis;
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
  import ActionEditor from "./ActionEditor.svelte";
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
      $store.filters,
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

  const isDraft = (isnew) => {
    console.log("re.selectedElement.isNew" + isnew);
    if (isnew === true) return "not_stored";
    return "";
  };

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
    if (config.beforeSelect) {
      config.beforeSelect(element);
    }

    store.setSelectedElement(element, () => {
      refreshUI();
    });
  };

  const addFilter = (filter) => {
    if (filter.includes(":")) {
      let keyval = filter.split(":");
      store.addFilter(keyval[0], keyval[1]);

      loadData(() => {
        selectElement($store.elements[0]);
      });
    }
  };

  const removeFilter = (filter) => {
    if (filter.includes(":")) {
      let keyval = filter.split(":");
      store.removeFilter(keyval[0]);

      loadData(() => {
        selectElement($store.elements[0]);
      });
    }
  };

  const refreshUI = () => {
    setTimeout(function () {
      let chips = document.querySelectorAll(".chips");
      let filters = [];

      // restore current filters
      $store.filters.forEach((f) => {
        filters.push({
          tag: f.key + ":" + f.val,
        });
      });

      M.Chips.init(chips, {
        data: filters,
        onChipAdd: (ev, chip) => addFilter(chip.firstChild.nodeValue),
        onChipDelete: (ev, chip) => removeFilter(chip.firstChild.nodeValue),
      });

      var elems = document.querySelectorAll("select");
      M.FormSelect.init(elems, {});

      M.updateTextFields();
      var elems2 = document.querySelectorAll(".collapsible");
      if (elems2 != undefined) {
        var instances = M.Collapsible.init(elems2, {});
      }

      var textareas = document.querySelectorAll(".materialize-textarea");
      textareas.forEach((e) => {
        M.textareaAutoResize(e);
      });

      if (config.refreshUI) {
        config.refreshUI();
      }
    }, 50);
  };

  const update = () => {
    config.update(
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
        name: null,
        color: "",
        icon: "filter_alt",
        fnc: () => {
          store.toggleFilter();
          refreshUI();
        },
      },
      {
        name: null,
        icon: "add",
        color: "",
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
          href="javascript:void(0)"
          class="collection-item blue-grey lighten-5"
          on:click="{selectElement(element)}"
        >
          {#if element && config.badge}
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

  <div class="col s9">

    {#if $store.filterActive}
      <div class="card-panel blue-grey lighten-5" style="padding:15px;">
        <div class="chips chips-placeholder search"></div>
      </div>
    {/if}

    <!-- START: OBJECT PAGE (Detail)-->
    {#if $store.selectedElement}
      <div
        class="card-panel blue-grey darken-3 {isDraft($store.selectedElement.isNew)}"
      >
        {#if $store.selectedElement.isNew}
          <div class="row center-align" style="padding:0; margin-top: -2.6em;">
            <div
              class="chip"
              style="color: #121212; margin: 0; background-color: orange;
              text-align:center;"
            >
              Not Stored
            </div>
          </div>
        {/if}

        <div class="row">
          <slot name="hero" />
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
            <textarea
              placeholder="Description"
              id="itemTemplate_description"
              type="text"
              class="materialize-textarea"
              bind:value="{$store.selectedElement.description}"
            ></textarea>
            <label class="active" for="itemTemplate_description">
              Description
            </label>
          </div>
        </div>

        {#if config.hideDetails == undefined || !config.hideDetails}
          <div class="row">
            <div class="input-field">
              <textarea
                placeholder="Details"
                id="itemTemplate_detail"
                type="text"
                class="materialize-textarea"
                bind:value="{$store.selectedElement.detail}"
              ></textarea>
              <label class="active" for="itemTemplate_detail">
                Detail (look)
              </label>
            </div>
          </div>
        {/if}

        <slot name="content" />
      </div>

      <slot name="extensions" />
    {/if}
  </div>

</div>
