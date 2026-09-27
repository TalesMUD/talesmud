import assert from 'assert';
import { hotbarSlotFromKey, isTextEntry, topOpenPanel } from './keyboardShortcuts.js';

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
