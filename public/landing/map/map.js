/**
 * Veilspan public world map — WoW-style continent → zone drill-down.
 * PUBLIC fields only (map-data.json is built with internal lore stripped).
 */
(() => {
  const NS = 'http://www.w3.org/2000/svg';
  const ASSET_V = 'map16';
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
    hoverCityId: null,
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
    const ck = Math.max(1.15, Math.min(2.05, 0.92 / state.scale));
    document.querySelectorAll('#layer-chips .map-chip').forEach((n) => {
      const x = Number(n.getAttribute('data-x'));
      const y = Number(n.getAttribute('data-y'));
      n.setAttribute('transform', `translate(${x} ${y}) scale(${ck})`);
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

  function screenToWorld(clientX, clientY) {
    const rect = viewport.getBoundingClientRect();
    return {
      x: (clientX - rect.left - state.tx) / state.scale,
      y: (clientY - rect.top - state.ty) / state.scale,
    };
  }

  function pointInEllipse(px, py, cx, cy, rx, ry) {
    if (!rx || !ry) return false;
    const dx = (px - cx) / rx;
    const dy = (py - cy) / ry;
    return dx * dx + dy * dy <= 1;
  }

  function zoneAtWorld(wx, wy) {
    const zones = state.data?.zones || [];
    let best = null;
    let bestArea = Infinity;
    for (const z of zones) {
      const { cx, cy, rx, ry } = z.shape || {};
      if (pointInEllipse(wx, wy, cx, cy, rx, ry)) {
        const area = rx * ry;
        if (area < bestArea) {
          best = z;
          bestArea = area;
        }
      }
    }
    return best;
  }

  function cityAtWorld(wx, wy) {
    const cities = state.data?.cities || [];
    const hit = 28 / Math.max(state.scale, 0.25);
    let best = null;
    let bestD = hit * hit;
    for (const c of cities) {
      if (c.x == null || c.y == null) continue;
      const d = (c.x - wx) ** 2 + (c.y - wy) ** 2;
      if (d < bestD) {
        best = c;
        bestD = d;
      }
    }
    return best;
  }

  function layoutChip(g) {
    const bg = g.querySelector('.map-chip-bg');
    const copy = g.querySelector('.map-chip-copy');
    if (!bg || !copy) return;
    let bb;
    try { bb = copy.getBBox(); } catch (_) { return; }
    if (!bb.width) return;
    bg.setAttribute('x', bb.x - 10);
    bg.setAttribute('y', bb.y - 5);
    bg.setAttribute('width', bb.width + 20);
    bg.setAttribute('height', bb.height + 10);
  }

  function makeChip(parent, x, y, title, meta, extraClass) {
    const g = el('g', {
      class: `map-chip${extraClass ? ` ${extraClass}` : ''}`,
      'data-x': x,
      'data-y': y,
      transform: `translate(${x} ${y})`,
    }, parent);
    el('rect', { class: 'map-chip-bg', x: -40, y: -14, width: 80, height: 28, rx: 5, ry: 5 }, g);
    const copy = el('g', { class: 'map-chip-copy' }, g);
    const t = el('text', { class: 'map-chip-title', x: 0, y: meta ? -2 : 4 }, copy);
    t.textContent = title;
    if (meta) {
      const m = el('text', { class: 'map-chip-meta', x: 0, y: 12 }, copy);
      m.textContent = meta;
    }
    requestAnimationFrame(() => layoutChip(g));
    return g;
  }

  function setHoverVisual(kind, payload, wx, wy) {
    const clip = document.getElementById('hoverClipEllipse');
    const light = document.getElementById('hover-light');
    const sat = document.getElementById('layer-sat');
    const hi = document.getElementById('layer-highlight');
    const chips = document.getElementById('layer-chips');
    if (!clip || !light) return;

    document.querySelectorAll('.map-chip.is-hover').forEach((n) => n.classList.remove('is-hover'));

    if (!kind || state.view !== 'continent') {
      state.hoverZoneId = null;
      state.hoverCityId = null;
      hi.classList.remove('is-on');
      sat.classList.remove('is-on');
      light.setAttribute('opacity', '0');
      return;
    }

    let cx, cy, rx, ry;
    if (kind === 'city') {
      state.hoverCityId = payload.id;
      state.hoverZoneId = payload.zoneId || null;
      const zone = payload.zoneId ? zoneById(payload.zoneId) : null;
      if (zone?.shape) {
        ({ cx, cy, rx, ry } = zone.shape);
      } else {
        cx = payload.x;
        cy = payload.y;
        rx = ry = 90;
      }
      const chip = chips?.querySelector(`.map-chip[data-city="${payload.id}"]`);
      if (chip) chip.classList.add('is-hover');
    } else {
      state.hoverZoneId = payload.id;
      state.hoverCityId = null;
      ({ cx, cy, rx, ry } = payload.shape);
      const chip = chips?.querySelector(`.map-chip[data-zone="${payload.id}"]`);
      if (chip) chip.classList.add('is-hover');
    }

    clip.setAttribute('cx', cx);
    clip.setAttribute('cy', cy);
    clip.setAttribute('rx', rx);
    clip.setAttribute('ry', ry);
    const radius = Math.max(rx, ry) * 1.35;
    light.setAttribute('cx', wx);
    light.setAttribute('cy', wy);
    light.setAttribute('rx', radius);
    light.setAttribute('ry', radius);
    light.setAttribute('opacity', '1');
    hi.classList.add('is-on');
    sat.classList.add('is-on');
  }

  function updateHoverHighlight(clientX, clientY) {
    if (state.view !== 'continent' || state.dragging) {
      setHoverVisual(null);
      return;
    }
    const { x: wx, y: wy } = screenToWorld(clientX, clientY);
    const city = cityAtWorld(wx, wy);
    if (city) {
      setHoverVisual('city', city, wx, wy);
      return;
    }
    const zone = zoneAtWorld(wx, wy);
    if (zone) {
      setHoverVisual('zone', zone, wx, wy);
      return;
    }
    setHoverVisual(null);
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
      document.querySelectorAll('.map-chip.is-selected').forEach((n) => n.classList.remove('is-selected'));
      const cityChip = document.querySelector(`.map-chip[data-city="${c.id}"]`);
      if (cityChip) cityChip.classList.add('is-selected');
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
      document.querySelectorAll('.map-chip.is-selected').forEach((n) => n.classList.remove('is-selected'));
      const shape = document.getElementById(`zone-${z.id}`);
      if (shape) shape.classList.add('selected');
      const chip = document.querySelector(`.map-chip[data-zone="${z.id}"]`);
      if (chip) chip.classList.add('is-selected');

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
          html += `<li data-poi="${escapeHtml(p.id)}"><strong>${escapeHtml(p.name)}</strong> <span class="state">(${escapeHtml(p.kind)})</span>`;
          if (p.blurb) html += `<div>${escapeHtml(p.blurb)}</div>`;
          html += `</li>`;
        }
        html += `</ul>`;
      }
      if (z.zoneMap) {
        const label = state.view === 'zone' && state.zoneId === z.id ? 'Zoom zone map' : 'Enter zone map';
        html += `<p><button type="button" class="chip" id="btn-enter-zone" data-zone="${escapeHtml(z.id)}">${label}</button></p>`;
      }
      panelBody.innerHTML = html;
      bindEnterBtn();
      bindPanelPois(z);
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
      ev.preventDefault();
      ev.stopPropagation();
      const z = zoneById(btn.getAttribute('data-zone'));
      if (z) enterZone(z);
    });
  }

  function bindPanelPois(z) {
    panelBody.querySelectorAll('[data-poi]').forEach((row) => {
      row.addEventListener('click', (ev) => {
        ev.preventDefault();
        ev.stopPropagation();
        const p = (z.pois || []).find((x) => x.id === row.getAttribute('data-poi'));
        if (!p) return;
        if (state.view !== 'zone' || state.zoneId !== z.id) enterZone(z);
        openPanel('poi', { ...p, zoneTitle: z.title });
      });
    });
  }

  function closestFromPoint(clientX, clientY, selector) {
    const nodes = document.elementsFromPoint(clientX, clientY);
    for (const n of nodes) {
      if (!n || typeof n.closest !== 'function') continue;
      if (n.closest('#detail-panel, .nav, .map-chrome')) return null;
      const hit = n.closest(selector);
      if (hit) return hit;
    }
    return null;
  }

  function handleMapTap(clientX, clientY) {
    if (state.view === 'zone') {
      const poiEl = closestFromPoint(clientX, clientY, '.poi-marker');
      if (poiEl) {
        const z = zoneById(state.zoneId);
        const p = (z?.pois || []).find((x) => x.id === poiEl.getAttribute('data-id'));
        if (p) {
          openPanel('poi', { ...p, zoneTitle: z.title });
          return;
        }
      }
      const z = zoneById(state.zoneId);
      if (z) openPanel('zone', z);
      return;
    }
    const cityEl = closestFromPoint(clientX, clientY, '.city-marker');
    if (cityEl) {
      const c = cityById(cityEl.getAttribute('data-id'));
      if (c) {
        openPanel('city', c);
        return;
      }
    }
    const zoneEl = closestFromPoint(clientX, clientY, '.zone-shape');
    if (zoneEl) {
      const z = zoneById(zoneEl.getAttribute('data-id'));
      if (z) openPanel('zone', z);
    }
  }

  function handleMapDblClick(clientX, clientY) {
    if (state.view === 'zone') return;
    const cityEl = closestFromPoint(clientX, clientY, '.city-marker');
    if (cityEl) {
      const c = cityById(cityEl.getAttribute('data-id'));
      const z = c ? zoneById(c.zoneId) : null;
      if (z) enterZone(z);
      return;
    }
    const zoneEl = closestFromPoint(clientX, clientY, '.zone-shape');
    if (zoneEl) {
      const z = zoneById(zoneEl.getAttribute('data-id'));
      if (z) enterZone(z);
    }
  }

  function closePanel() {
    panel.classList.remove('open');
    panel.setAttribute('aria-hidden', 'true');
    document.querySelectorAll('.zone-shape.selected').forEach((n) => n.classList.remove('selected'));
    document.querySelectorAll('.map-chip.is-selected').forEach((n) => n.classList.remove('is-selected'));
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
        const size = 44;
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
      const chip = makeChip(wrap, 0, 32, p.name, null, 'poi-chip');
      chip.classList.add('is-hover');
      wrap.addEventListener('click', (ev) => {
        ev.stopPropagation();
        ev.preventDefault();
      });
    }
  }

  function renderZones() {
    const g = document.getElementById('layer-zones');
    const fogG = document.getElementById('layer-fog');
    const chips = document.getElementById('layer-chips');
    g.innerHTML = '';
    if (fogG) fogG.innerHTML = '';
    if (chips) chips.innerHTML = '';

    for (const z of state.data.zones) {
      const { cx, cy, rx, ry } = z.shape;
      const vis = z.visibility || (z.fog ? 'fog' : 'live');
      if (z.fog && fogG) {
        el('ellipse', {
          class: 'fog-wash',
          cx, cy, rx: rx * 1.05, ry: ry * 1.05,
        }, fogG);
      }
      const shape = el('ellipse', {
        id: `zone-${z.id}`,
        class: `zone-shape ${vis}`,
        cx, cy, rx, ry,
        'data-id': z.id,
      }, g);
      shape.addEventListener('click', (ev) => {
        ev.stopPropagation();
        ev.preventDefault();
      });

      const short = {
        Z00: 'Catacombs', Z01: 'Meadows', Z02: 'Oldtown', Z03: 'Gloomfen',
        Z04: 'Ashenveil', Z05: 'Silverbrook', Z06: 'Ironspine', Z07: 'Verdant Reach',
        Z08: 'Gearwind', Z09: 'Thornfield', Z10: 'Kazgrath', Z11: 'Aelindor',
        Z12: 'Veridane', Z19: 'Depths',
      };
      const meta = z.fog ? 'Uncharted' : (z.levelRange ? `Lv ${z.levelRange}` : '');
      const chip = makeChip(
        chips,
        cx,
        cy + Math.min(ry * 0.22, 22),
        short[z.id] || z.title || z.id,
        meta,
        z.fog ? 'fog-chip' : ''
      );
      chip.setAttribute('data-zone', z.id);
    }
  }

  function renderCities() {
    const cities = state.data.cities || [];
    let g = document.getElementById('layer-cities');
    if (!g) {
      g = el('g', { id: 'layer-cities' }, continentRoot);
    }
    g.innerHTML = '';
    const chips = document.getElementById('layer-chips');
    const sizeFor = (imp) => {
      if (imp === 'capital') return 36;
      if (imp === 'hub') return 30;
      return 24;
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
        class: 'city-hit',
        r: size,
      }, wrap);
      wrap.addEventListener('click', (ev) => {
        ev.stopPropagation();
        ev.preventDefault();
      });
      if (chips) {
        const chip = makeChip(
          chips,
          c.x,
          c.y + 26,
          c.name || c.id,
          c.faction || (imp === 'capital' ? 'Capital' : ''),
          'city-chip'
        );
        chip.setAttribute('data-city', c.id);
      }
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
      if (e.button != null && e.button !== 0) return;
      state.pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
      state.moved = false;
      state.downX = e.clientX;
      state.downY = e.clientY;
      state.lastX = e.clientX;
      state.lastY = e.clientY;
      if (state.pointers.size === 2) {
        const pts = [...state.pointers.values()];
        state.pinchStartDist = Math.hypot(pts[0].x - pts[1].x, pts[0].y - pts[1].y);
        state.pinchStartScale = state.scale;
      }
    });

    viewport.addEventListener('pointermove', (e) => {
      if (!state.dragging) updateHoverHighlight(e.clientX, e.clientY);
      if (!state.pointers.has(e.pointerId)) return;
      state.pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
      const dist = Math.hypot(e.clientX - state.downX, e.clientY - state.downY);
      if (dist > 8 && !state.dragging) {
        state.moved = true;
        state.dragging = true;
        viewport.classList.add('dragging');
        try { viewport.setPointerCapture(e.pointerId); } catch (_) { /* ignore */ }
      }
      if (state.pointers.size === 2) {
        const pts = [...state.pointers.values()];
        const distPinch = Math.hypot(pts[0].x - pts[1].x, pts[0].y - pts[1].y);
        if (state.pinchStartDist > 0) {
          const midX = (pts[0].x + pts[1].x) / 2;
          const midY = (pts[0].y + pts[1].y) / 2;
          const target = state.pinchStartScale * (distPinch / state.pinchStartDist);
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
      const wasDrag = state.moved;
      const x = e.clientX;
      const y = e.clientY;
      state.pointers.delete(e.pointerId);
      if (state.pointers.size < 2) state.pinchStartDist = 0;
      if (state.pointers.size === 0) {
        state.dragging = false;
        viewport.classList.remove('dragging');
        if (!wasDrag) handleMapTap(x, y);
      }
    };
    viewport.addEventListener('pointerup', endPointer);
    viewport.addEventListener('pointercancel', endPointer);
    viewport.addEventListener('pointerleave', () => {
      if (!state.dragging) updateHoverHighlight(-1, -1);
    });
    viewport.addEventListener('dblclick', (e) => {
      e.preventDefault();
      handleMapDblClick(e.clientX, e.clientY);
    });

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
    let lastVW = 0;
    let lastVH = 0;
    const onViewportResize = () => {
      const { w, h } = viewportSize();
      if (Math.abs(w - lastVW) < 2 && Math.abs(h - lastVH) < 2) return;
      lastVW = w;
      lastVH = h;
      fitView();
    };
    window.addEventListener('resize', onViewportResize);
    if (window.ResizeObserver) {
      new ResizeObserver(onViewportResize).observe(viewport);
    }
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
    const relayout = () => document.querySelectorAll('.map-chip').forEach(layoutChip);
    if (document.fonts && document.fonts.ready) document.fonts.ready.then(relayout);
    setTimeout(relayout, 250);
    if (location.hash) applyHash();
  }

  boot().catch((err) => {
    console.error(err);
    panelBody.textContent = 'Map data failed to load.';
    panel.classList.add('open');
  });
})();
