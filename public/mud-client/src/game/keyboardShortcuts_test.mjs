import assert from 'assert';
import { combatStageFocus, hotbarSlotFromKey, isTextEntry, topOpenPanel, prefersReducedMotion } from './keyboardShortcuts.js';
import { settingsStore } from './SettingsStore.js';

globalThis.localStorage = { setItem() {}, getItem() { return null; } };
globalThis.window = { matchMedia: () => ({ matches: true }) };
assert.equal(prefersReducedMotion(), true, 'system follows OS preference');
settingsStore.setSetting('interface', 'reducedMotion', 'off');
assert.equal(prefersReducedMotion(), false, 'off overrides reduced OS motion');
window.matchMedia = () => ({ matches: false });
settingsStore.setSetting('interface', 'reducedMotion', 'on');
assert.equal(prefersReducedMotion(), true, 'on overrides normal OS motion');
settingsStore.setSetting('interface', 'reducedMotion', 'system');
assert.equal(prefersReducedMotion(), false);

assert.equal(hotbarSlotFromKey('1'), 0);
assert.equal(hotbarSlotFromKey('9'), 8);
assert.equal(hotbarSlotFromKey('0'), -1);
assert.equal(hotbarSlotFromKey('a'), -1);
assert.equal(hotbarSlotFromKey('Enter'), -1);

assert.equal(isTextEntry({ tagName: 'TEXTAREA', isContentEditable: false, closest: () => null }), true);
assert.equal(isTextEntry({ tagName: 'DIV', isContentEditable: false, closest: () => null }), false);
assert.equal(isTextEntry({ tagName: 'DIV', isContentEditable: false, closest: () => ({}) }), true);

assert.equal(topOpenPanel({ accountMenu: true, cheatSheet: true }), 'cheatSheet');
assert.equal(topOpenPanel({ editMode: true, widgetFocus: true }), 'widgetFocus');
assert.equal(topOpenPanel({}), '');
assert.equal(topOpenPanel({ battleOutcome: true, inventory: true }), 'inventory');

console.log('keyboardShortcuts_test ok');

// Combat entry never steals command input, selection, or editable text.
for (const tag of ['INPUT', 'TEXTAREA', 'SELECT', 'DIV']) {
  const previous = { tagName: tag, isContentEditable: tag === 'DIV', isConnected: true, focus() { throw Error('stole typing focus'); } };
  const doc = { activeElement: previous, body: {} };
  const stage = { ownerDocument: doc, contains: () => false, focus() { throw Error('stole typing focus'); } };
  combatStageFocus(stage).destroy();
}
{
  const doc = { body: {}, activeElement: null };
  const previous = { tagName: 'BUTTON', isConnected: true, focus(options) { assert.equal(options.preventScroll, true); doc.activeElement = this; } };
  const stage = { ownerDocument: doc, contains: e => e === stage, focus(options) { assert.equal(options.preventScroll, true); doc.activeElement = this; } };
  doc.activeElement = previous;
  const action = combatStageFocus(stage);
  assert.equal(doc.activeElement, stage);
  assert.equal(isTextEntry(stage), false, '1–9 and Tab remain available on the stage');
  action.destroy();
  assert.equal(doc.activeElement, previous);
  const next = combatStageFocus(stage);
  const input = { tagName: 'INPUT' };
  doc.activeElement = input;
  next.destroy();
  assert.equal(doc.activeElement, input, 'leave preserves a newly focused input');
}
console.log('combat keyboard focus ok');
