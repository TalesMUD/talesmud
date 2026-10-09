<script>
  import { createEventDispatcher } from "svelte";
  import { get } from "svelte/store";
  import { getAuth } from "../auth.js";
  import ConfirmDialog from "./ConfirmDialog.svelte";
  import { opError, performOp } from "./opsFlow.js";

  /** "npc", "quest", or "character". */
  export let mode = "character";
  export let characterId = "";
  export let npcInstanceId = "";
  export let npcTemplateId = "";
  export let npcName = "";
  export let roomId = "";
  /** Quest log entries from the live character detail. */
  export let quests = [];
  /** Called after a successful op so the parent can refresh. */
  export let onDone = null;

  const { authToken } = getAuth();
  const dispatch = createEventDispatcher();

  let pending = null;
  let busy = false;
  let error = "";

  const undoHint = "This is written to the audit log. Undo from the toast if the world has not moved on.";
  const noUndoHint = "This cannot be undone.";

  function ask(job) {
    error = "";
    pending = job;
  }

  function characterJobs() {
    return [
      {
        label: "End combat",
        title: "End this fight?",
        entityName: characterId,
        detail: "Stop the fight with no rewards, penalties, or healing.",
        hint: noUndoHint,
        action: "end-combat",
        body: { characterId },
      },
      {
        label: "Re-grant starter kit",
        title: "Re-grant starter kit?",
        entityName: characterId,
        detail: "Add missing class kit pieces. Gear the character already holds stays put.",
        hint: "Undo removes only the pieces this action added.",
        action: "regrant-starter-kit",
        body: { characterId },
      },
    ];
  }

  function npcJobs() {
    const name = npcName || npcInstanceId || npcTemplateId;
    const jobs = [];
    if (npcInstanceId) {
      jobs.push(
        {
          label: "Heal",
          title: "Heal this NPC?",
          entityName: name,
          entityId: npcInstanceId,
          detail: "Set current hit points to maximum.",
          hint: "Undo restores the previous hit points if the NPC is still alive and not in a fight.",
          action: "npc-heal",
          body: { npcInstanceId },
        },
        {
          label: "Despawn",
          title: "Despawn this NPC?",
          entityName: name,
          entityId: npcInstanceId,
          detail: "Remove the running instance. A persisted unique is marked dead and kept as content.",
          hint: "Undo puts the instance back. A fight must be ended first.",
          action: "npc-despawn",
          body: { npcInstanceId },
        },
        {
          label: "End combat",
          title: "End this NPC's fight?",
          entityName: name,
          entityId: npcInstanceId,
          detail: "Stop the fight with no rewards, penalties, or healing.",
          hint: noUndoHint,
          action: "end-combat",
          body: { npcInstanceId },
        },
      );
    }
    if (npcTemplateId) {
      const body = { npcTemplateId };
      if (roomId) body.roomId = roomId;
      jobs.push({
        label: "Respawn",
        title: "Respawn this NPC?",
        entityName: name,
        entityId: npcTemplateId,
        detail: roomId
          ? `Spawn it in ${roomId} only if it is not already alive.`
          : "Spawn it in its spawn room only if it is not already alive.",
        hint: "Undo despawns an instance this action created. Registering an already-living unique cannot be undone.",
        action: "npc-respawn",
        body,
      });
    }
    return jobs;
  }

  function questLabel(quest) {
    return quest?.questName || quest?.questId || "quest";
  }

  async function runPending() {
    const job = pending;
    pending = null;
    if (!job) return;
    const token = get(authToken);
    if (!token) {
      error = "Sign in as an admin to run live ops.";
      return;
    }
    busy = true;
    error = "";
    try {
      await performOp(token, job.action, job.body);
      if (onDone) await onDone(job);
      dispatch("done", job);
    } catch (err) {
      error = opError(err);
    } finally {
      busy = false;
    }
  }
</script>

