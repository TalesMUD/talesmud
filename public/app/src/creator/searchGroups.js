const labels = {
  room: "Rooms",
  npc: "NPCs",
  item: "Items",
  lootTable: "Loot tables",
  spawner: "Spawners",
  dialog: "Dialogs",
  quest: "Quests",
  script: "Scripts",
  skill: "Skills",
  characterTemplate: "Character templates",
};

const backlinkTypes = new Set(["room", "npc", "item", "dialog", "quest", "script"]);

const refTypes = new Set([
  "room",
  "npc",
  "item",
  "dialog",
  "quest",
  "script",
  "skill",
  "spawner",
  "loottable",
  "loot-table",
  "loottables",
  "charactertemplate",
  "character-template",
  "character_template",
]);

function typeLabel(type) {
  return labels[type] || type || "Other";
}

function groupHits(hits) {
  const groups = [];
  const byType = new Map();
  (hits || []).forEach((hit, index) => {
    const type = hit?.type || "";
    let group = byType.get(type);
    if (!group) {
      group = { type, hits: [] };
      byType.set(type, group);
      groups.push(group);
    }
    group.hits.push({ ...hit, index });
  });
  return groups;
}

function isTypingTarget(target) {
  if (!target || typeof target.closest !== "function") return false;
  if (target.closest("input, textarea, select, [contenteditable], .cm-editor, .CodeMirror, .codejar-wrap")) {
    return true;
  }
  return !!target.isContentEditable;
}

function inboundCount(view) {
  let count = 0;
  for (const group of view?.inbound || []) {
    count += (group.refs || []).length;
  }
  return count;
}

function referencedByHint(view, limit = 4) {
  const refs = [];
  for (const group of view?.inbound || []) {
    for (const ref of group.refs || []) refs.push(ref);
  }
  if (refs.length === 0) return "";
  const shown = refs.slice(0, limit).map((ref) => `${ref.type} ${ref.name || ref.id}`);
  const extra = refs.length > limit ? ", …" : "";
  return `Referenced by ${refs.length}: ${shown.join(", ")}${extra}`;
}

export { backlinkTypes, groupHits, inboundCount, isTypingTarget, refTypes, referencedByHint, typeLabel };
