import assert from 'node:assert/strict';
import {
  normalizeBattleLayoutB,
  parseBattleLayoutOverride,
  battleLayoutFromSearch,
  battleLayoutFromStorage,
  resolveBattleLayoutB,
  writeBattleLayoutOverride,
  BATTLE_LAYOUT_STORAGE_KEY,
} from './battleLayout.js';

assert.equal(normalizeBattleLayoutB(true), true);
assert.equal(normalizeBattleLayoutB(false), false);
assert.equal(normalizeBattleLayoutB('1'), true);
assert.equal(normalizeBattleLayoutB(undefined), false);

assert.equal(parseBattleLayoutOverride('b'), true);
assert.equal(parseBattleLayoutOverride('1'), true);
assert.equal(parseBattleLayoutOverride('classic'), false);
assert.equal(parseBattleLayoutOverride('0'), false);
assert.equal(parseBattleLayoutOverride(''), null);
assert.equal(parseBattleLayoutOverride(null), null);

assert.equal(battleLayoutFromSearch('?battleLayout=b'), true);
assert.equal(battleLayoutFromSearch('?battlepoc=1'), true);
assert.equal(battleLayoutFromSearch('?battleLayout=classic'), false);
assert.equal(battleLayoutFromSearch('?battlepoc=0'), false);
assert.equal(battleLayoutFromSearch('?foo=1'), null);
assert.equal(battleLayoutFromSearch('?battleLayout=b&battlepoc=0'), true); // battleLayout first

const mem = {
  _d: Object.create(null),
  getItem(k) { return Object.prototype.hasOwnProperty.call(this._d, k) ? this._d[k] : null; },
  setItem(k, v) { this._d[k] = String(v); },
  removeItem(k) { delete this._d[k]; },
};
assert.equal(battleLayoutFromStorage(mem), null);
writeBattleLayoutOverride(true, mem);
assert.equal(mem.getItem(BATTLE_LAYOUT_STORAGE_KEY), '1');
assert.equal(battleLayoutFromStorage(mem), true);
writeBattleLayoutOverride(false, mem);
assert.equal(battleLayoutFromStorage(mem), false);
writeBattleLayoutOverride(null, mem);
assert.equal(battleLayoutFromStorage(mem), null);

// URL beats localStorage beats settings
writeBattleLayoutOverride(false, mem);
assert.equal(
  resolveBattleLayoutB({ search: '?battleLayout=b', storage: mem, settingsFlag: false }),
  true
);
assert.equal(
  resolveBattleLayoutB({ search: '', storage: mem, settingsFlag: true }),
  false
);
writeBattleLayoutOverride(null, mem);
assert.equal(
  resolveBattleLayoutB({ search: '', storage: mem, settingsFlag: true }),
  true
);
assert.equal(
  resolveBattleLayoutB({ search: '', storage: mem, settingsFlag: false }),
  false
);

console.log('battleLayout_test: ok');
