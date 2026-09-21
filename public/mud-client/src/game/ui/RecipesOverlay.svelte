<script>
  import { onDestroy } from 'svelte';
  import { itemArtSrc, onItemArtError } from '../itemArtSrc.js';

  export let store;
  export let sendMessage;

  let errorText = '';
  let escHandler = null;
  let selectedKey = '';

  $: recipesPayload = $store.recipes || null;
  $: recipes = recipesPayload?.recipes || [];
  $: inventory = $store.inventory || [];

  // Live have-counts from current inventory (refresh after craft / pickup).
  $: rows = recipes.map((r) => enrichRecipe(r, inventory));
  $: selected = selectedKey ? (rows.find((r) => r.key === selectedKey) || null) : null;

  $: if (recipesPayload && !escHandler) {
    escHandler = (e) => {
      if (e.key !== 'Escape') return;
      if (selectedKey) {
        e.preventDefault();
        e.stopPropagation();
        selectedKey = '';
        return;
      }
      e.preventDefault();
      close();
    };
    if (typeof window !== 'undefined') window.addEventListener('keydown', escHandler, true);
  }

  $: if (!recipesPayload && escHandler) {
    if (typeof window !== 'undefined') window.removeEventListener('keydown', escHandler, true);
    escHandler = null;
    selectedKey = '';
    errorText = '';
  }

  onDestroy(() => {
    if (escHandler && typeof window !== 'undefined') {
      window.removeEventListener('keydown', escHandler, true);
    }
    escHandler = null;
  });

  $: if ($store.craftError) {
    errorText = $store.craftError;
    if (store?.clearCraftError) store.clearCraftError();
  }

  function countHave(inv, templateId) {
    if (!templateId) return 0;
    let n = 0;
    for (const item of inv || []) {
      const tid = item?.templateId || item?.id || '';
      if (tid === templateId || (item?.id && String(item.id).startsWith(templateId))) {
        n += Number(item.quantity || 1);
      }
    }
    return n;
  }

  function enrichRecipe(recipe, inv) {
    const ingredients = (recipe.ingredients || []).map((ing) => {
      const have = countHave(inv, ing.item);
      return {
        ...ing,
        have,
        haveEnough: have >= Number(ing.qty || 1),
      };
    });
    const haveAll = ingredients.every((i) => i.haveEnough);
    // Empty station always OK; otherwise trust server stationOk for current room.
    const stationOK = !recipe.station || !!recipe.stationOk;
    return {
      ...recipe,
      ingredients,
      canCraft: stationOK && haveAll,
      _haveAll: haveAll,
      _stationOK: stationOK,
    };
  }

  function close() {
    errorText = '';
    selectedKey = '';
    if (store?.clearRecipes) store.clearRecipes();
  }

  function selectRecipe(row) {
    if (selectedKey === row.key) {
      selectedKey = '';
      return;
    }
    selectedKey = row.key;
    errorText = '';
  }

  function craftRecipe(row) {
    errorText = '';
    if (!row) return;
    if (!row._stationOK) {
      errorText = row.stationHint || `Requires a ${row.station || 'station'}.`;
      return;
    }
    if (!row._haveAll) {
      const missing = (row.ingredients || [])
        .filter((i) => !i.haveEnough)
        .map((i) => `${i.qty}x ${i.name} (have ${i.have})`)
        .join('; ');
      errorText = `Missing materials: ${missing}`;
      return;
    }
    const key = row.key || row.name;
    sendMessage(`craft ${key}`);
  }

  function stationBadge(row) {
    if (!row.station) return 'Anywhere';
    return String(row.stationLabel || row.station);
  }

  function formatCategory(cat) {
    if (!cat) return '';
    return String(cat).replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
  }

  function artFor(itemLike) {
    if (!itemLike) return itemArtSrc({});
    if (itemLike.image) return itemLike.image;
    return itemArtSrc({
      templateId: itemLike.item || itemLike.templateId,
      id: itemLike.item || itemLike.id,
      name: itemLike.name,
      type: itemLike.type,
    });
  }
</script>

