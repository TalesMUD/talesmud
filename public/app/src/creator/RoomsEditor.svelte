<style>

</style>

<script>
  import ExitEditor from "./ExitEditor.svelte";
  import Toolbar from "./Toolbar.svelte";
  import { onMount } from "svelte";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { v4 as uuidv4 } from "uuid";
  import ActionEditor from "./ActionEditor.svelte";

  import {
    getRoom,
    deleteRoom,
    getRooms,
    updateRoom,
    createRoom,
  } from "../api/rooms.js";

  const config = {
    title: "Manage Rooms",
    actions: [
      {
        icon: "add",
        name: "Create Exit",
        color: "blue",
        fnc: () => createExit(),
      },
    ],
    get: getRooms,
    getElement: getRoom,
    create: createRoom,
    update: updateRoom,
    delete: deleteRoom,
    refreshUI: () => {
      var elems = document.querySelectorAll("select");
      var instances = M.FormSelect.init(elems, {});

      // second time to fix the selects
      setTimeout(function () {
        var elems = document.querySelectorAll("select");
        var instances = M.FormSelect.init(elems, {});

        M.updateTextFields();
        var elems2 = document.querySelectorAll(".collapsible");
        if (elems2 != undefined) {
          var instances = M.Collapsible.init(elems2, {});
        }

        var textareas = document.querySelectorAll(".materialize-textarea");
        textareas.forEach((e) => {
          M.textareaAutoResize(e);
        });
      }, 50);
    },

    new: (select) => {
      select({
        name: "New Room",
        description: "",
        detail: "",
        areaType: "",
        area: "",
        id: uuidv4(),
        isNew: true,
        exits: [],
        actions: [],
      });
    },

    badge: (element) => {
      return element.area;
    },
  };
  // create store outside of the component to use it in the slot..
  const store = createStore();

  const deleteExit = (exit) => {
    store.update((state) => {
      state.selectedElement.exits = state.selectedElement.exits.filter(
        (x) => x.name != exit.name
      );
      return state;
    });
  };

  const createExit = () => {
    store.update((state) => {
      if (state.selectedElement.exits == null) {
        state.selectedElement.exits = [];
      }

      state.selectedElement.exits.push({
        name: "New Exit",
        description: "todo",
        target: "select target",
      });
      return state;
    });
    config.refreshUI();
  };

  const deleteAction = (exit) => {
    store.update((state) => {
      state.selectedElement.actions = state.selectedElement.actions.filter(
        (x) => x.name != exit.name
      );
      return state;
    });
  };

  const createAction = () => {
    store.update((state) => {
      if (state.selectedElement.actions == null) {
        state.selectedElement.actions = [];
      }

      state.selectedElement.actions.push({
        name: "New Action",
        description: "A Description",
        response: "Response",
        type: "room_response",
      });
      return state;
    });
    config.refreshUI();
  };
  onMount(async () => {});
  /////////

  const exitsToolbar = {
    title: "Exits",
    small: true,
    actions: [
      {
        icon: "add",
        fnc: () => {
          createExit();
        },
      },
    ],
  };
  const actionsToolbar = {
    title: "Actions",
    small: true,
    actions: [
      {
        icon: "add",
        fnc: () => {
          createAction();
        },
      },
    ],
  };
</script>

<CRUDEditor store="{store}" config="{config}">

  <div slot="content">

    <div class="row">

      <div class="no-padding input-field col s4">
        <input
          placeholder="Area"
          id="area"
          type="text"
          bind:value="{$store.selectedElement.area}"
        />
        <label class="active first_label" for="area">Area</label>
      </div>

      <div class="input-field col s4">
        <input
          placeholder="Area Type"
          id="area_type"
          type="text"
          bind:value="{$store.selectedElement.areaType}"
        />
        <label class="active" for="area_type">Area Type</label>
      </div>

      <div class="input-field col s4">
        <input
          placeholder="Room Type"
          id="room_type"
          type="text"
          bind:value="{$store.selectedElement.roomType}"
        />
        <label class="active" for="room_type">Room Type</label>
      </div>
    </div>

    {#if $store.selectedElement.coords}
      <div class="row">

        <div class="no-padding input-field col s2">
          <input
            placeholder="X"
            id="x"
            type="text"
            bind:value="{$store.selectedElement.coords.x}"
          />
          <label class="first_label" for="x">X</label>
        </div>

        <div class="input-field col s2">
          <input
            placeholder="Y"
            id="y"
            type="text"
            bind:value="{$store.selectedElement.coords.y}"
          />
          <label class="active" for="y">Y</label>
        </div>

        <div class="input-field col s2">
          <input
            placeholder="Z"
            id="z"
            type="text"
            bind:value="{$store.selectedElement.coords.z}"
          />
          <label class="active" for="z">Z</label>
        </div>
      </div>
    {/if}
  </div>

  <div slot="extensions">

    <Toolbar toolbar="{exitsToolbar}" />

    {#if $store.selectedElement.exits}
      <ul
        class="card-panel blue-grey darken-3 collapsible"
        style="padding: 0; border: none;"
      >
        {#each $store.selectedElement.exits as exit}
          <ExitEditor exit="{exit}" deleteExit="{deleteExit}" />
        {/each}
      </ul>
    {/if}

    <Toolbar toolbar="{actionsToolbar}" />

    {#if $store.selectedElement.actions}
      <ul
        class="card-panel blue-grey darken-3 collapsible"
        style="padding: 0; border: none;"
      >
        {#each $store.selectedElement.actions as action}
          <ActionEditor action="{action}" deleteAction="{deleteAction}" />
        {/each}
      </ul>
    {/if}
  </div>
</CRUDEditor>
