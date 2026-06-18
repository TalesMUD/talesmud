<script>
  import { createEventDispatcher } from "svelte";
  import ValidationPanel from "./ValidationPanel.svelte";

  export let open = false;
  export let preview = null;

  const dispatch = createEventDispatcher();
  $: result = preview ? { issues: preview.issues || [] } : null;
</script>

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-6" on:click={() => dispatch("close")}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="w-full max-w-3xl rounded-lg border border-slate-700 bg-slate-950 shadow-2xl" on:click|stopPropagation>
      <div class="flex items-center justify-between border-b border-slate-800 px-5 py-3">
        <h2 class="text-sm font-semibold">Room Preview</h2>
        <button class="btn btn-ghost p-1.5" type="button" on:click={() => dispatch("close")}>
          <span class="material-symbols-outlined text-lg">close</span>
        </button>
      </div>
      <div class="space-y-4 p-5">
        <div>
          <div class="text-lg font-semibold">{preview?.name || "Untitled room"}</div>
          <div class="text-xs text-slate-400">{preview?.area || "No area"}</div>
        </div>
        <div class="rounded-md border border-slate-800 bg-slate-900/70 p-4 text-sm leading-relaxed text-slate-200">
          {preview?.description || "(no room description)"}
        </div>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
          <div class="rounded-md border border-slate-800 bg-slate-900/70 p-3"><div class="text-slate-500">Exits</div><div class="mt-1 text-xl font-semibold">{preview?.exitsCount || 0}</div></div>
          <div class="rounded-md border border-slate-800 bg-slate-900/70 p-3"><div class="text-slate-500">Actions</div><div class="mt-1 text-xl font-semibold">{preview?.actionsCount || 0}</div></div>
          <div class="rounded-md border border-slate-800 bg-slate-900/70 p-3"><div class="text-slate-500">Items</div><div class="mt-1 text-xl font-semibold">{preview?.itemsCount || 0}</div></div>
          <div class="rounded-md border border-slate-800 bg-slate-900/70 p-3"><div class="text-slate-500">NPCs</div><div class="mt-1 text-xl font-semibold">{preview?.npcsCount || 0}</div></div>
        </div>
        <ValidationPanel result={result} />
      </div>
    </div>
  </div>
{/if}
