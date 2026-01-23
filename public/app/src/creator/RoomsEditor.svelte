<script>
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { v4 as uuidv4 } from "uuid";

  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import ExitEditor from "./ExitEditor.svelte";
  import ActionEditor from "./ActionEditor.svelte";
  import SpawnerEditor from "./SpawnerEditor.svelte";
  import RoomItemsModal from "./RoomItemsModal.svelte";

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
  import { getNPCs, getNPC, getUniqueNPCs } from "../api/npcs.js";
  import {
    getNPCSpawnersByRoom,
    createNPCSpawner,
    updateNPCSpawner,
    deleteNPCSpawner,
  } from "../api/npcspawners.js";
  import {
    getItem,
    getItemTemplates,
    createItemFromTemplate,
  } from "../api/items.js";

  const roomsValueHelp = writable([]);
  const scriptsValueHelp = writable([]);
  const npcTemplates = writable([]);
  const roomSpawners = writable([]);
  const itemTemplates = writable([]);
  const roomItems = writable([]);  // Full item objects for current room
  const uniqueNPCs = writable([]);  // All unique NPCs available
  const roomResidents = writable([]);  // Full NPC objects for current room's NPCs
  const store = createStore();
  let hasLoadedScripts = false;
  let hasLoadedRooms = false;
  let hasLoadedNPCs = false;
  let hasLoadedItemTemplates = false;
  let hasLoadedUniqueNPCs = false;

  // Tab state for extensions section
  let activeTab = "exits";

  // Items modal state
  let showItemsModal = false;

  // Residents (unique NPCs) state
  let selectedResidentId = "";

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

  const loadNPCTemplates = () => {
    if (hasLoadedNPCs) return;
    if (!$isAuthenticated || !$authToken) return;

    getNPCs(
      $authToken,
      [],
      (npcs) => {
        // Filter to only templates or all NPCs that could be used as templates
        // For now, include all NPCs since they can all potentially be spawned
        npcTemplates.set(npcs || []);
        hasLoadedNPCs = true;
      },
      (err) => {
        console.log("Failed to load NPCs for Rooms editor:", err);
      }
    );
  };

  const loadItemTemplates = () => {
    if (hasLoadedItemTemplates) return;
    if (!$isAuthenticated || !$authToken) return;

    getItemTemplates(
      $authToken,
      [],
      (templates) => {
        itemTemplates.set(templates || []);
        hasLoadedItemTemplates = true;
      },
      (err) => {
        console.log("Failed to load item templates for Rooms editor:", err);
      }
    );
  };

  const loadUniqueNPCs = () => {
    if (hasLoadedUniqueNPCs) return;
    if (!$isAuthenticated || !$authToken) return;

    getUniqueNPCs(
      $authToken,
      (npcs) => {
        uniqueNPCs.set(npcs || []);
        hasLoadedUniqueNPCs = true;
      },
      (err) => {
        console.log("Failed to load unique NPCs for Rooms editor:", err);
      }
    );
  };

  // Load full NPC objects for NPCs in the current room
  const loadRoomResidents = async (npcIds) => {
    if (!npcIds || npcIds.length === 0) {
      roomResidents.set([]);
      return;
    }
    if (!$isAuthenticated || !$authToken) return;

    const loadedNPCs = [];
    for (const id of npcIds) {
      try {
        await new Promise((resolve, reject) => {
          getNPC(
            $authToken,
            id,
            (npc) => {
              loadedNPCs.push(npc);
              resolve();
            },
            reject
          );
        });
      } catch (err) {
        console.log("Failed to load NPC:", id, err);
      }
    }
    roomResidents.set(loadedNPCs);
  };

  // Load full item objects for items in the current room
  const loadRoomItems = async (itemIds) => {
    if (!itemIds || itemIds.length === 0) {
      roomItems.set([]);
      return;
    }
    if (!$isAuthenticated || !$authToken) return;

    const loadedItems = [];
    for (const id of itemIds) {
      try {
        await new Promise((resolve, reject) => {
          getItem(
            $authToken,
            id,
            (item) => {
              loadedItems.push(item);
              resolve();
            },
            reject
          );
        });
      } catch (err) {
        console.log("Failed to load item:", id, err);
      }
    }
    roomItems.set(loadedItems);
  };

  // Load spawners for the currently selected room
  let currentRoomId = null;
  let pendingSpawners = []; // Spawners being created but not yet saved

  const loadSpawnersForRoom = (roomId) => {
    if (!roomId || !$isAuthenticated || !$authToken) {
      roomSpawners.set([]);
      return;
    }

    getNPCSpawnersByRoom(
      $authToken,
      roomId,
      (spawners) => {
        roomSpawners.set(spawners || []);
        pendingSpawners = [];
      },
      (err) => {
        console.log("Failed to load spawners for room:", err);
        roomSpawners.set([]);
      }
    );
  };

  // Watch for room selection changes
  $: if ($store.selectedElement?.id && $store.selectedElement.id !== currentRoomId) {
    currentRoomId = $store.selectedElement.id;
    if (!$store.selectedElement.isNew) {
      loadSpawnersForRoom(currentRoomId);
      // Load items for this room
      const itemIds = $store.selectedElement.items || [];
      loadRoomItems(itemIds);
      // Load NPCs (residents) for this room
      const npcIds = $store.selectedElement.npcs || [];
      loadRoomResidents(npcIds);
    } else {
      roomSpawners.set([]);
      pendingSpawners = [];
      roomItems.set([]);
      roomResidents.set([]);
    }
  }

  // Spawner CRUD operations
  const createSpawnerDraft = () => {
    const newSpawner = {
      id: uuidv4(),
      name: "",
      templateId: "",
      roomId: $store.selectedElement.id,
      maxInstances: 3,
      spawnInterval: 60000000000, // 60 seconds in nanoseconds
      initialCount: 1,
      respawnTimeOverride: null,
      isNew: true,
    };
    pendingSpawners = [...pendingSpawners, newSpawner];
  };

  const saveSpawner = (spawner) => {
    if (!$isAuthenticated || !$authToken) return;

    const spawnerData = {
      ...spawner,
      roomId: $store.selectedElement.id,
    };
    delete spawnerData.isNew;

    createNPCSpawner(
      $authToken,
      spawnerData,
      (saved) => {
        // Remove from pending and add to saved
        pendingSpawners = pendingSpawners.filter(s => s.id !== spawner.id);
        roomSpawners.update(list => [...list, saved]);
      },
      (err) => {
        console.error("Failed to create spawner:", err);
        alert("Failed to create spawner. Please try again.");
      }
    );
  };

  const updateSpawnerHandler = (spawner) => {
    if (!$isAuthenticated || !$authToken) return;

    updateNPCSpawner(
      $authToken,
      spawner.id,
      spawner,
      (updated) => {
        roomSpawners.update(list =>
          list.map(s => s.id === updated.id ? updated : s)
        );
      },
      (err) => {
        console.error("Failed to update spawner:", err);
        alert("Failed to update spawner. Please try again.");
      }
    );
  };

  const deleteSpawnerHandler = (spawner) => {
    // If it's a pending (unsaved) spawner, just remove from list
    if (spawner.isNew) {
      pendingSpawners = pendingSpawners.filter(s => s.id !== spawner.id);
      return;
    }

    if (!confirm("Are you sure you want to delete this spawner?")) return;
    if (!$isAuthenticated || !$authToken) return;

    deleteNPCSpawner(
      $authToken,
      spawner.id,
      () => {
        roomSpawners.update(list => list.filter(s => s.id !== spawner.id));
      },
      (err) => {
        console.error("Failed to delete spawner:", err);
        alert("Failed to delete spawner. Please try again.");
      }
    );
  };

  // Combine saved and pending spawners for display
  $: allSpawners = [...$roomSpawners, ...pendingSpawners];

  // Load data once auth token becomes available (auth init may complete after onMount).
  $: if ($isAuthenticated && $authToken && !hasLoadedScripts) {
    loadScripts();
  }

  $: if ($isAuthenticated && $authToken && !hasLoadedRooms) {
    loadRoomsValueHelp();
  }

  $: if ($isAuthenticated && $authToken && !hasLoadedNPCs) {
    loadNPCTemplates();
  }

  $: if ($isAuthenticated && $authToken && !hasLoadedItemTemplates) {
    loadItemTemplates();
  }

  $: if ($isAuthenticated && $authToken && !hasLoadedUniqueNPCs) {
    loadUniqueNPCs();
  }

  // Residents (unique NPCs) handlers
  function addResident() {
    if (!selectedResidentId) return;

    // Check if NPC is already in the room
    const currentNpcs = $store.selectedElement.npcs || [];
    if (currentNpcs.includes(selectedResidentId)) {
      alert("This NPC is already assigned to this room.");
      return;
    }

    // Add the NPC to the room
    store.update((state) => {
      if (!state.selectedElement.npcs) {
        state.selectedElement.npcs = [];
      }
      state.selectedElement.npcs = [...state.selectedElement.npcs, selectedResidentId];
      return state;
    });

    // Also add to the loaded residents
    const npc = $uniqueNPCs.find(n => n.id === selectedResidentId);
    if (npc) {
      roomResidents.update(list => [...list, npc]);
    }

    selectedResidentId = "";
  }

  function removeResident(npcId) {
    store.update((state) => {
      state.selectedElement.npcs = (state.selectedElement.npcs || []).filter(id => id !== npcId);
      return state;
    });
    roomResidents.update(list => list.filter(n => n.id !== npcId));
  }

  // Get current room NPC IDs
  $: currentRoomNpcIds = $store.selectedElement?.npcs || [];

  // Filter unique NPCs to only show those not already in the room
  $: availableResidents = $uniqueNPCs.filter(npc => !currentRoomNpcIds.includes(npc.id));

  // Items modal handlers
  function openItemsModal() {
    showItemsModal = true;
  }

  function handleItemsModalClose(event) {
    const updatedItemIds = event.detail;
    // Update the room's items array
    store.update((state) => {
      state.selectedElement.items = updatedItemIds;
      return state;
    });
    showItemsModal = false;
  }

  async function handleCreateItemFromTemplate(event) {
    const templateId = event.detail;
    if (!$isAuthenticated || !$authToken) return;

    createItemFromTemplate(
      $authToken,
      templateId,
      (newItem) => {
        // Add the new item to the room's items array
        store.update((state) => {
          if (!state.selectedElement.items) {
            state.selectedElement.items = [];
          }
          state.selectedElement.items = [...state.selectedElement.items, newItem.id];
          return state;
        });
        // Also add to the loaded items
        roomItems.update(items => [...items, newItem]);
      },
      (err) => {
        console.error("Failed to create item from template:", err);
        alert("Failed to create item. Please try again.");
      }
    );
  }

  // Get current room item IDs
  $: currentRoomItemIds = $store.selectedElement?.items || [];
