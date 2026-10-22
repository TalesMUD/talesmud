import test from "node:test";
import assert from "node:assert/strict";
import {
  migrateThemeSettings,
  resolveAppliedTheme,
} from "./themeChoice.js";

test("old dark-fantasy without an explicit flag becomes auto and follows the server", () => {
  const saved = migrateThemeSettings("dark-fantasy", undefined);
  assert.deepEqual(saved, { theme: "auto", themeExplicit: false });
  assert.equal(resolveAppliedTheme(saved.theme, "gilded"), "gilded");
});

test("an explicit dark-fantasy choice wins over the server default", () => {
  const saved = migrateThemeSettings("dark-fantasy", true);
  assert.equal(saved.theme, "dark-fantasy");
  assert.equal(saved.themeExplicit, true);
  assert.equal(resolveAppliedTheme(saved.theme, "gilded"), "dark-fantasy");
});

test("a saved clean-hud was a real choice and stays put", () => {
  const saved = migrateThemeSettings("clean-hud", undefined);
  assert.deepEqual(saved, { theme: "clean-hud", themeExplicit: true });
  assert.equal(resolveAppliedTheme(saved.theme, "gilded"), "clean-hud");
});

test("auto and an unknown server theme fall back to dark-fantasy", () => {
  assert.equal(resolveAppliedTheme("auto", "gilded"), "gilded");
  assert.equal(resolveAppliedTheme("auto", "neon"), "dark-fantasy");
  assert.equal(resolveAppliedTheme("auto", ""), "dark-fantasy");
  assert.equal(resolveAppliedTheme(undefined, undefined), "dark-fantasy");
});

test("an unknown stored theme migrates to auto", () => {
  const saved = migrateThemeSettings("lantern", false);
  assert.equal(saved.theme, "auto");
  assert.equal(resolveAppliedTheme(saved.theme, "clean-hud"), "clean-hud");
});
