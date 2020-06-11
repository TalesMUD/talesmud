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
  import ActionEditor from "./ActionEditor.svelte";
  import RoomsToolbar from "./RoomsToolbar.svelte";

  import { store } from "./RoomsEditorStore.js";
  import { PlusIcon } from "svelte-feather-icons";
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { createAuth, getAuth } from "../auth.js";
  import { v4 as uuidv4 } from "uuid";

  import axios from "axios";
  import {
    getRoom,
    deleteRoom,
    getRooms,
    updateRoom,
    createRoom,
  } from "../api/rooms.js";

  export let location;

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
    getRooms(
      $authToken,
      (rooms) => {
        store.setRooms(rooms);
        if (cb) cb();
      },
      (err) => console.log(err)
    );
  };

  onMount(async () => {
    document.addEventListener("DOMContentLoaded", function () {
      var elems = document.querySelectorAll(".collapsible");
      var instances = M.Collapsible.init(elems);
    });

    loadData(() => {
      store.setSelectedRoom($store.rooms[0], () => {
        var elems = document.querySelectorAll(".collapsible");
        var instances = M.Collapsible.init(elems);
      });
    });
  });

  const newRoom = () => {
    let newRoom = {
      name: "New Room",
      description: "",
      detail: "",
      areaType: "",
      area: "",
      id: uuidv4(),
      isNew: true,
      exits: [],
      actions: [],
    };

    selectRoom(newRoom);
  };

  const delRoom = async (room) => {
    deleteRoom(
      $authToken,
      $store.selectedRoom.id,
      () => {
        console.log("delete successful.");
        loadData(() => {
          store.setSelectedRoom($store.rooms[0]);
        });
      },
      () => {
        console.log("create error.");
      }
    );
  };

  const create = async () => {
    createRoom(
      $authToken,
      $store.selectedRoom,
      (room) => {
        console.log("create successful.");
        loadData();
        $store.selectedRoom = room;
      },
      () => {
        console.log("create error.");
      }
    );
  };
  const deleteExit = (exit) => {
    store.deleteExit(exit);
  };
  const createExit = () => {
    store.createExit();
    M.updateTextFields();
    var elems = document.querySelectorAll(".collapsible");
    if (elems != undefined) {
      var instances = M.Collapsible.init(elems, {});
    }
  };
  const selectRoom = (room) => {
    store.setSelectedRoom(room, () => {
      M.updateTextFields();
      var elems = document.querySelectorAll(".collapsible");
      if (elems != undefined) {
        var instances = M.Collapsible.init(elems, {});
      }
      M.updateTextFields();

      var el = document.querySelectorAll(".tabs");

      var instance = M.Tabs.init(el, {});
    });

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
    var targetInstances = M.Autocomplete.init(targets, options);
  };
  const update = () => {
    updateRoom(
      $authToken,
      $store.selectedRoom.id,
      $store.selectedRoom,
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

<RoomsToolbar createRoom="{newRoom}" addExit="{createExit}" />

<div class="row">

  <div class="col s3">
    <div class="collection">
      {#each $store.rooms as room}
        <a href="#!" class="collection-item" on:click="{selectRoom(room)}">
          {#if room.area}
            <span class="new badge" data-badge-caption="">{room.area}</span>
          {/if}
          {room.name}
        </a>
      {/each}
    </div>
  </div>

  {#if $store.selectedRoom}
    <div class="col s9">

      <div class="card-panel cyan darken-4">

        <div class="row">

          <span class="header">{$store.selectedRoom.name}</span>

          {#if $store.selectedRoom.isNew}
            <button
              on:click="{() => create()}"
              class="waves-effect waves-light btn-small green"
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
              on:click="{() => delRoom()}"
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
              id="room_name"
              type="text"
              bind:value="{$store.selectedRoom.name}"
            />
            <label class="first_label" for="room_name">Name</label>
          </div>

          {#if $store.selectedRoom.isNew}
            <div class="input-field col s6">
              <input
                placeholder="ID"
                id="room_id"
                type="text"
                bind:value="{$store.selectedRoom.id}"
              />
              <label class="active" for="room_id">ID</label>
            </div>
          {:else}
            <div class="input-field col s6">
              <input
                placeholder="ID"
                id="room_id"
                type="text"
                bind:value="{$store.selectedRoom.id}"
                disabled
              />
              <label class="active" for="room_id">ID</label>
            </div>
          {/if}
        </div>

        <div class="row">
          <div class="input-field">
            <textarea
              placeholder="Room Description"
              id="room_description"
              rows="8"
              class="materialize-textarea"
              bind:value="{$store.selectedRoom.description}"
            ></textarea>
            <label class="active" for="room_description">Description</label>
          </div>
        </div>

        <div class="row">
          <div class="input-field">
            <textarea
              placeholder="Room Details"
              id="room_detail"
              rows="4"
              class="materialize-textarea"
              bind:value="{$store.selectedRoom.detail}"
            ></textarea>
            <label class="active" for="room_detail">Detail (look)</label>
          </div>
        </div>

        <div class="row">

          <div class="no-padding input-field col s4">
            <input
              placeholder="Area"
              id="area"
              type="text"
              bind:value="{$store.selectedRoom.area}"
            />
            <label class="active first_label" for="area">Area</label>
          </div>

          <div class="input-field col s4">
            <input
              placeholder="Area Type"
              id="area_type"
              type="text"
              bind:value="{$store.selectedRoom.areaType}"
            />
            <label class="active" for="area_type">Area Type</label>
          </div>

          <div class="input-field col s4">
            <input
              placeholder="Room Type"
              id="room_type"
              type="text"
              bind:value="{$store.selectedRoom.roomType}"
            />
            <label class="active" for="room_type">Room Type</label>
          </div>
        </div>

        {#if $store.selectedRoom.coords}
          <div class="row">

            <div class="no-padding input-field col s2">
              <input
                placeholder="X"
                id="x"
                type="text"
                bind:value="{$store.selectedRoom.coords.x}"
              />
              <label class="first_label" for="x">X</label>
            </div>

            <div class="input-field col s2">
              <input
                placeholder="Y"
                id="y"
                type="text"
                bind:value="{$store.selectedRoom.coords.y}"
              />
              <label class="active" for="y">Y</label>
            </div>

            <div class="input-field col s2">
              <input
                placeholder="Z"
                id="z"
                type="text"
                bind:value="{$store.selectedRoom.coords.z}"
              />
              <label class="active" for="z">Z</label>
            </div>
          </div>
        {/if}

      </div>
      {#if $store.selectedRoom.exits}
        <h6>Exits</h6>
        <ul class="collapsible">
          {#each $store.selectedRoom.exits as exit}
            <ActionEditor exit="{exit}" deleteExit="{deleteExit}" />
          {/each}
        </ul>
      {/if}
    </div>
  {/if}
</div>
