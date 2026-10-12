<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { itemArtSrc, onItemArtError } from '../itemArtSrc.js';
  import { itemEffectLines } from '../itemEffects.js';
  import { portal } from '../portal.js';
  import { comparisonItems, comparisonRows, isTwoHanded, itemUsabilityReason, itemWeight, itemIsUsable, itemOffersUseOn } from './itemComparison.js';

  export let item;
  export let source = 'inventory';
  export let equippedItems = {};
  export let character = null;
  export let equipped = false;
  export let sellable = false;

  const dispatch = createEventDispatcher();
  let card;
  $: worn = source === 'inventory' ? comparisonItems(item, equippedItems, character) : [];
  $: rows = source === 'inventory' && item?.slot ? comparisonRows(item, equippedItems, character) : [];
  $: reason = itemUsabilityReason(item, character);
  $: weight = itemWeight(item);
  $: equippable = !!item?.slot && !['inventory', 'container', 'purse'].includes(item.slot);
  $: consumable = item?.type === 'consumable' || item?.consumable;
  $: usable = itemIsUsable(item);
  $: offersUseOn = itemOffersUseOn(item);
  $: effects = itemEffectLines(item);

  onMount(() => card?.focus());

  function close() { dispatch('close'); }
  function action(verb) { dispatch('action', { verb, item }); }
  function keydown(event) {
    if (event.key !== 'Escape') return;
    event.preventDefault();
    event.stopImmediatePropagation();
    close();
  }
  function color(quality) {
    return { magic: '#4ade80', rare: '#60a5fa', legendary: '#c084fc', mythic: '#fbbf24' }[quality] || '#e5e7eb';
  }
  function label(value) { return String(value || '').replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()); }
  function format(value) { return Number.isInteger(value) ? String(value) : Number(value).toFixed(1); }
</script>

