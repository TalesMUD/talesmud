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
        <h2 class="text-sm font-semibold">Merchant Preview</h2>
        <button class="btn btn-ghost p-1.5" type="button" on:click={() => dispatch("close")}>
          <span class="material-symbols-outlined text-lg">close</span>
        </button>
      </div>
      <div class="space-y-4 p-5">
        <div class="text-lg font-semibold">{preview?.name || "Untitled merchant"}</div>
        <div class="overflow-hidden rounded-md border border-slate-800">
          <table class="w-full text-xs">
            <thead class="bg-slate-900/80 text-slate-500">
              <tr>
                <th class="px-3 py-2 text-left">Item</th>
                <th class="px-3 py-2 text-left">Stock</th>
                <th class="px-3 py-2 text-left">Buy Price</th>
                <th class="px-3 py-2 text-left">Level</th>
              </tr>
            </thead>
            <tbody>
              {#each preview?.stock || [] as row}
                <tr class="border-t border-slate-800">
                  <td class="px-3 py-2 text-slate-200">{row.itemName}</td>
                  <td class="px-3 py-2 text-slate-400">{row.quantity === -1 ? "Unlimited" : row.quantity}/{row.maxQuantity || "-"}</td>
                  <td class="px-3 py-2 text-slate-400">{row.buyPrice}</td>
                  <td class="px-3 py-2 text-slate-400">{row.requiredLevel || 1}</td>
                </tr>
              {:else}
                <tr><td colspan="4" class="px-3 py-8 text-center text-slate-500">No merchant stock configured.</td></tr>
              {/each}
            </tbody>
          </table>
        </div>
        <ValidationPanel result={result} />
      </div>
    </div>
  </div>
{/if}
