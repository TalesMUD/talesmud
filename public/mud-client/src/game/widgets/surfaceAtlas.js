// Surface interiors remain real atlas rooms for selection and travel. Their
// render group uses the nearest exterior anchor supplied by the compiler.
export function surfaceId(place) {
  return place?.layer === 'overworld' && place.mapRole === 'interior'
    ? place.surfaceRoomId || place.id : place?.id;
}
export function surfaceGroups(places, layer = 'overworld') {
  const list = (places || []).filter(p => p.layer === layer);
  if (layer !== 'overworld') return list.map(p => ({ ...p, members: [p] }));
  const byId = new Map(list.map(p => [p.id, p]));
  const grouped = new Map();
  for (const place of list) {
    const id = surfaceId(place);
    if (!grouped.has(id)) {
      const anchor = byId.get(id) || place;
      grouped.set(id, { ...anchor, id, members: [] });
    }
    grouped.get(id).members.push(place);
  }
  return [...grouped.values()];
}
export function groupForRoom(groups, id) {
  const template = String(id || '').split('~')[0];
  return groups.find(g => g.id === id || g.id === template || g.members.some(p => p.id === id || p.id === template));
}
export function primaryGroupPlace(group, currentRoomId) {
  if (group.discovered) return group;
  const template = String(currentRoomId || '').split('~')[0];
  const known = group.members.find(p => p.discovered && p.id === template) || group.members.find(p => p.discovered);
  return known ? { ...known, x: group.x, y: group.y } : group;
}
export function interiorChoices(places, selected) {
  if (!selected || !selected.discovered || selected.layer !== 'overworld') return [];
  const anchorId = surfaceId(selected);
  return (places || []).filter(p => p.discovered && p.layer === 'overworld' && p.mapRole === 'interior' &&
    (selected.town && p.area === selected.area || surfaceId(p) === anchorId))
    .sort((a,b) => String(a.name || a.id).localeCompare(String(b.name || b.id)));
}
export function outdoorRoads(places, paths) {
  const byId = new Map((places || []).map(p => [p.id,p]));
  const seen = new Set(), roads = [];
  const compass = new Set(['north','south','east','west','northeast','northwest','southeast','southwest']);
  for (const path of paths || []) {
    const a = byId.get(path.from), b = byId.get(path.to);
    if (!a || !b || !a.discovered || !b.discovered || a.layer !== 'overworld' || b.layer !== 'overworld' ||
      a.mapRole === 'interior' || b.mapRole === 'interior' || path.hidden || ['hidden','stair','passage'].includes(path.kind) || !compass.has(String(path.dir).toLowerCase())) continue;
    const key = [a.id,b.id].sort().join('|');
    if (!seen.has(key) && (a.x !== b.x || a.y !== b.y)) { seen.add(key); roads.push({a,b}); }
  }
  return roads;
}
