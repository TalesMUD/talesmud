// Process-wide WebSocket gate. Survives Game.svelte remounts so two
// component instances cannot dual-open (that is the live replace loop).

export const WS_CLOSE_SESSION_REPLACED = 4001;

let live = null;
let inflight = false;
let takenOver = false;
let generation = 0;

export function wsTakenOver() {
  return takenOver;
}

export function markWsTakenOver() {
  takenOver = true;
  inflight = false;
}

export function wsBusy() {
  if (takenOver) return true;
  if (inflight) return true;
  if (!live) return false;
  return live.readyState === WebSocket.CONNECTING || live.readyState === WebSocket.OPEN;
}

/** Claim the right to call `new WebSocket`. Must run before any store writes. */
export function beginWsConnect() {
  if (takenOver) return false;
  if (wsBusy()) return false;
  inflight = true;
  generation += 1;
  return true;
}

export function wsGeneration() {
  return generation;
}

export function registerWs(socket) {
  live = socket;
  inflight = false;
}

export function clearWs(socket) {
  if (live === socket) live = null;
  inflight = false;
}

export function liveSocket() {
  return live;
}

export function closeLive() {
  const s = live;
  live = null;
  inflight = false;
  if (!s) return;
  if (s.readyState === WebSocket.CONNECTING || s.readyState === WebSocket.OPEN) {
    try {
      s.close();
    } catch (e) {
      /* ignore */
    }
  }
}
