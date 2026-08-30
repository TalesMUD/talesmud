// Pure layout helpers for the atlas widget.
// The canvas must never feed size back into its parent: only the bitmap
// (canvas.width/height) changes, never canvas.style or the wrap box.

export function readStageSize(wrap) {
  if (!wrap) return { w: 0, h: 0 };
  const w = Math.max(0, Math.floor(wrap.clientWidth || 0));
  const h = Math.max(0, Math.floor(wrap.clientHeight || 0));
  return { w, h };
}

export function shouldRepaintSize(prev, next) {
  if (!next || next.w < 1 || next.h < 1) return false;
  if (!prev) return true;
  return prev.w !== next.w || prev.h !== next.h;
}

export function applyCanvasBitmap(canvas, w, h, dpr) {
  if (!canvas || w < 1 || h < 1) return false;
  const scale = dpr > 0 ? dpr : 1;
  const pw = Math.max(1, Math.round(w * scale));
  const ph = Math.max(1, Math.round(h * scale));
  let changed = false;
  if (canvas.width !== pw) {
    canvas.width = pw;
    changed = true;
  }
  if (canvas.height !== ph) {
    canvas.height = ph;
    changed = true;
  }
  return changed;
}

// Models the old bug: writing CSS pixel size from getBoundingClientRect
// back onto an in-flow canvas makes the parent grow every frame.
export function applyInFlowStyleSize(wrap, cssHeight) {
  const grown = Math.ceil(cssHeight + 0.4);
  wrap.y += grown - wrap.clientHeight;
  wrap.clientHeight = grown;
  return wrap;
}
