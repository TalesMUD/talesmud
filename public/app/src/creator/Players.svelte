<script>
  import { onMount } from "svelte";
  import axios from "axios";
  import { getAuth } from "../auth.js";
  import { backend } from "../api/base.js";
  import { getRoomsAsync } from "../api/rooms.js";
  import { getLiveCharacters, getLiveCharacter, getLiveInstances } from "../api/live.js";
  import { itemTemplateColumns, roomColumns } from "./tableColumns.js";
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import ConfirmDialog from "./ConfirmDialog.svelte";
  import OpsButtons from "./OpsButtons.svelte";
  import { opError, performOp } from "./opsFlow.js";

  const { isAuthenticated, authToken } = getAuth();

  let characters = [];
  let instances = [];
  let rooms = [];
  let itemTemplates = [];
  let detail = null;
  let selectedId = "";
  let loading = false;
  let detailLoading = false;
  let error = "";
  let loaded = false;

  let onlineOnly = false;
  let guestOnly = false;
  let combatOnly = false;
  let includeOlder = false;
  let zone = "";

  let teleportRoomId = "";
  let teleportForce = false;
  let giveTemplateId = "";
  let giveQuantity = 1;
  let takeInstanceId = "";
  let takeQuantity = 1;
  let pending = null;
  let busy = false;

  function when(value) {
    if (!value || String(value).startsWith("0001-")) return "—";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "—";
    return date.toLocaleString();
  }

  function collectItems(list, out, slot) {
    for (const item of list || []) {
      if (!item) continue;
      out.push({
        id: item.id,
        name: item.name || "item",
        quantity: item.quantity || 1,
        templateId: item.templateId || "",
        slot: slot || "",
      });
      if (Array.isArray(item.items) && item.items.length) collectItems(item.items, out, "");
    }
  }

  $: heldItems = (() => {
    const out = [];
    collectItems(detail?.inventory?.items, out, "");
    const equipped = detail?.equippedItems;
    if (equipped && typeof equipped === "object") {
      for (const [slot, item] of Object.entries(equipped)) collectItems([item], out, slot);
    }
    return out.filter((item) => item.id);
  })();

  $: quests = Array.isArray(detail?.quests) ? detail.quests : [];

  function params() {
    const query = {};
    if (onlineOnly) query.online = "1";
    if (guestOnly) query.guest = "1";
    if (combatOnly) query.inCombat = "1";
    if (includeOlder) query.all = "1";
    if (zone.trim()) query.zone = zone.trim();
    return query;
  }

  async function loadLists() {
    if (!$authToken) return;
    loading = true;
    error = "";
    try {
      const [rows, copies] = await Promise.all([
        getLiveCharacters($authToken, params()),
        getLiveInstances($authToken),
      ]);
      characters = Array.isArray(rows) ? rows : [];
      instances = Array.isArray(copies) ? copies : [];
      loaded = true;
      if (selectedId && !characters.some((row) => row.id === selectedId)) {
        selectedId = "";
        detail = null;
      }
    } catch (err) {
      error = opError(err, "Live characters are unavailable.");
    } finally {
      loading = false;
    }
  }

  async function loadCatalogs() {
    if (!$authToken) return;
    try {
      const [roomRows, itemRows] = await Promise.all([
        getRoomsAsync($authToken, []),
        axios
          .get(`${backend}/items`, {
            headers: { Authorization: `Bearer ${$authToken}` },
            params: { isTemplate: "true" },
          })
          .then((res) => res.data),
      ]);
      rooms = Array.isArray(roomRows) ? roomRows : [];
      itemTemplates = Array.isArray(itemRows) ? itemRows : [];
    } catch (err) {
      error = opError(err, "Rooms or item templates could not be loaded.");
    }
  }

  async function openCharacter(id) {
    selectedId = id;
    detail = characters.find((row) => row.id === id) || { id, name: id };
    teleportRoomId = "";
    giveTemplateId = "";
    takeInstanceId = "";
    if (!$authToken) return;
    detailLoading = true;
    try {
      detail = await getLiveCharacter($authToken, id);
    } catch (err) {
      error = opError(err, "That character could not be loaded.");
    } finally {
      detailLoading = false;
    }
  }

  async function refreshSelected() {
    await loadLists();
    if (selectedId) await openCharacter(selectedId);
  }

  function ask(job) {
    error = "";
    pending = job;
  }

  function roomName(id) {
    return rooms.find((room) => room.id === id)?.name || id;
  }

  function templateName(id) {
    return itemTemplates.find((item) => item.id === id)?.name || id;
  }

  function requestTeleport() {
    if (!detail?.id || !teleportRoomId) return;
    ask({
      title: "Teleport this character?",
      entityName: detail.name,
      entityId: detail.id,
      detail: `Move them to ${roomName(teleportRoomId)} (${teleportRoomId}).${teleportForce ? " A current fight will be aborted first." : " A fight blocks the move unless you force it."}`,
      hint: "Undo moves them back if they are still in the destination and not in a new fight.",
      action: "teleport",
      body: { characterId: detail.id, roomId: teleportRoomId, force: teleportForce },
    });
  }

  function requestGive() {
    if (!detail?.id || !giveTemplateId) return;
    const quantity = Number(giveQuantity) || 1;
    ask({
      title: "Give this item?",
      entityName: detail.name,
      entityId: detail.id,
      detail: `Add ${quantity} × ${templateName(giveTemplateId)}. A unique they already hold is refused.`,
      hint: "Undo takes back the pieces this action added.",
      action: "give-item",
      body: { characterId: detail.id, itemTemplateId: giveTemplateId, quantity },
    });
  }

  function requestTake() {
    if (!detail?.id || !takeInstanceId) return;
    const held = heldItems.find((item) => item.id === takeInstanceId);
    const quantity = Number(takeQuantity) || 1;
    ask({
      title: "Take this item?",
      entityName: detail.name,
      entityId: detail.id,
      detail: `Remove ${quantity} × ${held?.name || "item"} from this character.`,
      hint: "Undo puts the same pieces back.",
      action: "take-item",
      body: { characterId: detail.id, itemInstanceId: takeInstanceId, quantity },
    });
  }

  function requestCleanup(copy) {
    if (copy) {
      ask({
        title: "Clean this instance copy?",
        entityName: copy.id,
        entityId: copy.sourceRoom || "",
        detail: "Move anyone inside to the copy's entrance, then delete the copy. A fight in the copy is not aborted.",
        hint: "This cannot be undone.",
        action: "instance-cleanup",
        body: { roomCopyId: copy.id },
      });
      return;
    }
    ask({
      title: "Clean empty instance copies?",
      entityName: "Empty instance copies",
      detail: "Delete instance copies with nobody inside. Occupied copies stay.",
      hint: "This cannot be undone.",
      action: "instance-cleanup",
      body: { allEmpty: true },
    });
  }

  async function runPending() {
    const job = pending;
    pending = null;
    if (!job || !$authToken) return;
    busy = true;
    error = "";
    try {
      await performOp($authToken, job.action, job.body);
      await refreshSelected();
    } catch (err) {
      error = opError(err);
    } finally {
      busy = false;
    }
  }

  onMount(() => {
    loadLists();
    loadCatalogs();
  });

  $: if ($isAuthenticated && $authToken && !loaded && !loading) {
    loadLists();
    loadCatalogs();
  }