<div class="space-y-3">
  {#if error}
    <div class="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm text-red-100">{error}</div>
  {/if}

  {#if mode === "character"}
    <div class="flex flex-wrap gap-2">
      {#each characterJobs() as job}
        <button class="btn btn-outline text-xs" type="button" disabled={busy || !characterId} on:click={() => ask(job)}>
          {job.label}
        </button>
      {/each}
    </div>
  {:else if mode === "npc"}
    <div class="flex flex-wrap gap-2">
      {#each npcJobs() as job}
        <button class="btn btn-outline text-xs" type="button" disabled={busy} on:click={() => ask(job)}>
          {job.label}
        </button>
      {/each}
    </div>
  {:else if mode === "quest"}
    {#if !quests?.length}
      <p class="text-sm text-slate-500">No quest progress.</p>
    {:else}
      <div class="space-y-3">
        {#each quests as quest}
          <div class="rounded-md border border-slate-700 p-3 space-y-2">
            <div class="flex flex-wrap items-baseline justify-between gap-2">
              <div>
                <div class="font-semibold">{questLabel(quest)}</div>
                <div class="text-xs text-slate-500">{quest.status}{quest.readyToTurnIn ? " · ready to turn in" : ""}</div>
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  class="btn btn-outline text-xs"
                  type="button"
                  disabled={busy || !characterId}
                  on:click={() => ask({
                    label: "Complete quest",
                    title: "Complete this quest?",
                    entityName: questLabel(quest),
                    entityId: quest.questId,
                    detail: "Mark every objective done and grant the quest rewards.",
                    hint: "Undo restores the previous progress. Rewards already granted stay with the character.",
                    action: "quest-step",
                    body: { characterId, questId: quest.questId, op: "complete" },
                  })}
                >Complete quest</button>
                <button
                  class="btn btn-outline text-xs"
                  type="button"
                  disabled={busy || !characterId}
                  on:click={() => ask({
                    label: "Abandon",
                    title: "Abandon this quest?",
                    entityName: questLabel(quest),
                    entityId: quest.questId,
                    detail: "Mark the quest abandoned.",
                    hint: undoHint,
                    action: "quest-step",
                    body: { characterId, questId: quest.questId, op: "abandon" },
                  })}
                >Abandon</button>
                <button
                  class="btn btn-outline text-xs"
                  type="button"
                  disabled={busy || !characterId}
                  on:click={() => ask({
                    label: "Reset quest",
                    title: "Reset this quest?",
                    entityName: questLabel(quest),
                    entityId: quest.questId,
                    detail: "Delete the quest progress row.",
                    hint: undoHint,
                    action: "quest-step",
                    body: { characterId, questId: quest.questId, op: "reset-quest" },
                  })}
                >Reset quest</button>
              </div>
            </div>
            {#each quest.objectives || [] as objective}
              <div class="flex flex-wrap items-center justify-between gap-2 text-sm">
                <span>
                  {objective.description || objective.objectiveId}
                  <span class="text-slate-500">({objective.current}/{objective.required})</span>
                </span>
                <span class="flex gap-2">
                  <button
                    class="btn btn-outline text-xs"
                    type="button"
                    disabled={busy || !characterId}
                    on:click={() => ask({
                      label: "Mark done",
                      title: "Mark this step done?",
                      entityName: objective.description || objective.objectiveId,
                      entityId: quest.questId,
                      detail: "Advance this objective only. Rewards are not granted.",
                      hint: undoHint,
                      action: "quest-step",
                      body: { characterId, questId: quest.questId, objectiveId: objective.objectiveId, op: "complete" },
                    })}
                  >Mark done</button>
                  <button
                    class="btn btn-outline text-xs"
                    type="button"
                    disabled={busy || !characterId}
                    on:click={() => ask({
                      label: "Reset step",
                      title: "Reset this step?",
                      entityName: objective.description || objective.objectiveId,
                      entityId: quest.questId,
                      detail: "Set this objective's progress back to zero.",
                      hint: undoHint,
                      action: "quest-step",
                      body: { characterId, questId: quest.questId, objectiveId: objective.objectiveId, op: "reset" },
                    })}
                  >Reset step</button>
                </span>
              </div>
            {/each}
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<ConfirmDialog
  open={!!pending}
  title={pending?.title || "Run live op?"}
  entityName={pending?.entityName || ""}
  entityId={pending?.entityId || ""}
  detail={pending?.detail || ""}
  hint={pending?.hint || undoHint}
  confirmLabel={pending?.label || "Run"}
  tone="ops"
  on:confirm={runPending}
  on:cancel={() => (pending = null)}
/>
