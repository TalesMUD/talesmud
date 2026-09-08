<script>
  import { onDestroy, onMount, tick } from 'svelte';
  import { readStageSize, shouldRepaintSize, applyCanvasBitmap } from '../widgets/atlasLayout.js';
  import { paintAtlas, isCurrentPlace, panToCenterPlace, onMapTilesReady, clampMapScale, setYouPortrait } from '../widgets/atlasRenderer.js';

  export let store = null;
  export let sendMessage = null;

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
  let tooltip = { visible: false, text: '', x: 0, y: 0 };

  let stageWrap, stageCanvas, modalEl, mapBodyEl, intelEl;
  let selectedId = null;
  let lastTap = { id: null, at: 0 };
  let stageObserver;
  const hitState = { items: [] };
  let lastStageSize = null;
  let drawRaf = 0;
  let escHandler = null;
  let wasOpen = false;
  let resizeHandler = null;

  $: open = !!(store && $store && $store.mapOverviewOpen);

  function portalToBody(node) {
    // Inline every critical property — Materialize + scoped CSS must not win.
    node.style.cssText = [
      'position:fixed',
      'top:0',
      'left:0',
      'right:0',
      'bottom:0',
      'width:100vw',
      'height:100vh',
      'z-index:200000',
      'display:flex',
      'align-items:center',
      'justify-content:center',
      'padding:12px',
      'box-sizing:border-box',
      'background:rgba(0,0,0,0.82)',
      'opacity:1',
      'visibility:visible',
      'pointer-events:auto',
    ].join(';');
    if (node.parentNode !== document.body) {
      document.body.appendChild(node);
    }
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node);
      },
    };
  }

  function sizeModal() {
    if (!modalEl || typeof window === 'undefined') return { w: 0, h: 0 };
    const w = Math.max(640, Math.round(window.innerWidth * 0.8));
    const h = Math.max(460, Math.round(window.innerHeight * 0.8));
    modalEl.style.boxSizing = 'border-box';
    modalEl.style.position = 'relative';
    modalEl.style.left = 'auto';
    modalEl.style.right = 'auto';
    modalEl.style.top = 'auto';
    modalEl.style.bottom = 'auto';
    modalEl.style.opacity = '1';
    modalEl.style.visibility = 'visible';
    modalEl.style.display = 'flex';
    modalEl.style.flexDirection = 'column';
    modalEl.style.width = w + 'px';
    modalEl.style.height = h + 'px';
    modalEl.style.minWidth = w + 'px';
    modalEl.style.minHeight = h + 'px';
    modalEl.style.maxWidth = w + 'px';
    modalEl.style.maxHeight = h + 'px';
    modalEl.style.flex = 'none';
    modalEl.style.transform = 'none';
    const narrow = window.innerWidth < 900;
    if (mapBodyEl) {
      mapBodyEl.style.flexDirection = narrow ? 'column' : 'row';
    }
    if (intelEl) {
      intelEl.style.width = narrow ? '100%' : '300px';
      intelEl.style.flex = narrow ? '0 0 38%' : '0 0 300px';
      intelEl.style.maxHeight = narrow ? '38%' : 'none';
    }
    return { w, h };
  }

  function paintAfterLayout() {
    tick().then(() => {
      sizeModal();
      requestAnimationFrame(() => {
        sizeModal();
        requestAnimationFrame(() => {
          lastStageSize = readStageSize(stageWrap);
          applyRecenterToYou(true);
          scheduleDraw();
        });
      });
    });
  }

  function emptyAtlas() {
    return { characterId: '', currentRoomId: '', currentLayer: '', layers: [], places: [], paths: [], regions: [] };
  }

  function resolveLayer(data, roomId, preferred) {
    const places = data.places || [];
    if (preferred && places.some(p => p.layer === preferred)) return preferred;
    const here = places.find(p => p.id === roomId) || places.find(p => isCurrentPlace(p.id, roomId));
    if (here && here.layer) return here.layer;
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
    if (atlasChanged) atlas = nextAtlas;
    if (roomChanged) {
      currentRoomId = newRoomId;
      if (isTraveling && newRoomId) advanceTravel(newRoomId);
    }
    const nextLayer = resolveLayer(atlas, currentRoomId, $store.atlasLayer || activeLayer);
    const layerChanged = nextLayer !== activeLayer;
    if (layerChanged) activeLayer = nextLayer;
    // Follow / recenter when you move or atlas/layer catches up (meadow clusters).
    if (open && (roomChanged || layerChanged || (atlasChanged && currentRoomId))) {
      tick().then(() => {
        applyRecenterToYou(true);
        scheduleDraw();
      });
    } else if (open && atlasChanged) {
      scheduleDraw();
    }
  }

  $: if (open && !wasOpen) {
    wasOpen = true;
    userScale = 1;
    lastStageSize = null;
    paintAfterLayout();
  } else if (!open && wasOpen) {
    wasOpen = false;
    scheduleDraw();
  }

  $: if (store && $store.character) setYouPortrait($store.character.portrait || '');

  $: visiblePlaces = (atlas.places || []).filter(p => p.layer === activeLayer);
  $: visibleRegions = (atlas.regions || []).filter(r => r.layer === activeLayer);
  $: if (store && $store.mapSelectedId && $store.mapSelectedId !== selectedId) {
    selectedId = $store.mapSelectedId;
  }
  $: selectedPlace = (atlas.places || []).find(p => p.id === selectedId) || null;
  $: if (open && !selectedId && currentRoomId) {
    selectedId = currentRoomId;
    if (store && store.selectMapPlace) store.selectMapPlace(currentRoomId);
  }
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

  function requestTravel() {
    if (selectedPlace && selectedPlace.discovered && selectedPlace.id !== currentRoomId) {
      startTravel(selectedPlace.id);
    }
  }

  function dirBadge(d) {
    const k = String(d || '').toLowerCase();
    const m = { north: 'N', south: 'S', east: 'E', west: 'W', up: 'UP', down: 'DWN', northeast: 'NE', northwest: 'NW', southeast: 'SE', southwest: 'SW' };
    return m[k] || String(d || '?').slice(0, 3).toUpperCase();
  }

  function dangerLabel(d) {
    return ({ safe: 'Safe', low: 'Low', hazard: 'Hazard', hostile: 'Hostile', uncharted: 'Unknown' })[d] || d || '—';
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
      maximized: true,
      panX,
      panY,
      userScale,
      travelPathRoomIds,
      travelTargetId,
      selectedId,
    });
    hitState.items = result.hits;
  }

  function scheduleDraw() {
    if (drawRaf) return;
    const raf = typeof requestAnimationFrame === 'function' ? requestAnimationFrame : (fn) => setTimeout(fn, 16);
    drawRaf = raf(() => {
      drawRaf = 0;
      if (open) paint(stageCanvas, stageWrap);
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
      else if (found.discovered) text += ' · inspect';
      else text += ' · uncharted';
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
      if (found) {
        const now = Date.now();
        const dbl = lastTap.id === found.id && now - lastTap.at < 420;
        lastTap = { id: found.id, at: now };
        if (dbl && found.discovered && found.id !== currentRoomId) {
          startTravel(found.id);
        } else {
          selectedId = found.id;
          if (store && store.selectMapPlace) store.selectMapPlace(found.id);
        }
        scheduleDraw();
      }
    }
    isPanning = false;
    didDrag = false;
  }

  function onWheel(e) {
    e.preventDefault();
    userScale = clampMapScale(userScale * (e.deltaY < 0 ? 1.12 : 0.89));
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

  function applyRecenterToYou(keepScale = true) {
    if (!keepScale) userScale = 1;
    const size = readStageSize(stageWrap);
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

  function closeOverview() {
    if (store && store.closeMapOverview) store.closeMapOverview();
    else if (store && store.setMapOverviewOpen) store.setMapOverviewOpen(false);
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

  let observedStage = null;
  $: if (stageWrap !== observedStage) {
    if (stageObserver) stageObserver.disconnect();
    observedStage = stageWrap;
    lastStageSize = null;
    if (stageWrap) {
      stageObserver = new ResizeObserver(() => {
        onStageResize(stageWrap, lastStageSize, (s) => { lastStageSize = s; });
      });
      stageObserver.observe(stageWrap);
      lastStageSize = readStageSize(stageWrap);
      scheduleDraw();
    }
  }

  onMount(() => {
    onMapTilesReady(() => scheduleDraw());
    escHandler = (e) => {
      if (e.key === 'Escape' && open) {
        e.preventDefault();
        closeOverview();
      }
    };
    window.addEventListener('keydown', escHandler);
    resizeHandler = () => {
      if (open) paintAfterLayout();
    };
    window.addEventListener('resize', resizeHandler);
  });

  onDestroy(() => {
    if (escHandler) window.removeEventListener('keydown', escHandler);
    if (resizeHandler) window.removeEventListener('resize', resizeHandler);
    if (drawRaf && typeof cancelAnimationFrame === 'function') cancelAnimationFrame(drawRaf);
    if (stageObserver) stageObserver.disconnect();
    cancelTravel();
  });
</script>

<style>
  /* Scoped fallbacks — critical layout also inlined so body portal cannot lose them. */
  .map-overlay {
    position: fixed;
    inset: 0;
    z-index: 200000;
    background: rgba(0, 0, 0, 0.82);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1em;
    overflow: hidden;
    opacity: 1;
    visibility: visible;
  }
  /* Never class="modal" — Materialize global .modal is opacity:0 / display:none. */
  .map-panel {
    position: relative;
    width: 80vw;
    height: 80vh;
    max-width: none;
    max-height: none;
    display: flex;
    flex-direction: column;
    background: #0b0e14;
    border: 1px solid rgba(212, 175, 55, 0.28);
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.55);
    opacity: 1;
    visibility: visible;
    transform: none;
  }
  .map-body {
    flex: 1 1 0;
    min-height: 0;
    display: flex;
    flex-direction: row;
  }
  .intel {
    flex: 0 0 300px;
    width: 300px;
    min-width: 0;
    min-height: 0;
    overflow: auto;
    padding: 12px 14px 16px;
    background: #10141c;
    border-left: 1px solid rgba(212, 175, 55, 0.18);
    color: #d7d0c4;
    font-size: 12px;
  }
  .intel-kicker {
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #8a8070;
    margin-bottom: 4px;
  }
  .intel-title {
    margin: 0 0 8px;
    font-family: Georgia, serif;
    font-size: 1.15rem;
    color: #f3ead4;
    font-weight: 700;
    line-height: 1.25;
  }
  .intel-summary { margin: 0 0 10px; color: #b7ae9e; line-height: 1.45; }
  .intel-meta, .intel-tags, .intel-muted, .intel-hint, .intel-empty { color: #8a8070; font-size: 11px; }
  .intel-empty { padding: 1.5em 0; }
  .intel-chips { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 8px; }
  .chip {
    border: 1px solid rgba(148,163,184,0.28);
    border-radius: 999px;
    padding: 1px 7px;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: #c5bba8;
  }
  .chip.you { border-color: #e0b84a; color: #e0b84a; }
  .chip.danger-safe { border-color: #5ee0a0; color: #5ee0a0; }
  .chip.danger-low { border-color: #c4b07a; color: #c4b07a; }
  .chip.danger-hazard { border-color: #f0b44a; color: #f0b44a; }
  .chip.danger-hostile { border-color: #f07171; color: #f07171; }
  .chip.danger-uncharted { border-color: #6b7280; color: #9ca3af; }
  .intel-section {
    margin: 12px 0 6px;
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: #9a8f78;
    border-bottom: 1px solid rgba(148,163,184,0.12);
    padding-bottom: 3px;
  }
  .exit-row, .res-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 0;
    border-bottom: 1px solid rgba(255,255,255,0.04);
  }
  .exit-dir {
    flex: 0 0 32px;
    text-align: center;
    font-weight: 700;
    font-size: 10px;
    color: #f3ead4;
    background: rgba(255,255,255,0.05);
    border-radius: 3px;
    padding: 3px 0;
  }
  .exit-body { display: flex; flex-direction: column; min-width: 0; }
  .exit-body strong { color: #eee6d6; font-size: 12px; }
  .exit-body em { font-style: normal; color: #8a8070; font-size: 10px; }
  .res-dot { width: 7px; height: 7px; border-radius: 50%; background: #5b9fd6; flex-shrink: 0; }
  .res-dot.enemy { background: #e07a7a; }
  .res-kind { margin-left: auto; font-size: 10px; color: #8a8070; text-transform: uppercase; }
  .travel-btn {
    margin-top: 14px;
    width: 100%;
    border: 1px solid #d4af37;
    background: #c9a227;
    color: #1a1408;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 8px 10px;
    border-radius: 4px;
    cursor: pointer;
  }
  .travel-btn:disabled { opacity: 0.5; cursor: default; }
  .intel-hint { margin-top: 6px; }
  .toolbar {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0.75em 1em;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    background: rgba(20, 26, 36, 0.9);
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
  .stage {
    flex: 1 1 0;
    min-height: 0;
    position: relative;
    overflow: hidden;
  }
  canvas {
    position: absolute;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
    cursor: grab;
    touch-action: none;
  }
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
</style>

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div
    id="map-overview-overlay"
    class="map-overlay"
    style="position:fixed;top:0;left:0;width:100vw;height:100vh;z-index:200000;display:flex;align-items:center;justify-content:center;padding:12px;box-sizing:border-box;background:rgba(0,0,0,0.82);opacity:1;visibility:visible;"
    use:portalToBody
    role="dialog"
    aria-modal="true"
    aria-label="World map"
    on:click={(e) => { if (e.target === e.currentTarget) closeOverview(); }}
  >
    <div
      class="map-panel"
      bind:this={modalEl}
      style="position:relative;left:auto;right:auto;top:auto;bottom:auto;width:80vw;height:80vh;max-width:none;max-height:none;display:flex;flex-direction:column;overflow:hidden;background:#0b0e14;border:1px solid rgba(212,175,55,0.28);border-radius:8px;flex:none;opacity:1;visibility:visible;transform:none;"
      on:click|stopPropagation
    >
      <div class="toolbar">
        <i class="material-icons">explore</i>
        Cartographer
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
      <div class="map-body" bind:this={mapBodyEl}>
        <div class="stage" style="flex:1 1 0;min-height:0;min-width:0;position:relative;overflow:hidden;" bind:this={stageWrap}>
          <canvas
            style="position:absolute;inset:0;display:block;width:100%;height:100%;"
            bind:this={stageCanvas}
            on:pointerdown={pointerDown}
            on:pointermove={(e) => pointerMove(e, stageCanvas)}
            on:pointerup={(e) => pointerUp(e, stageCanvas)}
            on:pointerleave={() => tooltip = { ...tooltip, visible: false }}
            on:wheel={onWheel}
          ></canvas>
          {#if tooltip.visible}
            <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">{tooltip.text}</div>
          {/if}
        </div>
        <aside class="intel" bind:this={intelEl}>
          {#if !selectedPlace}
            <div class="intel-empty">Select a room on the chart.</div>
          {:else if !selectedPlace.discovered}
            <div class="intel-kicker">Uncharted</div>
            <h2 class="intel-title">Fog of war</h2>
            <p class="intel-summary">{selectedPlace.summary || 'Walk closer to chart this ground.'}</p>
            <div class="intel-meta">Z {selectedPlace.z} · {selectedPlace.layer}</div>
          {:else}
            <div class="intel-kicker">{selectedPlace.areaName || selectedPlace.area || 'Unknown sector'}</div>
            <h2 class="intel-title">{selectedPlace.name || selectedPlace.id}</h2>
            <div class="intel-chips">
              <span class="chip danger-{selectedPlace.danger || 'low'}">{dangerLabel(selectedPlace.danger)}</span>
              <span class="chip">Z:{selectedPlace.z} {selectedPlace.layer}</span>
              <span class="chip">{selectedPlace.biome || 'wild'}</span>
              <span class="chip">{selectedPlace.kind || 'place'}</span>
              {#if selectedPlace.current}<span class="chip you">You are here</span>{/if}
            </div>
            <p class="intel-summary">{selectedPlace.summary || ''}</p>
            {#if selectedPlace.tags && selectedPlace.tags.length}
              <div class="intel-tags">{selectedPlace.tags.join(' · ')}</div>
            {/if}
            <div class="intel-section">Vectors &amp; exits</div>
            {#if selectedPlace.exits && selectedPlace.exits.length}
              {#each selectedPlace.exits as ex}
                <div class="exit-row">
                  <span class="exit-dir">{dirBadge(ex.dir)}</span>
                  <span class="exit-body">
                    <strong>{ex.toName || (ex.to ? 'Uncharted' : '—')}</strong>
                    <em>{ex.hidden ? 'hidden' : ex.dir}{#if ex.vertical} · stair{/if}</em>
                  </span>
                </div>
              {/each}
            {:else}
              <div class="intel-muted">No charted exits.</div>
            {/if}
            <div class="intel-section">Usually here</div>
            {#if selectedPlace.residents && selectedPlace.residents.length}
              {#each selectedPlace.residents as r}
                <div class="res-row">
                  <span class="res-dot {r.kind}"></span>
                  {r.name}
                  <span class="res-kind">{r.kind}</span>
                </div>
              {/each}
            {:else}
              <div class="intel-muted">None recorded.</div>
            {/if}
            {#if selectedPlace.id !== currentRoomId && selectedPlace.discovered}
              <button class="travel-btn" type="button" on:click={requestTravel} disabled={isTraveling}>
                {isTraveling ? 'Traveling…' : 'Travel'}
              </button>
              <div class="intel-hint">Double-click a room to travel.</div>
            {/if}
          {/if}
        </aside>
      </div>
    </div>
  </div>
{/if}
