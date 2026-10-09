// Drift page helpers. The list comes from GET /api/health/drift.
// Export all concatenates the per-entity YAML from /api/health/drift/export.

const tabs = {
  room: "rooms",
  npc: "npcs",
  item: "item-templates",
  dialog: "dialogs",
  quest: "quests",
  script: "scripts",
  skill: "skills",
  charactertemplate: "character-templates",
};

export function normDriftType(entityType) {
  return String(entityType || "").toLowerCase().replaceAll("_", "").replaceAll("-", "");
}

export function driftEntityPath(entityType, id) {
  if (!id) return "";
  const tab = tabs[normDriftType(entityType)];
  return tab ? `/creator/${tab}?id=${encodeURIComponent(id)}` : "";
}

export function canExportDrift(row) {
  return !!(row && row.kind !== "removed" && row.type && row.id);
}

export function driftCounts(changes) {
  const counts = { added: 0, changed: 0, removed: 0 };
  for (const row of changes || []) {
    if (Object.prototype.hasOwnProperty.call(counts, row?.kind)) counts[row.kind] += 1;
  }
  return counts;
}

export function formatDriftValue(value) {
  if (value === undefined || value === null || value === "") return "—";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

export function filterDrift(changes, { kind = "all", query = "" } = {}) {
  const needle = String(query || "").trim().toLowerCase();
  return (changes || []).filter((row) => {
    if (!row) return false;
    if (kind && kind !== "all" && row.kind !== kind) return false;
    if (!needle) return true;
    const fields = (row.fields || [])
      .map((field) => `${field?.path || ""} ${formatDriftValue(field?.before)} ${formatDriftValue(field?.after)}`)
      .join(" ");
    return `${row.type || ""} ${row.id || ""} ${row.name || ""} ${row.kind || ""} ${fields}`.toLowerCase().includes(needle);
  });
}

// One multi-document YAML file. Each block is labeled with the entity.
export function yamlBundle(parts) {
  const blocks = [];
  for (const part of parts || []) {
    const type = String(part?.type || "").trim();
    const id = String(part?.id || "").trim();
    if (!type || !id) continue;
    const body = String(part?.text || "").replace(/^\uFEFF/, "").trim();
    blocks.push(`# ${type} ${id}\n${body}`);
  }
  if (blocks.length === 0) return "";
  return `${blocks.join("\n---\n")}\n`;
}
