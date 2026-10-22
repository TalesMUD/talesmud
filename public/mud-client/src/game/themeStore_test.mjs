import test from "node:test";
import assert from "node:assert/strict";

const memory = new Map();
globalThis.localStorage = {
  getItem(key) {
    return memory.has(key) ? memory.get(key) : null;
  },
  setItem(key, value) {
    memory.set(key, String(value));
  },
  removeItem(key) {
    memory.delete(key);
  },
};
globalThis.document = { body: { dataset: {} } };

memory.set(
  "talesmud_settings_v1",
  JSON.stringify({
    version: 1,
    interface: { theme: "dark-fantasy" },
  })
);

const { settingsStore } = await import("./SettingsStore.js");

test("a saved dark-fantasy that was never chosen migrates to auto", () => {
  assert.equal(settingsStore.getSetting("interface", "theme"), "auto");
  assert.equal(settingsStore.getSetting("interface", "themeExplicit"), false);
  const saved = JSON.parse(memory.get("talesmud_settings_v1"));
  assert.equal(saved.interface.theme, "auto");
  assert.equal(saved.interface.themeExplicit, false);
});

test("auto paints the server default, and an explicit theme wins", () => {
  settingsStore.setServerDefault("gilded");
  assert.equal(document.body.dataset.theme, "gilded");
  settingsStore.setSetting("interface", "theme", "dark-fantasy");
  assert.equal(settingsStore.getSetting("interface", "themeExplicit"), true);
  assert.equal(document.body.dataset.theme, "dark-fantasy");
  settingsStore.setServerDefault("gilded");
  assert.equal(document.body.dataset.theme, "dark-fantasy");
  settingsStore.setSetting("interface", "theme", "auto");
  assert.equal(settingsStore.getSetting("interface", "themeExplicit"), false);
  assert.equal(document.body.dataset.theme, "gilded");
});

test("an unknown server theme falls back to dark-fantasy while auto is stored", () => {
  settingsStore.setServerDefault("neon");
  assert.equal(document.body.dataset.theme, "dark-fantasy");
  assert.equal(settingsStore.getSetting("interface", "theme"), "auto");
});
