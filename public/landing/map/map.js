/**
 * Veilspan public world map — data-driven SVG pan/zoom/LOD.
 * PUBLIC fields only (map-data.json is built with internal lore stripped).
 * Imagine hooks: world plate + zone plates under assets/.
 */
(() => {
  const NS = 'http://www.w3.org/2000/svg';
  const viewport = document.getElementById('map-viewport');
  const svg = document.getElementById('map-svg');
  const camera = document.getElementById('camera');
  const panel = document.getElementById('detail-panel');
  const panelBody = document.getElementById('panel-body');
  const panelTitle = document.getElementById('panel-title');
  const panelMeta = document.getElementById('panel-meta');
  const panelEyebrow = document.getElementById('panel-eyebrow');

  const state = {
    data: null,
    scale: 0.7,
    tx: 0,
    ty: 0,
    minScale: 0.28,
    maxScale: 3.6,
    lodMode: 'auto', // auto | continent | zones | pois
    dragging: false,
    lastX: 0,
    lastY: 0,
    pointers: new Map(),
    pinchStartDist: 0,
    pinchStartScale: 1,
    selectedId: null,
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
  };

  function el(name, attrs = {}, parent) {
    const node = document.createElementNS(NS, name);
    for (const [k, v] of Object.entries(attrs)) {
      if (v != null) node.setAttribute(k, String(v));
    }
    if (parent) parent.appendChild(node);
    return node;
  }

  function applyCamera() {
    camera.setAttribute('transform', `translate(${state.tx} ${state.ty}) scale(${state.scale})`);
    updateLodClass();
  }

  function currentLod() {
    if (state.lodMode !== 'auto') return state.lodMode;
    const lod = state.data?.lod || {};
    if (state.scale < (lod.continent?.maxScale ?? 0.55)) return 'continent';
    if (state.scale < (lod.pois?.minScale ?? 1.35)) return 'zones';
    return 'pois';
  }

  function updateLodClass() {
    const lod = currentLod();
    svg.classList.remove('lod-continent', 'lod-zones', 'lod-pois');
    svg.classList.add(`lod-${lod}`);
  }

  function fitView() {
    const vb = state.data.world.viewBox;
    const [, , w, h] = vb;
    const rect = viewport.getBoundingClientRect();
    const sx = rect.width / w;
    const sy = rect.height / h;
    state.scale = Math.min(sx, sy) * 0.92;
    state.tx = (rect.width - w * state.scale) / 2;
    state.ty = (rect.height - h * state.scale) / 2;
    applyCamera();
  }

  function zoomAt(clientX, clientY, factor) {
    const rect = viewport.getBoundingClientRect();
    const x = clientX - rect.left;
    const y = clientY - rect.top;
    const worldX = (x - state.tx) / state.scale;
    const worldY = (y - state.ty) / state.scale;
    const next = Math.min(state.maxScale, Math.max(state.minScale, state.scale * factor));
    state.scale = next;
    state.tx = x - worldX * state.scale;
    state.ty = y - worldY * state.scale;
    applyCamera();
  }

  function zoneById(id) {
    return state.data.zones.find((z) => z.id === id);
  }

  function openPanel(kind, payload) {
    panel.classList.add('open');
    panel.setAttribute('aria-hidden', 'false');
    if (kind === 'city') {
      const c = payload;
      state.selectedId = c.id;
      document.querySelectorAll('.zone-shape.selected').forEach((n) => n.classList.remove('selected'));
      const imp = c.importance || 'town';
      panelEyebrow.textContent = (c.visibility === 'fog' ? 'Uncharted · ' : '') + (imp === 'capital' ? 'Capital' : imp === 'hub' ? 'Hub' : 'Town');
      panelTitle.textContent = c.name || c.id;
      panelMeta.textContent = [c.faction, c.zoneId].filter(Boolean).join(' · ');
      let html = '';
      if (c.visibility === 'fog') {
        html += `<div class="fog-banner">Fog of war — charts incomplete. Rumors only.</div>`;
      }
      if (c.blurb) html += `<p>${c.blurb}</p>`;
      panelBody.innerHTML = html || '<p>No public charts yet.</p>';
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
      // Imagine hook note (dev-facing, subtle)
      if (z.imagePlate) {
        html += `<p style="margin-top:1.2rem;font-size:0.8rem;color:var(--text-dim)">Art slot: <code>${escapeHtml(z.imagePlate)}</code></p>`;
      }
      panelBody.innerHTML = html;
    } else if (kind === 'poi') {
      const p = payload;
      state.selectedId = p.id;
      panelEyebrow.textContent = `POI · ${p.kind || 'landmark'}`;
      panelTitle.textContent = p.name;
      panelMeta.textContent = p.zoneTitle ? `in ${p.zoneTitle}` : '';
      panelBody.innerHTML = `<p>${escapeHtml(p.blurb || 'A notable place on the charts.')}</p>`;
    }
  }

  function closePanel() {
    panel.classList.remove('open');
    panel.setAttribute('aria-hidden', 'true');
    document.querySelectorAll('.zone-shape.selected').forEach((n) => n.classList.remove('selected'));
    state.selectedId = null;
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function renderTerrain() {
    const g = document.getElementById('layer-terrain');
    g.innerHTML = '';
    // Soft region washes behind live zones
    for (const z of state.data.zones) {
      if (z.fog) continue;
      const { cx, cy, rx, ry } = z.shape;
      el('ellipse', {
        cx, cy, rx: rx * 1.8, ry: ry * 1.8,
        fill: z.fill, opacity: 0.18,
      }, g);
    }
  }

  function renderConnections() {
    const g = document.getElementById('layer-connections');
    g.innerHTML = '';
    for (const e of state.data.connections || []) {
      const a = zoneById(e.from);
      const b = zoneById(e.to);
      if (!a || !b) continue;
      el('line', {
        x1: a.shape.cx, y1: a.shape.cy,
        x2: b.shape.cx, y2: b.shape.cy,
        class: `connection ${e.state || 'open'}`,
      }, g);
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
        fill: z.fog ? 'url(#fogPattern)' : z.fill,
        'data-id': z.id,
      }, g);
      if (z.fog) shape.setAttribute('fill', '#6a6055');
      shape.addEventListener('click', (ev) => {
        ev.stopPropagation();
        openPanel('zone', z);
      });
      shape.addEventListener('mouseenter', () => {
        if (window.matchMedia('(hover:hover)').matches && !panel.classList.contains('open')) {
          // lightweight hover: title in eyebrow via title attr
        }
      });

      // Try zone plate image (Imagine hook) — mid-zoom+ (layer-zones-detail)
      if (!z.fog && z.imagePlate) {
        const plateHref = absMapAsset(z.imagePlate) + '?v=map4';
        const img = el('image', {
          class: 'zone-plate layer-zones-detail',
          href: plateHref,
          x: cx - rx, y: cy - ry, width: rx * 2, height: ry * 2,
          opacity: 0, preserveAspectRatio: 'xMidYMid slice',
          style: 'pointer-events:none',
        }, g);
        const probe = new Image();
        probe.onload = () => { img.setAttribute('opacity', '0.55'); };
        probe.onerror = () => { img.remove(); };
        probe.src = plateHref;
      }

      const label = el('text', {
        x: cx, y: cy + 5,
        class: `zone-label${z.fog ? ' fog-label' : ''}`,
        'text-anchor': 'middle',
      }, labels);
      label.textContent = z.fog ? (z.title || '???') : (z.title || z.id);

      if (!z.fog && z.levelRange) {
        const lvl = el('text', {
          x: cx, y: cy + 22,
          class: 'zone-level',
          'text-anchor': 'middle',
        }, labels);
        lvl.textContent = z.levelRange;
      }
    }
  }

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

  function poiIconHref(kind) {
    const key = poiIconMap[kind];
    return key ? `/map/assets/icons/${key}.png?v=map4` : null;
  }

  function renderPois() {
    const g = document.getElementById('layer-pois');
    g.innerHTML = '';
    for (const z of state.data.zones) {
      if (z.fog) continue;
      for (const p of z.pois || []) {
        const color = poiColors[p.kind] || '#f0c674';
        const iconHref = poiIconHref(p.kind);
        let marker;
        if (iconHref) {
          const size = 28;
          marker = el('image', {
            class: 'poi-icon',
            href: iconHref,
            x: p.x - size / 2,
            y: p.y - size / 2,
            width: size,
            height: size,
            'data-id': p.id,
          }, g);
        } else {
          marker = el('circle', {
            class: 'poi-dot',
            cx: p.x, cy: p.y, r: 5.5,
            fill: color,
            'data-id': p.id,
          }, g);
        }
        marker.addEventListener('click', (ev) => {
          ev.stopPropagation();
          openPanel('poi', { ...p, zoneTitle: z.title });
        });
        const t = el('text', {
          class: 'poi-label',
          x: p.x + 14, y: p.y + 4,
        }, g);
        t.textContent = p.name;
      }
    }
  }

  /** Capitals ≫ hub (Oldtown) ≫ towns. LOD: capitals always; hub mid+; towns in close. */
  function renderCities() {
    const cities = state.data.cities || [];
    if (!Array.isArray(cities) || !cities.length) return;
    let g = document.getElementById('layer-cities');
    if (!g) {
      const root = document.getElementById('world-root');
      // Insert above zones so markers read clearly, before POIs
      g = el('g', { id: 'layer-cities' }, null);
      const pois = document.getElementById('layer-pois');
      if (pois && pois.parentNode) pois.parentNode.insertBefore(g, pois);
      else root.appendChild(g);
    }
    g.innerHTML = '';
    const sizeFor = (imp) => {
      if (imp === 'capital') return 42;
      if (imp === 'hub') return 34;
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
      }, g);
      const href = fog
        ? `/map/assets/icons/town.png?v=map4`
        : `/map/assets/icons/town.png?v=map4`;
      const img = el('image', {
        class: 'poi-icon',
        href,
        x: c.x - size / 2,
        y: c.y - size / 2,
        width: size,
        height: size,
        opacity: fog ? '0.55' : '0.95',
      }, wrap);
      img.addEventListener('click', (ev) => {
        ev.stopPropagation();
        openPanel('city', c);
      });
      if (!c.optionalLabel || imp !== 'town') {
        const label = el('text', {
          class: `city-label city-label-${imp}${fog ? ' fog-label' : ''}`,
          x: c.x,
          y: c.y + size / 2 + (imp === 'capital' ? 16 : 12),
          'text-anchor': 'middle',
        }, wrap);
        label.textContent = fog && imp === 'capital' ? c.name : (c.name || c.id || '');
      }
    }
  }

  function absMapAsset(href) {
    if (!href) return href;
    if (href.startsWith('http') || href.startsWith('/')) return href;
    return `/map/${href.replace(/^\.\//, '')}`;
  }

  function tryWorldPlate() {
    const layers = state.data.world?.layers || {};
    const candidates = [
      absMapAsset(layers.worldPlate) + (layers.worldPlate ? '?v=map4' : ''),
      '/map/assets/world/world-plate.jpg?v=map4',
      absMapAsset(layers.worldPlateAlt) + (layers.worldPlateAlt ? '?v=map4' : ''),
      '/map/assets/world/world-plate-16x9.jpg?v=map4',
    ].filter((u, i, a) => u && a.indexOf(u) === i);
    const node = document.getElementById('world-plate');
    const parchment = document.getElementById('parchment');
    // Plate is already in HTML; mark styled until proven broken
    svg.classList.add('has-world-plate');
    const tryNext = (i) => {
      if (i >= candidates.length) {
        node.setAttribute('opacity', '0');
        parchment.setAttribute('opacity', '1');
        parchment.setAttribute('fill', '#0a0e14');
        svg.classList.remove('has-world-plate');
        return;
      }
      const href = candidates[i];
      const probe = new Image();
      probe.onload = () => {
        node.setAttribute('href', href);
        node.setAttribute('opacity', '0.92');
        parchment.setAttribute('opacity', '0.35');
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
      if (state.pointers.size < 2) {
        state.pinchStartDist = 0;
      }
      if (state.pointers.size === 0) {
        state.dragging = false;
        viewport.classList.remove('dragging');
      }
    };
    viewport.addEventListener('pointerup', endPointer);
    viewport.addEventListener('pointercancel', endPointer);

    viewport.addEventListener('click', () => {
      // click empty parchment closes on desktop when not dragging far
    });

    document.getElementById('panel-close').addEventListener('click', closePanel);
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') closePanel();
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
    document.getElementById('btn-reset').addEventListener('click', fitView);
    window.addEventListener('resize', fitView);
  }

  async function boot() {
    const res = await fetch('/map/map-data.json?v=map4', { cache: 'no-cache' });
    if (!res.ok) throw new Error('Failed to load map-data.json');
    state.data = await res.json();
    // Defense in depth: strip internal if someone ever ships raw lore by mistake
    for (const z of state.data.zones || []) {
      if (z && 'internal' in z) delete z.internal;
    }
    document.getElementById('chapter-label').textContent = state.data.chapter || 'World Map';
    const vb = state.data.world.viewBox;
    svg.setAttribute('viewBox', `0 0 ${vb[2]} ${vb[3]}`);
    document.getElementById('parchment').setAttribute('width', vb[2]);
    document.getElementById('parchment').setAttribute('height', vb[3]);
    document.getElementById('world-plate').setAttribute('width', vb[2]);
    document.getElementById('world-plate').setAttribute('height', vb[3]);

    renderTerrain();
    renderConnections();
    renderZones();
    renderPois();
    renderCities();
    tryWorldPlate();
    bindInput();
    fitView();
    // Start slightly zoomed into the Hearthlands (Z02)
    const hub = zoneById('Z02');
    if (hub) {
      const rect = viewport.getBoundingClientRect();
      state.scale = Math.max(state.scale, 0.85);
      state.tx = rect.width / 2 - hub.shape.cx * state.scale;
      state.ty = rect.height / 2 - hub.shape.cy * state.scale;
      applyCamera();
    }
  }

  boot().catch((err) => {
    console.error(err);
    panelBody.textContent = 'Map data failed to load.';
    panel.classList.add('open');
  });
})();
