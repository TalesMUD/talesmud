<script>
  import { get } from "svelte/store";
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { userRole } from "../stores.js";
  import { getNPCInspect } from "../api/npcs.js";
  import { getLiveNPCs } from "../api/live.js";
  import OpsButtons from "./OpsButtons.svelte";

  export let npcId = "";
  export let isNew = false;

  const { authToken } = getAuth();

  const kindLabel = { enemy: "Enemy", merchant: "Merchant", npc: "NPC" };
  const paths = {
    room: "/creator/rooms",
    dialog: "/creator/dialogs",
    script: "/creator/scripts",
    quest: "/creator/quests",
    item: "/creator/item-templates",
  };

  let view = null;
  let live = [];
  let error = "";
  let liveError = "";
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

  async function load(id) {
    const mine = ++seq;
    loadedFor = id;
    loading = true;
    error = "";
    liveError = "";
    const token = get(authToken);
    if (!token) {
      error = "Sign in to inspect this NPC.";
      loading = false;
      view = null;
      live = [];
      return;
    }
    let staticView = null;
    try {
      staticView = await getNPCInspect(token, id);
      if (mine !== seq) return;
      view = staticView;
    } catch (err) {
      if (mine !== seq) return;
      view = null;
      error = apiError(err, "Could not load the inspector.");
    }
    const templateId = staticView?.id || id;
    try {
      const rows = await getLiveNPCs(token, { templateId });
      if (mine !== seq) return;
      live = Array.isArray(rows) ? rows : [];
    } catch (err) {
      if (mine !== seq) return;
      live = [];
      liveError = apiError(err, "Live instances are unavailable.");
    }
    if (mine === seq) loading = false;
  }

  function refresh() {
    if (npcId && !isNew) load(npcId);
  }

  $: if (isNew) {
    loadedFor = npcId;
    view = null;
    live = [];
    error = "";
    liveError = "";
    loading = false;
  } else if (npcId && npcId !== loadedFor) {
    load(npcId);
  }

  function percent(chance, guaranteed) {
    if (guaranteed) return "100%";
    const value = Number(chance);
    if (!Number.isFinite(value)) return "—";
    const rounded = Math.round(value * 1000) / 10;
    return `${rounded}%`;
  }

  function hpWidth(row) {
    const max = Number(row?.maxHp) || 0;
    if (max <= 0) return 0;
    return Math.max(0, Math.min(100, (Number(row.hp) / max) * 100));
  }

  function swings(speed) {
    const value = Number(speed) || 0;
    if (value <= 0 || value === 1) return "1 swing / round";
    if (value > 1) return `${value} swings / round`;
    return `${value} (holds between swings)`;
  }

  function fleeText(value) {
    const n = Number(value) || 0;
    if (n <= 0) return "never flees";
    if (n <= 1) return `${Math.round(n * 100)}% HP`;
    return String(n);
  }

  function when(iso) {
    if (!iso) return "";
    const date = new Date(iso);
    if (Number.isNaN(date.getTime())) return String(iso);
    return date.toLocaleString();
  }

  function scalingLabel(stats) {
    if (!stats) return "";
    if (stats.scaling === "named") return "named override";
    if (stats.scaling === "tier") return "difficulty tier";
    if (stats.scaling === "unknown") return "unknown tier, no scaling";
    return "stored stats";
  }

  function factorText(factors) {
    if (!factors) return "";
    return `×${factors.hp} HP · ×${factors.attack} attack · ×${factors.defense} defense`;
  }

  $: spawnRoomId = view?.spawnRoom?.id || "";
</script>

