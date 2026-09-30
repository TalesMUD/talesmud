<script>
  import { onDestroy, onMount, tick } from 'svelte';
  import { readStageSize, shouldRepaintSize, applyCanvasBitmap } from '../widgets/atlasLayout.js';
  import { paintAtlas, isCurrentPlace, panToCenterPlace, onMapTilesReady, clampMapScale, setYouPortrait, collectTurnInMarkers } from '../widgets/atlasRenderer.js';
  import { prefersReducedMotion } from '../keyboardShortcuts.js';
  import { interiorChoices, surfaceId, surfaceGroups, groupForRoom } from '../widgets/surfaceAtlas.js';
  import { mobileStore } from '../mobile/mobileStore.js';

  const { isMobile } = mobileStore;

  export let store = null;
  export let sendMessage = null;

  let atlas = emptyAtlas();
  let currentRoomId = null;
  let activeLayer = '';

  let travelTargetId = null;
  let travelPath = [];
  let isTraveling = false;
  let travelPathRoomIds = new Set();
  let travelFollowActive = false;
  let overlayFading = false;
  let panAnimRaf = 0;
  let closeAfterTravelTimer = 0;

  let panX = 0;
  let panY = 0;
  let userScale = 1;
  let frameWorld = true;
  let stopTileListener;
  let ambientTimer;
  let isPanning = false;
  let didDrag = false;
  let panStart = { x: 0, y: 0, panX: 0, panY: 0 };
  let tooltip = { visible: false, text: '', x: 0, y: 0 };

  let stageWrap, stageCanvas, modalEl, mapBodyEl, intelEl;
  let selectedId = null;
  let lastTap = { id: null, at: 0 };
  let intelExpanded = false;
  let roomFilter = '';
  const pointers = new Map();
  let pinchStart = null;
  let stageObserver;
  const hitState = { items: [] };
  let lastStageSize = null;
  let drawRaf = 0;
  let escHandler = null;
  let wasOpen = false;
  let resizeHandler = null;

  $: open = !!(store && $store && $store.mapOverviewOpen);

  function isNarrow() {
    if (typeof window === 'undefined') return false;
    return window.innerWidth <= 768;
  }

  function viewportSize() {
    const vv = typeof window !== 'undefined' && window.visualViewport;
    return {
      w: vv ? Math.round(vv.width) : window.innerWidth,
      h: vv ? Math.round(vv.height) : window.innerHeight,
    };
  }

  function portalToBody(node) {
    applyOverlayChrome(node);
    if (node.parentNode !== document.body) {
      document.body.appendChild(node);
    }
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node);
      },
    };
  }

  function applyOverlayChrome(node) {
    if (!node) return;
    const narrow = isNarrow();
    const vp = viewportSize();
    node.style.cssText = [
      'position:fixed',
      'top:0',
      'left:0',
      'right:0',
      'bottom:0',
      narrow ? `width:${vp.w}px` : 'width:100vw',
      narrow ? `height:${vp.h}px` : 'height:100vh',
      'z-index:200000',
      'display:flex',
      narrow ? 'align-items:stretch' : 'align-items:center',
      narrow ? 'justify-content:stretch' : 'justify-content:center',
      narrow ? 'padding:0' : 'padding:12px',
      'box-sizing:border-box',
      'background:rgba(0,0,0,0.82)',
      'opacity:1',
      'visibility:visible',
      'pointer-events:auto',
    ].join(';');
  }

  function sizeModal() {
    if (!modalEl || typeof window === 'undefined') return { w: 0, h: 0 };
    const narrow = isNarrow();
    const vp = viewportSize();
    const w = narrow ? vp.w : Math.max(640, Math.round(vp.w * 0.8));
    const h = narrow ? vp.h : Math.max(460, Math.round(vp.h * 0.8));
    applyOverlayChrome(document.getElementById('map-overview-overlay'));
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
    modalEl.style.borderRadius = narrow ? '0' : '8px';
    if (mapBodyEl) {
      mapBodyEl.style.flexDirection = 'row';
      mapBodyEl.style.position = 'relative';
    }
    if (intelEl) {
      if (narrow) {
        intelEl.style.width = '100%';
        intelEl.style.flex = 'none';
        intelEl.style.maxHeight = '';
      } else {
        intelEl.style.width = '300px';
        intelEl.style.flex = '0 0 300px';
        intelEl.style.maxHeight = 'none';
        intelEl.style.position = '';
        intelEl.style.left = '';
        intelEl.style.right = '';
        intelEl.style.bottom = '';
      }
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
        if (travelFollowActive || isTraveling) {
          frameWorld = false;
          animatePanToYou({ durationMs: prefersReducedMotion() ? 0 : 280 });
        } else {
          applyRecenterToYou(true);
          scheduleDraw();
        }
      });
    } else if (open && atlasChanged) {
      scheduleDraw();
    }
  }

  $: if (open && !wasOpen) {
    wasOpen = true;
    userScale = 1;
    frameWorld = true;
    panX = panY = 0;
    intelExpanded = false;
    lastStageSize = null;
    paintAfterLayout();
  } else if (!open && wasOpen) {
    wasOpen = false;
    intelExpanded = false;
    scheduleDraw();
  }

  $: if (store && $store.character) setYouPortrait($store.character.portrait || '');

  $: visiblePlaces = (atlas.places || []).filter(p => p.layer === activeLayer);
  $: visibleRegions = (atlas.regions || []).filter(r => r.layer === activeLayer);
  $: turnInMarkers = collectTurnInMarkers(store && $store ? $store.quests : []);
  $: if (open && turnInMarkers) scheduleDraw();
  $: if (store && $store.mapSelectedId && $store.mapSelectedId !== selectedId) {
    selectedId = $store.mapSelectedId;
    if ($isMobile) intelExpanded = false;
  }
  $: selectedPlace = (atlas.places || []).find(p => p.id === selectedId) || null;
  $: townRooms = interiorChoices(atlas.places, selectedPlace);
  $: filteredTownRooms = townRooms.filter(p => String(p.name || p.id).toLowerCase().includes(roomFilter.toLowerCase()));
  $: selectedStreet = selectedPlace?.mapRole === 'interior' ? (atlas.places || []).find(p => p.id === surfaceId(selectedPlace) && p.discovered) : null;
  $: entranceRooms = selectedPlace?.discovered ? (atlas.places || []).filter(p => (selectedPlace.entrances || []).includes(p.id)) : [];
  $: canTravel = !!(selectedPlace && selectedPlace.discovered && selectedPlace.id !== currentRoomId);
  let lastNarrow = null;
  $: if (open && $isMobile !== lastNarrow) {
    lastNarrow = $isMobile;
    if (typeof window !== 'undefined' && modalEl) tick().then(() => paintAfterLayout());
  }
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

  function selectRoom(room, changeLayer = false) {
    selectedId = room.id;
    roomFilter = '';
    if (store && store.selectMapPlace) store.selectMapPlace(room.id);
    if (changeLayer && room.layer !== activeLayer) selectLayer(room.layer);
    scheduleDraw();
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

  function targetPanForYou() {
    const size = readStageSize(stageWrap);
    const here = resolveHerePlace();
    if (here && size.w >= 4 && size.h >= 4) {
      return panToCenterPlace(surfaceGroups(visiblePlaces, activeLayer), here, size.w, size.h, userScale, atlas.paths || []);
    }
    return { panX: 0, panY: 0 };
  }

  function stopPanAnim() {
    if (panAnimRaf && typeof cancelAnimationFrame === 'function') cancelAnimationFrame(panAnimRaf);
    panAnimRaf = 0;
  }

  function animatePanToYou({ durationMs = 280 } = {}) {
    const target = targetPanForYou();
    if (prefersReducedMotion() || !durationMs || durationMs <= 0) {
      stopPanAnim();
      panX = target.panX;
      panY = target.panY;
      scheduleDraw();
      return;
    }
    stopPanAnim();
    const fromX = panX;
    const fromY = panY;
    const dx = target.panX - fromX;
    const dy = target.panY - fromY;
    if (Math.abs(dx) < 0.5 && Math.abs(dy) < 0.5) {
      panX = target.panX;
      panY = target.panY;
      scheduleDraw();
      return;
    }
    const start = (typeof performance !== 'undefined' ? performance.now() : Date.now());
    const ease = (t) => 1 - Math.pow(1 - t, 3);
    const step = (now) => {
      const t = Math.min(1, (now - start) / durationMs);
      const e = ease(t);
      panX = fromX + dx * e;
      panY = fromY + dy * e;
      scheduleDraw();
      if (t < 1) {
        panAnimRaf = requestAnimationFrame(step);
      } else {
        panAnimRaf = 0;
        panX = target.panX;
        panY = target.panY;
        scheduleDraw();
      }
    };
    panAnimRaf = requestAnimationFrame(step);
  }

  function startTravel(targetId) {
    if (!currentRoomId || targetId === currentRoomId || overlayFading) return;
    const path = findPath(currentRoomId, targetId);
    if (!path || !path.length) return;
    if (closeAfterTravelTimer) {
      clearTimeout(closeAfterTravelTimer);
      closeAfterTravelTimer = 0;
    }
    travelTargetId = targetId;
    travelPath = path;
    isTraveling = true;
    travelFollowActive = true;
    travelPathRoomIds = new Set([currentRoomId, ...path.map(s => s.roomId)]);
    // Leave world-fit so the camera can track the walker.
    frameWorld = false;
    animatePanToYou({ durationMs: prefersReducedMotion() ? 0 : 320 });
    if (sendMessage) sendMessage(path[0].direction);
  }

  function cancelTravel({ arrived = false } = {}) {
    travelTargetId = null;
    travelPath = [];
    isTraveling = false;
    travelPathRoomIds = new Set();
    if (!arrived) travelFollowActive = false;
  }

  function finishTravelAndMaybeCloseOverlay() {
    cancelTravel({ arrived: true });
    travelFollowActive = false;
    // Overlay mode only: fade out Cartographer and close. Compact Map tab stays open.
    if (!open || overlayFading) return;
    overlayFading = true;
    const delay = prefersReducedMotion() ? 60 : 520;
    closeAfterTravelTimer = setTimeout(() => {
      closeAfterTravelTimer = 0;
      overlayFading = false;
      closeOverview();
    }, delay);
  }

  function advanceTravel(newRoomId) {
    if (!isTraveling || !travelPath.length) {
      cancelTravel();
      return;
    }
    if (travelPath[0].roomId === newRoomId) {
      travelPath = travelPath.slice(1);
      travelPathRoomIds = new Set([
        newRoomId,
        ...travelPath.map(s => s.roomId),
        ...(travelTargetId ? [travelTargetId] : []),
      ]);
      if (!travelPath.length) {
        // Final step arrived — linger briefly on destination, then fade overlay.
        animatePanToYou({ durationMs: prefersReducedMotion() ? 0 : 240 });
        finishTravelAndMaybeCloseOverlay();
      } else {
        setTimeout(() => { if (sendMessage && travelPath[0]) sendMessage(travelPath[0].direction); }, 160);
      }
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
      frameWorld,
      travelPathRoomIds,
      travelTargetId,
      selectedId,
      turnInMarkers,
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
      if (Math.abs(dx) <= h.half && Math.abs(dy) <= h.half) return h.place;
    }
    return null;
  }

  function pointerDist() {
    const pts = [...pointers.values()];
    if (pts.length < 2) return 0;
    return Math.hypot(pts[0].x - pts[1].x, pts[0].y - pts[1].y);
  }

  function pointerDown(e) {
    pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    try { e.currentTarget.setPointerCapture(e.pointerId); } catch (err) { /* ignore */ }
    if (pointers.size >= 2) {
      isPanning = false;
      didDrag = true;
      pinchStart = { dist: pointerDist() || 1, scale: userScale };
      return;
    }
    isPanning = true;
    didDrag = false;
    panStart = { x: e.clientX, y: e.clientY, panX, panY };
  }

  function pointerMove(e, canvas) {
    if (pointers.has(e.pointerId)) pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    if (pinchStart && pointers.size >= 2) {
      const d = pointerDist();
      if (pinchStart.dist > 8 && d > 8) {
        userScale = clampMapScale(pinchStart.scale * (d / pinchStart.dist));
        scheduleDraw();
      }
      return;
    }
    if (isPanning) {
      const dx = e.clientX - panStart.x;
      const dy = e.clientY - panStart.y;
      if (Math.abs(dx) > 3 || Math.abs(dy) > 3) didDrag = true;
      panX = panStart.panX + dx;
      panY = panStart.panY + dy;
      scheduleDraw();
      return;
    }
    if ($isMobile) return;
    const rect = canvas.getBoundingClientRect();
    const found = hitTest(canvas, e.clientX - rect.left, e.clientY - rect.top);
    if (found) {
      let text = found.discovered ? (found.name || found.id) : 'Uncharted';
      if (found.areaName && found.discovered) text += ' · ' + found.areaName;
      if (found.turnInLabel) text += ' · Turn in: ' + found.turnInLabel;
      if (found.current || isCurrentPlace(found.id, currentRoomId)) text += ' (you are here)';
      else if (found.discovered && !found.turnInLabel) text += ' · inspect';
      else text += ' · uncharted';
      tooltip = { visible: true, text, x: e.clientX - rect.left, y: e.clientY - rect.top };
    } else {
      tooltip = { ...tooltip, visible: false };
    }
  }

  function pointerUp(e, canvas) {
    pointers.delete(e.pointerId);
    if (pointers.size < 2) pinchStart = null;
    try { e.currentTarget.releasePointerCapture(e.pointerId); } catch (err) { /* ignore */ }
    if (isPanning && !didDrag) {
      const rect = canvas.getBoundingClientRect();
      const found = hitTest(canvas, e.clientX - rect.left, e.clientY - rect.top);
      if (found) {
        const now = Date.now();
        const dbl = !$isMobile && lastTap.id === found.id && now - lastTap.at < 420;
        lastTap = { id: found.id, at: now };
        if (dbl && found.discovered && found.id !== currentRoomId) {
          startTravel(found.id);
        } else {
          selectedId = found.id;
          intelExpanded = false;
          if (store && store.selectMapPlace) store.selectMapPlace(found.id);
        }
        scheduleDraw();
      }
    }
    isPanning = false;
    didDrag = false;
  }

  function toggleIntel() {
    intelExpanded = !intelExpanded;
  }

  function onWheel(e) {
    e.preventDefault();
    userScale = clampMapScale(userScale * (e.deltaY < 0 ? 1.12 : 0.89));
    scheduleDraw();
  }

  function resolveHerePlace() {
    return groupForRoom(surfaceGroups(visiblePlaces, activeLayer), currentRoomId) || null;
  }

  function applyRecenterToYou(keepScale = true) {
    if (!keepScale) userScale = 1;
    if (frameWorld) { panX = panY = 0; return; }
    const size = readStageSize(stageWrap);
    const here = resolveHerePlace();
    if (here && size.w >= 4 && size.h >= 4) {
      const pan = panToCenterPlace(surfaceGroups(visiblePlaces, activeLayer), here, size.w, size.h, userScale, atlas.paths || []);
      panX = pan.panX;
      panY = pan.panY;
    } else {
      panX = 0;
      panY = 0;
    }
  }

  function recenter() {
    frameWorld = false;
    applyRecenterToYou(false);
    scheduleDraw();
  }

  function fitWorld() {
    frameWorld = true;
    userScale = 1;
    panX = panY = 0;
    scheduleDraw();
  }

  function closeOverview() {
    if (closeAfterTravelTimer) {
      clearTimeout(closeAfterTravelTimer);
      closeAfterTravelTimer = 0;
    }
    overlayFading = false;
    travelFollowActive = false;
    stopPanAnim();
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
    stopTileListener = onMapTilesReady(() => scheduleDraw());
    ambientTimer = setInterval(() => {
      if(open && !document.hidden && !prefersReducedMotion()) scheduleDraw();
    },650);
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
    if (window.visualViewport) window.visualViewport.addEventListener('resize', resizeHandler);
  });

  onDestroy(() => {
    if (stopTileListener) stopTileListener();
    clearInterval(ambientTimer);
    if (escHandler) window.removeEventListener('keydown', escHandler);
    if (resizeHandler) window.removeEventListener('resize', resizeHandler);
    if (resizeHandler && window.visualViewport) window.visualViewport.removeEventListener('resize', resizeHandler);
    if (drawRaf && typeof cancelAnimationFrame === 'function') cancelAnimationFrame(drawRaf);
    stopPanAnim();
    if (closeAfterTravelTimer) clearTimeout(closeAfterTravelTimer);
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
  .map-overlay.fading {
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.48s ease;
  }
  .map-overlay:not(.fading) {
    transition: opacity 0.18s ease;
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
  .room-choices { max-height: 230px; overflow: auto; margin-bottom: 10px; }
  .room-choice, .room-link { display: flex; flex-direction: column; width: 100%; padding: 7px 8px; margin: 3px 0; border: 1px solid rgba(190,170,120,.22); border-radius: 4px; background: #1b252b; color: #e6dcc4; text-align: left; font: inherit; cursor: pointer; }
  .room-choice:hover, .room-choice.chosen, .room-link:hover { border-color: #c7ac69; background: #29352e; }
  .room-choice em { font-size: 10px; font-style: normal; color: #cdb77f; }
  .room-filter { width: 100%; box-sizing: border-box; margin-bottom: 6px; padding: 7px; border: 1px solid #475549; border-radius: 4px; background: #162128; color: #f0e3c4; font: inherit; }
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
  .sheet-handle { display: none; }
  .peek-head { display: contents; }
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
    background: #0a0c10;
  }
  .stage::after {
    content: '';
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 2;
    background: radial-gradient(ellipse at center, transparent 55%, rgba(0,0,0,0.28) 100%);
  }
  canvas {
    position: absolute;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
    cursor: grab;
    image-rendering: pixelated;
    touch-action: none;
  }
  canvas:active { cursor: grabbing; }
  .map-legend {
    position: absolute;
    left: 10px;
    bottom: 10px;
    z-index: 3;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 7px 9px;
    border-radius: 6px;
    background: rgba(8, 10, 14, 0.82);
    border: 1px solid rgba(212, 175, 55, 0.32);
    color: #e8dcc8;
    font-size: 10px;
    letter-spacing: 0.02em;
    pointer-events: none;
    box-shadow: 0 4px 14px rgba(0,0,0,0.35);
  }
  .legend-item { display: flex; align-items: center; gap: 6px; }
  .legend-swatch {
    width: 11px;
    height: 11px;
    border-radius: 50%;
    border: 1.5px solid transparent;
    flex: 0 0 auto;
  }
  .legend-swatch.you {
    background: #1a140c;
    border-color: #ffe69b;
    box-shadow: 0 0 8px rgba(255, 220, 120, 0.55);
  }
  .legend-swatch.turnin {
    background: #f5c542;
    border-color: #ffe08a;
    box-shadow: 0 0 8px rgba(245, 197, 66, 0.55);
  }
  .legend-swatch.selected {
    background: transparent;
    border-radius: 2px;
    border-color: #ffe29a;
    box-shadow: inset 0 0 0 1px rgba(255, 226, 154, 0.35);
  }
  .turnin-banner {
    position: absolute;
    top: 10px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 3;
    padding: 5px 10px;
    border-radius: 999px;
    background: rgba(48, 34, 8, 0.92);
    border: 1px solid rgba(245, 197, 66, 0.65);
    color: #ffe08a;
    font-size: 11px;
    font-weight: 600;
    pointer-events: none;
    white-space: nowrap;
    box-shadow: 0 4px 12px rgba(0,0,0,0.35);
  }
  .chip.turnin {
    border-color: #f5c542;
    color: #f5c542;
    background: rgba(245, 197, 66, 0.12);
  }
  .turnin-list {
    margin: 8px 0 2px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .turnin-row {
    display: flex;
    gap: 8px;
    align-items: flex-start;
    padding: 6px 8px;
    border-radius: 6px;
    background: rgba(245, 197, 66, 0.1);
    border: 1px solid rgba(245, 197, 66, 0.35);
    color: #fde68a;
    font-size: 12px;
  }
  .turnin-row strong { display: block; color: #fff3c4; font-size: 12px; }
  .turnin-row em { color: #d6b35c; font-style: normal; font-size: 11px; }
  .turnin-icon {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #f5c542;
    color: #1a1204;
    font-weight: 800;
    font-size: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto;
    margin-top: 1px;
  }
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
  .toolbar-title { white-space: nowrap; }

  @media (max-width: 768px) {
    .map-overlay.narrow {
      padding: 0 !important;
      align-items: stretch !important;
      justify-content: stretch !important;
      height: 100dvh !important;
    }
    .map-panel.narrow {
      width: 100% !important;
      height: 100% !important;
      min-width: 0 !important;
      min-height: 0 !important;
      max-width: none !important;
      max-height: none !important;
      border-radius: 0 !important;
      border: none !important;
      padding-top: env(safe-area-inset-top, 0px);
    }
    .map-body {
      position: relative;
      padding-bottom: calc(176px + env(safe-area-inset-bottom, 0px));
    }
    .toolbar {
      padding: 4px 6px;
      gap: 4px;
      min-height: 48px;
    }
    .toolbar-title { display: none; }
    .layer-tabs {
      flex: 1 1 auto;
      min-width: 0;
      overflow-x: auto;
      flex-wrap: nowrap;
      -webkit-overflow-scrolling: touch;
    }
    .icon-btn {
      min-width: 44px;
      min-height: 44px;
      justify-content: center;
      flex: 0 0 auto;
    }
    .layer-tab {
      min-height: 36px;
      padding: 6px 10px;
      font-size: 11px;
      flex: 0 0 auto;
    }
    .intel.sheet {
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
      width: 100% !important;
      flex: none !important;
      min-height: calc(176px + env(safe-area-inset-bottom, 0px));
      max-height: calc(176px + env(safe-area-inset-bottom, 0px));
      overflow: hidden;
      border-left: none;
      border-top: 1px solid rgba(212, 175, 55, 0.28);
      border-radius: 16px 16px 0 0;
      padding: 2px 14px calc(10px + env(safe-area-inset-bottom, 0px));
      z-index: 4;
      box-shadow: 0 -10px 28px rgba(0,0,0,0.45);
      overscroll-behavior: contain;
    }
    .intel.sheet.expanded {
      max-height: min(62dvh, 560px);
      overflow: auto;
    }
    .intel.sheet .intel-more {
      display: none;
    }
    .intel.sheet.expanded .intel-more {
      display: block;
    }
    .intel.sheet:not(.expanded) .intel-chips {
      max-height: 22px;
      overflow: hidden;
    }
    .peek-head {
      display: block;
      cursor: pointer;
      -webkit-tap-highlight-color: transparent;
    }
    .sheet-handle {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      border: none;
      background: transparent;
      padding: 6px 0 4px;
      cursor: pointer;
      touch-action: manipulation;
    }
    .sheet-grip {
      width: 42px;
      height: 4px;
      border-radius: 999px;
      background: rgba(212, 175, 55, 0.55);
    }
    .intel-title { font-size: 1.05rem; margin-bottom: 4px; }
    .travel-btn {
      margin-top: 8px;
      min-height: 48px;
      font-size: 15px;
      touch-action: manipulation;
    }
    .intel-empty { padding: 0.4em 0 0.8em; }
  }
</style>

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div
    id="map-overview-overlay"
    class="map-overlay"
    class:narrow={$isMobile}
    class:fading={overlayFading}
    style="position:fixed;top:0;left:0;width:100vw;height:100vh;z-index:200000;display:flex;align-items:center;justify-content:center;padding:12px;box-sizing:border-box;background:rgba(0,0,0,0.82);opacity:1;visibility:visible;"
    use:portalToBody
    role="dialog"
    aria-modal="true"
    aria-label="World map"
    on:click={(e) => { if (e.target === e.currentTarget) closeOverview(); }}
  >
    <div
      class="map-panel"
      class:narrow={$isMobile}
      bind:this={modalEl}
      style="position:relative;left:auto;right:auto;top:auto;bottom:auto;width:80vw;height:80vh;max-width:none;max-height:none;display:flex;flex-direction:column;overflow:hidden;background:#0b0e14;border:1px solid rgba(212,175,55,0.28);border-radius:8px;flex:none;opacity:1;visibility:visible;transform:none;"
      on:click|stopPropagation
    >
      <div class="toolbar">
        <i class="material-icons">explore</i>
        <span class="toolbar-title">Cartographer</span>
        {#if isTraveling}<span class="travel">Traveling…</span>{/if}
        {#if layers.length > 1}
          <div class="layer-tabs">
            {#each layers as layer}
              <button class="layer-tab" class:active={activeLayer === layer.id} on:click={() => selectLayer(layer.id)}>{layer.name}</button>
            {/each}
          </div>
        {/if}
        <span class="spacer"></span>
        <button class="icon-btn" title="Fit world" on:click={fitWorld}>
          <i class="material-icons">public</i>
        </button>
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
            on:pointercancel={(e) => pointerUp(e, stageCanvas)}
            on:pointerleave={() => tooltip = { ...tooltip, visible: false }}
            on:wheel={onWheel}
          ></canvas>
          <div class="map-legend" aria-hidden="true">
            <div class="legend-item"><span class="legend-swatch you"></span> You</div>
            <div class="legend-item"><span class="legend-swatch turnin"></span> Turn in</div>
            <div class="legend-item"><span class="legend-swatch selected"></span> Selected</div>
          </div>
          {#if turnInMarkers.length}
            <div class="turnin-banner">
              {turnInMarkers.length} quest{turnInMarkers.length === 1 ? '' : 's'} ready to turn in
            </div>
          {/if}
          {#if tooltip.visible}
            <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">{tooltip.text}</div>
          {/if}
        </div>
        <aside class="intel" class:sheet={$isMobile} class:expanded={intelExpanded} bind:this={intelEl}>
          {#if $isMobile}
            <button class="sheet-handle" type="button" on:click={toggleIntel} aria-label={intelExpanded ? 'Collapse intel' : 'Expand intel'}>
              <span class="sheet-grip"></span>
            </button>
          {/if}
          {#if !selectedPlace}
            <div class="intel-empty">Select a room on the chart.</div>
          {:else if !selectedPlace.discovered}
            <div class="peek-head" role="button" tabindex="0" on:click={$isMobile ? toggleIntel : undefined} on:keydown={(e) => { if ($isMobile && (e.key === 'Enter' || e.key === ' ')) { e.preventDefault(); toggleIntel(); } }}>
              <div class="intel-kicker">Uncharted</div>
              <h2 class="intel-title">Fog of war</h2>
            </div>
            <p class="intel-summary">{selectedPlace.summary || 'Walk closer to chart this ground.'}</p>
            <div class="intel-meta">Z {selectedPlace.z} · {selectedPlace.layer}</div>
          {:else}
            <div class="peek-head" role="button" tabindex="0" on:click={$isMobile ? toggleIntel : undefined} on:keydown={(e) => { if ($isMobile && (e.key === 'Enter' || e.key === ' ')) { e.preventDefault(); toggleIntel(); } }}>
              <div class="intel-kicker">{selectedPlace.areaName || selectedPlace.area || 'Unknown sector'}</div>
              <h2 class="intel-title">{selectedPlace.name || selectedPlace.id}</h2>
              <div class="intel-chips">
                <span class="chip danger-{selectedPlace.danger || 'low'}">{dangerLabel(selectedPlace.danger)}</span>
                <span class="chip">Z:{selectedPlace.z} {selectedPlace.layer}</span>
                <span class="chip">{selectedPlace.terrain || selectedPlace.biome || 'wild'}</span>
                <span class="chip">{selectedPlace.kind || 'place'}</span>
                {#if selectedPlace.current}<span class="chip you">You are here</span>{/if}
                {#if turnInMarkers.some(m => m.roomId === selectedPlace.id || isCurrentPlace(selectedPlace.id, m.roomId))}
                  <span class="chip turnin">Quest turn-in</span>
                {/if}
              </div>
            </div>
            {#if turnInMarkers.filter(m => m.roomId === selectedPlace.id || isCurrentPlace(selectedPlace.id, m.roomId)).length}
              <div class="turnin-list">
                {#each turnInMarkers.filter(m => m.roomId === selectedPlace.id || isCurrentPlace(selectedPlace.id, m.roomId)) as q}
                  <div class="turnin-row">
                    <span class="turnin-icon" aria-hidden="true">!</span>
                    <div>
                      <strong>{q.questName}</strong>
                      {#if q.npcName}<em> → {q.npcName}</em>{/if}
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
            <button class="travel-btn" type="button" on:click|stopPropagation={requestTravel} disabled={!canTravel || isTraveling}>
              {isTraveling ? 'Traveling…' : (canTravel ? 'Travel' : 'You are here')}
            </button>
            <div class="intel-more">
              {#if selectedStreet}
                <button class="room-link" type="button" on:click={() => selectRoom(selectedStreet)}>Back to {selectedStreet.name}</button>
              {/if}
              {#if townRooms.length}
                <div class="intel-section">{selectedPlace.town ? 'Buildings in town' : 'Inside this building'} · {townRooms.length}</div>
                {#if townRooms.length > 6}
                  <input class="room-filter" aria-label="Filter buildings" placeholder="Find a building…" bind:value={roomFilter} />
                {/if}
                <div class="room-choices">
                  {#each filteredTownRooms as room}
                    <button class="room-choice" class:chosen={room.id === selectedId} type="button" data-room-id={room.id} on:click={() => selectRoom(room)}>
                      <span>{room.name}</span>{#if isCurrentPlace(room.id, currentRoomId)}<em>You are here</em>{/if}
                    </button>
                  {/each}
                </div>
              {/if}
              {#if entranceRooms.length}
                <div class="intel-section">Below this entrance</div>
                {#each entranceRooms as room}
                  <button class="room-choice" type="button" data-entrance-id={room.id} on:click={() => selectRoom(room, true)}>{room.discovered ? room.name : 'Uncharted passage'} <em>Lower layer</em></button>
                {/each}
              {/if}
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
              {#if canTravel && !isTraveling}
                <div class="intel-hint">Double-click a room to travel, or use Travel above.</div>
              {/if}
            </div>
          {/if}
        </aside>
      </div>
    </div>
  </div>
{/if}
