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
  .title {
    font-size: 2em;
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
  label {
    color: #eee;
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
      isNew: true,
      exits: [],
      actions: [],
    };

    createRoom(
      $authToken,
      newRoom,
      (room) => {
        loadData();
        store.setSelectedRoom(room, () => {
          var elems = document.querySelectorAll(".collapsible");
          var instances = M.Collapsible.init(elems);
        });
      },
      (err) => console.log(err)
    );
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
          {room.name}
        </a>
      {/each}
    </div>
  </div>

  {#if $store.selectedRoom}
    <div>

      <div class="col s12 m8">
        <div class="card-panel cyan darken-4">

          <input
            placeholder="Room Name"
            id="room_name"
            type="text"
            bind:value="{$store.selectedRoom.name}"
            class="col s8"
          />

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

          <input
            placeholder="Room ID"
            id="room_id"
            type="text"
            value="{$store.selectedRoom.id}"
            disabled
          />

          <textarea
            placeholder="Room Description"
            id="room_description"
            type="text"
            class="materialize-textarea"
            bind:value="{$store.selectedRoom.description}"
          ></textarea>

        </div>

        {#if $store.selectedRoom.exits}
          <h6>Exits</h6>
          <ul class="collapsible popout">
            {#each $store.selectedRoom.exits as exit}
              <ActionEditor exit="{exit}" deleteExit="{deleteExit}" />
            {/each}
          </ul>
        {/if}
      </div>
    </div>
  {/if}
</div>
