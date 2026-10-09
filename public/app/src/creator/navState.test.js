import assert from "node:assert/strict";
import test from "node:test";
import {
  CREATOR_NAV,
  activeGroupId,
  defaultNavState,
  isNavActive,
  isTypingTarget,
  loadNavState,
  saveNavState,
  shouldToggleRail,
  toggleCollapsed,
  toggleGroup,
  visibleNav,
  withActiveGroupOpen,
} from "./navState.js";

function memoryStorage(initial = {}) {
  const data = { ...initial };
  return {
    getItem(key) {
      return Object.prototype.hasOwnProperty.call(data, key) ? data[key] : null;
    },
    setItem(key, value) {
      data[key] = String(value);
    },
  };
}

test("loadNavState ignores empty and corrupt storage", () => {
  assert.deepEqual(loadNavState(memoryStorage()), defaultNavState());
  assert.deepEqual(loadNavState(null), defaultNavState());
  const broken = memoryStorage({ "tales.creator.nav.v1": "{" });
  assert.deepEqual(loadNavState(broken), defaultNavState());
});

test("collapsed rail and folded groups round-trip", () => {
  const storage = memoryStorage();
  const folded = toggleGroup(toggleCollapsed(defaultNavState()), "narrative");
  saveNavState(storage, folded);
  assert.deepEqual(loadNavState(storage), {
    collapsed: true,
    groups: { narrative: true },
  });
});

test("toggleGroup opens a folded group and leaves others", () => {
  const once = toggleGroup(defaultNavState(), "world");
  const twice = toggleGroup(once, "actors");
  assert.deepEqual(twice.groups, { world: true, actors: true });
  assert.deepEqual(toggleGroup(twice, "world").groups, { actors: true });
});

test("active group opens without unfolding the rest", () => {
  const state = {
    collapsed: false,
    groups: { world: true, narrative: true },
  };
  const next = withActiveGroupOpen(state, CREATOR_NAV, "/creator/quests?id=QST1");
  assert.equal(next.groups.world, true);
  assert.equal(next.groups.narrative, undefined);
  assert.equal(withActiveGroupOpen(next, CREATOR_NAV, "/creator/quests"), next);
});

test("rooms highlight covers /creator and dialog graph stays distinct", () => {
  const rooms = CREATOR_NAV[0].items[0];
  const dialogs = CREATOR_NAV[2].items[0];
  const graph = CREATOR_NAV[2].items[1];
  const quests = CREATOR_NAV[2].items[2];
  assert.equal(isNavActive("/creator", rooms), true);
  assert.equal(isNavActive("/creator/", rooms), true);
  assert.equal(isNavActive("/creator/rooms?id=R1", rooms), true);
  assert.equal(isNavActive("/creator/npcs", rooms), false);
  assert.equal(isNavActive("/creator/dialog-graph", dialogs), false);
  assert.equal(isNavActive("/creator/dialog-graph?id=DLG1", graph), true);
  assert.equal(isNavActive("/creator/quests/debug?id=QST1", quests), true);
  assert.equal(isNavActive("/creator/quests/debug", dialogs), false);
  assert.equal(activeGroupId(CREATOR_NAV, "/creator/health"), "world");
  assert.equal(activeGroupId(CREATOR_NAV, "/creator/quests/debug"), "narrative");
});

test("players stay admin-only and audit stays visible", () => {
  const creator = visibleNav(CREATOR_NAV, false);
  const operate = creator.find((group) => group.id === "operate");
  assert.deepEqual(
    operate.items.map((item) => item.id),
    ["audit"]
  );
  const admin = visibleNav(CREATOR_NAV, true).find((group) => group.id === "operate");
  assert.deepEqual(
    admin.items.map((item) => item.id),
    ["players", "audit"]
  );
});

test("rail shortcut ignores typing targets and unmodified keys", () => {
  const button = { tagName: "BUTTON", className: "", parentElement: null };
  assert.equal(shouldToggleRail({ key: "b", ctrlKey: true, target: button }), true);
  assert.equal(shouldToggleRail({ key: "B", metaKey: true, target: button }), true);
  assert.equal(shouldToggleRail({ key: "b", ctrlKey: true, shiftKey: true, target: button }), false);
  assert.equal(shouldToggleRail({ key: "b", altKey: true, metaKey: true, target: button }), false);
  assert.equal(shouldToggleRail({ key: "b", target: button }), false);
  assert.equal(shouldToggleRail({ key: "b", ctrlKey: true, repeat: true, target: button }), false);

  const input = { tagName: "INPUT", className: "", parentElement: null };
  const area = { tagName: "TEXTAREA", className: "", parentElement: null };
  const select = { tagName: "SELECT", className: "", parentElement: null };
  const editable = { tagName: "DIV", className: "", isContentEditable: true, parentElement: null };
  const mirror = {
    tagName: "SPAN",
    className: "cm-line",
    parentElement: { tagName: "DIV", className: "cm-editor", isContentEditable: false, parentElement: null },
  };
  const codeMirror = {
    tagName: "PRE",
    className: "",
    parentElement: { tagName: "DIV", className: "CodeMirror", parentElement: null },
  };
  for (const target of [input, area, select, editable, mirror, codeMirror]) {
    assert.equal(shouldToggleRail({ key: "b", ctrlKey: true, target }), false);
    assert.equal(isTypingTarget(target), true);
  }
});
