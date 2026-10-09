// Reachability colors for the world-map tiles and exits.
// Callers pass the layer state in. Svelte does not subscribe to values
// that a template reads only inside a function, so a closed-over flag
// leaves the zone outline in place after the report loads.

export function roomMark(room, on, report, byId) {
  if (!on || !report) return "";
  const row = byId?.get(room?.id);
  if (!row) return "";
  if (!row.reachable) return "unreachable";
  if (row.instance) return "instance";
  return "reachable";
}

export function islandPinned(roomId, on, index, report) {
  if (!on || index < 0) return false;
  const island = report?.islands?.[index];
  return !!island?.roomIds?.includes(roomId);
}

export function edgeLook(edge, on, report, byId) {
  if (!on || !report) {
    return {
      color: edge.isCrossZone ? "#e06040" : (edge.isCardinal ? "#888" : "#b08050"),
      dash: edge.isCrossZone ? "8,4" : (edge.isCardinal ? "none" : "4,3"),
      width: edge.isCrossZone ? 2 : 1.5,
    };
  }
  const source = byId?.get(edge.sourceId);
  const target = byId?.get(edge.targetId);
  const unreachable = (source && !source.reachable) || (target && !target.reachable);
  const instance = !unreachable && (source?.instance || target?.instance);
  let color = "#16a34a";
  if (unreachable) color = "#ef4444";
  else if (instance) color = "#6366f1";
  let dash = "none";
  if (edge.isHidden) dash = "5 4";
  else if (!edge.isBidirectional) dash = "2 4";
  return { color, dash, width: unreachable ? 2 : 1.5 };
}
