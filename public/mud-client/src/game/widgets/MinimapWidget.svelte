<script>
  import { onDestroy } from 'svelte';

  export let store = null;
  export let sendMessage = null;

  let canvas;
  let container;
  let resizeObserver;
  let tooltip = { visible: false, text: '', x: 0, y: 0 };

  const CELL = 56;
  const BIOME = {
    meadow: { fill: 'rgba(176, 196, 132, 0.42)', ink: '#6f8448' },
    forest: { fill: 'rgba(96, 130, 86, 0.40)', ink: '#3f5c38' },
    water: { fill: 'rgba(110, 158, 176, 0.38)', ink: '#3d6a7a' },
    dungeon: { fill: 'rgba(138, 118, 98, 0.38)', ink: '#5c4a3a' },
    settlement: { fill: 'rgba(196, 168, 118, 0.42)', ink: '#7a5c32' },
    wild: { fill: 'rgba(176, 168, 132, 0.36)', ink: '#5e5844' },
  };

  let atlas = emptyAtlas();
  let currentRoomId = null;
  let activeLayer = 'overworld';

  let travelTargetId = null;
  let travelPath = [];
  let isTraveling = false;
  let travelPathRoomIds = new Set();

  let panOffsetX = 0;
  let panOffsetY = 0;
  let isPanning = false;
  let panStartX = 0;
  let panStartY = 0;
  let panStartOffsetX = 0;
  let panStartOffsetY = 0;
  let didDrag = false;
  let scale = 1;
  let maximized = false;

  let hitPlaces = [];

  function emptyAtlas() {
    return { characterId: '', currentRoomId: '', currentLayer: 'overworld', layers: [], places: [], paths: [], regions: [] };
  }

  function portal(node) {
    document.body.appendChild(node);
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node);
      }
    };
  }

  $: if (store) {
    atlas = $store.atlas && Array.isArray($store.atlas.places) ? $store.atlas : emptyAtlas();
    const newRoomId = $store.currentRoomId || atlas.currentRoomId || null;
    const storeLayer = $store.atlasLayer || atlas.currentLayer || 'overworld';
    if (newRoomId !== currentRoomId) {
      currentRoomId = newRoomId;
      panOffsetX = 0;
      panOffsetY = 0;
      const here = (atlas.places || []).find(p => p.id === newRoomId);
      if (here && here.layer) {
        activeLayer = here.layer;
        if (store.setAtlasLayer) store.setAtlasLayer(here.layer);
      }
      if (isTraveling && newRoomId) {
        advanceTravel(newRoomId);
      }
    } else {
      activeLayer = storeLayer;
    }
  }

  $: if (canvas && atlas) {
    draw();
  }

  $: visiblePlaces = (atlas.places || []).filter(p => p.layer === activeLayer);
  $: visibleRegions = (atlas.regions || []).filter(r => r.layer === activeLayer);
  $: visiblePaths = (atlas.paths || []).filter(p => {
    const from = placeById(p.from);
    const to = placeById(p.to);
    return from && to && (from.layer === activeLayer || to.layer === activeLayer);
  });

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
      if (!byId[path.from] || !byId[path.to]) continue;
      if (!byId[path.from].discovered) continue;
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
          const path = [];
          let cur = targetId;
          while (parent[cur]) {
            path.unshift({ roomId: cur, direction: parent[cur].direction });
            cur = parent[cur].parentId;
          }
          return path;
        }
      }
    }
    return null;
  }

  function startTravel(targetId) {
    if (!currentRoomId || targetId === currentRoomId) return;
    const path = findPath(currentRoomId, targetId);
    if (!path || path.length === 0) return;
    travelTargetId = targetId;
    travelPath = path;
    isTraveling = true;
    travelPathRoomIds = new Set(path.map(s => s.roomId));
    sendNextStep();
  }

  function cancelTravel() {
    travelTargetId = null;
    travelPath = [];
    isTraveling = false;
    travelPathRoomIds = new Set();
  }

  function sendNextStep() {
    if (!isTraveling || travelPath.length === 0) {
      cancelTravel();
      return;
    }
    if (sendMessage) sendMessage(travelPath[0].direction);
  }

  function advanceTravel(newRoomId) {
    if (!isTraveling || travelPath.length === 0) {
      cancelTravel();
      return;
    }
    if (travelPath[0].roomId === newRoomId) {
      travelPath = travelPath.slice(1);
      travelPathRoomIds = new Set(travelPath.map(s => s.roomId));
      if (travelPath.length === 0) cancelTravel();
      else setTimeout(() => sendNextStep(), 160);
    } else {
      cancelTravel();
    }
  }

  function hash32(s) {
    let h = 2166136261;
    for (let i = 0; i < s.length; i++) h = Math.imul(h ^ s.charCodeAt(i), 16777619);
    return h >>> 0;
  }

  function handlePointerDown(e) {
    if (!canvas) return;
    isPanning = true;
    didDrag = false;
    panStartX = e.clientX;
    panStartY = e.clientY;
    panStartOffsetX = panOffsetX;
    panStartOffsetY = panOffsetY;
    canvas.setPointerCapture(e.pointerId);
  }

  function handlePointerMove(e) {
    if (!canvas) return;
    if (isPanning) {
      const dx = e.clientX - panStartX;
      const dy = e.clientY - panStartY;
      if (Math.abs(dx) > 3 || Math.abs(dy) > 3) didDrag = true;
      panOffsetX = panStartOffsetX + dx;
      panOffsetY = panStartOffsetY + dy;
      draw();
      return;
    }
    const rect = canvas.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const found = hitTest(mx, my);
    if (found) {
      let text = found.discovered ? (found.name || found.id) : 'Uncharted';
      if (found.areaName && found.discovered) text += ' · ' + found.areaName;
      if (found.current) text += ' (you are here)';
      else if (found.id === travelTargetId && isTraveling) text += ' (traveling…)';
      else if (found.discovered && found.canTravel) text += ' (click to travel)';
      tooltip = { visible: true, text, x: mx, y: my };
    } else {
      tooltip = { ...tooltip, visible: false };
    }
  }

  function handlePointerUp(e) {
    if (!canvas) return;
    canvas.releasePointerCapture(e.pointerId);
    if (isPanning && !didDrag) {
      const rect = canvas.getBoundingClientRect();
      const found = hitTest(e.clientX - rect.left, e.clientY - rect.top);
      if (found && found.discovered && found.id !== currentRoomId) {
        if (found.id === travelTargetId) cancelTravel();
        else {
          cancelTravel();
          startTravel(found.id);
        }
        draw();
      }
    }
    isPanning = false;
    didDrag = false;
  }

  function handleMouseLeave() {
    tooltip = { ...tooltip, visible: false };
  }

  function handleWheel(e) {
    e.preventDefault();
    const next = Math.min(2.4, Math.max(0.45, scale * (e.deltaY < 0 ? 1.12 : 0.89)));
    scale = next;
    draw();
  }

  function hitTest(mx, my) {
    for (let i = hitPlaces.length - 1; i >= 0; i--) {
      const h = hitPlaces[i];
      const dx = mx - h.px;
      const dy = my - h.py;
      if (dx * dx + dy * dy <= h.r * h.r) return h.place;
    }
    return null;
  }

  function recenter() {
    panOffsetX = 0;
    panOffsetY = 0;
    draw();
  }

  function toggleMaximize() {
    maximized = !maximized;
    setTimeout(() => resizeCanvas(), 30);
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) toggleMaximize();
  }

  function handleBackdropKeydown(e) {
    if (e.key === 'Escape') toggleMaximize();
  }

  function selectLayer(id) {
    activeLayer = id;
    if (store && store.setAtlasLayer) store.setAtlasLayer(id);
    panOffsetX = 0;
    panOffsetY = 0;
    draw();
  }

  function project(place, origin, w, h) {
    const spacing = CELL * scale;
    return {
      px: Math.round(w / 2 + panOffsetX + (place.x - origin.x) * spacing),
      py: Math.round(h / 2 + panOffsetY + (place.y - origin.y) * spacing),
    };
  }

  function drawSmoothHull(ctx, pts) {
    if (!pts || pts.length === 0) return;
    if (pts.length === 1) {
      ctx.arc(pts[0][0], pts[0][1], CELL * scale * 0.7, 0, Math.PI * 2);
      return;
    }
    ctx.moveTo(pts[0][0], pts[0][1]);
    for (let i = 0; i < pts.length; i++) {
      const p0 = pts[i];
      const p1 = pts[(i + 1) % pts.length];
      const mx = (p0[0] + p1[0]) / 2;
      const my = (p0[1] + p1[1]) / 2;
      ctx.quadraticCurveTo(p0[0], p0[1], mx, my);
    }
    ctx.closePath();
  }

  function drawPlaceGlyph(ctx, place, px, py) {
    const r = place.landmark ? 8 : place.kind === 'uncharted' ? 5 : 6;
    ctx.save();
    ctx.translate(px, py);
    if (place.current) {
      ctx.beginPath();
      ctx.arc(0, 0, r + 6, 0, Math.PI * 2);
      ctx.strokeStyle = 'rgba(245, 158, 11, 0.85)';
      ctx.lineWidth = 2;
      ctx.stroke();
    }
    if (place.id === travelTargetId) {
      ctx.beginPath();
      ctx.arc(0, 0, r + 5, 0, Math.PI * 2);
      ctx.strokeStyle = '#22d3ee';
      ctx.lineWidth = 2;
      ctx.stroke();
    }
    ctx.fillStyle = place.current ? '#f59e0b' : (BIOME[place.biome] || BIOME.wild).ink;
    ctx.strokeStyle = '#2a241c';
    ctx.lineWidth = 1.2;
    if (place.kind === 'uncharted') {
      ctx.setLineDash([2, 2]);
      ctx.beginPath();
      ctx.arc(0, 0, r, 0, Math.PI * 2);
      ctx.strokeStyle = '#6b6358';
      ctx.stroke();
      ctx.setLineDash([]);
    } else if (place.kind === 'dungeon') {
      ctx.beginPath();
      ctx.moveTo(0, -r);
      ctx.lineTo(r, 0);
      ctx.lineTo(0, r);
      ctx.lineTo(-r, 0);
      ctx.closePath();
      ctx.fill();
      ctx.stroke();
    } else if (place.kind === 'settlement') {
      ctx.beginPath();
      ctx.moveTo(0, -r);
      ctx.lineTo(r, -1);
      ctx.lineTo(r, r);
      ctx.lineTo(-r, r);
      ctx.lineTo(-r, -1);
      ctx.closePath();
      ctx.fill();
      ctx.stroke();
    } else if (place.kind === 'water') {
      ctx.beginPath();
      ctx.moveTo(0, r);
      ctx.quadraticCurveTo(r, 0, 0, -r);
      ctx.quadraticCurveTo(-r, 0, 0, r);
      ctx.fill();
      ctx.stroke();
    } else if (place.landmark) {
      star(ctx, 0, 0, r + 1, r * 0.45, 5);
      ctx.fill();
      ctx.stroke();
    } else {
      ctx.beginPath();
      ctx.arc(0, 0, r, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();
    }
    ctx.restore();
    return r + (place.current ? 8 : 4);
  }

  function star(ctx, x, y, r, ir, n) {
    ctx.beginPath();
    for (let i = 0; i < n * 2; i++) {
      const ang = (Math.PI / n) * i - Math.PI / 2;
      const rad = i % 2 === 0 ? r : ir;
      const fn = i === 0 ? ctx.moveTo : ctx.lineTo;
      fn.call(ctx, x + Math.cos(ang) * rad, y + Math.sin(ang) * rad);
    }
    ctx.closePath();
  }

  function originPlace() {
    const here = (atlas.places || []).find(p => p.id === currentRoomId && p.layer === activeLayer);
    if (here) return here;
    const layerPlaces = visiblePlaces;
    if (layerPlaces.length === 0) return { x: 0, y: 0 };
    const discovered = layerPlaces.filter(p => p.discovered);
    const src = discovered.length ? discovered : layerPlaces;
    const sx = src.reduce((a, p) => a + p.x, 0) / src.length;
    const sy = src.reduce((a, p) => a + p.y, 0) / src.length;
    return { x: sx, y: sy };
  }

  function draw() {
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;
    const w = canvas.width / dpr;
    const h = canvas.height / dpr;
    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, w, h);

    const paper = ctx.createRadialGradient(w * 0.5, h * 0.4, 12, w * 0.5, h * 0.5, Math.max(w, h) * 0.75);
    paper.addColorStop(0, '#e7d7b1');
    paper.addColorStop(1, '#cbb892');
    ctx.fillStyle = paper;
    ctx.fillRect(0, 0, w, h);
    ctx.fillStyle = 'rgba(90, 70, 40, 0.04)';
    for (let y = 0; y < h; y += 7) {
      ctx.fillRect(0, y, w, 1);
    }

    const origin = originPlace();
    const spacing = CELL * scale;
    hitPlaces = [];

    for (const region of visibleRegions) {
      const pts = (region.hull || []).map(([x, y]) => {
        const px = w / 2 + panOffsetX + (x - origin.x) * spacing;
        const py = h / 2 + panOffsetY + (y - origin.y) * spacing;
        return [px, py];
      });
      const biome = BIOME[region.biome] || BIOME.wild;
      ctx.beginPath();
      drawSmoothHull(ctx, pts);
      ctx.fillStyle = biome.fill;
      ctx.fill();
      ctx.strokeStyle = biome.ink;
      ctx.globalAlpha = 0.45;
      ctx.lineWidth = 1.5;
      ctx.stroke();
      ctx.globalAlpha = 1;
      if (pts.length) {
        const cx = pts.reduce((a, p) => a + p[0], 0) / pts.length;
        const cy = pts.reduce((a, p) => a + p[1], 0) / pts.length;
        ctx.save();
        ctx.font = '600 11px Georgia, serif';
        ctx.fillStyle = 'rgba(62, 48, 28, 0.72)';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(region.name || '', cx, cy - spacing * 0.55);
        ctx.restore();
      }
    }

    const byId = {};
    for (const p of atlas.places || []) byId[p.id] = p;

    for (const path of visiblePaths) {
      const a = byId[path.from];
      const b = byId[path.to];
      if (!a || !b) continue;
      const pa = project(a, origin, w, h);
      const pb = project(b, origin, w, h);
      const mx = (pa.px + pb.px) / 2;
      const my = (pa.py + pb.py) / 2;
      const dx = pb.py - pa.py;
      const dy = pa.px - pb.px;
      const len = Math.hypot(dx, dy) || 1;
      const wobble = ((hash32(path.from + path.to + path.dir) % 11) - 5) / 5 * 10 * scale;
      const cx = mx + dx / len * wobble;
      const cy = my + dy / len * wobble;
      const onTravel = travelPathRoomIds.has(path.from) || travelPathRoomIds.has(path.to);
      ctx.beginPath();
      ctx.moveTo(pa.px, pa.py);
      ctx.quadraticCurveTo(cx, cy, pb.px, pb.py);
      ctx.strokeStyle = onTravel ? '#0ea5e9' : (path.kind === 'road' ? '#6b4f2e' : '#5a4a38');
      ctx.lineWidth = onTravel ? 3 : path.kind === 'road' ? 2.4 : 1.4;
      ctx.globalAlpha = a.discovered && b.discovered ? 0.9 : 0.35;
      if (path.kind === 'stair') ctx.setLineDash([4, 3]);
      else if (path.kind === 'hidden' || path.kind === 'passage') ctx.setLineDash([2, 3]);
      else ctx.setLineDash([]);
      ctx.stroke();
      ctx.setLineDash([]);
      ctx.globalAlpha = 1;
    }

    for (const place of visiblePlaces) {
      const { px, py } = project(place, origin, w, h);
      const r = drawPlaceGlyph(ctx, place, px, py);
      hitPlaces.push({ px, py, r: r + 4, place });
      if (place.current && place.name) {
        ctx.save();
        ctx.font = '700 11px Georgia, serif';
        ctx.fillStyle = '#3b2c16';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'top';
        ctx.fillText(place.name, px, py + r + 2);
        ctx.restore();
      } else if (place.landmark && place.discovered && place.name && maximized) {
        ctx.save();
        ctx.font = '600 10px Georgia, serif';
        ctx.fillStyle = 'rgba(59, 44, 22, 0.85)';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'top';
        ctx.fillText(place.name, px, py + r + 1);
        ctx.restore();
      }
    }

    ctx.restore();
  }

  function resizeCanvas() {
    if (!canvas || !container) return;
    const rect = container.getBoundingClientRect();
    if (rect.width === 0 || rect.height === 0) return;
    const dpr = window.devicePixelRatio || 1;
    canvas.width = Math.round(rect.width * dpr);
    canvas.height = Math.round(rect.height * dpr);
    canvas.style.width = rect.width + 'px';
    canvas.style.height = rect.height + 'px';
    draw();
  }

  let observedContainer = null;
  $: if (container !== observedContainer) {
    if (resizeObserver) resizeObserver.disconnect();
    observedContainer = container;
    if (container) {
      resizeObserver = new ResizeObserver(() => resizeCanvas());
      resizeObserver.observe(container);
      resizeCanvas();
    }
  }

  onDestroy(() => {
    if (resizeObserver) resizeObserver.disconnect();
    cancelTravel();
    maximized = false;
  });