<div class="space-y-4">
  {#if isNew}
    <div class="rounded-md border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
      Save the NPC before opening the inspector.
    </div>
  {:else if error}
    <div class="rounded-md border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-100">{error}</div>
  {:else if loading && !view}
    <div class="py-8 text-center text-sm text-slate-500">Loading inspector…</div>
  {:else if view}
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap items-center gap-2">
        {#each view.kinds || [] as kind}
          <span class="rounded-full bg-slate-800 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-slate-200">{kindLabel[kind] || kind}</span>
        {/each}
        <span class="rounded-full bg-slate-800 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-slate-300">lvl {view.level || 1}</span>
        {#if view.difficulty}
          <span class="rounded-full bg-amber-500/15 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-amber-200">{view.difficulty}</span>
        {/if}
        <span class="rounded-full bg-slate-800 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-slate-400">{view.isTemplate ? "Template" : "Unique"}</span>
        {#if view.behavior?.creatureType}
          <span class="text-xs text-slate-400">{view.behavior.creatureType}{view.behavior.combatStyle ? ` · ${view.behavior.combatStyle}` : ""}</span>
        {/if}
      </div>
      <button class="btn btn-outline text-xs" type="button" on:click={refresh} disabled={loading}>Refresh</button>
    </div>

    {#if view.requestedId && view.requestedId !== view.id}
      <p class="text-xs text-slate-400">This row is an instance of <span class="font-mono">{view.id}</span>. The summary is the template.</p>
    {/if}

    {#if view.stats}
      <div class="card p-4 space-y-3">
        <div class="flex flex-wrap items-baseline justify-between gap-2">
          <div class="label-caps">Combat stats</div>
          <span class="text-[10px] uppercase tracking-wide text-slate-500">{scalingLabel(view.stats)}</span>
        </div>
        {#if view.stats.unknownTier}
          <p class="text-xs text-amber-200">Unknown difficulty tier. No multipliers are applied.</p>
        {/if}
        <div class="grid grid-cols-[8rem_1fr_1fr] gap-x-3 gap-y-1 text-sm">
          <span></span>
          <span class="text-[10px] uppercase tracking-wide text-slate-500">Content base</span>
          <span class="text-[10px] uppercase tracking-wide text-slate-500">Effective</span>
          <span class="text-slate-400">Max HP</span>
          <span>{view.stats.base ? view.stats.base.maxHitPoints : "—"}</span>
          <span>{view.stats.effective?.maxHitPoints ?? view.maxHitPoints}</span>
          <span class="text-slate-400">Attack</span>
          <span>{view.stats.base ? view.stats.base.attackPower : "—"}</span>
          <span>{view.stats.effective?.attackPower ?? "—"}</span>
          <span class="text-slate-400">Defense</span>
          <span>{view.stats.base ? view.stats.base.defense : "—"}</span>
          <span>{view.stats.effective?.defense ?? "—"}</span>
          <span class="text-slate-400">Attack speed</span>
          <span>{view.stats.attackSpeed || "default"}</span>
          <span>{swings(view.stats.attackSpeed)}</span>
        </div>
        {#if view.stats.factors}
          <p class="text-[11px] text-slate-500">{factorText(view.stats.factors)}</p>
        {:else if !view.stats.base}
          <p class="text-[11px] text-slate-500">No content base. These are the stored combat stats.</p>
        {/if}
        {#if view.behavior}
          <p class="text-xs text-slate-400">
            Aggro on sight: {view.behavior.aggroOnSight ? "yes" : "no"}
            {#if view.behavior.aggroRadius} · radius {view.behavior.aggroRadius}{/if}
            · Flee: {fleeText(view.behavior.fleeThreshold)}
            · XP {view.behavior.xpReward || 0}
            · Gold {view.behavior.goldMin || 0}–{view.behavior.goldMax || 0}
          </p>
        {/if}
      </div>
    {:else}
      <div class="card p-4">
        <div class="label-caps">Template</div>
        <p class="mt-2 text-sm text-slate-300">Max HP {view.maxHitPoints || 0}. No enemy trait.</p>
      </div>
    {/if}

    <div class="card p-4 space-y-2">
      <div class="label-caps">Dialogs</div>
      <p class="text-sm">
        Dialog
        {#if view.dialog}
          <a class="text-primary hover:underline" href={href("dialog", view.dialog.id)} on:click={(e) => open(e, "dialog", view.dialog.id)}>{view.dialog.name || view.dialog.id}</a>
          <span class="font-mono text-[10px] text-slate-500">{view.dialog.id}</span>
          {#if view.dialog.missing}<span class="text-amber-300">missing</span>{/if}
        {:else}
          <span class="text-slate-500">none</span>
        {/if}
      </p>
      <p class="text-sm">
        Idle dialog
        {#if view.idleDialog}
          <a class="text-primary hover:underline" href={href("dialog", view.idleDialog.id)} on:click={(e) => open(e, "dialog", view.idleDialog.id)}>{view.idleDialog.name || view.idleDialog.id}</a>
          <span class="font-mono text-[10px] text-slate-500">{view.idleDialog.id}</span>
          {#if view.idleDialog.missing}<span class="text-amber-300">missing</span>{/if}
        {:else}
          <span class="text-slate-500">none</span>
        {/if}
      </p>
    </div>

    {#if view.scripts?.length}
      <div class="card p-4 space-y-2">
        <div class="label-caps">Scripts</div>
        {#each view.scripts as hook}
          <div class="flex flex-wrap items-center justify-between gap-2 rounded-md border border-slate-800 px-3 py-2 text-sm">
            <span>
              {hook.hook}
              {#if hook.hook === "onLowHealth" && hook.threshold}
                <span class="text-slate-500">≤ {Math.round(hook.threshold * 100)}%</span>
              {/if}
            </span>
            {#if hook.id}
              <a class="text-primary hover:underline" href={href("script", hook.id)} on:click={(e) => open(e, "script", hook.id)}>{hook.name || hook.id}</a>
              <span class="font-mono text-[10px] text-slate-500">{hook.id}</span>
              {#if hook.missing}<span class="text-amber-300">missing</span>{/if}
            {:else}
              <span class="text-slate-500">—</span>
            {/if}
          </div>
        {/each}
      </div>
    {/if}

    {#if view.loot}
      <div class="card p-4 space-y-2">
        <div class="label-caps">Loot</div>
        {#if view.loot.tableId}
          <p class="text-sm">
            <span class="font-mono text-xs">{view.loot.tableId}</span>
            {#if view.loot.tableName}<span>{view.loot.tableName}</span>{/if}
            {#if view.loot.missing}<span class="text-amber-300">missing table</span>{/if}
          </p>
        {:else}
          <p class="text-sm text-slate-500">No loot table.</p>
        {/if}
        {#if view.loot.entries?.length}
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-[10px] uppercase tracking-wide text-slate-500">
                <th class="py-1 font-medium">Item</th>
                <th class="py-1 font-medium">Chance</th>
                <th class="py-1 font-medium">Notes</th>
              </tr>
            </thead>
            <tbody>
              {#each view.loot.entries as drop}
                <tr class="border-t border-slate-800">
                  <td class="py-1.5">
                    {#if drop.itemId}
                      <a class="text-primary hover:underline" href={href("item", drop.itemId)} on:click={(e) => open(e, "item", drop.itemId)}>{drop.itemName || drop.itemId}</a>
                      <span class="font-mono text-[10px] text-slate-500">{drop.itemId}</span>
                    {/if}
                    {#if drop.missing}<span class="text-amber-300">missing</span>{/if}
                  </td>
                  <td class="py-1.5">{percent(drop.dropChance, drop.guaranteed)}</td>
                  <td class="py-1.5 text-slate-400">
                    {#if drop.guaranteed}guaranteed{/if}
                    {#if drop.rarity}<span class="uppercase">{drop.rarity}</span>{/if}
                    {#if drop.bossOnly}boss only{/if}
                    {#if drop.minQuantity && drop.maxQuantity && (drop.minQuantity !== 1 || drop.maxQuantity !== 1)}
                      ×{drop.minQuantity}–{drop.maxQuantity}
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
        {#if view.loot.guaranteed?.length}
          <p class="text-xs text-slate-400">
            Always drops
            {#each view.loot.guaranteed as item}
              <a class="text-primary hover:underline" href={href("item", item.id)} on:click={(e) => open(e, "item", item.id)}>{item.name || item.id}</a>
            {/each}
          </p>
        {/if}
      </div>
    {/if}

    <div class="card p-4 space-y-2">
      <div class="label-caps">Spawners</div>
      {#if view.spawnRoom}
        <p class="text-sm">
          Spawn room
          <a class="text-primary hover:underline" href={href("room", view.spawnRoom.id)} on:click={(e) => open(e, "room", view.spawnRoom.id)}>{view.spawnRoom.name || view.spawnRoom.id}</a>
          <span class="font-mono text-[10px] text-slate-500">{view.spawnRoom.id}</span>
          {#if view.spawnRoom.missing}<span class="text-amber-300">missing</span>{/if}
        </p>
      {/if}
      {#if view.templateRespawn}
        <p class="text-xs text-slate-400">Template respawn {view.templateRespawn}.</p>
      {/if}
      {#if view.spawners?.length}
        {#each view.spawners as spawner}
          <div class="rounded-md border border-slate-800 px-3 py-2 text-sm">
            <div class="font-medium">{spawner.name || spawner.id} <span class="font-mono text-[10px] text-slate-500">{spawner.id}</span></div>
            <div class="text-xs text-slate-400">
              {#if spawner.roomId}
                <a class="text-primary hover:underline" href={href("room", spawner.roomId)} on:click={(e) => open(e, "room", spawner.roomId)}>{spawner.roomName || spawner.roomId}</a>
              {:else}
                no room
              {/if}
              · max {spawner.maxInstances}
              · {spawner.respawnTime ? `respawn ${spawner.respawnTime}` : "no respawn"}
              {#if spawner.spawnInterval} · interval {spawner.spawnInterval}{/if}
            </div>
          </div>
        {/each}
      {:else if (view.kinds || []).includes("enemy")}
        <p class="text-sm text-amber-200">No spawner. Once killed, this NPC stays dead until something respawns it.</p>
      {:else}
        <p class="text-sm text-slate-500">No spawner references this NPC.</p>
      {/if}
    </div>

    <div class="card p-4 space-y-2">
      <div class="label-caps">Quests</div>
      {#if view.quests?.length}
        {#each view.quests as quest}
          <p class="text-sm">
            <span class="rounded bg-slate-800 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-slate-300">{quest.role}</span>
            <a class="text-primary hover:underline" href={href("quest", quest.id)} on:click={(e) => open(e, "quest", quest.id)}>{quest.name || quest.id}</a>
            <span class="font-mono text-[10px] text-slate-500">{quest.id}</span>
            {#if quest.description}<span class="text-slate-400">· {quest.description}</span>{/if}
          </p>
        {/each}
      {:else}
        <p class="text-sm text-slate-500">No kill, talk, or deliver objective names this NPC.</p>
      {/if}
    </div>

    <div class="card p-4 space-y-3">
      <div class="label-caps">Live instances</div>
      {#if liveError}
        <p class="text-sm text-amber-200">{liveError}</p>
      {:else if !live.length}
        <p class="text-sm text-slate-400">No running instance.</p>
        {#if $userRole === "admin"}
          <OpsButtons mode="npc" npcTemplateId={view.id} npcName={view.name} roomId={spawnRoomId} onDone={refresh} />
        {:else}
          <p class="text-xs text-slate-500">Live ops are admin only.</p>
        {/if}
      {:else}
        {#each live as row (row.id)}
          <div class="rounded-md border border-slate-800 px-3 py-3 space-y-2">
            <div class="flex flex-wrap items-baseline justify-between gap-2 text-sm">
              <span class="font-mono text-xs">{row.id}</span>
              <span>
                {#if row.roomId}
                  <a class="text-primary hover:underline" href={href("room", row.roomId)} on:click={(e) => open(e, "room", row.roomId)}>{row.roomName || row.roomId}</a>
                {:else}
                  <span class="text-slate-500">no room</span>
                {/if}
              </span>
            </div>
            <div class="flex items-center gap-2 text-sm">
              <span class="w-10 text-slate-400">HP</span>
              <div class="h-2 flex-1 overflow-hidden rounded-full bg-slate-800">
                <div class="h-full bg-amber-500" style={`width:${hpWidth(row)}%`}></div>
              </div>
              <span>{row.hp}/{row.maxHp}</span>
            </div>
            <p class="text-xs text-slate-400">
              {#if row.dead}
                Dead
                {#if row.deadUntil} until {when(row.deadUntil)}{:else if row.noRespawn} · does not respawn{/if}
              {:else if row.inCombat}
                In combat{row.combatWith?.length ? ` with ${row.combatWith.join(", ")}` : ""}
              {:else}
                Not in combat
              {/if}
              {#if row.lastEvent} · {row.lastEvent}{/if}
            </p>
            {#if $userRole === "admin"}
              <OpsButtons
                mode="npc"
                npcInstanceId={row.id}
                npcTemplateId={view.id}
                npcName={row.name || view.name}
                roomId={spawnRoomId || row.roomId}
                onDone={refresh}
              />
            {/if}
          </div>
        {/each}
        {#if $userRole !== "admin"}
          <p class="text-xs text-slate-500">Live ops are admin only.</p>
        {/if}
      {/if}
    </div>
  {/if}
</div>
