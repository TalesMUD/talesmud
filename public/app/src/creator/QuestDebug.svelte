<script>
  import { onMount } from "svelte";
  import { navigateTo } from "yrv";
  import { getAuth } from "../auth.js";
  import { getQuestDebug, getQuests } from "../api/quests.js";
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import ConfirmDialog from "./ConfirmDialog.svelte";
  import { questColumns } from "./tableColumns.js";

  const { isAuthenticated, authToken } = getAuth();

  let questId = "";
  let quests = [];
  let questsLoaded = false;
  let report = null;
  let loading = false;
  let error = "";
  let selectedId = "";
  let resetOpen = false;
  let resetNote = "";

  const editorPaths = {
    room: "/creator/rooms",
    npc: "/creator/npcs",
    item: "/creator/item-templates",
    dialog: "/creator/dialogs",
    script: "/creator/scripts",
    quest: "/creator/quests/debug",
    characterTemplate: "/creator/character-templates",
  };

  function editorHref(type, id) {
    const base = editorPaths[type];
    if (!base || !id) return "";
    return `${base}?id=${encodeURIComponent(id)}`;
  }

  function openEditor(event, type, id) {
    const path = editorHref(type, id);
    if (!path) return;
    event.preventDefault();
    navigateTo(path);
  }

  function verdictLabel(verdict) {
    if (verdict === "red") return "Not completable";
    if (verdict === "amber") return "Unverified";
    return "Can complete";
  }

  function verdictClass(verdict) {
    if (verdict === "red") return "bg-red-500/15 text-red-200";
    if (verdict === "amber") return "bg-amber-500/15 text-amber-200";
    return "bg-emerald-500/15 text-emerald-200";
  }

  function dot(verdict) {
    if (verdict === "red") return "✕";
    if (verdict === "amber") return "!";
    return "✓";
  }

  function since(value) {
    if (!value) return "—";
    const then = new Date(value).getTime();
    if (Number.isNaN(then)) return "—";
    const days = Math.max(0, Math.floor((Date.now() - then) / 86400000));
    if (days === 0) return "today";
    return days === 1 ? "1 d" : `${days} d`;
  }

  function loadQuestList() {
    if (questsLoaded || !$authToken) return;
    questsLoaded = true;
    getQuests(
      $authToken,
      [],
      (data) => {
        quests = data || [];
      },
      () => {
        questsLoaded = false;
      }
    );
  }

  async function loadReport() {
    if (!$authToken || !questId) {
      report = null;
      return;
    }
    const id = questId;
    loading = true;
    error = "";
    resetNote = "";
    try {
      report = await getQuestDebug($authToken, id);
      selectedId = "";
    } catch (err) {
      report = null;
      error = err?.response?.data?.error || "Quest debug is unavailable.";
    } finally {
      loading = false;
    }
  }

  function editQuest(event) {
    event.preventDefault();
    if (!report?.id) return;
    navigateTo(`/creator/quests?id=${encodeURIComponent(report.id)}`);
  }

  function chooseQuest(event) {
    const id = event.detail || "";
    questId = id;
    report = null;
    error = "";
    selectedId = "";
    const path = id ? `/creator/quests/debug?id=${encodeURIComponent(id)}` : "/creator/quests/debug";
    navigateTo(path);
    if (id) loadReport();
  }

  function selectCharacter(row) {
    selectedId = row?.id || "";
    resetNote = "";
  }

  function askReset() {
    resetOpen = true;
  }

  function confirmReset() {
    resetOpen = false;
    const op = report?.ops?.questStep?.path || "/api/ops/quest-step";
    if (!selectedId) {
      resetNote = "Select a character in the live list first.";
      return;
    }
    resetNote = `Quest step reset is not connected. The live op is POST ${op}.`;
  }

  $: selected = (report?.characters || []).find((row) => row.id === selectedId) || null;

  onMount(() => {
    questId = new URLSearchParams(window.location.search).get("id") || "";
  });

  $: if ($isAuthenticated && $authToken) {
    loadQuestList();
    if (questId && !report && !loading && !error) {
      loadReport();
    }
  }
</script>

