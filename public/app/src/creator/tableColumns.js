/**
 * Column definitions for entity data tables in the Creator UI.
 *
 * @typedef {Object} ColumnDef
 * @property {string} key         - dot-path to the value (e.g. "coords.x")
 * @property {string} label       - display header text
 * @property {string} [type]      - "text" | "number" | "select" | "boolean" | "computed"
 * @property {number} [width]     - optional width in px
 * @property {boolean} [sortable] - whether column is sortable (default true)
 * @property {boolean} [filterable] - whether column is filterable (default true)
 * @property {Function} [accessor] - (element) => display value
 * @property {Function} [sortAccessor] - (element) => sortable value
 * @property {string[]} [options]  - dropdown options for "select" type filters
 * @property {boolean} [mono]      - render in monospace font
 * @property {number} [priority]   - display priority (1=always shown, 2=shown when space allows, 3=hidden in compact mode)
 */

export const roomColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 180, priority: 1 },
  { key: "area", label: "Area", width: 160, type: "select", options: [], priority: 2 },
  {
    key: "coords", label: "Coords", width: 90, type: "computed", priority: 3,
    accessor: (el) => el.coords ? `${el.coords.x},${el.coords.y},${el.coords.z}` : "-",
    sortAccessor: (el) => el.coords ? el.coords.x * 10000 + el.coords.y * 100 + el.coords.z : 0,
    filterable: false,
  },
  {
    key: "_exitsCount", label: "Exits", width: 60, type: "number", priority: 3,
    accessor: (el) => el.exits?.length || 0,
    sortAccessor: (el) => el.exits?.length || 0,
    filterable: false,
  },
  {
    key: "_actionsCount", label: "Actions", width: 70, type: "number", priority: 3,
    accessor: (el) => el.actions?.length || 0,
    sortAccessor: (el) => el.actions?.length || 0,
    filterable: false,
  },
  {
    key: "_spawnersCount", label: "Spawners", width: 80, type: "number", priority: 3,
    accessor: (el) => el._spawnersCount ?? "-",
    sortAccessor: (el) => el._spawnersCount ?? -1,
    filterable: false,
  },
];

export const itemTemplateColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 180, priority: 1 },
  { key: "type", label: "Type", width: 110, type: "select", options: [], priority: 2 },
  { key: "subType", label: "Subtype", width: 110, type: "select", options: [], priority: 3 },
  {
    key: "quality", label: "Quality", width: 100, type: "select", priority: 2,
    options: ["normal", "magic", "rare", "legendary", "mythic"],
  },
  { key: "slot", label: "Slot", width: 100, type: "select", options: [], priority: 3 },
  { key: "level", label: "Lvl", width: 50, type: "number", priority: 2 },
];

export const npcColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 180, priority: 1 },
  {
    key: "_npcType", label: "NPC Type", width: 90, type: "select", priority: 1,
    options: ["Neutral", "Merchant", "Enemy"],
    accessor: (el) => {
      const hasEnemy = !!el.enemyTrait;
      const hasMerchant = !!el.merchantTrait;
      if (hasEnemy) return "Enemy";
      if (hasMerchant) return "Merchant";
      return "Neutral";
    },
  },
  {
    key: "race.name", label: "Race", width: 90, type: "select", options: [], priority: 2,
    accessor: (el) => el.race?.name || "-",
  },
  {
    key: "class.name", label: "Class", width: 90, type: "select", options: [], priority: 2,
    accessor: (el) => el.class?.name || "-",
  },
  { key: "level", label: "Lvl", width: 50, type: "number", priority: 2 },
  {
    key: "isTemplate", label: "Template", width: 80, type: "select", priority: 3,
    options: ["Template", "Unique"],
    accessor: (el) => el.isTemplate ? "Template" : "Unique",
  },
];

export const dialogColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 280, priority: 1 },
  {
    key: "_optionsCount", label: "Options", width: 70, type: "number", priority: 2,
    accessor: (el) => el.options?.length || 0,
    sortAccessor: (el) => el.options?.length || 0,
    filterable: false,
  },
];

export const scriptColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 280, priority: 1 },
  { key: "type", label: "Type", width: 130, type: "select", options: [], priority: 2 },
];

export const itemColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 180, priority: 1 },
  { key: "type", label: "Type", width: 110, type: "select", options: [], priority: 2 },
  { key: "subType", label: "Subtype", width: 110, type: "select", options: [], priority: 3 },
  {
    key: "quality", label: "Quality", width: 100, type: "select", priority: 2,
    options: ["normal", "magic", "rare", "legendary", "mythic"],
  },
  { key: "slot", label: "Slot", width: 100, type: "select", options: [], priority: 3 },
  { key: "level", label: "Lvl", width: 50, type: "number", priority: 2 },
];

export const questColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 200, priority: 1 },
  { key: "area", label: "Area", width: 140, type: "select", options: [], priority: 2 },
  {
    key: "category", label: "Category", width: 90, type: "select", priority: 2,
    options: ["main", "side", "daily"],
  },
  { key: "level", label: "Lvl", width: 50, type: "number", priority: 2 },
  {
    key: "source.type", label: "Source", width: 90, type: "select", priority: 2,
    options: ["npc", "item", "auto", "script"],
    accessor: (el) => el.source?.type || "-",
  },
  {
    key: "_objectiveCount", label: "Objectives", width: 80, type: "number", priority: 3,
    accessor: (el) => el.objectives?.length || 0,
    sortAccessor: (el) => el.objectives?.length || 0,
    filterable: false,
  },
  {
    key: "repeatable", label: "Repeat", width: 65, type: "boolean", priority: 3,
    accessor: (el) => el.repeatable ? "Yes" : "No",
  },
];

export const characterTemplateColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 200, priority: 1 },
  {
    key: "archetype", label: "Archetype", width: 100, type: "select", priority: 2,
    options: ["warrior", "rogue", "mage", "cleric", "ranger", "druid", "custom"],
  },
  { key: "level", label: "Lvl", width: 50, type: "number", priority: 2 },
];

export const skillColumns = [
  {
    key: "id", label: "ID", width: 90, mono: true, priority: 1,
    accessor: (el) => el.id?.length > 10 ? el.id.slice(0, 8) + "\u2026" : el.id,
  },
  { key: "name", label: "Name", width: 160, priority: 1 },
  {
    key: "_classes", label: "Classes", width: 140, type: "computed", priority: 1,
    accessor: (el) => el.classIds?.join(", ") || "-",
    sortAccessor: (el) => el.classIds?.[0] || "",
  },
  { key: "levelRequired", label: "Lvl", width: 50, type: "number", priority: 2 },
  {
    key: "resourceType", label: "Resource", width: 90, type: "select", priority: 2,
    options: ["mana", "cooldown"],
  },
  {
    key: "effect", label: "Effect", width: 80, type: "select", priority: 2,
    options: ["damage", "heal", "buff", "debuff", "dot", "hot"],
  },
  {
    key: "target", label: "Target", width: 90, type: "select", priority: 3,
    options: ["enemy", "self", "all_enemies"],
  },
  { key: "basePower", label: "Power", width: 60, type: "number", priority: 3 },
];
