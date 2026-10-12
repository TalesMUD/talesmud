<script>
  import { v4 as uuidv4 } from "uuid";
  import { navigateTo } from "yrv";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import { getAuth } from "../auth.js";
  import { getRooms } from "../api/rooms.js";
  import { getNPCs } from "../api/npcs.js";
  import { getLiveNPCs } from "../api/live.js";
  import {
    getNPCSpawner,
    getNPCSpawners,
    createNPCSpawner,
    updateNPCSpawner,
    deleteNPCSpawner,
  } from "../api/npcspawners.js";
  import { npcColumns, roomColumns, spawnerColumns } from "./tableColumns.js";
  import {
    decorateSpawner,
    filterSpawners,
    filterUniqueGaps,
    uniqueTemplatesWithoutSpawner,
  } from "./spawnerList.js";

  const { isAuthenticated, authToken } = getAuth();
  const store = createStore();
  const columns = spawnerColumns.map((column) => ({ ...column }));

  let cache = { rooms: [], npcs: [], spawners: [], live: null, liveError: "" };
  let templateId = "";
  let roomId = "";
  let zone = "";
  let uniqueGap = false;
  let intervalSeconds = 60;
  let overrideSeconds = 60;
  let useOverride = false;
  let delaySeconds = 0;

  function indexById(rows) {
    const map = {};
    for (const row of rows || []) {
      if (row?.id) map[row.id] = row;
    }
    return map;
  }

  function contentNPCs(rows) {
    return (rows || []).filter((npc) => npc && !npc.templateId && !String(npc.id || "").includes("~"));
  }

  function decoratedRows() {
    const roomsById = indexById(cache.rooms);
    const npcsById = indexById(cache.npcs);
    return (cache.spawners || []).map((spawner) => decorateSpawner(spawner, {
      roomsById,
      npcsById,
      live: cache.live,
    }));
  }

  function tableRows() {
    if (uniqueGap) return [];
    return filterSpawners(decoratedRows(), { templateId, roomId, zone });
  }

  function zones() {
    return [...new Set((cache.rooms || []).map((room) => room?.area).filter(Boolean))].sort();
  }

  function loadAll(token) {
    const listed = (fn) => new Promise((resolve, reject) => fn(token, [], resolve, reject));
    return Promise.all([
      listed(getRooms),
      listed(getNPCs),
      listed(getNPCSpawners),
      getLiveNPCs(token)
        .then((rows) => ({ rows: rows || [], error: "" }))
        .catch(() => ({ rows: null, error: "Live NPC count is unavailable." })),
    ]).then(([rooms, npcs, spawners, live]) => ({
      rooms: rooms || [],
      npcs: npcs || [],
      spawners: spawners || [],
      live: live.rows,
      liveError: live.error,
    }));
  }

  function remember(next) {
    cache = next;
    const zoneCol = columns.find((column) => column.key === "zone");
    if (zoneCol) {
      zoneCol.options = [...new Set(decoratedRows().map((row) => row.zone).filter(Boolean))].sort();
    }
  }

  function parseSeconds(value) {
    if (value == null || value === "") return 0;
    if (typeof value === "number") return Math.round(value / 1e9);
    const str = String(value);
    if (/^\d+$/.test(str)) return Math.round(Number(str) / 1e9);
    let total = 0;
    const hours = str.match(/(\d+)h/);
    const minutes = str.match(/(\d+)m(?!s)/);
    const seconds = str.match(/(\d+(?:\.\d+)?)s/);
    if (hours) total += parseInt(hours[1], 10) * 3600;
    if (minutes) total += parseInt(minutes[1], 10) * 60;
    if (seconds) total += parseFloat(seconds[1]);
    return total || 0;
  }

  function readDurations(element) {
    intervalSeconds = parseSeconds(element?.spawnInterval) || 60;
    useOverride = element?.respawnTimeOverride != null && parseSeconds(element.respawnTimeOverride) > 0;
    overrideSeconds = useOverride ? parseSeconds(element.respawnTimeOverride) : 60;
    delaySeconds = parseSeconds(element?.respawnDelay) || 0;
  }

  function payload(element) {
    const copy = { ...element };
    copy.spawnInterval = Math.round(Number(intervalSeconds) || 0) * 1e9;
    copy.respawnTimeOverride = useOverride ? Math.round(Number(overrideSeconds) || 0) * 1e9 : null;
    copy.respawnDelay = Math.round(Number(delaySeconds) || 0) * 1e9;
    delete copy.roomName;
    delete copy.zone;
    delete copy.templateName;
    delete copy.respawnLabel;
    delete copy.liveCount;
    delete copy.isNew;
    return copy;
  }

  const config = {
    title: "Spawners",
    entityType: "spawner",
    subtitle: "Where NPC templates appear, how many stay up, and how fast they return.",
    listTitle: "Spawners",
    columns,
    hideDetails: true,
    labels: {
      create: "Create Spawner",
      update: "Update Spawner",
      delete: "Delete",
    },
    get: (token, _filters, cb, errorCb) => {
      loadAll(token).then((next) => {
        remember(next);
        cb(tableRows());
      }).catch(errorCb);
    },
    getElement: getNPCSpawner,
    create: (token, element, cb, errorCb) => createNPCSpawner(token, payload(element), cb, errorCb),
    update: (token, id, element, cb, errorCb) => updateNPCSpawner(token, id, payload(element), cb, errorCb),
    delete: deleteNPCSpawner,
    beforeSelect: readDurations,
    new: (select) => {
      select({
        id: uuidv4(),
        name: "",
        templateId: "",
        roomId: "",
        maxInstances: 1,
        initialCount: 1,
        spawnInterval: 60 * 1e9,
        respawnDelay: 0,
        isNew: true,
      });
    },
  };

  function applyFilters() {
    if (!$isAuthenticated) return;
    store.setElements(tableRows());
  }

  function open(event, path) {
    if (!path) return;
    event.preventDefault();
    navigateTo(path);
  }

  function inspectorPath(kind, id) {
    if (!id) return "";
    const base = kind === "room" ? "/creator/rooms" : "/creator/npcs";
    return `${base}?id=${encodeURIComponent(id)}&view=inspector`;
  }

  function roomName(id) {
    return indexById(cache.rooms)[id]?.name || id;
  }

  function syncDurations() {
    if (!selected) return;
    selected.spawnInterval = Math.round(Number(intervalSeconds) || 0) * 1e9;
    selected.respawnTimeOverride = useOverride ? Math.round(Number(overrideSeconds) || 0) * 1e9 : null;
    selected.respawnDelay = Math.round(Number(delaySeconds) || 0) * 1e9;
  }

  function draftFor(npc) {
    uniqueGap = false;
    const element = {
      id: uuidv4(),
      name: `${npc.name || npc.id} spawner`,
      templateId: npc.id,
      roomId: npc.spawnRoomId || "",
      maxInstances: 1,
      initialCount: 1,
      spawnInterval: 60 * 1e9,
      respawnDelay: 0,
      isNew: true,
    };
    readDurations(element);
    store.setElements(tableRows());
    store.setSelectedElement(element);
    store.openDetail();
  }

  $: gaps = uniqueGap
    ? filterUniqueGaps(uniqueTemplatesWithoutSpawner(cache.npcs, cache.spawners), {
        templateId,
        roomId,
        zone,
        roomsById: indexById(cache.rooms),
      })
    : [];
  $: pickerNPCs = contentNPCs(cache.npcs);
  $: selected = $store.selectedElement;
