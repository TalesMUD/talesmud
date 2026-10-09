<script>
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import { listAudit, undoAudit } from "../api/audit.js";
  import { showOpsToast } from "./opsToast.js";
  import { opError } from "./opsFlow.js";

  const { isAuthenticated, authToken } = getAuth();

  const entityTypes = [
    "",
    "rooms",
    "items",
    "npcs",
    "loottables",
    "spawners",
    "dialogs",
    "quests",
    "scripts",
    "skills",
    "character-templates",
    "settings",
    "character",
    "npc",
    "combat",
    "instance",
    "quest-progress",
  ];

  let rows = [];
  let loading = false;
  let error = "";
  let entityType = "";
  let entityId = "";
  let textFilter = "";
  let limit = "100";
  let openId = "";
  let undoing = "";
  let loaded = false;

  $: visible = rows.filter((row) => {
    if (!textFilter) return true;
    const haystack = `${row.summary || ""} ${row.action || ""} ${row.actorName || ""} ${row.entityId || ""}`.toLowerCase();
    return haystack.includes(textFilter.toLowerCase());
  });

  function when(value) {
    if (!value || String(value).startsWith("0001-")) return "—";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "—";
    return date.toLocaleString();
  }

  function pretty(value) {
    if (value == null || value === "") return "—";
    try {
      const parsed = typeof value === "string" ? JSON.parse(value) : value;
      return JSON.stringify(parsed, null, 2);
    } catch {
      return typeof value === "string" ? value : String(value);
    }
  }

  async function load() {
    if (!$authToken) return;
    loading = true;
    error = "";
    try {
      const params = { limit };
      if (entityType) params.entityType = entityType;
      if (entityId.trim()) params.entityId = entityId.trim();
      const data = await listAudit($authToken, params);
      rows = Array.isArray(data) ? data : [];
      loaded = true;
    } catch (err) {
      error = opError(err, "The audit log is unavailable.");
      rows = [];
    } finally {
      loading = false;
    }
  }

  async function undo(row) {
    if (!$authToken || undoing) return;
    undoing = row.id;
    error = "";
    try {
      const data = await undoAudit($authToken, row.id);
      showOpsToast({
        summary: data.summary || "Undone.",
        auditId: data.auditId,
        undoable: !!data.undoable,
        token: $authToken,
      });
      await load();
    } catch (err) {
      error = opError(err, "Undo failed.");
    } finally {
      undoing = "";
    }
  }

  onMount(load);

  $: if ($isAuthenticated && $authToken && !loaded && !loading) {
    load();
  }
</script>

<div class="flex flex-col h-[calc(100vh-128px)]">
  <div class="px-6 pt-5 pb-3 flex-shrink-0">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
      <div class="space-y-1">
        <h1 class="text-2xl font-bold tracking-tight">Audit log</h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm">Creator writes and live ops, newest first. Undo is admin only.</p>
      </div>
      <button class="btn btn-outline" type="button" on:click={load} disabled={loading}>
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
      <div class="grid grid-cols-1 md:grid-cols-4 gap-3">
        <select class="input-base" bind:value={entityType} on:change={load}>
          {#each entityTypes as type}
            <option value={type}>{type || "All entity types"}</option>
          {/each}
        </select>
        <input class="input-base" type="search" placeholder="Entity id" bind:value={entityId} on:change={load} />
        <input class="input-base" type="search" placeholder="Filter summary, action, actor..." bind:value={textFilter} />
        <select class="input-base" bind:value={limit} on:change={load}>
          <option value="50">50</option>
          <option value="100">100</option>
          <option value="250">250</option>
          <option value="500">500</option>
        </select>
      </div>
    </div>

    <div class="card overflow-x-auto">
      <table class="w-full text-sm" style="min-width: 860px;">
        <thead>
          <tr class="text-left text-[10px] uppercase tracking-wider text-slate-500">
            <th class="px-3 py-2">When</th>
            <th class="px-3 py-2">Actor</th>
            <th class="px-3 py-2">Action</th>
            <th class="px-3 py-2">Entity</th>
            <th class="px-3 py-2">Summary</th>
            <th class="px-3 py-2"></th>
          </tr>
        </thead>
        <tbody>
          {#if loading && rows.length === 0}
            <tr><td class="px-3 py-6 text-slate-500" colspan="6">Loading…</td></tr>
          {:else if visible.length === 0}
            <tr><td class="px-3 py-6 text-slate-500" colspan="6">No audit rows.</td></tr>
          {:else}
            {#each visible as row}
              <tr class="border-t border-slate-800 align-top">
                <td class="px-3 py-2 whitespace-nowrap">{when(row.time)}</td>
                <td class="px-3 py-2">{row.actorName || row.actorUserId || "—"}</td>
                <td class="px-3 py-2">
                  <div>{row.action}</div>
                  <div class="text-[10px] uppercase tracking-wider text-slate-500">{row.source}{row.undoneBy ? " · undone" : ""}</div>
                </td>
                <td class="px-3 py-2">
                  <div>{row.entityType}</div>
                  <div class="font-mono text-xs text-slate-500 break-all">{row.entityId}</div>
                </td>
                <td class="px-3 py-2">
                  <div>{row.summary || "—"}</div>
                  <button class="text-xs text-primary mt-1" type="button" on:click={() => (openId = openId === row.id ? "" : row.id)}>
                    {openId === row.id ? "Hide JSON" : "Before / after"}
                  </button>
                  {#if openId === row.id}
                    <div class="grid grid-cols-1 lg:grid-cols-2 gap-2 mt-2">
                      <pre class="audit-json">{pretty(row.before)}</pre>
                      <pre class="audit-json">{pretty(row.after)}</pre>
                    </div>
                  {/if}
                </td>
                <td class="px-3 py-2 whitespace-nowrap">
                  {#if row.undoable && !row.undoneBy}
                    <button class="btn btn-outline text-xs" type="button" disabled={undoing === row.id} on:click={() => undo(row)}>
                      {undoing === row.id ? "Undoing…" : "Undo"}
                    </button>
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</div>

<style>
  .audit-json {
    max-height: 240px;
    overflow: auto;
    margin: 0;
    padding: 8px;
    border-radius: 6px;
    background: #0f172a;
    color: #cbd5e1;
    font-size: 11px;
    line-height: 1.4;
    white-space: pre-wrap;
    word-break: break-word;
  }
</style>
