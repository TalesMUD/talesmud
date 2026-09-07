<script>
  import { onDestroy, onMount, tick } from 'svelte';
  import { readStageSize, shouldRepaintSize, applyCanvasBitmap } from './atlasLayout.js';
  import { paintAtlas, isCurrentPlace, panToCenterPlace } from './atlasRenderer.js';

  export let store = null;
  export let sendMessage = null;
  /** When true, hide the inline widget chrome; only render the full Map overview overlay. */
  export let overlayHost = false;

  let atlas = emptyAtlas();
  let currentRoomId = null;
  let activeLayer = '';

  let travelTargetId = null;
  let travelPath = [];
  let isTraveling = false;
  let travelPathRoomIds = new Set();

  let panX = 0;
  let panY = 0;
  let userScale = 1;
  let isPanning = false;
  let didDrag = false;
  let panStart = { x: 0, y: 0, panX: 0, panY: 0 };
  let maximized = false;
  let tooltip = { visible: false, text: '', x: 0, y: 0 };

  let widgetWrap, widgetCanvas;
  let modalWrap, modalCanvas;
  let widgetObserver, modalObserver;
  const hitState = { items: [] };
  let lastWidgetSize = null;
  let lastModalSize = null;
  let drawRaf = 0;
  let escHandler = null;

  // External Map pin / host opens overview via store.mapOverviewOpen
  $: if (store && $store) {
    const want = !!$store.mapOverviewOpen;
    if (want !== maximized) {
      maximized = want;
      if (maximized) {
        userScale = 1;
        lastModalSize = null;
        tick().then(() => {
          applyRecenterToYou(true);
          scheduleDraw();
        });
      } else {
        scheduleDraw();
      }
    }
  }

  function emptyAtlas() {
    return { characterId: '', currentRoomId: '', currentLayer: '', layers: [], places: [], paths: [], regions: [] };
  }

  /** Always mount fullscreen Map on document.body above inventory / HUD / BattleStage. */
  function portal(node) {
    if (node.parentNode !== document.body) {
      document.body.appendChild(node);
    }
    node.style.zIndex = '200000';
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node);
      }
    };
  }

  function resolveLayer(data, roomId, preferred) {
    const places = data.places || [];
    const here = places.find(p => p.id === roomId);
    if (here && here.layer) return here.layer;
    if (preferred && places.some(p => p.layer === preferred)) return preferred;
    if (data.currentLayer && places.some(p => p.layer === data.currentLayer)) return data.currentLayer;
    if (data.layers && data.layers[0]) return data.layers[0].id;
    const first = places[0];
    return (first && first.layer) || 'overworld';
  }

  $: if (store) {
    const nextAtlas = $store.atlas && Array.isArray($store.atlas.places) ? $store.atlas : emptyAtlas();
    const newRoomId = $store.currentRoomId || nextAtlas.currentRoomId || null;
    const roomChanged = newRoomId !== currentRoomId;
    const atlasChanged = nextAtlas !== atlas;
    if (atlasChanged) {
      atlas = nextAtlas;
    }
    if (roomChanged) {
      currentRoomId = newRoomId;
      userScale = 1;
      if (isTraveling && newRoomId) advanceTravel(newRoomId);
    }
    const nextLayer = resolveLayer(atlas, currentRoomId, $store.atlasLayer || activeLayer);
    const layerChanged = nextLayer !== activeLayer;
    if (layerChanged) {
      activeLayer = nextLayer;
    }
    // Recenter when you move or atlas/layer catches up (separateAreas clusters).
    if (roomChanged || layerChanged || (atlasChanged && currentRoomId)) {
      tick().then(() => {
        applyRecenterToYou(true);
        scheduleDraw();
      });
    } else if (atlasChanged) {
      scheduleDraw();
    }
  }

  $: visiblePlaces = (atlas.places || []).filter(p => p.layer === activeLayer);
  $: visibleRegions = (atlas.regions || []).filter(r => r.layer === activeLayer);
  $: layers = atlas.layers || [];

  function placeById(id) {
    return (atlas.places || []).find(p => p.id === id);
  }

  function findPath(startId, targetId) {
    if (!startId || !targetId || startId === targetId) return null;
    const byId = {};
    for (const p of atlas.places || []) byId[p.id] = p;
    if (!byId[startId] || !byId[targetId] || !byId[startId].discovered || !byId[targetId].discovered) {
      return null;
    }
    const adj = {};
    for (const path of atlas.paths || []) {
      if (!byId[path.from] || !byId[path.to] || !byId[path.from].discovered) continue;
      if (!adj[path.from]) adj[path.from] = [];
      adj[path.from].push({ to: path.to, dir: path.dir });
    }
    const queue = [startId];
    const seen = new Set([startId]);
    const parent = {};
    while (queue.length) {
      const id = queue.shift();
      for (const edge of adj[id] || []) {
        if (seen.has(edge.to)) continue;
        if (!byId[edge.to] || (!byId[edge.to].discovered && edge.to !== targetId)) continue;
        seen.add(edge.to);
        parent[edge.to] = { parentId: id, direction: edge.dir };
        queue.push(edge.to);
        if (edge.to === targetId) {
          const steps = [];
          let cur = targetId;
          while (parent[cur]) {
            steps.unshift({ roomId: cur, direction: parent[cur].direction });
            cur = parent[cur].parentId;
          }
          return steps;
        }
      }
    }
    return null;
  }

  function startTravel(targetId) {
    if (!currentRoomId || targetId === currentRoomId) return;
    const path = findPath(currentRoomId, targetId);
    if (!path || !path.length) return;
    travelTargetId = targetId;
    travelPath = path;
    isTraveling = true;
    travelPathRoomIds = new Set(path.map(s => s.roomId));
    if (sendMessage) sendMessage(path[0].direction);
  }

  function cancelTravel() {
    travelTargetId = null;
    travelPath = [];
    isTraveling = false;
    travelPathRoomIds = new Set();
  }

  function advanceTravel(newRoomId) {
    if (!isTraveling || !travelPath.length) {
      cancelTravel();
      return;
    }
    if (travelPath[0].roomId === newRoomId) {
      travelPath = travelPath.slice(1);
      travelPathRoomIds = new Set(travelPath.map(s => s.roomId));
      if (!travelPath.length) cancelTravel();
      else setTimeout(() => { if (sendMessage && travelPath[0]) sendMessage(travelPath[0].direction); }, 160);
    } else {
      cancelTravel();
    }
  }

  function paint(canvas, wrap) {
    if (!canvas || !wrap) return;
    const size = readStageSize(wrap);
    if (size.w < 4 || size.h < 4) return;
    const dpr = window.devicePixelRatio || 1;
    applyCanvasBitmap(canvas, size.w, size.h, dpr);
    const w = size.w;
    const h = size.h;

    const ctx = canvas.getContext('2d');
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    const result = paintAtlas(ctx, {
      w,
      h,
      atlas,
      activeLayer,
      visiblePlaces,
      visibleRegions,
      currentRoomId,
      maximized,
      panX,
      panY,
      userScale,
      travelPathRoomIds,
      travelTargetId,
    });
    hitState.items = result.hits;
  }

  function scheduleDraw() {
    if (drawRaf) return;
    const raf = typeof requestAnimationFrame === 'function' ? requestAnimationFrame : (fn) => setTimeout(fn, 16);
    drawRaf = raf(() => {
      drawRaf = 0;
      paint(widgetCanvas, widgetWrap);
      if (maximized) paint(modalCanvas, modalWrap);
    });
  }

  function hitTest(canvas, mx, my) {
    if (!canvas) return null;
    const items = hitState.items;
    for (let i = items.length - 1; i >= 0; i--) {
      const h = items[i];
      const dx = mx - h.px;
      const dy = my - h.py;
      if (dx * dx + dy * dy <= h.r * h.r) return h.place;
    }
    return null;
  }

  function pointerDown(e) {
    isPanning = true;
    didDrag = false;
    panStart = { x: e.clientX, y: e.clientY, panX, panY };
    e.currentTarget.setPointerCapture(e.pointerId);
  }

  function pointerMove(e, canvas) {
    if (isPanning) {
      const dx = e.clientX - panStart.x;
      const dy = e.clientY - panStart.y;
      if (Math.abs(dx) > 3 || Math.abs(dy) > 3) didDrag = true;
      panX = panStart.panX + dx;
      panY = panStart.panY + dy;
      scheduleDraw();
      return;
    }
    const rect = canvas.getBoundingClientRect();
    const found = hitTest(canvas, e.clientX - rect.left, e.clientY - rect.top);
    if (found) {
      let text = found.discovered ? (found.name || found.id) : 'Uncharted';
      if (found.areaName && found.discovered) text += ' · ' + found.areaName;
      if (found.current || isCurrentPlace(found.id, currentRoomId)) text += ' (you are here)';
      else if (found.discovered) text += ' (click to travel)';
      tooltip = { visible: true, text, x: e.clientX - rect.left, y: e.clientY - rect.top };
    } else {
      tooltip = { ...tooltip, visible: false };
    }
  }

  function pointerUp(e, canvas) {
    e.currentTarget.releasePointerCapture(e.pointerId);
    if (isPanning && !didDrag) {
      const rect = canvas.getBoundingClientRect();
      const found = hitTest(canvas, e.clientX - rect.left, e.clientY - rect.top);
      if (found && found.discovered && found.id !== currentRoomId) {
        if (found.id === travelTargetId) cancelTravel();
        else {
          cancelTravel();
          startTravel(found.id);
        }
        scheduleDraw();
      }
    }
    isPanning = false;
    didDrag = false;
  }

  function onWheel(e) {
    e.preventDefault();
    userScale = Math.min(2.6, Math.max(0.55, userScale * (e.deltaY < 0 ? 1.12 : 0.89)));
    scheduleDraw();
  }

  function resolveHerePlace() {
    const places = visiblePlaces || [];
    return (
      places.find((p) => p.id === currentRoomId) ||
      places.find((p) => isCurrentPlace(p.id, currentRoomId)) ||
      null
    );
  }

  /** Pan so you-are-here is centered. keepScale=false also resets zoom. */
  function applyRecenterToYou(keepScale = true) {
    if (!keepScale) userScale = 1;
    // Prefer modal stage when overview is open (overlayHost owns it).
    const wrap = (maximized && modalWrap) ? modalWrap : widgetWrap;
    const size = readStageSize(wrap);
    const here = resolveHerePlace();
    if (here && size.w >= 4 && size.h >= 4) {
      const pan = panToCenterPlace(visiblePlaces, here, size.w, size.h, userScale, atlas.paths || []);
      panX = pan.panX;
      panY = pan.panY;
    } else {
      panX = 0;
      panY = 0;
    }
  }

  function recenter() {
    applyRecenterToYou(false);
    scheduleDraw();
  }

  async function toggleMaximize() {
    const next = !maximized;
    // Always route through store so Game.svelte overlayHost owns the fullscreen modal.
    if (store && store.setMapOverviewOpen) {
      store.setMapOverviewOpen(next);
      return;
    }
    maximized = next;
    if (maximized) {
      userScale = 1;
      await tick();
      applyRecenterToYou(true);
    }
    lastModalSize = null;
    scheduleDraw();
  }

  function closeOverview() {
    if (store && store.closeMapOverview) store.closeMapOverview();
    else if (store && store.setMapOverviewOpen) store.setMapOverviewOpen(false);
    else maximized = false;
  }

  function selectLayer(id) {
    if (id === activeLayer) return;
    activeLayer = id;
    if (store && store.setAtlasLayer) store.setAtlasLayer(id);
    userScale = 1;
    tick().then(() => {
      applyRecenterToYou(true);
      scheduleDraw();
    });
  }

  function onStageResize(wrap, lastRef, setLast) {
    const next = readStageSize(wrap);
    if (!shouldRepaintSize(lastRef, next)) return;
    setLast(next);
    scheduleDraw();
  }

  let observedWidget = null;
  let observedModal = null;
  $: if (widgetWrap !== observedWidget) {
    if (widgetObserver) widgetObserver.disconnect();
    observedWidget = widgetWrap;
    lastWidgetSize = null;
    if (widgetWrap) {
      widgetObserver = new ResizeObserver(() => {
        onStageResize(widgetWrap, lastWidgetSize, (s) => { lastWidgetSize = s; });
      });
      widgetObserver.observe(widgetWrap);
      lastWidgetSize = readStageSize(widgetWrap);
      scheduleDraw();
    }
  }
  $: if (modalWrap !== observedModal) {
    if (modalObserver) modalObserver.disconnect();
    observedModal = modalWrap;
    lastModalSize = null;
    if (modalWrap) {
      modalObserver = new ResizeObserver(() => {
        onStageResize(modalWrap, lastModalSize, (s) => { lastModalSize = s; });
      });
      modalObserver.observe(modalWrap);
      lastModalSize = readStageSize(modalWrap);
      scheduleDraw();
    }
  }

  onMount(() => {
    escHandler = (e) => {
      if (e.key === 'Escape' && maximized && overlayHost) {
        e.preventDefault();
        closeOverview();
      }
    };
    window.addEventListener('keydown', escHandler);
  });

  onDestroy(() => {
    if (escHandler) window.removeEventListener('keydown', escHandler);
    if (drawRaf && typeof cancelAnimationFrame === 'function') cancelAnimationFrame(drawRaf);
    if (widgetObserver) widgetObserver.disconnect();
    if (modalObserver) modalObserver.disconnect();
    cancelTravel();
    // Layout remounts / tab switches must NOT close the fullscreen overlay.
    // Only the Game.svelte overlayHost owns close-on-destroy (and even that is
    // optional — prefer Esc/X). Never auto-close from the compact widget.
    if (overlayHost) {
      maximized = false;
    }
  });