</script>

<CRUDEditor {config} {store}>
  <div slot="toolbar" class="card mb-3 space-y-3 p-3" data-spawner-filters>
    <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
      <div class="space-y-1.5">
        <div class="label-caps">Template</div>
        <EntitySelectButton
          value={templateId}
          elements={pickerNPCs}
          columns={npcColumns}
          title="Filter by template"
          placeholder="Any template"
          on:change={(e) => { templateId = e.detail || ""; applyFilters(); }}
        />
      </div>
      <div class="space-y-1.5">
        <div class="label-caps">Room</div>
        <EntitySelectButton
          value={roomId}
          elements={cache.rooms}
          columns={roomColumns}
          title="Filter by room"
          placeholder="Any room"
          on:change={(e) => { roomId = e.detail || ""; applyFilters(); }}
        />
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="spawner-zone">Zone</label>
        <select id="spawner-zone" class="input-base" bind:value={zone} on:change={(e) => { zone = e.currentTarget.value; applyFilters(); }}>
          <option value="">All zones</option>
          {#each zones() as area}
            <option value={area}>{area}</option>
          {/each}
        </select>
      </div>
    </div>
    <label class="flex items-center gap-2 text-sm" for="spawner-unique-gap">
      <input id="spawner-unique-gap" type="checkbox" bind:checked={uniqueGap} on:change={(e) => { uniqueGap = e.currentTarget.checked; applyFilters(); }} />
      Unique template without spawner
    </label>
    {#if cache.liveError}
      <p class="text-xs text-amber-300">{cache.liveError}</p>
    {/if}
    {#if uniqueGap}
      {#if gaps.length === 0}
        <p class="text-sm text-slate-500">Every unique NPC in this filter has a spawner.</p>
      {:else}
        <ul class="space-y-2">
          {#each gaps as npc}
            <li class="flex flex-wrap items-center gap-2 text-sm">
              <a class="text-primary hover:underline" href={inspectorPath("npc", npc.id)} on:click={(e) => open(e, inspectorPath("npc", npc.id))}>{npc.name || npc.id}</a>
              <span class="font-mono text-[10px] text-slate-500">{npc.id}</span>
              {#if npc.spawnRoomId}
                <a class="text-primary hover:underline" href={inspectorPath("room", npc.spawnRoomId)} on:click={(e) => open(e, inspectorPath("room", npc.spawnRoomId))}>{roomName(npc.spawnRoomId)}</a>
              {:else}
                <span class="text-slate-500">no spawn room</span>
              {/if}
              <button class="text-xs text-primary hover:underline" type="button" on:click={() => draftFor(npc)}>Create spawner</button>
            </li>
          {/each}
        </ul>
      {/if}
    {/if}
  </div>

  <div slot="content" class="space-y-4">
    {#if selected}
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div class="space-y-1.5">
          <div class="label-caps">Room</div>
          <EntitySelectButton
            value={selected.roomId}
            elements={cache.rooms}
            columns={roomColumns}
            title="Select room"
            placeholder="Select a room..."
            allowNone={false}
            on:change={(e) => selected.roomId = e.detail}
          />
          {#if selected.roomId}
            <a class="text-xs text-primary hover:underline" href={inspectorPath("room", selected.roomId)} on:click={(e) => open(e, inspectorPath("room", selected.roomId))}>Room inspector</a>
          {/if}
        </div>
        <div class="space-y-1.5">
          <div class="label-caps">NPC template</div>
          <EntitySelectButton
            value={selected.templateId}
            elements={pickerNPCs}
            columns={npcColumns}
            title="Select NPC template"
            placeholder="Select NPC template..."
            allowNone={false}
            on:change={(e) => selected.templateId = e.detail}
          />
          {#if selected.templateId}
            <a class="text-xs text-primary hover:underline" href={inspectorPath("npc", selected.templateId)} on:click={(e) => open(e, inspectorPath("npc", selected.templateId))}>NPC inspector</a>
          {/if}
        </div>
      </div>

      <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="spawner-max">Max</label>
          <input id="spawner-max" class="input-base" type="number" min="1" bind:value={selected.maxInstances} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="spawner-initial">Initial</label>
          <input id="spawner-initial" class="input-base" type="number" min="0" bind:value={selected.initialCount} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="spawner-interval">Interval (sec)</label>
          <input id="spawner-interval" class="input-base" type="number" min="1" bind:value={intervalSeconds} on:input={syncDurations} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="spawner-delay">Respawn delay (sec)</label>
          <input id="spawner-delay" class="input-base" type="number" min="0" bind:value={delaySeconds} on:input={syncDurations} />
        </div>
      </div>

      <label class="flex items-center gap-2 text-sm" for="spawner-override">
        <input id="spawner-override" type="checkbox" bind:checked={useOverride} on:change={syncDurations} />
        Override respawn time
      </label>
      {#if useOverride}
        <div class="space-y-1.5">
          <label class="label-caps" for="spawner-override-seconds">Respawn override (sec)</label>
          <input id="spawner-override-seconds" class="input-base w-32" type="number" min="1" bind:value={overrideSeconds} on:input={syncDurations} />
        </div>
      {/if}
      {#if selected.liveCount != null}
        <p class="text-xs text-slate-500">Live in this room: {selected.liveCount}. Dead copies are not counted.</p>
      {/if}
    {/if}
  </div>
</CRUDEditor>
