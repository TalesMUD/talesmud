/**
 * Veilspan public world map — WoW-style continent → zone drill-down.
 * PUBLIC fields only (map-data.json is built with internal lore stripped).
 */
(() => {
  const NS = 'http://www.w3.org/2000/svg';
  const ASSET_V = 'map10';
  const viewport = document.getElementById('map-viewport');
  const svg = document.getElementById('map-svg');
  const camera = document.getElementById('camera');
  const continentRoot = document.getElementById('continent-root');
  const zoneRoot = document.getElementById('zone-root');
  const zonePlate = document.getElementById('zone-plate');
  const panel = document.getElementById('detail-panel');
  const panelBody = document.getElementById('panel-body');
  const panelTitle = document.getElementById('panel-title');
  const panelMeta = document.getElementById('panel-meta');
  const panelEyebrow = document.getElementById('panel-eyebrow');
  const chapterLabel = document.getElementById('chapter-label');
  const btnBack = document.getElementById('btn-back');
  const appEl = document.getElementById('app');

  const state = {
    data: null,
    view: 'continent', // continent | zone
    zoneId: null,
    scale: 0.5,
    tx: 0,
    ty: 0,
    minScale: 0.22,
    maxScale: 4.0,
    lodMode: 'auto',
    dragging: false,
    moved: false,
    downX: 0,
    downY: 0,
    lastX: 0,
    lastY: 0,
    pointers: new Map(),
    pinchStartDist: 0,
    pinchStartScale: 1,
    selectedId: null,
    hoverZoneId: null,
  };

  const poiColors = {
    inn: '#e8a849',
    bind: '#3ddc84',
    merchant: '#c8a84e',
    npc: '#8be9fd',
    hub: '#f0c674',
    dungeon: '#bd93f9',
    boss: '#ff5555',
    landmark: '#a8c8e8',
    gate: '#7ae8a4',
    town: '#f0c674',
  };

  const poiIconMap = {
    inn: 'inn',
    bind: 'bindstone',
    bindstone: 'bindstone',
    dungeon: 'dungeon',
    boss: 'boss',
    hub: 'town',
    town: 'town',
    gate: 'town',
    landmark: 'road',
    road: 'road',
    merchant: 'town',
    npc: 'town',
  };

  function el(name, attrs = {}, parent) {
    const node = document.createElementNS(NS, name);
    for (const [k, v] of Object.entries(attrs)) {
      if (v != null) node.setAttribute(k, String(v));
    }
    if (parent) parent.appendChild(node);
    return node;
  }

  function absMapAsset(href) {
    if (!href) return href;
    if (href.startsWith('http') || href.startsWith('/')) return href;
    return `/map/${href.replace(/^\.\//, '')}`;
  }

  function withV(href) {
    if (!href) return href;
    return href.includes('?') ? href : `${href}?v=${ASSET_V}`;
  }

  function poiIconHref(kind) {
    const key = poiIconMap[kind];
    return key ? withV(`/map/assets/icons/${key}.png`) : null;
  }

  function markerScale() {
    // Keep settlement icons readable at continent zoom; cap so they don't explode when close.
    return Math.max(0.5, Math.min(2.4, 0.9 / state.scale));
  }

  function updateMarkerScales() {
    const k = markerScale();
    document.querySelectorAll('.city-marker').forEach((n) => {
      const x = Number(n.getAttribute('data-x'));
      const y = Number(n.getAttribute('data-y'));
      n.setAttribute('transform', `translate(${x} ${y}) scale(${k})`);
    });
    if (state.view === 'zone') {
      const zk = Math.max(0.7, Math.min(1.6, 0.9 / state.scale));
      document.querySelectorAll('#zone-pois .poi-marker').forEach((n) => {
        const x = Number(n.getAttribute('data-x'));
        const y = Number(n.getAttribute('data-y'));
        n.setAttribute('transform', `translate(${x} ${y}) scale(${zk})`);
      });
    }
  }

  function applyCamera() {
    camera.setAttribute('transform', `translate(${state.tx} ${state.ty}) scale(${state.scale})`);
    updateLodClass();
    updateMarkerScales();
  }

  function currentLod() {
    if (state.view === 'zone') return 'pois';
    if (state.lodMode !== 'auto') return state.lodMode;
    const lod = state.data?.lod || {};
    if (state.scale < (lod.continent?.maxScale ?? 0.45)) return 'continent';
    if (state.scale < (lod.pois?.minScale ?? 1.8)) return 'zones';
    return 'pois';
  }

  function updateLodClass() {
    const lod = currentLod();
    svg.classList.remove('lod-continent', 'lod-zones', 'lod-pois');
    svg.classList.add(`lod-${lod}`);
  }

  function worldSize() {
    if (state.view === 'zone') {
      const z = zoneById(state.zoneId);
      const vb = z?.zoneMap?.viewBox || [0, 0, 1024, 1024];
      return { w: vb[2], h: vb[3] };
    }
    const vb = state.data.world.viewBox;
    return { w: vb[2], h: vb[3] };
  }

  function viewportSize() {
    const r = viewport.getBoundingClientRect();
    return { w: Math.max(1, r.width), h: Math.max(1, r.height) };
  }

  function syncSvgViewBox() {
    const { w, h } = viewportSize();
    svg.setAttribute('viewBox', `0 0 ${w} ${h}`);
  }

  function fitView() {
    syncSvgViewBox();
    const { w, h } = worldSize();
    const { w: vw, h: vh } = viewportSize();
    const pad = state.view === 'zone' ? 0.97 : 0.98;
    state.scale = Math.min(vw / w, vh / h) * pad;
    state.tx = (vw - w * state.scale) / 2;
    state.ty = (vh - h * state.scale) / 2;
    applyCamera();
  }

  function zoomAt(clientX, clientY, factor) {
    const { w: vw, h: vh } = viewportSize();
    const rect = viewport.getBoundingClientRect();
    const x = clientX - rect.left;
    const y = clientY - rect.top;
    const worldX = (x - state.tx) / state.scale;
    const worldY = (y - state.ty) / state.scale;
    const { w, h } = worldSize();
    const fit = Math.min(vw / w, vh / h);
    let minS = state.view === 'zone' ? fit * 0.82 : Math.min(fit * 0.55, state.minScale);
    let maxS = state.view === 'zone' ? fit * 3.4 : state.maxScale;
    const next = Math.min(maxS, Math.max(minS, state.scale * factor));
    state.scale = next;
    state.tx = x - worldX * state.scale;
    state.ty = y - worldY * state.scale;
    applyCamera();

    if (state.view === 'zone' && state.scale <= fit * 0.84) {
      const zid = state.zoneId;
      exitZone();
      const z = zoneById(zid);
      if (z) focusZone(z);
    }
  }

  function zoneById(id) {
    return state.data.zones.find((z) => z.id === id);
  }

  function cityById(id) {
    return (state.data.cities || []).find((c) => c.id === id);
  }

  function focusZone(z) {
    syncSvgViewBox();
    const { w: vw, h: vh } = viewportSize();
    const { cx, cy, rx, ry } = z.shape;
    const targetW = Math.max(rx * 3.4, 420);
    const targetH = Math.max(ry * 3.4, 420);
    state.scale = Math.min(vw / targetW, vh / targetH, state.maxScale);
    state.tx = vw / 2 - cx * state.scale;
    state.ty = vh / 2 - cy * state.scale;
    applyCamera();
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function openPanel(kind, payload) {
    panel.classList.add('open');
    panel.setAttribute('aria-hidden', 'false');
    if (kind === 'city') {
      const c = payload;
      state.selectedId = c.id;
      const imp = c.importance || 'town';
      panelEyebrow.textContent = (c.visibility === 'fog' ? 'Uncharted · ' : '') +
        (imp === 'capital' ? 'Capital' : imp === 'hub' ? 'Hub' : 'Town');
      panelTitle.textContent = c.name || c.id;
      panelMeta.textContent = [c.faction, c.zoneId].filter(Boolean).join(' · ');
      let html = '';
      if (c.visibility === 'fog') {
        html += `<div class="fog-banner">Fog of war — charts incomplete. Rumors only.</div>`;
      }
      if (c.blurb) html += `<p>${escapeHtml(c.blurb)}</p>`;
      const z = zoneById(c.zoneId);
      if (z && z.zoneMap && state.view === 'continent') {
        html += `<p><button type="button" class="chip" id="btn-enter-zone" data-zone="${escapeHtml(z.id)}">Enter zone map</button></p>`;
      }
      panelBody.innerHTML = html || '<p>No public charts yet.</p>';
      bindEnterBtn();
      return;
    }
    if (kind === 'zone') {
      const z = payload;
      state.selectedId = z.id;
      document.querySelectorAll('.zone-shape.selected').forEach((n) => n.classList.remove('selected'));
      const shape = document.getElementById(`zone-${z.id}`);
      if (shape) shape.classList.add('selected');

      panelEyebrow.textContent = z.fog ? `Uncharted · ${z.id}` : `Zone · ${z.id}`;
      panelTitle.textContent = z.title || z.id;
      panelMeta.textContent = z.levelRange ? `Levels ${z.levelRange}` : '';

      let html = '';
      if (z.fog) {
        html += `<div class="fog-banner">Fog of war — charts incomplete. Rumors only.</div>`;
      }
      for (const line of z.overview || []) {
        html += `<p>${escapeHtml(line)}</p>`;
      }
      if ((z.discoverable || []).length) {
        html += `<h3>To discover</h3><ul class="tag-list">`;
        for (const t of z.discoverable) html += `<li>${escapeHtml(t)}</li>`;
        html += `</ul>`;
      }
      if ((z.connections || []).length) {
        html += `<h3>Roads &amp; gates</h3><ul class="conn-list">`;
        for (const c of z.connections) {
          html += `<li><strong>${escapeHtml(c.label || c.to)}</strong>`;
          if (c.tease) html += ` — ${escapeHtml(c.tease)}`;
          html += `<span class="state">[${escapeHtml(c.state || 'open')}]</span></li>`;
        }
        html += `</ul>`;
      }
      if (!z.fog && (z.pois || []).length) {
        html += `<h3>Points of interest</h3><ul class="poi-list">`;
        for (const p of z.pois) {
          html += `<li><strong>${escapeHtml(p.name)}</strong> <span class="state">(${escapeHtml(p.kind)})</span>`;
          if (p.blurb) html += `<div>${escapeHtml(p.blurb)}</div>`;
          html += `</li>`;
        }
        html += `</ul>`;
      }
      if (z.zoneMap && state.view === 'continent') {
        html += `<p><button type="button" class="chip" id="btn-enter-zone" data-zone="${escapeHtml(z.id)}">Enter zone map</button></p>`;
      }
      panelBody.innerHTML = html;
      bindEnterBtn();
    } else if (kind === 'poi') {
      const p = payload;
      state.selectedId = p.id;
      panelEyebrow.textContent = `POI · ${p.kind || 'landmark'}`;
      panelTitle.textContent = p.name;
      panelMeta.textContent = p.zoneTitle ? `in ${p.zoneTitle}` : '';
      panelBody.innerHTML = `<p>${escapeHtml(p.blurb || 'A notable place on the charts.')}</p>`;
    }
  }

  function bindEnterBtn() {
    const btn = document.getElementById('btn-enter-zone');
    if (!btn) return;
    btn.addEventListener('click', (ev) => {
      ev.stopPropagation();
      const z = zoneById(btn.getAttribute('data-zone'));
      if (z) enterZone(z);
    });
  }

  function closePanel() {
    panel.classList.remove('open');
    panel.setAttribute('aria-hidden', 'true');
    document.querySelectorAll('.zone-shape.selected').forEach((n) => n.classList.remove('selected'));
    state.selectedId = null;
  }

  function enterZone(z, opts = {}) {
    if (!z) return;
    const href = z.zoneMap?.href || z.imagePlate;
    if (!href) {
      focusZone(z);
      openPanel('zone', z);
      return;
    }
    state.view = 'zone';
    state.zoneId = z.id;
    appEl.classList.add('view-zone');
    continentRoot.classList.add('hidden');
    zoneRoot.classList.remove('hidden');
    btnBack.classList.remove('hidden');
    btnBack.setAttribute('aria-hidden', 'false');
    chapterLabel.textContent = z.title || z.id;

    const vb = z.zoneMap?.viewBox || [0, 0, 1024, 1024];
    svg.setAttribute('aria-label', `${z.title || z.id} zone map`);
    zonePlate.setAttribute('href', withV(absMapAsset(href)));
    zonePlate.setAttribute('width', vb[2]);
    zonePlate.setAttribute('height', vb[3]);

    renderZonePois(z);
    fitView();
    openPanel('zone', z);
    if (!opts.skipHash) {
      const hash = `#${z.id}`;
      if (location.hash !== hash) history.replaceState(null, '', hash);
    }
  }

  function exitZone(opts = {}) {
    const prev = state.zoneId;
    state.view = 'continent';
    state.zoneId = null;
    appEl.classList.remove('view-zone');
    continentRoot.classList.remove('hidden');
    zoneRoot.classList.add('hidden');
    btnBack.classList.add('hidden');
    btnBack.setAttribute('aria-hidden', 'true');
    chapterLabel.textContent = state.data.chapter || 'World Map';
    svg.setAttribute('aria-label', 'Veilspan world map');
    document.getElementById('zone-pois').innerHTML = '';
    fitView();
    if (prev) {
      const z = zoneById(prev);
      if (z) focusZone(z);
    }
    if (!opts.skipHash && location.hash) {
      history.replaceState(null, '', location.pathname + location.search);
    }
  }

  function renderZonePois(z) {
    const g = document.getElementById('zone-pois');
    g.innerHTML = '';
    if (z.fog) return;
    for (const p of z.pois || []) {
      if (p.x == null || p.y == null) continue;
      const color = poiColors[p.kind] || '#f0c674';
      const iconHref = poiIconHref(p.kind);
      const wrap = el('g', {
        class: 'poi-marker',
        'data-id': p.id,
        'data-x': p.x,
        'data-y': p.y,
        transform: `translate(${p.x} ${p.y})`,
      }, g);
      if (iconHref) {
        const size = 52;
        el('circle', {
          class: 'poi-badge',
          r: size / 2 + 4,
          fill: 'rgba(6,8,12,0.7)',
          stroke: 'rgba(232,168,73,0.4)',
          'stroke-width': 2,
        }, wrap);
        el('image', {
          class: 'poi-icon',
          href: iconHref,
          x: -size / 2,
          y: -size / 2,
          width: size,
          height: size,
        }, wrap);
      } else {
        el('circle', {
          class: 'poi-dot',
          cx: 0, cy: 0, r: 10,
          fill: color,
        }, wrap);
      }
      const t = el('text', {
        class: 'poi-label',
        x: 0,
        y: 36,
        'text-anchor': 'middle',
      }, wrap);
      t.textContent = p.name;
      wrap.addEventListener('click', (ev) => {
        ev.stopPropagation();
        openPanel('poi', { ...p, zoneTitle: z.title });
      });
    }
  }

  function renderZones() {
    const g = document.getElementById('layer-zones');
    const labels = document.getElementById('layer-labels');
    g.innerHTML = '';
    labels.innerHTML = '';

    for (const z of state.data.zones) {
      const { cx, cy, rx, ry } = z.shape;
      const vis = z.visibility || (z.fog ? 'fog' : 'live');
      const shape = el('ellipse', {
        id: `zone-${z.id}`,
        class: `zone-shape ${vis}`,
        cx, cy, rx, ry,
        'data-id': z.id,
      }, g);
      shape.addEventListener('click', (ev) => {
        ev.stopPropagation();
        if (state.moved) return;
        openPanel('zone', z);
        if (z.zoneMap) enterZone(z);
        else focusZone(z);
      });
      shape.addEventListener('mouseenter', () => {
        state.hoverZoneId = z.id;
      });
      shape.addEventListener('mouseleave', () => {
        if (state.hoverZoneId === z.id) state.hoverZoneId = null;
      });

      const label = el('text', {
        x: cx, y: cy + 6,
        class: `zone-label${z.fog ? ' fog-label' : ''}`,
        'text-anchor': 'middle',
      }, labels);
      label.textContent = z.title || z.id;

      if (!z.fog && z.levelRange) {
        const lvl = el('text', {
          x: cx, y: cy + 24,
          class: 'zone-level',
          'text-anchor': 'middle',
        }, labels);
        lvl.textContent = z.levelRange;
      }
    }
  }

  function renderCities() {
    const cities = state.data.cities || [];
    let g = document.getElementById('layer-cities');
    if (!g) {
      g = el('g', { id: 'layer-cities' }, continentRoot);
    }
    g.innerHTML = '';
    const sizeFor = (imp) => {
      if (imp === 'capital') return 56;
      if (imp === 'hub') return 46;
      return 34;
    };
    for (const c of cities) {
      if (c.x == null || c.y == null) continue;
      const imp = c.importance || 'town';
      const size = sizeFor(imp);
      const fog = c.visibility === 'fog' || c.visibility === 'stub';
      const wrap = el('g', {
        class: `city-marker city-${imp}${fog ? ' city-fog' : ''}`,
        'data-id': c.id,
        'data-importance': imp,
        'data-x': c.x,
        'data-y': c.y,
        transform: `translate(${c.x} ${c.y})`,
      }, g);
      el('circle', {
        class: 'city-badge',
        r: size / 2 + 3,
        fill: 'rgba(6,8,12,0.72)',
        stroke: fog ? 'rgba(232,168,73,0.35)' : 'rgba(61,220,132,0.45)',
        'stroke-width': 2,
      }, wrap);
      el('image', {
        class: 'poi-icon',
        href: withV('/map/assets/icons/town.png'),
        x: -size / 2,
        y: -size / 2,
        width: size,
        height: size,
        opacity: fog ? '0.7' : '1',
      }, wrap);
      const showLabel = c.id === 'anvil-rest' || c.id === 'fenwatch';
      if (showLabel) {
        const label = el('text', {
          class: `city-label city-label-${imp}${fog ? ' fog-label' : ''}`,
          x: 0,
          y: size / 2 + 14,
          'text-anchor': 'middle',
        }, wrap);
        label.textContent = c.name || c.id;
      }
      wrap.addEventListener('click', (ev) => {
        ev.stopPropagation();
        if (state.moved) return;
        openPanel('city', c);
        const z = zoneById(c.zoneId);
        if (z && z.zoneMap) enterZone(z);
      });
    }
  }

  function tryWorldPlate() {
    const layers = state.data.world?.layers || {};
    const candidates = [
      withV(absMapAsset(layers.worldPlate)),
      withV('/map/assets/world/world-plate.jpg'),
      withV(absMapAsset(layers.worldPlateAlt)),
    ].filter((u, i, a) => u && a.indexOf(u) === i);
    const node = document.getElementById('world-plate');
    const parchment = document.getElementById('parchment');
    svg.classList.add('has-world-plate');
    const tryNext = (i) => {
      if (i >= candidates.length) {
        node.setAttribute('opacity', '0');
        parchment.setAttribute('opacity', '1');
        svg.classList.remove('has-world-plate');
        return;
      }
      const href = candidates[i];
      const probe = new Image();
      probe.onload = () => {
        node.setAttribute('href', href);
        node.setAttribute('opacity', '1');
        parchment.setAttribute('opacity', '0.08');
        parchment.setAttribute('fill', '#0a0e14');
        svg.classList.add('has-world-plate');
      };
      probe.onerror = () => tryNext(i + 1);
      probe.src = href;
    };
    tryNext(0);
  }

  function bindInput() {
    viewport.addEventListener('wheel', (e) => {
      e.preventDefault();
      const factor = e.deltaY > 0 ? 0.9 : 1.1;
      zoomAt(e.clientX, e.clientY, factor);
    }, { passive: false });

    viewport.addEventListener('pointerdown', (e) => {
      viewport.setPointerCapture(e.pointerId);
      state.pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
      state.moved = false;
      state.downX = e.clientX;
      state.downY = e.clientY;
      if (state.pointers.size === 1) {
        state.dragging = true;
        state.lastX = e.clientX;
        state.lastY = e.clientY;
        viewport.classList.add('dragging');
      } else if (state.pointers.size === 2) {
        const pts = [...state.pointers.values()];
        state.pinchStartDist = Math.hypot(pts[0].x - pts[1].x, pts[0].y - pts[1].y);
        state.pinchStartScale = state.scale;
      }
    });

    viewport.addEventListener('pointermove', (e) => {
      if (!state.pointers.has(e.pointerId)) return;
      state.pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
      if (Math.hypot(e.clientX - state.downX, e.clientY - state.downY) > 8) {
        state.moved = true;
      }
      if (state.pointers.size === 2) {
        const pts = [...state.pointers.values()];
        const dist = Math.hypot(pts[0].x - pts[1].x, pts[0].y - pts[1].y);
        if (state.pinchStartDist > 0) {
          const midX = (pts[0].x + pts[1].x) / 2;
          const midY = (pts[0].y + pts[1].y) / 2;
          const target = state.pinchStartScale * (dist / state.pinchStartDist);
          const factor = target / state.scale;
          zoomAt(midX, midY, factor);
        }
        return;
      }
      if (!state.dragging) return;
      const dx = e.clientX - state.lastX;
      const dy = e.clientY - state.lastY;
      state.lastX = e.clientX;
      state.lastY = e.clientY;
      state.tx += dx;
      state.ty += dy;
      applyCamera();
    });

    const endPointer = (e) => {
      state.pointers.delete(e.pointerId);
      if (state.pointers.size < 2) state.pinchStartDist = 0;
      if (state.pointers.size === 0) {
        state.dragging = false;
        viewport.classList.remove('dragging');
      }
    };
    viewport.addEventListener('pointerup', endPointer);
    viewport.addEventListener('pointercancel', endPointer);

    document.getElementById('panel-close').addEventListener('click', closePanel);
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        if (panel.classList.contains('open')) closePanel();
        else if (state.view === 'zone') exitZone();
      }
    });

    document.querySelectorAll('[data-lod]').forEach((btn) => {
      btn.addEventListener('click', () => {
        document.querySelectorAll('[data-lod]').forEach((b) => b.classList.remove('active'));
        btn.classList.add('active');
        state.lodMode = btn.getAttribute('data-lod');
        updateLodClass();
      });
    });
    document.getElementById('btn-zoom-in').addEventListener('click', () => {
      const r = viewport.getBoundingClientRect();
      zoomAt(r.left + r.width / 2, r.top + r.height / 2, 1.2);
    });
    document.getElementById('btn-zoom-out').addEventListener('click', () => {
      const r = viewport.getBoundingClientRect();
      zoomAt(r.left + r.width / 2, r.top + r.height / 2, 1 / 1.2);
    });
    document.getElementById('btn-reset').addEventListener('click', () => {
      if (state.view === 'zone') fitView();
      else fitView();
    });
    btnBack.addEventListener('click', () => exitZone());
    window.addEventListener('resize', fitView);
    window.addEventListener('hashchange', () => applyHash());
  }

  function applyHash() {
    const raw = (location.hash || '').replace('#', '').trim();
    if (!raw) {
      if (state.view === 'zone') exitZone({ skipHash: true });
      return;
    }
    const id = raw.toUpperCase();
    const z = zoneById(id) || (state.data.zones || []).find((x) => x.slug === raw.toLowerCase());
    if (z) enterZone(z, { skipHash: true });
  }

  async function boot() {
    const res = await fetch(`/map/map-data.json?v=${ASSET_V}`, { cache: 'no-cache' });
    if (!res.ok) throw new Error('Failed to load map-data.json');
    state.data = await res.json();
    for (const z of state.data.zones || []) {
      if (z && 'internal' in z) delete z.internal;
    }
    chapterLabel.textContent = state.data.chapter || 'World Map';
    const vb = state.data.world.viewBox;
    const parchment = document.getElementById('parchment');
    parchment.setAttribute('width', vb[2]);
    parchment.setAttribute('height', vb[3]);
    const plate = document.getElementById('world-plate');
    plate.setAttribute('width', vb[2]);
    plate.setAttribute('height', vb[3]);

    renderZones();
    renderCities();
    tryWorldPlate();
    bindInput();
    fitView();
    if (location.hash) applyHash();
  }

  boot().catch((err) => {
    console.error(err);
    panelBody.textContent = 'Map data failed to load.';
    panel.classList.add('open');
  });
})();