</script>

<div class="flex flex-col h-[calc(100vh-128px)]">
  <div class="px-6 pt-5 pb-3 flex-shrink-0">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
      <div class="space-y-1">
        <h1 class="text-2xl font-bold tracking-tight">Players</h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm">Online characters and anyone seen in the last 30 days.</p>
      </div>
      <button class="btn btn-outline" type="button" on:click={loadLists} disabled={loading}>
        <span class="material-symbols-outlined text-sm">refresh</span>
        Refresh
      </button>
    </div>
  </div>

  <div class="flex-1 overflow-y-auto px-6 pb-5 thin-scrollbar space-y-4">
    {#if error}
      <div class="rounded-md border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-100">{error}</div>
    {/if}

    <div class="card p-4">
      <div class="grid grid-cols-1 md:grid-cols-5 gap-3 items-center">
        <label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={onlineOnly} on:change={loadLists} /> Online</label>
        <label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={guestOnly} on:change={loadLists} /> Guest</label>
        <label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={combatOnly} on:change={loadLists} /> In combat</label>
        <label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={includeOlder} on:change={loadLists} /> Include older</label>
        <input class="input-base" type="search" placeholder="Zone" bind:value={zone} on:change={loadLists} />
      </div>
    </div>

    <div class="flex flex-col xl:flex-row gap-4 items-start">
      <div class="card overflow-x-auto flex-1 min-w-0 w-full">
        <table class="w-full text-sm" style="min-width: 760px;">
          <thead>
            <tr class="text-left text-[10px] uppercase tracking-wider text-slate-500">
              <th class="px-3 py-2">Name</th>
              <th class="px-3 py-2">User</th>
              <th class="px-3 py-2">Lvl</th>
              <th class="px-3 py-2">Class</th>
              <th class="px-3 py-2">Room</th>
              <th class="px-3 py-2">Zone</th>
              <th class="px-3 py-2">State</th>
            </tr>
          </thead>
          <tbody>
            {#if loading && characters.length === 0}
              <tr><td class="px-3 py-6 text-slate-500" colspan="7">Loading…</td></tr>
            {:else if characters.length === 0}
              <tr><td class="px-3 py-6 text-slate-500" colspan="7">No characters in this window.</td></tr>
            {:else}
              {#each characters as row}
                <tr
                  class="border-t border-slate-800 cursor-pointer hover:bg-slate-800/40 {selectedId === row.id ? 'bg-slate-800/60' : ''}"
                  on:click={() => openCharacter(row.id)}
                >
                  <td class="px-3 py-2">
                    <div class="font-semibold">{row.name}</div>
                    <div class="font-mono text-[10px] text-slate-500">{row.id}</div>
                  </td>
                  <td class="px-3 py-2">{row.userName || "—"}{row.guest ? " · guest" : ""}</td>
                  <td class="px-3 py-2">{row.level}</td>
                  <td class="px-3 py-2">{row.classId || "—"}</td>
                  <td class="px-3 py-2">{row.roomName || row.roomId || "—"}</td>
                  <td class="px-3 py-2">{row.zone || "—"}</td>
                  <td class="px-3 py-2">
                    {row.online ? "online" : "offline"}
                    {row.inCombat ? " · combat" : ""}
                    {#if row.partyName}<div class="text-xs text-slate-500">{row.partyName}</div>{/if}
                    <div class="text-[10px] text-slate-500">{when(row.lastSeen)}</div>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>

      {#if selectedId}
        <aside class="card p-4 w-full xl:w-[420px] xl:sticky xl:top-0 space-y-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-lg font-bold">{detail?.name || selectedId}</h2>
              <p class="font-mono text-xs text-slate-500 break-all">{detail?.id}</p>
              <p class="text-sm text-slate-400 mt-1">
                {detail?.online ? "Online" : "Offline"}
                · L{detail?.level || 0}
                {detail?.classId ? ` · ${detail.classId}` : ""}
                {detail?.inCombat ? " · in combat" : ""}
              </p>
              <p class="text-sm text-slate-400">{detail?.roomName || "No room"} <span class="font-mono text-xs">{detail?.roomId || ""}</span></p>
              {#if detail?.zone}<p class="text-xs text-slate-500">{detail.zone}</p>{/if}
            </div>
            <button class="btn btn-outline text-xs" type="button" on:click={() => { selectedId = ""; detail = null; }}>Close</button>
          </div>

          {#if detailLoading}
            <p class="text-sm text-slate-500">Loading character…</p>
          {/if}

          {#if detail?.combat}
            <div class="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm">
              Fight {detail.combat.id} in {detail.combat.roomId || "unknown room"}.
              {#each detail.combat.enemies || [] as enemy}
                <div>{enemy.name}: {enemy.hp}/{enemy.max}</div>
              {/each}
            </div>
          {/if}

          <section class="space-y-2">
            <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Teleport</h3>
            <EntitySelectButton
              value={teleportRoomId}
              elements={rooms}
              columns={roomColumns}
              title="Select Room"
              placeholder="Select a room..."
              allowNone={true}
              on:change={(event) => (teleportRoomId = event.detail)}
            />
            <label class="flex items-center gap-2 text-sm">
              <input type="checkbox" bind:checked={teleportForce} />
              Force (abort a current fight, then move)
            </label>
            <button class="btn btn-outline text-xs" type="button" disabled={busy || !teleportRoomId} on:click={requestTeleport}>Teleport…</button>
          </section>

          <section class="space-y-2">
            <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Give item</h3>
            <EntitySelectButton
              value={giveTemplateId}
              elements={itemTemplates}
              columns={itemTemplateColumns}
              title="Select Item Template"
              placeholder="Select an item template..."
              allowNone={true}
              on:change={(event) => (giveTemplateId = event.detail)}
            />
            <input class="input-base" type="number" min="1" bind:value={giveQuantity} />
            <button class="btn btn-outline text-xs" type="button" disabled={busy || !giveTemplateId} on:click={requestGive}>Give item…</button>
          </section>

          <section class="space-y-2">
            <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Take item</h3>
            <select class="input-base" bind:value={takeInstanceId}>
              <option value="">This character's items…</option>
              {#each heldItems as item}
                <option value={item.id}>
                  {item.slot ? `${item.slot}: ` : ""}{item.name} ×{item.quantity}
                </option>
              {/each}
            </select>
            <input class="input-base" type="number" min="1" bind:value={takeQuantity} />
            <button class="btn btn-outline text-xs" type="button" disabled={busy || !takeInstanceId} on:click={requestTake}>Take item…</button>
          </section>

          <section class="space-y-2">
            <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Character</h3>
            <OpsButtons mode="character" characterId={detail?.id || selectedId} onDone={refreshSelected} />
          </section>

          <section class="space-y-2">
            <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Quests</h3>
            <OpsButtons mode="quest" characterId={detail?.id || selectedId} {quests} onDone={refreshSelected} />
          </section>

          {#if detail?.revealedExits}
            <section>
              <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-500 mb-1">Revealed exits</h3>
              <pre class="text-xs text-slate-400 whitespace-pre-wrap break-all">{JSON.stringify(detail.revealedExits)}</pre>
            </section>
          {/if}
        </aside>
      {/if}
    </div>

    <div class="card p-4 space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-lg font-bold">Instance copies</h2>
          <p class="text-sm text-slate-500">Empty copies can be removed. Occupied copies are skipped when cleaning all of them.</p>
        </div>
        <button class="btn btn-outline text-xs" type="button" disabled={busy} on:click={() => requestCleanup(null)}>Clean empty copies…</button>
      </div>
      {#if instances.length === 0}
        <p class="text-sm text-slate-500">No instance copies.</p>
      {:else}
        <div class="overflow-x-auto">
          <table class="w-full text-sm" style="min-width: 640px;">
            <thead>
              <tr class="text-left text-[10px] uppercase tracking-wider text-slate-500">
                <th class="px-3 py-2">Copy</th>
                <th class="px-3 py-2">Source</th>
                <th class="px-3 py-2">Players</th>
                <th class="px-3 py-2">Created</th>
                <th class="px-3 py-2"></th>
              </tr>
            </thead>
            <tbody>
              {#each instances as copy}
                <tr class="border-t border-slate-800">
                  <td class="px-3 py-2 font-mono text-xs break-all">{copy.id}</td>
                  <td class="px-3 py-2 font-mono text-xs">{copy.sourceRoom || "—"}</td>
                  <td class="px-3 py-2">{(copy.players || []).length}</td>
                  <td class="px-3 py-2">{when(copy.created)}</td>
                  <td class="px-3 py-2">
                    <button class="btn btn-outline text-xs" type="button" disabled={busy} on:click={() => requestCleanup(copy)}>Clean…</button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </div>
</div>

<ConfirmDialog
  open={!!pending}
  title={pending?.title || "Run live op?"}
  entityName={pending?.entityName || ""}
  entityId={pending?.entityId || ""}
  detail={pending?.detail || ""}
  hint={pending?.hint || ""}
  confirmLabel="Run"
  tone="ops"
  on:confirm={runPending}
  on:cancel={() => (pending = null)}
/>
