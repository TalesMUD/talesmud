<script>
  import { get } from "svelte/store";
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { userRole } from "../stores.js";
  import { getRoomInspect } from "../api/rooms.js";
  import { getLiveCharacters, getLiveInstances, getLiveNPCs } from "../api/live.js";
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import OpsButtons from "./OpsButtons.svelte";

  export let roomId = "";
  export let isNew = false;

  const { authToken } = getAuth();

  const paths = {
    room: "/creator/rooms",
    npc: "/creator/npcs",
    item: "/creator/item-templates",
    script: "/creator/scripts",
    quest: "/creator/quests",
  };

  const characterColumns = [
    { key: "id", label: "ID", mono: true, priority: 1 },
    { key: "name", label: "Name", priority: 1 },
    { key: "roomId", label: "Room", mono: true, priority: 2 },
    { key: "level", label: "Level", type: "number", priority: 2 },
  ];

  let view = null;
  let npcs = [];
  let people = [];
  let copies = [];
  let characters = [];
  let characterId = "";
  let error = "";
  let liveNote = "";
  let loading = false;
  let loadedFor = "";
  let seq = 0;

  function href(type, id) {
    const base = paths[type];
    if (!base || !id) return "";
    return `${base}?id=${encodeURIComponent(id)}`;
  }

  function open(event, type, id) {
    const path = href(type, id);
    if (!path) return;
    event.preventDefault();
    navigateTo(path);
  }

  function apiError(err, fallback) {
    return err?.response?.data?.error || err?.message || fallback;
  }

  function templateOf(id) {
    const value = String(id || "");
    const mark = value.lastIndexOf("~");
    return mark > 0 ? value.slice(0, mark) : value;
  }

  function cloneHere(copy, id) {
    const mine = (copy?.cloneIds || []).find((cloneId) => templateOf(cloneId) === id);
    return mine || (copy?.cloneIds || [])[0] || "";
  }

  async function load(id) {
    const mine = ++seq;
    loadedFor = id;
    loading = true;
    error = "";
    liveNote = "";
    const token = get(authToken);
    if (!token) {
      error = "Sign in to inspect this room.";
      loading = false;
      view = null;
      return;
    }
    let staticView = null;
    try {
      staticView = await getRoomInspect(token, id);
      if (mine !== seq) return;
      view = staticView;
    } catch (err) {
      if (mine !== seq) return;
      view = null;
      error = apiError(err, "Could not load the inspector.");
      loading = false;
      return;
    }
    const here = staticView?.id || id;
    try {
      const rows = await getLiveNPCs(token);
      if (mine !== seq) return;
      npcs = (Array.isArray(rows) ? rows : []).filter((row) => row?.roomId === here || templateOf(row?.roomId) === here);
    } catch (err) {
      if (mine !== seq) return;
      npcs = [];
      liveNote = apiError(err, "Live NPCs are unavailable.");
    }
    try {
      const rows = await getLiveInstances(token);
      if (mine !== seq) return;
      copies = (Array.isArray(rows) ? rows : []).filter((row) => {
        if (row?.sourceRoom === here) return true;
        return (row?.cloneIds || []).some((cloneId) => templateOf(cloneId) === here);
      });
    } catch (err) {
      if (mine !== seq) return;
      copies = [];
      liveNote = liveNote || apiError(err, "Instance copies are unavailable.");
    }
    if (get(userRole) === "admin") {
      try {
        const rows = await getLiveCharacters(token, { all: "1" });
        if (mine !== seq) return;
        const list = Array.isArray(rows) ? rows : [];
        characters = list;
        people = list.filter((row) => row?.roomId === here || templateOf(row?.roomId) === here);
      } catch (err) {
        if (mine !== seq) return;
        characters = [];
        people = [];
        liveNote = liveNote || apiError(err, "The character list is unavailable.");
      }
    } else {
      characters = [];
      people = [];
    }
    if (mine === seq) loading = false;
  }

  function refresh() {
    if (roomId && !isNew) load(roomId);
  }

  $: if (isNew) {
    loadedFor = roomId;
    view = null;
    npcs = [];
    people = [];
    copies = [];
    error = "";
    liveNote = "";
    loading = false;
  } else if (roomId && roomId !== loadedFor) {
    load(roomId);
  }

  $: selectedCharacter = (characters || []).find((row) => row.id === characterId) || null;

  function exitLabel(exit) {
    const where = exit.roomName || exit.roomId || "nowhere";
    return `${exit.name || "exit"} → ${where}`;
  }