</script>

<CRUDEditor store={store} config={config}>
  <!-- Hero slot: Map Preview & Coordinates at top right -->
  <div slot="hero" class="hero-map-section">
    <div class="map-coords-container">
      <!-- Mini Map Preview -->
      <div class="p-3 rounded-lg bg-slate-800/50 border border-slate-700">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-bold uppercase text-primary">Map Preview</span>
          <span class="text-[10px] text-slate-400 font-mono">
            ({$store.selectedElement.coords?.x ?? 0}, {$store.selectedElement.coords?.y ?? 0}, {$store.selectedElement.coords?.z ?? 0})
          </span>
        </div>
        <div class="mini-map-grid">
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
        <div class="mt-1 text-[9px] text-slate-500 text-center">
          ● = This Room | ◆ = Has Exit
        </div>
      </div>

      <!-- Coordinates Input -->
      <div class="p-3 rounded-lg bg-slate-800/50 border border-slate-700">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-bold uppercase text-primary">Grid Coordinates</span>
        </div>
        <div class="grid grid-cols-3 gap-2">
          <div class="space-y-1">
            <label class="text-[9px] uppercase text-slate-500" for="coord_x">X</label>
            <input
              id="coord_x"
              type="number"
              class="input-base text-center font-mono text-sm"
              bind:value={$store.selectedElement.coords.x}
            />
          </div>
          <div class="space-y-1">
            <label class="text-[9px] uppercase text-slate-500" for="coord_y">Y</label>
            <input
              id="coord_y"
              type="number"
              class="input-base text-center font-mono text-sm"
              bind:value={$store.selectedElement.coords.y}
            />
          </div>
          <div class="space-y-1">
            <label class="text-[9px] uppercase text-slate-500" for="coord_z">Z</label>
            <input
              id="coord_z"
              type="number"
              class="input-base text-center font-mono text-sm"
              bind:value={$store.selectedElement.coords.z}
            />
          </div>
        </div>
        <p class="mt-2 text-[9px] text-slate-500">
          Z: -1=underground, 0=ground, 1+=above
        </p>
      </div>
    </div>
  </div>

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
  </div>

  <div slot="extensions" class="space-y-4">
    <!-- Tabbed Navigation -->
    <div class="flex items-center gap-1 border-b border-slate-200 dark:border-slate-700">
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === "exits"}
        on:click={() => activeTab = "exits"}
      >
        <span class="material-symbols-outlined text-base">explore</span>
        Exits
        {#if $store.selectedElement.exits?.length}
          <span class="tab-badge">{$store.selectedElement.exits.length}</span>
        {/if}
      </button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === "actions"}
        on:click={() => activeTab = "actions"}
      >
        <span class="material-symbols-outlined text-base">bolt</span>
        Actions
        {#if $store.selectedElement.actions?.length}
          <span class="tab-badge">{$store.selectedElement.actions.length}</span>
        {/if}
      </button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === "spawners"}
        on:click={() => activeTab = "spawners"}
      >
        <span class="material-symbols-outlined text-base">groups</span>
        Spawners
        {#if allSpawners.length}
          <span class="tab-badge">{allSpawners.length}</span>
        {/if}
      </button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === "items"}
        on:click={() => activeTab = "items"}
      >
        <span class="material-symbols-outlined text-base">inventory_2</span>
        Items
        {#if currentRoomItemIds.length}
          <span class="tab-badge">{currentRoomItemIds.length}</span>
        {/if}
      </button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === "residents"}
        on:click={() => activeTab = "residents"}
      >
        <span class="material-symbols-outlined text-base">person</span>
        Residents
        {#if currentRoomNpcIds.length}
          <span class="tab-badge">{currentRoomNpcIds.length}</span>
        {/if}
      </button>
    </div>

    <!-- Tab Content -->
    <div class="tab-content">
      {#if activeTab === "exits"}
        <div class="card p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">explore</span>
              Exits
            </h3>
            <button class="text-xs text-primary hover:underline" type="button" on:click={createExit}>
              + Add Exit
            </button>
          </div>
          <div class="space-y-3">
            {#if $store.selectedElement.exits?.length}
              {#each $store.selectedElement.exits as exit}
                <ExitEditor
                  exit={exit}
                  valueHelp={roomsValueHelp}
                  store={store}
                  deleteExit={deleteExit}
                />
              {/each}
            {:else}
              <p class="text-xs text-slate-500 dark:text-slate-400 italic">No exits configured. Add exits to connect this room to others.</p>
            {/if}
          </div>
        </div>
      {:else if activeTab === "actions"}
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
            {#if $store.selectedElement.actions?.length}
              {#each $store.selectedElement.actions as action}
                <ActionEditor action={action} deleteAction={deleteAction} />
              {/each}
            {:else}
              <p class="text-xs text-slate-500 dark:text-slate-400 italic">No actions configured. Add actions for interactive room elements.</p>
            {/if}
          </div>
        </div>
      {:else if activeTab === "spawners"}
        <div class="card p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">groups</span>
              NPC Spawners
            </h3>
            <button
              class="text-xs text-primary hover:underline"
              type="button"
              on:click={createSpawnerDraft}
              disabled={$store.selectedElement.isNew}
            >
              + Add Spawner
            </button>
          </div>

          {#if $store.selectedElement.isNew}
            <div class="p-4 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-300 text-xs">
              <span class="material-symbols-outlined text-sm align-middle mr-1">info</span>
              Save the room first before adding spawners.
            </div>
          {:else}
            <div class="space-y-3">
              {#if allSpawners.length}
                {#each allSpawners as spawner (spawner.id)}
                  <SpawnerEditor
                    {spawner}
                    npcTemplates={$npcTemplates}
                    isNew={spawner.isNew || false}
                    onSave={saveSpawner}
                    onDelete={deleteSpawnerHandler}
                  />
                {/each}
              {:else}
                <p class="text-xs text-slate-500 dark:text-slate-400 italic">No spawners configured. Add spawners to automatically spawn NPCs in this room.</p>
              {/if}
            </div>
          {/if}
        </div>
      {:else if activeTab === "items"}
        <div class="card p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">inventory_2</span>
              Room Items
            </h3>
            <button
              class="text-xs text-primary hover:underline"
              type="button"
              on:click={openItemsModal}
            >
              Manage Items
            </button>
          </div>

          <div class="space-y-2">
            {#if currentRoomItemIds.length}
              {#each $roomItems as item (item.id)}
                <div class="flex items-center justify-between p-3 rounded-lg bg-slate-800/50 border border-slate-700">
                  <div class="flex-1 min-w-0">
                    <span class="text-sm font-medium text-white">
                      {item.name}
                      {#if item.instanceSuffix}
                        <span class="text-xs text-slate-500 font-mono">-{item.instanceSuffix}</span>
                      {/if}
                    </span>
                    <div class="flex items-center gap-2 mt-0.5">
                      {#if item.type}
                        <span class="text-[10px] text-slate-500 uppercase">{item.type}</span>
                      {/if}
                      {#if item.quality && item.quality !== "normal"}
                        <span class="text-[10px] uppercase" style="color: {item.quality === 'magic' ? '#3b82f6' : item.quality === 'rare' ? '#eab308' : item.quality === 'legendary' ? '#f97316' : item.quality === 'mythic' ? '#a855f7' : '#9ca3af'}">{item.quality}</span>
                      {/if}
                    </div>
                  </div>
                </div>
              {/each}
              {#if currentRoomItemIds.length > $roomItems.length}
                <p class="text-xs text-slate-500 italic">
                  Loading {currentRoomItemIds.length - $roomItems.length} more items...
                </p>
              {/if}
            {:else}
              <p class="text-xs text-slate-500 dark:text-slate-400 italic">No items in this room. Click "Manage Items" to add items.</p>
            {/if}
          </div>
        </div>
      {:else if activeTab === "residents"}
        <div class="card p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">person</span>
              Residents (Unique NPCs)
            </h3>
          </div>

          <p class="text-xs text-slate-500">
            Assign unique NPCs to this room. These are one-of-a-kind NPCs like merchants, quest givers, or unique bosses.
          </p>

          <!-- Add Resident -->
          <div class="flex gap-2">
            <select
              class="input-base flex-1 text-sm"
              bind:value={selectedResidentId}
            >
              <option value="">Select a unique NPC...</option>
              {#each availableResidents as npc}
                <option value={npc.id}>
                  {npc.name} (Lvl {npc.level || 1})
                  {#if npc.class?.name} - {npc.class.name}{/if}
                </option>
              {/each}
            </select>
            <button
              class="px-4 py-2 text-xs font-medium bg-primary text-white rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
              type="button"
              on:click={addResident}
              disabled={!selectedResidentId}
            >
              Add
            </button>
          </div>

          <!-- Current Residents List -->
          <div class="space-y-2">
            {#if currentRoomNpcIds.length}
              {#each $roomResidents as npc (npc.id)}
                <div class="flex items-center justify-between p-3 rounded-lg bg-slate-800/50 border border-slate-700">
                  <div class="flex-1 min-w-0">
                    <span class="text-sm font-medium text-white">
                      {npc.name}
                    </span>
                    <div class="flex items-center gap-2 mt-0.5">
                      <span class="text-[10px] text-slate-500">Lvl {npc.level || 1}</span>
                      {#if npc.race?.name}
                        <span class="text-[10px] text-slate-500">{npc.race.name}</span>
                      {/if}
                      {#if npc.class?.name}
                        <span class="text-[10px] text-slate-500">{npc.class.name}</span>
                      {/if}
                      {#if npc.enemyTrait}
                        <span class="text-[10px] text-red-400 uppercase">Enemy</span>
                      {/if}
                      {#if npc.merchantTrait}
                        <span class="text-[10px] text-green-400 uppercase">Merchant</span>
                      {/if}
                    </div>
                  </div>
                  <button
                    class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-400/10 rounded transition-colors"
                    type="button"
                    on:click={() => removeResident(npc.id)}
                    title="Remove from room"
                  >
                    <span class="material-symbols-outlined text-lg">close</span>
                  </button>
                </div>
              {/each}
              {#if currentRoomNpcIds.length > $roomResidents.length}
                <p class="text-xs text-slate-500 italic">
                  Loading {currentRoomNpcIds.length - $roomResidents.length} more NPCs...
                </p>
              {/if}
            {:else}
              <p class="text-xs text-slate-500 dark:text-slate-400 italic">No unique NPCs assigned to this room. Use the dropdown above to add one.</p>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  </div>
</CRUDEditor>

<RoomItemsModal
  open={showItemsModal}
  itemIds={currentRoomItemIds}
  items={$roomItems}
  itemTemplates={$itemTemplates}
  on:close={handleItemsModalClose}
  on:createFromTemplate={handleCreateItemFromTemplate}
/>

<style>
  /* Hero section for map preview in right column */
  .hero-map-section {
    margin-top: 8px;
  }

  .map-coords-container {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .mini-map-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 2px;
    width: 100%;
    max-width: 90px;
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

  /* Tab Styles */
  .tab-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 16px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #64748b;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-bottom: -1px;
  }

  .tab-btn:hover {
    color: #94a3b8;
  }

  .tab-btn.active {
    color: #00bcd4;
    border-bottom-color: #00bcd4;
  }

  .tab-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    font-size: 10px;
    font-weight: 700;
    background: rgba(100, 116, 139, 0.3);
    border-radius: 9px;
  }

  .tab-btn.active .tab-badge {
    background: rgba(0, 188, 212, 0.2);
    color: #00bcd4;
  }

  .tab-content {
    min-height: 200px;
  }
</style>