<svelte:window on:keydown={keydown} />

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="item-card-backdrop" use:portal on:click={(event) => { if (event.target === event.currentTarget) close(); }}>
  <section class="item-card" class:item-card-unique={item.unique} role="dialog" aria-modal="true" aria-label="{item.name} details" tabindex="-1" bind:this={card}>
    <header class="item-card-header">
      <img class="item-card-art" src={itemArtSrc(item)} alt="" on:error={(event) => onItemArtError(event, item)} />
      <div class="item-card-heading">
        <h2 style="color: {item.unique ? '#fde68a' : color(item.quality)}">{item.name}{#if item.unique}<span class="item-card-unique-mark">UNIQUE</span>{/if}</h2>
        <div class="item-card-meta">{[item.quality || 'normal', item.type, item.subType].filter(Boolean).map(label).join(' · ')}</div>
        {#if equippable}<div class="item-card-slot">{label(item.slot)}{#if isTwoHanded(item)} · Two handed{/if}</div>{/if}
      </div>
      <button class="item-card-close" type="button" aria-label="Close item details" on:click={close}>×</button>
    </header>

    {#if item.description}<p class="item-card-description">{item.description}</p>{/if}
    {#if offersUseOn}<p class="item-card-use-tip">Tip: use this on another item — tap <strong>Use on…</strong>, or type <code>use {item.name.toLowerCase().split(/\s+/)[0]} on &lt;item&gt;</code>.</p>{/if}

    {#if source === 'inventory' && equippable}
      <div class="item-card-compare-title">Compared with {worn.length ? 'equipped' : 'empty slot'}</div>
      {#if worn.length}
        <div class="item-card-worn">
          {#each worn as old (old.id || old.name)}
            <span><img src={itemArtSrc(old)} alt="" on:error={(event) => onItemArtError(event, old)} />{old.name}</span>
          {/each}
        </div>
      {/if}
    {/if}

    {#if source === 'inventory' && equippable && rows.length}
      <div class="item-card-stats">
        {#each rows as row (row.key)}
          <div class="item-card-stat">
            <span>{row.label}</span>
            <span class="worn-value">{format(row.worn)}</span>
            <strong>→ {format(row.value)}</strong>
            <span class="diff" class:better={row.lowerIsBetter ? row.delta < 0 : row.delta > 0} class:worse={row.lowerIsBetter ? row.delta > 0 : row.delta < 0}>
              {row.delta > 0 ? '+' : ''}{format(row.delta)}
            </span>
          </div>
        {/each}
      </div>
    {:else if item.attributes && Object.keys(item.attributes).length}
      <div class="item-card-stats">
        {#each Object.entries(item.attributes) as [key, value]}
          <div class="item-card-stat"><span>{label(key)}</span><strong>{value}</strong></div>
        {/each}
      </div>
    {/if}

    {#if effects.length}
      <ul class="item-card-effects" aria-label="Effects">
        {#each effects as eff, i (i)}
          <li><span class="eff-mark" aria-hidden="true">✦</span><span><strong>{eff.label}</strong>{#if eff.name}{' — '}<em>{eff.name}</em>{/if}{#if eff.text}{eff.name ? ': ' : ' — '}{eff.text}{/if}</span></li>
        {/each}
      </ul>
    {/if}

    <div class="item-card-facts">
      <span>Weight <strong>{weight === null ? '—' : format(weight)}</strong></span>
      <span>Value <strong>{item.basePrice == null ? '—' : `${item.basePrice} gold`}</strong></span>
      {#if item.level}<span>Level <strong>{item.level}</strong></span>{/if}
    </div>
    {#if reason && source === 'inventory'}<p class="item-card-restriction">{reason}</p>{/if}

    <footer class="item-card-actions">
      {#if source === 'equipment' || equipped}
        <button type="button" class="primary" on:click={() => action('unequip')}>Unequip</button>
      {:else}
        <button type="button" on:click={() => action('examine')}>Examine</button>
        {#if equippable}<button type="button" class="primary" disabled={!!reason} on:click={() => action('equip')}>Equip</button>{/if}
        {#if usable}<button type="button" on:click={() => action('use')}>Use</button>{/if}
        {#if offersUseOn}<button type="button" on:click={() => action('useon')}>Use on…</button>{/if}
        {#if sellable}<button type="button" on:click={() => action('sell')}>Sell</button>{/if}
        <button type="button" on:click={() => action('drop')}>Drop</button>
      {/if}
      <button type="button" on:click={close}>Close</button>
    </footer>
  </section>
</div>

<style>
  .item-card-backdrop{position:fixed;inset:0;z-index:2000;background:rgba(0,0,0,.72);display:flex;align-items:center;justify-content:center;padding:16px}
  .item-card{width:min(660px,100%);max-height:min(86dvh,760px);overflow:auto;background:#11161c;border:1px solid rgba(212,175,55,.5);border-radius:12px;box-shadow:0 24px 70px #000c;color:#e5e7eb;padding:20px;outline:none}
  .item-card.item-card-unique{border:2px solid #facc15;box-shadow:0 0 0 1px #7a5a16,0 24px 70px #000c,0 0 22px rgba(250,204,21,.35)}
  .item-card-unique-mark{display:inline-block;margin-left:8px;font-size:.62rem;letter-spacing:.14em;color:#fde68a;border:1px solid rgba(250,204,21,.75);border-radius:999px;padding:1px 6px;vertical-align:middle}
  .item-card-effects{list-style:none;margin:2px 0 14px;padding:9px 12px;border-left:3px solid #e8c25a;border-radius:4px;background:linear-gradient(90deg,rgba(232,194,90,.13),rgba(96,165,250,.06));font-size:.84rem;line-height:1.45;color:#e6dcc4}
  .item-card-effects li{display:flex;gap:8px;align-items:baseline}
  .item-card-effects li+li{margin-top:4px}
  .item-card-effects .eff-mark{color:#f3d27a;flex:none}
  .item-card-effects strong{color:#f6d77e;font-weight:700}
  .item-card-effects em{font-style:normal;color:#9cc3f5;font-weight:600}
  .item-card-header{display:flex;gap:16px;align-items:center}
  .item-card-art{width:84px;height:84px;object-fit:contain;image-rendering:pixelated;background:#080b0f;border:1px solid #3d3423;border-radius:8px;flex:none}
  .item-card-heading{min-width:0;flex:1}
  h2{font-family:var(--font-display,serif);font-size:1.2rem;line-height:1.3;margin:0 0 4px}
  .item-card-meta,.item-card-slot{font-size:.78rem;color:#a9a397}
  .item-card-close{font-size:1.5rem;line-height:1;padding:2px 8px;align-self:flex-start}
  .item-card-description{font-size:.9rem;line-height:1.45;color:#d1c8b8;margin:16px 0}
  .item-card-use-tip{font-size:.78rem;color:#cfb573;margin:0 0 12px;line-height:1.4}
  .item-card-use-tip code{font-size:.74rem;background:#1b2026;padding:1px 5px;border-radius:4px;color:#ffe3a4}
  .item-card-compare-title{font-size:.74rem;text-transform:uppercase;letter-spacing:.1em;color:#cfb573;margin:14px 0 5px}
  .item-card-worn{display:flex;gap:8px;flex-wrap:wrap;font-size:.82rem;color:#bbb4a6}
  .item-card-worn span{display:inline-flex;align-items:center;gap:5px;background:#24201b;border:1px solid #4a3b25;border-radius:5px;padding:3px 7px}
  .item-card-worn img{width:26px;height:26px;object-fit:contain;image-rendering:pixelated}
  .item-card-stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(126px,1fr));gap:6px;margin:14px 0}
  .item-card-stat{display:flex;align-items:center;gap:6px;background:#1b2026;border:1px solid #30363b;border-radius:6px;padding:7px 9px;font-size:.78rem}
  .item-card-stat span:first-child{color:#b9bec5;flex:1}
  .item-card-stat .worn-value{color:#87929c;min-width:20px;text-align:right}
  .item-card-stat strong{color:#f0e5d1}
  .diff{color:#9ca3af;min-width:34px;text-align:right;font-weight:700}.diff.better{color:#4ade80}.diff.worse{color:#fb7185}
  .item-card-facts{display:flex;gap:16px;flex-wrap:wrap;border-top:1px solid #343a40;padding-top:12px;font-size:.78rem;color:#aeb3b9}
  .item-card-facts strong{color:#f3e4bd;margin-left:4px}
  .item-card-restriction{color:#fca5a5;font-size:.82rem;margin:12px 0 0}
  .item-card-actions{display:flex;flex-wrap:wrap;gap:8px;justify-content:flex-end;margin-top:18px}
  .item-card-actions button{margin:0;font-size:.82rem;padding:8px 14px;background:#252b31;color:#e7e2d8;border:1px solid #4a5157}
  .item-card-actions button.primary{background:#3d3018;border-color:#b38c3f;color:#ffe3a4}
  .item-card-actions button:disabled{opacity:.4;cursor:not-allowed}
  @media(max-width:600px){.item-card{padding:14px}.item-card-art{width:64px;height:64px}.item-card-header{gap:10px}}
</style>
