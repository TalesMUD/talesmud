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
    <div class="w-full max-w-2xl rounded-lg border border-slate-700 bg-slate-950 shadow-2xl" on:click|stopPropagation>
      <div class="flex items-center justify-between border-b border-slate-800 px-5 py-3">
        <div>
          <h2 class="text-sm font-semibold">Dialog Preview</h2>
          <p class="text-[10px] font-mono text-slate-500">{preview?.nodeId || "main"}</p>
        </div>
        <button class="btn btn-ghost p-1.5" type="button" on:click={() => dispatch("close")}>
          <span class="material-symbols-outlined text-lg">close</span>
        </button>
      </div>
      <div class="space-y-4 p-5">
        <div class="text-lg font-semibold">{preview?.name || "Untitled dialog"}</div>
        <div class="rounded-md border border-slate-800 bg-slate-900/70 p-4 text-sm leading-relaxed text-slate-200">
          {preview?.text || "(no dialog text)"}
        </div>
        <div class="text-xs text-slate-400">Options: {preview?.optionsCount || 0}</div>
        <ValidationPanel result={result} />
      </div>
    </div>
  </div>
{/if}
