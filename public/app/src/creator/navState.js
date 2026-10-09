// Creator sidebar state. Plain module so node:test can cover it without Svelte.

export const NAV_STORAGE_KEY = "tales.creator.nav.v1";

// Folded groups are stored as { [groupId]: true }. Missing keys are open.
export function defaultNavState() {
  return { collapsed: false, groups: {} };
}

export function loadNavState(storage) {
  if (!storage) return defaultNavState();
  try {
    const raw = storage.getItem(NAV_STORAGE_KEY);
    if (!raw) return defaultNavState();
    const parsed = JSON.parse(raw);
    const groups = {};
    if (parsed && typeof parsed.groups === "object" && parsed.groups) {
      for (const [key, value] of Object.entries(parsed.groups)) {
        if (value === true) groups[key] = true;
      }
    }
    return {
      collapsed: parsed?.collapsed === true,
      groups,
    };
  } catch {
    return defaultNavState();
  }
}

export function saveNavState(storage, state) {
  if (!storage) return;
  const groups = {};
  const source = state?.groups || {};
  for (const [key, value] of Object.entries(source)) {
    if (value === true) groups[key] = true;
  }
  storage.setItem(
    NAV_STORAGE_KEY,
    JSON.stringify({
      collapsed: state?.collapsed === true,
      groups,
    })
  );
}

export function toggleCollapsed(state) {
  return { ...state, collapsed: !state.collapsed, groups: { ...state.groups } };
}

export function toggleGroup(state, groupId) {
  const groups = { ...state.groups };
  if (groups[groupId] === true) delete groups[groupId];
  else groups[groupId] = true;
  return { ...state, groups };
}

export function withActiveGroupOpen(state, groups, pathname) {
  const id = activeGroupId(groups, pathname);
  if (!id || state.groups[id] !== true) return state;
  const nextGroups = { ...state.groups };
  delete nextGroups[id];
  return { ...state, groups: nextGroups };
}

export const CREATOR_NAV = [
  {
    id: "world",
    label: "World",
    items: [
      { id: "rooms", label: "Rooms", href: "/creator/rooms", icon: "door_open" },
      { id: "world", label: "World / Zones", href: "/creator/world", icon: "map", title: "World map" },
      { id: "health", label: "Health", href: "/creator/health", icon: "monitor_heart" },
    ],
  },
  {
    id: "actors",
    label: "Actors",
    items: [
      { id: "npcs", label: "NPCs", href: "/creator/npcs", icon: "group" },
      {
        id: "character-templates",
        label: "Character Templates",
        href: "/creator/character-templates",
        icon: "badge",
      },
      {
        id: "items",
        label: "Items",
        href: "/creator/item-templates",
        icon: "inventory_2",
        title: "Item templates",
      },
    ],
  },
  {
    id: "narrative",
    label: "Narrative",
    items: [
      { id: "dialogs", label: "Dialogs", href: "/creator/dialogs", icon: "forum" },
      { id: "dialog-graph", label: "Dialog Graph", href: "/creator/dialog-graph", icon: "account_tree" },
      { id: "quests", label: "Quests", href: "/creator/quests", icon: "flag" },
    ],
  },
  {
    id: "systems",
    label: "Systems",
    items: [
      { id: "skills", label: "Skills", href: "/creator/skills", icon: "bolt" },
      { id: "scripts", label: "Scripts", href: "/creator/scripts", icon: "code" },
      { id: "settings", label: "Settings", href: "/creator/settings", icon: "settings" },
    ],
  },
  {
    id: "operate",
    label: "Operate",
    items: [
      { id: "players", label: "Players", href: "/creator/players", icon: "person", admin: true },
      { id: "audit", label: "Audit log", href: "/creator/audit", icon: "history" },
      { id: "drift", label: "Drift", href: "/creator/drift", icon: "difference" },
    ],
  },
];

export function visibleNav(groups, isAdmin) {
  return (groups || [])
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => !item.admin || isAdmin),
    }))
    .filter((group) => group.items.length > 0);
}

export function normalizePath(pathname) {
  const raw = String(pathname || "").split("?")[0].split("#")[0];
  if (raw.length > 1 && raw.endsWith("/")) return raw.slice(0, -1);
  return raw || "/";
}

export function isNavActive(pathname, item) {
  const path = normalizePath(pathname);
  const href = item?.href || "";
  if (href === "/creator/rooms") {
    return path === "/creator" || path === "/creator/rooms";
  }
  return path === href || (href !== "" && path.startsWith(`${href}/`));
}

export function activeGroupId(groups, pathname) {
  for (const group of groups || []) {
    for (const item of group.items) {
      if (isNavActive(pathname, item)) return group.id;
    }
  }
  return "";
}

export function isTypingTarget(node) {
  let el = node;
  for (let guard = 0; el && guard < 25; guard += 1) {
    if (el.nodeType && el.nodeType !== 1) {
      el = el.parentElement || el.parentNode;
      continue;
    }
    const tag = String(el.tagName || "").toLowerCase();
    if (tag === "input" || tag === "textarea" || tag === "select") return true;
    if (el.isContentEditable) return true;
    const cls = typeof el.className === "string" ? el.className : "";
    if (/(?:^|\s)(?:CodeMirror|cm-editor|cm-content)(?:\s|$)/.test(cls)) return true;
    el = el.parentElement || null;
  }
  return false;
}

// Ctrl+B / Cmd+B toggles the rail. Shift and Alt stay with the browser.
export function shouldToggleRail(event) {
  if (!event || event.altKey || event.shiftKey || event.repeat) return false;
  if (String(event.key || "").toLowerCase() !== "b") return false;
  if (!(event.ctrlKey || event.metaKey)) return false;
  if (isTypingTarget(event.target)) return false;
  return true;
}
