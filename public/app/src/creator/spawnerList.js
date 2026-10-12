// Spawner list helpers. Plain module so node:test can cover filters without Svelte.

export function durationLabel(value) {
  if (value == null || value === "" || value === 0) return "";
  if (typeof value === "string") return value;
  const ns = Number(value);
  if (!Number.isFinite(ns) || ns <= 0) return "";
  let seconds = Math.round(ns / 1e9);
  const hours = Math.floor(seconds / 3600);
  seconds %= 3600;
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  let out = "";
  if (hours) out += `${hours}h`;
  if (minutes || hours) out += `${minutes}m`;
  if (secs || minutes || hours || !out) out += `${secs}s`;
  return out;
}

export function respawnLabel(spawner, template) {
  if (!spawner) return "";
  if (spawner.respawnTimeOverride) return durationLabel(spawner.respawnTimeOverride);
  if (spawner.respawnDelay) return durationLabel(spawner.respawnDelay);
  if (template?.respawnTime) return durationLabel(template.respawnTime);
  return "";
}

export function liveCountFor(spawner, live) {
  if (!spawner) return 0;
  let count = 0;
  for (const row of live || []) {
    if (!row || row.dead) continue;
    if (row.templateId !== spawner.templateId) continue;
    if (row.roomId !== spawner.roomId) continue;
    count += 1;
  }
  return count;
}

export function decorateSpawner(spawner, { roomsById = {}, npcsById = {}, live = null } = {}) {
  const room = roomsById[spawner?.roomId] || null;
  const npc = npcsById[spawner?.templateId] || null;
  return {
    ...spawner,
    roomName: room?.name || spawner?.roomId || "",
    zone: room?.area || "",
    templateName: npc?.name || spawner?.templateId || "",
    respawnLabel: respawnLabel(spawner, npc),
    liveCount: live == null ? null : liveCountFor(spawner, live),
  };
}

export function filterSpawners(spawners, { templateId = "", roomId = "", zone = "" } = {}) {
  return (spawners || []).filter((spawner) => {
    if (!spawner) return false;
    if (templateId && spawner.templateId !== templateId) return false;
    if (roomId && spawner.roomId !== roomId) return false;
    if (zone && (spawner.zone || "") !== zone) return false;
    return true;
  });
}

// A unique content NPC is not a template and not a spawned instance.
export function isUniqueContentNPC(npc) {
  if (!npc || npc.isTemplate) return false;
  if (npc.templateId) return false;
  if (npc.id && String(npc.id).includes("~")) return false;
  return Boolean(npc.id);
}

export function uniqueTemplatesWithoutSpawner(npcs, spawners) {
  const covered = new Set();
  for (const spawner of spawners || []) {
    if (spawner?.templateId) covered.add(spawner.templateId);
  }
  return (npcs || []).filter((npc) => isUniqueContentNPC(npc) && !covered.has(npc.id));
}

export function filterUniqueGaps(npcs, { templateId = "", roomId = "", zone = "", roomsById = {} } = {}) {
  return (npcs || []).filter((npc) => {
    if (templateId && npc.id !== templateId) return false;
    const home = npc.spawnRoomId || "";
    if (roomId && home !== roomId) return false;
    if (zone && (roomsById[home]?.area || "") !== zone) return false;
    return true;
  });
}
