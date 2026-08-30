<script>
  import { onDestroy, tick } from 'svelte';

  export let store = null;
  export let sendMessage = null;

  const BIOME = {
    meadow: { fill: 'rgba(132, 184, 92, 0.28)', ink: '#86efac', path: '#4ade80' },
    forest: { fill: 'rgba(52, 120, 72, 0.32)', ink: '#34d399', path: '#22c55e' },
    water: { fill: 'rgba(56, 132, 176, 0.30)', ink: '#7dd3fc', path: '#38bdf8' },
    dungeon: { fill: 'rgba(148, 92, 64, 0.30)', ink: '#fdba74', path: '#fb923c' },
    settlement: { fill: 'rgba(196, 148, 72, 0.28)', ink: '#fcd34d', path: '#fbbf24' },
    wild: { fill: 'rgba(120, 130, 150, 0.22)', ink: '#cbd5e1', path: '#94a3b8' },
  };

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
  let hits = [];

  function emptyAtlas() {
    return { characterId: '', currentRoomId: '', currentLayer: '', layers: [], places: [], paths: [], regions: [] };
  }

  function portal(node) {
    document.body.appendChild(node);
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
    atlas = $store.atlas && Array.isArray($store.atlas.places) ? $store.atlas : emptyAtlas();
    const newRoomId = $store.currentRoomId || atlas.currentRoomId || null;
    if (newRoomId !== currentRoomId) {
      currentRoomId = newRoomId;
      panX = 0;
      panY = 0;
      userScale = 1;
      if (isTraveling && newRoomId) advanceTravel(newRoomId);
    }
    activeLayer = resolveLayer(atlas, currentRoomId, $store.atlasLayer || activeLayer);
  }

  $: visiblePlaces = (atlas.places || []).filter(p => p.layer === activeLayer);
  $: visibleRegions = (atlas.regions || []).filter(r => r.layer === activeLayer);
  $: layers = atlas.layers || [];

  $: if (widgetCanvas || modalCanvas || visiblePlaces || panX || panY || userScale || maximized) {
    draw();
  }

  function placeById(id) {
    return (atlas.places || []).find(p => p.id === id);
  }

  function camera(places, regions, w, h) {
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    const consider = (x, y) => {
      if (x < minX) minX = x;
      if (y < minY) minY = y;
      if (x > maxX) maxX = x;
      if (y > maxY) maxY = y;
    };
    for (const p of places) consider(p.x, p.y);
    for (const region of regions || []) {
      for (const pt of region.hull || []) consider(pt[0], pt[1]);
    }
    if (!isFinite(minX)) {
      minX = maxX = minY = maxY = 0;
    }
    const dx = Math.max(1.4, maxX - minX);
    const dy = Math.max(1.4, maxY - minY);
    const pad = Math.max(28, Math.min(w, h) * 0.16);
    const fit = Math.min((w - pad * 2) / dx, (h - pad * 2) / dy);
    const spacing = Math.max(22, Math.min(fit * userScale, 110));
    return { spacing, ox: (minX + maxX) / 2, oy: (minY + maxY) / 2 };
  }

  function project(place, cam, w, h) {
    return {
      px: w / 2 + panX + (place.x - cam.ox) * cam.spacing,
      py: h / 2 + panY + (place.y - cam.oy) * cam.spacing,
    };
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

  function hash32(s) {
    let h = 2166136261;
    for (let i = 0; i < s.length; i++) h = Math.imul(h ^ s.charCodeAt(i), 16777619);
    return h >>> 0;
  }

  function paint(canvas, wrap) {
    if (!canvas || !wrap) return;
    const rect = wrap.getBoundingClientRect();
    if (rect.width < 4 || rect.height < 4) return;
    const dpr = window.devicePixelRatio || 1;
    const w = rect.width;
    const h = rect.height;
    const pw = Math.round(w * dpr);
    const ph = Math.round(h * dpr);
    if (canvas.width !== pw) canvas.width = pw;
    if (canvas.height !== ph) canvas.height = ph;
    canvas.style.width = w + 'px';
    canvas.style.height = h + 'px';

    const ctx = canvas.getContext('2d');
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, w, h);

    ctx.fillStyle = '#0e141c';
    ctx.fillRect(0, 0, w, h);
    ctx.fillStyle = 'rgba(148, 163, 184, 0.07)';
    for (let x = 16; x < w; x += 22) {
      for (let y = 16; y < h; y += 22) {
        ctx.fillRect(x, y, 1, 1);
      }
    }

    if (!visiblePlaces.length) {
      ctx.fillStyle = '#64748b';
      ctx.font = '12px sans-serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(currentRoomId ? 'Charting this floor…' : 'Walk to fill your atlas', w / 2, h / 2);
      return;
    }

    const cam = camera(visiblePlaces, visibleRegions, w, h);
    const byId = {};
    for (const p of atlas.places || []) byId[p.id] = p;

    for (const region of visibleRegions) {
      const pts = (region.hull || []).map(([x, y]) => [
        w / 2 + panX + (x - cam.ox) * cam.spacing,
        h / 2 + panY + (y - cam.oy) * cam.spacing,
      ]);
      if (!pts.length) continue;
      const biome = BIOME[region.biome] || BIOME.wild;
      ctx.beginPath();
      if (pts.length === 1) {
        ctx.arc(pts[0][0], pts[0][1], cam.spacing * 0.55, 0, Math.PI * 2);
      } else {
        ctx.moveTo((pts[0][0] + pts[pts.length - 1][0]) / 2, (pts[0][1] + pts[pts.length - 1][1]) / 2);
        for (let i = 0; i < pts.length; i++) {
          const p0 = pts[i];
          const p1 = pts[(i + 1) % pts.length];
          ctx.quadraticCurveTo(p0[0], p0[1], (p0[0] + p1[0]) / 2, (p0[1] + p1[1]) / 2);
        }
        ctx.closePath();
      }
      ctx.fillStyle = biome.fill;
      ctx.fill();
      ctx.strokeStyle = biome.ink;
      ctx.globalAlpha = 0.55;
      ctx.lineWidth = 1.4;
      ctx.stroke();
      ctx.globalAlpha = 1;
      const cx = pts.reduce((a, p) => a + p[0], 0) / pts.length;
      const cy = pts.reduce((a, p) => a + p[1], 0) / pts.length;
      ctx.font = '600 11px sans-serif';
      ctx.fillStyle = 'rgba(226, 232, 240, 0.72)';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(region.name || '', cx, cy - cam.spacing * 0.42);
    }

    const layerPaths = (atlas.paths || []).filter(path => {
      const a = byId[path.from];
      const b = byId[path.to];
      return a && b && (a.layer === activeLayer || b.layer === activeLayer);
    });

    for (const path of layerPaths) {
      const a = byId[path.from];
      const b = byId[path.to];
      const pa = project(a, cam, w, h);
      const pb = project(b, cam, w, h);
      const mx = (pa.px + pb.px) / 2;
      const my = (pa.py + pb.py) / 2;
      const dx = pb.py - pa.py;
      const dy = pa.px - pb.px;
      const len = Math.hypot(dx, dy) || 1;
      const wobble = ((hash32(path.from + '>' + path.to) % 9) - 4) / 4 * Math.min(14, cam.spacing * 0.18);
      const onTravel = travelPathRoomIds.has(path.from) && travelPathRoomIds.has(path.to)
        || travelPathRoomIds.has(path.to) && path.from === currentRoomId;
      ctx.beginPath();
      ctx.moveTo(pa.px, pa.py);
      ctx.quadraticCurveTo(mx + dx / len * wobble, my + dy / len * wobble, pb.px, pb.py);
      ctx.strokeStyle = onTravel ? '#38bdf8' : (BIOME[a.biome] || BIOME.wild).path;
      ctx.lineWidth = onTravel ? 3 : path.kind === 'road' ? 2.6 : 2;
      ctx.globalAlpha = a.discovered && b.discovered ? 0.95 : 0.4;
      if (path.kind === 'stair' || path.kind === 'hidden' || path.kind === 'passage') {
        ctx.setLineDash([5, 4]);
      } else {
        ctx.setLineDash([]);
      }
      ctx.stroke();
      ctx.setLineDash([]);
      ctx.globalAlpha = 1;
    }

    const localHits = [];
    for (const place of visiblePlaces) {
      const { px, py } = project(place, cam, w, h);
      const r = drawGlyph(ctx, place, px, py, cam.spacing);
      localHits.push({ px, py, r: r + 6, place });
      if (place.discovered && place.name) {
        const showName = place.current || place.landmark || maximized || visiblePlaces.length <= 8;
        if (showName) {
          ctx.font = place.current ? '700 11px sans-serif' : '600 10px sans-serif';
          ctx.textAlign = 'center';
          ctx.textBaseline = 'top';
          ctx.lineWidth = 3;
          ctx.strokeStyle = 'rgba(8, 12, 18, 0.85)';
          ctx.strokeText(place.name, px, py + r + 2);
          ctx.fillStyle = place.current ? '#fbbf24' : '#e2e8f0';
          ctx.fillText(place.name, px, py + r + 2);
        }
      }
    }
    hits = localHits;
  }

  function drawGlyph(ctx, place, px, py, spacing) {
    const r = Math.max(6, Math.min(11, spacing * 0.16)) * (place.landmark || place.current ? 1.25 : 1);
    ctx.save();
    ctx.translate(px, py);
    if (place.current) {
      ctx.beginPath();
      ctx.arc(0, 0, r + 7, 0, Math.PI * 2);
      ctx.strokeStyle = 'rgba(251, 191, 36, 0.9)';
      ctx.lineWidth = 2.4;
      ctx.stroke();
      ctx.beginPath();
      ctx.arc(0, 0, r + 3, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(251, 191, 36, 0.18)';
      ctx.fill();
    }
    if (place.id === travelTargetId) {
      ctx.beginPath();
      ctx.arc(0, 0, r + 6, 0, Math.PI * 2);
      ctx.strokeStyle = '#22d3ee';
      ctx.lineWidth = 2;
      ctx.stroke();
    }
    const ink = place.current ? '#fbbf24' : (BIOME[place.biome] || BIOME.wild).ink;
    ctx.fillStyle = ink;
    ctx.strokeStyle = '#020617';
    ctx.lineWidth = 1.4;
    if (place.kind === 'uncharted') {
      ctx.setLineDash([2, 2]);
      ctx.beginPath();
      ctx.arc(0, 0, r, 0, Math.PI * 2);
      ctx.strokeStyle = '#94a3b8';
      ctx.stroke();
      ctx.setLineDash([]);
    } else if (place.kind === 'dungeon') {
      ctx.beginPath();
      ctx.moveTo(0, -r); ctx.lineTo(r, 0); ctx.lineTo(0, r); ctx.lineTo(-r, 0);
      ctx.closePath(); ctx.fill(); ctx.stroke();
    } else if (place.kind === 'settlement') {
      ctx.beginPath();
      ctx.moveTo(0, -r);
      ctx.lineTo(r, -1); ctx.lineTo(r, r); ctx.lineTo(-r, r); ctx.lineTo(-r, -1);
      ctx.closePath(); ctx.fill(); ctx.stroke();
    } else if (place.kind === 'water') {
      ctx.beginPath();
      ctx.moveTo(0, r);
      ctx.quadraticCurveTo(r, 0, 0, -r);
      ctx.quadraticCurveTo(-r, 0, 0, r);
      ctx.fill(); ctx.stroke();
    } else if (place.landmark) {
      ctx.beginPath();
      for (let i = 0; i < 10; i++) {
        const ang = (Math.PI / 5) * i - Math.PI / 2;
        const rad = i % 2 === 0 ? r + 1 : r * 0.45;
        if (i === 0) ctx.moveTo(Math.cos(ang) * rad, Math.sin(ang) * rad);
        else ctx.lineTo(Math.cos(ang) * rad, Math.sin(ang) * rad);
      }
      ctx.closePath(); ctx.fill(); ctx.stroke();
    } else {
      ctx.beginPath();
      ctx.arc(0, 0, r, 0, Math.PI * 2);
      ctx.fill(); ctx.stroke();
    }
    ctx.restore();
    return r;
  }

  function draw() {
    paint(widgetCanvas, widgetWrap);
    if (maximized) paint(modalCanvas, modalWrap);
  }

  function hitTest(canvas, mx, my) {
    if (!canvas) return null;
    for (let i = hits.length - 1; i >= 0; i--) {
      const h = hits[i];
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
      draw();
      return;
    }
    const rect = canvas.getBoundingClientRect();
    const found = hitTest(canvas, e.clientX - rect.left, e.clientY - rect.top);
    if (found) {
      let text = found.discovered ? (found.name || found.id) : 'Uncharted';
      if (found.areaName && found.discovered) text += ' · ' + found.areaName;
      if (found.current) text += ' (you are here)';
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
        draw();
      }
    }
    isPanning = false;
    didDrag = false;
  }

  function onWheel(e) {
    e.preventDefault();
    userScale = Math.min(2.6, Math.max(0.55, userScale * (e.deltaY < 0 ? 1.12 : 0.89)));
    draw();
  }

  function recenter() {
    panX = 0;
    panY = 0;
    userScale = 1;
    draw();
  }

  async function toggleMaximize() {
    maximized = !maximized;
    if (maximized) {
      panX = 0;
      panY = 0;
      userScale = 1;
      await tick();
      draw();
    }
  }

  function selectLayer(id) {
    activeLayer = id;
    if (store && store.setAtlasLayer) store.setAtlasLayer(id);
    panX = 0;
    panY = 0;
    userScale = 1;
    draw();
  }

  let observedWidget = null;
  let observedModal = null;
  $: if (widgetWrap !== observedWidget) {
    if (widgetObserver) widgetObserver.disconnect();
    observedWidget = widgetWrap;
    if (widgetWrap) {
      widgetObserver = new ResizeObserver(() => draw());
      widgetObserver.observe(widgetWrap);
      draw();
    }
  }
  $: if (modalWrap !== observedModal) {
    if (modalObserver) modalObserver.disconnect();
    observedModal = modalWrap;
    if (modalWrap) {
      modalObserver = new ResizeObserver(() => draw());
      modalObserver.observe(modalWrap);
      draw();
    }
  }

  onDestroy(() => {
    if (widgetObserver) widgetObserver.disconnect();
    if (modalObserver) modalObserver.disconnect();
    cancelTravel();
    maximized = false;
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
  .stage { flex: 1 1 auto; min-height: 0; position: relative; }
  canvas { display: block; width: 100%; height: 100%; cursor: grab; touch-action: none; }
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
    z-index: 10000;
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
</style>

<div class="atlas">
  <div class="toolbar">
    <button class="icon-btn" title="Expand map" on:click={toggleMaximize}>
      <i class="material-icons">open_in_full</i>
    </button>
    <i class="material-icons">map</i>
    Atlas
    {#if isTraveling}<span class="travel">Traveling…</span>{/if}
    <div class="layer-tabs">
      {#each layers as layer}
        <button class="layer-tab" class:active={activeLayer === layer.id} on:click={() => selectLayer(layer.id)}>{layer.name}</button>
      {/each}
    </div>
    <span class="spacer"></span>
    {#if panX !== 0 || panY !== 0 || userScale !== 1}
      <button class="icon-btn" title="Fit map" on:click={recenter}>
        <i class="material-icons">my_location</i>
      </button>
    {/if}
    {#if isTraveling}
      <button class="icon-btn cancel" title="Cancel travel" on:click={cancelTravel}>
        <i class="material-icons">close</i>
      </button>
    {/if}
  </div>
  <div class="stage" bind:this={widgetWrap}>
    <canvas
      bind:this={widgetCanvas}
      on:pointerdown={pointerDown}
      on:pointermove={(e) => pointerMove(e, widgetCanvas)}
      on:pointerup={(e) => pointerUp(e, widgetCanvas)}
      on:pointerleave={() => tooltip = { ...tooltip, visible: false }}
      on:wheel={onWheel}
    ></canvas>
    {#if tooltip.visible && !maximized}
      <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">{tooltip.text}</div>
    {/if}
  </div>
</div>

{#if maximized}
  <div class="backdrop" use:portal on:click={(e) => { if (e.target === e.currentTarget) toggleMaximize(); }} on:keydown={(e) => { if (e.key === 'Escape') toggleMaximize(); }}>
    <div class="modal">
      <div class="toolbar">
        <i class="material-icons">map</i>
        Atlas
        {#if isTraveling}<span class="travel">Traveling…</span>{/if}
        <div class="layer-tabs">
          {#each layers as layer}
            <button class="layer-tab" class:active={activeLayer === layer.id} on:click={() => selectLayer(layer.id)}>{layer.name}</button>
          {/each}
        </div>
        <span class="spacer"></span>
        {#if panX !== 0 || panY !== 0 || userScale !== 1}
          <button class="icon-btn" title="Fit map" on:click={recenter}>
            <i class="material-icons">my_location</i>
          </button>
        {/if}
        {#if isTraveling}
          <button class="icon-btn cancel" title="Cancel travel" on:click={cancelTravel}>
            <i class="material-icons">close</i>
          </button>
        {/if}
        <button class="icon-btn" title="Close" on:click={toggleMaximize}>
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