</script>

<div class="space-y-4">
  {#if isNew}
    <div class="rounded-md border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
      Save the room before opening the inspector.
    </div>
  {:else if error}
    <div class="rounded-md border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-100">{error}</div>
  {:else if loading && !view}
    <div class="py-8 text-center text-sm text-slate-500">Loading inspector…</div>
  {:else if view}
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap items-center gap-2">
        {#if view.reach?.reachable}
          <span class="rounded-full bg-emerald-500/15 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-emerald-200">Reachable</span>
        {:else}
          <span class="rounded-full bg-amber-500/15 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-amber-200">Unreachable</span>
        {/if}
        {#if view.area}<span class="text-xs text-slate-400">{view.area}</span>{/if}
        {#if view.roomType}<span class="text-xs text-slate-500">{view.roomType}</span>{/if}
        {#each view.tags || [] as tag}
          <span class="rounded-full bg-slate-800 px-2 py-0.5 text-[10px] uppercase tracking-wide text-slate-300">{tag}</span>
        {/each}
      </div>
      <button class="btn btn-outline text-xs" type="button" on:click={refresh} disabled={loading}>Refresh</button>
    </div>

    {#if view.requestedId && view.requestedId !== view.id}
      <p class="text-xs text-slate-400">This id is a copy of <span class="font-mono">{view.id}</span>. The summary is the template room.</p>
    {/if}

    <div class="card p-4 space-y-1">
      <div class="label-caps">Reachability</div>
      {#if view.reach?.reachable}
        <p class="text-sm text-slate-300">Reachable from {view.reach.startRoomId || "the start room"}.</p>
      {:else}
        <p class="text-sm text-amber-200">Not reachable from {view.reach?.startRoomId || "the start room"}.</p>
        {#if view.reach?.reason}<p class="text-xs text-slate-400">{view.reach.reason}</p>{/if}
      {/if}
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div class="card p-4 space-y-2">
        <div class="label-caps">Exits out</div>
        {#if view.exitsOut?.length}
          {#each view.exitsOut as exit}
            <div class="text-sm">
              {#if exit.roomId}
                <a class="text-primary hover:underline" href={href("room", exit.roomId)} on:click={(e) => open(e, "room", exit.roomId)}>{exitLabel(exit)}</a>
              {:else}
                <span>{exit.name || "exit"} → nowhere</span>
              {/if}
              {#if exit.hidden}<span class="ml-2 text-amber-200">hidden</span>{/if}
              {#if exit.instance}<span class="ml-2 text-slate-400">instance</span>{/if}
              {#if exit.missing}<span class="ml-2 text-amber-300">missing</span>{/if}
              {#if exit.hidden}
                <div class="text-xs text-slate-500">
                  {#if exit.revealedBy?.length}
                    Revealed by
                    {#each exit.revealedBy as script, i}
                      {#if i > 0}<span>, </span>{/if}<a class="text-primary hover:underline" href={href("script", script.id)} on:click={(e) => open(e, "script", script.id)}>{script.name || script.id}</a>
                    {/each}
                  {:else}
                    No script reveals this exit.
                  {/if}
                </div>
              {/if}
            </div>
          {/each}
        {:else}
          <p class="text-sm text-slate-500">No exits out.</p>
        {/if}
      </div>
      <div class="card p-4 space-y-2">
        <div class="label-caps">Exits in</div>
        {#if view.exitsIn?.length}
          {#each view.exitsIn as exit}
            <div class="text-sm">
              {#if exit.roomId}
                <a class="text-primary hover:underline" href={href("room", exit.roomId)} on:click={(e) => open(e, "room", exit.roomId)}>{exit.name || "exit"} from {exit.roomName || exit.roomId}</a>
              {:else}
                <span>{exit.name || "exit"}</span>
              {/if}
              {#if exit.hidden}<span class="ml-2 text-amber-200">hidden</span>{/if}
              {#if exit.instance}<span class="ml-2 text-slate-400">instance</span>{/if}
              {#if exit.hidden}
                <div class="text-xs text-slate-500">
                  {#if exit.revealedBy?.length}
                    Revealed by
                    {#each exit.revealedBy as script, i}
                      {#if i > 0}<span>, </span>{/if}<a class="text-primary hover:underline" href={href("script", script.id)} on:click={(e) => open(e, "script", script.id)}>{script.name || script.id}</a>
                    {/each}
                  {:else}
                    No script reveals this exit.
                  {/if}
                </div>
              {/if}
            </div>
          {/each}
        {:else}
          <p class="text-sm text-slate-500">No exits in.</p>
        {/if}
      </div>
    </div>

    <div class="card p-4 space-y-2">
      <div class="label-caps">NPCs and spawners</div>
      {#if view.actors?.length}
        {#each view.actors as actor}
          <p class="text-sm">
            <span class="rounded bg-slate-800 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-slate-300">{actor.how}</span>
            <a class="text-primary hover:underline" href={href("npc", actor.id)} on:click={(e) => open(e, "npc", actor.id)}>{actor.name || actor.id}</a>
            <span class="font-mono text-[10px] text-slate-500">{actor.id}</span>
            {#if actor.level}<span class="text-slate-500">lvl {actor.level}</span>{/if}
            {#if actor.missing}<span class="text-amber-300">missing</span>{/if}
          </p>
        {/each}
      {:else}
        <p class="text-sm text-slate-500">No content NPC is placed here.</p>
      {/if}
      {#if view.spawners?.length}
        {#each view.spawners as spawner}
          <p class="text-sm text-slate-300">
            Spawner <span class="font-mono text-xs">{spawner.id}</span>
            {#if spawner.templateId}
              · <a class="text-primary hover:underline" href={href("npc", spawner.templateId)} on:click={(e) => open(e, "npc", spawner.templateId)}>{spawner.templateName || spawner.name || spawner.templateId}</a>
            {/if}
            · max {spawner.maxInstances}
            · {spawner.respawnTime ? `respawn ${spawner.respawnTime}` : "no respawn"}
            {#if spawner.templateMissing}<span class="text-amber-300">missing template</span>{/if}
          </p>
        {/each}
      {:else}
        <p class="text-sm text-slate-500">No spawner in this room.</p>
      {/if}
    </div>

    <div class="card p-4 space-y-2">
      <div class="label-caps">Items</div>
      {#if view.items?.length}
        {#each view.items as item}
          <p class="text-sm">
            <a class="text-primary hover:underline" href={href("item", item.id)} on:click={(e) => open(e, "item", item.id)}>{item.name || item.id}</a>
            <span class="font-mono text-[10px] text-slate-500">{item.id}</span>
            {#if item.missing}<span class="text-amber-300">missing</span>{/if}
          </p>
        {/each}
      {:else}
        <p class="text-sm text-slate-500">No items.</p>
      {/if}
    </div>

    <div class="card p-4 space-y-2">
      <div class="label-caps">Scripts</div>
      <p class="text-sm">
        On enter
        {#if view.onEnter}
          <a class="text-primary hover:underline" href={href("script", view.onEnter.id)} on:click={(e) => open(e, "script", view.onEnter.id)}>{view.onEnter.name || view.onEnter.id}</a>
          {#if view.onEnter.missing}<span class="text-amber-300">missing</span>{/if}
        {:else}
          <span class="text-slate-500">none</span>
        {/if}
      </p>
      {#if view.actions?.length}
        {#each view.actions as action}
          <p class="text-sm">
            {action.name || "action"}
            {#if action.type}<span class="text-slate-500">{action.type}</span>{/if}
            {#if action.script}
              <a class="text-primary hover:underline" href={href("script", action.script.id)} on:click={(e) => open(e, "script", action.script.id)}>{action.script.name || action.script.id}</a>
              {#if action.script.missing}<span class="text-amber-300">missing</span>{/if}
            {/if}
          </p>
        {/each}
      {:else}
        <p class="text-sm text-slate-500">No actions.</p>
      {/if}
    </div>

    <div class="card p-4 space-y-2">
      <div class="label-caps">Quests</div>
      {#if view.quests?.length}
        {#each view.quests as quest}
          <p class="text-sm">
            <span class="rounded bg-slate-800 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-slate-300">{quest.role}</span>
            <a class="text-primary hover:underline" href={href("quest", quest.id)} on:click={(e) => open(e, "quest", quest.id)}>{quest.name || quest.id}</a>
            {#if quest.description}<span class="text-slate-400">· {quest.description}</span>{/if}
          </p>
        {/each}
      {:else}
        <p class="text-sm text-slate-500">No visit objective names this room.</p>
      {/if}
    </div>

    <div class="card p-4 space-y-3">
      <div class="label-caps">Live</div>
      {#if liveNote}<p class="text-sm text-amber-200">{liveNote}</p>{/if}
      <div>
        <div class="text-xs uppercase tracking-wide text-slate-500 mb-1">Characters here</div>
        {#if $userRole !== "admin"}
          <p class="text-sm text-slate-500">The character list is admin only.</p>
        {:else if people.length}
          {#each people as person}
            <p class="text-sm">{person.name || person.id} <span class="font-mono text-[10px] text-slate-500">{person.roomId}</span>{person.online ? " · online" : ""}{person.inCombat ? " · in combat" : ""}</p>
          {/each}
        {:else}
          <p class="text-sm text-slate-500">No character is in this room.</p>
        {/if}
      </div>
      <div>
        <div class="text-xs uppercase tracking-wide text-slate-500 mb-1">NPC instances</div>
        {#if npcs.length}
          {#each npcs as row}
            <p class="text-sm">
              <a class="text-primary hover:underline" href={href("npc", row.templateId || row.id)} on:click={(e) => open(e, "npc", row.templateId || row.id)}>{row.name || row.id}</a>
              <span class="font-mono text-[10px] text-slate-500">{row.roomId}</span>
              · {row.hp}/{row.maxHp}
              {#if row.dead} · dead{:else if row.inCombat} · in combat{row.combatWith?.length ? ` with ${row.combatWith.join(", ")}` : ""}{/if}
            </p>
          {/each}
        {:else}
          <p class="text-sm text-slate-500">No running NPC instance is here.</p>
        {/if}
      </div>
      <div class="space-y-2">
        <div class="text-xs uppercase tracking-wide text-slate-500">Instance copies</div>
        {#if $userRole === "admin"}
          <EntitySelectButton
            value={characterId}
            elements={characters}
            columns={characterColumns}
            title="Select Character"
            placeholder="Character to teleport…"
            on:change={(event) => (characterId = event.detail)}
          />
          <OpsButtons
            mode="room"
            roomId={view.id}
            {characterId}
            characterName={selectedCharacter?.name || ""}
            onDone={refresh}
          />
        {:else}
          <p class="text-xs text-slate-500">Teleport and instance cleanup are admin only.</p>
        {/if}
        {#if copies.length}
          {#each copies as copy (copy.id)}
            <div class="rounded-md border border-slate-800 px-3 py-2 space-y-2 text-sm">
              <div>
                <span class="font-mono text-xs">{cloneHere(copy, view.id) || copy.id}</span>
                {#if copy.players?.length}<span class="text-slate-400">· {copy.players.length} player{copy.players.length === 1 ? "" : "s"}</span>{/if}
              </div>
              {#if $userRole === "admin"}
                <OpsButtons
                  mode="room"
                  roomId={cloneHere(copy, view.id) || view.id}
                  roomCopyId={cloneHere(copy, view.id)}
                  {characterId}
                  characterName={selectedCharacter?.name || ""}
                  onDone={refresh}
                />
              {/if}
            </div>
          {/each}
        {:else}
          <p class="text-sm text-slate-500">No instance copy of this room.</p>
        {/if}
      </div>
    </div>

    <!-- Worker B embeds BacklinksPanel for inbound references. -->
    <div data-backlinks-slot data-entity-type="room" data-entity-id={view.id}></div>
  {/if}
</div>
