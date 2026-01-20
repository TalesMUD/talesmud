<script>
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { v4 as uuidv4 } from "uuid";

  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import ExitEditor from "./ExitEditor.svelte";
  import ActionEditor from "./ActionEditor.svelte";

  import { getAuth } from "../auth.js";
  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  import {
    getRoom,
    deleteRoom,
    getRoomsValueHelp,
    getRooms,
    updateRoom,
    createRoom,
  } from "../api/rooms.js";
  import { getScripts } from "../api/scripts.js";

  const roomsValueHelp = writable([]);
  const scriptsValueHelp = writable([]);
  const store = createStore();
  let hasLoadedScripts = false;
  let hasLoadedRooms = false;

  const config = {
    title: "Manage Rooms",
    subtitle: "Configure environment, exits, and NPC populations.",
    listTitle: "Rooms",
    labels: {
      create: "Create Room",
      update: "Update Room",
      delete: "Delete",
    },
    get: getRooms,
    getElement: getRoom,
    create: createRoom,
    update: updateRoom,
    delete: deleteRoom,
    beforeSelect: (element) => {
      if (element.meta === undefined) {
        element.meta = {
          background: "",
        };
      }
      if (element.onEnterScriptID === undefined) {
        element.onEnterScriptID = "";
      }
      // Ensure coords object exists for editing
      if (element.coords === undefined || element.coords === null) {
        element.coords = { x: 0, y: 0, z: 0 };
      }
    },
    badge: (element) => element.area,
  };

  const createNewRoom = () => {
    config.new((element) => {
      if (config.beforeSelect) {
        config.beforeSelect(element);
      }
      store.setSelectedElement(element);
    });
  };

  config.extraActions = [
    {
      label: "Create Room",
      icon: "add_box",
      variant: "btn-outline",
      onClick: createNewRoom,
    },
  ];

  config.new = (select) => {
    select({
      name: "New Room",
      description: "",
      detail: "",
      areaType: "",
      area: "",
      roomType: "",
      id: uuidv4(),
      isNew: true,
      exits: [],
      actions: [],
      coords: { x: 0, y: 0, z: 0 },
      meta: {
        background: "",
      },
    });
  };

  const deleteExit = (exit) => {
    store.update((state) => {
      state.selectedElement.exits = state.selectedElement.exits.filter(
        (x) => x.name !== exit.name
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
  };

  const deleteAction = (action) => {
    store.update((state) => {
      state.selectedElement.actions = state.selectedElement.actions.filter(
        (x) => x.name !== action.name
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
        params: new Map(),
      });
      return state;
    });
  };

  const loadRoomsValueHelp = () => {
    if (hasLoadedRooms) return;
    if (!$isAuthenticated || !$authToken) return;

    getRoomsValueHelp(
      $authToken,
      (roomsvh) => {
        console.log("Loaded rooms value help:", roomsvh?.length || 0, "rooms");
        roomsValueHelp.set(roomsvh || []);
        hasLoadedRooms = true;
      },
      (err) => {
        console.log("Failed to load rooms value help:", err);
      }
    );
  };

  onMount(() => {
    loadRoomsValueHelp();
  });

  const loadScripts = () => {
    if (hasLoadedScripts) return;
    if (!$isAuthenticated || !$authToken) return;

    getScripts(
      $authToken,
      [],
      (scripts) => {
        scriptsValueHelp.set(scripts || []);
        hasLoadedScripts = true;
      },
      (err) => {
        console.log("Failed to load scripts for Rooms editor:", err);
      }
    );
  };

  // Load data once auth token becomes available (auth init may complete after onMount).
  $: if ($isAuthenticated && $authToken && !hasLoadedScripts) {
    loadScripts();
  }

  $: if ($isAuthenticated && $authToken && !hasLoadedRooms) {
    loadRoomsValueHelp();
  }
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="area">Area</label>
        <input
          id="area"
          type="text"
          class="input-base"
          bind:value={$store.selectedElement.area}
        />
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="area_type">Area Type</label>
        <input
          id="area_type"
          type="text"
          class="input-base"
          bind:value={$store.selectedElement.areaType}
        />
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="room_type">Room Type</label>
        <input
          id="room_type"
          type="text"
          class="input-base"
          bind:value={$store.selectedElement.roomType}
        />
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-1.5">
        <label class="label-caps" for="room_background">Background Image ID</label>
        <input
          id="room_background"
          type="text"
          class="input-base"
          bind:value={$store.selectedElement.meta.background}
        />
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="room_on_enter_script">On Enter Script</label>
        <select
          id="room_on_enter_script"
          class="input-base"
          bind:value={$store.selectedElement.onEnterScriptID}
        >
          <option value="">None</option>
          {#if $scriptsValueHelp}
            {#each $scriptsValueHelp as script}
              <option value={script.id}>{script.name} ({script.id})</option>
            {/each}
          {/if}
        </select>
        <p class="text-[10px] text-slate-400">
          Runs when a player enters this room (e.g., walking in or selecting a character).
        </p>
      </div>
    </div>

    <!-- Map Coordinates & Preview Section -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Mini Map Preview -->
      <div class="p-4 rounded-lg bg-slate-800/50 border border-slate-700">
        <div class="flex items-center justify-between mb-3">
          <span class="text-[10px] font-bold uppercase text-primary">Map Preview</span>
          <span class="text-[10px] text-slate-400 font-mono">
            ({$store.selectedElement.coords?.x ?? 0}, {$store.selectedElement.coords?.y ?? 0}, {$store.selectedElement.coords?.z ?? 0})
          </span>
        </div>
        <div class="mini-map-grid">
          <!-- 3x3 grid showing current room and adjacent positions -->
          {#each [-1, 0, 1] as dy}
            {#each [-1, 0, 1] as dx}
              {@const isCenter = dx === 0 && dy === 0}
              {@const exitDir = dx === -1 ? 'west' : dx === 1 ? 'east' : dy === -1 ? 'north' : dy === 1 ? 'south' : null}
              {@const hasExit = exitDir && $store.selectedElement.exits?.some(e => e.name?.toLowerCase() === exitDir)}
              {@const isDiagonal = dx !== 0 && dy !== 0}
              <div
                class="mini-map-cell"
                class:current={isCenter}
                class:has-exit={hasExit && !isDiagonal}
                class:diagonal={isDiagonal}
              >
                {#if isCenter}
                  <span class="text-[10px] font-bold text-primary">●</span>
                {:else if hasExit && !isDiagonal}
                  <span class="text-[8px] text-cyan-400">◆</span>
                {:else if !isDiagonal}
                  <span class="text-[8px] text-slate-600">·</span>
                {/if}
              </div>
            {/each}
          {/each}
        </div>
        <div class="mt-2 text-[9px] text-slate-500 text-center">
          ● = This Room | ◆ = Has Exit
        </div>
      </div>

      <!-- Coordinates Input -->
      <div class="p-4 rounded-lg bg-slate-800/50 border border-slate-700">
        <div class="flex items-center justify-between mb-3">
          <span class="text-[10px] font-bold uppercase text-primary">Grid Coordinates</span>
          <span class="text-[9px] text-slate-500">Z: -1=underground, 0=ground, 1+=above</span>
        </div>
        <div class="grid grid-cols-3 gap-3">
          <div class="space-y-1.5">
            <label class="label-caps" for="coord_x">X (East/West)</label>
            <input
              id="coord_x"
              type="number"
              class="input-base text-center font-mono"
              bind:value={$store.selectedElement.coords.x}
            />
          </div>
          <div class="space-y-1.5">
            <label class="label-caps" for="coord_y">Y (North/South)</label>
            <input
              id="coord_y"
              type="number"
              class="input-base text-center font-mono"
              bind:value={$store.selectedElement.coords.y}
            />
          </div>
          <div class="space-y-1.5">
            <label class="label-caps" for="coord_z">Z (Level)</label>
            <input
              id="coord_z"
              type="number"
              class="input-base text-center font-mono"
              bind:value={$store.selectedElement.coords.z}
            />
          </div>
        </div>
        <p class="mt-2 text-[9px] text-slate-500">
          Set coordinates to place this room on the world map grid.
          Rooms without coordinates won't appear on the visual map.
        </p>
      </div>
    </div>
  </div>

  <div slot="extensions" class="space-y-6">
    <div class="card p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
          <span class="material-symbols-outlined text-lg">explore</span>
          Exits
        </h3>
        <button class="text-xs text-primary hover:underline" type="button" on:click={createExit}>
          + Link Manual
        </button>
      </div>
      <div class="space-y-3">
        {#if $store.selectedElement.exits}
          {#each $store.selectedElement.exits as exit}
            <ExitEditor
              exit={exit}
              valueHelp={roomsValueHelp}
              store={store}
              deleteExit={deleteExit}
            />
          {/each}
        {/if}
      </div>
    </div>

    <div class="card p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
          <span class="material-symbols-outlined text-lg">bolt</span>
          Actions
        </h3>
        <button class="text-xs text-primary hover:underline" type="button" on:click={createAction}>
          + Add Action
        </button>
      </div>
      <div class="space-y-3">
        {#if $store.selectedElement.actions}
          {#each $store.selectedElement.actions as action}
            <ActionEditor action={action} deleteAction={deleteAction} />
          {/each}
        {/if}
      </div>
    </div>
  </div>
</CRUDEditor>

<style>
  .mini-map-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 2px;
    width: 100%;
    max-width: 120px;
    margin: 0 auto;
    aspect-ratio: 1;
  }

  .mini-map-cell {
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(30, 41, 59, 0.5);
    border: 1px solid rgba(51, 65, 85, 0.5);
    border-radius: 2px;
    aspect-ratio: 1;
  }

  .mini-map-cell.current {
    background: rgba(0, 188, 212, 0.2);
    border-color: rgba(0, 188, 212, 0.5);
  }

  .mini-map-cell.has-exit {
    background: rgba(0, 188, 212, 0.1);
    border-color: rgba(0, 188, 212, 0.3);
  }

  .mini-map-cell.diagonal {
    background: transparent;
    border-color: transparent;
  }
</style>