<div class="flex flex-col h-[calc(100vh-128px)]">
  <div class="px-6 pt-5 pb-3 flex-shrink-0">
    <div class="flex flex-col lg:flex-row lg:items-end justify-between gap-4">
      <div class="space-y-1 min-w-0">
        <div class="label-caps">Quest debugger</div>
        {#if report}
          <h1 class="text-2xl font-bold tracking-tight">
            {report.name || "Quest"}
            <span class="font-mono text-base text-slate-400">{report.id}</span>
          </h1>
          <p class="text-sm text-slate-400">
            {report.category || "quest"}
            {#if report.level} · lvl {report.level}{/if}
            {#if report.area} · {report.area}{/if}
            {#if report.offer?.reason} · {report.offer.reason}{/if}
            {#if report.rewards && (report.rewards.xp || report.rewards.gold || (report.rewards.items || []).length)}
              · {report.rewards.xp || 0} XP, {report.rewards.gold || 0} gold
            {/if}
          </p>
        {:else}
          <h1 class="text-2xl font-bold tracking-tight">Quest debugger</h1>
          <p class="text-sm text-slate-500">Step chain, reachability, and who is on the quest.</p>
        {/if}
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <EntitySelectButton
          value={questId}
          elements={quests}
          columns={questColumns}
          title="Select Quest"
          placeholder="Select a quest..."
          on:change={chooseQuest}
        />
        {#if report}
          <span class="rounded-full px-3 py-1 text-xs font-bold {verdictClass(report.verdict)}">{verdictLabel(report.verdict)}</span>
          <a class="btn btn-outline" href={`/creator/quests?id=${encodeURIComponent(report.id)}`} on:click={editQuest}>Edit quest</a>
        {/if}
      </div>
    </div>
  </div>

  <div class="flex-1 overflow-y-auto px-6 pb-6 thin-scrollbar">
    {#if !$isAuthenticated}
      <div class="py-12 text-center text-sm text-slate-500">Please log in to access Creator tools.</div>
    {:else if error}
      <div class="rounded-md border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-100">{error}</div>
    {:else if loading && !report}
      <div class="py-12 text-center text-sm text-slate-500">Checking this quest…</div>
    {:else if !questId}
      <div class="py-12 text-center text-sm text-slate-500">Choose a quest to see whether it can complete.</div>
    {:else if report}
      <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)] gap-4 items-start">
        <div class="space-y-4">
          <div class="card p-4">
            <div class="label-caps">Flow</div>
            <p class="mt-1 text-sm text-slate-400">{report.reason}</p>
            <div class="mt-3 space-y-2">
              {#each report.steps || [] as step, index (`${step.kind}-${step.objectiveId || index}`)}
                <div class="rounded-lg border border-slate-800 bg-slate-950/40 p-3">
                  <div class="flex gap-3">
                    <div class="mt-0.5 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-xs font-bold {verdictClass(step.verdict)}">{dot(step.verdict)}</div>
                    <div class="min-w-0">
                      <div class="font-semibold">{step.label}</div>
                      {#if step.reason}<div class="text-sm text-slate-400">{step.reason}</div>{/if}
                      {#if (step.places || []).length}
                        <ul class="mt-2 space-y-1">
                          {#each step.places as place, placeIndex (`${place.type}-${place.id}-${place.nodeId || ""}-${place.how}-${placeIndex}`)}
                            <li class="text-sm text-slate-300">
                              <span class={place.reachable ? "text-emerald-300" : "text-amber-300"}>{place.reachable ? "●" : "○"}</span>
                              {#if editorHref(place.type, place.id)}
                                <a class="font-mono text-emerald-300 hover:underline" href={editorHref(place.type, place.id)} on:click={(event) => openEditor(event, place.type, place.id)}>{place.id}</a>
                              {:else}
                                <span class="font-mono text-slate-400">{place.id || place.type}</span>
                              {/if}
                              {#if place.name && place.name !== place.id}<span class="text-slate-400"> {place.name}</span>{/if}
                              {#if place.nodeId}<span class="font-mono text-slate-500"> {place.nodeId}</span>{/if}
                              {#if place.how}<span class="text-slate-500"> · {place.how}</span>{/if}
                              {#if place.detail}<span class="text-slate-500"> · {place.detail}</span>{/if}
                            </li>
                          {/each}
                        </ul>
                      {/if}
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="card p-4">
            <div class="label-caps">Live progress · {(report.characters || []).length} characters</div>
            {#if (report.characters || []).length === 0}
              <p class="mt-3 text-sm text-slate-400">No characters are on this quest.</p>
            {:else}
              <div class="mt-3 overflow-x-auto">
                <table class="w-full text-sm">
                  <thead class="text-left text-xs uppercase tracking-wide text-slate-500">
                    <tr>
                      <th class="py-1 pr-3">Character</th>
                      <th class="py-1 pr-3">Status</th>
                      <th class="py-1 pr-3">Step</th>
                      <th class="py-1 pr-3">Since</th>
                      <th class="py-1">Where</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each report.characters as row (row.id)}
                      <tr class="cursor-pointer border-t border-slate-800 {selectedId === row.id ? "bg-slate-800/60" : ""}" on:click={() => selectCharacter(row)}>
                        <td class="py-2 pr-3">
                          <span class="font-semibold">{row.name}</span>
                          <span class="ml-1 rounded-full bg-slate-800 px-2 py-0.5 text-xs text-slate-300">lvl {row.level}</span>
                          {#if row.guest}<span class="ml-1 text-xs text-slate-500">guest</span>{/if}
                        </td>
                        <td class="py-2 pr-3 text-slate-300">{row.status || "—"}</td>
                        <td class="py-2 pr-3 font-mono text-slate-300">{row.step || "—"}</td>
                        <td class="py-2 pr-3 text-slate-400">{since(row.acceptedAt)}</td>
                        <td class="py-2 font-mono text-slate-400">{row.currentRoomId || "—"}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
            <div class="mt-3">
              <button class="btn btn-outline" type="button" data-op={report.ops?.questStep?.path || "/api/ops/quest-step"} on:click={askReset}>
                Reset quest step
              </button>
            </div>
            {#if resetNote}<p class="mt-2 text-xs text-slate-400">{resetNote}</p>{/if}
            <p class="mt-2 text-xs text-slate-500">Reset is a live op and is not sent from this panel.</p>
          </div>

          <div class="card p-4">
            <div class="label-caps">Prerequisites</div>
            {#if (report.prerequisites || []).length === 0}
              <p class="mt-2 text-sm text-slate-400">None.</p>
            {:else}
              <ul class="mt-2 space-y-1 text-sm">
                {#each report.prerequisites as link (link.id)}
                  <li>
                    <a class="font-mono text-emerald-300 hover:underline" href={editorHref("quest", link.id)} on:click={(event) => openEditor(event, "quest", link.id)}>{link.id}</a>
                    <span class="text-slate-400"> {link.name}</span>
                    <span class="ml-1 {verdictClass(link.verdict)} rounded-full px-2 py-0.5 text-xs">{verdictLabel(link.verdict)}</span>
                  </li>
                {/each}
              </ul>
            {/if}
            <div class="label-caps mt-4">Next quests</div>
            {#if (report.next || []).length === 0}
              <p class="mt-2 text-sm text-slate-400">None.</p>
            {:else}
              <ul class="mt-2 space-y-1 text-sm">
                {#each report.next as link (link.id)}
                  <li>
                    <a class="font-mono text-emerald-300 hover:underline" href={editorHref("quest", link.id)} on:click={(event) => openEditor(event, "quest", link.id)}>{link.id}</a>
                    <span class="text-slate-400"> {link.name}</span>
                    <span class="ml-1 {verdictClass(link.verdict)} rounded-full px-2 py-0.5 text-xs">{verdictLabel(link.verdict)}</span>
                  </li>
                {/each}
              </ul>
            {/if}
          </div>

          <div class="card p-4">
            <div class="label-caps">Offer and turn-in</div>
            <p class="mt-2 text-sm text-slate-300">{report.offer?.reason}</p>
            <p class="mt-1 text-sm text-slate-300">{report.turnIn?.reason}</p>
            <div class="label-caps mt-4">Rewards</div>
            <p class="mt-2 text-sm text-slate-300">{report.rewards?.xp || 0} XP · {report.rewards?.gold || 0} gold</p>
            {#each report.rewards?.items || [] as item (item.id)}
              <div class="mt-1 text-sm">
                {#if editorHref("item", item.id)}
                  <a class="font-mono text-emerald-300 hover:underline" href={editorHref("item", item.id)} on:click={(event) => openEditor(event, "item", item.id)}>{item.id}</a>
                {:else}
                  <span class="font-mono">{item.id}</span>
                {/if}
                <span class="text-slate-400"> {item.name}</span>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<ConfirmDialog
  open={resetOpen}
  title="Reset quest step?"
  entityType="quest"
  entityName={selected?.name || report?.name || ""}
  entityId={selected?.id || report?.id || ""}
  detail={selected ? `${selected.name} · ${selected.step || selected.status || "on this quest"}` : "Select a character in the live list first."}
  confirmLabel="Reset step"
  hint="This asks for POST /api/ops/quest-step. That live op is not connected, so nothing is sent."
  on:confirm={confirmReset}
  on:cancel={() => (resetOpen = false)}
/>