<style>
  .recipes-overlay {
    position: absolute;
    inset: 0;
    z-index: 120;
    background: rgba(0, 0, 0, 0.82);
    backdrop-filter: blur(6px);
    display: flex;
    flex-direction: column;
    padding: 1em;
    overflow: hidden;
  }
  .recipes-panel {
    flex: 1;
    min-height: 0;
    max-width: 640px;
    width: 100%;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    background: rgba(12, 16, 24, 0.97);
    border: 1px solid rgba(212, 175, 55, 0.28);
    border-radius: 10px;
    overflow: hidden;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.55);
    position: relative;
  }
  .recipes-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
    padding: 0.75em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: rgba(20, 26, 36, 0.9);
    flex-shrink: 0;
  }
  .recipes-title {
    flex: 1;
    font-weight: 700;
    color: #f8fafc;
    letter-spacing: 0.02em;
    display: flex;
    align-items: center;
    gap: 0.4em;
  }
  .recipes-title i { color: #fbbf24; font-size: 1.2em; }
  .recipes-close {
    border: none;
    background: transparent;
    color: #94a3b8;
    font-size: 1.4em;
    cursor: pointer;
    line-height: 1;
  }
  .recipes-hint {
    padding: 0.45em 1em 0;
    color: #94a3b8;
    font-size: 0.8em;
  }
  .recipes-error {
    margin: 0.5em 0.85em 0;
    padding: 0.5em 0.7em;
    border-radius: 6px;
    background: rgba(127, 29, 29, 0.45);
    border: 1px solid rgba(248, 113, 113, 0.4);
    color: #fecaca;
    font-size: 0.85em;
  }
  .recipes-list {
    flex: 1;
    min-height: 0;
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 0.45em;
    padding: 0.7em 0.85em 1em;
  }
  .recipe-row {
    display: grid;
    grid-template-columns: 56px 1fr auto;
    gap: 0.65em;
    align-items: center;
    text-align: left;
    padding: 0.55em 0.65em;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.18);
    background: rgba(255, 255, 255, 0.03);
    color: #e2e8f0;
    cursor: pointer;
  }
  .recipe-row:hover {
    border-color: rgba(96, 165, 250, 0.45);
    background: rgba(59, 130, 246, 0.1);
  }
  .recipe-row.selected {
    border-color: rgba(251, 191, 36, 0.65);
    background: rgba(251, 191, 36, 0.1);
  }
  .recipe-row.blocked {
    opacity: 0.72;
  }
  .recipe-icon {
    width: 56px;
    height: 56px;
    border-radius: 6px;
    background: #0b1119;
    border: 1px solid rgba(148, 163, 184, 0.25);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .recipe-icon img {
    width: 48px;
    height: 48px;
    object-fit: contain;
    image-rendering: pixelated;
  }
  .recipe-body { min-width: 0; display: flex; flex-direction: column; gap: 0.2em; }
  .recipe-name {
    font-weight: 700;
    font-size: 0.92em;
    color: #f8fafc;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .recipe-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35em;
    font-size: 0.72em;
  }
  .chip {
    border-radius: 999px;
    padding: 0.12em 0.5em;
    border: 1px solid rgba(148, 163, 184, 0.28);
    color: #cbd5e1;
    background: rgba(255, 255, 255, 0.04);
  }
  .chip.forge {
    border-color: rgba(248, 113, 113, 0.45);
    color: #fecaca;
    background: rgba(127, 29, 29, 0.25);
  }
  .chip.ok {
    border-color: rgba(34, 197, 94, 0.4);
    color: #86efac;
  }
  .chip.warn {
    border-color: rgba(245, 158, 11, 0.45);
    color: #fcd34d;
  }
  .recipe-ings {
    font-size: 0.75em;
    color: #94a3b8;
    line-height: 1.35;
  }
  .ing-ok { color: #86efac; }
  .ing-miss { color: #f87171; }
  .craft-btn {
    border: 1px solid rgba(34, 197, 94, 0.4);
    background: rgba(34, 197, 94, 0.14);
    color: #86efac;
    border-radius: 6px;
    padding: 0.45em 0.75em;
    cursor: pointer;
    font-weight: 700;
    font-size: 0.82em;
    font-family: inherit;
    white-space: nowrap;
  }
  .craft-btn:hover:not(:disabled) { background: rgba(34, 197, 94, 0.24); }
  .craft-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    border-color: rgba(148, 163, 184, 0.25);
    background: rgba(255, 255, 255, 0.04);
    color: #94a3b8;
  }
  .recipes-empty {
    text-align: center;
    color: #94a3b8;
    padding: 2em 1em;
  }
  .detail-backdrop {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    z-index: 5;
  }
  .detail-overlay {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: calc(100% - 2em);
    max-width: 360px;
    max-height: calc(100% - 2em);
    overflow-y: auto;
    background: rgba(12, 16, 24, 0.98);
    border: 1px solid rgba(212, 175, 55, 0.35);
    border-radius: 10px;
    padding: 1em;
    z-index: 6;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.7);
  }
  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 0.75em;
    padding-bottom: 0.6em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
  }
  .detail-title-row { display: flex; gap: 0.6em; align-items: center; min-width: 0; }
  .detail-art-wrap {
    width: 56px; height: 56px; flex-shrink: 0;
    display: flex; align-items: center; justify-content: center;
    background: rgba(0,0,0,0.35); border-radius: 8px;
    border: 1px solid rgba(255,255,255,0.1);
  }
  .detail-art { width: 48px; height: 48px; object-fit: contain; image-rendering: pixelated; }
  .detail-name { font-size: 1em; font-weight: 700; color: #f8fafc; }
  .detail-close {
    border: none; background: transparent; color: #64748b;
    font-size: 1.2em; cursor: pointer; line-height: 1;
  }
  .detail-desc {
    font-size: 0.85em; color: #94a3b8; font-style: italic;
    margin-bottom: 0.65em; line-height: 1.45;
  }
  .detail-section-label {
    font-size: 0.72em; text-transform: uppercase; letter-spacing: 0.04em;
    color: #64748b; margin-bottom: 0.35em; font-weight: 700;
  }
  .ing-list { display: flex; flex-direction: column; gap: 0.35em; margin-bottom: 0.75em; }
  .ing-row {
    display: flex; align-items: center; gap: 0.5em;
    padding: 0.35em 0.45em; border-radius: 6px;
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(148,163,184,0.15);
  }
  .ing-row img { width: 28px; height: 28px; object-fit: contain; image-rendering: pixelated; }
  .ing-name { flex: 1; font-size: 0.85em; color: #e2e8f0; }
  .ing-count { font-size: 0.8em; font-weight: 700; font-variant-numeric: tabular-nums; }
  .station-box {
    font-size: 0.82em; color: #cbd5e1; margin-bottom: 0.75em;
    padding: 0.45em 0.55em; border-radius: 6px;
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(148,163,184,0.18);
  }
  .detail-actions { display: flex; gap: 0.4em; flex-wrap: wrap; }
  .detail-action-btn {
    display: flex; align-items: center; gap: 0.3em;
    padding: 0.5em 0.85em; border-radius: 6px;
    border: 1px solid rgba(148,163,184,0.25);
    background: rgba(255,255,255,0.04); color: #e2e8f0;
    cursor: pointer; font-size: 0.9em; font-family: inherit; font-weight: 600;
  }
  .detail-action-btn.craft {
    color: #86efac; border-color: rgba(34,197,94,0.4); background: rgba(34,197,94,0.12);
  }
  .detail-action-btn.craft:disabled { opacity: 0.4; cursor: not-allowed; }
  .detail-action-btn.cancel { color: #94a3b8; }

  @media screen and (max-width: 520px) {
    .recipes-panel { max-width: none; }
    .recipe-row { grid-template-columns: 48px 1fr; }
    .craft-btn { grid-column: 1 / -1; }
    .detail-overlay {
      top: auto; bottom: 0; left: 0; right: 0; transform: none;
      width: 100%; max-width: none; max-height: 75%;
      border-radius: 12px 12px 0 0;
    }
  }
</style>

{#if recipesPayload}
  <div class="recipes-overlay" role="dialog" aria-label="Crafting recipes">
    <div class="recipes-panel">
      <div class="recipes-header">
        <div class="recipes-title">
          <i class="material-icons">construction</i>
          Crafting Recipes
        </div>
        <button class="recipes-close" type="button" on:click={close} aria-label="Close recipes">×</button>
      </div>
      <div class="recipes-hint">Everyone can craft — no profession required. Use Craft when you have the materials.</div>
      {#if errorText}
        <div class="recipes-error">{errorText}</div>
      {/if}
      <div class="recipes-list">
        {#each rows as row (row.key || row.id || row.name)}
          {@const blocked = !row.canCraft}
          {@const isSelected = selected && selected.key === row.key}
          <button
            class="recipe-row"
            class:blocked={blocked}
            class:selected={isSelected}
            type="button"
            on:click={() => selectRecipe(row)}
            aria-pressed={isSelected}
          >
            <div class="recipe-icon">
              <img src={artFor(row.output)} alt="" on:error={(e) => onItemArtError(e, { templateId: row.output?.item, name: row.output?.name })} />
            </div>
            <div class="recipe-body">
              <div class="recipe-name">{row.name}</div>
              <div class="recipe-meta">
                {#if row.category}
                  <span class="chip">{formatCategory(row.category)}</span>
                {/if}
                <span class="chip" class:forge={!!row.station} class:ok={!row.station}>
                  {stationBadge(row)}
                </span>
                {#if !row._stationOK}
                  <span class="chip warn">Wrong station</span>
                {:else if !row._haveAll}
                  <span class="chip warn">Need mats</span>
                {:else}
                  <span class="chip ok">Ready</span>
                {/if}
              </div>
              <div class="recipe-ings">
                {#each row.ingredients || [] as ing, i}
                  <span class={ing.haveEnough ? 'ing-ok' : 'ing-miss'}>{ing.have}/{ing.qty} {ing.name}</span>{#if i < (row.ingredients.length - 1)}<span>, </span>{/if}
                {/each}
                <span> → {row.output?.name || '?'}</span>
              </div>
            </div>
            <button
              class="craft-btn"
              type="button"
              disabled={!row.canCraft}
              on:click|stopPropagation={() => craftRecipe(row)}
            >Craft</button>
          </button>
        {:else}
          <div class="recipes-empty">No recipes known yet.</div>
        {/each}
      </div>

      {#if selected}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div class="detail-backdrop" on:click={() => (selectedKey = '')}></div>
        <div class="detail-overlay" role="dialog" aria-label="{selected.name} details">
          <div class="detail-header">
            <div class="detail-title-row">
              <div class="detail-art-wrap">
                <img class="detail-art" src={artFor(selected.output)} alt="" />
              </div>
              <div>
                <div class="detail-name">{selected.name}</div>
                <div class="recipe-meta" style="margin-top:0.25em">
                  {#if selected.category}<span class="chip">{formatCategory(selected.category)}</span>{/if}
                  <span class="chip" class:forge={!!selected.station}>{stationBadge(selected)}</span>
                </div>
              </div>
            </div>
            <button class="detail-close" type="button" on:click={() => (selectedKey = '')}>×</button>
          </div>
          {#if selected.description}
            <div class="detail-desc">{selected.description}</div>
          {/if}
          <div class="detail-section-label">Ingredients</div>
          <div class="ing-list">
            {#each selected.ingredients || [] as ing}
              <div class="ing-row">
                <img src={artFor(ing)} alt="" />
                <span class="ing-name">{ing.name}</span>
                <span class="ing-count" class:ing-ok={ing.haveEnough} class:ing-miss={!ing.haveEnough}>{ing.have}/{ing.qty}</span>
              </div>
            {/each}
          </div>
          <div class="station-box">
            {#if selected.station}
              Station: {stationBadge(selected)}
              {#if selected.stationHint}<div style="margin-top:0.25em;color:#94a3b8">{selected.stationHint}</div>{/if}
              {#if !selected._stationOK}<div style="margin-top:0.25em;color:#f87171">You are not at a suitable station.</div>{/if}
            {:else}
              Station: anywhere
            {/if}
          </div>
          <div class="detail-actions">
            <button
              class="detail-action-btn craft"
              type="button"
              disabled={!selected.canCraft}
              on:click={() => craftRecipe(selected)}
            >
              <i class="material-icons" style="font-size:1em">construction</i>
              Craft
            </button>
            <button class="detail-action-btn cancel" type="button" on:click={() => (selectedKey = '')}>Cancel</button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}