</script>

<style>
  .atlas {
    width: 100%;
    height: 100%;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: var(--panel-bg, #0d1117);
    border-radius: var(--panel-radius, 8px);
    border: 1px solid var(--panel-border, rgba(255, 255, 255, 0.1));
    overflow: hidden;
  }
  .toolbar {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    color: #cbd5e1;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .toolbar i { font-size: 16px; color: #94a3b8; }
  .spacer { flex: 1; }
  .icon-btn {
    cursor: pointer;
    background: none;
    border: none;
    color: #94a3b8;
    padding: 2px;
    border-radius: 4px;
    display: flex;
    align-items: center;
  }
  .icon-btn:hover { color: #e2e8f0; background: rgba(255,255,255,0.08); }
  .icon-btn.cancel { color: #f87171; }
  .layer-tabs { display: flex; gap: 4px; }
  .layer-tab {
    border: 1px solid rgba(148, 163, 184, 0.35);
    background: transparent;
    color: #94a3b8;
    font-size: 10px;
    text-transform: none;
    letter-spacing: 0;
    padding: 2px 7px;
    border-radius: 999px;
    cursor: pointer;
  }
  .layer-tab.active { background: #f59e0b; border-color: #f59e0b; color: #111827; }
  .travel { font-size: 10px; color: #22d3ee; text-transform: none; letter-spacing: 0; }
  .stage { flex: 1 1 0; min-height: 0; position: relative; overflow: hidden; }
  canvas { position: absolute; inset: 0; display: block; width: 100%; height: 100%; cursor: grab; touch-action: none; }
  canvas:active { cursor: grabbing; }
  .tooltip {
    position: absolute;
    background: rgba(2, 6, 23, 0.92);
    color: #e2e8f0;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    pointer-events: none;
    white-space: nowrap;
    transform: translate(-50%, -110%);
    border: 1px solid rgba(148,163,184,0.25);
  }
  .backdrop {
    position: fixed;
    inset: 0;
    /* Above BattleStage (9000) and any HUD / inventory popup */
    z-index: 200000;
    background: rgba(0, 0, 0, 0.72);
    display: flex;
    padding: 3vh 3vw;
  }
  .modal {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: #0b1220;
    border: 1px solid rgba(255,255,255,0.12);
    border-radius: 12px;
    overflow: hidden;
  }
  .open-hint {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 12px;
  }
  .open-map-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: rgba(245, 158, 11, 0.15);
    border: 1px solid rgba(245, 158, 11, 0.45);
    color: #fbbf24;
    border-radius: 8px;
    padding: 10px 14px;
    cursor: pointer;
    font-size: 12px;
    font-weight: 600;
  }
  .open-map-btn:hover { background: rgba(245, 158, 11, 0.25); }
  .open-map-btn i { font-size: 18px; }
  .atlas-compact {
    min-height: 0;
  }
  .stage-compact {
    min-height: 96px;
  }
  .compact-open {
    position: absolute;
    right: 8px;
    bottom: 8px;
    z-index: 2;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    background: rgba(2, 6, 23, 0.82);
    border: 1px solid rgba(245, 158, 11, 0.45);
    color: #fbbf24;
    border-radius: 6px;
    padding: 4px 8px;
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
  }
  .compact-open:hover { background: rgba(245, 158, 11, 0.2); }
  .compact-open i { font-size: 14px; }
  .compact-open-live {
    border-color: rgba(34, 211, 238, 0.55);
    color: #67e8f9;
  }
</style>

{#if !overlayHost}
<div class="atlas atlas-compact">
  <div class="toolbar">
    <i class="material-icons">map</i>
    Map
    {#if isTraveling}<span class="travel">Traveling…</span>{/if}
    {#if maximized}<span class="travel">Fullscreen</span>{/if}
    <span class="spacer"></span>
    <button class="icon-btn" title={maximized ? 'Close Map' : 'Open Map'} on:click={toggleMaximize}>
      <i class="material-icons">{maximized ? 'close_fullscreen' : 'open_in_full'}</i>
    </button>
  </div>
  <!-- Always keep the mini canvas; fullscreen lives on Game overlayHost only. -->
  <div class="stage stage-compact" bind:this={widgetWrap}>
    <canvas
      bind:this={widgetCanvas}
      on:pointerdown={pointerDown}
      on:pointermove={(e) => pointerMove(e, widgetCanvas)}
      on:pointerup={(e) => pointerUp(e, widgetCanvas)}
      on:pointerleave={() => tooltip = { ...tooltip, visible: false }}
      on:wheel={onWheel}
      on:dblclick={toggleMaximize}
    ></canvas>
    {#if maximized}
      <button class="compact-open compact-open-live" type="button" title="Map is open — click to close" on:click={toggleMaximize}>
        <i class="material-icons">map</i>
        Map open
      </button>
    {:else}
      <button class="compact-open" type="button" title="Open fullscreen Map" on:click={toggleMaximize}>
        <i class="material-icons">open_in_full</i>
        Open Map
      </button>
    {/if}
    {#if tooltip.visible}
      <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">{tooltip.text}</div>
    {/if}
  </div>
</div>
{/if}

{#if maximized && overlayHost}
  <div
    class="backdrop"
    use:portal
    role="dialog"
    aria-modal="true"
    aria-label="Map"
    on:click={(e) => { if (e.target === e.currentTarget) closeOverview(); }}
  >
    <div class="modal">
      <div class="toolbar">
        <i class="material-icons">map</i>
        Map
        {#if isTraveling}<span class="travel">Traveling…</span>{/if}
        {#if layers.length > 1}
          <div class="layer-tabs">
            {#each layers as layer}
              <button class="layer-tab" class:active={activeLayer === layer.id} on:click={() => selectLayer(layer.id)}>{layer.name}</button>
            {/each}
          </div>
        {/if}
        <span class="spacer"></span>
        <button class="icon-btn" title="Recenter on you" on:click={recenter}>
          <i class="material-icons">my_location</i>
        </button>
        {#if isTraveling}
          <button class="icon-btn cancel" title="Cancel travel" on:click={cancelTravel}>
            <i class="material-icons">close</i>
          </button>
        {/if}
        <button class="icon-btn" title="Close (Esc)" on:click={closeOverview}>
          <i class="material-icons">close</i>
        </button>
      </div>
      <div class="stage" bind:this={modalWrap}>
        <canvas
          bind:this={modalCanvas}
          on:pointerdown={pointerDown}
          on:pointermove={(e) => pointerMove(e, modalCanvas)}
          on:pointerup={(e) => pointerUp(e, modalCanvas)}
          on:pointerleave={() => tooltip = { ...tooltip, visible: false }}
          on:wheel={onWheel}
        ></canvas>
        {#if tooltip.visible}
          <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">{tooltip.text}</div>
        {/if}
      </div>
    </div>
  </div>
{/if}
