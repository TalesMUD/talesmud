import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { THEME_IDS, ENGINE_DEFAULT_THEME, migrateThemeSettings, resolveAppliedTheme } from './themeChoice.js';
import { contrastRatio } from './mobile/phoneSheets.js';
const read = name => readFileSync(new URL(name, import.meta.url), 'utf8');
const css = read('../../public/themes/quiet-dark/theme.css');

test('quiet-dark is selectable and follows deployment choice without changing engine fallback', () => {
  assert.ok(THEME_IDS.includes('quiet-dark'));
  assert.equal(ENGINE_DEFAULT_THEME, 'dark-fantasy');
  assert.equal(resolveAppliedTheme('auto', 'quiet-dark'), 'quiet-dark');
  assert.equal(resolveAppliedTheme('auto', undefined), 'dark-fantasy');
  for (const theme of THEME_IDS) {
    assert.equal(resolveAppliedTheme(theme, 'quiet-dark'), theme);
    assert.deepEqual(migrateThemeSettings(theme, true), { theme, themeExplicit: true });
  }
  assert.match(read('ui/SettingsModal.svelte'), /set\('theme', 'quiet-dark'\)/);
  assert.match(read('../../public/index.html'), /themes\/quiet-dark\/theme\.css\?v=rank1/);
  assert.doesNotMatch(read('../../public/index.html'), /phone1/);
});

test('quiet-dark touch controls and both hotbars retain at least 44px targets', () => {
  const touch = css.slice(css.indexOf('@media (pointer: coarse)'));
  assert.match(touch, /max-width: 1024px/);
  assert.match(touch, /min-width: 44px !important/);
  assert.match(touch, /min-height: 44px !important/);
  assert.match(touch, /\.combat-hotbar \.hb-slot/);
  assert.match(touch, /\.combat-hotbar \{ flex-wrap: wrap !important/);
  assert.match(css, /font-size: 14px/);
  assert.match(css, /prefers-reduced-motion: reduce/);
});

test('quiet-dark text palette clears 4.5:1 on all tonal surfaces', () => {
  for (const foreground of ['#ece6dc', '#b5ada1', '#d9b46a', '#e7796c', '#7cc27a', '#7aa6e0']) {
    assert.ok(css.includes(foreground));
    for (const background of ['#0d0c0b', '#151312', '#1c1a18', '#25221f']) {
      assert.ok(contrastRatio(foreground, background) >= 4.5, `${foreground} on ${background}`);
    }
  }
  assert.ok(contrastRatio('#1a1408', '#d9b46a') >= 4.5);
});
