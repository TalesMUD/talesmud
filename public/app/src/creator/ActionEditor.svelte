<script>
  import { onMount } from "svelte";

  export let action;
  export let deleteAction;

  onMount(async () => {
    if (action.params === null) {
      action.params = new Map();
    }
  });
</script>

<div class="p-4 rounded-lg border border-slate-200 dark:border-slate-700/50 bg-slate-50 dark:bg-slate-900/50 space-y-3">
  <div class="flex items-center justify-between">
    <div class="text-xs font-bold uppercase tracking-wider text-slate-400">
      {action.name || "Room Action"}
    </div>
    <button
      class="text-xs text-accent-red hover:underline"
      type="button"
      on:click={() => deleteAction(action)}
    >
      Remove
    </button>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <div class="space-y-1.5">
      <label class="label-caps" for={`action-name-${action.name}`}>Name</label>
      <input
        class="input-base text-xs"
        id={`action-name-${action.name}`}
        type="text"
        bind:value={action.name}
      />
    </div>
    <div class="space-y-1.5">
      <label class="label-caps" for={`action-type-${action.name}`}>Type</label>
      <select class="input-base text-xs" id={`action-type-${action.name}`} bind:value={action.type}>
        <option value="response">Respond to Player</option>
        <option value="response_room">Respond to Room</option>
        <option value="script">Run Script</option>
      </select>
    </div>
  </div>

  <div class="space-y-1.5">
    <label class="label-caps" for={`action-desc-${action.name}`}>Description</label>
    <input
      class="input-base text-xs"
      id={`action-desc-${action.name}`}
      type="text"
      bind:value={action.description}
    />
  </div>

  <div class="space-y-1.5">
    <label class="label-caps" for={`action-response-${action.name}`}>Response</label>
    <input
      class="input-base text-xs"
      id={`action-response-${action.name}`}
      type="text"
      bind:value={action.response}
    />
  </div>

  {#if action.type === "script"}
    <div class="space-y-1.5">
      <label class="label-caps" for={`action-script-${action.name}`}>Script ID</label>
      <input
        class="input-base text-xs"
        id={`action-script-${action.name}`}
        type="text"
        bind:value={action.scriptId}
      />
    </div>
  {/if}
</div>