</script>

<style>
  .minimap-container {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
    background: #cbb892;
    border-radius: var(--panel-radius, 8px);
    border: 1px solid var(--panel-border, rgba(255, 255, 255, 0.1));
    box-shadow: var(--panel-shadow, none);
  }
  .minimap-header {
    position: absolute;
    top: 0; left: 0; right: 0;
    padding: 6px 10px;
    font-family: var(--font-display, Georgia, serif);
    font-size: 11px;
    font-weight: 600;
    color: #4a3b28;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    display: flex;
    align-items: center;
    gap: 6px;
    z-index: 2;
    pointer-events: none;
  }
  .minimap-header i { font-size: 14px; color: #6b5a44; }
  .header-spacer { flex: 1; }
  .header-btn {
    pointer-events: auto;
    cursor: pointer;
    background: none;
    border: none;
    color: #5c4a32;
    padding: 2px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
  }
  .header-btn:hover { background: rgba(90, 60, 20, 0.12); }
  .header-btn.cancel { color: #b45309; }
  .layer-tabs {
    pointer-events: auto;
    display: flex;
    gap: 2px;
  }
  .layer-tab {
    border: 1px solid rgba(90, 70, 40, 0.35);
    background: rgba(255, 248, 230, 0.5);
    color: #5c4a32;
    font-size: 10px;
    text-transform: none;
    letter-spacing: 0;
    padding: 1px 6px;
    border-radius: 999px;
    cursor: pointer;
  }
  .layer-tab.active {
    background: #5c4a32;
    color: #f4e6c5;
  }
  .travel-indicator {
    pointer-events: auto;
    font-size: 10px;
    font-weight: 600;
    color: #0e7490;
    text-transform: none;
    letter-spacing: 0;
  }
  .canvas-wrap { position: absolute; inset: 0; }
  canvas { display: block; cursor: grab; touch-action: none; }
  canvas:active { cursor: grabbing; }
  .tooltip {
    position: absolute;
    background: rgba(42, 32, 20, 0.92);
    color: #f4e6c5;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    pointer-events: none;
    white-space: nowrap;
    z-index: 10;
    transform: translate(-50%, -100%);
    margin-top: -8px;
  }
  .empty-state {
    display: flex; align-items: center; justify-content: center;
    height: 100%; color: #6b5a44; font-size: 12px; text-align: center; padding: 16px;
  }
  .maximized-placeholder {
    display: flex; align-items: center; justify-content: center;
    height: 100%; color: #6b5a44; font-size: 12px;
  }
  .modal-backdrop {
    position: fixed; inset: 0; z-index: 10000;
    background: rgba(0, 0, 0, 0.7);
    display: flex; align-items: center; justify-content: center; padding: 24px;
  }
  .modal-panel {
    width: 100%; height: 100%;
    background: #cbb892;
    border: 1px solid #8a7350;
    border-radius: 12px;
    display: flex; flex-direction: column; overflow: hidden; position: relative;
  }
  .modal-header {
    display: flex; align-items: center; gap: 8px;
    padding: 10px 14px; border-bottom: 1px solid rgba(90,70,40,0.25);
    font-size: 13px; font-weight: 600; color: #3b2c16; flex-shrink: 0;
  }
  .modal-canvas-wrap { flex: 1; position: relative; overflow: hidden; }
</style>

{#if maximized}
  <div class="modal-backdrop" use:portal on:click={handleBackdropClick} on:keydown={handleBackdropKeydown}>
    <div class="modal-panel">
      <div class="modal-header">
        <i class="material-icons">map</i>
        Atlas
        {#if isTraveling}<span class="travel-indicator">Traveling…</span>{/if}
        <div class="layer-tabs">
          {#each atlas.layers || [] as layer}
            <button class="layer-tab" class:active={activeLayer === layer.id} on:click={() => selectLayer(layer.id)}>
              {layer.name}
            </button>
          {/each}
        </div>
        <span class="header-spacer"></span>
        {#if panOffsetX !== 0 || panOffsetY !== 0}
          <button class="header-btn" title="Re-center" on:click={recenter}>
            <i class="material-icons" style="font-size: inherit">my_location</i>
          </button>
        {/if}
        {#if isTraveling}
          <button class="header-btn cancel" title="Cancel travel" on:click={cancelTravel}>
            <i class="material-icons" style="font-size: inherit">close</i>
          </button>
        {/if}
        <button class="header-btn" title="Close map" on:click={toggleMaximize}>
          <i class="material-icons" style="font-size: inherit">close</i>
        </button>
      </div>
      <div class="modal-canvas-wrap" bind:this={container}>
        <canvas
          bind:this={canvas}
          on:pointerdown={handlePointerDown}
          on:pointermove={handlePointerMove}
          on:pointerup={handlePointerUp}
          on:pointerleave={handleMouseLeave}
          on:wheel={handleWheel}
        ></canvas>
        {#if tooltip.visible}
          <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">{tooltip.text}</div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<div class="minimap-container">
  <div class="minimap-header">
    <button class="header-btn" title="Expand map" on:click={toggleMaximize}>
      <i class="material-icons" style="font-size: inherit">open_in_full</i>
    </button>
    <i class="material-icons">map</i>
    Atlas
    {#if isTraveling && !maximized}<span class="travel-indicator">Traveling…</span>{/if}
    <div class="layer-tabs">
      {#each atlas.layers || [] as layer}
        <button class="layer-tab" class:active={activeLayer === layer.id} on:click={() => selectLayer(layer.id)}>
          {layer.name}
        </button>
      {/each}
    </div>
    <span class="header-spacer"></span>
    {#if !maximized && (panOffsetX !== 0 || panOffsetY !== 0)}
      <button class="header-btn" title="Re-center" on:click={recenter}>
        <i class="material-icons" style="font-size: inherit">my_location</i>
      </button>
    {/if}
    {#if isTraveling && !maximized}
      <button class="header-btn cancel" title="Cancel travel" on:click={cancelTravel}>
        <i class="material-icons" style="font-size: inherit">close</i>
      </button>
    {/if}
  </div>

  {#if maximized}
    <div class="maximized-placeholder">Map is expanded</div>
  {:else if currentRoomId && visiblePlaces.length > 0}
    <div class="canvas-wrap" bind:this={container}>
      <canvas
        bind:this={canvas}
        on:pointerdown={handlePointerDown}
        on:pointermove={handlePointerMove}
        on:pointerup={handlePointerUp}
        on:pointerleave={handleMouseLeave}
        on:wheel={handleWheel}
      ></canvas>
      {#if tooltip.visible}
        <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">{tooltip.text}</div>
      {/if}
    </div>
  {:else}
    <div class="empty-state">Walk the world to fill your atlas.</div>
  {/if}
</div>
